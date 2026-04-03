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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// Ensure tlsPolicyResource satisfies the resource interfaces.
var _ resource.Resource = &tlsPolicyResource{}
var _ resource.ResourceWithConfigure = &tlsPolicyResource{}
var _ resource.ResourceWithImportState = &tlsPolicyResource{}
var _ resource.ResourceWithUpgradeState = &tlsPolicyResource{}

// tlsPolicyResource implements the flashblade_tls_policy resource.
type tlsPolicyResource struct {
	client *client.FlashBladeClient
}

// NewTlsPolicyResource is the factory function registered in the provider.
func NewTlsPolicyResource() resource.Resource {
	return &tlsPolicyResource{}
}

// ---------- model structs ----------------------------------------------------

// tlsPolicyModel is the Terraform state model for the flashblade_tls_policy resource.
type tlsPolicyModel struct {
	ID                               types.String   `tfsdk:"id"`
	Name                             types.String   `tfsdk:"name"`
	ApplianceCertificate             types.String   `tfsdk:"appliance_certificate"`
	ClientCertificatesRequired       types.Bool     `tfsdk:"client_certificates_required"`
	DisabledTlsCiphers               types.List     `tfsdk:"disabled_tls_ciphers"`
	Enabled                          types.Bool     `tfsdk:"enabled"`
	EnabledTlsCiphers                types.List     `tfsdk:"enabled_tls_ciphers"`
	IsLocal                          types.Bool     `tfsdk:"is_local"`
	MinTlsVersion                    types.String   `tfsdk:"min_tls_version"`
	PolicyType                       types.String   `tfsdk:"policy_type"`
	TrustedClientCertificateAuthority types.String  `tfsdk:"trusted_client_certificate_authority"`
	VerifyClientCertificateTrust     types.Bool     `tfsdk:"verify_client_certificate_trust"`
	Timeouts                         timeouts.Value `tfsdk:"timeouts"`
}

// ---------- resource interface methods --------------------------------------

// Metadata sets the Terraform type name.
func (r *tlsPolicyResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_tls_policy"
}

// Schema defines the resource schema.
func (r *tlsPolicyResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     0,
		Description: "Manages a FlashBlade TLS policy that defines TLS version, cipher suites, and mTLS settings for network interfaces.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the TLS policy.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the TLS policy. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"appliance_certificate": schema.StringAttribute{
				Optional:    true,
				Description: "The name of the certificate used by the appliance for TLS connections.",
			},
			"client_certificates_required": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "When true, clients must present a certificate for mTLS. Defaults to false.",
			},
			"disabled_tls_ciphers": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "List of TLS cipher suites to disable.",
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the TLS policy is enabled. Defaults to true.",
			},
			"enabled_tls_ciphers": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "List of explicitly enabled TLS cipher suites.",
			},
			"is_local": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether this TLS policy is local to the array.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"min_tls_version": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The minimum TLS version required (e.g. TLSv1.2, TLSv1.3).",
			},
			"policy_type": schema.StringAttribute{
				Computed:    true,
				Description: "The type of the TLS policy.",
			},
			"trusted_client_certificate_authority": schema.StringAttribute{
				Optional:    true,
				Description: "The name of the certificate authority used to verify client certificates for mTLS.",
			},
			"verify_client_certificate_trust": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "When true, client certificates are verified against the trusted CA.",
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
func (r *tlsPolicyResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}

// Configure injects the FlashBladeClient into the resource.
func (r *tlsPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new TLS policy.
func (r *tlsPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data tlsPolicyModel
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

	post := client.TlsPolicyPost{}

	if !data.ApplianceCertificate.IsNull() && !data.ApplianceCertificate.IsUnknown() {
		post.ApplianceCertificate = &client.NamedReference{Name: data.ApplianceCertificate.ValueString()}
	}
	if !data.ClientCertificatesRequired.IsNull() && !data.ClientCertificatesRequired.IsUnknown() {
		post.ClientCertificatesRequired = data.ClientCertificatesRequired.ValueBool()
	}
	if !data.DisabledTlsCiphers.IsNull() && !data.DisabledTlsCiphers.IsUnknown() {
		post.DisabledTlsCiphers = listToStringSlice(ctx, data.DisabledTlsCiphers)
	}
	if !data.Enabled.IsNull() && !data.Enabled.IsUnknown() {
		post.Enabled = data.Enabled.ValueBool()
	}
	if !data.EnabledTlsCiphers.IsNull() && !data.EnabledTlsCiphers.IsUnknown() {
		post.EnabledTlsCiphers = listToStringSlice(ctx, data.EnabledTlsCiphers)
	}
	if !data.MinTlsVersion.IsNull() && !data.MinTlsVersion.IsUnknown() {
		post.MinTlsVersion = data.MinTlsVersion.ValueString()
	}
	if !data.TrustedClientCertificateAuthority.IsNull() && !data.TrustedClientCertificateAuthority.IsUnknown() {
		post.TrustedClientCertificateAuthority = &client.NamedReference{Name: data.TrustedClientCertificateAuthority.ValueString()}
	}
	if !data.VerifyClientCertificateTrust.IsNull() && !data.VerifyClientCertificateTrust.IsUnknown() {
		post.VerifyClientCertificateTrust = data.VerifyClientCertificateTrust.ValueBool()
	}

	policy, err := r.client.PostTlsPolicy(ctx, data.Name.ValueString(), post)
	if err != nil {
		resp.Diagnostics.AddError("Error creating TLS policy", err.Error())
		return
	}

	mapTlsPolicyToModel(policy, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes Terraform state from the API, logging field-level drift.
func (r *tlsPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data tlsPolicyModel
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
	policy, err := r.client.GetTlsPolicy(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading TLS policy", err.Error())
		return
	}

	// Drift detection on all mutable/computed fields.
	apiApplianceCert := ""
	if policy.ApplianceCertificate != nil {
		apiApplianceCert = policy.ApplianceCertificate.Name
	}
	if data.ApplianceCertificate.ValueString() != apiApplianceCert {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": name,
			"field":    "appliance_certificate",
			"was":      data.ApplianceCertificate.ValueString(),
			"now":      apiApplianceCert,
		})
	}

	if data.ClientCertificatesRequired.ValueBool() != policy.ClientCertificatesRequired {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": name,
			"field":    "client_certificates_required",
			"was":      data.ClientCertificatesRequired.ValueBool(),
			"now":      policy.ClientCertificatesRequired,
		})
	}

	wasDisabled := strings.Join(listToStringSlice(ctx, data.DisabledTlsCiphers), ",")
	nowDisabled := strings.Join(policy.DisabledTlsCiphers, ",")
	if wasDisabled != nowDisabled {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": name,
			"field":    "disabled_tls_ciphers",
			"was":      wasDisabled,
			"now":      nowDisabled,
		})
	}

	if data.Enabled.ValueBool() != policy.Enabled {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": name,
			"field":    "enabled",
			"was":      data.Enabled.ValueBool(),
			"now":      policy.Enabled,
		})
	}

	wasEnabled := strings.Join(listToStringSlice(ctx, data.EnabledTlsCiphers), ",")
	nowEnabled := strings.Join(policy.EnabledTlsCiphers, ",")
	if wasEnabled != nowEnabled {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": name,
			"field":    "enabled_tls_ciphers",
			"was":      wasEnabled,
			"now":      nowEnabled,
		})
	}

	if data.IsLocal.ValueBool() != policy.IsLocal {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": name,
			"field":    "is_local",
			"was":      data.IsLocal.ValueBool(),
			"now":      policy.IsLocal,
		})
	}

	if data.MinTlsVersion.ValueString() != policy.MinTlsVersion {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": name,
			"field":    "min_tls_version",
			"was":      data.MinTlsVersion.ValueString(),
			"now":      policy.MinTlsVersion,
		})
	}

	if data.PolicyType.ValueString() != policy.PolicyType {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": name,
			"field":    "policy_type",
			"was":      data.PolicyType.ValueString(),
			"now":      policy.PolicyType,
		})
	}

	apiTrustedCA := ""
	if policy.TrustedClientCertificateAuthority != nil {
		apiTrustedCA = policy.TrustedClientCertificateAuthority.Name
	}
	if data.TrustedClientCertificateAuthority.ValueString() != apiTrustedCA {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": name,
			"field":    "trusted_client_certificate_authority",
			"was":      data.TrustedClientCertificateAuthority.ValueString(),
			"now":      apiTrustedCA,
		})
	}

	if data.VerifyClientCertificateTrust.ValueBool() != policy.VerifyClientCertificateTrust {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": name,
			"field":    "verify_client_certificate_trust",
			"was":      data.VerifyClientCertificateTrust.ValueBool(),
			"now":      policy.VerifyClientCertificateTrust,
		})
	}

	mapTlsPolicyToModel(policy, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update applies changes to an existing TLS policy.
func (r *tlsPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state tlsPolicyModel
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

	patch := client.TlsPolicyPatch{}

	if !plan.ApplianceCertificate.Equal(state.ApplianceCertificate) {
		if plan.ApplianceCertificate.IsNull() {
			// Set to null: outer non-nil, inner nil.
			inner := (*client.NamedReference)(nil)
			patch.ApplianceCertificate = &inner
		} else {
			ref := &client.NamedReference{Name: plan.ApplianceCertificate.ValueString()}
			patch.ApplianceCertificate = &ref
		}
	}

	if !plan.ClientCertificatesRequired.Equal(state.ClientCertificatesRequired) {
		v := plan.ClientCertificatesRequired.ValueBool()
		patch.ClientCertificatesRequired = &v
	}

	if !plan.DisabledTlsCiphers.Equal(state.DisabledTlsCiphers) {
		v := listToStringSlice(ctx, plan.DisabledTlsCiphers)
		patch.DisabledTlsCiphers = &v
	}

	if !plan.Enabled.Equal(state.Enabled) {
		v := plan.Enabled.ValueBool()
		patch.Enabled = &v
	}

	if !plan.EnabledTlsCiphers.Equal(state.EnabledTlsCiphers) {
		v := listToStringSlice(ctx, plan.EnabledTlsCiphers)
		patch.EnabledTlsCiphers = &v
	}

	if !plan.MinTlsVersion.Equal(state.MinTlsVersion) {
		v := plan.MinTlsVersion.ValueString()
		patch.MinTlsVersion = &v
	}

	if !plan.TrustedClientCertificateAuthority.Equal(state.TrustedClientCertificateAuthority) {
		if plan.TrustedClientCertificateAuthority.IsNull() {
			inner := (*client.NamedReference)(nil)
			patch.TrustedClientCertificateAuthority = &inner
		} else {
			ref := &client.NamedReference{Name: plan.TrustedClientCertificateAuthority.ValueString()}
			patch.TrustedClientCertificateAuthority = &ref
		}
	}

	if !plan.VerifyClientCertificateTrust.Equal(state.VerifyClientCertificateTrust) {
		v := plan.VerifyClientCertificateTrust.ValueBool()
		patch.VerifyClientCertificateTrust = &v
	}

	policy, err := r.client.PatchTlsPolicy(ctx, state.Name.ValueString(), patch)
	if err != nil {
		resp.Diagnostics.AddError("Error updating TLS policy", err.Error())
		return
	}

	mapTlsPolicyToModel(policy, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes a TLS policy.
func (r *tlsPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data tlsPolicyModel
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

	err := r.client.DeleteTlsPolicy(ctx, data.Name.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting TLS policy", err.Error())
		return
	}
}

// ImportState imports an existing TLS policy by name.
func (r *tlsPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	name := req.ID
	policy, err := r.client.GetTlsPolicy(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError("Error importing TLS policy", err.Error())
		return
	}

	var data tlsPolicyModel
	data.Timeouts = nullTimeoutsValue()

	mapTlsPolicyToModel(policy, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ---------- helpers ---------------------------------------------------------

// mapTlsPolicyToModel maps a client.TlsPolicy to a tlsPolicyModel.
func mapTlsPolicyToModel(policy *client.TlsPolicy, data *tlsPolicyModel) {
	data.ID = types.StringValue(policy.ID)
	data.Name = types.StringValue(policy.Name)

	if policy.ApplianceCertificate != nil {
		data.ApplianceCertificate = types.StringValue(policy.ApplianceCertificate.Name)
	} else {
		data.ApplianceCertificate = types.StringNull()
	}

	data.ClientCertificatesRequired = types.BoolValue(policy.ClientCertificatesRequired)

	if len(policy.DisabledTlsCiphers) == 0 {
		data.DisabledTlsCiphers = types.ListValueMust(types.StringType, []attr.Value{})
	} else {
		elems := make([]attr.Value, len(policy.DisabledTlsCiphers))
		for i, c := range policy.DisabledTlsCiphers {
			elems[i] = types.StringValue(c)
		}
		data.DisabledTlsCiphers = types.ListValueMust(types.StringType, elems)
	}

	data.Enabled = types.BoolValue(policy.Enabled)

	if len(policy.EnabledTlsCiphers) == 0 {
		data.EnabledTlsCiphers = types.ListValueMust(types.StringType, []attr.Value{})
	} else {
		elems := make([]attr.Value, len(policy.EnabledTlsCiphers))
		for i, c := range policy.EnabledTlsCiphers {
			elems[i] = types.StringValue(c)
		}
		data.EnabledTlsCiphers = types.ListValueMust(types.StringType, elems)
	}

	data.IsLocal = types.BoolValue(policy.IsLocal)
	data.MinTlsVersion = types.StringValue(policy.MinTlsVersion)
	data.PolicyType = types.StringValue(policy.PolicyType)

	if policy.TrustedClientCertificateAuthority != nil {
		data.TrustedClientCertificateAuthority = types.StringValue(policy.TrustedClientCertificateAuthority.Name)
	} else {
		data.TrustedClientCertificateAuthority = types.StringNull()
	}

	data.VerifyClientCertificateTrust = types.BoolValue(policy.VerifyClientCertificateTrust)
}

// listToStringSlice converts a types.List to a []string.
// Returns an empty slice if the list is null or unknown.
func listToStringSlice(ctx context.Context, list types.List) []string {
	if list.IsNull() || list.IsUnknown() {
		return []string{}
	}
	var elems []string
	_ = list.ElementsAs(ctx, &elems, false)
	return elems
}
