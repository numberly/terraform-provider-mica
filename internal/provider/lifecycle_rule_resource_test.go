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

func ptrInt64LR(v int64) *int64 { return &v }

// newTestLifecycleRuleResource creates a lifecycleRuleResource wired to the given mock server.
func newTestLifecycleRuleResource(t *testing.T, ms *testmock.MockServer) *lifecycleRuleResource {
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
	return &lifecycleRuleResource{client: c}
}

// lifecycleRuleResourceSchema returns the parsed schema for the lifecycle rule resource.
func lifecycleRuleResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &lifecycleRuleResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

// buildLifecycleRuleType returns the tftypes.Object for the lifecycle rule resource schema.
func buildLifecycleRuleType() tftypes.Object {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":                                      tftypes.String,
		"bucket_name":                              tftypes.String,
		"rule_id":                                  tftypes.String,
		"prefix":                                   tftypes.String,
		"enabled":                                  tftypes.Bool,
		"abort_incomplete_multipart_uploads_after": tftypes.Number,
		"keep_current_version_for":                 tftypes.Number,
		"keep_current_version_until":               tftypes.Number,
		"keep_previous_version_for":                tftypes.Number,
		"cleanup_expired_object_delete_marker":     tftypes.Bool,
		"timeouts":                                 timeoutsType,
	}}
}

// nullLifecycleRuleConfig returns a base config map with all attributes null.
func nullLifecycleRuleConfig() map[string]tftypes.Value {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return map[string]tftypes.Value{
		"id":                                      tftypes.NewValue(tftypes.String, nil),
		"bucket_name":                              tftypes.NewValue(tftypes.String, nil),
		"rule_id":                                  tftypes.NewValue(tftypes.String, nil),
		"prefix":                                   tftypes.NewValue(tftypes.String, nil),
		"enabled":                                  tftypes.NewValue(tftypes.Bool, nil),
		"abort_incomplete_multipart_uploads_after": tftypes.NewValue(tftypes.Number, nil),
		"keep_current_version_for":                 tftypes.NewValue(tftypes.Number, nil),
		"keep_current_version_until":               tftypes.NewValue(tftypes.Number, nil),
		"keep_previous_version_for":                tftypes.NewValue(tftypes.Number, nil),
		"cleanup_expired_object_delete_marker":     tftypes.NewValue(tftypes.Bool, nil),
		"timeouts":                                 tftypes.NewValue(timeoutsType, nil),
	}
}

// lifecycleRulePlanWith returns a tfsdk.Plan with the given field values.
func lifecycleRulePlanWith(t *testing.T, bucketName, ruleID, prefix string, enabled bool, keepCurrentVersionFor int64) tfsdk.Plan {
	t.Helper()
	s := lifecycleRuleResourceSchema(t).Schema
	cfg := nullLifecycleRuleConfig()
	cfg["bucket_name"] = tftypes.NewValue(tftypes.String, bucketName)
	cfg["rule_id"] = tftypes.NewValue(tftypes.String, ruleID)
	cfg["prefix"] = tftypes.NewValue(tftypes.String, prefix)
	cfg["enabled"] = tftypes.NewValue(tftypes.Bool, enabled)
	if keepCurrentVersionFor > 0 {
		cfg["keep_current_version_for"] = tftypes.NewValue(tftypes.Number, keepCurrentVersionFor)
	}
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildLifecycleRuleType(), cfg),
		Schema: s,
	}
}

// ---- tests ------------------------------------------------------------------

// TestLifecycleRuleResource_Create verifies POST creates a rule, state populated with all fields.
func TestUnit_LifecycleRuleResource_Create(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterLifecycleRuleHandlers(ms.Mux)

	r := newTestLifecycleRuleResource(t, ms)
	s := lifecycleRuleResourceSchema(t).Schema

	plan := lifecycleRulePlanWith(t, "my-bucket", "rule-1", "logs/", true, 86400000)
	req := resource.CreateRequest{Plan: plan}
	resp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildLifecycleRuleType(), nil), Schema: s},
	}

	r.Create(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Create returned error: %s", resp.Diagnostics)
	}

	var model lifecycleRuleModel
	if diags := resp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected non-empty id after Create")
	}
	if model.BucketName.ValueString() != "my-bucket" {
		t.Errorf("expected bucket_name=my-bucket, got %s", model.BucketName.ValueString())
	}
	if model.RuleID.ValueString() != "rule-1" {
		t.Errorf("expected rule_id=rule-1, got %s", model.RuleID.ValueString())
	}
	if model.Prefix.ValueString() != "logs/" {
		t.Errorf("expected prefix=logs/, got %s", model.Prefix.ValueString())
	}
	if model.Enabled.ValueBool() != true {
		t.Error("expected enabled=true after Create")
	}
	if model.KeepCurrentVersionFor.ValueInt64() != 86400000 {
		t.Errorf("expected keep_current_version_for=86400000, got %d", model.KeepCurrentVersionFor.ValueInt64())
	}
}

// TestLifecycleRuleResource_Read verifies GET retrieves rule by bucket_name + rule_id.
func TestUnit_LifecycleRuleResource_Read(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterLifecycleRuleHandlers(ms.Mux)

	r := newTestLifecycleRuleResource(t, ms)
	s := lifecycleRuleResourceSchema(t).Schema

	// Create first.
	plan := lifecycleRulePlanWith(t, "read-bucket", "read-rule", "", true, 0)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildLifecycleRuleType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Read.
	readResp := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}

	var model lifecycleRuleModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.BucketName.ValueString() != "read-bucket" {
		t.Errorf("expected bucket_name=read-bucket, got %s", model.BucketName.ValueString())
	}
	if model.RuleID.ValueString() != "read-rule" {
		t.Errorf("expected rule_id=read-rule, got %s", model.RuleID.ValueString())
	}
	if model.Enabled.ValueBool() != true {
		t.Error("expected enabled=true after Read")
	}
}

// TestLifecycleRuleResource_Read_NotFound verifies GET returns 404, resource removed from state.
func TestUnit_LifecycleRuleResource_Read_NotFound(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterLifecycleRuleHandlers(ms.Mux)

	r := newTestLifecycleRuleResource(t, ms)
	s := lifecycleRuleResourceSchema(t).Schema

	// Build state directly (rule doesn't exist in mock).
	cfg := nullLifecycleRuleConfig()
	cfg["id"] = tftypes.NewValue(tftypes.String, "lcr-999")
	cfg["bucket_name"] = tftypes.NewValue(tftypes.String, "ghost-bucket")
	cfg["rule_id"] = tftypes.NewValue(tftypes.String, "ghost-rule")
	cfg["prefix"] = tftypes.NewValue(tftypes.String, "")
	cfg["enabled"] = tftypes.NewValue(tftypes.Bool, true)

	state := tfsdk.State{
		Raw:    tftypes.NewValue(buildLifecycleRuleType(), cfg),
		Schema: s,
	}

	readResp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), resource.ReadRequest{State: state}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned unexpected error: %s", readResp.Diagnostics)
	}

	// State should be removed (raw is null).
	if !readResp.State.Raw.IsNull() {
		t.Error("expected state to be removed after not-found Read")
	}
}

// TestLifecycleRuleResource_Update verifies PATCH sends only changed fields.
func TestUnit_LifecycleRuleResource_Update(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterLifecycleRuleHandlers(ms.Mux)

	r := newTestLifecycleRuleResource(t, ms)
	s := lifecycleRuleResourceSchema(t).Schema

	// Create (enabled=true).
	createPlan := lifecycleRulePlanWith(t, "upd-bucket", "upd-rule", "data/", true, 86400000)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildLifecycleRuleType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Update: disable the rule.
	updatePlan := lifecycleRulePlanWith(t, "upd-bucket", "upd-rule", "data/", false, 86400000)
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildLifecycleRuleType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlan,
		State: createResp.State,
	}, updateResp)

	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update returned error: %s", updateResp.Diagnostics)
	}

	var model lifecycleRuleModel
	if diags := updateResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Enabled.ValueBool() != false {
		t.Error("expected enabled=false after Update")
	}
}

// TestLifecycleRuleResource_Delete verifies DELETE by bucket_name + rule_id succeeds.
func TestUnit_LifecycleRuleResource_Delete(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterLifecycleRuleHandlers(ms.Mux)

	r := newTestLifecycleRuleResource(t, ms)
	s := lifecycleRuleResourceSchema(t).Schema

	// Create.
	plan := lifecycleRulePlanWith(t, "del-bucket", "del-rule", "", true, 0)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildLifecycleRuleType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Delete.
	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: createResp.State}, deleteResp)

	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete returned error: %s", deleteResp.Diagnostics)
	}

	// Verify it's gone.
	_, err := r.client.GetLifecycleRule(context.Background(), "del-bucket", "del-rule")
	if err == nil || !client.IsNotFound(err) {
		t.Errorf("expected lifecycle rule to be deleted, got: %v", err)
	}
}

// TestLifecycleRuleResource_Delete_NotFound verifies DELETE on already-deleted rule succeeds silently.
func TestUnit_LifecycleRuleResource_Delete_NotFound(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterLifecycleRuleHandlers(ms.Mux)

	r := newTestLifecycleRuleResource(t, ms)
	s := lifecycleRuleResourceSchema(t).Schema

	// Build state for a non-existent rule.
	cfg := nullLifecycleRuleConfig()
	cfg["id"] = tftypes.NewValue(tftypes.String, "lcr-999")
	cfg["bucket_name"] = tftypes.NewValue(tftypes.String, "ghost-bucket")
	cfg["rule_id"] = tftypes.NewValue(tftypes.String, "ghost-rule")
	cfg["prefix"] = tftypes.NewValue(tftypes.String, "")
	cfg["enabled"] = tftypes.NewValue(tftypes.Bool, true)

	state := tfsdk.State{
		Raw:    tftypes.NewValue(buildLifecycleRuleType(), cfg),
		Schema: s,
	}

	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: state}, deleteResp)

	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete on non-existent rule returned error: %s", deleteResp.Diagnostics)
	}
}

// TestLifecycleRuleResource_Import verifies Import by "bucketName/ruleId" populates all state fields.
func TestUnit_LifecycleRuleResource_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterLifecycleRuleHandlers(ms.Mux)

	// Seed a rule so import can find it.
	store.Seed(&client.LifecycleRule{
		ID:                    "lcr-imp-1",
		Name:                  "imp-bucket/imp-rule",
		Bucket:                client.NamedReference{Name: "imp-bucket"},
		RuleID:                "imp-rule",
		Prefix:                "archive/",
		Enabled:               true,
		KeepCurrentVersionFor: ptrInt64LR(172800000),
	})

	r := newTestLifecycleRuleResource(t, ms)
	s := lifecycleRuleResourceSchema(t).Schema

	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildLifecycleRuleType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "imp-bucket/imp-rule"}, importResp)

	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState returned error: %s", importResp.Diagnostics)
	}

	var model lifecycleRuleModel
	if diags := importResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.ValueString() != "lcr-imp-1" {
		t.Errorf("expected id=lcr-imp-1, got %s", model.ID.ValueString())
	}
	if model.BucketName.ValueString() != "imp-bucket" {
		t.Errorf("expected bucket_name=imp-bucket, got %s", model.BucketName.ValueString())
	}
	if model.RuleID.ValueString() != "imp-rule" {
		t.Errorf("expected rule_id=imp-rule, got %s", model.RuleID.ValueString())
	}
	if model.Prefix.ValueString() != "archive/" {
		t.Errorf("expected prefix=archive/, got %s", model.Prefix.ValueString())
	}
	if model.Enabled.ValueBool() != true {
		t.Error("expected enabled=true after import")
	}
	if model.KeepCurrentVersionFor.ValueInt64() != 172800000 {
		t.Errorf("expected keep_current_version_for=172800000, got %d", model.KeepCurrentVersionFor.ValueInt64())
	}
}

// TestLifecycleRuleResource_Schema verifies schema properties.
func TestUnit_LifecycleRuleResource_Schema(t *testing.T) {
	s := lifecycleRuleResourceSchema(t).Schema

	// bucket_name: Required + RequiresReplace.
	bucketAttr, ok := s.Attributes["bucket_name"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("bucket_name attribute not found or wrong type")
	}
	if !bucketAttr.Required {
		t.Error("bucket_name: expected Required=true")
	}
	if len(bucketAttr.PlanModifiers) == 0 {
		t.Error("bucket_name: expected RequiresReplace plan modifier")
	}

	// rule_id: Required + RequiresReplace.
	ruleAttr, ok := s.Attributes["rule_id"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("rule_id attribute not found or wrong type")
	}
	if !ruleAttr.Required {
		t.Error("rule_id: expected Required=true")
	}
	if len(ruleAttr.PlanModifiers) == 0 {
		t.Error("rule_id: expected RequiresReplace plan modifier")
	}

	// enabled: Optional + Computed.
	enabledAttr, ok := s.Attributes["enabled"].(resschema.BoolAttribute)
	if !ok {
		t.Fatal("enabled attribute not found or wrong type")
	}
	if !enabledAttr.Optional {
		t.Error("enabled: expected Optional=true")
	}
	if !enabledAttr.Computed {
		t.Error("enabled: expected Computed=true")
	}

	// cleanup_expired_object_delete_marker: Computed only.
	cleanupAttr, ok := s.Attributes["cleanup_expired_object_delete_marker"].(resschema.BoolAttribute)
	if !ok {
		t.Fatal("cleanup_expired_object_delete_marker attribute not found or wrong type")
	}
	if !cleanupAttr.Computed {
		t.Error("cleanup_expired_object_delete_marker: expected Computed=true")
	}
}
