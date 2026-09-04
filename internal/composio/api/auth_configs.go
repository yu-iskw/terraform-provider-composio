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

	"github.com/yu-iskw/terraform-provider-composio/internal/composio/models"
)

type CreateAuthConfigInput struct {
	ToolkitSlug              string
	Name                     string
	Managed                  bool
	AuthScheme               string
	Credentials              map[string]string
	RestrictToFollowingTools []string
	Scopes                   []string
}

type UpdateAuthConfigInput struct {
	Managed                  bool
	Name                     *string
	Credentials              map[string]string
	RestrictToFollowingTools *[]string
	Scopes                   *string
}

func (c *Client) CreateAuthConfig(ctx context.Context, in CreateAuthConfigInput) (string, error) {
	authCfg := map[string]any{}
	if in.Managed {
		authCfg["type"] = models.AuthConfigCreateManaged
		if len(in.Scopes) > 0 {
			authCfg["credentials"] = map[string]any{"scopes": in.Scopes}
		} else if len(in.Credentials) > 0 {
			authCfg["credentials"] = stringMapToAny(in.Credentials)
		}
	} else {
		authCfg["type"] = models.AuthConfigCreateCustom
		if in.AuthScheme != "" {
			authCfg["auth_scheme"] = in.AuthScheme
		}
		if len(in.Credentials) > 0 {
			authCfg["credentials"] = stringMapToAny(in.Credentials)
		}
	}
	if in.Name != "" {
		authCfg["name"] = in.Name
	}
	if in.RestrictToFollowingTools != nil {
		authCfg["restrict_to_following_tools"] = in.RestrictToFollowingTools
	}

	body := map[string]any{
		"toolkit":     map[string]string{"slug": in.ToolkitSlug},
		"auth_config": authCfg,
	}

	var raw json.RawMessage
	if err := c.Do(ctx, ScopeProject, http.MethodPost, "/auth_configs", body, &raw); err != nil {
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
	body := map[string]any{}
	if in.Managed {
		body["type"] = models.AuthConfigUpdateDefault
		if in.Scopes != nil {
			body["scopes"] = *in.Scopes
		}
	} else {
		body["type"] = models.AuthConfigUpdateCustom
		if in.Credentials != nil {
			body["credentials"] = stringMapToAny(in.Credentials)
		}
	}
	if in.Name != nil {
		body["name"] = *in.Name
	}
	if in.RestrictToFollowingTools != nil {
		body["restrict_to_following_tools"] = *in.RestrictToFollowingTools
	}
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
	Type                     string          `json:"type"`
	AuthScheme               string          `json:"auth_scheme"`
	IsComposioManaged        bool            `json:"is_composio_managed"`
	RestrictToFollowingTools []string        `json:"restrict_to_following_tools"`
	CreatedAt                string          `json:"created_at"`
	LastUpdatedAt            string          `json:"last_updated_at"`
	NoOfConnections          int             `json:"no_of_connections"`
	Toolkit                  json.RawMessage `json:"toolkit"`
	AuthConfig               json.RawMessage `json:"auth_config"`
}

func (w authConfigWire) toModel() models.AuthConfig {
	slug := ""
	if len(w.Toolkit) > 0 {
		var obj struct {
			Slug string `json:"slug"`
		}
		if json.Unmarshal(w.Toolkit, &obj) == nil && obj.Slug != "" {
			slug = obj.Slug
		} else {
			var s string
			if json.Unmarshal(w.Toolkit, &s) == nil {
				slug = s
			}
		}
	}
	return models.AuthConfig{
		ID:                       w.ID,
		Name:                     w.Name,
		ToolkitSlug:              slug,
		AuthScheme:               w.AuthScheme,
		IsComposioManaged:        w.IsComposioManaged,
		Status:                   w.Status,
		Type:                     w.Type,
		RestrictToFollowingTools: w.RestrictToFollowingTools,
		CreatedAt:                w.CreatedAt,
		LastUpdatedAt:            w.LastUpdatedAt,
		NoOfConnections:          w.NoOfConnections,
	}
}

func authConfigIDFromCreate(raw json.RawMessage) (string, error) {
	var wrap struct {
		AuthConfig authConfigWire `json:"auth_config"`
		ID         string         `json:"id"`
	}
	if err := json.Unmarshal(raw, &wrap); err != nil {
		return "", fmt.Errorf("decode create auth config: %w", err)
	}
	if wrap.AuthConfig.ID != "" {
		return wrap.AuthConfig.ID, nil
	}
	if wrap.ID != "" {
		return wrap.ID, nil
	}
	return "", fmt.Errorf("create auth config response missing id")
}

func stringMapToAny(in map[string]string) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
