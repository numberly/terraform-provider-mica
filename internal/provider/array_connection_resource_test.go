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

// newTestArrayConnectionResource creates an arrayConnectionResource wired to the given mock server.
func newTestArrayConnectionResource(t *testing.T, ms *testmock.MockServer) *arrayConnectionResource {
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
	return &arrayConnectionResource{client: c}
}

// arrayConnectionResourceSchema returns the parsed schema for the array connection resource.
func arrayConnectionResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &arrayConnectionResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

// buildArrayConnectionResourceType returns the tftypes.Object for the array connection resource schema.
func buildArrayConnectionResourceType() tftypes.Object {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	throttleType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"default_limit": tftypes.Number,
		"window_limit":  tftypes.Number,
		"window_start":  tftypes.String,
		"window_end":    tftypes.String,
	}}
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":                    tftypes.String,
		"remote_name":           tftypes.String,
		"management_address":    tftypes.String,
		"connection_key":        tftypes.String,
		"encrypted":             tftypes.Bool,
		"replication_addresses": tftypes.List{ElementType: tftypes.String},
		"throttle":              throttleType,
		"status":                tftypes.String,
		"type":                  tftypes.String,
		"os":                    tftypes.String,
		"version":               tftypes.String,
		"timeouts":              timeoutsType,
	}}
}

// nullArrayConnectionResourceConfig returns a base config map with all attributes null.
func nullArrayConnectionResourceConfig() map[string]tftypes.Value {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	throttleType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"default_limit": tftypes.Number,
		"window_limit":  tftypes.Number,
		"window_start":  tftypes.String,
		"window_end":    tftypes.String,
	}}
	return map[string]tftypes.Value{
		"id":                    tftypes.NewValue(tftypes.String, nil),
		"remote_name":           tftypes.NewValue(tftypes.String, nil),
		"management_address":    tftypes.NewValue(tftypes.String, nil),
		"connection_key":        tftypes.NewValue(tftypes.String, nil),
		"encrypted":             tftypes.NewValue(tftypes.Bool, nil),
		"replication_addresses": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"throttle":              tftypes.NewValue(throttleType, nil),
		"status":                tftypes.NewValue(tftypes.String, nil),
		"type":                  tftypes.NewValue(tftypes.String, nil),
		"os":                    tftypes.NewValue(tftypes.String, nil),
		"version":               tftypes.NewValue(tftypes.String, nil),
		"timeouts":              tftypes.NewValue(timeoutsType, nil),
	}
}

// arrayConnectionPlanWith returns a tfsdk.Plan with the given required field values.
func arrayConnectionPlanWith(t *testing.T, remoteName, mgmtAddr, connKey string) tfsdk.Plan {
	t.Helper()
	s := arrayConnectionResourceSchema(t).Schema
	cfg := nullArrayConnectionResourceConfig()
	cfg["remote_name"] = tftypes.NewValue(tftypes.String, remoteName)
	cfg["management_address"] = tftypes.NewValue(tftypes.String, mgmtAddr)
	cfg["connection_key"] = tftypes.NewValue(tftypes.String, connKey)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildArrayConnectionResourceType(), cfg),
		Schema: s,
	}
}

// ---- tests ------------------------------------------------------------------

// TestUnit_ArrayConnectionResource_Lifecycle: create → verify state → update management_address → verify → destroy.
func TestUnit_ArrayConnectionResource_Lifecycle(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterArrayConnectionHandlers(ms.Mux)

	r := newTestArrayConnectionResource(t, ms)
	s := arrayConnectionResourceSchema(t).Schema

	// Step 1: Create.
	plan := arrayConnectionPlanWith(t, "remote-fb", "10.0.0.1", "secret-key")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildArrayConnectionResourceType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create returned error: %s", createResp.Diagnostics)
	}

	var afterCreate arrayConnectionModel
	if diags := createResp.State.Get(context.Background(), &afterCreate); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}
	if afterCreate.ID.IsNull() || afterCreate.ID.ValueString() == "" {
		t.Error("expected non-empty id after Create")
	}
	if afterCreate.RemoteName.ValueString() != "remote-fb" {
		t.Errorf("expected remote_name=remote-fb, got %s", afterCreate.RemoteName.ValueString())
	}
	if afterCreate.ManagementAddress.ValueString() != "10.0.0.1" {
		t.Errorf("expected management_address=10.0.0.1, got %s", afterCreate.ManagementAddress.ValueString())
	}
	if afterCreate.Status.ValueString() != "connected" {
		t.Errorf("expected status=connected, got %s", afterCreate.Status.ValueString())
	}

	// Step 2: Read (idempotence check).
	readResp := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}

	var afterRead arrayConnectionModel
	if diags := readResp.State.Get(context.Background(), &afterRead); diags.HasError() {
		t.Fatalf("Get read state: %s", diags)
	}
	if afterRead.ManagementAddress.ValueString() != afterCreate.ManagementAddress.ValueString() {
		t.Errorf("address drift on Read: create=%s read=%s", afterCreate.ManagementAddress.ValueString(), afterRead.ManagementAddress.ValueString())
	}

	// Step 3: Update management_address.
	updatePlan := arrayConnectionPlanWith(t, "remote-fb", "10.0.0.99", "secret-key")
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildArrayConnectionResourceType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlan,
		State: readResp.State,
	}, updateResp)
	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update returned error: %s", updateResp.Diagnostics)
	}

	var afterUpdate arrayConnectionModel
	if diags := updateResp.State.Get(context.Background(), &afterUpdate); diags.HasError() {
		t.Fatalf("Get update state: %s", diags)
	}
	if afterUpdate.ManagementAddress.ValueString() != "10.0.0.99" {
		t.Errorf("expected management_address=10.0.0.99 after update, got %s", afterUpdate.ManagementAddress.ValueString())
	}

	// Step 4: Delete.
	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: updateResp.State}, deleteResp)
	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete returned error: %s", deleteResp.Diagnostics)
	}

	// Verify gone.
	_, err := r.client.GetArrayConnection(context.Background(), "remote-fb")
	if err == nil || !client.IsNotFound(err) {
		t.Errorf("expected array connection to be deleted, got: %v", err)
	}
}

// TestUnit_ArrayConnectionResource_Import: create → import by remote_name → check connection_key is empty.
func TestUnit_ArrayConnectionResource_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterArrayConnectionHandlers(ms.Mux)

	r := newTestArrayConnectionResource(t, ms)
	s := arrayConnectionResourceSchema(t).Schema

	// Create first.
	plan := arrayConnectionPlanWith(t, "remote-import", "192.168.2.20", "some-key")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildArrayConnectionResourceType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Import by remote_name.
	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildArrayConnectionResourceType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "remote-import"}, importResp)
	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState returned error: %s", importResp.Diagnostics)
	}

	var model arrayConnectionModel
	if diags := importResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get import state: %s", diags)
	}

	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected non-empty id after import")
	}
	if model.RemoteName.ValueString() != "remote-import" {
		t.Errorf("expected remote_name=remote-import, got %s", model.RemoteName.ValueString())
	}
	if model.ManagementAddress.ValueString() != "192.168.2.20" {
		t.Errorf("expected management_address=192.168.2.20, got %s", model.ManagementAddress.ValueString())
	}
	// connection_key must be empty string after import (write-only, not returned by API).
	if model.ConnectionKey.ValueString() != "" {
		t.Errorf("expected connection_key='' after import, got %q", model.ConnectionKey.ValueString())
	}
}

// TestUnit_ArrayConnectionResource_DriftDetection: create → Seed modified conn → Read → verify state updated.
func TestUnit_ArrayConnectionResource_DriftDetection(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterArrayConnectionHandlers(ms.Mux)

	r := newTestArrayConnectionResource(t, ms)
	s := arrayConnectionResourceSchema(t).Schema

	// Create.
	plan := arrayConnectionPlanWith(t, "remote-drift", "192.168.3.30", "drift-key")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildArrayConnectionResourceType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	var afterCreate arrayConnectionModel
	if diags := createResp.State.Get(context.Background(), &afterCreate); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}

	// Simulate drift: modify management_address in the mock store directly.
	store.Seed(&client.ArrayConnection{
		ID:                afterCreate.ID.ValueString(),
		Remote:            client.NamedReference{Name: "remote-drift"},
		ManagementAddress: "10.99.99.99", // changed outside Terraform
		Status:            "connected",
		Type:              "async-replication",
	})

	// Read to detect drift.
	readResp := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read drift detection: %s", readResp.Diagnostics)
	}

	var afterDrift arrayConnectionModel
	if diags := readResp.State.Get(context.Background(), &afterDrift); diags.HasError() {
		t.Fatalf("Get drift state: %s", diags)
	}

	// State should reflect the new API values.
	if afterDrift.ManagementAddress.ValueString() != "10.99.99.99" {
		t.Errorf("expected state to reflect drifted management_address=10.99.99.99, got %s", afterDrift.ManagementAddress.ValueString())
	}
}

// TestUnit_ArrayConnectionResource_ConnectionKeySensitive: verify connection_key is Sensitive
// and preserved from plan in subsequent Read (not overwritten by empty).
func TestUnit_ArrayConnectionResource_ConnectionKeySensitive(t *testing.T) {
	// Verify schema attribute is Sensitive.
	resp := arrayConnectionResourceSchema(t)
	attr, ok := resp.Schema.Attributes["connection_key"]
	if !ok {
		t.Fatal("connection_key attribute not found in schema")
	}
	if !attr.IsSensitive() {
		t.Error("connection_key should be Sensitive")
	}

	// Verify connection_key is preserved from state in Read (not overwritten to empty).
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterArrayConnectionHandlers(ms.Mux)

	r := newTestArrayConnectionResource(t, ms)
	s := resp.Schema

	plan := arrayConnectionPlanWith(t, "remote-key", "10.0.0.5", "my-secret-key")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildArrayConnectionResourceType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Read and verify connection_key is preserved from state.
	readResp := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read: %s", readResp.Diagnostics)
	}

	var model arrayConnectionModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	// connection_key must still be "my-secret-key" — never overwritten by Read.
	if model.ConnectionKey.ValueString() != "my-secret-key" {
		t.Errorf("expected connection_key=my-secret-key after Read, got %q", model.ConnectionKey.ValueString())
	}
}

// TestUnit_ArrayConnectionResource_PassiveAdoption: Create without connection_key adopts an existing connection.
func TestUnit_ArrayConnectionResource_PassiveAdoption(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterArrayConnectionHandlers(ms.Mux)

	// Pre-seed the passive-side connection (auto-created by FlashBlade when remote connects).
	store.Seed(&client.ArrayConnection{
		ID:                   "passive-conn-1",
		Remote:               client.NamedReference{Name: "remote-par5"},
		ManagementAddress:    "10.5.0.1",
		Encrypted:            true,
		Status:               "connected",
		Type:                 "async-replication",
		OS:                   "Purity//FB",
		Version:              "4.6.7",
		ReplicationAddresses: []string{},
	})

	r := newTestArrayConnectionResource(t, ms)
	s := arrayConnectionResourceSchema(t).Schema

	// Build plan WITHOUT connection_key and management_address — passive side.
	cfg := nullArrayConnectionResourceConfig()
	cfg["remote_name"] = tftypes.NewValue(tftypes.String, "remote-par5")
	cfg["encrypted"] = tftypes.NewValue(tftypes.Bool, true)
	cfg["replication_addresses"] = tftypes.NewValue(
		tftypes.List{ElementType: tftypes.String},
		[]tftypes.Value{tftypes.NewValue(tftypes.String, "10.6.21.22")},
	)
	plan := tfsdk.Plan{
		Raw:    tftypes.NewValue(buildArrayConnectionResourceType(), cfg),
		Schema: s,
	}

	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildArrayConnectionResourceType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Passive Create returned error: %s", createResp.Diagnostics)
	}

	var model arrayConnectionModel
	if diags := createResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.ValueString() != "passive-conn-1" {
		t.Errorf("expected id=passive-conn-1, got %s", model.ID.ValueString())
	}
	if model.RemoteName.ValueString() != "remote-par5" {
		t.Errorf("expected remote_name=remote-par5, got %s", model.RemoteName.ValueString())
	}
	if model.ManagementAddress.ValueString() != "10.5.0.1" {
		t.Errorf("expected management_address=10.5.0.1, got %s", model.ManagementAddress.ValueString())
	}
	if model.Status.ValueString() != "connected" {
		t.Errorf("expected status=connected, got %s", model.Status.ValueString())
	}
}
