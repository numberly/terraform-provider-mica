package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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

// Ensure networkAccessPolicyResource satisfies the resource interfaces.
var _ resource.Resource = &networkAccessPolicyResource{}
var _ resource.ResourceWithConfigure = &networkAccessPolicyResource{}
var _ resource.ResourceWithImportState = &networkAccessPolicyResource{}
var _ resource.ResourceWithUpgradeState = &networkAccessPolicyResource{}

// networkAccessPolicyResource implements the flashblade_network_access_policy resource.
// NAP policies are system-managed singletons: no POST or DELETE at the policy level.
// Create=GET+PATCH, Delete=PATCH-to-reset (set enabled=false).
type networkAccessPolicyResource struct {
	client *client.FlashBladeClient
}

// NewNetworkAccessPolicyResource is the factory function registered in the provider.
func NewNetworkAccessPolicyResource() resource.Resource {
	return &networkAccessPolicyResource{}
}

// ---------- model structs ----------------------------------------------------

// networkAccessPolicyModel is the top-level model for the flashblade_network_access_policy resource.
type networkAccessPolicyModel struct {
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
func (r *networkAccessPolicyResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_network_access_policy"
}

// Schema defines the resource schema.
func (r *networkAccessPolicyResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a FlashBlade network access policy (singleton). Network access policies are " +
			"system-managed — they cannot be created or deleted. Create adopts the existing policy via " +
			"GET+PATCH. Delete resets the policy to disabled state via PATCH.",
		Version: 0,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the network access policy.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the network access policy to manage. The policy must already exist on the FlashBlade array.",
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
				Description: "The type of the policy (e.g. 'network-access').",
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

func (r *networkAccessPolicyResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}

// Configure injects the FlashBladeClient into the resource.
func (r *networkAccessPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create adopts a system-managed NAP singleton via GET+PATCH.
// NAP policies cannot be created via POST — they are pre-provisioned by the array.
func (r *networkAccessPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data networkAccessPolicyModel
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

	// Step 1: Verify the policy exists (GET). NAP policies are system-managed singletons.
	_, err := r.client.GetNetworkAccessPolicy(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Network access policy not found",
				fmt.Sprintf("Network access policy %q was not found. NAP policies are system-managed and cannot be created — they must already exist on the FlashBlade array.", name),
			)
			return
		}
		resp.Diagnostics.AddError("Error reading network access policy", err.Error())
		return
	}

	// Step 2: PATCH with desired config.
	patch := client.NetworkAccessPolicyPatch{}
	if !data.Enabled.IsNull() && !data.Enabled.IsUnknown() {
		v := data.Enabled.ValueBool()
		patch.Enabled = &v
	}

	_, err = r.client.PatchNetworkAccessPolicy(ctx, name, patch)
	if err != nil {
		resp.Diagnostics.AddError("Error patching network access policy during create", err.Error())
		return
	}

	// Step 3: Read back full state.
	r.readIntoState(ctx, name, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes Terraform state from the API.
func (r *networkAccessPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data networkAccessPolicyModel
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
	policy, err := r.client.GetNetworkAccessPolicy(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading network access policy", err.Error())
		return
	}

	// Drift detection on enabled field.
	if !data.Enabled.IsNull() && !data.Enabled.IsUnknown() {
		if data.Enabled.ValueBool() != policy.Enabled {
			tflog.Info(ctx, "drift detected on network access policy", map[string]any{
				"resource":    name,
				"field":       "enabled",
				"state_value": data.Enabled.ValueBool(),
				"api_value":   policy.Enabled,
			})
		}
	}

	mapNAPToModel(policy, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update applies changes to an existing network access policy via PATCH.
func (r *networkAccessPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state networkAccessPolicyModel
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

	// Use old name to address the policy in the PATCH request.
	oldName := state.Name.ValueString()
	patch := client.NetworkAccessPolicyPatch{}

	if !plan.Name.Equal(state.Name) {
		v := plan.Name.ValueString()
		patch.Name = &v
	}
	if !plan.Enabled.Equal(state.Enabled) {
		v := plan.Enabled.ValueBool()
		patch.Enabled = &v
	}

	_, err := r.client.PatchNetworkAccessPolicy(ctx, oldName, patch)
	if err != nil {
		resp.Diagnostics.AddError("Error updating network access policy", err.Error())
		return
	}

	// After rename, the policy is known by the new name.
	newName := plan.Name.ValueString()
	r.readIntoState(ctx, newName, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete resets a NAP singleton to disabled state via PATCH.
// NAP policies cannot be deleted (no DELETE endpoint exists at the policy level).
func (r *networkAccessPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data networkAccessPolicyModel
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

	// Reset the singleton to disabled state — this is the "destroy" for a singleton.
	// No DELETE endpoint exists: PATCH to disabled is the closest equivalent.
	disabled := false
	patch := client.NetworkAccessPolicyPatch{Enabled: &disabled}
	_, err := r.client.PatchNetworkAccessPolicy(ctx, name, patch)
	if err != nil {
		if client.IsNotFound(err) {
			// Already gone — idempotent.
			return
		}
		resp.Diagnostics.AddError("Error resetting network access policy to disabled state", err.Error())
		return
	}

	tflog.Info(ctx, "network access policy reset to disabled state (singleton — no DELETE endpoint exists)", map[string]any{
		"name": name,
	})
}

// ImportState imports a network access policy by name.
func (r *networkAccessPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	name := req.ID

	var data networkAccessPolicyModel
	// Initialize timeouts with a null value so the framework can serialize it.
	data.Timeouts = timeouts.Value{
		Object: types.ObjectNull(map[string]attr.Type{
			"create": types.StringType,
			"read":   types.StringType,
			"update": types.StringType,
			"delete": types.StringType,
		}),
	}
	data.Name = types.StringValue(name)

	r.readIntoState(ctx, name, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ---------- helpers ---------------------------------------------------------

// readIntoState calls GetNetworkAccessPolicy and maps the result into the provided model.
func (r *networkAccessPolicyResource) readIntoState(ctx context.Context, name string, data *networkAccessPolicyModel, diags interface {
	AddError(string, string)
	HasError() bool
}) {
	policy, err := r.client.GetNetworkAccessPolicy(ctx, name)
	if err != nil {
		diags.AddError("Error reading network access policy after write", err.Error())
		return
	}
	mapNAPToModel(policy, data)
}

// mapNAPToModel maps a client.NetworkAccessPolicy to a networkAccessPolicyModel.
// Preserves user-managed fields (Timeouts).
func mapNAPToModel(policy *client.NetworkAccessPolicy, data *networkAccessPolicyModel) {
	data.ID = types.StringValue(policy.ID)
	data.Name = types.StringValue(policy.Name)
	data.Enabled = types.BoolValue(policy.Enabled)
	data.IsLocal = types.BoolValue(policy.IsLocal)
	data.PolicyType = types.StringValue(policy.PolicyType)
	data.Version = types.StringValue(policy.Version)
}
