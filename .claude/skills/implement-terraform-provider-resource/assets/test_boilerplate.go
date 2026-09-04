// Copyright 2026 yu-iskw
// Boilerplate: copy patterns into internal/provider when adding acceptance tests.
package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestAccExampleItem_basic(t *testing.T) {
	if !isIntegrationTestMode() {
		t.Skip("Skipping acceptance test unless TF_ACC=1")
	}

	providerConfig, err := getProviderConfig()
	if err != nil {
		t.Fatalf("getProviderConfig: %v", err)
	}

	name := acctest.RandomWithPrefix("tf-acc")
	createConfig, err := ReadAccTestResource([]string{"resources", "composio_auth_config", "lifecycle", "010_create.tf"})
	if err != nil {
		t.Fatalf("ReadAccTestResource: %v", err)
	}
	createConfig = substituteAccTestName(createConfig, name)

	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_15_0),
		},
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + createConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("composio_auth_config.test", "toolkit_slug", "github"),
					resource.TestCheckResourceAttr("composio_auth_config.test", "name", name),
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
