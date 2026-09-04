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
	"strings"
)

const redacted = "[redacted]"

var secretKeyNames = map[string]struct{}{
	"api_key":             {},
	"apikey":              {},
	"authorization":       {},
	"client_secret":       {},
	"clientsecret":        {},
	"oauth_client_secret": {},
	"password":            {},
	"refresh_token":       {},
	"signing_secret":      {},
	"token":               {},
	"x_api_key":           {},
	"x_org_api_key":       {},
	"x_user_api_key":      {},
}

func isSecretKey(key string) bool {
	k := strings.ToLower(strings.TrimSpace(key))
	k = strings.ReplaceAll(k, "-", "_")
	_, ok := secretKeyNames[k]
	if ok {
		return true
	}
	return strings.Contains(k, "secret") || strings.Contains(k, "password") || strings.HasSuffix(k, "_token")
}

func redactJSON(raw []byte) string {
	if len(raw) == 0 {
		return ""
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return string(raw)
	}
	redactValue(v)
	out, err := json.Marshal(v)
	if err != nil {
		return string(raw)
	}
	return string(out)
}

func redactValue(v any) {
	switch t := v.(type) {
	case map[string]any:
		for k, child := range t {
			if isSecretKey(k) {
				t[k] = redacted
				continue
			}
			redactValue(child)
		}
	case []any:
		for _, child := range t {
			redactValue(child)
		}
	}
}
