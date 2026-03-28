package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// Ensure bucketDataSource satisfies the datasource interfaces.
var _ datasource.DataSource = &bucketDataSource{}
var _ datasource.DataSourceWithConfigure = &bucketDataSource{}

// bucketDataSource implements the flashblade_bucket data source.
type bucketDataSource struct {
	client *client.FlashBladeClient
}

// NewBucketDataSource is the factory function registered in the provider.
func NewBucketDataSource() datasource.DataSource {
	return &bucketDataSource{}
}

// ---------- model structs ----------------------------------------------------

// bucketDataSourceModel is the top-level model for the flashblade_bucket data source.
type bucketDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Account          types.String `tfsdk:"account"`
	Created          types.Int64  `tfsdk:"created"`
	Destroyed        types.Bool   `tfsdk:"destroyed"`
	TimeRemaining    types.Int64  `tfsdk:"time_remaining"`
	Versioning       types.String `tfsdk:"versioning"`
	QuotaLimit       types.Int64  `tfsdk:"quota_limit"`
	HardLimitEnabled types.Bool   `tfsdk:"hard_limit_enabled"`
	ObjectCount      types.Int64  `tfsdk:"object_count"`
	BucketType       types.String `tfsdk:"bucket_type"`
	RetentionLock    types.String `tfsdk:"retention_lock"`
	Space            types.Object `tfsdk:"space"`
}

// ---------- data source interface methods -----------------------------------

// Metadata sets the Terraform type name.
func (d *bucketDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "flashblade_bucket"
}

// Schema defines the data source schema.
func (d *bucketDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing FlashBlade object store bucket by name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the bucket.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the bucket to look up.",
			},
			"account": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the object store account that owns this bucket.",
			},
			"created": schema.Int64Attribute{
				Computed:    true,
				Description: "Unix timestamp (milliseconds) when the bucket was created.",
			},
			"destroyed": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the bucket is soft-deleted.",
			},
			"time_remaining": schema.Int64Attribute{
				Computed:    true,
				Description: "Milliseconds remaining until auto-eradication of a soft-deleted bucket.",
			},
			"versioning": schema.StringAttribute{
				Computed:    true,
				Description: "The bucket versioning state ('none', 'enabled', or 'suspended').",
			},
			"quota_limit": schema.Int64Attribute{
				Computed:    true,
				Description: "The effective quota limit applied against the size of the bucket, in bytes.",
			},
			"hard_limit_enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "If true, the bucket's size cannot exceed the quota limit.",
			},
			"object_count": schema.Int64Attribute{
				Computed:    true,
				Description: "The count of objects in the bucket.",
			},
			"bucket_type": schema.StringAttribute{
				Computed:    true,
				Description: "The bucket type.",
			},
			"retention_lock": schema.StringAttribute{
				Computed:    true,
				Description: "The retention lock mode for the bucket.",
			},
			"space": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Storage space breakdown.",
				Attributes: map[string]schema.Attribute{
					"data_reduction": schema.Float64Attribute{
						Computed:    true,
						Description: "Data reduction ratio.",
					},
					"snapshots": schema.Int64Attribute{
						Computed:    true,
						Description: "Physical space used by snapshots in bytes.",
					},
					"total_physical": schema.Int64Attribute{
						Computed:    true,
						Description: "Total physical space used in bytes.",
					},
					"unique": schema.Int64Attribute{
						Computed:    true,
						Description: "Unique physical space used in bytes.",
					},
					"virtual": schema.Int64Attribute{
						Computed:    true,
						Description: "Virtual (logical) space used in bytes.",
					},
					"snapshots_effective": schema.Int64Attribute{
						Computed:    true,
						Description: "Effective snapshot space used in bytes.",
					},
				},
			},
		},
	}
}

// Configure injects the FlashBladeClient into the data source.
func (d *bucketDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read fetches bucket data by name and populates state.
func (d *bucketDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config bucketDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := config.Name.ValueString()
	bkt, err := d.client.GetBucket(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Bucket not found",
				fmt.Sprintf("No bucket with name %q exists on the FlashBlade array.", name),
			)
			return
		}
		resp.Diagnostics.AddError("Error reading bucket", err.Error())
		return
	}

	config.ID = types.StringValue(bkt.ID)
	config.Name = types.StringValue(bkt.Name)
	config.Account = types.StringValue(bkt.Account.Name)
	config.Created = types.Int64Value(bkt.Created)
	config.Destroyed = types.BoolValue(bkt.Destroyed)
	config.TimeRemaining = types.Int64Value(bkt.TimeRemaining)
	config.Versioning = types.StringValue(bkt.Versioning)
	config.QuotaLimit = types.Int64Value(bkt.QuotaLimit)
	config.HardLimitEnabled = types.BoolValue(bkt.HardLimitEnabled)
	config.ObjectCount = types.Int64Value(bkt.ObjectCount)
	config.BucketType = types.StringValue(bkt.BucketType)
	config.RetentionLock = types.StringValue(bkt.RetentionLock)

	spaceAttrTypes := map[string]attr.Type{
		"data_reduction":      types.Float64Type,
		"snapshots":           types.Int64Type,
		"total_physical":      types.Int64Type,
		"unique":              types.Int64Type,
		"virtual":             types.Int64Type,
		"snapshots_effective": types.Int64Type,
	}
	spaceObj, spaceDiags := types.ObjectValue(spaceAttrTypes, map[string]attr.Value{
		"data_reduction":      types.Float64Value(bkt.Space.DataReduction),
		"snapshots":           types.Int64Value(bkt.Space.Snapshots),
		"total_physical":      types.Int64Value(bkt.Space.TotalPhysical),
		"unique":              types.Int64Value(bkt.Space.Unique),
		"virtual":             types.Int64Value(bkt.Space.Virtual),
		"snapshots_effective": types.Int64Value(bkt.Space.SnapshotsEffective),
	})
	resp.Diagnostics.Append(spaceDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	config.Space = spaceObj

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
