package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/soulkyu/terraform-provider-flashblade/internal/client"
)

// Ensure smbSharePolicyDataSource satisfies the datasource interfaces.
var _ datasource.DataSource = &smbSharePolicyDataSource{}
var _ datasource.DataSourceWithConfigure = &smbSharePolicyDataSource{}

// smbSharePolicyDataSource implements the flashblade_smb_share_policy data source.
type smbSharePolicyDataSource struct {
	client *client.FlashBladeClient
}

// NewSmbSharePolicyDataSource is the factory function registered in the provider.
func NewSmbSharePolicyDataSource() datasource.DataSource {
	return &smbSharePolicyDataSource{}
}

// ---------- model structs ----------------------------------------------------

// smbSharePolicyDataSourceModel is the top-level model for the flashblade_smb_share_policy data source.
// Note: SMB share policy has no Version field (unlike NFS export policy).
type smbSharePolicyDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Enabled    types.Bool   `tfsdk:"enabled"`
	IsLocal    types.Bool   `tfsdk:"is_local"`
	PolicyType types.String `tfsdk:"policy_type"`
}

// ---------- data source interface methods -----------------------------------

// Metadata sets the Terraform type name.
func (d *smbSharePolicyDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "flashblade_smb_share_policy"
}

// Schema defines the data source schema.
func (d *smbSharePolicyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing FlashBlade SMB share policy by name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the SMB share policy.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the SMB share policy to look up.",
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
				Description: "The type of the policy (e.g. 'smb').",
			},
		},
	}
}

// Configure injects the FlashBladeClient into the data source.
func (d *smbSharePolicyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read fetches SMB share policy data by name and populates state.
func (d *smbSharePolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config smbSharePolicyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := config.Name.ValueString()
	policy, err := d.client.GetSmbSharePolicy(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"SMB share policy not found",
				fmt.Sprintf("No SMB share policy with name %q exists on the FlashBlade array.", name),
			)
			return
		}
		resp.Diagnostics.AddError("Error reading SMB share policy", err.Error())
		return
	}

	config.ID = types.StringValue(policy.ID)
	config.Name = types.StringValue(policy.Name)
	config.Enabled = types.BoolValue(policy.Enabled)
	config.IsLocal = types.BoolValue(policy.IsLocal)
	config.PolicyType = types.StringValue(policy.PolicyType)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
