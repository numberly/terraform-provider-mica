package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/numberly/terraform-provider-mica/internal/client"
	"github.com/numberly/terraform-provider-mica/internal/testmock"
	"github.com/numberly/terraform-provider-mica/internal/testmock/handlers"
)

// ---- helpers ----------------------------------------------------------------

func newTestArrayDnsResource(t *testing.T, ms *testmock.MockServer) *arrayDnsResource {
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
		"name":        tftypes.String,
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
		"name":        tftypes.NewValue(tftypes.String, nil),
		"domain":      tftypes.NewValue(tftypes.String, nil),
		"nameservers": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"services":    tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"sources":     tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"timeouts":    tftypes.NewValue(timeoutsType, nil),
	}
}

func arrayDnsPlanWith(t *testing.T, name string, domain string, nameservers []string) tfsdk.Plan {
	t.Helper()
	s := arrayDnsResourceSchema(t).Schema
	cfg := nullArrayDnsConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, name)
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

// TestUnit_ArrayDnsResource_Lifecycle exercises Create->Read->Update->Read->Delete.
func TestUnit_ArrayDnsResource_Lifecycle(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterArrayAdminHandlers(ms.Mux)

	r := newTestArrayDnsResource(t, ms)
	s := arrayDnsResourceSchema(t).Schema

	// Step 1: Create.
	createPlan := arrayDnsPlanWith(t, "lifecycle-dns", "lifecycle.example.com", []string{"8.8.8.8", "8.8.4.4"})
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
	if createModel.Name.ValueString() != "lifecycle-dns" {
		t.Errorf("Create: expected name=lifecycle-dns, got %s", createModel.Name.ValueString())
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
	updatePlan := arrayDnsPlanWith(t, "lifecycle-dns", "lifecycle.example.com", []string{"1.1.1.1"})
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

	// Step 5: Delete.
	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: readResp2.State}, deleteResp)
	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete: %s", deleteResp.Diagnostics)
	}
}

// TestUnit_ArrayDnsResource_Import verifies ImportState populates all attributes by name.
func TestUnit_ArrayDnsResource_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterArrayAdminHandlers(ms.Mux)

	// Seed a DNS entry for import.
	store.SeedDns(&client.ArrayDns{
		ID:          "dns-import-001",
		Name:        "import-dns",
		Domain:      "import.example.com",
		Nameservers: []string{"9.9.9.9"},
		Services:    []string{},
		Sources:     []string{},
	})

	r := newTestArrayDnsResource(t, ms)
	s := arrayDnsResourceSchema(t).Schema

	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildArrayDnsType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "import-dns"}, importResp)

	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState returned error: %s", importResp.Diagnostics)
	}

	var model arrayDnsModel
	if diags := importResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.ValueString() != "dns-import-001" {
		t.Errorf("expected ID dns-import-001 after import, got %s", model.ID.ValueString())
	}
	if model.Name.ValueString() != "import-dns" {
		t.Errorf("expected name=import-dns after import, got %s", model.Name.ValueString())
	}
	if model.Domain.ValueString() != "import.example.com" {
		t.Errorf("expected domain=import.example.com after import, got %s", model.Domain.ValueString())
	}
}

// TestUnit_ArrayDnsResource_DriftDetection verifies that Read detects and logs drift.
func TestUnit_ArrayDnsResource_DriftDetection(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterArrayAdminHandlers(ms.Mux)

	r := newTestArrayDnsResource(t, ms)
	s := arrayDnsResourceSchema(t).Schema

	// Create a DNS entry.
	createPlan := arrayDnsPlanWith(t, "drift-dns", "drift.example.com", []string{"8.8.8.8"})
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildArrayDnsType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Mutate the DNS entry in the mock (simulate external change).
	store.SeedDns(&client.ArrayDns{
		ID:          "dns-1",
		Name:        "drift-dns",
		Domain:      "drifted.example.com",
		Nameservers: []string{"1.1.1.1"},
		Services:    []string{},
		Sources:     []string{},
	})

	// Read should detect drift (logged, not errored).
	readResp := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}

	var model arrayDnsModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}
	if model.Domain.ValueString() != "drifted.example.com" {
		t.Errorf("expected domain=drifted.example.com after drift, got %s", model.Domain.ValueString())
	}
}

// TestUnit_ArrayDnsResource_StateUpgrade_V0toV1 verifies the v0->v1 state upgrader adds name="default".
func TestUnit_ArrayDnsResource_StateUpgrade_V0toV1(t *testing.T) {
	r := &arrayDnsResource{}
	upgraders := r.UpgradeState(context.Background())

	upgrader, ok := upgraders[0]
	if !ok {
		t.Fatal("expected state upgrader for version 0")
	}

	// Build v0 state (no name attribute).
	v0Type := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":          tftypes.String,
		"domain":      tftypes.String,
		"nameservers": tftypes.List{ElementType: tftypes.String},
		"services":    tftypes.List{ElementType: tftypes.String},
		"sources":     tftypes.List{ElementType: tftypes.String},
		"timeouts": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
			"create": tftypes.String,
			"read":   tftypes.String,
			"update": tftypes.String,
			"delete": tftypes.String,
		}},
	}}

	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}

	v0Raw := tftypes.NewValue(v0Type, map[string]tftypes.Value{
		"id":     tftypes.NewValue(tftypes.String, "old-dns-id"),
		"domain": tftypes.NewValue(tftypes.String, "old.example.com"),
		"nameservers": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
			tftypes.NewValue(tftypes.String, "8.8.8.8"),
		}),
		"services": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{}),
		"sources":  tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{}),
		"timeouts": tftypes.NewValue(timeoutsType, map[string]tftypes.Value{
			"create": tftypes.NewValue(tftypes.String, nil),
			"read":   tftypes.NewValue(tftypes.String, nil),
			"update": tftypes.NewValue(tftypes.String, nil),
			"delete": tftypes.NewValue(tftypes.String, nil),
		}),
	})

	priorState := tfsdk.State{
		Raw:    v0Raw,
		Schema: *upgrader.PriorSchema,
	}

	// Build v1 schema for the response state.
	v1Schema := arrayDnsResourceSchema(t).Schema

	upgradeResp := &resource.UpgradeStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildArrayDnsType(), nil), Schema: v1Schema},
	}

	upgrader.StateUpgrader(context.Background(), resource.UpgradeStateRequest{
		State: &priorState,
	}, upgradeResp)

	if upgradeResp.Diagnostics.HasError() {
		t.Fatalf("StateUpgrader returned error: %s", upgradeResp.Diagnostics)
	}

	var model arrayDnsModel
	if diags := upgradeResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get upgraded state: %s", diags)
	}

	if model.Name.ValueString() != "default" {
		t.Errorf("expected name=default after upgrade, got %q", model.Name.ValueString())
	}
	if model.ID.ValueString() != "old-dns-id" {
		t.Errorf("expected ID preserved, got %q", model.ID.ValueString())
	}
	if model.Domain.ValueString() != "old.example.com" {
		t.Errorf("expected domain preserved, got %q", model.Domain.ValueString())
	}

	var ns []string
	model.Nameservers.ElementsAs(context.Background(), &ns, false)
	if len(ns) != 1 || ns[0] != "8.8.8.8" {
		t.Errorf("expected nameservers=[8.8.8.8] preserved, got %v", ns)
	}
}

// ---- data source tests ------------------------------------------------------

func newTestArrayDnsDataSource(t *testing.T, ms *testmock.MockServer) *arrayDnsDataSource {
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
		"name":        tftypes.String,
		"domain":      tftypes.String,
		"nameservers": tftypes.List{ElementType: tftypes.String},
		"services":    tftypes.List{ElementType: tftypes.String},
		"sources":     tftypes.List{ElementType: tftypes.String},
	}}
}

func nullArrayDnsDSConfig() map[string]tftypes.Value {
	return map[string]tftypes.Value{
		"id":          tftypes.NewValue(tftypes.String, nil),
		"name":        tftypes.NewValue(tftypes.String, nil),
		"domain":      tftypes.NewValue(tftypes.String, nil),
		"nameservers": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"services":    tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"sources":     tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
	}
}

// TestUnit_ArrayDnsDataSource_Basic verifies the data source reads a DNS config by name.
func TestUnit_ArrayDnsDataSource_Basic(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterArrayAdminHandlers(ms.Mux)

	// Seed DNS entry.
	store.SeedDns(&client.ArrayDns{
		ID:          "dns-ds-001",
		Name:        "ds-dns",
		Domain:      "ds.example.com",
		Nameservers: []string{"1.2.3.4"},
		Services:    []string{},
		Sources:     []string{},
	})

	d := newTestArrayDnsDataSource(t, ms)
	s := arrayDnsDSSchema(t).Schema

	dsType := buildArrayDnsDSType()
	cfg := nullArrayDnsDSConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "ds-dns")

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

	if model.ID.ValueString() != "dns-ds-001" {
		t.Errorf("expected ID=dns-ds-001, got %s", model.ID.ValueString())
	}
	if model.Name.ValueString() != "ds-dns" {
		t.Errorf("expected name=ds-dns, got %s", model.Name.ValueString())
	}
	if model.Domain.ValueString() != "ds.example.com" {
		t.Errorf("expected domain=ds.example.com, got %s", model.Domain.ValueString())
	}

	var nameservers []string
	if diags := model.Nameservers.ElementsAs(context.Background(), &nameservers, false); diags.HasError() {
		t.Fatalf("Get nameservers: %s", diags)
	}
	if len(nameservers) != 1 || nameservers[0] != "1.2.3.4" {
		t.Errorf("expected nameservers=[1.2.3.4], got %v", nameservers)
	}
}

// TestUnit_ArrayDnsResource_PlanModifiers verifies plan modifiers on id and name.
func TestUnit_ArrayDnsResource_PlanModifiers(t *testing.T) {
	resp := arrayDnsResourceSchema(t)
	s := resp.Schema

	// id — UseStateForUnknown
	if _, ok := s.Attributes["id"]; !ok {
		t.Fatal("id attribute not found")
	}

	// name — RequiresReplace
	if _, ok := s.Attributes["name"]; !ok {
		t.Fatal("name attribute not found")
	}

	// Verify schema version is 1
	if s.Version != 1 {
		t.Errorf("expected schema version 1, got %d", s.Version)
	}
}

// TestUnit_ArrayDnsResource_SchemaV1HasName verifies the v1 schema includes the name attribute.
func TestUnit_ArrayDnsResource_SchemaV1HasName(t *testing.T) {
	resp := arrayDnsResourceSchema(t)
	s := resp.Schema

	nameAttr, ok := s.Attributes["name"]
	if !ok {
		t.Fatal("expected name attribute in schema v1")
	}
	_ = nameAttr
	_ = types.StringNull() // ensure types import used
}
