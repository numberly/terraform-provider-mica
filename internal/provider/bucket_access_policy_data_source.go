package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

var _ datasource.DataSource = &bucketAccessPolicyDataSource{}
var _ datasource.DataSourceWithConfigure = &bucketAccessPolicyDataSource{}

// bucketAccessPolicyDataSource implements the flashblade_bucket_access_policy data source.
type bucketAccessPolicyDataSource struct {
	client *client.FlashBladeClient
}

func NewBucketAccessPolicyDataSource() datasource.DataSource {
	return &bucketAccessPolicyDataSource{}
}

// ---------- model structs ----------------------------------------------------

// bucketAccessPolicyDataSourceModel is the model for the flashblade_bucket_access_policy data source.
type bucketAccessPolicyDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	BucketName types.String `tfsdk:"bucket_name"`
	Enabled    types.Bool   `tfsdk:"enabled"`
	RuleCount  types.Int64  `tfsdk:"rule_count"`
}

// ---------- data source interface methods -----------------------------------

func (d *bucketAccessPolicyDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "flashblade_bucket_access_policy"
}

// Schema defines the data source schema.
func (d *bucketAccessPolicyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing FlashBlade bucket access policy by bucket name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the bucket access policy.",
			},
			"bucket_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the bucket.",
			},
			"enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the bucket access policy is enabled.",
			},
			"rule_count": schema.Int64Attribute{
				Computed:    true,
				Description: "The number of rules on the bucket access policy.",
			},
		},
	}
}

// Configure injects the FlashBladeClient into the data source.
func (d *bucketAccessPolicyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read fetches a bucket access policy by bucket name and populates state.
func (d *bucketAccessPolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config bucketAccessPolicyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy, err := d.client.GetBucketAccessPolicy(ctx, config.BucketName.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Bucket access policy not found",
				fmt.Sprintf("No bucket access policy for bucket %q exists on the FlashBlade array.", config.BucketName.ValueString()),
			)
			return
		}
		resp.Diagnostics.AddError("Error reading bucket access policy", err.Error())
		return
	}

	config.ID = types.StringValue(policy.ID)
	config.BucketName = types.StringValue(policy.Bucket.Name)
	config.Enabled = types.BoolValue(policy.Enabled)
	config.RuleCount = types.Int64Value(int64(len(policy.Rules)))

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
