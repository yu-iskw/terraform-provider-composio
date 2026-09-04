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
	"embed"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

//go:embed docs/**/*.md
var embeddedDocs embed.FS

const (
	integrationTestModeEnvVar = "TF_ACC"
)

func isIntegrationTestMode() bool {
	return os.Getenv(integrationTestModeEnvVar) == "1"
}

// readMarkdownDescription reads the content of a markdown file from the embedded filesystem.
// The filename parameter should be in the format "internal/provider/docs/..." or "docs/..."
func readMarkdownDescription(ctx context.Context, filename string) (string, error) {
	// Extract the path relative to the 'internal/provider/' prefix
	relativePathInProvider := strings.TrimPrefix(filename, "internal/provider/")

	// If the path doesn't start with "docs/", the TrimPrefix didn't do anything,
	// meaning the input didn't have the "internal/provider/" prefix.
	// In that case, use the filename as-is (for backward compatibility).
	if !strings.HasPrefix(relativePathInProvider, "docs/") && strings.HasPrefix(filename, "docs/") {
		relativePathInProvider = filename
	}

	tflog.Debug(ctx, fmt.Sprintf("Attempting to read embedded markdown file: %s (original filename: %s)", relativePathInProvider, filename))

	// Read from the embedded filesystem
	content, err := embeddedDocs.ReadFile(relativePathInProvider)
	if err != nil {
		// Log the error with the path that failed
		tflog.Error(ctx, fmt.Sprintf("Error reading embedded markdown file %s: %s", relativePathInProvider, err.Error()))
		return "", fmt.Errorf("failed to read markdown file %s (tried embedded path: %s): %w", filename, relativePathInProvider, err)
	}

	return string(content), nil
}
