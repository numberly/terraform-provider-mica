package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// Interface assertions — all 4 mandatory per CONVENTIONS.md §"Resource Implementation".
var _ resource.Resource = &directoryServiceRoleResource{}
var _ resource.ResourceWithConfigure = &directoryServiceRoleResource{}
var _ resource.ResourceWithImportState = &directoryServiceRoleResource{}
var _ resource.ResourceWithUpgradeState = &directoryServiceRoleResource{}

type directoryServiceRoleResource struct {
	client *client.FlashBladeClient
}

func NewDirectoryServiceRoleResource() resource.Resource { return &directoryServiceRoleResource{} }

// directoryServiceRoleModel is the tfsdk state model.
type directoryServiceRoleModel struct {
	ID                       types.String   `tfsdk:"id"`
	Name                     types.String   `tfsdk:"name"`
	Group                    types.String   `tfsdk:"group"`
	GroupBase                types.String   `tfsdk:"group_base"`
	ManagementAccessPolicies types.List     `tfsdk:"management_access_policies"`
	Role                     types.Object   `tfsdk:"role"`
	Timeouts                 timeouts.Value `tfsdk:"timeouts"`
}

// directoryServiceRoleV0Model is the v0 state model (name was Computed, not Required).
// Fields match v1 structure — only the schema attribute modifiers changed between v0 and v1.
type directoryServiceRoleV0Model struct {
	ID                       types.String   `tfsdk:"id"`
	Name                     types.String   `tfsdk:"name"`
	Group                    types.String   `tfsdk:"group"`
	GroupBase                types.String   `tfsdk:"group_base"`
	ManagementAccessPolicies types.List     `tfsdk:"management_access_policies"`
	Role                     types.Object   `tfsdk:"role"`
	Timeouts                 timeouts.Value `tfsdk:"timeouts"`
}

// roleAttrTypes is the attr.Type map for the deprecated computed role sub-object.
func roleAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{"name": types.StringType}
}

func (r *directoryServiceRoleResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_directory_service_role"
}

func (r *directoryServiceRoleResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     1,
		Description: "Maps an LDAP group to one or more FlashBlade management access policies, identified by a user-supplied name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Globally-unique role ID assigned by FlashBlade.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Unique name for the directory service role. Required on create. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"group": schema.StringAttribute{
				Required:    true,
				Description: "CN of the LDAP group whose members receive the role. Mutable via PATCH.",
				// No plan modifier — drift detection only; PATCH applies changes.
			},
			"group_base": schema.StringAttribute{
				Required:    true,
				Description: "DN search base where the LDAP group is located. Mutable via PATCH.",
			},
			"management_access_policies": schema.ListAttribute{
				Required:    true,
				Description: "List of management access policy names (e.g. pure:policy/array_admin). Writable on POST only — changing this forces a new resource.",
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{listplanmodifier.RequiresReplace()},
				Validators:    []validator.List{listvalidator.SizeAtLeast(1)},
			},
			// D-02: role is deprecated per swagger → Computed-only; SC-3 replacement trigger lives on management_access_policies.
			"role": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Deprecated legacy backfill. Populated by the API when the role maps to exactly one legacy-named policy; otherwise null.",
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{Computed: true},
				},
				// No plan modifier — API-derived, may drift.
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{Create: true, Read: true, Update: true, Delete: true}),
		},
	}
}

func (r *directoryServiceRoleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.FlashBladeClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data Type", fmt.Sprintf("Expected *client.FlashBladeClient, got: %T.", req.ProviderData))
		return
	}
	r.client = c
}

// UpgradeState migrates v0 state (Computed name) to v1 (Required name).
// The v0 PriorSchema is the EXACT broken schema — do not "fix" it here.
// Existing v0 states already carry a name (populated by the API on create),
// so the upgrader copies every field forward verbatim.
func (r *directoryServiceRoleResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		// v0 → v1: name attribute changed from Computed to Required.
		0: {
			PriorSchema: &schema.Schema{
				Version:     0,
				Description: "Maps an LDAP group to one or more FlashBlade management access policies. The role name is server-generated from the first associated policy; use the output `name` attribute downstream.",
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed:      true,
						PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
					"name": schema.StringAttribute{
						Computed:      true,
						PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
					"group": schema.StringAttribute{
						Required: true,
					},
					"group_base": schema.StringAttribute{
						Required: true,
					},
					"management_access_policies": schema.ListAttribute{
						Required:      true,
						ElementType:   types.StringType,
						PlanModifiers: []planmodifier.List{listplanmodifier.RequiresReplace()},
						Validators:    []validator.List{listvalidator.SizeAtLeast(1)},
					},
					"role": schema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"name": schema.StringAttribute{Computed: true},
						},
					},
					"timeouts": timeouts.Attributes(ctx, timeouts.Opts{Create: true, Read: true, Update: true, Delete: true}),
				},
			},
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var old directoryServiceRoleV0Model
				resp.Diagnostics.Append(req.State.Get(ctx, &old)...)
				if resp.Diagnostics.HasError() {
					return
				}
				// Carry all fields forward verbatim — name was already populated by API in v0 state.
				// v0 and v1 models share identical Go field shapes (only schema flags differ),
				// so a direct type conversion is sufficient.
				newState := directoryServiceRoleModel(old)
				resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
			},
		},
	}
}

// Create builds DirectoryServiceRolePost from the plan and posts it.
func (r *directoryServiceRoleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data directoryServiceRoleModel
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

	// Map list<string> → []NamedReference{{Name: s}, ...}
	var policyNames []string
	resp.Diagnostics.Append(data.ManagementAccessPolicies.ElementsAs(ctx, &policyNames, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	policies := make([]client.NamedReference, len(policyNames))
	for i, n := range policyNames {
		policies[i] = client.NamedReference{Name: n}
	}

	body := client.DirectoryServiceRolePost{
		Group:                    data.Group.ValueString(),
		GroupBase:                data.GroupBase.ValueString(),
		ManagementAccessPolicies: policies,
	}
	role, err := r.client.PostDirectoryServiceRole(ctx, data.Name.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Error creating directory service role", err.Error())
		return
	}

	resp.Diagnostics.Append(mapDirectoryServiceRoleToModel(ctx, role, &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes state and logs drift via tflog.Debug.
func (r *directoryServiceRoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data directoryServiceRoleModel
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

	name := data.Name.ValueString()
	role, err := r.client.GetDirectoryServiceRole(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading directory service role", err.Error())
		return
	}

	// Drift detection — before overwriting state.
	if data.Group.ValueString() != role.Group {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": name, "field": "group",
			"was": data.Group.ValueString(), "now": role.Group,
		})
	}
	if data.GroupBase.ValueString() != role.GroupBase {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": name, "field": "group_base",
			"was": data.GroupBase.ValueString(), "now": role.GroupBase,
		})
	}
	// management_access_policies drift (computed by comparing list elements).
	var oldPolicies []string
	if !data.ManagementAccessPolicies.IsNull() && !data.ManagementAccessPolicies.IsUnknown() {
		_ = data.ManagementAccessPolicies.ElementsAs(ctx, &oldPolicies, false)
	}
	newPolicies := make([]string, len(role.ManagementAccessPolicies))
	for i, p := range role.ManagementAccessPolicies {
		newPolicies[i] = p.Name
	}
	if !stringSlicesEqual(oldPolicies, newPolicies) {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": name, "field": "management_access_policies",
			"was": oldPolicies, "now": newPolicies,
		})
	}
	// role.name drift (computed nested object).
	var oldRoleName string
	if !data.Role.IsNull() && !data.Role.IsUnknown() {
		attrs := data.Role.Attributes()
		if v, ok := attrs["name"].(types.String); ok {
			oldRoleName = v.ValueString()
		}
	}
	var newRoleName string
	if role.Role != nil {
		newRoleName = role.Role.Name
	}
	if oldRoleName != newRoleName {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": name, "field": "role.name",
			"was": oldRoleName, "now": newRoleName,
		})
	}

	resp.Diagnostics.Append(mapDirectoryServiceRoleToModel(ctx, role, &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update sends PATCH with only changed *string pointers. management_access_policies is
// RequiresReplace so Update is never called for it.
func (r *directoryServiceRoleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state directoryServiceRoleModel
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

	var body client.DirectoryServiceRolePatch
	if !plan.Group.Equal(state.Group) {
		v := plan.Group.ValueString()
		body.Group = &v
	}
	if !plan.GroupBase.Equal(state.GroupBase) {
		v := plan.GroupBase.ValueString()
		body.GroupBase = &v
	}

	name := state.Name.ValueString()
	role, err := r.client.PatchDirectoryServiceRole(ctx, name, body)
	if err != nil {
		resp.Diagnostics.AddError("Error updating directory service role", err.Error())
		return
	}

	plan.Timeouts = state.Timeouts
	resp.Diagnostics.Append(mapDirectoryServiceRoleToModel(ctx, role, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *directoryServiceRoleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data directoryServiceRoleModel
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

	if err := r.client.DeleteDirectoryServiceRole(ctx, data.Name.ValueString()); err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting directory service role", err.Error())
		return
	}
}

// ImportState imports by name.
func (r *directoryServiceRoleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	role, err := r.client.GetDirectoryServiceRole(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing directory service role", err.Error())
		return
	}
	var data directoryServiceRoleModel
	data.Timeouts = nullTimeoutsValue()
	resp.Diagnostics.Append(mapDirectoryServiceRoleToModel(ctx, role, &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ---------- helpers ---------------------------------------------------------

func mapDirectoryServiceRoleToModel(ctx context.Context, role *client.DirectoryServiceRole, data *directoryServiceRoleModel) (d diag.Diagnostics) {
	data.ID = types.StringValue(role.ID)
	data.Name = types.StringValue(role.Name)
	data.Group = types.StringValue(role.Group)
	data.GroupBase = types.StringValue(role.GroupBase)

	names := make([]attr.Value, len(role.ManagementAccessPolicies))
	for i, p := range role.ManagementAccessPolicies {
		names[i] = types.StringValue(p.Name)
	}
	listVal, diags := types.ListValue(types.StringType, names)
	d = append(d, diags...)
	data.ManagementAccessPolicies = listVal

	if role.Role == nil {
		data.Role = types.ObjectNull(roleAttrTypes())
	} else {
		obj, diags := types.ObjectValue(roleAttrTypes(), map[string]attr.Value{
			"name": types.StringValue(role.Role.Name),
		})
		d = append(d, diags...)
		data.Role = obj
	}
	return
}
