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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// Ensure smbSharePolicyRuleResource satisfies the resource interfaces.
var _ resource.Resource = &smbSharePolicyRuleResource{}
var _ resource.ResourceWithConfigure = &smbSharePolicyRuleResource{}
var _ resource.ResourceWithImportState = &smbSharePolicyRuleResource{}

// smbSharePolicyRuleResource implements the flashblade_smb_share_policy_rule resource.
type smbSharePolicyRuleResource struct {
	client *client.FlashBladeClient
}

// NewSmbSharePolicyRuleResource is the factory function registered in the provider.
func NewSmbSharePolicyRuleResource() resource.Resource {
	return &smbSharePolicyRuleResource{}
}

// ---------- model structs ----------------------------------------------------

// smbSharePolicyRuleModel is the top-level model for the flashblade_smb_share_policy_rule resource.
type smbSharePolicyRuleModel struct {
	ID          types.String   `tfsdk:"id"`
	PolicyName  types.String   `tfsdk:"policy_name"`
	Name        types.String   `tfsdk:"name"`
	Principal   types.String   `tfsdk:"principal"`
	Change      types.String   `tfsdk:"change"`
	FullControl types.String   `tfsdk:"full_control"`
	Read        types.String   `tfsdk:"read"`
	Timeouts    timeouts.Value `tfsdk:"timeouts"`
}

// ---------- resource interface methods --------------------------------------

// Metadata sets the Terraform type name.
func (r *smbSharePolicyRuleResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_smb_share_policy_rule"
}

// Schema defines the resource schema.
func (r *smbSharePolicyRuleResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a rule within a FlashBlade SMB share policy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the SMB share policy rule (server-assigned UUID).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"policy_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the SMB share policy this rule belongs to. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "The server-assigned rule name (stable identifier). Used for import and PATCH/DELETE API calls.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"principal": schema.StringAttribute{
				Required:    true,
				Description: "The user or group principal this rule applies to (e.g. 'Everyone', 'DOMAIN\\user').",
			},
			"change": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Permission to change files/directories: 'allow' or 'deny'.",
			},
			"full_control": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Full control permission: 'allow' or 'deny'.",
			},
			"read": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Read permission: 'allow' or 'deny'.",
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
func (r *smbSharePolicyRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new rule in the SMB share policy.
func (r *smbSharePolicyRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data smbSharePolicyRuleModel
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

	post := client.SmbSharePolicyRulePost{}
	if !data.Principal.IsNull() && !data.Principal.IsUnknown() {
		post.Principal = data.Principal.ValueString()
	}
	if !data.Change.IsNull() && !data.Change.IsUnknown() {
		post.Change = data.Change.ValueString()
	}
	if !data.FullControl.IsNull() && !data.FullControl.IsUnknown() {
		post.FullControl = data.FullControl.ValueString()
	}
	if !data.Read.IsNull() && !data.Read.IsUnknown() {
		post.Read = data.Read.ValueString()
	}

	created, err := r.client.PostSmbSharePolicyRule(ctx, policyName, post)
	if err != nil {
		resp.Diagnostics.AddError("Error creating SMB share policy rule", err.Error())
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
func (r *smbSharePolicyRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data smbSharePolicyRuleModel
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

	rule, err := r.client.GetSmbSharePolicyRuleByName(ctx, policyName, ruleName)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading SMB share policy rule", err.Error())
		return
	}

	// Drift detection on principal field.
	if !data.Principal.IsNull() && !data.Principal.IsUnknown() {
		if data.Principal.ValueString() != rule.Principal {
			tflog.Info(ctx, "drift detected on SMB share policy rule", map[string]any{
				"policy":      policyName,
				"rule":        ruleName,
				"field":       "principal",
				"state_value": data.Principal.ValueString(),
				"api_value":   rule.Principal,
			})
		}
	}

	mapSMBRuleToModel(rule, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update applies changes to an existing SMB share policy rule.
func (r *smbSharePolicyRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state smbSharePolicyRuleModel
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

	patch := client.SmbSharePolicyRulePatch{}

	if !plan.Principal.Equal(state.Principal) {
		v := plan.Principal.ValueString()
		patch.Principal = &v
	}
	if !plan.Change.Equal(state.Change) {
		v := plan.Change.ValueString()
		patch.Change = &v
	}
	if !plan.FullControl.Equal(state.FullControl) {
		v := plan.FullControl.ValueString()
		patch.FullControl = &v
	}
	if !plan.Read.Equal(state.Read) {
		v := plan.Read.ValueString()
		patch.Read = &v
	}

	_, err := r.client.PatchSmbSharePolicyRule(ctx, policyName, ruleName, patch)
	if err != nil {
		resp.Diagnostics.AddError("Error updating SMB share policy rule", err.Error())
		return
	}

	readDiags := r.readIntoState(ctx, policyName, ruleName, &plan)
	resp.Diagnostics.Append(readDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes an SMB share policy rule.
func (r *smbSharePolicyRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data smbSharePolicyRuleModel
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

	if err := r.client.DeleteSmbSharePolicyRule(ctx, policyName, ruleName); err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting SMB share policy rule", err.Error())
		return
	}
}

// ImportState imports an existing SMB share policy rule using composite ID "policy_name/rule_name".
// Note: rule_name is the server-assigned string name (not a numeric index like NFS).
func (r *smbSharePolicyRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := parseCompositeID(req.ID, 2)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid import ID format",
			fmt.Sprintf("Expected 'policy_name/rule_name', got: %q. Example: 'my-policy/smb-rule-abc12345'. %s", req.ID, err),
		)
		return
	}

	policyName := parts[0]
	ruleName := parts[1]

	rule, err := r.client.GetSmbSharePolicyRuleByName(ctx, policyName, ruleName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error finding SMB share policy rule by name",
			fmt.Sprintf("Could not find rule %q in policy %q: %s", ruleName, policyName, err),
		)
		return
	}

	var data smbSharePolicyRuleModel
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

// readIntoState calls GetSmbSharePolicyRuleByName and maps the result into the provided model.
// Returns diagnostics from the map operation.
func (r *smbSharePolicyRuleResource) readIntoState(ctx context.Context, policyName, ruleName string, data *smbSharePolicyRuleModel) diag.Diagnostics {
	var diags diag.Diagnostics
	rule, err := r.client.GetSmbSharePolicyRuleByName(ctx, policyName, ruleName)
	if err != nil {
		diags.AddError("Error reading SMB share policy rule after write", err.Error())
		return diags
	}
	mapSMBRuleToModel(rule, data)
	return diags
}

// mapSMBRuleToModel maps a client.SmbSharePolicyRule to an smbSharePolicyRuleModel.
// Empty strings from the API (JSON null → Go "") are mapped to types.StringNull()
// to avoid inconsistent state when the API does not store a permission value.
func mapSMBRuleToModel(rule *client.SmbSharePolicyRule, data *smbSharePolicyRuleModel) {
	data.ID = types.StringValue(rule.ID)
	data.Name = types.StringValue(rule.Name)
	data.PolicyName = types.StringValue(rule.Policy.Name)
	data.Principal = types.StringValue(rule.Principal)
	data.Change = stringOrNull(rule.Change)
	data.FullControl = stringOrNull(rule.FullControl)
	data.Read = stringOrNull(rule.Read)
}
