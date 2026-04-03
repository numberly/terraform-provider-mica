package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/numberly/opentofu-provider-flashblade/internal/client"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock/handlers"
)

// ---- helpers ----------------------------------------------------------------

// newTestCertificateResource creates a certificateResource wired to the given mock server.
func newTestCertificateResource(t *testing.T, ms *testmock.MockServer) *certificateResource {
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
	return &certificateResource{client: c}
}

// certificateResourceSchema returns the parsed schema for the certificate resource.
func certificateResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &certificateResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

// buildCertificateType returns the tftypes.Object for the certificate resource schema.
func buildCertificateType() tftypes.Object {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
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
		"passphrase":               tftypes.String,
		"private_key":              tftypes.String,
		"state":                    tftypes.String,
		"status":                   tftypes.String,
		"subject_alternative_names": tftypes.List{ElementType: tftypes.String},
		"valid_from":               tftypes.Number,
		"valid_to":                 tftypes.Number,
		"timeouts":                 timeoutsType,
	}}
}

// nullCertificateConfig returns a base config map with all attributes null.
func nullCertificateConfig() map[string]tftypes.Value {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
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
		"passphrase":               tftypes.NewValue(tftypes.String, nil),
		"private_key":              tftypes.NewValue(tftypes.String, nil),
		"state":                    tftypes.NewValue(tftypes.String, nil),
		"status":                   tftypes.NewValue(tftypes.String, nil),
		"subject_alternative_names": tftypes.NewValue(sanType, nil),
		"valid_from":               tftypes.NewValue(tftypes.Number, nil),
		"valid_to":                 tftypes.NewValue(tftypes.Number, nil),
		"timeouts":                 tftypes.NewValue(timeoutsType, nil),
	}
}

// certificatePlanWith returns a tfsdk.Plan with the given field values.
func certificatePlanWith(t *testing.T, name, certificate, privateKey, certType string) tfsdk.Plan {
	t.Helper()
	s := certificateResourceSchema(t).Schema
	cfg := nullCertificateConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, name)
	cfg["certificate"] = tftypes.NewValue(tftypes.String, certificate)
	if privateKey != "" {
		cfg["private_key"] = tftypes.NewValue(tftypes.String, privateKey)
	}
	if certType != "" {
		cfg["certificate_type"] = tftypes.NewValue(tftypes.String, certType)
	}
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildCertificateType(), cfg),
		Schema: s,
	}
}

// ---- tests ------------------------------------------------------------------

// TestUnit_CertificateResource_Lifecycle: create → read → update (cert renewal) → delete.
func TestUnit_CertificateResource_Lifecycle(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterCertificateHandlers(ms.Mux)

	r := newTestCertificateResource(t, ms)
	s := certificateResourceSchema(t).Schema

	const certPEM = "-----BEGIN CERTIFICATE-----\nMIIBtest1\n-----END CERTIFICATE-----"
	const privKey = "-----BEGIN PRIVATE KEY-----\nMIIEtest1\n-----END PRIVATE KEY-----"

	// Step 1: Create.
	plan := certificatePlanWith(t, "test-cert", certPEM, privKey, "")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildCertificateType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create returned error: %s", createResp.Diagnostics)
	}

	var afterCreate certificateModel
	if diags := createResp.State.Get(context.Background(), &afterCreate); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}
	if afterCreate.ID.IsNull() || afterCreate.ID.ValueString() == "" {
		t.Error("expected non-empty id after Create")
	}
	if afterCreate.Name.ValueString() != "test-cert" {
		t.Errorf("expected name=test-cert, got %s", afterCreate.Name.ValueString())
	}
	if afterCreate.Status.ValueString() != "imported" {
		t.Errorf("expected status=imported, got %s", afterCreate.Status.ValueString())
	}
	if afterCreate.IssuedBy.ValueString() == "" {
		t.Error("expected non-empty issued_by after Create")
	}
	// Verify private_key preserved from plan (not lost after Create).
	if afterCreate.PrivateKey.ValueString() != privKey {
		t.Errorf("expected private_key preserved from plan, got %q", afterCreate.PrivateKey.ValueString())
	}

	// Step 2: Read (idempotence check).
	readResp := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}

	var afterRead certificateModel
	if diags := readResp.State.Get(context.Background(), &afterRead); diags.HasError() {
		t.Fatalf("Get read state: %s", diags)
	}
	if afterRead.Certificate.ValueString() != afterCreate.Certificate.ValueString() {
		t.Errorf("certificate drift on Read: create=%s read=%s", afterCreate.Certificate.ValueString(), afterRead.Certificate.ValueString())
	}
	// Verify private_key preserved through Read.
	if afterRead.PrivateKey.ValueString() != privKey {
		t.Errorf("expected private_key preserved through Read, got %q", afterRead.PrivateKey.ValueString())
	}

	// Step 3: Update (cert renewal with new PEM).
	const newCertPEM = "-----BEGIN CERTIFICATE-----\nMIIBtest2\n-----END CERTIFICATE-----"
	updatePlan := certificatePlanWith(t, "test-cert", newCertPEM, privKey, "")
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildCertificateType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlan,
		State: readResp.State,
	}, updateResp)
	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update returned error: %s", updateResp.Diagnostics)
	}

	var afterUpdate certificateModel
	if diags := updateResp.State.Get(context.Background(), &afterUpdate); diags.HasError() {
		t.Fatalf("Get update state: %s", diags)
	}
	if afterUpdate.Certificate.ValueString() != newCertPEM {
		t.Errorf("expected certificate updated, got %s", afterUpdate.Certificate.ValueString())
	}

	// Step 4: Delete.
	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: updateResp.State}, deleteResp)
	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete returned error: %s", deleteResp.Diagnostics)
	}

	// Verify gone.
	_, err := r.client.GetCertificate(context.Background(), "test-cert")
	if err == nil || !client.IsNotFound(err) {
		t.Errorf("expected certificate to be deleted, got: %v", err)
	}
}

// TestUnit_CertificateResource_Import: seed cert in mock → import by name → verify state.
func TestUnit_CertificateResource_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterCertificateHandlers(ms.Mux)

	r := newTestCertificateResource(t, ms)
	s := certificateResourceSchema(t).Schema

	// Seed a certificate in the mock store.
	store.Seed(&client.Certificate{
		ID:              "cert-import-001",
		Name:            "import-cert",
		Certificate:     "-----BEGIN CERTIFICATE-----\nMIIBseed\n-----END CERTIFICATE-----",
		CertificateType: "appliance",
		CommonName:      "flashblade.example.com",
		IssuedBy:        "CN=Test CA",
		IssuedTo:        "CN=flashblade.example.com",
		KeyAlgorithm:    "RSA",
		KeySize:         2048,
		Status:          "imported",
		ValidFrom:       1700000000000,
		ValidTo:         1800000000000,
	})

	// Import by name.
	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildCertificateType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "import-cert"}, importResp)
	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState returned error: %s", importResp.Diagnostics)
	}

	var model certificateModel
	if diags := importResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get import state: %s", diags)
	}

	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected non-empty id after import")
	}
	if model.Name.ValueString() != "import-cert" {
		t.Errorf("expected name=import-cert, got %s", model.Name.ValueString())
	}
	if model.CertificateType.ValueString() != "appliance" {
		t.Errorf("expected certificate_type=appliance, got %s", model.CertificateType.ValueString())
	}
	if model.CommonName.ValueString() != "flashblade.example.com" {
		t.Errorf("expected common_name=flashblade.example.com, got %s", model.CommonName.ValueString())
	}
	if model.Status.ValueString() != "imported" {
		t.Errorf("expected status=imported, got %s", model.Status.ValueString())
	}
	// Write-only fields must be empty string after import (not the original key).
	if model.PrivateKey.ValueString() != "" {
		t.Errorf("expected private_key=\"\" after import, got %q", model.PrivateKey.ValueString())
	}
	if model.Passphrase.ValueString() != "" {
		t.Errorf("expected passphrase=\"\" after import, got %q", model.Passphrase.ValueString())
	}
	// Timeouts should be set to null via nullTimeoutsValue (no plan available during import).
	if !model.Timeouts.IsNull() {
		t.Error("expected timeouts to be null after import (nullTimeoutsValue)")
	}
}

// TestUnit_CertificateResource_DriftDetection: seed cert → modify in mock → Read → verify updated.
func TestUnit_CertificateResource_DriftDetection(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterCertificateHandlers(ms.Mux)

	r := newTestCertificateResource(t, ms)
	s := certificateResourceSchema(t).Schema

	const certPEM = "-----BEGIN CERTIFICATE-----\nMIIBdrift\n-----END CERTIFICATE-----"

	// Create.
	plan := certificatePlanWith(t, "drift-cert", certPEM, "-----BEGIN PRIVATE KEY-----\nMIIEdrift\n-----END PRIVATE KEY-----", "")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildCertificateType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	var afterCreate certificateModel
	if diags := createResp.State.Get(context.Background(), &afterCreate); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}

	// Simulate drift: modify status in the mock store directly.
	store.Seed(&client.Certificate{
		ID:              afterCreate.ID.ValueString(),
		Name:            "drift-cert",
		Certificate:     certPEM,
		CertificateType: "external", // changed outside Terraform
		IssuedBy:        "CN=Drifted CA",
		IssuedTo:        "CN=drift-cert",
		KeyAlgorithm:    "EC",
		KeySize:         256,
		Status:          "imported",
		ValidFrom:       1700000000000,
		ValidTo:         1800000000000,
	})

	// Read to detect drift.
	readResp := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read drift detection: %s", readResp.Diagnostics)
	}

	var afterDrift certificateModel
	if diags := readResp.State.Get(context.Background(), &afterDrift); diags.HasError() {
		t.Fatalf("Get drift state: %s", diags)
	}

	// State should reflect the new API values.
	if afterDrift.CertificateType.ValueString() != "external" {
		t.Errorf("expected state to reflect drifted certificate_type=external, got %s", afterDrift.CertificateType.ValueString())
	}
	if afterDrift.KeyAlgorithm.ValueString() != "EC" {
		t.Errorf("expected state to reflect drifted key_algorithm=EC, got %s", afterDrift.KeyAlgorithm.ValueString())
	}
	if afterDrift.IssuedBy.ValueString() != "CN=Drifted CA" {
		t.Errorf("expected state to reflect drifted issued_by, got %s", afterDrift.IssuedBy.ValueString())
	}
}
