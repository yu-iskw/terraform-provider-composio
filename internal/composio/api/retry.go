// Copyright 2026 yu-iskw
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"math"
	"net"
	"net/http"
	"time"
)

const (
	maxAttempts    = 4
	retryBaseDelay = 200 * time.Millisecond
	retryMaxDelay  = 8 * time.Second
)

func isRetryableMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func isRetryableStatus(status int) bool {
	switch status {
	case http.StatusRequestTimeout, http.StatusTooManyRequests,
		http.StatusInternalServerError, http.StatusBadGateway,
		http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

func isRetryableTransport(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr)
}

func backoff(attempt int, retryAfter time.Duration) time.Duration {
	if retryAfter > retryMaxDelay {
		return retryMaxDelay
	}
	if retryAfter > 0 {
		return retryAfter
	}
	if attempt < 1 {
		attempt = 1
	}
	exp := float64(retryBaseDelay) * math.Pow(2, float64(attempt-1))
	if exp > float64(retryMaxDelay) {
		exp = float64(retryMaxDelay)
	}
	return time.Duration(float64(exp) * jitter())
}

func jitter() float64 {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return 0.5
	}
	n := binary.LittleEndian.Uint64(b[:])
	return float64(n) / float64(^uint64(0))
}
