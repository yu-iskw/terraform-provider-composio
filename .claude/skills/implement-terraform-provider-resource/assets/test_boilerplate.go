// Copyright 2026 yu-iskw
// Boilerplate: copy patterns into internal/provider when adding acceptance tests.
package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const testAccExampleItemBasic = `
provider "template" {
  endpoint = "https://api.example.com"
  api_key  = "acceptance-test-key"
}

resource "template_example_item" "test" {
  name        = "acceptance-example"
  description = "from acceptance test boilerplate"
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
					resource.TestCheckResourceAttr("template_example_item.test", "name", "acceptance-example"),
				),
			},
			{
				ResourceName:      "template_example_item.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
