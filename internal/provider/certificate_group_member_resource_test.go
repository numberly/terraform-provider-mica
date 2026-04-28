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

// newTestCertificateGroupMemberResource creates a certificateGroupMemberResource wired to the given mock server.
func newTestCertificateGroupMemberResource(t *testing.T, ms *testmock.MockServer) *certificateGroupMemberResource {
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
	return &certificateGroupMemberResource{client: c}
}

// certificateGroupMemberResourceSchema returns the parsed schema for the certificate group member resource.
func certificateGroupMemberResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &certificateGroupMemberResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

// buildCertificateGroupMemberType returns the tftypes.Object for the certificate group member resource schema.
func buildCertificateGroupMemberType() tftypes.Object {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"delete": tftypes.String,
	}}
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"group_name":       tftypes.String,
		"certificate_name": tftypes.String,
		"timeouts":         timeoutsType,
	}}
}

// nullCertificateGroupMemberConfig returns a base config map with all attributes null.
func nullCertificateGroupMemberConfig() map[string]tftypes.Value {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"delete": tftypes.String,
	}}
	return map[string]tftypes.Value{
		"group_name":       tftypes.NewValue(tftypes.String, nil),
		"certificate_name": tftypes.NewValue(tftypes.String, nil),
		"timeouts":         tftypes.NewValue(timeoutsType, nil),
	}
}

// certificateGroupMemberPlanWith returns a tfsdk.Plan with the given field values.
func certificateGroupMemberPlanWith(t *testing.T, groupName, certName string) tfsdk.Plan {
	t.Helper()
	s := certificateGroupMemberResourceSchema(t).Schema
	cfg := nullCertificateGroupMemberConfig()
	cfg["group_name"] = tftypes.NewValue(tftypes.String, groupName)
	cfg["certificate_name"] = tftypes.NewValue(tftypes.String, certName)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildCertificateGroupMemberType(), cfg),
		Schema: s,
	}
}

// ---- tests ------------------------------------------------------------------

// TestUnit_CertificateGroupMemberResource_Lifecycle: create member → read → delete → verify gone.
func TestUnit_CertificateGroupMemberResource_Lifecycle(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	rawStore := handlers.RegisterCertificateGroupHandlers(ms.Mux)
	store := handlers.NewCertificateGroupStoreFacade(rawStore)

	// Seed a certificate group first.
	store.Seed(&client.CertificateGroup{
		ID:     "cg-lc-1",
		Name:   "my-group",
		Realms: []string{},
	})

	r := newTestCertificateGroupMemberResource(t, ms)
	s := certificateGroupMemberResourceSchema(t).Schema

	// Create.
	plan := certificateGroupMemberPlanWith(t, "my-group", "my-cert")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildCertificateGroupMemberType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create returned error: %s", createResp.Diagnostics)
	}

	var afterCreate certificateGroupMemberModel
	if diags := createResp.State.Get(context.Background(), &afterCreate); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}
	if afterCreate.GroupName.ValueString() != "my-group" {
		t.Errorf("expected group_name=my-group, got %s", afterCreate.GroupName.ValueString())
	}
	if afterCreate.CertName.ValueString() != "my-cert" {
		t.Errorf("expected certificate_name=my-cert, got %s", afterCreate.CertName.ValueString())
	}

	// Read (idempotent).
	readResp := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}

	var afterRead certificateGroupMemberModel
	if diags := readResp.State.Get(context.Background(), &afterRead); diags.HasError() {
		t.Fatalf("Get read state: %s", diags)
	}
	if afterRead.CertName.ValueString() != "my-cert" {
		t.Errorf("expected certificate_name=my-cert after Read, got %s", afterRead.CertName.ValueString())
	}

	// Delete.
	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: readResp.State}, deleteResp)
	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete returned error: %s", deleteResp.Diagnostics)
	}

	// Verify the member is gone.
	members, err := r.client.ListCertificateGroupMembers(context.Background(), "my-group")
	if err != nil {
		t.Fatalf("ListCertificateGroupMembers: %v", err)
	}
	for _, m := range members {
		if m.Certificate.Name == "my-cert" {
			t.Error("expected member to be deleted, but it still exists")
		}
	}
}

// TestUnit_CertificateGroupMemberResource_Import: seed group + member → import "groupName/certName" → verify state.
func TestUnit_CertificateGroupMemberResource_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	rawStore := handlers.RegisterCertificateGroupHandlers(ms.Mux)
	store := handlers.NewCertificateGroupStoreFacade(rawStore)

	// Seed group and member.
	store.Seed(&client.CertificateGroup{
		ID:     "cg-imp-1",
		Name:   "my-group",
		Realms: []string{},
	})
	store.SeedMember("my-group", client.CertificateGroupMember{
		Group:       client.NamedReference{Name: "my-group"},
		Certificate: client.NamedReference{Name: "my-cert"},
	})

	r := newTestCertificateGroupMemberResource(t, ms)
	s := certificateGroupMemberResourceSchema(t).Schema

	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildCertificateGroupMemberType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "my-group/my-cert"}, importResp)

	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState returned error: %s", importResp.Diagnostics)
	}

	var model certificateGroupMemberModel
	if diags := importResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get import state: %s", diags)
	}

	if model.GroupName.ValueString() != "my-group" {
		t.Errorf("expected group_name=my-group, got %s", model.GroupName.ValueString())
	}
	if model.CertName.ValueString() != "my-cert" {
		t.Errorf("expected certificate_name=my-cert, got %s", model.CertName.ValueString())
	}
	// Timeouts should be null for CRD-only import.
	if !model.Timeouts.IsNull() {
		t.Error("expected timeouts to be null after import")
	}
}
