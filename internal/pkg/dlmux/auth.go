//
// Copyright 2023 Signal Messenger, LLC
// SPDX-License-Identifier: AGPL-3.0-only
//

package dlmux

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"
	"time"
)

func verifyAuthHeader(authHeader string, sharedSecret []byte) (*string, error) {
	authTokens := strings.SplitN(authHeader, " ", 2)
	if strings.ToLower(authTokens[0]) != "basic" {
		return nil, errors.New("Unsupported authorization")
	}

	basicData, err := base64.StdEncoding.DecodeString(authTokens[1])
	if err != nil {
		return nil, err
	}

	tokens := strings.SplitN(string(basicData), ":", 3)
	if len(tokens) != 3 {
		return nil, errors.New("Invalid token count")
	}
	userId := tokens[0]
	timestamp, err := strconv.ParseInt(tokens[1], 10, 64)
	if err != nil {
		return nil, err
	}

	expectedMac, err := hex.DecodeString(tokens[2])
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	if time.Unix(timestamp, 0).Add(24 * time.Hour).Before(now) {
		return nil, errors.New("Expired token")
	}

	h := hmac.New(sha256.New, sharedSecret)
	h.Write([]byte(tokens[0] + ":" + tokens[1]))
	actualMac := h.Sum(nil)

	if !hmac.Equal(actualMac, expectedMac) {
		return nil, errors.New("Invalid signature")
	}

	return &userId, nil
}
