package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

var _ datasource.DataSource = &auditObjectStorePolicyDataSource{}
var _ datasource.DataSourceWithConfigure = &auditObjectStorePolicyDataSource{}

// auditObjectStorePolicyDataSource implements the flashblade_audit_object_store_policy data source.
type auditObjectStorePolicyDataSource struct {
	client *client.FlashBladeClient
}

func NewAuditObjectStorePolicyDataSource() datasource.DataSource {
	return &auditObjectStorePolicyDataSource{}
}

// ---------- model structs ----------------------------------------------------

// auditObjectStorePolicyDataSourceModel is the top-level model for the data source.
type auditObjectStorePolicyDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Enabled    types.Bool   `tfsdk:"enabled"`
	IsLocal    types.Bool   `tfsdk:"is_local"`
	PolicyType types.String `tfsdk:"policy_type"`
	LogTargets types.List   `tfsdk:"log_targets"`
}

// ---------- data source interface methods -----------------------------------

func (d *auditObjectStorePolicyDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "flashblade_audit_object_store_policy"
}

// Schema defines the data source schema.
func (d *auditObjectStorePolicyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing FlashBlade audit object store policy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the audit object store policy.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the audit object store policy to look up.",
			},
			"enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the audit object store policy is enabled.",
			},
			"is_local": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the policy is defined on the local array.",
			},
			"policy_type": schema.StringAttribute{
				Computed:    true,
				Description: "The type of the policy (e.g. 'audit').",
			},
			"log_targets": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of log target names configured for this policy.",
			},
		},
	}
}

// Configure injects the FlashBladeClient into the data source.
func (d *auditObjectStorePolicyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read fetches audit object store policy data and populates state.
func (d *auditObjectStorePolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config auditObjectStorePolicyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := config.Name.ValueString()
	policy, err := d.client.GetAuditObjectStorePolicy(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Audit object store policy not found",
				fmt.Sprintf("No audit object store policy with name %q exists on the FlashBlade array.", name),
			)
			return
		}
		resp.Diagnostics.AddError("Error reading audit object store policy", err.Error())
		return
	}

	config.ID = types.StringValue(policy.ID)
	config.Name = types.StringValue(policy.Name)
	config.Enabled = types.BoolValue(policy.Enabled)
	config.IsLocal = types.BoolValue(policy.IsLocal)
	config.PolicyType = types.StringValue(policy.PolicyType)

	config.LogTargets = namedRefsToListValue(policy.LogTargets)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
