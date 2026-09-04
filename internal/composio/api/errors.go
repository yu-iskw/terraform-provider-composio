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
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const maxErrorBody = 4096

type APIError struct {
	StatusCode   int
	Method       string
	Path         string
	RequestID    string
	Code         string
	Message      string
	RetryAfter   time.Duration
	ResponseBody string
}

func (e *APIError) Error() string {
	if e == nil {
		return "composio api error"
	}
	parts := []string{fmt.Sprintf("composio API %s %s: HTTP %d", e.Method, e.Path, e.StatusCode)}
	if e.Code != "" {
		parts = append(parts, "code "+e.Code)
	}
	if e.Message != "" {
		parts = append(parts, e.Message)
	}
	if e.RequestID != "" {
		parts = append(parts, "request_id "+e.RequestID)
	}
	return strings.Join(parts, ": ")
}

func IsNotFound(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound
}

type wireError struct {
	Error struct {
		Message      string          `json:"message"`
		Code         json.RawMessage `json:"code"`
		Slug         string          `json:"slug"`
		RequestID    string          `json:"request_id"`
		SuggestedFix string          `json:"suggested_fix"`
	} `json:"error"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

func parseAPIError(method, path string, status int, hdr http.Header, body []byte) *APIError {
	apiErr := &APIError{
		StatusCode:   status,
		Method:       method,
		Path:         path,
		RequestID:    firstHeader(hdr, "x-request-id", "x-composio-request-id"),
		RetryAfter:   parseRetryAfter(hdr.Get("Retry-After")),
		ResponseBody: redactJSON(clip(body, maxErrorBody)),
	}

	var wire wireError
	if json.Unmarshal(body, &wire) == nil {
		if wire.Error.Message != "" {
			apiErr.Message = wire.Error.Message
		} else if wire.Message != "" {
			apiErr.Message = wire.Message
		}
		if code := rawCode(wire.Error.Code); code != "" {
			apiErr.Code = code
		} else if wire.Error.Slug != "" {
			apiErr.Code = wire.Error.Slug
		}
		if wire.Error.RequestID != "" {
			apiErr.RequestID = wire.Error.RequestID
		} else if wire.RequestID != "" {
			apiErr.RequestID = wire.RequestID
		}
		if wire.Error.SuggestedFix != "" && apiErr.Message != "" {
			apiErr.Message = apiErr.Message + ". " + wire.Error.SuggestedFix
		}
	}
	if apiErr.Message == "" {
		apiErr.Message = http.StatusText(status)
	}
	return apiErr
}

func rawCode(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	var n json.Number
	if json.Unmarshal(raw, &n) == nil {
		return n.String()
	}
	return strings.TrimSpace(string(raw))
}

func firstHeader(hdr http.Header, keys ...string) string {
	for _, key := range keys {
		if v := hdr.Get(key); v != "" {
			return v
		}
	}
	return ""
}

func parseRetryAfter(v string) time.Duration {
	v = strings.TrimSpace(v)
	if v == "" {
		return 0
	}
	if secs, err := strconv.Atoi(v); err == nil && secs >= 0 {
		return time.Duration(secs) * time.Second
	}
	if at, err := http.ParseTime(v); err == nil {
		d := time.Until(at)
		if d < 0 {
			return 0
		}
		return d
	}
	return 0
}

func clip(b []byte, n int) []byte {
	if len(b) <= n {
		return b
	}
	return b[:n]
}
