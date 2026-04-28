package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

var _ datasource.DataSource = &tlsPolicyDataSource{}
var _ datasource.DataSourceWithConfigure = &tlsPolicyDataSource{}

// tlsPolicyDataSource implements the flashblade_tls_policy data source.
type tlsPolicyDataSource struct {
	client *client.FlashBladeClient
}

func NewTlsPolicyDataSource() datasource.DataSource {
	return &tlsPolicyDataSource{}
}

// ---------- model structs ----------------------------------------------------

// tlsPolicyDataSourceModel is the top-level model for the flashblade_tls_policy data source.
type tlsPolicyDataSourceModel struct {
	ID                               types.String `tfsdk:"id"`
	Name                             types.String `tfsdk:"name"`
	ApplianceCertificate             types.String `tfsdk:"appliance_certificate"`
	ClientCertificatesRequired       types.Bool   `tfsdk:"client_certificates_required"`
	DisabledTlsCiphers               types.List   `tfsdk:"disabled_tls_ciphers"`
	Enabled                          types.Bool   `tfsdk:"enabled"`
	EnabledTlsCiphers                types.List   `tfsdk:"enabled_tls_ciphers"`
	IsLocal                          types.Bool   `tfsdk:"is_local"`
	MinTlsVersion                    types.String `tfsdk:"min_tls_version"`
	PolicyType                       types.String `tfsdk:"policy_type"`
	TrustedClientCertificateAuthority types.String `tfsdk:"trusted_client_certificate_authority"`
	VerifyClientCertificateTrust     types.Bool   `tfsdk:"verify_client_certificate_trust"`
}

// ---------- data source interface methods -----------------------------------

func (d *tlsPolicyDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "flashblade_tls_policy"
}

// Schema defines the data source schema.
func (d *tlsPolicyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing FlashBlade TLS policy by name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the TLS policy.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the TLS policy to look up.",
			},
			"appliance_certificate": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the certificate used by the appliance for TLS connections.",
			},
			"client_certificates_required": schema.BoolAttribute{
				Computed:    true,
				Description: "When true, clients must present a certificate for mTLS.",
			},
			"disabled_tls_ciphers": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of TLS cipher suites that are disabled.",
			},
			"enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the TLS policy is enabled.",
			},
			"enabled_tls_ciphers": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of explicitly enabled TLS cipher suites.",
			},
			"is_local": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether this TLS policy is local to the array.",
			},
			"min_tls_version": schema.StringAttribute{
				Computed:    true,
				Description: "The minimum TLS version required (e.g. TLSv1.2, TLSv1.3).",
			},
			"policy_type": schema.StringAttribute{
				Computed:    true,
				Description: "The type of the TLS policy.",
			},
			"trusted_client_certificate_authority": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the certificate authority used to verify client certificates for mTLS.",
			},
			"verify_client_certificate_trust": schema.BoolAttribute{
				Computed:    true,
				Description: "When true, client certificates are verified against the trusted CA.",
			},
		},
	}
}

// Configure injects the FlashBladeClient into the data source.
func (d *tlsPolicyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read fetches the TLS policy from the API by name.
func (d *tlsPolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config tlsPolicyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := config.Name.ValueString()
	policy, err := d.client.GetTlsPolicy(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.Diagnostics.AddError("TLS policy not found", fmt.Sprintf("No TLS policy named %q", name))
		} else {
			resp.Diagnostics.AddError("Error reading TLS policy", err.Error())
		}
		return
	}

	config.ID = types.StringValue(policy.ID)
	config.Name = types.StringValue(policy.Name)

	if policy.ApplianceCertificate != nil {
		config.ApplianceCertificate = types.StringValue(policy.ApplianceCertificate.Name)
	} else {
		config.ApplianceCertificate = types.StringNull()
	}

	config.ClientCertificatesRequired = types.BoolValue(policy.ClientCertificatesRequired)

	config.DisabledTlsCiphers = stringsToListValue(policy.DisabledTlsCiphers)
	config.Enabled = types.BoolValue(policy.Enabled)
	config.EnabledTlsCiphers = stringsToListValue(policy.EnabledTlsCiphers)

	config.IsLocal = types.BoolValue(policy.IsLocal)
	config.MinTlsVersion = types.StringValue(policy.MinTlsVersion)
	config.PolicyType = types.StringValue(policy.PolicyType)

	if policy.TrustedClientCertificateAuthority != nil {
		config.TrustedClientCertificateAuthority = types.StringValue(policy.TrustedClientCertificateAuthority.Name)
	} else {
		config.TrustedClientCertificateAuthority = types.StringNull()
	}

	config.VerifyClientCertificateTrust = types.BoolValue(policy.VerifyClientCertificateTrust)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
