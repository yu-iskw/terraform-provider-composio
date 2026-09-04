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
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestToolkitDataSourceMetadata(t *testing.T) {
	ds := NewToolkitDataSource()
	var resp datasource.MetadataResponse
	ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "composio"}, &resp)
	if resp.TypeName != "composio_toolkit" {
		t.Fatalf("got %q", resp.TypeName)
	}
}

func TestAccPreCheckSkipsWithoutTFACC(t *testing.T) {
	if isIntegrationTestMode() {
		t.Skip("running under TF_ACC")
	}
	testAccPreCheck(t)
}

func TestAccToolkitDataSource(t *testing.T) {
	if !isIntegrationTestMode() {
		t.Skip("Skipping acceptance test for data_source_composio_toolkit")
	}

	providerConfig, err := getProviderConfig()
	if err != nil {
		t.Fatalf("getProviderConfig: %v", err)
	}

	dataConfig, err := ReadAccTestResource([]string{"data_sources", "composio_toolkit", "data", "010_data.tf"})
	if err != nil {
		t.Fatalf("ReadAccTestResource: %v", err)
	}

	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_15_0),
		},
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + dataConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.composio_toolkit.github", "slug", "github"),
					resource.TestCheckResourceAttr("data.composio_toolkit.github", "id", "github"),
					resource.TestCheckResourceAttrSet("data.composio_toolkit.github", "name"),
				),
			},
		},
	})
}
