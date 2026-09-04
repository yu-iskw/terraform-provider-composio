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
		writeJSON(t, w, `{"slug":"notion","name":"Notion","auth_config_details":[{"mode":"OAUTH2"},{"mode":"API_KEY"}]}`)
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

func TestEndpointRejectsPathAndQuery(t *testing.T) {
	for _, endpoint := range []string{
		"https://backend.composio.dev/api/v3.1",
		"https://backend.composio.dev/?x=1",
		"https://backend.composio.dev/#frag",
		"https://user:pass@backend.composio.dev",
	} {
		if _, err := New(Options{ProjectAPIKey: "k", Endpoint: endpoint}); err == nil {
			t.Fatalf("expected error for %s", endpoint)
		}
	}
	if _, err := New(Options{ProjectAPIKey: "k", Endpoint: "https://backend.composio.dev/"}); err != nil {
		t.Fatalf("origin with trailing slash must be accepted: %v", err)
	}
}

func TestClientRefusesCrossOriginRedirect(t *testing.T) {
	var leaked string
	dest := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		leaked = r.Header.Get("x-api-key")
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(dest.Close)

	src := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, dest.URL+"/api/v3.1/toolkits/github", http.StatusFound)
	}))
	t.Cleanup(src.Close)

	c := testClient(t, src, Options{ProjectAPIKey: "project-secret"})
	_, err := c.GetToolkit(context.Background(), "github")
	if err == nil {
		t.Fatal("expected cross-origin redirect to fail")
	}
	if leaked != "" {
		t.Fatalf("api key forwarded to %s", dest.URL)
	}
}

func TestRequestURLJoinsPinnedPrefix(t *testing.T) {
	got, err := requestURL("https://backend.composio.dev/", "/auth_configs/ac_1")
	if err != nil {
		t.Fatalf("requestURL: %v", err)
	}
	if got != "https://backend.composio.dev/api/v3.1/auth_configs/ac_1" {
		t.Fatalf("got %s", got)
	}
}

func TestCreateAuthConfigCustomUsesAuthScheme(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s", r.Method)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode: %v", err)
			return
		}
		ac, ok := body["auth_config"].(map[string]any)
		if !ok {
			t.Errorf("auth_config = %T", body["auth_config"])
			return
		}
		if ac["authScheme"] != "OAUTH2" {
			t.Errorf("authScheme = %v", ac["authScheme"])
		}
		if _, exists := ac["auth_scheme"]; exists {
			t.Errorf("unexpected snake_case auth_scheme in create body: %v", ac["auth_scheme"])
		}
		w.Header().Set("Content-Type", "application/json")
		writeJSON(t, w, `{"auth_config":{"id":"ac_custom"}}`)
	}))
	t.Cleanup(srv.Close)

	c := testClient(t, srv, Options{})
	id, err := c.CreateAuthConfig(context.Background(), CreateAuthConfigInput{
		ToolkitSlug: "github",
		AuthScheme:  "OAUTH2",
		Credentials: map[string]string{"client_id": "id"},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id != "ac_custom" {
		t.Fatalf("id = %s", id)
	}
}

func TestCreateAuthConfigEnabledForToolRouter(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode: %v", err)
			return
		}
		ac, ok := body["auth_config"].(map[string]any)
		if !ok {
			t.Errorf("auth_config = %T", body["auth_config"])
			return
		}
		if ac["is_enabled_for_tool_router"] != true {
			t.Errorf("is_enabled_for_tool_router = %v", ac["is_enabled_for_tool_router"])
		}
		if ac["type"] != "use_custom_auth" {
			t.Errorf("type = %v", ac["type"])
		}
		w.Header().Set("Content-Type", "application/json")
		writeJSON(t, w, `{"auth_config":{"id":"ac_router"}}`)
	}))
	t.Cleanup(srv.Close)

	enabled := true
	c := testClient(t, srv, Options{})
	id, err := c.CreateAuthConfig(context.Background(), CreateAuthConfigInput{
		ToolkitSlug:          "CUSTOM_ACME",
		AuthScheme:           "API_KEY",
		Credentials:          map[string]string{},
		EnabledForToolRouter: &enabled,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id != "ac_router" {
		t.Fatalf("id = %s", id)
	}
}

func TestUpdateAuthConfigEnabledForToolRouter(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("method = %s", r.Method)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode: %v", err)
			return
		}
		if body["type"] != "custom" {
			t.Errorf("type = %v", body["type"])
		}
		if body["is_enabled_for_tool_router"] != true {
			t.Errorf("is_enabled_for_tool_router = %v", body["is_enabled_for_tool_router"])
		}
		w.WriteHeader(http.StatusOK)
		writeJSON(t, w, `{"success":true,"message":"ok"}`)
	}))
	t.Cleanup(srv.Close)

	enabled := true
	c := testClient(t, srv, Options{})
	if err := c.UpdateAuthConfig(context.Background(), "ac_1", UpdateAuthConfigInput{
		Managed:              false,
		EnabledForToolRouter: &enabled,
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}
}

func TestGetAuthConfigReadsCredentialScopes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, `{
			"id":"ac_123",
			"name":"GitHub",
			"status":"ENABLED",
			"auth_scheme":"OAUTH2",
			"is_composio_managed":true,
			"credentials":{"scopes":"repo, read:org"},
			"toolkit":{"slug":"github"}
		}`)
	}))
	t.Cleanup(srv.Close)

	c := testClient(t, srv, Options{})
	got, err := c.GetAuthConfig(context.Background(), "ac_123")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Scopes == nil {
		t.Fatal("expected scopes from credentials")
	}
	if len(*got.Scopes) != 2 || (*got.Scopes)[0] != "repo" || (*got.Scopes)[1] != "read:org" {
		t.Fatalf("scopes = %#v", *got.Scopes)
	}
}

func TestUpdateAuthConfigClearsScopes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode: %v", err)
			return
		}
		if body["type"] != "default" {
			t.Errorf("type = %v", body["type"])
		}
		if body["scopes"] != "" {
			t.Errorf("scopes = %#v", body["scopes"])
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	c := testClient(t, srv, Options{})
	empty := []string{}
	if err := c.UpdateAuthConfig(context.Background(), "ac_123", UpdateAuthConfigInput{
		Managed: true,
		Scopes:  &empty,
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}
}

func TestUpdateAuthConfigClearsRestrictTools(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode: %v", err)
			return
		}
		raw, ok := body["restrict_to_following_tools"]
		if !ok {
			t.Error("restrict_to_following_tools omitted")
			return
		}
		got, ok := raw.([]any)
		if !ok || len(got) != 0 {
			t.Errorf("restrict_to_following_tools = %#v", raw)
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	c := testClient(t, srv, Options{})
	empty := []string{}
	if err := c.UpdateAuthConfig(context.Background(), "ac_123", UpdateAuthConfigInput{
		Managed:                  true,
		RestrictToFollowingTools: &empty,
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}
}

func TestCreateAuthConfigRequiresNestedID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		writeJSON(t, w, `{"id":"ac_top"}`)
	}))
	t.Cleanup(srv.Close)

	c := testClient(t, srv, Options{})
	if _, err := c.CreateAuthConfig(context.Background(), CreateAuthConfigInput{ToolkitSlug: "github", Managed: true}); err == nil {
		t.Fatal("top-level id must not be accepted")
	}
}

func TestParseCredentialScopes(t *testing.T) {
	got, ok := parseCredentialScopes([]byte(`{"scopes":["repo","gist"]}`))
	if !ok || len(got) != 2 || got[0] != "repo" {
		t.Fatalf("array scopes = %#v present=%v", got, ok)
	}
	got, ok = parseCredentialScopes([]byte(`{"client_id":"x"}`))
	if ok {
		t.Fatalf("missing scopes should be absent, got %#v", got)
	}
	got, ok = parseCredentialScopes([]byte(`{"scopes":""}`))
	if !ok || got == nil || len(got) != 0 {
		t.Fatalf("empty string scopes = %#v present=%v", got, ok)
	}
}
