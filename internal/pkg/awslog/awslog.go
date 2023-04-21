//
// Copyright 2023 Signal Messenger, LLC
// SPDX-License-Identifier: AGPL-3.0-only
//

package awslog

import (
	"github.com/aws/smithy-go/logging"
	"github.com/rs/zerolog"
)

func New(logger *zerolog.Logger) logging.LoggerFunc {
	return func(classification logging.Classification, format string, v ...interface{}) {
		var event *zerolog.Event
		switch classification {
		case logging.Warn:
			event = logger.Warn()
		case logging.Debug:
			event = logger.Debug()
		default:
			event = logger.Info()
		}
		event.Msgf(format, v...)
	}
}
