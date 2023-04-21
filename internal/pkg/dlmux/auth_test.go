//
// Copyright 2023 Signal Messenger, LLC
// SPDX-License-Identifier: AGPL-3.0-only
//

package dlmux

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"testing"
	"time"
)

const userId = "test"

func getAuthHeader(secret []byte, userId string, now time.Time) string {
	h := hmac.New(sha256.New, secret)
	timestamp := now.UnixMilli() / time.Second.Milliseconds()

	username := fmt.Sprintf("%s:%d", userId, timestamp)

	h.Write([]byte(username))
	password := h.Sum(nil)

	auth := fmt.Sprintf("%s:%s", username, hex.EncodeToString(password))
	return fmt.Sprintf(
		"Basic %s",
		base64.StdEncoding.EncodeToString([]byte(auth)))
}

func TestVerifyValidAuthHeader(t *testing.T) {
	sharedSecret := make([]byte, 32)
	authHeader := getAuthHeader(sharedSecret, userId, time.Now().UTC())

	outUserId, err := verifyAuthHeader(authHeader, sharedSecret)
	if err != nil {
		t.Fatalf("Didn't expect error %+v", err)
	}
	if outUserId == nil {
		t.Fatal("Expected userId")
	}

	if *outUserId != userId {
		t.Fatalf("Expected correct userId, but got %s", *outUserId)
	}
}

func TestVerifyExpiredAuthHeader(t *testing.T) {
	sharedSecret := make([]byte, 32)
	past := time.Now().UTC().Add(-72 * time.Hour)

	authHeader := getAuthHeader(sharedSecret, userId, past)

	outUserId, err := verifyAuthHeader(authHeader, sharedSecret)
	if err == nil {
		t.Fatal("Expected error")
	}
	if outUserId != nil {
		t.Fatalf("Didn't expect userId %s", *outUserId)
	}
}

func TestVerifyInvaliduthHeader(t *testing.T) {
	sharedSecret := make([]byte, 32)
	otherSharedSecret := make([]byte, 32)
	otherSharedSecret[0] = 1

	authHeader := getAuthHeader(sharedSecret, userId, time.Now().UTC())

	outUserId, err := verifyAuthHeader(authHeader, otherSharedSecret)
	if err == nil {
		t.Fatal("Expected error")
	}
	if outUserId != nil {
		t.Fatalf("Didn't expect userId %s", *outUserId)
	}
}
