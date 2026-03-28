package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// Ensure serverDataSource satisfies the datasource interfaces.
var _ datasource.DataSource = &serverDataSource{}
var _ datasource.DataSourceWithConfigure = &serverDataSource{}

// serverDataSource implements the flashblade_server data source.
type serverDataSource struct {
	client *client.FlashBladeClient
}

// NewServerDataSource is the factory function registered in the provider.
func NewServerDataSource() datasource.DataSource {
	return &serverDataSource{}
}

// ---------- model structs ----------------------------------------------------

// serverDataSourceModel is the top-level model for the flashblade_server data source.
type serverDataSourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// ---------- data source interface methods -----------------------------------

// Metadata sets the Terraform type name.
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

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
