package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/soulkyu/terraform-provider-flashblade/internal/client"
)

// Ensure objectStoreAccessPolicyResource satisfies the resource interfaces.
var _ resource.Resource = &objectStoreAccessPolicyResource{}
var _ resource.ResourceWithConfigure = &objectStoreAccessPolicyResource{}
var _ resource.ResourceWithImportState = &objectStoreAccessPolicyResource{}

// objectStoreAccessPolicyResource implements the flashblade_object_store_access_policy resource.
type objectStoreAccessPolicyResource struct {
	client *client.FlashBladeClient
}

// NewObjectStoreAccessPolicyResource is the factory function registered in the provider.
func NewObjectStoreAccessPolicyResource() resource.Resource {
	return &objectStoreAccessPolicyResource{}
}

// ---------- model structs ----------------------------------------------------

// objectStoreAccessPolicyModel is the top-level model for the flashblade_object_store_access_policy resource.
type objectStoreAccessPolicyModel struct {
	ID          types.String   `tfsdk:"id"`
	Name        types.String   `tfsdk:"name"`
	Description types.String   `tfsdk:"description"`
	ARN         types.String   `tfsdk:"arn"`
	Enabled     types.Bool     `tfsdk:"enabled"`
	IsLocal     types.Bool     `tfsdk:"is_local"`
	PolicyType  types.String   `tfsdk:"policy_type"`
	Timeouts    timeouts.Value `tfsdk:"timeouts"`
}

// ---------- resource interface methods --------------------------------------

// Metadata sets the Terraform type name.
func (r *objectStoreAccessPolicyResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_object_store_access_policy"
}

// Schema defines the resource schema.
func (r *objectStoreAccessPolicyResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a FlashBlade object store access policy (IAM-style S3 policy).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the object store access policy.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the object store access policy. Can be renamed in-place via PATCH.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "A human-readable description. POST-only field — changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"arn": schema.StringAttribute{
				Computed:    true,
				Description: "The Amazon Resource Name (ARN) for the policy.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "If true, the policy is enabled. This is read-only (not writable via PATCH).",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
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
				Description: "The type of the policy (e.g. 'object-store-access').",
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
func (r *objectStoreAccessPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new object store access policy.
func (r *objectStoreAccessPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data objectStoreAccessPolicyModel
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

	post := client.ObjectStoreAccessPolicyPost{}
	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		post.Description = data.Description.ValueString()
	}

	_, err := r.client.PostObjectStoreAccessPolicy(ctx, data.Name.ValueString(), post)
	if err != nil {
		resp.Diagnostics.AddError("Error creating object store access policy", err.Error())
		return
	}

	r.readIntoState(ctx, data.Name.ValueString(), &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes Terraform state from the API.
func (r *objectStoreAccessPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data objectStoreAccessPolicyModel
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
	policy, err := r.client.GetObjectStoreAccessPolicy(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading object store access policy", err.Error())
		return
	}

	// Drift detection on name field.
	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		if data.Name.ValueString() != policy.Name {
			tflog.Info(ctx, "drift detected on object store access policy", map[string]any{
				"resource":    name,
				"field":       "name",
				"state_value": data.Name.ValueString(),
				"api_value":   policy.Name,
			})
		}
	}

	mapOAPToModel(policy, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update applies changes to an existing object store access policy.
// Only name is patchable (rename in-place). Description has RequiresReplace.
func (r *objectStoreAccessPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state objectStoreAccessPolicyModel
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
	patch := client.ObjectStoreAccessPolicyPatch{}

	if !plan.Name.Equal(state.Name) {
		v := plan.Name.ValueString()
		patch.Name = &v
	}

	_, err := r.client.PatchObjectStoreAccessPolicy(ctx, oldName, patch)
	if err != nil {
		resp.Diagnostics.AddError("Error updating object store access policy", err.Error())
		return
	}

	// After rename the policy is known by the new name.
	newName := plan.Name.ValueString()
	r.readIntoState(ctx, newName, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes an object store access policy.
// Blocked if the policy is currently attached to one or more buckets.
func (r *objectStoreAccessPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data objectStoreAccessPolicyModel
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

	// Guard: check for buckets using this policy before deleting.
	members, err := r.client.ListObjectStoreAccessPolicyMembers(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError("Error checking policy members before deletion", err.Error())
		return
	}
	if len(members) > 0 {
		resp.Diagnostics.AddError(
			"Cannot delete object store access policy",
			fmt.Sprintf("Policy %q is attached to %d bucket(s). Detach the policy from all buckets before deleting.", name, len(members)),
		)
		return
	}

	if err := r.client.DeleteObjectStoreAccessPolicy(ctx, name); err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting object store access policy", err.Error())
		return
	}
}

// ImportState imports an existing object store access policy by name.
func (r *objectStoreAccessPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	name := req.ID

	var data objectStoreAccessPolicyModel
	// Initialize timeouts with a null value so the framework can serialize it.
	data.Timeouts = timeouts.Value{
		Object: types.ObjectNull(map[string]attr.Type{
			"create": types.StringType,
			"read":   types.StringType,
			"update": types.StringType,
			"delete": types.StringType,
		}),
	}
	// Set Name so Read can look up the policy.
	data.Name = types.StringValue(name)

	r.readIntoState(ctx, name, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ---------- helpers ---------------------------------------------------------

// readIntoState calls GetObjectStoreAccessPolicy and maps the result into the provided model.
func (r *objectStoreAccessPolicyResource) readIntoState(ctx context.Context, name string, data *objectStoreAccessPolicyModel, diags interface {
	AddError(string, string)
	HasError() bool
}) {
	policy, err := r.client.GetObjectStoreAccessPolicy(ctx, name)
	if err != nil {
		diags.AddError("Error reading object store access policy after write", err.Error())
		return
	}
	mapOAPToModel(policy, data)
}

// mapOAPToModel maps a client.ObjectStoreAccessPolicy to an objectStoreAccessPolicyModel.
// It preserves user-managed fields (Timeouts).
func mapOAPToModel(policy *client.ObjectStoreAccessPolicy, data *objectStoreAccessPolicyModel) {
	data.ID = types.StringValue(policy.ID)
	data.Name = types.StringValue(policy.Name)
	data.ARN = types.StringValue(policy.ARN)
	data.Enabled = types.BoolValue(policy.Enabled)
	data.IsLocal = types.BoolValue(policy.IsLocal)
	data.PolicyType = types.StringValue(policy.PolicyType)

	if policy.Description != "" {
		data.Description = types.StringValue(policy.Description)
	} else if data.Description.IsUnknown() {
		data.Description = types.StringNull()
	}
}
