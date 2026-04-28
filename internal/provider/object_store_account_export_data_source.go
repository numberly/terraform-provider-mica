package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

var _ datasource.DataSource = &objectStoreAccountExportDataSource{}
var _ datasource.DataSourceWithConfigure = &objectStoreAccountExportDataSource{}

// objectStoreAccountExportDataSource implements the flashblade_object_store_account_export data source.
type objectStoreAccountExportDataSource struct {
	client *client.FlashBladeClient
}

func NewObjectStoreAccountExportDataSource() datasource.DataSource {
	return &objectStoreAccountExportDataSource{}
}

// ---------- model structs ----------------------------------------------------

// objectStoreAccountExportDataSourceModel is the top-level model for the flashblade_object_store_account_export data source.
type objectStoreAccountExportDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	AccountName types.String `tfsdk:"account_name"`
	ServerName  types.String `tfsdk:"server_name"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	PolicyName  types.String `tfsdk:"policy_name"`
}

// ---------- data source interface methods -----------------------------------

func (d *objectStoreAccountExportDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "flashblade_object_store_account_export"
}

// Schema defines the data source schema.
func (d *objectStoreAccountExportDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing FlashBlade object store account export by name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the object store account export.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The combined name of the export to look up (e.g. 'account/export_name').",
			},
			"account_name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the object store account being exported.",
			},
			"server_name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the server the account is exported to.",
			},
			"enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the export is enabled.",
			},
			"policy_name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the S3 export policy applied to the export.",
			},
		},
	}
}

// Configure injects the FlashBladeClient into the data source.
func (d *objectStoreAccountExportDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read fetches object store account export data by name and populates state.
func (d *objectStoreAccountExportDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config objectStoreAccountExportDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := config.Name.ValueString()
	export, err := d.client.GetObjectStoreAccountExport(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Object store account export not found",
				fmt.Sprintf("No object store account export with name %q exists on the FlashBlade array.", name),
			)
			return
		}
		resp.Diagnostics.AddError("Error reading object store account export", err.Error())
		return
	}

	config.ID = types.StringValue(export.ID)
	config.Name = types.StringValue(export.Name)
	config.Enabled = types.BoolValue(export.Enabled)

	if export.Member != nil {
		config.AccountName = types.StringValue(export.Member.Name)
	} else {
		config.AccountName = types.StringValue("")
	}
	if export.Server != nil {
		config.ServerName = types.StringValue(export.Server.Name)
	} else {
		config.ServerName = types.StringValue("")
	}
	if export.Policy != nil {
		config.PolicyName = types.StringValue(export.Policy.Name)
	} else {
		config.PolicyName = types.StringValue("")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
