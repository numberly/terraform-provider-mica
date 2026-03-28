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

// Ensure s3ExportPolicyResource satisfies the resource interfaces.
var _ resource.Resource = &s3ExportPolicyResource{}
var _ resource.ResourceWithConfigure = &s3ExportPolicyResource{}
var _ resource.ResourceWithImportState = &s3ExportPolicyResource{}

// s3ExportPolicyResource implements the flashblade_s3_export_policy resource.
type s3ExportPolicyResource struct {
	client *client.FlashBladeClient
}

// NewS3ExportPolicyResource is the factory function registered in the provider.
func NewS3ExportPolicyResource() resource.Resource {
	return &s3ExportPolicyResource{}
}

// ---------- model structs ----------------------------------------------------

// s3ExportPolicyModel is the top-level model for the flashblade_s3_export_policy resource.
type s3ExportPolicyModel struct {
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
func (r *s3ExportPolicyResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_s3_export_policy"
}

// Schema defines the resource schema.
func (r *s3ExportPolicyResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a FlashBlade S3 export policy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the S3 export policy.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the S3 export policy. Can be changed in-place via PATCH (rename).",
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
				Description: "The type of the policy (e.g. 's3-export').",
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

// Configure injects the FlashBladeClient into the resource.
func (r *s3ExportPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new S3 export policy.
func (r *s3ExportPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data s3ExportPolicyModel
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

	post := client.S3ExportPolicyPost{}
	if !data.Enabled.IsNull() && !data.Enabled.IsUnknown() {
		v := data.Enabled.ValueBool()
		post.Enabled = &v
	}

	_, err := r.client.PostS3ExportPolicy(ctx, data.Name.ValueString(), post)
	if err != nil {
		resp.Diagnostics.AddError("Error creating S3 export policy", err.Error())
		return
	}

	r.readIntoState(ctx, data.Name.ValueString(), &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes Terraform state from the API.
func (r *s3ExportPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data s3ExportPolicyModel
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
	policy, err := r.client.GetS3ExportPolicy(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading S3 export policy", err.Error())
		return
	}

	// Drift detection on enabled field.
	if !data.Enabled.IsNull() && !data.Enabled.IsUnknown() {
		if data.Enabled.ValueBool() != policy.Enabled {
			tflog.Info(ctx, "drift detected on S3 export policy", map[string]any{
				"resource":    name,
				"field":       "enabled",
				"state_value": data.Enabled.ValueBool(),
				"api_value":   policy.Enabled,
			})
		}
	}

	mapS3PolicyToModel(policy, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update applies changes to an existing S3 export policy.
func (r *s3ExportPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state s3ExportPolicyModel
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
	patch := client.S3ExportPolicyPatch{}

	if !plan.Name.Equal(state.Name) {
		v := plan.Name.ValueString()
		patch.Name = &v
	}
	if !plan.Enabled.Equal(state.Enabled) {
		v := plan.Enabled.ValueBool()
		patch.Enabled = &v
	}

	_, err := r.client.PatchS3ExportPolicy(ctx, oldName, patch)
	if err != nil {
		resp.Diagnostics.AddError("Error updating S3 export policy", err.Error())
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

// Delete removes an S3 export policy.
func (r *s3ExportPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data s3ExportPolicyModel
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

	if err := r.client.DeleteS3ExportPolicy(ctx, name); err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting S3 export policy", err.Error())
		return
	}
}

// ImportState imports an existing S3 export policy by name.
func (r *s3ExportPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	name := req.ID

	var data s3ExportPolicyModel
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

// readIntoState calls GetS3ExportPolicy and maps the result into the provided model.
func (r *s3ExportPolicyResource) readIntoState(ctx context.Context, name string, data *s3ExportPolicyModel, diags interface {
	AddError(string, string)
	HasError() bool
}) {
	policy, err := r.client.GetS3ExportPolicy(ctx, name)
	if err != nil {
		diags.AddError("Error reading S3 export policy after write", err.Error())
		return
	}
	mapS3PolicyToModel(policy, data)
}

// mapS3PolicyToModel maps a client.S3ExportPolicy to an s3ExportPolicyModel.
// It preserves user-managed fields (Timeouts).
func mapS3PolicyToModel(policy *client.S3ExportPolicy, data *s3ExportPolicyModel) {
	data.ID = types.StringValue(policy.ID)
	data.Name = types.StringValue(policy.Name)
	data.Enabled = types.BoolValue(policy.Enabled)
	data.IsLocal = types.BoolValue(policy.IsLocal)
	data.PolicyType = types.StringValue(policy.PolicyType)
	data.Version = types.StringValue(policy.Version)
}
