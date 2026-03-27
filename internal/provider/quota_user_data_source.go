package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/soulkyu/terraform-provider-flashblade/internal/client"
)

// Ensure quotaUserDataSource satisfies the datasource interfaces.
var _ datasource.DataSource = &quotaUserDataSource{}
var _ datasource.DataSourceWithConfigure = &quotaUserDataSource{}

// quotaUserDataSource implements the flashblade_quota_user data source.
type quotaUserDataSource struct {
	client *client.FlashBladeClient
}

// NewQuotaUserDataSource is the factory function registered in the provider.
func NewQuotaUserDataSource() datasource.DataSource {
	return &quotaUserDataSource{}
}

// ---------- model structs ----------------------------------------------------

// quotaUserDataSourceModel is the top-level model for the flashblade_quota_user data source.
type quotaUserDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	FileSystemName types.String `tfsdk:"file_system_name"`
	UID            types.String `tfsdk:"uid"`
	Quota          types.Int64  `tfsdk:"quota"`
	Usage          types.Int64  `tfsdk:"usage"`
}

// ---------- data source interface methods -----------------------------------

// Metadata sets the Terraform type name.
func (d *quotaUserDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "flashblade_quota_user"
}

// Schema defines the data source schema.
func (d *quotaUserDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing per-filesystem user quota from a FlashBlade array.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Synthetic composite identifier in the form file_system_name/uid.",
			},
			"file_system_name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the file system the quota is attached to.",
			},
			"uid": schema.StringAttribute{
				Required:    true,
				Description: "User ID (UID) the quota applies to.",
			},
			"quota": schema.Int64Attribute{
				Computed:    true,
				Description: "Quota limit in bytes.",
			},
			"usage": schema.Int64Attribute{
				Computed:    true,
				Description: "Current usage in bytes.",
			},
		},
	}
}

// Configure injects the FlashBladeClient into the data source.
func (d *quotaUserDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read fetches user quota data and populates state.
func (d *quotaUserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config quotaUserDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	fsName := config.FileSystemName.ValueString()
	uid := config.UID.ValueString()

	qu, err := d.client.GetQuotaUser(ctx, fsName, uid)
	if err != nil {
		if client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"User quota not found",
				fmt.Sprintf("No user quota for UID %q on file system %q exists on the FlashBlade array.", uid, fsName),
			)
			return
		}
		resp.Diagnostics.AddError("Error reading user quota", err.Error())
		return
	}

	config.ID = types.StringValue(fsName + "/" + uid)
	config.FileSystemName = types.StringValue(fsName)
	config.UID = types.StringValue(uid)
	config.Quota = types.Int64Value(qu.Quota)
	config.Usage = types.Int64Value(qu.Usage)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
