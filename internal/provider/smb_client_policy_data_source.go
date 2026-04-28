package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

var _ datasource.DataSource = &smbClientPolicyDataSource{}
var _ datasource.DataSourceWithConfigure = &smbClientPolicyDataSource{}

// smbClientPolicyDataSource implements the flashblade_smb_client_policy data source.
type smbClientPolicyDataSource struct {
	client *client.FlashBladeClient
}

func NewSmbClientPolicyDataSource() datasource.DataSource {
	return &smbClientPolicyDataSource{}
}

// ---------- model structs ----------------------------------------------------

// smbClientPolicyDataSourceModel is the top-level model for the flashblade_smb_client_policy data source.
type smbClientPolicyDataSourceModel struct {
	ID                            types.String `tfsdk:"id"`
	Name                          types.String `tfsdk:"name"`
	Enabled                       types.Bool   `tfsdk:"enabled"`
	IsLocal                       types.Bool   `tfsdk:"is_local"`
	PolicyType                    types.String `tfsdk:"policy_type"`
	Version                       types.String `tfsdk:"version"`
	AccessBasedEnumerationEnabled types.Bool   `tfsdk:"access_based_enumeration_enabled"`
}

// ---------- data source interface methods -----------------------------------

func (d *smbClientPolicyDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "flashblade_smb_client_policy"
}

// Schema defines the data source schema.
func (d *smbClientPolicyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing FlashBlade SMB client policy by name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the SMB client policy.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the SMB client policy to look up.",
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
			"version": schema.StringAttribute{
				Computed:    true,
				Description: "The version of the SMB client policy (read-only, server-assigned).",
			},
			"access_based_enumeration_enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "If true, access-based enumeration is enabled for this policy.",
			},
		},
	}
}

// Configure injects the FlashBladeClient into the data source.
func (d *smbClientPolicyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read fetches SMB client policy data by name and populates state.
func (d *smbClientPolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config smbClientPolicyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := config.Name.ValueString()
	policy, err := d.client.GetSmbClientPolicy(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"SMB client policy not found",
				fmt.Sprintf("No SMB client policy with name %q exists on the FlashBlade array.", name),
			)
			return
		}
		resp.Diagnostics.AddError("Error reading SMB client policy", err.Error())
		return
	}

	config.ID = types.StringValue(policy.ID)
	config.Name = types.StringValue(policy.Name)
	config.Enabled = types.BoolValue(policy.Enabled)
	config.IsLocal = types.BoolValue(policy.IsLocal)
	config.PolicyType = types.StringValue(policy.PolicyType)
	config.Version = types.StringValue(policy.Version)
	config.AccessBasedEnumerationEnabled = types.BoolValue(policy.AccessBasedEnumerationEnabled)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
