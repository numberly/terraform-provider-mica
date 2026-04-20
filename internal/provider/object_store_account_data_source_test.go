package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/numberly/opentofu-provider-flashblade/internal/client"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock/handlers"
)

// ---- helpers ----------------------------------------------------------------

// newTestOSADataSource creates an objectStoreAccountDataSource wired to the given mock server.
func newTestOSADataSource(t *testing.T, ms *testmock.MockServer) *objectStoreAccountDataSource {
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
	return &objectStoreAccountDataSource{client: c}
}

// osaDataSourceSchema returns the schema for the object store account data source.
func osaDataSourceSchema(t *testing.T) datasource.SchemaResponse {
	t.Helper()
	d := &objectStoreAccountDataSource{}
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	return resp
}

// buildObjectStoreAccountDSType returns the tftypes.Object for the object store account data source schema.
func buildObjectStoreAccountDSType() tftypes.Object {
	spaceType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"data_reduction":      tftypes.Number,
		"snapshots":           tftypes.Number,
		"total_physical":      tftypes.Number,
		"unique":              tftypes.Number,
		"virtual":             tftypes.Number,
		"snapshots_effective": tftypes.Number,
	}}
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":                 tftypes.String,
		"name":               tftypes.String,
		"created":            tftypes.Number,
		"quota_limit":        tftypes.Number,
		"hard_limit_enabled": tftypes.Bool,
		"object_count":       tftypes.Number,
		"space":              spaceType,
	}}
}

// nullOSADSConfig returns a base config map with all data source attributes null.
func nullOSADSConfig() map[string]tftypes.Value {
	spaceType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"data_reduction":      tftypes.Number,
		"snapshots":           tftypes.Number,
		"total_physical":      tftypes.Number,
		"unique":              tftypes.Number,
		"virtual":             tftypes.Number,
		"snapshots_effective": tftypes.Number,
	}}
	return map[string]tftypes.Value{
		"id":                 tftypes.NewValue(tftypes.String, nil),
		"name":               tftypes.NewValue(tftypes.String, nil),
		"created":            tftypes.NewValue(tftypes.Number, nil),
		"quota_limit":        tftypes.NewValue(tftypes.Number, nil),
		"hard_limit_enabled": tftypes.NewValue(tftypes.Bool, nil),
		"object_count":       tftypes.NewValue(tftypes.Number, nil),
		"space":              tftypes.NewValue(spaceType, nil),
	}
}

// ---- tests ------------------------------------------------------------------

// TestUnit_ObjectStoreAccountDataSource verifies data source reads account by name and returns all attributes.
func TestUnit_ObjectStoreAccountDataSource(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterObjectStoreAccountHandlers(ms.Mux)

	// Create an account via the resource client so the data source can find it.
	c, err := client.NewClient(context.Background(), client.Config{
		Endpoint:           ms.URL(),
		APIToken:           "test-token",
		InsecureSkipVerify: true,
		MaxRetries:         1,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	_, err = c.PostObjectStoreAccount(context.Background(), "ds-test-account", client.ObjectStoreAccountPost{
		QuotaLimit: "21474836480",
	})
	if err != nil {
		t.Fatalf("PostObjectStoreAccount: %v", err)
	}

	d := newTestOSADataSource(t, ms)
	s := osaDataSourceSchema(t).Schema

	cfg := nullOSADSConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "ds-test-account")

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildObjectStoreAccountDSType(), nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(buildObjectStoreAccountDSType(), cfg), Schema: s},
	}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("DataSource Read returned error: %s", readResp.Diagnostics)
	}

	var model objectStoreAccountDataSourceModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Name.ValueString() != "ds-test-account" {
		t.Errorf("expected name=ds-test-account, got %s", model.Name.ValueString())
	}
	if model.QuotaLimit.ValueInt64() != 21474836480 {
		t.Errorf("expected quota_limit=21474836480, got %d", model.QuotaLimit.ValueInt64())
	}
	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected ID to be populated")
	}
}

// TestUnit_ObjectStoreAccountDataSource_NotFound verifies that a missing account returns an error diagnostic.
func TestUnit_ObjectStoreAccountDataSource_NotFound(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterObjectStoreAccountHandlers(ms.Mux)

	d := newTestOSADataSource(t, ms)
	s := osaDataSourceSchema(t).Schema

	cfg := nullOSADSConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "nonexistent-account")

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildObjectStoreAccountDSType(), nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(buildObjectStoreAccountDSType(), cfg), Schema: s},
	}, readResp)

	if !readResp.Diagnostics.HasError() {
		t.Error("expected error diagnostic for not-found account, got none")
	}

	found := false
	for _, diag := range readResp.Diagnostics {
		if diag.Summary() == "Object store account not found" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'Object store account not found' diagnostic, got: %s", readResp.Diagnostics)
	}
}
