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

// newTestArrayConnectionDataSource creates an arrayConnectionDataSource wired to the given mock server.
func newTestArrayConnectionDataSource(t *testing.T, ms *testmock.MockServer) *arrayConnectionDataSource {
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
	return &arrayConnectionDataSource{client: c}
}

// arrayConnectionDSSchema returns the schema for the array connection data source.
func arrayConnectionDSSchema(t *testing.T) datasource.SchemaResponse {
	t.Helper()
	d := &arrayConnectionDataSource{}
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	return resp
}

// buildArrayConnectionDSType returns the tftypes.Object for the array connection data source schema.
func buildArrayConnectionDSType() tftypes.Object {
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":                    tftypes.String,
		"remote_name":           tftypes.String,
		"remote_id":             tftypes.String,
		"status":                tftypes.String,
		"management_address":    tftypes.String,
		"replication_addresses": tftypes.List{ElementType: tftypes.String},
		"encrypted":             tftypes.Bool,
		"type":                  tftypes.String,
		"version":               tftypes.String,
	}}
}

// nullArrayConnectionDSConfig returns a base config map with all attributes null.
func nullArrayConnectionDSConfig() map[string]tftypes.Value {
	return map[string]tftypes.Value{
		"id":                    tftypes.NewValue(tftypes.String, nil),
		"remote_name":           tftypes.NewValue(tftypes.String, nil),
		"remote_id":             tftypes.NewValue(tftypes.String, nil),
		"status":                tftypes.NewValue(tftypes.String, nil),
		"management_address":    tftypes.NewValue(tftypes.String, nil),
		"replication_addresses": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"encrypted":             tftypes.NewValue(tftypes.Bool, nil),
		"type":                  tftypes.NewValue(tftypes.String, nil),
		"version":               tftypes.NewValue(tftypes.String, nil),
	}
}

// ---- tests ------------------------------------------------------------------

// TestUnit_ArrayConnection_Read verifies the data source reads an array connection by remote_name
// and returns all attributes correctly.
func TestUnit_ArrayConnection_Read(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterArrayConnectionHandlers(ms.Mux)
	store.Seed(&client.ArrayConnection{
		ID:     "conn-123",
		Status: "connected",
		Remote: client.NamedReference{
			Name: "remote-fb",
			ID:   "remote-id-456",
		},
		ManagementAddress:    "10.0.0.1",
		ReplicationAddresses: []string{"10.0.1.1", "10.0.1.2"},
		Encrypted:            true,
		Type:                 "async-replication",
		Version:              "4.3.0",
	})

	d := newTestArrayConnectionDataSource(t, ms)
	s := arrayConnectionDSSchema(t).Schema

	cfg := nullArrayConnectionDSConfig()
	cfg["remote_name"] = tftypes.NewValue(tftypes.String, "remote-fb")

	objType := buildArrayConnectionDSType()
	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(objType, cfg), Schema: s},
	}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("DataSource Read returned error: %s", readResp.Diagnostics)
	}

	var model arrayConnectionDataSourceModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.ValueString() != "conn-123" {
		t.Errorf("expected id=conn-123, got %s", model.ID.ValueString())
	}
	if model.Status.ValueString() != "connected" {
		t.Errorf("expected status=connected, got %s", model.Status.ValueString())
	}
	if model.RemoteName.ValueString() != "remote-fb" {
		t.Errorf("expected remote_name=remote-fb, got %s", model.RemoteName.ValueString())
	}
	if model.RemoteID.ValueString() != "remote-id-456" {
		t.Errorf("expected remote_id=remote-id-456, got %s", model.RemoteID.ValueString())
	}
	if model.ManagementAddress.ValueString() != "10.0.0.1" {
		t.Errorf("expected management_address=10.0.0.1, got %s", model.ManagementAddress.ValueString())
	}
	if model.Encrypted.ValueBool() != true {
		t.Error("expected encrypted=true")
	}
	if model.Type.ValueString() != "async-replication" {
		t.Errorf("expected type=async-replication, got %s", model.Type.ValueString())
	}
	if model.Version.ValueString() != "4.3.0" {
		t.Errorf("expected version=4.3.0, got %s", model.Version.ValueString())
	}
	if model.ReplicationAddresses.IsNull() {
		t.Error("expected replication_addresses to be populated")
	}
}

// TestUnit_ArrayConnection_NotFound verifies that a missing array connection returns an error diagnostic.
func TestUnit_ArrayConnection_NotFound(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterArrayConnectionHandlers(ms.Mux)

	d := newTestArrayConnectionDataSource(t, ms)
	s := arrayConnectionDSSchema(t).Schema

	cfg := nullArrayConnectionDSConfig()
	cfg["remote_name"] = tftypes.NewValue(tftypes.String, "nonexistent")

	objType := buildArrayConnectionDSType()
	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(objType, cfg), Schema: s},
	}, readResp)

	if !readResp.Diagnostics.HasError() {
		t.Error("expected error diagnostic for not-found array connection, got none")
	}

	found := false
	for _, diag := range readResp.Diagnostics {
		if diag.Summary() == "Array connection not found" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'Array connection not found' diagnostic, got: %s", readResp.Diagnostics)
	}
}

// TestUnit_ArrayConnection_Schema verifies schema attributes: remote_name is Required, others Computed.
func TestUnit_ArrayConnection_Schema(t *testing.T) {
	resp := arrayConnectionDSSchema(t)
	s := resp.Schema

	// remote_name must be Required
	attr, ok := s.Attributes["remote_name"]
	if !ok {
		t.Fatal("remote_name attribute not found in schema")
	}
	if !attr.IsRequired() {
		t.Error("remote_name should be Required")
	}

	// All other attributes must be Computed
	computedAttrs := []string{"id", "status", "remote_id", "management_address", "replication_addresses", "encrypted", "type", "version"}
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
