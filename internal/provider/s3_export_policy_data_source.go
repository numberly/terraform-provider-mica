package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// Ensure s3ExportPolicyDataSource satisfies the datasource interfaces.
var _ datasource.DataSource = &s3ExportPolicyDataSource{}
var _ datasource.DataSourceWithConfigure = &s3ExportPolicyDataSource{}

// s3ExportPolicyDataSource implements the flashblade_s3_export_policy data source.
type s3ExportPolicyDataSource struct {
	client *client.FlashBladeClient
}

// NewS3ExportPolicyDataSource is the factory function registered in the provider.
func NewS3ExportPolicyDataSource() datasource.DataSource {
	return &s3ExportPolicyDataSource{}
}

// ---------- model structs ----------------------------------------------------

// s3ExportPolicyDataSourceModel is the top-level model for the flashblade_s3_export_policy data source.
type s3ExportPolicyDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Enabled    types.Bool   `tfsdk:"enabled"`
	IsLocal    types.Bool   `tfsdk:"is_local"`
	PolicyType types.String `tfsdk:"policy_type"`
	Version    types.String `tfsdk:"version"`
}

// ---------- data source interface methods -----------------------------------

func (d *s3ExportPolicyDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "flashblade_s3_export_policy"
}

// Schema defines the data source schema.
func (d *s3ExportPolicyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing FlashBlade S3 export policy by name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the S3 export policy.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the S3 export policy to look up.",
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
				Description: "The type of the policy (e.g. 's3-export').",
			},
			"version": schema.StringAttribute{
				Computed:    true,
				Description: "The version token that changes on each policy update.",
			},
		},
	}
}

// Configure injects the FlashBladeClient into the data source.
func (d *s3ExportPolicyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read fetches S3 export policy data by name and populates state.
func (d *s3ExportPolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config s3ExportPolicyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := config.Name.ValueString()
	policy, err := d.client.GetS3ExportPolicy(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"S3 export policy not found",
				fmt.Sprintf("No S3 export policy with name %q exists on the FlashBlade array.", name),
			)
			return
		}
		resp.Diagnostics.AddError("Error reading S3 export policy", err.Error())
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
