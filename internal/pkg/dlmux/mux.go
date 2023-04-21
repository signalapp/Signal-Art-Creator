//
// Copyright 2023 Signal Messenger, LLC
// SPDX-License-Identifier: AGPL-3.0-only
//

package dlmux

import (
	"context"
	"encoding/hex"
	"github.com/DataDog/datadog-go/v5/statsd"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/fasthttp/router"
	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog"
	pkgConfig "github.com/signalapp/art-service/internal/pkg/config"
)

type abstractRedisKV interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(
		ctx context.Context,
		key string,
		value interface{},
		expiration time.Duration) *redis.StatusCmd
}

type abstractRedisPubSub interface {
	Subscribe(ctx context.Context, channels ...string) *redis.PubSub
	Publish(
		ctx context.Context,
		channel string,
		message interface{}) *redis.IntCmd
}

type abstractRedis interface {
	abstractRedisKV
	abstractRedisPubSub
}

type Config struct {
	region       string
	bucket       string
	uploadURL    string
	rdb          abstractRedis
	provisioning pkgConfig.ProvisioningConfig
	rateLimiter  *rateLimiter
	authSecret   []byte
	awsConfig    aws.Config
}

func NewConfig(
	app *pkgConfig.Config,
	aws aws.Config,
) (*Config, error) {
	authSecret, err := hex.DecodeString(app.AuthSecret)
	if err != nil {
		return nil, err
	}

	var rdb abstractRedis
	if app.RedisURI != nil {
		rdb = redis.NewClient(&redis.Options{
			Addr: *app.RedisURI,
		})
	} else {
		rdb = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs: []string{*app.RedisClusterURI},
		})
	}

	return &Config{
		region:       app.Region,
		bucket:       app.Bucket,
		rdb:          rdb,
		provisioning: app.Provisioning,
		uploadURL:    app.UploadURL,
		authSecret:   authSecret,
		rateLimiter:  newRateLimiter(app, rdb),
		awsConfig:    aws,
	}, nil
}

func NewRouter(statsdClient statsd.ClientInterface, logger *zerolog.Logger, config *Config) *router.Router {
	r := router.New()
	r.GET("/api/healthz", serveHealthz(logger))
	r.GET("/api/form", serveS3Signature(statsdClient, logger, config))
	r.GET("/api/socket", serveProvisioningSocket(logger, config))
	r.GET("/{filepath:*}", serveStatics(logger))

	return r
}
