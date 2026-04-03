package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/numberly/opentofu-provider-flashblade/internal/client"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock/handlers"
)

// ---- helpers ----------------------------------------------------------------

// newTestTlsPolicyMemberResource creates a tlsPolicyMemberResource wired to the given mock server.
func newTestTlsPolicyMemberResource(t *testing.T, ms *testmock.MockServer) *tlsPolicyMemberResource {
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
	return &tlsPolicyMemberResource{client: c}
}

// tlsPolicyMemberResourceSchema returns the parsed schema for the TLS policy member resource.
func tlsPolicyMemberResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &tlsPolicyMemberResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

// buildTlsPolicyMemberType returns the tftypes.Object for the TLS policy member resource schema.
func buildTlsPolicyMemberType() tftypes.Object {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"delete": tftypes.String,
	}}
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"policy_name": tftypes.String,
		"member_name": tftypes.String,
		"timeouts":    timeoutsType,
	}}
}

// nullTlsPolicyMemberConfig returns a base config map with all attributes null.
func nullTlsPolicyMemberConfig() map[string]tftypes.Value {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"delete": tftypes.String,
	}}
	return map[string]tftypes.Value{
		"policy_name": tftypes.NewValue(tftypes.String, nil),
		"member_name": tftypes.NewValue(tftypes.String, nil),
		"timeouts":    tftypes.NewValue(timeoutsType, nil),
	}
}

// tlsPolicyMemberPlanWith returns a tfsdk.Plan with the given field values.
func tlsPolicyMemberPlanWith(t *testing.T, policyName, memberName string) tfsdk.Plan {
	t.Helper()
	s := tlsPolicyMemberResourceSchema(t).Schema
	cfg := nullTlsPolicyMemberConfig()
	cfg["policy_name"] = tftypes.NewValue(tftypes.String, policyName)
	cfg["member_name"] = tftypes.NewValue(tftypes.String, memberName)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildTlsPolicyMemberType(), cfg),
		Schema: s,
	}
}

// ---- tests ------------------------------------------------------------------

// TestUnit_TlsPolicyMemberResource_Lifecycle: create member → read → delete → verify gone.
func TestUnit_TlsPolicyMemberResource_Lifecycle(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterTlsPolicyHandlers(ms.Mux)

	// Seed a TLS policy first.
	store.Seed(&client.TlsPolicy{
		ID:         "tls-1",
		Name:       "lifecycle-policy",
		Enabled:    true,
		IsLocal:    true,
		PolicyType: "global",
	})

	r := newTestTlsPolicyMemberResource(t, ms)
	s := tlsPolicyMemberResourceSchema(t).Schema

	// Create.
	plan := tlsPolicyMemberPlanWith(t, "lifecycle-policy", "data-vip1")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildTlsPolicyMemberType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create returned error: %s", createResp.Diagnostics)
	}

	var afterCreate tlsPolicyMemberModel
	if diags := createResp.State.Get(context.Background(), &afterCreate); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}
	if afterCreate.PolicyName.ValueString() != "lifecycle-policy" {
		t.Errorf("expected policy_name=lifecycle-policy, got %s", afterCreate.PolicyName.ValueString())
	}
	if afterCreate.MemberName.ValueString() != "data-vip1" {
		t.Errorf("expected member_name=data-vip1, got %s", afterCreate.MemberName.ValueString())
	}

	// Read (idempotent).
	readResp := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}

	var afterRead tlsPolicyMemberModel
	if diags := readResp.State.Get(context.Background(), &afterRead); diags.HasError() {
		t.Fatalf("Get read state: %s", diags)
	}
	if afterRead.MemberName.ValueString() != "data-vip1" {
		t.Errorf("expected member_name=data-vip1 after Read, got %s", afterRead.MemberName.ValueString())
	}

	// Delete.
	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: readResp.State}, deleteResp)
	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete returned error: %s", deleteResp.Diagnostics)
	}

	// Verify the member is gone.
	members, err := r.client.ListTlsPolicyMembers(context.Background(), "lifecycle-policy")
	if err != nil {
		t.Fatalf("ListTlsPolicyMembers: %v", err)
	}
	for _, m := range members {
		if m.Member.Name == "data-vip1" {
			t.Error("expected member to be deleted, but it still exists")
		}
	}
}

// TestUnit_TlsPolicyMemberResource_Read_NotFound: seed policy with no members → read → state removed.
func TestUnit_TlsPolicyMemberResource_Read_NotFound(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterTlsPolicyHandlers(ms.Mux)

	// Seed policy with no members.
	store.Seed(&client.TlsPolicy{
		ID:         "tls-2",
		Name:       "empty-policy",
		Enabled:    true,
		IsLocal:    true,
		PolicyType: "global",
	})

	r := newTestTlsPolicyMemberResource(t, ms)
	s := tlsPolicyMemberResourceSchema(t).Schema

	cfg := nullTlsPolicyMemberConfig()
	cfg["policy_name"] = tftypes.NewValue(tftypes.String, "empty-policy")
	cfg["member_name"] = tftypes.NewValue(tftypes.String, "ghost-vip")

	state := tfsdk.State{
		Raw:    tftypes.NewValue(buildTlsPolicyMemberType(), cfg),
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

// TestUnit_TlsPolicyMemberResource_Import: seed policy + member → import "policyName/memberName" → verify state.
func TestUnit_TlsPolicyMemberResource_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterTlsPolicyHandlers(ms.Mux)

	// Seed policy and member.
	store.Seed(&client.TlsPolicy{
		ID:         "tls-3",
		Name:       "imp-policy",
		Enabled:    true,
		IsLocal:    true,
		PolicyType: "global",
	})
	store.SeedMember("imp-policy", client.TlsPolicyMember{
		Policy: client.NamedReference{Name: "imp-policy"},
		Member: client.NamedReference{Name: "imp-vip"},
	})

	r := newTestTlsPolicyMemberResource(t, ms)
	s := tlsPolicyMemberResourceSchema(t).Schema

	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildTlsPolicyMemberType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "imp-policy/imp-vip"}, importResp)

	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState returned error: %s", importResp.Diagnostics)
	}

	var model tlsPolicyMemberModel
	if diags := importResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get import state: %s", diags)
	}

	if model.PolicyName.ValueString() != "imp-policy" {
		t.Errorf("expected policy_name=imp-policy, got %s", model.PolicyName.ValueString())
	}
	if model.MemberName.ValueString() != "imp-vip" {
		t.Errorf("expected member_name=imp-vip, got %s", model.MemberName.ValueString())
	}
	// Timeouts should be null for CRD-only import.
	if !model.Timeouts.IsNull() {
		t.Error("expected timeouts to be null after import")
	}
}
