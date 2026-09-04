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
