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
	"github.com/soulkyu/terraform-provider-flashblade/internal/client"
	"github.com/soulkyu/terraform-provider-flashblade/internal/testmock"
	"github.com/soulkyu/terraform-provider-flashblade/internal/testmock/handlers"
)

// ---- helpers ----------------------------------------------------------------

// newTestNFSRuleResource creates an nfsExportPolicyRuleResource wired to the given mock server.
func newTestNFSRuleResource(t *testing.T, ms *testmock.MockServer) *nfsExportPolicyRuleResource {
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
	return &nfsExportPolicyRuleResource{client: c}
}

// nfsRuleResourceSchema returns the parsed schema for the NFS export policy rule resource.
func nfsRuleResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &nfsExportPolicyRuleResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

// buildNFSRuleType returns the tftypes.Object for the NFS export policy rule resource.
func buildNFSRuleType() tftypes.Object {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":                          tftypes.String,
		"policy_name":                 tftypes.String,
		"name":                        tftypes.String,
		"index":                       tftypes.Number,
		"policy_version":              tftypes.String,
		"access":                      tftypes.String,
		"client":                      tftypes.String,
		"permission":                  tftypes.String,
		"anonuid":                     tftypes.Number,
		"anongid":                     tftypes.Number,
		"atime":                       tftypes.Bool,
		"fileid_32bit":                tftypes.Bool,
		"secure":                      tftypes.Bool,
		"security":                    tftypes.List{ElementType: tftypes.String},
		"required_transport_security": tftypes.String,
		"timeouts":                    timeoutsType,
	}}
}

// nullNFSRuleConfig returns a base config map with all attributes null.
func nullNFSRuleConfig() map[string]tftypes.Value {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return map[string]tftypes.Value{
		"id":                          tftypes.NewValue(tftypes.String, nil),
		"policy_name":                 tftypes.NewValue(tftypes.String, nil),
		"name":                        tftypes.NewValue(tftypes.String, nil),
		"index":                       tftypes.NewValue(tftypes.Number, nil),
		"policy_version":              tftypes.NewValue(tftypes.String, nil),
		"access":                      tftypes.NewValue(tftypes.String, nil),
		"client":                      tftypes.NewValue(tftypes.String, nil),
		"permission":                  tftypes.NewValue(tftypes.String, nil),
		"anonuid":                     tftypes.NewValue(tftypes.Number, nil),
		"anongid":                     tftypes.NewValue(tftypes.Number, nil),
		"atime":                       tftypes.NewValue(tftypes.Bool, nil),
		"fileid_32bit":                tftypes.NewValue(tftypes.Bool, nil),
		"secure":                      tftypes.NewValue(tftypes.Bool, nil),
		"security":                    tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"required_transport_security": tftypes.NewValue(tftypes.String, nil),
		"timeouts":                    tftypes.NewValue(timeoutsType, nil),
	}
}

// nfsRulePlan builds a tfsdk.Plan with the given policy_name and rule fields.
func nfsRulePlan(t *testing.T, policyName, access, clientStr, permission string) tfsdk.Plan {
	t.Helper()
	s := nfsRuleResourceSchema(t).Schema
	cfg := nullNFSRuleConfig()
	cfg["policy_name"] = tftypes.NewValue(tftypes.String, policyName)
	if access != "" {
		cfg["access"] = tftypes.NewValue(tftypes.String, access)
	}
	if clientStr != "" {
		cfg["client"] = tftypes.NewValue(tftypes.String, clientStr)
	}
	if permission != "" {
		cfg["permission"] = tftypes.NewValue(tftypes.String, permission)
	}
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildNFSRuleType(), cfg),
		Schema: s,
	}
}

// createTestPolicy is a helper that creates an NFS export policy via the client.
func createTestPolicy(t *testing.T, c *client.FlashBladeClient, name string) {
	t.Helper()
	enabled := true
	_, err := c.PostNfsExportPolicy(context.Background(), name, client.NfsExportPolicyPost{Enabled: &enabled})
	if err != nil {
		t.Fatalf("PostNfsExportPolicy(%q): %v", name, err)
	}
}

// ---- tests ------------------------------------------------------------------

// TestNfsExportPolicyRuleResource_Create verifies Create populates id, name, index, and rule fields.
func TestNfsExportPolicyRuleResource_Create(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterNfsExportPolicyHandlers(ms.Mux)

	r := newTestNFSRuleResource(t, ms)
	s := nfsRuleResourceSchema(t).Schema

	createTestPolicy(t, r.client, "rule-create-policy")

	plan := nfsRulePlan(t, "rule-create-policy", "root-squash", "*", "rw")
	resp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNFSRuleType(), nil), Schema: s},
	}

	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Create returned error: %s", resp.Diagnostics)
	}

	var model nfsExportPolicyRuleModel
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
	if model.PolicyName.ValueString() != "rule-create-policy" {
		t.Errorf("expected policy_name=rule-create-policy, got %s", model.PolicyName.ValueString())
	}
	if model.Access.ValueString() != "root-squash" {
		t.Errorf("expected access=root-squash, got %s", model.Access.ValueString())
	}
	if model.Client.ValueString() != "*" {
		t.Errorf("expected client=*, got %s", model.Client.ValueString())
	}
	if model.Permission.ValueString() != "rw" {
		t.Errorf("expected permission=rw, got %s", model.Permission.ValueString())
	}
}

// TestNfsExportPolicyRuleResource_Update verifies PATCH updates mutable rule fields.
func TestNfsExportPolicyRuleResource_Update(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterNfsExportPolicyHandlers(ms.Mux)

	r := newTestNFSRuleResource(t, ms)
	s := nfsRuleResourceSchema(t).Schema

	createTestPolicy(t, r.client, "rule-update-policy")

	// Create rule first.
	createPlan := nfsRulePlan(t, "rule-update-policy", "root-squash", "*", "rw")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNFSRuleType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Update client to "10.0.0.0/8".
	updatePlan := nfsRulePlan(t, "rule-update-policy", "root-squash", "10.0.0.0/8", "rw")
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNFSRuleType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlan,
		State: createResp.State,
	}, updateResp)

	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update returned error: %s", updateResp.Diagnostics)
	}

	var model nfsExportPolicyRuleModel
	if diags := updateResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}
	if model.Client.ValueString() != "10.0.0.0/8" {
		t.Errorf("expected client=10.0.0.0/8 after update, got %s", model.Client.ValueString())
	}
}

// TestNfsExportPolicyRuleResource_Delete verifies DELETE removes the rule.
func TestNfsExportPolicyRuleResource_Delete(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterNfsExportPolicyHandlers(ms.Mux)

	r := newTestNFSRuleResource(t, ms)
	s := nfsRuleResourceSchema(t).Schema

	createTestPolicy(t, r.client, "rule-delete-policy")

	// Create rule first.
	createPlan := nfsRulePlan(t, "rule-delete-policy", "root-squash", "*", "rw")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNFSRuleType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	var createdModel nfsExportPolicyRuleModel
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
	_, err := r.client.GetNfsExportPolicyRuleByName(context.Background(), "rule-delete-policy", ruleName)
	if err == nil || !client.IsNotFound(err) {
		t.Errorf("expected rule to be deleted, got: %v", err)
	}
}

// TestNfsExportPolicyRuleResource_Import verifies ImportState with composite ID "policy_name/index".
func TestNfsExportPolicyRuleResource_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterNfsExportPolicyHandlers(ms.Mux)

	r := newTestNFSRuleResource(t, ms)
	s := nfsRuleResourceSchema(t).Schema

	createTestPolicy(t, r.client, "rule-import-policy")

	// Create rule.
	createPlan := nfsRulePlan(t, "rule-import-policy", "root-squash", "*", "rw")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNFSRuleType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	var createdModel nfsExportPolicyRuleModel
	if diags := createResp.State.Get(context.Background(), &createdModel); diags.HasError() {
		t.Fatalf("Get created state: %s", diags)
	}
	index := strconv.FormatInt(createdModel.Index.ValueInt64(), 10)

	// Import using "policy_name/index" composite ID.
	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNFSRuleType(), nil), Schema: s},
	}
	importID := "rule-import-policy/" + index
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: importID}, importResp)

	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState returned error: %s", importResp.Diagnostics)
	}

	var model nfsExportPolicyRuleModel
	if diags := importResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.PolicyName.ValueString() != "rule-import-policy" {
		t.Errorf("expected policy_name=rule-import-policy after import, got %s", model.PolicyName.ValueString())
	}
	if model.Name.IsNull() || model.Name.ValueString() == "" {
		t.Error("expected server-assigned name to be populated after import")
	}
	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected ID to be populated after import")
	}
	if model.Access.ValueString() != "root-squash" {
		t.Errorf("expected access=root-squash after import, got %s", model.Access.ValueString())
	}
}

// TestUnit_NFSRule_PlanModifiers verifies all RequiresReplace and UseStateForUnknown
// plan modifiers in the nfs_export_policy_rule resource schema.
func TestUnit_NFSRule_PlanModifiers(t *testing.T) {
	s := nfsRuleResourceSchema(t).Schema

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
