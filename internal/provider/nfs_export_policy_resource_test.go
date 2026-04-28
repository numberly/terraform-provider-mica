package provider

import (
	"context"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/numberly/terraform-provider-mica/internal/client"
	"github.com/numberly/terraform-provider-mica/internal/testmock"
	"github.com/numberly/terraform-provider-mica/internal/testmock/handlers"
)

// ---- helpers ----------------------------------------------------------------

// newTestNFSPolicyResource creates an nfsExportPolicyResource wired to the given mock server.
func newTestNFSPolicyResource(t *testing.T, ms *testmock.MockServer) *nfsExportPolicyResource {
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
	return &nfsExportPolicyResource{client: c}
}

// nfsPolicyResourceSchema returns the parsed schema for the NFS export policy resource.
func nfsPolicyResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &nfsExportPolicyResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

// buildNfsExportPolicyType returns the tftypes.Object for the NFS export policy resource.
func buildNfsExportPolicyType() tftypes.Object {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":          tftypes.String,
		"name":        tftypes.String,
		"enabled":     tftypes.Bool,
		"is_local":    tftypes.Bool,
		"policy_type": tftypes.String,
		"version":     tftypes.String,
		"timeouts":    timeoutsType,
	}}
}

// nullNFSPolicyConfig returns a base config map with all attributes null.
func nullNFSPolicyConfig() map[string]tftypes.Value {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return map[string]tftypes.Value{
		"id":          tftypes.NewValue(tftypes.String, nil),
		"name":        tftypes.NewValue(tftypes.String, nil),
		"enabled":     tftypes.NewValue(tftypes.Bool, nil),
		"is_local":    tftypes.NewValue(tftypes.Bool, nil),
		"policy_type": tftypes.NewValue(tftypes.String, nil),
		"version":     tftypes.NewValue(tftypes.String, nil),
		"timeouts":    tftypes.NewValue(timeoutsType, nil),
	}
}

// nfsPolicyPlanWithName returns a tfsdk.Plan with the given policy name.
func nfsPolicyPlanWithName(t *testing.T, name string) tfsdk.Plan {
	t.Helper()
	s := nfsPolicyResourceSchema(t).Schema
	cfg := nullNFSPolicyConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, name)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildNfsExportPolicyType(), cfg),
		Schema: s,
	}
}

// nfsPolicyPlanWithNameAndEnabled returns a tfsdk.Plan with name and enabled flag.
func nfsPolicyPlanWithNameAndEnabled(t *testing.T, name string, enabled bool) tfsdk.Plan {
	t.Helper()
	s := nfsPolicyResourceSchema(t).Schema
	cfg := nullNFSPolicyConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, name)
	cfg["enabled"] = tftypes.NewValue(tftypes.Bool, enabled)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildNfsExportPolicyType(), cfg),
		Schema: s,
	}
}

// ---- tests ------------------------------------------------------------------

// TestNfsExportPolicyResource_Create verifies Create populates ID, name, and enabled.
func TestUnit_NfsExportPolicyResource_Create(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterNfsExportPolicyHandlers(ms.Mux)

	r := newTestNFSPolicyResource(t, ms)
	s := nfsPolicyResourceSchema(t).Schema

	plan := nfsPolicyPlanWithNameAndEnabled(t, "test-policy", true)
	resp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNfsExportPolicyType(), nil), Schema: s},
	}

	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Create returned error: %s", resp.Diagnostics)
	}

	var model nfsExportPolicyModel
	if diags := resp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected non-empty ID after Create")
	}
	if model.Name.ValueString() != "test-policy" {
		t.Errorf("expected name=test-policy, got %s", model.Name.ValueString())
	}
	if !model.Enabled.ValueBool() {
		t.Error("expected enabled=true after Create")
	}
}

// TestNfsExportPolicyResource_Update verifies PATCH updates enabled flag and supports rename.
func TestUnit_NfsExportPolicyResource_Update(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterNfsExportPolicyHandlers(ms.Mux)

	r := newTestNFSPolicyResource(t, ms)
	s := nfsPolicyResourceSchema(t).Schema

	// Create first.
	createPlan := nfsPolicyPlanWithNameAndEnabled(t, "update-policy", true)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNfsExportPolicyType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Update enabled=false.
	newPlan := nfsPolicyPlanWithNameAndEnabled(t, "update-policy", false)
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNfsExportPolicyType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  newPlan,
		State: createResp.State,
	}, updateResp)

	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update returned error: %s", updateResp.Diagnostics)
	}

	var model nfsExportPolicyModel
	if diags := updateResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}
	if model.Enabled.ValueBool() {
		t.Error("expected enabled=false after update")
	}

	// Now rename the policy in-place.
	renamePlan := nfsPolicyPlanWithNameAndEnabled(t, "update-policy-renamed", false)
	renameResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNfsExportPolicyType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  renamePlan,
		State: updateResp.State,
	}, renameResp)

	if renameResp.Diagnostics.HasError() {
		t.Fatalf("Rename update returned error: %s", renameResp.Diagnostics)
	}

	var renamedModel nfsExportPolicyModel
	if diags := renameResp.State.Get(context.Background(), &renamedModel); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}
	if renamedModel.Name.ValueString() != "update-policy-renamed" {
		t.Errorf("expected name=update-policy-renamed after rename, got %s", renamedModel.Name.ValueString())
	}
}

// TestNfsExportPolicyResource_Delete verifies DELETE removes the policy.
func TestUnit_NfsExportPolicyResource_Delete(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterNfsExportPolicyHandlers(ms.Mux)
	// Register file-systems handler for the delete-guard (ListNfsExportPolicyMembers).
	handlers.RegisterFileSystemHandlers(ms.Mux)

	r := newTestNFSPolicyResource(t, ms)
	s := nfsPolicyResourceSchema(t).Schema

	// Create first.
	plan := nfsPolicyPlanWithName(t, "delete-policy")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNfsExportPolicyType(), nil), Schema: s},
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

	// Verify policy is gone.
	_, err := r.client.GetNfsExportPolicy(context.Background(), "delete-policy")
	if err == nil || !client.IsNotFound(err) {
		t.Errorf("expected policy to be deleted, got: %v", err)
	}
}

// TestNfsExportPolicyResource_Import verifies ImportState populates all attributes.
func TestUnit_NfsExportPolicyResource_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterNfsExportPolicyHandlers(ms.Mux)

	r := newTestNFSPolicyResource(t, ms)
	s := nfsPolicyResourceSchema(t).Schema

	// Create first.
	plan := nfsPolicyPlanWithName(t, "import-policy")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNfsExportPolicyType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Import by name.
	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNfsExportPolicyType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "import-policy"}, importResp)

	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState returned error: %s", importResp.Diagnostics)
	}

	var model nfsExportPolicyModel
	if diags := importResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Name.ValueString() != "import-policy" {
		t.Errorf("expected name=import-policy after import, got %s", model.Name.ValueString())
	}
	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected ID to be populated after import")
	}
}

// TestUnit_NfsExportPolicy_Lifecycle exercises the full Create->Read->Update->Read->Delete sequence.
func TestUnit_Unit_NfsExportPolicy_Lifecycle(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterNfsExportPolicyHandlers(ms.Mux)
	handlers.RegisterFileSystemHandlers(ms.Mux)

	r := newTestNFSPolicyResource(t, ms)
	s := nfsPolicyResourceSchema(t).Schema

	// Step 1: Create.
	createPlan := nfsPolicyPlanWithNameAndEnabled(t, "lifecycle-nfs-policy", true)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNfsExportPolicyType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}
	var createModel nfsExportPolicyModel
	if diags := createResp.State.Get(context.Background(), &createModel); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}
	if createModel.Name.ValueString() != "lifecycle-nfs-policy" {
		t.Errorf("Create: expected name=lifecycle-nfs-policy, got %s", createModel.Name.ValueString())
	}
	if !createModel.Enabled.ValueBool() {
		t.Error("Create: expected enabled=true")
	}

	// Step 2: Read post-create.
	readResp1 := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp1)
	if readResp1.Diagnostics.HasError() {
		t.Fatalf("Read post-create: %s", readResp1.Diagnostics)
	}
	var readModel1 nfsExportPolicyModel
	if diags := readResp1.State.Get(context.Background(), &readModel1); diags.HasError() {
		t.Fatalf("Get read1 state: %s", diags)
	}
	if !readModel1.Enabled.ValueBool() {
		t.Error("Read1: expected enabled=true")
	}

	// Step 3: Update enabled=false.
	updatePlan := nfsPolicyPlanWithNameAndEnabled(t, "lifecycle-nfs-policy", false)
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNfsExportPolicyType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlan,
		State: readResp1.State,
	}, updateResp)
	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update: %s", updateResp.Diagnostics)
	}
	var updateModel nfsExportPolicyModel
	if diags := updateResp.State.Get(context.Background(), &updateModel); diags.HasError() {
		t.Fatalf("Get update state: %s", diags)
	}
	if updateModel.Enabled.ValueBool() {
		t.Error("Update: expected enabled=false")
	}

	// Step 4: Read post-update.
	readResp2 := &resource.ReadResponse{State: updateResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: updateResp.State}, readResp2)
	if readResp2.Diagnostics.HasError() {
		t.Fatalf("Read post-update: %s", readResp2.Diagnostics)
	}
	var readModel2 nfsExportPolicyModel
	if diags := readResp2.State.Get(context.Background(), &readModel2); diags.HasError() {
		t.Fatalf("Get read2 state: %s", diags)
	}
	if readModel2.Enabled.ValueBool() {
		t.Error("Read2: expected enabled=false")
	}

	// Step 5: Delete.
	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: readResp2.State}, deleteResp)
	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete: %s", deleteResp.Diagnostics)
	}
	_, err := r.client.GetNfsExportPolicy(context.Background(), "lifecycle-nfs-policy")
	if err == nil || !client.IsNotFound(err) {
		t.Errorf("expected policy to be deleted, got: %v", err)
	}
}

// TestUnit_NfsExportPolicy_ImportIdempotency verifies ImportState->Read produces state matching original Create.
func TestUnit_Unit_NfsExportPolicy_ImportIdempotency(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterNfsExportPolicyHandlers(ms.Mux)

	r := newTestNFSPolicyResource(t, ms)
	s := nfsPolicyResourceSchema(t).Schema

	// Create.
	createPlan := nfsPolicyPlanWithNameAndEnabled(t, "idempotent-nfs-policy", true)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNfsExportPolicyType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}
	var createModel nfsExportPolicyModel
	if diags := createResp.State.Get(context.Background(), &createModel); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}

	// ImportState.
	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNfsExportPolicyType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "idempotent-nfs-policy"}, importResp)
	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState: %s", importResp.Diagnostics)
	}

	// Read to populate full state.
	readResp := &resource.ReadResponse{State: importResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: importResp.State}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read post-import: %s", readResp.Diagnostics)
	}
	var importedModel nfsExportPolicyModel
	if diags := readResp.State.Get(context.Background(), &importedModel); diags.HasError() {
		t.Fatalf("Get imported state: %s", diags)
	}

	// Verify 0-diff.
	if importedModel.Name.ValueString() != createModel.Name.ValueString() {
		t.Errorf("name mismatch: create=%s import=%s", createModel.Name.ValueString(), importedModel.Name.ValueString())
	}
	if importedModel.Enabled.ValueBool() != createModel.Enabled.ValueBool() {
		t.Errorf("enabled mismatch: create=%v import=%v", createModel.Enabled.ValueBool(), importedModel.Enabled.ValueBool())
	}
	if importedModel.ID.ValueString() != createModel.ID.ValueString() {
		t.Errorf("id mismatch: create=%s import=%s", createModel.ID.ValueString(), importedModel.ID.ValueString())
	}
}

// ---- data source tests -------------------------------------------------------

// newTestNFSPolicyDataSource creates an nfsExportPolicyDataSource wired to the given mock server.
func newTestNFSPolicyDataSource(t *testing.T, ms *testmock.MockServer) *nfsExportPolicyDataSource {
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
	return &nfsExportPolicyDataSource{client: c}
}

// nfsPolicyDataSourceSchema returns the schema for the NFS export policy data source.
func nfsPolicyDataSourceSchema(t *testing.T) datasource.SchemaResponse {
	t.Helper()
	d := &nfsExportPolicyDataSource{}
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	return resp
}

// buildNfsExportPolicyDSType returns the tftypes.Object for the NFS export policy data source.
func buildNfsExportPolicyDSType() tftypes.Object {
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":          tftypes.String,
		"name":        tftypes.String,
		"enabled":     tftypes.Bool,
		"is_local":    tftypes.Bool,
		"policy_type": tftypes.String,
		"version":     tftypes.String,
	}}
}

// nullNFSPolicyDSConfig returns a base config map with all data source attributes null.
func nullNFSPolicyDSConfig() map[string]tftypes.Value {
	return map[string]tftypes.Value{
		"id":          tftypes.NewValue(tftypes.String, nil),
		"name":        tftypes.NewValue(tftypes.String, nil),
		"enabled":     tftypes.NewValue(tftypes.Bool, nil),
		"is_local":    tftypes.NewValue(tftypes.Bool, nil),
		"policy_type": tftypes.NewValue(tftypes.String, nil),
		"version":     tftypes.NewValue(tftypes.String, nil),
	}
}

// TestNfsExportPolicyDataSource verifies data source reads policy by name and returns all attributes.
func TestUnit_NfsExportPolicyDataSource(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterNfsExportPolicyHandlers(ms.Mux)

	// Create a policy via the resource client so the data source can find it.
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
	_, err = c.PostNfsExportPolicy(context.Background(), "ds-test-policy", client.NfsExportPolicyPost{
		Enabled: &enabled,
	})
	if err != nil {
		t.Fatalf("PostNfsExportPolicy: %v", err)
	}

	d := newTestNFSPolicyDataSource(t, ms)
	s := nfsPolicyDataSourceSchema(t).Schema

	cfg := nullNFSPolicyDSConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "ds-test-policy")

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNfsExportPolicyDSType(), nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(buildNfsExportPolicyDSType(), cfg), Schema: s},
	}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("DataSource Read returned error: %s", readResp.Diagnostics)
	}

	var model nfsExportPolicyDataSourceModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Name.ValueString() != "ds-test-policy" {
		t.Errorf("expected name=ds-test-policy, got %s", model.Name.ValueString())
	}
	if !model.Enabled.ValueBool() {
		t.Error("expected enabled=true")
	}
	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected ID to be populated")
	}
}

// TestUnit_NfsExportPolicy_Create_Conflict verifies that a 409 Conflict on POST
// produces an error diagnostic with a meaningful message.
func TestUnit_Unit_NfsExportPolicy_Create_Conflict(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	ms.RegisterHandler("/api/2.22/nfs-export-policies", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handlers.WriteJSONError(w, http.StatusConflict, "Policy with the given name already exists.")
			return
		}
		handlers.WriteJSONListResponse(w, http.StatusOK, []client.NfsExportPolicy{})
	})

	r := newTestNFSPolicyResource(t, ms)
	s := nfsPolicyResourceSchema(t).Schema

	plan := nfsPolicyPlanWithName(t, "conflict-policy")
	resp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNfsExportPolicyType(), nil), Schema: s},
	}

	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("expected Create to produce an error diagnostic on 409 Conflict, got none")
	}
}

// TestUnit_NfsExportPolicy_Read_NotFound verifies that a not-found response (empty items)
// during Read removes the resource from Terraform state without an error diagnostic.
func TestUnit_Unit_NfsExportPolicy_Read_NotFound(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	ms.RegisterHandler("/api/2.22/nfs-export-policies", func(w http.ResponseWriter, r *http.Request) {
		handlers.WriteJSONListResponse(w, http.StatusOK, []client.NfsExportPolicy{})
	})

	r := newTestNFSPolicyResource(t, ms)
	s := nfsPolicyResourceSchema(t).Schema

	cfg := nullNFSPolicyConfig()
	cfg["id"] = tftypes.NewValue(tftypes.String, "nfs-gone-id")
	cfg["name"] = tftypes.NewValue(tftypes.String, "gone-policy")
	state := tfsdk.State{Raw: tftypes.NewValue(buildNfsExportPolicyType(), cfg), Schema: s}

	readResp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), resource.ReadRequest{State: state}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}
	if !readResp.State.Raw.IsNull() {
		t.Error("expected state to be removed (null) when NFS export policy not found")
	}
}

// TestUnit_NFSPolicy_PlanModifiers verifies all UseStateForUnknown plan modifiers
// in the nfs_export_policy resource schema.
func TestUnit_Unit_NFSPolicy_PlanModifiers(t *testing.T) {
	s := nfsPolicyResourceSchema(t).Schema

	// id — UseStateForUnknown
	idAttr, ok := s.Attributes["id"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("id attribute not found or wrong type")
	}
	if len(idAttr.PlanModifiers) == 0 {
		t.Error("expected UseStateForUnknown plan modifier on id attribute")
	}

	// is_local — UseStateForUnknown
	ilAttr, ok := s.Attributes["is_local"].(resschema.BoolAttribute)
	if !ok {
		t.Fatal("is_local attribute not found or wrong type")
	}
	if len(ilAttr.PlanModifiers) == 0 {
		t.Error("expected UseStateForUnknown plan modifier on is_local attribute")
	}

	// policy_type — UseStateForUnknown
	ptAttr, ok := s.Attributes["policy_type"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("policy_type attribute not found or wrong type")
	}
	if len(ptAttr.PlanModifiers) == 0 {
		t.Error("expected UseStateForUnknown plan modifier on policy_type attribute")
	}
}
