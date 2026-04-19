package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	resschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/numberly/opentofu-provider-flashblade/internal/client"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock/handlers"
)

// ---- helpers ----------------------------------------------------------------

// newTestSMBClientRuleResource creates an smbClientPolicyRuleResource wired to the given mock server.
func newTestSMBClientRuleResource(t *testing.T, ms *testmock.MockServer) *smbClientPolicyRuleResource {
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
	return &smbClientPolicyRuleResource{client: c}
}

// smbClientRuleResourceSchema returns the parsed schema for the SMB client policy rule resource.
func smbClientRuleResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &smbClientPolicyRuleResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

// buildSmbClientPolicyRuleType returns the tftypes.Object for the SMB client policy rule resource.
func buildSmbClientPolicyRuleType() tftypes.Object {
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
		"client":      tftypes.String,
		"encryption":  tftypes.String,
		"permission":  tftypes.String,
		"timeouts":    timeoutsType,
	}}
}

// nullSMBClientRuleConfig returns a base config map with all attributes null.
func nullSMBClientRuleConfig() map[string]tftypes.Value {
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
		"client":      tftypes.NewValue(tftypes.String, nil),
		"encryption":  tftypes.NewValue(tftypes.String, nil),
		"permission":  tftypes.NewValue(tftypes.String, nil),
		"timeouts":    tftypes.NewValue(timeoutsType, nil),
	}
}

// smbClientRulePlan returns a tfsdk.Plan with the given fields set.
func smbClientRulePlan(t *testing.T, policyName, clientMatch, encryption, permission string) tfsdk.Plan {
	t.Helper()
	s := smbClientRuleResourceSchema(t).Schema
	cfg := nullSMBClientRuleConfig()
	cfg["policy_name"] = tftypes.NewValue(tftypes.String, policyName)
	if clientMatch != "" {
		cfg["client"] = tftypes.NewValue(tftypes.String, clientMatch)
	}
	if encryption != "" {
		cfg["encryption"] = tftypes.NewValue(tftypes.String, encryption)
	}
	if permission != "" {
		cfg["permission"] = tftypes.NewValue(tftypes.String, permission)
	}
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildSmbClientPolicyRuleType(), cfg),
		Schema: s,
	}
}

// createSMBClientPolicyForRuleTest creates an SMB client policy in the mock server so rule tests can attach rules to it.
func createSMBClientPolicyForRuleTest(t *testing.T, ms *testmock.MockServer, policyName string) {
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
	_, err = c.PostSmbClientPolicy(context.Background(), policyName, client.SmbClientPolicyPost{Enabled: &enabled})
	if err != nil {
		t.Fatalf("PostSmbClientPolicy: %v", err)
	}
}

// ---- tests ------------------------------------------------------------------

// TestUnit_SmbClientPolicyRule_CRUD exercises the full Create->Read->Update(client)->Delete sequence.
func TestUnit_SmbClientPolicyRule_CRUD(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterSmbClientPolicyHandlers(ms.Mux)

	r := newTestSMBClientRuleResource(t, ms)
	s := smbClientRuleResourceSchema(t).Schema

	createSMBClientPolicyForRuleTest(t, ms, "client-rule-test-policy")

	// Step 1: Create with client="*", encryption="optional", permission="rw".
	createPlan := smbClientRulePlan(t, "client-rule-test-policy", "*", "optional", "rw")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSmbClientPolicyRuleType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}
	var createModel smbClientPolicyRuleModel
	if diags := createResp.State.Get(context.Background(), &createModel); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}
	if createModel.Client.ValueString() != "*" {
		t.Errorf("Create: expected client=*, got %s", createModel.Client.ValueString())
	}
	if createModel.Encryption.ValueString() != "optional" {
		t.Errorf("Create: expected encryption=optional, got %s", createModel.Encryption.ValueString())
	}
	if createModel.Permission.ValueString() != "rw" {
		t.Errorf("Create: expected permission=rw, got %s", createModel.Permission.ValueString())
	}
	if createModel.ID.IsNull() || createModel.ID.ValueString() == "" {
		t.Error("expected non-empty ID after Create")
	}
	if createModel.Name.IsNull() || createModel.Name.ValueString() == "" {
		t.Error("expected server-assigned name to be populated")
	}

	// Step 2: Read post-create.
	readResp1 := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp1)
	if readResp1.Diagnostics.HasError() {
		t.Fatalf("Read post-create: %s", readResp1.Diagnostics)
	}

	// Step 3: Update client to "10.0.0.0/8".
	updateCfg := nullSMBClientRuleConfig()
	updateCfg["policy_name"] = tftypes.NewValue(tftypes.String, "client-rule-test-policy")
	updateCfg["client"] = tftypes.NewValue(tftypes.String, "10.0.0.0/8")
	updateCfg["encryption"] = tftypes.NewValue(tftypes.String, "optional")
	updateCfg["permission"] = tftypes.NewValue(tftypes.String, "rw")
	updatePlan := tfsdk.Plan{
		Raw:    tftypes.NewValue(buildSmbClientPolicyRuleType(), updateCfg),
		Schema: s,
	}
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSmbClientPolicyRuleType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlan,
		State: readResp1.State,
	}, updateResp)
	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update: %s", updateResp.Diagnostics)
	}
	var updateModel smbClientPolicyRuleModel
	if diags := updateResp.State.Get(context.Background(), &updateModel); diags.HasError() {
		t.Fatalf("Get update state: %s", diags)
	}
	if updateModel.Client.ValueString() != "10.0.0.0/8" {
		t.Errorf("Update: expected client=10.0.0.0/8, got %s", updateModel.Client.ValueString())
	}

	// Step 4: Delete.
	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: updateResp.State}, deleteResp)
	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete: %s", deleteResp.Diagnostics)
	}
}

// TestUnit_SmbClientPolicyRule_Import verifies ImportState using composite ID "policy_name/rule_name".
func TestUnit_SmbClientPolicyRule_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterSmbClientPolicyHandlers(ms.Mux)

	r := newTestSMBClientRuleResource(t, ms)
	s := smbClientRuleResourceSchema(t).Schema

	createSMBClientPolicyForRuleTest(t, ms, "import-client-rule-policy")

	// Create first.
	createPlan := smbClientRulePlan(t, "import-client-rule-policy", "*", "optional", "rw")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSmbClientPolicyRuleType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}
	var createdModel smbClientPolicyRuleModel
	if diags := createResp.State.Get(context.Background(), &createdModel); diags.HasError() {
		t.Fatalf("Get created state: %s", diags)
	}

	// Import by composite ID "policy_name/rule_name".
	compositeID := "import-client-rule-policy/" + createdModel.Name.ValueString()
	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSmbClientPolicyRuleType(), nil), Schema: s},
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
	var importedModel smbClientPolicyRuleModel
	if diags := readResp.State.Get(context.Background(), &importedModel); diags.HasError() {
		t.Fatalf("Get imported state: %s", diags)
	}

	// Verify 0-diff.
	if importedModel.PolicyName.ValueString() != "import-client-rule-policy" {
		t.Errorf("policy_name mismatch: expected import-client-rule-policy, got %s", importedModel.PolicyName.ValueString())
	}
	if importedModel.Client.ValueString() != createdModel.Client.ValueString() {
		t.Errorf("client mismatch: create=%s import=%s", createdModel.Client.ValueString(), importedModel.Client.ValueString())
	}
	if importedModel.Encryption.ValueString() != createdModel.Encryption.ValueString() {
		t.Errorf("encryption mismatch: create=%s import=%s", createdModel.Encryption.ValueString(), importedModel.Encryption.ValueString())
	}
	if importedModel.Permission.ValueString() != createdModel.Permission.ValueString() {
		t.Errorf("permission mismatch: create=%s import=%s", createdModel.Permission.ValueString(), importedModel.Permission.ValueString())
	}
	if importedModel.ID.IsNull() || importedModel.ID.ValueString() == "" {
		t.Error("expected ID to be populated after import")
	}
}

// TestUnit_SmbClientPolicyRule_PlanModifiers verifies all RequiresReplace and UseStateForUnknown
// plan modifiers in the smb_client_policy_rule resource schema.
func TestUnit_SmbClientPolicyRule_PlanModifiers(t *testing.T) {
	s := smbClientRuleResourceSchema(t).Schema

	// id
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

	// index — UseStateForUnknown
	indexAttr, ok := s.Attributes["index"].(resschema.Int64Attribute)
	if !ok {
		t.Fatal("index attribute not found or wrong type")
	}
	if len(indexAttr.PlanModifiers) == 0 {
		t.Error("expected UseStateForUnknown plan modifier on index attribute")
	}
}

// TestUnit_SmbClientPolicyRule_Idempotent verifies that Read after Create shows no attribute drift.
func TestUnit_SmbClientPolicyRule_Idempotent(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterSmbClientPolicyHandlers(ms.Mux)

	r := newTestSMBClientRuleResource(t, ms)
	s := smbClientRuleResourceSchema(t).Schema

	createSMBClientPolicyForRuleTest(t, ms, "idempotent-rule-policy")

	plan := smbClientRulePlan(t, "idempotent-rule-policy", "*", "optional", "rw")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSmbClientPolicyRuleType(), nil), Schema: s},
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

	var beforeModel, afterModel smbClientPolicyRuleModel
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
	if beforeModel.Client.ValueString() != afterModel.Client.ValueString() {
		t.Errorf("Client changed after Read: %s -> %s", beforeModel.Client.ValueString(), afterModel.Client.ValueString())
	}
	if beforeModel.Encryption.ValueString() != afterModel.Encryption.ValueString() {
		t.Errorf("Encryption changed after Read: %s -> %s", beforeModel.Encryption.ValueString(), afterModel.Encryption.ValueString())
	}
	if beforeModel.Permission.ValueString() != afterModel.Permission.ValueString() {
		t.Errorf("Permission changed after Read: %s -> %s", beforeModel.Permission.ValueString(), afterModel.Permission.ValueString())
	}
}
