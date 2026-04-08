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

// newTestCertificateGroupResource creates a certificateGroupResource wired to the given mock server.
func newTestCertificateGroupResource(t *testing.T, ms *testmock.MockServer) *certificateGroupResource {
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
	return &certificateGroupResource{client: c}
}

// certificateGroupResourceSchema returns the parsed schema for the certificate group resource.
func certificateGroupResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &certificateGroupResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

// buildCertificateGroupType returns the tftypes.Object for the certificate group resource schema.
func buildCertificateGroupType() tftypes.Object {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"delete": tftypes.String,
	}}
	realmsType := tftypes.List{ElementType: tftypes.String}
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":       tftypes.String,
		"name":     tftypes.String,
		"realms":   realmsType,
		"timeouts": timeoutsType,
	}}
}

// nullCertificateGroupConfig returns a base config map with all attributes null.
func nullCertificateGroupConfig() map[string]tftypes.Value {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"delete": tftypes.String,
	}}
	realmsType := tftypes.List{ElementType: tftypes.String}
	return map[string]tftypes.Value{
		"id":       tftypes.NewValue(tftypes.String, nil),
		"name":     tftypes.NewValue(tftypes.String, nil),
		"realms":   tftypes.NewValue(realmsType, nil),
		"timeouts": tftypes.NewValue(timeoutsType, nil),
	}
}

// certificateGroupPlanWith returns a tfsdk.Plan with the given name.
func certificateGroupPlanWith(t *testing.T, name string) tfsdk.Plan {
	t.Helper()
	s := certificateGroupResourceSchema(t).Schema
	cfg := nullCertificateGroupConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, name)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildCertificateGroupType(), cfg),
		Schema: s,
	}
}

// ---- tests ------------------------------------------------------------------

// TestUnit_CertificateGroupResource_Lifecycle: create → read → delete cycle.
func TestUnit_CertificateGroupResource_Lifecycle(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterCertificateGroupHandlers(ms.Mux)

	r := newTestCertificateGroupResource(t, ms)
	s := certificateGroupResourceSchema(t).Schema

	// Step 1: Create.
	plan := certificateGroupPlanWith(t, "test-group")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildCertificateGroupType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create returned error: %s", createResp.Diagnostics)
	}

	var afterCreate certificateGroupModel
	if diags := createResp.State.Get(context.Background(), &afterCreate); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}
	if afterCreate.ID.IsNull() || afterCreate.ID.ValueString() == "" {
		t.Error("expected non-empty id after Create")
	}
	if afterCreate.Name.ValueString() != "test-group" {
		t.Errorf("expected name=test-group, got %s", afterCreate.Name.ValueString())
	}

	// Step 2: Read (idempotence check).
	readResp := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}

	var afterRead certificateGroupModel
	if diags := readResp.State.Get(context.Background(), &afterRead); diags.HasError() {
		t.Fatalf("Get read state: %s", diags)
	}
	if afterRead.Name.ValueString() != afterCreate.Name.ValueString() {
		t.Errorf("name drift on Read: create=%s read=%s", afterCreate.Name.ValueString(), afterRead.Name.ValueString())
	}

	// Step 3: Delete.
	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: readResp.State}, deleteResp)
	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete returned error: %s", deleteResp.Diagnostics)
	}

	// Verify gone.
	_, err := r.client.GetCertificateGroup(context.Background(), "test-group")
	if err == nil || !client.IsNotFound(err) {
		t.Errorf("expected certificate group to be deleted, got: %v", err)
	}
}

// TestUnit_CertificateGroupResource_Import: seed group → import by name → verify all fields populated.
func TestUnit_CertificateGroupResource_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	rawStore := handlers.RegisterCertificateGroupHandlers(ms.Mux)
	store := handlers.NewCertificateGroupStoreFacade(rawStore)

	r := newTestCertificateGroupResource(t, ms)
	s := certificateGroupResourceSchema(t).Schema

	// Seed a certificate group.
	store.Seed(&client.CertificateGroup{
		ID:     "cg-import-1",
		Name:   "import-group",
		Realms: []string{},
	})

	// Import by name.
	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildCertificateGroupType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "import-group"}, importResp)
	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState returned error: %s", importResp.Diagnostics)
	}

	var model certificateGroupModel
	if diags := importResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get import state: %s", diags)
	}

	if model.ID.ValueString() != "cg-import-1" {
		t.Errorf("expected id=cg-import-1, got %s", model.ID.ValueString())
	}
	if model.Name.ValueString() != "import-group" {
		t.Errorf("expected name=import-group, got %s", model.Name.ValueString())
	}
	// Realms should be an empty list (not null).
	if model.Realms.IsNull() {
		t.Error("expected realms to be empty list, not null")
	}
	if len(model.Realms.Elements()) != 0 {
		t.Errorf("expected 0 realms, got %d", len(model.Realms.Elements()))
	}
	// Timeouts should be null (CRD inline null timeouts).
	if !model.Timeouts.IsNull() {
		t.Error("expected timeouts to be null after import")
	}
}

// TestUnit_CertificateGroupResource_DriftDetection: create → modify mock store → Read → verify updated state.
func TestUnit_CertificateGroupResource_DriftDetection(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	rawStore := handlers.RegisterCertificateGroupHandlers(ms.Mux)
	store := handlers.NewCertificateGroupStoreFacade(rawStore)

	r := newTestCertificateGroupResource(t, ms)
	s := certificateGroupResourceSchema(t).Schema

	// Create a certificate group.
	plan := certificateGroupPlanWith(t, "drift-group")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildCertificateGroupType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Modify mock store directly — simulate out-of-band change (add realms).
	store.Seed(&client.CertificateGroup{
		ID:     "certgroup-1",
		Name:   "drift-group",
		Realms: []string{"management", "replication"},
	})

	// Read should pick up the drift.
	readResp := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read: %s", readResp.Diagnostics)
	}

	var afterRead certificateGroupModel
	if diags := readResp.State.Get(context.Background(), &afterRead); diags.HasError() {
		t.Fatalf("Get read state: %s", diags)
	}

	// State should reflect updated realms from API.
	if len(afterRead.Realms.Elements()) != 2 {
		t.Errorf("expected 2 realms after drift Read, got %d", len(afterRead.Realms.Elements()))
	}
}
