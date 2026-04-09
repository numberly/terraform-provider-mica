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

// newTestVirtualHostResource creates an objectStoreVirtualHostResource wired to the given mock server.
func newTestVirtualHostResource(t *testing.T, ms *testmock.MockServer) *objectStoreVirtualHostResource {
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
	return &objectStoreVirtualHostResource{client: c}
}

// virtualHostResourceSchema returns the parsed schema for the virtual host resource.
func virtualHostResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &objectStoreVirtualHostResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

// buildVirtualHostType returns the tftypes.Object for the virtual host resource schema.
func buildVirtualHostType() tftypes.Object {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":               tftypes.String,
		"name":             tftypes.String,
		"hostname":         tftypes.String,
		"attached_servers": tftypes.List{ElementType: tftypes.String},
		"timeouts":         timeoutsType,
	}}
}

// nullVirtualHostConfig returns a base config map with all attributes null.
func nullVirtualHostConfig() map[string]tftypes.Value {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, nil),
		"name":             tftypes.NewValue(tftypes.String, nil),
		"hostname":         tftypes.NewValue(tftypes.String, nil),
		"attached_servers": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"timeouts":         tftypes.NewValue(timeoutsType, nil),
	}
}

// virtualHostPlanWith returns a tfsdk.Plan with the given name, hostname, and attached_servers.
func virtualHostPlanWith(t *testing.T, name, hostname string, servers []string) tfsdk.Plan {
	t.Helper()
	s := virtualHostResourceSchema(t).Schema
	cfg := nullVirtualHostConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, name)
	cfg["hostname"] = tftypes.NewValue(tftypes.String, hostname)

	if servers != nil {
		serverValues := make([]tftypes.Value, len(servers))
		for i, srv := range servers {
			serverValues[i] = tftypes.NewValue(tftypes.String, srv)
		}
		cfg["attached_servers"] = tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, serverValues)
	} else {
		// Empty list (not null) to match the default.
		cfg["attached_servers"] = tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{})
	}

	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildVirtualHostType(), cfg),
		Schema: s,
	}
}

// ---- tests ------------------------------------------------------------------

// TestUnit_ObjectStoreVirtualHost_CreateReadUpdateDelete tests the full lifecycle:
// Create with hostname + attached_servers, update attached_servers, destroy.
func TestUnit_ObjectStoreVirtualHost_CreateReadUpdateDelete(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterObjectStoreVirtualHostHandlers(ms.Mux)

	r := newTestVirtualHostResource(t, ms)
	s := virtualHostResourceSchema(t).Schema

	// --- Create ---
	plan := virtualHostPlanWith(t, "s3.example.com", "s3.example.com", []string{"server1"})
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildVirtualHostType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)

	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create returned error: %s", createResp.Diagnostics)
	}

	var model objectStoreVirtualHostModel
	if diags := createResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected non-empty ID after Create")
	}
	if model.Name.IsNull() || model.Name.ValueString() == "" {
		t.Error("expected non-empty Name (server-assigned) after Create")
	}
	if model.Hostname.ValueString() != "s3.example.com" {
		t.Errorf("expected hostname=s3.example.com, got %s", model.Hostname.ValueString())
	}

	// Verify attached_servers.
	var servers []string
	if diags := model.AttachedServers.ElementsAs(context.Background(), &servers, false); diags.HasError() {
		t.Fatalf("ElementsAs: %s", diags)
	}
	if len(servers) != 1 || servers[0] != "server1" {
		t.Errorf("expected attached_servers=[server1], got %v", servers)
	}

	// --- Read (verify no drift) ---
	readResp := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}

	var readModel objectStoreVirtualHostModel
	if diags := readResp.State.Get(context.Background(), &readModel); diags.HasError() {
		t.Fatalf("Get state after Read: %s", diags)
	}
	if readModel.Hostname.ValueString() != "s3.example.com" {
		t.Errorf("Read: expected hostname=s3.example.com, got %s", readModel.Hostname.ValueString())
	}

	// --- Update: change attached_servers to ["server1", "server2"] ---
	updatePlan := virtualHostPlanWith(t, "s3.example.com", "s3.example.com", []string{"server1", "server2"})
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildVirtualHostType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlan,
		State: createResp.State,
	}, updateResp)

	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update returned error: %s", updateResp.Diagnostics)
	}

	var updatedModel objectStoreVirtualHostModel
	if diags := updateResp.State.Get(context.Background(), &updatedModel); diags.HasError() {
		t.Fatalf("Get state after Update: %s", diags)
	}

	var updatedServers []string
	if diags := updatedModel.AttachedServers.ElementsAs(context.Background(), &updatedServers, false); diags.HasError() {
		t.Fatalf("ElementsAs: %s", diags)
	}
	if len(updatedServers) != 2 {
		t.Errorf("expected 2 attached_servers after update, got %d", len(updatedServers))
	}

	// --- Delete ---
	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: updateResp.State}, deleteResp)

	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete returned error: %s", deleteResp.Diagnostics)
	}

	// Verify the host is gone.
	_, err := r.client.GetObjectStoreVirtualHost(context.Background(), model.Name.ValueString())
	if err == nil || !client.IsNotFound(err) {
		t.Errorf("expected virtual host to be deleted, got: %v", err)
	}
}

// TestUnit_ObjectStoreVirtualHost_Import tests import by server-assigned name with no drift.
func TestUnit_ObjectStoreVirtualHost_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterObjectStoreVirtualHostHandlers(ms.Mux)

	r := newTestVirtualHostResource(t, ms)
	s := virtualHostResourceSchema(t).Schema

	// Create first.
	plan := virtualHostPlanWith(t, "s3.import.test", "s3.import.test", []string{"server1"})
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildVirtualHostType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	var createdModel objectStoreVirtualHostModel
	if diags := createResp.State.Get(context.Background(), &createdModel); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	// Import by server-assigned name.
	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildVirtualHostType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: createdModel.Name.ValueString()}, importResp)

	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState returned error: %s", importResp.Diagnostics)
	}

	var importedModel objectStoreVirtualHostModel
	if diags := importResp.State.Get(context.Background(), &importedModel); diags.HasError() {
		t.Fatalf("Get state after import: %s", diags)
	}

	if importedModel.Name.ValueString() != createdModel.Name.ValueString() {
		t.Errorf("expected name=%s after import, got %s", createdModel.Name.ValueString(), importedModel.Name.ValueString())
	}
	if importedModel.Hostname.ValueString() != "s3.import.test" {
		t.Errorf("expected hostname=s3.import.test after import, got %s", importedModel.Hostname.ValueString())
	}
	if importedModel.ID.IsNull() || importedModel.ID.ValueString() == "" {
		t.Error("expected ID to be populated after import")
	}

	// Verify no drift: imported attached_servers should match created.
	var importedServers []string
	if diags := importedModel.AttachedServers.ElementsAs(context.Background(), &importedServers, false); diags.HasError() {
		t.Fatalf("ElementsAs: %s", diags)
	}
	var createdServers []string
	if diags := createdModel.AttachedServers.ElementsAs(context.Background(), &createdServers, false); diags.HasError() {
		t.Fatalf("ElementsAs: %s", diags)
	}
	if len(importedServers) != len(createdServers) {
		t.Errorf("expected %d attached_servers after import, got %d", len(createdServers), len(importedServers))
	}
}

// TestUnit_ObjectStoreVirtualHost_UpdateHostname tests in-place hostname update.
func TestUnit_ObjectStoreVirtualHost_UpdateHostname(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterObjectStoreVirtualHostHandlers(ms.Mux)

	r := newTestVirtualHostResource(t, ms)
	s := virtualHostResourceSchema(t).Schema

	// Create with initial hostname.
	plan := virtualHostPlanWith(t, "s3.example.com", "s3.example.com", nil)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildVirtualHostType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Update hostname.
	updatePlan := virtualHostPlanWith(t, "s3.example.com", "s3-new.example.com", nil)
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildVirtualHostType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlan,
		State: createResp.State,
	}, updateResp)

	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update returned error: %s", updateResp.Diagnostics)
	}

	var model objectStoreVirtualHostModel
	if diags := updateResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Hostname.ValueString() != "s3-new.example.com" {
		t.Errorf("expected hostname=s3-new.example.com, got %s", model.Hostname.ValueString())
	}
}

// TestUnit_ObjectStoreVirtualHost_EmptyServers tests that creating with no attached_servers causes no drift.
func TestUnit_ObjectStoreVirtualHost_EmptyServers(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterObjectStoreVirtualHostHandlers(ms.Mux)

	r := newTestVirtualHostResource(t, ms)
	s := virtualHostResourceSchema(t).Schema

	// Create with no attached_servers (empty list).
	plan := virtualHostPlanWith(t, "s3.empty.test", "s3.empty.test", nil)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildVirtualHostType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)

	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create returned error: %s", createResp.Diagnostics)
	}

	var model objectStoreVirtualHostModel
	if diags := createResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	// Verify attached_servers is an empty list, not null.
	if model.AttachedServers.IsNull() {
		t.Error("expected attached_servers to be empty list, not null")
	}
	if len(model.AttachedServers.Elements()) != 0 {
		t.Errorf("expected empty attached_servers, got %d elements", len(model.AttachedServers.Elements()))
	}

	// Read again — verify no drift.
	readResp := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}

	var readModel objectStoreVirtualHostModel
	if diags := readResp.State.Get(context.Background(), &readModel); diags.HasError() {
		t.Fatalf("Get state after Read: %s", diags)
	}

	// The attached_servers should still be an empty list.
	if readModel.AttachedServers.IsNull() {
		t.Error("expected attached_servers to be empty list after Read, not null")
	}

	// Verify lists are equal (both empty) — no drift.
	if !model.AttachedServers.Equal(readModel.AttachedServers) {
		t.Error("drift detected: attached_servers differs between Create and Read")
	}
}

// TestUnit_ObjectStoreVirtualHost_Idempotent verifies that Read after Create shows no attribute drift.
func TestUnit_ObjectStoreVirtualHost_Idempotent(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterObjectStoreVirtualHostHandlers(ms.Mux)

	r := newTestVirtualHostResource(t, ms)
	s := virtualHostResourceSchema(t).Schema

	plan := virtualHostPlanWith(t, "s3.idempotent.test", "s3.idempotent.test", []string{"server1"})
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildVirtualHostType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Read the state back — should not change anything.
	readResp := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}

	var beforeModel, afterModel objectStoreVirtualHostModel
	if diags := createResp.State.Get(context.Background(), &beforeModel); diags.HasError() {
		t.Fatalf("Get before state: %s", diags)
	}
	if diags := readResp.State.Get(context.Background(), &afterModel); diags.HasError() {
		t.Fatalf("Get after state: %s", diags)
	}

	if beforeModel.ID.ValueString() != afterModel.ID.ValueString() {
		t.Errorf("ID changed after Read: %s -> %s", beforeModel.ID.ValueString(), afterModel.ID.ValueString())
	}
	if beforeModel.Name.ValueString() != afterModel.Name.ValueString() {
		t.Errorf("Name changed after Read: %s -> %s", beforeModel.Name.ValueString(), afterModel.Name.ValueString())
	}
	if beforeModel.Hostname.ValueString() != afterModel.Hostname.ValueString() {
		t.Errorf("Hostname changed after Read: %s -> %s", beforeModel.Hostname.ValueString(), afterModel.Hostname.ValueString())
	}
}

// TestUnit_ObjectStoreVirtualHostResource_StateUpgrade_V0toV1 verifies v0 state (name was Computed)
// upgrades correctly to v1 state (name is Required).
func TestUnit_ObjectStoreVirtualHostResource_StateUpgrade_V0toV1(t *testing.T) {
	r := &objectStoreVirtualHostResource{}
	upgraders := r.UpgradeState(context.Background())

	upgrader, ok := upgraders[0]
	if !ok {
		t.Fatal("expected v0→v1 state upgrader")
	}

	// Build v0 state with the prior schema.
	v0Type := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":       tftypes.String,
		"name":     tftypes.String,
		"hostname": tftypes.String,
		"attached_servers": tftypes.List{ElementType: tftypes.String},
		"timeouts": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
			"create": tftypes.String,
			"read":   tftypes.String,
			"update": tftypes.String,
			"delete": tftypes.String,
		}},
	}}

	v0Values := map[string]tftypes.Value{
		"id":       tftypes.NewValue(tftypes.String, "vh-id-001"),
		"name":     tftypes.NewValue(tftypes.String, "s3.example.com"),
		"hostname": tftypes.NewValue(tftypes.String, "s3.example.com"),
		"attached_servers": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
			tftypes.NewValue(tftypes.String, "server1"),
		}),
		"timeouts": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
			"create": tftypes.String,
			"read":   tftypes.String,
			"update": tftypes.String,
			"delete": tftypes.String,
		}}, map[string]tftypes.Value{
			"create": tftypes.NewValue(tftypes.String, nil),
			"read":   tftypes.NewValue(tftypes.String, nil),
			"update": tftypes.NewValue(tftypes.String, nil),
			"delete": tftypes.NewValue(tftypes.String, nil),
		}),
	}

	priorState := tfsdk.State{
		Raw:    tftypes.NewValue(v0Type, v0Values),
		Schema: *upgrader.PriorSchema,
	}

	resp := &resource.UpgradeStateResponse{
		State: tfsdk.State{
			Raw:    tftypes.NewValue(buildVirtualHostType(), nil),
			Schema: virtualHostResourceSchema(t).Schema,
		},
	}

	upgrader.StateUpgrader(context.Background(), resource.UpgradeStateRequest{State: &priorState}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("StateUpgrader returned error: %s", resp.Diagnostics)
	}

	var model objectStoreVirtualHostModel
	if diags := resp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get upgraded state: %s", diags)
	}

	if model.ID.ValueString() != "vh-id-001" {
		t.Errorf("expected ID=vh-id-001, got %s", model.ID.ValueString())
	}
	if model.Name.ValueString() != "s3.example.com" {
		t.Errorf("expected Name=s3.example.com, got %s", model.Name.ValueString())
	}
	if model.Hostname.ValueString() != "s3.example.com" {
		t.Errorf("expected Hostname=s3.example.com, got %s", model.Hostname.ValueString())
	}
}
