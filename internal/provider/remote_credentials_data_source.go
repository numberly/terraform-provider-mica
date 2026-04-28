package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

var _ datasource.DataSource = &remoteCredentialsDataSource{}
var _ datasource.DataSourceWithConfigure = &remoteCredentialsDataSource{}

// remoteCredentialsDataSource implements the flashblade_object_store_remote_credentials data source.
type remoteCredentialsDataSource struct {
	client *client.FlashBladeClient
}

func NewRemoteCredentialsDataSource() datasource.DataSource {
	return &remoteCredentialsDataSource{}
}

// ---------- model structs ----------------------------------------------------

// remoteCredentialsDataSourceModel is the top-level model for the flashblade_object_store_remote_credentials data source.
type remoteCredentialsDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	AccessKeyID types.String `tfsdk:"access_key_id"`
	RemoteName  types.String `tfsdk:"remote_name"`
}

// ---------- data source interface methods -----------------------------------

func (d *remoteCredentialsDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "flashblade_object_store_remote_credentials"
}

// Schema defines the data source schema.
func (d *remoteCredentialsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads existing FlashBlade object store remote credentials by name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the remote credentials.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the remote credentials to look up.",
			},
			"access_key_id": schema.StringAttribute{
				Computed:    true,
				Description: "The access key ID of the remote credentials.",
			},
			"remote_name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the remote array connection.",
			},
		},
	}
}

// Configure injects the FlashBladeClient into the data source.
func (d *remoteCredentialsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read fetches remote credentials data by name and populates state.
func (d *remoteCredentialsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config remoteCredentialsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := config.Name.ValueString()
	cred, err := d.client.GetRemoteCredentials(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Remote credentials not found",
				fmt.Sprintf("No remote credentials with name %q exist on the FlashBlade array.", name),
			)
			return
		}
		resp.Diagnostics.AddError("Error reading remote credentials", err.Error())
		return
	}

	config.ID = types.StringValue(cred.ID)
	config.Name = types.StringValue(cred.Name)
	config.AccessKeyID = types.StringValue(cred.AccessKeyID)
	config.RemoteName = types.StringValue(cred.Remote.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
