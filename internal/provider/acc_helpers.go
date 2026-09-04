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
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

const accTestNamePlaceholder = "__NAME__"

func getPathToAccTests() (string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("failed to get current file path")
	}

	accTestsPath := path.Join(path.Dir(filename), "acc_tests")
	if _, err := os.Stat(accTestsPath); os.IsNotExist(err) {
		return "", fmt.Errorf("acc_tests directory does not exist at %s", accTestsPath)
	}
	return accTestsPath, nil
}

func getPathToAccTestResource(elements []string) (string, error) {
	pathToAccTests, err := getPathToAccTests()
	if err != nil {
		return "", err
	}

	allElements := append([]string{pathToAccTests}, elements...)
	accTestResourcePath := path.Join(allElements...)

	cleanedAccTestsPath := path.Clean(pathToAccTests)
	cleanedResourcePath := path.Clean(accTestResourcePath)
	if !strings.HasPrefix(cleanedResourcePath, cleanedAccTestsPath) {
		return "", fmt.Errorf("attempted to access file outside acc_tests directory: %s", accTestResourcePath)
	}

	if _, err := os.Stat(accTestResourcePath); os.IsNotExist(err) {
		return "", fmt.Errorf("acc_tests resource does not exist at %s", accTestResourcePath)
	}
	return accTestResourcePath, nil
}

// ReadAccTestResource reads a .tf fixture under internal/provider/acc_tests.
func ReadAccTestResource(elements []string) (string, error) {
	p, err := getPathToAccTestResource(elements)
	if err != nil {
		return "", err
	}
	resource, err := os.ReadFile(filepath.Clean(p))
	if err != nil {
		return "", err
	}
	return string(resource), nil
}

// getProviderConfig returns the Acc provider block. Credentials come from
// COMPOSIO_API_KEY (and related env vars) via provider Configure — never from HCL.
func getProviderConfig() (string, error) {
	return ReadAccTestResource([]string{"provider.tf"})
}

func substituteAccTestName(config, name string) string {
	return strings.ReplaceAll(config, accTestNamePlaceholder, name)
}
