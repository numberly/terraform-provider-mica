package provider

import (
	"context"
	"testing"

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

// newTestOAPRuleResource creates an objectStoreAccessPolicyRuleResource wired to the given mock server.
func newTestOAPRuleResource(t *testing.T, ms *testmock.MockServer) *objectStoreAccessPolicyRuleResource {
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
	return &objectStoreAccessPolicyRuleResource{client: c}
}

// oapRuleResourceSchema returns the parsed schema for the OAP rule resource.
func oapRuleResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &objectStoreAccessPolicyRuleResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

// buildObjectStoreAccessPolicyRuleType returns the tftypes.Object for the OAP rule resource.
func buildObjectStoreAccessPolicyRuleType() tftypes.Object {
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
		"effect":      tftypes.String,
		"actions":     tftypes.List{ElementType: tftypes.String},
		"resources":   tftypes.List{ElementType: tftypes.String},
		"conditions":  tftypes.String,
		"timeouts":    timeoutsType,
	}}
}

// nullOAPRuleConfig returns a base config map with all attributes null.
func nullOAPRuleConfig() map[string]tftypes.Value {
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
		"effect":      tftypes.NewValue(tftypes.String, nil),
		"actions":     tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"resources":   tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"conditions":  tftypes.NewValue(tftypes.String, nil),
		"timeouts":    tftypes.NewValue(timeoutsType, nil),
	}
}

// oapRulePlan returns a tfsdk.Plan for an OAP rule.
func oapRulePlan(t *testing.T, policyName, ruleName, effect string, actions, resources []string) tfsdk.Plan {
	t.Helper()
	s := oapRuleResourceSchema(t).Schema
	cfg := nullOAPRuleConfig()
	cfg["policy_name"] = tftypes.NewValue(tftypes.String, policyName)
	cfg["name"] = tftypes.NewValue(tftypes.String, ruleName)
	cfg["effect"] = tftypes.NewValue(tftypes.String, effect)

	actionVals := make([]tftypes.Value, len(actions))
	for i, a := range actions {
		actionVals[i] = tftypes.NewValue(tftypes.String, a)
	}
	cfg["actions"] = tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, actionVals)

	resourceVals := make([]tftypes.Value, len(resources))
	for i, r := range resources {
		resourceVals[i] = tftypes.NewValue(tftypes.String, r)
	}
	cfg["resources"] = tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, resourceVals)

	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildObjectStoreAccessPolicyRuleType(), cfg),
		Schema: s,
	}
}

// oapRulePlanWithConditions returns a tfsdk.Plan with conditions set.
func oapRulePlanWithConditions(t *testing.T, policyName, ruleName, effect string, actions, resources []string, conditions string) tfsdk.Plan {
	t.Helper()
	s := oapRuleResourceSchema(t).Schema
	cfg := nullOAPRuleConfig()
	cfg["policy_name"] = tftypes.NewValue(tftypes.String, policyName)
	cfg["name"] = tftypes.NewValue(tftypes.String, ruleName)
	cfg["effect"] = tftypes.NewValue(tftypes.String, effect)
	cfg["conditions"] = tftypes.NewValue(tftypes.String, conditions)

	actionVals := make([]tftypes.Value, len(actions))
	for i, a := range actions {
		actionVals[i] = tftypes.NewValue(tftypes.String, a)
	}
	cfg["actions"] = tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, actionVals)

	resourceVals := make([]tftypes.Value, len(resources))
	for i, r := range resources {
		resourceVals[i] = tftypes.NewValue(tftypes.String, r)
	}
	cfg["resources"] = tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, resourceVals)

	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildObjectStoreAccessPolicyRuleType(), cfg),
		Schema: s,
	}
}

// createTestOAPPolicy creates a policy in the mock server.
func createTestOAPPolicy(t *testing.T, ms *testmock.MockServer, policyName string) {
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
	_, err = c.PostObjectStoreAccessPolicy(context.Background(), policyName, client.ObjectStoreAccessPolicyPost{})
	if err != nil {
		t.Fatalf("PostObjectStoreAccessPolicy(%s): %v", policyName, err)
	}
}

// ---- tests ------------------------------------------------------------------

// TestObjectStoreAccessPolicyRuleResource_Create verifies Create populates ID, effect, actions, resources.
func TestUnit_ObjectStoreAccessPolicyRuleResource_Create(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterObjectStoreAccessPolicyHandlers(ms.Mux)

	// Create the parent policy first.
	createTestOAPPolicy(t, ms, "rule-test-policy")

	r := newTestOAPRuleResource(t, ms)
	s := oapRuleResourceSchema(t).Schema

	plan := oapRulePlan(t, "rule-test-policy", "rule-1", "allow",
		[]string{"s3:GetObject", "s3:PutObject"},
		[]string{"arn:aws:s3:::my-bucket/*"},
	)
	resp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildObjectStoreAccessPolicyRuleType(), nil), Schema: s},
	}

	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Create returned error: %s", resp.Diagnostics)
	}

	var model objectStoreAccessPolicyRuleModel
	if diags := resp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected non-empty ID after Create")
	}
	if model.ID.ValueString() != "rule-test-policy/rule-1" {
		t.Errorf("expected ID=rule-test-policy/rule-1, got %s", model.ID.ValueString())
	}
	if model.Name.ValueString() != "rule-1" {
		t.Errorf("expected name=rule-1, got %s", model.Name.ValueString())
	}
	if model.Effect.ValueString() != "allow" {
		t.Errorf("expected effect=allow, got %s", model.Effect.ValueString())
	}

	var actions []string
	model.Actions.ElementsAs(context.Background(), &actions, false)
	if len(actions) != 2 {
		t.Errorf("expected 2 actions, got %d", len(actions))
	}
	if !model.Conditions.IsNull() {
		t.Error("expected conditions to be null since none were provided")
	}
}

// TestObjectStoreAccessPolicyRuleResource_Update verifies PATCH updates actions list.
func TestUnit_ObjectStoreAccessPolicyRuleResource_Update(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterObjectStoreAccessPolicyHandlers(ms.Mux)

	createTestOAPPolicy(t, ms, "rule-update-policy")

	r := newTestOAPRuleResource(t, ms)
	s := oapRuleResourceSchema(t).Schema

	// Create first.
	createPlan := oapRulePlan(t, "rule-update-policy", "update-rule", "allow",
		[]string{"s3:GetObject"},
		[]string{"arn:aws:s3:::test-bucket/*"},
	)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildObjectStoreAccessPolicyRuleType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Update actions list.
	updatePlan := oapRulePlan(t, "rule-update-policy", "update-rule", "allow",
		[]string{"s3:GetObject", "s3:PutObject", "s3:DeleteObject"},
		[]string{"arn:aws:s3:::test-bucket/*"},
	)
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildObjectStoreAccessPolicyRuleType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlan,
		State: createResp.State,
	}, updateResp)

	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update returned error: %s", updateResp.Diagnostics)
	}

	var model objectStoreAccessPolicyRuleModel
	if diags := updateResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	var actions []string
	model.Actions.ElementsAs(context.Background(), &actions, false)
	if len(actions) != 3 {
		t.Errorf("expected 3 actions after update, got %d: %v", len(actions), actions)
	}
}

// TestObjectStoreAccessPolicyRuleResource_Delete verifies DELETE removes the rule.
func TestUnit_ObjectStoreAccessPolicyRuleResource_Delete(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterObjectStoreAccessPolicyHandlers(ms.Mux)

	createTestOAPPolicy(t, ms, "rule-delete-policy")

	r := newTestOAPRuleResource(t, ms)
	s := oapRuleResourceSchema(t).Schema

	// Create first.
	plan := oapRulePlan(t, "rule-delete-policy", "delete-rule", "deny",
		[]string{"s3:*"},
		[]string{"arn:aws:s3:::protected/*"},
	)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildObjectStoreAccessPolicyRuleType(), nil), Schema: s},
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

	// Verify rule is gone.
	c, err := client.NewClient(context.Background(), client.Config{
		Endpoint:           ms.URL(),
		APIToken:           "test-token",
		InsecureSkipVerify: true,
		MaxRetries:         1,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	_, err = c.GetObjectStoreAccessPolicyRuleByName(context.Background(), "rule-delete-policy", "delete-rule")
	if err == nil || !client.IsNotFound(err) {
		t.Errorf("expected rule to be deleted, got: %v", err)
	}
}

// TestObjectStoreAccessPolicyRuleResource_Import verifies ImportState using composite "policy_name/rule_name".
func TestUnit_ObjectStoreAccessPolicyRuleResource_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterObjectStoreAccessPolicyHandlers(ms.Mux)

	createTestOAPPolicy(t, ms, "rule-import-policy")

	r := newTestOAPRuleResource(t, ms)
	s := oapRuleResourceSchema(t).Schema

	// Create first.
	plan := oapRulePlan(t, "rule-import-policy", "import-rule", "allow",
		[]string{"s3:GetObject"},
		[]string{"arn:aws:s3:::shared/*"},
	)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildObjectStoreAccessPolicyRuleType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Import by "policy_name/rule_name".
	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildObjectStoreAccessPolicyRuleType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "rule-import-policy/import-rule"}, importResp)

	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState returned error: %s", importResp.Diagnostics)
	}

	var model objectStoreAccessPolicyRuleModel
	if diags := importResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Name.ValueString() != "import-rule" {
		t.Errorf("expected name=import-rule after import, got %s", model.Name.ValueString())
	}
	if model.PolicyName.ValueString() != "rule-import-policy" {
		t.Errorf("expected policy_name=rule-import-policy after import, got %s", model.PolicyName.ValueString())
	}
	if model.Effect.ValueString() != "allow" {
		t.Errorf("expected effect=allow after import, got %s", model.Effect.ValueString())
	}
	if model.ID.ValueString() != "rule-import-policy/import-rule" {
		t.Errorf("expected id=rule-import-policy/import-rule, got %s", model.ID.ValueString())
	}
}

// TestObjectStoreAccessPolicyRuleResource_ConditionsRoundTrip verifies conditions JSON round-trips correctly.
func TestUnit_ObjectStoreAccessPolicyRuleResource_ConditionsRoundTrip(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterObjectStoreAccessPolicyHandlers(ms.Mux)

	createTestOAPPolicy(t, ms, "conditions-policy")

	r := newTestOAPRuleResource(t, ms)
	s := oapRuleResourceSchema(t).Schema

	conditions := `{"StringEquals":{"s3:prefix":["photos/","videos/"]}}`

	plan := oapRulePlanWithConditions(t, "conditions-policy", "conditions-rule", "allow",
		[]string{"s3:GetObject"},
		[]string{"arn:aws:s3:::media-bucket/*"},
		conditions,
	)
	resp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildObjectStoreAccessPolicyRuleType(), nil), Schema: s},
	}

	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Create returned error: %s", resp.Diagnostics)
	}

	var model objectStoreAccessPolicyRuleModel
	if diags := resp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Conditions.IsNull() {
		t.Error("expected conditions to be populated, got null")
	} else {
		gotConditions := model.Conditions.ValueString()
		if gotConditions != conditions {
			t.Errorf("expected conditions=%s, got %s", conditions, gotConditions)
		}
	}
}

// TestUnit_OAPRule_Lifecycle exercises the full Create->Read->Update->Read->Delete sequence.
func TestUnit_Unit_OAPRule_Lifecycle(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterObjectStoreAccessPolicyHandlers(ms.Mux)

	createTestOAPPolicy(t, ms, "lifecycle-oap-rule-policy")

	r := newTestOAPRuleResource(t, ms)
	s := oapRuleResourceSchema(t).Schema

	// Step 1: Create.
	createPlan := oapRulePlan(t, "lifecycle-oap-rule-policy", "lifecycle-rule", "allow",
		[]string{"s3:GetObject"},
		[]string{"arn:aws:s3:::test-bucket/*"},
	)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildObjectStoreAccessPolicyRuleType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}
	var createModel objectStoreAccessPolicyRuleModel
	if diags := createResp.State.Get(context.Background(), &createModel); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}
	if createModel.Name.ValueString() != "lifecycle-rule" {
		t.Errorf("Create: expected name=lifecycle-rule, got %s", createModel.Name.ValueString())
	}

	// Step 2: Read post-create.
	readResp1 := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp1)
	if readResp1.Diagnostics.HasError() {
		t.Fatalf("Read post-create: %s", readResp1.Diagnostics)
	}
	var readModel1 objectStoreAccessPolicyRuleModel
	if diags := readResp1.State.Get(context.Background(), &readModel1); diags.HasError() {
		t.Fatalf("Get read1 state: %s", diags)
	}
	if readModel1.Effect.ValueString() != "allow" {
		t.Errorf("Read1: expected effect=allow, got %s", readModel1.Effect.ValueString())
	}

	// Step 3: Update actions list (effect is RequiresReplace, use resources update).
	updatePlan := oapRulePlan(t, "lifecycle-oap-rule-policy", "lifecycle-rule", "allow",
		[]string{"s3:GetObject", "s3:PutObject"},
		[]string{"arn:aws:s3:::test-bucket/*"},
	)
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildObjectStoreAccessPolicyRuleType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlan,
		State: readResp1.State,
	}, updateResp)
	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update: %s", updateResp.Diagnostics)
	}
	var updateModel objectStoreAccessPolicyRuleModel
	if diags := updateResp.State.Get(context.Background(), &updateModel); diags.HasError() {
		t.Fatalf("Get update state: %s", diags)
	}
	var updatedActions []string
	updateModel.Actions.ElementsAs(context.Background(), &updatedActions, false)
	if len(updatedActions) != 2 {
		t.Errorf("Update: expected 2 actions, got %d", len(updatedActions))
	}

	// Step 4: Read post-update.
	readResp2 := &resource.ReadResponse{State: updateResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: updateResp.State}, readResp2)
	if readResp2.Diagnostics.HasError() {
		t.Fatalf("Read post-update: %s", readResp2.Diagnostics)
	}
	var readModel2 objectStoreAccessPolicyRuleModel
	if diags := readResp2.State.Get(context.Background(), &readModel2); diags.HasError() {
		t.Fatalf("Get read2 state: %s", diags)
	}
	var readActions []string
	readModel2.Actions.ElementsAs(context.Background(), &readActions, false)
	if len(readActions) != 2 {
		t.Errorf("Read2: expected 2 actions, got %d", len(readActions))
	}

	// Step 5: Delete.
	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: readResp2.State}, deleteResp)
	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete: %s", deleteResp.Diagnostics)
	}
}

// TestUnit_OAPRule_ImportIdempotency verifies ImportState->Read produces state matching original Create.
func TestUnit_Unit_OAPRule_ImportIdempotency(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterObjectStoreAccessPolicyHandlers(ms.Mux)

	createTestOAPPolicy(t, ms, "idempotent-oap-rule-policy")

	r := newTestOAPRuleResource(t, ms)
	s := oapRuleResourceSchema(t).Schema

	// Create.
	createPlan := oapRulePlan(t, "idempotent-oap-rule-policy", "idempotent-rule", "allow",
		[]string{"s3:GetObject"},
		[]string{"arn:aws:s3:::test-bucket/*"},
	)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildObjectStoreAccessPolicyRuleType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}
	var createModel objectStoreAccessPolicyRuleModel
	if diags := createResp.State.Get(context.Background(), &createModel); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}

	// ImportState using composite ID "policy_name/rule_name".
	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildObjectStoreAccessPolicyRuleType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "idempotent-oap-rule-policy/idempotent-rule"}, importResp)
	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState: %s", importResp.Diagnostics)
	}

	// Read to populate full state.
	readResp := &resource.ReadResponse{State: importResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: importResp.State}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read post-import: %s", readResp.Diagnostics)
	}
	var importedModel objectStoreAccessPolicyRuleModel
	if diags := readResp.State.Get(context.Background(), &importedModel); diags.HasError() {
		t.Fatalf("Get imported state: %s", diags)
	}

	// Verify 0-diff.
	if importedModel.PolicyName.ValueString() != createModel.PolicyName.ValueString() {
		t.Errorf("policy_name mismatch: create=%s import=%s", createModel.PolicyName.ValueString(), importedModel.PolicyName.ValueString())
	}
	if importedModel.Name.ValueString() != createModel.Name.ValueString() {
		t.Errorf("name mismatch: create=%s import=%s", createModel.Name.ValueString(), importedModel.Name.ValueString())
	}
	if importedModel.Effect.ValueString() != createModel.Effect.ValueString() {
		t.Errorf("effect mismatch: create=%s import=%s", createModel.Effect.ValueString(), importedModel.Effect.ValueString())
	}
}

// TestUnit_OAPRule_PlanModifiers verifies all RequiresReplace and UseStateForUnknown
// plan modifiers in the OAP rule resource schema.
func TestUnit_Unit_OAPRule_PlanModifiers(t *testing.T) {
	s := oapRuleResourceSchema(t).Schema

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

	// name — RequiresReplace
	nameAttr, ok := s.Attributes["name"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("name attribute not found or wrong type")
	}
	if len(nameAttr.PlanModifiers) == 0 {
		t.Error("expected RequiresReplace plan modifier on name attribute")
	}

	// effect — RequiresReplace
	effectAttr, ok := s.Attributes["effect"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("effect attribute not found or wrong type")
	}
	if len(effectAttr.PlanModifiers) == 0 {
		t.Error("expected RequiresReplace plan modifier on effect attribute")
	}
}

// TestUnit_OAPRule_EffectValidator verifies the effect field rejects invalid values
// and accepts "allow" and "deny".
func TestUnit_Unit_OAPRule_EffectValidator(t *testing.T) {
	s := oapRuleResourceSchema(t).Schema

	eAttr, ok := s.Attributes["effect"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("effect attribute not found or wrong type")
	}
	if len(eAttr.Validators) == 0 {
		t.Fatal("expected at least one validator on effect attribute")
	}

	v := eAttr.Validators[0]

	// "invalid" should produce an error.
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), validator.StringRequest{
		ConfigValue: types.StringValue("invalid"),
	}, resp)
	if !resp.Diagnostics.HasError() {
		t.Error("expected validator to reject 'invalid' effect value")
	}

	// Valid values should not produce errors.
	for _, valid := range []string{"allow", "deny"} {
		resp := &validator.StringResponse{}
		v.ValidateString(context.Background(), validator.StringRequest{
			ConfigValue: types.StringValue(valid),
		}, resp)
		if resp.Diagnostics.HasError() {
			t.Errorf("expected validator to accept %q effect value, got error: %s", valid, resp.Diagnostics)
		}
	}
}
