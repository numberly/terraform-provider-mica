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

var _ resource.Resource = &qosPolicyMemberResource{}
var _ resource.ResourceWithConfigure = &qosPolicyMemberResource{}
var _ resource.ResourceWithImportState = &qosPolicyMemberResource{}

// qosPolicyMemberResource implements the flashblade_qos_policy_member resource.
type qosPolicyMemberResource struct {
	client *client.FlashBladeClient
}

func NewQosPolicyMemberResource() resource.Resource {
	return &qosPolicyMemberResource{}
}

// ---------- model structs ----------------------------------------------------

// qosPolicyMemberModel is the Terraform state model for the flashblade_qos_policy_member resource.
type qosPolicyMemberModel struct {
	PolicyName types.String   `tfsdk:"policy_name"`
	MemberName types.String   `tfsdk:"member_name"`
	MemberType types.String   `tfsdk:"member_type"`
	Timeouts   timeouts.Value `tfsdk:"timeouts"`
}

// ---------- resource interface methods --------------------------------------

func (r *qosPolicyMemberResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_qos_policy_member"
}

// Schema defines the resource schema.
func (r *qosPolicyMemberResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     0,
		Description: "Assigns a file system or realm as a member of a FlashBlade QoS policy. Note: bucket assignment is not supported on FlashBlade API v2.22 — only file-systems and realms are valid member types.",
		Attributes: map[string]schema.Attribute{
			"policy_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the QoS policy. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"member_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the file system or realm to assign. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"member_type": schema.StringAttribute{
				Required:    true,
				Description: "The type of the member. Valid values: file-systems, realms. Note: buckets are not supported on API v2.22.",
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
func (r *qosPolicyMemberResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create adds a member to a QoS policy.
func (r *qosPolicyMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data qosPolicyMemberModel
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

	member, err := r.client.PostQosPolicyMember(ctx, data.PolicyName.ValueString(), data.MemberName.ValueString(), data.MemberType.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error creating QoS policy member", err.Error())
		return
	}

	mapQosPolicyMemberToModel(member, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes Terraform state from the API by listing members and finding the match.
func (r *qosPolicyMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data qosPolicyMemberModel
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

	members, err := r.client.ListQosPolicyMembers(ctx, data.PolicyName.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading QoS policy members", err.Error())
		return
	}

	var found *client.QosPolicyMember
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

	mapQosPolicyMemberToModel(found, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update is not supported for QoS policy members (CRD-only resource).
func (r *qosPolicyMemberResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"QoS policy members cannot be updated. Changing policy_name or member_name forces a new resource.",
	)
}

// Delete removes a member from a QoS policy.
func (r *qosPolicyMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data qosPolicyMemberModel
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

	err := r.client.DeleteQosPolicyMember(ctx, data.PolicyName.ValueString(), data.MemberName.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting QoS policy member", err.Error())
		return
	}
}

// ImportState imports an existing QoS policy member by composite ID "policyName/memberName".
func (r *qosPolicyMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := parseCompositeID(req.ID, 2)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("Expected format: policyName/memberName. Error: %s", err))
		return
	}

	policyName := parts[0]
	memberName := parts[1]

	members, err := r.client.ListQosPolicyMembers(ctx, policyName)
	if err != nil {
		resp.Diagnostics.AddError("Error importing QoS policy member", err.Error())
		return
	}

	var found *client.QosPolicyMember
	for i := range members {
		if members[i].Member.Name == memberName {
			found = &members[i]
			break
		}
	}

	if found == nil {
		resp.Diagnostics.AddError(
			"QoS policy member not found",
			fmt.Sprintf("No member %q found in QoS policy %q.", memberName, policyName),
		)
		return
	}

	var data qosPolicyMemberModel
	data.Timeouts = timeouts.Value{
		Object: types.ObjectNull(map[string]attr.Type{
			"create": types.StringType,
			"read":   types.StringType,
			"delete": types.StringType,
		}),
	}

	mapQosPolicyMemberToModel(found, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ---------- helpers ---------------------------------------------------------

// mapQosPolicyMemberToModel maps a client.QosPolicyMember to the Terraform model.
func mapQosPolicyMemberToModel(member *client.QosPolicyMember, data *qosPolicyMemberModel) {
	data.PolicyName = types.StringValue(member.Policy.Name)
	data.MemberName = types.StringValue(member.Member.Name)
	// MemberType is not returned by the mock but would be set from the API response.
	// Keep existing value if already set, otherwise leave as null.
	if data.MemberType.IsNull() || data.MemberType.IsUnknown() {
		data.MemberType = types.StringNull()
	}
}
