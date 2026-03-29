package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// Ensure snapshotPolicyResource satisfies the resource interfaces.
var _ resource.Resource = &snapshotPolicyResource{}
var _ resource.ResourceWithConfigure = &snapshotPolicyResource{}
var _ resource.ResourceWithImportState = &snapshotPolicyResource{}

// snapshotPolicyResource implements the flashblade_snapshot_policy resource.
type snapshotPolicyResource struct {
	client *client.FlashBladeClient
}

// NewSnapshotPolicyResource is the factory function registered in the provider.
func NewSnapshotPolicyResource() resource.Resource {
	return &snapshotPolicyResource{}
}

// ---------- model structs ----------------------------------------------------

// snapshotPolicyModel is the top-level model for the flashblade_snapshot_policy resource.
type snapshotPolicyModel struct {
	ID            types.String   `tfsdk:"id"`
	Name          types.String   `tfsdk:"name"`
	Enabled       types.Bool     `tfsdk:"enabled"`
	IsLocal       types.Bool     `tfsdk:"is_local"`
	PolicyType    types.String   `tfsdk:"policy_type"`
	RetentionLock types.String   `tfsdk:"retention_lock"`
	Timeouts      timeouts.Value `tfsdk:"timeouts"`
}

// ---------- resource interface methods --------------------------------------

// Metadata sets the Terraform type name.
func (r *snapshotPolicyResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_snapshot_policy"
}

// Schema defines the resource schema.
func (r *snapshotPolicyResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a FlashBlade snapshot policy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the snapshot policy.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
				Description: "The name of the snapshot policy. " +
					"Changing this forces a new resource. Snapshot policy names cannot be renamed in-place (API limitation).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "If true, the policy is enabled and its rules are enforced.",
			},
			"is_local": schema.BoolAttribute{
				Computed:    true,
				Description: "If true, the policy is local to this array (not replicated).",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"policy_type": schema.StringAttribute{
				Computed:    true,
				Description: "The type of the policy (e.g. 'snapshot').",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"retention_lock": schema.StringAttribute{
				Computed:    true,
				Description: "The retention lock mode of the policy (e.g. 'none', 'ratcheted').",
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
		},
	}
}


// Configure injects the FlashBladeClient into the resource.
func (r *snapshotPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new snapshot policy.
func (r *snapshotPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data snapshotPolicyModel
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

	post := client.SnapshotPolicyPost{}
	if !data.Enabled.IsNull() && !data.Enabled.IsUnknown() {
		v := data.Enabled.ValueBool()
		post.Enabled = &v
	}

	_, err := r.client.PostSnapshotPolicy(ctx, data.Name.ValueString(), post)
	if err != nil {
		resp.Diagnostics.AddError("Error creating snapshot policy", err.Error())
		return
	}

	r.readIntoState(ctx, data.Name.ValueString(), &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes Terraform state from the API.
func (r *snapshotPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data snapshotPolicyModel
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
	policy, err := r.client.GetSnapshotPolicy(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading snapshot policy", err.Error())
		return
	}

	// Drift detection on enabled field.
	if !data.Enabled.IsNull() && !data.Enabled.IsUnknown() {
		if data.Enabled.ValueBool() != policy.Enabled {
			tflog.Info(ctx, "drift detected on snapshot policy", map[string]any{
				"resource":    name,
				"field":       "enabled",
				"state_value": data.Enabled.ValueBool(),
				"api_value":   policy.Enabled,
			})
		}
	}

	mapSnapshotPolicyToModel(policy, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update applies changes to an existing snapshot policy.
// Since name has RequiresReplace, the only in-place update is enabled.
func (r *snapshotPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state snapshotPolicyModel
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

	// Name is RequiresReplace, so it cannot change here — address via current name.
	name := state.Name.ValueString()
	patch := client.SnapshotPolicyPatch{}

	if !plan.Enabled.Equal(state.Enabled) {
		v := plan.Enabled.ValueBool()
		patch.Enabled = &v
	}

	_, err := r.client.PatchSnapshotPolicy(ctx, name, patch)
	if err != nil {
		resp.Diagnostics.AddError("Error updating snapshot policy", err.Error())
		return
	}

	r.readIntoState(ctx, name, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes a snapshot policy.
// Blocked if the policy is currently attached to one or more file systems.
func (r *snapshotPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data snapshotPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, diags := data.Timeouts.Delete(ctx, 10*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	name := data.Name.ValueString()

	// Guard: check for file systems using this policy before deleting.
	members, err := r.client.ListSnapshotPolicyMembers(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError("Error checking policy members before deletion", err.Error())
		return
	}
	if len(members) > 0 {
		resp.Diagnostics.AddError(
			"Cannot delete snapshot policy",
			fmt.Sprintf("Policy %q is attached to %d file system(s). Detach the policy from all file systems before deleting.", name, len(members)),
		)
		return
	}

	if err := r.client.DeleteSnapshotPolicy(ctx, name); err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting snapshot policy", err.Error())
		return
	}
}

// ImportState imports an existing snapshot policy by name.
func (r *snapshotPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	name := req.ID

	var data snapshotPolicyModel
	// Initialize timeouts with a null value so the framework can serialize it.
	data.Timeouts = nullTimeoutsValue()
	// Set Name so Read can look up the policy.
	data.Name = types.StringValue(name)

	r.readIntoState(ctx, name, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ---------- helpers ---------------------------------------------------------

// readIntoState calls GetSnapshotPolicy and maps the result into the provided model.
func (r *snapshotPolicyResource) readIntoState(ctx context.Context, name string, data *snapshotPolicyModel, diags DiagnosticReporter) {
	policy, err := r.client.GetSnapshotPolicy(ctx, name)
	if err != nil {
		diags.AddError("Error reading snapshot policy after write", err.Error())
		return
	}
	mapSnapshotPolicyToModel(policy, data)
}

// mapSnapshotPolicyToModel maps a client.SnapshotPolicy to a snapshotPolicyModel.
// It preserves user-managed fields (Timeouts).
func mapSnapshotPolicyToModel(policy *client.SnapshotPolicy, data *snapshotPolicyModel) {
	data.ID = types.StringValue(policy.ID)
	data.Name = types.StringValue(policy.Name)
	data.Enabled = types.BoolValue(policy.Enabled)
	data.IsLocal = types.BoolValue(policy.IsLocal)
	data.PolicyType = types.StringValue(policy.PolicyType)
	data.RetentionLock = types.StringValue(policy.RetentionLock)
}
