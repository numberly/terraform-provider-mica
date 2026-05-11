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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

var _ resource.Resource = &smbClientPolicyResource{}
var _ resource.ResourceWithConfigure = &smbClientPolicyResource{}
var _ resource.ResourceWithImportState = &smbClientPolicyResource{}
var _ resource.ResourceWithUpgradeState = &smbClientPolicyResource{}

// smbClientPolicyResource implements the flashblade_smb_client_policy resource.
type smbClientPolicyResource struct {
	client *client.FlashBladeClient
}

func NewSmbClientPolicyResource() resource.Resource {
	return &smbClientPolicyResource{}
}

// ---------- model structs ----------------------------------------------------

// smbClientPolicyV0Model is the schema-v0 model (prior to API 2.23 workload field).
type smbClientPolicyV0Model struct {
	ID                            types.String   `tfsdk:"id"`
	Name                          types.String   `tfsdk:"name"`
	Enabled                       types.Bool     `tfsdk:"enabled"`
	IsLocal                       types.Bool     `tfsdk:"is_local"`
	PolicyType                    types.String   `tfsdk:"policy_type"`
	Version                       types.String   `tfsdk:"version"`
	AccessBasedEnumerationEnabled types.Bool     `tfsdk:"access_based_enumeration_enabled"`
	Timeouts                      timeouts.Value `tfsdk:"timeouts"`
}

// smbClientPolicyModel is the top-level model for the flashblade_smb_client_policy resource (schema v1).
type smbClientPolicyModel struct {
	ID                            types.String   `tfsdk:"id"`
	Name                          types.String   `tfsdk:"name"`
	Enabled                       types.Bool     `tfsdk:"enabled"`
	IsLocal                       types.Bool     `tfsdk:"is_local"`
	PolicyType                    types.String   `tfsdk:"policy_type"`
	Version                       types.String   `tfsdk:"version"`
	AccessBasedEnumerationEnabled types.Bool     `tfsdk:"access_based_enumeration_enabled"`
	Workload                      types.Object   `tfsdk:"workload"`
	Timeouts                      timeouts.Value `tfsdk:"timeouts"`
}

// ---------- resource interface methods --------------------------------------

func (r *smbClientPolicyResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_smb_client_policy"
}

// Schema defines the resource schema.
func (r *smbClientPolicyResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     1,
		Description: "Manages a FlashBlade SMB client policy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the SMB client policy.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the SMB client policy. Can be changed in-place via PATCH (rename).",
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
			"version": schema.StringAttribute{
				Computed:    true,
				Description: "The version of the SMB client policy (read-only, server-assigned).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"access_based_enumeration_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "If true, access-based enumeration is enabled for this policy.",
			},
			"workload": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "The workload that owns this SMB client policy (read-only, API-managed). Populated by the API when the policy is associated with a workload.",
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed:    true,
						Description: "The workload unique identifier.",
					},
					"name": schema.StringAttribute{
						Computed:    true,
						Description: "The workload name.",
					},
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
func (r *smbClientPolicyResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		// PriorSchema verified against smbClientPolicyV0Model fields:
		//   ID (Computed), Name (Required), Enabled (Optional+Computed),
		//   IsLocal (Computed), PolicyType (Computed), Version (Computed),
		//   AccessBasedEnumerationEnabled (Optional+Computed), Timeouts (Optional)
		0: {
			PriorSchema: &schema.Schema{
				Version:     0,
				Description: "Manages a FlashBlade SMB client policy.",
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed:    true,
						Description: "The unique identifier of the SMB client policy.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"name": schema.StringAttribute{
						Required:    true,
						Description: "The name of the SMB client policy. Can be changed in-place via PATCH (rename).",
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
					"version": schema.StringAttribute{
						Computed:    true,
						Description: "The version of the SMB client policy (read-only, server-assigned).",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"access_based_enumeration_enabled": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						Description: "If true, access-based enumeration is enabled for this policy.",
					},
					"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
						Create: true,
						Read:   true,
						Update: true,
						Delete: true,
					}),
				},
			},
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var prior smbClientPolicyV0Model
				resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
				if resp.Diagnostics.HasError() {
					return
				}

				next := smbClientPolicyModel{
					ID:                            prior.ID,
					Name:                          prior.Name,
					Enabled:                       prior.Enabled,
					IsLocal:                       prior.IsLocal,
					PolicyType:                    prior.PolicyType,
					Version:                       prior.Version,
					AccessBasedEnumerationEnabled: prior.AccessBasedEnumerationEnabled,
					// workload: new computed field in v1 — null until Read hydrates from API.
					Workload: types.ObjectNull(smbClientPolicyWorkloadAttrTypes()),
					Timeouts: prior.Timeouts,
				}

				resp.Diagnostics.Append(resp.State.Set(ctx, next)...)
			},
		},
	}
}

// Configure injects the FlashBladeClient into the resource.
func (r *smbClientPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *smbClientPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data smbClientPolicyModel
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

	post := client.SmbClientPolicyPost{}
	if !data.Enabled.IsNull() && !data.Enabled.IsUnknown() {
		v := data.Enabled.ValueBool()
		post.Enabled = &v
	}
	if !data.AccessBasedEnumerationEnabled.IsNull() && !data.AccessBasedEnumerationEnabled.IsUnknown() {
		v := data.AccessBasedEnumerationEnabled.ValueBool()
		post.AccessBasedEnumerationEnabled = &v
	}

	_, err := r.client.PostSmbClientPolicy(ctx, data.Name.ValueString(), post)
	if err != nil {
		resp.Diagnostics.AddError("Error creating SMB client policy", err.Error())
		return
	}

	resp.Diagnostics.Append(r.readIntoState(ctx, data.Name.ValueString(), &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *smbClientPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data smbClientPolicyModel
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
	policy, err := r.client.GetSmbClientPolicy(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading SMB client policy", err.Error())
		return
	}

	// Drift detection on enabled field.
	if !data.Enabled.IsNull() && !data.Enabled.IsUnknown() {
		if data.Enabled.ValueBool() != policy.Enabled {
			tflog.Debug(ctx, "drift detected on SMB client policy", map[string]any{
				"resource": name,
				"field":    "enabled",
				"was":      data.Enabled.ValueBool(),
				"now":      policy.Enabled,
			})
		}
	}

	// Drift detection on access_based_enumeration_enabled field.
	if !data.AccessBasedEnumerationEnabled.IsNull() && !data.AccessBasedEnumerationEnabled.IsUnknown() {
		if data.AccessBasedEnumerationEnabled.ValueBool() != policy.AccessBasedEnumerationEnabled {
			tflog.Debug(ctx, "drift detected on SMB client policy", map[string]any{
				"resource": name,
				"field":    "access_based_enumeration_enabled",
				"was":      data.AccessBasedEnumerationEnabled.ValueBool(),
				"now":      policy.AccessBasedEnumerationEnabled,
			})
		}
	}

	// Drift detection on workload field (computed — can change if policy migrated).
	if !data.Workload.IsNull() && !data.Workload.IsUnknown() {
		var prev struct {
			ID   string `tfsdk:"id"`
			Name string `tfsdk:"name"`
		}
		if diags := data.Workload.As(ctx, &prev, basetypes.ObjectAsOptions{}); !diags.HasError() {
			newID := ""
			newName := ""
			if policy.Workload != nil {
				newID = policy.Workload.ID
				newName = policy.Workload.Name
			}
			if prev.ID != newID || prev.Name != newName {
				tflog.Debug(ctx, "drift detected on SMB client policy", map[string]any{
					"resource": name,
					"field":    "workload",
					"was":      fmt.Sprintf("{id:%s name:%s}", prev.ID, prev.Name),
					"now":      fmt.Sprintf("{id:%s name:%s}", newID, newName),
				})
			}
		}
	}

	mapSMBClientPolicyToModel(policy, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update applies changes to an existing SMB client policy.
func (r *smbClientPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state smbClientPolicyModel
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
	patch := client.SmbClientPolicyPatch{}

	if !plan.Name.Equal(state.Name) {
		v := plan.Name.ValueString()
		patch.Name = &v
	}
	if !plan.Enabled.Equal(state.Enabled) {
		v := plan.Enabled.ValueBool()
		patch.Enabled = &v
	}
	if !plan.AccessBasedEnumerationEnabled.Equal(state.AccessBasedEnumerationEnabled) {
		v := plan.AccessBasedEnumerationEnabled.ValueBool()
		patch.AccessBasedEnumerationEnabled = &v
	}

	_, err := r.client.PatchSmbClientPolicy(ctx, oldName, patch)
	if err != nil {
		resp.Diagnostics.AddError("Error updating SMB client policy", err.Error())
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

// Delete removes an SMB client policy.
// Blocked if the policy is currently attached to one or more file systems.
func (r *smbClientPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data smbClientPolicyModel
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
	members, err := r.client.ListSmbClientPolicyMembers(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError("Error checking policy members before deletion", err.Error())
		return
	}
	if len(members) > 0 {
		resp.Diagnostics.AddError(
			"Cannot delete SMB client policy",
			fmt.Sprintf("Policy %q is attached to %d file system(s). Detach the policy from all file systems before deleting.", name, len(members)),
		)
		return
	}

	if err := r.client.DeleteSmbClientPolicy(ctx, name); err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting SMB client policy", err.Error())
		return
	}
}

func (r *smbClientPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	name := req.ID

	var data smbClientPolicyModel
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

// readIntoState calls GetSmbClientPolicy and maps the result into the provided model.
func (r *smbClientPolicyResource) readIntoState(ctx context.Context, name string, data *smbClientPolicyModel) diag.Diagnostics {
	var diags diag.Diagnostics

	policy, err := r.client.GetSmbClientPolicy(ctx, name)
	if err != nil {
		diags.AddError("Error reading SMB client policy after write", err.Error())
		return diags
	}
	mapSMBClientPolicyToModel(policy, data)
	return diags
}

// smbClientPolicyWorkloadAttrTypes returns the attribute types for the workload object.
func smbClientPolicyWorkloadAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":   types.StringType,
		"name": types.StringType,
	}
}

// mapSMBClientPolicyToModel maps a client.SmbClientPolicy to an smbClientPolicyModel.
// It preserves user-managed fields (Timeouts).
func mapSMBClientPolicyToModel(policy *client.SmbClientPolicy, data *smbClientPolicyModel) {
	data.ID = types.StringValue(policy.ID)
	data.Name = types.StringValue(policy.Name)
	data.Enabled = types.BoolValue(policy.Enabled)
	data.IsLocal = types.BoolValue(policy.IsLocal)
	data.PolicyType = types.StringValue(policy.PolicyType)
	data.Version = types.StringValue(policy.Version)
	data.AccessBasedEnumerationEnabled = types.BoolValue(policy.AccessBasedEnumerationEnabled)

	// Workload — set if present in API response, null otherwise (Computed-only).
	if policy.Workload != nil {
		workloadObj, _ := types.ObjectValue(smbClientPolicyWorkloadAttrTypes(), map[string]attr.Value{
			"id":   types.StringValue(policy.Workload.ID),
			"name": types.StringValue(policy.Workload.Name),
		})
		data.Workload = workloadObj
	} else {
		data.Workload = types.ObjectNull(smbClientPolicyWorkloadAttrTypes())
	}
}
