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

func newTestMapDsrMembershipResource(t *testing.T, ms *testmock.MockServer) *mapDsrMembershipResource {
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
	return &mapDsrMembershipResource{client: c}
}

func mapDsrMembershipResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &mapDsrMembershipResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

// buildMapDsrMembershipType returns the tftypes.Object for the DSRM resource schema.
// Must match the schema exactly: id, policy, role, timeouts (CRD — create/read/delete only).
func buildMapDsrMembershipType() tftypes.Object {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"delete": tftypes.String,
	}}
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":       tftypes.String,
		"policy":   tftypes.String,
		"role":     tftypes.String,
		"timeouts": timeoutsType,
	}}
}

// nullMapDsrMembershipConfig returns a base config map with all attributes null.
func nullMapDsrMembershipConfig() map[string]tftypes.Value {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"delete": tftypes.String,
	}}
	return map[string]tftypes.Value{
		"id":       tftypes.NewValue(tftypes.String, nil),
		"policy":   tftypes.NewValue(tftypes.String, nil),
		"role":     tftypes.NewValue(tftypes.String, nil),
		"timeouts": tftypes.NewValue(timeoutsType, nil),
	}
}

// mapDsrMembershipPlanWith returns a tfsdk.Plan with the given policy and role values.
func mapDsrMembershipPlanWith(t *testing.T, policy, role string) tfsdk.Plan {
	t.Helper()
	s := mapDsrMembershipResourceSchema(t).Schema
	cfg := nullMapDsrMembershipConfig()
	cfg["policy"] = tftypes.NewValue(tftypes.String, policy)
	cfg["role"] = tftypes.NewValue(tftypes.String, role)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildMapDsrMembershipType(), cfg),
		Schema: s,
	}
}

func TestUnit_ManagementAccessPolicyDirectoryServiceRoleMembershipResource_Lifecycle(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	_ = handlers.RegisterManagementAccessPolicyDirectoryServiceRoleMembershipsHandlers(ms.Mux)
	r := newTestMapDsrMembershipResource(t, ms)
	s := mapDsrMembershipResourceSchema(t).Schema

	// Create — associates pure:policy/storage_admin with admin-role
	plan := mapDsrMembershipPlanWith(t, "pure:policy/storage_admin", "admin-role")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{
			Raw:    tftypes.NewValue(buildMapDsrMembershipType(), nil),
			Schema: s,
		},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create returned error: %s", createResp.Diagnostics)
	}

	var afterCreate mapDsrMembershipModel
	if diags := createResp.State.Get(context.Background(), &afterCreate); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}
	if afterCreate.Policy.ValueString() != "pure:policy/storage_admin" {
		t.Errorf("expected policy=pure:policy/storage_admin, got %s", afterCreate.Policy.ValueString())
	}
	if afterCreate.Role.ValueString() != "admin-role" {
		t.Errorf("expected role=admin-role, got %s", afterCreate.Role.ValueString())
	}
	// D-05: composite ID = role_name/policy_name (role FIRST)
	wantID := "admin-role/pure:policy/storage_admin"
	if afterCreate.ID.ValueString() != wantID {
		t.Errorf("expected id=%s, got %s", wantID, afterCreate.ID.ValueString())
	}

	// Read — should find the association
	readResp := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}
	if readResp.State.Raw.IsNull() {
		t.Error("expected state to be populated after Read, got null")
	}

	var afterRead mapDsrMembershipModel
	if diags := readResp.State.Get(context.Background(), &afterRead); diags.HasError() {
		t.Fatalf("Get read state: %s", diags)
	}
	if afterRead.Policy.ValueString() != "pure:policy/storage_admin" {
		t.Errorf("expected policy=pure:policy/storage_admin after Read, got %s", afterRead.Policy.ValueString())
	}
	if afterRead.Role.ValueString() != "admin-role" {
		t.Errorf("expected role=admin-role after Read, got %s", afterRead.Role.ValueString())
	}

	// Delete — removes the association
	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: readResp.State}, deleteResp)
	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete returned error: %s", deleteResp.Diagnostics)
	}
}

func TestUnit_ManagementAccessPolicyDirectoryServiceRoleMembershipResource_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterManagementAccessPolicyDirectoryServiceRoleMembershipsHandlers(ms.Mux)
	// Seed the association before importing.
	store.Seed("pure:policy/array_admin", "admin-role")

	r := newTestMapDsrMembershipResource(t, ms)
	s := mapDsrMembershipResourceSchema(t).Schema

	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{
			Raw:    tftypes.NewValue(buildMapDsrMembershipType(), nil),
			Schema: s,
		},
	}
	// Import by composite ID "admin-role/pure:policy/array_admin" (role FIRST per D-05).
	// The embedded ":" and "/" in the policy name are handled correctly by SplitN("/", 2).
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "admin-role/pure:policy/array_admin"}, importResp)

	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState returned error: %s", importResp.Diagnostics)
	}

	var model mapDsrMembershipModel
	if diags := importResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get import state: %s", diags)
	}

	if model.Role.ValueString() != "admin-role" {
		t.Errorf("expected role=admin-role, got %s", model.Role.ValueString())
	}
	// Policy name must preserve the colon and slash: "pure:policy/array_admin"
	if model.Policy.ValueString() != "pure:policy/array_admin" {
		t.Errorf("expected policy=pure:policy/array_admin, got %s", model.Policy.ValueString())
	}
	wantID := "admin-role/pure:policy/array_admin"
	if model.ID.ValueString() != wantID {
		t.Errorf("expected id=%s, got %s", wantID, model.ID.ValueString())
	}
	if !model.Timeouts.IsNull() {
		t.Error("expected timeouts to be null after import")
	}
}

func TestUnit_ManagementAccessPolicyDirectoryServiceRoleMembershipResource_MissingAssociation(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	// NOTE: do NOT seed the membership — we are testing the "association removed
	// outside Terraform" path (DSRM-05). The mock GET will return an empty list + 200,
	// which getOneByName[T] translates into client.IsNotFound(err), which Read turns
	// into resp.State.RemoveResource(ctx).
	handlers.RegisterManagementAccessPolicyDirectoryServiceRoleMembershipsHandlers(ms.Mux)
	r := newTestMapDsrMembershipResource(t, ms)
	s := mapDsrMembershipResourceSchema(t).Schema

	// Build a prior state as if the membership had been previously imported.
	// The composite ID format is role_name/policy_name (role FIRST per D-05).
	stateVal := tftypes.NewValue(buildMapDsrMembershipType(), map[string]tftypes.Value{
		"policy": tftypes.NewValue(tftypes.String, "pure:policy/array_admin"),
		"role":   tftypes.NewValue(tftypes.String, "array_admin"),
		"id":     tftypes.NewValue(tftypes.String, "array_admin/pure:policy/array_admin"),
		"timeouts": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
			"create": tftypes.String,
			"read":   tftypes.String,
			"delete": tftypes.String,
		}}, nil),
	})
	priorState := tfsdk.State{Schema: s, Raw: stateVal}
	readResp := &resource.ReadResponse{State: priorState}
	r.Read(context.Background(), resource.ReadRequest{State: priorState}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors: %v", readResp.Diagnostics.Errors())
	}
	// RemoveResource sets state.Raw to a null value — this is how the framework
	// signals "the resource no longer exists and should be dropped from state".
	if !readResp.State.Raw.IsNull() {
		t.Errorf("expected state.Raw to be null after RemoveResource, got: %v", readResp.State.Raw)
	}
}
