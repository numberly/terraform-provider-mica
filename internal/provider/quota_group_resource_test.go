package provider

import (
	"context"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/numberly/opentofu-provider-flashblade/internal/client"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock/handlers"
)

// ---- helpers ----------------------------------------------------------------

// newTestQuotaGroupResource creates a quotaGroupResource wired to the given mock server.
func newTestQuotaGroupResource(t *testing.T, ms *testmock.MockServer) *quotaGroupResource {
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
	return &quotaGroupResource{client: c}
}

// quotaGroupResourceSchema returns the parsed schema for the quota group resource.
func quotaGroupResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &quotaGroupResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

// buildQuotaGroupType returns the tftypes.Object for the quota group resource.
func buildQuotaGroupType() tftypes.Object {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":               tftypes.String,
		"file_system_name": tftypes.String,
		"gid":              tftypes.String,
		"quota":            tftypes.Number,
		"usage":            tftypes.Number,
		"timeouts":         timeoutsType,
	}}
}

// nullQuotaGroupConfig returns a base config map with all attributes null.
func nullQuotaGroupConfig() map[string]tftypes.Value {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, nil),
		"file_system_name": tftypes.NewValue(tftypes.String, nil),
		"gid":              tftypes.NewValue(tftypes.String, nil),
		"quota":            tftypes.NewValue(tftypes.Number, nil),
		"usage":            tftypes.NewValue(tftypes.Number, nil),
		"timeouts":         tftypes.NewValue(timeoutsType, nil),
	}
}

// quotaGroupPlan returns a tfsdk.Plan with the given fs name, gid, and quota.
func quotaGroupPlan(t *testing.T, fsName, gid string, quota int64) tfsdk.Plan {
	t.Helper()
	s := quotaGroupResourceSchema(t).Schema
	cfg := nullQuotaGroupConfig()
	cfg["file_system_name"] = tftypes.NewValue(tftypes.String, fsName)
	cfg["gid"] = tftypes.NewValue(tftypes.String, gid)
	cfg["quota"] = tftypes.NewValue(tftypes.Number, quota)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildQuotaGroupType(), cfg),
		Schema: s,
	}
}

// ---- tests ------------------------------------------------------------------

// TestQuotaGroupResource_Create verifies Create populates ID and quota.
func TestQuotaGroupResource_Create(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterQuotaHandlers(ms.Mux)

	r := newTestQuotaGroupResource(t, ms)
	s := quotaGroupResourceSchema(t).Schema

	plan := quotaGroupPlan(t, "testfs", "2000", 1073741824)
	resp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildQuotaGroupType(), nil), Schema: s},
	}

	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Create returned error: %s", resp.Diagnostics)
	}

	var model quotaGroupModel
	if diags := resp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.ValueString() != "testfs/2000" {
		t.Errorf("expected ID=testfs/2000, got %s", model.ID.ValueString())
	}
	if model.FileSystemName.ValueString() != "testfs" {
		t.Errorf("expected file_system_name=testfs, got %s", model.FileSystemName.ValueString())
	}
	if model.GID.ValueString() != "2000" {
		t.Errorf("expected gid=2000, got %s", model.GID.ValueString())
	}
	if model.Quota.ValueInt64() != 1073741824 {
		t.Errorf("expected quota=1073741824, got %d", model.Quota.ValueInt64())
	}
}

// TestQuotaGroupResource_Update verifies PATCH updates the quota limit.
func TestQuotaGroupResource_Update(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterQuotaHandlers(ms.Mux)

	r := newTestQuotaGroupResource(t, ms)
	s := quotaGroupResourceSchema(t).Schema

	// Create first.
	createPlan := quotaGroupPlan(t, "testfs", "2000", 1073741824)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildQuotaGroupType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Update quota to 2GB.
	newPlan := quotaGroupPlan(t, "testfs", "2000", 2147483648)
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildQuotaGroupType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  newPlan,
		State: createResp.State,
	}, updateResp)

	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update returned error: %s", updateResp.Diagnostics)
	}

	var model quotaGroupModel
	if diags := updateResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}
	if model.Quota.ValueInt64() != 2147483648 {
		t.Errorf("expected quota=2147483648 after update, got %d", model.Quota.ValueInt64())
	}
}

// TestQuotaGroupResource_Delete verifies Delete removes the quota.
func TestQuotaGroupResource_Delete(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterQuotaHandlers(ms.Mux)

	r := newTestQuotaGroupResource(t, ms)
	s := quotaGroupResourceSchema(t).Schema

	// Create first.
	plan := quotaGroupPlan(t, "testfs", "2000", 1073741824)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildQuotaGroupType(), nil), Schema: s},
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

	// Verify quota is gone.
	_, err := r.client.GetQuotaGroup(context.Background(), "testfs", "2000")
	if err == nil || !client.IsNotFound(err) {
		t.Errorf("expected quota to be deleted, got: %v", err)
	}
}

// TestQuotaGroupResource_Import verifies ImportState populates all attributes from composite ID.
func TestQuotaGroupResource_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterQuotaHandlers(ms.Mux)

	r := newTestQuotaGroupResource(t, ms)
	s := quotaGroupResourceSchema(t).Schema

	// Create first via client directly.
	c, err := client.NewClient(context.Background(), client.Config{
		Endpoint:           ms.URL(),
		APIToken:           "test-token",
		InsecureSkipVerify: true,
		MaxRetries:         1,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	_, err = c.PostQuotaGroup(context.Background(), "testfs", "2000", client.QuotaGroupPost{Quota: 1073741824})
	if err != nil {
		t.Fatalf("PostQuotaGroup: %v", err)
	}

	// Import by composite ID "testfs/2000".
	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildQuotaGroupType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "testfs/2000"}, importResp)

	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState returned error: %s", importResp.Diagnostics)
	}

	var model quotaGroupModel
	if diags := importResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.ValueString() != "testfs/2000" {
		t.Errorf("expected ID=testfs/2000 after import, got %s", model.ID.ValueString())
	}
	if model.FileSystemName.ValueString() != "testfs" {
		t.Errorf("expected file_system_name=testfs after import, got %s", model.FileSystemName.ValueString())
	}
	if model.GID.ValueString() != "2000" {
		t.Errorf("expected gid=2000 after import, got %s", model.GID.ValueString())
	}
	if model.Quota.ValueInt64() != 1073741824 {
		t.Errorf("expected quota=1073741824 after import, got %d", model.Quota.ValueInt64())
	}
}

// ---- data source tests -------------------------------------------------------

// newTestQuotaGroupDataSource creates a quotaGroupDataSource wired to the given mock server.
func newTestQuotaGroupDataSource(t *testing.T, ms *testmock.MockServer) *quotaGroupDataSource {
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
	return &quotaGroupDataSource{client: c}
}

// quotaGroupDataSourceSchema returns the schema for the quota group data source.
func quotaGroupDataSourceSchema(t *testing.T) datasource.SchemaResponse {
	t.Helper()
	d := &quotaGroupDataSource{}
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	return resp
}

// buildQuotaGroupDSType returns the tftypes.Object for the quota group data source.
func buildQuotaGroupDSType() tftypes.Object {
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":               tftypes.String,
		"file_system_name": tftypes.String,
		"gid":              tftypes.String,
		"quota":            tftypes.Number,
		"usage":            tftypes.Number,
	}}
}

// nullQuotaGroupDSConfig returns a base config map with all data source attributes null.
func nullQuotaGroupDSConfig() map[string]tftypes.Value {
	return map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, nil),
		"file_system_name": tftypes.NewValue(tftypes.String, nil),
		"gid":              tftypes.NewValue(tftypes.String, nil),
		"quota":            tftypes.NewValue(tftypes.Number, nil),
		"usage":            tftypes.NewValue(tftypes.Number, nil),
	}
}

// TestQuotaGroupDataSource verifies data source reads quota by file_system_name+gid and returns all attributes.
func TestQuotaGroupDataSource(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterQuotaHandlers(ms.Mux)

	// Create a quota via the client so the data source can find it.
	c, err := client.NewClient(context.Background(), client.Config{
		Endpoint:           ms.URL(),
		APIToken:           "test-token",
		InsecureSkipVerify: true,
		MaxRetries:         1,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	_, err = c.PostQuotaGroup(context.Background(), "testfs", "2000", client.QuotaGroupPost{Quota: 1073741824})
	if err != nil {
		t.Fatalf("PostQuotaGroup: %v", err)
	}

	d := newTestQuotaGroupDataSource(t, ms)
	s := quotaGroupDataSourceSchema(t).Schema

	cfg := nullQuotaGroupDSConfig()
	cfg["file_system_name"] = tftypes.NewValue(tftypes.String, "testfs")
	cfg["gid"] = tftypes.NewValue(tftypes.String, "2000")

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildQuotaGroupDSType(), nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(buildQuotaGroupDSType(), cfg), Schema: s},
	}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("DataSource Read returned error: %s", readResp.Diagnostics)
	}

	var model quotaGroupDataSourceModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.ValueString() != "testfs/2000" {
		t.Errorf("expected ID=testfs/2000, got %s", model.ID.ValueString())
	}
	if model.FileSystemName.ValueString() != "testfs" {
		t.Errorf("expected file_system_name=testfs, got %s", model.FileSystemName.ValueString())
	}
	if model.GID.ValueString() != "2000" {
		t.Errorf("expected gid=2000, got %s", model.GID.ValueString())
	}
	if model.Quota.ValueInt64() != 1073741824 {
		t.Errorf("expected quota=1073741824, got %d", model.Quota.ValueInt64())
	}
}

// TestUnit_QuotaGroup_Create_Conflict verifies that a 409 Conflict on POST produces
// an error diagnostic.
func TestUnit_QuotaGroup_Create_Conflict(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	ms.RegisterHandler("/api/2.22/quotas/groups", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handlers.WriteJSONError(w, http.StatusConflict, "A quota for this group already exists on the file system.")
			return
		}
		handlers.WriteJSONListResponse(w, http.StatusOK, []client.QuotaGroup{})
	})

	r := newTestQuotaGroupResource(t, ms)
	s := quotaGroupResourceSchema(t).Schema

	plan := quotaGroupPlan(t, "testfs", "2000", 1073741824)
	resp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildQuotaGroupType(), nil), Schema: s},
	}

	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("expected Create to produce an error diagnostic on 409 Conflict, got none")
	}
}

// TestUnit_QuotaGroup_Read_NotFound verifies that a not-found response during Read
// removes the resource from Terraform state without an error diagnostic.
func TestUnit_QuotaGroup_Read_NotFound(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	ms.RegisterHandler("/api/2.22/quotas/groups", func(w http.ResponseWriter, r *http.Request) {
		handlers.WriteJSONListResponse(w, http.StatusOK, []client.QuotaGroup{})
	})

	r := newTestQuotaGroupResource(t, ms)
	s := quotaGroupResourceSchema(t).Schema

	cfg := nullQuotaGroupConfig()
	cfg["id"] = tftypes.NewValue(tftypes.String, "testfs/2000")
	cfg["file_system_name"] = tftypes.NewValue(tftypes.String, "testfs")
	cfg["gid"] = tftypes.NewValue(tftypes.String, "2000")
	cfg["quota"] = tftypes.NewValue(tftypes.Number, int64(1073741824))
	state := tfsdk.State{Raw: tftypes.NewValue(buildQuotaGroupType(), cfg), Schema: s}

	readResp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), resource.ReadRequest{State: state}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}
	if !readResp.State.Raw.IsNull() {
		t.Error("expected state to be removed (null) when quota group not found")
	}
}

// TestUnit_QuotaGroup_Lifecycle exercises the full Create->Read->Update->Read->Delete sequence.
func TestUnit_QuotaGroup_Lifecycle(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterQuotaHandlers(ms.Mux)

	r := newTestQuotaGroupResource(t, ms)
	s := quotaGroupResourceSchema(t).Schema

	// Step 1: Create.
	createPlan := quotaGroupPlan(t, "lifecycle-fs", "3000", 1073741824)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildQuotaGroupType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}
	var createModel quotaGroupModel
	if diags := createResp.State.Get(context.Background(), &createModel); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}
	if createModel.Quota.ValueInt64() != 1073741824 {
		t.Errorf("Create: expected quota=1GiB, got %d", createModel.Quota.ValueInt64())
	}

	// Step 2: Read post-create.
	readResp1 := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp1)
	if readResp1.Diagnostics.HasError() {
		t.Fatalf("Read post-create: %s", readResp1.Diagnostics)
	}
	var readModel1 quotaGroupModel
	if diags := readResp1.State.Get(context.Background(), &readModel1); diags.HasError() {
		t.Fatalf("Get read1 state: %s", diags)
	}
	if readModel1.GID.ValueString() != "3000" {
		t.Errorf("Read1: expected gid=3000, got %s", readModel1.GID.ValueString())
	}

	// Step 3: Update quota to 2GiB.
	updatePlan := quotaGroupPlan(t, "lifecycle-fs", "3000", 2147483648)
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildQuotaGroupType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlan,
		State: readResp1.State,
	}, updateResp)
	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update: %s", updateResp.Diagnostics)
	}
	var updateModel quotaGroupModel
	if diags := updateResp.State.Get(context.Background(), &updateModel); diags.HasError() {
		t.Fatalf("Get update state: %s", diags)
	}
	if updateModel.Quota.ValueInt64() != 2147483648 {
		t.Errorf("Update: expected quota=2GiB, got %d", updateModel.Quota.ValueInt64())
	}

	// Step 4: Read post-update.
	readResp2 := &resource.ReadResponse{State: updateResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: updateResp.State}, readResp2)
	if readResp2.Diagnostics.HasError() {
		t.Fatalf("Read post-update: %s", readResp2.Diagnostics)
	}
	var readModel2 quotaGroupModel
	if diags := readResp2.State.Get(context.Background(), &readModel2); diags.HasError() {
		t.Fatalf("Get read2 state: %s", diags)
	}
	if readModel2.Quota.ValueInt64() != 2147483648 {
		t.Errorf("Read2: expected quota=2GiB, got %d", readModel2.Quota.ValueInt64())
	}

	// Step 5: Delete.
	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: readResp2.State}, deleteResp)
	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete: %s", deleteResp.Diagnostics)
	}
	_, err := r.client.GetQuotaGroup(context.Background(), "lifecycle-fs", "3000")
	if err == nil || !client.IsNotFound(err) {
		t.Errorf("expected quota to be deleted, got: %v", err)
	}
}

// TestUnit_QuotaGroup_ImportIdempotency verifies ImportState->Read produces state matching original Create.
func TestUnit_QuotaGroup_ImportIdempotency(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterQuotaHandlers(ms.Mux)

	r := newTestQuotaGroupResource(t, ms)
	s := quotaGroupResourceSchema(t).Schema

	// Create.
	createPlan := quotaGroupPlan(t, "idempotent-fs", "5000", 1073741824)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildQuotaGroupType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}
	var createModel quotaGroupModel
	if diags := createResp.State.Get(context.Background(), &createModel); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}

	// ImportState using composite ID "fs_name/gid".
	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildQuotaGroupType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "idempotent-fs/5000"}, importResp)
	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState: %s", importResp.Diagnostics)
	}

	// Read to populate full state.
	readResp := &resource.ReadResponse{State: importResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: importResp.State}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read post-import: %s", readResp.Diagnostics)
	}
	var importedModel quotaGroupModel
	if diags := readResp.State.Get(context.Background(), &importedModel); diags.HasError() {
		t.Fatalf("Get imported state: %s", diags)
	}

	// Verify 0-diff.
	if importedModel.FileSystemName.ValueString() != createModel.FileSystemName.ValueString() {
		t.Errorf("file_system_name mismatch: create=%s import=%s", createModel.FileSystemName.ValueString(), importedModel.FileSystemName.ValueString())
	}
	if importedModel.GID.ValueString() != createModel.GID.ValueString() {
		t.Errorf("gid mismatch: create=%s import=%s", createModel.GID.ValueString(), importedModel.GID.ValueString())
	}
	if importedModel.Quota.ValueInt64() != createModel.Quota.ValueInt64() {
		t.Errorf("quota mismatch: create=%d import=%d", createModel.Quota.ValueInt64(), importedModel.Quota.ValueInt64())
	}
}

// TestUnit_QuotaGroup_PlanModifiers verifies all RequiresReplace and UseStateForUnknown
// plan modifiers in the quota_group resource schema.
func TestUnit_QuotaGroup_PlanModifiers(t *testing.T) {
	s := quotaGroupResourceSchema(t).Schema

	// id — UseStateForUnknown
	idAttr, ok := s.Attributes["id"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("id attribute not found or wrong type")
	}
	if len(idAttr.PlanModifiers) == 0 {
		t.Error("expected UseStateForUnknown plan modifier on id attribute")
	}

	// file_system_name — RequiresReplace
	fsnAttr, ok := s.Attributes["file_system_name"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("file_system_name attribute not found or wrong type")
	}
	if len(fsnAttr.PlanModifiers) == 0 {
		t.Error("expected RequiresReplace plan modifier on file_system_name attribute")
	}

	// gid — RequiresReplace
	gidAttr, ok := s.Attributes["gid"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("gid attribute not found or wrong type")
	}
	if len(gidAttr.PlanModifiers) == 0 {
		t.Error("expected RequiresReplace plan modifier on gid attribute")
	}

	// usage — UseStateForUnknown
	usageAttr, ok := s.Attributes["usage"].(resschema.Int64Attribute)
	if !ok {
		t.Fatal("usage attribute not found or wrong type")
	}
	if len(usageAttr.PlanModifiers) == 0 {
		t.Error("expected UseStateForUnknown plan modifier on usage attribute")
	}
}

// TestUnit_QuotaGroup_QuotaValidator verifies the quota field rejects negative values
// and accepts 0 and positive values.
func TestUnit_QuotaGroup_QuotaValidator(t *testing.T) {
	s := quotaGroupResourceSchema(t).Schema

	qAttr, ok := s.Attributes["quota"].(resschema.Int64Attribute)
	if !ok {
		t.Fatal("quota attribute not found or wrong type")
	}
	if len(qAttr.Validators) == 0 {
		t.Fatal("expected at least one validator on quota attribute")
	}

	v := qAttr.Validators[0]

	// -1 should produce an error.
	resp := &validator.Int64Response{}
	v.ValidateInt64(context.Background(), validator.Int64Request{
		ConfigValue: types.Int64Value(-1),
	}, resp)
	if !resp.Diagnostics.HasError() {
		t.Error("expected validator to reject -1 quota value")
	}

	// 0 should be valid.
	resp2 := &validator.Int64Response{}
	v.ValidateInt64(context.Background(), validator.Int64Request{
		ConfigValue: types.Int64Value(0),
	}, resp2)
	if resp2.Diagnostics.HasError() {
		t.Errorf("expected validator to accept 0 quota value, got error: %s", resp2.Diagnostics)
	}

	// 1048576 should be valid.
	resp3 := &validator.Int64Response{}
	v.ValidateInt64(context.Background(), validator.Int64Request{
		ConfigValue: types.Int64Value(1048576),
	}, resp3)
	if resp3.Diagnostics.HasError() {
		t.Errorf("expected validator to accept 1048576 quota value, got error: %s", resp3.Diagnostics)
	}
}
