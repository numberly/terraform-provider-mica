package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/numberly/terraform-provider-mica/internal/client"
	"github.com/numberly/terraform-provider-mica/internal/testmock"
	"github.com/numberly/terraform-provider-mica/internal/testmock/handlers"
)

// ---- helpers ----------------------------------------------------------------

// newTestTlsPolicyResource creates a tlsPolicyResource wired to the given mock server.
func newTestTlsPolicyResource(t *testing.T, ms *testmock.MockServer) *tlsPolicyResource {
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
	return &tlsPolicyResource{client: c}
}

// tlsPolicyResourceSchema returns the parsed schema for the TLS policy resource.
func tlsPolicyResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &tlsPolicyResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

// buildTlsPolicyType returns the tftypes.Object for the TLS policy resource schema.
func buildTlsPolicyType() tftypes.Object {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
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
		"timeouts":                            timeoutsType,
	}}
}

// nullTlsPolicyConfig returns a base config map with all attributes null.
func nullTlsPolicyConfig() map[string]tftypes.Value {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
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
		"timeouts":                            tftypes.NewValue(timeoutsType, nil),
	}
}

// tlsPolicyPlanWith returns a tfsdk.Plan with the given field values.
func tlsPolicyPlanWith(t *testing.T, name, minTlsVersion, applianceCert string, enabled bool) tfsdk.Plan {
	t.Helper()
	s := tlsPolicyResourceSchema(t).Schema
	cfg := nullTlsPolicyConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, name)
	if minTlsVersion != "" {
		cfg["min_tls_version"] = tftypes.NewValue(tftypes.String, minTlsVersion)
	}
	if applianceCert != "" {
		cfg["appliance_certificate"] = tftypes.NewValue(tftypes.String, applianceCert)
	}
	cfg["enabled"] = tftypes.NewValue(tftypes.Bool, enabled)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildTlsPolicyType(), cfg),
		Schema: s,
	}
}

// ---- tests ------------------------------------------------------------------

// TestUnit_TlsPolicyResource_Lifecycle: create → read → update min_tls_version → delete.
func TestUnit_TlsPolicyResource_Lifecycle(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterTlsPolicyHandlers(ms.Mux)

	r := newTestTlsPolicyResource(t, ms)
	s := tlsPolicyResourceSchema(t).Schema

	// Step 1: Create.
	plan := tlsPolicyPlanWith(t, "strict-tls", "TLSv1.2", "my-cert", true)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildTlsPolicyType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create returned error: %s", createResp.Diagnostics)
	}

	var afterCreate tlsPolicyModel
	if diags := createResp.State.Get(context.Background(), &afterCreate); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}
	if afterCreate.ID.IsNull() || afterCreate.ID.ValueString() == "" {
		t.Error("expected non-empty id after Create")
	}
	if afterCreate.Name.ValueString() != "strict-tls" {
		t.Errorf("expected name=strict-tls, got %s", afterCreate.Name.ValueString())
	}
	if afterCreate.MinTlsVersion.ValueString() != "TLSv1.2" {
		t.Errorf("expected min_tls_version=TLSv1.2, got %s", afterCreate.MinTlsVersion.ValueString())
	}
	if afterCreate.Enabled.ValueBool() != true {
		t.Error("expected enabled=true after Create")
	}
	if afterCreate.ApplianceCertificate.ValueString() != "my-cert" {
		t.Errorf("expected appliance_certificate=my-cert, got %s", afterCreate.ApplianceCertificate.ValueString())
	}

	// Step 2: Read (idempotence check).
	readResp := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}

	var afterRead tlsPolicyModel
	if diags := readResp.State.Get(context.Background(), &afterRead); diags.HasError() {
		t.Fatalf("Get read state: %s", diags)
	}
	if afterRead.MinTlsVersion.ValueString() != afterCreate.MinTlsVersion.ValueString() {
		t.Errorf("min_tls_version drift on Read: create=%s read=%s",
			afterCreate.MinTlsVersion.ValueString(), afterRead.MinTlsVersion.ValueString())
	}

	// Step 3: Update min_tls_version.
	updatePlan := tlsPolicyPlanWith(t, "strict-tls", "TLSv1.3", "my-cert", true)
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildTlsPolicyType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlan,
		State: readResp.State,
	}, updateResp)
	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update returned error: %s", updateResp.Diagnostics)
	}

	var afterUpdate tlsPolicyModel
	if diags := updateResp.State.Get(context.Background(), &afterUpdate); diags.HasError() {
		t.Fatalf("Get update state: %s", diags)
	}
	if afterUpdate.MinTlsVersion.ValueString() != "TLSv1.3" {
		t.Errorf("expected min_tls_version=TLSv1.3 after Update, got %s", afterUpdate.MinTlsVersion.ValueString())
	}

	// Step 4: Delete.
	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: updateResp.State}, deleteResp)
	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete returned error: %s", deleteResp.Diagnostics)
	}

	// Verify gone.
	_, err := r.client.GetTlsPolicy(context.Background(), "strict-tls")
	if err == nil || !client.IsNotFound(err) {
		t.Errorf("expected TLS policy to be deleted, got: %v", err)
	}
}

// TestUnit_TlsPolicyResource_Import: seed policy in mock → import by name → verify all fields populated.
func TestUnit_TlsPolicyResource_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterTlsPolicyHandlers(ms.Mux)

	r := newTestTlsPolicyResource(t, ms)
	s := tlsPolicyResourceSchema(t).Schema

	// Seed a TLS policy in the mock store.
	store.Seed(&client.TlsPolicy{
		ID:             "tls-import-001",
		Name:           "import-policy",
		MinTlsVersion:  "TLSv1.2",
		Enabled:        true,
		IsLocal:        true,
		PolicyType:     "global",
		ApplianceCertificate: &client.NamedReference{Name: "appliance-cert"},
	})

	// Import by name.
	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildTlsPolicyType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "import-policy"}, importResp)
	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState returned error: %s", importResp.Diagnostics)
	}

	var model tlsPolicyModel
	if diags := importResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get import state: %s", diags)
	}

	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected non-empty id after import")
	}
	if model.Name.ValueString() != "import-policy" {
		t.Errorf("expected name=import-policy, got %s", model.Name.ValueString())
	}
	if model.MinTlsVersion.ValueString() != "TLSv1.2" {
		t.Errorf("expected min_tls_version=TLSv1.2, got %s", model.MinTlsVersion.ValueString())
	}
	if model.Enabled.ValueBool() != true {
		t.Error("expected enabled=true after import")
	}
	if model.ApplianceCertificate.ValueString() != "appliance-cert" {
		t.Errorf("expected appliance_certificate=appliance-cert, got %s", model.ApplianceCertificate.ValueString())
	}
	// Timeouts should be null via nullTimeoutsValue (no plan available during import).
	if !model.Timeouts.IsNull() {
		t.Error("expected timeouts to be null after import (nullTimeoutsValue)")
	}
}

// TestUnit_TlsPolicyResource_DriftDetection: create → modify mock → Read → verify updated state.
func TestUnit_TlsPolicyResource_DriftDetection(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterTlsPolicyHandlers(ms.Mux)

	r := newTestTlsPolicyResource(t, ms)
	s := tlsPolicyResourceSchema(t).Schema

	// Create policy.
	plan := tlsPolicyPlanWith(t, "drift-policy", "TLSv1.2", "", true)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildTlsPolicyType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Modify mock store directly — simulate out-of-band change.
	store.Seed(&client.TlsPolicy{
		ID:            "tls-1",
		Name:          "drift-policy",
		MinTlsVersion: "TLSv1.3",
		Enabled:       false,
		IsLocal:       true,
		PolicyType:    "global",
	})

	// Read should pick up the drift.
	readResp := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read: %s", readResp.Diagnostics)
	}

	var afterRead tlsPolicyModel
	if diags := readResp.State.Get(context.Background(), &afterRead); diags.HasError() {
		t.Fatalf("Get read state: %s", diags)
	}

	// State should now reflect updated API values.
	if afterRead.MinTlsVersion.ValueString() != "TLSv1.3" {
		t.Errorf("expected min_tls_version=TLSv1.3 after drift Read, got %s", afterRead.MinTlsVersion.ValueString())
	}
	if afterRead.Enabled.ValueBool() != false {
		t.Error("expected enabled=false after drift Read")
	}
}
