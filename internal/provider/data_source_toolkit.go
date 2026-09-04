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

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yu-iskw/terraform-provider-composio/internal/composio/api"
)

var (
	_ datasource.DataSource              = &toolkitDataSource{}
	_ datasource.DataSourceWithConfigure = &toolkitDataSource{}
)

func NewToolkitDataSource() datasource.DataSource {
	return &toolkitDataSource{}
}

type toolkitDataSource struct {
	client *api.Client
}

type toolkitDataSourceModel struct {
	Slug          types.String `tfsdk:"slug"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	Logo          types.String `tfsdk:"logo"`
	Type          types.String `tfsdk:"type"`
	NoAuth        types.Bool   `tfsdk:"no_auth"`
	AuthSchemes   types.Set    `tfsdk:"auth_schemes"`
	ToolsCount    types.Int64  `tfsdk:"tools_count"`
	TriggersCount types.Int64  `tfsdk:"triggers_count"`
	Version       types.String `tfsdk:"version"`
	AppURL        types.String `tfsdk:"app_url"`
	ID            types.String `tfsdk:"id"`
}

func (d *toolkitDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_toolkit"
}

func (d *toolkitDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a Composio toolkit by slug. Requires a project API key (`x-api-key`).",
		Attributes: map[string]schema.Attribute{
			"slug": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Toolkit slug, for example `github`.",
				Validators:          []validator.String{ValidateNonEmptyString{}},
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Same as `slug`.",
			},
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Display name.",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Toolkit description.",
			},
			"logo": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Logo URL.",
			},
			"type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "`native` or `custom`.",
			},
			"no_auth": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "True when the toolkit can be used without authentication.",
			},
			"auth_schemes": schema.SetAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Supported authentication schemes.",
			},
			"tools_count": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Number of tools in the toolkit.",
			},
			"triggers_count": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Number of trigger types in the toolkit.",
			},
			"version": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Toolkit version reported by Composio.",
			},
			"app_url": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Upstream application URL, when present.",
			},
		},
	}
}

func (d *toolkitDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *api.Client, got: %T.", req.ProviderData))
		return
	}
	if !client.HasProjectKey() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing Project API Key",
			"data.composio_toolkit requires `api_key` or COMPOSIO_API_KEY.",
		)
		return
	}
	d.client = client
}

func (d *toolkitDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config toolkitDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tk, err := d.client.GetToolkit(ctx, config.Slug.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Composio toolkit", formatAPIError(err))
		return
	}

	state := toolkitDataSourceModel{
		ID:            types.StringValue(tk.Slug),
		Slug:          types.StringValue(tk.Slug),
		Name:          types.StringValue(tk.Name),
		Description:   types.StringValue(tk.Description),
		Logo:          types.StringValue(tk.Logo),
		Type:          types.StringValue(tk.Type),
		NoAuth:        types.BoolValue(tk.NoAuth),
		AuthSchemes:   stringSet(tk.AuthSchemes),
		ToolsCount:    types.Int64Value(int64(tk.ToolsCount)),
		TriggersCount: types.Int64Value(int64(tk.TriggersCount)),
		Version:       types.StringValue(tk.Version),
		AppURL:        types.StringValue(tk.AppURL),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
