package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

var _ resource.Resource = &smbSharePolicyResource{}
var _ resource.ResourceWithConfigure = &smbSharePolicyResource{}
var _ resource.ResourceWithImportState = &smbSharePolicyResource{}
var _ resource.ResourceWithUpgradeState = &smbSharePolicyResource{}

// smbSharePolicyResource implements the flashblade_smb_share_policy resource.
type smbSharePolicyResource struct {
	client *client.FlashBladeClient
}

func NewSmbSharePolicyResource() resource.Resource {
	return &smbSharePolicyResource{}
}

// ---------- model structs ----------------------------------------------------

// smbSharePolicyModel is the top-level model for the flashblade_smb_share_policy resource.
// Note: SMB share policy has no Version field (unlike NFS export policy).
type smbSharePolicyModel struct {
	ID         types.String   `tfsdk:"id"`
	Name       types.String   `tfsdk:"name"`
	Enabled    types.Bool     `tfsdk:"enabled"`
	IsLocal    types.Bool     `tfsdk:"is_local"`
	PolicyType types.String   `tfsdk:"policy_type"`
	Timeouts   timeouts.Value `tfsdk:"timeouts"`
}

// ---------- resource interface methods --------------------------------------

func (r *smbSharePolicyResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_smb_share_policy"
}

// Schema defines the resource schema.
func (r *smbSharePolicyResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     0,
		Description: "Manages a FlashBlade SMB share policy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the SMB share policy.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the SMB share policy. Can be changed in-place via PATCH (rename).",
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
				Description: "The type of the policy (e.g. 'smb').",
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


// UpgradeState returns state upgraders for schema migrations.
func (r *smbSharePolicyResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}

// Configure injects the FlashBladeClient into the resource.
func (r *smbSharePolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *smbSharePolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data smbSharePolicyModel
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

	post := client.SmbSharePolicyPost{}
	if !data.Enabled.IsNull() && !data.Enabled.IsUnknown() {
		v := data.Enabled.ValueBool()
		post.Enabled = &v
	}

	_, err := r.client.PostSmbSharePolicy(ctx, data.Name.ValueString(), post)
	if err != nil {
		resp.Diagnostics.AddError("Error creating SMB share policy", err.Error())
		return
	}

	resp.Diagnostics.Append(r.readIntoState(ctx, data.Name.ValueString(), &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *smbSharePolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data smbSharePolicyModel
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
	policy, err := r.client.GetSmbSharePolicy(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading SMB share policy", err.Error())
		return
	}

	// Drift detection on enabled field.
	if !data.Enabled.IsNull() && !data.Enabled.IsUnknown() {
		if data.Enabled.ValueBool() != policy.Enabled {
			tflog.Debug(ctx, "drift detected on SMB share policy", map[string]any{
				"resource":    name,
				"field":       "enabled",
				"was":         data.Enabled.ValueBool(),
				"now":           policy.Enabled,
			})
		}
	}

	mapSMBPolicyToModel(policy, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update applies changes to an existing SMB share policy.
func (r *smbSharePolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state smbSharePolicyModel
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
	patch := client.SmbSharePolicyPatch{}

	if !plan.Name.Equal(state.Name) {
		v := plan.Name.ValueString()
		patch.Name = &v
	}
	if !plan.Enabled.Equal(state.Enabled) {
		v := plan.Enabled.ValueBool()
		patch.Enabled = &v
	}

	_, err := r.client.PatchSmbSharePolicy(ctx, oldName, patch)
	if err != nil {
		resp.Diagnostics.AddError("Error updating SMB share policy", err.Error())
		return
	}

	// After rename the policy is now known by the new name.
	newName := plan.Name.ValueString()
	resp.Diagnostics.Append(r.readIntoState(ctx, newName, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes an SMB share policy.
// Blocked if the policy is currently attached to one or more file systems.
func (r *smbSharePolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data smbSharePolicyModel
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
	members, err := r.client.ListSmbSharePolicyMembers(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError("Error checking policy members before deletion", err.Error())
		return
	}
	if len(members) > 0 {
		resp.Diagnostics.AddError(
			"Cannot delete SMB share policy",
			fmt.Sprintf("Policy %q is attached to %d file system(s). Detach the policy from all file systems before deleting.", name, len(members)),
		)
		return
	}

	if err := r.client.DeleteSmbSharePolicy(ctx, name); err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting SMB share policy", err.Error())
		return
	}
}

func (r *smbSharePolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	name := req.ID

	var data smbSharePolicyModel
	// Initialize timeouts with a null value so the framework can serialize it.
	data.Timeouts = nullTimeoutsValue()
	// Set Name so Read can look up the policy.
	data.Name = types.StringValue(name)

	resp.Diagnostics.Append(r.readIntoState(ctx, name, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ---------- helpers ---------------------------------------------------------

// readIntoState calls GetSmbSharePolicy and maps the result into the provided model.
func (r *smbSharePolicyResource) readIntoState(ctx context.Context, name string, data *smbSharePolicyModel) diag.Diagnostics {
	var diags diag.Diagnostics

	policy, err := r.client.GetSmbSharePolicy(ctx, name)
	if err != nil {
		diags.AddError("Error reading SMB share policy after write", err.Error())
		return diags
	}
	mapSMBPolicyToModel(policy, data)
	return diags
}


// mapSMBPolicyToModel maps a client.SmbSharePolicy to an smbSharePolicyModel.
// It preserves user-managed fields (Timeouts).
func mapSMBPolicyToModel(policy *client.SmbSharePolicy, data *smbSharePolicyModel) {
	data.ID = types.StringValue(policy.ID)
	data.Name = types.StringValue(policy.Name)
	data.Enabled = types.BoolValue(policy.Enabled)
	data.IsLocal = types.BoolValue(policy.IsLocal)
	data.PolicyType = types.StringValue(policy.PolicyType)
}
