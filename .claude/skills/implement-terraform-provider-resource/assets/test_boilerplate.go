// Copyright 2026 yu-iskw
// Boilerplate: copy patterns into internal/provider when adding acceptance tests.
package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const testAccExampleItemBasic = `
provider "composio" {
  api_key = "acceptance-test-key"
}

resource "composio_auth_config" "test" {
  toolkit_slug = "github"
  managed_auth = {}
}
`

func TestAccExampleItem_basic(t *testing.T) {
	if !isIntegrationTestMode() {
		t.Skip("Skipping acceptance test unless TF_ACC=1")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccExampleItemBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("composio_auth_config.test", "toolkit_slug", "github"),
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
