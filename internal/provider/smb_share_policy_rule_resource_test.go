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

// newTestSMBRuleResource creates an smbSharePolicyRuleResource wired to the given mock server.
func newTestSMBRuleResource(t *testing.T, ms *testmock.MockServer) *smbSharePolicyRuleResource {
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
	return &smbSharePolicyRuleResource{client: c}
}

// smbRuleResourceSchema returns the parsed schema for the SMB share policy rule resource.
func smbRuleResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &smbSharePolicyRuleResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

// buildSMBRuleType returns the tftypes.Object for the SMB share policy rule resource.
func buildSMBRuleType() tftypes.Object {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":           tftypes.String,
		"policy_name":  tftypes.String,
		"name":         tftypes.String,
		"principal":    tftypes.String,
		"change":       tftypes.String,
		"full_control": tftypes.String,
		"read":         tftypes.String,
		"timeouts":     timeoutsType,
	}}
}

// nullSMBRuleConfig returns a base config map with all attributes null.
func nullSMBRuleConfig() map[string]tftypes.Value {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return map[string]tftypes.Value{
		"id":           tftypes.NewValue(tftypes.String, nil),
		"policy_name":  tftypes.NewValue(tftypes.String, nil),
		"name":         tftypes.NewValue(tftypes.String, nil),
		"principal":    tftypes.NewValue(tftypes.String, nil),
		"change":       tftypes.NewValue(tftypes.String, nil),
		"full_control": tftypes.NewValue(tftypes.String, nil),
		"read":         tftypes.NewValue(tftypes.String, nil),
		"timeouts":     tftypes.NewValue(timeoutsType, nil),
	}
}

// smbRulePlan returns a tfsdk.Plan with the given fields set.
func smbRulePlan(t *testing.T, policyName, principal, change, fullControl, readPerm string) tfsdk.Plan {
	t.Helper()
	s := smbRuleResourceSchema(t).Schema
	cfg := nullSMBRuleConfig()
	cfg["policy_name"] = tftypes.NewValue(tftypes.String, policyName)
	if principal != "" {
		cfg["principal"] = tftypes.NewValue(tftypes.String, principal)
	}
	if change != "" {
		cfg["change"] = tftypes.NewValue(tftypes.String, change)
	}
	if fullControl != "" {
		cfg["full_control"] = tftypes.NewValue(tftypes.String, fullControl)
	}
	if readPerm != "" {
		cfg["read"] = tftypes.NewValue(tftypes.String, readPerm)
	}
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildSMBRuleType(), cfg),
		Schema: s,
	}
}

// createSMBPolicyForRuleTest creates an SMB policy in the mock server so rule tests can attach rules to it.
func createSMBPolicyForRuleTest(t *testing.T, ms *testmock.MockServer, policyName string) {
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
	enabled := true
	_, err = c.PostSmbSharePolicy(context.Background(), policyName, client.SmbSharePolicyPost{Enabled: &enabled})
	if err != nil {
		t.Fatalf("PostSmbSharePolicy: %v", err)
	}
}

// ---- tests ------------------------------------------------------------------

// TestSmbSharePolicyRuleResource_Create verifies Create populates ID, policy_name, name, and rule fields.
func TestSmbSharePolicyRuleResource_Create(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterSmbSharePolicyHandlers(ms.Mux)

	createSMBPolicyForRuleTest(t, ms, "rule-test-policy")

	r := newTestSMBRuleResource(t, ms)
	s := smbRuleResourceSchema(t).Schema

	plan := smbRulePlan(t, "rule-test-policy", "Everyone", "allow", "deny", "allow")
	resp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSMBRuleType(), nil), Schema: s},
	}

	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Create returned error: %s", resp.Diagnostics)
	}

	var model smbSharePolicyRuleModel
	if diags := resp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected non-empty ID after Create")
	}
	if model.Name.IsNull() || model.Name.ValueString() == "" {
		t.Error("expected server-assigned name to be populated")
	}
	if model.PolicyName.ValueString() != "rule-test-policy" {
		t.Errorf("expected policy_name=rule-test-policy, got %s", model.PolicyName.ValueString())
	}
	if model.Principal.ValueString() != "Everyone" {
		t.Errorf("expected principal=Everyone, got %s", model.Principal.ValueString())
	}
	if model.Change.ValueString() != "allow" {
		t.Errorf("expected change=allow, got %s", model.Change.ValueString())
	}
	if model.FullControl.ValueString() != "deny" {
		t.Errorf("expected full_control=deny, got %s", model.FullControl.ValueString())
	}
	if model.Read.ValueString() != "allow" {
		t.Errorf("expected read=allow, got %s", model.Read.ValueString())
	}
}

// TestSmbSharePolicyRuleResource_Update verifies PATCH updates rule fields in-place.
func TestSmbSharePolicyRuleResource_Update(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterSmbSharePolicyHandlers(ms.Mux)

	createSMBPolicyForRuleTest(t, ms, "update-rule-policy")

	r := newTestSMBRuleResource(t, ms)
	s := smbRuleResourceSchema(t).Schema

	// Create first.
	createPlan := smbRulePlan(t, "update-rule-policy", "Everyone", "allow", "deny", "allow")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSMBRuleType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Update change from "allow" to "deny".
	updateCfg := nullSMBRuleConfig()
	updateCfg["policy_name"] = tftypes.NewValue(tftypes.String, "update-rule-policy")
	updateCfg["principal"] = tftypes.NewValue(tftypes.String, "Everyone")
	updateCfg["change"] = tftypes.NewValue(tftypes.String, "deny")
	updateCfg["full_control"] = tftypes.NewValue(tftypes.String, "deny")
	updateCfg["read"] = tftypes.NewValue(tftypes.String, "allow")
	updatePlan := tfsdk.Plan{
		Raw:    tftypes.NewValue(buildSMBRuleType(), updateCfg),
		Schema: s,
	}

	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSMBRuleType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlan,
		State: createResp.State,
	}, updateResp)

	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update returned error: %s", updateResp.Diagnostics)
	}

	var model smbSharePolicyRuleModel
	if diags := updateResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Change.ValueString() != "deny" {
		t.Errorf("expected change=deny after update, got %s", model.Change.ValueString())
	}
}

// TestSmbSharePolicyRuleResource_Delete verifies DELETE removes the rule.
func TestSmbSharePolicyRuleResource_Delete(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterSmbSharePolicyHandlers(ms.Mux)

	createSMBPolicyForRuleTest(t, ms, "delete-rule-policy")

	r := newTestSMBRuleResource(t, ms)
	s := smbRuleResourceSchema(t).Schema

	// Create first.
	createPlan := smbRulePlan(t, "delete-rule-policy", "Everyone", "allow", "allow", "allow")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSMBRuleType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	var createdModel smbSharePolicyRuleModel
	if diags := createResp.State.Get(context.Background(), &createdModel); diags.HasError() {
		t.Fatalf("Get created state: %s", diags)
	}
	ruleName := createdModel.Name.ValueString()

	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: createResp.State}, deleteResp)

	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete returned error: %s", deleteResp.Diagnostics)
	}

	// Verify rule is gone.
	_, err := r.client.GetSmbSharePolicyRuleByName(context.Background(), "delete-rule-policy", ruleName)
	if err == nil || !client.IsNotFound(err) {
		t.Errorf("expected rule to be deleted, got: %v", err)
	}
}

// TestSmbSharePolicyRuleResource_Import verifies ImportState using composite ID "policy_name/rule_name".
func TestSmbSharePolicyRuleResource_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterSmbSharePolicyHandlers(ms.Mux)

	createSMBPolicyForRuleTest(t, ms, "import-rule-policy")

	r := newTestSMBRuleResource(t, ms)
	s := smbRuleResourceSchema(t).Schema

	// Create first.
	createPlan := smbRulePlan(t, "import-rule-policy", "DOMAIN\\user", "allow", "deny", "allow")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSMBRuleType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	var createdModel smbSharePolicyRuleModel
	if diags := createResp.State.Get(context.Background(), &createdModel); diags.HasError() {
		t.Fatalf("Get created state: %s", diags)
	}

	// Import by composite ID "policy_name/rule_name".
	compositeID := "import-rule-policy/" + createdModel.Name.ValueString()
	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSMBRuleType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: compositeID}, importResp)

	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState returned error: %s", importResp.Diagnostics)
	}

	var model smbSharePolicyRuleModel
	if diags := importResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get imported state: %s", diags)
	}

	if model.PolicyName.ValueString() != "import-rule-policy" {
		t.Errorf("expected policy_name=import-rule-policy after import, got %s", model.PolicyName.ValueString())
	}
	if model.Name.ValueString() != createdModel.Name.ValueString() {
		t.Errorf("expected name=%s after import, got %s", createdModel.Name.ValueString(), model.Name.ValueString())
	}
	if model.Principal.ValueString() != "DOMAIN\\user" {
		t.Errorf("expected principal=DOMAIN\\user after import, got %s", model.Principal.ValueString())
	}
	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected ID to be populated after import")
	}
}

// TestUnit_SMBRule_PlanModifiers verifies all RequiresReplace and UseStateForUnknown
// plan modifiers in the smb_share_policy_rule resource schema.
func TestUnit_SMBRule_PlanModifiers(t *testing.T) {
	s := smbRuleResourceSchema(t).Schema

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

	// name — UseStateForUnknown (computed, server-assigned)
	nameAttr, ok := s.Attributes["name"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("name attribute not found or wrong type")
	}
	if len(nameAttr.PlanModifiers) == 0 {
		t.Error("expected UseStateForUnknown plan modifier on name attribute")
	}
}
