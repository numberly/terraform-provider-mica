package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/numberly/opentofu-provider-flashblade/internal/client"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock/handlers"
)

// ---- helpers ----------------------------------------------------------------

// newTestTlsPolicyDataSource creates a tlsPolicyDataSource wired to the given mock server.
func newTestTlsPolicyDataSource(t *testing.T, ms *testmock.MockServer) *tlsPolicyDataSource {
	t.Helper()
	c, err := client.NewClient(context.Background(), client.Config{
		Endpoint:           ms.URL(),
		APIToken:           "test-token",
		InsecureSkipVerify: true,
		MaxRetries:         1,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return &tlsPolicyDataSource{client: c}
}

// tlsPolicyDSSchema returns the parsed schema for the TLS policy data source.
func tlsPolicyDSSchema(t *testing.T) datasource.SchemaResponse {
	t.Helper()
	d := &tlsPolicyDataSource{}
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	return resp
}

// buildTlsPolicyDSType returns the tftypes.Object for the TLS policy data source schema.
func buildTlsPolicyDSType() tftypes.Object {
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":                                  tftypes.String,
		"name":                                tftypes.String,
		"appliance_certificate":               tftypes.String,
		"client_certificates_required":        tftypes.Bool,
		"disabled_tls_ciphers":                tftypes.List{ElementType: tftypes.String},
		"enabled":                             tftypes.Bool,
		"enabled_tls_ciphers":                 tftypes.List{ElementType: tftypes.String},
		"is_local":                            tftypes.Bool,
		"min_tls_version":                     tftypes.String,
		"policy_type":                         tftypes.String,
		"trusted_client_certificate_authority": tftypes.String,
		"verify_client_certificate_trust":     tftypes.Bool,
	}}
}

// nullTlsPolicyDSConfig returns a base config map with all attributes null.
func nullTlsPolicyDSConfig() map[string]tftypes.Value {
	cipherListType := tftypes.List{ElementType: tftypes.String}
	return map[string]tftypes.Value{
		"id":                                  tftypes.NewValue(tftypes.String, nil),
		"name":                                tftypes.NewValue(tftypes.String, nil),
		"appliance_certificate":               tftypes.NewValue(tftypes.String, nil),
		"client_certificates_required":        tftypes.NewValue(tftypes.Bool, nil),
		"disabled_tls_ciphers":                tftypes.NewValue(cipherListType, nil),
		"enabled":                             tftypes.NewValue(tftypes.Bool, nil),
		"enabled_tls_ciphers":                 tftypes.NewValue(cipherListType, nil),
		"is_local":                            tftypes.NewValue(tftypes.Bool, nil),
		"min_tls_version":                     tftypes.NewValue(tftypes.String, nil),
		"policy_type":                         tftypes.NewValue(tftypes.String, nil),
		"trusted_client_certificate_authority": tftypes.NewValue(tftypes.String, nil),
		"verify_client_certificate_trust":     tftypes.NewValue(tftypes.Bool, nil),
	}
}

// ---- tests ------------------------------------------------------------------

// TestUnit_TlsPolicyDataSource_Basic: seed policy → read by name → verify all computed fields.
func TestUnit_TlsPolicyDataSource_Basic(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterTlsPolicyHandlers(ms.Mux)

	// Seed a TLS policy in the mock store.
	store.Seed(&client.TlsPolicy{
		ID:                               "tls-ds-001",
		Name:                             "ds-policy",
		MinTlsVersion:                    "TLSv1.2",
		Enabled:                          true,
		IsLocal:                          true,
		PolicyType:                       "global",
		ClientCertificatesRequired:       false,
		VerifyClientCertificateTrust:     false,
		ApplianceCertificate:             &client.NamedReference{Name: "ds-cert"},
		TrustedClientCertificateAuthority: &client.NamedReference{Name: "trusted-ca"},
		DisabledTlsCiphers:               []string{"TLS_RSA_WITH_RC4_128_MD5"},
		EnabledTlsCiphers:                []string{},
	})

	d := newTestTlsPolicyDataSource(t, ms)
	s := tlsPolicyDSSchema(t).Schema

	cfg := nullTlsPolicyDSConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "ds-policy")

	config := tfsdk.Config{
		Raw:    tftypes.NewValue(buildTlsPolicyDSType(), cfg),
		Schema: s,
	}

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildTlsPolicyDSType(), nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{Config: config}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}

	var model tlsPolicyDataSourceModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.ValueString() != "tls-ds-001" {
		t.Errorf("expected id=tls-ds-001, got %s", model.ID.ValueString())
	}
	if model.Name.ValueString() != "ds-policy" {
		t.Errorf("expected name=ds-policy, got %s", model.Name.ValueString())
	}
	if model.MinTlsVersion.ValueString() != "TLSv1.2" {
		t.Errorf("expected min_tls_version=TLSv1.2, got %s", model.MinTlsVersion.ValueString())
	}
	if model.Enabled.ValueBool() != true {
		t.Error("expected enabled=true")
	}
	if model.IsLocal.ValueBool() != true {
		t.Error("expected is_local=true")
	}
	if model.PolicyType.ValueString() != "global" {
		t.Errorf("expected policy_type=global, got %s", model.PolicyType.ValueString())
	}
	if model.ApplianceCertificate.ValueString() != "ds-cert" {
		t.Errorf("expected appliance_certificate=ds-cert, got %s", model.ApplianceCertificate.ValueString())
	}
	if model.TrustedClientCertificateAuthority.ValueString() != "trusted-ca" {
		t.Errorf("expected trusted_client_certificate_authority=trusted-ca, got %s",
			model.TrustedClientCertificateAuthority.ValueString())
	}
}
