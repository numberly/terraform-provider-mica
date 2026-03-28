package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// Ensure objectStoreAccessPolicyDataSource satisfies the datasource interfaces.
var _ datasource.DataSource = &objectStoreAccessPolicyDataSource{}
var _ datasource.DataSourceWithConfigure = &objectStoreAccessPolicyDataSource{}

// objectStoreAccessPolicyDataSource implements the flashblade_object_store_access_policy data source.
type objectStoreAccessPolicyDataSource struct {
	client *client.FlashBladeClient
}

// NewObjectStoreAccessPolicyDataSource is the factory function registered in the provider.
func NewObjectStoreAccessPolicyDataSource() datasource.DataSource {
	return &objectStoreAccessPolicyDataSource{}
}

// ---------- model structs ----------------------------------------------------

// objectStoreAccessPolicyDataSourceModel is the top-level model for the flashblade_object_store_access_policy data source.
type objectStoreAccessPolicyDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	ARN         types.String `tfsdk:"arn"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	IsLocal     types.Bool   `tfsdk:"is_local"`
	PolicyType  types.String `tfsdk:"policy_type"`
}

// ---------- data source interface methods -----------------------------------

// Metadata sets the Terraform type name.
func (d *objectStoreAccessPolicyDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "flashblade_object_store_access_policy"
}

// Schema defines the data source schema.
func (d *objectStoreAccessPolicyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing FlashBlade object store access policy by name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the object store access policy.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the object store access policy to look up.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "A human-readable description of the policy.",
			},
			"arn": schema.StringAttribute{
				Computed:    true,
				Description: "The Amazon Resource Name (ARN) for the policy.",
			},
			"enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "If true, the policy is enabled.",
			},
			"is_local": schema.BoolAttribute{
				Computed:    true,
				Description: "If true, the policy is local to this array (not replicated).",
			},
			"policy_type": schema.StringAttribute{
				Computed:    true,
				Description: "The type of the policy (e.g. 'object-store-access').",
			},
		},
	}
}

// Configure injects the FlashBladeClient into the data source.
func (d *objectStoreAccessPolicyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read fetches object store access policy data by name and populates state.
func (d *objectStoreAccessPolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config objectStoreAccessPolicyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := config.Name.ValueString()
	policy, err := d.client.GetObjectStoreAccessPolicy(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Object store access policy not found",
				fmt.Sprintf("No object store access policy with name %q exists on the FlashBlade array.", name),
			)
			return
		}
		resp.Diagnostics.AddError("Error reading object store access policy", err.Error())
		return
	}

	config.ID = types.StringValue(policy.ID)
	config.Name = types.StringValue(policy.Name)
	config.ARN = types.StringValue(policy.ARN)
	config.Enabled = types.BoolValue(policy.Enabled)
	config.IsLocal = types.BoolValue(policy.IsLocal)
	config.PolicyType = types.StringValue(policy.PolicyType)

	if policy.Description != "" {
		config.Description = types.StringValue(policy.Description)
	} else {
		config.Description = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
