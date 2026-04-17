package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

var _ resource.Resource = &directoryServiceManagementResource{}
var _ resource.ResourceWithConfigure = &directoryServiceManagementResource{}
var _ resource.ResourceWithImportState = &directoryServiceManagementResource{}
var _ resource.ResourceWithUpgradeState = &directoryServiceManagementResource{}

// managementDirectoryServiceName is the hardcoded singleton name per D-01.
const managementDirectoryServiceName = "management"

type directoryServiceManagementResource struct {
	client *client.FlashBladeClient
}

func NewDirectoryServiceManagementResource() resource.Resource {
	return &directoryServiceManagementResource{}
}

type directoryServiceManagementModel struct {
	ID                    types.String   `tfsdk:"id"`
	Enabled               types.Bool     `tfsdk:"enabled"`
	URIs                  types.List     `tfsdk:"uris"`
	BaseDN                types.String   `tfsdk:"base_dn"`
	BindUser              types.String   `tfsdk:"bind_user"`
	BindPassword          types.String   `tfsdk:"bind_password"`
	CACertificate         types.String   `tfsdk:"ca_certificate"`
	CACertificateGroup    types.String   `tfsdk:"ca_certificate_group"`
	UserLoginAttribute    types.String   `tfsdk:"user_login_attribute"`
	UserObjectClass       types.String   `tfsdk:"user_object_class"`
	SSHPublicKeyAttribute types.String   `tfsdk:"ssh_public_key_attribute"`
	Services              types.List     `tfsdk:"services"`
	Timeouts              timeouts.Value `tfsdk:"timeouts"`
}

func (r *directoryServiceManagementResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_directory_service_management"
}

func (r *directoryServiceManagementResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     0,
		Description: "Manages the FlashBlade LDAP management directory service (admin authentication). Singleton resource — the underlying name is always \"management\". Delete resets the configuration (enabled=false, empty URIs, cleared references).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the management directory service.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "If true, the management directory service authenticates FlashBlade admin logins against LDAP.",
			},
			"uris": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of LDAP server URIs. Each entry must start with ldap:// or ldaps://.",
				Validators: []validator.List{
					LDAPURIValidator(),
				},
			},
			"base_dn": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Base Distinguished Name (DN) used when searching the directory.",
			},
			"bind_user": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Distinguished Name (DN) of the user used to bind to the directory.",
			},
			"bind_password": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Sensitive:   true,
				Description: "Password used to bind to the directory. Write-only — never returned by the API.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"ca_certificate": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Name of a CA certificate used to validate the LDAPS server certificate. Clear by omitting the attribute.",
			},
			"ca_certificate_group": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Name of a CA certificate group used to validate the LDAPS server certificate. Clear by omitting the attribute.",
			},
			"user_login_attribute": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "LDAP attribute that holds the user's login name. API default: sAMAccountName for AD, uid otherwise.",
			},
			"user_object_class": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "LDAP object class for management users. API default: User (AD), posixAccount/shadowAccount (OpenLDAP), person (other).",
			},
			"ssh_public_key_attribute": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "LDAP attribute that holds the user's SSH public key (e.g. sshPublicKey).",
			},
			"services": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Services that use this directory service configuration. Read-only. No plan modifier — drift is visible.",
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

func (r *directoryServiceManagementResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}

func (r *directoryServiceManagementResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *directoryServiceManagementResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data directoryServiceManagementModel
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

	// Build patch body from plan — treat as full create (no prior state).
	var zeroState directoryServiceManagementModel
	patch, d := buildDSMPatchFromPlan(ctx, data, zeroState)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.PatchDirectoryServiceManagement(ctx, managementDirectoryServiceName, patch)
	if err != nil {
		resp.Diagnostics.AddError("Error creating directory service management config", err.Error())
		return
	}

	// Preserve write-only sensitive value from plan before mapping (API never returns it).
	savedPassword := data.BindPassword

	resp.Diagnostics.Append(mapDirectoryServiceToModel(ctx, result, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Restore bind_password — mapDirectoryServiceToModel does not touch it,
	// but be explicit to document the intent.
	data.BindPassword = savedPassword

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *directoryServiceManagementResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data directoryServiceManagementModel
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

	ds, err := r.client.GetDirectoryServiceManagement(ctx, managementDirectoryServiceName)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading directory service management config", err.Error())
		return
	}

	// Drift detection on all mutable/computed fields.
	if data.Enabled.ValueBool() != ds.Enabled {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": managementDirectoryServiceName,
			"field":    "enabled",
			"was":      data.Enabled.ValueBool(),
			"now":      ds.Enabled,
		})
	}

	// Compute new URIs list for comparison.
	var newURIs types.List
	if len(ds.URIs) > 0 {
		newURIs, diags = types.ListValueFrom(ctx, types.StringType, ds.URIs)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		newURIs = emptyStringList()
	}
	if !data.URIs.Equal(newURIs) {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": managementDirectoryServiceName,
			"field":    "uris",
			"was":      data.URIs.String(),
			"now":      newURIs.String(),
		})
	}

	if data.BaseDN.ValueString() != ds.BaseDN {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": managementDirectoryServiceName,
			"field":    "base_dn",
			"was":      data.BaseDN.ValueString(),
			"now":      ds.BaseDN,
		})
	}

	if data.BindUser.ValueString() != ds.BindUser {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": managementDirectoryServiceName,
			"field":    "bind_user",
			"was":      data.BindUser.ValueString(),
			"now":      ds.BindUser,
		})
	}

	newCACert := ""
	if ds.CACertificate != nil {
		newCACert = ds.CACertificate.Name
	}
	if data.CACertificate.ValueString() != newCACert {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": managementDirectoryServiceName,
			"field":    "ca_certificate",
			"was":      data.CACertificate.ValueString(),
			"now":      newCACert,
		})
	}

	newCACertGroup := ""
	if ds.CACertificateGroup != nil {
		newCACertGroup = ds.CACertificateGroup.Name
	}
	if data.CACertificateGroup.ValueString() != newCACertGroup {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": managementDirectoryServiceName,
			"field":    "ca_certificate_group",
			"was":      data.CACertificateGroup.ValueString(),
			"now":      newCACertGroup,
		})
	}

	if data.UserLoginAttribute.ValueString() != ds.Management.UserLoginAttribute {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": managementDirectoryServiceName,
			"field":    "user_login_attribute",
			"was":      data.UserLoginAttribute.ValueString(),
			"now":      ds.Management.UserLoginAttribute,
		})
	}

	if data.UserObjectClass.ValueString() != ds.Management.UserObjectClass {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": managementDirectoryServiceName,
			"field":    "user_object_class",
			"was":      data.UserObjectClass.ValueString(),
			"now":      ds.Management.UserObjectClass,
		})
	}

	if data.SSHPublicKeyAttribute.ValueString() != ds.Management.SSHPublicKeyAttribute {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": managementDirectoryServiceName,
			"field":    "ssh_public_key_attribute",
			"was":      data.SSHPublicKeyAttribute.ValueString(),
			"now":      ds.Management.SSHPublicKeyAttribute,
		})
	}

	// Compute new services list for comparison.
	var newServices types.List
	if len(ds.Services) > 0 {
		newServices, diags = types.ListValueFrom(ctx, types.StringType, ds.Services)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		newServices = emptyStringList()
	}
	if !data.Services.Equal(newServices) {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": managementDirectoryServiceName,
			"field":    "services",
			"was":      data.Services.String(),
			"now":      newServices.String(),
		})
	}

	// Preserve write-only bind_password — API never returns it.
	savedPassword := data.BindPassword
	resp.Diagnostics.Append(mapDirectoryServiceToModel(ctx, ds, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.BindPassword = savedPassword

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *directoryServiceManagementResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state directoryServiceManagementModel
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

	patch := client.DirectoryServicePatch{}

	if !plan.Enabled.Equal(state.Enabled) {
		v := plan.Enabled.ValueBool()
		patch.Enabled = &v
	}

	if !plan.URIs.Equal(state.URIs) {
		uris, d := listToStrings(ctx, plan.URIs)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}
		patch.URIs = &uris
	}

	if !plan.BaseDN.Equal(state.BaseDN) {
		v := plan.BaseDN.ValueString()
		patch.BaseDN = &v
	}

	if !plan.BindUser.Equal(state.BindUser) {
		v := plan.BindUser.ValueString()
		patch.BindUser = &v
	}

	if !plan.BindPassword.Equal(state.BindPassword) {
		v := plan.BindPassword.ValueString()
		patch.BindPassword = &v
	}

	if !plan.CACertificate.Equal(state.CACertificate) {
		if plan.CACertificate.IsNull() || plan.CACertificate.ValueString() == "" {
			var nilRef *client.NamedReference
			patch.CACertificate = &nilRef
		} else {
			ref := &client.NamedReference{Name: plan.CACertificate.ValueString()}
			patch.CACertificate = &ref
		}
	}

	if !plan.CACertificateGroup.Equal(state.CACertificateGroup) {
		if plan.CACertificateGroup.IsNull() || plan.CACertificateGroup.ValueString() == "" {
			var nilRef *client.NamedReference
			patch.CACertificateGroup = &nilRef
		} else {
			ref := &client.NamedReference{Name: plan.CACertificateGroup.ValueString()}
			patch.CACertificateGroup = &ref
		}
	}

	// Management sub-object: build only if any sub-field changed.
	mgmtChanged := !plan.UserLoginAttribute.Equal(state.UserLoginAttribute) ||
		!plan.UserObjectClass.Equal(state.UserObjectClass) ||
		!plan.SSHPublicKeyAttribute.Equal(state.SSHPublicKeyAttribute)
	if mgmtChanged {
		mgmtPatch := &client.DirectoryServiceManagementPatch{}
		if !plan.UserLoginAttribute.Equal(state.UserLoginAttribute) {
			v := plan.UserLoginAttribute.ValueString()
			mgmtPatch.UserLoginAttribute = &v
		}
		if !plan.UserObjectClass.Equal(state.UserObjectClass) {
			v := plan.UserObjectClass.ValueString()
			mgmtPatch.UserObjectClass = &v
		}
		if !plan.SSHPublicKeyAttribute.Equal(state.SSHPublicKeyAttribute) {
			v := plan.SSHPublicKeyAttribute.ValueString()
			mgmtPatch.SSHPublicKeyAttribute = &v
		}
		patch.Management = mgmtPatch
	}

	result, err := r.client.PatchDirectoryServiceManagement(ctx, managementDirectoryServiceName, patch)
	if err != nil {
		resp.Diagnostics.AddError("Error updating directory service management config", err.Error())
		return
	}

	resp.Diagnostics.Append(mapDirectoryServiceToModel(ctx, result, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Preserve write-only bind_password from plan.
	plan.BindPassword = state.BindPassword
	if !req.Plan.Raw.IsNull() {
		var planData directoryServiceManagementModel
		req.Plan.Get(ctx, &planData)
		plan.BindPassword = planData.BindPassword
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *directoryServiceManagementResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data directoryServiceManagementModel
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

	falseVal := false
	emptyStr := ""
	emptyURIs := []string{}
	var nilRef *client.NamedReference
	patch := client.DirectoryServicePatch{
		Enabled:            &falseVal,
		URIs:               &emptyURIs,
		BaseDN:             &emptyStr,
		BindUser:           &emptyStr,
		CACertificate:      &nilRef,
		CACertificateGroup: &nilRef,
		Management: &client.DirectoryServiceManagementPatch{
			UserLoginAttribute:    &emptyStr,
			UserObjectClass:       &emptyStr,
			SSHPublicKeyAttribute: &emptyStr,
		},
		// BindPassword intentionally omitted — never re-sent.
	}
	if _, err := r.client.PatchDirectoryServiceManagement(ctx, managementDirectoryServiceName, patch); err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error resetting directory service management config", err.Error())
	}
}

func (r *directoryServiceManagementResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	name := req.ID
	if name == "" {
		name = managementDirectoryServiceName
	}
	ds, err := r.client.GetDirectoryServiceManagement(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError("Error importing directory service management config", err.Error())
		return
	}

	var data directoryServiceManagementModel
	data.Timeouts = nullTimeoutsValue()
	data.BindPassword = types.StringValue("") // write-once: leave empty per D-00-e

	resp.Diagnostics.Append(mapDirectoryServiceToModel(ctx, ds, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Re-set BindPassword to empty after mapping (mapper does not overwrite password).
	data.BindPassword = types.StringValue("")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// mapDirectoryServiceToModel maps all API fields to the Terraform model.
// Does NOT touch data.BindPassword — caller is responsible for preserving it.
func mapDirectoryServiceToModel(ctx context.Context, ds *client.DirectoryService, data *directoryServiceManagementModel) diag.Diagnostics {
	var diags diag.Diagnostics

	data.ID = types.StringValue(ds.ID)
	data.Enabled = types.BoolValue(ds.Enabled)
	data.BaseDN = types.StringValue(ds.BaseDN)
	data.BindUser = types.StringValue(ds.BindUser)
	data.UserLoginAttribute = types.StringValue(ds.Management.UserLoginAttribute)
	data.UserObjectClass = types.StringValue(ds.Management.UserObjectClass)
	data.SSHPublicKeyAttribute = types.StringValue(ds.Management.SSHPublicKeyAttribute)

	// CACertificate: use empty string when nil (Computed field, avoids perpetual diff).
	if ds.CACertificate != nil {
		data.CACertificate = types.StringValue(ds.CACertificate.Name)
	} else {
		data.CACertificate = types.StringValue("")
	}

	// CACertificateGroup: same pattern.
	if ds.CACertificateGroup != nil {
		data.CACertificateGroup = types.StringValue(ds.CACertificateGroup.Name)
	} else {
		data.CACertificateGroup = types.StringValue("")
	}

	// URIs list.
	if len(ds.URIs) > 0 {
		uriList, d := types.ListValueFrom(ctx, types.StringType, ds.URIs)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		data.URIs = uriList
	} else {
		data.URIs = emptyStringList()
	}

	// Services list.
	if len(ds.Services) > 0 {
		svcList, d := types.ListValueFrom(ctx, types.StringType, ds.Services)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		data.Services = svcList
	} else {
		data.Services = emptyStringList()
	}

	return diags
}

// buildDSMPatchFromPlan builds a DirectoryServicePatch from the plan.
// For fields not set in plan (null/unknown), the patch field is left nil (omitted).
// This is used for Create where we send all plan-specified fields.
func buildDSMPatchFromPlan(ctx context.Context, plan directoryServiceManagementModel, _ directoryServiceManagementModel) (client.DirectoryServicePatch, diag.Diagnostics) {
	var diags diag.Diagnostics
	patch := client.DirectoryServicePatch{}

	if !plan.Enabled.IsNull() && !plan.Enabled.IsUnknown() {
		v := plan.Enabled.ValueBool()
		patch.Enabled = &v
	}

	if !plan.URIs.IsNull() && !plan.URIs.IsUnknown() {
		uris, d := listToStrings(ctx, plan.URIs)
		diags.Append(d...)
		if diags.HasError() {
			return patch, diags
		}
		patch.URIs = &uris
	}

	if !plan.BaseDN.IsNull() && !plan.BaseDN.IsUnknown() {
		v := plan.BaseDN.ValueString()
		patch.BaseDN = &v
	}

	if !plan.BindUser.IsNull() && !plan.BindUser.IsUnknown() {
		v := plan.BindUser.ValueString()
		patch.BindUser = &v
	}

	if !plan.BindPassword.IsNull() && !plan.BindPassword.IsUnknown() && plan.BindPassword.ValueString() != "" {
		v := plan.BindPassword.ValueString()
		patch.BindPassword = &v
	}

	if !plan.CACertificate.IsNull() && !plan.CACertificate.IsUnknown() && plan.CACertificate.ValueString() != "" {
		ref := &client.NamedReference{Name: plan.CACertificate.ValueString()}
		patch.CACertificate = &ref
	}

	if !plan.CACertificateGroup.IsNull() && !plan.CACertificateGroup.IsUnknown() && plan.CACertificateGroup.ValueString() != "" {
		ref := &client.NamedReference{Name: plan.CACertificateGroup.ValueString()}
		patch.CACertificateGroup = &ref
	}

	// Management sub-object.
	anyMgmt := (!plan.UserLoginAttribute.IsNull() && !plan.UserLoginAttribute.IsUnknown()) ||
		(!plan.UserObjectClass.IsNull() && !plan.UserObjectClass.IsUnknown()) ||
		(!plan.SSHPublicKeyAttribute.IsNull() && !plan.SSHPublicKeyAttribute.IsUnknown())
	if anyMgmt {
		mgmt := &client.DirectoryServiceManagementPatch{}
		if !plan.UserLoginAttribute.IsNull() && !plan.UserLoginAttribute.IsUnknown() {
			v := plan.UserLoginAttribute.ValueString()
			mgmt.UserLoginAttribute = &v
		}
		if !plan.UserObjectClass.IsNull() && !plan.UserObjectClass.IsUnknown() {
			v := plan.UserObjectClass.ValueString()
			mgmt.UserObjectClass = &v
		}
		if !plan.SSHPublicKeyAttribute.IsNull() && !plan.SSHPublicKeyAttribute.IsUnknown() {
			v := plan.SSHPublicKeyAttribute.ValueString()
			mgmt.SSHPublicKeyAttribute = &v
		}
		patch.Management = mgmt
	}

	return patch, diags
}
