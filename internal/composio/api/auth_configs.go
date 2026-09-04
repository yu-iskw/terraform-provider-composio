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
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/yu-iskw/terraform-provider-composio/internal/composio/models"
)

const (
	authConfigCreateManaged = "use_composio_managed_auth"
	authConfigCreateCustom  = "use_custom_auth"
	authConfigUpdateDefault = "default"
	authConfigUpdateCustom  = "custom"
)

type CreateAuthConfigInput struct {
	ToolkitSlug              string
	Name                     string
	Managed                  bool
	AuthScheme               string
	Credentials              map[string]string
	RestrictToFollowingTools []string
	Scopes                   []string
	EnabledForToolRouter     *bool
}

type UpdateAuthConfigInput struct {
	Managed                  bool
	Name                     *string
	Credentials              map[string]string
	RestrictToFollowingTools *[]string
	Scopes                   *[]string
	EnabledForToolRouter     *bool
}

func (c *Client) CreateAuthConfig(ctx context.Context, in CreateAuthConfigInput) (string, error) {
	req := createAuthConfigRequest{
		Toolkit: toolkitRef{Slug: in.ToolkitSlug},
		AuthConfig: createAuthConfigBody{
			Type: authConfigCreateCustom,
		},
	}
	if in.Managed {
		req.AuthConfig.Type = authConfigCreateManaged
		if len(in.Scopes) > 0 {
			req.AuthConfig.Credentials = map[string]any{"scopes": in.Scopes}
		}
	} else {
		req.AuthConfig.AuthScheme = in.AuthScheme
		if len(in.Credentials) > 0 {
			req.AuthConfig.Credentials = stringMapToAny(in.Credentials)
		}
	}
	if in.Name != "" {
		req.AuthConfig.Name = in.Name
	}
	if in.RestrictToFollowingTools != nil {
		req.AuthConfig.RestrictToFollowingTools = in.RestrictToFollowingTools
	}
	if in.EnabledForToolRouter != nil {
		req.AuthConfig.IsEnabledForToolRouter = in.EnabledForToolRouter
	}

	var raw json.RawMessage
	if err := c.Do(ctx, ScopeProject, http.MethodPost, "/auth_configs", req, &raw); err != nil {
		return "", err
	}
	id, err := authConfigIDFromCreate(raw)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (c *Client) GetAuthConfig(ctx context.Context, id string) (models.AuthConfig, error) {
	var raw authConfigWire
	path := "/auth_configs/" + url.PathEscape(id)
	if err := c.Do(ctx, ScopeProject, http.MethodGet, path, nil, &raw); err != nil {
		return models.AuthConfig{}, err
	}
	return raw.toModel(), nil
}

func (c *Client) UpdateAuthConfig(ctx context.Context, id string, in UpdateAuthConfigInput) error {
	body := updateAuthConfigBody{Type: authConfigUpdateCustom}
	if in.Managed {
		body.Type = authConfigUpdateDefault
		if in.Scopes != nil {
			joined := strings.Join(*in.Scopes, ",")
			body.Scopes = &joined
		}
	} else if in.Credentials != nil {
		body.Credentials = stringMapToAny(in.Credentials)
	}
	body.Name = in.Name
	body.RestrictToFollowingTools = in.RestrictToFollowingTools
	body.IsEnabledForToolRouter = in.EnabledForToolRouter
	path := "/auth_configs/" + url.PathEscape(id)
	return c.Do(ctx, ScopeProject, http.MethodPatch, path, body, nil)
}

func (c *Client) DeleteAuthConfig(ctx context.Context, id string) error {
	path := "/auth_configs/" + url.PathEscape(id)
	err := c.Do(ctx, ScopeProject, http.MethodDelete, path, nil, nil)
	if IsNotFound(err) {
		return nil
	}
	return err
}

func (c *Client) SetAuthConfigStatus(ctx context.Context, id, status string) error {
	path := "/auth_configs/" + url.PathEscape(id) + "/" + url.PathEscape(status)
	return c.Do(ctx, ScopeProject, http.MethodPatch, path, nil, nil)
}

type authConfigWire struct {
	ID                       string          `json:"id"`
	Name                     string          `json:"name"`
	Status                   string          `json:"status"`
	AuthScheme               string          `json:"auth_scheme"`
	IsComposioManaged        bool            `json:"is_composio_managed"`
	RestrictToFollowingTools []string        `json:"restrict_to_following_tools"`
	CreatedAt                string          `json:"created_at"`
	IsEnabledForToolRouter   *bool           `json:"is_enabled_for_tool_router"`
	Credentials              json.RawMessage `json:"credentials"`
	Toolkit                  struct {
		Slug string `json:"slug"`
	} `json:"toolkit"`
}

type toolkitRef struct {
	Slug string `json:"slug"`
}

type createAuthConfigRequest struct {
	Toolkit    toolkitRef           `json:"toolkit"`
	AuthConfig createAuthConfigBody `json:"auth_config"`
}

type createAuthConfigBody struct {
	Type                     string         `json:"type"`
	Name                     string         `json:"name,omitempty"`
	AuthScheme               string         `json:"authScheme,omitempty"`
	Credentials              map[string]any `json:"credentials,omitempty"`
	RestrictToFollowingTools []string       `json:"restrict_to_following_tools,omitempty"`
	IsEnabledForToolRouter   *bool          `json:"is_enabled_for_tool_router,omitempty"`
}

type updateAuthConfigBody struct {
	Type                     string         `json:"type"`
	Name                     *string        `json:"name,omitempty"`
	Scopes                   *string        `json:"scopes,omitempty"`
	Credentials              map[string]any `json:"credentials,omitempty"`
	RestrictToFollowingTools *[]string      `json:"restrict_to_following_tools,omitempty"`
	IsEnabledForToolRouter   *bool          `json:"is_enabled_for_tool_router,omitempty"`
}

func (w authConfigWire) toModel() models.AuthConfig {
	scopes, scopesSet := parseCredentialScopes(w.Credentials)
	var scopePtr *[]string
	if scopesSet {
		scopePtr = &scopes
	}
	return models.AuthConfig{
		ID:                       w.ID,
		Name:                     w.Name,
		ToolkitSlug:              w.Toolkit.Slug,
		AuthScheme:               w.AuthScheme,
		IsComposioManaged:        w.IsComposioManaged,
		Status:                   w.Status,
		RestrictToFollowingTools: w.RestrictToFollowingTools,
		Scopes:                   scopePtr,
		CreatedAt:                w.CreatedAt,
		IsEnabledForToolRouter:   w.IsEnabledForToolRouter,
	}
}

func authConfigIDFromCreate(raw json.RawMessage) (string, error) {
	var wrap struct {
		AuthConfig struct {
			ID string `json:"id"`
		} `json:"auth_config"`
	}
	if err := json.Unmarshal(raw, &wrap); err != nil {
		return "", fmt.Errorf("decode create auth config: %w", err)
	}
	if wrap.AuthConfig.ID == "" {
		return "", fmt.Errorf("create auth config response missing auth_config.id")
	}
	return wrap.AuthConfig.ID, nil
}

func stringMapToAny(in map[string]string) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func parseCredentialScopes(raw json.RawMessage) ([]string, bool) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, false
	}
	var obj map[string]json.RawMessage
	if json.Unmarshal(raw, &obj) != nil {
		return nil, false
	}
	s, ok := obj["scopes"]
	if !ok || len(s) == 0 || string(s) == "null" {
		return nil, false
	}
	var str string
	if json.Unmarshal(s, &str) == nil {
		return splitCommaList(str), true
	}
	var arr []string
	if json.Unmarshal(s, &arr) == nil {
		return arr, true
	}
	return nil, false
}

func splitCommaList(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return []string{}
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
