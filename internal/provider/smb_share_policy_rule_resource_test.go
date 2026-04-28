package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	resschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/numberly/terraform-provider-mica/internal/client"
	"github.com/numberly/terraform-provider-mica/internal/testmock"
	"github.com/numberly/terraform-provider-mica/internal/testmock/handlers"
)

// ---- helpers ----------------------------------------------------------------

// newTestSMBRuleResource creates an smbSharePolicyRuleResource wired to the given mock server.
func newTestSMBRuleResource(t *testing.T, ms *testmock.MockServer) *smbSharePolicyRuleResource {
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

// buildSmbSharePolicyRuleType returns the tftypes.Object for the SMB share policy rule resource.
func buildSmbSharePolicyRuleType() tftypes.Object {
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
		Raw:    tftypes.NewValue(buildSmbSharePolicyRuleType(), cfg),
		Schema: s,
	}
}

// createSMBPolicyForRuleTest creates an SMB policy in the mock server so rule tests can attach rules to it.
func createSMBPolicyForRuleTest(t *testing.T, ms *testmock.MockServer, policyName string) {
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
	enabled := true
	_, err = c.PostSmbSharePolicy(context.Background(), policyName, client.SmbSharePolicyPost{Enabled: &enabled})
	if err != nil {
		t.Fatalf("PostSmbSharePolicy: %v", err)
	}
}

// ---- tests ------------------------------------------------------------------

// TestSmbSharePolicyRuleResource_Create verifies Create populates ID, policy_name, name, and rule fields.
func TestUnit_SmbSharePolicyRuleResource_Create(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterSmbSharePolicyHandlers(ms.Mux)

	createSMBPolicyForRuleTest(t, ms, "rule-test-policy")

	r := newTestSMBRuleResource(t, ms)
	s := smbRuleResourceSchema(t).Schema

	plan := smbRulePlan(t, "rule-test-policy", "Everyone", "allow", "deny", "allow")
	resp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSmbSharePolicyRuleType(), nil), Schema: s},
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
func TestUnit_SmbSharePolicyRuleResource_Update(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterSmbSharePolicyHandlers(ms.Mux)

	createSMBPolicyForRuleTest(t, ms, "update-rule-policy")

	r := newTestSMBRuleResource(t, ms)
	s := smbRuleResourceSchema(t).Schema

	// Create first.
	createPlan := smbRulePlan(t, "update-rule-policy", "Everyone", "allow", "deny", "allow")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSmbSharePolicyRuleType(), nil), Schema: s},
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
		Raw:    tftypes.NewValue(buildSmbSharePolicyRuleType(), updateCfg),
		Schema: s,
	}

	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSmbSharePolicyRuleType(), nil), Schema: s},
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
func TestUnit_SmbSharePolicyRuleResource_Delete(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterSmbSharePolicyHandlers(ms.Mux)

	createSMBPolicyForRuleTest(t, ms, "delete-rule-policy")

	r := newTestSMBRuleResource(t, ms)
	s := smbRuleResourceSchema(t).Schema

	// Create first.
	createPlan := smbRulePlan(t, "delete-rule-policy", "Everyone", "allow", "allow", "allow")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSmbSharePolicyRuleType(), nil), Schema: s},
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
func TestUnit_SmbSharePolicyRuleResource_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterSmbSharePolicyHandlers(ms.Mux)

	createSMBPolicyForRuleTest(t, ms, "import-rule-policy")

	r := newTestSMBRuleResource(t, ms)
	s := smbRuleResourceSchema(t).Schema

	// Create first.
	createPlan := smbRulePlan(t, "import-rule-policy", "DOMAIN\\user", "allow", "deny", "allow")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSmbSharePolicyRuleType(), nil), Schema: s},
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
		State: tfsdk.State{Raw: tftypes.NewValue(buildSmbSharePolicyRuleType(), nil), Schema: s},
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

// TestUnit_SmbSharePolicyRule_Lifecycle exercises the full Create->Read->Update->Read->Delete sequence.
func TestUnit_Unit_SmbSharePolicyRule_Lifecycle(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterSmbSharePolicyHandlers(ms.Mux)

	r := newTestSMBRuleResource(t, ms)
	s := smbRuleResourceSchema(t).Schema

	createSMBPolicyForRuleTest(t, ms, "lifecycle-smb-rule-policy")

	// Step 1: Create.
	createPlan := smbRulePlan(t, "lifecycle-smb-rule-policy", "Everyone", "allow", "deny", "allow")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSmbSharePolicyRuleType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}
	var createModel smbSharePolicyRuleModel
	if diags := createResp.State.Get(context.Background(), &createModel); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}
	if createModel.Change.ValueString() != "allow" {
		t.Errorf("Create: expected change=allow, got %s", createModel.Change.ValueString())
	}

	// Step 2: Read post-create.
	readResp1 := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp1)
	if readResp1.Diagnostics.HasError() {
		t.Fatalf("Read post-create: %s", readResp1.Diagnostics)
	}
	var readModel1 smbSharePolicyRuleModel
	if diags := readResp1.State.Get(context.Background(), &readModel1); diags.HasError() {
		t.Fatalf("Get read1 state: %s", diags)
	}
	if readModel1.Principal.ValueString() != "Everyone" {
		t.Errorf("Read1: expected principal=Everyone, got %s", readModel1.Principal.ValueString())
	}

	// Step 3: Update change to "deny".
	updateCfg := nullSMBRuleConfig()
	updateCfg["policy_name"] = tftypes.NewValue(tftypes.String, "lifecycle-smb-rule-policy")
	updateCfg["principal"] = tftypes.NewValue(tftypes.String, "Everyone")
	updateCfg["change"] = tftypes.NewValue(tftypes.String, "deny")
	updateCfg["full_control"] = tftypes.NewValue(tftypes.String, "deny")
	updateCfg["read"] = tftypes.NewValue(tftypes.String, "allow")
	updatePlan := tfsdk.Plan{Raw: tftypes.NewValue(buildSmbSharePolicyRuleType(), updateCfg), Schema: s}
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSmbSharePolicyRuleType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlan,
		State: readResp1.State,
	}, updateResp)
	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update: %s", updateResp.Diagnostics)
	}
	var updateModel smbSharePolicyRuleModel
	if diags := updateResp.State.Get(context.Background(), &updateModel); diags.HasError() {
		t.Fatalf("Get update state: %s", diags)
	}
	if updateModel.Change.ValueString() != "deny" {
		t.Errorf("Update: expected change=deny, got %s", updateModel.Change.ValueString())
	}

	// Step 4: Read post-update.
	readResp2 := &resource.ReadResponse{State: updateResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: updateResp.State}, readResp2)
	if readResp2.Diagnostics.HasError() {
		t.Fatalf("Read post-update: %s", readResp2.Diagnostics)
	}
	var readModel2 smbSharePolicyRuleModel
	if diags := readResp2.State.Get(context.Background(), &readModel2); diags.HasError() {
		t.Fatalf("Get read2 state: %s", diags)
	}
	if readModel2.Change.ValueString() != "deny" {
		t.Errorf("Read2: expected change=deny, got %s", readModel2.Change.ValueString())
	}

	// Step 5: Delete.
	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: readResp2.State}, deleteResp)
	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete: %s", deleteResp.Diagnostics)
	}
}

// TestUnit_SmbSharePolicyRule_ImportIdempotency verifies ImportState->Read produces state matching original Create.
func TestUnit_Unit_SmbSharePolicyRule_ImportIdempotency(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterSmbSharePolicyHandlers(ms.Mux)

	r := newTestSMBRuleResource(t, ms)
	s := smbRuleResourceSchema(t).Schema

	createSMBPolicyForRuleTest(t, ms, "idempotent-smb-rule-policy")

	// Create.
	createPlan := smbRulePlan(t, "idempotent-smb-rule-policy", "Everyone", "allow", "deny", "allow")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSmbSharePolicyRuleType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}
	var createModel smbSharePolicyRuleModel
	if diags := createResp.State.Get(context.Background(), &createModel); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}

	// ImportState using composite ID "policy_name/rule_name".
	compositeID := "idempotent-smb-rule-policy/" + createModel.Name.ValueString()
	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSmbSharePolicyRuleType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: compositeID}, importResp)
	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState: %s", importResp.Diagnostics)
	}

	// Read to populate full state.
	readResp := &resource.ReadResponse{State: importResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: importResp.State}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read post-import: %s", readResp.Diagnostics)
	}
	var importedModel smbSharePolicyRuleModel
	if diags := readResp.State.Get(context.Background(), &importedModel); diags.HasError() {
		t.Fatalf("Get imported state: %s", diags)
	}

	// Verify 0-diff.
	if importedModel.PolicyName.ValueString() != createModel.PolicyName.ValueString() {
		t.Errorf("policy_name mismatch: create=%s import=%s", createModel.PolicyName.ValueString(), importedModel.PolicyName.ValueString())
	}
	if importedModel.Principal.ValueString() != createModel.Principal.ValueString() {
		t.Errorf("principal mismatch: create=%s import=%s", createModel.Principal.ValueString(), importedModel.Principal.ValueString())
	}
	if importedModel.Change.ValueString() != createModel.Change.ValueString() {
		t.Errorf("change mismatch: create=%s import=%s", createModel.Change.ValueString(), importedModel.Change.ValueString())
	}
}

// TestUnit_SMBRule_PlanModifiers verifies all RequiresReplace and UseStateForUnknown
// plan modifiers in the smb_share_policy_rule resource schema.
func TestUnit_Unit_SMBRule_PlanModifiers(t *testing.T) {
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
