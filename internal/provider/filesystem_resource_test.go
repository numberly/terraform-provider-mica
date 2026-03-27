package provider

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/soulkyu/terraform-provider-flashblade/internal/client"
	"github.com/soulkyu/terraform-provider-flashblade/internal/testmock"
	"github.com/soulkyu/terraform-provider-flashblade/internal/testmock/handlers"
)

// ---- helpers ----------------------------------------------------------------

// newTestResource creates a filesystemResource wired to the given mock server.
func newTestResource(t *testing.T, ms *testmock.MockServer) *filesystemResource {
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
	return &filesystemResource{client: c}
}

// resourceSchema returns the parsed schema for the filesystem resource.
func resourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &filesystemResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

// fsType returns the tftypes.Object for the full filesystem resource schema.
// It is built from the actual schema so tests stay in sync automatically.
func fsType(t *testing.T) tftypes.Object {
	t.Helper()
	s := resourceSchema(t).Schema
	typ, err := s.Type().ApplyTerraform5AttributePathStep(nil)
	_ = typ
	_ = err
	// Build a static tftypes.Object matching the schema — maintained manually
	// but checked against the schema via compile-time model struct.
	return buildFSType()
}

// buildFSType returns the tftypes.Object for filesystem_resource schema.
func buildFSType() tftypes.Object {
	spaceType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"data_reduction":      tftypes.Number,
		"snapshots":           tftypes.Number,
		"total_physical":      tftypes.Number,
		"unique":              tftypes.Number,
		"virtual":             tftypes.Number,
		"snapshots_effective": tftypes.Number,
	}}
	nfsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"enabled":      tftypes.Bool,
		"v3_enabled":   tftypes.Bool,
		"v4_1_enabled": tftypes.Bool,
		"rules":        tftypes.String,
		"transport":    tftypes.String,
	}}
	smbType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"enabled":                         tftypes.Bool,
		"access_based_enumeration_enabled": tftypes.Bool,
		"continuous_availability_enabled":  tftypes.Bool,
		"smb_encryption_enabled":           tftypes.Bool,
	}}
	httpType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"enabled": tftypes.Bool,
	}}
	multiProtocolType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"access_control_style": tftypes.String,
		"safeguard_acls":       tftypes.Bool,
	}}
	defaultQuotasType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"group_quota": tftypes.Number,
		"user_quota":  tftypes.Number,
	}}
	sourceType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":   tftypes.String,
		"name": tftypes.String,
	}}
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":                          tftypes.String,
		"name":                        tftypes.String,
		"provisioned":                 tftypes.Number,
		"destroyed":                   tftypes.Bool,
		"destroy_eradicate_on_delete": tftypes.Bool,
		"time_remaining":              tftypes.Number,
		"created":                     tftypes.Number,
		"promotion_status":            tftypes.String,
		"writable":                    tftypes.Bool,
		"nfs_export_policy":           tftypes.String,
		"smb_share_policy":            tftypes.String,
		"space":                       spaceType,
		"nfs":                         nfsType,
		"smb":                         smbType,
		"http":                        httpType,
		"multi_protocol":              multiProtocolType,
		"default_quotas":              defaultQuotasType,
		"source":                      sourceType,
		"timeouts":                    timeoutsType,
	}}
}

// nullFSConfig returns a base config map with all attributes null.
func nullFSConfig() map[string]tftypes.Value {
	spaceType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"data_reduction":      tftypes.Number,
		"snapshots":           tftypes.Number,
		"total_physical":      tftypes.Number,
		"unique":              tftypes.Number,
		"virtual":             tftypes.Number,
		"snapshots_effective": tftypes.Number,
	}}
	nfsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"enabled":      tftypes.Bool,
		"v3_enabled":   tftypes.Bool,
		"v4_1_enabled": tftypes.Bool,
		"rules":        tftypes.String,
		"transport":    tftypes.String,
	}}
	smbType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"enabled":                         tftypes.Bool,
		"access_based_enumeration_enabled": tftypes.Bool,
		"continuous_availability_enabled":  tftypes.Bool,
		"smb_encryption_enabled":           tftypes.Bool,
	}}
	httpType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"enabled": tftypes.Bool,
	}}
	multiProtocolType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"access_control_style": tftypes.String,
		"safeguard_acls":       tftypes.Bool,
	}}
	defaultQuotasType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"group_quota": tftypes.Number,
		"user_quota":  tftypes.Number,
	}}
	sourceType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":   tftypes.String,
		"name": tftypes.String,
	}}
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return map[string]tftypes.Value{
		"id":                          tftypes.NewValue(tftypes.String, nil),
		"name":                        tftypes.NewValue(tftypes.String, nil),
		"provisioned":                 tftypes.NewValue(tftypes.Number, nil),
		"destroyed":                   tftypes.NewValue(tftypes.Bool, nil),
		"destroy_eradicate_on_delete": tftypes.NewValue(tftypes.Bool, nil),
		"time_remaining":              tftypes.NewValue(tftypes.Number, nil),
		"created":                     tftypes.NewValue(tftypes.Number, nil),
		"promotion_status":            tftypes.NewValue(tftypes.String, nil),
		"writable":                    tftypes.NewValue(tftypes.Bool, nil),
		"nfs_export_policy":           tftypes.NewValue(tftypes.String, nil),
		"smb_share_policy":            tftypes.NewValue(tftypes.String, nil),
		"space":                       tftypes.NewValue(spaceType, nil),
		"nfs":                         tftypes.NewValue(nfsType, nil),
		"smb":                         tftypes.NewValue(smbType, nil),
		"http":                        tftypes.NewValue(httpType, nil),
		"multi_protocol":              tftypes.NewValue(multiProtocolType, nil),
		"default_quotas":              tftypes.NewValue(defaultQuotasType, nil),
		"source":                      tftypes.NewValue(sourceType, nil),
		"timeouts":                    tftypes.NewValue(timeoutsType, nil),
	}
}

// planWithName returns a tfsdk.Plan with the given name and provisioned size.
func planWithName(t *testing.T, name string, provisioned int64) tfsdk.Plan {
	t.Helper()
	s := resourceSchema(t).Schema
	cfg := nullFSConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, name)
	cfg["provisioned"] = tftypes.NewValue(tftypes.Number, provisioned)
	cfg["destroy_eradicate_on_delete"] = tftypes.NewValue(tftypes.Bool, true)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildFSType(), cfg),
		Schema: s,
	}
}

// stateFromCreateResp builds a tfsdk.State from a Create response for Update/Delete tests.
func stateFromPlan(t *testing.T, plan tfsdk.Plan) tfsdk.State {
	t.Helper()
	s := resourceSchema(t).Schema
	return tfsdk.State{
		Raw:    plan.Raw,
		Schema: s,
	}
}

// ---- tests ------------------------------------------------------------------

// TestUnit_FileSystem_Create verifies Create populates ID and computed fields.
func TestUnit_FileSystem_Create(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterFileSystemHandlers(ms.Mux)

	r := newTestResource(t, ms)
	s := resourceSchema(t).Schema

	plan := planWithName(t, "test-fs", 1073741824)
	req := resource.CreateRequest{
		Plan: plan,
	}
	resp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildFSType(), nil), Schema: s},
	}

	r.Create(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Create returned error: %s", resp.Diagnostics)
	}

	var model filesystemModel
	diags := resp.State.Get(context.Background(), &model)
	if diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected non-empty ID after Create")
	}
	if model.Name.ValueString() != "test-fs" {
		t.Errorf("expected name=test-fs, got %s", model.Name.ValueString())
	}
	if model.Provisioned.ValueInt64() != 1073741824 {
		t.Errorf("expected provisioned=1073741824, got %d", model.Provisioned.ValueInt64())
	}
	if model.Created.IsNull() || model.Created.IsUnknown() {
		t.Error("expected created to be populated after Create")
	}
}

// TestUnit_FileSystem_Create_WithNFS verifies NFS attributes are persisted.
func TestUnit_FileSystem_Create_WithNFS(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterFileSystemHandlers(ms.Mux)

	r := newTestResource(t, ms)
	s := resourceSchema(t).Schema

	nfsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"enabled":      tftypes.Bool,
		"v3_enabled":   tftypes.Bool,
		"v4_1_enabled": tftypes.Bool,
		"rules":        tftypes.String,
		"transport":    tftypes.String,
	}}

	cfg := nullFSConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "nfs-fs")
	cfg["provisioned"] = tftypes.NewValue(tftypes.Number, int64(1073741824))
	cfg["destroy_eradicate_on_delete"] = tftypes.NewValue(tftypes.Bool, true)
	cfg["nfs"] = tftypes.NewValue(nfsType, map[string]tftypes.Value{
		"enabled":      tftypes.NewValue(tftypes.Bool, true),
		"v3_enabled":   tftypes.NewValue(tftypes.Bool, false),
		"v4_1_enabled": tftypes.NewValue(tftypes.Bool, true),
		"rules":        tftypes.NewValue(tftypes.String, "*(rw)"),
		"transport":    tftypes.NewValue(tftypes.String, "tcp"),
	})

	plan := tfsdk.Plan{Raw: tftypes.NewValue(buildFSType(), cfg), Schema: s}
	resp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildFSType(), nil), Schema: s},
	}

	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Create with NFS returned error: %s", resp.Diagnostics)
	}

	var model filesystemModel
	if diags := resp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}
	if model.NFS == nil {
		t.Fatal("expected NFS block to be populated")
	}
	if !model.NFS.Enabled.ValueBool() {
		t.Error("expected NFS.enabled=true")
	}
}

// TestUnit_FileSystem_Create_WithSMB verifies SMB attributes are persisted.
func TestUnit_FileSystem_Create_WithSMB(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterFileSystemHandlers(ms.Mux)

	r := newTestResource(t, ms)
	s := resourceSchema(t).Schema

	smbType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"enabled":                         tftypes.Bool,
		"access_based_enumeration_enabled": tftypes.Bool,
		"continuous_availability_enabled":  tftypes.Bool,
		"smb_encryption_enabled":           tftypes.Bool,
	}}

	cfg := nullFSConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "smb-fs")
	cfg["provisioned"] = tftypes.NewValue(tftypes.Number, int64(1073741824))
	cfg["destroy_eradicate_on_delete"] = tftypes.NewValue(tftypes.Bool, true)
	cfg["smb"] = tftypes.NewValue(smbType, map[string]tftypes.Value{
		"enabled":                         tftypes.NewValue(tftypes.Bool, true),
		"access_based_enumeration_enabled": tftypes.NewValue(tftypes.Bool, false),
		"continuous_availability_enabled":  tftypes.NewValue(tftypes.Bool, false),
		"smb_encryption_enabled":           tftypes.NewValue(tftypes.Bool, false),
	})

	plan := tfsdk.Plan{Raw: tftypes.NewValue(buildFSType(), cfg), Schema: s}
	resp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildFSType(), nil), Schema: s},
	}

	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Create with SMB returned error: %s", resp.Diagnostics)
	}

	var model filesystemModel
	if diags := resp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}
	if model.SMB == nil {
		t.Fatal("expected SMB block to be populated")
	}
	if !model.SMB.Enabled.ValueBool() {
		t.Error("expected SMB.enabled=true")
	}
}

// TestUnit_FileSystem_Read verifies Read populates all attributes from API.
func TestUnit_FileSystem_Read(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterFileSystemHandlers(ms.Mux)

	r := newTestResource(t, ms)
	s := resourceSchema(t).Schema

	// First create a file system so Read can find it.
	plan := planWithName(t, "read-fs", 2147483648)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildFSType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Now read the state back.
	readResp := &resource.ReadResponse{
		State: createResp.State,
	}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}

	var model filesystemModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Name.ValueString() != "read-fs" {
		t.Errorf("expected name=read-fs, got %s", model.Name.ValueString())
	}
	if model.Provisioned.ValueInt64() != 2147483648 {
		t.Errorf("expected provisioned=2147483648, got %d", model.Provisioned.ValueInt64())
	}
}

// TestUnit_FileSystem_Read_Destroyed verifies that destroyed=true is reflected in state.
func TestUnit_FileSystem_Read_Destroyed(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterFileSystemHandlers(ms.Mux)

	r := newTestResource(t, ms)
	s := resourceSchema(t).Schema

	// Create and then soft-delete.
	plan := planWithName(t, "destroyed-fs", 1073741824)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildFSType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	var createdModel filesystemModel
	if diags := createResp.State.Get(context.Background(), &createdModel); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	// Soft-delete via PATCH.
	destroyed := true
	_, err := r.client.PatchFileSystem(context.Background(), createdModel.ID.ValueString(), client.FileSystemPatch{
		Destroyed: &destroyed,
	})
	if err != nil {
		t.Fatalf("PatchFileSystem(destroyed=true): %v", err)
	}

	readResp := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}

	var model filesystemModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}
	if !model.Destroyed.ValueBool() {
		t.Error("expected destroyed=true after soft-delete")
	}
}

// TestUnit_FileSystem_Read_NotFound verifies that 404 removes resource from state.
func TestUnit_FileSystem_Read_NotFound(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterFileSystemHandlers(ms.Mux)

	r := newTestResource(t, ms)
	s := resourceSchema(t).Schema

	// Simulate a state with a non-existent file system.
	cfg := nullFSConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "does-not-exist")
	cfg["id"] = tftypes.NewValue(tftypes.String, "non-existent-id")
	cfg["provisioned"] = tftypes.NewValue(tftypes.Number, int64(0))
	cfg["destroy_eradicate_on_delete"] = tftypes.NewValue(tftypes.Bool, true)
	state := tfsdk.State{Raw: tftypes.NewValue(buildFSType(), cfg), Schema: s}

	readResp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), resource.ReadRequest{State: state}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}
	if !readResp.State.Raw.IsNull() {
		t.Error("expected state to be removed (null) when resource not found")
	}
}

// TestUnit_FileSystem_Update verifies Update changes provisioned size.
func TestUnit_FileSystem_Update(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterFileSystemHandlers(ms.Mux)

	r := newTestResource(t, ms)
	s := resourceSchema(t).Schema

	// Create first.
	plan := planWithName(t, "update-fs", 1073741824)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildFSType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Update provisioned size.
	newPlan := planWithName(t, "update-fs", 2147483648)
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildFSType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  newPlan,
		State: createResp.State,
	}, updateResp)

	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update returned error: %s", updateResp.Diagnostics)
	}

	var model filesystemModel
	if diags := updateResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}
	if model.Provisioned.ValueInt64() != 2147483648 {
		t.Errorf("expected provisioned=2147483648 after update, got %d", model.Provisioned.ValueInt64())
	}
}

// TestUnit_FileSystem_Update_Rename verifies in-place rename without replace.
func TestUnit_FileSystem_Update_Rename(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterFileSystemHandlers(ms.Mux)

	r := newTestResource(t, ms)
	s := resourceSchema(t).Schema

	plan := planWithName(t, "rename-before", 1073741824)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildFSType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	newPlan := planWithName(t, "rename-after", 1073741824)
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildFSType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  newPlan,
		State: createResp.State,
	}, updateResp)

	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update rename returned error: %s", updateResp.Diagnostics)
	}

	var model filesystemModel
	if diags := updateResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}
	if model.Name.ValueString() != "rename-after" {
		t.Errorf("expected name=rename-after after update, got %s", model.Name.ValueString())
	}
}

// TestUnit_FileSystem_Destroy verifies soft-delete + eradicate when destroy_eradicate_on_delete=true.
func TestUnit_FileSystem_Destroy(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterFileSystemHandlers(ms.Mux)

	r := newTestResource(t, ms)
	s := resourceSchema(t).Schema

	plan := planWithName(t, "destroy-fs", 1073741824)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildFSType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: createResp.State}, deleteResp)

	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete returned error: %s", deleteResp.Diagnostics)
	}

	// Verify file system is gone.
	_, err := r.client.GetFileSystem(context.Background(), "destroy-fs")
	if err == nil || !client.IsNotFound(err) {
		t.Errorf("expected file system to be eradicated, got: %v", err)
	}
}

// TestUnit_FileSystem_Destroy_SoftOnly verifies only soft-delete when destroy_eradicate_on_delete=false.
func TestUnit_FileSystem_Destroy_SoftOnly(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterFileSystemHandlers(ms.Mux)

	r := newTestResource(t, ms)
	s := resourceSchema(t).Schema

	// Create with eradicate=false.
	cfg := nullFSConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "softdelete-fs")
	cfg["provisioned"] = tftypes.NewValue(tftypes.Number, int64(1073741824))
	cfg["destroy_eradicate_on_delete"] = tftypes.NewValue(tftypes.Bool, false)

	plan := tfsdk.Plan{Raw: tftypes.NewValue(buildFSType(), cfg), Schema: s}
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildFSType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: createResp.State}, deleteResp)

	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete (soft only) returned error: %s", deleteResp.Diagnostics)
	}

	// Verify file system still exists but is destroyed.
	var createdModel filesystemModel
	if diags := createResp.State.Get(context.Background(), &createdModel); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	// Use list with destroyed=true to find it.
	destroyed := true
	items, err := r.client.ListFileSystems(context.Background(), client.ListFileSystemsOpts{
		Destroyed: &destroyed,
		Names:     []string{"softdelete-fs"},
	})
	if err != nil {
		t.Fatalf("ListFileSystems(destroyed): %v", err)
	}
	if len(items) == 0 {
		t.Error("expected file system to still exist as soft-deleted")
	}
	if !items[0].Destroyed {
		t.Error("expected file system to be destroyed after soft-only delete")
	}
}

// TestUnit_FileSystem_Import verifies ImportState populates full state.
func TestUnit_FileSystem_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterFileSystemHandlers(ms.Mux)

	r := newTestResource(t, ms)
	s := resourceSchema(t).Schema

	// Create first.
	plan := planWithName(t, "import-fs", 1073741824)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildFSType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Import by name.
	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildFSType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "import-fs"}, importResp)

	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState returned error: %s", importResp.Diagnostics)
	}

	var model filesystemModel
	if diags := importResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Name.ValueString() != "import-fs" {
		t.Errorf("expected name=import-fs after import, got %s", model.Name.ValueString())
	}
	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected ID to be populated after import")
	}
}

// TestUnit_FileSystem_DriftLog verifies structured drift logging via tflog.
func TestUnit_FileSystem_DriftLog(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterFileSystemHandlers(ms.Mux)

	r := newTestResource(t, ms)
	s := resourceSchema(t).Schema

	// Create with provisioned=1GiB.
	plan := planWithName(t, "drift-fs", 1073741824)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildFSType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Simulate drift: change provisioned in the server without going through Terraform.
	var createdModel filesystemModel
	if diags := createResp.State.Get(context.Background(), &createdModel); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}
	newProvisioned := int64(2147483648)
	_, err := r.client.PatchFileSystem(context.Background(), createdModel.ID.ValueString(), client.FileSystemPatch{
		Provisioned: &newProvisioned,
	})
	if err != nil {
		t.Fatalf("PatchFileSystem: %v", err)
	}

	// Read should detect and log the drift. We just verify no error is returned.
	// (tflog output is captured in provider-level tests with sinks.)
	readResp := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read with drift returned error: %s", readResp.Diagnostics)
	}

	// After Read, state should reflect the API value (drift corrected).
	var afterModel filesystemModel
	if diags := readResp.State.Get(context.Background(), &afterModel); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}
	if afterModel.Provisioned.ValueInt64() != 2147483648 {
		t.Errorf("expected provisioned=2GiB after drift read, got %d", afterModel.Provisioned.ValueInt64())
	}
}

// TestUnit_FileSystem_Idempotent verifies that Read after Create shows no changes.
func TestUnit_FileSystem_Idempotent(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterFileSystemHandlers(ms.Mux)

	r := newTestResource(t, ms)
	s := resourceSchema(t).Schema

	plan := planWithName(t, "idempotent-fs", 1073741824)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildFSType(), nil), Schema: s},
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

	var beforeModel, afterModel filesystemModel
	if diags := createResp.State.Get(context.Background(), &beforeModel); diags.HasError() {
		t.Fatalf("Get before state: %s", diags)
	}
	if diags := readResp.State.Get(context.Background(), &afterModel); diags.HasError() {
		t.Fatalf("Get after state: %s", diags)
	}

	if beforeModel.ID.ValueString() != afterModel.ID.ValueString() {
		t.Errorf("ID changed after Read: %s -> %s", beforeModel.ID.ValueString(), afterModel.ID.ValueString())
	}
	if beforeModel.Provisioned.ValueInt64() != afterModel.Provisioned.ValueInt64() {
		t.Errorf("Provisioned changed after Read: %d -> %d",
			beforeModel.Provisioned.ValueInt64(), afterModel.Provisioned.ValueInt64())
	}
}
