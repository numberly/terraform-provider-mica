package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

var _ datasource.DataSource = &subnetDataSource{}
var _ datasource.DataSourceWithConfigure = &subnetDataSource{}

// subnetDataSource implements the flashblade_subnet data source.
type subnetDataSource struct {
	client *client.FlashBladeClient
}

func NewSubnetDataSource() datasource.DataSource {
	return &subnetDataSource{}
}

// ---------- model structs ----------------------------------------------------

// subnetDataSourceModel is the top-level model for the flashblade_subnet data source.
type subnetDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Prefix     types.String `tfsdk:"prefix"`
	Gateway    types.String `tfsdk:"gateway"`
	MTU        types.Int64  `tfsdk:"mtu"`
	VLAN       types.Int64  `tfsdk:"vlan"`
	LagName    types.String `tfsdk:"lag_name"`
	Enabled    types.Bool   `tfsdk:"enabled"`
	Services   types.List   `tfsdk:"services"`
	Interfaces types.List   `tfsdk:"interfaces"`
}

// ---------- data source interface methods -----------------------------------

func (d *subnetDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "flashblade_subnet"
}

// Schema defines the data source schema.
func (d *subnetDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing FlashBlade subnet by name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the subnet.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the subnet to look up.",
			},
			"prefix": schema.StringAttribute{
				Computed:    true,
				Description: "IPv4 or IPv6 subnet address in CIDR notation.",
			},
			"gateway": schema.StringAttribute{
				Computed:    true,
				Description: "IPv4 or IPv6 gateway address for the subnet.",
			},
			"mtu": schema.Int64Attribute{
				Computed:    true,
				Description: "Maximum transmission unit (MTU) in bytes.",
			},
			"vlan": schema.Int64Attribute{
				Computed:    true,
				Description: "VLAN ID. 0 means untagged.",
			},
			"lag_name": schema.StringAttribute{
				Computed:    true,
				Description: "Name of the link aggregation group (LAG) this subnet is attached to.",
			},
			"enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the subnet is enabled.",
			},
			"services": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of services associated with this subnet.",
			},
			"interfaces": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of network interface names attached to this subnet.",
			},
		},
	}
}

// Configure injects the FlashBladeClient into the data source.
func (d *subnetDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read fetches subnet data by name and populates state.
func (d *subnetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config subnetDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := config.Name.ValueString()
	subnet, err := d.client.GetSubnet(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Subnet not found",
				fmt.Sprintf("No subnet with name %q exists on the FlashBlade array.", name),
			)
			return
		}
		resp.Diagnostics.AddError("Error reading subnet", err.Error())
		return
	}

	mapSubnetToDataSourceModel(ctx, subnet, &config, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

// mapSubnetToDataSourceModel maps a client.Subnet to a subnetDataSourceModel.
func mapSubnetToDataSourceModel(ctx context.Context, subnet *client.Subnet, data *subnetDataSourceModel, diags *diag.Diagnostics) {
	data.ID = types.StringValue(subnet.ID)
	data.Name = types.StringValue(subnet.Name)
	data.Prefix = types.StringValue(subnet.Prefix)
	data.Gateway = stringOrNull(subnet.Gateway)
	data.MTU = types.Int64Value(subnet.MTU)
	data.VLAN = types.Int64Value(subnet.VLAN)
	data.Enabled = types.BoolValue(subnet.Enabled)
	data.LagName = refToLagName(subnet.LinkAggregationGroup)

	// Map services list.
	if len(subnet.Services) > 0 {
		svcList, svcDiags := types.ListValueFrom(ctx, types.StringType, subnet.Services)
		diags.Append(svcDiags...)
		if diags.HasError() {
			return
		}
		data.Services = svcList
	} else {
		data.Services = types.ListNull(types.StringType)
	}

	// Map interfaces list (extract .Name from each NamedReference).
	if len(subnet.Interfaces) > 0 {
		ifaceNames := make([]string, 0, len(subnet.Interfaces))
		for _, iface := range subnet.Interfaces {
			ifaceNames = append(ifaceNames, iface.Name)
		}
		ifaceList, ifaceDiags := types.ListValueFrom(ctx, types.StringType, ifaceNames)
		diags.Append(ifaceDiags...)
		if diags.HasError() {
			return
		}
		data.Interfaces = ifaceList
	} else {
		data.Interfaces = types.ListNull(types.StringType)
	}
}
