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

func newTestAuditObjectStorePolicyMemberResource(t *testing.T, ms *testmock.MockServer) *auditObjectStorePolicyMemberResource {
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
	return &auditObjectStorePolicyMemberResource{client: c}
}

func auditObjectStorePolicyMemberResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &auditObjectStorePolicyMemberResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

func buildAuditObjectStorePolicyMemberType() tftypes.Object {
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

func nullAuditObjectStorePolicyMemberConfig() map[string]tftypes.Value {
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

func auditObjectStorePolicyMemberPlanWith(t *testing.T, policyName, memberName string) tfsdk.Plan {
	t.Helper()
	s := auditObjectStorePolicyMemberResourceSchema(t).Schema
	cfg := nullAuditObjectStorePolicyMemberConfig()
	cfg["policy_name"] = tftypes.NewValue(tftypes.String, policyName)
	cfg["member_name"] = tftypes.NewValue(tftypes.String, memberName)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildAuditObjectStorePolicyMemberType(), cfg),
		Schema: s,
	}
}

// ---- tests ------------------------------------------------------------------

func TestUnit_AuditObjectStorePolicyMemberResource_Lifecycle(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterAuditObjectStorePolicyHandlers(ms.Mux)

	store.Seed(&client.AuditObjectStorePolicy{
		ID: "pol-1", Name: "audit-pol", Enabled: true, IsLocal: true, PolicyType: "audit",
	})

	r := newTestAuditObjectStorePolicyMemberResource(t, ms)
	s := auditObjectStorePolicyMemberResourceSchema(t).Schema
	objType := buildAuditObjectStorePolicyMemberType()

	// Create
	plan := auditObjectStorePolicyMemberPlanWith(t, "audit-pol", "test-bucket")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	var model auditObjectStorePolicyMemberModel
	if diags := createResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}
	if model.PolicyName.ValueString() != "audit-pol" {
		t.Errorf("expected policy_name=audit-pol, got %s", model.PolicyName.ValueString())
	}
	if model.MemberName.ValueString() != "test-bucket" {
		t.Errorf("expected member_name=test-bucket, got %s", model.MemberName.ValueString())
	}

	// Read
	readResp := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read: %s", readResp.Diagnostics)
	}

	// Delete
	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: createResp.State}, deleteResp)
	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete: %s", deleteResp.Diagnostics)
	}

	// Verify deleted — Read should remove state
	cfg := nullAuditObjectStorePolicyMemberConfig()
	cfg["policy_name"] = tftypes.NewValue(tftypes.String, "audit-pol")
	cfg["member_name"] = tftypes.NewValue(tftypes.String, "test-bucket")
	state := tfsdk.State{Raw: tftypes.NewValue(objType, cfg), Schema: s}
	readResp2 := &resource.ReadResponse{State: state}
	r.Read(context.Background(), resource.ReadRequest{State: state}, readResp2)
	if !readResp2.State.Raw.IsNull() {
		t.Error("expected state to be removed after delete")
	}
}

func TestUnit_AuditObjectStorePolicyMemberResource_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterAuditObjectStorePolicyHandlers(ms.Mux)

	store.Seed(&client.AuditObjectStorePolicy{
		ID: "pol-2", Name: "imp-policy", Enabled: true, IsLocal: true, PolicyType: "audit",
	})
	store.SeedMember("imp-policy", client.AuditObjectStorePolicyMember{
		Member: client.NamedReference{Name: "imp-bucket"},
		Policy: client.NamedReference{Name: "imp-policy"},
	})

	r := newTestAuditObjectStorePolicyMemberResource(t, ms)
	s := auditObjectStorePolicyMemberResourceSchema(t).Schema

	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildAuditObjectStorePolicyMemberType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "imp-policy/imp-bucket"}, importResp)

	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState: %s", importResp.Diagnostics)
	}

	var model auditObjectStorePolicyMemberModel
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
