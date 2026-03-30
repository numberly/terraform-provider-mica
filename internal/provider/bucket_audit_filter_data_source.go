package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// Ensure bucketAuditFilterDataSource satisfies the datasource interfaces.
var _ datasource.DataSource = &bucketAuditFilterDataSource{}
var _ datasource.DataSourceWithConfigure = &bucketAuditFilterDataSource{}

// bucketAuditFilterDataSource implements the flashblade_bucket_audit_filter data source.
type bucketAuditFilterDataSource struct {
	client *client.FlashBladeClient
}

// NewBucketAuditFilterDataSource is the factory function registered in the provider.
func NewBucketAuditFilterDataSource() datasource.DataSource {
	return &bucketAuditFilterDataSource{}
}

// ---------- model structs ----------------------------------------------------

// bucketAuditFilterDataSourceModel is the model for the flashblade_bucket_audit_filter data source.
type bucketAuditFilterDataSourceModel struct {
	BucketName types.String `tfsdk:"bucket_name"`
	Actions    types.List   `tfsdk:"actions"`
	S3Prefixes types.List   `tfsdk:"s3_prefixes"`
}

// ---------- data source interface methods -----------------------------------

// Metadata sets the Terraform type name.
func (d *bucketAuditFilterDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "flashblade_bucket_audit_filter"
}

// Schema defines the data source schema.
func (d *bucketAuditFilterDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing FlashBlade bucket audit filter by bucket name.",
		Attributes: map[string]schema.Attribute{
			"bucket_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the bucket.",
			},
			"actions": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of S3 actions being audited.",
			},
			"s3_prefixes": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of S3 object key prefixes being filtered.",
			},
		},
	}
}

// Configure injects the FlashBladeClient into the data source.
func (d *bucketAuditFilterDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read fetches a bucket audit filter by bucket name and populates state.
func (d *bucketAuditFilterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config bucketAuditFilterDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	filter, err := d.client.GetBucketAuditFilter(ctx, config.BucketName.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Bucket audit filter not found",
				fmt.Sprintf("No bucket audit filter for bucket %q exists on the FlashBlade array.", config.BucketName.ValueString()),
			)
			return
		}
		resp.Diagnostics.AddError("Error reading bucket audit filter", err.Error())
		return
	}

	config.BucketName = types.StringValue(filter.Bucket.Name)
	config.Actions = types.ListValueMust(types.StringType, stringSliceToAttrValues(filter.Actions))

	prefixes := filter.S3Prefixes
	if prefixes == nil {
		prefixes = []string{}
	}
	config.S3Prefixes = types.ListValueMust(types.StringType, stringSliceToAttrValues(prefixes))

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
