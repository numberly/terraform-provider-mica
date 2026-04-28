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

// newTestArrayConnectionKeyResource creates an arrayConnectionKeyResource wired to the given mock server.
func newTestArrayConnectionKeyResource(t *testing.T, ms *testmock.MockServer) *arrayConnectionKeyResource {
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
	return &arrayConnectionKeyResource{client: c}
}

// arrayConnectionKeyResourceSchema returns the parsed schema for the array connection key resource.
func arrayConnectionKeyResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &arrayConnectionKeyResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

// buildArrayConnectionKeyType returns the tftypes.Object for the array connection key resource schema.
func buildArrayConnectionKeyType() tftypes.Object {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
	}}
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":             tftypes.String,
		"connection_key": tftypes.String,
		"created":        tftypes.Number,
		"expires":        tftypes.Number,
		"timeouts":       timeoutsType,
	}}
}

// nullArrayConnectionKeyConfig returns a base config map with all attributes null.
func nullArrayConnectionKeyConfig() map[string]tftypes.Value {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
	}}
	return map[string]tftypes.Value{
		"id":             tftypes.NewValue(tftypes.String, nil),
		"connection_key": tftypes.NewValue(tftypes.String, nil),
		"created":        tftypes.NewValue(tftypes.Number, nil),
		"expires":        tftypes.NewValue(tftypes.Number, nil),
		"timeouts":       tftypes.NewValue(timeoutsType, nil),
	}
}

// arrayConnectionKeyPlan returns a tfsdk.Plan with all-null values (no required fields).
func arrayConnectionKeyPlan(t *testing.T) tfsdk.Plan {
	t.Helper()
	s := arrayConnectionKeyResourceSchema(t).Schema
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildArrayConnectionKeyType(), nullArrayConnectionKeyConfig()),
		Schema: s,
	}
}

// ---- tests ------------------------------------------------------------------

// TestUnit_ArrayConnectionKeyResource_Lifecycle: Create (POST) → Read (GET) → Delete (no-op).
func TestUnit_ArrayConnectionKeyResource_Lifecycle(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterArrayConnectionKeyHandlers(ms.Mux)

	r := newTestArrayConnectionKeyResource(t, ms)
	s := arrayConnectionKeyResourceSchema(t).Schema

	// Step 1: Create — triggers POST.
	plan := arrayConnectionKeyPlan(t)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildArrayConnectionKeyType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create returned error: %s", createResp.Diagnostics)
	}

	var afterCreate arrayConnectionKeyModel
	if diags := createResp.State.Get(context.Background(), &afterCreate); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}
	if afterCreate.ConnectionKey.IsNull() || afterCreate.ConnectionKey.ValueString() == "" {
		t.Error("expected non-empty connection_key after Create")
	}
	if afterCreate.ID.ValueString() != afterCreate.ConnectionKey.ValueString() {
		t.Errorf("expected id == connection_key, got id=%s key=%s",
			afterCreate.ID.ValueString(), afterCreate.ConnectionKey.ValueString())
	}
	if afterCreate.Created.ValueInt64() == 0 {
		t.Error("expected non-zero created timestamp after Create")
	}
	if afterCreate.Expires.ValueInt64() == 0 {
		t.Error("expected non-zero expires timestamp after Create")
	}

	// Step 2: Read — triggers GET, verifies idempotence.
	readResp := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}

	var afterRead arrayConnectionKeyModel
	if diags := readResp.State.Get(context.Background(), &afterRead); diags.HasError() {
		t.Fatalf("Get read state: %s", diags)
	}
	if afterRead.ConnectionKey.ValueString() != afterCreate.ConnectionKey.ValueString() {
		t.Errorf("Read changed connection_key: create=%s read=%s",
			afterCreate.ConnectionKey.ValueString(), afterRead.ConnectionKey.ValueString())
	}

	// Verify connection_key is marked Sensitive in schema.
	schemaResp := arrayConnectionKeyResourceSchema(t)
	attr, ok := schemaResp.Schema.Attributes["connection_key"]
	if !ok {
		t.Fatal("connection_key attribute not found in schema")
	}
	if !attr.IsSensitive() {
		t.Error("connection_key should be Sensitive in schema")
	}

	// Step 3: Delete — no-op (should not return error).
	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: readResp.State}, deleteResp)
	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete returned error (expected no-op): %s", deleteResp.Diagnostics)
	}
}

// TestUnit_ArrayConnectionKeyResource_ReadPreservesState: Read is a no-op that preserves state
// (connection keys are ephemeral — API may not return them after consumption/expiry).
func TestUnit_ArrayConnectionKeyResource_ReadPreservesState(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterArrayConnectionKeyHandlers(ms.Mux)

	r := newTestArrayConnectionKeyResource(t, ms)
	s := arrayConnectionKeyResourceSchema(t).Schema

	// Create.
	plan := arrayConnectionKeyPlan(t)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildArrayConnectionKeyType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Read should preserve state (no-op).
	readResp := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read: %s", readResp.Diagnostics)
	}

	var afterRead arrayConnectionKeyModel
	if diags := readResp.State.Get(context.Background(), &afterRead); diags.HasError() {
		t.Fatalf("Get state after Read: %s", diags)
	}

	if afterRead.ConnectionKey.IsNull() || afterRead.ConnectionKey.ValueString() == "" {
		t.Error("expected connection_key preserved in state after Read, got empty/null")
	}
	if afterRead.Created.IsNull() {
		t.Error("expected created preserved in state after Read")
	}
}

// TestUnit_ArrayConnectionKeyResource_DeleteNoOp: Delete does not call any mock endpoint.
func TestUnit_ArrayConnectionKeyResource_DeleteNoOp(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterArrayConnectionKeyHandlers(ms.Mux)

	// Seed an initial key so we have something in state.
	store.Seed(&client.ArrayConnectionKey{
		ConnectionKey: "existing-key",
		Created:       1000000000000,
		Expires:       1000003600000,
	})

	r := newTestArrayConnectionKeyResource(t, ms)
	s := arrayConnectionKeyResourceSchema(t).Schema

	// Build state with the seeded key.
	stateValues := nullArrayConnectionKeyConfig()
	stateValues["id"] = tftypes.NewValue(tftypes.String, "existing-key")
	stateValues["connection_key"] = tftypes.NewValue(tftypes.String, "existing-key")
	stateValues["created"] = tftypes.NewValue(tftypes.Number, int64(1000000000000))
	stateValues["expires"] = tftypes.NewValue(tftypes.Number, int64(1000003600000))
	state := tfsdk.State{
		Raw:    tftypes.NewValue(buildArrayConnectionKeyType(), stateValues),
		Schema: s,
	}

	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: state}, deleteResp)
	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete returned unexpected error: %s", deleteResp.Diagnostics)
	}

	// The key must still be present in the mock (Delete didn't call the API).
	c, err := client.NewClient(context.Background(), client.Config{
		Endpoint:           ms.URL(),
		APIToken:           "test-token",
		InsecureSkipVerify: true,
		MaxRetries:         1,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	got, err := c.GetArrayConnectionKey(context.Background())
	if err != nil {
		t.Fatalf("GetArrayConnectionKey after Delete: %v", err)
	}
	if got.ConnectionKey != "existing-key" {
		t.Errorf("expected key to still exist after no-op Delete, got %q", got.ConnectionKey)
	}
}
