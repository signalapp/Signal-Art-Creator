//
// Copyright 2023 Signal Messenger, LLC
// SPDX-License-Identifier: AGPL-3.0-only
//

package fasthttplog

import (
	"github.com/rs/zerolog"
	"github.com/valyala/fasthttp"
)

type FastHttpLogger struct {
	logger *zerolog.Logger
}

var _ fasthttp.Logger = (*FastHttpLogger)(nil)

func New(logger *zerolog.Logger) *FastHttpLogger {
	return &FastHttpLogger{logger: logger}
}

func (l *FastHttpLogger) Printf(format string, args ...interface{}) {
	l.logger.Info().Msgf(format, args...)
}
