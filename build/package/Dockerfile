#
# Copyright 2023 Signal Messenger, LLC
# SPDX-License-Identifier: AGPL-3.0-only
#

FROM node:19 AS build_web
WORKDIR /
COPY ./protos ./protos
COPY ./web ./web
WORKDIR /web
RUN npm ci
RUN npm run build

FROM golang:1.19 AS build
WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY cmd ./cmd/
COPY internal ./internal/
COPY --from=build_web /web/dist ./web/dist
COPY --from=build_web /web/web.go ./web/web.go
RUN GOARCH=amd64 go build -o /artd ./cmd/artd

FROM gcr.io/distroless/base:latest-amd64
WORKDIR /
COPY --from=build /artd /artd
ARG CONFIG_FILE=configs/you_need_to_set_this.yaml
COPY $CONFIG_FILE /config.yaml
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/artd", "-l", "[::]:8080", "/config.yaml"]
