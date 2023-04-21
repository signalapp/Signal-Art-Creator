//
// Copyright 2023 Signal Messenger, LLC
// SPDX-License-Identifier: AGPL-3.0-only
//

package dlmux

import (
	"mime"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
	"github.com/signalapp/art-service/web"
	"github.com/valyala/fasthttp"
)

type candidateFile struct {
	path     string
	encoding *string
}

var gzipEncoding = "gzip"
var brotliEncoding = "br"

func serveStatics(
	logger *zerolog.Logger,
) fasthttp.RequestHandler {
	statics := web.Statics

	return func(ctx *fasthttp.RequestCtx) {
		path := string(ctx.Path())
		if strings.HasPrefix(path, "/assets") {
			ctx.Response.Header.Add(
				fasthttp.HeaderCacheControl,
				"public, max-age=604800, immutable")
		}

		basePath := path
		if path == "/" {
			basePath = "/index.html"
		}

		var candidateFiles = make([]candidateFile, 1, 3)
		candidateFiles[0] = candidateFile{basePath, nil}

		if ctx.Request.Header.HasAcceptEncoding(gzipEncoding) {
			candidateFiles = append(candidateFiles,
				candidateFile{basePath + ".gz", &gzipEncoding})
		}
		if ctx.Request.Header.HasAcceptEncoding(brotliEncoding) {
			candidateFiles = append(candidateFiles,
				candidateFile{basePath + ".br", &brotliEncoding})
		}

		for i := len(candidateFiles) - 1; i >= 0; i-- {
			candidate := candidateFiles[i]

			file, err := statics.Open("dist/" + candidate.path[1:])
			if err != nil {
				continue
			}

			stat, err := file.Stat()
			if err != nil {
				continue
			}

			contentType := mime.TypeByExtension(filepath.Ext(basePath))
			if contentType == "" {
				contentType = "application/octet-stream"
			}

			ctx.SetContentType(contentType)
			if candidate.encoding != nil {
				ctx.Response.Header.SetContentEncoding(*candidate.encoding)
			}
			ctx.SetBodyStream(file, int(stat.Size()))
			logger.Debug().Msgf("serving %s => %s", path, candidate.path)
			return
		}

		logger.Info().Msgf("missing static file for %s", path)
		ctx.NotFound()
	}
}
