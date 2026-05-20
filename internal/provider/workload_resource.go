package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

var _ resource.Resource = &workloadResource{}
var _ resource.ResourceWithConfigure = &workloadResource{}
var _ resource.ResourceWithImportState = &workloadResource{}
var _ resource.ResourceWithUpgradeState = &workloadResource{}

// workloadResource implements the flashblade_workload resource.
type workloadResource struct {
	client *client.FlashBladeClient
}

// workloadModel is the top-level Terraform state model for flashblade_workload.
type workloadModel struct {
	ID                     types.String `tfsdk:"id"`
	Name                   types.String `tfsdk:"name"`
	PresetName             types.String `tfsdk:"preset_name"`
	Parameters             types.List   `tfsdk:"parameters"`
	DestroyEradicateOnDelete types.Bool   `tfsdk:"destroy_eradicate_on_delete"`
	Context                types.Object `tfsdk:"context"`
	Created                types.Int64  `tfsdk:"created"`
	Destroyed              types.Bool   `tfsdk:"destroyed"`
	Status                 types.String `tfsdk:"status"`
	StatusDetails          types.List   `tfsdk:"status_details"`
	TimeRemaining          types.Int64  `tfsdk:"time_remaining"`
	Timeouts               timeouts.Value `tfsdk:"timeouts"`
}

// workloadParameterModel is the model for a single workload parameter entry.
type workloadParameterModel struct {
	Name              types.String `tfsdk:"name"`
	ValueString       types.String `tfsdk:"value_string"`
	ValueBool         types.Bool   `tfsdk:"value_bool"`
	ValueInteger      types.Int64  `tfsdk:"value_integer"`
	ValueResourceName types.String `tfsdk:"value_resource_name"`
	ValueResourceID   types.String `tfsdk:"value_resource_id"`
	ValueResourceType types.String `tfsdk:"value_resource_type"`
}

func workloadParameterAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":                types.StringType,
		"value_string":        types.StringType,
		"value_bool":          types.BoolType,
		"value_integer":       types.Int64Type,
		"value_resource_name": types.StringType,
		"value_resource_id":   types.StringType,
		"value_resource_type": types.StringType,
	}
}

func workloadContextAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":   types.StringType,
		"name": types.StringType,
	}
}

// NewWorkloadResource creates a new workloadResource provider factory.
func NewWorkloadResource() resource.Resource {
	return &workloadResource{}
}

func (r *workloadResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_workload"
}

func (r *workloadResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     0,
		Description: "Manages a FlashBlade workload. A workload organises storage resources (volumes, file systems, etc.) and their related configuration objects into a logical grouping driven by a preset.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the workload.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the workload. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"preset_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the preset to deploy this workload from. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"parameters": schema.ListNestedAttribute{
				Optional:    true,
				Description: "Parameter values to pass to the preset when creating the workload. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "The name of the preset parameter.",
						},
						"value_string": schema.StringAttribute{
							Optional:    true,
							Description: "String value for this parameter.",
						},
						"value_bool": schema.BoolAttribute{
							Optional:    true,
							Description: "Boolean value for this parameter.",
						},
						"value_integer": schema.Int64Attribute{
							Optional:    true,
							Description: "Integer value for this parameter.",
						},
						"value_resource_name": schema.StringAttribute{
							Optional:    true,
							Description: "Resource reference name for this parameter.",
						},
						"value_resource_id": schema.StringAttribute{
							Optional:    true,
							Description: "Resource reference ID for this parameter.",
						},
						"value_resource_type": schema.StringAttribute{
							Optional:    true,
							Description: "Resource reference type for this parameter.",
						},
					},
				},
			},
			"destroy_eradicate_on_delete": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "When true, permanently eradicates the workload on destroy (two-phase: soft-delete then eradicate). When false, only soft-deletes the workload (leaves it in the destroyed queue).",
			},
			"context": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "The fleet context that owns this workload (read-only, API-managed).",
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed:    true,
						Description: "The context unique identifier.",
					},
					"name": schema.StringAttribute{
						Computed:    true,
						Description: "The context name.",
					},
				},
			},
			"created": schema.Int64Attribute{
				Computed:    true,
				Description: "The workload creation time, measured in milliseconds since the UNIX epoch.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"destroyed": schema.BoolAttribute{
				Computed:    true,
				Description: "True if the workload has been soft-deleted and is pending eradication.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The workload status (e.g. creating, ready, destroying, destroyed, eradicating, recovering).",
			},
			"status_details": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Additional information about the workload status.",
			},
			"time_remaining": schema.Int64Attribute{
				Computed:    true,
				Description: "Time remaining in milliseconds before the destroyed workload is permanently eradicated.",
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

func (r *workloadResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}

func (r *workloadResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *workloadResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data workloadModel
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
	presetName := data.PresetName.ValueString()

	post := client.WorkloadPost{}

	if !data.Parameters.IsNull() && !data.Parameters.IsUnknown() {
		var paramModels []workloadParameterModel
		resp.Diagnostics.Append(data.Parameters.ElementsAs(ctx, &paramModels, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		post.Parameters = convertParameterModelsToClient(paramModels)
	}

	if _, err := r.client.PostWorkload(ctx, name, presetName, post); err != nil {
		resp.Diagnostics.AddError("Error creating workload", err.Error())
		return
	}

	resp.Diagnostics.Append(r.readIntoState(ctx, name, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *workloadResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data workloadModel
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
	wl, err := r.client.GetWorkload(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading workload", err.Error())
		return
	}

	// Drift detection on computed/volatile fields.
	if data.Status.ValueString() != wl.Status {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": name, "field": "status",
			"was": data.Status.ValueString(), "now": wl.Status,
		})
	}
	if data.Destroyed.ValueBool() != wl.Destroyed {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": name, "field": "destroyed",
			"was": data.Destroyed.ValueBool(), "now": wl.Destroyed,
		})
	}
	if data.Created.ValueInt64() != wl.Created {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": name, "field": "created",
			"was": data.Created.ValueInt64(), "now": wl.Created,
		})
	}
	if data.TimeRemaining.ValueInt64() != wl.TimeRemaining {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": name, "field": "time_remaining",
			"was": data.TimeRemaining.ValueInt64(), "now": wl.TimeRemaining,
		})
	}

	resp.Diagnostics.Append(mapWorkloadToModel(wl, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *workloadResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data workloadModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateTimeout, diags := data.Timeouts.Update(ctx, 20*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	// The only mutable attribute (besides destroy_eradicate_on_delete which is
	// provider-local) is name, but name has RequiresReplace, so Update is only
	// ever called for destroy_eradicate_on_delete changes. No API call needed.
	name := data.Name.ValueString()
	resp.Diagnostics.Append(r.readIntoState(ctx, name, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *workloadResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data workloadModel
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
	eradicate := data.DestroyEradicateOnDelete.ValueBool()

	if err := r.client.DestroyAndEradicateWorkload(ctx, name, eradicate); err != nil {
		resp.Diagnostics.AddError("Error deleting workload", err.Error())
		return
	}
}

func (r *workloadResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	name := req.ID
	wl, err := r.client.GetWorkload(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError("Error importing workload", err.Error())
		return
	}

	var data workloadModel
	data.Timeouts = nullTimeoutsValue()
	data.DestroyEradicateOnDelete = types.BoolValue(false)
	// Parameters cannot be recovered on import (API doesn't echo them back).
	data.Parameters = types.ListValueMust(
		types.ObjectType{AttrTypes: workloadParameterAttrTypes()},
		[]attr.Value{},
	)

	resp.Diagnostics.Append(mapWorkloadToModel(wl, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve the preset name from API preset reference (may be null if preset deleted).
	if wl.Preset != nil && wl.Preset.Name != "" {
		data.PresetName = types.StringValue(wl.Preset.Name)
	} else {
		data.PresetName = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// readIntoState fetches the workload from the API and maps it into data.
func (r *workloadResource) readIntoState(ctx context.Context, name string, data *workloadModel) diag.Diagnostics {
	var diags diag.Diagnostics
	wl, err := r.client.GetWorkload(ctx, name)
	if err != nil {
		diags.AddError("Error reading workload after write", err.Error())
		return diags
	}
	diags.Append(mapWorkloadToModel(wl, data)...)
	return diags
}

// mapWorkloadToModel maps a client.Workload API response into the Terraform state model.
// It preserves provider-local fields (parameters, destroy_eradicate_on_delete, preset_name)
// that are not returned by the API.
func mapWorkloadToModel(wl *client.Workload, data *workloadModel) diag.Diagnostics {
	var diags diag.Diagnostics

	data.ID = types.StringValue(wl.ID)
	data.Name = types.StringValue(wl.Name)
	data.Created = types.Int64Value(wl.Created)
	data.Destroyed = types.BoolValue(wl.Destroyed)
	data.Status = types.StringValue(wl.Status)
	data.TimeRemaining = types.Int64Value(wl.TimeRemaining)

	// StatusDetails: always a list (never null).
	if wl.StatusDetails != nil {
		elems := make([]attr.Value, len(wl.StatusDetails))
		for i, s := range wl.StatusDetails {
			elems[i] = types.StringValue(s)
		}
		statusDetails, d := types.ListValue(types.StringType, elems)
		diags.Append(d...)
		data.StatusDetails = statusDetails
	} else {
		data.StatusDetails = types.ListValueMust(types.StringType, []attr.Value{})
	}

	// Context: optional reference.
	if wl.Context != nil {
		ctxObj, d := types.ObjectValue(workloadContextAttrTypes(), map[string]attr.Value{
			"id":   types.StringValue(wl.Context.ID),
			"name": types.StringValue(wl.Context.Name),
		})
		diags.Append(d...)
		data.Context = ctxObj
	} else {
		data.Context = types.ObjectNull(workloadContextAttrTypes())
	}

	return diags
}

// convertParameterModelsToClient converts Terraform parameter models to API structs.
func convertParameterModelsToClient(models []workloadParameterModel) []client.WorkloadParameter {
	params := make([]client.WorkloadParameter, len(models))
	for i, m := range models {
		p := client.WorkloadParameter{
			Name:  m.Name.ValueString(),
			Value: client.WorkloadParameterValue{},
		}
		if !m.ValueString.IsNull() && !m.ValueString.IsUnknown() {
			s := m.ValueString.ValueString()
			p.Value.String = &s
		}
		if !m.ValueBool.IsNull() && !m.ValueBool.IsUnknown() {
			b := m.ValueBool.ValueBool()
			p.Value.Boolean = &b
		}
		if !m.ValueInteger.IsNull() && !m.ValueInteger.IsUnknown() {
			n := m.ValueInteger.ValueInt64()
			p.Value.Integer = &n
		}
		if (!m.ValueResourceName.IsNull() && !m.ValueResourceName.IsUnknown()) ||
			(!m.ValueResourceID.IsNull() && !m.ValueResourceID.IsUnknown()) {
			ref := &client.WorkloadParameterResourceRef{}
			if !m.ValueResourceName.IsNull() && !m.ValueResourceName.IsUnknown() {
				ref.Name = m.ValueResourceName.ValueString()
			}
			if !m.ValueResourceID.IsNull() && !m.ValueResourceID.IsUnknown() {
				ref.ID = m.ValueResourceID.ValueString()
			}
			if !m.ValueResourceType.IsNull() && !m.ValueResourceType.IsUnknown() {
				ref.ResourceType = m.ValueResourceType.ValueString()
			}
			p.Value.ResourceReference = ref
		}
		params[i] = p
	}
	return params
}

