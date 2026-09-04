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
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestParseRetryAfterSeconds(t *testing.T) {
	d := parseRetryAfter("3")
	if d != 3*time.Second {
		t.Fatalf("got %s", d)
	}
}

func TestIsRetryableMethod(t *testing.T) {
	if !isRetryableMethod(http.MethodGet) {
		t.Fatal("GET should retry")
	}
	if isRetryableMethod(http.MethodPost) {
		t.Fatal("POST should not retry")
	}
}

func TestRedactJSONByKey(t *testing.T) {
	in := []byte(`{"client_secret":"shh","ok":"public","nested":{"password":"p"}}`)
	out := redactJSON(in)
	if strings.Contains(out, "shh") || strings.Contains(out, `"p"`) {
		t.Fatalf("leaked: %s", out)
	}
	if !strings.Contains(out, "public") {
		t.Fatalf("lost public field: %s", out)
	}
}

func TestBackoffHonorsRetryAfter(t *testing.T) {
	d := backoff(1, 2*time.Second)
	if d != 2*time.Second {
		t.Fatalf("got %s", d)
	}
}

func TestBackoffClampsRetryAfter(t *testing.T) {
	d := backoff(1, 30*time.Second)
	if d != retryMaxDelay {
		t.Fatalf("got %s", d)
	}
}
