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

// newTestServerDataSource creates a serverDataSource wired to the given mock server.
func newTestServerDataSource(t *testing.T, ms *testmock.MockServer) *serverDataSource {
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
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":                 tftypes.String,
		"name":               tftypes.String,
		"created":            tftypes.Number,
		"dns":                tftypes.List{ElementType: tftypes.String},
		"directory_services": tftypes.List{ElementType: tftypes.String},
		"network_interfaces": tftypes.List{ElementType: tftypes.String},
	}}
}

// nullServerDSConfig returns a base config map with all data source attributes null.
func nullServerDSConfig() map[string]tftypes.Value {
	return map[string]tftypes.Value{
		"id":                 tftypes.NewValue(tftypes.String, nil),
		"name":               tftypes.NewValue(tftypes.String, nil),
		"created":            tftypes.NewValue(tftypes.Number, nil),
		"dns":                tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"directory_services": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"network_interfaces": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
	}
}

// ---- tests ------------------------------------------------------------------

// TestUnit_ServerDataSource verifies data source reads a server by name and returns all attributes.
func TestUnit_ServerDataSource(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterServerHandlers(ms.Mux)
	store.AddServer("srv-numberly-backup-pr")
	handlers.RegisterNetworkInterfaceHandlers(ms.Mux)

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
	// Verify DNS contains expected name from mock seed.
	var dnsNames []string
	if diags := model.DNS.ElementsAs(context.Background(), &dnsNames, false); diags.HasError() {
		t.Fatalf("DNS ElementsAs: %s", diags)
	}
	if len(dnsNames) != 1 || dnsNames[0] != "management" {
		t.Errorf("expected dns=[management], got %v", dnsNames)
	}
	// directory_services should be populated from seed data.
	if model.DirectoryServices.IsNull() {
		t.Error("expected directory_services to be populated from seed data")
	}
	var dsNames []string
	if diags := model.DirectoryServices.ElementsAs(context.Background(), &dsNames, false); diags.HasError() {
		t.Fatalf("DirectoryServices ElementsAs: %s", diags)
	}
	if len(dsNames) != 1 || dsNames[0] != "srv-backup_nfs" {
		t.Errorf("expected directory_services=[srv-backup_nfs], got %v", dsNames)
	}
	// network_interfaces should be empty list (not null) when no VIPs attached.
	if model.NetworkInterfaces.IsNull() {
		t.Error("expected network_interfaces to be empty list (not null) when no VIPs attached")
	}
}

// TestUnit_ServerDataSource_NotFound verifies that a missing server returns an error diagnostic.
func TestUnit_ServerDataSource_NotFound(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterServerHandlers(ms.Mux)
	handlers.RegisterNetworkInterfaceHandlers(ms.Mux)

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

// TestUnit_ServerDataSource_VIPEnrichment verifies that the data source populates
// network_interfaces when VIPs are attached to the requested server.
func TestUnit_ServerDataSource_VIPEnrichment(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	srvStore := handlers.RegisterServerHandlers(ms.Mux)
	srvStore.AddServer("ds-enrich-server")
	niStore := handlers.RegisterNetworkInterfaceHandlers(ms.Mux)

	ni1 := niStore.AddNetworkInterface("ds-vip1.eth0", "10.0.4.1", "subnet-d", "vip", "data")
	ni1.AttachedServers = []client.NamedReference{{Name: "ds-enrich-server"}}

	ni2 := niStore.AddNetworkInterface("ds-vip2.eth0", "10.0.4.2", "subnet-d", "vip", "data")
	ni2.AttachedServers = []client.NamedReference{{Name: "ds-enrich-server"}}

	// VIP not attached to our server.
	niStore.AddNetworkInterface("ds-vip3.eth0", "10.0.4.3", "subnet-d", "vip", "data")

	d := newTestServerDataSource(t, ms)
	s := serverDataSourceSchema(t).Schema

	cfg := nullServerDSConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "ds-enrich-server")

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

	if model.NetworkInterfaces.IsNull() {
		t.Fatal("expected network_interfaces to be populated after data source Read with attached VIPs")
	}

	var niNames []string
	if diags := model.NetworkInterfaces.ElementsAs(context.Background(), &niNames, false); diags.HasError() {
		t.Fatalf("ElementsAs: %s", diags)
	}
	if len(niNames) != 2 {
		t.Fatalf("expected 2 network interfaces, got %d: %v", len(niNames), niNames)
	}
	// Order not guaranteed — check both names are present.
	nameSet := map[string]bool{}
	for _, n := range niNames {
		nameSet[n] = true
	}
	if !nameSet["ds-vip1.eth0"] {
		t.Error("expected ds-vip1.eth0 in network_interfaces")
	}
	if !nameSet["ds-vip2.eth0"] {
		t.Error("expected ds-vip2.eth0 in network_interfaces")
	}
}
