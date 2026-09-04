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

package models

const (
	AuthConfigStatusEnabled  = "ENABLED"
	AuthConfigStatusDisabled = "DISABLED"

	AuthConfigCreateManaged = "use_composio_managed_auth"
	AuthConfigCreateCustom  = "use_custom_auth"

	AuthConfigUpdateDefault = "default"
	AuthConfigUpdateCustom  = "custom"
)

// AuthConfig is a durable Composio authentication configuration.
type AuthConfig struct {
	ID                       string
	Name                     string
	ToolkitSlug              string
	AuthScheme               string
	IsComposioManaged        bool
	Status                   string
	RestrictToFollowingTools []string
	CreatedAt                string
	LastUpdatedAt            string
	NoOfConnections          int
	Type                     string
}

func (a AuthConfig) Enabled() bool {
	return a.Status != AuthConfigStatusDisabled
}
