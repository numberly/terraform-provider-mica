package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// Ensure objectStoreUserDataSource satisfies the datasource interfaces.
var _ datasource.DataSource = &objectStoreUserDataSource{}
var _ datasource.DataSourceWithConfigure = &objectStoreUserDataSource{}

// objectStoreUserDataSource implements the flashblade_object_store_user data source.
type objectStoreUserDataSource struct {
	client *client.FlashBladeClient
}

// NewObjectStoreUserDataSource is the factory function registered in the provider.
func NewObjectStoreUserDataSource() datasource.DataSource {
	return &objectStoreUserDataSource{}
}

// ---------- model structs ----------------------------------------------------

// objectStoreUserDataSourceModel is the top-level model for the flashblade_object_store_user data source.
type objectStoreUserDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	FullAccess types.Bool   `tfsdk:"full_access"`
}

// ---------- data source interface methods -----------------------------------

// Metadata sets the Terraform type name.
func (d *objectStoreUserDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "flashblade_object_store_user"
}

// Schema defines the data source schema.
func (d *objectStoreUserDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing FlashBlade object store user by name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the object store user.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the object store user in the format account/username.",
			},
			"full_access": schema.BoolAttribute{
				Computed:    true,
				Description: "If true, the user has full access to all object store operations.",
			},
		},
	}
}

// Configure injects the FlashBladeClient into the data source.
func (d *objectStoreUserDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read fetches object store user data by name and populates state.
func (d *objectStoreUserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config objectStoreUserDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := config.Name.ValueString()
	user, err := d.client.GetObjectStoreUser(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Object store user not found",
				fmt.Sprintf("No object store user with name %q exists on the FlashBlade array.", name),
			)
			return
		}
		resp.Diagnostics.AddError("Error reading object store user", err.Error())
		return
	}

	config.ID = types.StringValue(user.ID)
	config.Name = types.StringValue(user.Name)
	config.FullAccess = types.BoolValue(user.FullAccess)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
