package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

var _ datasource.DataSource = &fileSystemExportDataSource{}
var _ datasource.DataSourceWithConfigure = &fileSystemExportDataSource{}

// fileSystemExportDataSource implements the flashblade_file_system_export data source.
type fileSystemExportDataSource struct {
	client *client.FlashBladeClient
}

func NewFileSystemExportDataSource() datasource.DataSource {
	return &fileSystemExportDataSource{}
}

// ---------- model structs ----------------------------------------------------

// fileSystemExportDataSourceModel is the top-level model for the flashblade_file_system_export data source.
type fileSystemExportDataSourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	ExportName      types.String `tfsdk:"export_name"`
	FileSystemName  types.String `tfsdk:"file_system_name"`
	ServerName      types.String `tfsdk:"server_name"`
	SharePolicyName types.String `tfsdk:"share_policy_name"`
	Enabled         types.Bool   `tfsdk:"enabled"`
	PolicyType      types.String `tfsdk:"policy_type"`
	Status          types.String `tfsdk:"status"`
}

// ---------- data source interface methods -----------------------------------

func (d *fileSystemExportDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "flashblade_file_system_export"
}

// Schema defines the data source schema.
func (d *fileSystemExportDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing FlashBlade file system export by name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the file system export.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The combined name of the export to look up (e.g. 'filesystem/export_name').",
			},
			"export_name": schema.StringAttribute{
				Computed:    true,
				Description: "The export name part.",
			},
			"file_system_name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the file system being exported.",
			},
			"server_name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the server the file system is exported to.",
			},
			"share_policy_name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the SMB share policy applied to the export.",
			},
			"enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the export is enabled.",
			},
			"policy_type": schema.StringAttribute{
				Computed:    true,
				Description: "The policy type ('nfs' or 'smb').",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The status of the export.",
			},
		},
	}
}

// Configure injects the FlashBladeClient into the data source.
func (d *fileSystemExportDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read fetches file system export data by name and populates state.
func (d *fileSystemExportDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config fileSystemExportDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := config.Name.ValueString()
	export, err := d.client.GetFileSystemExport(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"File system export not found",
				fmt.Sprintf("No file system export with name %q exists on the FlashBlade array.", name),
			)
			return
		}
		resp.Diagnostics.AddError("Error reading file system export", err.Error())
		return
	}

	config.ID = types.StringValue(export.ID)
	config.Name = types.StringValue(export.Name)
	config.ExportName = types.StringValue(export.ExportName)
	config.Enabled = types.BoolValue(export.Enabled)
	config.PolicyType = types.StringValue(export.PolicyType)
	config.Status = types.StringValue(export.Status)

	if export.Member != nil {
		config.FileSystemName = types.StringValue(export.Member.Name)
	} else {
		config.FileSystemName = types.StringValue("")
	}
	if export.Server != nil {
		config.ServerName = types.StringValue(export.Server.Name)
	} else {
		config.ServerName = types.StringValue("")
	}
	if export.SharePolicy != nil {
		config.SharePolicyName = types.StringValue(export.SharePolicy.Name)
	} else {
		config.SharePolicyName = types.StringValue("")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
