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
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yu-iskw/terraform-provider-composio/internal/composio/api"
)

func configureProjectClient(providerData any, name string) (*api.Client, diag.Diagnostics) {
	var diags diag.Diagnostics
	if providerData == nil {
		return nil, diags
	}
	client, ok := providerData.(*api.Client)
	if !ok {
		diags.AddError("Unexpected Configure Type", fmt.Sprintf("Expected *api.Client, got: %T.", providerData))
		return nil, diags
	}
	if !client.HasProjectKey() {
		diags.AddError("Missing Project API Key", name+" requires `api_key` or COMPOSIO_API_KEY.")
		return nil, diags
	}
	return client, diags
}

func stringSet(values []string) types.Set {
	if values == nil {
		values = []string{}
	}
	elems := make([]attr.Value, 0, len(values))
	for _, v := range values {
		elems = append(elems, types.StringValue(v))
	}
	return types.SetValueMust(types.StringType, elems)
}
