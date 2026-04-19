package provider

import (
	"context"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/numberly/opentofu-provider-flashblade/internal/client"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock/handlers"
)

// ---- helpers ----------------------------------------------------------------

// newTestOAPResource creates an objectStoreAccessPolicyResource wired to the given mock server.
func newTestOAPResource(t *testing.T, ms *testmock.MockServer) *objectStoreAccessPolicyResource {
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
	return &objectStoreAccessPolicyResource{client: c}
}

// oapResourceSchema returns the parsed schema for the OAP resource.
func oapResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &objectStoreAccessPolicyResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

// buildObjectStoreAccessPolicyType returns the tftypes.Object for the OAP resource.
func buildObjectStoreAccessPolicyType() tftypes.Object {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":          tftypes.String,
		"name":        tftypes.String,
		"description": tftypes.String,
		"arn":         tftypes.String,
		"enabled":     tftypes.Bool,
		"is_local":    tftypes.Bool,
		"policy_type": tftypes.String,
		"timeouts":    timeoutsType,
	}}
}

// nullOAPConfig returns a base config map with all attributes null.
func nullOAPConfig() map[string]tftypes.Value {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return map[string]tftypes.Value{
		"id":          tftypes.NewValue(tftypes.String, nil),
		"name":        tftypes.NewValue(tftypes.String, nil),
		"description": tftypes.NewValue(tftypes.String, nil),
		"arn":         tftypes.NewValue(tftypes.String, nil),
		"enabled":     tftypes.NewValue(tftypes.Bool, nil),
		"is_local":    tftypes.NewValue(tftypes.Bool, nil),
		"policy_type": tftypes.NewValue(tftypes.String, nil),
		"timeouts":    tftypes.NewValue(timeoutsType, nil),
	}
}

// oapPlanWithName returns a tfsdk.Plan with the given policy name.
func oapPlanWithName(t *testing.T, name string) tfsdk.Plan {
	t.Helper()
	s := oapResourceSchema(t).Schema
	cfg := nullOAPConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, name)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildObjectStoreAccessPolicyType(), cfg),
		Schema: s,
	}
}

// oapPlanWithNameAndDescription returns a tfsdk.Plan with name and description.
func oapPlanWithNameAndDescription(t *testing.T, name, description string) tfsdk.Plan {
	t.Helper()
	s := oapResourceSchema(t).Schema
	cfg := nullOAPConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, name)
	cfg["description"] = tftypes.NewValue(tftypes.String, description)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildObjectStoreAccessPolicyType(), cfg),
		Schema: s,
	}
}

// ---- tests ------------------------------------------------------------------

// TestObjectStoreAccessPolicyResource_Create verifies Create populates ID, name, enabled, and description.
func TestObjectStoreAccessPolicyResource_Create(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterObjectStoreAccessPolicyHandlers(ms.Mux)

	r := newTestOAPResource(t, ms)
	s := oapResourceSchema(t).Schema

	plan := oapPlanWithNameAndDescription(t, "test-oap-policy", "test description")
	resp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildObjectStoreAccessPolicyType(), nil), Schema: s},
	}

	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Create returned error: %s", resp.Diagnostics)
	}

	var model objectStoreAccessPolicyModel
	if diags := resp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected non-empty ID after Create")
	}
	if model.Name.ValueString() != "test-oap-policy" {
		t.Errorf("expected name=test-oap-policy, got %s", model.Name.ValueString())
	}
	if model.Description.ValueString() != "test description" {
		t.Errorf("expected description='test description', got %s", model.Description.ValueString())
	}
	if !model.Enabled.ValueBool() {
		t.Error("expected enabled=true after Create")
	}
	if model.PolicyType.ValueString() != "object-store-access" {
		t.Errorf("expected policy_type=object-store-access, got %s", model.PolicyType.ValueString())
	}
}

// TestObjectStoreAccessPolicyResource_Update verifies PATCH supports rename (name change).
func TestObjectStoreAccessPolicyResource_Update(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterObjectStoreAccessPolicyHandlers(ms.Mux)

	r := newTestOAPResource(t, ms)
	s := oapResourceSchema(t).Schema

	// Create first.
	createPlan := oapPlanWithName(t, "oap-rename-before")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildObjectStoreAccessPolicyType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Rename the policy in-place via PATCH.
	renamePlan := oapPlanWithName(t, "oap-rename-after")
	renameResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildObjectStoreAccessPolicyType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  renamePlan,
		State: createResp.State,
	}, renameResp)

	if renameResp.Diagnostics.HasError() {
		t.Fatalf("Update (rename) returned error: %s", renameResp.Diagnostics)
	}

	var model objectStoreAccessPolicyModel
	if diags := renameResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Name.ValueString() != "oap-rename-after" {
		t.Errorf("expected name=oap-rename-after after rename, got %s", model.Name.ValueString())
	}
}

// TestObjectStoreAccessPolicyResource_Delete verifies DELETE removes the policy.
func TestObjectStoreAccessPolicyResource_Delete(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterObjectStoreAccessPolicyHandlers(ms.Mux)
	// Register buckets handler for the delete-guard (nil accounts store — GET-only, no POST needed).
	handlers.RegisterBucketHandlers(ms.Mux, nil)

	r := newTestOAPResource(t, ms)
	s := oapResourceSchema(t).Schema

	// Create first.
	plan := oapPlanWithName(t, "delete-oap-policy")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildObjectStoreAccessPolicyType(), nil), Schema: s},
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

	// Verify policy is gone.
	_, err := r.client.GetObjectStoreAccessPolicy(context.Background(), "delete-oap-policy")
	if err == nil || !client.IsNotFound(err) {
		t.Errorf("expected policy to be deleted, got: %v", err)
	}
}

// TestObjectStoreAccessPolicyResource_Import verifies ImportState populates all attributes.
func TestObjectStoreAccessPolicyResource_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterObjectStoreAccessPolicyHandlers(ms.Mux)

	r := newTestOAPResource(t, ms)
	s := oapResourceSchema(t).Schema

	// Create first.
	plan := oapPlanWithName(t, "import-oap-policy")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildObjectStoreAccessPolicyType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Import by name.
	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildObjectStoreAccessPolicyType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "import-oap-policy"}, importResp)

	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState returned error: %s", importResp.Diagnostics)
	}

	var model objectStoreAccessPolicyModel
	if diags := importResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Name.ValueString() != "import-oap-policy" {
		t.Errorf("expected name=import-oap-policy after import, got %s", model.Name.ValueString())
	}
	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected ID to be populated after import")
	}
}

// ---- data source tests -------------------------------------------------------

// newTestOAPDataSource creates an objectStoreAccessPolicyDataSource wired to the given mock server.
func newTestOAPDataSource(t *testing.T, ms *testmock.MockServer) *objectStoreAccessPolicyDataSource {
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
	return &objectStoreAccessPolicyDataSource{client: c}
}

// oapDataSourceSchema returns the schema for the OAP data source.
func oapDataSourceSchema(t *testing.T) datasource.SchemaResponse {
	t.Helper()
	d := &objectStoreAccessPolicyDataSource{}
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	return resp
}

// buildObjectStoreAccessPolicyDSType returns the tftypes.Object for the OAP data source.
func buildObjectStoreAccessPolicyDSType() tftypes.Object {
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":          tftypes.String,
		"name":        tftypes.String,
		"description": tftypes.String,
		"arn":         tftypes.String,
		"enabled":     tftypes.Bool,
		"is_local":    tftypes.Bool,
		"policy_type": tftypes.String,
	}}
}

// nullOAPDSConfig returns a base config map with all data source attributes null.
func nullOAPDSConfig() map[string]tftypes.Value {
	return map[string]tftypes.Value{
		"id":          tftypes.NewValue(tftypes.String, nil),
		"name":        tftypes.NewValue(tftypes.String, nil),
		"description": tftypes.NewValue(tftypes.String, nil),
		"arn":         tftypes.NewValue(tftypes.String, nil),
		"enabled":     tftypes.NewValue(tftypes.Bool, nil),
		"is_local":    tftypes.NewValue(tftypes.Bool, nil),
		"policy_type": tftypes.NewValue(tftypes.String, nil),
	}
}

// TestObjectStoreAccessPolicyDataSource verifies data source reads policy by name and returns all attributes.
func TestObjectStoreAccessPolicyDataSource(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterObjectStoreAccessPolicyHandlers(ms.Mux)

	// Create a policy via the client so the data source can find it.
	c, err := client.NewClient(context.Background(), client.Config{
		Endpoint:           ms.URL(),
		APIToken:           "test-token",
		InsecureSkipVerify: true,
		MaxRetries:         1,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	_, err = c.PostObjectStoreAccessPolicy(context.Background(), "ds-oap-test", client.ObjectStoreAccessPolicyPost{
		Description: "datasource test",
	})
	if err != nil {
		t.Fatalf("PostObjectStoreAccessPolicy: %v", err)
	}

	d := newTestOAPDataSource(t, ms)
	s := oapDataSourceSchema(t).Schema

	cfg := nullOAPDSConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "ds-oap-test")

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildObjectStoreAccessPolicyDSType(), nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(buildObjectStoreAccessPolicyDSType(), cfg), Schema: s},
	}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("DataSource Read returned error: %s", readResp.Diagnostics)
	}

	var model objectStoreAccessPolicyDataSourceModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Name.ValueString() != "ds-oap-test" {
		t.Errorf("expected name=ds-oap-test, got %s", model.Name.ValueString())
	}
	if !model.Enabled.ValueBool() {
		t.Error("expected enabled=true")
	}
	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected ID to be populated")
	}
	if model.Description.ValueString() != "datasource test" {
		t.Errorf("expected description='datasource test', got %s", model.Description.ValueString())
	}
}

// TestUnit_OAP_Create_Conflict verifies that a 409 Conflict on POST produces
// an error diagnostic.
func TestUnit_OAP_Create_Conflict(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	ms.RegisterHandler("/api/2.22/object-store-access-policies", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handlers.WriteJSONError(w, http.StatusConflict, "Policy with the given name already exists.")
			return
		}
		handlers.WriteJSONListResponse(w, http.StatusOK, []client.ObjectStoreAccessPolicy{})
	})

	r := newTestOAPResource(t, ms)
	s := oapResourceSchema(t).Schema

	plan := oapPlanWithName(t, "conflict-oap")
	resp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildObjectStoreAccessPolicyType(), nil), Schema: s},
	}

	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("expected Create to produce an error diagnostic on 409 Conflict, got none")
	}
}

// TestUnit_OAP_Read_NotFound verifies that a not-found response (empty items)
// during Read removes the resource from Terraform state without an error diagnostic.
func TestUnit_OAP_Read_NotFound(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	ms.RegisterHandler("/api/2.22/object-store-access-policies", func(w http.ResponseWriter, r *http.Request) {
		handlers.WriteJSONListResponse(w, http.StatusOK, []client.ObjectStoreAccessPolicy{})
	})

	r := newTestOAPResource(t, ms)
	s := oapResourceSchema(t).Schema

	cfg := nullOAPConfig()
	cfg["id"] = tftypes.NewValue(tftypes.String, "oap-gone-id")
	cfg["name"] = tftypes.NewValue(tftypes.String, "gone-oap")
	state := tfsdk.State{Raw: tftypes.NewValue(buildObjectStoreAccessPolicyType(), cfg), Schema: s}

	readResp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), resource.ReadRequest{State: state}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}
	if !readResp.State.Raw.IsNull() {
		t.Error("expected state to be removed (null) when OAP not found")
	}
}

// TestUnit_ObjectStoreAccessPolicy_Lifecycle exercises the full Create->Read->Update->Read->Delete sequence.
func TestUnit_ObjectStoreAccessPolicy_Lifecycle(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterObjectStoreAccessPolicyHandlers(ms.Mux)
	handlers.RegisterBucketHandlers(ms.Mux, nil)

	r := newTestOAPResource(t, ms)
	s := oapResourceSchema(t).Schema

	// Step 1: Create.
	createPlan := oapPlanWithName(t, "lifecycle-oap-policy")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildObjectStoreAccessPolicyType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}
	var createModel objectStoreAccessPolicyModel
	if diags := createResp.State.Get(context.Background(), &createModel); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}
	if createModel.Name.ValueString() != "lifecycle-oap-policy" {
		t.Errorf("Create: expected name=lifecycle-oap-policy, got %s", createModel.Name.ValueString())
	}

	// Step 2: Read post-create.
	readResp1 := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp1)
	if readResp1.Diagnostics.HasError() {
		t.Fatalf("Read post-create: %s", readResp1.Diagnostics)
	}
	var readModel1 objectStoreAccessPolicyModel
	if diags := readResp1.State.Get(context.Background(), &readModel1); diags.HasError() {
		t.Fatalf("Get read1 state: %s", diags)
	}
	if readModel1.ID.IsNull() || readModel1.ID.ValueString() == "" {
		t.Error("Read1: expected non-empty ID")
	}

	// Step 3: Update (rename policy — description is RequiresReplace, name is mutable).
	updatePlan := oapPlanWithName(t, "lifecycle-oap-policy-renamed")
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildObjectStoreAccessPolicyType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlan,
		State: readResp1.State,
	}, updateResp)
	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update: %s", updateResp.Diagnostics)
	}
	var updateModel objectStoreAccessPolicyModel
	if diags := updateResp.State.Get(context.Background(), &updateModel); diags.HasError() {
		t.Fatalf("Get update state: %s", diags)
	}
	if updateModel.Name.ValueString() != "lifecycle-oap-policy-renamed" {
		t.Errorf("Update: expected name=lifecycle-oap-policy-renamed, got %s", updateModel.Name.ValueString())
	}

	// Step 4: Read post-update.
	readResp2 := &resource.ReadResponse{State: updateResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: updateResp.State}, readResp2)
	if readResp2.Diagnostics.HasError() {
		t.Fatalf("Read post-update: %s", readResp2.Diagnostics)
	}
	var readModel2 objectStoreAccessPolicyModel
	if diags := readResp2.State.Get(context.Background(), &readModel2); diags.HasError() {
		t.Fatalf("Get read2 state: %s", diags)
	}
	if readModel2.Name.ValueString() != "lifecycle-oap-policy-renamed" {
		t.Errorf("Read2: expected name=lifecycle-oap-policy-renamed, got %s", readModel2.Name.ValueString())
	}

	// Step 5: Delete.
	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: readResp2.State}, deleteResp)
	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete: %s", deleteResp.Diagnostics)
	}
	_, err := r.client.GetObjectStoreAccessPolicy(context.Background(), "lifecycle-oap-policy-renamed")
	if err == nil || !client.IsNotFound(err) {
		t.Errorf("expected policy to be deleted, got: %v", err)
	}
}

// TestUnit_ObjectStoreAccessPolicy_ImportIdempotency verifies ImportState->Read produces state matching original Create.
func TestUnit_ObjectStoreAccessPolicy_ImportIdempotency(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterObjectStoreAccessPolicyHandlers(ms.Mux)

	r := newTestOAPResource(t, ms)
	s := oapResourceSchema(t).Schema

	// Create.
	createPlan := oapPlanWithName(t, "idempotent-oap-policy")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildObjectStoreAccessPolicyType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}
	var createModel objectStoreAccessPolicyModel
	if diags := createResp.State.Get(context.Background(), &createModel); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}

	// ImportState.
	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildObjectStoreAccessPolicyType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "idempotent-oap-policy"}, importResp)
	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState: %s", importResp.Diagnostics)
	}

	// Read to populate full state.
	readResp := &resource.ReadResponse{State: importResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: importResp.State}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read post-import: %s", readResp.Diagnostics)
	}
	var importedModel objectStoreAccessPolicyModel
	if diags := readResp.State.Get(context.Background(), &importedModel); diags.HasError() {
		t.Fatalf("Get imported state: %s", diags)
	}

	// Verify 0-diff.
	if importedModel.Name.ValueString() != createModel.Name.ValueString() {
		t.Errorf("name mismatch: create=%s import=%s", createModel.Name.ValueString(), importedModel.Name.ValueString())
	}
	if importedModel.ID.ValueString() != createModel.ID.ValueString() {
		t.Errorf("id mismatch: create=%s import=%s", createModel.ID.ValueString(), importedModel.ID.ValueString())
	}
}

// TestUnit_OAP_PlanModifiers verifies all RequiresReplace and UseStateForUnknown
// plan modifiers in the object_store_access_policy resource schema.
func TestUnit_OAP_PlanModifiers(t *testing.T) {
	s := oapResourceSchema(t).Schema

	// id — UseStateForUnknown
	idAttr, ok := s.Attributes["id"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("id attribute not found or wrong type")
	}
	if len(idAttr.PlanModifiers) == 0 {
		t.Error("expected UseStateForUnknown plan modifier on id attribute")
	}

	// description — RequiresReplace + UseStateForUnknown
	descAttr, ok := s.Attributes["description"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("description attribute not found or wrong type")
	}
	if len(descAttr.PlanModifiers) < 2 {
		t.Error("expected RequiresReplace and UseStateForUnknown plan modifiers on description attribute")
	}

	// is_local — UseStateForUnknown
	ilAttr, ok := s.Attributes["is_local"].(resschema.BoolAttribute)
	if !ok {
		t.Fatal("is_local attribute not found or wrong type")
	}
	if len(ilAttr.PlanModifiers) == 0 {
		t.Error("expected UseStateForUnknown plan modifier on is_local attribute")
	}

	// policy_type — UseStateForUnknown
	ptAttr, ok := s.Attributes["policy_type"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("policy_type attribute not found or wrong type")
	}
	if len(ptAttr.PlanModifiers) == 0 {
		t.Error("expected UseStateForUnknown plan modifier on policy_type attribute")
	}
}
