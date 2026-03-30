package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// Ensure qosPolicyDataSource satisfies the datasource interfaces.
var _ datasource.DataSource = &qosPolicyDataSource{}
var _ datasource.DataSourceWithConfigure = &qosPolicyDataSource{}

// qosPolicyDataSource implements the flashblade_qos_policy data source.
type qosPolicyDataSource struct {
	client *client.FlashBladeClient
}

// NewQosPolicyDataSource is the factory function registered in the provider.
func NewQosPolicyDataSource() datasource.DataSource {
	return &qosPolicyDataSource{}
}

// ---------- model structs ----------------------------------------------------

// qosPolicyDataSourceModel is the model for the flashblade_qos_policy data source.
type qosPolicyDataSourceModel struct {
	Name                types.String `tfsdk:"name"`
	ID                  types.String `tfsdk:"id"`
	Enabled             types.Bool   `tfsdk:"enabled"`
	MaxTotalBytesPerSec types.Int64  `tfsdk:"max_total_bytes_per_sec"`
	MaxTotalOpsPerSec   types.Int64  `tfsdk:"max_total_ops_per_sec"`
	IsLocal             types.Bool   `tfsdk:"is_local"`
	PolicyType          types.String `tfsdk:"policy_type"`
}

// ---------- data source interface methods -----------------------------------

// Metadata sets the Terraform type name.
func (d *qosPolicyDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "flashblade_qos_policy"
}

// Schema defines the data source schema.
func (d *qosPolicyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing FlashBlade QoS policy by name.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the QoS policy.",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the QoS policy.",
			},
			"enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the QoS policy is enabled.",
			},
			"max_total_bytes_per_sec": schema.Int64Attribute{
				Computed:    true,
				Description: "Maximum total bandwidth in bytes per second.",
			},
			"max_total_ops_per_sec": schema.Int64Attribute{
				Computed:    true,
				Description: "Maximum total operations (IOPS) per second.",
			},
			"is_local": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the QoS policy is local to this array.",
			},
			"policy_type": schema.StringAttribute{
				Computed:    true,
				Description: "The type of the QoS policy (e.g. bandwidth-limit).",
			},
		},
	}
}

// Configure injects the FlashBladeClient into the data source.
func (d *qosPolicyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read fetches a QoS policy by name and populates state.
func (d *qosPolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config qosPolicyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy, err := d.client.GetQosPolicy(ctx, config.Name.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"QoS policy not found",
				fmt.Sprintf("No QoS policy %q exists on the FlashBlade array.", config.Name.ValueString()),
			)
			return
		}
		resp.Diagnostics.AddError("Error reading QoS policy", err.Error())
		return
	}

	config.ID = types.StringValue(policy.ID)
	config.Name = types.StringValue(policy.Name)
	config.Enabled = types.BoolValue(policy.Enabled)
	config.IsLocal = types.BoolValue(policy.IsLocal)
	config.PolicyType = types.StringValue(policy.PolicyType)

	if policy.MaxTotalBytesPerSec != 0 {
		config.MaxTotalBytesPerSec = types.Int64Value(policy.MaxTotalBytesPerSec)
	} else {
		config.MaxTotalBytesPerSec = types.Int64Null()
	}

	if policy.MaxTotalOpsPerSec != 0 {
		config.MaxTotalOpsPerSec = types.Int64Value(policy.MaxTotalOpsPerSec)
	} else {
		config.MaxTotalOpsPerSec = types.Int64Null()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
