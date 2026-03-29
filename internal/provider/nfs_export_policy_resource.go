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

// Ensure nfsExportPolicyResource satisfies the resource interfaces.
var _ resource.Resource = &nfsExportPolicyResource{}
var _ resource.ResourceWithConfigure = &nfsExportPolicyResource{}
var _ resource.ResourceWithImportState = &nfsExportPolicyResource{}
var _ resource.ResourceWithUpgradeState = &nfsExportPolicyResource{}

// nfsExportPolicyResource implements the flashblade_nfs_export_policy resource.
type nfsExportPolicyResource struct {
	client *client.FlashBladeClient
}

// NewNfsExportPolicyResource is the factory function registered in the provider.
func NewNfsExportPolicyResource() resource.Resource {
	return &nfsExportPolicyResource{}
}

// ---------- model structs ----------------------------------------------------

// nfsExportPolicyModel is the top-level model for the flashblade_nfs_export_policy resource.
type nfsExportPolicyModel struct {
	ID         types.String   `tfsdk:"id"`
	Name       types.String   `tfsdk:"name"`
	Enabled    types.Bool     `tfsdk:"enabled"`
	IsLocal    types.Bool     `tfsdk:"is_local"`
	PolicyType types.String   `tfsdk:"policy_type"`
	Version    types.String   `tfsdk:"version"`
	Timeouts   timeouts.Value `tfsdk:"timeouts"`
}

// ---------- resource interface methods --------------------------------------

// Metadata sets the Terraform type name.
func (r *nfsExportPolicyResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_nfs_export_policy"
}

// Schema defines the resource schema.
func (r *nfsExportPolicyResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a FlashBlade NFS export policy.",
		Version:     0,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the NFS export policy.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the NFS export policy. Can be changed in-place via PATCH (rename).",
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
				Description: "The type of the policy (e.g. 'nfs').",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"version": schema.StringAttribute{
				Computed:    true,
				Description: "The version token that changes on each policy update.",
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

func (r *nfsExportPolicyResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}

// Configure injects the FlashBladeClient into the resource.
func (r *nfsExportPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new NFS export policy.
func (r *nfsExportPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data nfsExportPolicyModel
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

	post := client.NfsExportPolicyPost{}
	if !data.Enabled.IsNull() && !data.Enabled.IsUnknown() {
		v := data.Enabled.ValueBool()
		post.Enabled = &v
	}

	_, err := r.client.PostNfsExportPolicy(ctx, data.Name.ValueString(), post)
	if err != nil {
		resp.Diagnostics.AddError("Error creating NFS export policy", err.Error())
		return
	}

	r.readIntoState(ctx, data.Name.ValueString(), &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes Terraform state from the API.
func (r *nfsExportPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data nfsExportPolicyModel
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
	policy, err := r.client.GetNfsExportPolicy(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading NFS export policy", err.Error())
		return
	}

	// Drift detection on enabled field.
	if !data.Enabled.IsNull() && !data.Enabled.IsUnknown() {
		if data.Enabled.ValueBool() != policy.Enabled {
			tflog.Info(ctx, "drift detected on NFS export policy", map[string]any{
				"resource":    name,
				"field":       "enabled",
				"state_value": data.Enabled.ValueBool(),
				"api_value":   policy.Enabled,
			})
		}
	}

	mapNFSPolicyToModel(policy, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update applies changes to an existing NFS export policy.
func (r *nfsExportPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state nfsExportPolicyModel
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

	// Use OLD name to address the policy in the PATCH request.
	oldName := state.Name.ValueString()
	patch := client.NfsExportPolicyPatch{}

	if !plan.Name.Equal(state.Name) {
		v := plan.Name.ValueString()
		patch.Name = &v
	}
	if !plan.Enabled.Equal(state.Enabled) {
		v := plan.Enabled.ValueBool()
		patch.Enabled = &v
	}

	_, err := r.client.PatchNfsExportPolicy(ctx, oldName, patch)
	if err != nil {
		resp.Diagnostics.AddError("Error updating NFS export policy", err.Error())
		return
	}

	// After rename the policy is now known by the new name.
	newName := plan.Name.ValueString()
	r.readIntoState(ctx, newName, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes an NFS export policy.
// Blocked if the policy is currently attached to one or more file systems.
func (r *nfsExportPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data nfsExportPolicyModel
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
	members, err := r.client.ListNfsExportPolicyMembers(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError("Error checking policy members before deletion", err.Error())
		return
	}
	if len(members) > 0 {
		resp.Diagnostics.AddError(
			"Cannot delete NFS export policy",
			fmt.Sprintf("Policy %q is attached to %d file system(s). Detach the policy from all file systems before deleting.", name, len(members)),
		)
		return
	}

	if err := r.client.DeleteNfsExportPolicy(ctx, name); err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting NFS export policy", err.Error())
		return
	}
}

// ImportState imports an existing NFS export policy by name.
func (r *nfsExportPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	name := req.ID

	var data nfsExportPolicyModel
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

// readIntoState calls GetNfsExportPolicy and maps the result into the provided model.
func (r *nfsExportPolicyResource) readIntoState(ctx context.Context, name string, data *nfsExportPolicyModel, diags DiagnosticReporter) {
	policy, err := r.client.GetNfsExportPolicy(ctx, name)
	if err != nil {
		diags.AddError("Error reading NFS export policy after write", err.Error())
		return
	}
	mapNFSPolicyToModel(policy, data)
}

// mapNFSPolicyToModel maps a client.NfsExportPolicy to an nfsExportPolicyModel.
// It preserves user-managed fields (Timeouts).
func mapNFSPolicyToModel(policy *client.NfsExportPolicy, data *nfsExportPolicyModel) {
	data.ID = types.StringValue(policy.ID)
	data.Name = types.StringValue(policy.Name)
	data.Enabled = types.BoolValue(policy.Enabled)
	data.IsLocal = types.BoolValue(policy.IsLocal)
	data.PolicyType = types.StringValue(policy.PolicyType)
	data.Version = types.StringValue(policy.Version)
}
