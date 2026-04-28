package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

var _ resource.Resource = &auditObjectStorePolicyResource{}
var _ resource.ResourceWithConfigure = &auditObjectStorePolicyResource{}
var _ resource.ResourceWithImportState = &auditObjectStorePolicyResource{}
var _ resource.ResourceWithUpgradeState = &auditObjectStorePolicyResource{}

// auditObjectStorePolicyResource implements the flashblade_audit_object_store_policy resource.
type auditObjectStorePolicyResource struct {
	client *client.FlashBladeClient
}

func NewAuditObjectStorePolicyResource() resource.Resource {
	return &auditObjectStorePolicyResource{}
}

// ---------- model structs ----------------------------------------------------

// auditObjectStorePolicyModel is the top-level model for the flashblade_audit_object_store_policy resource.
type auditObjectStorePolicyModel struct {
	ID         types.String   `tfsdk:"id"`
	Name       types.String   `tfsdk:"name"`
	Enabled    types.Bool     `tfsdk:"enabled"`
	IsLocal    types.Bool     `tfsdk:"is_local"`
	PolicyType types.String   `tfsdk:"policy_type"`
	LogTargets types.List     `tfsdk:"log_targets"`
	Timeouts   timeouts.Value `tfsdk:"timeouts"`
}

// ---------- resource interface methods --------------------------------------

func (r *auditObjectStorePolicyResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_audit_object_store_policy"
}

// Schema defines the resource schema.
func (r *auditObjectStorePolicyResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     0,
		Description: "Manages a FlashBlade audit object store policy that controls audit logging for object store operations.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the audit object store policy.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the audit object store policy. Not renameable; changing forces replacement.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the audit object store policy is enabled.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"is_local": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the policy is defined on the local array (read-only).",
			},
			"policy_type": schema.StringAttribute{
				Computed:    true,
				Description: "The type of the policy (e.g. 'audit'). Read-only, set by the array.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"log_targets": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of log target names to receive audit events from this policy.",
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
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
func (r *auditObjectStorePolicyResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}

// Configure injects the FlashBladeClient into the resource.
func (r *auditObjectStorePolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *auditObjectStorePolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data auditObjectStorePolicyModel
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

	post := client.AuditObjectStorePolicyPost{}
	if !data.Enabled.IsNull() && !data.Enabled.IsUnknown() {
		v := data.Enabled.ValueBool()
		post.Enabled = &v
	}

	logTargetNames := stringSliceFromList(data.LogTargets)
	if len(logTargetNames) > 0 {
		refs := make([]client.NamedReference, len(logTargetNames))
		for i, n := range logTargetNames {
			refs[i] = client.NamedReference{Name: n}
		}
		post.LogTargets = refs
	}

	policy, err := r.client.PostAuditObjectStorePolicy(ctx, name, post)
	if err != nil {
		resp.Diagnostics.AddError("Error creating audit object store policy", err.Error())
		return
	}

	mapAuditObjectStorePolicyToModel(policy, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *auditObjectStorePolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data auditObjectStorePolicyModel
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
	policy, err := r.client.GetAuditObjectStorePolicy(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading audit object store policy", err.Error())
		return
	}

	// Drift detection on enabled.
	if !data.Enabled.IsNull() && !data.Enabled.IsUnknown() {
		if data.Enabled.ValueBool() != policy.Enabled {
			tflog.Debug(ctx, "drift detected", map[string]any{
				"resource": name,
				"field":    "enabled",
				"was":      data.Enabled.ValueBool(),
				"now":      policy.Enabled,
			})
		}
	}

	// Drift detection on log_targets.
	apiTargetNames := namedRefsToNames(policy.LogTargets)
	stateTargetNames := stringSliceFromList(data.LogTargets)
	if !stringSlicesEqual(stateTargetNames, apiTargetNames) {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": name,
			"field":    "log_targets",
			"was":      stateTargetNames,
			"now":      apiTargetNames,
		})
	}

	mapAuditObjectStorePolicyToModel(policy, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update applies changes to an existing audit object store policy.
func (r *auditObjectStorePolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state auditObjectStorePolicyModel
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
	patch := client.AuditObjectStorePolicyPatch{}
	needsPatch := false

	if !plan.Enabled.Equal(state.Enabled) {
		v := plan.Enabled.ValueBool()
		patch.Enabled = &v
		needsPatch = true
	}

	if !plan.LogTargets.Equal(state.LogTargets) {
		names := stringSliceFromList(plan.LogTargets)
		refs := make([]client.NamedReference, len(names))
		for i, n := range names {
			refs[i] = client.NamedReference{Name: n}
		}
		patch.LogTargets = &refs
		needsPatch = true
	}

	if needsPatch {
		_, err := r.client.PatchAuditObjectStorePolicy(ctx, name, patch)
		if err != nil {
			resp.Diagnostics.AddError("Error updating audit object store policy", err.Error())
			return
		}
	}

	resp.Diagnostics.Append(r.readIntoState(ctx, name, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes an audit object store policy.
func (r *auditObjectStorePolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data auditObjectStorePolicyModel
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
	if err := r.client.DeleteAuditObjectStorePolicy(ctx, name); err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting audit object store policy", err.Error())
		return
	}
}

func (r *auditObjectStorePolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	name := req.ID

	var data auditObjectStorePolicyModel
	data.Timeouts = nullTimeoutsValue()
	data.Name = types.StringValue(name)

	resp.Diagnostics.Append(r.readIntoState(ctx, name, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ---------- helpers ---------------------------------------------------------

// readIntoState calls GetAuditObjectStorePolicy and maps the result into the provided model.
func (r *auditObjectStorePolicyResource) readIntoState(ctx context.Context, name string, data *auditObjectStorePolicyModel) diag.Diagnostics {
	var diags diag.Diagnostics

	policy, err := r.client.GetAuditObjectStorePolicy(ctx, name)
	if err != nil {
		diags.AddError("Error reading audit object store policy after write", err.Error())
		return diags
	}
	mapAuditObjectStorePolicyToModel(policy, data)
	return diags
}


// mapAuditObjectStorePolicyToModel converts a client.AuditObjectStorePolicy to the Terraform model.
// It preserves user-managed fields (Timeouts).
func mapAuditObjectStorePolicyToModel(policy *client.AuditObjectStorePolicy, data *auditObjectStorePolicyModel) {
	data.ID = types.StringValue(policy.ID)
	data.Name = types.StringValue(policy.Name)
	data.Enabled = types.BoolValue(policy.Enabled)
	data.IsLocal = types.BoolValue(policy.IsLocal)
	data.PolicyType = types.StringValue(policy.PolicyType)

	// Map LogTargets: []NamedReference -> list of name strings.
	names := namedRefsToNames(policy.LogTargets)
	if len(names) > 0 {
		vals := make([]attr.Value, len(names))
		for i, n := range names {
			vals[i] = types.StringValue(n)
		}
		data.LogTargets = types.ListValueMust(types.StringType, vals)
	} else {
		data.LogTargets = types.ListValueMust(types.StringType, []attr.Value{})
	}
}
