package provider

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/soulkyu/terraform-provider-flashblade/internal/client"
	"github.com/soulkyu/terraform-provider-flashblade/internal/testmock"
	"github.com/soulkyu/terraform-provider-flashblade/internal/testmock/handlers"
)

// ---- helpers ----------------------------------------------------------------

func newTestArrayNtpResource(t *testing.T, ms *testmock.MockServer) *arrayNtpResource {
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
	return &arrayNtpResource{client: c}
}

func arrayNtpResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &arrayNtpResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

func buildArrayNtpType() tftypes.Object {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":          tftypes.String,
		"ntp_servers": tftypes.List{ElementType: tftypes.String},
		"timeouts":    timeoutsType,
	}}
}

func nullArrayNtpConfig() map[string]tftypes.Value {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return map[string]tftypes.Value{
		"id":          tftypes.NewValue(tftypes.String, nil),
		"ntp_servers": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"timeouts":    tftypes.NewValue(timeoutsType, nil),
	}
}

func arrayNtpPlanWith(t *testing.T, servers []string) tfsdk.Plan {
	t.Helper()
	s := arrayNtpResourceSchema(t).Schema
	cfg := nullArrayNtpConfig()
	nsValues := make([]tftypes.Value, len(servers))
	for i, ns := range servers {
		nsValues[i] = tftypes.NewValue(tftypes.String, ns)
	}
	cfg["ntp_servers"] = tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nsValues)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildArrayNtpType(), cfg),
		Schema: s,
	}
}

// ---- resource tests ---------------------------------------------------------

// TestArrayNtpResource_Create verifies that Create sets ntp_servers.
func TestArrayNtpResource_Create(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterArrayAdminHandlers(ms.Mux)

	r := newTestArrayNtpResource(t, ms)
	s := arrayNtpResourceSchema(t).Schema

	plan := arrayNtpPlanWith(t, []string{"time.google.com"})
	resp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildArrayNtpType(), nil), Schema: s},
	}

	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Create returned error: %s", resp.Diagnostics)
	}

	var model arrayNtpModel
	if diags := resp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected non-empty ID after Create")
	}

	var servers []string
	if diags := model.NtpServers.ElementsAs(context.Background(), &servers, false); diags.HasError() {
		t.Fatalf("Get ntp_servers: %s", diags)
	}
	if len(servers) != 1 || servers[0] != "time.google.com" {
		t.Errorf("expected ntp_servers=[time.google.com], got %v", servers)
	}
}

// TestArrayNtpResource_Update verifies that Update adds an NTP server via PATCH.
func TestArrayNtpResource_Update(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterArrayAdminHandlers(ms.Mux)

	r := newTestArrayNtpResource(t, ms)
	s := arrayNtpResourceSchema(t).Schema

	// Create first.
	createPlan := arrayNtpPlanWith(t, []string{"time.google.com"})
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildArrayNtpType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Update to add a second server.
	updatePlan := arrayNtpPlanWith(t, []string{"time.google.com", "pool.ntp.org"})
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildArrayNtpType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlan,
		State: createResp.State,
	}, updateResp)

	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update returned error: %s", updateResp.Diagnostics)
	}

	var model arrayNtpModel
	if diags := updateResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	var servers []string
	if diags := model.NtpServers.ElementsAs(context.Background(), &servers, false); diags.HasError() {
		t.Fatalf("Get ntp_servers: %s", diags)
	}
	if len(servers) != 2 {
		t.Errorf("expected 2 ntp_servers after update, got %v", servers)
	}
}

// TestArrayNtpResource_Delete verifies that Delete clears ntp_servers.
func TestArrayNtpResource_Delete(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterArrayAdminHandlers(ms.Mux)

	r := newTestArrayNtpResource(t, ms)
	s := arrayNtpResourceSchema(t).Schema

	// Create first.
	createPlan := arrayNtpPlanWith(t, []string{"time.google.com"})
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildArrayNtpType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Delete clears servers.
	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: createResp.State}, deleteResp)

	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete returned error: %s", deleteResp.Diagnostics)
	}

	// Verify reset via client.
	arrayInfo, err := r.client.GetArrayNtp(context.Background())
	if err != nil {
		t.Fatalf("GetArrayNtp after delete: %v", err)
	}
	if len(arrayInfo.NtpServers) != 0 {
		t.Errorf("expected ntp_servers to be empty after delete, got %v", arrayInfo.NtpServers)
	}
}

// TestArrayNtpResource_Import verifies ImportState populates all attributes.
func TestArrayNtpResource_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterArrayAdminHandlers(ms.Mux)

	r := newTestArrayNtpResource(t, ms)
	s := arrayNtpResourceSchema(t).Schema

	// Set up NTP config via direct client call.
	servers := []string{"ntp1.example.com"}
	if _, err := r.client.PatchArrayNtp(context.Background(), client.ArrayNtpPatch{
		NtpServers: &servers,
	}); err != nil {
		t.Fatalf("PatchArrayNtp: %v", err)
	}

	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildArrayNtpType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "default"}, importResp)

	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState returned error: %s", importResp.Diagnostics)
	}

	var model arrayNtpModel
	if diags := importResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected non-empty ID after import")
	}

	var ntpServers []string
	if diags := model.NtpServers.ElementsAs(context.Background(), &ntpServers, false); diags.HasError() {
		t.Fatalf("Get ntp_servers: %s", diags)
	}
	if len(ntpServers) != 1 || ntpServers[0] != "ntp1.example.com" {
		t.Errorf("expected ntp_servers=[ntp1.example.com] after import, got %v", ntpServers)
	}
}

// ---- data source tests ------------------------------------------------------

func newTestArrayNtpDataSource(t *testing.T, ms *testmock.MockServer) *arrayNtpDataSource {
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
	return &arrayNtpDataSource{client: c}
}

func arrayNtpDSSchema(t *testing.T) datasource.SchemaResponse {
	t.Helper()
	d := &arrayNtpDataSource{}
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	return resp
}

func buildArrayNtpDSType() tftypes.Object {
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":          tftypes.String,
		"ntp_servers": tftypes.List{ElementType: tftypes.String},
	}}
}

// TestArrayNtpDataSource verifies data source reads current NTP config.
func TestArrayNtpDataSource(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterArrayAdminHandlers(ms.Mux)

	// Set up NTP config via client.
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
	servers := []string{"time.cloudflare.com"}
	if _, err := c.PatchArrayNtp(context.Background(), client.ArrayNtpPatch{
		NtpServers: &servers,
	}); err != nil {
		t.Fatalf("PatchArrayNtp: %v", err)
	}

	d := newTestArrayNtpDataSource(t, ms)
	s := arrayNtpDSSchema(t).Schema

	dsType := buildArrayNtpDSType()
	cfg := map[string]tftypes.Value{
		"id":          tftypes.NewValue(tftypes.String, nil),
		"ntp_servers": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
	}

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(dsType, nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(dsType, cfg), Schema: s},
	}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("DataSource Read returned error: %s", readResp.Diagnostics)
	}

	var model arrayNtpDataSourceModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected non-empty ID")
	}

	var ntpServers []string
	if diags := model.NtpServers.ElementsAs(context.Background(), &ntpServers, false); diags.HasError() {
		t.Fatalf("Get ntp_servers: %s", diags)
	}
	if len(ntpServers) != 1 || ntpServers[0] != "time.cloudflare.com" {
		t.Errorf("expected ntp_servers=[time.cloudflare.com], got %v", ntpServers)
	}

	// Ensure the unused attr import is referenced.
	_ = attr.Value(nil)
}

// TestUnit_ArrayNtp_Lifecycle exercises the full Create->Read->Update->Read->Delete sequence.
func TestUnit_ArrayNtp_Lifecycle(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterArrayAdminHandlers(ms.Mux)

	r := newTestArrayNtpResource(t, ms)
	s := arrayNtpResourceSchema(t).Schema

	// Step 1: Create.
	createPlan := arrayNtpPlanWith(t, []string{"ntp1.example.com"})
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildArrayNtpType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}
	var createModel arrayNtpModel
	if diags := createResp.State.Get(context.Background(), &createModel); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}
	var createServers []string
	createModel.NtpServers.ElementsAs(context.Background(), &createServers, false)
	if len(createServers) != 1 || createServers[0] != "ntp1.example.com" {
		t.Errorf("Create: expected ntp_servers=[ntp1.example.com], got %v", createServers)
	}

	// Step 2: Read post-create.
	readResp1 := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp1)
	if readResp1.Diagnostics.HasError() {
		t.Fatalf("Read post-create: %s", readResp1.Diagnostics)
	}
	var readModel1 arrayNtpModel
	if diags := readResp1.State.Get(context.Background(), &readModel1); diags.HasError() {
		t.Fatalf("Get read1 state: %s", diags)
	}
	if readModel1.ID.IsNull() || readModel1.ID.ValueString() == "" {
		t.Error("Read1: expected non-empty ID")
	}

	// Step 3: Update NTP servers.
	updatePlan := arrayNtpPlanWith(t, []string{"ntp1.example.com", "ntp2.example.com"})
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildArrayNtpType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlan,
		State: readResp1.State,
	}, updateResp)
	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update: %s", updateResp.Diagnostics)
	}
	var updateModel arrayNtpModel
	if diags := updateResp.State.Get(context.Background(), &updateModel); diags.HasError() {
		t.Fatalf("Get update state: %s", diags)
	}
	var updateServers []string
	updateModel.NtpServers.ElementsAs(context.Background(), &updateServers, false)
	if len(updateServers) != 2 {
		t.Errorf("Update: expected 2 NTP servers, got %d", len(updateServers))
	}

	// Step 4: Read post-update.
	readResp2 := &resource.ReadResponse{State: updateResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: updateResp.State}, readResp2)
	if readResp2.Diagnostics.HasError() {
		t.Fatalf("Read post-update: %s", readResp2.Diagnostics)
	}
	var readModel2 arrayNtpModel
	if diags := readResp2.State.Get(context.Background(), &readModel2); diags.HasError() {
		t.Fatalf("Get read2 state: %s", diags)
	}
	var readServers2 []string
	readModel2.NtpServers.ElementsAs(context.Background(), &readServers2, false)
	if len(readServers2) != 2 {
		t.Errorf("Read2: expected 2 NTP servers, got %d", len(readServers2))
	}

	// Step 5: Delete (reset to defaults).
	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: readResp2.State}, deleteResp)
	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete: %s", deleteResp.Diagnostics)
	}
}

// TestUnit_ArrayNtp_ImportIdempotency verifies ImportState->Read produces state matching original Create.
func TestUnit_ArrayNtp_ImportIdempotency(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterArrayAdminHandlers(ms.Mux)

	r := newTestArrayNtpResource(t, ms)
	s := arrayNtpResourceSchema(t).Schema

	// Create.
	createPlan := arrayNtpPlanWith(t, []string{"time.cloudflare.com"})
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildArrayNtpType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}
	var createModel arrayNtpModel
	if diags := createResp.State.Get(context.Background(), &createModel); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}

	// ImportState using "default" singleton ID.
	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildArrayNtpType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "default"}, importResp)
	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState: %s", importResp.Diagnostics)
	}

	// Read to populate full state.
	readResp := &resource.ReadResponse{State: importResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: importResp.State}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read post-import: %s", readResp.Diagnostics)
	}
	var importedModel arrayNtpModel
	if diags := readResp.State.Get(context.Background(), &importedModel); diags.HasError() {
		t.Fatalf("Get imported state: %s", diags)
	}

	// Verify 0-diff.
	var createServers, importServers []string
	createModel.NtpServers.ElementsAs(context.Background(), &createServers, false)
	importedModel.NtpServers.ElementsAs(context.Background(), &importServers, false)
	if len(createServers) != len(importServers) {
		t.Errorf("ntp_servers count mismatch: create=%d import=%d", len(createServers), len(importServers))
	}
	if len(createServers) > 0 && len(importServers) > 0 && createServers[0] != importServers[0] {
		t.Errorf("ntp_servers[0] mismatch: create=%s import=%s", createServers[0], importServers[0])
	}
}

// TestUnit_ArrayNTP_PlanModifiers verifies all UseStateForUnknown plan modifiers
// in the array_ntp resource schema.
func TestUnit_ArrayNTP_PlanModifiers(t *testing.T) {
	s := arrayNtpResourceSchema(t).Schema

	// id — UseStateForUnknown
	idAttr, ok := s.Attributes["id"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("id attribute not found or wrong type")
	}
	if len(idAttr.PlanModifiers) == 0 {
		t.Error("expected UseStateForUnknown plan modifier on id attribute")
	}
}
