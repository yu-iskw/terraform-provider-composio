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
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yu-iskw/terraform-provider-composio/internal/composio/api"
	"github.com/yu-iskw/terraform-provider-composio/internal/composio/models"
)

var (
	_ resource.Resource                   = &authConfigResource{}
	_ resource.ResourceWithConfigure      = &authConfigResource{}
	_ resource.ResourceWithImportState    = &authConfigResource{}
	_ resource.ResourceWithModifyPlan     = &authConfigResource{}
	_ resource.ResourceWithValidateConfig = &authConfigResource{}
)

func NewAuthConfigResource() resource.Resource {
	return &authConfigResource{}
}

type authConfigResource struct {
	client *api.Client
}

type authConfigResourceModel struct {
	ID                types.String      `tfsdk:"id"`
	ToolkitSlug       types.String      `tfsdk:"toolkit_slug"`
	Name              types.String      `tfsdk:"name"`
	Enabled           types.Bool        `tfsdk:"enabled"`
	ManagedAuth       *managedAuthModel `tfsdk:"managed_auth"`
	CustomAuth        *customAuthModel  `tfsdk:"custom_auth"`
	AuthScheme        types.String      `tfsdk:"auth_scheme"`
	IsComposioManaged types.Bool        `tfsdk:"is_composio_managed"`
	Status            types.String      `tfsdk:"status"`
	CreatedAt         types.String      `tfsdk:"created_at"`
}

type managedAuthModel struct {
	RestrictToFollowingTools types.Set `tfsdk:"restrict_to_following_tools"`
	Scopes                   types.Set `tfsdk:"scopes"`
}

type customAuthModel struct {
	AuthScheme               types.String `tfsdk:"auth_scheme"`
	Credentials              types.Map    `tfsdk:"credentials"`
	RestrictToFollowingTools types.Set    `tfsdk:"restrict_to_following_tools"`
}

func (r *authConfigResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_auth_config"
}

func (r *authConfigResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Composio auth config. Requires a project API key (`x-api-key`). Import uses the auth config id. Changing `toolkit_slug` forces replacement. `Sensitive` hides values in UI output. It does not keep them out of Terraform state except for write-only `custom_auth.credentials`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Composio auth config id.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"toolkit_slug": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Toolkit slug, for example `github`. Cannot be changed in place.",
				Validators:          []validator.String{ValidateNonEmptyString{}},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Display name. The API may assign a default when omitted.",
			},
			"enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "When false, the provider disables the auth config when Composio still reports it enabled. Disabled configs cannot start new connections.",
			},
			"managed_auth": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Use Composio-managed authentication. Exactly one of `managed_auth` or `custom_auth` must be set. Switching to `custom_auth` forces replacement.",
				Attributes: map[string]schema.Attribute{
					"restrict_to_following_tools": schema.SetAttribute{
						Optional:            true,
						ElementType:         types.StringType,
						MarkdownDescription: "Tool slugs this auth config may use. Order is not significant.",
					},
					"scopes": schema.SetAttribute{
						Optional:            true,
						ElementType:         types.StringType,
						MarkdownDescription: "OAuth scopes for managed auth. Order is not significant.",
					},
				},
			},
			"custom_auth": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Bring your own credentials. Exactly one of `managed_auth` or `custom_auth` must be set. Switching to `managed_auth` forces replacement.",
				Attributes: map[string]schema.Attribute{
					"auth_scheme": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "Authentication scheme, for example `OAUTH2` or `API_KEY`.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"credentials": schema.MapAttribute{
						Optional:            true,
						Sensitive:           true,
						WriteOnly:           true,
						ElementType:         types.StringType,
						MarkdownDescription: "Write-only credential map. Never stored in state. Sent on create when set. Sent on update only when Terraform also plans another change. A credentials-only edit does not produce a plan. Terraform 1.11 or later is required.",
					},
					"restrict_to_following_tools": schema.SetAttribute{
						Optional:            true,
						ElementType:         types.StringType,
						MarkdownDescription: "Tool slugs this auth config may use. Order is not significant.",
					},
				},
			},
			"auth_scheme": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Scheme reported by Composio.",
			},
			"is_composio_managed": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "True when Composio manages the credentials.",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "`ENABLED` or `DISABLED`.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Creation timestamp from Composio.",
			},
		},
	}
}

func (r *authConfigResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var cfg authConfigResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	hasManaged := cfg.ManagedAuth != nil
	hasCustom := cfg.CustomAuth != nil
	if hasManaged == hasCustom {
		resp.Diagnostics.AddError(
			"Invalid Auth Config Mode",
			"Set exactly one of `managed_auth` or `custom_auth`.",
		)
	}
}

func (r *authConfigResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *api.Client, got: %T.", req.ProviderData))
		return
	}
	if !client.HasProjectKey() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing Project API Key",
			"composio_auth_config requires `api_key` or COMPOSIO_API_KEY.",
		)
		return
	}
	r.client = client
}

func (r *authConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan authConfigResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var config authConfigResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in, diags := createInputFromModels(ctx, plan, config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.client.CreateAuthConfig(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Composio auth config", formatAPIError(err))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(id))...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.loadAfterWrite(ctx, id, plan.Enabled.ValueBool(), &plan); err != nil {
		resp.Diagnostics.AddError("Unable to read Composio auth config after create", formatAPIError(err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *authConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state authConfigResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	remote, err := r.client.GetAuthConfig(ctx, state.ID.ValueString())
	if api.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Composio auth config", formatAPIError(err))
		return
	}
	state.applyRemote(remote)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *authConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan authConfigResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var config authConfigResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in, diags := updateInputFromModels(ctx, plan, config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := plan.ID.ValueString()
	if err := r.client.UpdateAuthConfig(ctx, id, in); err != nil {
		resp.Diagnostics.AddError("Unable to update Composio auth config", formatAPIError(err))
		return
	}
	if err := r.loadAfterWrite(ctx, id, plan.Enabled.ValueBool(), &plan); err != nil {
		resp.Diagnostics.AddError("Unable to read Composio auth config after update", formatAPIError(err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *authConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state authConfigResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteAuthConfig(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Unable to delete Composio auth config", formatAPIError(err))
	}
}

func (r *authConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *authConfigResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}
	var state authConfigResourceModel
	var plan authConfigResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if (state.ManagedAuth != nil) != (plan.ManagedAuth != nil) {
		resp.RequiresReplace = append(resp.RequiresReplace, path.Root("managed_auth"), path.Root("custom_auth"))
	}
}

func (r *authConfigResource) reconcileStatus(ctx context.Context, id string, enabled bool) error {
	status := models.AuthConfigStatusEnabled
	if !enabled {
		status = models.AuthConfigStatusDisabled
	}
	return r.client.SetAuthConfigStatus(ctx, id, status)
}

func (r *authConfigResource) loadAfterWrite(ctx context.Context, id string, enabled bool, dest *authConfigResourceModel) error {
	remote, err := r.client.GetAuthConfig(ctx, id)
	if err != nil {
		return err
	}
	if remote.Enabled() != enabled {
		if err := r.reconcileStatus(ctx, id, enabled); err != nil {
			return err
		}
		remote, err = r.client.GetAuthConfig(ctx, id)
		if err != nil {
			return err
		}
	}
	dest.applyRemote(remote)
	return nil
}

func createInputFromModels(ctx context.Context, plan, config authConfigResourceModel) (api.CreateAuthConfigInput, diag.Diagnostics) {
	var diags diag.Diagnostics
	in := api.CreateAuthConfigInput{
		ToolkitSlug: plan.ToolkitSlug.ValueString(),
		Name:        plan.Name.ValueString(),
	}
	switch {
	case plan.ManagedAuth != nil:
		in.Managed = true
		in.RestrictToFollowingTools, diags = setToStrings(ctx, plan.ManagedAuth.RestrictToFollowingTools)
		if diags.HasError() {
			return in, diags
		}
		in.Scopes, diags = setToStrings(ctx, plan.ManagedAuth.Scopes)
	case plan.CustomAuth != nil:
		in.Managed = false
		in.AuthScheme = plan.CustomAuth.AuthScheme.ValueString()
		in.RestrictToFollowingTools, diags = setToStrings(ctx, plan.CustomAuth.RestrictToFollowingTools)
		if diags.HasError() {
			return in, diags
		}
		creds, d := credentialsFromConfig(ctx, config.CustomAuth)
		diags.Append(d...)
		in.Credentials = creds
	}
	return in, diags
}

func updateInputFromModels(ctx context.Context, plan, config authConfigResourceModel) (api.UpdateAuthConfigInput, diag.Diagnostics) {
	var diags diag.Diagnostics
	in := api.UpdateAuthConfigInput{Managed: plan.ManagedAuth != nil}
	if !plan.Name.IsNull() && !plan.Name.IsUnknown() {
		name := plan.Name.ValueString()
		in.Name = &name
	}
	if plan.ManagedAuth != nil {
		tools, d := setToStrings(ctx, plan.ManagedAuth.RestrictToFollowingTools)
		diags.Append(d...)
		in.RestrictToFollowingTools = &tools
		scopes, d := setToStrings(ctx, plan.ManagedAuth.Scopes)
		diags.Append(d...)
		if len(scopes) > 0 {
			joined := strings.Join(scopes, ",")
			in.Scopes = &joined
		}
	}
	if plan.CustomAuth != nil {
		tools, d := setToStrings(ctx, plan.CustomAuth.RestrictToFollowingTools)
		diags.Append(d...)
		in.RestrictToFollowingTools = &tools
		creds, d := credentialsFromConfig(ctx, config.CustomAuth)
		diags.Append(d...)
		if len(creds) > 0 {
			in.Credentials = creds
		}
	}
	return in, diags
}

func credentialsFromConfig(ctx context.Context, custom *customAuthModel) (map[string]string, diag.Diagnostics) {
	var diags diag.Diagnostics
	if custom == nil || custom.Credentials.IsNull() || custom.Credentials.IsUnknown() {
		return nil, diags
	}
	out := map[string]string{}
	diags.Append(custom.Credentials.ElementsAs(ctx, &out, false)...)
	return out, diags
}

func setToStrings(ctx context.Context, s types.Set) ([]string, diag.Diagnostics) {
	var diags diag.Diagnostics
	if s.IsNull() || s.IsUnknown() {
		return nil, diags
	}
	var out []string
	diags.Append(s.ElementsAs(ctx, &out, false)...)
	return out, diags
}

func (m *authConfigResourceModel) applyRemote(remote models.AuthConfig) {
	m.ID = types.StringValue(remote.ID)
	m.ToolkitSlug = types.StringValue(remote.ToolkitSlug)
	m.Name = types.StringValue(remote.Name)
	m.Enabled = types.BoolValue(remote.Enabled())
	m.AuthScheme = types.StringValue(remote.AuthScheme)
	m.IsComposioManaged = types.BoolValue(remote.IsComposioManaged)
	m.Status = types.StringValue(remote.Status)
	m.CreatedAt = types.StringValue(remote.CreatedAt)

	priorTools := types.SetNull(types.StringType)
	if m.ManagedAuth != nil {
		priorTools = m.ManagedAuth.RestrictToFollowingTools
	} else if m.CustomAuth != nil {
		priorTools = m.CustomAuth.RestrictToFollowingTools
	}
	tools := optionalStringSet(remote.RestrictToFollowingTools, priorTools)
	if remote.IsComposioManaged {
		scopes := types.SetNull(types.StringType)
		if m.ManagedAuth != nil {
			scopes = m.ManagedAuth.Scopes
		}
		m.ManagedAuth = &managedAuthModel{
			RestrictToFollowingTools: tools,
			Scopes:                   scopes,
		}
		m.CustomAuth = nil
	} else {
		scheme := types.StringValue(remote.AuthScheme)
		if m.CustomAuth != nil && !m.CustomAuth.AuthScheme.IsNull() {
			scheme = m.CustomAuth.AuthScheme
		}
		m.CustomAuth = &customAuthModel{
			AuthScheme:               scheme,
			Credentials:              types.MapNull(types.StringType),
			RestrictToFollowingTools: tools,
		}
		m.ManagedAuth = nil
	}
}

func optionalStringSet(values []string, prior types.Set) types.Set {
	if len(values) == 0 && (prior.IsNull() || prior.IsUnknown()) {
		return types.SetNull(types.StringType)
	}
	return stringSet(values)
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

func formatAPIError(err error) string {
	var apiErr *api.APIError
	if !errors.As(err, &apiErr) {
		return err.Error()
	}
	b := strings.Builder{}
	b.WriteString(apiErr.Message)
	b.WriteString("\n\n")
	fmt.Fprintf(&b, "HTTP status: %d\n", apiErr.StatusCode)
	if apiErr.RequestID != "" {
		b.WriteString("Request ID: ")
		b.WriteString(apiErr.RequestID)
		b.WriteString("\n")
	}
	if apiErr.Code != "" {
		b.WriteString("Error code: ")
		b.WriteString(apiErr.Code)
		b.WriteString("\n")
	}
	return strings.TrimSpace(b.String())
}
