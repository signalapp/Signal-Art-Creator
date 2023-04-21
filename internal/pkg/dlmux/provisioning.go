//
// Copyright 2023 Signal Messenger, LLC
// SPDX-License-Identifier: AGPL-3.0-only
//

package dlmux

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/rs/zerolog"
	"github.com/signalapp/art-service/internal/pkg/messages"
	"github.com/valyala/fasthttp"
	"google.golang.org/protobuf/proto"
)

const readLimit int64 = 1024
const tokenLen int = 32
const handshakeTimeout time.Duration = 10 * time.Second
const pingInterval time.Duration = 15 * time.Second
const pongWait time.Duration = 30 * time.Second
const writeWait time.Duration = 10 * time.Second

var goodbyeError = websocket.CloseError{
	Code: websocket.CloseNormalClosure,
	Text: "goodbye",
}

type message struct {
	closeErr *websocket.CloseError
	data     []byte
}

type client struct {
	logger *zerolog.Logger
	rdb    abstractRedisPubSub
	conn   *websocket.Conn

	logId       string
	token       string
	channelName string

	recv <-chan message
	send chan<- message
	done <-chan bool
}

func pumpSocketReads(
	logger *zerolog.Logger,
	logId string,
	conn *websocket.Conn,
	recv chan<- message,
) {
	defer func() {
		logger.Debug().Msgf("%s: closing read pump", logId)
		close(recv)
	}()

	conn.SetReadLimit(readLimit)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		logger.Debug().Msgf("%s: got pong", logId)
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			if closeErr, ok := err.(*websocket.CloseError); ok {
				if websocket.IsUnexpectedCloseError(
					closeErr,
					websocket.CloseNormalClosure,
					websocket.CloseGoingAway,
					websocket.CloseNoStatusReceived,
				) {
					logger.Warn().Msgf("%s: remote close %+v", logId, closeErr)
				} else {
					logger.Debug().Msgf("%s: remote close %+v", logId, closeErr)
				}

				// Don't let readers acknowledge 1006 (socket is dead)
				if websocket.IsUnexpectedCloseError(
					closeErr,
					websocket.CloseAbnormalClosure,
				) {
					recv <- message{closeErr: closeErr}
				}
			} else {
				logger.Warn().Msgf("%s: remote read err %+v", logId, err)
			}
			return
		}

		if messageType != websocket.BinaryMessage {
			logger.Warn().Msgf("%s: remote sent non-binary message", logId)
			continue
		}

		logger.Debug().Msgf("%s: remote sent message len=%d", logId, len(p))
		recv <- message{data: p}
	}
}

func pumpSocketWrites(
	logger *zerolog.Logger,
	logId string,
	conn *websocket.Conn,
	send <-chan message,
	done chan<- bool,
) {
	ticker := time.NewTicker(pingInterval)

	defer func() {
		logger.Debug().Msgf("%s: closing write pump", logId)
		conn.Close()
		ticker.Stop()
		done <- true
	}()

	for {
		var msgType int
		var msgData []byte
		select {
		case <-ticker.C:
			logger.Debug().Msgf("%s: sending ping", logId)
			msgType = websocket.PingMessage
		case msg, ok := <-send:
			// Closed
			if !ok {
				logger.Debug().Msgf("%s: send channel closed", logId)
				return
			}

			if msg.closeErr != nil {
				if websocket.IsUnexpectedCloseError(
					msg.closeErr,
					websocket.CloseNormalClosure,
					websocket.CloseGoingAway,
				) {
					logger.Warn().Msgf("%s: sending close %+v", logId, msg.closeErr)
				} else {
					logger.Debug().Msgf("%s: sending close %+v", logId, msg.closeErr)
				}
				msgType = websocket.CloseMessage
				msgData = websocket.FormatCloseMessage(
					msg.closeErr.Code, msg.closeErr.Text)
			} else {
				logger.Debug().Msgf("%s: sending data len=%d", logId, len(msg.data))
				msgType = websocket.BinaryMessage
				msgData = msg.data
			}
		}

		conn.SetWriteDeadline(time.Now().Add(writeWait))

		// Let read pump handle errors
		_ = conn.WriteMessage(msgType, msgData)
	}
}

func newClient(
	logger *zerolog.Logger,
	rdb abstractRedisPubSub,
	conn *websocket.Conn,
	logId string,
	token string,
	channelName string,
) client {
	recv := make(chan message)
	send := make(chan message)
	done := make(chan bool, 1)

	go pumpSocketReads(logger, logId, conn, recv)
	go pumpSocketWrites(logger, logId, conn, send, done)

	return client{
		logger: logger,
		conn:   conn,
		rdb:    rdb,

		logId:       logId,
		token:       token,
		channelName: channelName,

		recv: recv,
		send: send,
		done: done,
	}
}

func (c *client) cleanup() {
	close(c.send)

	// wait for channel to be closed
	for range c.recv {
	}

	// wait for write pump to end, otherwise the hijacked
	// fasthttp handler for websocket is going to return and close the underlying
	// resource resulting in a panic.
	<-c.done

	c.logger.Debug().Msgf("%s: pumps are clean", c.logId)
}

func (c *client) serveReader(ctx context.Context) {
	rdb := c.rdb
	pubsub := rdb.Subscribe(ctx, c.channelName)
	defer pubsub.Close()

	// Process incoming redis messages
	tokenMessage := &messages.ProvisioningToken{
		Token: &c.token,
	}
	handshake, err := proto.Marshal(tokenMessage)
	if err != nil {
		c.logger.Error().Msgf("%s: failed to serialize handshake", c.logId)
		c.send <- message{
			closeErr: &websocket.CloseError{
				Code: websocket.CloseInternalServerErr,
				Text: "failed to serialize handshake",
			},
		}
		return
	}

	c.send <- message{data: handshake}

	for {
		select {
		case <-ctx.Done():
			c.logger.Warn().Msgf("%s: context is done", c.logId)
			return
		case msg, ok := <-c.recv:
			// Channel closed
			if !ok {
				return
			}

			// Acknowledge close
			if msg.closeErr != nil {
				c.send <- message{
					closeErr: &goodbyeError,
				}
				return
			}

			c.send <- message{
				closeErr: &websocket.CloseError{
					Code: websocket.CloseProtocolError,
					Text: "readonly socket",
				},
			}

			return
		case msg := <-pubsub.Channel():
			if msg.Payload == "" {
				// Channel closed on the other side of pubsub
				c.send <- message{
					closeErr: &websocket.CloseError{
						Code: websocket.CloseNormalClosure,
						Text: "normal",
					},
				}
				return
			}

			data, err := hex.DecodeString(msg.Payload)
			if err != nil {
				c.logger.Error().Msgf("%s: failed to decode redis message %+v",
					c.logId, err)
				c.send <- message{
					closeErr: &websocket.CloseError{
						Code: websocket.CloseInternalServerErr,
						Text: "failed to decode redis message",
					},
				}
				return
			}

			c.send <- message{data: data}
		}
	}
}

func (c *client) serveWriter(ctx context.Context) {
	rdb := c.rdb

	defer func() {
		err := rdb.Publish(ctx, c.channelName, "").Err()
		if err != nil {
			c.logger.Error().Msgf("%s: failed to publish close message", c.logId)
		}
	}()

	// Process incoming websocket messages
	for {
		select {
		case <-ctx.Done():
			c.logger.Warn().Msgf("%s: context is done", c.logId)
			return
		case msg, ok := <-c.recv:
			// Closed
			if !ok {
				return
			}

			// Acknowledge close
			if msg.closeErr != nil {
				c.send <- message{
					closeErr: &goodbyeError,
				}
				return
			}

			if len(msg.data) == 0 {
				c.logger.Warn().Msgf(
					"%s: ignoring empty websocket message", c.logId)
				continue
			}

			numSubscribers, err := rdb.Publish(
				ctx,
				c.channelName,
				hex.EncodeToString(msg.data)).Result()
			if err != nil {
				c.logger.Error().Msgf("%s: failed to publish message to redis %+v",
					c.logId, err)
				c.send <- message{
					closeErr: &websocket.CloseError{
						Code: websocket.CloseInternalServerErr,
						Text: "redis publish error",
					},
				}
				return
			}

			if numSubscribers == 0 {
				c.logger.Error().Msgf("%s: no redis subscribers", c.logId)
				c.send <- message{
					closeErr: &websocket.CloseError{
						Code: websocket.CloseProtocolError,
						Text: "no subscribers",
					},
				}
				return
			}

			// We only allow delivery of a single message.
			c.send <- message{
				closeErr: &goodbyeError,
			}
			return
		}
	}
}

func serveProvisioningSocket(
	logger *zerolog.Logger,
	muxConfig *Config,
) fasthttp.RequestHandler {
	var upgrader = websocket.FastHTTPUpgrader{
		HandshakeTimeout: handshakeTimeout,
		CheckOrigin: func(ctx *fasthttp.RequestCtx) bool {
			origin := ctx.Request.Header.Peek(fasthttp.HeaderOrigin)
			if origin == nil {
				logger.Error().Msgf("missing origin header")
				return false
			}

			if muxConfig.provisioning.Origin != string(origin) {
				logger.Error().Msgf("invalid origin header %s", string(origin))
				return false
			}

			return true
		},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	return func(ctx *fasthttp.RequestCtx) {
		tokenBytes := ctx.QueryArgs().Peek("token")

		var logId string
		var isReader bool
		var token string
		if tokenBytes == nil {
			// No token, provided - generate new token
			generatedToken, err := generateProvisioningToken()
			if err != nil {
				logger.Error().Msgf("failed to generate token %+v", err)
				ctx.Error(fasthttp.StatusMessage(fasthttp.StatusInternalServerError),
					fasthttp.StatusInternalServerError)
				return
			}
			token = generatedToken
			logId = fmt.Sprintf("reader(%s)", token[0:8])
			isReader = true
		} else {
			// Parse and validate incoming token
			if len(tokenBytes) != 2*tokenLen {
				logger.Warn().Msgf("client sent bad token")
				ctx.Error(fasthttp.StatusMessage(fasthttp.StatusBadRequest),
					fasthttp.StatusBadRequest)
			}

			_, err := hex.DecodeString(string(tokenBytes))
			if err != nil {
				logger.Warn().Msgf("client sent bad token %+v", err)
				ctx.Error(fasthttp.StatusMessage(fasthttp.StatusBadRequest),
					fasthttp.StatusBadRequest)
			}
			token = string(tokenBytes)
			logId = fmt.Sprintf("writer(%s)", token[0:8])
			isReader = false
		}

		err := upgrader.Upgrade(ctx, func(conn *websocket.Conn) {
			conn.SetReadLimit(readLimit)
			channelName := fmt.Sprintf(
				"%s%s",
				muxConfig.provisioning.PubSubPrefix,
				token)

			c := newClient(
				logger,
				muxConfig.rdb,
				conn,
				logId,
				token,
				channelName,
			)

			defer c.cleanup()

			if isReader {
				c.serveReader(ctx)
			} else {
				c.serveWriter(ctx)
			}
		})

		if err != nil {
			logger.Error().Msgf("failed to upgrade to websocket %+v", err)
			return
		}
	}
}

func generateProvisioningToken() (string, error) {
	raw := make([]byte, tokenLen)
	read, err := rand.Read(raw)
	if err != nil {
		return "", err
	}

	if read != tokenLen {
		return "", errors.New("Failed to generate random bytes")
	}

	return hex.EncodeToString(raw), nil
}
