//
// Copyright 2023 Signal Messenger, LLC
// SPDX-License-Identifier: AGPL-3.0-only
//

package dlmux

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	pkgConfig "github.com/signalapp/art-service/internal/pkg/config"
	"github.com/valyala/fasthttp"
)

type rateLimiter struct {
	bucketName        string
	bucketSize        int
	leakRatePerMinute float64
	rdb               abstractRedisKV
}

type leakyBucket struct {
	BucketSize           int     `json:"bucketSize"`
	LeakRatePerMillis    float64 `json:"leakRatePerMillis"`
	SpaceRemaining       int     `json:"spaceRemaining"`
	LastUpdateTimeMillis int64   `json:"lastUpdateTimeMillis"`
}

func (b *leakyBucket) add(amount int) bool {
	elapsedTime := time.Since(time.UnixMilli(b.LastUpdateTimeMillis))

	newSpaceRemaining := b.SpaceRemaining +
		int(float64(elapsedTime.Milliseconds())*b.LeakRatePerMillis)
	if newSpaceRemaining < b.BucketSize {
		b.SpaceRemaining = newSpaceRemaining
	} else {
		b.SpaceRemaining = b.BucketSize
	}
	b.LastUpdateTimeMillis = time.Now().UnixMilli()

	if b.SpaceRemaining >= amount {
		b.SpaceRemaining -= amount
		return true
	} else {
		return false
	}
}

func newRateLimiter(app *pkgConfig.Config, rdb abstractRedis) *rateLimiter {
	return &rateLimiter{
		bucketName:        app.RateLimiter.BucketName,
		bucketSize:        app.RateLimiter.BucketSize,
		leakRatePerMinute: app.RateLimiter.LeakRatePerMinute,
		rdb:               rdb,
	}
}

func (r *rateLimiter) Validate(
	ctx *fasthttp.RequestCtx,
	userId string,
) (bool, error) {
	key := fmt.Sprintf("leaky_bucket::%s::%s", r.bucketName, userId)
	leakRatePerMillis := r.leakRatePerMinute / float64(time.Minute.Milliseconds())

	rawValue, err := r.rdb.Get(ctx, key).Result()
	var bucket leakyBucket
	if err == redis.Nil {
		bucket = leakyBucket{
			BucketSize:           r.bucketSize,
			LeakRatePerMillis:    leakRatePerMillis,
			SpaceRemaining:       r.bucketSize,
			LastUpdateTimeMillis: time.Now().UnixMilli(),
		}
	} else if err == nil {
		err = json.Unmarshal([]byte(rawValue), &bucket)
		if err != nil {
			return false, err
		}
	} else {
		return false, err
	}

	// Rate-limited!
	if !bucket.add(1) {
		return false, nil
	}

	newValueBytes, err := json.Marshal(bucket)
	if err != nil {
		return false, err
	}

	expires := int(float64(r.bucketSize) / leakRatePerMillis)
	r.rdb.Set(
		ctx,
		key,
		string(newValueBytes),
		time.Duration(expires)*time.Millisecond,
	)

	return true, nil
}
