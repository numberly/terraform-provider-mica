package provider

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	resschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/numberly/opentofu-provider-flashblade/internal/client"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock/handlers"
)

// ---- helpers ----------------------------------------------------------------

// newTestS3RuleResource creates an s3ExportPolicyRuleResource wired to the given mock server.
func newTestS3RuleResource(t *testing.T, ms *testmock.MockServer) *s3ExportPolicyRuleResource {
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
	return &s3ExportPolicyRuleResource{client: c}
}

// s3RuleResourceSchema returns the parsed schema for the S3 export policy rule resource.
func s3RuleResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &s3ExportPolicyRuleResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

// buildS3RuleType returns the tftypes.Object for the S3 export policy rule resource.
func buildS3RuleType() tftypes.Object {
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
		"index":       tftypes.Number,
		"effect":      tftypes.String,
		"actions":     tftypes.List{ElementType: tftypes.String},
		"resources":   tftypes.List{ElementType: tftypes.String},
		"timeouts":    timeoutsType,
	}}
}

// nullS3RuleConfig returns a base config map with all attributes null.
func nullS3RuleConfig() map[string]tftypes.Value {
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
		"index":       tftypes.NewValue(tftypes.Number, nil),
		"effect":      tftypes.NewValue(tftypes.String, nil),
		"actions":     tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"resources":   tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"timeouts":    tftypes.NewValue(timeoutsType, nil),
	}
}

// s3RulePlan builds a tfsdk.Plan with the given fields.
func s3RulePlan(t *testing.T, policyName, effect string, actions, resources []string) tfsdk.Plan {
	t.Helper()
	s := s3RuleResourceSchema(t).Schema
	cfg := nullS3RuleConfig()
	cfg["policy_name"] = tftypes.NewValue(tftypes.String, policyName)
	cfg["effect"] = tftypes.NewValue(tftypes.String, effect)

	actionValues := make([]tftypes.Value, len(actions))
	for i, a := range actions {
		actionValues[i] = tftypes.NewValue(tftypes.String, a)
	}
	cfg["actions"] = tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, actionValues)

	resourceValues := make([]tftypes.Value, len(resources))
	for i, r := range resources {
		resourceValues[i] = tftypes.NewValue(tftypes.String, r)
	}
	cfg["resources"] = tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, resourceValues)

	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildS3RuleType(), cfg),
		Schema: s,
	}
}

// ---- tests ------------------------------------------------------------------

// TestUnit_S3ExportPolicyRule_CreateReadUpdateDelete exercises the full lifecycle.
func TestUnit_S3ExportPolicyRule_CreateReadUpdateDelete(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterS3ExportPolicyHandlers(ms.Mux)

	r := newTestS3RuleResource(t, ms)
	s := s3RuleResourceSchema(t).Schema

	// Create the parent policy first.
	createTestS3Policy(t, r.client, "s3-rule-lifecycle-policy")

	// Step 1: Create rule with effect=allow, actions=["s3:GetObject"], resources=["*"].
	createPlan := s3RulePlan(t, "s3-rule-lifecycle-policy", "allow", []string{"s3:GetObject"}, []string{"*"})
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildS3RuleType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}
	var createModel s3ExportPolicyRuleModel
	if diags := createResp.State.Get(context.Background(), &createModel); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}
	if createModel.Effect.ValueString() != "allow" {
		t.Errorf("Create: expected effect=allow, got %s", createModel.Effect.ValueString())
	}
	if createModel.ID.IsNull() || createModel.ID.ValueString() == "" {
		t.Error("Create: expected non-empty ID")
	}
	if createModel.Name.IsNull() || createModel.Name.ValueString() == "" {
		t.Error("Create: expected non-empty server-assigned name")
	}

	// Step 2: Read post-create.
	readResp1 := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp1)
	if readResp1.Diagnostics.HasError() {
		t.Fatalf("Read post-create: %s", readResp1.Diagnostics)
	}
	var readModel1 s3ExportPolicyRuleModel
	if diags := readResp1.State.Get(context.Background(), &readModel1); diags.HasError() {
		t.Fatalf("Get read1 state: %s", diags)
	}
	if readModel1.Effect.ValueString() != "allow" {
		t.Errorf("Read1: expected effect=allow, got %s", readModel1.Effect.ValueString())
	}

	// Step 3: Update effect to deny (in-place, no replace).
	updatePlan := s3RulePlan(t, "s3-rule-lifecycle-policy", "deny", []string{"s3:GetObject"}, []string{"*"})
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildS3RuleType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlan,
		State: readResp1.State,
	}, updateResp)
	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update: %s", updateResp.Diagnostics)
	}
	var updateModel s3ExportPolicyRuleModel
	if diags := updateResp.State.Get(context.Background(), &updateModel); diags.HasError() {
		t.Fatalf("Get update state: %s", diags)
	}
	if updateModel.Effect.ValueString() != "deny" {
		t.Errorf("Update: expected effect=deny, got %s", updateModel.Effect.ValueString())
	}
	// Verify the rule name is unchanged (in-place update, not recreated).
	if updateModel.Name.ValueString() != createModel.Name.ValueString() {
		t.Errorf("Update: rule name changed from %s to %s (expected in-place update)", createModel.Name.ValueString(), updateModel.Name.ValueString())
	}

	// Step 4: Read post-update.
	readResp2 := &resource.ReadResponse{State: updateResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: updateResp.State}, readResp2)
	if readResp2.Diagnostics.HasError() {
		t.Fatalf("Read post-update: %s", readResp2.Diagnostics)
	}
	var readModel2 s3ExportPolicyRuleModel
	if diags := readResp2.State.Get(context.Background(), &readModel2); diags.HasError() {
		t.Fatalf("Get read2 state: %s", diags)
	}
	if readModel2.Effect.ValueString() != "deny" {
		t.Errorf("Read2: expected effect=deny, got %s", readModel2.Effect.ValueString())
	}

	// Step 5: Delete.
	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: readResp2.State}, deleteResp)
	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete: %s", deleteResp.Diagnostics)
	}
	_, err := r.client.GetS3ExportPolicyRuleByName(context.Background(), "s3-rule-lifecycle-policy", createModel.Name.ValueString())
	if err == nil || !client.IsNotFound(err) {
		t.Errorf("expected rule to be deleted, got: %v", err)
	}
}

// TestUnit_S3ExportPolicyRule_Import verifies ImportState with composite ID "policy_name/rule_index".
func TestUnit_S3ExportPolicyRule_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterS3ExportPolicyHandlers(ms.Mux)

	r := newTestS3RuleResource(t, ms)
	s := s3RuleResourceSchema(t).Schema

	createTestS3Policy(t, r.client, "s3-rule-import-policy")

	// Create rule.
	createPlan := s3RulePlan(t, "s3-rule-import-policy", "allow", []string{"s3:GetObject"}, []string{"*"})
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildS3RuleType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	var createModel s3ExportPolicyRuleModel
	if diags := createResp.State.Get(context.Background(), &createModel); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}
	index := strconv.FormatInt(createModel.Index.ValueInt64(), 10)

	// Import using "policy_name/index" composite ID.
	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildS3RuleType(), nil), Schema: s},
	}
	importID := "s3-rule-import-policy/" + index
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: importID}, importResp)
	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState: %s", importResp.Diagnostics)
	}

	var model s3ExportPolicyRuleModel
	if diags := importResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.PolicyName.ValueString() != "s3-rule-import-policy" {
		t.Errorf("expected policy_name=s3-rule-import-policy, got %s", model.PolicyName.ValueString())
	}
	if model.Name.IsNull() || model.Name.ValueString() == "" {
		t.Error("expected server-assigned name to be populated after import")
	}
	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected ID to be populated after import")
	}
	if model.Effect.ValueString() != "allow" {
		t.Errorf("expected effect=allow after import, got %s", model.Effect.ValueString())
	}
}

// TestUnit_S3ExportPolicyRule_IndependentDelete verifies deleting one rule does not affect others.
func TestUnit_S3ExportPolicyRule_IndependentDelete(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterS3ExportPolicyHandlers(ms.Mux)

	r := newTestS3RuleResource(t, ms)
	s := s3RuleResourceSchema(t).Schema

	createTestS3Policy(t, r.client, "s3-rule-indep-policy")

	// Create rule 1.
	plan1 := s3RulePlan(t, "s3-rule-indep-policy", "allow", []string{"s3:GetObject"}, []string{"*"})
	resp1 := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildS3RuleType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan1}, resp1)
	if resp1.Diagnostics.HasError() {
		t.Fatalf("Create rule 1: %s", resp1.Diagnostics)
	}
	var model1 s3ExportPolicyRuleModel
	if diags := resp1.State.Get(context.Background(), &model1); diags.HasError() {
		t.Fatalf("Get rule 1 state: %s", diags)
	}

	// Create rule 2.
	plan2 := s3RulePlan(t, "s3-rule-indep-policy", "deny", []string{"s3:PutObject"}, []string{"bucket/*"})
	resp2 := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildS3RuleType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan2}, resp2)
	if resp2.Diagnostics.HasError() {
		t.Fatalf("Create rule 2: %s", resp2.Diagnostics)
	}
	var model2 s3ExportPolicyRuleModel
	if diags := resp2.State.Get(context.Background(), &model2); diags.HasError() {
		t.Fatalf("Get rule 2 state: %s", diags)
	}

	// Delete rule 1.
	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: resp1.State}, deleteResp)
	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete rule 1: %s", deleteResp.Diagnostics)
	}

	// Verify rule 1 is gone.
	_, err := r.client.GetS3ExportPolicyRuleByName(context.Background(), "s3-rule-indep-policy", model1.Name.ValueString())
	if err == nil || !client.IsNotFound(err) {
		t.Errorf("expected rule 1 to be deleted, got: %v", err)
	}

	// Verify rule 2 still exists.
	rule2, err := r.client.GetS3ExportPolicyRuleByName(context.Background(), "s3-rule-indep-policy", model2.Name.ValueString())
	if err != nil {
		t.Fatalf("expected rule 2 to still exist, got error: %v", err)
	}
	if rule2.Effect != "deny" {
		t.Errorf("rule 2 effect changed: expected deny, got %s", rule2.Effect)
	}
}

// TestUnit_S3ExportPolicyRule_PlanModifiers verifies RequiresReplace and UseStateForUnknown modifiers.
func TestUnit_S3ExportPolicyRule_PlanModifiers(t *testing.T) {
	s := s3RuleResourceSchema(t).Schema

	// id -- UseStateForUnknown
	idAttr, ok := s.Attributes["id"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("id attribute not found or wrong type")
	}
	if len(idAttr.PlanModifiers) == 0 {
		t.Error("expected UseStateForUnknown plan modifier on id attribute")
	}

	// policy_name -- RequiresReplace
	pnAttr, ok := s.Attributes["policy_name"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("policy_name attribute not found or wrong type")
	}
	if len(pnAttr.PlanModifiers) == 0 {
		t.Error("expected RequiresReplace plan modifier on policy_name attribute")
	}

	// name -- UseStateForUnknown (computed, server-assigned)
	nameAttr, ok := s.Attributes["name"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("name attribute not found or wrong type")
	}
	if len(nameAttr.PlanModifiers) == 0 {
		t.Error("expected UseStateForUnknown plan modifier on name attribute")
	}

	// effect -- no RequiresReplace (patchable in-place)
	effectAttr, ok := s.Attributes["effect"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("effect attribute not found or wrong type")
	}
	if len(effectAttr.PlanModifiers) > 0 {
		t.Error("expected no plan modifiers on effect attribute (it should be patchable in-place)")
	}
}

// TestUnit_S3ExportPolicyRule_Idempotent verifies that Read after Create shows no attribute drift.
func TestUnit_S3ExportPolicyRule_Idempotent(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterS3ExportPolicyHandlers(ms.Mux)

	r := newTestS3RuleResource(t, ms)
	s := s3RuleResourceSchema(t).Schema

	createTestS3Policy(t, r.client, "idempotent-rule-policy")

	plan := s3RulePlan(t, "idempotent-rule-policy", "allow", []string{"s3:GetObject"}, []string{"*"})
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildS3RuleType(), nil), Schema: s},
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

	var beforeModel, afterModel s3ExportPolicyRuleModel
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
	if beforeModel.PolicyName.ValueString() != afterModel.PolicyName.ValueString() {
		t.Errorf("PolicyName changed after Read: %s -> %s", beforeModel.PolicyName.ValueString(), afterModel.PolicyName.ValueString())
	}
	if beforeModel.Effect.ValueString() != afterModel.Effect.ValueString() {
		t.Errorf("Effect changed after Read: %s -> %s", beforeModel.Effect.ValueString(), afterModel.Effect.ValueString())
	}
	if beforeModel.Index.ValueInt64() != afterModel.Index.ValueInt64() {
		t.Errorf("Index changed after Read: %d -> %d", beforeModel.Index.ValueInt64(), afterModel.Index.ValueInt64())
	}
}
