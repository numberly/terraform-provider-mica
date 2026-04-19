package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

var _ datasource.DataSource = &quotaGroupDataSource{}
var _ datasource.DataSourceWithConfigure = &quotaGroupDataSource{}

// quotaGroupDataSource implements the flashblade_quota_group data source.
type quotaGroupDataSource struct {
	client *client.FlashBladeClient
}

func NewQuotaGroupDataSource() datasource.DataSource {
	return &quotaGroupDataSource{}
}

// ---------- model structs ----------------------------------------------------

// quotaGroupDataSourceModel is the top-level model for the flashblade_quota_group data source.
type quotaGroupDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	FileSystemName types.String `tfsdk:"file_system_name"`
	GID            types.String `tfsdk:"gid"`
	Quota          types.Int64  `tfsdk:"quota"`
	Usage          types.Int64  `tfsdk:"usage"`
}

// ---------- data source interface methods -----------------------------------

func (d *quotaGroupDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "flashblade_quota_group"
}

// Schema defines the data source schema.
func (d *quotaGroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing per-filesystem group quota from a FlashBlade array.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Synthetic composite identifier in the form file_system_name/gid.",
			},
			"file_system_name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the file system the quota is attached to.",
			},
			"gid": schema.StringAttribute{
				Required:    true,
				Description: "Group ID (GID) the quota applies to.",
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
func (d *quotaGroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read fetches group quota data and populates state.
func (d *quotaGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config quotaGroupDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	fsName := config.FileSystemName.ValueString()
	gid := config.GID.ValueString()

	qg, err := d.client.GetQuotaGroup(ctx, fsName, gid)
	if err != nil {
		if client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Group quota not found",
				fmt.Sprintf("No group quota for GID %q on file system %q exists on the FlashBlade array.", gid, fsName),
			)
			return
		}
		resp.Diagnostics.AddError("Error reading group quota", err.Error())
		return
	}

	config.ID = types.StringValue(fsName + "/" + gid)
	config.FileSystemName = types.StringValue(fsName)
	config.GID = types.StringValue(gid)
	config.Quota = types.Int64Value(qg.Quota)
	config.Usage = types.Int64Value(qg.Usage)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
