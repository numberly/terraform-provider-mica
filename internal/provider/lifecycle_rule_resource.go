package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// Ensure lifecycleRuleResource satisfies the resource interfaces.
var _ resource.Resource = &lifecycleRuleResource{}
var _ resource.ResourceWithConfigure = &lifecycleRuleResource{}
var _ resource.ResourceWithImportState = &lifecycleRuleResource{}
var _ resource.ResourceWithUpgradeState = &lifecycleRuleResource{}

// lifecycleRuleResource implements the flashblade_lifecycle_rule resource.
type lifecycleRuleResource struct {
	client *client.FlashBladeClient
}

// NewLifecycleRuleResource is the factory function registered in the provider.
func NewLifecycleRuleResource() resource.Resource {
	return &lifecycleRuleResource{}
}

// ---------- model structs ----------------------------------------------------

// lifecycleRuleModel is the Terraform state model for the flashblade_lifecycle_rule resource.
type lifecycleRuleModel struct {
	ID                                   types.String   `tfsdk:"id"`
	BucketName                           types.String   `tfsdk:"bucket_name"`
	RuleID                               types.String   `tfsdk:"rule_id"`
	Prefix                               types.String   `tfsdk:"prefix"`
	Enabled                              types.Bool     `tfsdk:"enabled"`
	AbortIncompleteMultipartUploadsAfter types.Int64    `tfsdk:"abort_incomplete_multipart_uploads_after"`
	KeepCurrentVersionFor                types.Int64    `tfsdk:"keep_current_version_for"`
	KeepCurrentVersionUntil              types.Int64    `tfsdk:"keep_current_version_until"`
	KeepPreviousVersionFor               types.Int64    `tfsdk:"keep_previous_version_for"`
	CleanupExpiredObjectDeleteMarker     types.Bool     `tfsdk:"cleanup_expired_object_delete_marker"`
	Timeouts                             timeouts.Value `tfsdk:"timeouts"`
}

// ---------- resource interface methods --------------------------------------

// Metadata sets the Terraform type name.
func (r *lifecycleRuleResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_lifecycle_rule"
}

// Schema defines the resource schema.
func (r *lifecycleRuleResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     0,
		Description: "Manages a FlashBlade bucket lifecycle rule for automatic object expiration and cleanup.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the lifecycle rule.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"bucket_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the bucket this rule belongs to. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"rule_id": schema.StringAttribute{
				Required:    true,
				Description: "The rule identifier within the bucket. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"prefix": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Object key prefix filter for the rule. Defaults to empty string (all objects).",
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether the lifecycle rule is enabled. Defaults to true.",
			},
			"abort_incomplete_multipart_uploads_after": schema.Int64Attribute{
				Optional:    true,
				Description: "Duration in milliseconds after which incomplete multipart uploads are aborted.",
			},
			"keep_current_version_for": schema.Int64Attribute{
				Optional:    true,
				Description: "Duration in milliseconds to keep current object versions before expiration.",
			},
			"keep_current_version_until": schema.Int64Attribute{
				Optional:    true,
				Description: "Timestamp in milliseconds until which current object versions are kept.",
			},
			"keep_previous_version_for": schema.Int64Attribute{
				Optional:    true,
				Description: "Duration in milliseconds to keep previous object versions before expiration.",
			},
			"cleanup_expired_object_delete_marker": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether expired object delete markers are cleaned up. Read-only, managed by the array.",
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

// UpgradeState returns state upgraders for schema migrations.
func (r *lifecycleRuleResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}

// Configure injects the FlashBladeClient into the resource.
func (r *lifecycleRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.client = c
}

// ---------- CRUD methods ----------------------------------------------------

// Create creates a new lifecycle rule.
func (r *lifecycleRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data lifecycleRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := data.Timeouts.Create(ctx, 20*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	body := client.LifecycleRulePost{
		Bucket: client.NamedReference{Name: data.BucketName.ValueString()},
		RuleID: data.RuleID.ValueString(),
		Prefix: data.Prefix.ValueString(),
	}

	if !data.AbortIncompleteMultipartUploadsAfter.IsNull() && !data.AbortIncompleteMultipartUploadsAfter.IsUnknown() {
		v := data.AbortIncompleteMultipartUploadsAfter.ValueInt64()
		body.AbortIncompleteMultipartUploadsAfter = &v
	}
	if !data.KeepCurrentVersionFor.IsNull() && !data.KeepCurrentVersionFor.IsUnknown() {
		v := data.KeepCurrentVersionFor.ValueInt64()
		body.KeepCurrentVersionFor = &v
	}
	if !data.KeepCurrentVersionUntil.IsNull() && !data.KeepCurrentVersionUntil.IsUnknown() {
		v := data.KeepCurrentVersionUntil.ValueInt64()
		body.KeepCurrentVersionUntil = &v
	}
	if !data.KeepPreviousVersionFor.IsNull() && !data.KeepPreviousVersionFor.IsUnknown() {
		v := data.KeepPreviousVersionFor.ValueInt64()
		body.KeepPreviousVersionFor = &v
	}

	rule, err := r.client.PostLifecycleRule(ctx, body, false)
	if err != nil {
		resp.Diagnostics.AddError("Error creating lifecycle rule", err.Error())
		return
	}

	mapLifecycleRuleToModel(rule, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes Terraform state from the API.
func (r *lifecycleRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data lifecycleRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	readTimeout, diags := data.Timeouts.Read(ctx, 5*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, readTimeout)
	defer cancel()

	rule, err := r.client.GetLifecycleRule(ctx, data.BucketName.ValueString(), data.RuleID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading lifecycle rule", err.Error())
		return
	}

	mapLifecycleRuleToModel(rule, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update applies changes to an existing lifecycle rule.
func (r *lifecycleRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state lifecycleRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateTimeout, diags := plan.Timeouts.Update(ctx, 20*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	patch := client.LifecycleRulePatch{}
	needsPatch := false

	if !plan.Enabled.Equal(state.Enabled) {
		v := plan.Enabled.ValueBool()
		patch.Enabled = &v
		needsPatch = true
	}
	if !plan.Prefix.Equal(state.Prefix) {
		v := plan.Prefix.ValueString()
		patch.Prefix = &v
		needsPatch = true
	}
	if !plan.AbortIncompleteMultipartUploadsAfter.Equal(state.AbortIncompleteMultipartUploadsAfter) {
		v := plan.AbortIncompleteMultipartUploadsAfter.ValueInt64()
		patch.AbortIncompleteMultipartUploadsAfter = &v
		needsPatch = true
	}
	if !plan.KeepCurrentVersionFor.Equal(state.KeepCurrentVersionFor) {
		v := plan.KeepCurrentVersionFor.ValueInt64()
		patch.KeepCurrentVersionFor = &v
		needsPatch = true
	}
	if !plan.KeepCurrentVersionUntil.Equal(state.KeepCurrentVersionUntil) {
		v := plan.KeepCurrentVersionUntil.ValueInt64()
		patch.KeepCurrentVersionUntil = &v
		needsPatch = true
	}
	if !plan.KeepPreviousVersionFor.Equal(state.KeepPreviousVersionFor) {
		v := plan.KeepPreviousVersionFor.ValueInt64()
		patch.KeepPreviousVersionFor = &v
		needsPatch = true
	}

	if needsPatch {
		_, err := r.client.PatchLifecycleRule(ctx, state.BucketName.ValueString(), state.RuleID.ValueString(), patch, false)
		if err != nil {
			resp.Diagnostics.AddError("Error updating lifecycle rule", err.Error())
			return
		}
	}

	// Re-read to refresh computed fields.
	rule, err := r.client.GetLifecycleRule(ctx, plan.BucketName.ValueString(), plan.RuleID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading lifecycle rule after update", err.Error())
		return
	}

	mapLifecycleRuleToModel(rule, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes a lifecycle rule.
func (r *lifecycleRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data lifecycleRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, diags := data.Timeouts.Delete(ctx, 30*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	err := r.client.DeleteLifecycleRule(ctx, data.BucketName.ValueString(), data.RuleID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting lifecycle rule", err.Error())
		return
	}
}

// ImportState imports an existing lifecycle rule by composite ID "bucketName/ruleID".
func (r *lifecycleRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := parseCompositeID(req.ID, 2)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("Expected format: bucketName/ruleID. Error: %s", err))
		return
	}

	rule, err := r.client.GetLifecycleRule(ctx, parts[0], parts[1])
	if err != nil {
		resp.Diagnostics.AddError("Error importing lifecycle rule", err.Error())
		return
	}

	var data lifecycleRuleModel
	data.Timeouts = nullTimeoutsValue()

	mapLifecycleRuleToModel(rule, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ---------- helpers ---------------------------------------------------------

// mapLifecycleRuleToModel maps a client.LifecycleRule to the Terraform model.
func mapLifecycleRuleToModel(rule *client.LifecycleRule, data *lifecycleRuleModel) {
	data.ID = types.StringValue(rule.ID)
	data.BucketName = types.StringValue(rule.Bucket.Name)
	data.RuleID = types.StringValue(rule.RuleID)
	data.Prefix = types.StringValue(rule.Prefix)
	data.Enabled = types.BoolValue(rule.Enabled)
	data.CleanupExpiredObjectDeleteMarker = types.BoolValue(rule.CleanupExpiredObjectDeleteMarker)

	if rule.AbortIncompleteMultipartUploadsAfter != nil {
		data.AbortIncompleteMultipartUploadsAfter = types.Int64Value(*rule.AbortIncompleteMultipartUploadsAfter)
	} else {
		data.AbortIncompleteMultipartUploadsAfter = types.Int64Null()
	}

	if rule.KeepCurrentVersionFor != nil {
		data.KeepCurrentVersionFor = types.Int64Value(*rule.KeepCurrentVersionFor)
	} else {
		data.KeepCurrentVersionFor = types.Int64Null()
	}

	if rule.KeepCurrentVersionUntil != nil {
		data.KeepCurrentVersionUntil = types.Int64Value(*rule.KeepCurrentVersionUntil)
	} else {
		data.KeepCurrentVersionUntil = types.Int64Null()
	}

	if rule.KeepPreviousVersionFor != nil {
		data.KeepPreviousVersionFor = types.Int64Value(*rule.KeepPreviousVersionFor)
	} else {
		data.KeepPreviousVersionFor = types.Int64Null()
	}
}
