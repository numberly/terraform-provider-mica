package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

var _ datasource.DataSource = &bucketCorsPolicyDataSource{}
var _ datasource.DataSourceWithConfigure = &bucketCorsPolicyDataSource{}

// bucketCorsPolicyDataSource implements the flashblade_bucket_cors_policy data source.
type bucketCorsPolicyDataSource struct {
	client *client.FlashBladeClient
}

func NewBucketCorsPolicyDataSource() datasource.DataSource {
	return &bucketCorsPolicyDataSource{}
}

// bucketCorsPolicyDataSourceModel is the model for the flashblade_bucket_cors_policy data source.
type bucketCorsPolicyDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	BucketName types.String `tfsdk:"bucket_name"`
	IsLocal    types.Bool   `tfsdk:"is_local"`
	PolicyType types.String `tfsdk:"policy_type"`
}

func (d *bucketCorsPolicyDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "flashblade_bucket_cors_policy"
}

func (d *bucketCorsPolicyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing FlashBlade bucket CORS policy by bucket name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the CORS policy.",
			},
			"bucket_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the bucket.",
			},
			"is_local": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the CORS policy is local to this array.",
			},
			"policy_type": schema.StringAttribute{
				Computed:    true,
				Description: "The policy type.",
			},
		},
	}
}

func (d *bucketCorsPolicyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *bucketCorsPolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config bucketCorsPolicyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy, err := d.client.GetCorsPolicy(ctx, config.BucketName.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Bucket CORS policy not found",
				fmt.Sprintf("No CORS policy for bucket %q exists on the FlashBlade array.", config.BucketName.ValueString()),
			)
			return
		}
		resp.Diagnostics.AddError("Error reading bucket CORS policy", err.Error())
		return
	}

	config.ID = types.StringValue(policy.ID)
	config.BucketName = types.StringValue(policy.Bucket.Name)
	config.IsLocal = types.BoolValue(policy.IsLocal)
	config.PolicyType = types.StringValue(policy.PolicyType)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
