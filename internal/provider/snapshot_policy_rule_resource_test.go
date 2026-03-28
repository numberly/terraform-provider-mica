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

// newTestSnapshotRuleResource creates a snapshotPolicyRuleResource wired to the given mock server.
func newTestSnapshotRuleResource(t *testing.T, ms *testmock.MockServer) *snapshotPolicyRuleResource {
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
	return &snapshotPolicyRuleResource{client: c}
}

// snapshotRuleResourceSchema returns the parsed schema for the snapshot policy rule resource.
func snapshotRuleResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &snapshotPolicyRuleResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

// buildSnapshotRuleType returns the tftypes.Object for the snapshot policy rule resource.
func buildSnapshotRuleType() tftypes.Object {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":          tftypes.String,
		"policy_name": tftypes.String,
		"name":        tftypes.String,
		"at":          tftypes.Number,
		"every":       tftypes.Number,
		"keep_for":    tftypes.Number,
		"suffix":      tftypes.String,
		"client_name": tftypes.String,
		"timeouts":    timeoutsType,
	}}
}

// nullSnapshotRuleConfig returns a base config map with all attributes null.
func nullSnapshotRuleConfig() map[string]tftypes.Value {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return map[string]tftypes.Value{
		"id":          tftypes.NewValue(tftypes.String, nil),
		"policy_name": tftypes.NewValue(tftypes.String, nil),
		"name":        tftypes.NewValue(tftypes.String, nil),
		"at":          tftypes.NewValue(tftypes.Number, nil),
		"every":       tftypes.NewValue(tftypes.Number, nil),
		"keep_for":    tftypes.NewValue(tftypes.Number, nil),
		"suffix":      tftypes.NewValue(tftypes.String, nil),
		"client_name": tftypes.NewValue(tftypes.String, nil),
		"timeouts":    tftypes.NewValue(timeoutsType, nil),
	}
}

// snapshotRulePlan builds a tfsdk.Plan for a snapshot policy rule.
func snapshotRulePlan(t *testing.T, policyName string, every, keepFor int64) tfsdk.Plan {
	t.Helper()
	s := snapshotRuleResourceSchema(t).Schema
	cfg := nullSnapshotRuleConfig()
	cfg["policy_name"] = tftypes.NewValue(tftypes.String, policyName)
	if every != 0 {
		cfg["every"] = tftypes.NewValue(tftypes.Number, every)
	}
	if keepFor != 0 {
		cfg["keep_for"] = tftypes.NewValue(tftypes.Number, keepFor)
	}
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildSnapshotRuleType(), cfg),
		Schema: s,
	}
}

// createTestSnapshotPolicy is a helper that creates a snapshot policy via the client.
func createTestSnapshotPolicy(t *testing.T, c *client.FlashBladeClient, name string) {
	t.Helper()
	enabled := true
	_, err := c.PostSnapshotPolicy(context.Background(), name, client.SnapshotPolicyPost{Enabled: &enabled})
	if err != nil {
		t.Fatalf("PostSnapshotPolicy(%q): %v", name, err)
	}
}

// ---- tests ------------------------------------------------------------------

// TestSnapshotPolicyRuleResource_Create verifies Create adds a rule via parent PATCH add_rules.
func TestSnapshotPolicyRuleResource_Create(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterSnapshotPolicyHandlers(ms.Mux)

	r := newTestSnapshotRuleResource(t, ms)
	s := snapshotRuleResourceSchema(t).Schema

	createTestSnapshotPolicy(t, r.client, "snap-rule-create-policy")

	// every=86400000 (daily), keep_for=604800000 (7 days)
	plan := snapshotRulePlan(t, "snap-rule-create-policy", 86400000, 604800000)
	resp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSnapshotRuleType(), nil), Schema: s},
	}

	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Create returned error: %s", resp.Diagnostics)
	}

	var model snapshotPolicyRuleModel
	if diags := resp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected non-empty ID after Create")
	}
	if model.Name.IsNull() || model.Name.ValueString() == "" {
		t.Error("expected non-empty name after Create (server-assigned)")
	}
	if model.PolicyName.ValueString() != "snap-rule-create-policy" {
		t.Errorf("expected policy_name=snap-rule-create-policy, got %s", model.PolicyName.ValueString())
	}
	if model.Every.IsNull() || model.Every.ValueInt64() != 86400000 {
		t.Errorf("expected every=86400000, got %v", model.Every)
	}
	if model.KeepFor.IsNull() || model.KeepFor.ValueInt64() != 604800000 {
		t.Errorf("expected keep_for=604800000, got %v", model.KeepFor)
	}
}

// TestSnapshotPolicyRuleResource_Update verifies Update replaces the rule via remove+add PATCH.
func TestSnapshotPolicyRuleResource_Update(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterSnapshotPolicyHandlers(ms.Mux)

	r := newTestSnapshotRuleResource(t, ms)
	s := snapshotRuleResourceSchema(t).Schema

	createTestSnapshotPolicy(t, r.client, "snap-rule-update-policy")

	// Create rule first.
	createPlan := snapshotRulePlan(t, "snap-rule-update-policy", 86400000, 604800000)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSnapshotRuleType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Update keep_for to 1209600000 (14 days).
	updatePlan := snapshotRulePlan(t, "snap-rule-update-policy", 86400000, 1209600000)
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSnapshotRuleType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlan,
		State: createResp.State,
	}, updateResp)

	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update returned error: %s", updateResp.Diagnostics)
	}

	var model snapshotPolicyRuleModel
	if diags := updateResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}
	if model.KeepFor.IsNull() || model.KeepFor.ValueInt64() != 1209600000 {
		t.Errorf("expected keep_for=1209600000 after update, got %v", model.KeepFor)
	}
}

// TestSnapshotPolicyRuleResource_Delete verifies Delete removes the rule via remove_rules PATCH.
func TestSnapshotPolicyRuleResource_Delete(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterSnapshotPolicyHandlers(ms.Mux)

	r := newTestSnapshotRuleResource(t, ms)
	s := snapshotRuleResourceSchema(t).Schema

	createTestSnapshotPolicy(t, r.client, "snap-rule-delete-policy")

	// Create rule first.
	createPlan := snapshotRulePlan(t, "snap-rule-delete-policy", 86400000, 604800000)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSnapshotRuleType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: createResp.State}, deleteResp)

	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete returned error: %s", deleteResp.Diagnostics)
	}

	// Verify policy's rules array is empty.
	policy, err := r.client.GetSnapshotPolicy(context.Background(), "snap-rule-delete-policy")
	if err != nil {
		t.Fatalf("GetSnapshotPolicy after delete: %v", err)
	}
	if len(policy.Rules) != 0 {
		t.Errorf("expected policy.Rules to be empty after delete, got %d rules", len(policy.Rules))
	}
}

// TestSnapshotPolicyRuleResource_Import verifies ImportState with composite ID "policy_name/0".
func TestSnapshotPolicyRuleResource_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterSnapshotPolicyHandlers(ms.Mux)

	r := newTestSnapshotRuleResource(t, ms)
	s := snapshotRuleResourceSchema(t).Schema

	createTestSnapshotPolicy(t, r.client, "snap-rule-import-policy")

	// Create rule.
	createPlan := snapshotRulePlan(t, "snap-rule-import-policy", 86400000, 604800000)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSnapshotRuleType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Import using "policy_name/0" composite ID.
	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSnapshotRuleType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "snap-rule-import-policy/0"}, importResp)

	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState returned error: %s", importResp.Diagnostics)
	}

	var model snapshotPolicyRuleModel
	if diags := importResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.PolicyName.ValueString() != "snap-rule-import-policy" {
		t.Errorf("expected policy_name=snap-rule-import-policy after import, got %s", model.PolicyName.ValueString())
	}
	if model.Name.IsNull() || model.Name.ValueString() == "" {
		t.Error("expected server-assigned name to be populated after import")
	}
	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected ID to be populated after import")
	}
	if model.Every.IsNull() || model.Every.ValueInt64() != 86400000 {
		t.Errorf("expected every=86400000 after import, got %v", model.Every)
	}
}

// TestUnit_SnapshotPolicyRule_Lifecycle exercises the full Create->Read->Update->Read->Delete sequence.
func TestUnit_SnapshotPolicyRule_Lifecycle(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterSnapshotPolicyHandlers(ms.Mux)

	r := newTestSnapshotRuleResource(t, ms)
	s := snapshotRuleResourceSchema(t).Schema

	createTestSnapshotPolicy(t, r.client, "lifecycle-snap-rule-policy")

	// Step 1: Create (daily snapshots, 7 day retention).
	createPlan := snapshotRulePlan(t, "lifecycle-snap-rule-policy", 86400000, 604800000)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSnapshotRuleType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}
	var createModel snapshotPolicyRuleModel
	if diags := createResp.State.Get(context.Background(), &createModel); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}
	if createModel.Every.ValueInt64() != 86400000 {
		t.Errorf("Create: expected every=86400000, got %d", createModel.Every.ValueInt64())
	}

	// Step 2: Read post-create.
	readResp1 := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp1)
	if readResp1.Diagnostics.HasError() {
		t.Fatalf("Read post-create: %s", readResp1.Diagnostics)
	}
	var readModel1 snapshotPolicyRuleModel
	if diags := readResp1.State.Get(context.Background(), &readModel1); diags.HasError() {
		t.Fatalf("Get read1 state: %s", diags)
	}
	if readModel1.KeepFor.ValueInt64() != 604800000 {
		t.Errorf("Read1: expected keep_for=604800000, got %d", readModel1.KeepFor.ValueInt64())
	}

	// Step 3: Update keep_for to 14 days.
	updatePlan := snapshotRulePlan(t, "lifecycle-snap-rule-policy", 86400000, 1209600000)
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSnapshotRuleType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlan,
		State: readResp1.State,
	}, updateResp)
	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update: %s", updateResp.Diagnostics)
	}
	var updateModel snapshotPolicyRuleModel
	if diags := updateResp.State.Get(context.Background(), &updateModel); diags.HasError() {
		t.Fatalf("Get update state: %s", diags)
	}
	if updateModel.KeepFor.ValueInt64() != 1209600000 {
		t.Errorf("Update: expected keep_for=1209600000, got %d", updateModel.KeepFor.ValueInt64())
	}

	// Step 4: Read post-update.
	readResp2 := &resource.ReadResponse{State: updateResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: updateResp.State}, readResp2)
	if readResp2.Diagnostics.HasError() {
		t.Fatalf("Read post-update: %s", readResp2.Diagnostics)
	}
	var readModel2 snapshotPolicyRuleModel
	if diags := readResp2.State.Get(context.Background(), &readModel2); diags.HasError() {
		t.Fatalf("Get read2 state: %s", diags)
	}
	if readModel2.KeepFor.ValueInt64() != 1209600000 {
		t.Errorf("Read2: expected keep_for=1209600000, got %d", readModel2.KeepFor.ValueInt64())
	}

	// Step 5: Delete.
	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: readResp2.State}, deleteResp)
	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete: %s", deleteResp.Diagnostics)
	}
}

// TestUnit_SnapshotPolicyRule_ImportIdempotency verifies ImportState->Read produces state matching original Create.
func TestUnit_SnapshotPolicyRule_ImportIdempotency(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterSnapshotPolicyHandlers(ms.Mux)

	r := newTestSnapshotRuleResource(t, ms)
	s := snapshotRuleResourceSchema(t).Schema

	createTestSnapshotPolicy(t, r.client, "idempotent-snap-rule-policy")

	// Create.
	createPlan := snapshotRulePlan(t, "idempotent-snap-rule-policy", 86400000, 604800000)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSnapshotRuleType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}
	var createModel snapshotPolicyRuleModel
	if diags := createResp.State.Get(context.Background(), &createModel); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}

	// ImportState using "policy_name/0" composite ID (rule index 0).
	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSnapshotRuleType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "idempotent-snap-rule-policy/0"}, importResp)
	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState: %s", importResp.Diagnostics)
	}

	// Read to populate full state.
	readResp := &resource.ReadResponse{State: importResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: importResp.State}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read post-import: %s", readResp.Diagnostics)
	}
	var importedModel snapshotPolicyRuleModel
	if diags := readResp.State.Get(context.Background(), &importedModel); diags.HasError() {
		t.Fatalf("Get imported state: %s", diags)
	}

	// Verify 0-diff.
	if importedModel.PolicyName.ValueString() != createModel.PolicyName.ValueString() {
		t.Errorf("policy_name mismatch: create=%s import=%s", createModel.PolicyName.ValueString(), importedModel.PolicyName.ValueString())
	}
	if importedModel.Every.ValueInt64() != createModel.Every.ValueInt64() {
		t.Errorf("every mismatch: create=%d import=%d", createModel.Every.ValueInt64(), importedModel.Every.ValueInt64())
	}
	if importedModel.KeepFor.ValueInt64() != createModel.KeepFor.ValueInt64() {
		t.Errorf("keep_for mismatch: create=%d import=%d", createModel.KeepFor.ValueInt64(), importedModel.KeepFor.ValueInt64())
	}
}

// TestUnit_SnapshotRule_PlanModifiers verifies all RequiresReplace and UseStateForUnknown
// plan modifiers in the snapshot_policy_rule resource schema.
func TestUnit_SnapshotRule_PlanModifiers(t *testing.T) {
	s := snapshotRuleResourceSchema(t).Schema

	// id — UseStateForUnknown
	idAttr, ok := s.Attributes["id"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("id attribute not found or wrong type")
	}
	if len(idAttr.PlanModifiers) == 0 {
		t.Error("expected UseStateForUnknown plan modifier on id attribute")
	}

	// policy_name — RequiresReplace
	pnAttr, ok := s.Attributes["policy_name"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("policy_name attribute not found or wrong type")
	}
	if len(pnAttr.PlanModifiers) == 0 {
		t.Error("expected RequiresReplace plan modifier on policy_name attribute")
	}

	// name — UseStateForUnknown
	nameAttr, ok := s.Attributes["name"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("name attribute not found or wrong type")
	}
	if len(nameAttr.PlanModifiers) == 0 {
		t.Error("expected UseStateForUnknown plan modifier on name attribute")
	}
}
