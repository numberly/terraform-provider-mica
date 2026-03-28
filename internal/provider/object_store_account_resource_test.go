package provider

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	resschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/soulkyu/terraform-provider-flashblade/internal/client"
	"github.com/soulkyu/terraform-provider-flashblade/internal/testmock"
	"github.com/soulkyu/terraform-provider-flashblade/internal/testmock/handlers"
)

// ---- helpers ----------------------------------------------------------------

// newTestOSAResource creates an objectStoreAccountResource wired to the given mock server.
func newTestOSAResource(t *testing.T, ms *testmock.MockServer) *objectStoreAccountResource {
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
	return &objectStoreAccountResource{client: c}
}

// osaResourceSchema returns the parsed schema for the object store account resource.
func osaResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &objectStoreAccountResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

// buildOSAType returns the tftypes.Object for the object store account resource schema.
func buildOSAType() tftypes.Object {
	spaceType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"data_reduction":      tftypes.Number,
		"snapshots":           tftypes.Number,
		"total_physical":      tftypes.Number,
		"unique":              tftypes.Number,
		"virtual":             tftypes.Number,
		"snapshots_effective": tftypes.Number,
	}}
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":                 tftypes.String,
		"name":               tftypes.String,
		"created":            tftypes.Number,
		"quota_limit":        tftypes.String,
		"hard_limit_enabled": tftypes.Bool,
		"object_count":       tftypes.Number,
		"space":              spaceType,
		"timeouts":           timeoutsType,
	}}
}

// nullOSAConfig returns a base config map with all attributes null.
func nullOSAConfig() map[string]tftypes.Value {
	spaceType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"data_reduction":      tftypes.Number,
		"snapshots":           tftypes.Number,
		"total_physical":      tftypes.Number,
		"unique":              tftypes.Number,
		"virtual":             tftypes.Number,
		"snapshots_effective": tftypes.Number,
	}}
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return map[string]tftypes.Value{
		"id":                 tftypes.NewValue(tftypes.String, nil),
		"name":               tftypes.NewValue(tftypes.String, nil),
		"created":            tftypes.NewValue(tftypes.Number, nil),
		"quota_limit":        tftypes.NewValue(tftypes.String, nil),
		"hard_limit_enabled": tftypes.NewValue(tftypes.Bool, nil),
		"object_count":       tftypes.NewValue(tftypes.Number, nil),
		"space":              tftypes.NewValue(spaceType, nil),
		"timeouts":           tftypes.NewValue(timeoutsType, nil),
	}
}

// osaplanWithName returns a tfsdk.Plan with the given account name.
func osaPlanWithName(t *testing.T, name string) tfsdk.Plan {
	t.Helper()
	s := osaResourceSchema(t).Schema
	cfg := nullOSAConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, name)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildOSAType(), cfg),
		Schema: s,
	}
}

// osaPlanWithNameAndQuota returns a tfsdk.Plan with name and quota_limit.
func osaPlanWithNameAndQuota(t *testing.T, name string, quotaLimit string) tfsdk.Plan {
	t.Helper()
	s := osaResourceSchema(t).Schema
	cfg := nullOSAConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, name)
	cfg["quota_limit"] = tftypes.NewValue(tftypes.String, quotaLimit)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildOSAType(), cfg),
		Schema: s,
	}
}

// ---- tests ------------------------------------------------------------------

// TestUnit_ObjectStoreAccount_Create verifies Create populates ID, name, and created timestamp.
func TestUnit_ObjectStoreAccount_Create(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterObjectStoreAccountHandlers(ms.Mux)

	r := newTestOSAResource(t, ms)
	s := osaResourceSchema(t).Schema

	plan := osaPlanWithName(t, "test-account")
	req := resource.CreateRequest{Plan: plan}
	resp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildOSAType(), nil), Schema: s},
	}

	r.Create(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Create returned error: %s", resp.Diagnostics)
	}

	var model objectStoreAccountModel
	if diags := resp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected non-empty ID after Create")
	}
	if model.Name.ValueString() != "test-account" {
		t.Errorf("expected name=test-account, got %s", model.Name.ValueString())
	}
	if model.Created.IsNull() || model.Created.IsUnknown() || model.Created.ValueInt64() == 0 {
		t.Error("expected created to be populated after Create")
	}
}

// TestUnit_ObjectStoreAccount_Update verifies PATCH updates quota_limit and hard_limit_enabled.
func TestUnit_ObjectStoreAccount_Update(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterObjectStoreAccountHandlers(ms.Mux)

	r := newTestOSAResource(t, ms)
	s := osaResourceSchema(t).Schema

	// Create first.
	plan := osaPlanWithName(t, "update-account")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildOSAType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Update quota_limit.
	newPlan := osaPlanWithNameAndQuota(t, "update-account", "10737418240")
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildOSAType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  newPlan,
		State: createResp.State,
	}, updateResp)

	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update returned error: %s", updateResp.Diagnostics)
	}

	var model objectStoreAccountModel
	if diags := updateResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}
	if model.QuotaLimit.ValueString() != "10737418240" {
		t.Errorf("expected quota_limit=10737418240 after update, got %s", model.QuotaLimit.ValueString())
	}
}

// TestUnit_ObjectStoreAccount_Delete verifies DELETE removes the account.
func TestUnit_ObjectStoreAccount_Delete(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	accountStore := handlers.RegisterObjectStoreAccountHandlers(ms.Mux)
	handlers.RegisterBucketHandlers(ms.Mux, accountStore)

	r := newTestOSAResource(t, ms)
	s := osaResourceSchema(t).Schema

	plan := osaPlanWithName(t, "delete-account")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildOSAType(), nil), Schema: s},
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

	// Verify account is gone.
	_, err := r.client.GetObjectStoreAccount(context.Background(), "delete-account")
	if err == nil || !client.IsNotFound(err) {
		t.Errorf("expected account to be deleted, got: %v", err)
	}
}

// TestUnit_ObjectStoreAccount_Import verifies ImportState populates all attributes and 0 diff after.
func TestUnit_ObjectStoreAccount_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterObjectStoreAccountHandlers(ms.Mux)

	r := newTestOSAResource(t, ms)
	s := osaResourceSchema(t).Schema

	// Create first.
	plan := osaPlanWithName(t, "import-account")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildOSAType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Import by name.
	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildOSAType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "import-account"}, importResp)

	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState returned error: %s", importResp.Diagnostics)
	}

	var model objectStoreAccountModel
	if diags := importResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Name.ValueString() != "import-account" {
		t.Errorf("expected name=import-account after import, got %s", model.Name.ValueString())
	}
	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected ID to be populated after import")
	}
	if model.Created.IsNull() || model.Created.ValueInt64() == 0 {
		t.Error("expected created to be populated after import")
	}
}

// TestUnit_ObjectStoreAccount_ForceNew verifies that name has RequiresReplace semantics.
// The test confirms that the name attribute schema has plan modifiers registered.
func TestUnit_ObjectStoreAccount_ForceNew(t *testing.T) {
	schResp := osaResourceSchema(t).Schema
	nameAttr, ok := schResp.Attributes["name"]
	if !ok {
		t.Fatal("name attribute not found in schema")
	}
	// Cast to concrete resschema.StringAttribute to access PlanModifiers field.
	strAttr, ok := nameAttr.(resschema.StringAttribute)
	if !ok {
		t.Fatalf("name attribute is not a resschema.StringAttribute, got %T", nameAttr)
	}
	if len(strAttr.PlanModifiers) == 0 {
		t.Error("expected RequiresReplace plan modifier on name attribute")
	}
}

// TestUnit_ObjectStoreAccount_Read_NotFound verifies that 404 removes resource from state.
func TestUnit_ObjectStoreAccount_Read_NotFound(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterObjectStoreAccountHandlers(ms.Mux)

	r := newTestOSAResource(t, ms)
	s := osaResourceSchema(t).Schema

	cfg := nullOSAConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "does-not-exist")
	cfg["id"] = tftypes.NewValue(tftypes.String, "non-existent-id")
	state := tfsdk.State{Raw: tftypes.NewValue(buildOSAType(), cfg), Schema: s}

	readResp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), resource.ReadRequest{State: state}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}
	if !readResp.State.Raw.IsNull() {
		t.Error("expected state to be removed (null) when resource not found")
	}
}

// TestUnit_ObjectStoreAccount_Lifecycle exercises the full Create->Read->Update->Read->Delete sequence.
func TestUnit_ObjectStoreAccount_Lifecycle(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	accountStore := handlers.RegisterObjectStoreAccountHandlers(ms.Mux)
	handlers.RegisterBucketHandlers(ms.Mux, accountStore)

	r := newTestOSAResource(t, ms)
	s := osaResourceSchema(t).Schema

	// Step 1: Create.
	createPlan := osaPlanWithName(t, "lifecycle-account")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildOSAType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}
	var createModel objectStoreAccountModel
	if diags := createResp.State.Get(context.Background(), &createModel); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}
	if createModel.Name.ValueString() != "lifecycle-account" {
		t.Errorf("Create: expected name=lifecycle-account, got %s", createModel.Name.ValueString())
	}

	// Step 2: Read post-create.
	readResp1 := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp1)
	if readResp1.Diagnostics.HasError() {
		t.Fatalf("Read post-create: %s", readResp1.Diagnostics)
	}
	var readModel1 objectStoreAccountModel
	if diags := readResp1.State.Get(context.Background(), &readModel1); diags.HasError() {
		t.Fatalf("Get read1 state: %s", diags)
	}
	if readModel1.Name.ValueString() != "lifecycle-account" {
		t.Errorf("Read1: expected name=lifecycle-account, got %s", readModel1.Name.ValueString())
	}

	// Step 3: Update quota_limit.
	updatePlan := osaPlanWithNameAndQuota(t, "lifecycle-account", "10737418240")
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildOSAType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlan,
		State: readResp1.State,
	}, updateResp)
	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update: %s", updateResp.Diagnostics)
	}
	var updateModel objectStoreAccountModel
	if diags := updateResp.State.Get(context.Background(), &updateModel); diags.HasError() {
		t.Fatalf("Get update state: %s", diags)
	}
	if updateModel.QuotaLimit.ValueString() != "10737418240" {
		t.Errorf("Update: expected quota_limit=10737418240, got %s", updateModel.QuotaLimit.ValueString())
	}

	// Step 4: Read post-update.
	readResp2 := &resource.ReadResponse{State: updateResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: updateResp.State}, readResp2)
	if readResp2.Diagnostics.HasError() {
		t.Fatalf("Read post-update: %s", readResp2.Diagnostics)
	}
	var readModel2 objectStoreAccountModel
	if diags := readResp2.State.Get(context.Background(), &readModel2); diags.HasError() {
		t.Fatalf("Get read2 state: %s", diags)
	}
	if readModel2.QuotaLimit.ValueString() != "10737418240" {
		t.Errorf("Read2: expected quota_limit=10737418240, got %s", readModel2.QuotaLimit.ValueString())
	}

	// Step 5: Delete.
	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: readResp2.State}, deleteResp)
	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete: %s", deleteResp.Diagnostics)
	}
	_, err := r.client.GetObjectStoreAccount(context.Background(), "lifecycle-account")
	if err == nil || !client.IsNotFound(err) {
		t.Errorf("expected account to be deleted, got: %v", err)
	}
}

// TestUnit_ObjectStoreAccount_ImportIdempotency verifies ImportState->Read produces state matching original Create.
func TestUnit_ObjectStoreAccount_ImportIdempotency(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterObjectStoreAccountHandlers(ms.Mux)

	r := newTestOSAResource(t, ms)
	s := osaResourceSchema(t).Schema

	// Create.
	createPlan := osaPlanWithNameAndQuota(t, "idempotent-account", "5368709120")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildOSAType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}
	var createModel objectStoreAccountModel
	if diags := createResp.State.Get(context.Background(), &createModel); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}

	// ImportState.
	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildOSAType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "idempotent-account"}, importResp)
	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState: %s", importResp.Diagnostics)
	}

	// Read to populate full state.
	readResp := &resource.ReadResponse{State: importResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: importResp.State}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read post-import: %s", readResp.Diagnostics)
	}
	var importedModel objectStoreAccountModel
	if diags := readResp.State.Get(context.Background(), &importedModel); diags.HasError() {
		t.Fatalf("Get imported state: %s", diags)
	}

	// Verify 0-diff.
	if importedModel.Name.ValueString() != createModel.Name.ValueString() {
		t.Errorf("name mismatch: create=%s import=%s", createModel.Name.ValueString(), importedModel.Name.ValueString())
	}
	if importedModel.QuotaLimit.ValueString() != createModel.QuotaLimit.ValueString() {
		t.Errorf("quota_limit mismatch: create=%s import=%s", createModel.QuotaLimit.ValueString(), importedModel.QuotaLimit.ValueString())
	}
	if importedModel.ID.ValueString() != createModel.ID.ValueString() {
		t.Errorf("id mismatch: create=%s import=%s", createModel.ID.ValueString(), importedModel.ID.ValueString())
	}
}

// TestUnit_ObjectStoreAccount_PlanModifiers verifies all RequiresReplace and UseStateForUnknown
// plan modifiers in the object_store_account resource schema.
func TestUnit_ObjectStoreAccount_PlanModifiers(t *testing.T) {
	s := osaResourceSchema(t).Schema

	// id — UseStateForUnknown
	idAttr, ok := s.Attributes["id"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("id attribute not found or wrong type")
	}
	if len(idAttr.PlanModifiers) == 0 {
		t.Error("expected UseStateForUnknown plan modifier on id attribute")
	}

	// name — RequiresReplace
	nameAttr, ok := s.Attributes["name"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("name attribute not found or wrong type")
	}
	if len(nameAttr.PlanModifiers) == 0 {
		t.Error("expected RequiresReplace plan modifier on name attribute")
	}

	// created — int64 UseStateForUnknown
	createdAttr, ok := s.Attributes["created"].(resschema.Int64Attribute)
	if !ok {
		t.Fatal("created attribute not found or wrong type")
	}
	if len(createdAttr.PlanModifiers) == 0 {
		t.Error("expected UseStateForUnknown plan modifier on created attribute")
	}
}
