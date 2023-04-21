//
// Copyright 2023 Signal Messenger, LLC
// SPDX-License-Identifier: AGPL-3.0-only
//

package main

import (
	"context"
	"io"
	"net"
	"os"
	"regexp"
	"time"

	"github.com/DataDog/datadog-go/v5/statsd"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/jessevdk/go-flags"
	"github.com/rs/zerolog"
	"github.com/valyala/fasthttp"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/signalapp/art-service/internal/pkg/awslog"
	appConfig "github.com/signalapp/art-service/internal/pkg/config"
	"github.com/signalapp/art-service/internal/pkg/dlmux"
	"github.com/signalapp/art-service/internal/pkg/fasthttplog"
	"github.com/signalapp/art-service/internal/pkg/metrics"
)

func main() {
	var opt struct {
		Verbose bool `short:"v" description:"Enable debug logging"`

		Profile string `long:"profile" description:"AWS Profile"`

		ListenAddr string `short:"l" long:"listen" default:"[::1]:3000" description:"net.Dial compatible address string"`

		JsonLogPath flags.Filename `long:"jsonlog" description:"Path to JSON log output"`

		DatadogStatsdAddr string `long:"datadog-statsd-addr" description:"Address to reach a datadog compatible statsd over UDP"`

		DatadogTags []string `long:"datadog-tags" description:"Tags to assign to every metric"`

		Args struct {
			ConfigFile flags.Filename `positional-arg-name:"config_file.yaml"`
		} `positional-args:"yes" required:"yes"`
	}

	parser := flags.NewParser(&opt, flags.Default)

	args, err := parser.Parse()
	if flags.WroteHelp(err) {
		os.Exit(1)
	}
	if err != nil {
		logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
		logger.Fatal().Err(err).Msg("failed to parse args")
	}

	consoleWriter := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}
	if err != nil {
		logger := zerolog.New(consoleWriter).With().Timestamp().Logger()
		logger.Fatal().Err(err).Msg("failed to parse args")
	}
	var logger zerolog.Logger
	var logWriter io.Writer
	if len(opt.JsonLogPath) > 0 {
		logWriter = zerolog.MultiLevelWriter(
			consoleWriter,
			&lumberjack.Logger{
				Filename:   string(opt.JsonLogPath),
				MaxBackups: 10,
				Compress:   true,
			},
		)
		logger = zerolog.New(logWriter).With().Timestamp().Logger().Level(zerolog.InfoLevel)
	} else {
		logWriter = consoleWriter
		logger = zerolog.New(logWriter).With().Caller().Timestamp().Logger()
	}

	if !opt.Verbose {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	if len(args) > 0 {
		logger.Info().Msg("too many args")
		parser.WriteHelp(os.Stderr)
		os.Exit(1)
	}

	configFile, err := os.Open(string(opt.Args.ConfigFile))
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to open config file")
	}

	var config appConfig.Config
	err = config.ParseFromFile(configFile)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to parse config file")
	}

	logger.Info().Msg("initializing AWS SDK and loading credentials")
	aws, err := awsConfig.LoadDefaultConfig(context.Background(),
		awsConfig.WithSharedConfigProfile(opt.Profile),
		awsConfig.WithLogger(awslog.New(&logger)))
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to load AWS config")
	}

	logger.Info().Msg("initializing datadog statsd")
	var requiredMetricsPathsPattern *regexp.Regexp = regexp.MustCompile(
		"^(/api/*)")
	var ignoredMetricsPathsPattern *regexp.Regexp = regexp.MustCompile(
		"^(/api/healthz)")
	var statsdClient statsd.ClientInterface
	if len(opt.DatadogStatsdAddr) > 0 {
		logger.Info().Msgf(
			"initializing datadog statsd client to %s", opt.DatadogStatsdAddr)
		statsdClient, err = statsd.New(
			opt.DatadogStatsdAddr,
			statsd.WithNamespace("artd"),
			statsd.WithTags(opt.DatadogTags))
		if err != nil {
			logger.Fatal().Err(err).Msgf(
				"failed to create datadog statsd client for address %s",
				opt.DatadogStatsdAddr)
		}
	} else {
		statsdClient = &statsd.NoOpClient{}
	}

	muxConfig, err := dlmux.NewConfig(&config, aws)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to create mux config")
	}

	logger.Info().Msg("configuring art-server http routes")
	r := dlmux.NewRouter(statsdClient, &logger, muxConfig)

	logger.Info().Msg("setting up server")
	server := &fasthttp.Server{
		Handler: metrics.NewHandler(
			statsdClient,
			requiredMetricsPathsPattern,
			ignoredMetricsPathsPattern,
			r.Handler),
		Logger:      fasthttplog.New(&logger),
		ReadTimeout: time.Minute,
	}

	logger.Info().Msgf("listening on %v", opt.ListenAddr)
	ln, err := net.Listen("tcp", opt.ListenAddr)
	if err != nil {
		logger.Fatal().Err(err).Msgf("failed to listen on %v", opt.ListenAddr)
	}
	err = server.Serve(ln)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to listen and serve")
	}
}
