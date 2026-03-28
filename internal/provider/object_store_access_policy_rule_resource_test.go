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

// newTestOAPRuleResource creates an objectStoreAccessPolicyRuleResource wired to the given mock server.
func newTestOAPRuleResource(t *testing.T, ms *testmock.MockServer) *objectStoreAccessPolicyRuleResource {
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

// buildOAPRuleType returns the tftypes.Object for the OAP rule resource.
func buildOAPRuleType() tftypes.Object {
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
		Raw:    tftypes.NewValue(buildOAPRuleType(), cfg),
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
		Raw:    tftypes.NewValue(buildOAPRuleType(), cfg),
		Schema: s,
	}
}

// createTestOAPPolicy creates a policy in the mock server.
func createTestOAPPolicy(t *testing.T, ms *testmock.MockServer, policyName string) {
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
	_, err = c.PostObjectStoreAccessPolicy(context.Background(), policyName, client.ObjectStoreAccessPolicyPost{})
	if err != nil {
		t.Fatalf("PostObjectStoreAccessPolicy(%s): %v", policyName, err)
	}
}

// ---- tests ------------------------------------------------------------------

// TestObjectStoreAccessPolicyRuleResource_Create verifies Create populates ID, effect, actions, resources.
func TestObjectStoreAccessPolicyRuleResource_Create(t *testing.T) {
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
		State: tfsdk.State{Raw: tftypes.NewValue(buildOAPRuleType(), nil), Schema: s},
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
	if model.Conditions.IsNull() {
		// Expected null since no conditions provided
	}
}

// TestObjectStoreAccessPolicyRuleResource_Update verifies PATCH updates actions list.
func TestObjectStoreAccessPolicyRuleResource_Update(t *testing.T) {
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
		State: tfsdk.State{Raw: tftypes.NewValue(buildOAPRuleType(), nil), Schema: s},
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
		State: tfsdk.State{Raw: tftypes.NewValue(buildOAPRuleType(), nil), Schema: s},
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
func TestObjectStoreAccessPolicyRuleResource_Delete(t *testing.T) {
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
		State: tfsdk.State{Raw: tftypes.NewValue(buildOAPRuleType(), nil), Schema: s},
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
	_, err = c.GetObjectStoreAccessPolicyRuleByName(context.Background(), "rule-delete-policy", "delete-rule")
	if err == nil || !client.IsNotFound(err) {
		t.Errorf("expected rule to be deleted, got: %v", err)
	}
}

// TestObjectStoreAccessPolicyRuleResource_Import verifies ImportState using composite "policy_name/rule_name".
func TestObjectStoreAccessPolicyRuleResource_Import(t *testing.T) {
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
		State: tfsdk.State{Raw: tftypes.NewValue(buildOAPRuleType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Import by "policy_name/rule_name".
	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildOAPRuleType(), nil), Schema: s},
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
func TestObjectStoreAccessPolicyRuleResource_ConditionsRoundTrip(t *testing.T) {
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
		State: tfsdk.State{Raw: tftypes.NewValue(buildOAPRuleType(), nil), Schema: s},
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
