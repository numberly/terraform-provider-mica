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

// newTestTargetDataSource creates a targetDataSource wired to the given mock server.
func newTestTargetDataSource(t *testing.T, ms *testmock.MockServer) *targetDataSource {
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
	return &targetDataSource{client: c}
}

// targetDSSchema returns the parsed schema for the target data source.
func targetDSSchema(t *testing.T) datasource.SchemaResponse {
	t.Helper()
	d := &targetDataSource{}
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	return resp
}

// buildTargetDSType returns the tftypes.Object for the target data source schema.
func buildTargetDSType() tftypes.Object {
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":                   tftypes.String,
		"name":                 tftypes.String,
		"address":              tftypes.String,
		"ca_certificate_group": tftypes.String,
		"status":               tftypes.String,
		"status_details":       tftypes.String,
	}}
}

// nullTargetDSConfig returns a base config map with all attributes null.
func nullTargetDSConfig() map[string]tftypes.Value {
	return map[string]tftypes.Value{
		"id":                   tftypes.NewValue(tftypes.String, nil),
		"name":                 tftypes.NewValue(tftypes.String, nil),
		"address":              tftypes.NewValue(tftypes.String, nil),
		"ca_certificate_group": tftypes.NewValue(tftypes.String, nil),
		"status":               tftypes.NewValue(tftypes.String, nil),
		"status_details":       tftypes.NewValue(tftypes.String, nil),
	}
}

// TestUnit_TargetDataSource_Basic seeds a target in the mock and reads it via the data source.
func TestUnit_TargetDataSource_Basic(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterTargetHandlers(ms.Mux)

	// Seed a target for the data source to read.
	store.Seed(&client.Target{
		ID:            "tgt-ds-001",
		Name:          "ds-target",
		Address:       "s3.example.com",
		Status:        "connected",
		StatusDetails: "all good",
	})

	d := newTestTargetDataSource(t, ms)
	s := targetDSSchema(t).Schema
	objType := buildTargetDSType()

	cfg := nullTargetDSConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "ds-target")

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(objType, cfg), Schema: s},
	}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}

	var model targetDataSourceModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.ValueString() != "tgt-ds-001" {
		t.Errorf("expected id=tgt-ds-001, got %s", model.ID.ValueString())
	}
	if model.Address.ValueString() != "s3.example.com" {
		t.Errorf("expected address=s3.example.com, got %s", model.Address.ValueString())
	}
	if model.Status.ValueString() != "connected" {
		t.Errorf("expected status=connected, got %s", model.Status.ValueString())
	}
	if model.StatusDetails.ValueString() != "all good" {
		t.Errorf("expected status_details='all good', got %s", model.StatusDetails.ValueString())
	}
	if !model.CACertificateGroup.IsNull() {
		t.Errorf("expected ca_certificate_group=null (not set), got %s", model.CACertificateGroup.ValueString())
	}
}
