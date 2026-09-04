package provider

import (
	"context"
	"fmt"

	yourservice "github.com/example/terraform-provider-template/internal/your_service"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &exampleItemDataSource{}
	_ datasource.DataSourceWithConfigure = &exampleItemDataSource{}
)

type exampleItemDataSource struct {
	client *yourservice.Client
}

type exampleItemDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Endpoint    types.String `tfsdk:"endpoint"`
}

func NewExampleItemDataSource() datasource.DataSource {
	return &exampleItemDataSource{}
}

func (d *exampleItemDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_example_item"
}

func (d *exampleItemDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an example item from the template client. Replace this with a real API-backed data source.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Stable identifier derived from the item name.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Example item name.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Example item description.",
			},
			"endpoint": schema.StringAttribute{
				Computed:    true,
				Description: "Endpoint configured for the template client.",
			},
		},
	}
}

func (d *exampleItemDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*yourservice.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *yourservice.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *exampleItemDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config exampleItemDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := yourservice.DefaultEndpoint
	if d.client != nil {
		endpoint = d.client.Endpoint
	}

	state := exampleItemDataSourceModel{
		ID:          types.StringValue(exampleItemID(config.Name.ValueString())),
		Name:        config.Name,
		Description: types.StringValue("Example item named " + config.Name.ValueString()),
		Endpoint:    types.StringValue(endpoint),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
