package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

var _ resource.Resource = &snmpManagerResource{}
var _ resource.ResourceWithConfigure = &snmpManagerResource{}
var _ resource.ResourceWithImportState = &snmpManagerResource{}
var _ resource.ResourceWithUpgradeState = &snmpManagerResource{}

// snmpManagerResource implements the flashblade_snmp_manager resource.
type snmpManagerResource struct {
	client *client.FlashBladeClient
}

func NewSnmpManagerResource() resource.Resource {
	return &snmpManagerResource{}
}

// ---------- model structs ----------------------------------------------------

// snmpManagerModel is the top-level model for the flashblade_snmp_manager resource.
type snmpManagerModel struct {
	ID           types.String   `tfsdk:"id"`
	Name         types.String   `tfsdk:"name"`
	Host         types.String   `tfsdk:"host"`
	Notification types.String   `tfsdk:"notification"`
	Version      types.String   `tfsdk:"version"`
	V2c          types.Object   `tfsdk:"v2c"`
	V3           types.Object   `tfsdk:"v3"`
	Timeouts     timeouts.Value `tfsdk:"timeouts"`
}

// snmpV2cModel maps the v2c nested object.
type snmpV2cModel struct {
	Community types.String `tfsdk:"community"`
}

// snmpV3Model maps the v3 nested object.
type snmpV3Model struct {
	User              types.String `tfsdk:"user"`
	AuthProtocol      types.String `tfsdk:"auth_protocol"`
	AuthPassphrase    types.String `tfsdk:"auth_passphrase"`
	PrivacyProtocol   types.String `tfsdk:"privacy_protocol"`
	PrivacyPassphrase types.String `tfsdk:"privacy_passphrase"`
}

// snmpV2cAttrTypes is the attribute type map for the v2c nested object.
var snmpV2cAttrTypes = map[string]attr.Type{
	"community": types.StringType,
}

// snmpV3AttrTypes is the attribute type map for the v3 nested object.
var snmpV3AttrTypes = map[string]attr.Type{
	"user":               types.StringType,
	"auth_protocol":      types.StringType,
	"auth_passphrase":    types.StringType,
	"privacy_protocol":   types.StringType,
	"privacy_passphrase": types.StringType,
}

// ---------- resource interface methods --------------------------------------

func (r *snmpManagerResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_snmp_manager"
}

func (r *snmpManagerResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     0,
		Description: "Manages a FlashBlade SNMP trap/inform destination (SNMP manager) for alerts.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the SNMP manager.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the SNMP manager. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"host": schema.StringAttribute{
				Required:    true,
				Description: "DNS name or IP address (with optional :port) of the SNMP receiver.",
			},
			"notification": schema.StringAttribute{
				Required:    true,
				Description: "Notification delivery mode: `inform` (acknowledged) or `trap` (fire-and-forget).",
				Validators: []validator.String{
					stringvalidator.OneOf("inform", "trap"),
				},
			},
			"version": schema.StringAttribute{
				Required:    true,
				Description: "SNMP protocol version: `v2c` or `v3`. Switching in place is permitted (no resource replacement).",
				Validators: []validator.String{
					stringvalidator.OneOf("v2c", "v3"),
				},
			},
			"v2c": schema.SingleNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "SNMPv2c configuration. Required when `version = \"v2c\"`.",
				Attributes: map[string]schema.Attribute{
					"community": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Sensitive:   true,
						Description: "Community string. Write-once: never returned by the API on GET; state preserves the user-supplied value.",
						Validators: []validator.String{
							stringvalidator.LengthAtMost(32),
						},
					},
				},
			},
			"v3": schema.SingleNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "SNMPv3 configuration. Required when `version = \"v3\"`.",
				Attributes: map[string]schema.Attribute{
					"user": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "SNMPv3 username.",
					},
					"auth_protocol": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Authentication protocol: `MD5` or `SHA`.",
						Validators: []validator.String{
							stringvalidator.OneOf("MD5", "SHA"),
						},
					},
					"auth_passphrase": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Sensitive:   true,
						Description: "Authentication passphrase (max 32 chars). Write-once: never returned by the API on GET; state preserves the user-supplied value.",
						Validators: []validator.String{
							stringvalidator.LengthAtMost(32),
						},
					},
					"privacy_protocol": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Privacy protocol: `AES` or `DES`.",
						Validators: []validator.String{
							stringvalidator.OneOf("AES", "DES"),
						},
					},
					"privacy_passphrase": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Sensitive:   true,
						Description: "Privacy passphrase (8..63 chars). Write-once: never returned by the API on GET; state preserves the user-supplied value.",
						Validators: []validator.String{
							stringvalidator.LengthBetween(8, 63),
						},
					},
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

func (r *snmpManagerResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}

func (r *snmpManagerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *snmpManagerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan snmpManagerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := plan.Timeouts.Create(ctx, 20*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	// Extract plan-supplied v2c/v3 (with sensitive fields) BEFORE the POST so we
	// can preserve them in state (the API never echoes sensitive fields).
	planV2c, d := v2cFromObject(ctx, plan.V2c)
	resp.Diagnostics.Append(d...)
	planV3, d := v3ModelFromObject(ctx, plan.V3)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	post := client.SnmpManagerPost{
		Host:         plan.Host.ValueString(),
		Notification: plan.Notification.ValueString(),
		Version:      plan.Version.ValueString(),
	}
	if planV2c != nil {
		post.V2c = &client.SnmpV2c{Community: planV2c.Community.ValueString()}
	}
	if planV3 != nil {
		post.V3 = &client.SnmpV3Post{
			User:              planV3.User.ValueString(),
			AuthProtocol:      planV3.AuthProtocol.ValueString(),
			AuthPassphrase:    planV3.AuthPassphrase.ValueString(),
			PrivacyProtocol:   planV3.PrivacyProtocol.ValueString(),
			PrivacyPassphrase: planV3.PrivacyPassphrase.ValueString(),
		}
	}

	mgr, err := r.client.PostSnmpManager(ctx, plan.Name.ValueString(), post)
	if err != nil {
		resp.Diagnostics.AddError("Error creating SNMP manager", err.Error())
		return
	}

	resp.Diagnostics.Append(mapSnmpManagerToModel(ctx, mgr, &plan, planV2c, planV3, nil)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes Terraform state from the API, logging field-level drift on
// the 6 non-sensitive leaves. Sensitive fields (community, auth_passphrase,
// privacy_passphrase) are NEVER overwritten from the API response.
func (r *snmpManagerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data snmpManagerModel
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
	mgr, err := r.client.GetSnmpManager(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading SNMP manager", err.Error())
		return
	}

	// Decode existing nested objects to preserve sensitive write-once fields.
	prevV2c, d := v2cFromObject(ctx, data.V2c)
	resp.Diagnostics.Append(d...)
	prevV3, d := v3ModelFromObject(ctx, data.V3)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Host.ValueString() != mgr.Host {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": name, "field": "host",
			"was": data.Host.ValueString(), "now": mgr.Host,
		})
	}
	if data.Notification.ValueString() != mgr.Notification {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": name, "field": "notification",
			"was": data.Notification.ValueString(), "now": mgr.Notification,
		})
	}
	if data.Version.ValueString() != mgr.Version {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": name, "field": "version",
			"was": data.Version.ValueString(), "now": mgr.Version,
		})
	}

	oldV3User, oldV3Auth, oldV3Priv := "", "", ""
	if prevV3 != nil {
		oldV3User = prevV3.User.ValueString()
		oldV3Auth = prevV3.AuthProtocol.ValueString()
		oldV3Priv = prevV3.PrivacyProtocol.ValueString()
	}
	newV3User, newV3Auth, newV3Priv := "", "", ""
	if mgr.V3 != nil {
		newV3User = mgr.V3.User
		newV3Auth = mgr.V3.AuthProtocol
		newV3Priv = mgr.V3.PrivacyProtocol
	}
	if oldV3User != newV3User {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": name, "field": "v3.user",
			"was": oldV3User, "now": newV3User,
		})
	}
	if oldV3Auth != newV3Auth {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": name, "field": "v3.auth_protocol",
			"was": oldV3Auth, "now": newV3Auth,
		})
	}
	if oldV3Priv != newV3Priv {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": name, "field": "v3.privacy_protocol",
			"was": oldV3Priv, "now": newV3Priv,
		})
	}

	resp.Diagnostics.Append(mapSnmpManagerToModel(ctx, mgr, &data, prevV2c, prevV3, prevV3)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *snmpManagerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state snmpManagerModel
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

	planV2c, d := v2cFromObject(ctx, plan.V2c)
	resp.Diagnostics.Append(d...)
	planV3, d := v3ModelFromObject(ctx, plan.V3)
	resp.Diagnostics.Append(d...)
	stateV2c, d := v2cFromObject(ctx, state.V2c)
	resp.Diagnostics.Append(d...)
	stateV3, d := v3ModelFromObject(ctx, state.V3)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	patch := client.SnmpManagerPatch{}
	if !plan.Host.Equal(state.Host) {
		v := plan.Host.ValueString()
		patch.Host = &v
	}
	if !plan.Notification.Equal(state.Notification) {
		v := plan.Notification.ValueString()
		patch.Notification = &v
	}
	versionChanged := !plan.Version.Equal(state.Version)
	if versionChanged {
		v := plan.Version.ValueString()
		patch.Version = &v
	}

	// Nested blocks are atomic: send the full block if any leaf inside changed,
	// or if the version flipped and the target block is now active.
	if v2cBlockChanged(planV2c, stateV2c) || (versionChanged && plan.Version.ValueString() == "v2c") {
		if planV2c != nil {
			patch.V2c = &client.SnmpV2c{Community: planV2c.Community.ValueString()}
		}
	}
	if v3BlockChanged(planV3, stateV3) || (versionChanged && plan.Version.ValueString() == "v3") {
		if planV3 != nil {
			patch.V3 = &client.SnmpV3{
				User:              planV3.User.ValueString(),
				AuthProtocol:      planV3.AuthProtocol.ValueString(),
				AuthPassphrase:    planV3.AuthPassphrase.ValueString(),
				PrivacyProtocol:   planV3.PrivacyProtocol.ValueString(),
				PrivacyPassphrase: planV3.PrivacyPassphrase.ValueString(),
			}
		}
	}

	_, err := r.client.PatchSnmpManager(ctx, state.Name.ValueString(), patch)
	if err != nil {
		resp.Diagnostics.AddError("Error updating SNMP manager", err.Error())
		return
	}

	// Re-read to refresh computed fields. Sensitive fields preserved from plan.
	mgr, err := r.client.GetSnmpManager(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading SNMP manager after update", err.Error())
		return
	}
	resp.Diagnostics.Append(mapSnmpManagerToModel(ctx, mgr, &plan, planV2c, planV3, nil)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *snmpManagerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data snmpManagerModel
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

	err := r.client.DeleteSnmpManager(ctx, data.Name.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting SNMP manager", err.Error())
		return
	}
}

// ImportState imports an existing SNMP manager by name. Sensitive write-once
// fields (community, auth_passphrase, privacy_passphrase) are set to null;
// the operator must re-supply them via HCL + apply.
func (r *snmpManagerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	name := req.ID
	mgr, err := r.client.GetSnmpManager(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError("Error importing SNMP manager", err.Error())
		return
	}

	var data snmpManagerModel
	data.Timeouts = nullTimeoutsValue()
	resp.Diagnostics.Append(mapSnmpManagerToModel(ctx, mgr, &data, nil, nil, nil)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ---------- helpers ---------------------------------------------------------

// v2cFromObject decodes a types.Object into snmpV2cModel.
// Returns nil when the object is null or unknown.
func v2cFromObject(ctx context.Context, obj types.Object) (*snmpV2cModel, diag.Diagnostics) {
	if obj.IsNull() || obj.IsUnknown() {
		return nil, nil
	}
	var m snmpV2cModel
	diags := obj.As(ctx, &m, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
	if diags.HasError() {
		return nil, diags
	}
	return &m, diags
}

// v3ModelFromObject decodes a types.Object into snmpV3Model.
// Returns nil when the object is null or unknown.
func v3ModelFromObject(ctx context.Context, obj types.Object) (*snmpV3Model, diag.Diagnostics) {
	if obj.IsNull() || obj.IsUnknown() {
		return nil, nil
	}
	var m snmpV3Model
	diags := obj.As(ctx, &m, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
	if diags.HasError() {
		return nil, diags
	}
	return &m, diags
}

func v2cBlockChanged(plan, state *snmpV2cModel) bool {
	if plan == nil && state == nil {
		return false
	}
	if plan == nil || state == nil {
		return true
	}
	return !plan.Community.Equal(state.Community)
}

func v3BlockChanged(plan, state *snmpV3Model) bool {
	if plan == nil && state == nil {
		return false
	}
	if plan == nil || state == nil {
		return true
	}
	return !plan.User.Equal(state.User) ||
		!plan.AuthProtocol.Equal(state.AuthProtocol) ||
		!plan.AuthPassphrase.Equal(state.AuthPassphrase) ||
		!plan.PrivacyProtocol.Equal(state.PrivacyProtocol) ||
		!plan.PrivacyPassphrase.Equal(state.PrivacyPassphrase)
}

// mapSnmpManagerToModel maps a client.SnmpManager into the Terraform model.
// `preservedV2c` and `preservedV3` carry the user-supplied (Create/Update) or
// previously-stored (Read) sensitive values; the API never returns them on GET.
// `readPriorV3` carries the v3 model from the prior state (Read only) so that
// when the API now reports `v3 = nil` (e.g. version flipped to v2c) we still
// surface a non-null v3 in state when appropriate. Pass nil otherwise.
func mapSnmpManagerToModel(
	ctx context.Context,
	mgr *client.SnmpManager,
	data *snmpManagerModel,
	preservedV2c *snmpV2cModel,
	preservedV3 *snmpV3Model,
	_ *snmpV3Model,
) diag.Diagnostics {
	var diags diag.Diagnostics

	data.ID = types.StringValue(mgr.ID)
	data.Name = types.StringValue(mgr.Name)
	data.Host = types.StringValue(mgr.Host)
	data.Notification = types.StringValue(mgr.Notification)
	data.Version = types.StringValue(mgr.Version)

	// v2c mapping.
	if mgr.V2c != nil {
		v2 := snmpV2cModel{
			// skip sensitive write-once: community is never returned by the API on GET.
			// Preserve the user-supplied or previously-stored value.
			Community: types.StringNull(),
		}
		if preservedV2c != nil && !preservedV2c.Community.IsNull() && !preservedV2c.Community.IsUnknown() {
			v2.Community = preservedV2c.Community
		}
		obj, d := types.ObjectValueFrom(ctx, snmpV2cAttrTypes, v2)
		diags.Append(d...)
		if !diags.HasError() {
			data.V2c = obj
		}
	} else {
		data.V2c = types.ObjectNull(snmpV2cAttrTypes)
	}

	// v3 mapping.
	if mgr.V3 != nil {
		v3 := snmpV3Model{
			User:            stringFromAPI(mgr.V3.User),
			AuthProtocol:    stringFromAPI(mgr.V3.AuthProtocol),
			PrivacyProtocol: stringFromAPI(mgr.V3.PrivacyProtocol),
			// skip sensitive write-once: auth_passphrase is never returned by the API on GET.
			AuthPassphrase: types.StringNull(),
			// skip sensitive write-once: privacy_passphrase is never returned by the API on GET.
			PrivacyPassphrase: types.StringNull(),
		}
		if preservedV3 != nil {
			if !preservedV3.AuthPassphrase.IsNull() && !preservedV3.AuthPassphrase.IsUnknown() {
				v3.AuthPassphrase = preservedV3.AuthPassphrase
			}
			if !preservedV3.PrivacyPassphrase.IsNull() && !preservedV3.PrivacyPassphrase.IsUnknown() {
				v3.PrivacyPassphrase = preservedV3.PrivacyPassphrase
			}
		}
		obj, d := types.ObjectValueFrom(ctx, snmpV3AttrTypes, v3)
		diags.Append(d...)
		if !diags.HasError() {
			data.V3 = obj
		}
	} else {
		data.V3 = types.ObjectNull(snmpV3AttrTypes)
	}

	return diags
}

// stringFromAPI maps an API string to a Terraform types.String; empty string
// becomes null for cleaner state representation.
func stringFromAPI(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}
