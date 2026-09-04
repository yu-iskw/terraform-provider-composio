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
	"strings"
	"testing"
)

func TestGetProviderConfig(t *testing.T) {
	cfg, err := getProviderConfig()
	if err != nil {
		t.Fatalf("getProviderConfig: %v", err)
	}
	if !strings.Contains(cfg, `provider "composio"`) {
		t.Fatalf("expected composio provider block, got %q", cfg)
	}
}

func TestReadAccTestResourceToolkit(t *testing.T) {
	cfg, err := ReadAccTestResource([]string{"data_sources", "composio_toolkit", "data", "010_data.tf"})
	if err != nil {
		t.Fatalf("ReadAccTestResource: %v", err)
	}
	if !strings.Contains(cfg, `data "composio_toolkit"`) {
		t.Fatalf("unexpected fixture: %q", cfg)
	}
}

func TestSubstituteAccTestName(t *testing.T) {
	got := substituteAccTestName(`name = "__NAME__"`, "tf-acc-xyz")
	if got != `name = "tf-acc-xyz"` {
		t.Fatalf("got %q", got)
	}
}
