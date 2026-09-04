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
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yu-iskw/terraform-provider-composio/internal/composio/api"
	"github.com/yu-iskw/terraform-provider-composio/internal/composio/models"
)

var (
	_ resource.Resource                   = &customToolkitResource{}
	_ resource.ResourceWithConfigure      = &customToolkitResource{}
	_ resource.ResourceWithImportState    = &customToolkitResource{}
	_ resource.ResourceWithValidateConfig = &customToolkitResource{}
)

func NewCustomToolkitResource() resource.Resource {
	return &customToolkitResource{}
}

type customToolkitResource struct {
	client *api.Client
}

type customToolkitResourceModel struct {
	ID         types.String            `tfsdk:"id"`
	Slug       types.String            `tfsdk:"slug"`
	Name       types.String            `tfsdk:"name"`
	AppURL     types.String            `tfsdk:"app_url"`
	AuthScheme *customToolkitAuthModel `tfsdk:"auth_scheme"`
	Type       types.String            `tfsdk:"type"`
	ToolsCount types.Int64             `tfsdk:"tools_count"`
	Version    types.String            `tfsdk:"version"`
	Logo       types.String            `tfsdk:"logo"`
}

type customToolkitAuthModel struct {
	Mode         types.String `tfsdk:"mode"`
	Headers      types.Map    `tfsdk:"headers"`
	DiscoveryURL types.String `tfsdk:"discovery_url"`
}

func (r *customToolkitResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom_toolkit"
}

func (r *customToolkitResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Registers a Composio Custom MCP toolkit (experimental). Maps to `POST /api/v3.1/custom/toolkits/upsert` and `DELETE /api/v3.1/custom/toolkits/{slug}`. " +
			"Composio prefixes the slug with `CUSTOM_` (for example `ACME` becomes `CUSTOM_ACME`). " +
			"`app_url` and `auth_scheme` cannot change in place; Terraform forces replacement. " +
			"Deleting the toolkit also removes its auth configs and connected accounts. " +
			"Sessions and MCP URLs remain application runtime concerns — this resource only registers the remote MCP server.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Normalized toolkit slug returned by Composio (includes the `CUSTOM_` prefix).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"slug": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Project-local slug without the `CUSTOM_` prefix (for example `ACME`). Changing this forces replacement.",
				Validators:          []validator.String{ValidateNonEmptyString{}},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Human-readable toolkit name.",
				Validators:          []validator.String{ValidateNonEmptyString{}},
			},
			"app_url": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Public HTTPS URL of the remote MCP server. Immutable after create; changing forces replacement.",
				Validators:          []validator.String{ValidateNonEmptyString{}},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"auth_scheme": schema.SingleNestedAttribute{
				Required:            true,
				MarkdownDescription: "Authentication scheme for the remote MCP server. Immutable after create; changing forces replacement.",
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"mode": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "One of `NO_AUTH`, `API_KEY`, or `DCR_OAUTH`.",
						Validators:          []validator.String{ValidateNonEmptyString{}},
					},
					"headers": schema.MapAttribute{
						Optional:            true,
						ElementType:         types.StringType,
						MarkdownDescription: "Required for `API_KEY`. At least one header value must contain `{{generic_api_key}}`.",
					},
					"discovery_url": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Required for `DCR_OAUTH`. OAuth authorization-server discovery URL.",
					},
				},
			},
			"type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Toolkit provenance from Composio (`custom` for Custom MCP).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tools_count": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Number of tools Composio reports after sync. Initial sync is automatic; this may be zero briefly after create.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"version": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Toolkit version reported by Composio.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"logo": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Logo URL reported by Composio.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *customToolkitResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config customToolkitResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() || config.AuthScheme == nil {
		return
	}
	mode := strings.TrimSpace(config.AuthScheme.Mode.ValueString())
	if config.AuthScheme.Mode.IsUnknown() || config.AuthScheme.Mode.IsNull() {
		return
	}
	switch mode {
	case api.CustomAuthModeNoAuth:
		if !config.AuthScheme.Headers.IsNull() && !config.AuthScheme.Headers.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				path.Root("auth_scheme").AtName("headers"),
				"Invalid Auth Scheme",
				"`headers` is only valid when mode is `API_KEY`.",
			)
		}
		if !config.AuthScheme.DiscoveryURL.IsNull() && !config.AuthScheme.DiscoveryURL.IsUnknown() && strings.TrimSpace(config.AuthScheme.DiscoveryURL.ValueString()) != "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("auth_scheme").AtName("discovery_url"),
				"Invalid Auth Scheme",
				"`discovery_url` is only valid when mode is `DCR_OAUTH`.",
			)
		}
	case api.CustomAuthModeAPIKey:
		if config.AuthScheme.Headers.IsNull() || config.AuthScheme.Headers.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				path.Root("auth_scheme").AtName("headers"),
				"Missing Headers",
				"`headers` is required when mode is `API_KEY`.",
			)
			return
		}
		headers := map[string]string{}
		resp.Diagnostics.Append(config.AuthScheme.Headers.ElementsAs(ctx, &headers, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		if !headersContainAPIKeyPlaceholder(headers) {
			resp.Diagnostics.AddAttributeError(
				path.Root("auth_scheme").AtName("headers"),
				"Invalid Headers",
				fmt.Sprintf("At least one header value must contain %q.", api.CustomGenericPlaceholder),
			)
		}
		if !config.AuthScheme.DiscoveryURL.IsNull() && !config.AuthScheme.DiscoveryURL.IsUnknown() && strings.TrimSpace(config.AuthScheme.DiscoveryURL.ValueString()) != "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("auth_scheme").AtName("discovery_url"),
				"Invalid Auth Scheme",
				"`discovery_url` is only valid when mode is `DCR_OAUTH`.",
			)
		}
	case api.CustomAuthModeDCROAuth:
		if config.AuthScheme.DiscoveryURL.IsNull() || config.AuthScheme.DiscoveryURL.IsUnknown() || strings.TrimSpace(config.AuthScheme.DiscoveryURL.ValueString()) == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("auth_scheme").AtName("discovery_url"),
				"Missing Discovery URL",
				"`discovery_url` is required when mode is `DCR_OAUTH`.",
			)
		}
		if !config.AuthScheme.Headers.IsNull() && !config.AuthScheme.Headers.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				path.Root("auth_scheme").AtName("headers"),
				"Invalid Auth Scheme",
				"`headers` is only valid when mode is `API_KEY`.",
			)
		}
	default:
		resp.Diagnostics.AddAttributeError(
			path.Root("auth_scheme").AtName("mode"),
			"Invalid Auth Mode",
			"mode must be one of `NO_AUTH`, `API_KEY`, or `DCR_OAUTH`.",
		)
	}
}

func (r *customToolkitResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := configureProjectClient(req.ProviderData, "composio_custom_toolkit")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() || client == nil {
		return
	}
	r.client = client
}

func (r *customToolkitResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan customToolkitResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in, diags := upsertInputFromModel(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := r.client.UpsertCustomToolkit(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Composio custom toolkit", conflictHint(err))
		return
	}
	plan.ID = types.StringValue(out.Slug)
	if err := r.loadAfterWrite(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("Unable to read Composio custom toolkit after create", err.Error())
		if _, delErr := r.client.DeleteCustomToolkit(ctx, out.Slug); delErr != nil {
			resp.Diagnostics.AddError("Unable to roll back Composio custom toolkit after a read failure", delErr.Error())
		}
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *customToolkitResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state customToolkitResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	remote, err := r.client.GetToolkit(ctx, state.ID.ValueString())
	if api.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Composio custom toolkit", err.Error())
		return
	}
	state.applyRemote(remote)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *customToolkitResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan customToolkitResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in, diags := upsertInputFromModel(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := r.client.UpsertCustomToolkit(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update Composio custom toolkit", conflictHint(err))
		return
	}
	if out.Slug != "" {
		plan.ID = types.StringValue(out.Slug)
	}
	if err := r.loadAfterWrite(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("Unable to read Composio custom toolkit after update", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *customToolkitResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state customToolkitResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if _, err := r.client.DeleteCustomToolkit(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Unable to delete Composio custom toolkit", err.Error())
	}
}

func (r *customToolkitResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *customToolkitResource) loadAfterWrite(ctx context.Context, dest *customToolkitResourceModel) error {
	remote, err := r.client.GetToolkit(ctx, dest.ID.ValueString())
	if err != nil {
		return err
	}
	dest.applyRemote(remote)
	return nil
}

func upsertInputFromModel(ctx context.Context, plan customToolkitResourceModel) (api.UpsertCustomToolkitInput, diag.Diagnostics) {
	var diags diag.Diagnostics
	in := api.UpsertCustomToolkitInput{
		Slug:   plan.Slug.ValueString(),
		Name:   plan.Name.ValueString(),
		AppURL: plan.AppURL.ValueString(),
	}
	if plan.AuthScheme == nil {
		diags.AddError("Missing Auth Scheme", "`auth_scheme` is required.")
		return in, diags
	}
	scheme := api.CustomAuthScheme{Mode: plan.AuthScheme.Mode.ValueString()}
	if !plan.AuthScheme.Headers.IsNull() && !plan.AuthScheme.Headers.IsUnknown() {
		headers := map[string]string{}
		diags.Append(plan.AuthScheme.Headers.ElementsAs(ctx, &headers, false)...)
		scheme.Headers = headers
	}
	if !plan.AuthScheme.DiscoveryURL.IsNull() && !plan.AuthScheme.DiscoveryURL.IsUnknown() {
		scheme.DiscoveryURL = plan.AuthScheme.DiscoveryURL.ValueString()
	}
	in.AuthSchemes = []api.CustomAuthScheme{scheme}
	return in, diags
}

func (m *customToolkitResourceModel) applyRemote(remote models.Toolkit) {
	if remote.Slug != "" {
		m.ID = types.StringValue(remote.Slug)
	}
	m.Name = types.StringValue(remote.Name)
	if remote.AppURL != "" {
		m.AppURL = types.StringValue(remote.AppURL)
	}
	m.Type = types.StringValue(remote.Type)
	m.ToolsCount = types.Int64Value(int64(remote.ToolsCount))
	m.Version = types.StringValue(remote.Version)
	m.Logo = types.StringValue(remote.Logo)
	// auth_scheme and input slug are configuration; keep prior state (GET does not round-trip headers).
	if m.Slug.IsNull() || m.Slug.IsUnknown() || m.Slug.ValueString() == "" {
		m.Slug = types.StringValue(stripCustomPrefix(remote.Slug))
	}
}

func stripCustomPrefix(slug string) string {
	const prefix = "CUSTOM_"
	if strings.HasPrefix(strings.ToUpper(slug), prefix) {
		return slug[len(prefix):]
	}
	return slug
}

func headersContainAPIKeyPlaceholder(headers map[string]string) bool {
	for _, v := range headers {
		if strings.Contains(v, api.CustomGenericPlaceholder) {
			return true
		}
	}
	return false
}

func conflictHint(err error) string {
	if api.IsConflict(err) {
		return err.Error() + ". app_url and auth_scheme are immutable; delete the toolkit and recreate it (Terraform replace) to change them."
	}
	return err.Error()
}
