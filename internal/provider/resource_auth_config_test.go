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
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/yu-iskw/terraform-provider-composio/internal/composio/api"
	"github.com/yu-iskw/terraform-provider-composio/internal/composio/models"
)

func TestAuthConfigResourceMetadata(t *testing.T) {
	res := NewAuthConfigResource()
	var resp fwresource.MetadataResponse
	res.Metadata(context.Background(), fwresource.MetadataRequest{ProviderTypeName: "composio"}, &resp)
	if resp.TypeName != "composio_auth_config" {
		t.Fatalf("got %q", resp.TypeName)
	}
}

func TestAccAuthConfigResource_lifecycle(t *testing.T) {
	if !isIntegrationTestMode() {
		t.Skip("Skipping acceptance test for resource_composio_auth_config")
	}

	providerConfig, err := getProviderConfig()
	if err != nil {
		t.Fatalf("getProviderConfig: %v", err)
	}

	createConfig, err := ReadAccTestResource([]string{"resources", "composio_auth_config", "lifecycle", "010_create.tf"})
	if err != nil {
		t.Fatalf("ReadAccTestResource create: %v", err)
	}
	updateConfig, err := ReadAccTestResource([]string{"resources", "composio_auth_config", "lifecycle", "020_update.tf"})
	if err != nil {
		t.Fatalf("ReadAccTestResource update: %v", err)
	}

	name := acctest.RandomWithPrefix("tf-acc")
	createConfig = substituteAccTestName(createConfig, name)
	updateConfig = substituteAccTestName(updateConfig, name)

	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_15_0),
		},
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAuthConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + createConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("composio_auth_config.test", "toolkit_slug", "github"),
					resource.TestCheckResourceAttr("composio_auth_config.test", "name", name),
					resource.TestCheckResourceAttr("composio_auth_config.test", "enabled", "true"),
					resource.TestCheckResourceAttr("composio_auth_config.test", "is_composio_managed", "true"),
					resource.TestCheckResourceAttrSet("composio_auth_config.test", "id"),
					resource.TestCheckResourceAttr("composio_auth_config.test", "managed_auth.restrict_to_following_tools.#", "2"),
				),
			},
			{
				Config: providerConfig + updateConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("composio_auth_config.test", "name", name),
					resource.TestCheckResourceAttr("composio_auth_config.test", "managed_auth.restrict_to_following_tools.#", "1"),
					resource.TestCheckTypeSetElemAttr("composio_auth_config.test", "managed_auth.restrict_to_following_tools.*", "GITHUB_CREATE_ISSUE"),
				),
			},
			{
				ResourceName:      "composio_auth_config.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAuthConfigDestroy(s *terraform.State) error {
	client, err := api.New(api.Options{
		Endpoint:      os.Getenv("COMPOSIO_ENDPOINT"),
		ProjectAPIKey: os.Getenv("COMPOSIO_API_KEY"),
	})
	if err != nil {
		return fmt.Errorf("create API client for CheckDestroy: %w", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "composio_auth_config" {
			continue
		}
		id := rs.Primary.ID
		if id == "" {
			continue
		}
		_, err := client.GetAuthConfig(context.Background(), id)
		if err == nil {
			return fmt.Errorf("auth config %s still exists", id)
		}
		if !api.IsNotFound(err) {
			return fmt.Errorf("checking auth config %s destroyed: %w", id, err)
		}
	}
	return nil
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

func TestAPIErrorStringOmitsResponseBody(t *testing.T) {
	msg := (&api.APIError{
		StatusCode:   400,
		Method:       "GET",
		Path:         "/auth_configs/x",
		Message:      "tool restriction is not valid for toolkit github",
		RequestID:    "req_9",
		Code:         "400",
		ResponseBody: `{"client_secret":"shh"}`,
	}).Error()
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

func TestApplyRemoteReadsRestrictTools(t *testing.T) {
	m := authConfigResourceModel{ManagedAuth: &managedAuthModel{RestrictToFollowingTools: types.SetNull(types.StringType)}}
	m.applyRemote(models.AuthConfig{
		ID:                       "ac_tools",
		IsComposioManaged:        true,
		Status:                   models.AuthConfigStatusEnabled,
		RestrictToFollowingTools: []string{"GITHUB_CREATE_ISSUE"},
	})
	var got []string
	if diags := m.ManagedAuth.RestrictToFollowingTools.ElementsAs(context.Background(), &got, false); diags.HasError() {
		t.Fatalf("tools: %v", diags)
	}
	if len(got) != 1 || got[0] != "GITHUB_CREATE_ISSUE" {
		t.Fatalf("got %#v", got)
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
	nullSet := types.SetNull(types.StringType)
	state := authConfigResourceModel{
		Name:    types.StringValue("GitHub"),
		Enabled: types.BoolValue(true),
		ManagedAuth: &managedAuthModel{
			RestrictToFollowingTools: nullSet,
			Scopes:                   nullSet,
		},
	}
	plan := authConfigResourceModel{
		Name:    types.StringValue("GitHub"),
		Enabled: types.BoolValue(false),
		ManagedAuth: &managedAuthModel{
			RestrictToFollowingTools: nullSet,
			Scopes:                   nullSet,
		},
	}
	if authConfigPatchNeeded(plan, state) {
		t.Fatal("enabled-only change must not PATCH auth config")
	}
	plan.Name = types.StringValue("Renamed")
	if !authConfigPatchNeeded(plan, state) {
		t.Fatal("name change must PATCH")
	}
}

func TestSetsEqualTreatsUntypedNullAsEqual(t *testing.T) {
	var zero types.Set
	typed := types.SetNull(types.StringType)
	if zero.Equal(typed) {
		t.Fatal("framework Equal must stay false for untyped null; otherwise setsEqual is redundant")
	}
	if !setsEqual(zero, typed) {
		t.Fatal("untyped and typed null sets must compare equal")
	}
	if !setsEqual(typed, typed) {
		t.Fatal("typed null sets must compare equal")
	}
}

func TestAuthConfigPatchNeededIgnoresUnknownComputedName(t *testing.T) {
	nullSet := types.SetNull(types.StringType)
	state := authConfigResourceModel{
		Name:        types.StringValue("GitHub"),
		ManagedAuth: &managedAuthModel{RestrictToFollowingTools: nullSet, Scopes: nullSet},
	}
	plan := authConfigResourceModel{
		Name:        types.StringUnknown(),
		ManagedAuth: &managedAuthModel{RestrictToFollowingTools: nullSet, Scopes: types.SetUnknown(types.StringType)},
	}
	if authConfigPatchNeeded(plan, state) {
		t.Fatal("unknown computed name/scopes must not PATCH")
	}
	plan.ManagedAuth.RestrictToFollowingTools = types.SetUnknown(types.StringType)
	if authConfigPatchNeeded(plan, state) {
		t.Fatal("unknown computed restrict_to_following_tools must not PATCH")
	}
}

func TestAuthConfigPatchNeededCustomToolsChange(t *testing.T) {
	nullSet := types.SetNull(types.StringType)
	state := authConfigResourceModel{
		CustomAuth: &customAuthModel{RestrictToFollowingTools: nullSet},
	}
	plan := authConfigResourceModel{
		Enabled:    types.BoolValue(false),
		CustomAuth: &customAuthModel{RestrictToFollowingTools: nullSet},
	}
	if authConfigPatchNeeded(plan, state) {
		t.Fatal("custom enabled-only change must not PATCH auth config")
	}
	plan.CustomAuth.RestrictToFollowingTools = types.SetValueMust(types.StringType, []attr.Value{types.StringValue("GITHUB_GET_ISSUE")})
	if !authConfigPatchNeeded(plan, state) {
		t.Fatal("custom tools change must PATCH")
	}
}

func TestApplyRemoteKeepsConfiguredToolkitSlugCasing(t *testing.T) {
	m := authConfigResourceModel{ToolkitSlug: types.StringValue("GITHUB")}
	m.applyRemote(models.AuthConfig{
		ID:                "ac_case",
		ToolkitSlug:       "github",
		IsComposioManaged: true,
		Status:            models.AuthConfigStatusEnabled,
	})
	if m.ToolkitSlug.ValueString() != "GITHUB" {
		t.Fatalf("configured casing must be kept, got %q", m.ToolkitSlug.ValueString())
	}
}

func TestUpdateInputOmitsNullRestrictTools(t *testing.T) {
	plan := authConfigResourceModel{
		ManagedAuth: &managedAuthModel{
			RestrictToFollowingTools: types.SetNull(types.StringType),
			Scopes:                   types.SetNull(types.StringType),
		},
	}
	in, diags := updateInputFromModels(context.Background(), plan, authConfigResourceModel{})
	if diags.HasError() {
		t.Fatalf("diags: %v", diags)
	}
	if in.RestrictToFollowingTools != nil {
		t.Fatal("null restrict_to_following_tools must be omitted from PATCH")
	}

	plan = authConfigResourceModel{
		CustomAuth: &customAuthModel{
			RestrictToFollowingTools: types.SetNull(types.StringType),
		},
	}
	in, diags = updateInputFromModels(context.Background(), plan, authConfigResourceModel{})
	if diags.HasError() {
		t.Fatalf("diags: %v", diags)
	}
	if in.RestrictToFollowingTools != nil {
		t.Fatal("null custom restrict_to_following_tools must be omitted from PATCH")
	}
}

func TestUpdateInputClearsRestrictTools(t *testing.T) {
	plan := authConfigResourceModel{
		ManagedAuth: &managedAuthModel{
			RestrictToFollowingTools: types.SetValueMust(types.StringType, []attr.Value{}),
			Scopes:                   types.SetNull(types.StringType),
		},
	}
	in, diags := updateInputFromModels(context.Background(), plan, authConfigResourceModel{})
	if diags.HasError() {
		t.Fatalf("diags: %v", diags)
	}
	if in.RestrictToFollowingTools == nil {
		t.Fatal("empty configured tools must be sent so the API can clear them")
	}
	if len(*in.RestrictToFollowingTools) != 0 {
		t.Fatalf("tools = %#v", *in.RestrictToFollowingTools)
	}
}

func TestCreateInputMapsRestrictTools(t *testing.T) {
	tools := types.SetValueMust(types.StringType, []attr.Value{types.StringValue("GITHUB_GET_REPOS")})
	in, diags := createInputFromModels(context.Background(), authConfigResourceModel{
		ToolkitSlug: types.StringValue("github"),
		ManagedAuth: &managedAuthModel{RestrictToFollowingTools: tools, Scopes: types.SetNull(types.StringType)},
	}, authConfigResourceModel{})
	if diags.HasError() {
		t.Fatalf("diags: %v", diags)
	}
	if len(in.RestrictToFollowingTools) != 1 || in.RestrictToFollowingTools[0] != "GITHUB_GET_REPOS" {
		t.Fatalf("managed tools = %#v", in.RestrictToFollowingTools)
	}

	in, diags = createInputFromModels(context.Background(), authConfigResourceModel{
		ToolkitSlug: types.StringValue("github"),
		CustomAuth:  &customAuthModel{AuthScheme: types.StringValue("OAUTH2"), RestrictToFollowingTools: tools},
	}, authConfigResourceModel{})
	if diags.HasError() {
		t.Fatalf("diags: %v", diags)
	}
	if len(in.RestrictToFollowingTools) != 1 || in.RestrictToFollowingTools[0] != "GITHUB_GET_REPOS" {
		t.Fatalf("custom tools = %#v", in.RestrictToFollowingTools)
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
