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

	"github.com/hashicorp/terraform-plugin-framework/provider"
)

func TestProviderSchemaUsesComposioConfiguration(t *testing.T) {
	p := New("test")()

	var resp provider.SchemaResponse
	p.Schema(context.Background(), provider.SchemaRequest{}, &resp)

	for _, name := range []string{"endpoint", "api_key", "org_api_key", "max_concurrent_requests", "request_timeout"} {
		if _, ok := resp.Schema.Attributes[name]; !ok {
			t.Fatalf("expected provider schema to include %s", name)
		}
	}
	if _, ok := resp.Schema.Attributes["requests_per_second"]; ok {
		t.Fatal("did not expect requests_per_second")
	}
	if _, ok := resp.Schema.Attributes["token"]; ok {
		t.Fatal("did not expect token")
	}
}

func TestProviderRegistersAuthConfigAndToolkit(t *testing.T) {
	p := New("test")()
	if got := len(p.Resources(context.Background())); got != 2 {
		t.Fatalf("resources = %d", got)
	}
	if got := len(p.DataSources(context.Background())); got != 1 {
		t.Fatalf("data sources = %d", got)
	}
}
