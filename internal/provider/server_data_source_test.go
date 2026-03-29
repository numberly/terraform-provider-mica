package provider

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/numberly/opentofu-provider-flashblade/internal/client"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock/handlers"
)

// ---- helpers ----------------------------------------------------------------

// newTestServerDataSource creates a serverDataSource wired to the given mock server.
func newTestServerDataSource(t *testing.T, ms *testmock.MockServer) *serverDataSource {
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
	return &serverDataSource{client: c}
}

// serverDataSourceSchema returns the schema for the server data source.
func serverDataSourceSchema(t *testing.T) datasource.SchemaResponse {
	t.Helper()
	d := &serverDataSource{}
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	return resp
}

// buildServerDSType returns the tftypes.Object for the server data source schema.
func buildServerDSType() tftypes.Object {
	dnsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"domain":      tftypes.String,
		"nameservers": tftypes.List{ElementType: tftypes.String},
		"services":    tftypes.List{ElementType: tftypes.String},
	}}
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":      tftypes.String,
		"name":    tftypes.String,
		"created": tftypes.Number,
		"dns":     tftypes.List{ElementType: dnsType},
	}}
}

// nullServerDSConfig returns a base config map with all data source attributes null.
func nullServerDSConfig() map[string]tftypes.Value {
	dnsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"domain":      tftypes.String,
		"nameservers": tftypes.List{ElementType: tftypes.String},
		"services":    tftypes.List{ElementType: tftypes.String},
	}}
	return map[string]tftypes.Value{
		"id":      tftypes.NewValue(tftypes.String, nil),
		"name":    tftypes.NewValue(tftypes.String, nil),
		"created": tftypes.NewValue(tftypes.Number, nil),
		"dns":     tftypes.NewValue(tftypes.List{ElementType: dnsType}, nil),
	}
}

// ---- tests ------------------------------------------------------------------

// TestUnit_ServerDataSource verifies data source reads a server by name and returns all attributes.
func TestUnit_ServerDataSource(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterServerHandlers(ms.Mux)
	store.AddServer("srv-numberly-backup-pr")

	d := newTestServerDataSource(t, ms)
	s := serverDataSourceSchema(t).Schema

	cfg := nullServerDSConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "srv-numberly-backup-pr")

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildServerDSType(), nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(buildServerDSType(), cfg), Schema: s},
	}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("DataSource Read returned error: %s", readResp.Diagnostics)
	}

	var model serverDataSourceModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Name.ValueString() != "srv-numberly-backup-pr" {
		t.Errorf("expected name=srv-numberly-backup-pr, got %s", model.Name.ValueString())
	}
	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected ID to be populated")
	}
	if model.Created.IsNull() || model.Created.ValueInt64() == 0 {
		t.Error("expected created to be populated")
	}
	if model.DNS.IsNull() {
		t.Error("expected dns to be populated")
	}
}

// TestUnit_ServerDataSource_NotFound verifies that a missing server returns an error diagnostic.
func TestUnit_ServerDataSource_NotFound(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterServerHandlers(ms.Mux)

	d := newTestServerDataSource(t, ms)
	s := serverDataSourceSchema(t).Schema

	cfg := nullServerDSConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "nonexistent-server")

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildServerDSType(), nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(buildServerDSType(), cfg), Schema: s},
	}, readResp)

	if !readResp.Diagnostics.HasError() {
		t.Error("expected error diagnostic for not-found server, got none")
	}

	found := false
	for _, diag := range readResp.Diagnostics {
		if diag.Summary() == "Server not found" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'Server not found' diagnostic, got: %s", readResp.Diagnostics)
	}
}
