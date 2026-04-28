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

func newTestQosPolicyDataSource(t *testing.T, ms *testmock.MockServer) *qosPolicyDataSource {
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
	return &qosPolicyDataSource{client: c}
}

func qosPolicyDSSchema(t *testing.T) datasource.SchemaResponse {
	t.Helper()
	d := &qosPolicyDataSource{}
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	return resp
}

func buildQosPolicyDSType() tftypes.Object {
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"name":                    tftypes.String,
		"id":                      tftypes.String,
		"enabled":                 tftypes.Bool,
		"max_total_bytes_per_sec": tftypes.Number,
		"max_total_ops_per_sec":   tftypes.Number,
		"is_local":                tftypes.Bool,
		"policy_type":             tftypes.String,
	}}
}

func nullQosPolicyDSConfig() map[string]tftypes.Value {
	return map[string]tftypes.Value{
		"name":                    tftypes.NewValue(tftypes.String, nil),
		"id":                      tftypes.NewValue(tftypes.String, nil),
		"enabled":                 tftypes.NewValue(tftypes.Bool, nil),
		"max_total_bytes_per_sec": tftypes.NewValue(tftypes.Number, nil),
		"max_total_ops_per_sec":   tftypes.NewValue(tftypes.Number, nil),
		"is_local":                tftypes.NewValue(tftypes.Bool, nil),
		"policy_type":             tftypes.NewValue(tftypes.String, nil),
	}
}

func TestUnit_QosPolicyDataSource_Read(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterQosPolicyHandlers(ms.Mux)
	store.Seed(&client.QosPolicy{
		ID:                  "qos-ds-1",
		Name:                "ds-policy",
		Enabled:             true,
		IsLocal:             true,
		MaxTotalBytesPerSec: 2097152,
		MaxTotalOpsPerSec:   10000,
		PolicyType:          "bandwidth-limit",
	})

	d := newTestQosPolicyDataSource(t, ms)
	s := qosPolicyDSSchema(t).Schema
	objType := buildQosPolicyDSType()

	cfg := nullQosPolicyDSConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "ds-policy")

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(objType, cfg), Schema: s},
	}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}

	var model qosPolicyDataSourceModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.ValueString() != "qos-ds-1" {
		t.Errorf("expected id=qos-ds-1, got %s", model.ID.ValueString())
	}
	if model.Name.ValueString() != "ds-policy" {
		t.Errorf("expected name=ds-policy, got %s", model.Name.ValueString())
	}
	if model.Enabled.ValueBool() != true {
		t.Error("expected enabled=true")
	}
	if model.MaxTotalBytesPerSec.ValueInt64() != 2097152 {
		t.Errorf("expected max_total_bytes_per_sec=2097152, got %d", model.MaxTotalBytesPerSec.ValueInt64())
	}
	if model.MaxTotalOpsPerSec.ValueInt64() != 10000 {
		t.Errorf("expected max_total_ops_per_sec=10000, got %d", model.MaxTotalOpsPerSec.ValueInt64())
	}
	if model.IsLocal.ValueBool() != true {
		t.Error("expected is_local=true")
	}
	if model.PolicyType.ValueString() != "bandwidth-limit" {
		t.Errorf("expected policy_type=bandwidth-limit, got %s", model.PolicyType.ValueString())
	}
}

func TestUnit_QosPolicyDataSource_Read_NotFound(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterQosPolicyHandlers(ms.Mux)

	d := newTestQosPolicyDataSource(t, ms)
	s := qosPolicyDSSchema(t).Schema
	objType := buildQosPolicyDSType()

	cfg := nullQosPolicyDSConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "nope")

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(objType, cfg), Schema: s},
	}, readResp)

	if !readResp.Diagnostics.HasError() {
		t.Error("expected error diagnostic for not-found QoS policy, got none")
	}

	found := false
	for _, diag := range readResp.Diagnostics {
		if diag.Summary() == "QoS policy not found" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'QoS policy not found' diagnostic, got: %s", readResp.Diagnostics)
	}
}
