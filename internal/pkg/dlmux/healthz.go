//
// Copyright 2023 Signal Messenger, LLC
// SPDX-License-Identifier: AGPL-3.0-only
//

package dlmux

import (
	"github.com/rs/zerolog"
	"github.com/valyala/fasthttp"
)

func serveHealthz(logger *zerolog.Logger) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		_, err := ctx.WriteString("OK")
		if err != nil {
			logger.Trace().Err(err).Msg("failed to write response to health check")
		}
	}
}
