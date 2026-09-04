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
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func testClient(t *testing.T, srv *httptest.Server, opts Options) *Client {
	t.Helper()
	opts.Endpoint = srv.URL
	if opts.ProjectAPIKey == "" && opts.OrgAPIKey == "" {
		opts.ProjectAPIKey = "project-secret"
	}
	opts.Timeout = 5 * time.Second
	c, err := New(opts)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c
}

func writeJSON(t *testing.T, w http.ResponseWriter, body string) {
	t.Helper()
	if _, err := io.WriteString(w, body); err != nil {
		t.Errorf("write response: %v", err)
	}
}

func TestNewRequiresCredential(t *testing.T) {
	_, err := New(Options{})
	if err == nil {
		t.Fatal("expected error when both keys are empty")
	}
}

func TestDoSendsProjectAPIKey(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != "project-secret" {
			t.Errorf("x-api-key = %q", r.Header.Get("x-api-key"))
		}
		if r.Header.Get("x-org-api-key") != "" {
			t.Errorf("unexpected org header %q", r.Header.Get("x-org-api-key"))
		}
		if r.URL.Path != "/api/v3.1/toolkits/github" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "" {
			t.Error("Authorization header must not be set")
		}
		w.Header().Set("Content-Type", "application/json")
		writeJSON(t, w, `{"slug":"github","name":"GitHub"}`)
	}))
	t.Cleanup(srv.Close)

	c := testClient(t, srv, Options{ProjectAPIKey: "project-secret"})
	tk, err := c.GetToolkit(context.Background(), "github")
	if err != nil {
		t.Fatalf("GetToolkit: %v", err)
	}
	if tk.Slug != "github" || tk.Name != "GitHub" {
		t.Fatalf("toolkit = %+v", tk)
	}
}

func TestDoSendsOrgAPIKey(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-org-api-key") != "org-secret" {
			t.Errorf("x-org-api-key = %q", r.Header.Get("x-org-api-key"))
		}
		if r.Header.Get("x-api-key") != "" {
			t.Errorf("unexpected project header")
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	c := testClient(t, srv, Options{OrgAPIKey: "org-secret"})
	if err := c.Do(context.Background(), ScopeOrganization, http.MethodGet, "/org/owner/project/list", nil, nil); err != nil {
		t.Fatalf("Do: %v", err)
	}
}

func TestDoProjectScopeRequiresKey(t *testing.T) {
	c, err := New(Options{OrgAPIKey: "org-only", Endpoint: "https://backend.composio.dev"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	err = c.Do(context.Background(), ScopeProject, http.MethodGet, "/auth_configs/x", nil, nil)
	if err == nil {
		t.Fatal("expected missing api_key error")
	}
}

func TestRetryAfter429(t *testing.T) {
	var n atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if n.Add(1) == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			writeJSON(t, w, `{"error":{"message":"slow down","code":429,"slug":"rate_limited"}}`)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		writeJSON(t, w, `{"slug":"github","name":"GitHub"}`)
	}))
	t.Cleanup(srv.Close)

	c := testClient(t, srv, Options{})
	if _, err := c.GetToolkit(context.Background(), "github"); err != nil {
		t.Fatalf("GetToolkit: %v", err)
	}
	if n.Load() != 2 {
		t.Fatalf("attempts = %d", n.Load())
	}
}

func TestDoesNotRetryPOST(t *testing.T) {
	var n atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
		writeJSON(t, w, `{"error":{"message":"down"}}`)
	}))
	t.Cleanup(srv.Close)

	c := testClient(t, srv, Options{})
	_, err := c.CreateAuthConfig(context.Background(), CreateAuthConfigInput{ToolkitSlug: "github", Managed: true})
	if err == nil {
		t.Fatal("expected error")
	}
	if n.Load() != 1 {
		t.Fatalf("POST retries are forbidden, attempts = %d", n.Load())
	}
}

func TestNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		writeJSON(t, w, `{"error":{"message":"missing","code":404,"request_id":"req_1"}}`)
	}))
	t.Cleanup(srv.Close)

	c := testClient(t, srv, Options{})
	_, err := c.GetAuthConfig(context.Background(), "ac_missing")
	if !IsNotFound(err) {
		t.Fatalf("IsNotFound = false, err = %v", err)
	}
	var apiErr *APIError
	if !asAPIError(err, &apiErr) {
		t.Fatal("expected APIError")
	}
	if apiErr.RequestID != "req_1" {
		t.Fatalf("request id = %q", apiErr.RequestID)
	}
}

func asAPIError(err error, target **APIError) bool {
	e, ok := err.(*APIError)
	if !ok {
		return false
	}
	*target = e
	return true
}

func TestRedactsSecretsInAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(t, w, `{"error":{"message":"bad creds","code":400},"client_secret":"super-secret","nested":{"api_key":"abc"}}`)
	}))
	t.Cleanup(srv.Close)

	c := testClient(t, srv, Options{})
	_, err := c.GetAuthConfig(context.Background(), "ac_x")
	var apiErr *APIError
	if !asAPIError(err, &apiErr) {
		t.Fatalf("err = %v", err)
	}
	if strings.Contains(apiErr.ResponseBody, "super-secret") || strings.Contains(apiErr.ResponseBody, "abc") {
		t.Fatalf("secret leaked in %s", apiErr.ResponseBody)
	}
	if !strings.Contains(apiErr.ResponseBody, redacted) {
		t.Fatalf("expected redaction in %s", apiErr.ResponseBody)
	}
	if strings.Contains(apiErr.Error(), "super-secret") {
		t.Fatalf("secret leaked in Error(): %s", apiErr.Error())
	}
}

func TestCreateAuthConfigManagedThenGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v3.1/auth_configs":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Errorf("decode create body: %v", err)
				return
			}
			toolkit, ok := body["toolkit"].(map[string]any)
			if !ok {
				t.Errorf("toolkit = %T", body["toolkit"])
				return
			}
			if toolkit["slug"] != "github" {
				t.Errorf("toolkit slug = %v", toolkit["slug"])
			}
			ac, ok := body["auth_config"].(map[string]any)
			if !ok {
				t.Errorf("auth_config = %T", body["auth_config"])
				return
			}
			if ac["type"] != "use_composio_managed_auth" {
				t.Errorf("type = %v", ac["type"])
			}
			w.Header().Set("Content-Type", "application/json")
			writeJSON(t, w, `{"auth_config":{"id":"ac_123","auth_scheme":"OAUTH2","is_composio_managed":true}}`)
		case r.Method == http.MethodGet && r.URL.Path == "/api/v3.1/auth_configs/ac_123":
			w.Header().Set("Content-Type", "application/json")
			writeJSON(t, w, `{
				"id":"ac_123",
				"name":"GitHub",
				"status":"ENABLED",
				"auth_scheme":"OAUTH2",
				"is_composio_managed":true,
				"restrict_to_following_tools":["GITHUB_CREATE_ISSUE"],
				"toolkit":{"slug":"github"},
				"created_at":"2026-01-01T00:00:00Z"
			}`)
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	c := testClient(t, srv, Options{})
	id, err := c.CreateAuthConfig(context.Background(), CreateAuthConfigInput{
		ToolkitSlug:              "github",
		Managed:                  true,
		RestrictToFollowingTools: []string{"GITHUB_CREATE_ISSUE"},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id != "ac_123" {
		t.Fatalf("id = %s", id)
	}
	got, err := c.GetAuthConfig(context.Background(), id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.ToolkitSlug != "github" || !got.Enabled() || got.AuthScheme != "OAUTH2" {
		t.Fatalf("model = %+v", got)
	}
}

func TestDeleteAuthConfigTreats404AsSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		writeJSON(t, w, `{"error":{"message":"gone"}}`)
	}))
	t.Cleanup(srv.Close)

	c := testClient(t, srv, Options{})
	if err := c.DeleteAuthConfig(context.Background(), "ac_gone"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestSetAuthConfigStatusPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/api/v3.1/auth_configs/ac_123/DISABLED" {
			t.Errorf("got %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	c := testClient(t, srv, Options{})
	if err := c.SetAuthConfigStatus(context.Background(), "ac_123", "DISABLED"); err != nil {
		t.Fatalf("status: %v", err)
	}
}

func TestToolkitAuthSchemesFromObjects(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, `{"slug":"notion","name":"Notion","auth_schemes":[{"mode":"OAUTH2"},{"mode":"API_KEY"}]}`)
	}))
	t.Cleanup(srv.Close)

	c := testClient(t, srv, Options{})
	tk, err := c.GetToolkit(context.Background(), "notion")
	if err != nil {
		t.Fatalf("GetToolkit: %v", err)
	}
	if len(tk.AuthSchemes) != 2 || tk.AuthSchemes[0] != "OAUTH2" {
		t.Fatalf("schemes = %v", tk.AuthSchemes)
	}
}

func TestToolkitNestedMetaAndAuthConfigDetails(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, `{
			"slug":"github",
			"name":"GitHub",
			"no_auth":false,
			"auth_config_details":[{"mode":"OAUTH2"},{"mode":"API_KEY"}],
			"meta":{
				"description":"GitHub toolkit",
				"logo":"https://example.com/github.png",
				"app_url":"https://github.com",
				"tools_count":12,
				"triggers_count":3,
				"version":"20260301_00"
			}
		}`)
	}))
	t.Cleanup(srv.Close)

	c := testClient(t, srv, Options{})
	tk, err := c.GetToolkit(context.Background(), "github")
	if err != nil {
		t.Fatalf("GetToolkit: %v", err)
	}
	if tk.Description != "GitHub toolkit" || tk.Logo == "" || tk.AppURL != "https://github.com" {
		t.Fatalf("meta = %+v", tk)
	}
	if tk.ToolsCount != 12 || tk.TriggersCount != 3 || tk.Version != "20260301_00" {
		t.Fatalf("counts = %+v", tk)
	}
	if len(tk.AuthSchemes) != 2 || tk.AuthSchemes[0] != "OAUTH2" {
		t.Fatalf("schemes = %v", tk.AuthSchemes)
	}
}

func TestInvalidEndpoint(t *testing.T) {
	_, err := New(Options{ProjectAPIKey: "k", Endpoint: "not-a-url"})
	if err == nil {
		t.Fatal("expected invalid endpoint")
	}
}
