package provider

import (
	"context"
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

// newTestSnapshotPolicyResource creates a snapshotPolicyResource wired to the given mock server.
func newTestSnapshotPolicyResource(t *testing.T, ms *testmock.MockServer) *snapshotPolicyResource {
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
	return &snapshotPolicyResource{client: c}
}

// snapshotPolicyResourceSchema returns the parsed schema for the snapshot policy resource.
func snapshotPolicyResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &snapshotPolicyResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

// buildSnapshotPolicyType returns the tftypes.Object for the snapshot policy resource.
func buildSnapshotPolicyType() tftypes.Object {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":             tftypes.String,
		"name":           tftypes.String,
		"enabled":        tftypes.Bool,
		"is_local":       tftypes.Bool,
		"policy_type":    tftypes.String,
		"retention_lock": tftypes.String,
		"timeouts":       timeoutsType,
	}}
}

// nullSnapshotPolicyConfig returns a base config map with all attributes null.
func nullSnapshotPolicyConfig() map[string]tftypes.Value {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return map[string]tftypes.Value{
		"id":             tftypes.NewValue(tftypes.String, nil),
		"name":           tftypes.NewValue(tftypes.String, nil),
		"enabled":        tftypes.NewValue(tftypes.Bool, nil),
		"is_local":       tftypes.NewValue(tftypes.Bool, nil),
		"policy_type":    tftypes.NewValue(tftypes.String, nil),
		"retention_lock": tftypes.NewValue(tftypes.String, nil),
		"timeouts":       tftypes.NewValue(timeoutsType, nil),
	}
}

// snapshotPolicyPlanWithName returns a tfsdk.Plan with the given policy name.
func snapshotPolicyPlanWithName(t *testing.T, name string) tfsdk.Plan {
	t.Helper()
	s := snapshotPolicyResourceSchema(t).Schema
	cfg := nullSnapshotPolicyConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, name)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildSnapshotPolicyType(), cfg),
		Schema: s,
	}
}

// snapshotPolicyPlanWithNameAndEnabled returns a tfsdk.Plan with name and enabled flag.
func snapshotPolicyPlanWithNameAndEnabled(t *testing.T, name string, enabled bool) tfsdk.Plan {
	t.Helper()
	s := snapshotPolicyResourceSchema(t).Schema
	cfg := nullSnapshotPolicyConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, name)
	cfg["enabled"] = tftypes.NewValue(tftypes.Bool, enabled)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildSnapshotPolicyType(), cfg),
		Schema: s,
	}
}

// ---- tests ------------------------------------------------------------------

// TestSnapshotPolicyResource_Create verifies Create populates ID, name, and enabled.
func TestSnapshotPolicyResource_Create(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterSnapshotPolicyHandlers(ms.Mux)

	r := newTestSnapshotPolicyResource(t, ms)
	s := snapshotPolicyResourceSchema(t).Schema

	plan := snapshotPolicyPlanWithNameAndEnabled(t, "test-snap-policy", true)
	resp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSnapshotPolicyType(), nil), Schema: s},
	}

	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Create returned error: %s", resp.Diagnostics)
	}

	var model snapshotPolicyModel
	if diags := resp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected non-empty ID after Create")
	}
	if model.Name.ValueString() != "test-snap-policy" {
		t.Errorf("expected name=test-snap-policy, got %s", model.Name.ValueString())
	}
	if !model.Enabled.ValueBool() {
		t.Error("expected enabled=true after Create")
	}
}

// TestSnapshotPolicyResource_Update verifies PATCH updates enabled flag in-place.
// Name change is RequiresReplace so we only test enabled update here.
func TestSnapshotPolicyResource_Update(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterSnapshotPolicyHandlers(ms.Mux)

	r := newTestSnapshotPolicyResource(t, ms)
	s := snapshotPolicyResourceSchema(t).Schema

	// Create first.
	createPlan := snapshotPolicyPlanWithNameAndEnabled(t, "update-snap-policy", true)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSnapshotPolicyType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Update enabled=false (in-place, name does NOT change).
	newPlan := snapshotPolicyPlanWithNameAndEnabled(t, "update-snap-policy", false)
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSnapshotPolicyType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  newPlan,
		State: createResp.State,
	}, updateResp)

	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update returned error: %s", updateResp.Diagnostics)
	}

	var model snapshotPolicyModel
	if diags := updateResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}
	if model.Enabled.ValueBool() {
		t.Error("expected enabled=false after update")
	}
	if model.Name.ValueString() != "update-snap-policy" {
		t.Errorf("expected name unchanged, got %s", model.Name.ValueString())
	}
}

// TestSnapshotPolicyResource_Delete verifies DELETE removes the policy.
func TestSnapshotPolicyResource_Delete(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterSnapshotPolicyHandlers(ms.Mux)

	r := newTestSnapshotPolicyResource(t, ms)
	s := snapshotPolicyResourceSchema(t).Schema

	// Create first.
	plan := snapshotPolicyPlanWithName(t, "delete-snap-policy")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSnapshotPolicyType(), nil), Schema: s},
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
	_, err := r.client.GetSnapshotPolicy(context.Background(), "delete-snap-policy")
	if err == nil || !client.IsNotFound(err) {
		t.Errorf("expected policy to be deleted, got: %v", err)
	}
}

// TestSnapshotPolicyResource_Import verifies ImportState populates all attributes.
func TestSnapshotPolicyResource_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterSnapshotPolicyHandlers(ms.Mux)

	r := newTestSnapshotPolicyResource(t, ms)
	s := snapshotPolicyResourceSchema(t).Schema

	// Create first.
	plan := snapshotPolicyPlanWithName(t, "import-snap-policy")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSnapshotPolicyType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Import by name.
	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSnapshotPolicyType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "import-snap-policy"}, importResp)

	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState returned error: %s", importResp.Diagnostics)
	}

	var model snapshotPolicyModel
	if diags := importResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Name.ValueString() != "import-snap-policy" {
		t.Errorf("expected name=import-snap-policy after import, got %s", model.Name.ValueString())
	}
	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected ID to be populated after import")
	}
}

// ---- data source tests -------------------------------------------------------

// newTestSnapshotPolicyDataSource creates a snapshotPolicyDataSource wired to the given mock server.
func newTestSnapshotPolicyDataSource(t *testing.T, ms *testmock.MockServer) *snapshotPolicyDataSource {
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
	return &snapshotPolicyDataSource{client: c}
}

// snapshotPolicyDataSourceSchema returns the schema for the snapshot policy data source.
func snapshotPolicyDataSourceSchema(t *testing.T) datasource.SchemaResponse {
	t.Helper()
	d := &snapshotPolicyDataSource{}
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	return resp
}

// buildSnapshotPolicyDSType returns the tftypes.Object for the snapshot policy data source.
func buildSnapshotPolicyDSType() tftypes.Object {
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":             tftypes.String,
		"name":           tftypes.String,
		"enabled":        tftypes.Bool,
		"is_local":       tftypes.Bool,
		"policy_type":    tftypes.String,
		"retention_lock": tftypes.String,
	}}
}

// nullSnapshotPolicyDSConfig returns a base config map with all data source attributes null.
func nullSnapshotPolicyDSConfig() map[string]tftypes.Value {
	return map[string]tftypes.Value{
		"id":             tftypes.NewValue(tftypes.String, nil),
		"name":           tftypes.NewValue(tftypes.String, nil),
		"enabled":        tftypes.NewValue(tftypes.Bool, nil),
		"is_local":       tftypes.NewValue(tftypes.Bool, nil),
		"policy_type":    tftypes.NewValue(tftypes.String, nil),
		"retention_lock": tftypes.NewValue(tftypes.String, nil),
	}
}

// TestSnapshotPolicyDataSource verifies data source reads policy by name and returns all attributes.
func TestSnapshotPolicyDataSource(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterSnapshotPolicyHandlers(ms.Mux)

	// Create a policy via the resource client so the data source can find it.
	c, err := client.NewClient(context.Background(), client.Config{
		Endpoint:           ms.URL(),
		APIToken:           "test-token",
		InsecureSkipVerify: true,
		MaxRetries:         1,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	enabled := true
	_, err = c.PostSnapshotPolicy(context.Background(), "ds-test-snap-policy", client.SnapshotPolicyPost{
		Enabled: &enabled,
	})
	if err != nil {
		t.Fatalf("PostSnapshotPolicy: %v", err)
	}

	d := newTestSnapshotPolicyDataSource(t, ms)
	s := snapshotPolicyDataSourceSchema(t).Schema

	cfg := nullSnapshotPolicyDSConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "ds-test-snap-policy")

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSnapshotPolicyDSType(), nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(buildSnapshotPolicyDSType(), cfg), Schema: s},
	}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("DataSource Read returned error: %s", readResp.Diagnostics)
	}

	var model snapshotPolicyDataSourceModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Name.ValueString() != "ds-test-snap-policy" {
		t.Errorf("expected name=ds-test-snap-policy, got %s", model.Name.ValueString())
	}
	if !model.Enabled.ValueBool() {
		t.Error("expected enabled=true")
	}
	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected ID to be populated")
	}
}

// TestUnit_SnapshotPolicy_Lifecycle exercises the full Create->Read->Update->Read->Delete sequence.
func TestUnit_SnapshotPolicy_Lifecycle(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterSnapshotPolicyHandlers(ms.Mux)

	r := newTestSnapshotPolicyResource(t, ms)
	s := snapshotPolicyResourceSchema(t).Schema

	// Step 1: Create.
	createPlan := snapshotPolicyPlanWithNameAndEnabled(t, "lifecycle-snap-policy", true)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSnapshotPolicyType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}
	var createModel snapshotPolicyModel
	if diags := createResp.State.Get(context.Background(), &createModel); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}
	if createModel.Name.ValueString() != "lifecycle-snap-policy" {
		t.Errorf("Create: expected name=lifecycle-snap-policy, got %s", createModel.Name.ValueString())
	}

	// Step 2: Read post-create.
	readResp1 := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp1)
	if readResp1.Diagnostics.HasError() {
		t.Fatalf("Read post-create: %s", readResp1.Diagnostics)
	}
	var readModel1 snapshotPolicyModel
	if diags := readResp1.State.Get(context.Background(), &readModel1); diags.HasError() {
		t.Fatalf("Get read1 state: %s", diags)
	}
	if !readModel1.Enabled.ValueBool() {
		t.Error("Read1: expected enabled=true")
	}

	// Step 3: Update enabled=false.
	updatePlan := snapshotPolicyPlanWithNameAndEnabled(t, "lifecycle-snap-policy", false)
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSnapshotPolicyType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlan,
		State: readResp1.State,
	}, updateResp)
	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update: %s", updateResp.Diagnostics)
	}
	var updateModel snapshotPolicyModel
	if diags := updateResp.State.Get(context.Background(), &updateModel); diags.HasError() {
		t.Fatalf("Get update state: %s", diags)
	}
	if updateModel.Enabled.ValueBool() {
		t.Error("Update: expected enabled=false")
	}

	// Step 4: Read post-update.
	readResp2 := &resource.ReadResponse{State: updateResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: updateResp.State}, readResp2)
	if readResp2.Diagnostics.HasError() {
		t.Fatalf("Read post-update: %s", readResp2.Diagnostics)
	}
	var readModel2 snapshotPolicyModel
	if diags := readResp2.State.Get(context.Background(), &readModel2); diags.HasError() {
		t.Fatalf("Get read2 state: %s", diags)
	}
	if readModel2.Enabled.ValueBool() {
		t.Error("Read2: expected enabled=false")
	}

	// Step 5: Delete.
	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: readResp2.State}, deleteResp)
	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete: %s", deleteResp.Diagnostics)
	}
	_, err := r.client.GetSnapshotPolicy(context.Background(), "lifecycle-snap-policy")
	if err == nil || !client.IsNotFound(err) {
		t.Errorf("expected policy to be deleted, got: %v", err)
	}
}

// TestUnit_SnapshotPolicy_ImportIdempotency verifies ImportState->Read produces state matching original Create.
func TestUnit_SnapshotPolicy_ImportIdempotency(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterSnapshotPolicyHandlers(ms.Mux)

	r := newTestSnapshotPolicyResource(t, ms)
	s := snapshotPolicyResourceSchema(t).Schema

	// Create.
	createPlan := snapshotPolicyPlanWithNameAndEnabled(t, "idempotent-snap-policy", true)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSnapshotPolicyType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}
	var createModel snapshotPolicyModel
	if diags := createResp.State.Get(context.Background(), &createModel); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}

	// ImportState.
	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSnapshotPolicyType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "idempotent-snap-policy"}, importResp)
	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState: %s", importResp.Diagnostics)
	}

	// Read to populate full state.
	readResp := &resource.ReadResponse{State: importResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: importResp.State}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read post-import: %s", readResp.Diagnostics)
	}
	var importedModel snapshotPolicyModel
	if diags := readResp.State.Get(context.Background(), &importedModel); diags.HasError() {
		t.Fatalf("Get imported state: %s", diags)
	}

	// Verify 0-diff.
	if importedModel.Name.ValueString() != createModel.Name.ValueString() {
		t.Errorf("name mismatch: create=%s import=%s", createModel.Name.ValueString(), importedModel.Name.ValueString())
	}
	if importedModel.Enabled.ValueBool() != createModel.Enabled.ValueBool() {
		t.Errorf("enabled mismatch: create=%v import=%v", createModel.Enabled.ValueBool(), importedModel.Enabled.ValueBool())
	}
	if importedModel.ID.ValueString() != createModel.ID.ValueString() {
		t.Errorf("id mismatch: create=%s import=%s", createModel.ID.ValueString(), importedModel.ID.ValueString())
	}
}

// TestUnit_SnapshotPolicy_PlanModifiers verifies all RequiresReplace and UseStateForUnknown
// plan modifiers in the snapshot_policy resource schema.
func TestUnit_SnapshotPolicy_PlanModifiers(t *testing.T) {
	s := snapshotPolicyResourceSchema(t).Schema

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

	// is_local — UseStateForUnknown
	ilAttr, ok := s.Attributes["is_local"].(resschema.BoolAttribute)
	if !ok {
		t.Fatal("is_local attribute not found or wrong type")
	}
	if len(ilAttr.PlanModifiers) == 0 {
		t.Error("expected UseStateForUnknown plan modifier on is_local attribute")
	}

	// retention_lock — UseStateForUnknown
	rlAttr, ok := s.Attributes["retention_lock"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("retention_lock attribute not found or wrong type")
	}
	if len(rlAttr.PlanModifiers) == 0 {
		t.Error("expected UseStateForUnknown plan modifier on retention_lock attribute")
	}
}
