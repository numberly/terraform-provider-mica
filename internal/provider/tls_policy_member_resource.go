package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

var _ resource.Resource = &tlsPolicyMemberResource{}
var _ resource.ResourceWithConfigure = &tlsPolicyMemberResource{}
var _ resource.ResourceWithImportState = &tlsPolicyMemberResource{}
var _ resource.ResourceWithUpgradeState = &tlsPolicyMemberResource{}

// tlsPolicyMemberResource implements the flashblade_tls_policy_member resource.
type tlsPolicyMemberResource struct {
	client *client.FlashBladeClient
}

func NewTlsPolicyMemberResource() resource.Resource {
	return &tlsPolicyMemberResource{}
}

// ---------- model structs ----------------------------------------------------

// tlsPolicyMemberModel is the Terraform state model for the flashblade_tls_policy_member resource.
type tlsPolicyMemberModel struct {
	PolicyName types.String   `tfsdk:"policy_name"`
	MemberName types.String   `tfsdk:"member_name"`
	Timeouts   timeouts.Value `tfsdk:"timeouts"`
}

// ---------- resource interface methods --------------------------------------

func (r *tlsPolicyMemberResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_tls_policy_member"
}

// Schema defines the resource schema.
func (r *tlsPolicyMemberResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     0,
		Description: "Assigns a network interface to a FlashBlade TLS policy. This is a CRD-only resource — all fields force a new resource on change.",
		Attributes: map[string]schema.Attribute{
			"policy_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the TLS policy. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"member_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the network interface to assign. Changing this forces a new resource.",
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
func (r *tlsPolicyMemberResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create assigns a network interface to a TLS policy.
func (r *tlsPolicyMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data tlsPolicyMemberModel
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

	member, err := r.client.PostTlsPolicyMember(ctx, data.PolicyName.ValueString(), data.MemberName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error creating TLS policy member", err.Error())
		return
	}

	data.PolicyName = types.StringValue(member.Policy.Name)
	data.MemberName = types.StringValue(member.Member.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes Terraform state from the API by listing members and finding the match.
func (r *tlsPolicyMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data tlsPolicyMemberModel
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

	members, err := r.client.ListTlsPolicyMembers(ctx, data.PolicyName.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading TLS policy members", err.Error())
		return
	}

	var found *client.TlsPolicyMember
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

// Update is not supported for TLS policy members (CRD-only resource).
func (r *tlsPolicyMemberResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"TLS policy members cannot be updated. Changing policy_name or member_name forces a new resource.",
	)
}

// Delete removes a network interface from a TLS policy.
func (r *tlsPolicyMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data tlsPolicyMemberModel
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

	err := r.client.DeleteTlsPolicyMember(ctx, data.PolicyName.ValueString(), data.MemberName.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting TLS policy member", err.Error())
		return
	}
}

// ImportState imports an existing TLS policy member by composite ID "policyName/memberName".
func (r *tlsPolicyMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := parseCompositeID(req.ID, 2)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("Expected format: policyName/memberName. Error: %s", err))
		return
	}

	policyName := parts[0]
	memberName := parts[1]

	members, err := r.client.ListTlsPolicyMembers(ctx, policyName)
	if err != nil {
		resp.Diagnostics.AddError("Error importing TLS policy member", err.Error())
		return
	}

	var found *client.TlsPolicyMember
	for i := range members {
		if members[i].Member.Name == memberName {
			found = &members[i]
			break
		}
	}

	if found == nil {
		resp.Diagnostics.AddError(
			"TLS policy member not found",
			fmt.Sprintf("No member %q found in TLS policy %q.", memberName, policyName),
		)
		return
	}

	var data tlsPolicyMemberModel
	data.Timeouts = timeouts.Value{
		Object: types.ObjectNull(map[string]attr.Type{
			"create": types.StringType,
			"read":   types.StringType,
			"delete": types.StringType,
		}),
	}

	data.PolicyName = types.StringValue(found.Policy.Name)
	data.MemberName = types.StringValue(found.Member.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// UpgradeState returns state upgraders for schema migrations.
func (r *tlsPolicyMemberResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}
