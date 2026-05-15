package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

var _ datasource.DataSource = &resiliencyGroupDataSource{}
var _ datasource.DataSourceWithConfigure = &resiliencyGroupDataSource{}

// resiliencyGroupDataSource implements the flashblade_resiliency_group data source.
type resiliencyGroupDataSource struct {
	client *client.FlashBladeClient
}

func NewResiliencyGroupDataSource() datasource.DataSource {
	return &resiliencyGroupDataSource{}
}

// ---------- model structs ----------------------------------------------------

// resiliencyGroupDataSourceModel is the top-level model for the
// flashblade_resiliency_group data source.
type resiliencyGroupDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Status        types.String `tfsdk:"status"`
	StatusDetails types.String `tfsdk:"status_details"`
}

// ---------- data source interface methods -----------------------------------

func (d *resiliencyGroupDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "flashblade_resiliency_group"
}

// Schema defines the data source schema.
func (d *resiliencyGroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing FlashBlade resiliency group by name. " +
			"Resiliency groups are hardware-managed high-availability groupings reported by the array (API 2.23+).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The non-modifiable globally unique identifier of the resiliency group.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the resiliency group to look up.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Status of the resiliency group. One of `critical`, `healthy`, `unhealthy`.",
			},
			"status_details": schema.StringAttribute{
				Computed:    true,
				Description: "Human-readable description of the status when the group is not healthy. Empty when healthy.",
			},
		},
	}
}

// Configure injects the FlashBladeClient into the data source.
func (d *resiliencyGroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read fetches resiliency group data by name and populates state.
func (d *resiliencyGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config resiliencyGroupDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := config.Name.ValueString()
	rg, err := d.client.GetResiliencyGroup(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Resiliency group not found",
				fmt.Sprintf("No resiliency group with name %q exists on the FlashBlade array.", name),
			)
			return
		}
		resp.Diagnostics.AddError("Error reading resiliency group", err.Error())
		return
	}

	config.ID = stringOrNull(rg.ID)
	config.Name = types.StringValue(rg.Name)
	config.Status = stringOrNull(rg.Status)
	config.StatusDetails = stringOrNull(rg.StatusDetails)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
