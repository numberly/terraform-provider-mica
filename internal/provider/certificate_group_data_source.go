package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

var _ datasource.DataSource = &certificateGroupDataSource{}
var _ datasource.DataSourceWithConfigure = &certificateGroupDataSource{}

// certificateGroupDataSource implements the flashblade_certificate_group data source.
type certificateGroupDataSource struct {
	client *client.FlashBladeClient
}

func NewCertificateGroupDataSource() datasource.DataSource {
	return &certificateGroupDataSource{}
}

// ---------- model structs ----------------------------------------------------

// certificateGroupDataSourceModel is the top-level model for the flashblade_certificate_group data source.
type certificateGroupDataSourceModel struct {
	ID     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Realms types.List   `tfsdk:"realms"`
}

// ---------- data source interface methods -----------------------------------

func (d *certificateGroupDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "flashblade_certificate_group"
}

// Schema defines the data source schema.
func (d *certificateGroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing FlashBlade certificate group by name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the certificate group.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the certificate group to look up.",
			},
			"realms": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "The list of realms associated with this certificate group.",
			},
		},
	}
}

// Configure injects the FlashBladeClient into the data source.
func (d *certificateGroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read fetches certificate group data by name and populates state.
func (d *certificateGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config certificateGroupDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := config.Name.ValueString()
	group, err := d.client.GetCertificateGroup(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Certificate group not found",
				fmt.Sprintf("No certificate group named %q exists on the FlashBlade array.", name),
			)
			return
		}
		resp.Diagnostics.AddError("Error reading certificate group", err.Error())
		return
	}

	config.ID = types.StringValue(group.ID)
	config.Name = types.StringValue(group.Name)

	if len(group.Realms) == 0 {
		config.Realms = types.ListValueMust(types.StringType, []attr.Value{})
	} else {
		elems := make([]attr.Value, len(group.Realms))
		for i, realm := range group.Realms {
			elems[i] = types.StringValue(realm)
		}
		config.Realms = types.ListValueMust(types.StringType, elems)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
