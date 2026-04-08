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

// newTestCertificateGroupDataSource creates a certificateGroupDataSource wired to the given mock server.
func newTestCertificateGroupDataSource(t *testing.T, ms *testmock.MockServer) *certificateGroupDataSource {
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
	return &certificateGroupDataSource{client: c}
}

// certificateGroupDSSchema returns the parsed schema for the certificate group data source.
func certificateGroupDSSchema(t *testing.T) datasource.SchemaResponse {
	t.Helper()
	d := &certificateGroupDataSource{}
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	return resp
}

// buildCertificateGroupDSType returns the tftypes.Object for the certificate group data source schema.
func buildCertificateGroupDSType() tftypes.Object {
	realmsType := tftypes.List{ElementType: tftypes.String}
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":     tftypes.String,
		"name":   tftypes.String,
		"realms": realmsType,
	}}
}

// nullCertificateGroupDSConfig returns a base config map with all attributes null.
func nullCertificateGroupDSConfig() map[string]tftypes.Value {
	realmsType := tftypes.List{ElementType: tftypes.String}
	return map[string]tftypes.Value{
		"id":     tftypes.NewValue(tftypes.String, nil),
		"name":   tftypes.NewValue(tftypes.String, nil),
		"realms": tftypes.NewValue(realmsType, nil),
	}
}

// ---- tests ------------------------------------------------------------------

// TestUnit_CertificateGroupDataSource_Basic: seed group → read → verify all fields populated.
func TestUnit_CertificateGroupDataSource_Basic(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	rawStore := handlers.RegisterCertificateGroupHandlers(ms.Mux)
	store := handlers.NewCertificateGroupStoreFacade(rawStore)

	d := newTestCertificateGroupDataSource(t, ms)
	s := certificateGroupDSSchema(t).Schema

	// Seed a certificate group with realms.
	store.Seed(&client.CertificateGroup{
		ID:     "cg-ds-1",
		Name:   "ds-group",
		Realms: []string{"management", "replication"},
	})

	// Build config with name set.
	cfg := nullCertificateGroupDSConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "ds-group")

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{
			Raw:    tftypes.NewValue(buildCertificateGroupDSType(), cfg),
			Schema: s,
		},
	}

	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{
			Raw:    tftypes.NewValue(buildCertificateGroupDSType(), cfg),
			Schema: s,
		},
	}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}

	var model certificateGroupDataSourceModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.ValueString() != "cg-ds-1" {
		t.Errorf("expected id=cg-ds-1, got %s", model.ID.ValueString())
	}
	if model.Name.ValueString() != "ds-group" {
		t.Errorf("expected name=ds-group, got %s", model.Name.ValueString())
	}
	if len(model.Realms.Elements()) != 2 {
		t.Errorf("expected 2 realms, got %d", len(model.Realms.Elements()))
	}
}
