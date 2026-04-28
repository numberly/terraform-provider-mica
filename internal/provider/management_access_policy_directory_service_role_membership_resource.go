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

	"github.com/numberly/terraform-provider-mica/internal/client"
)

// Interface assertions — all 4 mandatory per CONVENTIONS.md §"Resource Implementation" and D-00-b.
var _ resource.Resource = &mapDsrMembershipResource{}
var _ resource.ResourceWithConfigure = &mapDsrMembershipResource{}
var _ resource.ResourceWithImportState = &mapDsrMembershipResource{}
var _ resource.ResourceWithUpgradeState = &mapDsrMembershipResource{}

type mapDsrMembershipResource struct {
	client *client.FlashBladeClient
}

func NewManagementAccessPolicyDirectoryServiceRoleMembershipResource() resource.Resource {
	return &mapDsrMembershipResource{}
}

type mapDsrMembershipModel struct {
	ID       types.String   `tfsdk:"id"`
	Policy   types.String   `tfsdk:"policy"`
	Role     types.String   `tfsdk:"role"`
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *mapDsrMembershipResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_management_access_policy_directory_service_role_membership"
}

func (r *mapDsrMembershipResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version: 0,
		Description: "Associates a management access policy with a directory service role. " +
			"Enables attaching additional policies to a role post-creation without replacing the role. " +
			"Composite import ID format: role_name/policy_name (role FIRST — required because built-in " +
			"policy names contain colons and slashes, e.g. pure:policy/array_admin).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Composite ID: role_name/policy_name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"policy": schema.StringAttribute{
				Required:    true,
				Description: "Name of the management access policy to associate. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role": schema.StringAttribute{
				Required:    true,
				Description: "Name of the directory service role. Changing this forces a new resource.",
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

// UpgradeState returns empty map — satisfies the 4th interface assertion at SchemaVersion 0.
func (r *mapDsrMembershipResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}

func (r *mapDsrMembershipResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create associates the policy with the directory service role.
func (r *mapDsrMembershipResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data mapDsrMembershipModel
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

	m, err := r.client.PostManagementAccessPolicyDirectoryServiceRoleMembership(ctx, data.Policy.ValueString(), data.Role.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error creating management access policy / directory service role membership", err.Error())
		return
	}

	data.Policy = types.StringValue(m.Policy.Name)
	data.Role = types.StringValue(m.Role.Name)
	// D-05: composite ID = role_name/policy_name (role FIRST so SplitN("/", 2) correctly
	// splits built-in policy names like "pure:policy/array_admin" as the second part).
	data.ID = types.StringValue(compositeID(m.Role.Name, m.Policy.Name))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read verifies the association still exists. On empty response (removed outside Terraform) it
// calls resp.State.RemoveResource(ctx) per D-08.
func (r *mapDsrMembershipResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data mapDsrMembershipModel
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

	m, err := r.client.GetManagementAccessPolicyDirectoryServiceRoleMembership(ctx, data.Policy.ValueString(), data.Role.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading management access policy / directory service role membership", err.Error())
		return
	}

	data.Policy = types.StringValue(m.Policy.Name)
	data.Role = types.StringValue(m.Role.Name)
	data.ID = types.StringValue(compositeID(m.Role.Name, m.Policy.Name))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update is not supported — both fields have RequiresReplace(), so this is never called.
func (r *mapDsrMembershipResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"Changing policy or role forces a new resource — update should never be called for this resource.",
	)
}

func (r *mapDsrMembershipResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data mapDsrMembershipModel
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

	err := r.client.DeleteManagementAccessPolicyDirectoryServiceRoleMembership(ctx, data.Policy.ValueString(), data.Role.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting management access policy / directory service role membership", err.Error())
		return
	}
}

// ImportState imports an existing association by composite ID "role_name/policy_name" (role FIRST — D-05).
// The slash separator with role first ensures built-in policy names like "pure:policy/array_admin"
// are parsed correctly by strings.SplitN(id, "/", 2).
func (r *mapDsrMembershipResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := parseCompositeID(req.ID, 2)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Expected format: role_name/policy_name (e.g. array_admin/pure:policy/array_admin). Error: %s", err),
		)
		return
	}
	roleName := parts[0]   // D-05: role FIRST
	policyName := parts[1] // policy may contain ":" and "/" (e.g. pure:policy/array_admin)

	m, err := r.client.GetManagementAccessPolicyDirectoryServiceRoleMembership(ctx, policyName, roleName)
	if err != nil {
		resp.Diagnostics.AddError("Error importing management access policy / directory service role membership", err.Error())
		return
	}

	data := mapDsrMembershipModel{
		ID:       types.StringValue(compositeID(m.Role.Name, m.Policy.Name)),
		Policy:   types.StringValue(m.Policy.Name),
		Role:     types.StringValue(m.Role.Name),
		Timeouts: nullTimeoutsValueNoUpdate(),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
