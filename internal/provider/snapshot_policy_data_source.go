package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

var _ datasource.DataSource = &snapshotPolicyDataSource{}
var _ datasource.DataSourceWithConfigure = &snapshotPolicyDataSource{}

// snapshotPolicyDataSource implements the flashblade_snapshot_policy data source.
type snapshotPolicyDataSource struct {
	client *client.FlashBladeClient
}

func NewSnapshotPolicyDataSource() datasource.DataSource {
	return &snapshotPolicyDataSource{}
}

// ---------- model structs ----------------------------------------------------

// snapshotPolicyDataSourceModel is the top-level model for the flashblade_snapshot_policy data source.
type snapshotPolicyDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	IsLocal       types.Bool   `tfsdk:"is_local"`
	PolicyType    types.String `tfsdk:"policy_type"`
	RetentionLock types.String `tfsdk:"retention_lock"`
}

// ---------- data source interface methods -----------------------------------

func (d *snapshotPolicyDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "flashblade_snapshot_policy"
}

// Schema defines the data source schema.
func (d *snapshotPolicyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing FlashBlade snapshot policy by name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the snapshot policy.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the snapshot policy to look up.",
			},
			"enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "If true, the policy is enabled and its rules are enforced.",
			},
			"is_local": schema.BoolAttribute{
				Computed:    true,
				Description: "If true, the policy is local to this array (not replicated).",
			},
			"policy_type": schema.StringAttribute{
				Computed:    true,
				Description: "The type of the policy (e.g. 'snapshot').",
			},
			"retention_lock": schema.StringAttribute{
				Computed:    true,
				Description: "The retention lock mode of the policy (e.g. 'none', 'ratcheted').",
			},
		},
	}
}

// Configure injects the FlashBladeClient into the data source.
func (d *snapshotPolicyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read fetches snapshot policy data by name and populates state.
func (d *snapshotPolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config snapshotPolicyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := config.Name.ValueString()
	policy, err := d.client.GetSnapshotPolicy(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Snapshot policy not found",
				fmt.Sprintf("No snapshot policy with name %q exists on the FlashBlade array.", name),
			)
			return
		}
		resp.Diagnostics.AddError("Error reading snapshot policy", err.Error())
		return
	}

	config.ID = types.StringValue(policy.ID)
	config.Name = types.StringValue(policy.Name)
	config.Enabled = types.BoolValue(policy.Enabled)
	config.IsLocal = types.BoolValue(policy.IsLocal)
	config.PolicyType = types.StringValue(policy.PolicyType)
	config.RetentionLock = types.StringValue(policy.RetentionLock)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
