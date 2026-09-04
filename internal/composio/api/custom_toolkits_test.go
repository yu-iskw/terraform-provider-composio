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
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCustomGenericPlaceholder(t *testing.T) {
	// Composio Custom MCP requires this exact header substitution token.
	const want = "{{generic_api_key}}"
	if CustomGenericPlaceholder != want {
		t.Fatalf("got %q, want %q", CustomGenericPlaceholder, want)
	}
}

func TestUpsertCustomToolkitSendsBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s", r.Method)
		}
		if r.URL.Path != "/api/v3.1/custom/toolkits/upsert" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if r.Header.Get("x-api-key") != "project-secret" {
			t.Errorf("x-api-key = %q", r.Header.Get("x-api-key"))
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		var got map[string]any
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("unmarshal: %v body=%s", err, body)
		}
		if got["slug"] != "ACME" {
			t.Errorf("slug = %v", got["slug"])
		}
		cfg, ok := got["toolkit_config"].(map[string]any)
		if !ok {
			t.Fatalf("toolkit_config type %T", got["toolkit_config"])
		}
		if cfg["name"] != "Acme" || cfg["app_url"] != "https://mcp.example.com/mcp" {
			t.Errorf("config = %#v", cfg)
		}
		schemes, ok := cfg["auth_schemes"].([]any)
		if !ok || len(schemes) != 1 {
			t.Fatalf("auth_schemes = %#v", cfg["auth_schemes"])
		}
		scheme, ok := schemes[0].(map[string]any)
		if !ok || scheme["mode"] != "API_KEY" {
			t.Fatalf("scheme = %#v", schemes[0])
		}
		headers, ok := scheme["headers"].(map[string]any)
		if !ok || headers["Authorization"] != "Bearer "+CustomGenericPlaceholder {
			t.Errorf("headers = %#v", scheme["headers"])
		}
		w.Header().Set("Content-Type", "application/json")
		writeJSON(t, w, `{"slug":"CUSTOM_ACME"}`)
	}))
	t.Cleanup(srv.Close)

	c := testClient(t, srv, Options{ProjectAPIKey: "project-secret"})
	out, err := c.UpsertCustomToolkit(context.Background(), UpsertCustomToolkitInput{
		Slug:   "ACME",
		Name:   "Acme",
		AppURL: "https://mcp.example.com/mcp",
		AuthSchemes: []CustomAuthScheme{{
			Mode:    CustomAuthModeAPIKey,
			Headers: map[string]string{"Authorization": "Bearer " + CustomGenericPlaceholder},
		}},
	})
	if err != nil {
		t.Fatalf("UpsertCustomToolkit: %v", err)
	}
	if out.Slug != "CUSTOM_ACME" {
		t.Fatalf("slug = %q", out.Slug)
	}
}

func TestUpsertCustomToolkitMissingSlug(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		writeJSON(t, w, `{}`)
	}))
	t.Cleanup(srv.Close)

	c := testClient(t, srv, Options{ProjectAPIKey: "project-secret"})
	_, err := c.UpsertCustomToolkit(context.Background(), UpsertCustomToolkitInput{
		Slug:        "ACME",
		Name:        "Acme",
		AppURL:      "https://mcp.example.com/mcp",
		AuthSchemes: []CustomAuthScheme{{Mode: CustomAuthModeNoAuth}},
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUpsertCustomToolkitConflict(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		writeJSON(t, w, `{"error":{"message":"app_url cannot change","code":409,"slug":"conflict","status":409}}`)
	}))
	t.Cleanup(srv.Close)

	c := testClient(t, srv, Options{ProjectAPIKey: "project-secret"})
	_, err := c.UpsertCustomToolkit(context.Background(), UpsertCustomToolkitInput{
		Slug:        "ACME",
		Name:        "Acme",
		AppURL:      "https://mcp.example.com/other",
		AuthSchemes: []CustomAuthScheme{{Mode: CustomAuthModeNoAuth}},
	})
	if !IsConflict(err) {
		t.Fatalf("IsConflict = false, err=%v", err)
	}
}

func TestSyncCustomToolkit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3.1/custom/toolkits/sync" {
			t.Errorf("path = %s", r.URL.Path)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read: %v", err)
		}
		var got syncCustomToolkitRequest
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if got.Slug != "CUSTOM_ACME" || got.ConnectedAccountID != "ca_1" {
			t.Errorf("body = %+v", got)
		}
		w.Header().Set("Content-Type", "application/json")
		writeJSON(t, w, `{"slug":"CUSTOM_ACME","version":"20260728_00","synced_count":12}`)
	}))
	t.Cleanup(srv.Close)

	c := testClient(t, srv, Options{ProjectAPIKey: "project-secret"})
	out, err := c.SyncCustomToolkit(context.Background(), SyncCustomToolkitInput{
		Slug:               "CUSTOM_ACME",
		ConnectedAccountID: "ca_1",
	})
	if err != nil {
		t.Fatalf("SyncCustomToolkit: %v", err)
	}
	if out.Slug != "CUSTOM_ACME" || out.Version != "20260728_00" || out.SyncedCount != 12 {
		t.Fatalf("out = %+v", out)
	}
}

func TestDeleteCustomToolkit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s", r.Method)
		}
		if r.URL.Path != "/api/v3.1/custom/toolkits/CUSTOM_ACME" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		writeJSON(t, w, `{
			"slug":"CUSTOM_ACME",
			"deleted":"true",
			"revoke_job_ids":["job_1"],
			"auth_configs_soft_deleted":1,
			"connected_accounts_soft_deleted":2
		}`)
	}))
	t.Cleanup(srv.Close)

	c := testClient(t, srv, Options{ProjectAPIKey: "project-secret"})
	out, err := c.DeleteCustomToolkit(context.Background(), "CUSTOM_ACME")
	if err != nil {
		t.Fatalf("DeleteCustomToolkit: %v", err)
	}
	if !out.Deleted || out.AuthConfigsDeleted != 1 || out.ConnectedAccountsDeleted != 2 {
		t.Fatalf("out = %+v", out)
	}
	if len(out.RevokeJobIDs) != 1 || out.RevokeJobIDs[0] != "job_1" {
		t.Fatalf("revoke jobs = %#v", out.RevokeJobIDs)
	}
}

func TestDeleteCustomToolkitNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		writeJSON(t, w, `{"error":{"message":"missing","code":404,"slug":"not_found","status":404}}`)
	}))
	t.Cleanup(srv.Close)

	c := testClient(t, srv, Options{ProjectAPIKey: "project-secret"})
	out, err := c.DeleteCustomToolkit(context.Background(), "CUSTOM_GONE")
	if err != nil {
		t.Fatalf("DeleteCustomToolkit: %v", err)
	}
	if !out.Deleted || out.Slug != "CUSTOM_GONE" {
		t.Fatalf("out = %+v", out)
	}
}
