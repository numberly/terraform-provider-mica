package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

var _ datasource.DataSource = &objectStoreVirtualHostDataSource{}
var _ datasource.DataSourceWithConfigure = &objectStoreVirtualHostDataSource{}

// objectStoreVirtualHostDataSource implements the flashblade_object_store_virtual_host data source.
type objectStoreVirtualHostDataSource struct {
	client *client.FlashBladeClient
}

func NewObjectStoreVirtualHostDataSource() datasource.DataSource {
	return &objectStoreVirtualHostDataSource{}
}

// ---------- model structs ----------------------------------------------------

// objectStoreVirtualHostDataSourceModel is the top-level model for the data source.
type objectStoreVirtualHostDataSourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Filter          types.String `tfsdk:"filter"`
	Hostname        types.String `tfsdk:"hostname"`
	AttachedServers types.List   `tfsdk:"attached_servers"`
}

// ---------- data source interface methods -----------------------------------

func (d *objectStoreVirtualHostDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "flashblade_object_store_virtual_host"
}

// Schema defines the data source schema.
func (d *objectStoreVirtualHostDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing FlashBlade object store virtual host.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the virtual host.",
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "The name of the virtual host to look up.",
			},
			"filter": schema.StringAttribute{
				Optional:    true,
				Description: "A filter expression for listing virtual hosts.",
			},
			"hostname": schema.StringAttribute{
				Computed:    true,
				Description: "The hostname of the virtual host.",
			},
			"attached_servers": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of server names attached to this virtual host.",
			},
		},
	}
}

// Configure injects the FlashBladeClient into the data source.
func (d *objectStoreVirtualHostDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read fetches virtual host data and populates state.
func (d *objectStoreVirtualHostDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config objectStoreVirtualHostDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var vh *client.ObjectStoreVirtualHost

	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		// Lookup by name.
		result, err := d.client.GetObjectStoreVirtualHost(ctx, config.Name.ValueString())
		if err != nil {
			if client.IsNotFound(err) {
				resp.Diagnostics.AddError(
					"Object store virtual host not found",
					fmt.Sprintf("No object store virtual host with name %q exists on the FlashBlade array.", config.Name.ValueString()),
				)
				return
			}
			resp.Diagnostics.AddError("Error reading object store virtual host", err.Error())
			return
		}
		vh = result
	} else {
		// List with optional filter.
		opts := client.ListObjectStoreVirtualHostsOpts{}
		if !config.Filter.IsNull() && config.Filter.ValueString() != "" {
			opts.Filter = config.Filter.ValueString()
		}
		hosts, err := d.client.ListObjectStoreVirtualHosts(ctx, opts)
		if err != nil {
			resp.Diagnostics.AddError("Error listing object store virtual hosts", err.Error())
			return
		}
		if len(hosts) == 0 {
			resp.Diagnostics.AddError(
				"No object store virtual hosts found",
				"No virtual hosts matched the provided filter.",
			)
			return
		}
		vh = &hosts[0]
	}

	config.ID = types.StringValue(vh.ID)
	config.Name = types.StringValue(vh.Name)
	config.Hostname = types.StringValue(vh.Hostname)

	if len(vh.AttachedServers) > 0 {
		serverNames := make([]string, len(vh.AttachedServers))
		for i, s := range vh.AttachedServers {
			serverNames[i] = s.Name
		}
		serverList, diags := types.ListValueFrom(ctx, types.StringType, serverNames)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		config.AttachedServers = serverList
	} else {
		config.AttachedServers = types.ListValueMust(types.StringType, []attr.Value{})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
