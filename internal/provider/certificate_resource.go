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
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

var _ resource.Resource = &certificateResource{}
var _ resource.ResourceWithConfigure = &certificateResource{}
var _ resource.ResourceWithImportState = &certificateResource{}
var _ resource.ResourceWithUpgradeState = &certificateResource{}

// certificateResource implements the flashblade_certificate resource.
type certificateResource struct {
	client *client.FlashBladeClient
}

func NewCertificateResource() resource.Resource {
	return &certificateResource{}
}

// ---------- model structs ----------------------------------------------------

// certificateModel is the top-level model for the flashblade_certificate resource.
type certificateModel struct {
	ID                      types.String   `tfsdk:"id"`
	Name                    types.String   `tfsdk:"name"`
	Certificate             types.String   `tfsdk:"certificate"`
	CertificateType         types.String   `tfsdk:"certificate_type"`
	CommonName              types.String   `tfsdk:"common_name"`
	Country                 types.String   `tfsdk:"country"`
	Email                   types.String   `tfsdk:"email"`
	IntermediateCertificate types.String   `tfsdk:"intermediate_certificate"`
	IssuedBy                types.String   `tfsdk:"issued_by"`
	IssuedTo                types.String   `tfsdk:"issued_to"`
	KeyAlgorithm            types.String   `tfsdk:"key_algorithm"`
	KeySize                 types.Int64    `tfsdk:"key_size"`
	Locality                types.String   `tfsdk:"locality"`
	Organization            types.String   `tfsdk:"organization"`
	OrganizationalUnit      types.String   `tfsdk:"organizational_unit"`
	Passphrase              types.String   `tfsdk:"passphrase"`
	PrivateKey              types.String   `tfsdk:"private_key"`
	State                   types.String   `tfsdk:"state"`
	Status                  types.String   `tfsdk:"status"`
	SubjectAlternativeNames types.List     `tfsdk:"subject_alternative_names"`
	ValidFrom               types.Int64    `tfsdk:"valid_from"`
	ValidTo                 types.Int64    `tfsdk:"valid_to"`
	Timeouts                timeouts.Value `tfsdk:"timeouts"`
}

// ---------- resource interface methods --------------------------------------

func (r *certificateResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_certificate"
}

// Schema defines the resource schema.
func (r *certificateResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     0,
		Description: "Manages a FlashBlade TLS certificate (import-mode: PEM certificate with private key).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the certificate.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the certificate. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"certificate": schema.StringAttribute{
				Required:    true,
				Description: "The PEM-encoded X.509 certificate body.",
			},
			"certificate_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The certificate type. Valid values: 'array' (FlashBlade identity, requires private_key) or 'external' (trusted external server such as AD). When unset, the provider infers 'array' if private_key is provided; otherwise the API defaults to 'external'. Immutable after creation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"common_name": schema.StringAttribute{
				Computed:    true,
				Description: "The common name (CN) extracted from the certificate.",
			},
			"country": schema.StringAttribute{
				Computed:    true,
				Description: "The country (C) field extracted from the certificate.",
			},
			"email": schema.StringAttribute{
				Computed:    true,
				Description: "The email address extracted from the certificate.",
			},
			"intermediate_certificate": schema.StringAttribute{
				Optional:    true,
				Description: "The PEM-encoded intermediate certificate chain.",
			},
			"issued_by": schema.StringAttribute{
				Computed:    true,
				Description: "The issuer of the certificate. Changes when the certificate is renewed.",
			},
			"issued_to": schema.StringAttribute{
				Computed:    true,
				Description: "The subject of the certificate. Changes when the certificate is renewed.",
			},
			"key_algorithm": schema.StringAttribute{
				Computed:    true,
				Description: "The key algorithm (e.g. RSA, EC). Changes when the certificate is renewed.",
			},
			"key_size": schema.Int64Attribute{
				Computed:    true,
				Description: "The key size in bits. Changes when the certificate is renewed.",
			},
			"locality": schema.StringAttribute{
				Computed:    true,
				Description: "The locality (L) field extracted from the certificate.",
			},
			"organization": schema.StringAttribute{
				Computed:    true,
				Description: "The organization (O) field extracted from the certificate.",
			},
			"organizational_unit": schema.StringAttribute{
				Computed:    true,
				Description: "The organizational unit (OU) field extracted from the certificate.",
			},
			"passphrase": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "The passphrase protecting the private key. Not returned by the API after creation.",
			},
			"private_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "The PEM-encoded private key. Not returned by the API after creation.",
			},
			"state": schema.StringAttribute{
				Computed:    true,
				Description: "The state/province (ST) field extracted from the certificate.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The certificate status (e.g. imported, self-signed). Changes when the certificate is renewed.",
			},
			"subject_alternative_names": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "The subject alternative names (SANs) extracted from the certificate.",
			},
			"valid_from": schema.Int64Attribute{
				Computed:    true,
				Description: "The Unix timestamp (milliseconds) from which the certificate is valid. Changes when renewed.",
			},
			"valid_to": schema.Int64Attribute{
				Computed:    true,
				Description: "The Unix timestamp (milliseconds) until which the certificate is valid. Changes when renewed.",
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
func (r *certificateResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}

// Configure injects the FlashBladeClient into the resource.
func (r *certificateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create imports a new certificate.
func (r *certificateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data certificateModel
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

	post := client.CertificatePost{
		Certificate: data.Certificate.ValueString(),
	}
	privateKeySet := !data.PrivateKey.IsNull() && data.PrivateKey.ValueString() != ""
	switch {
	case !data.CertificateType.IsNull() && !data.CertificateType.IsUnknown():
		post.CertificateType = data.CertificateType.ValueString()
	case privateKeySet:
		// The FlashBlade API defaults to "external" when certificate_type is
		// omitted, and rejects private_key for external certificates. Infer
		// "array" when a private key is supplied so the common case works
		// without the user needing to set certificate_type explicitly.
		// Note: the swagger description names this value "appliance", but the
		// real API only accepts "array" (verified against the Pure-Storage
		// Ansible FlashBlade collection and live array behavior).
		post.CertificateType = "array"
	}
	wasIntermediateNull := data.IntermediateCertificate.IsNull()
	if !wasIntermediateNull {
		post.IntermediateCertificate = data.IntermediateCertificate.ValueString()
	}
	if !data.Passphrase.IsNull() {
		post.Passphrase = data.Passphrase.ValueString()
	}
	if privateKeySet {
		post.PrivateKey = data.PrivateKey.ValueString()
	}

	cert, err := r.client.PostCertificate(ctx, data.Name.ValueString(), post)
	if err != nil {
		resp.Diagnostics.AddError("Error creating certificate", err.Error())
		return
	}

	// Preserve write-only fields from plan — API never returns them.
	privateKey := data.PrivateKey
	passphrase := data.Passphrase

	mapCertificateToModel(cert, &data)

	// The API returns "" for fields that were never set; Terraform requires the
	// state to stay null when the config was null (Optional, non-Computed).
	if wasIntermediateNull && cert.IntermediateCertificate == "" {
		data.IntermediateCertificate = types.StringNull()
	}

	data.PrivateKey = privateKey
	data.Passphrase = passphrase

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes Terraform state from the API, logging field-level drift.
func (r *certificateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data certificateModel
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
	cert, err := r.client.GetCertificate(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading certificate", err.Error())
		return
	}

	// Drift detection: compare old state vs API response and log each changed field.
	if data.Certificate.ValueString() != cert.Certificate {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": name,
			"field":    "certificate",
			"was":      data.Certificate.ValueString(),
			"now":      cert.Certificate,
		})
	}
	if data.CertificateType.ValueString() != cert.CertificateType {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": name,
			"field":    "certificate_type",
			"was":      data.CertificateType.ValueString(),
			"now":      cert.CertificateType,
		})
	}
	if data.CommonName.ValueString() != cert.CommonName {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": name,
			"field":    "common_name",
			"was":      data.CommonName.ValueString(),
			"now":      cert.CommonName,
		})
	}
	if data.Country.ValueString() != cert.Country {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": name,
			"field":    "country",
			"was":      data.Country.ValueString(),
			"now":      cert.Country,
		})
	}
	if data.Email.ValueString() != cert.Email {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": name,
			"field":    "email",
			"was":      data.Email.ValueString(),
			"now":      cert.Email,
		})
	}
	if data.IntermediateCertificate.ValueString() != cert.IntermediateCertificate {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": name,
			"field":    "intermediate_certificate",
			"was":      data.IntermediateCertificate.ValueString(),
			"now":      cert.IntermediateCertificate,
		})
	}
	if data.IssuedBy.ValueString() != cert.IssuedBy {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": name,
			"field":    "issued_by",
			"was":      data.IssuedBy.ValueString(),
			"now":      cert.IssuedBy,
		})
	}
	if data.IssuedTo.ValueString() != cert.IssuedTo {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": name,
			"field":    "issued_to",
			"was":      data.IssuedTo.ValueString(),
			"now":      cert.IssuedTo,
		})
	}
	if data.KeyAlgorithm.ValueString() != cert.KeyAlgorithm {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": name,
			"field":    "key_algorithm",
			"was":      data.KeyAlgorithm.ValueString(),
			"now":      cert.KeyAlgorithm,
		})
	}
	if data.KeySize.ValueInt64() != int64(cert.KeySize) {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": name,
			"field":    "key_size",
			"was":      data.KeySize.ValueInt64(),
			"now":      int64(cert.KeySize),
		})
	}
	if data.Locality.ValueString() != cert.Locality {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": name,
			"field":    "locality",
			"was":      data.Locality.ValueString(),
			"now":      cert.Locality,
		})
	}
	if data.Organization.ValueString() != cert.Organization {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": name,
			"field":    "organization",
			"was":      data.Organization.ValueString(),
			"now":      cert.Organization,
		})
	}
	if data.OrganizationalUnit.ValueString() != cert.OrganizationalUnit {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": name,
			"field":    "organizational_unit",
			"was":      data.OrganizationalUnit.ValueString(),
			"now":      cert.OrganizationalUnit,
		})
	}
	if data.State.ValueString() != cert.State {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": name,
			"field":    "state",
			"was":      data.State.ValueString(),
			"now":      cert.State,
		})
	}
	if data.Status.ValueString() != cert.Status {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": name,
			"field":    "status",
			"was":      data.Status.ValueString(),
			"now":      cert.Status,
		})
	}
	if data.ValidFrom.ValueInt64() != cert.ValidFrom {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": name,
			"field":    "valid_from",
			"was":      data.ValidFrom.ValueInt64(),
			"now":      cert.ValidFrom,
		})
	}
	if data.ValidTo.ValueInt64() != cert.ValidTo {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": name,
			"field":    "valid_to",
			"was":      data.ValidTo.ValueInt64(),
			"now":      cert.ValidTo,
		})
	}

	// Preserve write-only fields from prior state.
	privateKey := data.PrivateKey
	passphrase := data.Passphrase
	wasIntermediateNull := data.IntermediateCertificate.IsNull()

	mapCertificateToModel(cert, &data)

	if wasIntermediateNull && cert.IntermediateCertificate == "" {
		data.IntermediateCertificate = types.StringNull()
	}

	data.PrivateKey = privateKey
	data.Passphrase = passphrase

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update applies cert renewal to an existing certificate.
// Only calls the API if certificate or intermediate_certificate actually changed (renewal).
// If only private_key/passphrase changed (e.g., after import), the state is updated without
// an API call — the API already has the key and doesn't accept it in isolation.
func (r *certificateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state certificateModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	certChanged := !plan.Certificate.Equal(state.Certificate)
	intermediateChanged := !plan.IntermediateCertificate.Equal(state.IntermediateCertificate)

	// Only call the API if the actual certificate content changed (renewal scenario).
	if certChanged || intermediateChanged {
		updateTimeout, diags := plan.Timeouts.Update(ctx, 20*time.Minute)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		ctx, cancel := context.WithTimeout(ctx, updateTimeout)
		defer cancel()

		patch := client.CertificatePatch{}

		if certChanged {
			v := plan.Certificate.ValueString()
			patch.Certificate = &v
		}
		if intermediateChanged {
			v := plan.IntermediateCertificate.ValueString()
			patch.IntermediateCertificate = &v
		}
		// Include private_key and passphrase only when renewing the cert.
		if !plan.PrivateKey.IsNull() && plan.PrivateKey.ValueString() != "" {
			v := plan.PrivateKey.ValueString()
			patch.PrivateKey = &v
		}
		if !plan.Passphrase.IsNull() && plan.Passphrase.ValueString() != "" {
			v := plan.Passphrase.ValueString()
			patch.Passphrase = &v
		}

		cert, err := r.client.PatchCertificate(ctx, state.Name.ValueString(), patch)
		if err != nil {
			resp.Diagnostics.AddError("Error updating certificate", err.Error())
			return
		}

		// Preserve write-only fields before mapping overwrites the model.
		privateKey := plan.PrivateKey
		passphrase := plan.Passphrase
		wasIntermediateNull := plan.IntermediateCertificate.IsNull()

		mapCertificateToModel(cert, &plan)

		if wasIntermediateNull && cert.IntermediateCertificate == "" {
			plan.IntermediateCertificate = types.StringNull()
		}

		plan.PrivateKey = privateKey
		plan.Passphrase = passphrase
	} else {
		// No API call needed — only write-only fields changed (e.g., import → apply).
		// Copy all computed fields from current state so Terraform gets known values.
		plan.CommonName = state.CommonName
		plan.Country = state.Country
		plan.Email = state.Email
		plan.IssuedBy = state.IssuedBy
		plan.IssuedTo = state.IssuedTo
		plan.KeyAlgorithm = state.KeyAlgorithm
		plan.KeySize = state.KeySize
		plan.Locality = state.Locality
		plan.Organization = state.Organization
		plan.OrganizationalUnit = state.OrganizationalUnit
		plan.State = state.State
		plan.Status = state.Status
		plan.SubjectAlternativeNames = state.SubjectAlternativeNames
		plan.ValidFrom = state.ValidFrom
		plan.ValidTo = state.ValidTo
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes a certificate.
func (r *certificateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data certificateModel
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

	err := r.client.DeleteCertificate(ctx, data.Name.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting certificate", err.Error())
		return
	}
}

func (r *certificateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	name := req.ID
	cert, err := r.client.GetCertificate(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError("Error importing certificate", err.Error())
		return
	}

	var data certificateModel
	data.Timeouts = nullTimeoutsValue()

	mapCertificateToModel(cert, &data)

	// intermediate_certificate is Optional (non-Computed): keep it null when the
	// API reports no intermediate, so a follow-up plan doesn't show bogus drift.
	if cert.IntermediateCertificate == "" {
		data.IntermediateCertificate = types.StringNull()
	}

	// Write-only sensitive fields are not returned by the API — set to empty string after import.
	data.PrivateKey = types.StringValue("")
	data.Passphrase = types.StringValue("")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ---------- helpers ---------------------------------------------------------

// mapCertificateToModel maps a client.Certificate to a certificateModel.
// Does NOT set PrivateKey or Passphrase — write-only fields preserved by the caller.
func mapCertificateToModel(cert *client.Certificate, data *certificateModel) {
	data.ID = types.StringValue(cert.ID)
	data.Name = types.StringValue(cert.Name)
	data.Certificate = types.StringValue(cert.Certificate)
	data.CertificateType = types.StringValue(cert.CertificateType)
	data.CommonName = types.StringValue(cert.CommonName)
	data.Country = types.StringValue(cert.Country)
	data.Email = types.StringValue(cert.Email)
	data.IntermediateCertificate = types.StringValue(cert.IntermediateCertificate)
	data.IssuedBy = types.StringValue(cert.IssuedBy)
	data.IssuedTo = types.StringValue(cert.IssuedTo)
	data.KeyAlgorithm = types.StringValue(cert.KeyAlgorithm)
	data.KeySize = types.Int64Value(int64(cert.KeySize))
	data.Locality = types.StringValue(cert.Locality)
	data.Organization = types.StringValue(cert.Organization)
	data.OrganizationalUnit = types.StringValue(cert.OrganizationalUnit)
	data.State = types.StringValue(cert.State)
	data.Status = types.StringValue(cert.Status)
	data.ValidFrom = types.Int64Value(cert.ValidFrom)
	data.ValidTo = types.Int64Value(cert.ValidTo)

	if len(cert.SubjectAlternativeNames) == 0 {
		data.SubjectAlternativeNames = types.ListValueMust(types.StringType, []attr.Value{})
	} else {
		elems := make([]attr.Value, len(cert.SubjectAlternativeNames))
		for i, san := range cert.SubjectAlternativeNames {
			elems[i] = types.StringValue(san)
		}
		data.SubjectAlternativeNames = types.ListValueMust(types.StringType, elems)
	}
}
