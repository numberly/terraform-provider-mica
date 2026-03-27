package provider

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/soulkyu/terraform-provider-flashblade/internal/client"
	"github.com/soulkyu/terraform-provider-flashblade/internal/testmock"
	"github.com/soulkyu/terraform-provider-flashblade/internal/testmock/handlers"
)

// ---- helpers ----------------------------------------------------------------

// newTestOAPResource creates an objectStoreAccessPolicyResource wired to the given mock server.
func newTestOAPResource(t *testing.T, ms *testmock.MockServer) *objectStoreAccessPolicyResource {
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
	return &objectStoreAccessPolicyResource{client: c}
}

// oapResourceSchema returns the parsed schema for the OAP resource.
func oapResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &objectStoreAccessPolicyResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

// buildOAPType returns the tftypes.Object for the OAP resource.
func buildOAPType() tftypes.Object {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":          tftypes.String,
		"name":        tftypes.String,
		"description": tftypes.String,
		"arn":         tftypes.String,
		"enabled":     tftypes.Bool,
		"is_local":    tftypes.Bool,
		"policy_type": tftypes.String,
		"timeouts":    timeoutsType,
	}}
}

// nullOAPConfig returns a base config map with all attributes null.
func nullOAPConfig() map[string]tftypes.Value {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return map[string]tftypes.Value{
		"id":          tftypes.NewValue(tftypes.String, nil),
		"name":        tftypes.NewValue(tftypes.String, nil),
		"description": tftypes.NewValue(tftypes.String, nil),
		"arn":         tftypes.NewValue(tftypes.String, nil),
		"enabled":     tftypes.NewValue(tftypes.Bool, nil),
		"is_local":    tftypes.NewValue(tftypes.Bool, nil),
		"policy_type": tftypes.NewValue(tftypes.String, nil),
		"timeouts":    tftypes.NewValue(timeoutsType, nil),
	}
}

// oapPlanWithName returns a tfsdk.Plan with the given policy name.
func oapPlanWithName(t *testing.T, name string) tfsdk.Plan {
	t.Helper()
	s := oapResourceSchema(t).Schema
	cfg := nullOAPConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, name)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildOAPType(), cfg),
		Schema: s,
	}
}

// oapPlanWithNameAndDescription returns a tfsdk.Plan with name and description.
func oapPlanWithNameAndDescription(t *testing.T, name, description string) tfsdk.Plan {
	t.Helper()
	s := oapResourceSchema(t).Schema
	cfg := nullOAPConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, name)
	cfg["description"] = tftypes.NewValue(tftypes.String, description)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildOAPType(), cfg),
		Schema: s,
	}
}

// ---- tests ------------------------------------------------------------------

// TestObjectStoreAccessPolicyResource_Create verifies Create populates ID, name, enabled, and description.
func TestObjectStoreAccessPolicyResource_Create(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterObjectStoreAccessPolicyHandlers(ms.Mux)

	r := newTestOAPResource(t, ms)
	s := oapResourceSchema(t).Schema

	plan := oapPlanWithNameAndDescription(t, "test-oap-policy", "test description")
	resp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildOAPType(), nil), Schema: s},
	}

	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Create returned error: %s", resp.Diagnostics)
	}

	var model objectStoreAccessPolicyModel
	if diags := resp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected non-empty ID after Create")
	}
	if model.Name.ValueString() != "test-oap-policy" {
		t.Errorf("expected name=test-oap-policy, got %s", model.Name.ValueString())
	}
	if model.Description.ValueString() != "test description" {
		t.Errorf("expected description='test description', got %s", model.Description.ValueString())
	}
	if !model.Enabled.ValueBool() {
		t.Error("expected enabled=true after Create")
	}
	if model.PolicyType.ValueString() != "object-store-access" {
		t.Errorf("expected policy_type=object-store-access, got %s", model.PolicyType.ValueString())
	}
}

// TestObjectStoreAccessPolicyResource_Update verifies PATCH supports rename (name change).
func TestObjectStoreAccessPolicyResource_Update(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterObjectStoreAccessPolicyHandlers(ms.Mux)

	r := newTestOAPResource(t, ms)
	s := oapResourceSchema(t).Schema

	// Create first.
	createPlan := oapPlanWithName(t, "oap-rename-before")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildOAPType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Rename the policy in-place via PATCH.
	renamePlan := oapPlanWithName(t, "oap-rename-after")
	renameResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildOAPType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  renamePlan,
		State: createResp.State,
	}, renameResp)

	if renameResp.Diagnostics.HasError() {
		t.Fatalf("Update (rename) returned error: %s", renameResp.Diagnostics)
	}

	var model objectStoreAccessPolicyModel
	if diags := renameResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Name.ValueString() != "oap-rename-after" {
		t.Errorf("expected name=oap-rename-after after rename, got %s", model.Name.ValueString())
	}
}

// TestObjectStoreAccessPolicyResource_Delete verifies DELETE removes the policy.
func TestObjectStoreAccessPolicyResource_Delete(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterObjectStoreAccessPolicyHandlers(ms.Mux)
	// Register buckets handler for the delete-guard (nil accounts store — GET-only, no POST needed).
	handlers.RegisterBucketHandlers(ms.Mux, nil)

	r := newTestOAPResource(t, ms)
	s := oapResourceSchema(t).Schema

	// Create first.
	plan := oapPlanWithName(t, "delete-oap-policy")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildOAPType(), nil), Schema: s},
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
	_, err := r.client.GetObjectStoreAccessPolicy(context.Background(), "delete-oap-policy")
	if err == nil || !client.IsNotFound(err) {
		t.Errorf("expected policy to be deleted, got: %v", err)
	}
}

// TestObjectStoreAccessPolicyResource_Import verifies ImportState populates all attributes.
func TestObjectStoreAccessPolicyResource_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterObjectStoreAccessPolicyHandlers(ms.Mux)

	r := newTestOAPResource(t, ms)
	s := oapResourceSchema(t).Schema

	// Create first.
	plan := oapPlanWithName(t, "import-oap-policy")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildOAPType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Import by name.
	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildOAPType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "import-oap-policy"}, importResp)

	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState returned error: %s", importResp.Diagnostics)
	}

	var model objectStoreAccessPolicyModel
	if diags := importResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Name.ValueString() != "import-oap-policy" {
		t.Errorf("expected name=import-oap-policy after import, got %s", model.Name.ValueString())
	}
	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected ID to be populated after import")
	}
}

// ---- data source tests -------------------------------------------------------

// newTestOAPDataSource creates an objectStoreAccessPolicyDataSource wired to the given mock server.
func newTestOAPDataSource(t *testing.T, ms *testmock.MockServer) *objectStoreAccessPolicyDataSource {
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
	return &objectStoreAccessPolicyDataSource{client: c}
}

// oapDataSourceSchema returns the schema for the OAP data source.
func oapDataSourceSchema(t *testing.T) datasource.SchemaResponse {
	t.Helper()
	d := &objectStoreAccessPolicyDataSource{}
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	return resp
}

// buildOAPDSType returns the tftypes.Object for the OAP data source.
func buildOAPDSType() tftypes.Object {
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":          tftypes.String,
		"name":        tftypes.String,
		"description": tftypes.String,
		"arn":         tftypes.String,
		"enabled":     tftypes.Bool,
		"is_local":    tftypes.Bool,
		"policy_type": tftypes.String,
	}}
}

// nullOAPDSConfig returns a base config map with all data source attributes null.
func nullOAPDSConfig() map[string]tftypes.Value {
	return map[string]tftypes.Value{
		"id":          tftypes.NewValue(tftypes.String, nil),
		"name":        tftypes.NewValue(tftypes.String, nil),
		"description": tftypes.NewValue(tftypes.String, nil),
		"arn":         tftypes.NewValue(tftypes.String, nil),
		"enabled":     tftypes.NewValue(tftypes.Bool, nil),
		"is_local":    tftypes.NewValue(tftypes.Bool, nil),
		"policy_type": tftypes.NewValue(tftypes.String, nil),
	}
}

// TestObjectStoreAccessPolicyDataSource verifies data source reads policy by name and returns all attributes.
func TestObjectStoreAccessPolicyDataSource(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterObjectStoreAccessPolicyHandlers(ms.Mux)

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
	_, err = c.PostObjectStoreAccessPolicy(context.Background(), "ds-oap-test", client.ObjectStoreAccessPolicyPost{
		Description: "datasource test",
	})
	if err != nil {
		t.Fatalf("PostObjectStoreAccessPolicy: %v", err)
	}

	d := newTestOAPDataSource(t, ms)
	s := oapDataSourceSchema(t).Schema

	cfg := nullOAPDSConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "ds-oap-test")

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildOAPDSType(), nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(buildOAPDSType(), cfg), Schema: s},
	}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("DataSource Read returned error: %s", readResp.Diagnostics)
	}

	var model objectStoreAccessPolicyDataSourceModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Name.ValueString() != "ds-oap-test" {
		t.Errorf("expected name=ds-oap-test, got %s", model.Name.ValueString())
	}
	if !model.Enabled.ValueBool() {
		t.Error("expected enabled=true")
	}
	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected ID to be populated")
	}
	if model.Description.ValueString() != "datasource test" {
		t.Errorf("expected description='datasource test', got %s", model.Description.ValueString())
	}
}
