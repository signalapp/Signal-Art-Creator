//
// Copyright 2023 Signal Messenger, LLC
// SPDX-License-Identifier: AGPL-3.0-only
//

package dlmux

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/DataDog/datadog-go/v5/statsd"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/rs/zerolog"
	"github.com/valyala/fasthttp"
)

const maxManifestSize int = 10 * 1024
const maxStickerSize int = 300 * 1024

const signatureAlgorithm string = "AWS4-HMAC-SHA256"

type UploadItem struct {
	Id            int    `json:"id"`
	Key           string `json:"key"`
	Credential    string `json:"credential"`
	Acl           string `json:"acl"`
	Algorithm     string `json:"algorithm"`
	Date          string `json:"date"`
	Policy        string `json:"policy"`
	Signature     string `json:"signature"`
	SecurityToken string `json:"securityToken"`
}

type aws4Signer struct {
	credentials *aws.Credentials
	dateOnly    string
	region      string
}
type uploadItemGenerator struct {
	credentials *aws.Credentials
	region      string
	bucket      string
}

type Response struct {
	Manifest  UploadItem   `json:"manifest"`
	Art       []UploadItem `json:"art"`
	PackId    string       `json:"packId"`
	UploadURL string       `json:"uploadURL"`
}

func serveS3Signature(
	statsdClient statsd.ClientInterface,
	logger *zerolog.Logger,
	muxConfig *Config,
) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		packId, err := generatePackId()
		if err != nil {
			logger.Error().Msgf("failed to generate pack id %v", err)
			ctx.Error(fasthttp.StatusMessage(fasthttp.StatusInternalServerError),
				fasthttp.StatusInternalServerError)
			return
		}

		//
		// Verify query params
		//

		artType := ctx.QueryArgs().Peek("artType")
		var packPrefix string
		var maxArtSize int
		if string(artType) == "sticker" {
			packPrefix = "stickers"
			maxArtSize = maxStickerSize
		} else {
			logger.Error().Msgf("unsupported artType %s", artType)
			ctx.Error(fasthttp.StatusMessage(fasthttp.StatusBadRequest),
				fasthttp.StatusBadRequest)
			return
		}

		artCount, err := ctx.QueryArgs().GetUint("artCount")
		if err != nil {
			logger.Error().Msgf("unsupported or missing artCount %+v", err)
			ctx.Error(fasthttp.StatusMessage(fasthttp.StatusBadRequest),
				fasthttp.StatusBadRequest)
			return
		}

		//
		// Verify auth
		//

		authHeader := ctx.Request.Header.Peek(fasthttp.HeaderAuthorization)
		if authHeader == nil {
			logger.Error().Msgf("missing auth header")
			ctx.Error(fasthttp.StatusMessage(fasthttp.StatusUnauthorized),
				fasthttp.StatusUnauthorized)
			return
		}

		userId, err := verifyAuthHeader(string(authHeader), muxConfig.authSecret)
		if err != nil {
			errorTag := fmt.Sprintf("error:%s", err.Error())
			tags := []string{errorTag}
			statsdClient.Incr("verifyAuthHeader.failure", tags, 1)
			ctx.Error(fasthttp.StatusMessage(fasthttp.StatusUnauthorized),
				fasthttp.StatusUnauthorized)
			return
		} else {
			statsdClient.Incr("verifyAuthHeader.success", []string{}, 1)
		}

		//
		// Apply rate limiting
		//

		isAllowed, err := muxConfig.rateLimiter.Validate(ctx, *userId)
		if err != nil {
			logger.Error().Err(err).Msgf("unable to apply rate limiting %+v", err)
			ctx.Error(fasthttp.StatusMessage(fasthttp.StatusInternalServerError),
				fasthttp.StatusInternalServerError)
			return
		}
		if !isAllowed {
			logger.Error().Err(err).Msgf("rate limited %+v", userId)
			ctx.Error(fasthttp.StatusMessage(fasthttp.StatusTooManyRequests),
				fasthttp.StatusTooManyRequests)
			return
		}

		//
		// Generate upload items
		//

		credentials, err := muxConfig.awsConfig.Credentials.Retrieve(ctx)
		if err != nil {
			logger.Error().Err(err).Msgf("unable to retrieve credentials %+v", err)
			ctx.Error(fasthttp.StatusMessage(fasthttp.StatusInternalServerError),
				fasthttp.StatusInternalServerError)
			return
		}

		uploadItemGenerator := uploadItemGenerator{
			credentials: &credentials,
			region:      muxConfig.region,
			bucket:      muxConfig.bucket,
		}

		packLocation := fmt.Sprintf("%s/%s", packPrefix, packId)
		manifestKey := fmt.Sprintf("%s/manifest.proto", packLocation)

		manifest, err := uploadItemGenerator.generate(
			manifestKey, -1, maxManifestSize)
		if err != nil {
			logger.Error().Err(err).Msg("unable to generate manifest")
			ctx.Error(fasthttp.StatusMessage(fasthttp.StatusInternalServerError),
				fasthttp.StatusInternalServerError)
			return
		}

		art := make([]UploadItem, artCount)
		for i := 0; i < artCount; i++ {
			artKey := fmt.Sprintf("%s/full/%d", packLocation, i)
			art[i], err = uploadItemGenerator.generate(artKey, i, maxArtSize)
			if err != nil {
				logger.Error().Err(err).Msgf("unable to generate art %d", i)
				ctx.Error(fasthttp.StatusMessage(fasthttp.StatusInternalServerError),
					fasthttp.StatusInternalServerError)
				return
			}
		}

		response, err := json.Marshal(Response{
			Manifest:  manifest,
			Art:       art,
			PackId:    packId,
			UploadURL: muxConfig.uploadURL,
		})
		if err != nil {
			logger.Error().Err(err).Msg("unable to response into json")
			ctx.Error(fasthttp.StatusMessage(fasthttp.StatusInternalServerError),
				fasthttp.StatusInternalServerError)
			return
		}

		ctx.SetContentType("application/json")
		_, err = ctx.Write(response)
		if err != nil {
			logger.Trace().Err(err).Msg(
				"an error occurred while writing the http response",
			)
		}
	}
}

func (s *aws4Signer) sign(stringToSign string) string {
	dateHmac := hmac.New(sha256.New, []byte("AWS4"+s.credentials.SecretAccessKey))
	dateHmac.Write([]byte(s.dateOnly))
	kDate := dateHmac.Sum(nil)

	regionHmac := hmac.New(sha256.New, kDate)
	regionHmac.Write([]byte(s.region))
	kRegion := regionHmac.Sum(nil)

	serviceHmac := hmac.New(sha256.New, kRegion)
	serviceHmac.Write([]byte("s3"))
	kService := serviceHmac.Sum(nil)

	signingHmac := hmac.New(sha256.New, kService)
	signingHmac.Write([]byte("aws4_request"))
	kSigning := signingHmac.Sum(nil)

	signatureHmac := hmac.New(sha256.New, kSigning)
	signatureHmac.Write([]byte(stringToSign))
	signature := signatureHmac.Sum(nil)

	return hex.EncodeToString(signature)
}

func (g *uploadItemGenerator) generate(
	key string,
	id int,
	maxSize int,
) (UploadItem, error) {
	accessKeyId := g.credentials.AccessKeyID
	securityToken := g.credentials.SessionToken
	now := time.Now().UTC()
	dateOnly := now.Format("20060102")
	date := now.Format("20060102T150405Z0700")
	credential := fmt.Sprintf(
		"%s/%s/%s/s3/aws4_request", accessKeyId, dateOnly, g.region)

	policyDocStruct := struct {
		Expiration string        `json:"expiration"`
		Conditions []interface{} `json:"conditions"`
	}{
		Expiration: time.Now().UTC().Add(time.Hour).Format(time.RFC3339),
		Conditions: []interface{}{
			[]interface{}{"content-length-range", 1, maxSize},
			[]interface{}{"starts-with", "$Content-Type", ""},
			map[string]string{"key": key},
			map[string]string{"acl": "private"},
			map[string]string{"bucket": g.bucket},
			map[string]string{"X-Amz-Algorithm": signatureAlgorithm},
			map[string]string{"X-Amz-Credential": credential},
			map[string]string{"X-Amz-Security-Token": securityToken},
			map[string]string{"X-Amz-Date": date},
		},
	}

	policyDoc, err := json.Marshal(policyDocStruct)
	if err != nil {
		return UploadItem{}, err
	}

	base64EncodedPolicyDoc := base64.StdEncoding.EncodeToString(policyDoc)
	signer := aws4Signer{
		credentials: g.credentials,
		dateOnly:    dateOnly,
		region:      g.region,
	}
	signature := signer.sign(base64EncodedPolicyDoc)

	return UploadItem{
		Id:            id,
		Key:           key,
		Credential:    credential,
		Acl:           "private",
		Algorithm:     signatureAlgorithm,
		Date:          date,
		Policy:        base64EncodedPolicyDoc,
		Signature:     signature,
		SecurityToken: securityToken,
	}, nil
}

func generatePackId() (string, error) {
	raw := make([]byte, 16)
	read, err := rand.Read(raw)
	if err != nil {
		return "", err
	}

	if read != 16 {
		return "", errors.New("Failed to generate random bytes")
	}

	return hex.EncodeToString(raw), nil
}
