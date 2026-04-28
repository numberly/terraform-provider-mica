package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/numberly/terraform-provider-mica/internal/client"
	"github.com/numberly/terraform-provider-mica/internal/testmock"
	"github.com/numberly/terraform-provider-mica/internal/testmock/handlers"
)

// ─── Test Helpers ──────────────────────────────────────────────────────────────

func newTestDirectoryServiceManagementResource(t *testing.T, ms *testmock.MockServer) *directoryServiceManagementResource {
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
	return &directoryServiceManagementResource{client: c}
}

func directoryServiceManagementResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &directoryServiceManagementResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

func buildDirectoryServiceManagementType() tftypes.Object {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":                      tftypes.String,
		"enabled":                 tftypes.Bool,
		"uris":                    tftypes.List{ElementType: tftypes.String},
		"base_dn":                 tftypes.String,
		"bind_user":               tftypes.String,
		"bind_password":           tftypes.String,
		"ca_certificate":          tftypes.String,
		"ca_certificate_group":    tftypes.String,
		"user_login_attribute":    tftypes.String,
		"user_object_class":       tftypes.String,
		"ssh_public_key_attribute": tftypes.String,
		"services":                tftypes.List{ElementType: tftypes.String},
		"timeouts":                timeoutsType,
	}}
}

func nullDirectoryServiceManagementConfig() map[string]tftypes.Value {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return map[string]tftypes.Value{
		"id":                      tftypes.NewValue(tftypes.String, nil),
		"enabled":                 tftypes.NewValue(tftypes.Bool, nil),
		"uris":                    tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"base_dn":                 tftypes.NewValue(tftypes.String, nil),
		"bind_user":               tftypes.NewValue(tftypes.String, nil),
		"bind_password":           tftypes.NewValue(tftypes.String, nil),
		"ca_certificate":          tftypes.NewValue(tftypes.String, nil),
		"ca_certificate_group":    tftypes.NewValue(tftypes.String, nil),
		"user_login_attribute":    tftypes.NewValue(tftypes.String, nil),
		"user_object_class":       tftypes.NewValue(tftypes.String, nil),
		"ssh_public_key_attribute": tftypes.NewValue(tftypes.String, nil),
		"services":                tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"timeouts":                tftypes.NewValue(timeoutsType, nil),
	}
}

// directoryServiceManagementPlanWith builds a tfsdk.Plan for the resource.
// Accepted keys in vals: "enabled" (bool), "uris" ([]string), "base_dn" (string),
// "bind_user" (string), "bind_password" (string), "user_login_attribute" (string).
func directoryServiceManagementPlanWith(t *testing.T, vals map[string]any) tfsdk.Plan {
	t.Helper()
	s := directoryServiceManagementResourceSchema(t).Schema
	cfg := nullDirectoryServiceManagementConfig()

	if v, ok := vals["enabled"]; ok {
		cfg["enabled"] = tftypes.NewValue(tftypes.Bool, v.(bool))
	}
	if v, ok := vals["uris"]; ok {
		strs := v.([]string)
		elems := make([]tftypes.Value, len(strs))
		for i, s := range strs {
			elems[i] = tftypes.NewValue(tftypes.String, s)
		}
		cfg["uris"] = tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, elems)
	}
	if v, ok := vals["base_dn"]; ok {
		cfg["base_dn"] = tftypes.NewValue(tftypes.String, v.(string))
	}
	if v, ok := vals["bind_user"]; ok {
		cfg["bind_user"] = tftypes.NewValue(tftypes.String, v.(string))
	}
	if v, ok := vals["bind_password"]; ok {
		cfg["bind_password"] = tftypes.NewValue(tftypes.String, v.(string))
	}
	if v, ok := vals["ca_certificate"]; ok {
		cfg["ca_certificate"] = tftypes.NewValue(tftypes.String, v.(string))
	}
	if v, ok := vals["ca_certificate_group"]; ok {
		cfg["ca_certificate_group"] = tftypes.NewValue(tftypes.String, v.(string))
	}
	if v, ok := vals["user_login_attribute"]; ok {
		cfg["user_login_attribute"] = tftypes.NewValue(tftypes.String, v.(string))
	}
	if v, ok := vals["user_object_class"]; ok {
		cfg["user_object_class"] = tftypes.NewValue(tftypes.String, v.(string))
	}
	if v, ok := vals["ssh_public_key_attribute"]; ok {
		cfg["ssh_public_key_attribute"] = tftypes.NewValue(tftypes.String, v.(string))
	}

	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildDirectoryServiceManagementType(), cfg),
		Schema: s,
	}
}

// ─── Tests ─────────────────────────────────────────────────────────────────────

func TestUnit_DirectoryServiceManagementResource_Lifecycle(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterDirectoryServicesHandlers(ms.Mux)

	// Seed so PATCH can find the singleton.
	store.Seed(&client.DirectoryService{Name: "management"})

	r := newTestDirectoryServiceManagementResource(t, ms)
	s := directoryServiceManagementResourceSchema(t).Schema

	// ── Create ──────────────────────────────────────────────────────────────
	plan := directoryServiceManagementPlanWith(t, map[string]any{
		"enabled":       true,
		"uris":          []string{"ldaps://ldap.example.com:636"},
		"base_dn":       "dc=example,dc=com",
		"bind_user":     "cn=binder,dc=example,dc=com",
		"bind_password": "s3cret",
	})
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildDirectoryServiceManagementType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	var afterCreate directoryServiceManagementModel
	if diags := createResp.State.Get(context.Background(), &afterCreate); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}
	if afterCreate.BaseDN.ValueString() != "dc=example,dc=com" {
		t.Errorf("base_dn: want %q, got %q", "dc=example,dc=com", afterCreate.BaseDN.ValueString())
	}
	if !afterCreate.Enabled.ValueBool() {
		t.Errorf("enabled: want true, got %v", afterCreate.Enabled.ValueBool())
	}
	if afterCreate.ID.IsNull() || afterCreate.ID.ValueString() == "" {
		t.Error("expected non-empty id after Create")
	}

	// ── Update: change base_dn ───────────────────────────────────────────────
	newPlan := directoryServiceManagementPlanWith(t, map[string]any{
		"enabled":       true,
		"uris":          []string{"ldaps://ldap.example.com:636"},
		"base_dn":       "dc=new,dc=com",
		"bind_user":     "cn=binder,dc=example,dc=com",
		"bind_password": "s3cret",
	})
	updateResp := &resource.UpdateResponse{
		State: createResp.State,
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  newPlan,
		State: createResp.State,
	}, updateResp)
	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update: %s", updateResp.Diagnostics)
	}
	var afterUpdate directoryServiceManagementModel
	if diags := updateResp.State.Get(context.Background(), &afterUpdate); diags.HasError() {
		t.Fatalf("Get update state: %s", diags)
	}
	if afterUpdate.BaseDN.ValueString() != "dc=new,dc=com" {
		t.Errorf("after update base_dn: want %q, got %q", "dc=new,dc=com", afterUpdate.BaseDN.ValueString())
	}

	// ── Delete → verify store was reset ─────────────────────────────────────
	delResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: updateResp.State}, delResp)
	if delResp.Diagnostics.HasError() {
		t.Fatalf("Delete: %s", delResp.Diagnostics)
	}

	// Verify store state via GET.
	ds, err := r.client.GetDirectoryServiceManagement(context.Background(), "management")
	if err != nil {
		t.Fatalf("post-delete GET: %v", err)
	}
	if ds.Enabled {
		t.Errorf("after delete enabled: want false, got %v", ds.Enabled)
	}
	if len(ds.URIs) != 0 {
		t.Errorf("after delete uris: want empty, got %v", ds.URIs)
	}
	if ds.BaseDN != "" {
		t.Errorf("after delete base_dn: want empty, got %q", ds.BaseDN)
	}
}

func TestUnit_DirectoryServiceManagementResource_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterDirectoryServicesHandlers(ms.Mux)
	store.Seed(&client.DirectoryService{
		Name:     "management",
		Enabled:  true,
		URIs:     []string{"ldaps://ldap.example.com:636"},
		BaseDN:   "dc=example,dc=com",
		BindUser: "cn=admin,dc=example,dc=com",
		Management: client.DirectoryServiceManagement{
			UserLoginAttribute: "sAMAccountName",
		},
	})

	r := newTestDirectoryServiceManagementResource(t, ms)
	s := directoryServiceManagementResourceSchema(t).Schema

	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildDirectoryServiceManagementType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "management"}, importResp)
	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState: %s", importResp.Diagnostics)
	}

	var data directoryServiceManagementModel
	if diags := importResp.State.Get(context.Background(), &data); diags.HasError() {
		t.Fatalf("Get import state: %s", diags)
	}

	if data.BaseDN.ValueString() != "dc=example,dc=com" {
		t.Errorf("base_dn: want %q, got %q", "dc=example,dc=com", data.BaseDN.ValueString())
	}
	if !data.Enabled.ValueBool() {
		t.Errorf("enabled: want true, got %v", data.Enabled.ValueBool())
	}
	if data.UserLoginAttribute.ValueString() != "sAMAccountName" {
		t.Errorf("user_login_attribute: want %q, got %q", "sAMAccountName", data.UserLoginAttribute.ValueString())
	}
	if data.BindPassword.ValueString() != "" {
		t.Errorf("bind_password: want empty, got %q", data.BindPassword.ValueString())
	}
	if data.ID.IsNull() || data.ID.ValueString() == "" {
		t.Error("expected non-empty id after import")
	}
	// Timeouts should be null (nullTimeoutsValue).
	if !data.Timeouts.IsNull() {
		t.Logf("note: timeouts after import: %+v (expected null)", data.Timeouts)
	}
}

func TestUnit_DirectoryServiceManagementResource_DriftDetection(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterDirectoryServicesHandlers(ms.Mux)
	store.Seed(&client.DirectoryService{
		Name:    "management",
		BaseDN:  "dc=old,dc=com",
		Enabled: true,
	})

	r := newTestDirectoryServiceManagementResource(t, ms)
	s := directoryServiceManagementResourceSchema(t).Schema

	// Build initial state matching the seed.
	initial := directoryServiceManagementModel{
		ID:                    types.StringValue("ds-1"),
		Enabled:               types.BoolValue(true),
		BaseDN:                types.StringValue("dc=old,dc=com"),
		BindUser:              types.StringValue(""),
		BindPassword:          types.StringValue(""),
		CACertificate:         types.StringValue(""),
		CACertificateGroup:    types.StringValue(""),
		UserLoginAttribute:    types.StringValue(""),
		UserObjectClass:       types.StringValue(""),
		SSHPublicKeyAttribute: types.StringValue(""),
		URIs:                  emptyStringList(),
		Services:              emptyStringList(),
		Timeouts:              nullTimeoutsValue(),
	}
	state := tfsdk.State{Raw: tftypes.NewValue(buildDirectoryServiceManagementType(), nil), Schema: s}
	if diags := state.Set(context.Background(), &initial); diags.HasError() {
		t.Fatalf("Set initial state: %s", diags)
	}

	// Mutate store out-of-band (simulate external drift).
	store.Seed(&client.DirectoryService{Name: "management", BaseDN: "dc=new,dc=com", Enabled: true})

	readResp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), resource.ReadRequest{State: state}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read: %s", readResp.Diagnostics)
	}

	var after directoryServiceManagementModel
	if diags := readResp.State.Get(context.Background(), &after); diags.HasError() {
		t.Fatalf("Get drift state: %s", diags)
	}
	if after.BaseDN.ValueString() != "dc=new,dc=com" {
		t.Errorf("post-drift base_dn: want %q, got %q", "dc=new,dc=com", after.BaseDN.ValueString())
	}
	// Confirm enabled still reflects API value.
	if !after.Enabled.ValueBool() {
		t.Errorf("post-drift enabled: want true, got %v", after.Enabled.ValueBool())
	}
}
