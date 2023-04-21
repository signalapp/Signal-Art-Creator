//
// Copyright 2023 Signal Messenger, LLC
// SPDX-License-Identifier: AGPL-3.0-only
//

package config

import (
	"io"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v2"
)

type ProvisioningConfig struct {
	PubSubPrefix string `yaml:"pubSubPrefix" validate:"required"`
	Origin       string `yaml:"origin" validate:"required"`
}

type Config struct {
	Region          string  `yaml:"region" validate:"required"`
	Bucket          string  `yaml:"bucket" validate:"required"`
	AuthSecret      string  `yaml:"authSecret" validate:"hexadecimal"`
	UploadURL       string  `yaml:"uploadURL" validate:"required"`
	RedisURI        *string `yaml:"redisURI" validate:"required_without=RedisClusterURI"`
	RedisClusterURI *string `yaml:"redisClusterURI" validate:"required_without=RedisURI"`
	RateLimiter     struct {
		BucketName        string  `yaml:"bucketName" validate:"required"`
		BucketSize        int     `yaml:"bucketSize" validate:"required"`
		LeakRatePerMinute float64 `yaml:"leakRatePerMinute" validate:"required"`
	} `yaml:"rateLimiter" validate:"required"`
	Provisioning ProvisioningConfig `yaml:"provisioning" validate:"required"`
}

func NewConfig() *Config {
	return &Config{}
}

func (config *Config) ParseFromFile(r io.Reader) error {
	decoder := yaml.NewDecoder(r)
	decoder.SetStrict(true)
	err := decoder.Decode(config)
	if err != nil {
		return err
	}
	return validator.New().Struct(config)
}
