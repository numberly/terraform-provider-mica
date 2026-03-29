package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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

// Ensure smbClientPolicyRuleResource satisfies the resource interfaces.
var _ resource.Resource = &smbClientPolicyRuleResource{}
var _ resource.ResourceWithConfigure = &smbClientPolicyRuleResource{}
var _ resource.ResourceWithImportState = &smbClientPolicyRuleResource{}
var _ resource.ResourceWithUpgradeState = &smbClientPolicyRuleResource{}

// smbClientPolicyRuleResource implements the flashblade_smb_client_policy_rule resource.
type smbClientPolicyRuleResource struct {
	client *client.FlashBladeClient
}

// NewSmbClientPolicyRuleResource is the factory function registered in the provider.
func NewSmbClientPolicyRuleResource() resource.Resource {
	return &smbClientPolicyRuleResource{}
}

// ---------- model structs ----------------------------------------------------

// smbClientPolicyRuleModel is the top-level model for the flashblade_smb_client_policy_rule resource.
type smbClientPolicyRuleModel struct {
	ID         types.String   `tfsdk:"id"`
	PolicyName types.String   `tfsdk:"policy_name"`
	Name       types.String   `tfsdk:"name"`
	Index      types.Int64    `tfsdk:"index"`
	Client     types.String   `tfsdk:"client"`
	Encryption types.String   `tfsdk:"encryption"`
	Permission types.String   `tfsdk:"permission"`
	Timeouts   timeouts.Value `tfsdk:"timeouts"`
}

// ---------- resource interface methods --------------------------------------

// Metadata sets the Terraform type name.
func (r *smbClientPolicyRuleResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_smb_client_policy_rule"
}

// Schema defines the resource schema.
func (r *smbClientPolicyRuleResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     0,
		Description: "Manages a rule within a FlashBlade SMB client policy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the SMB client policy rule (server-assigned UUID).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"policy_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the SMB client policy this rule belongs to. Changing this forces a new resource.",
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
			"index": schema.Int64Attribute{
				Computed:    true,
				Description: "The server-assigned rule index within the policy.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"client": schema.StringAttribute{
				Required:    true,
				Description: "The client match expression (e.g. '*', '10.0.0.0/8').",
			},
			"encryption": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Encryption requirement: 'optional', 'required', or 'disabled'.",
				Validators: []validator.String{
					stringvalidator.OneOf("optional", "required", "disabled"),
				},
			},
			"permission": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Permission level: 'rw' or 'ro'.",
				Validators: []validator.String{
					stringvalidator.OneOf("rw", "ro"),
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


// UpgradeState returns state upgraders for schema migrations.
func (r *smbClientPolicyRuleResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}

// Configure injects the FlashBladeClient into the resource.
func (r *smbClientPolicyRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new rule in the SMB client policy.
func (r *smbClientPolicyRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data smbClientPolicyRuleModel
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

	post := client.SmbClientPolicyRulePost{}
	if !data.Client.IsNull() && !data.Client.IsUnknown() {
		post.Client = data.Client.ValueString()
	}
	if !data.Encryption.IsNull() && !data.Encryption.IsUnknown() {
		post.Encryption = data.Encryption.ValueString()
	}
	if !data.Permission.IsNull() && !data.Permission.IsUnknown() {
		post.Permission = data.Permission.ValueString()
	}

	created, err := r.client.PostSmbClientPolicyRule(ctx, policyName, post)
	if err != nil {
		resp.Diagnostics.AddError("Error creating SMB client policy rule", err.Error())
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
func (r *smbClientPolicyRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data smbClientPolicyRuleModel
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

	rule, err := r.client.GetSmbClientPolicyRuleByName(ctx, policyName, ruleName)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading SMB client policy rule", err.Error())
		return
	}

	// Drift detection on client field.
	if !data.Client.IsNull() && !data.Client.IsUnknown() {
		if data.Client.ValueString() != rule.Client {
			tflog.Info(ctx, "drift detected on SMB client policy rule", map[string]any{
				"policy":      policyName,
				"rule":        ruleName,
				"field":       "client",
				"state_value": data.Client.ValueString(),
				"api_value":   rule.Client,
			})
		}
	}

	mapSMBClientRuleToModel(rule, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update applies changes to an existing SMB client policy rule.
func (r *smbClientPolicyRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state smbClientPolicyRuleModel
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

	patch := client.SmbClientPolicyRulePatch{}

	if !plan.Client.Equal(state.Client) {
		v := plan.Client.ValueString()
		patch.Client = &v
	}
	if !plan.Encryption.Equal(state.Encryption) {
		v := plan.Encryption.ValueString()
		patch.Encryption = &v
	}
	if !plan.Permission.Equal(state.Permission) {
		v := plan.Permission.ValueString()
		patch.Permission = &v
	}

	_, err := r.client.PatchSmbClientPolicyRule(ctx, policyName, ruleName, patch)
	if err != nil {
		resp.Diagnostics.AddError("Error updating SMB client policy rule", err.Error())
		return
	}

	readDiags := r.readIntoState(ctx, policyName, ruleName, &plan)
	resp.Diagnostics.Append(readDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes an SMB client policy rule.
func (r *smbClientPolicyRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data smbClientPolicyRuleModel
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

	if err := r.client.DeleteSmbClientPolicyRule(ctx, policyName, ruleName); err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting SMB client policy rule", err.Error())
		return
	}
}

// ImportState imports an existing SMB client policy rule using composite ID "policy_name/rule_name".
func (r *smbClientPolicyRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := parseCompositeID(req.ID, 2)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid import ID format",
			fmt.Sprintf("Expected 'policy_name/rule_name', got: %q. Example: 'my-policy/smb-client-rule-abc12345'. %s", req.ID, err),
		)
		return
	}

	policyName := parts[0]
	ruleName := parts[1]

	rule, err := r.client.GetSmbClientPolicyRuleByName(ctx, policyName, ruleName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error finding SMB client policy rule by name",
			fmt.Sprintf("Could not find rule %q in policy %q: %s", ruleName, policyName, err),
		)
		return
	}

	var data smbClientPolicyRuleModel
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

// readIntoState calls GetSmbClientPolicyRuleByName and maps the result into the provided model.
func (r *smbClientPolicyRuleResource) readIntoState(ctx context.Context, policyName, ruleName string, data *smbClientPolicyRuleModel) diag.Diagnostics {
	var diags diag.Diagnostics
	rule, err := r.client.GetSmbClientPolicyRuleByName(ctx, policyName, ruleName)
	if err != nil {
		diags.AddError("Error reading SMB client policy rule after write", err.Error())
		return diags
	}
	mapSMBClientRuleToModel(rule, data)
	return diags
}

// mapSMBClientRuleToModel maps a client.SmbClientPolicyRule to an smbClientPolicyRuleModel.
// Empty strings from the API (JSON null -> Go "") are mapped to types.StringNull()
// to avoid inconsistent state when the API does not store a field value.
func mapSMBClientRuleToModel(rule *client.SmbClientPolicyRule, data *smbClientPolicyRuleModel) {
	data.ID = types.StringValue(rule.ID)
	data.Name = types.StringValue(rule.Name)
	data.PolicyName = types.StringValue(rule.Policy.Name)
	data.Index = types.Int64Value(int64(rule.Index))
	data.Client = types.StringValue(rule.Client)
	data.Encryption = stringOrNull(rule.Encryption)
	data.Permission = stringOrNull(rule.Permission)
}
