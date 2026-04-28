package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

var _ datasource.DataSource = &linkAggregationGroupDataSource{}
var _ datasource.DataSourceWithConfigure = &linkAggregationGroupDataSource{}

// linkAggregationGroupDataSource implements the flashblade_link_aggregation_group data source.
type linkAggregationGroupDataSource struct {
	client *client.FlashBladeClient
}

func NewLinkAggregationGroupDataSource() datasource.DataSource {
	return &linkAggregationGroupDataSource{}
}

// ---------- model structs ----------------------------------------------------

// lagDataSourceModel is the top-level model for the flashblade_link_aggregation_group data source.
type lagDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Status     types.String `tfsdk:"status"`
	MacAddress types.String `tfsdk:"mac_address"`
	PortSpeed  types.Int64  `tfsdk:"port_speed"`
	LagSpeed   types.Int64  `tfsdk:"lag_speed"`
	Ports      types.List   `tfsdk:"ports"`
}

// ---------- data source interface methods -----------------------------------

func (d *linkAggregationGroupDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "flashblade_link_aggregation_group"
}

// Schema defines the data source schema.
func (d *linkAggregationGroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing FlashBlade link aggregation group (LAG) by name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the link aggregation group.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the LAG to look up (e.g. lag0).",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Status of the LAG (e.g. healthy, degraded).",
			},
			"mac_address": schema.StringAttribute{
				Computed:    true,
				Description: "MAC address of the LAG.",
			},
			"port_speed": schema.Int64Attribute{
				Computed:    true,
				Description: "Speed of each individual port in bits per second.",
			},
			"lag_speed": schema.Int64Attribute{
				Computed:    true,
				Description: "Aggregate speed of the LAG in bits per second (sum of port speeds).",
			},
			"ports": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of port names (e.g. eth0, eth1) in this LAG.",
			},
		},
	}
}

// Configure injects the FlashBladeClient into the data source.
func (d *linkAggregationGroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.FlashBladeClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data Type",
			fmt.Sprintf("Expected *client.FlashBladeClient, got: %T. This is a bug in the provider.", req.ProviderData),
		)
		return
	}
	d.client = c
}

// Read fetches LAG data by name and populates state.
func (d *linkAggregationGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config lagDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := config.Name.ValueString()
	lag, err := d.client.GetLinkAggregationGroup(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Link aggregation group not found",
				fmt.Sprintf("No link aggregation group with name %q exists on the FlashBlade array.", name),
			)
			return
		}
		resp.Diagnostics.AddError("Error reading link aggregation group", err.Error())
		return
	}

	config.ID = stringOrNull(lag.ID)
	config.Name = types.StringValue(lag.Name)
	config.Status = stringOrNull(lag.Status)
	config.MacAddress = stringOrNull(lag.MacAddress)
	config.PortSpeed = types.Int64Value(lag.PortSpeed)
	config.LagSpeed = types.Int64Value(lag.LagSpeed)

	if len(lag.Ports) > 0 {
		portNames := make([]string, len(lag.Ports))
		for i, p := range lag.Ports {
			portNames[i] = p.Name
		}
		portsList, diags := types.ListValueFrom(ctx, types.StringType, portNames)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		config.Ports = portsList
	} else {
		config.Ports = types.ListNull(types.StringType)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
