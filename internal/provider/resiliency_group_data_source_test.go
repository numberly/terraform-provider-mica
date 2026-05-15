package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/numberly/terraform-provider-mica/internal/client"
	"github.com/numberly/terraform-provider-mica/internal/testmock"
	"github.com/numberly/terraform-provider-mica/internal/testmock/handlers"
)

// ---- helpers ----------------------------------------------------------------

// newTestResiliencyGroupDataSource creates a resiliencyGroupDataSource wired to
// the given mock server.
func newTestResiliencyGroupDataSource(t *testing.T, ms *testmock.MockServer) *resiliencyGroupDataSource {
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
	return &resiliencyGroupDataSource{client: c}
}

// resiliencyGroupDSSchema returns the schema for the resiliency group data source.
func resiliencyGroupDSSchema(t *testing.T) datasource.SchemaResponse {
	t.Helper()
	d := &resiliencyGroupDataSource{}
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	return resp
}

// buildResiliencyGroupDSType returns the tftypes.Object for the data source schema.
func buildResiliencyGroupDSType() tftypes.Object {
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":             tftypes.String,
		"name":           tftypes.String,
		"status":         tftypes.String,
		"status_details": tftypes.String,
	}}
}

// nullResiliencyGroupDSConfig returns a base config map with all attributes null.
func nullResiliencyGroupDSConfig() map[string]tftypes.Value {
	return map[string]tftypes.Value{
		"id":             tftypes.NewValue(tftypes.String, nil),
		"name":           tftypes.NewValue(tftypes.String, nil),
		"status":         tftypes.NewValue(tftypes.String, nil),
		"status_details": tftypes.NewValue(tftypes.String, nil),
	}
}

// ---- tests ------------------------------------------------------------------

// TestUnit_ResiliencyGroupDataSource_Basic verifies the data source reads a
// seeded resiliency group by name and exposes its computed fields.
func TestUnit_ResiliencyGroupDataSource_Basic(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterResiliencyGroupHandlers(ms.Mux)
	store.Seed(&client.ResiliencyGroup{
		ID:            "rg-001",
		Name:          "rg0",
		Status:        "unhealthy",
		StatusDetails: "blade pulled",
	})

	d := newTestResiliencyGroupDataSource(t, ms)
	s := resiliencyGroupDSSchema(t).Schema

	cfg := nullResiliencyGroupDSConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "rg0")

	objType := buildResiliencyGroupDSType()
	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(objType, cfg), Schema: s},
	}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("DataSource Read returned error: %s", readResp.Diagnostics)
	}

	var model resiliencyGroupDataSourceModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Name.ValueString() != "rg0" {
		t.Errorf("expected name=rg0, got %s", model.Name.ValueString())
	}
	if model.ID.ValueString() != "rg-001" {
		t.Errorf("expected id=rg-001, got %s", model.ID.ValueString())
	}
	if model.Status.ValueString() != "unhealthy" {
		t.Errorf("expected status=unhealthy, got %s", model.Status.ValueString())
	}
	if model.StatusDetails.ValueString() != "blade pulled" {
		t.Errorf("expected status_details='blade pulled', got %s", model.StatusDetails.ValueString())
	}
}

// TestUnit_ResiliencyGroupDataSource_NotFound verifies error diagnostic when
// the resiliency group does not exist.
func TestUnit_ResiliencyGroupDataSource_NotFound(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterResiliencyGroupHandlers(ms.Mux)

	d := newTestResiliencyGroupDataSource(t, ms)
	s := resiliencyGroupDSSchema(t).Schema

	cfg := nullResiliencyGroupDSConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "ghost-rg")

	objType := buildResiliencyGroupDSType()
	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(objType, cfg), Schema: s},
	}, readResp)

	if !readResp.Diagnostics.HasError() {
		t.Fatal("expected error diagnostic for not-found resiliency group, got none")
	}

	found := false
	for _, diag := range readResp.Diagnostics {
		if diag.Summary() == "Resiliency group not found" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'Resiliency group not found' diagnostic, got: %s", readResp.Diagnostics)
	}
}

// TestUnit_ResiliencyGroupDataSource_Schema verifies that `name` is Required
// and the rest of the attributes are Computed.
func TestUnit_ResiliencyGroupDataSource_Schema(t *testing.T) {
	resp := resiliencyGroupDSSchema(t)
	s := resp.Schema

	attr, ok := s.Attributes["name"]
	if !ok {
		t.Fatal("name attribute not found in schema")
	}
	if !attr.IsRequired() {
		t.Error("name should be Required")
	}

	computedAttrs := []string{"id", "status", "status_details"}
	for _, name := range computedAttrs {
		a, ok := s.Attributes[name]
		if !ok {
			t.Errorf("attribute %q not found in schema", name)
			continue
		}
		if !a.IsComputed() {
			t.Errorf("attribute %q should be Computed", name)
		}
	}
}
