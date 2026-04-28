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

func newTestDirectoryServiceRoleResource(t *testing.T, ms *testmock.MockServer) *directoryServiceRoleResource {
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
	return &directoryServiceRoleResource{client: c}
}

func directoryServiceRoleResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &directoryServiceRoleResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

func buildDirectoryServiceRoleType() tftypes.Object {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	roleType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"name": tftypes.String,
	}}
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":                         tftypes.String,
		"name":                       tftypes.String,
		"group":                      tftypes.String,
		"group_base":                 tftypes.String,
		"management_access_policies": tftypes.List{ElementType: tftypes.String},
		"role":                       roleType,
		"timeouts":                   timeoutsType,
	}}
}

func nullDirectoryServiceRoleConfig() map[string]tftypes.Value {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	roleType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"name": tftypes.String,
	}}
	return map[string]tftypes.Value{
		"id":                         tftypes.NewValue(tftypes.String, nil),
		"name":                       tftypes.NewValue(tftypes.String, nil),
		"group":                      tftypes.NewValue(tftypes.String, nil),
		"group_base":                 tftypes.NewValue(tftypes.String, nil),
		"management_access_policies": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"role":                       tftypes.NewValue(roleType, nil),
		"timeouts":                   tftypes.NewValue(timeoutsType, nil),
	}
}

func directoryServiceRolePlanWith(t *testing.T, name, group, groupBase string, policies []string) tfsdk.Plan {
	t.Helper()
	s := directoryServiceRoleResourceSchema(t).Schema
	cfg := nullDirectoryServiceRoleConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, name)
	cfg["group"] = tftypes.NewValue(tftypes.String, group)
	cfg["group_base"] = tftypes.NewValue(tftypes.String, groupBase)

	policyVals := make([]tftypes.Value, len(policies))
	for i, p := range policies {
		policyVals[i] = tftypes.NewValue(tftypes.String, p)
	}
	cfg["management_access_policies"] = tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, policyVals)

	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildDirectoryServiceRoleType(), cfg),
		Schema: s,
	}
}

func TestUnit_DirectoryServiceRoleResource_Lifecycle(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	_ = handlers.RegisterDirectoryServiceRolesHandlers(ms.Mux)

	r := newTestDirectoryServiceRoleResource(t, ms)
	s := directoryServiceRoleResourceSchema(t).Schema

	// Create — name is user-supplied per v1 schema (D-04/D-08).
	plan := directoryServiceRolePlanWith(t, "infra-admins", "cn=admins,ou=groups,dc=corp", "ou=groups,dc=corp", []string{"pure:policy/array_admin"})
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildDirectoryServiceRoleType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create returned error: %v", createResp.Diagnostics)
	}

	var afterCreate directoryServiceRoleModel
	if diags := createResp.State.Get(context.Background(), &afterCreate); diags.HasError() {
		t.Fatalf("Get create state: %v", diags)
	}
	if afterCreate.ID.IsNull() || afterCreate.ID.ValueString() == "" {
		t.Error("expected non-empty id after Create")
	}
	// Name is user-supplied via plan (POST /directory-services/roles?names=infra-admins).
	if afterCreate.Name.ValueString() != "infra-admins" {
		t.Errorf("expected name=infra-admins, got %q", afterCreate.Name.ValueString())
	}
	if afterCreate.Group.ValueString() != "cn=admins,ou=groups,dc=corp" {
		t.Errorf("expected group=cn=admins, got %q", afterCreate.Group.ValueString())
	}
	if afterCreate.GroupBase.ValueString() != "ou=groups,dc=corp" {
		t.Errorf("expected group_base=ou=groups,dc=corp, got %q", afterCreate.GroupBase.ValueString())
	}

	// Read
	readResp := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %v", readResp.Diagnostics)
	}

	var afterRead directoryServiceRoleModel
	if diags := readResp.State.Get(context.Background(), &afterRead); diags.HasError() {
		t.Fatalf("Get read state: %v", diags)
	}
	if afterRead.Group.ValueString() != afterCreate.Group.ValueString() {
		t.Errorf("group drift on Read: create=%q read=%q", afterCreate.Group.ValueString(), afterRead.Group.ValueString())
	}

	// Update (change group) — name unchanged (immutable, RequiresReplace).
	updatePlan := directoryServiceRolePlanWith(t, "infra-admins", "cn=new-admins,ou=groups,dc=corp", "ou=groups,dc=corp", []string{"pure:policy/array_admin"})
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildDirectoryServiceRoleType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlan,
		State: readResp.State,
	}, updateResp)
	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update returned error: %v", updateResp.Diagnostics)
	}

	var afterUpdate directoryServiceRoleModel
	if diags := updateResp.State.Get(context.Background(), &afterUpdate); diags.HasError() {
		t.Fatalf("Get update state: %v", diags)
	}
	if afterUpdate.Group.ValueString() != "cn=new-admins,ou=groups,dc=corp" {
		t.Errorf("expected group=cn=new-admins after update, got %q", afterUpdate.Group.ValueString())
	}

	// Delete
	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: updateResp.State}, deleteResp)
	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete returned error: %v", deleteResp.Diagnostics)
	}

	// Confirm deletion
	_, err := r.client.GetDirectoryServiceRole(context.Background(), "infra-admins")
	if err == nil || !client.IsNotFound(err) {
		t.Errorf("expected role to be deleted, got: %v", err)
	}
}

func TestUnit_DirectoryServiceRoleResource_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterDirectoryServiceRolesHandlers(ms.Mux)

	// Seed a role directly into the mock store.
	store.Seed(&client.DirectoryServiceRole{
		ID:        "dsr-import-1",
		Name:      "array_admin",
		Group:     "cn=fb-admins,ou=groups,dc=corp",
		GroupBase: "ou=groups,dc=corp",
		ManagementAccessPolicies: []client.NamedReference{
			{Name: "pure:policy/array_admin"},
		},
	})

	r := newTestDirectoryServiceRoleResource(t, ms)
	s := directoryServiceRoleResourceSchema(t).Schema

	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildDirectoryServiceRoleType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "array_admin"}, importResp)
	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState returned error: %v", importResp.Diagnostics)
	}

	var model directoryServiceRoleModel
	if diags := importResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get import state: %v", diags)
	}
	if model.ID.ValueString() != "dsr-import-1" {
		t.Errorf("expected id=dsr-import-1, got %q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "array_admin" {
		t.Errorf("expected name=array_admin, got %q", model.Name.ValueString())
	}
	if model.Group.ValueString() != "cn=fb-admins,ou=groups,dc=corp" {
		t.Errorf("expected group=cn=fb-admins, got %q", model.Group.ValueString())
	}
	if model.GroupBase.ValueString() != "ou=groups,dc=corp" {
		t.Errorf("expected group_base=ou=groups,dc=corp, got %q", model.GroupBase.ValueString())
	}
	// Timeouts must be null after import (no plan available).
	if !model.Timeouts.IsNull() {
		t.Error("expected null timeouts after import")
	}
}

func TestUnit_DirectoryServiceRoleResource_DriftDetection(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterDirectoryServiceRolesHandlers(ms.Mux)

	// Seed initial state.
	store.Seed(&client.DirectoryServiceRole{
		ID:        "dsr-drift-1",
		Name:      "array_admin",
		Group:     "cn=original,ou=groups,dc=corp",
		GroupBase: "ou=groups,dc=corp",
		ManagementAccessPolicies: []client.NamedReference{
			{Name: "pure:policy/array_admin"},
		},
	})

	r := newTestDirectoryServiceRoleResource(t, ms)
	s := directoryServiceRoleResourceSchema(t).Schema

	// Build initial state matching seed.
	plan := directoryServiceRolePlanWith(t, "array_admin", "cn=original,ou=groups,dc=corp", "ou=groups,dc=corp", []string{"pure:policy/array_admin"})
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildDirectoryServiceRoleType(), nil), Schema: s},
	}
	// Use Read path directly: build a state with the seeded values.
	// We construct state manually matching the seed.
	cfg := nullDirectoryServiceRoleConfig()
	cfg["id"] = tftypes.NewValue(tftypes.String, "dsr-drift-1")
	cfg["name"] = tftypes.NewValue(tftypes.String, "array_admin")
	cfg["group"] = tftypes.NewValue(tftypes.String, "cn=original,ou=groups,dc=corp")
	cfg["group_base"] = tftypes.NewValue(tftypes.String, "ou=groups,dc=corp")
	cfg["management_access_policies"] = tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
		tftypes.NewValue(tftypes.String, "pure:policy/array_admin"),
	})
	initialState := tfsdk.State{
		Raw:    tftypes.NewValue(buildDirectoryServiceRoleType(), cfg),
		Schema: s,
	}
	_ = plan
	_ = createResp

	// Now mutate the mock store to simulate out-of-band change.
	store.Seed(&client.DirectoryServiceRole{
		ID:        "dsr-drift-1",
		Name:      "array_admin",
		Group:     "cn=mutated,ou=groups,dc=corp", // changed outside Terraform
		GroupBase: "ou=groups,dc=corp",
		ManagementAccessPolicies: []client.NamedReference{
			{Name: "pure:policy/array_admin"},
		},
	})

	// Read should detect drift and update state to reflect the new value.
	readResp := &resource.ReadResponse{State: initialState}
	r.Read(context.Background(), resource.ReadRequest{State: initialState}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read drift detection: %v", readResp.Diagnostics)
	}

	var afterDrift directoryServiceRoleModel
	if diags := readResp.State.Get(context.Background(), &afterDrift); diags.HasError() {
		t.Fatalf("Get drift state: %v", diags)
	}
	if afterDrift.Group.ValueString() != "cn=mutated,ou=groups,dc=corp" {
		t.Errorf("expected state to reflect drifted group=cn=mutated, got %q", afterDrift.Group.ValueString())
	}
}

// TestUnit_DirectoryServiceRoleResource_StateUpgrade_V0toV1 verifies that a v0 state
// (name was Computed, populated server-side) is correctly carried forward verbatim
// into the v1 schema where name is Required.
func TestUnit_DirectoryServiceRoleResource_StateUpgrade_V0toV1(t *testing.T) {
	r := &directoryServiceRoleResource{}
	upgraders := r.UpgradeState(context.Background())

	upgrader, ok := upgraders[0]
	if !ok {
		t.Fatal("expected v0->v1 upgrader at key 0")
	}
	if upgrader.PriorSchema == nil {
		t.Fatal("expected PriorSchema to be set for v0->v1 upgrader")
	}

	// Build a v0 state tftypes value using the same shape as the PriorSchema.
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	roleType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"name": tftypes.String,
	}}
	v0Type := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":                         tftypes.String,
		"name":                       tftypes.String,
		"group":                      tftypes.String,
		"group_base":                 tftypes.String,
		"management_access_policies": tftypes.List{ElementType: tftypes.String},
		"role":                       roleType,
		"timeouts":                   timeoutsType,
	}}
	v0Val := tftypes.NewValue(v0Type, map[string]tftypes.Value{
		"id":         tftypes.NewValue(tftypes.String, "dsr-1"),
		"name":       tftypes.NewValue(tftypes.String, "legacy-role"),
		"group":      tftypes.NewValue(tftypes.String, "cn=admins,ou=groups,dc=corp"),
		"group_base": tftypes.NewValue(tftypes.String, "ou=groups,dc=corp"),
		"management_access_policies": tftypes.NewValue(
			tftypes.List{ElementType: tftypes.String},
			[]tftypes.Value{tftypes.NewValue(tftypes.String, "pure:policy/array_admin")},
		),
		"role": tftypes.NewValue(roleType, map[string]tftypes.Value{
			"name": tftypes.NewValue(tftypes.String, "array_admin"),
		}),
		"timeouts": tftypes.NewValue(timeoutsType, nil),
	})

	priorState := tfsdk.State{
		Raw:    v0Val,
		Schema: *upgrader.PriorSchema,
	}

	req := resource.UpgradeStateRequest{State: &priorState}
	resp := &resource.UpgradeStateResponse{
		State: tfsdk.State{
			Raw:    tftypes.NewValue(buildDirectoryServiceRoleType(), nil),
			Schema: directoryServiceRoleResourceSchema(t).Schema,
		},
	}

	upgrader.StateUpgrader(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("StateUpgrader returned error: %s", resp.Diagnostics)
	}

	var model directoryServiceRoleModel
	if diags := resp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get upgraded state: %s", diags)
	}

	// All fields must be carried forward verbatim.
	if model.ID.ValueString() != "dsr-1" {
		t.Errorf("id: want %q got %q", "dsr-1", model.ID.ValueString())
	}
	if model.Name.ValueString() != "legacy-role" {
		t.Errorf("name: want %q got %q", "legacy-role", model.Name.ValueString())
	}
	if model.Group.ValueString() != "cn=admins,ou=groups,dc=corp" {
		t.Errorf("group: want %q got %q", "cn=admins,ou=groups,dc=corp", model.Group.ValueString())
	}
	if model.GroupBase.ValueString() != "ou=groups,dc=corp" {
		t.Errorf("group_base: want %q got %q", "ou=groups,dc=corp", model.GroupBase.ValueString())
	}

	var policies []string
	if diags := model.ManagementAccessPolicies.ElementsAs(context.Background(), &policies, false); diags.HasError() {
		t.Fatalf("ElementsAs policies: %s", diags)
	}
	if len(policies) != 1 || policies[0] != "pure:policy/array_admin" {
		t.Errorf("management_access_policies: want [pure:policy/array_admin] got %v", policies)
	}

	// role nested object must be preserved.
	if model.Role.IsNull() {
		t.Error("expected role object to be preserved, got null")
	}
}
