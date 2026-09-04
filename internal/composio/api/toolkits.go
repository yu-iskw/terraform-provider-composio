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
	"net/http"
	"net/url"

	"github.com/yu-iskw/terraform-provider-composio/internal/composio/models"
)

func (c *Client) GetToolkit(ctx context.Context, slug string) (models.Toolkit, error) {
	var raw toolkitWire
	path := "/toolkits/" + url.PathEscape(slug)
	if err := c.Do(ctx, ScopeProject, http.MethodGet, path, nil, &raw); err != nil {
		return models.Toolkit{}, err
	}
	return raw.toModel(), nil
}

type toolkitWire struct {
	Slug              string            `json:"slug"`
	Name              string            `json:"name"`
	Type              string            `json:"type"`
	NoAuth            bool              `json:"no_auth"`
	AuthConfigDetails []toolkitAuthMode `json:"auth_config_details"`
	Meta              *toolkitMeta      `json:"meta"`
}

type toolkitAuthMode struct {
	Mode string `json:"mode"`
}

type toolkitMeta struct {
	Description   string `json:"description"`
	Logo          string `json:"logo"`
	AppURL        string `json:"app_url"`
	ToolsCount    int    `json:"tools_count"`
	TriggersCount int    `json:"triggers_count"`
	Version       string `json:"version"`
}

func (w toolkitWire) toModel() models.Toolkit {
	tk := models.Toolkit{
		Slug:   w.Slug,
		Name:   w.Name,
		Type:   w.Type,
		NoAuth: w.NoAuth,
	}
	if w.Meta != nil {
		tk.Description = w.Meta.Description
		tk.Logo = w.Meta.Logo
		tk.AppURL = w.Meta.AppURL
		tk.ToolsCount = w.Meta.ToolsCount
		tk.TriggersCount = w.Meta.TriggersCount
		tk.Version = w.Meta.Version
	}
	if len(w.AuthConfigDetails) > 0 {
		tk.AuthSchemes = make([]string, 0, len(w.AuthConfigDetails))
		for _, d := range w.AuthConfigDetails {
			if d.Mode != "" {
				tk.AuthSchemes = append(tk.AuthSchemes, d.Mode)
			}
		}
	}
	return tk
}
