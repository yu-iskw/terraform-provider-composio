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
	Slug              string          `json:"slug"`
	Name              string          `json:"name"`
	Description       string          `json:"description"`
	Logo              string          `json:"logo"`
	Type              string          `json:"type"`
	NoAuth            bool            `json:"no_auth"`
	AuthSchemes       json.RawMessage `json:"auth_schemes"`
	AuthConfigDetails json.RawMessage `json:"auth_config_details"`
	ToolsCount        int             `json:"tools_count"`
	TriggersCount     int             `json:"triggers_count"`
	Version           string          `json:"version"`
	AppURL            string          `json:"app_url"`
	Meta              *toolkitMeta    `json:"meta"`
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
	desc := w.Description
	logo := w.Logo
	appURL := w.AppURL
	tools := w.ToolsCount
	triggers := w.TriggersCount
	version := w.Version
	if w.Meta != nil {
		desc = firstNonEmpty(desc, w.Meta.Description)
		logo = firstNonEmpty(logo, w.Meta.Logo)
		appURL = firstNonEmpty(appURL, w.Meta.AppURL)
		if tools == 0 {
			tools = w.Meta.ToolsCount
		}
		if triggers == 0 {
			triggers = w.Meta.TriggersCount
		}
		version = firstNonEmpty(version, w.Meta.Version)
	}
	schemes := parseAuthSchemes(w.AuthSchemes)
	if len(schemes) == 0 {
		schemes = parseAuthSchemes(w.AuthConfigDetails)
	}
	return models.Toolkit{
		Slug:          w.Slug,
		Name:          w.Name,
		Description:   desc,
		Logo:          logo,
		Type:          w.Type,
		NoAuth:        w.NoAuth,
		AuthSchemes:   schemes,
		ToolsCount:    tools,
		TriggersCount: triggers,
		Version:       version,
		AppURL:        appURL,
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func parseAuthSchemes(raw json.RawMessage) []string {
	if len(raw) == 0 {
		return nil
	}
	var strings []string
	if json.Unmarshal(raw, &strings) == nil {
		return strings
	}
	var objs []struct {
		Mode string `json:"mode"`
		Name string `json:"name"`
	}
	if json.Unmarshal(raw, &objs) == nil {
		out := make([]string, 0, len(objs))
		for _, o := range objs {
			if o.Mode != "" {
				out = append(out, o.Mode)
				continue
			}
			if o.Name != "" {
				out = append(out, o.Name)
			}
		}
		return out
	}
	return nil
}
