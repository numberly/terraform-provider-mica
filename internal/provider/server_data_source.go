package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

var _ datasource.DataSource = &serverDataSource{}
var _ datasource.DataSourceWithConfigure = &serverDataSource{}

// serverDataSource implements the flashblade_server data source.
type serverDataSource struct {
	client *client.FlashBladeClient
}

func NewServerDataSource() datasource.DataSource {
	return &serverDataSource{}
}

// ---------- model structs ----------------------------------------------------

// serverDataSourceModel is the top-level model for the flashblade_server data source.
type serverDataSourceModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Created           types.Int64  `tfsdk:"created"`
	DNS               types.List   `tfsdk:"dns"`
	DirectoryServices types.List   `tfsdk:"directory_services"`
	NetworkInterfaces types.List   `tfsdk:"network_interfaces"`
}

// ---------- data source interface methods -----------------------------------

func (d *serverDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "flashblade_server"
}

// Schema defines the data source schema.
func (d *serverDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing FlashBlade server by name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the server.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the server to look up.",
			},
			"created": schema.Int64Attribute{
				Computed:    true,
				Description: "Unix timestamp (milliseconds) when the server was created.",
			},
			"dns": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of DNS configuration names associated with this server.",
			},
			"directory_services": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of directory service names associated with this server.",
			},
			"network_interfaces": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Names of network interfaces (VIPs) attached to this server. Discovered automatically from the array.",
			},
		},
	}
}

// Configure injects the FlashBladeClient into the data source.
func (d *serverDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read fetches server data by name and populates state.
func (d *serverDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config serverDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := config.Name.ValueString()
	srv, err := d.client.GetServer(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Server not found",
				fmt.Sprintf("No server with name %q exists on the FlashBlade array.", name),
			)
			return
		}
		resp.Diagnostics.AddError("Error reading server", err.Error())
		return
	}

	config.ID = types.StringValue(srv.ID)
	config.Name = types.StringValue(srv.Name)
	config.Created = types.Int64Value(srv.Created)

	// Map DNS names (flat list of strings).
	if len(srv.DNS) > 0 {
		names := make([]string, len(srv.DNS))
		for i, d := range srv.DNS {
			names[i] = d.Name
		}
		dnsList, listDiags := types.ListValueFrom(ctx, types.StringType, names)
		resp.Diagnostics.Append(listDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		config.DNS = dnsList
	} else {
		config.DNS = types.ListNull(types.StringType)
	}

	// Map directory_services names (computed, read-only).
	if len(srv.DirectoryServices) > 0 {
		names := make([]string, len(srv.DirectoryServices))
		for i, ds := range srv.DirectoryServices {
			names[i] = ds.Name
		}
		dsList, listDiags := types.ListValueFrom(ctx, types.StringType, names)
		resp.Diagnostics.Append(listDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		config.DirectoryServices = dsList
	} else {
		config.DirectoryServices = types.ListNull(types.StringType)
	}

	enrichDataSourceNetworkInterfaces(ctx, d.client, &config, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)

}

// enrichDataSourceNetworkInterfaces calls ListNetworkInterfaces and filters by server name.
// Sets data.NetworkInterfaces to an empty list (not null) if no VIPs are attached.
// Appends a warning diagnostic (not error) if the API call fails.
func enrichDataSourceNetworkInterfaces(ctx context.Context, c *client.FlashBladeClient, data *serverDataSourceModel, diags *diag.Diagnostics) {
	nis, err := c.ListNetworkInterfaces(ctx)
	if err != nil {
		diags.AddWarning(
			"Could not list network interfaces",
			fmt.Sprintf("VIP enrichment for server %q failed: %s. network_interfaces will be empty.", data.Name.ValueString(), err.Error()),
		)
		data.NetworkInterfaces = types.ListValueMust(types.StringType, []attr.Value{})
		return
	}

	serverName := data.Name.ValueString()
	var matchingNames []string
	for _, ni := range nis {
		for _, as := range ni.AttachedServers {
			if as.Name == serverName {
				matchingNames = append(matchingNames, ni.Name)
				break
			}
		}
	}

	if matchingNames == nil {
		matchingNames = []string{}
	}

	niList, listDiags := types.ListValueFrom(ctx, types.StringType, matchingNames)
	diags.Append(listDiags...)
	if diags.HasError() {
		return
	}
	data.NetworkInterfaces = niList
}
