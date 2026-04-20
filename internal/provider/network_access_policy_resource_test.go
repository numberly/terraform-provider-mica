package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/numberly/opentofu-provider-flashblade/internal/client"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock/handlers"
)

// ---- helpers ----------------------------------------------------------------

// newTestNAPResource creates a networkAccessPolicyResource wired to the given mock server.
func newTestNAPResource(t *testing.T, ms *testmock.MockServer) *networkAccessPolicyResource {
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
	return &networkAccessPolicyResource{client: c}
}

// napResourceSchema returns the parsed schema for the NAP resource.
func napResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &networkAccessPolicyResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

// buildNetworkAccessPolicyType returns the tftypes.Object for the NAP resource.
func buildNetworkAccessPolicyType() tftypes.Object {
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

// nullNAPConfig returns a base config map with all attributes null.
func nullNAPConfig() map[string]tftypes.Value {
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

// napPlanWithName returns a tfsdk.Plan for the NAP resource with the given name.
func napPlanWithName(t *testing.T, name string) tfsdk.Plan {
	t.Helper()
	s := napResourceSchema(t).Schema
	cfg := nullNAPConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, name)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildNetworkAccessPolicyType(), cfg),
		Schema: s,
	}
}

// napPlanWithNameAndEnabled returns a tfsdk.Plan for the NAP resource with name and enabled flag.
func napPlanWithNameAndEnabled(t *testing.T, name string, enabled bool) tfsdk.Plan {
	t.Helper()
	s := napResourceSchema(t).Schema
	cfg := nullNAPConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, name)
	cfg["enabled"] = tftypes.NewValue(tftypes.Bool, enabled)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildNetworkAccessPolicyType(), cfg),
		Schema: s,
	}
}

// ---- tests ------------------------------------------------------------------

// TestNetworkAccessPolicyResource_Create verifies Create (GET+PATCH) populates state from the pre-seeded "default" singleton.
func TestUnit_NetworkAccessPolicyResource_Create(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterNetworkAccessPolicyHandlers(ms.Mux)

	r := newTestNAPResource(t, ms)
	s := napResourceSchema(t).Schema

	// The mock pre-seeds a "default" policy. Create should adopt it via GET+PATCH.
	plan := napPlanWithNameAndEnabled(t, "default", true)
	resp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNetworkAccessPolicyType(), nil), Schema: s},
	}

	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Create returned error: %s", resp.Diagnostics)
	}

	var model networkAccessPolicyModel
	if diags := resp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected non-empty ID after Create")
	}
	if model.Name.ValueString() != "default" {
		t.Errorf("expected name=default, got %s", model.Name.ValueString())
	}
	if !model.Enabled.ValueBool() {
		t.Error("expected enabled=true after Create")
	}
}

// TestNetworkAccessPolicyResource_Create_NotFound verifies Create fails with a clear error for non-existent policies.
func TestUnit_NetworkAccessPolicyResource_Create_NotFound(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterNetworkAccessPolicyHandlers(ms.Mux)

	r := newTestNAPResource(t, ms)
	s := napResourceSchema(t).Schema

	// "nonexistent-policy" is not pre-seeded in the mock.
	plan := napPlanWithName(t, "nonexistent-policy")
	resp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNetworkAccessPolicyType(), nil), Schema: s},
	}

	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected Create to fail for non-existent policy, but got no error")
	}
}

// TestNetworkAccessPolicyResource_Update verifies PATCH updates enabled flag.
func TestUnit_NetworkAccessPolicyResource_Update(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterNetworkAccessPolicyHandlers(ms.Mux)

	r := newTestNAPResource(t, ms)
	s := napResourceSchema(t).Schema

	// Adopt the default singleton.
	createPlan := napPlanWithNameAndEnabled(t, "default", true)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNetworkAccessPolicyType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Update enabled=false.
	updatePlan := napPlanWithNameAndEnabled(t, "default", false)
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNetworkAccessPolicyType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlan,
		State: createResp.State,
	}, updateResp)

	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update returned error: %s", updateResp.Diagnostics)
	}

	var model networkAccessPolicyModel
	if diags := updateResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}
	if model.Enabled.ValueBool() {
		t.Error("expected enabled=false after update")
	}
}

// TestNetworkAccessPolicyResource_Delete verifies Delete resets singleton to disabled=false via PATCH.
// The policy should still exist on the array (it is a singleton).
func TestUnit_NetworkAccessPolicyResource_Delete(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterNetworkAccessPolicyHandlers(ms.Mux)

	r := newTestNAPResource(t, ms)
	s := napResourceSchema(t).Schema

	// Adopt the default singleton.
	plan := napPlanWithNameAndEnabled(t, "default", true)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNetworkAccessPolicyType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Delete (reset to disabled state).
	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: createResp.State}, deleteResp)

	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete returned error: %s", deleteResp.Diagnostics)
	}

	// Verify the policy still exists on the array (singleton — not actually deleted).
	policy, err := r.client.GetNetworkAccessPolicy(context.Background(), "default")
	if err != nil {
		t.Fatalf("expected policy to still exist after Delete (singleton): %v", err)
	}
	// Verify the policy was reset to disabled state.
	if policy.Enabled {
		t.Error("expected policy to be disabled after Delete (reset to disabled state)")
	}
}

// TestNetworkAccessPolicyResource_Import verifies ImportState populates all attributes.
func TestUnit_NetworkAccessPolicyResource_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterNetworkAccessPolicyHandlers(ms.Mux)

	r := newTestNAPResource(t, ms)
	s := napResourceSchema(t).Schema

	// Import the pre-seeded "default" singleton directly by name.
	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNetworkAccessPolicyType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "default"}, importResp)

	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState returned error: %s", importResp.Diagnostics)
	}

	var model networkAccessPolicyModel
	if diags := importResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Name.ValueString() != "default" {
		t.Errorf("expected name=default after import, got %s", model.Name.ValueString())
	}
	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected ID to be populated after import")
	}
}

// ---- data source tests -------------------------------------------------------

// newTestNAPDataSource creates a networkAccessPolicyDataSource wired to the given mock server.
func newTestNAPDataSource(t *testing.T, ms *testmock.MockServer) *networkAccessPolicyDataSource {
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
	return &networkAccessPolicyDataSource{client: c}
}

// napDataSourceSchema returns the schema for the NAP data source.
func napDataSourceSchema(t *testing.T) datasource.SchemaResponse {
	t.Helper()
	d := &networkAccessPolicyDataSource{}
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	return resp
}

// buildNetworkAccessPolicyDSType returns the tftypes.Object for the NAP data source.
func buildNetworkAccessPolicyDSType() tftypes.Object {
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":          tftypes.String,
		"name":        tftypes.String,
		"enabled":     tftypes.Bool,
		"is_local":    tftypes.Bool,
		"policy_type": tftypes.String,
		"version":     tftypes.String,
	}}
}

// nullNAPDSConfig returns a base config map with all data source attributes null.
func nullNAPDSConfig() map[string]tftypes.Value {
	return map[string]tftypes.Value{
		"id":          tftypes.NewValue(tftypes.String, nil),
		"name":        tftypes.NewValue(tftypes.String, nil),
		"enabled":     tftypes.NewValue(tftypes.Bool, nil),
		"is_local":    tftypes.NewValue(tftypes.Bool, nil),
		"policy_type": tftypes.NewValue(tftypes.String, nil),
		"version":     tftypes.NewValue(tftypes.String, nil),
	}
}

// TestNetworkAccessPolicyDataSource verifies data source reads the pre-seeded "default" policy by name.
func TestUnit_NetworkAccessPolicyDataSource(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterNetworkAccessPolicyHandlers(ms.Mux)

	d := newTestNAPDataSource(t, ms)
	s := napDataSourceSchema(t).Schema

	// The "default" policy is pre-seeded in the mock — data source should find it directly.
	cfg := nullNAPDSConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "default")

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNetworkAccessPolicyDSType(), nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(buildNetworkAccessPolicyDSType(), cfg), Schema: s},
	}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("DataSource Read returned error: %s", readResp.Diagnostics)
	}

	var model networkAccessPolicyDataSourceModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Name.ValueString() != "default" {
		t.Errorf("expected name=default, got %s", model.Name.ValueString())
	}
	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected ID to be populated")
	}
	if !model.Enabled.ValueBool() {
		t.Error("expected enabled=true for pre-seeded default policy")
	}
}

// TestUnit_NetworkAccessPolicy_Lifecycle exercises the full Create->Read->Update->Read->Delete sequence.
// NAP is a singleton (GET+PATCH). Create adopts the existing "default" policy.
func TestUnit_Unit_NetworkAccessPolicy_Lifecycle(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterNetworkAccessPolicyHandlers(ms.Mux)

	r := newTestNAPResource(t, ms)
	s := napResourceSchema(t).Schema

	// Step 1: Create (adopt singleton via GET+PATCH).
	createPlan := napPlanWithNameAndEnabled(t, "default", true)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNetworkAccessPolicyType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}
	var createModel networkAccessPolicyModel
	if diags := createResp.State.Get(context.Background(), &createModel); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}
	if createModel.Name.ValueString() != "default" {
		t.Errorf("Create: expected name=default, got %s", createModel.Name.ValueString())
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
	var readModel1 networkAccessPolicyModel
	if diags := readResp1.State.Get(context.Background(), &readModel1); diags.HasError() {
		t.Fatalf("Get read1 state: %s", diags)
	}
	if !readModel1.Enabled.ValueBool() {
		t.Error("Read1: expected enabled=true")
	}

	// Step 3: Update enabled=false.
	updatePlan := napPlanWithNameAndEnabled(t, "default", false)
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNetworkAccessPolicyType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlan,
		State: readResp1.State,
	}, updateResp)
	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update: %s", updateResp.Diagnostics)
	}
	var updateModel networkAccessPolicyModel
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
	var readModel2 networkAccessPolicyModel
	if diags := readResp2.State.Get(context.Background(), &readModel2); diags.HasError() {
		t.Fatalf("Get read2 state: %s", diags)
	}
	if readModel2.Enabled.ValueBool() {
		t.Error("Read2: expected enabled=false")
	}

	// Step 5: Delete (PATCH reset — singleton).
	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: readResp2.State}, deleteResp)
	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete: %s", deleteResp.Diagnostics)
	}
}

// TestUnit_NetworkAccessPolicy_ImportIdempotency verifies ImportState->Read produces state matching original Create.
func TestUnit_Unit_NetworkAccessPolicy_ImportIdempotency(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterNetworkAccessPolicyHandlers(ms.Mux)

	r := newTestNAPResource(t, ms)
	s := napResourceSchema(t).Schema

	// Create (adopt singleton).
	createPlan := napPlanWithNameAndEnabled(t, "default", true)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNetworkAccessPolicyType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}
	var createModel networkAccessPolicyModel
	if diags := createResp.State.Get(context.Background(), &createModel); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}

	// ImportState.
	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNetworkAccessPolicyType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "default"}, importResp)
	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState: %s", importResp.Diagnostics)
	}

	// Read to populate full state.
	readResp := &resource.ReadResponse{State: importResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: importResp.State}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read post-import: %s", readResp.Diagnostics)
	}
	var importedModel networkAccessPolicyModel
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

// TestUnit_NAP_PlanModifiers verifies all UseStateForUnknown plan modifiers
// in the network_access_policy resource schema.
func TestUnit_Unit_NAP_PlanModifiers(t *testing.T) {
	s := napResourceSchema(t).Schema

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
