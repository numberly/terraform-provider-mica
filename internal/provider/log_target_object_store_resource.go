package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// Ensure logTargetObjectStoreResource satisfies the resource interfaces.
var _ resource.Resource = &logTargetObjectStoreResource{}
var _ resource.ResourceWithConfigure = &logTargetObjectStoreResource{}
var _ resource.ResourceWithImportState = &logTargetObjectStoreResource{}
var _ resource.ResourceWithUpgradeState = &logTargetObjectStoreResource{}

// logTargetObjectStoreResource implements the flashblade_log_target_object_store resource.
type logTargetObjectStoreResource struct {
	client *client.FlashBladeClient
}

// NewLogTargetObjectStoreResource is the factory function registered in the provider.
func NewLogTargetObjectStoreResource() resource.Resource {
	return &logTargetObjectStoreResource{}
}

// ---------- model structs ----------------------------------------------------

// logTargetObjectStoreModel is the top-level model for the flashblade_log_target_object_store resource.
type logTargetObjectStoreModel struct {
	ID                types.String   `tfsdk:"id"`
	Name              types.String   `tfsdk:"name"`
	BucketName        types.String   `tfsdk:"bucket_name"`
	LogNamePrefix     types.String   `tfsdk:"log_name_prefix"`
	LogRotateDuration types.Int64    `tfsdk:"log_rotate_duration"`
	Timeouts          timeouts.Value `tfsdk:"timeouts"`
}

// ---------- resource interface methods --------------------------------------

// Metadata sets the Terraform type name.
func (r *logTargetObjectStoreResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_log_target_object_store"
}

// Schema defines the resource schema.
func (r *logTargetObjectStoreResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     0,
		Description: "Manages a FlashBlade log target object store that stores audit logs in a bucket.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the log target object store.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the log target object store. Not renameable; changing forces replacement.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"bucket_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the bucket where audit logs will be stored.",
			},
			"log_name_prefix": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The prefix of audit log object names in the bucket.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"log_rotate_duration": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "The rotation interval for audit logs in milliseconds.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
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
func (r *logTargetObjectStoreResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}

// Configure injects the FlashBladeClient into the resource.
func (r *logTargetObjectStoreResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new log target object store.
func (r *logTargetObjectStoreResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data logTargetObjectStoreModel
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

	name := data.Name.ValueString()
	post := client.LogTargetObjectStorePost{
		Bucket: client.NamedReference{Name: data.BucketName.ValueString()},
	}
	if !data.LogNamePrefix.IsNull() && !data.LogNamePrefix.IsUnknown() {
		post.LogNamePrefix = client.AuditLogNamePrefix{Prefix: data.LogNamePrefix.ValueString()}
	}
	if !data.LogRotateDuration.IsNull() && !data.LogRotateDuration.IsUnknown() {
		post.LogRotate = client.AuditLogRotate{Duration: data.LogRotateDuration.ValueInt64()}
	}

	item, err := r.client.PostLogTargetObjectStore(ctx, name, post)
	if err != nil {
		resp.Diagnostics.AddError("Error creating log target object store", err.Error())
		return
	}

	mapLogTargetObjectStoreToModel(item, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes Terraform state from the API.
func (r *logTargetObjectStoreResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data logTargetObjectStoreModel
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
	item, err := r.client.GetLogTargetObjectStore(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading log target object store", err.Error())
		return
	}

	// Drift detection on bucket_name.
	if !data.BucketName.IsNull() && !data.BucketName.IsUnknown() {
		if data.BucketName.ValueString() != item.Bucket.Name {
			tflog.Debug(ctx, "drift detected on log target object store", map[string]any{
				"resource": name,
				"field":    "bucket_name",
				"was":      data.BucketName.ValueString(),
				"now":      item.Bucket.Name,
			})
		}
	}

	// Drift detection on log_name_prefix.
	if !data.LogNamePrefix.IsNull() && !data.LogNamePrefix.IsUnknown() {
		if data.LogNamePrefix.ValueString() != item.LogNamePrefix.Prefix {
			tflog.Debug(ctx, "drift detected on log target object store", map[string]any{
				"resource": name,
				"field":    "log_name_prefix",
				"was":      data.LogNamePrefix.ValueString(),
				"now":      item.LogNamePrefix.Prefix,
			})
		}
	}

	// Drift detection on log_rotate_duration.
	if !data.LogRotateDuration.IsNull() && !data.LogRotateDuration.IsUnknown() {
		if data.LogRotateDuration.ValueInt64() != item.LogRotate.Duration {
			tflog.Debug(ctx, "drift detected on log target object store", map[string]any{
				"resource": name,
				"field":    "log_rotate_duration",
				"was":      data.LogRotateDuration.ValueInt64(),
				"now":      item.LogRotate.Duration,
			})
		}
	}

	mapLogTargetObjectStoreToModel(item, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update applies changes to an existing log target object store.
func (r *logTargetObjectStoreResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state logTargetObjectStoreModel
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

	name := state.Name.ValueString()
	patch := client.LogTargetObjectStorePatch{}
	needsPatch := false

	if !plan.BucketName.Equal(state.BucketName) {
		bucketRef := client.NamedReference{Name: plan.BucketName.ValueString()}
		patch.Bucket = &bucketRef
		needsPatch = true
	}

	if !plan.LogNamePrefix.Equal(state.LogNamePrefix) {
		prefix := client.AuditLogNamePrefix{Prefix: plan.LogNamePrefix.ValueString()}
		patch.LogNamePrefix = &prefix
		needsPatch = true
	}

	if !plan.LogRotateDuration.Equal(state.LogRotateDuration) {
		rotate := client.AuditLogRotate{Duration: plan.LogRotateDuration.ValueInt64()}
		patch.LogRotate = &rotate
		needsPatch = true
	}

	if needsPatch {
		_, err := r.client.PatchLogTargetObjectStore(ctx, name, patch)
		if err != nil {
			resp.Diagnostics.AddError("Error updating log target object store", err.Error())
			return
		}
	}

	r.readIntoState(ctx, name, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes a log target object store.
func (r *logTargetObjectStoreResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data logTargetObjectStoreModel
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

	name := data.Name.ValueString()
	if err := r.client.DeleteLogTargetObjectStore(ctx, name); err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting log target object store", err.Error())
		return
	}
}

// ImportState imports an existing log target object store by name.
func (r *logTargetObjectStoreResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	name := req.ID

	var data logTargetObjectStoreModel
	data.Timeouts = nullTimeoutsValue()
	data.Name = types.StringValue(name)

	r.readIntoState(ctx, name, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ---------- helpers ---------------------------------------------------------

// readIntoState calls GetLogTargetObjectStore and maps the result into the provided model.
func (r *logTargetObjectStoreResource) readIntoState(ctx context.Context, name string, data *logTargetObjectStoreModel, diags DiagnosticReporter) {
	item, err := r.client.GetLogTargetObjectStore(ctx, name)
	if err != nil {
		diags.AddError("Error reading log target object store after write", err.Error())
		return
	}
	mapLogTargetObjectStoreToModel(item, data)
}

// mapLogTargetObjectStoreToModel converts a client.LogTargetObjectStore to the Terraform model.
// It preserves user-managed fields (Timeouts).
func mapLogTargetObjectStoreToModel(item *client.LogTargetObjectStore, data *logTargetObjectStoreModel) {
	data.ID = types.StringValue(item.ID)
	data.Name = types.StringValue(item.Name)
	data.BucketName = types.StringValue(item.Bucket.Name)
	data.LogNamePrefix = types.StringValue(item.LogNamePrefix.Prefix)
	data.LogRotateDuration = types.Int64Value(item.LogRotate.Duration)
}
