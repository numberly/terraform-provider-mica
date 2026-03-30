package provider

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	resschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/numberly/opentofu-provider-flashblade/internal/client"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock/handlers"
)

// ---- helpers ----------------------------------------------------------------

// newTestQosPolicyResource creates a qosPolicyResource wired to the given mock server.
func newTestQosPolicyResource(t *testing.T, ms *testmock.MockServer) *qosPolicyResource {
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
	return &qosPolicyResource{client: c}
}

// qosPolicyResourceSchema returns the parsed schema for the QoS policy resource.
func qosPolicyResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &qosPolicyResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

// buildQosPolicyType returns the tftypes.Object for the QoS policy resource schema.
func buildQosPolicyType() tftypes.Object {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":                     tftypes.String,
		"name":                   tftypes.String,
		"enabled":                tftypes.Bool,
		"max_total_bytes_per_sec": tftypes.Number,
		"max_total_ops_per_sec":  tftypes.Number,
		"is_local":               tftypes.Bool,
		"policy_type":            tftypes.String,
		"timeouts":               timeoutsType,
	}}
}

// nullQosPolicyConfig returns a base config map with all attributes null.
func nullQosPolicyConfig() map[string]tftypes.Value {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return map[string]tftypes.Value{
		"id":                     tftypes.NewValue(tftypes.String, nil),
		"name":                   tftypes.NewValue(tftypes.String, nil),
		"enabled":                tftypes.NewValue(tftypes.Bool, nil),
		"max_total_bytes_per_sec": tftypes.NewValue(tftypes.Number, nil),
		"max_total_ops_per_sec":  tftypes.NewValue(tftypes.Number, nil),
		"is_local":               tftypes.NewValue(tftypes.Bool, nil),
		"policy_type":            tftypes.NewValue(tftypes.String, nil),
		"timeouts":               tftypes.NewValue(timeoutsType, nil),
	}
}

// qosPolicyPlanWith returns a tfsdk.Plan with the given field values.
func qosPolicyPlanWith(t *testing.T, name string, enabled bool, maxBytes int64, maxOps int64) tfsdk.Plan {
	t.Helper()
	s := qosPolicyResourceSchema(t).Schema
	cfg := nullQosPolicyConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, name)
	cfg["enabled"] = tftypes.NewValue(tftypes.Bool, enabled)
	if maxBytes > 0 {
		cfg["max_total_bytes_per_sec"] = tftypes.NewValue(tftypes.Number, maxBytes)
	}
	if maxOps > 0 {
		cfg["max_total_ops_per_sec"] = tftypes.NewValue(tftypes.Number, maxOps)
	}
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildQosPolicyType(), cfg),
		Schema: s,
	}
}

// ---- tests ------------------------------------------------------------------

// TestQosPolicyResource_Create verifies POST creates a policy, state populated with all fields.
func TestQosPolicyResource_Create(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterQosPolicyHandlers(ms.Mux)

	r := newTestQosPolicyResource(t, ms)
	s := qosPolicyResourceSchema(t).Schema

	plan := qosPolicyPlanWith(t, "my-qos-policy", true, 1048576, 0)
	req := resource.CreateRequest{Plan: plan}
	resp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildQosPolicyType(), nil), Schema: s},
	}

	r.Create(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Create returned error: %s", resp.Diagnostics)
	}

	var model qosPolicyModel
	if diags := resp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected non-empty id after Create")
	}
	if model.Name.ValueString() != "my-qos-policy" {
		t.Errorf("expected name=my-qos-policy, got %s", model.Name.ValueString())
	}
	if model.Enabled.ValueBool() != true {
		t.Error("expected enabled=true after Create")
	}
	if model.MaxTotalBytesPerSec.ValueInt64() != 1048576 {
		t.Errorf("expected max_total_bytes_per_sec=1048576, got %d", model.MaxTotalBytesPerSec.ValueInt64())
	}
	if !model.MaxTotalOpsPerSec.IsNull() {
		t.Errorf("expected max_total_ops_per_sec=null, got %d", model.MaxTotalOpsPerSec.ValueInt64())
	}
	if model.IsLocal.ValueBool() != true {
		t.Error("expected is_local=true after Create")
	}
	if model.PolicyType.ValueString() != "bandwidth-limit" {
		t.Errorf("expected policy_type=bandwidth-limit, got %s", model.PolicyType.ValueString())
	}
}

// TestQosPolicyResource_Read verifies GET retrieves policy by name.
func TestQosPolicyResource_Read(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterQosPolicyHandlers(ms.Mux)

	r := newTestQosPolicyResource(t, ms)
	s := qosPolicyResourceSchema(t).Schema

	// Create first.
	plan := qosPolicyPlanWith(t, "read-policy", true, 0, 5000)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildQosPolicyType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Read.
	readResp := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}

	var model qosPolicyModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Name.ValueString() != "read-policy" {
		t.Errorf("expected name=read-policy, got %s", model.Name.ValueString())
	}
	if model.MaxTotalOpsPerSec.ValueInt64() != 5000 {
		t.Errorf("expected max_total_ops_per_sec=5000, got %d", model.MaxTotalOpsPerSec.ValueInt64())
	}
}

// TestQosPolicyResource_Read_NotFound verifies GET returns 404, resource removed from state.
func TestQosPolicyResource_Read_NotFound(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterQosPolicyHandlers(ms.Mux)

	r := newTestQosPolicyResource(t, ms)
	s := qosPolicyResourceSchema(t).Schema

	cfg := nullQosPolicyConfig()
	cfg["id"] = tftypes.NewValue(tftypes.String, "qos-999")
	cfg["name"] = tftypes.NewValue(tftypes.String, "ghost-policy")
	cfg["enabled"] = tftypes.NewValue(tftypes.Bool, true)

	state := tfsdk.State{
		Raw:    tftypes.NewValue(buildQosPolicyType(), cfg),
		Schema: s,
	}

	readResp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), resource.ReadRequest{State: state}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned unexpected error: %s", readResp.Diagnostics)
	}

	if !readResp.State.Raw.IsNull() {
		t.Error("expected state to be removed after not-found Read")
	}
}

// TestQosPolicyResource_Update verifies PATCH sends only changed fields.
func TestQosPolicyResource_Update(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterQosPolicyHandlers(ms.Mux)

	r := newTestQosPolicyResource(t, ms)
	s := qosPolicyResourceSchema(t).Schema

	// Create (no ops limit).
	createPlan := qosPolicyPlanWith(t, "upd-policy", true, 1048576, 0)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildQosPolicyType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Update: add ops limit.
	updatePlan := qosPolicyPlanWith(t, "upd-policy", true, 1048576, 10000)
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildQosPolicyType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlan,
		State: createResp.State,
	}, updateResp)

	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update returned error: %s", updateResp.Diagnostics)
	}

	var model qosPolicyModel
	if diags := updateResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.MaxTotalOpsPerSec.ValueInt64() != 10000 {
		t.Errorf("expected max_total_ops_per_sec=10000, got %d", model.MaxTotalOpsPerSec.ValueInt64())
	}
	if model.MaxTotalBytesPerSec.ValueInt64() != 1048576 {
		t.Errorf("expected max_total_bytes_per_sec=1048576 (unchanged), got %d", model.MaxTotalBytesPerSec.ValueInt64())
	}
}

// TestQosPolicyResource_Delete verifies DELETE by name succeeds.
func TestQosPolicyResource_Delete(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterQosPolicyHandlers(ms.Mux)

	r := newTestQosPolicyResource(t, ms)
	s := qosPolicyResourceSchema(t).Schema

	// Create.
	plan := qosPolicyPlanWith(t, "del-policy", true, 0, 0)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildQosPolicyType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Delete.
	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: createResp.State}, deleteResp)

	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete returned error: %s", deleteResp.Diagnostics)
	}

	// Verify it's gone.
	_, err := r.client.GetQosPolicy(context.Background(), "del-policy")
	if err == nil || !client.IsNotFound(err) {
		t.Errorf("expected QoS policy to be deleted, got: %v", err)
	}
}

// TestQosPolicyResource_Import verifies Import by name populates all state fields.
func TestQosPolicyResource_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterQosPolicyHandlers(ms.Mux)

	// Seed a policy so import can find it.
	store.Seed(&client.QosPolicy{
		ID:                  "qos-imp-1",
		Name:                "imported-policy",
		Enabled:             true,
		IsLocal:             true,
		MaxTotalBytesPerSec: 2097152,
		MaxTotalOpsPerSec:   5000,
		PolicyType:          "bandwidth-limit",
	})

	r := newTestQosPolicyResource(t, ms)
	s := qosPolicyResourceSchema(t).Schema

	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildQosPolicyType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "imported-policy"}, importResp)

	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState returned error: %s", importResp.Diagnostics)
	}

	var model qosPolicyModel
	if diags := importResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.ValueString() != "qos-imp-1" {
		t.Errorf("expected id=qos-imp-1, got %s", model.ID.ValueString())
	}
	if model.Name.ValueString() != "imported-policy" {
		t.Errorf("expected name=imported-policy, got %s", model.Name.ValueString())
	}
	if model.Enabled.ValueBool() != true {
		t.Error("expected enabled=true after import")
	}
	if model.MaxTotalBytesPerSec.ValueInt64() != 2097152 {
		t.Errorf("expected max_total_bytes_per_sec=2097152, got %d", model.MaxTotalBytesPerSec.ValueInt64())
	}
	if model.MaxTotalOpsPerSec.ValueInt64() != 5000 {
		t.Errorf("expected max_total_ops_per_sec=5000, got %d", model.MaxTotalOpsPerSec.ValueInt64())
	}
	if model.IsLocal.ValueBool() != true {
		t.Error("expected is_local=true after import")
	}
	if model.PolicyType.ValueString() != "bandwidth-limit" {
		t.Errorf("expected policy_type=bandwidth-limit, got %s", model.PolicyType.ValueString())
	}
}

// TestQosPolicyResource_Schema verifies schema properties.
func TestQosPolicyResource_Schema(t *testing.T) {
	s := qosPolicyResourceSchema(t).Schema

	// name: Required + RequiresReplace.
	nameAttr, ok := s.Attributes["name"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("name attribute not found or wrong type")
	}
	if !nameAttr.Required {
		t.Error("name: expected Required=true")
	}
	if len(nameAttr.PlanModifiers) == 0 {
		t.Error("name: expected RequiresReplace plan modifier")
	}

	// enabled: Optional + Computed.
	enabledAttr, ok := s.Attributes["enabled"].(resschema.BoolAttribute)
	if !ok {
		t.Fatal("enabled attribute not found or wrong type")
	}
	if !enabledAttr.Optional {
		t.Error("enabled: expected Optional=true")
	}
	if !enabledAttr.Computed {
		t.Error("enabled: expected Computed=true")
	}

	// is_local: Computed only.
	isLocalAttr, ok := s.Attributes["is_local"].(resschema.BoolAttribute)
	if !ok {
		t.Fatal("is_local attribute not found or wrong type")
	}
	if !isLocalAttr.Computed {
		t.Error("is_local: expected Computed=true")
	}

	// policy_type: Computed only.
	policyTypeAttr, ok := s.Attributes["policy_type"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("policy_type attribute not found or wrong type")
	}
	if !policyTypeAttr.Computed {
		t.Error("policy_type: expected Computed=true")
	}
}
