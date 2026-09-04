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
)

func TestIsIntegrationTestMode(t *testing.T) {
	original_value := os.Getenv(integrationTestModeEnvVar)

	t.Setenv(integrationTestModeEnvVar, "1")
	if !isIntegrationTestMode() {
		t.Errorf("Expected: true, Got: %t", isIntegrationTestMode())
	}
	t.Setenv(integrationTestModeEnvVar, "0")
	if isIntegrationTestMode() {
		t.Errorf("Expected: false, Got: %t", isIntegrationTestMode())
	}

	t.Setenv(integrationTestModeEnvVar, original_value)
}

func TestReadMarkdownDescriptionEmbedded(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		filename string
		wantErr  bool
	}{
		{
			name:     "resource with internal/provider prefix",
			filename: "internal/provider/docs/resources/example_item.md",
			wantErr:  false,
		},
		{
			name:     "resource with docs prefix",
			filename: "docs/resources/example_item.md",
			wantErr:  false,
		},
		{
			name:     "data source",
			filename: "internal/provider/docs/data_sources/example_item.md",
			wantErr:  false,
		},
		{
			name:     "non-existent file",
			filename: "internal/provider/docs/resources/nonexistent.md",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := readMarkdownDescription(ctx, tt.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("readMarkdownDescription() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(content) == 0 {
				t.Errorf("readMarkdownDescription() returned empty content for %s", tt.filename)
			}
		})
	}
}
