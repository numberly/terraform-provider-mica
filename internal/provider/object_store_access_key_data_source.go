package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

var _ datasource.DataSource = &objectStoreAccessKeyDataSource{}
var _ datasource.DataSourceWithConfigure = &objectStoreAccessKeyDataSource{}

// objectStoreAccessKeyDataSource implements the flashblade_object_store_access_key data source.
type objectStoreAccessKeyDataSource struct {
	client *client.FlashBladeClient
}

func NewAccessKeyDataSource() datasource.DataSource {
	return &objectStoreAccessKeyDataSource{}
}

// ---------- model structs ----------------------------------------------------

// objectStoreAccessKeyDataSourceModel is the top-level model for the data source.
type objectStoreAccessKeyDataSourceModel struct {
	Name               types.String `tfsdk:"name"`
	AccessKeyID        types.String `tfsdk:"access_key_id"`
	SecretAccessKey    types.String `tfsdk:"secret_access_key"`
	Created            types.Int64  `tfsdk:"created"`
	Enabled            types.Bool   `tfsdk:"enabled"`
	ObjectStoreAccount types.String `tfsdk:"object_store_account"`
}

// ---------- data source interface methods -----------------------------------

func (d *objectStoreAccessKeyDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "flashblade_object_store_access_key"
}

// Schema defines the data source schema.
func (d *objectStoreAccessKeyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing FlashBlade object store access key by name. Note: secret_access_key is always empty — the API does not return it on GET.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The access key name (format: <account>/admin/<key-id>).",
			},
			"access_key_id": schema.StringAttribute{
				Computed:    true,
				Description: "The access key ID (public part of the credential pair).",
			},
			"secret_access_key": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "The secret access key. Always empty from the data source — the API does not return it on GET. Use the resource to capture it at creation time.",
			},
			"created": schema.Int64Attribute{
				Computed:    true,
				Description: "Unix timestamp (milliseconds) when the access key was created.",
			},
			"enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "If true, the access key is enabled.",
			},
			"object_store_account": schema.StringAttribute{
				Computed:    true,
				Description: "The object store account this access key belongs to.",
			},
		},
	}
}

// Configure injects the FlashBladeClient into the data source.
func (d *objectStoreAccessKeyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read fetches access key data by name and populates state.
// secret_access_key will always be empty — the API never returns it on GET.
func (d *objectStoreAccessKeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config objectStoreAccessKeyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := config.Name.ValueString()
	key, err := d.client.GetObjectStoreAccessKey(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Object store access key not found",
				fmt.Sprintf("No object store access key with name %q exists on the FlashBlade array.", name),
			)
			return
		}
		resp.Diagnostics.AddError("Error reading object store access key", err.Error())
		return
	}

	config.Name = types.StringValue(key.Name)
	config.AccessKeyID = types.StringValue(key.AccessKeyID)
	// SecretAccessKey is always empty from GET — expected behavior.
	config.SecretAccessKey = types.StringValue(key.SecretAccessKey)
	config.Created = types.Int64Value(key.Created)
	config.Enabled = types.BoolValue(key.Enabled)
	// Derive account name from user reference.
	config.ObjectStoreAccount = types.StringValue(extractAccountFromUserName(key.User.Name))

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

// extractAccountFromUserName extracts the account portion from "<account>/admin".
func extractAccountFromUserName(userName string) string {
	if idx := len(userName); idx > 0 {
		// Split on first "/" to get account part.
		for i := 0; i < len(userName); i++ {
			if userName[i] == '/' {
				return userName[:i]
			}
		}
	}
	return userName
}
