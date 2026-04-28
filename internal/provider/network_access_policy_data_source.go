package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

var _ datasource.DataSource = &networkAccessPolicyDataSource{}
var _ datasource.DataSourceWithConfigure = &networkAccessPolicyDataSource{}

// networkAccessPolicyDataSource implements the flashblade_network_access_policy data source.
type networkAccessPolicyDataSource struct {
	client *client.FlashBladeClient
}

func NewNetworkAccessPolicyDataSource() datasource.DataSource {
	return &networkAccessPolicyDataSource{}
}

// ---------- model structs ----------------------------------------------------

// networkAccessPolicyDataSourceModel is the top-level model for the flashblade_network_access_policy data source.
type networkAccessPolicyDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Enabled    types.Bool   `tfsdk:"enabled"`
	IsLocal    types.Bool   `tfsdk:"is_local"`
	PolicyType types.String `tfsdk:"policy_type"`
	Version    types.String `tfsdk:"version"`
}

// ---------- data source interface methods -----------------------------------

func (d *networkAccessPolicyDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "flashblade_network_access_policy"
}

// Schema defines the data source schema.
func (d *networkAccessPolicyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing FlashBlade network access policy by name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the network access policy.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the network access policy to look up.",
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
				Description: "The type of the policy (e.g. 'network-access').",
			},
			"version": schema.StringAttribute{
				Computed:    true,
				Description: "The version token that changes on each policy update.",
			},
		},
	}
}

// Configure injects the FlashBladeClient into the data source.
func (d *networkAccessPolicyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read fetches network access policy data by name and populates state.
func (d *networkAccessPolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config networkAccessPolicyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := config.Name.ValueString()
	policy, err := d.client.GetNetworkAccessPolicy(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Network access policy not found",
				fmt.Sprintf("No network access policy with name %q exists on the FlashBlade array.", name),
			)
			return
		}
		resp.Diagnostics.AddError("Error reading network access policy", err.Error())
		return
	}

	config.ID = types.StringValue(policy.ID)
	config.Name = types.StringValue(policy.Name)
	config.Enabled = types.BoolValue(policy.Enabled)
	config.IsLocal = types.BoolValue(policy.IsLocal)
	config.PolicyType = types.StringValue(policy.PolicyType)
	config.Version = types.StringValue(policy.Version)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
