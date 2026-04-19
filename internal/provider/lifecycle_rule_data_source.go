package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

var _ datasource.DataSource = &lifecycleRuleDataSource{}
var _ datasource.DataSourceWithConfigure = &lifecycleRuleDataSource{}

// lifecycleRuleDataSource implements the flashblade_lifecycle_rule data source.
type lifecycleRuleDataSource struct {
	client *client.FlashBladeClient
}

func NewLifecycleRuleDataSource() datasource.DataSource {
	return &lifecycleRuleDataSource{}
}

// ---------- model structs ----------------------------------------------------

// lifecycleRuleDataSourceModel is the model for the flashblade_lifecycle_rule data source.
type lifecycleRuleDataSourceModel struct {
	ID                                   types.String `tfsdk:"id"`
	BucketName                           types.String `tfsdk:"bucket_name"`
	RuleID                               types.String `tfsdk:"rule_id"`
	Prefix                               types.String `tfsdk:"prefix"`
	Enabled                              types.Bool   `tfsdk:"enabled"`
	AbortIncompleteMultipartUploadsAfter types.Int64  `tfsdk:"abort_incomplete_multipart_uploads_after"`
	KeepCurrentVersionFor                types.Int64  `tfsdk:"keep_current_version_for"`
	KeepCurrentVersionUntil              types.Int64  `tfsdk:"keep_current_version_until"`
	KeepPreviousVersionFor               types.Int64  `tfsdk:"keep_previous_version_for"`
	CleanupExpiredObjectDeleteMarker     types.Bool   `tfsdk:"cleanup_expired_object_delete_marker"`
}

// ---------- data source interface methods -----------------------------------

func (d *lifecycleRuleDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "flashblade_lifecycle_rule"
}

// Schema defines the data source schema.
func (d *lifecycleRuleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing FlashBlade bucket lifecycle rule by bucket name and rule ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the lifecycle rule.",
			},
			"bucket_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the bucket.",
			},
			"rule_id": schema.StringAttribute{
				Required:    true,
				Description: "The rule identifier within the bucket.",
			},
			"prefix": schema.StringAttribute{
				Computed:    true,
				Description: "Object key prefix filter for the rule.",
			},
			"enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the lifecycle rule is enabled.",
			},
			"abort_incomplete_multipart_uploads_after": schema.Int64Attribute{
				Computed:    true,
				Description: "Duration in milliseconds after which incomplete multipart uploads are aborted.",
			},
			"keep_current_version_for": schema.Int64Attribute{
				Computed:    true,
				Description: "Duration in milliseconds to keep current object versions before expiration.",
			},
			"keep_current_version_until": schema.Int64Attribute{
				Computed:    true,
				Description: "Timestamp in milliseconds until which current object versions are kept.",
			},
			"keep_previous_version_for": schema.Int64Attribute{
				Computed:    true,
				Description: "Duration in milliseconds to keep previous object versions before expiration.",
			},
			"cleanup_expired_object_delete_marker": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether expired object delete markers are cleaned up.",
			},
		},
	}
}

// Configure injects the FlashBladeClient into the data source.
func (d *lifecycleRuleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read fetches a lifecycle rule by bucket name and rule ID and populates state.
func (d *lifecycleRuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config lifecycleRuleDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule, err := d.client.GetLifecycleRule(ctx, config.BucketName.ValueString(), config.RuleID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Lifecycle rule not found",
				fmt.Sprintf("No lifecycle rule %q on bucket %q exists on the FlashBlade array.", config.RuleID.ValueString(), config.BucketName.ValueString()),
			)
			return
		}
		resp.Diagnostics.AddError("Error reading lifecycle rule", err.Error())
		return
	}

	config.ID = types.StringValue(rule.ID)
	config.BucketName = types.StringValue(rule.Bucket.Name)
	config.RuleID = types.StringValue(rule.RuleID)
	config.Prefix = types.StringValue(rule.Prefix)
	config.Enabled = types.BoolValue(rule.Enabled)
	config.CleanupExpiredObjectDeleteMarker = types.BoolValue(rule.CleanupExpiredObjectDeleteMarker)

	if rule.AbortIncompleteMultipartUploadsAfter != nil {
		config.AbortIncompleteMultipartUploadsAfter = types.Int64Value(*rule.AbortIncompleteMultipartUploadsAfter)
	} else {
		config.AbortIncompleteMultipartUploadsAfter = types.Int64Null()
	}

	if rule.KeepCurrentVersionFor != nil {
		config.KeepCurrentVersionFor = types.Int64Value(*rule.KeepCurrentVersionFor)
	} else {
		config.KeepCurrentVersionFor = types.Int64Null()
	}

	if rule.KeepCurrentVersionUntil != nil {
		config.KeepCurrentVersionUntil = types.Int64Value(*rule.KeepCurrentVersionUntil)
	} else {
		config.KeepCurrentVersionUntil = types.Int64Null()
	}

	if rule.KeepPreviousVersionFor != nil {
		config.KeepPreviousVersionFor = types.Int64Value(*rule.KeepPreviousVersionFor)
	} else {
		config.KeepPreviousVersionFor = types.Int64Null()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
