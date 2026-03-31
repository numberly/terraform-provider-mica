package provider

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	resschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/numberly/opentofu-provider-flashblade/internal/client"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock/handlers"
)

// ---- helpers ----------------------------------------------------------------

// newTestServerResource creates a serverResource wired to the given mock server.
func newTestServerResource(t *testing.T, ms *testmock.MockServer) *serverResource {
	t.Helper()
	c, err := client.NewClient(context.Background(), client.Config{
		Endpoint:           ms.URL(),
		APIToken:           "test-token",
		InsecureSkipVerify: true,
		MaxRetries:         1,
		RetryBaseDelay:     1 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return &serverResource{client: c}
}

// serverResourceSchema returns the parsed schema for the server resource.
func serverResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &serverResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

// buildServerType returns the tftypes.Object for the server resource schema (v2).
func buildServerType() tftypes.Object {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":                 tftypes.String,
		"name":               tftypes.String,
		"created":            tftypes.Number,
		"dns":                tftypes.List{ElementType: tftypes.String},
		"directory_services": tftypes.List{ElementType: tftypes.String},
		"cascade_delete":     tftypes.List{ElementType: tftypes.String},
		"timeouts":           timeoutsType,
		"network_interfaces": tftypes.List{ElementType: tftypes.String},
	}}
}

// nullServerConfig returns a base config map with all resource attributes null (v2 schema).
func nullServerConfig() map[string]tftypes.Value {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return map[string]tftypes.Value{
		"id":                 tftypes.NewValue(tftypes.String, nil),
		"name":               tftypes.NewValue(tftypes.String, nil),
		"created":            tftypes.NewValue(tftypes.Number, nil),
		"dns":                tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"directory_services": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"cascade_delete":     tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"timeouts":           tftypes.NewValue(timeoutsType, nil),
		"network_interfaces": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
	}
}

// serverPlanWithName returns a tfsdk.Plan with the given server name and no DNS config.
func serverPlanWithName(t *testing.T, name string) tfsdk.Plan {
	t.Helper()
	s := serverResourceSchema(t).Schema
	cfg := nullServerConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, name)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildServerType(), cfg),
		Schema: s,
	}
}

// serverPlanWithDNS returns a tfsdk.Plan with name and a flat list of DNS config names.
func serverPlanWithDNS(t *testing.T, name string, dnsNames []string) tfsdk.Plan {
	t.Helper()
	s := serverResourceSchema(t).Schema
	cfg := nullServerConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, name)
	if dnsNames != nil {
		vals := make([]tftypes.Value, len(dnsNames))
		for i, n := range dnsNames {
			vals[i] = tftypes.NewValue(tftypes.String, n)
		}
		cfg["dns"] = tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, vals)
	}
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildServerType(), cfg),
		Schema: s,
	}
}

// ---- tests ------------------------------------------------------------------

// TestUnit_Server_Create verifies Create populates ID, name, dns, and created timestamp.
func TestUnit_Server_Create(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterServerHandlers(ms.Mux)
	handlers.RegisterNetworkInterfaceHandlers(ms.Mux)

	r := newTestServerResource(t, ms)
	s := serverResourceSchema(t).Schema

	plan := serverPlanWithDNS(t, "test-server", []string{"management"})
	req := resource.CreateRequest{Plan: plan}
	resp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildServerType(), nil), Schema: s},
	}

	r.Create(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Create returned error: %s", resp.Diagnostics)
	}

	var model serverResourceModel
	if diags := resp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected non-empty ID after Create")
	}
	if model.Name.ValueString() != "test-server" {
		t.Errorf("expected name=test-server, got %s", model.Name.ValueString())
	}
	if model.Created.IsNull() || model.Created.IsUnknown() || model.Created.ValueInt64() == 0 {
		t.Error("expected created to be populated after Create")
	}
	if model.DNS.IsNull() {
		t.Error("expected dns to be populated after Create")
	}
	// network_interfaces should be an empty list (not null) when no VIPs are attached.
	if model.NetworkInterfaces.IsNull() {
		t.Error("expected network_interfaces to be empty list (not null) after Create with no VIPs")
	}
}

// TestUnit_Server_Read verifies Read populates all attributes from a seeded server.
func TestUnit_Server_Read(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterServerHandlers(ms.Mux)
	store.AddServer("read-server")
	handlers.RegisterNetworkInterfaceHandlers(ms.Mux)

	r := newTestServerResource(t, ms)
	s := serverResourceSchema(t).Schema

	cfg := nullServerConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "read-server")
	cfg["id"] = tftypes.NewValue(tftypes.String, "placeholder")
	state := tfsdk.State{Raw: tftypes.NewValue(buildServerType(), cfg), Schema: s}

	readResp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), resource.ReadRequest{State: state}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}

	var model serverResourceModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Name.ValueString() != "read-server" {
		t.Errorf("expected name=read-server, got %s", model.Name.ValueString())
	}
	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected ID to be populated")
	}
	if model.Created.IsNull() || model.Created.ValueInt64() == 0 {
		t.Error("expected created to be populated")
	}
	if model.DNS.IsNull() {
		t.Error("expected dns to be populated from seed data")
	}
	// network_interfaces should be an empty list (not null).
	if model.NetworkInterfaces.IsNull() {
		t.Error("expected network_interfaces to be empty list (not null) when no VIPs are attached")
	}
}

// TestUnit_Server_Update verifies PATCH updates dns configuration.
func TestUnit_Server_Update(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterServerHandlers(ms.Mux)
	handlers.RegisterNetworkInterfaceHandlers(ms.Mux)

	r := newTestServerResource(t, ms)
	s := serverResourceSchema(t).Schema

	// Create first.
	createPlan := serverPlanWithDNS(t, "update-server", []string{"management"})
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildServerType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Update DNS.
	updatePlan := serverPlanWithDNS(t, "update-server", []string{"updated-dns"})
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildServerType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlan,
		State: createResp.State,
	}, updateResp)

	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update returned error: %s", updateResp.Diagnostics)
	}

	var model serverResourceModel
	if diags := updateResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	// Verify DNS was updated.
	if model.DNS.IsNull() {
		t.Fatal("expected dns to be populated after Update")
	}
	var dnsNames []string
	if diags := model.DNS.ElementsAs(context.Background(), &dnsNames, false); diags.HasError() {
		t.Fatalf("ElementsAs: %s", diags)
	}
	if len(dnsNames) != 1 {
		t.Fatalf("expected 1 dns entry, got %d", len(dnsNames))
	}
	if dnsNames[0] != "updated-dns" {
		t.Errorf("expected dns name=updated-dns, got %s", dnsNames[0])
	}
}

// TestUnit_Server_Delete verifies DELETE removes the server.
func TestUnit_Server_Delete(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterServerHandlers(ms.Mux)
	handlers.RegisterNetworkInterfaceHandlers(ms.Mux)

	r := newTestServerResource(t, ms)
	s := serverResourceSchema(t).Schema

	plan := serverPlanWithName(t, "delete-server")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildServerType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: createResp.State}, deleteResp)

	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete returned error: %s", deleteResp.Diagnostics)
	}

	// Verify server is gone.
	_, err := r.client.GetServer(context.Background(), "delete-server")
	if err == nil || !client.IsNotFound(err) {
		t.Errorf("expected server to be deleted, got: %v", err)
	}
}

// TestUnit_Server_Import verifies ImportState populates all attributes and no drift.
func TestUnit_Server_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterServerHandlers(ms.Mux)
	handlers.RegisterNetworkInterfaceHandlers(ms.Mux)

	r := newTestServerResource(t, ms)
	s := serverResourceSchema(t).Schema

	// Create first.
	plan := serverPlanWithDNS(t, "import-server", []string{"management"})
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildServerType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Import by name.
	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildServerType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "import-server"}, importResp)

	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState returned error: %s", importResp.Diagnostics)
	}

	var model serverResourceModel
	if diags := importResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Name.ValueString() != "import-server" {
		t.Errorf("expected name=import-server after import, got %s", model.Name.ValueString())
	}
	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected ID to be populated after import")
	}
	if model.Created.IsNull() || model.Created.ValueInt64() == 0 {
		t.Error("expected created to be populated after import")
	}
	if model.DNS.IsNull() {
		t.Error("expected dns to be populated after import")
	}
	if model.NetworkInterfaces.IsNull() {
		t.Error("expected network_interfaces to be empty list (not null) after import")
	}
}

// TestUnit_Server_NotFound verifies that 404 removes resource from state.
func TestUnit_Server_NotFound(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterServerHandlers(ms.Mux)
	handlers.RegisterNetworkInterfaceHandlers(ms.Mux)

	r := newTestServerResource(t, ms)
	s := serverResourceSchema(t).Schema

	cfg := nullServerConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "does-not-exist")
	cfg["id"] = tftypes.NewValue(tftypes.String, "non-existent-id")
	state := tfsdk.State{Raw: tftypes.NewValue(buildServerType(), cfg), Schema: s}

	readResp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), resource.ReadRequest{State: state}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}
	if !readResp.State.Raw.IsNull() {
		t.Error("expected state to be removed (null) when resource not found")
	}
}

// TestUnit_Server_PlanModifiers verifies RequiresReplace on name and UseStateForUnknown on computed fields.
func TestUnit_Server_PlanModifiers(t *testing.T) {
	s := serverResourceSchema(t).Schema

	// name — RequiresReplace
	nameAttr, ok := s.Attributes["name"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("name attribute not found or wrong type")
	}
	if len(nameAttr.PlanModifiers) == 0 {
		t.Error("expected RequiresReplace plan modifier on name attribute")
	}

	// id — UseStateForUnknown
	idAttr, ok := s.Attributes["id"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("id attribute not found or wrong type")
	}
	if len(idAttr.PlanModifiers) == 0 {
		t.Error("expected UseStateForUnknown plan modifier on id attribute")
	}

	// created — int64 UseStateForUnknown
	createdAttr, ok := s.Attributes["created"].(resschema.Int64Attribute)
	if !ok {
		t.Fatal("created attribute not found or wrong type")
	}
	if len(createdAttr.PlanModifiers) == 0 {
		t.Error("expected UseStateForUnknown plan modifier on created attribute")
	}

	// network_interfaces — UseStateForUnknown
	niAttr, ok := s.Attributes["network_interfaces"].(resschema.ListAttribute)
	if !ok {
		t.Fatal("network_interfaces attribute not found or wrong type")
	}
	if len(niAttr.PlanModifiers) == 0 {
		t.Error("expected UseStateForUnknown plan modifier on network_interfaces attribute")
	}

	// dns — UseStateForUnknown
	dnsAttr, ok := s.Attributes["dns"].(resschema.ListAttribute)
	if !ok {
		t.Fatal("dns attribute not found or wrong type")
	}
	if len(dnsAttr.PlanModifiers) == 0 {
		t.Error("expected UseStateForUnknown plan modifier on dns attribute")
	}

	// directory_services — UseStateForUnknown
	dsAttr, ok := s.Attributes["directory_services"].(resschema.ListAttribute)
	if !ok {
		t.Fatal("directory_services attribute not found or wrong type")
	}
	if len(dsAttr.PlanModifiers) == 0 {
		t.Error("expected UseStateForUnknown plan modifier on directory_services attribute")
	}
}

// TestUnit_Server_Idempotent verifies that Read after Create shows no attribute drift.
func TestUnit_Server_Idempotent(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterServerHandlers(ms.Mux)
	handlers.RegisterNetworkInterfaceHandlers(ms.Mux)

	r := newTestServerResource(t, ms)
	s := serverResourceSchema(t).Schema

	plan := serverPlanWithName(t, "idempotent-server")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildServerType(), nil), Schema: s},
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

	var beforeModel, afterModel serverResourceModel
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
}

// TestUnit_Server_StateUpgradeV0ToV1 verifies the StateUpgrader migrates old state (without
// network_interfaces) to schema version 1 with an empty network_interfaces list.
// The v0->v1 upgrader outputs serverV1StateModel (nested DNS preserved, network_interfaces added).
func TestUnit_Server_StateUpgradeV0ToV1(t *testing.T) {
	r := &serverResource{}
	upgraders := r.UpgradeState(context.Background())

	upgrader, ok := upgraders[0]
	if !ok {
		t.Fatal("expected StateUpgrader for version 0")
	}

	// Build v0 state (no network_interfaces field, nested DNS).
	dnsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"domain":      tftypes.String,
		"nameservers": tftypes.List{ElementType: tftypes.String},
		"services":    tftypes.List{ElementType: tftypes.String},
	}}
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}

	v0Type := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":             tftypes.String,
		"name":           tftypes.String,
		"created":        tftypes.Number,
		"dns":            tftypes.List{ElementType: dnsType},
		"cascade_delete": tftypes.List{ElementType: tftypes.String},
		"timeouts":       timeoutsType,
	}}

	v0Raw := tftypes.NewValue(v0Type, map[string]tftypes.Value{
		"id":             tftypes.NewValue(tftypes.String, "srv-old-id"),
		"name":           tftypes.NewValue(tftypes.String, "my-server"),
		"created":        tftypes.NewValue(tftypes.Number, 1700000000000),
		"dns":            tftypes.NewValue(tftypes.List{ElementType: dnsType}, nil),
		"cascade_delete": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"timeouts":       tftypes.NewValue(timeoutsType, nil),
	})

	priorState := tfsdk.State{
		Raw:    v0Raw,
		Schema: *upgrader.PriorSchema,
	}

	// v1 target schema (nested DNS + network_interfaces).
	v1Schema := *upgraders[1].PriorSchema

	upgradeReq := resource.UpgradeStateRequest{
		State: &priorState,
	}

	// Build the v1 tftypes target.
	v1Type := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":                 tftypes.String,
		"name":               tftypes.String,
		"created":            tftypes.Number,
		"dns":                tftypes.List{ElementType: dnsType},
		"cascade_delete":     tftypes.List{ElementType: tftypes.String},
		"network_interfaces": tftypes.List{ElementType: tftypes.String},
		"timeouts":           timeoutsType,
	}}
	upgradeResp := &resource.UpgradeStateResponse{
		State: tfsdk.State{
			Raw:    tftypes.NewValue(v1Type, nil),
			Schema: v1Schema,
		},
	}

	upgrader.StateUpgrader(context.Background(), upgradeReq, upgradeResp)

	if upgradeResp.Diagnostics.HasError() {
		t.Fatalf("StateUpgrader v0->v1 returned error: %s", upgradeResp.Diagnostics)
	}

	var model serverV1StateModel
	if diags := upgradeResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get upgraded state: %s", diags)
	}

	// Verify existing fields preserved.
	if model.ID.ValueString() != "srv-old-id" {
		t.Errorf("expected id=srv-old-id, got %s", model.ID.ValueString())
	}
	if model.Name.ValueString() != "my-server" {
		t.Errorf("expected name=my-server, got %s", model.Name.ValueString())
	}

	// Verify network_interfaces is empty list (not null).
	if model.NetworkInterfaces.IsNull() {
		t.Error("expected network_interfaces to be empty list (not null) after upgrade from v0")
	}
	if model.NetworkInterfaces.IsUnknown() {
		t.Error("expected network_interfaces to be known after upgrade from v0")
	}
	if len(model.NetworkInterfaces.Elements()) != 0 {
		t.Errorf("expected network_interfaces to have 0 elements, got %d", len(model.NetworkInterfaces.Elements()))
	}
}

// TestUnit_Server_StateUpgradeV1ToV2 verifies the v1->v2 upgrader converts nested DNS to null
// and adds directory_services as null.
func TestUnit_Server_StateUpgradeV1ToV2(t *testing.T) {
	r := &serverResource{}
	upgraders := r.UpgradeState(context.Background())

	upgrader, ok := upgraders[1]
	if !ok {
		t.Fatal("expected StateUpgrader for version 1")
	}

	// Build v1 raw state (nested DNS, network_interfaces, no directory_services).
	dnsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"domain":      tftypes.String,
		"nameservers": tftypes.List{ElementType: tftypes.String},
		"services":    tftypes.List{ElementType: tftypes.String},
	}}
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}

	v1Type := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":                 tftypes.String,
		"name":               tftypes.String,
		"created":            tftypes.Number,
		"dns":                tftypes.List{ElementType: dnsType},
		"cascade_delete":     tftypes.List{ElementType: tftypes.String},
		"network_interfaces": tftypes.List{ElementType: tftypes.String},
		"timeouts":           timeoutsType,
	}}

	v1Raw := tftypes.NewValue(v1Type, map[string]tftypes.Value{
		"id":      tftypes.NewValue(tftypes.String, "srv-v1-id"),
		"name":    tftypes.NewValue(tftypes.String, "v1-server"),
		"created": tftypes.NewValue(tftypes.Number, 1700000000000),
		"dns":     tftypes.NewValue(tftypes.List{ElementType: dnsType}, nil),
		"cascade_delete": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
			tftypes.NewValue(tftypes.String, "export1"),
		}),
		"network_interfaces": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
			tftypes.NewValue(tftypes.String, "vip1.eth0"),
		}),
		"timeouts": tftypes.NewValue(timeoutsType, nil),
	})

	priorState := tfsdk.State{
		Raw:    v1Raw,
		Schema: *upgrader.PriorSchema,
	}

	// v2 target schema.
	v2Schema := serverResourceSchema(t).Schema

	upgradeReq := resource.UpgradeStateRequest{
		State: &priorState,
	}
	upgradeResp := &resource.UpgradeStateResponse{
		State: tfsdk.State{
			Raw:    tftypes.NewValue(buildServerType(), nil),
			Schema: v2Schema,
		},
	}

	upgrader.StateUpgrader(context.Background(), upgradeReq, upgradeResp)

	if upgradeResp.Diagnostics.HasError() {
		t.Fatalf("StateUpgrader v1->v2 returned error: %s", upgradeResp.Diagnostics)
	}

	var model serverResourceModel
	if diags := upgradeResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get upgraded state: %s", diags)
	}

	// Verify preserved fields.
	if model.ID.ValueString() != "srv-v1-id" {
		t.Errorf("expected id=srv-v1-id, got %s", model.ID.ValueString())
	}
	if model.Name.ValueString() != "v1-server" {
		t.Errorf("expected name=v1-server, got %s", model.Name.ValueString())
	}

	// DNS must be null (will be refreshed on next plan/apply).
	if !model.DNS.IsNull() {
		t.Error("expected dns to be null after v1->v2 upgrade")
	}

	// directory_services must be null.
	if !model.DirectoryServices.IsNull() {
		t.Error("expected directory_services to be null after v1->v2 upgrade")
	}

	// NetworkInterfaces must be preserved.
	if model.NetworkInterfaces.IsNull() {
		t.Fatal("expected network_interfaces to be preserved after v1->v2 upgrade")
	}
	var niNames []string
	if diags := model.NetworkInterfaces.ElementsAs(context.Background(), &niNames, false); diags.HasError() {
		t.Fatalf("ElementsAs: %s", diags)
	}
	if len(niNames) != 1 || niNames[0] != "vip1.eth0" {
		t.Errorf("expected network_interfaces=[vip1.eth0], got %v", niNames)
	}

	// CascadeDelete must be preserved.
	if model.CascadeDelete.IsNull() {
		t.Fatal("expected cascade_delete to be preserved after v1->v2 upgrade")
	}
	var cascadeNames []string
	if diags := model.CascadeDelete.ElementsAs(context.Background(), &cascadeNames, false); diags.HasError() {
		t.Fatalf("ElementsAs: %s", diags)
	}
	if len(cascadeNames) != 1 || cascadeNames[0] != "export1" {
		t.Errorf("expected cascade_delete=[export1], got %v", cascadeNames)
	}
}

// TestUnit_Server_VIPEnrichment verifies that Create/Read populates network_interfaces
// when VIPs are attached to the server.
func TestUnit_Server_VIPEnrichment(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterServerHandlers(ms.Mux)
	niStore := handlers.RegisterNetworkInterfaceHandlers(ms.Mux)

	// Seed two VIPs, one attached to "vip-server", one to another server.
	ni1 := niStore.AddNetworkInterface("vip1.eth0", "10.0.1.1", "subnet-a", "vip", "data")
	ni1.AttachedServers = []client.NamedReference{{Name: "vip-server"}}

	ni2 := niStore.AddNetworkInterface("vip2.eth0", "10.0.1.2", "subnet-a", "vip", "data")
	ni2.AttachedServers = []client.NamedReference{{Name: "other-server"}}

	r := newTestServerResource(t, ms)
	s := serverResourceSchema(t).Schema

	// Create the server.
	plan := serverPlanWithName(t, "vip-server")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildServerType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	var model serverResourceModel
	if diags := createResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	// network_interfaces should contain only vip1.eth0 (attached to vip-server).
	if model.NetworkInterfaces.IsNull() {
		t.Fatal("expected network_interfaces to be populated after Create with attached VIPs")
	}

	var niNames []string
	if diags := model.NetworkInterfaces.ElementsAs(context.Background(), &niNames, false); diags.HasError() {
		t.Fatalf("ElementsAs: %s", diags)
	}
	if len(niNames) != 1 {
		t.Fatalf("expected 1 network interface, got %d: %v", len(niNames), niNames)
	}
	if niNames[0] != "vip1.eth0" {
		t.Errorf("expected vip1.eth0, got %s", niNames[0])
	}
}

// TestUnit_Server_VIPEnrichment_Read verifies Read also populates network_interfaces.
func TestUnit_Server_VIPEnrichment_Read(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	srvStore := handlers.RegisterServerHandlers(ms.Mux)
	srvStore.AddServer("enrich-read-server")
	niStore := handlers.RegisterNetworkInterfaceHandlers(ms.Mux)

	ni := niStore.AddNetworkInterface("eth0.enrich", "10.0.2.1", "subnet-b", "vip", "data")
	ni.AttachedServers = []client.NamedReference{{Name: "enrich-read-server"}}

	r := newTestServerResource(t, ms)
	s := serverResourceSchema(t).Schema

	cfg := nullServerConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "enrich-read-server")
	cfg["id"] = tftypes.NewValue(tftypes.String, "placeholder")
	state := tfsdk.State{Raw: tftypes.NewValue(buildServerType(), cfg), Schema: s}

	readResp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), resource.ReadRequest{State: state}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}

	var model serverResourceModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.NetworkInterfaces.IsNull() {
		t.Fatal("expected network_interfaces to be populated after Read with attached VIPs")
	}

	var niNames []string
	if diags := model.NetworkInterfaces.ElementsAs(context.Background(), &niNames, false); diags.HasError() {
		t.Fatalf("ElementsAs: %s", diags)
	}
	if len(niNames) != 1 {
		t.Fatalf("expected 1 network interface, got %d: %v", len(niNames), niNames)
	}
	if niNames[0] != "eth0.enrich" {
		t.Errorf("expected eth0.enrich, got %s", niNames[0])
	}
}

// TestUnit_Server_NoVIPs verifies that a server with no attached VIPs gets empty (not null) list.
func TestUnit_Server_NoVIPs(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterServerHandlers(ms.Mux)
	niStore := handlers.RegisterNetworkInterfaceHandlers(ms.Mux)
	// VIP attached to a different server.
	ni := niStore.AddNetworkInterface("vip.other", "10.0.3.1", "subnet-c", "vip", "data")
	ni.AttachedServers = []client.NamedReference{{Name: "other-server"}}

	r := newTestServerResource(t, ms)
	s := serverResourceSchema(t).Schema

	plan := serverPlanWithName(t, "no-vip-server")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildServerType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	var model serverResourceModel
	if diags := createResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.NetworkInterfaces.IsNull() {
		t.Error("expected network_interfaces to be empty list (not null) when no VIPs are attached")
	}
	if len(model.NetworkInterfaces.Elements()) != 0 {
		t.Errorf("expected 0 network interfaces, got %d", len(model.NetworkInterfaces.Elements()))
	}
}

// TestUnit_Server_SchemaVersion verifies that schema version is 2.
func TestUnit_Server_SchemaVersion(t *testing.T) {
	s := serverResourceSchema(t).Schema
	if s.Version != 2 {
		t.Errorf("expected schema version 2, got %d", s.Version)
	}
}
