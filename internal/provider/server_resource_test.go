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
	c, err := client.NewClient(client.Config{
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

// buildServerType returns the tftypes.Object for the server resource schema.
func buildServerType() tftypes.Object {
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
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":             tftypes.String,
		"name":           tftypes.String,
		"created":        tftypes.Number,
		"dns":            tftypes.List{ElementType: dnsType},
		"cascade_delete": tftypes.List{ElementType: tftypes.String},
		"timeouts":       timeoutsType,
	}}
}

// nullServerConfig returns a base config map with all resource attributes null.
func nullServerConfig() map[string]tftypes.Value {
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
	return map[string]tftypes.Value{
		"id":             tftypes.NewValue(tftypes.String, nil),
		"name":           tftypes.NewValue(tftypes.String, nil),
		"created":        tftypes.NewValue(tftypes.Number, nil),
		"dns":            tftypes.NewValue(tftypes.List{ElementType: dnsType}, nil),
		"cascade_delete": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"timeouts":       tftypes.NewValue(timeoutsType, nil),
	}
}

// serverPlanWithName returns a tfsdk.Plan with the given server name and DNS config.
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

// serverPlanWithDNS returns a tfsdk.Plan with name and a DNS configuration.
func serverPlanWithDNS(t *testing.T, name, domain string, nameservers []string) tfsdk.Plan {
	t.Helper()
	s := serverResourceSchema(t).Schema

	dnsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"domain":      tftypes.String,
		"nameservers": tftypes.List{ElementType: tftypes.String},
		"services":    tftypes.List{ElementType: tftypes.String},
	}}

	// Build nameservers tftypes list.
	nsValues := make([]tftypes.Value, len(nameservers))
	for i, ns := range nameservers {
		nsValues[i] = tftypes.NewValue(tftypes.String, ns)
	}

	dnsObj := tftypes.NewValue(dnsType, map[string]tftypes.Value{
		"domain":      tftypes.NewValue(tftypes.String, domain),
		"nameservers": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nsValues),
		"services":    tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
	})

	cfg := nullServerConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, name)
	cfg["dns"] = tftypes.NewValue(tftypes.List{ElementType: dnsType}, []tftypes.Value{dnsObj})

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

	r := newTestServerResource(t, ms)
	s := serverResourceSchema(t).Schema

	plan := serverPlanWithDNS(t, "test-server", "example.com", []string{"10.0.0.1", "10.0.0.2"})
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
}

// TestUnit_Server_Read verifies Read populates all attributes from a seeded server.
func TestUnit_Server_Read(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterServerHandlers(ms.Mux)
	store.AddServer("read-server")

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
}

// TestUnit_Server_Update verifies PATCH updates dns configuration.
func TestUnit_Server_Update(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterServerHandlers(ms.Mux)

	r := newTestServerResource(t, ms)
	s := serverResourceSchema(t).Schema

	// Create first.
	createPlan := serverPlanWithDNS(t, "update-server", "old.example.com", []string{"10.0.0.1"})
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildServerType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Update DNS.
	updatePlan := serverPlanWithDNS(t, "update-server", "new.example.com", []string{"10.0.0.2"})
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
	var dnsModels []serverDNSModel
	if diags := model.DNS.ElementsAs(context.Background(), &dnsModels, false); diags.HasError() {
		t.Fatalf("ElementsAs: %s", diags)
	}
	if len(dnsModels) != 1 {
		t.Fatalf("expected 1 dns entry, got %d", len(dnsModels))
	}
	if dnsModels[0].Domain.ValueString() != "new.example.com" {
		t.Errorf("expected domain=new.example.com, got %s", dnsModels[0].Domain.ValueString())
	}
}

// TestUnit_Server_Delete verifies DELETE removes the server.
func TestUnit_Server_Delete(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterServerHandlers(ms.Mux)

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

	r := newTestServerResource(t, ms)
	s := serverResourceSchema(t).Schema

	// Create first.
	plan := serverPlanWithDNS(t, "import-server", "import.local", []string{"10.0.0.1"})
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
}

// TestUnit_Server_NotFound verifies that 404 removes resource from state.
func TestUnit_Server_NotFound(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterServerHandlers(ms.Mux)

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
}
