//
// Copyright 2023 Signal Messenger, LLC
// SPDX-License-Identifier: AGPL-3.0-only
//

package metrics

import (
	"fmt"
	"regexp"
	"time"

	"github.com/DataDog/datadog-go/v5/statsd"
	"github.com/valyala/fasthttp"
)

func NewHandler(
	statsdClient statsd.ClientInterface,
	requiredPathsPattern *regexp.Regexp,
	ignoredPathsPattern *regexp.Regexp,
	innerHandler fasthttp.RequestHandler,
) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		innerHandler(ctx)

		if !requiredPathsPattern.Match(ctx.Path()) {
			return
		}
		if ignoredPathsPattern.Match(ctx.Path()) {
			return
		}

		duration := time.Now().Sub(ctx.Time())
		statusCodeTag := fmt.Sprintf("status_code:%d", ctx.Response.StatusCode())
		pathTag := fmt.Sprintf("path:%s", ctx.Path())
		tags := []string{statusCodeTag, pathTag}

		// Don't report variadic path names
		if ctx.Response.StatusCode() == 404 {
			tags = tags[:1]
		}
		_ = statsdClient.Timing("http.response", duration, tags, 1)
		_ = statsdClient.Incr("http.response", tags, 1)
	}
}
