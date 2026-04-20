package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

var _ datasource.DataSource = &arrayNtpDataSource{}
var _ datasource.DataSourceWithConfigure = &arrayNtpDataSource{}

// arrayNtpDataSource implements the flashblade_array_ntp data source.
type arrayNtpDataSource struct {
	client *client.FlashBladeClient
}

func NewArrayNtpDataSource() datasource.DataSource {
	return &arrayNtpDataSource{}
}

// ---------- model structs ----------------------------------------------------

// arrayNtpDataSourceModel is the top-level model for the flashblade_array_ntp data source.
type arrayNtpDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	NtpServers types.List   `tfsdk:"ntp_servers"`
}

// ---------- data source interface methods -----------------------------------

func (d *arrayNtpDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "flashblade_array_ntp"
}

// Schema defines the data source schema.
func (d *arrayNtpDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads the current NTP server configuration of a FlashBlade array.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the array.",
			},
			"ntp_servers": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of NTP server hostnames or IP addresses.",
			},
		},
	}
}

// Configure injects the FlashBladeClient into the data source.
func (d *arrayNtpDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read fetches the NTP configuration and populates state.
func (d *arrayNtpDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config arrayNtpDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	arrayInfo, err := d.client.GetArrayNtp(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading NTP configuration", err.Error())
		return
	}

	config.ID = types.StringValue(arrayInfo.ID)

	if len(arrayInfo.NtpServers) > 0 {
		servers, diags := types.ListValueFrom(ctx, types.StringType, arrayInfo.NtpServers)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		config.NtpServers = servers
	} else {
		config.NtpServers = emptyStringList()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
