package provider

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// Ensure bucketResource satisfies the resource interfaces.
var _ resource.Resource = &bucketResource{}
var _ resource.ResourceWithConfigure = &bucketResource{}
var _ resource.ResourceWithImportState = &bucketResource{}
var _ resource.ResourceWithUpgradeState = &bucketResource{}

// bucketResource implements the flashblade_bucket resource.
type bucketResource struct {
	client *client.FlashBladeClient
}

// NewBucketResource is the factory function registered in the provider.
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
	Space                    types.Object   `tfsdk:"space"`
	Timeouts                 timeouts.Value `tfsdk:"timeouts"`
}

// ---------- resource interface methods --------------------------------------

// Metadata sets the Terraform type name.
func (r *bucketResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_bucket"
}

// Schema defines the resource schema.
func (r *bucketResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a FlashBlade object store bucket.",
		Version:     0,
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
					int64UseStateForUnknown(),
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

// Create creates a new bucket.
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

	r.readIntoState(ctx, data.Name.ValueString(), &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes Terraform state from the API.
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
			tflog.Info(ctx, "drift detected on bucket", map[string]any{
				"resource":    name,
				"field":       "versioning",
				"state_value": data.Versioning.ValueString(),
				"api_value":   bkt.Versioning,
			})
		}
	}
	if !data.QuotaLimit.IsNull() && !data.QuotaLimit.IsUnknown() {
		if data.QuotaLimit.ValueInt64() != bkt.QuotaLimit {
			tflog.Info(ctx, "drift detected on bucket", map[string]any{
				"resource":    name,
				"field":       "quota_limit",
				"state_value": data.QuotaLimit.ValueInt64(),
				"api_value":   bkt.QuotaLimit,
			})
		}
	}
	if !data.HardLimitEnabled.IsNull() && !data.HardLimitEnabled.IsUnknown() {
		if data.HardLimitEnabled.ValueBool() != bkt.HardLimitEnabled {
			tflog.Info(ctx, "drift detected on bucket", map[string]any{
				"resource":    name,
				"field":       "hard_limit_enabled",
				"state_value": data.HardLimitEnabled.ValueBool(),
				"api_value":   bkt.HardLimitEnabled,
			})
		}
	}

	mapBucketToModel(bkt, &data)
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

	if !plan.Versioning.Equal(state.Versioning) {
		v := plan.Versioning.ValueString()
		patch.Versioning = &v
	}
	if !plan.QuotaLimit.Equal(state.QuotaLimit) {
		v := strconv.FormatInt(plan.QuotaLimit.ValueInt64(), 10)
		patch.QuotaLimit = &v
	}
	if !plan.HardLimitEnabled.Equal(state.HardLimitEnabled) {
		v := plan.HardLimitEnabled.ValueBool()
		patch.HardLimitEnabled = &v
	}
	if !plan.RetentionLock.Equal(state.RetentionLock) {
		v := plan.RetentionLock.ValueString()
		patch.RetentionLock = &v
	}

	_, err := r.client.PatchBucket(ctx, state.ID.ValueString(), patch)
	if err != nil {
		resp.Diagnostics.AddError("Error updating bucket", err.Error())
		return
	}

	r.readIntoState(ctx, plan.Name.ValueString(), &plan, &resp.Diagnostics)
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

	// Guard: bucket must be empty before deletion.
	if !data.ObjectCount.IsNull() && !data.ObjectCount.IsUnknown() && data.ObjectCount.ValueInt64() > 0 {
		resp.Diagnostics.AddError(
			"Cannot delete bucket: bucket contains objects",
			fmt.Sprintf("Bucket %q contains %d object(s). Empty the bucket before deleting it.", data.Name.ValueString(), data.ObjectCount.ValueInt64()),
		)
		return
	}

	id := data.ID.ValueString()
	name := data.Name.ValueString()

	// Phase 1: Soft-delete.
	destroyed := true
	_, err := r.client.PatchBucket(ctx, id, client.BucketPatch{Destroyed: &destroyed})
	if err != nil {
		if client.IsNotFound(err) {
			// Already gone — no error.
			return
		}
		resp.Diagnostics.AddError("Error soft-deleting bucket", err.Error())
		return
	}

	// Phase 2: Eradicate only if destroy_eradicate_on_delete is true.
	eradicate := data.DestroyEradicateOnDelete.ValueBool()
	if eradicate {
		if err := r.client.DeleteBucket(ctx, id); err != nil {
			if !client.IsNotFound(err) {
				resp.Diagnostics.AddError("Error eradicating bucket", err.Error())
				return
			}
		}
		if err := r.client.PollBucketUntilEradicated(ctx, name); err != nil {
			resp.Diagnostics.AddError("Error waiting for bucket eradication", err.Error())
			return
		}
	}
}

// ImportState imports an existing bucket by name.
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
	data.Timeouts = timeouts.Value{
		Object: types.ObjectNull(map[string]attr.Type{
			"create": types.StringType,
			"read":   types.StringType,
			"update": types.StringType,
			"delete": types.StringType,
		}),
	}

	mapBucketToModel(bkt, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ---------- helpers ---------------------------------------------------------

// bucketSpaceAttrTypes returns the attribute types for the bucket space object.
func bucketSpaceAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"data_reduction":      types.Float64Type,
		"snapshots":           types.Int64Type,
		"total_physical":      types.Int64Type,
		"unique":              types.Int64Type,
		"virtual":             types.Int64Type,
		"snapshots_effective": types.Int64Type,
	}
}

// readIntoState calls GetBucket and maps the result into the provided model.
func (r *bucketResource) readIntoState(ctx context.Context, name string, data *bucketModel, diags interface {
	AddError(string, string)
	HasError() bool
}) {
	bkt, err := r.client.GetBucket(ctx, name)
	if err != nil {
		diags.AddError("Error reading bucket after write", err.Error())
		return
	}
	mapBucketToModel(bkt, data)
}

// mapBucketToModel maps a client.Bucket to a bucketModel.
// It preserves user-managed fields (DestroyEradicateOnDelete, Timeouts).
func mapBucketToModel(bkt *client.Bucket, data *bucketModel) {
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

	spaceObj, diags := types.ObjectValue(bucketSpaceAttrTypes(), map[string]attr.Value{
		"data_reduction":      types.Float64Value(bkt.Space.DataReduction),
		"snapshots":           types.Int64Value(bkt.Space.Snapshots),
		"total_physical":      types.Int64Value(bkt.Space.TotalPhysical),
		"unique":              types.Int64Value(bkt.Space.Unique),
		"virtual":             types.Int64Value(bkt.Space.Virtual),
		"snapshots_effective": types.Int64Value(bkt.Space.SnapshotsEffective),
	})
	// ObjectValue only fails if keys/types mismatch — treat as a coding error.
	if diags.HasError() {
		panic("mapBucketToModel: failed to build space object: " + diags[0].Detail())
	}
	data.Space = spaceObj
}
