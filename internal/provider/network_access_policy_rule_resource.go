package provider

import (
	"context"
	"fmt"
	"strconv"
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

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// Ensure networkAccessPolicyRuleResource satisfies the resource interfaces.
var _ resource.Resource = &networkAccessPolicyRuleResource{}
var _ resource.ResourceWithConfigure = &networkAccessPolicyRuleResource{}
var _ resource.ResourceWithImportState = &networkAccessPolicyRuleResource{}

// networkAccessPolicyRuleResource implements the flashblade_network_access_policy_rule resource.
type networkAccessPolicyRuleResource struct {
	client *client.FlashBladeClient
}

// NewNetworkAccessPolicyRuleResource is the factory function registered in the provider.
func NewNetworkAccessPolicyRuleResource() resource.Resource {
	return &networkAccessPolicyRuleResource{}
}

// ---------- model structs ----------------------------------------------------

// networkAccessPolicyRuleModel is the top-level model for the flashblade_network_access_policy_rule resource.
type networkAccessPolicyRuleModel struct {
	ID            types.String   `tfsdk:"id"`
	PolicyName    types.String   `tfsdk:"policy_name"`
	Name          types.String   `tfsdk:"name"`
	Index         types.Int64    `tfsdk:"index"`
	Client        types.String   `tfsdk:"client"`
	Effect        types.String   `tfsdk:"effect"`
	Interfaces    types.List     `tfsdk:"interfaces"`
	PolicyVersion types.String   `tfsdk:"policy_version"`
	Timeouts      timeouts.Value `tfsdk:"timeouts"`
}

// ---------- resource interface methods --------------------------------------

// Metadata sets the Terraform type name.
func (r *networkAccessPolicyRuleResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_network_access_policy_rule"
}

// Schema defines the resource schema.
func (r *networkAccessPolicyRuleResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a rule within a FlashBlade network access policy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the network access policy rule (server-assigned UUID).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"policy_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the network access policy this rule belongs to. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "The server-assigned rule identifier. Used internally for PATCH/DELETE API calls.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"index": schema.Int64Attribute{
				Computed:    true,
				Description: "The server-assigned ordering index for this rule within the policy. Used for import.",
			},
			"client": schema.StringAttribute{
				Required:    true,
				Description: "IP address, CIDR range, or '*' matching the clients to which this rule applies.",
			},
			"effect": schema.StringAttribute{
				Required:    true,
				Description: "The effect of the rule: 'allow' or 'deny'.",
			},
			"interfaces": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of protocol interfaces this rule applies to (e.g. ['nfs', 'smb', 's3']). If empty, applies to all.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"policy_version": schema.StringAttribute{
				Computed:    true,
				Description: "The version of the parent policy at the time this rule was last read.",
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
func (r *networkAccessPolicyRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new rule in the network access policy.
func (r *networkAccessPolicyRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data networkAccessPolicyRuleModel
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

	post := client.NetworkAccessPolicyRulePost{
		Client: data.Client.ValueString(),
		Effect: data.Effect.ValueString(),
	}

	if !data.Interfaces.IsNull() && !data.Interfaces.IsUnknown() {
		var interfaces []string
		resp.Diagnostics.Append(data.Interfaces.ElementsAs(ctx, &interfaces, false)...)
		if !resp.Diagnostics.HasError() {
			post.Interfaces = interfaces
		}
	}
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.PostNetworkAccessPolicyRule(ctx, policyName, post)
	if err != nil {
		resp.Diagnostics.AddError("Error creating network access policy rule", err.Error())
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
func (r *networkAccessPolicyRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data networkAccessPolicyRuleModel
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

	rule, err := r.client.GetNetworkAccessPolicyRuleByName(ctx, policyName, ruleName)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading network access policy rule", err.Error())
		return
	}

	// Drift detection on mutable fields.
	if !data.Client.IsNull() && !data.Client.IsUnknown() {
		if data.Client.ValueString() != rule.Client {
			tflog.Info(ctx, "drift detected on network access policy rule", map[string]any{
				"policy":      policyName,
				"rule":        ruleName,
				"field":       "client",
				"state_value": data.Client.ValueString(),
				"api_value":   rule.Client,
			})
		}
	}

	mapDiags := mapNAPRuleToModel(ctx, rule, &data)
	resp.Diagnostics.Append(mapDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update applies changes to an existing network access policy rule via PATCH.
func (r *networkAccessPolicyRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state networkAccessPolicyRuleModel
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

	patch := client.NetworkAccessPolicyRulePatch{}

	if !plan.Client.Equal(state.Client) {
		v := plan.Client.ValueString()
		patch.Client = &v
	}
	if !plan.Effect.Equal(state.Effect) {
		v := plan.Effect.ValueString()
		patch.Effect = &v
	}
	if !plan.Interfaces.Equal(state.Interfaces) {
		var interfaces []string
		resp.Diagnostics.Append(plan.Interfaces.ElementsAs(ctx, &interfaces, false)...)
		if !resp.Diagnostics.HasError() {
			patch.Interfaces = interfaces
		}
	}
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.PatchNetworkAccessPolicyRule(ctx, policyName, ruleName, patch)
	if err != nil {
		resp.Diagnostics.AddError("Error updating network access policy rule", err.Error())
		return
	}

	readDiags := r.readIntoState(ctx, policyName, ruleName, &plan)
	resp.Diagnostics.Append(readDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes a network access policy rule.
func (r *networkAccessPolicyRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data networkAccessPolicyRuleModel
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

	if err := r.client.DeleteNetworkAccessPolicyRule(ctx, policyName, ruleName); err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting network access policy rule", err.Error())
		return
	}
}

// ImportState imports an existing network access policy rule using composite ID "policy_name/rule_index".
func (r *networkAccessPolicyRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := parseCompositeID(req.ID, 2)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid import ID format",
			fmt.Sprintf("Expected 'policy_name/rule_index', got: %q. Example: 'default/0'. %s", req.ID, err),
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

	rule, err := r.client.GetNetworkAccessPolicyRuleByIndex(ctx, policyName, index)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error finding network access policy rule by index",
			fmt.Sprintf("Could not find rule at index %d in policy %q: %s", index, policyName, err),
		)
		return
	}

	var data networkAccessPolicyRuleModel
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

// readIntoState calls GetNetworkAccessPolicyRuleByName and maps the result into the provided model.
// Returns diagnostics from the map operation.
func (r *networkAccessPolicyRuleResource) readIntoState(ctx context.Context, policyName, ruleName string, data *networkAccessPolicyRuleModel) diag.Diagnostics {
	var diags diag.Diagnostics
	rule, err := r.client.GetNetworkAccessPolicyRuleByName(ctx, policyName, ruleName)
	if err != nil {
		diags.AddError("Error reading network access policy rule after write", err.Error())
		return diags
	}
	diags.Append(mapNAPRuleToModel(ctx, rule, data)...)
	return diags
}

// mapNAPRuleToModel maps a client.NetworkAccessPolicyRule to a networkAccessPolicyRuleModel.
func mapNAPRuleToModel(ctx context.Context, rule *client.NetworkAccessPolicyRule, data *networkAccessPolicyRuleModel) diag.Diagnostics {
	var diags diag.Diagnostics

	data.ID = types.StringValue(rule.ID)
	data.Name = types.StringValue(rule.Name)
	data.Index = types.Int64Value(int64(rule.Index))
	data.Client = types.StringValue(rule.Client)
	data.Effect = types.StringValue(rule.Effect)
	data.PolicyVersion = types.StringValue(rule.PolicyVersion)

	if rule.Policy != nil {
		data.PolicyName = types.StringValue(rule.Policy.Name)
	}

	if len(rule.Interfaces) > 0 {
		interfacesList, ifDiags := types.ListValueFrom(ctx, types.StringType, rule.Interfaces)
		diags.Append(ifDiags...)
		if !diags.HasError() {
			data.Interfaces = interfacesList
		}
	} else {
		data.Interfaces = types.ListValueMust(types.StringType, []attr.Value{})
	}

	return diags
}
