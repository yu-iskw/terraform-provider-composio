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
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
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
	ID                   types.String      `tfsdk:"id"`
	ToolkitSlug          types.String      `tfsdk:"toolkit_slug"`
	Name                 types.String      `tfsdk:"name"`
	Enabled              types.Bool        `tfsdk:"enabled"`
	EnabledForToolRouter types.Bool        `tfsdk:"enabled_for_tool_router"`
	ManagedAuth          *managedAuthModel `tfsdk:"managed_auth"`
	CustomAuth           *customAuthModel  `tfsdk:"custom_auth"`
	AuthScheme           types.String      `tfsdk:"auth_scheme"`
	IsComposioManaged    types.Bool        `tfsdk:"is_composio_managed"`
	Status               types.String      `tfsdk:"status"`
	CreatedAt            types.String      `tfsdk:"created_at"`
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "When false, the provider disables the auth config when Composio still reports it enabled. Disabled configs cannot start new connections.",
			},
			"enabled_for_tool_router": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "When true, sessions can match connected accounts for this auth config by `user_id` automatically. Required for authenticated Custom MCP toolkits used in sessions. Maps to Composio `is_enabled_for_tool_router`.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"managed_auth": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Use Composio-managed authentication. Exactly one of `managed_auth` or `custom_auth` must be set. Switching to `custom_auth` forces replacement.",
				Attributes: map[string]schema.Attribute{
					"restrict_to_following_tools": restrictToFollowingToolsAttribute(),
					"scopes": schema.SetAttribute{
						Optional:            true,
						Computed:            true,
						ElementType:         types.StringType,
						MarkdownDescription: "OAuth scopes for managed auth. If omitted, Terraform stores the scopes Composio reports. Set to `[]` to clear. Order is not significant.",
						PlanModifiers: []planmodifier.Set{
							setplanmodifier.UseStateForUnknown(),
						},
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
						MarkdownDescription: "Write-only credential map. Never stored in state. Sent on create when set. Sent on update only when Terraform also plans another change. A credentials-only edit does not produce a plan. Terraform 1.15 or later is required.",
					},
					"restrict_to_following_tools": restrictToFollowingToolsAttribute(),
				},
			},
			"auth_scheme": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Scheme reported by Composio.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"is_composio_managed": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "True when Composio manages the credentials.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "`ENABLED` or `DISABLED`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Creation timestamp from Composio.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func restrictToFollowingToolsAttribute() schema.SetAttribute {
	return schema.SetAttribute{
		Optional:            true,
		Computed:            true,
		ElementType:         types.StringType,
		MarkdownDescription: "Tool slugs this auth config may use. If omitted, Terraform stores the tools Composio reports. Set to `[]` to clear. Order is not significant.",
		PlanModifiers: []planmodifier.Set{
			setplanmodifier.UseStateForUnknown(),
		},
	}
}

func (r *authConfigResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var managed, custom types.Object
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("managed_auth"), &managed)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("custom_auth"), &custom)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if managed.IsUnknown() || custom.IsUnknown() {
		return
	}
	if managed.IsNull() == custom.IsNull() {
		resp.Diagnostics.AddError(
			"Invalid Auth Config Mode",
			"Set exactly one of `managed_auth` or `custom_auth`.",
		)
	}
}

func (r *authConfigResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := configureProjectClient(req.ProviderData, "composio_auth_config")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() || client == nil {
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
		resp.Diagnostics.AddError("Unable to create Composio auth config", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(id))...)
	if resp.Diagnostics.HasError() {
		if delErr := r.client.DeleteAuthConfig(ctx, id); delErr != nil {
			resp.Diagnostics.AddError("Unable to roll back Composio auth config after a state write failure", delErr.Error())
		}
		return
	}
	if err := r.loadAfterWrite(ctx, id, &plan); err != nil {
		resp.Diagnostics.AddError("Unable to read Composio auth config after create", err.Error())
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
		resp.Diagnostics.AddError("Unable to read Composio auth config", err.Error())
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
	var state authConfigResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := plan.ID.ValueString()
	if authConfigPatchNeeded(plan, state) {
		in, diags := updateInputFromModels(ctx, plan, config)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		if err := r.client.UpdateAuthConfig(ctx, id, in); err != nil {
			resp.Diagnostics.AddError("Unable to update Composio auth config", err.Error())
			return
		}
	}
	if err := r.loadAfterWrite(ctx, id, &plan); err != nil {
		resp.Diagnostics.AddError("Unable to read Composio auth config after update", err.Error())
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
		resp.Diagnostics.AddError("Unable to delete Composio auth config", err.Error())
	}
}

func (r *authConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *authConfigResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}
	var stateManaged, stateCustom, planManaged, planCustom types.Object
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("managed_auth"), &stateManaged)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("custom_auth"), &stateCustom)...)
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("managed_auth"), &planManaged)...)
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("custom_auth"), &planCustom)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if stateManaged.IsUnknown() || planManaged.IsUnknown() || stateCustom.IsUnknown() || planCustom.IsUnknown() {
		return
	}
	if stateManaged.IsNull() != planManaged.IsNull() {
		resp.RequiresReplace = append(resp.RequiresReplace, path.Root("managed_auth"), path.Root("custom_auth"))
	}

	var planEnabled, stateEnabled types.Bool
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("enabled"), &planEnabled)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("enabled"), &stateEnabled)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !planEnabled.IsUnknown() && !planEnabled.Equal(stateEnabled) {
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("status"), types.StringUnknown())...)
	}
}

func (r *authConfigResource) loadAfterWrite(ctx context.Context, id string, dest *authConfigResourceModel) error {
	remote, err := r.client.GetAuthConfig(ctx, id)
	if err != nil {
		return err
	}
	wantEnabled := dest.Enabled.ValueBool()
	if remote.Enabled() != wantEnabled {
		status := models.AuthConfigStatusEnabled
		if !wantEnabled {
			status = models.AuthConfigStatusDisabled
		}
		if err := r.client.SetAuthConfigStatus(ctx, id, status); err != nil {
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
	if !plan.EnabledForToolRouter.IsNull() && !plan.EnabledForToolRouter.IsUnknown() {
		v := plan.EnabledForToolRouter.ValueBool()
		in.EnabledForToolRouter = &v
	}
	in.RestrictToFollowingTools, diags = setToStrings(ctx, plan.restrictToFollowingTools())
	if diags.HasError() {
		return in, diags
	}
	switch {
	case plan.ManagedAuth != nil:
		in.Managed = true
		in.Scopes, diags = setToStrings(ctx, plan.ManagedAuth.Scopes)
	case plan.CustomAuth != nil:
		in.Managed = false
		in.AuthScheme = plan.CustomAuth.AuthScheme.ValueString()
		creds, d := credentialsFromConfig(ctx, config.CustomAuth)
		diags.Append(d...)
		if creds == nil {
			creds = map[string]string{}
		}
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
	tools := plan.restrictToFollowingTools()
	if !tools.IsNull() && !tools.IsUnknown() {
		converted, d := setToStrings(ctx, tools)
		diags.Append(d...)
		in.RestrictToFollowingTools = &converted
	}
	if plan.ManagedAuth != nil {
		if !plan.ManagedAuth.Scopes.IsNull() && !plan.ManagedAuth.Scopes.IsUnknown() {
			scopes, d := setToStrings(ctx, plan.ManagedAuth.Scopes)
			diags.Append(d...)
			in.Scopes = &scopes
		}
	}
	if plan.CustomAuth != nil {
		creds, d := credentialsFromConfig(ctx, config.CustomAuth)
		diags.Append(d...)
		if len(creds) > 0 {
			in.Credentials = creds
		}
	}
	if !plan.EnabledForToolRouter.IsNull() && !plan.EnabledForToolRouter.IsUnknown() {
		v := plan.EnabledForToolRouter.ValueBool()
		in.EnabledForToolRouter = &v
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

func (m authConfigResourceModel) restrictToFollowingTools() types.Set {
	if m.ManagedAuth != nil {
		return m.ManagedAuth.RestrictToFollowingTools
	}
	if m.CustomAuth != nil {
		return m.CustomAuth.RestrictToFollowingTools
	}
	return types.SetNull(types.StringType)
}

func (m *authConfigResourceModel) applyRemote(remote models.AuthConfig) {
	m.ID = types.StringValue(remote.ID)
	if remote.ToolkitSlug != "" {
		configured := m.ToolkitSlug.ValueString()
		if configured == "" || !strings.EqualFold(configured, remote.ToolkitSlug) {
			m.ToolkitSlug = types.StringValue(remote.ToolkitSlug)
		}
	}
	m.Name = types.StringValue(remote.Name)
	m.Enabled = types.BoolValue(remote.Enabled())
	if remote.IsEnabledForToolRouter != nil {
		m.EnabledForToolRouter = types.BoolValue(*remote.IsEnabledForToolRouter)
	} else if m.EnabledForToolRouter.IsUnknown() {
		m.EnabledForToolRouter = types.BoolNull()
	}
	m.AuthScheme = types.StringValue(remote.AuthScheme)
	m.IsComposioManaged = types.BoolValue(remote.IsComposioManaged)
	m.Status = types.StringValue(remote.Status)
	m.CreatedAt = types.StringValue(remote.CreatedAt)

	tools := optionalStringSet(remote.RestrictToFollowingTools, m.restrictToFollowingTools())
	if remote.IsComposioManaged {
		priorScopes := types.SetNull(types.StringType)
		if m.ManagedAuth != nil {
			priorScopes = m.ManagedAuth.Scopes
		}
		m.ManagedAuth = &managedAuthModel{
			RestrictToFollowingTools: tools,
			Scopes:                   remoteStringSet(remote.Scopes, priorScopes),
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

func remoteStringSet(values *[]string, prior types.Set) types.Set {
	if values == nil {
		if prior.IsUnknown() {
			return types.SetNull(types.StringType)
		}
		return prior
	}
	return optionalStringSet(*values, prior)
}

func authConfigPatchNeeded(plan, state authConfigResourceModel) bool {
	if !plan.Name.IsUnknown() && !plan.Name.Equal(state.Name) {
		return true
	}
	if !plan.EnabledForToolRouter.IsUnknown() && !plan.EnabledForToolRouter.Equal(state.EnabledForToolRouter) {
		return true
	}
	planTools := plan.restrictToFollowingTools()
	if !planTools.IsUnknown() && !setsEqual(planTools, state.restrictToFollowingTools()) {
		return true
	}
	if plan.ManagedAuth != nil && state.ManagedAuth != nil {
		return !plan.ManagedAuth.Scopes.IsUnknown() && !setsEqual(plan.ManagedAuth.Scopes, state.ManagedAuth.Scopes)
	}
	if plan.CustomAuth != nil && state.CustomAuth != nil {
		return false
	}
	return true
}

// setsEqual treats two nulls or two unknowns as equal even when one is a
// zero-value set. types.Set.Equal returns false if either side has no
// element type, which would force a PATCH on enabled-only updates.
func setsEqual(a, b types.Set) bool {
	if a.IsNull() && b.IsNull() {
		return true
	}
	if a.IsUnknown() && b.IsUnknown() {
		return true
	}
	return a.Equal(b)
}
