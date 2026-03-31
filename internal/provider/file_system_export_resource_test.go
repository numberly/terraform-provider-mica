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

// newTestFSExportResource creates a fileSystemExportResource wired to the given mock server.
func newTestFSExportResource(t *testing.T, ms *testmock.MockServer) *fileSystemExportResource {
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
	return &fileSystemExportResource{client: c}
}

// fsExportResourceSchema returns the parsed schema for the file system export resource.
func fsExportResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &fileSystemExportResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

// buildFSExportType returns the tftypes.Object for the file system export resource schema.
func buildFSExportType() tftypes.Object {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":                tftypes.String,
		"name":              tftypes.String,
		"export_name":       tftypes.String,
		"file_system_name":  tftypes.String,
		"server_name":       tftypes.String,
		"policy_name":       tftypes.String,
		"share_policy_name": tftypes.String,
		"enabled":           tftypes.Bool,
		"policy_type":       tftypes.String,
		"status":            tftypes.String,
		"timeouts":          timeoutsType,
	}}
}

// nullFSExportConfig returns a base config map with all attributes null.
func nullFSExportConfig() map[string]tftypes.Value {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return map[string]tftypes.Value{
		"id":                tftypes.NewValue(tftypes.String, nil),
		"name":              tftypes.NewValue(tftypes.String, nil),
		"export_name":       tftypes.NewValue(tftypes.String, nil),
		"file_system_name":  tftypes.NewValue(tftypes.String, nil),
		"server_name":       tftypes.NewValue(tftypes.String, nil),
		"policy_name":       tftypes.NewValue(tftypes.String, nil),
		"share_policy_name": tftypes.NewValue(tftypes.String, nil),
		"enabled":           tftypes.NewValue(tftypes.Bool, nil),
		"policy_type":       tftypes.NewValue(tftypes.String, nil),
		"status":            tftypes.NewValue(tftypes.String, nil),
		"timeouts":          tftypes.NewValue(timeoutsType, nil),
	}
}

// fsExportPlanWith returns a tfsdk.Plan with fs_name, server_name, and policy_name set.
func fsExportPlanWith(t *testing.T, fsName, serverName, policyName string) tfsdk.Plan {
	t.Helper()
	s := fsExportResourceSchema(t).Schema
	cfg := nullFSExportConfig()
	cfg["file_system_name"] = tftypes.NewValue(tftypes.String, fsName)
	cfg["server_name"] = tftypes.NewValue(tftypes.String, serverName)
	cfg["policy_name"] = tftypes.NewValue(tftypes.String, policyName)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildFSExportType(), cfg),
		Schema: s,
	}
}

// fsExportPlanWithSharePolicy returns a tfsdk.Plan with all required fields plus share_policy_name.
// export_name must be provided to match state and avoid spurious PATCH of export_name.
func fsExportPlanWithSharePolicy(t *testing.T, fsName, serverName, policyName, exportName, sharePolicyName string) tfsdk.Plan {
	t.Helper()
	s := fsExportResourceSchema(t).Schema
	cfg := nullFSExportConfig()
	cfg["file_system_name"] = tftypes.NewValue(tftypes.String, fsName)
	cfg["server_name"] = tftypes.NewValue(tftypes.String, serverName)
	cfg["policy_name"] = tftypes.NewValue(tftypes.String, policyName)
	cfg["export_name"] = tftypes.NewValue(tftypes.String, exportName)
	cfg["share_policy_name"] = tftypes.NewValue(tftypes.String, sharePolicyName)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildFSExportType(), cfg),
		Schema: s,
	}
}

// ---- tests ------------------------------------------------------------------

// TestUnit_FileSystemExport_Create verifies Create populates all expected attributes.
func TestUnit_FileSystemExport_Create(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterFileSystemExportHandlers(ms.Mux)

	r := newTestFSExportResource(t, ms)
	s := fsExportResourceSchema(t).Schema

	plan := fsExportPlanWith(t, "test-fs", "server1", "nfs-policy")
	req := resource.CreateRequest{Plan: plan}
	resp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildFSExportType(), nil), Schema: s},
	}

	r.Create(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Create returned error: %s", resp.Diagnostics)
	}

	var model fileSystemExportModel
	if diags := resp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected non-empty ID after Create")
	}
	if model.Name.ValueString() != "test-fs/test-fs" {
		t.Errorf("expected name=test-fs/test-fs, got %s", model.Name.ValueString())
	}
	if model.ExportName.ValueString() != "test-fs" {
		t.Errorf("expected export_name=test-fs, got %s", model.ExportName.ValueString())
	}
	if model.FileSystemName.ValueString() != "test-fs" {
		t.Errorf("expected file_system_name=test-fs, got %s", model.FileSystemName.ValueString())
	}
	if model.ServerName.ValueString() != "server1" {
		t.Errorf("expected server_name=server1, got %s", model.ServerName.ValueString())
	}
	if model.PolicyName.ValueString() != "nfs-policy" {
		t.Errorf("expected policy_name=nfs-policy, got %s", model.PolicyName.ValueString())
	}
	if !model.Enabled.ValueBool() {
		t.Error("expected enabled=true after Create")
	}
	if model.Status.ValueString() != "exported" {
		t.Errorf("expected status=exported, got %s", model.Status.ValueString())
	}
}

// TestUnit_FileSystemExport_Read verifies Read populates all attributes from a seeded export.
func TestUnit_FileSystemExport_Read(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterFileSystemExportHandlers(ms.Mux)

	seeded := store.AddFileSystemExport("read-fs", "nfs-policy", "server1")

	r := newTestFSExportResource(t, ms)
	s := fsExportResourceSchema(t).Schema

	cfg := nullFSExportConfig()
	cfg["id"] = tftypes.NewValue(tftypes.String, seeded.ID)
	cfg["name"] = tftypes.NewValue(tftypes.String, seeded.Name)
	cfg["file_system_name"] = tftypes.NewValue(tftypes.String, "read-fs")
	state := tfsdk.State{Raw: tftypes.NewValue(buildFSExportType(), cfg), Schema: s}

	readResp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), resource.ReadRequest{State: state}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}

	var model fileSystemExportModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Name.ValueString() != "read-fs/read-fs" {
		t.Errorf("expected name=read-fs/read-fs, got %s", model.Name.ValueString())
	}
	if model.ExportName.ValueString() != "read-fs" {
		t.Errorf("expected export_name=read-fs, got %s", model.ExportName.ValueString())
	}
	if model.ServerName.ValueString() != "server1" {
		t.Errorf("expected server_name=server1, got %s", model.ServerName.ValueString())
	}
	if model.PolicyName.ValueString() != "nfs-policy" {
		t.Errorf("expected policy_name=nfs-policy, got %s", model.PolicyName.ValueString())
	}
	if !model.Enabled.ValueBool() {
		t.Error("expected enabled=true")
	}
}

// TestUnit_FileSystemExport_Update verifies PATCH updates share_policy_name.
func TestUnit_FileSystemExport_Update(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterFileSystemExportHandlers(ms.Mux)

	r := newTestFSExportResource(t, ms)
	s := fsExportResourceSchema(t).Schema

	// Create first.
	createPlan := fsExportPlanWith(t, "update-fs", "server1", "nfs-policy")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildFSExportType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Update share_policy_name. export_name must match state to avoid spurious PATCH.
	updatePlan := fsExportPlanWithSharePolicy(t, "update-fs", "server1", "nfs-policy", "update-fs", "smb-share-policy")
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildFSExportType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlan,
		State: createResp.State,
	}, updateResp)

	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update returned error: %s", updateResp.Diagnostics)
	}

	var model fileSystemExportModel
	if diags := updateResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.SharePolicyName.ValueString() != "smb-share-policy" {
		t.Errorf("expected share_policy_name=smb-share-policy, got %s", model.SharePolicyName.ValueString())
	}
}

// TestUnit_FileSystemExport_Delete verifies DELETE removes the export without error.
func TestUnit_FileSystemExport_Delete(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterFileSystemExportHandlers(ms.Mux)

	r := newTestFSExportResource(t, ms)
	s := fsExportResourceSchema(t).Schema

	// Create first.
	createPlan := fsExportPlanWith(t, "delete-fs", "server1", "nfs-policy")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildFSExportType(), nil), Schema: s},
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

	// Verify export is gone.
	_, err := r.client.GetFileSystemExport(context.Background(), "delete-fs/delete-fs")
	if err == nil || !client.IsNotFound(err) {
		t.Errorf("expected export to be deleted, got: %v", err)
	}
}

// TestUnit_FileSystemExport_Import verifies ImportState populates all fields from combined name.
func TestUnit_FileSystemExport_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterFileSystemExportHandlers(ms.Mux)

	store.AddFileSystemExport("import-fs", "nfs-policy", "server1")

	r := newTestFSExportResource(t, ms)
	s := fsExportResourceSchema(t).Schema

	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildFSExportType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "import-fs/import-fs"}, importResp)

	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState returned error: %s", importResp.Diagnostics)
	}

	var model fileSystemExportModel
	if diags := importResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Name.ValueString() != "import-fs/import-fs" {
		t.Errorf("expected name=import-fs/import-fs, got %s", model.Name.ValueString())
	}
	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected ID to be populated after import")
	}
	if model.FileSystemName.ValueString() != "import-fs" {
		t.Errorf("expected file_system_name=import-fs, got %s", model.FileSystemName.ValueString())
	}
	if model.ServerName.ValueString() != "server1" {
		t.Errorf("expected server_name=server1, got %s", model.ServerName.ValueString())
	}
	if model.PolicyName.ValueString() != "nfs-policy" {
		t.Errorf("expected policy_name=nfs-policy, got %s", model.PolicyName.ValueString())
	}
}

// TestUnit_FileSystemExport_NotFound verifies Read removes resource from state when not found.
func TestUnit_FileSystemExport_NotFound(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterFileSystemExportHandlers(ms.Mux)

	r := newTestFSExportResource(t, ms)
	s := fsExportResourceSchema(t).Schema

	cfg := nullFSExportConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "does-not-exist/does-not-exist")
	cfg["id"] = tftypes.NewValue(tftypes.String, "non-existent-id")
	cfg["file_system_name"] = tftypes.NewValue(tftypes.String, "does-not-exist")
	state := tfsdk.State{Raw: tftypes.NewValue(buildFSExportType(), cfg), Schema: s}

	readResp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), resource.ReadRequest{State: state}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}
	if !readResp.State.Raw.IsNull() {
		t.Error("expected state to be removed (null) when resource not found")
	}
}

// TestUnit_FileSystemExport_Idempotent verifies that Read after Create shows no attribute drift.
func TestUnit_FileSystemExport_Idempotent(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterFileSystemExportHandlers(ms.Mux)

	r := newTestFSExportResource(t, ms)
	s := fsExportResourceSchema(t).Schema

	plan := fsExportPlanWith(t, "idempotent-fs", "server1", "nfs-policy")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildFSExportType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Read the state back — should not change anything.
	readResp := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}

	var beforeModel, afterModel fileSystemExportModel
	if diags := createResp.State.Get(context.Background(), &beforeModel); diags.HasError() {
		t.Fatalf("Get before state: %s", diags)
	}
	if diags := readResp.State.Get(context.Background(), &afterModel); diags.HasError() {
		t.Fatalf("Get after state: %s", diags)
	}

	if beforeModel.ID.ValueString() != afterModel.ID.ValueString() {
		t.Errorf("ID changed after Read: %s -> %s", beforeModel.ID.ValueString(), afterModel.ID.ValueString())
	}
	if beforeModel.Name.ValueString() != afterModel.Name.ValueString() {
		t.Errorf("Name changed after Read: %s -> %s", beforeModel.Name.ValueString(), afterModel.Name.ValueString())
	}
	if beforeModel.FileSystemName.ValueString() != afterModel.FileSystemName.ValueString() {
		t.Errorf("FileSystemName changed after Read: %s -> %s", beforeModel.FileSystemName.ValueString(), afterModel.FileSystemName.ValueString())
	}
	if beforeModel.Enabled.ValueBool() != afterModel.Enabled.ValueBool() {
		t.Errorf("Enabled changed after Read: %v -> %v", beforeModel.Enabled.ValueBool(), afterModel.Enabled.ValueBool())
	}
}
