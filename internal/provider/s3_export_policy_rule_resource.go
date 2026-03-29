package provider

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// Ensure s3ExportPolicyRuleResource satisfies the resource interfaces.
var _ resource.Resource = &s3ExportPolicyRuleResource{}
var _ resource.ResourceWithConfigure = &s3ExportPolicyRuleResource{}
var _ resource.ResourceWithImportState = &s3ExportPolicyRuleResource{}

// s3ExportPolicyRuleResource implements the flashblade_s3_export_policy_rule resource.
type s3ExportPolicyRuleResource struct {
	client *client.FlashBladeClient
}

// NewS3ExportPolicyRuleResource is the factory function registered in the provider.
func NewS3ExportPolicyRuleResource() resource.Resource {
	return &s3ExportPolicyRuleResource{}
}

// ---------- model structs ----------------------------------------------------

// s3ExportPolicyRuleModel is the top-level model for the flashblade_s3_export_policy_rule resource.
type s3ExportPolicyRuleModel struct {
	ID         types.String   `tfsdk:"id"`
	PolicyName types.String   `tfsdk:"policy_name"`
	Name       types.String   `tfsdk:"name"`
	Index      types.Int64    `tfsdk:"index"`
	Effect     types.String   `tfsdk:"effect"`
	Actions    types.List     `tfsdk:"actions"`
	Resources  types.List     `tfsdk:"resources"`
	Timeouts   timeouts.Value `tfsdk:"timeouts"`
}

// ---------- resource interface methods --------------------------------------

// Metadata sets the Terraform type name.
func (r *s3ExportPolicyRuleResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_s3_export_policy_rule"
}

// Schema defines the resource schema.
func (r *s3ExportPolicyRuleResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a rule within a FlashBlade S3 export policy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the S3 export policy rule (server-assigned UUID).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"policy_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the S3 export policy this rule belongs to. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The rule name. Passed as ?names= query param on POST.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					AlphanumericValidator(),
				},
			},
			"index": schema.Int64Attribute{
				Computed:    true,
				Description: "The server-assigned ordering index for this rule within the policy. Used for import.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"effect": schema.StringAttribute{
				Required:    true,
				Description: "The effect of the rule: 'allow' or 'deny'. Can be updated in-place.",
				Validators: []validator.String{
					stringvalidator.OneOf("allow", "deny"),
				},
			},
			"actions": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "The S3 actions this rule applies to (e.g. 's3:GetObject').",
			},
			"resources": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "The S3 resources this rule applies to (e.g. '*').",
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
func (r *s3ExportPolicyRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new rule in the S3 export policy.
func (r *s3ExportPolicyRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data s3ExportPolicyRuleModel
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

	post := client.S3ExportPolicyRulePost{
		Effect: data.Effect.ValueString(),
	}

	var actions []string
	resp.Diagnostics.Append(data.Actions.ElementsAs(ctx, &actions, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	post.Actions = actions

	var resources []string
	resp.Diagnostics.Append(data.Resources.ElementsAs(ctx, &resources, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	post.Resources = resources

	ruleName := data.Name.ValueString()
	created, err := r.client.PostS3ExportPolicyRule(ctx, policyName, ruleName, post)
	if err != nil {
		resp.Diagnostics.AddError("Error creating S3 export policy rule", err.Error())
		return
	}

	// Read back full state by server-assigned rule name.
	readDiags := r.readIntoState(ctx, policyName, created.Name, &data)
	resp.Diagnostics.Append(readDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes Terraform state from the API.
func (r *s3ExportPolicyRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data s3ExportPolicyRuleModel
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

	rule, err := r.client.GetS3ExportPolicyRuleByName(ctx, policyName, ruleName)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading S3 export policy rule", err.Error())
		return
	}

	// Drift detection on effect field.
	if !data.Effect.IsNull() && !data.Effect.IsUnknown() {
		if data.Effect.ValueString() != rule.Effect {
			tflog.Info(ctx, "drift detected on S3 export policy rule", map[string]any{
				"policy":      policyName,
				"rule":        ruleName,
				"field":       "effect",
				"state_value": data.Effect.ValueString(),
				"api_value":   rule.Effect,
			})
		}
	}

	mapDiags := mapS3RuleToModel(ctx, rule, &data)
	resp.Diagnostics.Append(mapDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update applies changes to an existing S3 export policy rule.
func (r *s3ExportPolicyRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state s3ExportPolicyRuleModel
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

	patch := client.S3ExportPolicyRulePatch{}

	if !plan.Effect.Equal(state.Effect) {
		v := plan.Effect.ValueString()
		patch.Effect = &v
	}
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

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.PatchS3ExportPolicyRule(ctx, policyName, ruleName, patch)
	if err != nil {
		resp.Diagnostics.AddError("Error updating S3 export policy rule", err.Error())
		return
	}

	readDiags := r.readIntoState(ctx, policyName, ruleName, &plan)
	resp.Diagnostics.Append(readDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes an S3 export policy rule.
func (r *s3ExportPolicyRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data s3ExportPolicyRuleModel
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

	if err := r.client.DeleteS3ExportPolicyRule(ctx, policyName, ruleName); err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting S3 export policy rule", err.Error())
		return
	}
}

// ImportState imports an existing S3 export policy rule using composite ID "policy_name/rule_index".
func (r *s3ExportPolicyRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := parseCompositeID(req.ID, 2)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid import ID format",
			fmt.Sprintf("Expected 'policy_name/rule_index', got: %q. Example: 'my-policy/0'. %s", req.ID, err),
		)
		return
	}

	policyName := parts[0]
	indexStr := parts[1]

	index, err := strconv.Atoi(indexStr)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid rule index in import ID",
			fmt.Sprintf("The rule index %q is not a valid integer: %s", indexStr, err),
		)
		return
	}

	rule, err := r.client.GetS3ExportPolicyRuleByIndex(ctx, policyName, index)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error finding S3 export policy rule by index",
			fmt.Sprintf("Could not find rule at index %d in policy %q: %s", index, policyName, err),
		)
		return
	}

	var data s3ExportPolicyRuleModel
	// Initialize timeouts with a null value so the framework can serialize it.
	data.Timeouts = nullTimeoutsValue()
	data.PolicyName = types.StringValue(policyName)
	data.Name = types.StringValue(rule.Name)

	readDiags := r.readIntoState(ctx, policyName, rule.Name, &data)
	resp.Diagnostics.Append(readDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ---------- helpers ---------------------------------------------------------

// readIntoState calls GetS3ExportPolicyRuleByName and maps the result into the provided model.
// Returns diagnostics from the map operation.
func (r *s3ExportPolicyRuleResource) readIntoState(ctx context.Context, policyName, ruleName string, data *s3ExportPolicyRuleModel) diag.Diagnostics {
	var diags diag.Diagnostics
	rule, err := r.client.GetS3ExportPolicyRuleByName(ctx, policyName, ruleName)
	if err != nil {
		diags.AddError("Error reading S3 export policy rule after write", err.Error())
		return diags
	}
	diags.Append(mapS3RuleToModel(ctx, rule, data)...)
	return diags
}

// mapS3RuleToModel maps a client.S3ExportPolicyRule to an s3ExportPolicyRuleModel.
func mapS3RuleToModel(ctx context.Context, rule *client.S3ExportPolicyRule, data *s3ExportPolicyRuleModel) diag.Diagnostics {
	var diags diag.Diagnostics

	data.ID = types.StringValue(rule.ID)
	data.Name = types.StringValue(rule.Name)
	data.Index = types.Int64Value(int64(rule.Index))
	data.PolicyName = types.StringValue(rule.Policy.Name)
	data.Effect = types.StringValue(rule.Effect)

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

	return diags
}
