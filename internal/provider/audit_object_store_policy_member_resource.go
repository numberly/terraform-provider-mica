package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// Ensure auditObjectStorePolicyMemberResource satisfies the resource interfaces.
var _ resource.Resource = &auditObjectStorePolicyMemberResource{}
var _ resource.ResourceWithConfigure = &auditObjectStorePolicyMemberResource{}
var _ resource.ResourceWithImportState = &auditObjectStorePolicyMemberResource{}
var _ resource.ResourceWithUpgradeState = &auditObjectStorePolicyMemberResource{}

// auditObjectStorePolicyMemberResource implements the flashblade_audit_object_store_policy_member resource.
type auditObjectStorePolicyMemberResource struct {
	client *client.FlashBladeClient
}

// NewAuditObjectStorePolicyMemberResource is the factory function registered in the provider.
func NewAuditObjectStorePolicyMemberResource() resource.Resource {
	return &auditObjectStorePolicyMemberResource{}
}

// ---------- model structs ----------------------------------------------------

// auditObjectStorePolicyMemberModel is the Terraform state model.
type auditObjectStorePolicyMemberModel struct {
	PolicyName types.String   `tfsdk:"policy_name"`
	MemberName types.String   `tfsdk:"member_name"`
	Timeouts   timeouts.Value `tfsdk:"timeouts"`
}

// ---------- resource interface methods --------------------------------------

// Metadata sets the Terraform type name.
func (r *auditObjectStorePolicyMemberResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_audit_object_store_policy_member"
}

// Schema defines the resource schema.
func (r *auditObjectStorePolicyMemberResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     0,
		Description: "Assigns a bucket as a member of a FlashBlade audit object store policy.",
		Attributes: map[string]schema.Attribute{
			"policy_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the audit object store policy. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"member_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the bucket to assign to the policy. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Delete: true,
			}),
		},
	}
}

// Configure injects the FlashBladeClient into the resource.
func (r *auditObjectStorePolicyMemberResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ---------- CRD methods (no Update) ------------------------------------------

// Create adds a bucket member to an audit object store policy.
func (r *auditObjectStorePolicyMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data auditObjectStorePolicyMemberModel
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

	member, err := r.client.PostAuditObjectStorePolicyMember(ctx, data.PolicyName.ValueString(), data.MemberName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error creating audit object store policy member", err.Error())
		return
	}

	data.PolicyName = types.StringValue(member.Policy.Name)
	data.MemberName = types.StringValue(member.Member.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes Terraform state by listing members and finding the match.
func (r *auditObjectStorePolicyMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data auditObjectStorePolicyMemberModel
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

	members, err := r.client.ListAuditObjectStorePolicyMembers(ctx, data.PolicyName.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading audit object store policy members", err.Error())
		return
	}

	var found *client.AuditObjectStorePolicyMember
	for i := range members {
		if members[i].Member.Name == data.MemberName.ValueString() {
			found = &members[i]
			break
		}
	}

	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data.PolicyName = types.StringValue(found.Policy.Name)
	data.MemberName = types.StringValue(found.Member.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update is not supported for policy members (CRD-only resource).
func (r *auditObjectStorePolicyMemberResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"Audit object store policy members cannot be updated. Changing policy_name or member_name forces a new resource.",
	)
}

// Delete removes a bucket from an audit object store policy.
func (r *auditObjectStorePolicyMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data auditObjectStorePolicyMemberModel
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

	err := r.client.DeleteAuditObjectStorePolicyMember(ctx, data.PolicyName.ValueString(), data.MemberName.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting audit object store policy member", err.Error())
		return
	}
}

// ImportState imports by composite ID "policyName/memberName".
func (r *auditObjectStorePolicyMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := parseCompositeID(req.ID, 2)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("Expected format: policyName/memberName. Error: %s", err))
		return
	}

	policyName := parts[0]
	memberName := parts[1]

	members, err := r.client.ListAuditObjectStorePolicyMembers(ctx, policyName)
	if err != nil {
		resp.Diagnostics.AddError("Error importing audit object store policy member", err.Error())
		return
	}

	var found *client.AuditObjectStorePolicyMember
	for i := range members {
		if members[i].Member.Name == memberName {
			found = &members[i]
			break
		}
	}

	if found == nil {
		resp.Diagnostics.AddError(
			"Audit object store policy member not found",
			fmt.Sprintf("No member %q found in audit object store policy %q.", memberName, policyName),
		)
		return
	}

	var data auditObjectStorePolicyMemberModel
	data.Timeouts = nullTimeoutsValueCRD()
	data.PolicyName = types.StringValue(found.Policy.Name)
	data.MemberName = types.StringValue(found.Member.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// UpgradeState returns state upgraders for the audit object store policy member resource.
func (r *auditObjectStorePolicyMemberResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}
