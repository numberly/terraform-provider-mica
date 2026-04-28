package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/numberly/terraform-provider-mica/internal/client"
	"github.com/numberly/terraform-provider-mica/internal/testmock"
	"github.com/numberly/terraform-provider-mica/internal/testmock/handlers"
)

// ---- helpers ----------------------------------------------------------------

// newTestQosPolicyMemberResource creates a qosPolicyMemberResource wired to the given mock server.
func newTestQosPolicyMemberResource(t *testing.T, ms *testmock.MockServer) *qosPolicyMemberResource {
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
	return &qosPolicyMemberResource{client: c}
}

// qosPolicyMemberResourceSchema returns the parsed schema for the QoS policy member resource.
func qosPolicyMemberResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &qosPolicyMemberResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

// buildQosPolicyMemberType returns the tftypes.Object for the QoS policy member resource schema.
func buildQosPolicyMemberType() tftypes.Object {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"delete": tftypes.String,
	}}
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"policy_name": tftypes.String,
		"member_name": tftypes.String,
		"member_type": tftypes.String,
		"timeouts":    timeoutsType,
	}}
}

// nullQosPolicyMemberConfig returns a base config map with all attributes null.
func nullQosPolicyMemberConfig() map[string]tftypes.Value {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"delete": tftypes.String,
	}}
	return map[string]tftypes.Value{
		"policy_name": tftypes.NewValue(tftypes.String, nil),
		"member_name": tftypes.NewValue(tftypes.String, nil),
		"member_type": tftypes.NewValue(tftypes.String, nil),
		"timeouts":    tftypes.NewValue(timeoutsType, nil),
	}
}

// qosPolicyMemberPlanWith returns a tfsdk.Plan with the given field values.
func qosPolicyMemberPlanWith(t *testing.T, policyName, memberName string) tfsdk.Plan {
	t.Helper()
	s := qosPolicyMemberResourceSchema(t).Schema
	cfg := nullQosPolicyMemberConfig()
	cfg["policy_name"] = tftypes.NewValue(tftypes.String, policyName)
	cfg["member_name"] = tftypes.NewValue(tftypes.String, memberName)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildQosPolicyMemberType(), cfg),
		Schema: s,
	}
}

// ---- tests ------------------------------------------------------------------

// TestQosPolicyMemberResource_Create verifies POST creates a member assignment.
func TestUnit_QosPolicyMemberResource_Create(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterQosPolicyHandlers(ms.Mux)

	// Seed a QoS policy first.
	store.Seed(&client.QosPolicy{
		ID:         "qos-1",
		Name:       "bw-policy",
		Enabled:    true,
		IsLocal:    true,
		PolicyType: "bandwidth-limit",
	})

	r := newTestQosPolicyMemberResource(t, ms)
	s := qosPolicyMemberResourceSchema(t).Schema

	plan := qosPolicyMemberPlanWith(t, "bw-policy", "my-bucket")
	req := resource.CreateRequest{Plan: plan}
	resp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildQosPolicyMemberType(), nil), Schema: s},
	}

	r.Create(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Create returned error: %s", resp.Diagnostics)
	}

	var model qosPolicyMemberModel
	if diags := resp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.PolicyName.ValueString() != "bw-policy" {
		t.Errorf("expected policy_name=bw-policy, got %s", model.PolicyName.ValueString())
	}
	if model.MemberName.ValueString() != "my-bucket" {
		t.Errorf("expected member_name=my-bucket, got %s", model.MemberName.ValueString())
	}
}

// TestQosPolicyMemberResource_Read verifies reading an existing member.
func TestUnit_QosPolicyMemberResource_Read(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterQosPolicyHandlers(ms.Mux)

	// Seed policy and member.
	store.Seed(&client.QosPolicy{
		ID:         "qos-2",
		Name:       "read-policy",
		Enabled:    true,
		IsLocal:    true,
		PolicyType: "bandwidth-limit",
	})
	store.SeedMember("read-policy", client.QosPolicyMember{
		Member: client.NamedReference{Name: "read-bucket"},
		Policy: client.NamedReference{Name: "read-policy"},
	})

	r := newTestQosPolicyMemberResource(t, ms)
	s := qosPolicyMemberResourceSchema(t).Schema

	cfg := nullQosPolicyMemberConfig()
	cfg["policy_name"] = tftypes.NewValue(tftypes.String, "read-policy")
	cfg["member_name"] = tftypes.NewValue(tftypes.String, "read-bucket")

	state := tfsdk.State{
		Raw:    tftypes.NewValue(buildQosPolicyMemberType(), cfg),
		Schema: s,
	}

	readResp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), resource.ReadRequest{State: state}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}

	var model qosPolicyMemberModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.PolicyName.ValueString() != "read-policy" {
		t.Errorf("expected policy_name=read-policy, got %s", model.PolicyName.ValueString())
	}
	if model.MemberName.ValueString() != "read-bucket" {
		t.Errorf("expected member_name=read-bucket, got %s", model.MemberName.ValueString())
	}
}

// TestQosPolicyMemberResource_Read_NotFound verifies reading non-existent member removes state.
func TestUnit_QosPolicyMemberResource_Read_NotFound(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterQosPolicyHandlers(ms.Mux)

	// Seed policy with no members.
	store.Seed(&client.QosPolicy{
		ID:         "qos-3",
		Name:       "empty-policy",
		Enabled:    true,
		IsLocal:    true,
		PolicyType: "bandwidth-limit",
	})

	r := newTestQosPolicyMemberResource(t, ms)
	s := qosPolicyMemberResourceSchema(t).Schema

	cfg := nullQosPolicyMemberConfig()
	cfg["policy_name"] = tftypes.NewValue(tftypes.String, "empty-policy")
	cfg["member_name"] = tftypes.NewValue(tftypes.String, "ghost-bucket")

	state := tfsdk.State{
		Raw:    tftypes.NewValue(buildQosPolicyMemberType(), cfg),
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

// TestQosPolicyMemberResource_Delete verifies DELETE removes a member.
func TestUnit_QosPolicyMemberResource_Delete(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterQosPolicyHandlers(ms.Mux)

	// Seed policy.
	store.Seed(&client.QosPolicy{
		ID:         "qos-4",
		Name:       "del-policy",
		Enabled:    true,
		IsLocal:    true,
		PolicyType: "bandwidth-limit",
	})

	r := newTestQosPolicyMemberResource(t, ms)
	s := qosPolicyMemberResourceSchema(t).Schema

	// Create member first.
	plan := qosPolicyMemberPlanWith(t, "del-policy", "del-bucket")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildQosPolicyMemberType(), nil), Schema: s},
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

	// Verify the member is gone.
	members, err := r.client.ListQosPolicyMembers(context.Background(), "del-policy")
	if err != nil {
		t.Fatalf("ListQosPolicyMembers: %v", err)
	}
	for _, m := range members {
		if m.Member.Name == "del-bucket" {
			t.Error("expected member to be deleted, but it still exists")
		}
	}
}

// TestQosPolicyMemberResource_Import verifies Import by "policyName/memberName" populates state.
func TestUnit_QosPolicyMemberResource_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterQosPolicyHandlers(ms.Mux)

	// Seed policy and member.
	store.Seed(&client.QosPolicy{
		ID:         "qos-5",
		Name:       "imp-policy",
		Enabled:    true,
		IsLocal:    true,
		PolicyType: "bandwidth-limit",
	})
	store.SeedMember("imp-policy", client.QosPolicyMember{
		Member: client.NamedReference{Name: "imp-bucket"},
		Policy: client.NamedReference{Name: "imp-policy"},
	})

	r := newTestQosPolicyMemberResource(t, ms)
	s := qosPolicyMemberResourceSchema(t).Schema

	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildQosPolicyMemberType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "imp-policy/imp-bucket"}, importResp)

	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState returned error: %s", importResp.Diagnostics)
	}

	var model qosPolicyMemberModel
	if diags := importResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.PolicyName.ValueString() != "imp-policy" {
		t.Errorf("expected policy_name=imp-policy, got %s", model.PolicyName.ValueString())
	}
	if model.MemberName.ValueString() != "imp-bucket" {
		t.Errorf("expected member_name=imp-bucket, got %s", model.MemberName.ValueString())
	}
}
