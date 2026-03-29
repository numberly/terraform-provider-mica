package provider

import (
	"context"
	"testing"
	"time"

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

// newTestQuotaUserResource creates a quotaUserResource wired to the given mock server.
func newTestQuotaUserResource(t *testing.T, ms *testmock.MockServer) *quotaUserResource {
	t.Helper()
	c, err := client.NewClient(context.Background(), client.Config{
		Endpoint:           ms.URL(),
		APIToken:           "test-token",
		InsecureSkipVerify: true,
		MaxRetries:         1,
		RetryBaseDelay:     1 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return &quotaUserResource{client: c}
}

// quotaUserResourceSchema returns the parsed schema for the quota user resource.
func quotaUserResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &quotaUserResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

// buildQuotaUserType returns the tftypes.Object for the quota user resource.
func buildQuotaUserType() tftypes.Object {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":               tftypes.String,
		"file_system_name": tftypes.String,
		"uid":              tftypes.String,
		"quota":            tftypes.Number,
		"usage":            tftypes.Number,
		"timeouts":         timeoutsType,
	}}
}

// nullQuotaUserConfig returns a base config map with all attributes null.
func nullQuotaUserConfig() map[string]tftypes.Value {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, nil),
		"file_system_name": tftypes.NewValue(tftypes.String, nil),
		"uid":              tftypes.NewValue(tftypes.String, nil),
		"quota":            tftypes.NewValue(tftypes.Number, nil),
		"usage":            tftypes.NewValue(tftypes.Number, nil),
		"timeouts":         tftypes.NewValue(timeoutsType, nil),
	}
}

// quotaUserPlan returns a tfsdk.Plan with the given fs name, uid, and quota.
func quotaUserPlan(t *testing.T, fsName, uid string, quota int64) tfsdk.Plan {
	t.Helper()
	s := quotaUserResourceSchema(t).Schema
	cfg := nullQuotaUserConfig()
	cfg["file_system_name"] = tftypes.NewValue(tftypes.String, fsName)
	cfg["uid"] = tftypes.NewValue(tftypes.String, uid)
	cfg["quota"] = tftypes.NewValue(tftypes.Number, quota)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildQuotaUserType(), cfg),
		Schema: s,
	}
}

// ---- tests ------------------------------------------------------------------

// TestQuotaUserResource_Create verifies Create populates ID and quota.
func TestQuotaUserResource_Create(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterQuotaHandlers(ms.Mux)

	r := newTestQuotaUserResource(t, ms)
	s := quotaUserResourceSchema(t).Schema

	plan := quotaUserPlan(t, "testfs", "1000", 1073741824)
	resp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildQuotaUserType(), nil), Schema: s},
	}

	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Create returned error: %s", resp.Diagnostics)
	}

	var model quotaUserModel
	if diags := resp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.ValueString() != "testfs/1000" {
		t.Errorf("expected ID=testfs/1000, got %s", model.ID.ValueString())
	}
	if model.FileSystemName.ValueString() != "testfs" {
		t.Errorf("expected file_system_name=testfs, got %s", model.FileSystemName.ValueString())
	}
	if model.UID.ValueString() != "1000" {
		t.Errorf("expected uid=1000, got %s", model.UID.ValueString())
	}
	if model.Quota.ValueInt64() != 1073741824 {
		t.Errorf("expected quota=1073741824, got %d", model.Quota.ValueInt64())
	}
}

// TestQuotaUserResource_Update verifies PATCH updates the quota limit.
func TestQuotaUserResource_Update(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterQuotaHandlers(ms.Mux)

	r := newTestQuotaUserResource(t, ms)
	s := quotaUserResourceSchema(t).Schema

	// Create first.
	createPlan := quotaUserPlan(t, "testfs", "1000", 1073741824)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildQuotaUserType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Update quota to 2GB.
	newPlan := quotaUserPlan(t, "testfs", "1000", 2147483648)
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildQuotaUserType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  newPlan,
		State: createResp.State,
	}, updateResp)

	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update returned error: %s", updateResp.Diagnostics)
	}

	var model quotaUserModel
	if diags := updateResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}
	if model.Quota.ValueInt64() != 2147483648 {
		t.Errorf("expected quota=2147483648 after update, got %d", model.Quota.ValueInt64())
	}
}

// TestQuotaUserResource_Delete verifies Delete removes the quota.
func TestQuotaUserResource_Delete(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterQuotaHandlers(ms.Mux)

	r := newTestQuotaUserResource(t, ms)
	s := quotaUserResourceSchema(t).Schema

	// Create first.
	plan := quotaUserPlan(t, "testfs", "1000", 1073741824)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildQuotaUserType(), nil), Schema: s},
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
	_, err := r.client.GetQuotaUser(context.Background(), "testfs", "1000")
	if err == nil || !client.IsNotFound(err) {
		t.Errorf("expected quota to be deleted, got: %v", err)
	}
}

// TestQuotaUserResource_Import verifies ImportState populates all attributes from composite ID.
func TestQuotaUserResource_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterQuotaHandlers(ms.Mux)

	r := newTestQuotaUserResource(t, ms)
	s := quotaUserResourceSchema(t).Schema

	// Create first via client directly.
	c, err := client.NewClient(context.Background(), client.Config{
		Endpoint:           ms.URL(),
		APIToken:           "test-token",
		InsecureSkipVerify: true,
		MaxRetries:         1,
		RetryBaseDelay:     1 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	_, err = c.PostQuotaUser(context.Background(), "testfs", "1000", client.QuotaUserPost{Quota: 1073741824})
	if err != nil {
		t.Fatalf("PostQuotaUser: %v", err)
	}

	// Import by composite ID "testfs/1000".
	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildQuotaUserType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "testfs/1000"}, importResp)

	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState returned error: %s", importResp.Diagnostics)
	}

	var model quotaUserModel
	if diags := importResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.ValueString() != "testfs/1000" {
		t.Errorf("expected ID=testfs/1000 after import, got %s", model.ID.ValueString())
	}
	if model.FileSystemName.ValueString() != "testfs" {
		t.Errorf("expected file_system_name=testfs after import, got %s", model.FileSystemName.ValueString())
	}
	if model.UID.ValueString() != "1000" {
		t.Errorf("expected uid=1000 after import, got %s", model.UID.ValueString())
	}
	if model.Quota.ValueInt64() != 1073741824 {
		t.Errorf("expected quota=1073741824 after import, got %d", model.Quota.ValueInt64())
	}
}

// ---- data source tests -------------------------------------------------------

// newTestQuotaUserDataSource creates a quotaUserDataSource wired to the given mock server.
func newTestQuotaUserDataSource(t *testing.T, ms *testmock.MockServer) *quotaUserDataSource {
	t.Helper()
	c, err := client.NewClient(context.Background(), client.Config{
		Endpoint:           ms.URL(),
		APIToken:           "test-token",
		InsecureSkipVerify: true,
		MaxRetries:         1,
		RetryBaseDelay:     1 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return &quotaUserDataSource{client: c}
}

// quotaUserDataSourceSchema returns the schema for the quota user data source.
func quotaUserDataSourceSchema(t *testing.T) datasource.SchemaResponse {
	t.Helper()
	d := &quotaUserDataSource{}
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	return resp
}

// buildQuotaUserDSType returns the tftypes.Object for the quota user data source.
func buildQuotaUserDSType() tftypes.Object {
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":               tftypes.String,
		"file_system_name": tftypes.String,
		"uid":              tftypes.String,
		"quota":            tftypes.Number,
		"usage":            tftypes.Number,
	}}
}

// nullQuotaUserDSConfig returns a base config map with all data source attributes null.
func nullQuotaUserDSConfig() map[string]tftypes.Value {
	return map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, nil),
		"file_system_name": tftypes.NewValue(tftypes.String, nil),
		"uid":              tftypes.NewValue(tftypes.String, nil),
		"quota":            tftypes.NewValue(tftypes.Number, nil),
		"usage":            tftypes.NewValue(tftypes.Number, nil),
	}
}

// TestQuotaUserDataSource verifies data source reads quota by file_system_name+uid and returns all attributes.
func TestQuotaUserDataSource(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterQuotaHandlers(ms.Mux)

	// Create a quota via the client so the data source can find it.
	c, err := client.NewClient(context.Background(), client.Config{
		Endpoint:           ms.URL(),
		APIToken:           "test-token",
		InsecureSkipVerify: true,
		MaxRetries:         1,
		RetryBaseDelay:     1 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	_, err = c.PostQuotaUser(context.Background(), "testfs", "1000", client.QuotaUserPost{Quota: 1073741824})
	if err != nil {
		t.Fatalf("PostQuotaUser: %v", err)
	}

	d := newTestQuotaUserDataSource(t, ms)
	s := quotaUserDataSourceSchema(t).Schema

	cfg := nullQuotaUserDSConfig()
	cfg["file_system_name"] = tftypes.NewValue(tftypes.String, "testfs")
	cfg["uid"] = tftypes.NewValue(tftypes.String, "1000")

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildQuotaUserDSType(), nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(buildQuotaUserDSType(), cfg), Schema: s},
	}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("DataSource Read returned error: %s", readResp.Diagnostics)
	}

	var model quotaUserDataSourceModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.ValueString() != "testfs/1000" {
		t.Errorf("expected ID=testfs/1000, got %s", model.ID.ValueString())
	}
	if model.FileSystemName.ValueString() != "testfs" {
		t.Errorf("expected file_system_name=testfs, got %s", model.FileSystemName.ValueString())
	}
	if model.UID.ValueString() != "1000" {
		t.Errorf("expected uid=1000, got %s", model.UID.ValueString())
	}
	if model.Quota.ValueInt64() != 1073741824 {
		t.Errorf("expected quota=1073741824, got %d", model.Quota.ValueInt64())
	}
}

// TestUnit_QuotaUser_Lifecycle exercises the full Create->Read->Update->Read->Delete sequence.
func TestUnit_QuotaUser_Lifecycle(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterQuotaHandlers(ms.Mux)

	r := newTestQuotaUserResource(t, ms)
	s := quotaUserResourceSchema(t).Schema

	// Step 1: Create.
	createPlan := quotaUserPlan(t, "lifecycle-fs", "4000", 1073741824)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildQuotaUserType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}
	var createModel quotaUserModel
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
	var readModel1 quotaUserModel
	if diags := readResp1.State.Get(context.Background(), &readModel1); diags.HasError() {
		t.Fatalf("Get read1 state: %s", diags)
	}
	if readModel1.UID.ValueString() != "4000" {
		t.Errorf("Read1: expected uid=4000, got %s", readModel1.UID.ValueString())
	}

	// Step 3: Update quota to 2GiB.
	updatePlan := quotaUserPlan(t, "lifecycle-fs", "4000", 2147483648)
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildQuotaUserType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlan,
		State: readResp1.State,
	}, updateResp)
	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update: %s", updateResp.Diagnostics)
	}
	var updateModel quotaUserModel
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
	var readModel2 quotaUserModel
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
	_, err := r.client.GetQuotaUser(context.Background(), "lifecycle-fs", "4000")
	if err == nil || !client.IsNotFound(err) {
		t.Errorf("expected quota to be deleted, got: %v", err)
	}
}

// TestUnit_QuotaUser_ImportIdempotency verifies ImportState->Read produces state matching original Create.
func TestUnit_QuotaUser_ImportIdempotency(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterQuotaHandlers(ms.Mux)

	r := newTestQuotaUserResource(t, ms)
	s := quotaUserResourceSchema(t).Schema

	// Create.
	createPlan := quotaUserPlan(t, "idempotent-fs", "6000", 1073741824)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildQuotaUserType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}
	var createModel quotaUserModel
	if diags := createResp.State.Get(context.Background(), &createModel); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}

	// ImportState using composite ID "fs_name/uid".
	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildQuotaUserType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "idempotent-fs/6000"}, importResp)
	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState: %s", importResp.Diagnostics)
	}

	// Read to populate full state.
	readResp := &resource.ReadResponse{State: importResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: importResp.State}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read post-import: %s", readResp.Diagnostics)
	}
	var importedModel quotaUserModel
	if diags := readResp.State.Get(context.Background(), &importedModel); diags.HasError() {
		t.Fatalf("Get imported state: %s", diags)
	}

	// Verify 0-diff.
	if importedModel.FileSystemName.ValueString() != createModel.FileSystemName.ValueString() {
		t.Errorf("file_system_name mismatch: create=%s import=%s", createModel.FileSystemName.ValueString(), importedModel.FileSystemName.ValueString())
	}
	if importedModel.UID.ValueString() != createModel.UID.ValueString() {
		t.Errorf("uid mismatch: create=%s import=%s", createModel.UID.ValueString(), importedModel.UID.ValueString())
	}
	if importedModel.Quota.ValueInt64() != createModel.Quota.ValueInt64() {
		t.Errorf("quota mismatch: create=%d import=%d", createModel.Quota.ValueInt64(), importedModel.Quota.ValueInt64())
	}
}

// TestUnit_QuotaUser_PlanModifiers verifies all RequiresReplace and UseStateForUnknown
// plan modifiers in the quota_user resource schema.
func TestUnit_QuotaUser_PlanModifiers(t *testing.T) {
	s := quotaUserResourceSchema(t).Schema

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

	// uid — RequiresReplace
	uidAttr, ok := s.Attributes["uid"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("uid attribute not found or wrong type")
	}
	if len(uidAttr.PlanModifiers) == 0 {
		t.Error("expected RequiresReplace plan modifier on uid attribute")
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

// TestUnit_QuotaUser_QuotaValidator verifies the quota field rejects negative values
// and accepts 0 and positive values.
func TestUnit_QuotaUser_QuotaValidator(t *testing.T) {
	s := quotaUserResourceSchema(t).Schema

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
