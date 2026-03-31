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

// Ensure networkInterfaceDataSource satisfies the datasource interfaces.
var _ datasource.DataSource = &networkInterfaceDataSource{}
var _ datasource.DataSourceWithConfigure = &networkInterfaceDataSource{}

// networkInterfaceDataSource implements the flashblade_network_interface data source.
type networkInterfaceDataSource struct {
	client *client.FlashBladeClient
}

// NewNetworkInterfaceDataSource is the factory function registered in the provider.
func NewNetworkInterfaceDataSource() datasource.DataSource {
	return &networkInterfaceDataSource{}
}

// ---------- model structs ----------------------------------------------------

// networkInterfaceDataSourceModel is the top-level model for the flashblade_network_interface data source.
type networkInterfaceDataSourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Address         types.String `tfsdk:"address"`
	SubnetName      types.String `tfsdk:"subnet_name"`
	Type            types.String `tfsdk:"type"`
	Services        types.String `tfsdk:"services"`
	AttachedServers types.List   `tfsdk:"attached_servers"`
	Enabled         types.Bool   `tfsdk:"enabled"`
	Gateway         types.String `tfsdk:"gateway"`
	MTU             types.Int64  `tfsdk:"mtu"`
	Netmask         types.String `tfsdk:"netmask"`
	VLAN            types.Int64  `tfsdk:"vlan"`
	Realms          types.List   `tfsdk:"realms"`
}

// ---------- data source interface methods -----------------------------------

// Metadata sets the Terraform type name.
func (d *networkInterfaceDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "flashblade_network_interface"
}

// Schema defines the data source schema.
func (d *networkInterfaceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing FlashBlade network interface by name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the network interface.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the network interface to look up.",
			},
			"address": schema.StringAttribute{
				Computed:    true,
				Description: "The IPv4 address of the network interface.",
			},
			"subnet_name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the subnet this interface is attached to.",
			},
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "The network interface type (e.g. vip).",
			},
			"services": schema.StringAttribute{
				Computed:    true,
				Description: "The service type for this network interface (data, sts, egress-only, replication).",
			},
			"attached_servers": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of server names attached to this interface.",
			},
			"enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the network interface is enabled.",
			},
			"gateway": schema.StringAttribute{
				Computed:    true,
				Description: "The gateway address for this network interface.",
			},
			"mtu": schema.Int64Attribute{
				Computed:    true,
				Description: "Maximum transmission unit (MTU) in bytes.",
			},
			"netmask": schema.StringAttribute{
				Computed:    true,
				Description: "The subnet mask for this network interface.",
			},
			"vlan": schema.Int64Attribute{
				Computed:    true,
				Description: "VLAN ID. 0 means untagged.",
			},
			"realms": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of realms associated with this network interface.",
			},
		},
	}
}

// Configure injects the FlashBladeClient into the data source.
func (d *networkInterfaceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read fetches network interface data by name and populates state.
func (d *networkInterfaceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config networkInterfaceDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := config.Name.ValueString()
	ni, err := d.client.GetNetworkInterface(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Network interface not found",
				fmt.Sprintf("No network interface with name %q exists on the FlashBlade array.", name),
			)
			return
		}
		resp.Diagnostics.AddError("Error reading network interface", err.Error())
		return
	}

	mapNetworkInterfaceToDataSourceModel(ctx, ni, &config, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

// mapNetworkInterfaceToDataSourceModel maps a *client.NetworkInterface to a *networkInterfaceDataSourceModel.
func mapNetworkInterfaceToDataSourceModel(ctx context.Context, ni *client.NetworkInterface, data *networkInterfaceDataSourceModel, diags *diag.Diagnostics) {
	data.ID = types.StringValue(ni.ID)
	data.Name = types.StringValue(ni.Name)
	data.Address = types.StringValue(ni.Address)
	data.Type = types.StringValue(ni.Type)
	data.Enabled = types.BoolValue(ni.Enabled)
	data.MTU = types.Int64Value(ni.MTU)
	data.VLAN = types.Int64Value(ni.VLAN)
	data.Gateway = stringOrNull(ni.Gateway)
	data.Netmask = stringOrNull(ni.Netmask)
	data.SubnetName = refToSubnetName(ni.Subnet)

	// Collapse []string services to a single string value.
	if len(ni.Services) > 0 {
		data.Services = types.StringValue(ni.Services[0])
	} else {
		data.Services = types.StringNull()
	}

	// AttachedServers: always use a list (empty, not null) to prevent spurious drift.
	if len(ni.AttachedServers) > 0 {
		names := make([]string, 0, len(ni.AttachedServers))
		for _, s := range ni.AttachedServers {
			names = append(names, s.Name)
		}
		serverList, serverDiags := types.ListValueFrom(ctx, types.StringType, names)
		diags.Append(serverDiags...)
		if diags.HasError() {
			return
		}
		data.AttachedServers = serverList
	} else {
		emptyList, emptyDiags := types.ListValueFrom(ctx, types.StringType, []string{})
		diags.Append(emptyDiags...)
		if diags.HasError() {
			return
		}
		data.AttachedServers = emptyList
	}

	// Realms: null when absent, list otherwise.
	if len(ni.Realms) > 0 {
		realmList, realmDiags := types.ListValueFrom(ctx, types.StringType, ni.Realms)
		diags.Append(realmDiags...)
		if diags.HasError() {
			return
		}
		data.Realms = realmList
	} else {
		data.Realms = types.ListNull(types.StringType)
	}
}
