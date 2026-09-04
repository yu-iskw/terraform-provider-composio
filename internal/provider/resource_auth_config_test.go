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

	"github.com/hashicorp/terraform-plugin-framework/resource"
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
	diags := m.applyRemote(context.Background(), models.AuthConfig{
		ID:                       "ac_1",
		Name:                     "GitHub",
		ToolkitSlug:              "github",
		AuthScheme:               "OAUTH2",
		IsComposioManaged:        true,
		Status:                   models.AuthConfigStatusEnabled,
		RestrictToFollowingTools: []string{"GITHUB_CREATE_ISSUE"},
		CreatedAt:                "2026-01-01T00:00:00Z",
	})
	if diags.HasError() {
		t.Fatalf("%v", diags)
	}
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
	diags := m.applyRemote(context.Background(), models.AuthConfig{
		ID:                "ac_2",
		Name:              "Notion",
		ToolkitSlug:       "notion",
		AuthScheme:        "OAUTH2",
		IsComposioManaged: false,
		Status:            models.AuthConfigStatusDisabled,
	})
	if diags.HasError() {
		t.Fatalf("%v", diags)
	}
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
