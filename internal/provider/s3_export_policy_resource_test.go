package provider

import (
	"context"
	"testing"
	"time"

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

// newTestS3PolicyResource creates an s3ExportPolicyResource wired to the given mock server.
func newTestS3PolicyResource(t *testing.T, ms *testmock.MockServer) *s3ExportPolicyResource {
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
	return &s3ExportPolicyResource{client: c}
}

// s3PolicyResourceSchema returns the parsed schema for the S3 export policy resource.
func s3PolicyResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &s3ExportPolicyResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

// buildS3PolicyType returns the tftypes.Object for the S3 export policy resource.
func buildS3PolicyType() tftypes.Object {
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

// nullS3PolicyConfig returns a base config map with all attributes null.
func nullS3PolicyConfig() map[string]tftypes.Value {
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

// s3PolicyPlanWithNameAndEnabled returns a tfsdk.Plan with name and enabled flag.
func s3PolicyPlanWithNameAndEnabled(t *testing.T, name string, enabled bool) tfsdk.Plan {
	t.Helper()
	s := s3PolicyResourceSchema(t).Schema
	cfg := nullS3PolicyConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, name)
	cfg["enabled"] = tftypes.NewValue(tftypes.Bool, enabled)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildS3PolicyType(), cfg),
		Schema: s,
	}
}

// createTestS3Policy is a helper that creates an S3 export policy via the client.
func createTestS3Policy(t *testing.T, c *client.FlashBladeClient, name string) {
	t.Helper()
	enabled := true
	_, err := c.PostS3ExportPolicy(context.Background(), name, client.S3ExportPolicyPost{Enabled: &enabled})
	if err != nil {
		t.Fatalf("PostS3ExportPolicy(%q): %v", name, err)
	}
}

// ---- tests ------------------------------------------------------------------

// TestUnit_S3ExportPolicy_CreateReadUpdateDelete exercises the full lifecycle.
func TestUnit_S3ExportPolicy_CreateReadUpdateDelete(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterS3ExportPolicyHandlers(ms.Mux)

	r := newTestS3PolicyResource(t, ms)
	s := s3PolicyResourceSchema(t).Schema

	// Step 1: Create with enabled=true.
	createPlan := s3PolicyPlanWithNameAndEnabled(t, "test-s3-policy", true)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildS3PolicyType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}
	var createModel s3ExportPolicyModel
	if diags := createResp.State.Get(context.Background(), &createModel); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}
	if createModel.Name.ValueString() != "test-s3-policy" {
		t.Errorf("Create: expected name=test-s3-policy, got %s", createModel.Name.ValueString())
	}
	if !createModel.Enabled.ValueBool() {
		t.Error("Create: expected enabled=true")
	}
	if createModel.ID.IsNull() || createModel.ID.ValueString() == "" {
		t.Error("Create: expected non-empty ID")
	}

	// Step 2: Read post-create.
	readResp1 := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp1)
	if readResp1.Diagnostics.HasError() {
		t.Fatalf("Read post-create: %s", readResp1.Diagnostics)
	}
	var readModel1 s3ExportPolicyModel
	if diags := readResp1.State.Get(context.Background(), &readModel1); diags.HasError() {
		t.Fatalf("Get read1 state: %s", diags)
	}
	if !readModel1.Enabled.ValueBool() {
		t.Error("Read1: expected enabled=true")
	}

	// Step 3: Update enabled=false.
	updatePlan := s3PolicyPlanWithNameAndEnabled(t, "test-s3-policy", false)
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildS3PolicyType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlan,
		State: readResp1.State,
	}, updateResp)
	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update: %s", updateResp.Diagnostics)
	}
	var updateModel s3ExportPolicyModel
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
	var readModel2 s3ExportPolicyModel
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
	_, err := r.client.GetS3ExportPolicy(context.Background(), "test-s3-policy")
	if err == nil || !client.IsNotFound(err) {
		t.Errorf("expected policy to be deleted, got: %v", err)
	}
}

// TestUnit_S3ExportPolicy_Import verifies ImportState populates all attributes.
func TestUnit_S3ExportPolicy_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterS3ExportPolicyHandlers(ms.Mux)

	r := newTestS3PolicyResource(t, ms)
	s := s3PolicyResourceSchema(t).Schema

	// Create first.
	createPlan := s3PolicyPlanWithNameAndEnabled(t, "import-s3-policy", true)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildS3PolicyType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}
	var createModel s3ExportPolicyModel
	if diags := createResp.State.Get(context.Background(), &createModel); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}

	// Import by name.
	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildS3PolicyType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "import-s3-policy"}, importResp)
	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState: %s", importResp.Diagnostics)
	}

	// Read to populate full state.
	readResp := &resource.ReadResponse{State: importResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: importResp.State}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read post-import: %s", readResp.Diagnostics)
	}
	var importedModel s3ExportPolicyModel
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

// TestUnit_S3ExportPolicy_PlanModifiers verifies UseStateForUnknown plan modifiers.
func TestUnit_S3ExportPolicy_PlanModifiers(t *testing.T) {
	s := s3PolicyResourceSchema(t).Schema

	// id -- UseStateForUnknown
	idAttr, ok := s.Attributes["id"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("id attribute not found or wrong type")
	}
	if len(idAttr.PlanModifiers) == 0 {
		t.Error("expected UseStateForUnknown plan modifier on id attribute")
	}

	// is_local -- UseStateForUnknown
	ilAttr, ok := s.Attributes["is_local"].(resschema.BoolAttribute)
	if !ok {
		t.Fatal("is_local attribute not found or wrong type")
	}
	if len(ilAttr.PlanModifiers) == 0 {
		t.Error("expected UseStateForUnknown plan modifier on is_local attribute")
	}

	// policy_type -- UseStateForUnknown
	ptAttr, ok := s.Attributes["policy_type"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("policy_type attribute not found or wrong type")
	}
	if len(ptAttr.PlanModifiers) == 0 {
		t.Error("expected UseStateForUnknown plan modifier on policy_type attribute")
	}
}

// ---- data source tests -------------------------------------------------------

// newTestS3PolicyDataSource creates an s3ExportPolicyDataSource wired to the given mock server.
func newTestS3PolicyDataSource(t *testing.T, ms *testmock.MockServer) *s3ExportPolicyDataSource {
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
	return &s3ExportPolicyDataSource{client: c}
}

// s3PolicyDataSourceSchema returns the schema for the S3 export policy data source.
func s3PolicyDataSourceSchema(t *testing.T) datasource.SchemaResponse {
	t.Helper()
	d := &s3ExportPolicyDataSource{}
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	return resp
}

// buildS3PolicyDSType returns the tftypes.Object for the S3 export policy data source.
func buildS3PolicyDSType() tftypes.Object {
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":          tftypes.String,
		"name":        tftypes.String,
		"enabled":     tftypes.Bool,
		"is_local":    tftypes.Bool,
		"policy_type": tftypes.String,
		"version":     tftypes.String,
	}}
}

// nullS3PolicyDSConfig returns a base config map with all data source attributes null.
func nullS3PolicyDSConfig() map[string]tftypes.Value {
	return map[string]tftypes.Value{
		"id":          tftypes.NewValue(tftypes.String, nil),
		"name":        tftypes.NewValue(tftypes.String, nil),
		"enabled":     tftypes.NewValue(tftypes.Bool, nil),
		"is_local":    tftypes.NewValue(tftypes.Bool, nil),
		"policy_type": tftypes.NewValue(tftypes.String, nil),
		"version":     tftypes.NewValue(tftypes.String, nil),
	}
}

// TestUnit_S3ExportPolicy_DataSource verifies data source reads policy by name.
func TestUnit_S3ExportPolicy_DataSource(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterS3ExportPolicyHandlers(ms.Mux)

	// Create a policy via the client so the data source can find it.
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
	_, err = c.PostS3ExportPolicy(context.Background(), "ds-test-s3-policy", client.S3ExportPolicyPost{
		Enabled: &enabled,
	})
	if err != nil {
		t.Fatalf("PostS3ExportPolicy: %v", err)
	}

	d := newTestS3PolicyDataSource(t, ms)
	s := s3PolicyDataSourceSchema(t).Schema

	cfg := nullS3PolicyDSConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "ds-test-s3-policy")

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildS3PolicyDSType(), nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(buildS3PolicyDSType(), cfg), Schema: s},
	}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("DataSource Read returned error: %s", readResp.Diagnostics)
	}

	var model s3ExportPolicyDataSourceModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Name.ValueString() != "ds-test-s3-policy" {
		t.Errorf("expected name=ds-test-s3-policy, got %s", model.Name.ValueString())
	}
	if !model.Enabled.ValueBool() {
		t.Error("expected enabled=true")
	}
	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected ID to be populated")
	}
}
