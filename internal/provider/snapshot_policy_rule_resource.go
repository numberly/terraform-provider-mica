package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/soulkyu/terraform-provider-flashblade/internal/client"
)

// Ensure snapshotPolicyRuleResource satisfies the resource interfaces.
var _ resource.Resource = &snapshotPolicyRuleResource{}
var _ resource.ResourceWithConfigure = &snapshotPolicyRuleResource{}
var _ resource.ResourceWithImportState = &snapshotPolicyRuleResource{}

// snapshotPolicyRuleResource implements the flashblade_snapshot_policy_rule resource.
// IMPORTANT: There is NO dedicated API endpoint for snapshot rules.
// All CRUD operations go through PATCH on the parent policy via add_rules / remove_rules.
type snapshotPolicyRuleResource struct {
	client *client.FlashBladeClient
}

// NewSnapshotPolicyRuleResource is the factory function registered in the provider.
func NewSnapshotPolicyRuleResource() resource.Resource {
	return &snapshotPolicyRuleResource{}
}

// ---------- model structs ----------------------------------------------------

// snapshotPolicyRuleModel is the top-level model for the flashblade_snapshot_policy_rule resource.
type snapshotPolicyRuleModel struct {
	// ID is synthetic: "{policy_name}/{rule_name}" — snapshot rules have no UUID from the API.
	ID         types.String   `tfsdk:"id"`
	PolicyName types.String   `tfsdk:"policy_name"`
	// Name is the server-assigned rule identifier within the policy.
	Name       types.String   `tfsdk:"name"`
	AtTime     types.Int64    `tfsdk:"at"`
	Every      types.Int64    `tfsdk:"every"`
	KeepFor    types.Int64    `tfsdk:"keep_for"`
	Suffix     types.String   `tfsdk:"suffix"`
	ClientName types.String   `tfsdk:"client_name"`
	Timeouts   timeouts.Value `tfsdk:"timeouts"`
}

// ---------- resource interface methods --------------------------------------

// Metadata sets the Terraform type name.
func (r *snapshotPolicyRuleResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_snapshot_policy_rule"
}

// Schema defines the resource schema.
func (r *snapshotPolicyRuleResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a rule within a FlashBlade snapshot policy. " +
			"Rules are managed via PATCH add_rules / remove_rules on the parent policy — there is no dedicated rules endpoint.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Synthetic ID in the form '{policy_name}/{rule_name}'.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"policy_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the snapshot policy this rule belongs to. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "The server-assigned rule identifier within the policy.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"at": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Schedule: run at this epoch millisecond offset within the day.",
			},
			"every": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Schedule: run every N milliseconds (e.g. 86400000 for daily).",
			},
			"keep_for": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Retention: keep snapshots for this many milliseconds (e.g. 604800000 for 7 days).",
			},
			"suffix": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "An optional suffix appended to snapshot names created by this rule.",
			},
			"client_name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "An optional client name pattern for this rule.",
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
func (r *snapshotPolicyRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create adds a new rule to the snapshot policy via PATCH add_rules.
func (r *snapshotPolicyRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data snapshotPolicyRuleModel
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
	rulePost := buildRulePost(&data)

	updatedPolicy, err := r.client.AddSnapshotPolicyRule(ctx, policyName, rulePost)
	if err != nil {
		resp.Diagnostics.AddError("Error creating snapshot policy rule", err.Error())
		return
	}

	// The new rule is the last one added in the response.
	if len(updatedPolicy.Rules) == 0 {
		resp.Diagnostics.AddError("Snapshot policy rule creation failed", "No rules found in policy after add_rules PATCH.")
		return
	}
	newRule := updatedPolicy.Rules[len(updatedPolicy.Rules)-1]

	readDiags := r.readIntoState(ctx, policyName, newRule.Name, &data)
	resp.Diagnostics.Append(readDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes Terraform state from the API by looking up the rule by its stored name.
func (r *snapshotPolicyRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data snapshotPolicyRuleModel
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

	policy, err := r.client.GetSnapshotPolicy(ctx, policyName)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading snapshot policy for rule", err.Error())
		return
	}

	rule := findRuleByName(policy, ruleName)
	if rule == nil {
		// Rule was deleted externally.
		tflog.Info(ctx, "snapshot policy rule not found, removing from state", map[string]any{
			"policy_name": policyName,
			"rule_name":   ruleName,
		})
		resp.State.RemoveResource(ctx)
		return
	}

	mapSnapshotRuleToModel(rule, policyName, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update replaces the rule atomically via PATCH remove_rules + add_rules (ReplaceSnapshotPolicyRule).
func (r *snapshotPolicyRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state snapshotPolicyRuleModel
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
	oldRuleName := state.Name.ValueString()
	newRulePost := buildRulePost(&plan)

	updatedPolicy, err := r.client.ReplaceSnapshotPolicyRule(ctx, policyName, oldRuleName, newRulePost)
	if err != nil {
		resp.Diagnostics.AddError("Error updating snapshot policy rule", err.Error())
		return
	}

	// After replace, the new rule is the last one in the array.
	if len(updatedPolicy.Rules) == 0 {
		resp.Diagnostics.AddError("Snapshot policy rule update failed", "No rules found in policy after replace PATCH.")
		return
	}
	newRule := updatedPolicy.Rules[len(updatedPolicy.Rules)-1]

	readDiags := r.readIntoState(ctx, policyName, newRule.Name, &plan)
	resp.Diagnostics.Append(readDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes the rule from the snapshot policy via PATCH remove_rules.
func (r *snapshotPolicyRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data snapshotPolicyRuleModel
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

	_, err := r.client.RemoveSnapshotPolicyRule(ctx, policyName, ruleName)
	if err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting snapshot policy rule", err.Error())
		return
	}
}

// ImportState imports an existing snapshot rule using composite ID "policy_name/rule_index".
func (r *snapshotPolicyRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid import ID format",
			fmt.Sprintf("Expected 'policy_name/rule_index', got: %q. Example: 'my-policy/0'", req.ID),
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

	rule, err := r.client.GetSnapshotPolicyRuleByIndex(ctx, policyName, index)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error finding snapshot policy rule by index",
			fmt.Sprintf("Could not find rule at index %d in policy %q: %s", index, policyName, err),
		)
		return
	}

	var data snapshotPolicyRuleModel
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
	data.Name = types.StringValue(rule.Name)

	readDiags := r.readIntoState(ctx, policyName, rule.Name, &data)
	resp.Diagnostics.Append(readDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ---------- helpers ---------------------------------------------------------

// readIntoState fetches the snapshot policy and finds the rule by name, mapping it into the model.
// Returns diagnostics from the lookup/map operation.
func (r *snapshotPolicyRuleResource) readIntoState(ctx context.Context, policyName, ruleName string, data *snapshotPolicyRuleModel) diag.Diagnostics {
	var diags diag.Diagnostics

	policy, err := r.client.GetSnapshotPolicy(ctx, policyName)
	if err != nil {
		diags.AddError("Error reading snapshot policy for rule after write", err.Error())
		return diags
	}

	rule := findRuleByName(policy, ruleName)
	if rule == nil {
		diags.AddError(
			"Snapshot policy rule not found after write",
			fmt.Sprintf("Rule %q was not found in policy %q after the operation completed.", ruleName, policyName),
		)
		return diags
	}

	mapSnapshotRuleToModel(rule, policyName, data)
	return diags
}

// buildRulePost constructs a SnapshotPolicyRulePost from the resource model.
func buildRulePost(data *snapshotPolicyRuleModel) client.SnapshotPolicyRulePost {
	post := client.SnapshotPolicyRulePost{}
	if !data.AtTime.IsNull() && !data.AtTime.IsUnknown() {
		v := data.AtTime.ValueInt64()
		post.AtTime = &v
	}
	if !data.Every.IsNull() && !data.Every.IsUnknown() {
		v := data.Every.ValueInt64()
		post.Every = &v
	}
	if !data.KeepFor.IsNull() && !data.KeepFor.IsUnknown() {
		v := data.KeepFor.ValueInt64()
		post.KeepFor = &v
	}
	if !data.Suffix.IsNull() && !data.Suffix.IsUnknown() {
		post.Suffix = data.Suffix.ValueString()
	}
	if !data.ClientName.IsNull() && !data.ClientName.IsUnknown() {
		post.ClientName = data.ClientName.ValueString()
	}
	return post
}

// findRuleByName searches the policy's rules slice for the rule with the given name.
// Returns nil if not found.
func findRuleByName(policy *client.SnapshotPolicy, ruleName string) *client.SnapshotPolicyRuleInPolicy {
	for i := range policy.Rules {
		if policy.Rules[i].Name == ruleName {
			return &policy.Rules[i]
		}
	}
	return nil
}

// mapSnapshotRuleToModel maps a client.SnapshotPolicyRuleInPolicy to a snapshotPolicyRuleModel.
func mapSnapshotRuleToModel(rule *client.SnapshotPolicyRuleInPolicy, policyName string, data *snapshotPolicyRuleModel) {
	data.PolicyName = types.StringValue(policyName)
	data.Name = types.StringValue(rule.Name)
	// Synthetic composite ID: policy_name/rule_name
	data.ID = types.StringValue(policyName + "/" + rule.Name)

	if rule.AtTime != nil {
		data.AtTime = types.Int64Value(*rule.AtTime)
	} else {
		data.AtTime = types.Int64Null()
	}
	if rule.Every != nil {
		data.Every = types.Int64Value(*rule.Every)
	} else {
		data.Every = types.Int64Null()
	}
	if rule.KeepFor != nil {
		data.KeepFor = types.Int64Value(*rule.KeepFor)
	} else {
		data.KeepFor = types.Int64Null()
	}
	data.Suffix = types.StringValue(rule.Suffix)
	data.ClientName = types.StringValue(rule.ClientName)
}
