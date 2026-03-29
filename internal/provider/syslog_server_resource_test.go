package provider

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/numberly/opentofu-provider-flashblade/internal/client"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock/handlers"
)

// ---- helpers ----------------------------------------------------------------

// newTestSyslogServerResource creates a syslogServerResource wired to the given mock server.
func newTestSyslogServerResource(t *testing.T, ms *testmock.MockServer) *syslogServerResource {
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
	return &syslogServerResource{client: c}
}

// syslogServerResourceSchema returns the parsed schema for the syslog server resource.
func syslogServerResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &syslogServerResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

// buildSyslogServerType returns the tftypes.Object for the syslog server resource schema.
func buildSyslogServerType() tftypes.Object {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":       tftypes.String,
		"name":     tftypes.String,
		"uri":      tftypes.String,
		"services": tftypes.List{ElementType: tftypes.String},
		"sources":  tftypes.List{ElementType: tftypes.String},
		"timeouts": timeoutsType,
	}}
}

// nullSyslogServerConfig returns a base config map with all attributes null.
func nullSyslogServerConfig() map[string]tftypes.Value {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return map[string]tftypes.Value{
		"id":       tftypes.NewValue(tftypes.String, nil),
		"name":     tftypes.NewValue(tftypes.String, nil),
		"uri":      tftypes.NewValue(tftypes.String, nil),
		"services": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"sources":  tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"timeouts": tftypes.NewValue(timeoutsType, nil),
	}
}

// syslogServerPlan returns a tfsdk.Plan with the given name, uri, services, and sources.
func syslogServerPlan(t *testing.T, name, uri string, services, sources []string) tfsdk.Plan {
	t.Helper()
	s := syslogServerResourceSchema(t).Schema
	cfg := nullSyslogServerConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, name)
	cfg["uri"] = tftypes.NewValue(tftypes.String, uri)

	if services != nil {
		vals := make([]tftypes.Value, len(services))
		for i, svc := range services {
			vals[i] = tftypes.NewValue(tftypes.String, svc)
		}
		cfg["services"] = tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, vals)
	} else {
		cfg["services"] = tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{})
	}

	if sources != nil {
		vals := make([]tftypes.Value, len(sources))
		for i, src := range sources {
			vals[i] = tftypes.NewValue(tftypes.String, src)
		}
		cfg["sources"] = tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, vals)
	} else {
		cfg["sources"] = tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{})
	}

	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildSyslogServerType(), cfg),
		Schema: s,
	}
}

// ---- tests ------------------------------------------------------------------

// TestUnit_SyslogServer_CRUD exercises the full Create->Read->Update->Delete sequence.
func TestUnit_SyslogServer_CRUD(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterSyslogServerHandlers(ms.Mux)

	r := newTestSyslogServerResource(t, ms)
	s := syslogServerResourceSchema(t).Schema

	// Step 1: Create with uri, services=["data-audit"].
	createPlan := syslogServerPlan(t, "test-syslog", "udp://syslog.example.com:514", []string{"data-audit"}, nil)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSyslogServerType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}
	var createModel syslogServerModel
	if diags := createResp.State.Get(context.Background(), &createModel); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}
	if createModel.Name.ValueString() != "test-syslog" {
		t.Errorf("Create: expected name=test-syslog, got %s", createModel.Name.ValueString())
	}
	if createModel.URI.ValueString() != "udp://syslog.example.com:514" {
		t.Errorf("Create: expected uri=udp://syslog.example.com:514, got %s", createModel.URI.ValueString())
	}
	if createModel.ID.IsNull() || createModel.ID.ValueString() == "" {
		t.Error("Create: expected non-empty ID")
	}

	// Verify services has one element.
	var createServices []string
	if diags := createModel.Services.ElementsAs(context.Background(), &createServices, false); diags.HasError() {
		t.Fatalf("ElementsAs services: %s", diags)
	}
	if len(createServices) != 1 || createServices[0] != "data-audit" {
		t.Errorf("Create: expected services=[data-audit], got %v", createServices)
	}

	// Verify sources is empty list (not null).
	if createModel.Sources.IsNull() {
		t.Error("Create: expected sources to be empty list, not null")
	}
	if len(createModel.Sources.Elements()) != 0 {
		t.Errorf("Create: expected empty sources, got %d elements", len(createModel.Sources.Elements()))
	}

	// Step 2: Read post-create (0-diff check).
	readResp1 := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp1)
	if readResp1.Diagnostics.HasError() {
		t.Fatalf("Read post-create: %s", readResp1.Diagnostics)
	}

	// Step 3: Update uri and services.
	updatePlan := syslogServerPlan(t, "test-syslog", "tcp://syslog.example.com:6514", []string{"data-audit", "management"}, nil)
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSyslogServerType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlan,
		State: readResp1.State,
	}, updateResp)
	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update: %s", updateResp.Diagnostics)
	}
	var updateModel syslogServerModel
	if diags := updateResp.State.Get(context.Background(), &updateModel); diags.HasError() {
		t.Fatalf("Get update state: %s", diags)
	}
	if updateModel.URI.ValueString() != "tcp://syslog.example.com:6514" {
		t.Errorf("Update: expected uri=tcp://syslog.example.com:6514, got %s", updateModel.URI.ValueString())
	}
	var updateServices []string
	if diags := updateModel.Services.ElementsAs(context.Background(), &updateServices, false); diags.HasError() {
		t.Fatalf("ElementsAs services: %s", diags)
	}
	if len(updateServices) != 2 {
		t.Errorf("Update: expected 2 services, got %d", len(updateServices))
	}

	// Step 4: Read post-update (0-diff check).
	readResp2 := &resource.ReadResponse{State: updateResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: updateResp.State}, readResp2)
	if readResp2.Diagnostics.HasError() {
		t.Fatalf("Read post-update: %s", readResp2.Diagnostics)
	}

	// Step 5: Delete.
	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: updateResp.State}, deleteResp)
	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete: %s", deleteResp.Diagnostics)
	}
	_, err := r.client.GetSyslogServer(context.Background(), "test-syslog")
	if err == nil || !client.IsNotFound(err) {
		t.Errorf("expected syslog server to be deleted, got: %v", err)
	}
}

// TestUnit_SyslogServer_Import verifies ImportState populates all attributes correctly.
func TestUnit_SyslogServer_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterSyslogServerHandlers(ms.Mux)

	r := newTestSyslogServerResource(t, ms)
	s := syslogServerResourceSchema(t).Schema

	// Create first.
	createPlan := syslogServerPlan(t, "import-syslog", "udp://syslog.example.com:514", []string{"data-audit"}, nil)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSyslogServerType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}
	var createModel syslogServerModel
	if diags := createResp.State.Get(context.Background(), &createModel); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}

	// Import by name.
	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSyslogServerType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "import-syslog"}, importResp)
	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState: %s", importResp.Diagnostics)
	}

	var importedModel syslogServerModel
	if diags := importResp.State.Get(context.Background(), &importedModel); diags.HasError() {
		t.Fatalf("Get imported state: %s", diags)
	}

	// Verify 0-diff.
	if importedModel.Name.ValueString() != createModel.Name.ValueString() {
		t.Errorf("name mismatch: create=%s import=%s", createModel.Name.ValueString(), importedModel.Name.ValueString())
	}
	if importedModel.URI.ValueString() != createModel.URI.ValueString() {
		t.Errorf("uri mismatch: create=%s import=%s", createModel.URI.ValueString(), importedModel.URI.ValueString())
	}
	if importedModel.ID.ValueString() != createModel.ID.ValueString() {
		t.Errorf("id mismatch: create=%s import=%s", createModel.ID.ValueString(), importedModel.ID.ValueString())
	}
	if !importedModel.Services.Equal(createModel.Services) {
		t.Errorf("services mismatch after import")
	}
	if !importedModel.Sources.Equal(createModel.Sources) {
		t.Errorf("sources mismatch after import")
	}
}

// ---- data source tests -------------------------------------------------------

// newTestSyslogServerDataSource creates a syslogServerDataSource wired to the given mock server.
func newTestSyslogServerDataSource(t *testing.T, ms *testmock.MockServer) *syslogServerDataSource {
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
	return &syslogServerDataSource{client: c}
}

// syslogServerDataSourceSchema returns the schema for the syslog server data source.
func syslogServerDataSourceSchema(t *testing.T) datasource.SchemaResponse {
	t.Helper()
	d := &syslogServerDataSource{}
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	return resp
}

// buildSyslogServerDSType returns the tftypes.Object for the syslog server data source.
func buildSyslogServerDSType() tftypes.Object {
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":       tftypes.String,
		"name":     tftypes.String,
		"uri":      tftypes.String,
		"services": tftypes.List{ElementType: tftypes.String},
		"sources":  tftypes.List{ElementType: tftypes.String},
	}}
}

// nullSyslogServerDSConfig returns a base config map with all data source attributes null.
func nullSyslogServerDSConfig() map[string]tftypes.Value {
	return map[string]tftypes.Value{
		"id":       tftypes.NewValue(tftypes.String, nil),
		"name":     tftypes.NewValue(tftypes.String, nil),
		"uri":      tftypes.NewValue(tftypes.String, nil),
		"services": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"sources":  tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
	}
}

// TestUnit_SyslogServer_DataSource verifies data source reads syslog server by name.
func TestUnit_SyslogServer_DataSource(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterSyslogServerHandlers(ms.Mux)

	// Create a syslog server via the client so the data source can find it.
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
	_, err = c.PostSyslogServer(context.Background(), "ds-syslog-test", client.SyslogServerPost{
		URI:      "udp://syslog.example.com:514",
		Services: []string{"data-audit"},
	})
	if err != nil {
		t.Fatalf("PostSyslogServer: %v", err)
	}

	d := newTestSyslogServerDataSource(t, ms)
	s := syslogServerDataSourceSchema(t).Schema

	cfg := nullSyslogServerDSConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "ds-syslog-test")

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSyslogServerDSType(), nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(buildSyslogServerDSType(), cfg), Schema: s},
	}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("DataSource Read returned error: %s", readResp.Diagnostics)
	}

	var model syslogServerDataSourceModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Name.ValueString() != "ds-syslog-test" {
		t.Errorf("expected name=ds-syslog-test, got %s", model.Name.ValueString())
	}
	if model.URI.ValueString() != "udp://syslog.example.com:514" {
		t.Errorf("expected uri=udp://syslog.example.com:514, got %s", model.URI.ValueString())
	}
	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected ID to be populated")
	}

	var services []string
	if diags := model.Services.ElementsAs(context.Background(), &services, false); diags.HasError() {
		t.Fatalf("ElementsAs services: %s", diags)
	}
	if len(services) != 1 || services[0] != "data-audit" {
		t.Errorf("expected services=[data-audit], got %v", services)
	}
}

// TestUnit_SyslogServer_Idempotent verifies that Read after Create shows no attribute drift.
func TestUnit_SyslogServer_Idempotent(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterSyslogServerHandlers(ms.Mux)

	r := newTestSyslogServerResource(t, ms)
	s := syslogServerResourceSchema(t).Schema

	plan := syslogServerPlan(t, "idempotent-syslog", "tcp://syslog.example.com:514", []string{"data-audit"}, nil)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSyslogServerType(), nil), Schema: s},
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

	var beforeModel, afterModel syslogServerModel
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
	if beforeModel.URI.ValueString() != afterModel.URI.ValueString() {
		t.Errorf("URI changed after Read: %s -> %s", beforeModel.URI.ValueString(), afterModel.URI.ValueString())
	}
}

// TestUnit_SyslogServer_Update verifies PATCH updates URI and name is unchanged.
func TestUnit_SyslogServer_Update(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterSyslogServerHandlers(ms.Mux)

	r := newTestSyslogServerResource(t, ms)
	s := syslogServerResourceSchema(t).Schema

	// Create with initial URI.
	createPlan := syslogServerPlan(t, "update-syslog", "tcp://syslog1.example.com:514", []string{"data-audit"}, nil)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSyslogServerType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Update URI.
	updatePlan := syslogServerPlan(t, "update-syslog", "tcp://syslog2.example.com:514", []string{"data-audit"}, nil)
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSyslogServerType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlan,
		State: createResp.State,
	}, updateResp)
	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update: %s", updateResp.Diagnostics)
	}

	var model syslogServerModel
	if diags := updateResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.URI.ValueString() != "tcp://syslog2.example.com:514" {
		t.Errorf("expected uri=tcp://syslog2.example.com:514, got %s", model.URI.ValueString())
	}
	if model.Name.ValueString() != "update-syslog" {
		t.Errorf("expected name unchanged=update-syslog, got %s", model.Name.ValueString())
	}
}
