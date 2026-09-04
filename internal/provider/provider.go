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

	yourservice "github.com/example/terraform-provider-template/internal/your_service"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure templateProvider satisfies the provider interface.
var _ provider.Provider = &templateProvider{}

// templateProvider defines the provider implementation.
type templateProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// templateProviderModel describes the provider data model.
type templateProviderModel struct {
	Endpoint              types.String  `tfsdk:"endpoint"`
	APIKey                types.String  `tfsdk:"api_key"`
	MaxConcurrentRequests types.Int64   `tfsdk:"max_concurrent_requests"`
	RequestsPerSecond     types.Float64 `tfsdk:"requests_per_second"`
}

func (p *templateProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "template"
	resp.Version = p.version
}

func (p *templateProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A template Terraform provider with example resources and data sources.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "Example API endpoint. Defaults to `https://api.example.com` when omitted.",
				Optional:            true,
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "Example API key used by the template client.",
				Required:            true,
				Sensitive:           true,
			},
			"max_concurrent_requests": schema.Int64Attribute{
				MarkdownDescription: "Maximum number of concurrent HTTP requests to the API. Defaults to 10 when omitted.",
				Optional:            true,
			},
			"requests_per_second": schema.Float64Attribute{
				MarkdownDescription: "Average sustained API requests per second (token bucket). Defaults to 10 when omitted.",
				Optional:            true,
			},
		},
	}
}

func (p *templateProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration
	var config templateProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate configuration
	if config.Endpoint.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("endpoint"),
			"Unknown API Endpoint",
			"The provider cannot configure the API client when `endpoint` is unknown.",
		)
		return
	}
	if config.APIKey.IsUnknown() || config.APIKey.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing API Key",
			"Please set the `api_key` attribute for the template provider.",
		)
		return
	}

	endpoint := ""
	if !config.Endpoint.IsNull() {
		endpoint = config.Endpoint.ValueString()
	}

	opts := &yourservice.Options{}
	if !config.MaxConcurrentRequests.IsNull() && !config.MaxConcurrentRequests.IsUnknown() {
		opts.MaxConcurrent = config.MaxConcurrentRequests.ValueInt64()
	}
	if !config.RequestsPerSecond.IsNull() && !config.RequestsPerSecond.IsUnknown() {
		opts.RPS = config.RequestsPerSecond.ValueFloat64()
	}

	client, err := yourservice.New(endpoint, config.APIKey.ValueString(), opts)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Unable to Configure API Client",
			err.Error(),
		)
		return
	}
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *templateProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewExampleItemResource,
	}
}

func (p *templateProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewExampleItemDataSource,
	}
}

func (p *templateProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &templateProvider{
			version: version,
		}
	}
}
