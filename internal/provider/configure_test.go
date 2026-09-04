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
	"testing"

	"github.com/yu-iskw/terraform-provider-composio/internal/composio/api"
)

func TestConfigureProjectClientRequiresProjectKey(t *testing.T) {
	c, err := api.New(api.Options{OrgAPIKey: "org-only", Endpoint: "https://backend.composio.dev"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	client, diags := configureProjectClient(c, "composio_auth_config")
	if client != nil {
		t.Fatal("expected nil client")
	}
	if !diags.HasError() {
		t.Fatal("expected missing project key")
	}
}

func TestConfigureProjectClientNilProviderData(t *testing.T) {
	client, diags := configureProjectClient(nil, "composio_auth_config")
	if client != nil || diags.HasError() {
		t.Fatal("nil provider data must be a no-op")
	}
}
