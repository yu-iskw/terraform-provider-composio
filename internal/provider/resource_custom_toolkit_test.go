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

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yu-iskw/terraform-provider-composio/internal/composio/api"
	"github.com/yu-iskw/terraform-provider-composio/internal/composio/models"
)

func TestCustomToolkitResourceMetadata(t *testing.T) {
	var resp fwresource.MetadataResponse
	NewCustomToolkitResource().Metadata(context.Background(), fwresource.MetadataRequest{ProviderTypeName: "composio"}, &resp)
	if resp.TypeName != "composio_custom_toolkit" {
		t.Fatalf("got %q", resp.TypeName)
	}
}

func TestStripCustomPrefix(t *testing.T) {
	if got := stripCustomPrefix("CUSTOM_ACME"); got != "ACME" {
		t.Fatalf("got %q", got)
	}
	if got := stripCustomPrefix("ACME"); got != "ACME" {
		t.Fatalf("got %q", got)
	}
}

func TestHeadersContainAPIKeyPlaceholder(t *testing.T) {
	if !headersContainAPIKeyPlaceholder(map[string]string{
		"Authorization": "Bearer " + api.CustomGenericPlaceholder,
	}) {
		t.Fatal("expected true")
	}
	if headersContainAPIKeyPlaceholder(map[string]string{"Authorization": "Bearer secret"}) {
		t.Fatal("expected false")
	}
}

func TestCustomToolkitApplyRemotePreservesAuthScheme(t *testing.T) {
	m := customToolkitResourceModel{
		Slug: types.StringValue("ACME"),
		AuthScheme: &customToolkitAuthModel{
			Mode:    types.StringValue(api.CustomAuthModeNoAuth),
			Headers: types.MapNull(types.StringType),
		},
	}
	m.applyRemote(models.Toolkit{
		Slug:       "CUSTOM_ACME",
		Name:       "Acme",
		Type:       "custom",
		AppURL:     "https://mcp.example.com/mcp",
		ToolsCount: 3,
		Version:    "20260728_00",
		Logo:       "https://example.com/logo.png",
	})
	if m.ID.ValueString() != "CUSTOM_ACME" {
		t.Fatalf("id = %q", m.ID.ValueString())
	}
	if m.Slug.ValueString() != "ACME" {
		t.Fatalf("slug = %q", m.Slug.ValueString())
	}
	if m.AuthScheme == nil || m.AuthScheme.Mode.ValueString() != api.CustomAuthModeNoAuth {
		t.Fatalf("auth scheme = %+v", m.AuthScheme)
	}
	if m.ToolsCount.ValueInt64() != 3 || m.Type.ValueString() != "custom" {
		t.Fatalf("computed = tools=%d type=%s", m.ToolsCount.ValueInt64(), m.Type.ValueString())
	}
}

func TestAccCustomToolkitResource_lifecycle(t *testing.T) {
	if !isIntegrationTestMode() {
		t.Skip("Skipping acceptance test for composio_custom_toolkit")
	}
	mcpURL := strings.TrimSpace(os.Getenv("COMPOSIO_ACC_CUSTOM_MCP_URL"))
	if mcpURL == "" {
		t.Skip("Skipping acceptance test: COMPOSIO_ACC_CUSTOM_MCP_URL is unset")
	}

	providerConfig, err := getProviderConfig()
	if err != nil {
		t.Fatalf("getProviderConfig: %v", err)
	}

	createConfig, err := ReadAccTestResource([]string{"resources", "composio_custom_toolkit", "lifecycle", "010_create.tf"})
	if err != nil {
		t.Fatalf("ReadAccTestResource create: %v", err)
	}
	updateConfig, err := ReadAccTestResource([]string{"resources", "composio_custom_toolkit", "lifecycle", "020_update.tf"})
	if err != nil {
		t.Fatalf("ReadAccTestResource update: %v", err)
	}

	suffix := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	slug := "TFACC" + strings.ToUpper(suffix)
	name := "tf-acc-" + suffix
	nameUpdated := name + "-updated"

	createConfig = substituteAccTestName(createConfig, name)
	createConfig = strings.ReplaceAll(createConfig, "__SLUG__", slug)
	createConfig = strings.ReplaceAll(createConfig, "__MCP_URL__", mcpURL)
	updateConfig = substituteAccTestName(updateConfig, nameUpdated)
	updateConfig = strings.ReplaceAll(updateConfig, "__SLUG__", slug)
	updateConfig = strings.ReplaceAll(updateConfig, "__MCP_URL__", mcpURL)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCustomToolkitDestroy,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + createConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("composio_custom_toolkit.test", "slug", slug),
					resource.TestCheckResourceAttr("composio_custom_toolkit.test", "name", name),
					resource.TestCheckResourceAttr("composio_custom_toolkit.test", "app_url", mcpURL),
					resource.TestCheckResourceAttr("composio_custom_toolkit.test", "auth_scheme.mode", "NO_AUTH"),
					resource.TestCheckResourceAttr("composio_custom_toolkit.test", "id", "CUSTOM_"+slug),
					resource.TestCheckResourceAttr("composio_custom_toolkit.test", "type", "custom"),
				),
			},
			{
				Config: providerConfig + updateConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("composio_custom_toolkit.test", "name", nameUpdated),
					resource.TestCheckResourceAttr("composio_custom_toolkit.test", "id", "CUSTOM_"+slug),
				),
			},
			{
				ResourceName:            "composio_custom_toolkit.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"auth_scheme"},
			},
		},
	})
}

func testAccCheckCustomToolkitDestroy(s *terraform.State) error {
	client, err := api.New(api.Options{
		Endpoint:      os.Getenv("COMPOSIO_ENDPOINT"),
		ProjectAPIKey: os.Getenv("COMPOSIO_API_KEY"),
	})
	if err != nil {
		return fmt.Errorf("create API client for CheckDestroy: %w", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "composio_custom_toolkit" {
			continue
		}
		id := rs.Primary.ID
		if id == "" {
			continue
		}
		_, err := client.GetToolkit(context.Background(), id)
		if err == nil {
			return fmt.Errorf("custom toolkit %s still exists", id)
		}
		if !api.IsNotFound(err) {
			return fmt.Errorf("checking custom toolkit %s destroyed: %w", id, err)
		}
	}
	return nil
}
