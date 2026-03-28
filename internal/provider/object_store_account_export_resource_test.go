package provider

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/numberly/opentofu-provider-flashblade/internal/client"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock/handlers"
)

// ---- helpers ----------------------------------------------------------------

// newTestAccountExportResource creates an objectStoreAccountExportResource wired to the given mock server.
func newTestAccountExportResource(t *testing.T, ms *testmock.MockServer) *objectStoreAccountExportResource {
	t.Helper()
	c, err := client.NewClient(client.Config{
		Endpoint:           ms.URL(),
		APIToken:           "test-token",
		InsecureSkipVerify: true,
		MaxRetries:         1,
		RetryBaseDelay:     1 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return &objectStoreAccountExportResource{client: c}
}

// accountExportResourceSchema returns the parsed schema for the object store account export resource.
func accountExportResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &objectStoreAccountExportResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

// buildAccountExportType returns the tftypes.Object for the account export resource schema.
func buildAccountExportType() tftypes.Object {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":           tftypes.String,
		"name":         tftypes.String,
		"account_name": tftypes.String,
		"server_name":  tftypes.String,
		"enabled":      tftypes.Bool,
		"policy_name":  tftypes.String,
		"timeouts":     timeoutsType,
	}}
}

// nullAccountExportConfig returns a base config map with all attributes null.
func nullAccountExportConfig() map[string]tftypes.Value {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return map[string]tftypes.Value{
		"id":           tftypes.NewValue(tftypes.String, nil),
		"name":         tftypes.NewValue(tftypes.String, nil),
		"account_name": tftypes.NewValue(tftypes.String, nil),
		"server_name":  tftypes.NewValue(tftypes.String, nil),
		"enabled":      tftypes.NewValue(tftypes.Bool, nil),
		"policy_name":  tftypes.NewValue(tftypes.String, nil),
		"timeouts":     tftypes.NewValue(timeoutsType, nil),
	}
}

// accountExportPlanWith returns a tfsdk.Plan with account_name, server_name, policy_name, and enabled.
func accountExportPlanWith(t *testing.T, accountName, serverName, policyName string, enabled bool) tfsdk.Plan {
	t.Helper()
	s := accountExportResourceSchema(t).Schema
	cfg := nullAccountExportConfig()
	cfg["account_name"] = tftypes.NewValue(tftypes.String, accountName)
	cfg["server_name"] = tftypes.NewValue(tftypes.String, serverName)
	cfg["policy_name"] = tftypes.NewValue(tftypes.String, policyName)
	cfg["enabled"] = tftypes.NewValue(tftypes.Bool, enabled)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildAccountExportType(), cfg),
		Schema: s,
	}
}

// ---- tests ------------------------------------------------------------------

// TestUnit_AccountExport_Create verifies Create populates all expected attributes.
func TestUnit_AccountExport_Create(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterObjectStoreAccountExportHandlers(ms.Mux)

	r := newTestAccountExportResource(t, ms)
	s := accountExportResourceSchema(t).Schema

	plan := accountExportPlanWith(t, "test-account", "server1", "s3-policy", true)
	req := resource.CreateRequest{Plan: plan}
	resp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildAccountExportType(), nil), Schema: s},
	}

	r.Create(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Create returned error: %s", resp.Diagnostics)
	}

	var model objectStoreAccountExportModel
	if diags := resp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected non-empty ID after Create")
	}
	if model.Name.ValueString() != "test-account/test-account" {
		t.Errorf("expected name=test-account/test-account, got %s", model.Name.ValueString())
	}
	if model.AccountName.ValueString() != "test-account" {
		t.Errorf("expected account_name=test-account, got %s", model.AccountName.ValueString())
	}
	if model.ServerName.ValueString() != "server1" {
		t.Errorf("expected server_name=server1, got %s", model.ServerName.ValueString())
	}
	if !model.Enabled.ValueBool() {
		t.Error("expected enabled=true after Create")
	}
	if model.PolicyName.ValueString() != "s3-policy" {
		t.Errorf("expected policy_name=s3-policy, got %s", model.PolicyName.ValueString())
	}
}

// TestUnit_AccountExport_Read verifies Read populates all attributes from a seeded export.
func TestUnit_AccountExport_Read(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterObjectStoreAccountExportHandlers(ms.Mux)

	seeded := store.AddObjectStoreAccountExport("read-account", "s3-policy", "server1")

	r := newTestAccountExportResource(t, ms)
	s := accountExportResourceSchema(t).Schema

	cfg := nullAccountExportConfig()
	cfg["id"] = tftypes.NewValue(tftypes.String, seeded.ID)
	cfg["name"] = tftypes.NewValue(tftypes.String, seeded.Name)
	cfg["account_name"] = tftypes.NewValue(tftypes.String, "read-account")
	state := tfsdk.State{Raw: tftypes.NewValue(buildAccountExportType(), cfg), Schema: s}

	readResp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), resource.ReadRequest{State: state}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}

	var model objectStoreAccountExportModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Name.ValueString() != "read-account/read-account" {
		t.Errorf("expected name=read-account/read-account, got %s", model.Name.ValueString())
	}
	if model.AccountName.ValueString() != "read-account" {
		t.Errorf("expected account_name=read-account, got %s", model.AccountName.ValueString())
	}
	if model.ServerName.ValueString() != "server1" {
		t.Errorf("expected server_name=server1, got %s", model.ServerName.ValueString())
	}
	if model.PolicyName.ValueString() != "s3-policy" {
		t.Errorf("expected policy_name=s3-policy, got %s", model.PolicyName.ValueString())
	}
	if !model.Enabled.ValueBool() {
		t.Error("expected enabled=true")
	}
}

// TestUnit_AccountExport_Update verifies PATCH updates enabled=false.
func TestUnit_AccountExport_Update(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterObjectStoreAccountExportHandlers(ms.Mux)

	r := newTestAccountExportResource(t, ms)
	s := accountExportResourceSchema(t).Schema

	// Create first with enabled=true.
	createPlan := accountExportPlanWith(t, "update-account", "server1", "s3-policy", true)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildAccountExportType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Update enabled=false.
	updatePlan := accountExportPlanWith(t, "update-account", "server1", "s3-policy", false)
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildAccountExportType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlan,
		State: createResp.State,
	}, updateResp)

	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update returned error: %s", updateResp.Diagnostics)
	}

	var model objectStoreAccountExportModel
	if diags := updateResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Enabled.ValueBool() {
		t.Error("expected enabled=false after Update")
	}
}

// TestUnit_AccountExport_Delete verifies DELETE sends the short export name (not combined "account/export")
// to the API and removes the export without error.
// The mock uses strict lookup: member_names + "/" + names must match the stored combined key.
// If the resource passes the combined name as exportName, the mock will not find the export
// and the subsequent GET will still return it (proving the delete failed).
func TestUnit_AccountExport_Delete(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterObjectStoreAccountExportHandlers(ms.Mux)

	r := newTestAccountExportResource(t, ms)
	s := accountExportResourceSchema(t).Schema

	// Create first.
	createPlan := accountExportPlanWith(t, "delete-account", "server1", "s3-policy", true)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildAccountExportType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Delete.
	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: createResp.State}, deleteResp)

	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete returned error: %s", deleteResp.Diagnostics)
	}

	// Verify export is actually gone from the mock store.
	// If Delete sent the wrong name format, the strict mock would not have found/deleted
	// the export, and GET will still return it.
	_, err := r.client.GetObjectStoreAccountExport(context.Background(), "delete-account/delete-account")
	if err == nil || !client.IsNotFound(err) {
		t.Errorf("expected export to be deleted (strict mock), but GET still returns it: %v", err)
	}
}

// TestUnit_AccountExport_Delete_NoSlash verifies Delete with a name that has no "/" passes name as-is (defensive).
func TestUnit_AccountExport_Delete_NoSlash(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterObjectStoreAccountExportHandlers(ms.Mux)

	// Seed an export with a simple name (no slash) to simulate edge case.
	store.AddObjectStoreAccountExportWithName("simple-account", "simple-export", "s3-policy", "server1")

	r := newTestAccountExportResource(t, ms)
	s := accountExportResourceSchema(t).Schema

	// Build state with name="simple-export" (no slash).
	cfg := nullAccountExportConfig()
	cfg["id"] = tftypes.NewValue(tftypes.String, "test-id")
	cfg["name"] = tftypes.NewValue(tftypes.String, "simple-export")
	cfg["account_name"] = tftypes.NewValue(tftypes.String, "simple-account")
	cfg["server_name"] = tftypes.NewValue(tftypes.String, "server1")
	cfg["enabled"] = tftypes.NewValue(tftypes.Bool, true)
	cfg["policy_name"] = tftypes.NewValue(tftypes.String, "s3-policy")
	state := tfsdk.State{Raw: tftypes.NewValue(buildAccountExportType(), cfg), Schema: s}

	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: state}, deleteResp)

	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete returned error: %s", deleteResp.Diagnostics)
	}
}

// TestUnit_AccountExport_Import verifies ImportState populates all fields from combined name.
func TestUnit_AccountExport_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterObjectStoreAccountExportHandlers(ms.Mux)

	store.AddObjectStoreAccountExport("import-account", "s3-policy", "server1")

	r := newTestAccountExportResource(t, ms)
	s := accountExportResourceSchema(t).Schema

	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildAccountExportType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "import-account/import-account"}, importResp)

	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState returned error: %s", importResp.Diagnostics)
	}

	var model objectStoreAccountExportModel
	if diags := importResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Name.ValueString() != "import-account/import-account" {
		t.Errorf("expected name=import-account/import-account, got %s", model.Name.ValueString())
	}
	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected ID to be populated after import")
	}
	if model.AccountName.ValueString() != "import-account" {
		t.Errorf("expected account_name=import-account, got %s", model.AccountName.ValueString())
	}
	if model.ServerName.ValueString() != "server1" {
		t.Errorf("expected server_name=server1, got %s", model.ServerName.ValueString())
	}
	if model.PolicyName.ValueString() != "s3-policy" {
		t.Errorf("expected policy_name=s3-policy, got %s", model.PolicyName.ValueString())
	}
}

// TestUnit_AccountExport_NotFound verifies Read removes resource from state when not found.
func TestUnit_AccountExport_NotFound(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterObjectStoreAccountExportHandlers(ms.Mux)

	r := newTestAccountExportResource(t, ms)
	s := accountExportResourceSchema(t).Schema

	cfg := nullAccountExportConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "does-not-exist/does-not-exist")
	cfg["id"] = tftypes.NewValue(tftypes.String, "non-existent-id")
	cfg["account_name"] = tftypes.NewValue(tftypes.String, "does-not-exist")
	state := tfsdk.State{Raw: tftypes.NewValue(buildAccountExportType(), cfg), Schema: s}

	readResp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), resource.ReadRequest{State: state}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}
	if !readResp.State.Raw.IsNull() {
		t.Error("expected state to be removed (null) when resource not found")
	}
}
