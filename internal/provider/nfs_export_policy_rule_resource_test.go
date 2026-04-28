package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
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

// newTestNFSRuleResource creates an nfsExportPolicyRuleResource wired to the given mock server.
func newTestNFSRuleResource(t *testing.T, ms *testmock.MockServer) *nfsExportPolicyRuleResource {
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

// buildNfsExportPolicyRuleType returns the tftypes.Object for the NFS export policy rule resource.
func buildNfsExportPolicyRuleType() tftypes.Object {
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
		Raw:    tftypes.NewValue(buildNfsExportPolicyRuleType(), cfg),
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
func TestUnit_NfsExportPolicyRuleResource_Create(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterNfsExportPolicyHandlers(ms.Mux)

	r := newTestNFSRuleResource(t, ms)
	s := nfsRuleResourceSchema(t).Schema

	createTestPolicy(t, r.client, "rule-create-policy")

	plan := nfsRulePlan(t, "rule-create-policy", "root-squash", "*", "rw")
	resp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNfsExportPolicyRuleType(), nil), Schema: s},
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
func TestUnit_NfsExportPolicyRuleResource_Update(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterNfsExportPolicyHandlers(ms.Mux)

	r := newTestNFSRuleResource(t, ms)
	s := nfsRuleResourceSchema(t).Schema

	createTestPolicy(t, r.client, "rule-update-policy")

	// Create rule first.
	createPlan := nfsRulePlan(t, "rule-update-policy", "root-squash", "*", "rw")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNfsExportPolicyRuleType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Update client to "10.0.0.0/8".
	updatePlan := nfsRulePlan(t, "rule-update-policy", "root-squash", "10.0.0.0/8", "rw")
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNfsExportPolicyRuleType(), nil), Schema: s},
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
func TestUnit_NfsExportPolicyRuleResource_Delete(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterNfsExportPolicyHandlers(ms.Mux)

	r := newTestNFSRuleResource(t, ms)
	s := nfsRuleResourceSchema(t).Schema

	createTestPolicy(t, r.client, "rule-delete-policy")

	// Create rule first.
	createPlan := nfsRulePlan(t, "rule-delete-policy", "root-squash", "*", "rw")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNfsExportPolicyRuleType(), nil), Schema: s},
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
func TestUnit_NfsExportPolicyRuleResource_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterNfsExportPolicyHandlers(ms.Mux)

	r := newTestNFSRuleResource(t, ms)
	s := nfsRuleResourceSchema(t).Schema

	createTestPolicy(t, r.client, "rule-import-policy")

	// Create rule.
	createPlan := nfsRulePlan(t, "rule-import-policy", "root-squash", "*", "rw")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNfsExportPolicyRuleType(), nil), Schema: s},
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
		State: tfsdk.State{Raw: tftypes.NewValue(buildNfsExportPolicyRuleType(), nil), Schema: s},
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

// TestUnit_NfsExportPolicyRule_Lifecycle exercises the full Create->Read->Update->Read->Delete sequence.
func TestUnit_Unit_NfsExportPolicyRule_Lifecycle(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterNfsExportPolicyHandlers(ms.Mux)

	r := newTestNFSRuleResource(t, ms)
	s := nfsRuleResourceSchema(t).Schema

	createTestPolicy(t, r.client, "lifecycle-nfs-rule-policy")

	// Step 1: Create.
	createPlan := nfsRulePlan(t, "lifecycle-nfs-rule-policy", "root-squash", "*", "rw")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNfsExportPolicyRuleType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}
	var createModel nfsExportPolicyRuleModel
	if diags := createResp.State.Get(context.Background(), &createModel); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}
	if createModel.Access.ValueString() != "root-squash" {
		t.Errorf("Create: expected access=root-squash, got %s", createModel.Access.ValueString())
	}

	// Step 2: Read post-create.
	readResp1 := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp1)
	if readResp1.Diagnostics.HasError() {
		t.Fatalf("Read post-create: %s", readResp1.Diagnostics)
	}
	var readModel1 nfsExportPolicyRuleModel
	if diags := readResp1.State.Get(context.Background(), &readModel1); diags.HasError() {
		t.Fatalf("Get read1 state: %s", diags)
	}
	if readModel1.Client.ValueString() != "*" {
		t.Errorf("Read1: expected client=*, got %s", readModel1.Client.ValueString())
	}

	// Step 3: Update client to specific subnet.
	updatePlan := nfsRulePlan(t, "lifecycle-nfs-rule-policy", "root-squash", "192.168.0.0/16", "rw")
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNfsExportPolicyRuleType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlan,
		State: readResp1.State,
	}, updateResp)
	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update: %s", updateResp.Diagnostics)
	}
	var updateModel nfsExportPolicyRuleModel
	if diags := updateResp.State.Get(context.Background(), &updateModel); diags.HasError() {
		t.Fatalf("Get update state: %s", diags)
	}
	if updateModel.Client.ValueString() != "192.168.0.0/16" {
		t.Errorf("Update: expected client=192.168.0.0/16, got %s", updateModel.Client.ValueString())
	}

	// Step 4: Read post-update.
	readResp2 := &resource.ReadResponse{State: updateResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: updateResp.State}, readResp2)
	if readResp2.Diagnostics.HasError() {
		t.Fatalf("Read post-update: %s", readResp2.Diagnostics)
	}
	var readModel2 nfsExportPolicyRuleModel
	if diags := readResp2.State.Get(context.Background(), &readModel2); diags.HasError() {
		t.Fatalf("Get read2 state: %s", diags)
	}
	if readModel2.Client.ValueString() != "192.168.0.0/16" {
		t.Errorf("Read2: expected client=192.168.0.0/16, got %s", readModel2.Client.ValueString())
	}

	// Step 5: Delete.
	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: readResp2.State}, deleteResp)
	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete: %s", deleteResp.Diagnostics)
	}
}

// TestUnit_NfsExportPolicyRule_ImportIdempotency verifies ImportState->Read produces state matching original Create.
func TestUnit_Unit_NfsExportPolicyRule_ImportIdempotency(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterNfsExportPolicyHandlers(ms.Mux)

	r := newTestNFSRuleResource(t, ms)
	s := nfsRuleResourceSchema(t).Schema

	createTestPolicy(t, r.client, "idempotent-nfs-rule-policy")

	// Create.
	createPlan := nfsRulePlan(t, "idempotent-nfs-rule-policy", "root-squash", "*", "rw")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNfsExportPolicyRuleType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}
	var createModel nfsExportPolicyRuleModel
	if diags := createResp.State.Get(context.Background(), &createModel); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}
	index := strconv.FormatInt(createModel.Index.ValueInt64(), 10)

	// ImportState using composite ID.
	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNfsExportPolicyRuleType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "idempotent-nfs-rule-policy/" + index}, importResp)
	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState: %s", importResp.Diagnostics)
	}

	// Read to populate full state.
	readResp := &resource.ReadResponse{State: importResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: importResp.State}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read post-import: %s", readResp.Diagnostics)
	}
	var importedModel nfsExportPolicyRuleModel
	if diags := readResp.State.Get(context.Background(), &importedModel); diags.HasError() {
		t.Fatalf("Get imported state: %s", diags)
	}

	// Verify 0-diff.
	if importedModel.PolicyName.ValueString() != createModel.PolicyName.ValueString() {
		t.Errorf("policy_name mismatch: create=%s import=%s", createModel.PolicyName.ValueString(), importedModel.PolicyName.ValueString())
	}
	if importedModel.Access.ValueString() != createModel.Access.ValueString() {
		t.Errorf("access mismatch: create=%s import=%s", createModel.Access.ValueString(), importedModel.Access.ValueString())
	}
	if importedModel.Client.ValueString() != createModel.Client.ValueString() {
		t.Errorf("client mismatch: create=%s import=%s", createModel.Client.ValueString(), importedModel.Client.ValueString())
	}
}

// TestUnit_NFSRule_PlanModifiers verifies all RequiresReplace and UseStateForUnknown
// plan modifiers in the nfs_export_policy_rule resource schema.
func TestUnit_Unit_NFSRule_PlanModifiers(t *testing.T) {
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

// nfsRuleBodyCaptor records the last PATCH body for the rules endpoint.
type nfsRuleBodyCaptor struct {
	inner     http.Handler
	lastPATCH []byte
}

func (c *nfsRuleBodyCaptor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/api/2.22/nfs-export-policies/rules" && r.Method == http.MethodPatch {
		buf, _ := io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewReader(buf))
		c.lastPATCH = buf
	}
	c.inner.ServeHTTP(w, r)
}

// newNFSRuleCaptorClient builds a FlashBladeClient pointed at a captor-wrapped mock mux.
func newNFSRuleCaptorClient(t *testing.T, captor *nfsRuleBodyCaptor) (*client.FlashBladeClient, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(captor)
	c, err := client.NewClient(context.Background(), client.Config{
		Endpoint:           srv.URL,
		APIToken:           "test-token",
		InsecureSkipVerify: true,
		MaxRetries:         1,
	})
	if err != nil {
		srv.Close()
		t.Fatalf("NewClient: %v", err)
	}
	return c, srv
}

// TestUnit_NfsExportPolicyRule_Patch_Security_Clear verifies that transitioning
// the security list from ["sys"] to [] emits "security":[] in the PATCH body
// (not omitted), and that the mock-stored rule ends up with an empty security
// list. Regression guard for R-009.
func TestUnit_NfsExportPolicyRule_Patch_Security_Clear(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterNfsExportPolicyHandlers(ms.Mux)

	captor := &nfsRuleBodyCaptor{inner: ms.Mux}
	c, captureSrv := newNFSRuleCaptorClient(t, captor)
	defer captureSrv.Close()

	// Create policy + rule via the client (exercises the POST path, gets a real
	// server-assigned rule name/id).
	createTestPolicy(t, c, "clear-sec-policy")
	created, err := c.PostNfsExportPolicyRule(context.Background(), "clear-sec-policy", client.NfsExportPolicyRulePost{
		Access:     "root-squash",
		Client:     "*",
		Permission: "rw",
		Security:   []string{"sys"},
	})
	if err != nil {
		t.Fatalf("PostNfsExportPolicyRule: %v", err)
	}
	ruleName := created.Name

	r := &nfsExportPolicyRuleResource{client: c}
	s := nfsRuleResourceSchema(t).Schema

	// Build state: fully populated (simulating state after a prior read).
	stateCfg := nullNFSRuleConfig()
	stateCfg["id"] = tftypes.NewValue(tftypes.String, created.ID)
	stateCfg["policy_name"] = tftypes.NewValue(tftypes.String, "clear-sec-policy")
	stateCfg["name"] = tftypes.NewValue(tftypes.String, ruleName)
	stateCfg["index"] = tftypes.NewValue(tftypes.Number, created.Index)
	stateCfg["policy_version"] = tftypes.NewValue(tftypes.String, "v1")
	stateCfg["access"] = tftypes.NewValue(tftypes.String, "root-squash")
	stateCfg["client"] = tftypes.NewValue(tftypes.String, "*")
	stateCfg["permission"] = tftypes.NewValue(tftypes.String, "rw")
	stateCfg["anonuid"] = tftypes.NewValue(tftypes.Number, 0)
	stateCfg["anongid"] = tftypes.NewValue(tftypes.Number, 0)
	stateCfg["atime"] = tftypes.NewValue(tftypes.Bool, true)
	stateCfg["fileid_32bit"] = tftypes.NewValue(tftypes.Bool, false)
	stateCfg["secure"] = tftypes.NewValue(tftypes.Bool, false)
	stateCfg["security"] = tftypes.NewValue(
		tftypes.List{ElementType: tftypes.String},
		[]tftypes.Value{tftypes.NewValue(tftypes.String, "sys")},
	)
	stateCfg["required_transport_security"] = tftypes.NewValue(tftypes.String, "")

	priorState := tfsdk.State{
		Raw:    tftypes.NewValue(buildNfsExportPolicyRuleType(), stateCfg),
		Schema: s,
	}

	// Build plan: same as state but security = [] (non-null empty list).
	planCfg := nullNFSRuleConfig()
	for k, v := range stateCfg {
		planCfg[k] = v
	}
	planCfg["security"] = tftypes.NewValue(
		tftypes.List{ElementType: tftypes.String},
		[]tftypes.Value{},
	)

	plan := tfsdk.Plan{
		Raw:    tftypes.NewValue(buildNfsExportPolicyRuleType(), planCfg),
		Schema: s,
	}

	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNfsExportPolicyRuleType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{Plan: plan, State: priorState}, updateResp)
	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update returned error: %s", updateResp.Diagnostics)
	}

	// Assertion 1: captured PATCH body contains "security":[] (not omitted, not null).
	if captor.lastPATCH == nil {
		t.Fatal("expected PATCH body to be captured")
	}
	var rawBody map[string]json.RawMessage
	if err := json.Unmarshal(captor.lastPATCH, &rawBody); err != nil {
		t.Fatalf("decode PATCH body: %v", err)
	}
	secRaw, ok := rawBody["security"]
	if !ok {
		t.Fatalf("expected 'security' key in PATCH body, got: %s", string(captor.lastPATCH))
	}
	if string(secRaw) != "[]" {
		t.Errorf("expected security=[] in PATCH body, got %s", string(secRaw))
	}

	// Assertion 2: mock-stored rule now has empty Security slice.
	got, err := c.GetNfsExportPolicyRuleByName(context.Background(), "clear-sec-policy", ruleName)
	if err != nil {
		t.Fatalf("GetNfsExportPolicyRuleByName: %v", err)
	}
	if len(got.Security) != 0 {
		t.Errorf("expected stored Security to be empty, got %v", got.Security)
	}

	// Assertion 3: returned state has non-null empty security list.
	var model nfsExportPolicyRuleModel
	if diags := updateResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}
	if model.Security.IsNull() {
		t.Error("expected Security in state to be non-null (empty list)")
	}
	if len(model.Security.Elements()) != 0 {
		t.Errorf("expected Security in state to be empty, got %d elements", len(model.Security.Elements()))
	}
}

// TestUnit_NfsExportPolicyRuleResource_StateUpgrade_V0toV1 verifies that the
// v0->v1 upgrader is an identity: every attribute present in v0 state lands
// in v1 state unchanged. Regression guard for R-009 schema bump.
func TestUnit_NfsExportPolicyRuleResource_StateUpgrade_V0toV1(t *testing.T) {
	r := &nfsExportPolicyRuleResource{}
	upgraders := r.UpgradeState(context.Background())

	upgrader, ok := upgraders[0]
	if !ok {
		t.Fatal("expected v0->v1 upgrader at key 0")
	}
	if upgrader.PriorSchema == nil {
		t.Fatal("expected PriorSchema to be set for v0->v1 upgrader")
	}

	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	stringList := tftypes.List{ElementType: tftypes.String}

	v0Type := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
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
		"security":                    stringList,
		"required_transport_security": tftypes.String,
		"timeouts":                    timeoutsType,
	}}

	v0Val := tftypes.NewValue(v0Type, map[string]tftypes.Value{
		"id":                          tftypes.NewValue(tftypes.String, "rule-001"),
		"policy_name":                 tftypes.NewValue(tftypes.String, "my-policy"),
		"name":                        tftypes.NewValue(tftypes.String, "rule-abc"),
		"index":                       tftypes.NewValue(tftypes.Number, 3),
		"policy_version":              tftypes.NewValue(tftypes.String, "v7"),
		"access":                      tftypes.NewValue(tftypes.String, "root-squash"),
		"client":                      tftypes.NewValue(tftypes.String, "10.0.0.0/8"),
		"permission":                  tftypes.NewValue(tftypes.String, "rw"),
		"anonuid":                     tftypes.NewValue(tftypes.Number, 65534),
		"anongid":                     tftypes.NewValue(tftypes.Number, 65534),
		"atime":                       tftypes.NewValue(tftypes.Bool, true),
		"fileid_32bit":                tftypes.NewValue(tftypes.Bool, false),
		"secure":                      tftypes.NewValue(tftypes.Bool, true),
		"security":                    tftypes.NewValue(stringList, []tftypes.Value{tftypes.NewValue(tftypes.String, "sys")}),
		"required_transport_security": tftypes.NewValue(tftypes.String, "krb5"),
		"timeouts":                    tftypes.NewValue(timeoutsType, nil),
	})

	priorState := tfsdk.State{
		Raw:    v0Val,
		Schema: *upgrader.PriorSchema,
	}

	currentSchema := nfsRuleResourceSchema(t).Schema
	resp := &resource.UpgradeStateResponse{
		State: tfsdk.State{
			Raw:    tftypes.NewValue(buildNfsExportPolicyRuleType(), nil),
			Schema: currentSchema,
		},
	}
	req := resource.UpgradeStateRequest{State: &priorState}

	upgrader.StateUpgrader(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("StateUpgrader returned error: %s", resp.Diagnostics)
	}

	var model nfsExportPolicyRuleModel
	if diags := resp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get upgraded state: %s", diags)
	}

	if model.ID.ValueString() != "rule-001" {
		t.Errorf("expected id=rule-001, got %s", model.ID.ValueString())
	}
	if model.PolicyName.ValueString() != "my-policy" {
		t.Errorf("expected policy_name=my-policy, got %s", model.PolicyName.ValueString())
	}
	if model.Name.ValueString() != "rule-abc" {
		t.Errorf("expected name=rule-abc, got %s", model.Name.ValueString())
	}
	if model.Index.ValueInt64() != 3 {
		t.Errorf("expected index=3, got %d", model.Index.ValueInt64())
	}
	if model.Access.ValueString() != "root-squash" {
		t.Errorf("expected access=root-squash, got %s", model.Access.ValueString())
	}
	if model.Client.ValueString() != "10.0.0.0/8" {
		t.Errorf("expected client=10.0.0.0/8, got %s", model.Client.ValueString())
	}
	if model.Permission.ValueString() != "rw" {
		t.Errorf("expected permission=rw, got %s", model.Permission.ValueString())
	}
	if model.Anonuid.ValueInt64() != 65534 {
		t.Errorf("expected anonuid=65534, got %d", model.Anonuid.ValueInt64())
	}
	if !model.Secure.ValueBool() {
		t.Error("expected secure=true")
	}
	if len(model.Security.Elements()) != 1 {
		t.Errorf("expected security list length=1, got %d", len(model.Security.Elements()))
	}
	if model.RequiredTransportSecurity.ValueString() != "krb5" {
		t.Errorf("expected required_transport_security=krb5, got %s", model.RequiredTransportSecurity.ValueString())
	}
}
