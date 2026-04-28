package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/numberly/terraform-provider-mica/internal/client"
	"github.com/numberly/terraform-provider-mica/internal/testmock"
	"github.com/numberly/terraform-provider-mica/internal/testmock/handlers"
)

func newTestDirectoryServiceRoleDataSource(t *testing.T, ms *testmock.MockServer) *directoryServiceRoleDataSource {
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
	return &directoryServiceRoleDataSource{client: c}
}

func directoryServiceRoleDSSchema(t *testing.T) datasource.SchemaResponse {
	t.Helper()
	d := &directoryServiceRoleDataSource{}
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	return resp
}

func buildDirectoryServiceRoleDSType() tftypes.Object {
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
	}}
}

func TestUnit_DirectoryServiceRoleDataSource_Basic(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterDirectoryServiceRolesHandlers(ms.Mux)

	// Seed a role with all computed fields populated.
	store.Seed(&client.DirectoryServiceRole{
		ID:        "dsr-seed-1",
		Name:      "array_admin",
		Group:     "cn=fb-admins,ou=groups,dc=corp",
		GroupBase: "ou=groups,dc=corp",
		ManagementAccessPolicies: []client.NamedReference{
			{Name: "pure:policy/array_admin"},
		},
		Role: &client.NamedReference{Name: "array_admin"},
	})

	d := newTestDirectoryServiceRoleDataSource(t, ms)
	s := directoryServiceRoleDSSchema(t).Schema

	roleType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"name": tftypes.String}}

	cfgVal := tftypes.NewValue(buildDirectoryServiceRoleDSType(), map[string]tftypes.Value{
		"id":                         tftypes.NewValue(tftypes.String, nil),
		"name":                       tftypes.NewValue(tftypes.String, "array_admin"),
		"group":                      tftypes.NewValue(tftypes.String, nil),
		"group_base":                 tftypes.NewValue(tftypes.String, nil),
		"management_access_policies": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"role":                       tftypes.NewValue(roleType, nil),
	})
	config := tfsdk.Config{Schema: s, Raw: cfgVal}
	state := tfsdk.State{Schema: s}
	readResp := &datasource.ReadResponse{State: state}
	d.Read(context.Background(), datasource.ReadRequest{Config: config}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors: %v", readResp.Diagnostics.Errors())
	}

	var got directoryServiceRoleDataSourceModel
	if diags := readResp.State.Get(context.Background(), &got); diags.HasError() {
		t.Fatalf("state.Get: %v", diags.Errors())
	}
	if got.ID.ValueString() != "dsr-seed-1" {
		t.Errorf("id: got %q, want %q", got.ID.ValueString(), "dsr-seed-1")
	}
	if got.Name.ValueString() != "array_admin" {
		t.Errorf("name: got %q, want %q", got.Name.ValueString(), "array_admin")
	}
	if got.Group.ValueString() != "cn=fb-admins,ou=groups,dc=corp" {
		t.Errorf("group: got %q", got.Group.ValueString())
	}
	if got.GroupBase.ValueString() != "ou=groups,dc=corp" {
		t.Errorf("group_base: got %q", got.GroupBase.ValueString())
	}
	var policies []string
	_ = got.ManagementAccessPolicies.ElementsAs(context.Background(), &policies, false)
	if len(policies) != 1 || policies[0] != "pure:policy/array_admin" {
		t.Errorf("management_access_policies: got %v, want [pure:policy/array_admin]", policies)
	}
	if got.Role.IsNull() {
		t.Fatalf("role: expected non-null nested object")
	}
	roleAttrs := got.Role.Attributes()
	if v, ok := roleAttrs["name"].(types.String); !ok || v.ValueString() != "array_admin" {
		t.Errorf("role.name: got %v, want array_admin", roleAttrs["name"])
	}
}
