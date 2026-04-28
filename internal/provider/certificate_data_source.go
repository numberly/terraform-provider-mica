package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

var _ datasource.DataSource = &certificateDataSource{}
var _ datasource.DataSourceWithConfigure = &certificateDataSource{}

// certificateDataSource implements the flashblade_certificate data source.
type certificateDataSource struct {
	client *client.FlashBladeClient
}

func NewCertificateDataSource() datasource.DataSource {
	return &certificateDataSource{}
}

// ---------- model structs ----------------------------------------------------

// certificateDataSourceModel is the top-level model for the flashblade_certificate data source.
// No private_key, passphrase, or timeouts — data sources don't manage write-only or timeout fields.
type certificateDataSourceModel struct {
	ID                      types.String `tfsdk:"id"`
	Name                    types.String `tfsdk:"name"`
	Certificate             types.String `tfsdk:"certificate"`
	CertificateType         types.String `tfsdk:"certificate_type"`
	CommonName              types.String `tfsdk:"common_name"`
	Country                 types.String `tfsdk:"country"`
	Email                   types.String `tfsdk:"email"`
	IntermediateCertificate types.String `tfsdk:"intermediate_certificate"`
	IssuedBy                types.String `tfsdk:"issued_by"`
	IssuedTo                types.String `tfsdk:"issued_to"`
	KeyAlgorithm            types.String `tfsdk:"key_algorithm"`
	KeySize                 types.Int64  `tfsdk:"key_size"`
	Locality                types.String `tfsdk:"locality"`
	Organization            types.String `tfsdk:"organization"`
	OrganizationalUnit      types.String `tfsdk:"organizational_unit"`
	State                   types.String `tfsdk:"state"`
	Status                  types.String `tfsdk:"status"`
	SubjectAlternativeNames types.List   `tfsdk:"subject_alternative_names"`
	ValidFrom               types.Int64  `tfsdk:"valid_from"`
	ValidTo                 types.Int64  `tfsdk:"valid_to"`
}

// ---------- data source interface methods -----------------------------------

func (d *certificateDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "flashblade_certificate"
}

// Schema defines the data source schema.
func (d *certificateDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing FlashBlade certificate by name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the certificate.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the certificate to look up.",
			},
			"certificate": schema.StringAttribute{
				Computed:    true,
				Description: "The PEM-encoded X.509 certificate body.",
			},
			"certificate_type": schema.StringAttribute{
				Computed:    true,
				Description: "The certificate type (e.g. appliance, external).",
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
				Computed:    true,
				Description: "The PEM-encoded intermediate certificate chain.",
			},
			"issued_by": schema.StringAttribute{
				Computed:    true,
				Description: "The issuer of the certificate.",
			},
			"issued_to": schema.StringAttribute{
				Computed:    true,
				Description: "The subject of the certificate.",
			},
			"key_algorithm": schema.StringAttribute{
				Computed:    true,
				Description: "The key algorithm (e.g. RSA, EC).",
			},
			"key_size": schema.Int64Attribute{
				Computed:    true,
				Description: "The key size in bits.",
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
			"state": schema.StringAttribute{
				Computed:    true,
				Description: "The state/province (ST) field extracted from the certificate.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The certificate status (e.g. imported, self-signed).",
			},
			"subject_alternative_names": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "The subject alternative names (SANs) extracted from the certificate.",
			},
			"valid_from": schema.Int64Attribute{
				Computed:    true,
				Description: "The Unix timestamp (milliseconds) from which the certificate is valid.",
			},
			"valid_to": schema.Int64Attribute{
				Computed:    true,
				Description: "The Unix timestamp (milliseconds) until which the certificate is valid.",
			},
		},
	}
}

// Configure injects the FlashBladeClient into the data source.
func (d *certificateDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.client = c
}

// Read fetches certificate data by name and populates state.
func (d *certificateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config certificateDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := config.Name.ValueString()
	cert, err := d.client.GetCertificate(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Certificate not found",
				fmt.Sprintf("No certificate named %q exists on the FlashBlade array.", name),
			)
			return
		}
		resp.Diagnostics.AddError("Error reading certificate", err.Error())
		return
	}

	config.ID = types.StringValue(cert.ID)
	config.Name = types.StringValue(cert.Name)
	config.Certificate = types.StringValue(cert.Certificate)
	config.CertificateType = types.StringValue(cert.CertificateType)
	config.CommonName = types.StringValue(cert.CommonName)
	config.Country = types.StringValue(cert.Country)
	config.Email = types.StringValue(cert.Email)
	config.IntermediateCertificate = types.StringValue(cert.IntermediateCertificate)
	config.IssuedBy = types.StringValue(cert.IssuedBy)
	config.IssuedTo = types.StringValue(cert.IssuedTo)
	config.KeyAlgorithm = types.StringValue(cert.KeyAlgorithm)
	config.KeySize = types.Int64Value(int64(cert.KeySize))
	config.Locality = types.StringValue(cert.Locality)
	config.Organization = types.StringValue(cert.Organization)
	config.OrganizationalUnit = types.StringValue(cert.OrganizationalUnit)
	config.State = types.StringValue(cert.State)
	config.Status = types.StringValue(cert.Status)
	config.ValidFrom = types.Int64Value(cert.ValidFrom)
	config.ValidTo = types.Int64Value(cert.ValidTo)

	if len(cert.SubjectAlternativeNames) == 0 {
		config.SubjectAlternativeNames = types.ListValueMust(types.StringType, []attr.Value{})
	} else {
		elems := make([]attr.Value, len(cert.SubjectAlternativeNames))
		for i, san := range cert.SubjectAlternativeNames {
			elems[i] = types.StringValue(san)
		}
		config.SubjectAlternativeNames = types.ListValueMust(types.StringType, elems)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
