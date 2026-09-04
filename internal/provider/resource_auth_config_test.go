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

package provider

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yu-iskw/terraform-provider-composio/internal/composio/api"
	"github.com/yu-iskw/terraform-provider-composio/internal/composio/models"
)

func TestAuthConfigResourceMetadata(t *testing.T) {
	res := NewAuthConfigResource()
	var resp resource.MetadataResponse
	res.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "composio"}, &resp)
	if resp.TypeName != "composio_auth_config" {
		t.Fatalf("got %q", resp.TypeName)
	}
}

func TestApplyRemoteManaged(t *testing.T) {
	m := authConfigResourceModel{}
	m.applyRemote(models.AuthConfig{
		ID:                       "ac_1",
		Name:                     "GitHub",
		ToolkitSlug:              "github",
		AuthScheme:               "OAUTH2",
		IsComposioManaged:        true,
		Status:                   models.AuthConfigStatusEnabled,
		RestrictToFollowingTools: []string{"GITHUB_CREATE_ISSUE"},
		CreatedAt:                "2026-01-01T00:00:00Z",
	})
	if m.ID.ValueString() != "ac_1" {
		t.Fatalf("id = %s", m.ID.ValueString())
	}
	if !m.Enabled.ValueBool() {
		t.Fatal("expected enabled")
	}
	if m.ManagedAuth == nil || m.CustomAuth != nil {
		t.Fatal("expected managed_auth only")
	}
}

func TestApplyRemoteCustom(t *testing.T) {
	m := authConfigResourceModel{}
	m.applyRemote(models.AuthConfig{
		ID:                "ac_2",
		Name:              "Notion",
		ToolkitSlug:       "notion",
		AuthScheme:        "OAUTH2",
		IsComposioManaged: false,
		Status:            models.AuthConfigStatusDisabled,
	})
	if m.Enabled.ValueBool() {
		t.Fatal("expected disabled")
	}
	if m.CustomAuth == nil || m.ManagedAuth != nil {
		t.Fatal("expected custom_auth only")
	}
}

func TestFormatAPIError(t *testing.T) {
	msg := formatAPIError(&api.APIError{
		StatusCode:   400,
		Message:      "tool restriction is not valid for toolkit github",
		RequestID:    "req_9",
		Code:         "400",
		ResponseBody: `{"client_secret":"shh"}`,
	})
	if !strings.Contains(msg, "tool restriction") || !strings.Contains(msg, "req_9") {
		t.Fatalf("msg = %s", msg)
	}
	if strings.Contains(msg, "shh") {
		t.Fatalf("secret leaked: %s", msg)
	}
}

func TestApplyRemoteKeepsOmittedToolsNull(t *testing.T) {
	m := authConfigResourceModel{}
	m.applyRemote(models.AuthConfig{
		ID:                "ac_3",
		IsComposioManaged: true,
		Status:            models.AuthConfigStatusEnabled,
	})
	if m.ManagedAuth == nil || !m.ManagedAuth.RestrictToFollowingTools.IsNull() {
		t.Fatal("omitted restrict_to_following_tools must stay null")
	}
}

func TestOptionalStringSetPreservesEmptyConfig(t *testing.T) {
	empty := types.SetValueMust(types.StringType, []attr.Value{})
	got := optionalStringSet(nil, empty)
	if got.IsNull() {
		t.Fatal("explicit empty set must stay empty")
	}
	if len(got.Elements()) != 0 {
		t.Fatalf("got %#v", got.Elements())
	}
}

func TestApplyRemoteReadsScopes(t *testing.T) {
	scopes := []string{"repo", "read:org"}
	m := authConfigResourceModel{}
	m.applyRemote(models.AuthConfig{
		ID:                "ac_scopes",
		IsComposioManaged: true,
		Status:            models.AuthConfigStatusEnabled,
		Scopes:            &scopes,
	})
	if m.ManagedAuth == nil {
		t.Fatal("expected managed_auth")
	}
	var got []string
	if diags := m.ManagedAuth.Scopes.ElementsAs(context.Background(), &got, false); diags.HasError() {
		t.Fatalf("scopes: %v", diags)
	}
	if len(got) != 2 || got[0] != "repo" {
		t.Fatalf("scopes = %#v", got)
	}
}

func TestApplyRemoteKeepsPriorScopesWhenRemoteOmits(t *testing.T) {
	prior := types.SetValueMust(types.StringType, []attr.Value{types.StringValue("repo")})
	m := authConfigResourceModel{ManagedAuth: &managedAuthModel{Scopes: prior}}
	m.applyRemote(models.AuthConfig{
		ID:                "ac_omit",
		IsComposioManaged: true,
		Status:            models.AuthConfigStatusEnabled,
	})
	var got []string
	if diags := m.ManagedAuth.Scopes.ElementsAs(context.Background(), &got, false); diags.HasError() {
		t.Fatalf("scopes: %v", diags)
	}
	if len(got) != 1 || got[0] != "repo" {
		t.Fatalf("prior scopes must remain when GET omits credentials.scopes, got %#v", got)
	}
}

func TestApplyRemoteClearsScopesWhenRemoteEmpty(t *testing.T) {
	empty := []string{}
	prior := types.SetValueMust(types.StringType, []attr.Value{types.StringValue("repo")})
	m := authConfigResourceModel{ManagedAuth: &managedAuthModel{Scopes: prior}}
	m.applyRemote(models.AuthConfig{
		ID:                "ac_empty",
		IsComposioManaged: true,
		Status:            models.AuthConfigStatusEnabled,
		Scopes:            &empty,
	})
	if m.ManagedAuth.Scopes.IsNull() {
		t.Fatal("cleared scopes should be an empty set, not null")
	}
	if len(m.ManagedAuth.Scopes.Elements()) != 0 {
		t.Fatalf("got %#v", m.ManagedAuth.Scopes.Elements())
	}
}

func TestAuthConfigPatchNeededSkipsEnabledOnly(t *testing.T) {
	state := authConfigResourceModel{
		Name:        types.StringValue("GitHub"),
		Enabled:     types.BoolValue(true),
		ManagedAuth: &managedAuthModel{Scopes: types.SetNull(types.StringType)},
	}
	plan := authConfigResourceModel{
		Name:        types.StringValue("GitHub"),
		Enabled:     types.BoolValue(false),
		ManagedAuth: &managedAuthModel{Scopes: types.SetNull(types.StringType)},
	}
	if authConfigPatchNeeded(plan, state, authConfigResourceModel{}) {
		t.Fatal("enabled-only change must not PATCH auth config")
	}
	plan.Name = types.StringValue("Renamed")
	if !authConfigPatchNeeded(plan, state, authConfigResourceModel{}) {
		t.Fatal("name change must PATCH")
	}
}

func TestUpdateInputClearsManagedScopes(t *testing.T) {
	plan := authConfigResourceModel{
		ManagedAuth: &managedAuthModel{
			RestrictToFollowingTools: types.SetNull(types.StringType),
			Scopes:                   types.SetValueMust(types.StringType, []attr.Value{}),
		},
	}
	in, diags := updateInputFromModels(context.Background(), plan, authConfigResourceModel{})
	if diags.HasError() {
		t.Fatalf("diags: %v", diags)
	}
	if in.Scopes == nil {
		t.Fatal("empty configured scopes must be sent so the API can clear them")
	}
	if len(*in.Scopes) != 0 {
		t.Fatalf("scopes = %#v", *in.Scopes)
	}
}
