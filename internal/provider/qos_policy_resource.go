package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

var _ resource.Resource = &qosPolicyResource{}
var _ resource.ResourceWithConfigure = &qosPolicyResource{}
var _ resource.ResourceWithImportState = &qosPolicyResource{}
var _ resource.ResourceWithUpgradeState = &qosPolicyResource{}

// qosPolicyResource implements the flashblade_qos_policy resource.
type qosPolicyResource struct {
	client *client.FlashBladeClient
}

func NewQosPolicyResource() resource.Resource {
	return &qosPolicyResource{}
}

// ---------- model structs ----------------------------------------------------

// qosPolicyModel is the Terraform state model for the flashblade_qos_policy resource.
type qosPolicyModel struct {
	ID                  types.String   `tfsdk:"id"`
	Name                types.String   `tfsdk:"name"`
	Enabled             types.Bool     `tfsdk:"enabled"`
	MaxTotalBytesPerSec types.Int64    `tfsdk:"max_total_bytes_per_sec"`
	MaxTotalOpsPerSec   types.Int64    `tfsdk:"max_total_ops_per_sec"`
	IsLocal             types.Bool     `tfsdk:"is_local"`
	PolicyType          types.String   `tfsdk:"policy_type"`
	Timeouts            timeouts.Value `tfsdk:"timeouts"`
}

// ---------- resource interface methods --------------------------------------

func (r *qosPolicyResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_qos_policy"
}

// Schema defines the resource schema.
func (r *qosPolicyResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     1,
		Description: "Manages a FlashBlade QoS policy for enforcing bandwidth and IOPS limits on buckets and file systems.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the QoS policy.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the QoS policy. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether the QoS policy is enabled. Defaults to true.",
			},
			"max_total_bytes_per_sec": schema.Int64Attribute{
				Optional:    true,
				Description: "Maximum total bandwidth in bytes per second.",
			},
			"max_total_ops_per_sec": schema.Int64Attribute{
				Optional:    true,
				Description: "Maximum total operations (IOPS) per second.",
			},
			"is_local": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the QoS policy is local to this array. Read-only.",
			},
			"policy_type": schema.StringAttribute{
				Computed:    true,
				Description: "The type of the QoS policy (e.g. bandwidth-limit). Read-only.",
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

// qosPolicyV0Model mirrors the resource state at schema Version 0. The shape is
// identical to the current model since the v0→v1 migration only changes wire
// semantics in QosPolicyPost (MaxTotal* int64 → *int64) — no Terraform attribute
// was added, removed, or retyped. See D-52-01.
type qosPolicyV0Model struct {
	ID                  types.String   `tfsdk:"id"`
	Name                types.String   `tfsdk:"name"`
	Enabled             types.Bool     `tfsdk:"enabled"`
	MaxTotalBytesPerSec types.Int64    `tfsdk:"max_total_bytes_per_sec"`
	MaxTotalOpsPerSec   types.Int64    `tfsdk:"max_total_ops_per_sec"`
	IsLocal             types.Bool     `tfsdk:"is_local"`
	PolicyType          types.String   `tfsdk:"policy_type"`
	Timeouts            timeouts.Value `tfsdk:"timeouts"`
}

// UpgradeState returns state upgraders for schema migrations.
// v0→v1: no-op identity — the Terraform attribute set is unchanged. The bump exists
// because wire-format semantics changed (QosPolicyPost.MaxTotal* int64 → *int64 so
// the 0=unlimited value is preserved through POST). See R-006 / D-52-01.
func (r *qosPolicyResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema: &schema.Schema{
				Version:     0,
				Description: "Manages a FlashBlade QoS policy for enforcing bandwidth and IOPS limits on buckets and file systems.",
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"name": schema.StringAttribute{
						Required: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"enabled": schema.BoolAttribute{
						Optional: true,
						Computed: true,
						Default:  booldefault.StaticBool(true),
					},
					"max_total_bytes_per_sec": schema.Int64Attribute{Optional: true},
					"max_total_ops_per_sec":   schema.Int64Attribute{Optional: true},
					"is_local":                schema.BoolAttribute{Computed: true},
					"policy_type": schema.StringAttribute{
						Computed: true,
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
			},
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var oldState qosPolicyV0Model
				resp.Diagnostics.Append(req.State.Get(ctx, &oldState)...)
				if resp.Diagnostics.HasError() {
					return
				}
				// Identity copy — no attribute shape change.
				newState := qosPolicyModel(oldState)
				resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
			},
		},
	}
}

// Configure injects the FlashBladeClient into the resource.
func (r *qosPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *qosPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data qosPolicyModel
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

	body := client.QosPolicyPost{
		Name: data.Name.ValueString(),
	}

	if !data.Enabled.IsNull() && !data.Enabled.IsUnknown() {
		v := data.Enabled.ValueBool()
		body.Enabled = &v
	}
	if !data.MaxTotalBytesPerSec.IsNull() && !data.MaxTotalBytesPerSec.IsUnknown() {
		v := data.MaxTotalBytesPerSec.ValueInt64()
		body.MaxTotalBytesPerSec = &v
	}
	if !data.MaxTotalOpsPerSec.IsNull() && !data.MaxTotalOpsPerSec.IsUnknown() {
		v := data.MaxTotalOpsPerSec.ValueInt64()
		body.MaxTotalOpsPerSec = &v
	}

	policy, err := r.client.PostQosPolicy(ctx, data.Name.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Error creating QoS policy", err.Error())
		return
	}

	mapQosPolicyToModel(policy, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *qosPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data qosPolicyModel
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

	policy, err := r.client.GetQosPolicy(ctx, data.Name.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading QoS policy", err.Error())
		return
	}

	mapQosPolicyToModel(policy, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update applies changes to an existing QoS policy.
func (r *qosPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state qosPolicyModel
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

	patch := client.QosPolicyPatch{}
	needsPatch := false

	if !plan.Enabled.Equal(state.Enabled) {
		v := plan.Enabled.ValueBool()
		patch.Enabled = &v
		needsPatch = true
	}
	if !plan.MaxTotalBytesPerSec.Equal(state.MaxTotalBytesPerSec) {
		v := plan.MaxTotalBytesPerSec.ValueInt64()
		patch.MaxTotalBytesPerSec = &v
		needsPatch = true
	}
	if !plan.MaxTotalOpsPerSec.Equal(state.MaxTotalOpsPerSec) {
		v := plan.MaxTotalOpsPerSec.ValueInt64()
		patch.MaxTotalOpsPerSec = &v
		needsPatch = true
	}

	if needsPatch {
		_, err := r.client.PatchQosPolicy(ctx, state.Name.ValueString(), patch)
		if err != nil {
			resp.Diagnostics.AddError("Error updating QoS policy", err.Error())
			return
		}
	}

	// Re-read to refresh computed fields.
	policy, err := r.client.GetQosPolicy(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading QoS policy after update", err.Error())
		return
	}

	mapQosPolicyToModel(policy, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes a QoS policy.
func (r *qosPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data qosPolicyModel
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

	err := r.client.DeleteQosPolicy(ctx, data.Name.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting QoS policy", err.Error())
		return
	}
}

func (r *qosPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	policy, err := r.client.GetQosPolicy(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing QoS policy", err.Error())
		return
	}

	var data qosPolicyModel
	data.Timeouts = nullTimeoutsValue()

	mapQosPolicyToModel(policy, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ---------- helpers ---------------------------------------------------------

// mapQosPolicyToModel maps a client.QosPolicy to the Terraform model.
func mapQosPolicyToModel(policy *client.QosPolicy, data *qosPolicyModel) {
	data.ID = types.StringValue(policy.ID)
	data.Name = types.StringValue(policy.Name)
	data.Enabled = types.BoolValue(policy.Enabled)
	data.IsLocal = types.BoolValue(policy.IsLocal)
	data.PolicyType = types.StringValue(policy.PolicyType)

	if policy.MaxTotalBytesPerSec != 0 {
		data.MaxTotalBytesPerSec = types.Int64Value(policy.MaxTotalBytesPerSec)
	} else if data.MaxTotalBytesPerSec.IsNull() || data.MaxTotalBytesPerSec.IsUnknown() {
		data.MaxTotalBytesPerSec = types.Int64Null()
	}

	if policy.MaxTotalOpsPerSec != 0 {
		data.MaxTotalOpsPerSec = types.Int64Value(policy.MaxTotalOpsPerSec)
	} else if data.MaxTotalOpsPerSec.IsNull() || data.MaxTotalOpsPerSec.IsUnknown() {
		data.MaxTotalOpsPerSec = types.Int64Null()
	}
}
