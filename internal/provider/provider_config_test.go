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

func TestProviderSchemaUsesGenericConfiguration(t *testing.T) {
	p := New("test")()

	var resp provider.SchemaResponse
	p.Schema(context.Background(), provider.SchemaRequest{}, &resp)

	if _, ok := resp.Schema.Attributes["endpoint"]; !ok {
		t.Fatal("expected provider schema to include endpoint")
	}
	if _, ok := resp.Schema.Attributes["api_key"]; !ok {
		t.Fatal("expected provider schema to include api_key")
	}
	if _, ok := resp.Schema.Attributes["max_concurrent_requests"]; !ok {
		t.Fatal("expected provider schema to include max_concurrent_requests")
	}
	if _, ok := resp.Schema.Attributes["requests_per_second"]; !ok {
		t.Fatal("expected provider schema to include requests_per_second")
	}
	if _, ok := resp.Schema.Attributes["host"]; ok {
		t.Fatal("did not expect provider schema to include host")
	}
	if _, ok := resp.Schema.Attributes["token"]; ok {
		t.Fatal("did not expect provider schema to include token")
	}
}
