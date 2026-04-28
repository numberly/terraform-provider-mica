package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

var _ datasource.DataSource = &targetDataSource{}
var _ datasource.DataSourceWithConfigure = &targetDataSource{}

// targetDataSource implements the flashblade_target data source.
type targetDataSource struct {
	client *client.FlashBladeClient
}

func NewTargetDataSource() datasource.DataSource {
	return &targetDataSource{}
}

// ---------- model structs ----------------------------------------------------

// targetDataSourceModel is the top-level model for the flashblade_target data source.
type targetDataSourceModel struct {
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Address            types.String `tfsdk:"address"`
	CACertificateGroup types.String `tfsdk:"ca_certificate_group"`
	Status             types.String `tfsdk:"status"`
	StatusDetails      types.String `tfsdk:"status_details"`
}

// ---------- data source interface methods -----------------------------------

func (d *targetDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "flashblade_target"
}

// Schema defines the data source schema.
func (d *targetDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing FlashBlade replication target by name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the target.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the target to look up.",
			},
			"address": schema.StringAttribute{
				Computed:    true,
				Description: "The hostname or IP address of the target S3 endpoint.",
			},
			"ca_certificate_group": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the CA certificate group used to validate the target's TLS certificate. Null when not set.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The connection status of the target (e.g. connected, connecting, error).",
			},
			"status_details": schema.StringAttribute{
				Computed:    true,
				Description: "Additional details about the connection status.",
			},
		},
	}
}

// Configure injects the FlashBladeClient into the data source.
func (d *targetDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read fetches target data by name and populates state.
func (d *targetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config targetDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := config.Name.ValueString()
	tgt, err := d.client.GetTarget(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Target not found",
				fmt.Sprintf("No target with name %q exists on the FlashBlade array.", name),
			)
			return
		}
		resp.Diagnostics.AddError("Error reading target", err.Error())
		return
	}

	config.ID = types.StringValue(tgt.ID)
	config.Name = types.StringValue(tgt.Name)
	config.Address = types.StringValue(tgt.Address)
	config.Status = types.StringValue(tgt.Status)
	config.StatusDetails = types.StringValue(tgt.StatusDetails)

	if tgt.CACertificateGroup != nil {
		config.CACertificateGroup = types.StringValue(tgt.CACertificateGroup.Name)
	} else {
		config.CACertificateGroup = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
