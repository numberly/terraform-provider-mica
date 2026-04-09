package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/numberly/opentofu-provider-flashblade/internal/client"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock/handlers"
)

// ---- helpers ----------------------------------------------------------------

// newTestAuditObjectStorePolicyResource creates an auditObjectStorePolicyResource wired to the given mock server.
func newTestAuditObjectStorePolicyResource(t *testing.T, ms *testmock.MockServer) *auditObjectStorePolicyResource {
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
	return &auditObjectStorePolicyResource{client: c}
}

// auditObjectStorePolicyResourceSchema returns the parsed schema for the resource.
func auditObjectStorePolicyResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &auditObjectStorePolicyResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

// buildAuditObjectStorePolicyType returns the tftypes.Object for the resource schema.
func buildAuditObjectStorePolicyType() tftypes.Object {
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
		"log_targets": tftypes.List{ElementType: tftypes.String},
		"timeouts":    timeoutsType,
	}}
}

// nullAuditObjectStorePolicyConfig returns a base config map with all attributes null.
func nullAuditObjectStorePolicyConfig() map[string]tftypes.Value {
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
		"log_targets": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"timeouts":    tftypes.NewValue(timeoutsType, nil),
	}
}

// auditObjectStorePolicyPlan builds a tfsdk.Plan for the given parameters.
func auditObjectStorePolicyPlan(t *testing.T, name string, enabled bool, logTargets []string) tfsdk.Plan {
	t.Helper()
	s := auditObjectStorePolicyResourceSchema(t).Schema
	cfg := nullAuditObjectStorePolicyConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, name)
	cfg["enabled"] = tftypes.NewValue(tftypes.Bool, enabled)

	if logTargets != nil {
		vals := make([]tftypes.Value, len(logTargets))
		for i, lt := range logTargets {
			vals[i] = tftypes.NewValue(tftypes.String, lt)
		}
		cfg["log_targets"] = tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, vals)
	} else {
		cfg["log_targets"] = tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{})
	}

	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildAuditObjectStorePolicyType(), cfg),
		Schema: s,
	}
}

// ---- resource tests ---------------------------------------------------------

// TestUnit_AuditObjectStorePolicyResource_Lifecycle exercises Create->Read->Update->Delete.
func TestUnit_AuditObjectStorePolicyResource_Lifecycle(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterAuditObjectStorePolicyHandlers(ms.Mux)

	r := newTestAuditObjectStorePolicyResource(t, ms)
	s := auditObjectStorePolicyResourceSchema(t).Schema

	// Step 1: Create with enabled=true, no log targets.
	createPlan := auditObjectStorePolicyPlan(t, "test-audit-policy", true, nil)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildAuditObjectStorePolicyType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	var createModel auditObjectStorePolicyModel
	if diags := createResp.State.Get(context.Background(), &createModel); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}
	if createModel.Name.ValueString() != "test-audit-policy" {
		t.Errorf("Create: expected name=test-audit-policy, got %s", createModel.Name.ValueString())
	}
	if !createModel.Enabled.ValueBool() {
		t.Error("Create: expected enabled=true")
	}
	if createModel.ID.IsNull() || createModel.ID.ValueString() == "" {
		t.Error("Create: expected non-empty ID")
	}
	if createModel.LogTargets.IsNull() {
		t.Error("Create: expected log_targets to be empty list, not null")
	}
	if len(createModel.LogTargets.Elements()) != 0 {
		t.Errorf("Create: expected empty log_targets, got %d elements", len(createModel.LogTargets.Elements()))
	}

	// Step 2: Read post-create (0-diff check).
	readResp1 := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp1)
	if readResp1.Diagnostics.HasError() {
		t.Fatalf("Read post-create: %s", readResp1.Diagnostics)
	}

	// Step 3: Update enabled=false and add a log target.
	updatePlan := auditObjectStorePolicyPlan(t, "test-audit-policy", false, []string{"log-target-1"})
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildAuditObjectStorePolicyType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlan,
		State: readResp1.State,
	}, updateResp)
	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update: %s", updateResp.Diagnostics)
	}

	var updateModel auditObjectStorePolicyModel
	if diags := updateResp.State.Get(context.Background(), &updateModel); diags.HasError() {
		t.Fatalf("Get update state: %s", diags)
	}
	if updateModel.Enabled.ValueBool() {
		t.Error("Update: expected enabled=false")
	}
	var updateLogTargets []string
	if diags := updateModel.LogTargets.ElementsAs(context.Background(), &updateLogTargets, false); diags.HasError() {
		t.Fatalf("ElementsAs log_targets: %s", diags)
	}
	if len(updateLogTargets) != 1 || updateLogTargets[0] != "log-target-1" {
		t.Errorf("Update: expected log_targets=[log-target-1], got %v", updateLogTargets)
	}

	// Step 4: Read post-update (0-diff check).
	readResp2 := &resource.ReadResponse{State: updateResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: updateResp.State}, readResp2)
	if readResp2.Diagnostics.HasError() {
		t.Fatalf("Read post-update: %s", readResp2.Diagnostics)
	}

	// Step 5: Delete.
	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: updateResp.State}, deleteResp)
	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete: %s", deleteResp.Diagnostics)
	}
	_, err := r.client.GetAuditObjectStorePolicy(context.Background(), "test-audit-policy")
	if err == nil || !client.IsNotFound(err) {
		t.Errorf("expected audit policy to be deleted, got: %v", err)
	}
}

// TestUnit_AuditObjectStorePolicyResource_Import verifies ImportState populates all attributes correctly.
func TestUnit_AuditObjectStorePolicyResource_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterAuditObjectStorePolicyHandlers(ms.Mux)

	r := newTestAuditObjectStorePolicyResource(t, ms)
	s := auditObjectStorePolicyResourceSchema(t).Schema

	// Create first.
	createPlan := auditObjectStorePolicyPlan(t, "import-audit-policy", true, []string{"target-x"})
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildAuditObjectStorePolicyType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	var createModel auditObjectStorePolicyModel
	if diags := createResp.State.Get(context.Background(), &createModel); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}

	// Import by name.
	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildAuditObjectStorePolicyType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "import-audit-policy"}, importResp)
	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState: %s", importResp.Diagnostics)
	}

	var importedModel auditObjectStorePolicyModel
	if diags := importResp.State.Get(context.Background(), &importedModel); diags.HasError() {
		t.Fatalf("Get imported state: %s", diags)
	}

	if importedModel.Name.ValueString() != createModel.Name.ValueString() {
		t.Errorf("name mismatch: create=%s import=%s", createModel.Name.ValueString(), importedModel.Name.ValueString())
	}
	if importedModel.ID.ValueString() != createModel.ID.ValueString() {
		t.Errorf("id mismatch: create=%s import=%s", createModel.ID.ValueString(), importedModel.ID.ValueString())
	}
	if importedModel.Enabled.ValueBool() != createModel.Enabled.ValueBool() {
		t.Errorf("enabled mismatch: create=%v import=%v", createModel.Enabled.ValueBool(), importedModel.Enabled.ValueBool())
	}
	if !importedModel.LogTargets.Equal(createModel.LogTargets) {
		t.Errorf("log_targets mismatch after import")
	}
}

// TestUnit_AuditObjectStorePolicyResource_DriftDetection verifies that Read after
// out-of-band changes updates the state correctly.
func TestUnit_AuditObjectStorePolicyResource_DriftDetection(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterAuditObjectStorePolicyHandlers(ms.Mux)

	r := newTestAuditObjectStorePolicyResource(t, ms)
	s := auditObjectStorePolicyResourceSchema(t).Schema

	// Create.
	createPlan := auditObjectStorePolicyPlan(t, "drift-audit-policy", true, nil)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildAuditObjectStorePolicyType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Simulate out-of-band change: disable the policy.
	store.Seed(&client.AuditObjectStorePolicy{
		ID:         "drift-id",
		Name:       "drift-audit-policy",
		Enabled:    false,
		IsLocal:    true,
		PolicyType: "audit",
		LogTargets: []client.NamedReference{{Name: "new-target"}},
	})

	// Read should pick up the drift.
	readResp := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read: %s", readResp.Diagnostics)
	}

	var readModel auditObjectStorePolicyModel
	if diags := readResp.State.Get(context.Background(), &readModel); diags.HasError() {
		t.Fatalf("Get read state: %s", diags)
	}

	if readModel.Enabled.ValueBool() {
		t.Error("DriftDetection: expected enabled=false after out-of-band change")
	}

	var logTargets []string
	if diags := readModel.LogTargets.ElementsAs(context.Background(), &logTargets, false); diags.HasError() {
		t.Fatalf("ElementsAs log_targets: %s", diags)
	}
	if len(logTargets) != 1 || logTargets[0] != "new-target" {
		t.Errorf("DriftDetection: expected log_targets=[new-target], got %v", logTargets)
	}
}

// ---- data source helpers ----------------------------------------------------

// newTestAuditObjectStorePolicyDataSource creates an auditObjectStorePolicyDataSource wired to the given mock server.
func newTestAuditObjectStorePolicyDataSource(t *testing.T, ms *testmock.MockServer) *auditObjectStorePolicyDataSource {
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
	return &auditObjectStorePolicyDataSource{client: c}
}

// auditObjectStorePolicyDataSourceSchema returns the schema for the data source.
func auditObjectStorePolicyDataSourceSchema(t *testing.T) datasource.SchemaResponse {
	t.Helper()
	d := &auditObjectStorePolicyDataSource{}
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	return resp
}

// buildAuditObjectStorePolicyDSType returns the tftypes.Object for the data source schema.
func buildAuditObjectStorePolicyDSType() tftypes.Object {
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":          tftypes.String,
		"name":        tftypes.String,
		"enabled":     tftypes.Bool,
		"is_local":    tftypes.Bool,
		"policy_type": tftypes.String,
		"log_targets": tftypes.List{ElementType: tftypes.String},
	}}
}

// nullAuditObjectStorePolicyDSConfig returns a base config map with all data source attributes null.
func nullAuditObjectStorePolicyDSConfig() map[string]tftypes.Value {
	return map[string]tftypes.Value{
		"id":          tftypes.NewValue(tftypes.String, nil),
		"name":        tftypes.NewValue(tftypes.String, nil),
		"enabled":     tftypes.NewValue(tftypes.Bool, nil),
		"is_local":    tftypes.NewValue(tftypes.Bool, nil),
		"policy_type": tftypes.NewValue(tftypes.String, nil),
		"log_targets": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
	}
}

// ---- data source tests ------------------------------------------------------

// TestUnit_AuditObjectStorePolicyDataSource_Basic verifies data source reads policy by name.
func TestUnit_AuditObjectStorePolicyDataSource_Basic(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterAuditObjectStorePolicyHandlers(ms.Mux)

	store.Seed(&client.AuditObjectStorePolicy{
		ID:         "ds-policy-id-1",
		Name:       "ds-audit-policy",
		Enabled:    true,
		IsLocal:    true,
		PolicyType: "audit",
		LogTargets: []client.NamedReference{{Name: "ds-log-target"}},
	})

	d := newTestAuditObjectStorePolicyDataSource(t, ms)
	s := auditObjectStorePolicyDataSourceSchema(t).Schema

	cfg := nullAuditObjectStorePolicyDSConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "ds-audit-policy")

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildAuditObjectStorePolicyDSType(), nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(buildAuditObjectStorePolicyDSType(), cfg), Schema: s},
	}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("DataSource Read returned error: %s", readResp.Diagnostics)
	}

	var model auditObjectStorePolicyDataSourceModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Name.ValueString() != "ds-audit-policy" {
		t.Errorf("expected name=ds-audit-policy, got %s", model.Name.ValueString())
	}
	if model.ID.ValueString() != "ds-policy-id-1" {
		t.Errorf("expected id=ds-policy-id-1, got %s", model.ID.ValueString())
	}
	if !model.Enabled.ValueBool() {
		t.Error("expected enabled=true")
	}
	if !model.IsLocal.ValueBool() {
		t.Error("expected is_local=true")
	}
	if model.PolicyType.ValueString() != "audit" {
		t.Errorf("expected policy_type=audit, got %s", model.PolicyType.ValueString())
	}

	var logTargets []string
	if diags := model.LogTargets.ElementsAs(context.Background(), &logTargets, false); diags.HasError() {
		t.Fatalf("ElementsAs log_targets: %s", diags)
	}
	if len(logTargets) != 1 || logTargets[0] != "ds-log-target" {
		t.Errorf("expected log_targets=[ds-log-target], got %v", logTargets)
	}
}
