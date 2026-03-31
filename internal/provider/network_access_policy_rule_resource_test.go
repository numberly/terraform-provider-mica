package provider

import (
	"context"
	"strconv"
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

// newTestNAPRuleResource creates a networkAccessPolicyRuleResource wired to the given mock server.
func newTestNAPRuleResource(t *testing.T, ms *testmock.MockServer) *networkAccessPolicyRuleResource {
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
	return &networkAccessPolicyRuleResource{client: c}
}

// napRuleResourceSchema returns the parsed schema for the NAP rule resource.
func napRuleResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &networkAccessPolicyRuleResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

// buildNAPRuleType returns the tftypes.Object for the NAP rule resource.
func buildNAPRuleType() tftypes.Object {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":             tftypes.String,
		"policy_name":    tftypes.String,
		"name":           tftypes.String,
		"index":          tftypes.Number,
		"client":         tftypes.String,
		"effect":         tftypes.String,
		"interfaces":     tftypes.List{ElementType: tftypes.String},
		"policy_version": tftypes.String,
		"timeouts":       timeoutsType,
	}}
}

// nullNAPRuleConfig returns a base config map with all attributes null.
func nullNAPRuleConfig() map[string]tftypes.Value {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return map[string]tftypes.Value{
		"id":             tftypes.NewValue(tftypes.String, nil),
		"policy_name":    tftypes.NewValue(tftypes.String, nil),
		"name":           tftypes.NewValue(tftypes.String, nil),
		"index":          tftypes.NewValue(tftypes.Number, nil),
		"client":         tftypes.NewValue(tftypes.String, nil),
		"effect":         tftypes.NewValue(tftypes.String, nil),
		"interfaces":     tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"policy_version": tftypes.NewValue(tftypes.String, nil),
		"timeouts":       tftypes.NewValue(timeoutsType, nil),
	}
}

// napRulePlan builds a tfsdk.Plan for a NAP rule with the given policy_name, client, effect, and optional interfaces.
func napRulePlan(t *testing.T, policyName, clientStr, effect string, interfaces []string) tfsdk.Plan {
	t.Helper()
	s := napRuleResourceSchema(t).Schema
	cfg := nullNAPRuleConfig()
	cfg["policy_name"] = tftypes.NewValue(tftypes.String, policyName)
	if clientStr != "" {
		cfg["client"] = tftypes.NewValue(tftypes.String, clientStr)
	}
	if effect != "" {
		cfg["effect"] = tftypes.NewValue(tftypes.String, effect)
	}
	if len(interfaces) > 0 {
		ifaceVals := make([]tftypes.Value, len(interfaces))
		for i, iface := range interfaces {
			ifaceVals[i] = tftypes.NewValue(tftypes.String, iface)
		}
		cfg["interfaces"] = tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, ifaceVals)
	}
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildNAPRuleType(), cfg),
		Schema: s,
	}
}

// ---- tests ------------------------------------------------------------------

// TestNetworkAccessPolicyRuleResource_Create verifies Create populates id, name, index, and rule fields.
// The mock pre-seeds a "default" policy, so rules can be created in it directly.
func TestNetworkAccessPolicyRuleResource_Create(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterNetworkAccessPolicyHandlers(ms.Mux)

	r := newTestNAPRuleResource(t, ms)
	s := napRuleResourceSchema(t).Schema

	plan := napRulePlan(t, "default", "*", "allow", []string{"nfs", "smb"})
	resp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNAPRuleType(), nil), Schema: s},
	}

	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Create returned error: %s", resp.Diagnostics)
	}

	var model networkAccessPolicyRuleModel
	if diags := resp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected non-empty ID after Create")
	}
	if model.Name.IsNull() || model.Name.ValueString() == "" {
		t.Error("expected non-empty name after Create (server-assigned)")
	}
	if model.Index.IsNull() {
		t.Error("expected index to be set after Create")
	}
	if model.PolicyName.ValueString() != "default" {
		t.Errorf("expected policy_name=default, got %s", model.PolicyName.ValueString())
	}
	if model.Client.ValueString() != "*" {
		t.Errorf("expected client=*, got %s", model.Client.ValueString())
	}
	if model.Effect.ValueString() != "allow" {
		t.Errorf("expected effect=allow, got %s", model.Effect.ValueString())
	}
	// Verify interfaces were stored.
	var ifaces []string
	if diags := model.Interfaces.ElementsAs(context.Background(), &ifaces, false); diags.HasError() {
		t.Fatalf("ElementsAs interfaces: %s", diags)
	}
	if len(ifaces) != 2 {
		t.Errorf("expected 2 interfaces, got %d: %v", len(ifaces), ifaces)
	}
}

// TestNetworkAccessPolicyRuleResource_Update verifies PATCH updates the client field.
func TestNetworkAccessPolicyRuleResource_Update(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterNetworkAccessPolicyHandlers(ms.Mux)

	r := newTestNAPRuleResource(t, ms)
	s := napRuleResourceSchema(t).Schema

	// Create rule first.
	createPlan := napRulePlan(t, "default", "*", "allow", nil)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNAPRuleType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Update client to "10.0.0.0/8".
	updatePlan := napRulePlan(t, "default", "10.0.0.0/8", "allow", nil)
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNAPRuleType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlan,
		State: createResp.State,
	}, updateResp)

	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update returned error: %s", updateResp.Diagnostics)
	}

	var model networkAccessPolicyRuleModel
	if diags := updateResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}
	if model.Client.ValueString() != "10.0.0.0/8" {
		t.Errorf("expected client=10.0.0.0/8 after update, got %s", model.Client.ValueString())
	}
}

// TestNetworkAccessPolicyRuleResource_Delete verifies DELETE removes the rule.
func TestNetworkAccessPolicyRuleResource_Delete(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterNetworkAccessPolicyHandlers(ms.Mux)

	r := newTestNAPRuleResource(t, ms)
	s := napRuleResourceSchema(t).Schema

	// Create rule first.
	createPlan := napRulePlan(t, "default", "*", "deny", nil)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNAPRuleType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	var createdModel networkAccessPolicyRuleModel
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
	_, err := r.client.GetNetworkAccessPolicyRuleByName(context.Background(), "default", ruleName)
	if err == nil || !client.IsNotFound(err) {
		t.Errorf("expected rule to be deleted, got: %v", err)
	}
}

// TestNetworkAccessPolicyRuleResource_Import verifies ImportState with composite ID "policy_name/index".
func TestNetworkAccessPolicyRuleResource_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterNetworkAccessPolicyHandlers(ms.Mux)

	r := newTestNAPRuleResource(t, ms)
	s := napRuleResourceSchema(t).Schema

	// Create rule in the pre-seeded "default" policy.
	createPlan := napRulePlan(t, "default", "*", "allow", nil)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNAPRuleType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	var createdModel networkAccessPolicyRuleModel
	if diags := createResp.State.Get(context.Background(), &createdModel); diags.HasError() {
		t.Fatalf("Get created state: %s", diags)
	}
	index := strconv.FormatInt(createdModel.Index.ValueInt64(), 10)

	// Import using "policy_name/index" composite ID.
	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNAPRuleType(), nil), Schema: s},
	}
	importID := "default/" + index
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: importID}, importResp)

	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState returned error: %s", importResp.Diagnostics)
	}

	var model networkAccessPolicyRuleModel
	if diags := importResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.PolicyName.ValueString() != "default" {
		t.Errorf("expected policy_name=default after import, got %s", model.PolicyName.ValueString())
	}
	if model.Name.IsNull() || model.Name.ValueString() == "" {
		t.Error("expected server-assigned name to be populated after import")
	}
	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected ID to be populated after import")
	}
	if model.Client.ValueString() != "*" {
		t.Errorf("expected client=* after import, got %s", model.Client.ValueString())
	}
	if model.Effect.ValueString() != "allow" {
		t.Errorf("expected effect=allow after import, got %s", model.Effect.ValueString())
	}
}

// TestUnit_NetworkAccessPolicyRule_Lifecycle exercises the full Create->Read->Update->Read->Delete sequence.
func TestUnit_NetworkAccessPolicyRule_Lifecycle(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterNetworkAccessPolicyHandlers(ms.Mux)

	r := newTestNAPRuleResource(t, ms)
	s := napRuleResourceSchema(t).Schema

	// Step 1: Create (mock pre-seeds "default" policy).
	createPlan := napRulePlan(t, "default", "*", "allow", []string{"nfs"})
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNAPRuleType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}
	var createModel networkAccessPolicyRuleModel
	if diags := createResp.State.Get(context.Background(), &createModel); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}
	if createModel.Client.ValueString() != "*" {
		t.Errorf("Create: expected client=*, got %s", createModel.Client.ValueString())
	}

	// Step 2: Read post-create.
	readResp1 := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp1)
	if readResp1.Diagnostics.HasError() {
		t.Fatalf("Read post-create: %s", readResp1.Diagnostics)
	}
	var readModel1 networkAccessPolicyRuleModel
	if diags := readResp1.State.Get(context.Background(), &readModel1); diags.HasError() {
		t.Fatalf("Get read1 state: %s", diags)
	}
	if readModel1.Effect.ValueString() != "allow" {
		t.Errorf("Read1: expected effect=allow, got %s", readModel1.Effect.ValueString())
	}

	// Step 3: Update client to specific subnet.
	updatePlan := napRulePlan(t, "default", "10.0.0.0/8", "allow", []string{"nfs"})
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNAPRuleType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlan,
		State: readResp1.State,
	}, updateResp)
	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update: %s", updateResp.Diagnostics)
	}
	var updateModel networkAccessPolicyRuleModel
	if diags := updateResp.State.Get(context.Background(), &updateModel); diags.HasError() {
		t.Fatalf("Get update state: %s", diags)
	}
	if updateModel.Client.ValueString() != "10.0.0.0/8" {
		t.Errorf("Update: expected client=10.0.0.0/8, got %s", updateModel.Client.ValueString())
	}

	// Step 4: Read post-update.
	readResp2 := &resource.ReadResponse{State: updateResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: updateResp.State}, readResp2)
	if readResp2.Diagnostics.HasError() {
		t.Fatalf("Read post-update: %s", readResp2.Diagnostics)
	}
	var readModel2 networkAccessPolicyRuleModel
	if diags := readResp2.State.Get(context.Background(), &readModel2); diags.HasError() {
		t.Fatalf("Get read2 state: %s", diags)
	}
	if readModel2.Client.ValueString() != "10.0.0.0/8" {
		t.Errorf("Read2: expected client=10.0.0.0/8, got %s", readModel2.Client.ValueString())
	}

	// Step 5: Delete.
	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: readResp2.State}, deleteResp)
	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete: %s", deleteResp.Diagnostics)
	}
}

// TestUnit_NetworkAccessPolicyRule_ImportIdempotency verifies ImportState->Read produces state matching original Create.
func TestUnit_NetworkAccessPolicyRule_ImportIdempotency(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterNetworkAccessPolicyHandlers(ms.Mux)

	r := newTestNAPRuleResource(t, ms)
	s := napRuleResourceSchema(t).Schema

	// Create.
	createPlan := napRulePlan(t, "default", "*", "allow", []string{"nfs", "smb"})
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNAPRuleType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}
	var createModel networkAccessPolicyRuleModel
	if diags := createResp.State.Get(context.Background(), &createModel); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}
	index := strconv.FormatInt(createModel.Index.ValueInt64(), 10)

	// ImportState using composite ID "policy_name/index".
	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNAPRuleType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "default/" + index}, importResp)
	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState: %s", importResp.Diagnostics)
	}

	// Read to populate full state.
	readResp := &resource.ReadResponse{State: importResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: importResp.State}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read post-import: %s", readResp.Diagnostics)
	}
	var importedModel networkAccessPolicyRuleModel
	if diags := readResp.State.Get(context.Background(), &importedModel); diags.HasError() {
		t.Fatalf("Get imported state: %s", diags)
	}

	// Verify 0-diff.
	if importedModel.PolicyName.ValueString() != createModel.PolicyName.ValueString() {
		t.Errorf("policy_name mismatch: create=%s import=%s", createModel.PolicyName.ValueString(), importedModel.PolicyName.ValueString())
	}
	if importedModel.Client.ValueString() != createModel.Client.ValueString() {
		t.Errorf("client mismatch: create=%s import=%s", createModel.Client.ValueString(), importedModel.Client.ValueString())
	}
	if importedModel.Effect.ValueString() != createModel.Effect.ValueString() {
		t.Errorf("effect mismatch: create=%s import=%s", createModel.Effect.ValueString(), importedModel.Effect.ValueString())
	}
}

// TestUnit_NAPRule_PlanModifiers verifies all RequiresReplace and UseStateForUnknown
// plan modifiers in the network_access_policy_rule resource schema.
func TestUnit_NAPRule_PlanModifiers(t *testing.T) {
	s := napRuleResourceSchema(t).Schema

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
