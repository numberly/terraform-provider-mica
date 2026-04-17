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

// newTestDirectoryServiceManagementDataSource creates a directoryServiceManagementDataSource
// wired to the given mock server.
func newTestDirectoryServiceManagementDataSource(t *testing.T, ms *testmock.MockServer) *directoryServiceManagementDataSource {
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
	return &directoryServiceManagementDataSource{client: c}
}

// directoryServiceManagementDSSchema returns the parsed schema for the data source.
func directoryServiceManagementDSSchema(t *testing.T) datasource.SchemaResponse {
	t.Helper()
	d := &directoryServiceManagementDataSource{}
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	return resp
}

// buildDirectoryServiceManagementDSType returns the tftypes.Object for the data source schema.
func buildDirectoryServiceManagementDSType() tftypes.Object {
	refType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"name": tftypes.String}}
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":                       tftypes.String,
		"enabled":                  tftypes.Bool,
		"uris":                     tftypes.List{ElementType: tftypes.String},
		"base_dn":                  tftypes.String,
		"bind_user":                tftypes.String,
		"ca_certificate":           refType,
		"ca_certificate_group":     refType,
		"user_login_attribute":     tftypes.String,
		"user_object_class":        tftypes.String,
		"ssh_public_key_attribute": tftypes.String,
		"services":                 tftypes.List{ElementType: tftypes.String},
	}}
}

// nullDirectoryServiceManagementDSConfig returns a base config map with all attributes null.
func nullDirectoryServiceManagementDSConfig() map[string]tftypes.Value {
	refType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"name": tftypes.String}}
	return map[string]tftypes.Value{
		"id":                       tftypes.NewValue(tftypes.String, nil),
		"enabled":                  tftypes.NewValue(tftypes.Bool, nil),
		"uris":                     tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"base_dn":                  tftypes.NewValue(tftypes.String, nil),
		"bind_user":                tftypes.NewValue(tftypes.String, nil),
		"ca_certificate":           tftypes.NewValue(refType, nil),
		"ca_certificate_group":     tftypes.NewValue(refType, nil),
		"user_login_attribute":     tftypes.NewValue(tftypes.String, nil),
		"user_object_class":        tftypes.NewValue(tftypes.String, nil),
		"ssh_public_key_attribute": tftypes.NewValue(tftypes.String, nil),
		"services":                 tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
	}
}

// TestUnit_DirectoryServiceManagementDataSource_Basic seeds the mock with a full
// DirectoryService and verifies that Read populates all schema fields correctly.
func TestUnit_DirectoryServiceManagementDataSource_Basic(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterDirectoryServicesHandlers(ms.Mux)

	store.Seed(&client.DirectoryService{
		Name:               "management",
		Enabled:            true,
		URIs:               []string{"ldaps://ldap.example.com:636"},
		BaseDN:             "dc=example,dc=com",
		BindUser:           "cn=admin",
		CACertificateGroup: &client.NamedReference{Name: "corp-ca"},
		Management: client.DirectoryServiceManagement{
			UserLoginAttribute:    "sAMAccountName",
			UserObjectClass:       "User",
			SSHPublicKeyAttribute: "sshPublicKey",
		},
		Services: []string{"management"},
	})

	d := newTestDirectoryServiceManagementDataSource(t, ms)
	s := directoryServiceManagementDSSchema(t).Schema
	objType := buildDirectoryServiceManagementDSType()
	cfg := nullDirectoryServiceManagementDSConfig()

	req := datasource.ReadRequest{
		Config: tfsdk.Config{
			Raw:    tftypes.NewValue(objType, cfg),
			Schema: s,
		},
	}
	resp := &datasource.ReadResponse{
		State: tfsdk.State{
			Raw:    tftypes.NewValue(objType, nil),
			Schema: s,
		},
	}
	d.Read(context.Background(), req, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Read: %s", resp.Diagnostics)
	}

	var got directoryServiceManagementDataSourceModel
	if diags := resp.State.Get(context.Background(), &got); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if got.Enabled.ValueBool() != true {
		t.Errorf("enabled: want true, got %v", got.Enabled.ValueBool())
	}
	if got.BaseDN.ValueString() != "dc=example,dc=com" {
		t.Errorf("base_dn: want %q, got %q", "dc=example,dc=com", got.BaseDN.ValueString())
	}
	if got.BindUser.ValueString() != "cn=admin" {
		t.Errorf("bind_user: want %q, got %q", "cn=admin", got.BindUser.ValueString())
	}
	if got.UserLoginAttribute.ValueString() != "sAMAccountName" {
		t.Errorf("user_login_attribute: want %q, got %q", "sAMAccountName", got.UserLoginAttribute.ValueString())
	}
	if got.UserObjectClass.ValueString() != "User" {
		t.Errorf("user_object_class: want %q, got %q", "User", got.UserObjectClass.ValueString())
	}
	if got.SSHPublicKeyAttribute.ValueString() != "sshPublicKey" {
		t.Errorf("ssh_public_key_attribute: want %q, got %q", "sshPublicKey", got.SSHPublicKeyAttribute.ValueString())
	}
	if got.URIs.IsNull() || len(got.URIs.Elements()) != 1 {
		t.Errorf("uris: want len 1, got %v", got.URIs)
	}
	if got.CACertificateGroup.IsNull() {
		t.Errorf("ca_certificate_group should be non-null (seeded with name corp-ca)")
	}
	if !got.CACertificate.IsNull() {
		t.Errorf("ca_certificate should be null (not seeded)")
	}
	if got.Services.IsNull() || len(got.Services.Elements()) != 1 {
		t.Errorf("services: want len 1, got %v", got.Services)
	}
}
