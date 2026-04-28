package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

var _ resource.Resource = &objectStoreUserPolicyResource{}
var _ resource.ResourceWithConfigure = &objectStoreUserPolicyResource{}
var _ resource.ResourceWithImportState = &objectStoreUserPolicyResource{}
var _ resource.ResourceWithUpgradeState = &objectStoreUserPolicyResource{}

// objectStoreUserPolicyResource implements the flashblade_object_store_user_policy resource.
type objectStoreUserPolicyResource struct {
	client *client.FlashBladeClient
}

func NewObjectStoreUserPolicyResource() resource.Resource {
	return &objectStoreUserPolicyResource{}
}

// ---------- model structs ----------------------------------------------------

// objectStoreUserPolicyModel is the Terraform state model for the flashblade_object_store_user_policy resource.
type objectStoreUserPolicyModel struct {
	UserName   types.String   `tfsdk:"user_name"`
	PolicyName types.String   `tfsdk:"policy_name"`
	Timeouts   timeouts.Value `tfsdk:"timeouts"`
}

// ---------- resource interface methods --------------------------------------

func (r *objectStoreUserPolicyResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_object_store_user_policy"
}

// Schema defines the resource schema.
func (r *objectStoreUserPolicyResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     0,
		Description: "Associates an object store user with an object store access policy on FlashBlade. Each association is a separate resource (CRD — no Update).",
		Attributes: map[string]schema.Attribute{
			"user_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the object store user (format: account/username). Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"policy_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the object store access policy. Changing this forces a new resource.",
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
func (r *objectStoreUserPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create associates an object store user with an access policy.
func (r *objectStoreUserPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data objectStoreUserPolicyModel
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

	member, err := r.client.PostObjectStoreUserPolicy(ctx, data.UserName.ValueString(), data.PolicyName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error creating object store user policy association", err.Error())
		return
	}

	mapObjectStoreUserPolicyToModel(member, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes Terraform state from the API by listing user policies and finding the match.
func (r *objectStoreUserPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data objectStoreUserPolicyModel
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

	members, err := r.client.ListObjectStoreUserPolicies(ctx, data.UserName.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading object store user policies", err.Error())
		return
	}

	var found *client.ObjectStoreUserPolicyMember
	for i := range members {
		if members[i].Member.Name == data.UserName.ValueString() && members[i].Policy.Name == data.PolicyName.ValueString() {
			found = &members[i]
			break
		}
	}

	if found == nil {
		tflog.Warn(ctx, "object_store_user_policy association removed outside Terraform", map[string]any{
			"user_name":   data.UserName.ValueString(),
			"policy_name": data.PolicyName.ValueString(),
		})
		resp.State.RemoveResource(ctx)
		return
	}

	mapObjectStoreUserPolicyToModel(found, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update is not supported for object store user policy associations (CRD-only resource).
func (r *objectStoreUserPolicyResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"Object store user policy associations cannot be updated. Changing user_name or policy_name forces a new resource.",
	)
}

func (r *objectStoreUserPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data objectStoreUserPolicyModel
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

	err := r.client.DeleteObjectStoreUserPolicy(ctx, data.UserName.ValueString(), data.PolicyName.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting object store user policy association", err.Error())
		return
	}
}

// ImportState imports an existing user-policy association.
// Import ID format: "account/username/policyname" (3 slash-separated parts).
func (r *objectStoreUserPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Split into at most 3 parts: account, username, policyname.
	// The first two parts together form the user name (account/username).
	parts := strings.SplitN(req.ID, "/", 3)
	if len(parts) != 3 {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Expected format: account/username/policyname, got %q", req.ID),
		)
		return
	}

	userName := parts[0] + "/" + parts[1]
	policyName := parts[2]

	members, err := r.client.ListObjectStoreUserPolicies(ctx, userName)
	if err != nil {
		resp.Diagnostics.AddError("Error importing object store user policy association", err.Error())
		return
	}

	var found *client.ObjectStoreUserPolicyMember
	for i := range members {
		if members[i].Member.Name == userName && members[i].Policy.Name == policyName {
			found = &members[i]
			break
		}
	}

	if found == nil {
		resp.Diagnostics.AddError(
			"Object store user policy association not found",
			fmt.Sprintf("No association found for user %q and policy %q.", userName, policyName),
		)
		return
	}

	var data objectStoreUserPolicyModel
	data.Timeouts = timeouts.Value{
		Object: types.ObjectNull(map[string]attr.Type{
			"create": types.StringType,
			"read":   types.StringType,
			"delete": types.StringType,
		}),
	}

	mapObjectStoreUserPolicyToModel(found, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// UpgradeState returns state upgraders for schema migrations.
func (r *objectStoreUserPolicyResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}

// ---------- helpers ---------------------------------------------------------

// mapObjectStoreUserPolicyToModel maps a client.ObjectStoreUserPolicyMember to the Terraform model.
func mapObjectStoreUserPolicyToModel(m *client.ObjectStoreUserPolicyMember, data *objectStoreUserPolicyModel) {
	data.UserName = types.StringValue(m.Member.Name)
	data.PolicyName = types.StringValue(m.Policy.Name)
}
