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

// newTestCertificateDataSource creates a certificateDataSource wired to the given mock server.
func newTestCertificateDataSource(t *testing.T, ms *testmock.MockServer) *certificateDataSource {
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
	return &certificateDataSource{client: c}
}

// certificateDSSchema returns the parsed schema for the certificate data source.
func certificateDSSchema(t *testing.T) datasource.SchemaResponse {
	t.Helper()
	d := &certificateDataSource{}
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	return resp
}

// buildCertificateDSType returns the tftypes.Object for the certificate data source schema.
func buildCertificateDSType() tftypes.Object {
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":                       tftypes.String,
		"name":                     tftypes.String,
		"certificate":              tftypes.String,
		"certificate_type":         tftypes.String,
		"common_name":              tftypes.String,
		"country":                  tftypes.String,
		"email":                    tftypes.String,
		"intermediate_certificate": tftypes.String,
		"issued_by":                tftypes.String,
		"issued_to":                tftypes.String,
		"key_algorithm":            tftypes.String,
		"key_size":                 tftypes.Number,
		"locality":                 tftypes.String,
		"organization":             tftypes.String,
		"organizational_unit":      tftypes.String,
		"state":                    tftypes.String,
		"status":                   tftypes.String,
		"subject_alternative_names": tftypes.List{ElementType: tftypes.String},
		"valid_from":               tftypes.Number,
		"valid_to":                 tftypes.Number,
	}}
}

// nullCertificateDSConfig returns a base config map with all attributes null.
func nullCertificateDSConfig() map[string]tftypes.Value {
	sanType := tftypes.List{ElementType: tftypes.String}
	return map[string]tftypes.Value{
		"id":                       tftypes.NewValue(tftypes.String, nil),
		"name":                     tftypes.NewValue(tftypes.String, nil),
		"certificate":              tftypes.NewValue(tftypes.String, nil),
		"certificate_type":         tftypes.NewValue(tftypes.String, nil),
		"common_name":              tftypes.NewValue(tftypes.String, nil),
		"country":                  tftypes.NewValue(tftypes.String, nil),
		"email":                    tftypes.NewValue(tftypes.String, nil),
		"intermediate_certificate": tftypes.NewValue(tftypes.String, nil),
		"issued_by":                tftypes.NewValue(tftypes.String, nil),
		"issued_to":                tftypes.NewValue(tftypes.String, nil),
		"key_algorithm":            tftypes.NewValue(tftypes.String, nil),
		"key_size":                 tftypes.NewValue(tftypes.Number, nil),
		"locality":                 tftypes.NewValue(tftypes.String, nil),
		"organization":             tftypes.NewValue(tftypes.String, nil),
		"organizational_unit":      tftypes.NewValue(tftypes.String, nil),
		"state":                    tftypes.NewValue(tftypes.String, nil),
		"status":                   tftypes.NewValue(tftypes.String, nil),
		"subject_alternative_names": tftypes.NewValue(sanType, nil),
		"valid_from":               tftypes.NewValue(tftypes.Number, nil),
		"valid_to":                 tftypes.NewValue(tftypes.Number, nil),
	}
}

// TestUnit_CertificateDataSource_Basic seeds a certificate and reads it via the data source.
func TestUnit_CertificateDataSource_Basic(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterCertificateHandlers(ms.Mux)

	// Seed a certificate for the data source to read.
	store.Seed(&client.Certificate{
		ID:                      "cert-ds-001",
		Name:                    "ds-cert",
		Certificate:             "-----BEGIN CERTIFICATE-----\nMIIBds\n-----END CERTIFICATE-----",
		CertificateType:         "array",
		CommonName:              "flashblade.example.com",
		Country:                 "US",
		Email:                   "admin@example.com",
		IntermediateCertificate: "",
		IssuedBy:                "CN=Test CA",
		IssuedTo:                "CN=flashblade.example.com",
		KeyAlgorithm:            "RSA",
		KeySize:                 4096,
		Locality:                "San Francisco",
		Organization:            "Example Corp",
		OrganizationalUnit:      "IT",
		State:                   "CA",
		Status:                  "imported",
		SubjectAlternativeNames: []string{"flashblade.example.com", "fb.example.com"},
		ValidFrom:               1700000000000,
		ValidTo:                 1800000000000,
	})

	d := newTestCertificateDataSource(t, ms)
	s := certificateDSSchema(t).Schema
	objType := buildCertificateDSType()

	cfg := nullCertificateDSConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "ds-cert")

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(objType, cfg), Schema: s},
	}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}

	var model certificateDataSourceModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.ValueString() != "cert-ds-001" {
		t.Errorf("expected id=cert-ds-001, got %s", model.ID.ValueString())
	}
	if model.Name.ValueString() != "ds-cert" {
		t.Errorf("expected name=ds-cert, got %s", model.Name.ValueString())
	}
	if model.CertificateType.ValueString() != "array" {
		t.Errorf("expected certificate_type=array, got %s", model.CertificateType.ValueString())
	}
	if model.CommonName.ValueString() != "flashblade.example.com" {
		t.Errorf("expected common_name=flashblade.example.com, got %s", model.CommonName.ValueString())
	}
	if model.IssuedBy.ValueString() != "CN=Test CA" {
		t.Errorf("expected issued_by=CN=Test CA, got %s", model.IssuedBy.ValueString())
	}
	if model.KeyAlgorithm.ValueString() != "RSA" {
		t.Errorf("expected key_algorithm=RSA, got %s", model.KeyAlgorithm.ValueString())
	}
	if model.KeySize.ValueInt64() != 4096 {
		t.Errorf("expected key_size=4096, got %d", model.KeySize.ValueInt64())
	}
	if model.Status.ValueString() != "imported" {
		t.Errorf("expected status=imported, got %s", model.Status.ValueString())
	}
	if model.ValidFrom.ValueInt64() != 1700000000000 {
		t.Errorf("expected valid_from=1700000000000, got %d", model.ValidFrom.ValueInt64())
	}
	if model.ValidTo.ValueInt64() != 1800000000000 {
		t.Errorf("expected valid_to=1800000000000, got %d", model.ValidTo.ValueInt64())
	}

	// Verify subject_alternative_names list.
	if model.SubjectAlternativeNames.IsNull() {
		t.Fatal("expected subject_alternative_names to be non-null")
	}
	var sans []string
	if diags := model.SubjectAlternativeNames.ElementsAs(context.Background(), &sans, false); diags.HasError() {
		t.Fatalf("ElementsAs SANs: %s", diags)
	}
	if len(sans) != 2 {
		t.Errorf("expected 2 SANs, got %d: %v", len(sans), sans)
	}
}
