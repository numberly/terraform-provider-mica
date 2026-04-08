package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// Ensure arrayConnectionDataSource satisfies the datasource interfaces.
var _ datasource.DataSource = &arrayConnectionDataSource{}
var _ datasource.DataSourceWithConfigure = &arrayConnectionDataSource{}

// arrayConnectionDataSource implements the flashblade_array_connection data source.
type arrayConnectionDataSource struct {
	client *client.FlashBladeClient
}

// NewArrayConnectionDataSource is the factory function registered in the provider.
func NewArrayConnectionDataSource() datasource.DataSource {
	return &arrayConnectionDataSource{}
}

// ---------- model structs ----------------------------------------------------

// arrayConnectionDataSourceModel is the top-level model for the flashblade_array_connection data source.
type arrayConnectionDataSourceModel struct {
	ID                   types.String `tfsdk:"id"`
	RemoteName           types.String `tfsdk:"remote_name"`
	RemoteID             types.String `tfsdk:"remote_id"`
	Status               types.String `tfsdk:"status"`
	ManagementAddress    types.String `tfsdk:"management_address"`
	ReplicationAddresses types.List   `tfsdk:"replication_addresses"`
	Encrypted            types.Bool   `tfsdk:"encrypted"`
	Type                 types.String `tfsdk:"type"`
	Version              types.String `tfsdk:"version"`
	CACertificateGroup   types.String `tfsdk:"ca_certificate_group"`
	OS                   types.String `tfsdk:"os"`
}

// ---------- data source interface methods -----------------------------------

// Metadata sets the Terraform type name.
func (d *arrayConnectionDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "flashblade_array_connection"
}

// Schema defines the data source schema.
func (d *arrayConnectionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing FlashBlade array connection by remote array name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the array connection.",
			},
			"remote_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the remote array to look up.",
			},
			"remote_id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the remote array.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Connection status (connected, partially_connected, connecting, incompatible, not_connected).",
			},
			"management_address": schema.StringAttribute{
				Computed:    true,
				Description: "Management IP address or hostname of the remote array.",
			},
			"replication_addresses": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of replication IP addresses for the remote array.",
			},
			"encrypted": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the array connection is encrypted.",
			},
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "The type of array connection (e.g. async-replication).",
			},
			"version": schema.StringAttribute{
				Computed:    true,
				Description: "The version of the array connection protocol.",
			},
			"ca_certificate_group": schema.StringAttribute{
				Computed:    true,
				Description: "Name of the CA certificate group for TLS verification.",
			},
			"os": schema.StringAttribute{
				Computed:    true,
				Description: "Operating system of the remote array.",
			},
		},
	}
}

// Configure injects the FlashBladeClient into the data source.
func (d *arrayConnectionDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read fetches array connection data by remote name and populates state.
func (d *arrayConnectionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config arrayConnectionDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	remoteName := config.RemoteName.ValueString()
	conn, err := d.client.GetArrayConnection(ctx, remoteName)
	if err != nil {
		if client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Array connection not found",
				fmt.Sprintf("No array connection with remote name %q exists on the FlashBlade array.", remoteName),
			)
			return
		}
		resp.Diagnostics.AddError("Error reading array connection", err.Error())
		return
	}

	config.ID = types.StringValue(conn.ID)
	config.RemoteName = types.StringValue(conn.Remote.Name)
	config.RemoteID = types.StringValue(conn.Remote.ID)
	config.Status = types.StringValue(conn.Status)
	config.ManagementAddress = types.StringValue(conn.ManagementAddress)
	config.Encrypted = types.BoolValue(conn.Encrypted)
	config.Type = types.StringValue(conn.Type)
	config.Version = types.StringValue(conn.Version)

	if len(conn.ReplicationAddresses) > 0 {
		replAddrs, diags := types.ListValueFrom(ctx, types.StringType, conn.ReplicationAddresses)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		config.ReplicationAddresses = replAddrs
	} else {
		config.ReplicationAddresses = types.ListNull(types.StringType)
	}

	if conn.CACertificateGroup != nil {
		config.CACertificateGroup = types.StringValue(conn.CACertificateGroup.Name)
	} else {
		config.CACertificateGroup = types.StringNull()
	}
	config.OS = types.StringValue(conn.OS)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
