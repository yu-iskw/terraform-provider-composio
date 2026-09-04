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
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

// Custom toolkit auth modes from Composio Custom MCP docs.
const (
	CustomAuthModeNoAuth   = "NO_AUTH"
	CustomAuthModeAPIKey   = "API_KEY"
	CustomAuthModeDCROAuth = "DCR_OAUTH"
)

// CustomGenericPlaceholder is Composio's documented Custom MCP header
// substitution token. Servers replace it with the connected-account secret;
// the value is a template marker, not a credential.
const CustomGenericPlaceholder = "{{generic_api_key}}"

type CustomAuthScheme struct {
	Mode         string
	Headers      map[string]string
	DiscoveryURL string
}

type UpsertCustomToolkitInput struct {
	Slug        string
	Name        string
	AppURL      string
	AuthSchemes []CustomAuthScheme
}

type UpsertCustomToolkitResult struct {
	Slug string
}

type SyncCustomToolkitInput struct {
	Slug               string
	ConnectedAccountID string
}

type SyncCustomToolkitResult struct {
	Slug        string
	Version     string
	SyncedCount int
}

type DeleteCustomToolkitResult struct {
	Slug                     string
	Deleted                  bool
	RevokeJobIDs             []string
	AuthConfigsDeleted       int
	ConnectedAccountsDeleted int
}

func (c *Client) UpsertCustomToolkit(ctx context.Context, in UpsertCustomToolkitInput) (UpsertCustomToolkitResult, error) {
	req := upsertCustomToolkitRequest{
		Slug: in.Slug,
		ToolkitConfig: upsertCustomToolkitConfig{
			Name:   in.Name,
			AppURL: in.AppURL,
		},
	}
	req.ToolkitConfig.AuthSchemes = make([]customAuthSchemeWire, 0, len(in.AuthSchemes))
	for _, s := range in.AuthSchemes {
		wire := customAuthSchemeWire{Mode: s.Mode}
		if len(s.Headers) > 0 {
			wire.Headers = s.Headers
		}
		if s.DiscoveryURL != "" {
			wire.DiscoveryURL = s.DiscoveryURL
		}
		req.ToolkitConfig.AuthSchemes = append(req.ToolkitConfig.AuthSchemes, wire)
	}

	var out upsertCustomToolkitResponse
	if err := c.Do(ctx, ScopeProject, http.MethodPost, "/custom/toolkits/upsert", req, &out); err != nil {
		return UpsertCustomToolkitResult{}, err
	}
	if out.Slug == "" {
		return UpsertCustomToolkitResult{}, fmt.Errorf("upsert custom toolkit response missing slug")
	}
	return UpsertCustomToolkitResult(out), nil
}

func (c *Client) SyncCustomToolkit(ctx context.Context, in SyncCustomToolkitInput) (SyncCustomToolkitResult, error) {
	body := syncCustomToolkitRequest{Slug: in.Slug}
	if in.ConnectedAccountID != "" {
		body.ConnectedAccountID = in.ConnectedAccountID
	}
	var out syncCustomToolkitResponse
	if err := c.Do(ctx, ScopeProject, http.MethodPost, "/custom/toolkits/sync", body, &out); err != nil {
		return SyncCustomToolkitResult{}, err
	}
	return SyncCustomToolkitResult(out), nil
}

func (c *Client) DeleteCustomToolkit(ctx context.Context, slug string) (DeleteCustomToolkitResult, error) {
	path := "/custom/toolkits/" + url.PathEscape(slug)
	var out deleteCustomToolkitResponse
	if err := c.Do(ctx, ScopeProject, http.MethodDelete, path, nil, &out); err != nil {
		if IsNotFound(err) {
			return DeleteCustomToolkitResult{Slug: slug, Deleted: true}, nil
		}
		return DeleteCustomToolkitResult{}, err
	}
	return DeleteCustomToolkitResult{
		Slug:                     out.Slug,
		Deleted:                  out.Deleted.Bool(),
		RevokeJobIDs:             out.RevokeJobIDs,
		AuthConfigsDeleted:       firstNonZero(out.AuthConfigsDeleted, out.AuthConfigsSoftDeleted),
		ConnectedAccountsDeleted: firstNonZero(out.ConnectedAccountsDeleted, out.ConnectedAccountsSoftDeleted),
	}, nil
}

func IsConflict(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusConflict
}

type upsertCustomToolkitRequest struct {
	Slug          string                    `json:"slug"`
	ToolkitConfig upsertCustomToolkitConfig `json:"toolkit_config"`
}

type upsertCustomToolkitConfig struct {
	Name        string                 `json:"name"`
	AppURL      string                 `json:"app_url"`
	AuthSchemes []customAuthSchemeWire `json:"auth_schemes"`
}

type customAuthSchemeWire struct {
	Mode         string            `json:"mode"`
	Headers      map[string]string `json:"headers,omitempty"`
	DiscoveryURL string            `json:"discovery_url,omitempty"`
}

type upsertCustomToolkitResponse struct {
	Slug string `json:"slug"`
}

type syncCustomToolkitRequest struct {
	Slug               string `json:"slug"`
	ConnectedAccountID string `json:"connected_account_id,omitempty"`
}

type syncCustomToolkitResponse struct {
	Slug        string `json:"slug"`
	Version     string `json:"version"`
	SyncedCount int    `json:"synced_count"`
}

type deleteCustomToolkitResponse struct {
	Slug                         string   `json:"slug"`
	Deleted                      flexBool `json:"deleted"`
	RevokeJobIDs                 []string `json:"revoke_job_ids"`
	AuthConfigsDeleted           int      `json:"auth_configs_deleted"`
	AuthConfigsSoftDeleted       int      `json:"auth_configs_soft_deleted"`
	ConnectedAccountsDeleted     int      `json:"connected_accounts_deleted"`
	ConnectedAccountsSoftDeleted int      `json:"connected_accounts_soft_deleted"`
}

// flexBool accepts JSON true/false or the string "true"/"false".
type flexBool bool

func (b *flexBool) UnmarshalJSON(data []byte) error {
	var v bool
	if err := json.Unmarshal(data, &v); err == nil {
		*b = flexBool(v)
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	switch s {
	case "true", "TRUE", "1":
		*b = true
	case "false", "FALSE", "0", "":
		*b = false
	default:
		return fmt.Errorf("invalid boolean %q", s)
	}
	return nil
}

func (b flexBool) Bool() bool { return bool(b) }

func firstNonZero(vals ...int) int {
	for _, v := range vals {
		if v != 0 {
			return v
		}
	}
	return 0
}
