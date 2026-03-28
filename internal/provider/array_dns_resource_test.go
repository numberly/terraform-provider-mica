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
	"github.com/numberly/opentofu-provider-flashblade/internal/client"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock/handlers"
)

// ---- helpers ----------------------------------------------------------------

func newTestArrayDnsResource(t *testing.T, ms *testmock.MockServer) *arrayDnsResource {
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
	return &arrayDnsResource{client: c}
}

func arrayDnsResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &arrayDnsResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

func buildArrayDnsType() tftypes.Object {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":          tftypes.String,
		"domain":      tftypes.String,
		"nameservers": tftypes.List{ElementType: tftypes.String},
		"services":    tftypes.List{ElementType: tftypes.String},
		"sources":     tftypes.List{ElementType: tftypes.String},
		"timeouts":    timeoutsType,
	}}
}

func nullArrayDnsConfig() map[string]tftypes.Value {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return map[string]tftypes.Value{
		"id":          tftypes.NewValue(tftypes.String, nil),
		"domain":      tftypes.NewValue(tftypes.String, nil),
		"nameservers": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"services":    tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"sources":     tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"timeouts":    tftypes.NewValue(timeoutsType, nil),
	}
}

func arrayDnsPlanWith(t *testing.T, domain string, nameservers []string) tfsdk.Plan {
	t.Helper()
	s := arrayDnsResourceSchema(t).Schema
	cfg := nullArrayDnsConfig()
	cfg["domain"] = tftypes.NewValue(tftypes.String, domain)
	nsValues := make([]tftypes.Value, len(nameservers))
	for i, ns := range nameservers {
		nsValues[i] = tftypes.NewValue(tftypes.String, ns)
	}
	cfg["nameservers"] = tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nsValues)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildArrayDnsType(), cfg),
		Schema: s,
	}
}

// ---- resource tests ---------------------------------------------------------

// TestArrayDnsResource_Create verifies that Create sets nameservers and domain.
func TestArrayDnsResource_Create(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterArrayAdminHandlers(ms.Mux)

	r := newTestArrayDnsResource(t, ms)
	s := arrayDnsResourceSchema(t).Schema

	plan := arrayDnsPlanWith(t, "example.com", []string{"8.8.8.8", "8.8.4.4"})
	resp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildArrayDnsType(), nil), Schema: s},
	}

	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Create returned error: %s", resp.Diagnostics)
	}

	var model arrayDnsModel
	if diags := resp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected non-empty ID after Create")
	}
	if model.Domain.ValueString() != "example.com" {
		t.Errorf("expected domain=example.com, got %s", model.Domain.ValueString())
	}

	var nameservers []string
	if diags := model.Nameservers.ElementsAs(context.Background(), &nameservers, false); diags.HasError() {
		t.Fatalf("Get nameservers: %s", diags)
	}
	if len(nameservers) != 2 || nameservers[0] != "8.8.8.8" || nameservers[1] != "8.8.4.4" {
		t.Errorf("expected nameservers=[8.8.8.8, 8.8.4.4], got %v", nameservers)
	}
}

// TestArrayDnsResource_Update verifies that Update changes nameservers via PATCH.
func TestArrayDnsResource_Update(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterArrayAdminHandlers(ms.Mux)

	r := newTestArrayDnsResource(t, ms)
	s := arrayDnsResourceSchema(t).Schema

	// Create first.
	createPlan := arrayDnsPlanWith(t, "example.com", []string{"8.8.8.8", "8.8.4.4"})
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildArrayDnsType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Update to single nameserver.
	updatePlan := arrayDnsPlanWith(t, "example.com", []string{"1.1.1.1"})
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildArrayDnsType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlan,
		State: createResp.State,
	}, updateResp)

	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update returned error: %s", updateResp.Diagnostics)
	}

	var model arrayDnsModel
	if diags := updateResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	var nameservers []string
	if diags := model.Nameservers.ElementsAs(context.Background(), &nameservers, false); diags.HasError() {
		t.Fatalf("Get nameservers: %s", diags)
	}
	if len(nameservers) != 1 || nameservers[0] != "1.1.1.1" {
		t.Errorf("expected nameservers=[1.1.1.1], got %v", nameservers)
	}
}

// TestArrayDnsResource_Delete verifies that Delete resets DNS to defaults.
func TestArrayDnsResource_Delete(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterArrayAdminHandlers(ms.Mux)

	r := newTestArrayDnsResource(t, ms)
	s := arrayDnsResourceSchema(t).Schema

	// Create first.
	createPlan := arrayDnsPlanWith(t, "example.com", []string{"8.8.8.8"})
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildArrayDnsType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Delete resets config.
	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: createResp.State}, deleteResp)

	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete returned error: %s", deleteResp.Diagnostics)
	}

	// Verify reset.
	dns, err := r.client.GetArrayDns(context.Background())
	if err != nil {
		t.Fatalf("GetArrayDns after delete: %v", err)
	}
	if dns.Domain != "" {
		t.Errorf("expected domain to be reset, got %q", dns.Domain)
	}
	if len(dns.Nameservers) != 0 {
		t.Errorf("expected nameservers to be empty after delete, got %v", dns.Nameservers)
	}
}

// TestArrayDnsResource_Import verifies ImportState populates all attributes.
func TestArrayDnsResource_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterArrayAdminHandlers(ms.Mux)

	r := newTestArrayDnsResource(t, ms)
	s := arrayDnsResourceSchema(t).Schema

	// Set up some DNS config via direct client call.
	domain := "import.example.com"
	ns := []string{"9.9.9.9"}
	_, err := r.client.PatchArrayDns(context.Background(), client.ArrayDnsPatch{
		Domain:      &domain,
		Nameservers: &ns,
	})
	if err != nil {
		t.Fatalf("PatchArrayDns: %v", err)
	}

	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildArrayDnsType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "default"}, importResp)

	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState returned error: %s", importResp.Diagnostics)
	}

	var model arrayDnsModel
	if diags := importResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected non-empty ID after import")
	}
	if model.Domain.ValueString() != "import.example.com" {
		t.Errorf("expected domain=import.example.com after import, got %s", model.Domain.ValueString())
	}
}

// ---- data source tests ------------------------------------------------------

func newTestArrayDnsDataSource(t *testing.T, ms *testmock.MockServer) *arrayDnsDataSource {
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
	return &arrayDnsDataSource{client: c}
}

func arrayDnsDSSchema(t *testing.T) datasource.SchemaResponse {
	t.Helper()
	d := &arrayDnsDataSource{}
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	return resp
}

func buildArrayDnsDSType() tftypes.Object {
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":          tftypes.String,
		"domain":      tftypes.String,
		"nameservers": tftypes.List{ElementType: tftypes.String},
		"services":    tftypes.List{ElementType: tftypes.String},
		"sources":     tftypes.List{ElementType: tftypes.String},
	}}
}

// TestArrayDnsDataSource verifies data source reads current DNS config.
func TestArrayDnsDataSource(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterArrayAdminHandlers(ms.Mux)

	// Set up DNS config via client.
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
	domain := "ds.example.com"
	ns := []string{"1.2.3.4"}
	if _, err := c.PatchArrayDns(context.Background(), client.ArrayDnsPatch{
		Domain:      &domain,
		Nameservers: &ns,
	}); err != nil {
		t.Fatalf("PatchArrayDns: %v", err)
	}

	d := newTestArrayDnsDataSource(t, ms)
	s := arrayDnsDSSchema(t).Schema

	dsType := buildArrayDnsDSType()
	cfg := map[string]tftypes.Value{
		"id":          tftypes.NewValue(tftypes.String, nil),
		"domain":      tftypes.NewValue(tftypes.String, nil),
		"nameservers": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"services":    tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"sources":     tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
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

	var model arrayDnsDataSourceModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Domain.ValueString() != "ds.example.com" {
		t.Errorf("expected domain=ds.example.com, got %s", model.Domain.ValueString())
	}
	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected non-empty ID")
	}
	var nameservers []string
	if diags := model.Nameservers.ElementsAs(context.Background(), &nameservers, false); diags.HasError() {
		t.Fatalf("Get nameservers: %s", diags)
	}
	if len(nameservers) != 1 || nameservers[0] != "1.2.3.4" {
		t.Errorf("expected nameservers=[1.2.3.4], got %v", nameservers)
	}

	// Ensure the unused tftypes/attr imports are referenced.
	_ = attr.Value(nil)
}

// TestUnit_ArrayDns_Lifecycle exercises the full Create->Read->Update->Read->Delete sequence.
// ArrayDns is a singleton: Create=PATCH, Delete=PATCH(reset).
func TestUnit_ArrayDns_Lifecycle(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterArrayAdminHandlers(ms.Mux)

	r := newTestArrayDnsResource(t, ms)
	s := arrayDnsResourceSchema(t).Schema

	// Step 1: Create (configure DNS).
	createPlan := arrayDnsPlanWith(t, "lifecycle.example.com", []string{"8.8.8.8", "8.8.4.4"})
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildArrayDnsType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}
	var createModel arrayDnsModel
	if diags := createResp.State.Get(context.Background(), &createModel); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}
	if createModel.Domain.ValueString() != "lifecycle.example.com" {
		t.Errorf("Create: expected domain=lifecycle.example.com, got %s", createModel.Domain.ValueString())
	}

	// Step 2: Read post-create.
	readResp1 := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp1)
	if readResp1.Diagnostics.HasError() {
		t.Fatalf("Read post-create: %s", readResp1.Diagnostics)
	}
	var readModel1 arrayDnsModel
	if diags := readResp1.State.Get(context.Background(), &readModel1); diags.HasError() {
		t.Fatalf("Get read1 state: %s", diags)
	}
	var ns1 []string
	readModel1.Nameservers.ElementsAs(context.Background(), &ns1, false)
	if len(ns1) != 2 {
		t.Errorf("Read1: expected 2 nameservers, got %d", len(ns1))
	}

	// Step 3: Update nameservers.
	updatePlan := arrayDnsPlanWith(t, "lifecycle.example.com", []string{"1.1.1.1"})
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildArrayDnsType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlan,
		State: readResp1.State,
	}, updateResp)
	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update: %s", updateResp.Diagnostics)
	}
	var updateModel arrayDnsModel
	if diags := updateResp.State.Get(context.Background(), &updateModel); diags.HasError() {
		t.Fatalf("Get update state: %s", diags)
	}
	var ns2 []string
	updateModel.Nameservers.ElementsAs(context.Background(), &ns2, false)
	if len(ns2) != 1 || ns2[0] != "1.1.1.1" {
		t.Errorf("Update: expected nameservers=[1.1.1.1], got %v", ns2)
	}

	// Step 4: Read post-update.
	readResp2 := &resource.ReadResponse{State: updateResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: updateResp.State}, readResp2)
	if readResp2.Diagnostics.HasError() {
		t.Fatalf("Read post-update: %s", readResp2.Diagnostics)
	}
	var readModel2 arrayDnsModel
	if diags := readResp2.State.Get(context.Background(), &readModel2); diags.HasError() {
		t.Fatalf("Get read2 state: %s", diags)
	}
	var ns3 []string
	readModel2.Nameservers.ElementsAs(context.Background(), &ns3, false)
	if len(ns3) != 1 || ns3[0] != "1.1.1.1" {
		t.Errorf("Read2: expected nameservers=[1.1.1.1], got %v", ns3)
	}

	// Step 5: Delete (reset to defaults).
	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: readResp2.State}, deleteResp)
	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete: %s", deleteResp.Diagnostics)
	}
}

// TestUnit_ArrayDns_ImportIdempotency verifies ImportState->Read produces state matching original Create.
func TestUnit_ArrayDns_ImportIdempotency(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterArrayAdminHandlers(ms.Mux)

	r := newTestArrayDnsResource(t, ms)
	s := arrayDnsResourceSchema(t).Schema

	// Create (configure DNS).
	createPlan := arrayDnsPlanWith(t, "idempotent.example.com", []string{"8.8.8.8"})
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildArrayDnsType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}
	var createModel arrayDnsModel
	if diags := createResp.State.Get(context.Background(), &createModel); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}

	// ImportState using "default" singleton ID.
	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildArrayDnsType(), nil), Schema: s},
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
	var importedModel arrayDnsModel
	if diags := readResp.State.Get(context.Background(), &importedModel); diags.HasError() {
		t.Fatalf("Get imported state: %s", diags)
	}

	// Verify 0-diff.
	if importedModel.Domain.ValueString() != createModel.Domain.ValueString() {
		t.Errorf("domain mismatch: create=%s import=%s", createModel.Domain.ValueString(), importedModel.Domain.ValueString())
	}
	var createNS, importNS []string
	createModel.Nameservers.ElementsAs(context.Background(), &createNS, false)
	importedModel.Nameservers.ElementsAs(context.Background(), &importNS, false)
	if len(createNS) != len(importNS) {
		t.Errorf("nameservers count mismatch: create=%d import=%d", len(createNS), len(importNS))
	}
}

// TestUnit_ArrayDNS_PlanModifiers verifies all UseStateForUnknown plan modifiers
// in the array_dns resource schema.
func TestUnit_ArrayDNS_PlanModifiers(t *testing.T) {
	s := arrayDnsResourceSchema(t).Schema

	// id — UseStateForUnknown
	idAttr, ok := s.Attributes["id"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("id attribute not found or wrong type")
	}
	if len(idAttr.PlanModifiers) == 0 {
		t.Error("expected UseStateForUnknown plan modifier on id attribute")
	}
}
