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

var _ datasource.DataSource = &workloadDataSource{}
var _ datasource.DataSourceWithConfigure = &workloadDataSource{}

// workloadDataSource implements the flashblade_workload data source.
type workloadDataSource struct {
	client *client.FlashBladeClient
}

// workloadDataSourceModel is the top-level model for the flashblade_workload data source.
type workloadDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	PresetName    types.String `tfsdk:"preset_name"`
	Context       types.Object `tfsdk:"context"`
	Created       types.Int64  `tfsdk:"created"`
	Destroyed     types.Bool   `tfsdk:"destroyed"`
	Status        types.String `tfsdk:"status"`
	StatusDetails types.List   `tfsdk:"status_details"`
	TimeRemaining types.Int64  `tfsdk:"time_remaining"`
}

// NewWorkloadDataSource creates a new workloadDataSource provider factory.
func NewWorkloadDataSource() datasource.DataSource {
	return &workloadDataSource{}
}

func (d *workloadDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "flashblade_workload"
}

func (d *workloadDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing FlashBlade workload by name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the workload.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the workload to look up.",
			},
			"preset_name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the preset this workload was deployed from.",
			},
			"context": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "The fleet context that owns this workload (read-only, API-managed).",
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed:    true,
						Description: "The context unique identifier.",
					},
					"name": schema.StringAttribute{
						Computed:    true,
						Description: "The context name.",
					},
				},
			},
			"created": schema.Int64Attribute{
				Computed:    true,
				Description: "The workload creation time, measured in milliseconds since the UNIX epoch.",
			},
			"destroyed": schema.BoolAttribute{
				Computed:    true,
				Description: "True if the workload has been soft-deleted and is pending eradication.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The workload status (e.g. creating, ready, destroying, destroyed, eradicating, recovering).",
			},
			"status_details": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Additional information about the workload status.",
			},
			"time_remaining": schema.Int64Attribute{
				Computed:    true,
				Description: "Time remaining in milliseconds before the destroyed workload is permanently eradicated.",
			},
		},
	}
}

func (d *workloadDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *workloadDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config workloadDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := config.Name.ValueString()
	wl, err := d.client.GetWorkload(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Workload not found",
				fmt.Sprintf("No workload with name %q exists on the FlashBlade array.", name),
			)
			return
		}
		resp.Diagnostics.AddError("Error reading workload", err.Error())
		return
	}

	config.ID = types.StringValue(wl.ID)
	config.Name = types.StringValue(wl.Name)
	config.Created = types.Int64Value(wl.Created)
	config.Destroyed = types.BoolValue(wl.Destroyed)
	config.Status = types.StringValue(wl.Status)
	config.TimeRemaining = types.Int64Value(wl.TimeRemaining)

	if wl.Preset != nil && wl.Preset.Name != "" {
		config.PresetName = types.StringValue(wl.Preset.Name)
	} else {
		config.PresetName = types.StringNull()
	}

	// StatusDetails: always a list.
	if wl.StatusDetails != nil {
		elems := make([]attr.Value, len(wl.StatusDetails))
		for i, s := range wl.StatusDetails {
			elems[i] = types.StringValue(s)
		}
		statusDetails, diags := types.ListValue(types.StringType, elems)
		resp.Diagnostics.Append(diags...)
		config.StatusDetails = statusDetails
	} else {
		config.StatusDetails = types.ListValueMust(types.StringType, []attr.Value{})
	}

	// Context: optional reference.
	if wl.Context != nil {
		ctxObj, diags := types.ObjectValue(workloadContextAttrTypes(), map[string]attr.Value{
			"id":   types.StringValue(wl.Context.ID),
			"name": types.StringValue(wl.Context.Name),
		})
		resp.Diagnostics.Append(diags...)
		config.Context = ctxObj
	} else {
		config.Context = types.ObjectNull(workloadContextAttrTypes())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
