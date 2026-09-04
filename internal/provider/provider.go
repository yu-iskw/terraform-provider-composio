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
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yu-iskw/terraform-provider-composio/internal/composio/api"
)

var _ provider.Provider = &composioProvider{}

type composioProvider struct {
	version string
}

type composioProviderModel struct {
	Endpoint              types.String `tfsdk:"endpoint"`
	APIKey                types.String `tfsdk:"api_key"`
	OrgAPIKey             types.String `tfsdk:"org_api_key"`
	MaxConcurrentRequests types.Int64  `tfsdk:"max_concurrent_requests"`
	RequestTimeout        types.String `tfsdk:"request_timeout"`
}

func (p *composioProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "composio"
	resp.Version = p.version
}

func (p *composioProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage durable Composio control-plane configuration. Terraform configures projects, auth configs, and related objects. Applications authorize users and execute tools.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "Composio API origin. Defaults to `https://backend.composio.dev`. Override only for tests or a proxy. The provider always calls `/api/v3.1` under this origin.",
				Optional:            true,
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "Project API key sent as `x-api-key`. Falls back to `COMPOSIO_API_KEY`. Required by project-scoped resources and data sources.",
				Optional:            true,
				Sensitive:           true,
			},
			"org_api_key": schema.StringAttribute{
				MarkdownDescription: "Organization API key sent as `x-org-api-key`. Falls back to `COMPOSIO_ORG_API_KEY`. Required by organization-scoped resources such as projects.",
				Optional:            true,
				Sensitive:           true,
			},
			"max_concurrent_requests": schema.Int64Attribute{
				MarkdownDescription: "Maximum in-flight HTTP requests. Defaults to 8.",
				Optional:            true,
			},
			"request_timeout": schema.StringAttribute{
				MarkdownDescription: "Per-request timeout as a Go duration, for example `30s`. Defaults to 30s.",
				Optional:            true,
			},
		},
	}
}

func (p *composioProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config composioProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.Endpoint.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("endpoint"), "Unknown API Endpoint", "The provider cannot configure the client when `endpoint` is unknown.")
		return
	}
	if config.APIKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("api_key"), "Unknown API Key", "The provider cannot configure the client when `api_key` is unknown.")
		return
	}
	if config.OrgAPIKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("org_api_key"), "Unknown Organization API Key", "The provider cannot configure the client when `org_api_key` is unknown.")
		return
	}
	if config.MaxConcurrentRequests.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("max_concurrent_requests"), "Unknown Concurrency Limit", "The provider cannot configure the client when `max_concurrent_requests` is unknown.")
		return
	}
	if config.RequestTimeout.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("request_timeout"), "Unknown Request Timeout", "The provider cannot configure the client when `request_timeout` is unknown.")
		return
	}

	endpoint := os.Getenv("COMPOSIO_ENDPOINT")
	if !config.Endpoint.IsNull() {
		endpoint = config.Endpoint.ValueString()
	}

	apiKey := os.Getenv("COMPOSIO_API_KEY")
	if !config.APIKey.IsNull() {
		apiKey = config.APIKey.ValueString()
	}

	orgKey := os.Getenv("COMPOSIO_ORG_API_KEY")
	if !config.OrgAPIKey.IsNull() {
		orgKey = config.OrgAPIKey.ValueString()
	}

	opts := api.Options{
		Endpoint:      endpoint,
		ProjectAPIKey: apiKey,
		OrgAPIKey:     orgKey,
		UserAgent:     fmt.Sprintf("terraform-provider-composio/%s terraform-plugin-framework", p.version),
	}
	if !config.MaxConcurrentRequests.IsNull() {
		opts.MaxConcurrent = config.MaxConcurrentRequests.ValueInt64()
	}
	if !config.RequestTimeout.IsNull() && config.RequestTimeout.ValueString() != "" {
		d, err := time.ParseDuration(config.RequestTimeout.ValueString())
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("request_timeout"),
				"Invalid Request Timeout",
				fmt.Sprintf("Could not parse %q as a duration: %s", config.RequestTimeout.ValueString(), err),
			)
			return
		}
		opts.Timeout = d
	}

	client, err := api.New(opts)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Configure Composio Client", err.Error())
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *composioProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAuthConfigResource,
	}
}

func (p *composioProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewToolkitDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &composioProvider{version: version}
	}
}
