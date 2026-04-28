package provider

import (
	"context"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	resschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/numberly/terraform-provider-mica/internal/client"
	"github.com/numberly/terraform-provider-mica/internal/testmock"
	"github.com/numberly/terraform-provider-mica/internal/testmock/handlers"
)

// ---- helpers ----------------------------------------------------------------

// newTestResource creates a filesystemResource wired to the given mock server.
func newTestResource(t *testing.T, ms *testmock.MockServer) *filesystemResource {
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

// buildFileSystemType returns the tftypes.Object for filesystem_resource schema.
func buildFileSystemType() tftypes.Object {
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
		Raw:    tftypes.NewValue(buildFileSystemType(), cfg),
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
		State: tfsdk.State{Raw: tftypes.NewValue(buildFileSystemType(), nil), Schema: s},
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

	plan := tfsdk.Plan{Raw: tftypes.NewValue(buildFileSystemType(), cfg), Schema: s}
	resp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildFileSystemType(), nil), Schema: s},
	}

	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Create with NFS returned error: %s", resp.Diagnostics)
	}

	var model filesystemModel
	if diags := resp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}
	if model.NFS.IsNull() || model.NFS.IsUnknown() {
		t.Fatal("expected NFS block to be populated")
	}
	var nfsModel filesystemNFSModel
	if diags := model.NFS.As(context.Background(), &nfsModel, basetypes.ObjectAsOptions{}); diags.HasError() {
		t.Fatalf("NFS.As: %s", diags)
	}
	if !nfsModel.Enabled.ValueBool() {
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

	plan := tfsdk.Plan{Raw: tftypes.NewValue(buildFileSystemType(), cfg), Schema: s}
	resp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildFileSystemType(), nil), Schema: s},
	}

	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Create with SMB returned error: %s", resp.Diagnostics)
	}

	var model filesystemModel
	if diags := resp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}
	if model.SMB.IsNull() || model.SMB.IsUnknown() {
		t.Fatal("expected SMB block to be populated")
	}
	var smbModel filesystemSMBModel
	if diags := model.SMB.As(context.Background(), &smbModel, basetypes.ObjectAsOptions{}); diags.HasError() {
		t.Fatalf("SMB.As: %s", diags)
	}
	if !smbModel.Enabled.ValueBool() {
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
		State: tfsdk.State{Raw: tftypes.NewValue(buildFileSystemType(), nil), Schema: s},
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
		State: tfsdk.State{Raw: tftypes.NewValue(buildFileSystemType(), nil), Schema: s},
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
	state := tfsdk.State{Raw: tftypes.NewValue(buildFileSystemType(), cfg), Schema: s}

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
		State: tfsdk.State{Raw: tftypes.NewValue(buildFileSystemType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Update provisioned size.
	newPlan := planWithName(t, "update-fs", 2147483648)
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildFileSystemType(), nil), Schema: s},
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
		State: tfsdk.State{Raw: tftypes.NewValue(buildFileSystemType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	newPlan := planWithName(t, "rename-after", 1073741824)
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildFileSystemType(), nil), Schema: s},
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
		State: tfsdk.State{Raw: tftypes.NewValue(buildFileSystemType(), nil), Schema: s},
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

	plan := tfsdk.Plan{Raw: tftypes.NewValue(buildFileSystemType(), cfg), Schema: s}
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildFileSystemType(), nil), Schema: s},
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
		State: tfsdk.State{Raw: tftypes.NewValue(buildFileSystemType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Import by name.
	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildFileSystemType(), nil), Schema: s},
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
		State: tfsdk.State{Raw: tftypes.NewValue(buildFileSystemType(), nil), Schema: s},
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
		State: tfsdk.State{Raw: tftypes.NewValue(buildFileSystemType(), nil), Schema: s},
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

// TestUnit_FileSystem_Lifecycle exercises the full Create->Read->Update->Read->Delete sequence.
func TestUnit_FileSystem_Lifecycle(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterFileSystemHandlers(ms.Mux)

	r := newTestResource(t, ms)
	s := resourceSchema(t).Schema

	// Step 1: Create.
	createPlan := planWithName(t, "lifecycle-fs", 1073741824)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildFileSystemType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}
	var createModel filesystemModel
	if diags := createResp.State.Get(context.Background(), &createModel); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}
	if createModel.Name.ValueString() != "lifecycle-fs" {
		t.Errorf("Create: expected name=lifecycle-fs, got %s", createModel.Name.ValueString())
	}
	if createModel.ID.IsNull() || createModel.ID.ValueString() == "" {
		t.Error("Create: expected non-empty ID")
	}

	// Step 2: Read post-create.
	readResp1 := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp1)
	if readResp1.Diagnostics.HasError() {
		t.Fatalf("Read post-create: %s", readResp1.Diagnostics)
	}
	var readModel1 filesystemModel
	if diags := readResp1.State.Get(context.Background(), &readModel1); diags.HasError() {
		t.Fatalf("Get read1 state: %s", diags)
	}
	if readModel1.Name.ValueString() != "lifecycle-fs" {
		t.Errorf("Read1: expected name=lifecycle-fs, got %s", readModel1.Name.ValueString())
	}

	// Step 3: Update provisioned to 2GiB.
	updatePlan := planWithName(t, "lifecycle-fs", 2147483648)
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildFileSystemType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlan,
		State: readResp1.State,
	}, updateResp)
	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update: %s", updateResp.Diagnostics)
	}
	var updateModel filesystemModel
	if diags := updateResp.State.Get(context.Background(), &updateModel); diags.HasError() {
		t.Fatalf("Get update state: %s", diags)
	}
	if updateModel.Provisioned.ValueInt64() != 2147483648 {
		t.Errorf("Update: expected provisioned=2GiB, got %d", updateModel.Provisioned.ValueInt64())
	}

	// Step 4: Read post-update.
	readResp2 := &resource.ReadResponse{State: updateResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: updateResp.State}, readResp2)
	if readResp2.Diagnostics.HasError() {
		t.Fatalf("Read post-update: %s", readResp2.Diagnostics)
	}
	var readModel2 filesystemModel
	if diags := readResp2.State.Get(context.Background(), &readModel2); diags.HasError() {
		t.Fatalf("Get read2 state: %s", diags)
	}
	if readModel2.Provisioned.ValueInt64() != 2147483648 {
		t.Errorf("Read2: expected provisioned=2GiB, got %d", readModel2.Provisioned.ValueInt64())
	}

	// Step 5: Delete (soft-delete + eradicate).
	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: readResp2.State}, deleteResp)
	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete: %s", deleteResp.Diagnostics)
	}
	_, err := r.client.GetFileSystem(context.Background(), "lifecycle-fs")
	if err == nil || !client.IsNotFound(err) {
		t.Errorf("expected file system to be deleted, got: %v", err)
	}
}

// TestUnit_FileSystem_ImportIdempotency verifies ImportState->Read produces state matching original Create.
func TestUnit_FileSystem_ImportIdempotency(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterFileSystemHandlers(ms.Mux)

	r := newTestResource(t, ms)
	s := resourceSchema(t).Schema

	// Create.
	createPlan := planWithName(t, "idempotent-fs", 1073741824)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildFileSystemType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}
	var createModel filesystemModel
	if diags := createResp.State.Get(context.Background(), &createModel); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}

	// ImportState.
	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildFileSystemType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "idempotent-fs"}, importResp)
	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState: %s", importResp.Diagnostics)
	}

	// Read to populate full state.
	readResp := &resource.ReadResponse{State: importResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: importResp.State}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read post-import: %s", readResp.Diagnostics)
	}
	var importedModel filesystemModel
	if diags := readResp.State.Get(context.Background(), &importedModel); diags.HasError() {
		t.Fatalf("Get imported state: %s", diags)
	}

	// Verify 0-diff: key user-configurable attributes match.
	if importedModel.Name.ValueString() != createModel.Name.ValueString() {
		t.Errorf("name mismatch: create=%s import=%s", createModel.Name.ValueString(), importedModel.Name.ValueString())
	}
	if importedModel.Provisioned.ValueInt64() != createModel.Provisioned.ValueInt64() {
		t.Errorf("provisioned mismatch: create=%d import=%d", createModel.Provisioned.ValueInt64(), importedModel.Provisioned.ValueInt64())
	}
	if importedModel.ID.ValueString() != createModel.ID.ValueString() {
		t.Errorf("id mismatch: create=%s import=%s", createModel.ID.ValueString(), importedModel.ID.ValueString())
	}
}

// TestUnit_FileSystem_Delete_Unprocessable verifies that a 422 Unprocessable returned
// by DELETE (eradication) produces an error diagnostic (not a silent failure or panic).
// This represents the case where a filesystem cannot be eradicated (e.g. still has mounts).
func TestUnit_FileSystem_Delete_Unprocessable(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()

	// Register a mock server that:
	// - PATCH: accepts soft-delete (returns destroyed=true)
	// - DELETE: returns 422 Unprocessable
	// - GET: returns the soft-deleted file system (needed by PATCH logic)
	ms.RegisterHandler("/api/2.22/file-systems", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// Return a soft-deleted filesystem (supports PATCH destroyed path).
			fs := client.FileSystem{
				ID:        "unprocessable-fs-id",
				Name:      "unprocessable-fs",
				Destroyed: true,
			}
			handlers.WriteJSONListResponse(w, http.StatusOK, []client.FileSystem{fs})
		case http.MethodPatch:
			// Accept soft-delete — return destroyed=true.
			destroyed := true
			fs := client.FileSystem{
				ID:        "unprocessable-fs-id",
				Name:      "unprocessable-fs",
				Destroyed: destroyed,
			}
			handlers.WriteJSONListResponse(w, http.StatusOK, []client.FileSystem{fs})
		case http.MethodDelete:
			// Simulate 422: filesystem cannot be eradicated (e.g. still has NFS mounts).
			handlers.WriteJSONError(w, http.StatusUnprocessableEntity, "File system cannot be eradicated while NFS mounts are active.")
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	r := newTestResource(t, ms)
	s := resourceSchema(t).Schema

	// Build state representing an existing filesystem with eradicate=true.
	cfg := nullFSConfig()
	cfg["id"] = tftypes.NewValue(tftypes.String, "unprocessable-fs-id")
	cfg["name"] = tftypes.NewValue(tftypes.String, "unprocessable-fs")
	cfg["provisioned"] = tftypes.NewValue(tftypes.Number, int64(1073741824))
	cfg["destroy_eradicate_on_delete"] = tftypes.NewValue(tftypes.Bool, true)
	state := tfsdk.State{Raw: tftypes.NewValue(buildFileSystemType(), cfg), Schema: s}

	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: state}, deleteResp)

	if !deleteResp.Diagnostics.HasError() {
		t.Error("expected Delete to produce an error diagnostic on 422 Unprocessable, got none")
	}
}

// TestUnit_FileSystem_PlanModifiers verifies all UseStateForUnknown plan modifiers
// in the filesystem resource schema.
func TestUnit_FileSystem_PlanModifiers(t *testing.T) {
	s := resourceSchema(t).Schema

	// id — UseStateForUnknown
	idAttr, ok := s.Attributes["id"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("id attribute not found or wrong type")
	}
	if len(idAttr.PlanModifiers) == 0 {
		t.Error("expected UseStateForUnknown plan modifier on id attribute")
	}

	// created — int64 UseStateForUnknown
	createdAttr, ok := s.Attributes["created"].(resschema.Int64Attribute)
	if !ok {
		t.Fatal("created attribute not found or wrong type")
	}
	if len(createdAttr.PlanModifiers) == 0 {
		t.Error("expected UseStateForUnknown plan modifier on created attribute")
	}

	// promotion_status — UseStateForUnknown
	psAttr, ok := s.Attributes["promotion_status"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("promotion_status attribute not found or wrong type")
	}
	if len(psAttr.PlanModifiers) == 0 {
		t.Error("expected UseStateForUnknown plan modifier on promotion_status attribute")
	}
}
