package provider

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

var _ resource.Resource = &bucketResource{}
var _ resource.ResourceWithConfigure = &bucketResource{}
var _ resource.ResourceWithImportState = &bucketResource{}
var _ resource.ResourceWithValidateConfig = &bucketResource{}
var _ resource.ResourceWithUpgradeState = &bucketResource{}

// bucketResource implements the flashblade_bucket resource.
type bucketResource struct {
	client *client.FlashBladeClient
}

func NewBucketResource() resource.Resource {
	return &bucketResource{}
}

// ---------- model structs ----------------------------------------------------

// bucketModel is the top-level model for the flashblade_bucket resource.
type bucketModel struct {
	ID                       types.String   `tfsdk:"id"`
	Name                     types.String   `tfsdk:"name"`
	Account                  types.String   `tfsdk:"account"`
	Created                  types.Int64    `tfsdk:"created"`
	Destroyed                types.Bool     `tfsdk:"destroyed"`
	DestroyEradicateOnDelete types.Bool     `tfsdk:"destroy_eradicate_on_delete"`
	TimeRemaining            types.Int64    `tfsdk:"time_remaining"`
	Versioning               types.String   `tfsdk:"versioning"`
	QuotaLimit               types.Int64    `tfsdk:"quota_limit"`
	HardLimitEnabled         types.Bool     `tfsdk:"hard_limit_enabled"`
	ObjectCount              types.Int64    `tfsdk:"object_count"`
	BucketType               types.String   `tfsdk:"bucket_type"`
	RetentionLock            types.String   `tfsdk:"retention_lock"`
	EradicationConfig        types.Object   `tfsdk:"eradication_config"`
	ObjectLockConfig         types.Object   `tfsdk:"object_lock_config"`
	PublicAccessConfig       types.Object   `tfsdk:"public_access_config"`
	PublicStatus             types.String   `tfsdk:"public_status"`
	Space                    types.Object   `tfsdk:"space"`
	Timeouts                 timeouts.Value `tfsdk:"timeouts"`
}

// ---------- resource interface methods --------------------------------------

func (r *bucketResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_bucket"
}

// Schema defines the resource schema.
func (r *bucketResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     0,
		Description: "Manages a FlashBlade object store bucket.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the bucket.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the bucket. Changing this forces a new resource (S3 clients hardcode bucket names).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"account": schema.StringAttribute{
				Required:    true,
				Description: "The name of the object store account that owns this bucket. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"created": schema.Int64Attribute{
				Computed:    true,
				Description: "Unix timestamp (milliseconds) when the bucket was created.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"destroyed": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the bucket is soft-deleted.",
			},
			"destroy_eradicate_on_delete": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "When true, Terraform will eradicate the bucket on destroy. When false (default), only soft-deletes. Buckets hold production data — eradication is opt-in.",
			},
			"time_remaining": schema.Int64Attribute{
				Computed:    true,
				Description: "Milliseconds remaining until auto-eradication of a soft-deleted bucket.",
			},
			"versioning": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The bucket versioning state ('none', 'enabled', or 'suspended').",
				Validators: []validator.String{
					stringvalidator.OneOf("none", "enabled", "suspended"),
				},
			},
			"quota_limit": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "The effective quota limit applied against the size of the bucket, in bytes.",
			},
			"hard_limit_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "If true, the bucket's size cannot exceed the quota limit.",
			},
			"object_count": schema.Int64Attribute{
				Computed:    true,
				Description: "The count of objects in the bucket.",
			},
			"bucket_type": schema.StringAttribute{
				Computed:    true,
				Description: "The bucket type (e.g. 'multi-site-writable').",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"retention_lock": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The retention lock mode for the bucket.",
			},
			"eradication_config": schema.SingleNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Eradication configuration for the bucket.",
				Attributes: map[string]schema.Attribute{
					"eradication_delay": schema.Int64Attribute{
						Optional:    true,
						Computed:    true,
						Description: "Eradication delay in milliseconds.",
					},
					"eradication_mode": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Eradication mode (e.g. 'retention-based', 'permission-based').",
					},
					"manual_eradication": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Manual eradication setting ('enabled' or 'disabled').",
					},
				},
			},
			"object_lock_config": schema.SingleNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "S3 object lock configuration for the bucket.",
				Attributes: map[string]schema.Attribute{
					"freeze_locked_objects": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Whether to freeze locked objects.",
					},
					"default_retention": schema.Int64Attribute{
						Optional:    true,
						Computed:    true,
						Description: "Default retention period in seconds.",
					},
					"default_retention_mode": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Default retention mode ('compliance' or 'governance').",
					},
					"object_lock_enabled": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Whether object lock is enabled.",
					},
				},
			},
			"public_access_config": schema.SingleNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Public access configuration for the bucket.",
				Attributes: map[string]schema.Attribute{
					"block_new_public_policies": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Whether to block new public policies.",
					},
					"block_public_access": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Whether to block public access.",
					},
				},
			},
			"public_status": schema.StringAttribute{
				Computed:    true,
				Description: "Bucket's public access status.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Update: true,
				Delete: true,
			}),
			"space": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Storage space breakdown (read-only, API-managed).",
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


// ValidateConfig emits plan-time warnings for replication readiness.
func (r *bucketResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data bucketModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !data.Versioning.IsNull() && !data.Versioning.IsUnknown() {
		if data.Versioning.ValueString() != "enabled" {
			resp.Diagnostics.AddWarning(
				"Bucket versioning not enabled",
				"Versioning must be set to \"enabled\" for buckets participating in cross-array replication. "+
					"If this bucket will be used with a bucket replica link, set versioning = \"enabled\".",
			)
		}
	}
}

// UpgradeState returns state upgraders for schema migrations.
func (r *bucketResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}

// Configure injects the FlashBladeClient into the resource.
func (r *bucketResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *bucketResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data bucketModel
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

	post := client.BucketPost{
		Account: client.NamedReference{Name: data.Account.ValueString()},
	}
	if !data.QuotaLimit.IsNull() && !data.QuotaLimit.IsUnknown() {
		post.QuotaLimit = strconv.FormatInt(data.QuotaLimit.ValueInt64(), 10)
	}
	if !data.HardLimitEnabled.IsNull() && !data.HardLimitEnabled.IsUnknown() {
		post.HardLimitEnabled = data.HardLimitEnabled.ValueBool()
	}
	if !data.RetentionLock.IsNull() && !data.RetentionLock.IsUnknown() {
		post.RetentionLock = data.RetentionLock.ValueString()
	}
	if cfg := extractEradicationConfig(data.EradicationConfig); cfg != nil {
		post.EradicationConfig = cfg
	}
	if cfg := extractObjectLockConfig(data.ObjectLockConfig); cfg != nil {
		post.ObjectLockConfig = cfg
	}
	// public_access_config is NOT valid on POST — skip

	bucket, err := r.client.PostBucket(ctx, data.Name.ValueString(), post)
	if err != nil {
		resp.Diagnostics.AddError("Error creating bucket", err.Error())
		return
	}

	// Versioning is not a valid POST parameter — apply via PATCH after creation.
	if !data.Versioning.IsNull() && !data.Versioning.IsUnknown() {
		v := data.Versioning.ValueString()
		_, err := r.client.PatchBucket(ctx, bucket.ID, client.BucketPatch{
			Versioning: &v,
		})
		if err != nil {
			resp.Diagnostics.AddError("Error setting bucket versioning", err.Error())
			return
		}
	}

	resp.Diagnostics.Append(r.readIntoState(ctx, data.Name.ValueString(), &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *bucketResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data bucketModel
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

	name := data.Name.ValueString()
	bkt, err := r.client.GetBucket(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading bucket", err.Error())
		return
	}

	// Drift detection on user-configurable fields.
	if !data.Versioning.IsNull() && !data.Versioning.IsUnknown() {
		if data.Versioning.ValueString() != bkt.Versioning {
			tflog.Debug(ctx, "drift detected on bucket", map[string]any{
				"resource":    name,
				"field":       "versioning",
				"was":         data.Versioning.ValueString(),
				"now":           bkt.Versioning,
			})
		}
	}
	if !data.QuotaLimit.IsNull() && !data.QuotaLimit.IsUnknown() {
		if data.QuotaLimit.ValueInt64() != bkt.QuotaLimit {
			tflog.Debug(ctx, "drift detected on bucket", map[string]any{
				"resource":    name,
				"field":       "quota_limit",
				"was":         data.QuotaLimit.ValueInt64(),
				"now":           bkt.QuotaLimit,
			})
		}
	}
	if !data.HardLimitEnabled.IsNull() && !data.HardLimitEnabled.IsUnknown() {
		if data.HardLimitEnabled.ValueBool() != bkt.HardLimitEnabled {
			tflog.Debug(ctx, "drift detected on bucket", map[string]any{
				"resource":    name,
				"field":       "hard_limit_enabled",
				"was":         data.HardLimitEnabled.ValueBool(),
				"now":           bkt.HardLimitEnabled,
			})
		}
	}

	resp.Diagnostics.Append(mapBucketToModel(bkt, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update applies changes to an existing bucket.
func (r *bucketResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state bucketModel
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

	patch := client.BucketPatch{}

	// Only include fields that changed AND have a known planned value.
	// Computed attributes may be unknown in the plan — skip those to avoid
	// sending spurious values in the PATCH body.
	if !plan.Versioning.IsUnknown() && !plan.Versioning.Equal(state.Versioning) {
		v := plan.Versioning.ValueString()
		patch.Versioning = &v
	}
	if !plan.QuotaLimit.IsUnknown() && !plan.QuotaLimit.Equal(state.QuotaLimit) {
		v := strconv.FormatInt(plan.QuotaLimit.ValueInt64(), 10)
		patch.QuotaLimit = &v
	}
	if !plan.HardLimitEnabled.IsUnknown() && !plan.HardLimitEnabled.Equal(state.HardLimitEnabled) {
		v := plan.HardLimitEnabled.ValueBool()
		patch.HardLimitEnabled = &v
	}
	if !plan.RetentionLock.IsUnknown() && !plan.RetentionLock.Equal(state.RetentionLock) {
		v := plan.RetentionLock.ValueString()
		patch.RetentionLock = &v
	}
	if !plan.EradicationConfig.IsUnknown() && !plan.EradicationConfig.Equal(state.EradicationConfig) {
		patch.EradicationConfig = extractEradicationConfig(plan.EradicationConfig)
	}
	if !plan.ObjectLockConfig.IsUnknown() && !plan.ObjectLockConfig.Equal(state.ObjectLockConfig) {
		patch.ObjectLockConfig = extractObjectLockConfig(plan.ObjectLockConfig)
	}
	if !plan.PublicAccessConfig.IsUnknown() && !plan.PublicAccessConfig.Equal(state.PublicAccessConfig) {
		patch.PublicAccessConfig = extractPublicAccessConfig(plan.PublicAccessConfig)
	}

	_, err := r.client.PatchBucket(ctx, state.ID.ValueString(), patch)
	if err != nil {
		resp.Diagnostics.AddError("Error updating bucket", err.Error())
		return
	}

	resp.Diagnostics.Append(r.readIntoState(ctx, plan.Name.ValueString(), &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes a bucket via two-phase soft-delete + optional eradication.
func (r *bucketResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data bucketModel
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

	id := data.ID.ValueString()
	name := data.Name.ValueString()

	// Fresh GET to check current object count (state data may be stale).
	freshBucket, err := r.client.GetBucket(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			// Already gone -- nothing to delete.
			return
		}
		resp.Diagnostics.AddError("Error reading bucket before deletion", err.Error())
		return
	}
	if freshBucket.ObjectCount > 0 {
		resp.Diagnostics.AddError(
			"Cannot delete bucket: bucket contains objects",
			fmt.Sprintf("Bucket %q contains %d object(s). Empty the bucket before deleting it.", name, freshBucket.ObjectCount),
		)
		return
	}

	eradicate := data.DestroyEradicateOnDelete.ValueBool()
	if err := r.client.DestroyAndEradicateBucket(ctx, id, name, eradicate); err != nil {
		resp.Diagnostics.AddError("Error deleting bucket", err.Error())
		return
	}
}

func (r *bucketResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	name := req.ID
	bkt, err := r.client.GetBucket(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError("Error importing bucket", err.Error())
		return
	}

	var data bucketModel
	// Default for the provider-only field.
	data.DestroyEradicateOnDelete = types.BoolValue(false)
	// Initialize timeouts with a proper null value.
	data.Timeouts = nullTimeoutsValue()

	resp.Diagnostics.Append(mapBucketToModel(bkt, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ---------- helpers ---------------------------------------------------------

// readIntoState calls GetBucket and maps the result into the provided model.
func (r *bucketResource) readIntoState(ctx context.Context, name string, data *bucketModel) diag.Diagnostics {
	var diags diag.Diagnostics

	bkt, err := r.client.GetBucket(ctx, name)
	if err != nil {
		diags.AddError("Error reading bucket after write", err.Error())
		return diags
	}
	for _, d := range mapBucketToModel(bkt, data) {
		diags.AddError(d.Summary(), d.Detail())
	}
	return diags
}


// mapBucketToModel maps a client.Bucket to a bucketModel.
// It preserves user-managed fields (DestroyEradicateOnDelete, Timeouts).
// Returns diagnostics instead of panicking on object construction errors.
func mapBucketToModel(bkt *client.Bucket, data *bucketModel) diag.Diagnostics {
	var diags diag.Diagnostics

	data.ID = types.StringValue(bkt.ID)
	data.Name = types.StringValue(bkt.Name)
	data.Account = types.StringValue(bkt.Account.Name)
	data.Created = types.Int64Value(bkt.Created)
	data.Destroyed = types.BoolValue(bkt.Destroyed)
	data.TimeRemaining = types.Int64Value(bkt.TimeRemaining)
	data.Versioning = types.StringValue(bkt.Versioning)
	data.QuotaLimit = types.Int64Value(bkt.QuotaLimit)
	data.HardLimitEnabled = types.BoolValue(bkt.HardLimitEnabled)
	data.ObjectCount = types.Int64Value(bkt.ObjectCount)
	data.BucketType = types.StringValue(bkt.BucketType)
	data.RetentionLock = types.StringValue(bkt.RetentionLock)

	spaceObj, spaceDiags := mapSpaceToObject(bkt.Space)
	diags.Append(spaceDiags...)
	if diags.HasError() {
		return diags
	}
	data.Space = spaceObj

	eradObj, eradDiags := mapEradicationConfigToObject(bkt.EradicationConfig)
	diags.Append(eradDiags...)
	data.EradicationConfig = eradObj

	olObj, olDiags := mapObjectLockConfigToObject(bkt.ObjectLockConfig)
	diags.Append(olDiags...)
	data.ObjectLockConfig = olObj

	paObj, paDiags := mapPublicAccessConfigToObject(bkt.PublicAccessConfig)
	diags.Append(paDiags...)
	data.PublicAccessConfig = paObj

	data.PublicStatus = types.StringValue(bkt.PublicStatus)

	return diags
}

// ---------- bucket config attr types and mapping helpers --------------------

// eradicationConfigAttrTypes returns the attribute type map for the eradication_config nested object.
func eradicationConfigAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"eradication_delay":  types.Int64Type,
		"eradication_mode":   types.StringType,
		"manual_eradication": types.StringType,
	}
}

// objectLockConfigAttrTypes returns the attribute type map for the object_lock_config nested object.
func objectLockConfigAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"freeze_locked_objects":  types.BoolType,
		"default_retention":      types.Int64Type,
		"default_retention_mode": types.StringType,
		"object_lock_enabled":    types.BoolType,
	}
}

// publicAccessConfigAttrTypes returns the attribute type map for the public_access_config nested object.
func publicAccessConfigAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"block_new_public_policies": types.BoolType,
		"block_public_access":       types.BoolType,
	}
}

// mapEradicationConfigToObject builds a types.Object from a client.EradicationConfig.
func mapEradicationConfigToObject(cfg client.EradicationConfig) (types.Object, diag.Diagnostics) {
	return types.ObjectValue(eradicationConfigAttrTypes(), map[string]attr.Value{
		"eradication_delay":  types.Int64Value(cfg.EradicationDelay),
		"eradication_mode":   types.StringValue(cfg.EradicationMode),
		"manual_eradication": types.StringValue(cfg.ManualEradication),
	})
}

// mapObjectLockConfigToObject builds a types.Object from a client.ObjectLockConfig.
func mapObjectLockConfigToObject(cfg client.ObjectLockConfig) (types.Object, diag.Diagnostics) {
	return types.ObjectValue(objectLockConfigAttrTypes(), map[string]attr.Value{
		"freeze_locked_objects":  types.BoolValue(cfg.FreezeLockedObjects),
		"default_retention":      types.Int64Value(cfg.DefaultRetention),
		"default_retention_mode": types.StringValue(cfg.DefaultRetentionMode),
		"object_lock_enabled":    types.BoolValue(cfg.ObjectLockEnabled),
	})
}

// mapPublicAccessConfigToObject builds a types.Object from a client.PublicAccessConfig.
func mapPublicAccessConfigToObject(cfg client.PublicAccessConfig) (types.Object, diag.Diagnostics) {
	return types.ObjectValue(publicAccessConfigAttrTypes(), map[string]attr.Value{
		"block_new_public_policies": types.BoolValue(cfg.BlockNewPublicPolicies),
		"block_public_access":       types.BoolValue(cfg.BlockPublicAccess),
	})
}

// extractEradicationConfig extracts a client.EradicationConfig from a plan types.Object.
// Returns nil if the object is null or unknown.
func extractEradicationConfig(obj types.Object) *client.EradicationConfig {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}
	attrs := obj.Attributes()
	cfg := &client.EradicationConfig{}
	if v, ok := attrs["eradication_delay"].(types.Int64); ok && !v.IsNull() && !v.IsUnknown() {
		cfg.EradicationDelay = v.ValueInt64()
	}
	if v, ok := attrs["eradication_mode"].(types.String); ok && !v.IsNull() && !v.IsUnknown() {
		cfg.EradicationMode = v.ValueString()
	}
	if v, ok := attrs["manual_eradication"].(types.String); ok && !v.IsNull() && !v.IsUnknown() {
		cfg.ManualEradication = v.ValueString()
	}
	return cfg
}

// extractObjectLockConfig extracts a client.ObjectLockConfig from a plan types.Object.
// Returns nil if the object is null or unknown.
func extractObjectLockConfig(obj types.Object) *client.ObjectLockConfig {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}
	attrs := obj.Attributes()
	cfg := &client.ObjectLockConfig{}
	if v, ok := attrs["freeze_locked_objects"].(types.Bool); ok && !v.IsNull() && !v.IsUnknown() {
		cfg.FreezeLockedObjects = v.ValueBool()
	}
	if v, ok := attrs["default_retention"].(types.Int64); ok && !v.IsNull() && !v.IsUnknown() {
		cfg.DefaultRetention = v.ValueInt64()
	}
	if v, ok := attrs["default_retention_mode"].(types.String); ok && !v.IsNull() && !v.IsUnknown() {
		cfg.DefaultRetentionMode = v.ValueString()
	}
	if v, ok := attrs["object_lock_enabled"].(types.Bool); ok && !v.IsNull() && !v.IsUnknown() {
		cfg.ObjectLockEnabled = v.ValueBool()
	}
	return cfg
}

// extractPublicAccessConfig extracts a client.PublicAccessConfig from a plan types.Object.
// Returns nil if the object is null or unknown.
func extractPublicAccessConfig(obj types.Object) *client.PublicAccessConfig {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}
	attrs := obj.Attributes()
	cfg := &client.PublicAccessConfig{}
	if v, ok := attrs["block_new_public_policies"].(types.Bool); ok && !v.IsNull() && !v.IsUnknown() {
		cfg.BlockNewPublicPolicies = v.ValueBool()
	}
	if v, ok := attrs["block_public_access"].(types.Bool); ok && !v.IsNull() && !v.IsUnknown() {
		cfg.BlockPublicAccess = v.ValueBool()
	}
	return cfg
}
