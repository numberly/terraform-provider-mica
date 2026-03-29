package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// Ensure objectStoreAccessPolicyRuleResource satisfies the resource interfaces.
var _ resource.Resource = &objectStoreAccessPolicyRuleResource{}
var _ resource.ResourceWithConfigure = &objectStoreAccessPolicyRuleResource{}
var _ resource.ResourceWithImportState = &objectStoreAccessPolicyRuleResource{}
var _ resource.ResourceWithUpgradeState = &objectStoreAccessPolicyRuleResource{}

// objectStoreAccessPolicyRuleResource implements the flashblade_object_store_access_policy_rule resource.
type objectStoreAccessPolicyRuleResource struct {
	client *client.FlashBladeClient
}

// NewObjectStoreAccessPolicyRuleResource is the factory function registered in the provider.
func NewObjectStoreAccessPolicyRuleResource() resource.Resource {
	return &objectStoreAccessPolicyRuleResource{}
}

// ---------- model structs ----------------------------------------------------

// objectStoreAccessPolicyRuleModel is the top-level model for the flashblade_object_store_access_policy_rule resource.
type objectStoreAccessPolicyRuleModel struct {
	ID         types.String   `tfsdk:"id"`
	PolicyName types.String   `tfsdk:"policy_name"`
	Name       types.String   `tfsdk:"name"`
	Effect     types.String   `tfsdk:"effect"`
	Actions    types.List     `tfsdk:"actions"`
	Resources  types.List     `tfsdk:"resources"`
	Conditions types.String   `tfsdk:"conditions"`
	Timeouts   timeouts.Value `tfsdk:"timeouts"`
}

// ---------- resource interface methods --------------------------------------

// Metadata sets the Terraform type name.
func (r *objectStoreAccessPolicyRuleResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_object_store_access_policy_rule"
}

// Schema defines the resource schema.
func (r *objectStoreAccessPolicyRuleResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a rule within a FlashBlade object store access policy (IAM-style S3 policy rule).",
		Version:     0,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Synthetic identifier in the form 'policy_name/rule_name'.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"policy_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the object store access policy this rule belongs to. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the rule. Changing this forces a new resource (rules cannot be renamed).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"effect": schema.StringAttribute{
				Required:    true,
				Description: "The effect of the rule: 'allow' or 'deny'. Read-only after creation — changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("allow", "deny"),
				},
			},
			"actions": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "List of S3 actions this rule applies to (e.g. ['s3:GetObject', 's3:PutObject']).",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"resources": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "List of ARN-like resource patterns this rule applies to.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"conditions": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "JSON-encoded IAM conditions object (use jsonencode()). Null or empty if no conditions.",
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

func (r *objectStoreAccessPolicyRuleResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}

// Configure injects the FlashBladeClient into the resource.
func (r *objectStoreAccessPolicyRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new rule in the object store access policy.
func (r *objectStoreAccessPolicyRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data objectStoreAccessPolicyRuleModel
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

	policyName := data.PolicyName.ValueString()
	ruleName := data.Name.ValueString()

	post := client.ObjectStoreAccessPolicyRulePost{
		Effect: data.Effect.ValueString(),
	}

	if !data.Actions.IsNull() && !data.Actions.IsUnknown() {
		var actions []string
		resp.Diagnostics.Append(data.Actions.ElementsAs(ctx, &actions, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		post.Actions = actions
	}

	if !data.Resources.IsNull() && !data.Resources.IsUnknown() {
		var resources []string
		resp.Diagnostics.Append(data.Resources.ElementsAs(ctx, &resources, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		post.Resources = resources
	}

	if !data.Conditions.IsNull() && !data.Conditions.IsUnknown() {
		condStr := data.Conditions.ValueString()
		if condStr != "" {
			post.Conditions = json.RawMessage(condStr)
		}
	}

	_, err := r.client.PostObjectStoreAccessPolicyRule(ctx, policyName, ruleName, post)
	if err != nil {
		resp.Diagnostics.AddError("Error creating object store access policy rule", err.Error())
		return
	}

	readDiags := r.readIntoState(ctx, policyName, ruleName, &data)
	resp.Diagnostics.Append(readDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes Terraform state from the API.
func (r *objectStoreAccessPolicyRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data objectStoreAccessPolicyRuleModel
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

	policyName := data.PolicyName.ValueString()
	ruleName := data.Name.ValueString()

	rule, err := r.client.GetObjectStoreAccessPolicyRuleByName(ctx, policyName, ruleName)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading object store access policy rule", err.Error())
		return
	}

	// Drift detection on actions field.
	if !data.Actions.IsNull() && !data.Actions.IsUnknown() {
		var stateActions []string
		data.Actions.ElementsAs(ctx, &stateActions, false)
		if len(stateActions) != len(rule.Actions) {
			tflog.Info(ctx, "drift detected on object store access policy rule", map[string]any{
				"policy": policyName,
				"rule":   ruleName,
				"field":  "actions",
			})
		}
	}

	mapDiags := mapOAPRuleToModel(ctx, rule, &data)
	resp.Diagnostics.Append(mapDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update applies changes to an existing object store access policy rule.
// Only actions, resources, and conditions are patchable. Effect has RequiresReplace.
func (r *objectStoreAccessPolicyRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state objectStoreAccessPolicyRuleModel
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

	policyName := state.PolicyName.ValueString()
	ruleName := state.Name.ValueString()

	patch := client.ObjectStoreAccessPolicyRulePatch{}

	if !plan.Actions.Equal(state.Actions) {
		var actions []string
		resp.Diagnostics.Append(plan.Actions.ElementsAs(ctx, &actions, false)...)
		if !resp.Diagnostics.HasError() {
			patch.Actions = actions
		}
	}

	if !plan.Resources.Equal(state.Resources) {
		var resources []string
		resp.Diagnostics.Append(plan.Resources.ElementsAs(ctx, &resources, false)...)
		if !resp.Diagnostics.HasError() {
			patch.Resources = resources
		}
	}

	if !plan.Conditions.Equal(state.Conditions) {
		condStr := plan.Conditions.ValueString()
		if condStr != "" {
			patch.Conditions = json.RawMessage(condStr)
		} else {
			patch.Conditions = json.RawMessage("null")
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.PatchObjectStoreAccessPolicyRule(ctx, policyName, ruleName, patch)
	if err != nil {
		resp.Diagnostics.AddError("Error updating object store access policy rule", err.Error())
		return
	}

	readDiags := r.readIntoState(ctx, policyName, ruleName, &plan)
	resp.Diagnostics.Append(readDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes an object store access policy rule.
func (r *objectStoreAccessPolicyRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data objectStoreAccessPolicyRuleModel
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

	policyName := data.PolicyName.ValueString()
	ruleName := data.Name.ValueString()

	if err := r.client.DeleteObjectStoreAccessPolicyRule(ctx, policyName, ruleName); err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting object store access policy rule", err.Error())
		return
	}
}

// ImportState imports an existing object store access policy rule using composite ID "policy_name/rule_name".
func (r *objectStoreAccessPolicyRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := parseCompositeID(req.ID, 2)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid import ID format",
			fmt.Sprintf("Expected 'policy_name/rule_name', got: %q. Example: 'my-policy/my-rule'. %s", req.ID, err),
		)
		return
	}

	policyName := parts[0]
	ruleName := parts[1]

	var data objectStoreAccessPolicyRuleModel
	// Initialize timeouts with a null value so the framework can serialize it.
	data.Timeouts = timeouts.Value{
		Object: types.ObjectNull(map[string]attr.Type{
			"create": types.StringType,
			"read":   types.StringType,
			"update": types.StringType,
			"delete": types.StringType,
		}),
	}
	data.PolicyName = types.StringValue(policyName)
	data.Name = types.StringValue(ruleName)

	readDiags := r.readIntoState(ctx, policyName, ruleName, &data)
	resp.Diagnostics.Append(readDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ---------- helpers ---------------------------------------------------------

// readIntoState calls GetObjectStoreAccessPolicyRuleByName and maps the result into the provided model.
// Returns diagnostics from the map operation.
func (r *objectStoreAccessPolicyRuleResource) readIntoState(ctx context.Context, policyName, ruleName string, data *objectStoreAccessPolicyRuleModel) diag.Diagnostics {
	var diags diag.Diagnostics
	rule, err := r.client.GetObjectStoreAccessPolicyRuleByName(ctx, policyName, ruleName)
	if err != nil {
		diags.AddError("Error reading object store access policy rule after write", err.Error())
		return diags
	}
	diags.Append(mapOAPRuleToModel(ctx, rule, data)...)
	return diags
}

// mapOAPRuleToModel maps a client.ObjectStoreAccessPolicyRule to an objectStoreAccessPolicyRuleModel.
// Handles conditions json.RawMessage to types.String conversion.
func mapOAPRuleToModel(ctx context.Context, rule *client.ObjectStoreAccessPolicyRule, data *objectStoreAccessPolicyRuleModel) diag.Diagnostics {
	var diags diag.Diagnostics

	// Synthetic ID: "policyName/ruleName"
	policyName := ""
	if rule.Policy != nil {
		policyName = rule.Policy.Name
	}
	data.ID = types.StringValue(compositeID(policyName, rule.Name))
	data.Name = types.StringValue(rule.Name)
	data.Effect = types.StringValue(rule.Effect)

	if rule.Policy != nil {
		data.PolicyName = types.StringValue(rule.Policy.Name)
	}

	if len(rule.Actions) > 0 {
		actionsList, actDiags := types.ListValueFrom(ctx, types.StringType, rule.Actions)
		diags.Append(actDiags...)
		if !diags.HasError() {
			data.Actions = actionsList
		}
	} else {
		data.Actions = types.ListValueMust(types.StringType, []attr.Value{})
	}

	if len(rule.Resources) > 0 {
		resourcesList, resDiags := types.ListValueFrom(ctx, types.StringType, rule.Resources)
		diags.Append(resDiags...)
		if !diags.HasError() {
			data.Resources = resourcesList
		}
	} else {
		data.Resources = types.ListValueMust(types.StringType, []attr.Value{})
	}

	// Handle conditions: nil or empty `{}` -> null; otherwise string representation.
	if len(rule.Conditions) == 0 || string(rule.Conditions) == "null" || string(rule.Conditions) == "{}" {
		data.Conditions = types.StringNull()
	} else {
		data.Conditions = types.StringValue(string(rule.Conditions))
	}

	return diags
}

