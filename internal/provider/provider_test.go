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
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

const integrationTestModeEnvVar = "TF_ACC"

func isIntegrationTestMode() bool {
	return os.Getenv(integrationTestModeEnvVar) == "1"
}

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"composio": providerserver.NewProtocol6WithError(New("test")()),
}

func TestAccProviderFactories(t *testing.T) {
	if _, ok := testAccProtoV6ProviderFactories["composio"]; !ok {
		t.Fatal("expected composio provider factory")
	}
}

func testAccPreCheck(t *testing.T) {
	t.Helper()
	if !isIntegrationTestMode() {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' is set to '1'")
	}
	if os.Getenv("COMPOSIO_API_KEY") == "" {
		t.Fatal("COMPOSIO_API_KEY must be set for acceptance tests")
	}
}

func TestIsIntegrationTestMode(t *testing.T) {
	t.Setenv(integrationTestModeEnvVar, "1")
	if !isIntegrationTestMode() {
		t.Fatalf("expected true, got %t", isIntegrationTestMode())
	}
	t.Setenv(integrationTestModeEnvVar, "0")
	if isIntegrationTestMode() {
		t.Fatalf("expected false, got %t", isIntegrationTestMode())
	}
}

func TestProviderMetadata(t *testing.T) {
	p := New("test")()

	var resp provider.MetadataResponse
	p.Metadata(context.Background(), provider.MetadataRequest{}, &resp)

	if resp.TypeName != "composio" {
		t.Fatalf("expected provider type name composio, got %q", resp.TypeName)
	}

	if resp.Version != "test" {
		t.Fatalf("expected provider version test, got %q", resp.Version)
	}
}
