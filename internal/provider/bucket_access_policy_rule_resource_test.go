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

// newTestBAPRuleResource creates a bucketAccessPolicyRuleResource wired to the given mock server.
func newTestBAPRuleResource(t *testing.T, ms *testmock.MockServer) *bucketAccessPolicyRuleResource {
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
	return &bucketAccessPolicyRuleResource{client: c}
}

// bapRuleResourceSchema returns the parsed schema for the bucket access policy rule resource.
func bapRuleResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &bucketAccessPolicyRuleResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

// buildBAPRuleType returns the tftypes.Object for the bucket access policy rule resource schema.
func buildBAPRuleType() tftypes.Object {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"delete": tftypes.String,
	}}
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"name":        tftypes.String,
		"bucket_name": tftypes.String,
		"actions":     tftypes.List{ElementType: tftypes.String},
		"effect":      tftypes.String,
		"principals":  tftypes.List{ElementType: tftypes.String},
		"resources":   tftypes.List{ElementType: tftypes.String},
		"timeouts":    timeoutsType,
	}}
}

// nullBAPRuleConfig returns a base config map with all attributes null.
func nullBAPRuleConfig() map[string]tftypes.Value {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"delete": tftypes.String,
	}}
	return map[string]tftypes.Value{
		"name":        tftypes.NewValue(tftypes.String, nil),
		"bucket_name": tftypes.NewValue(tftypes.String, nil),
		"actions":     tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"effect":      tftypes.NewValue(tftypes.String, nil),
		"principals":  tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"resources":   tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"timeouts":    tftypes.NewValue(timeoutsType, nil),
	}
}

// bapRulePlanWith returns a tfsdk.Plan for a bucket access policy rule.
func bapRulePlanWith(t *testing.T, bucketName string, actions, principals, resources []string) tfsdk.Plan {
	t.Helper()
	s := bapRuleResourceSchema(t).Schema
	cfg := nullBAPRuleConfig()
	cfg["bucket_name"] = tftypes.NewValue(tftypes.String, bucketName)

	actionVals := make([]tftypes.Value, len(actions))
	for i, a := range actions {
		actionVals[i] = tftypes.NewValue(tftypes.String, a)
	}
	cfg["actions"] = tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, actionVals)

	principalVals := make([]tftypes.Value, len(principals))
	for i, p := range principals {
		principalVals[i] = tftypes.NewValue(tftypes.String, p)
	}
	cfg["principals"] = tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, principalVals)

	resourceVals := make([]tftypes.Value, len(resources))
	for i, r := range resources {
		resourceVals[i] = tftypes.NewValue(tftypes.String, r)
	}
	cfg["resources"] = tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, resourceVals)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildBAPRuleType(), cfg),
		Schema: s,
	}
}

// createTestBAPPolicy creates a bucket access policy in the mock server for rule tests.
func createTestBAPPolicy(t *testing.T, ms *testmock.MockServer, bucketName string) {
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
	_, err = c.PostBucketAccessPolicy(context.Background(), bucketName, client.BucketAccessPolicyPost{})
	if err != nil {
		t.Fatalf("PostBucketAccessPolicy(%s): %v", bucketName, err)
	}
}

// ---- tests ------------------------------------------------------------------

// TestBucketAccessPolicyRuleResource_Metadata verifies type name.
func TestBucketAccessPolicyRuleResource_Metadata(t *testing.T) {
	r := &bucketAccessPolicyRuleResource{}
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{}, &resp)
	if resp.TypeName != "flashblade_bucket_access_policy_rule" {
		t.Errorf("expected type name flashblade_bucket_access_policy_rule, got %s", resp.TypeName)
	}
}

// TestBucketAccessPolicyRuleResource_Schema verifies schema properties.
func TestBucketAccessPolicyRuleResource_Schema(t *testing.T) {
	s := bapRuleResourceSchema(t).Schema

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

	// actions: Required.
	actionsAttr, ok := s.Attributes["actions"].(resschema.ListAttribute)
	if !ok {
		t.Fatal("actions attribute not found or wrong type")
	}
	if !actionsAttr.Required {
		t.Error("actions: expected Required=true")
	}

	// name: Computed.
	nameAttr, ok := s.Attributes["name"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("name attribute not found or wrong type")
	}
	if !nameAttr.Computed {
		t.Error("name: expected Computed=true")
	}

	// effect: Computed (read-only, set by API).
	effectAttr, ok := s.Attributes["effect"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("effect attribute not found or wrong type")
	}
	if !effectAttr.Computed {
		t.Error("effect: expected Computed=true")
	}
}

// TestBucketAccessPolicyRuleResource_Create verifies POST creates a rule.
func TestBucketAccessPolicyRuleResource_Create(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterBucketAccessPolicyHandlers(ms.Mux)

	// Create parent policy first.
	createTestBAPPolicy(t, ms, "rule-test-bucket")

	r := newTestBAPRuleResource(t, ms)
	s := bapRuleResourceSchema(t).Schema

	plan := bapRulePlanWith(t, "rule-test-bucket",
		[]string{"s3:GetObject", "s3:PutObject"},
		[]string{"*"},
		[]string{"rule-test-bucket/*"},
	)
	resp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildBAPRuleType(), nil), Schema: s},
	}

	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Create returned error: %s", resp.Diagnostics)
	}

	var model bucketAccessPolicyRuleModel
	if diags := resp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Name.IsNull() || model.Name.ValueString() == "" {
		t.Error("expected non-empty name after Create")
	}
	if model.BucketName.ValueString() != "rule-test-bucket" {
		t.Errorf("expected bucket_name=rule-test-bucket, got %s", model.BucketName.ValueString())
	}
	if model.Effect.ValueString() != "allow" {
		t.Errorf("expected effect=allow, got %s", model.Effect.ValueString())
	}

	var actions []string
	model.Actions.ElementsAs(context.Background(), &actions, false)
	if len(actions) != 2 {
		t.Errorf("expected 2 actions, got %d", len(actions))
	}

	var principals []string
	model.Principals.ElementsAs(context.Background(), &principals, false)
	if len(principals) != 1 || principals[0] != "*" {
		t.Errorf("expected principals=[*], got %v", principals)
	}
}

// TestBucketAccessPolicyRuleResource_Read verifies GET retrieves rule.
func TestBucketAccessPolicyRuleResource_Read(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterBucketAccessPolicyHandlers(ms.Mux)

	createTestBAPPolicy(t, ms, "read-rule-bucket")

	r := newTestBAPRuleResource(t, ms)
	s := bapRuleResourceSchema(t).Schema

	// Create first.
	plan := bapRulePlanWith(t, "read-rule-bucket",
		[]string{"s3:GetObject"},
		[]string{"*"},
		[]string{"read-rule-bucket/*"},
	)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildBAPRuleType(), nil), Schema: s},
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

	var model bucketAccessPolicyRuleModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.BucketName.ValueString() != "read-rule-bucket" {
		t.Errorf("expected bucket_name=read-rule-bucket, got %s", model.BucketName.ValueString())
	}
	if model.Effect.ValueString() != "allow" {
		t.Errorf("expected effect=allow, got %s", model.Effect.ValueString())
	}
}

// TestBucketAccessPolicyRuleResource_ReadNotFound verifies Read removes resource when not found.
func TestBucketAccessPolicyRuleResource_ReadNotFound(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterBucketAccessPolicyHandlers(ms.Mux)

	r := newTestBAPRuleResource(t, ms)
	s := bapRuleResourceSchema(t).Schema

	// Build state for a non-existent rule.
	cfg := nullBAPRuleConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "ghost-rule")
	cfg["bucket_name"] = tftypes.NewValue(tftypes.String, "ghost-bucket")
	cfg["effect"] = tftypes.NewValue(tftypes.String, "allow")
	cfg["actions"] = tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
		tftypes.NewValue(tftypes.String, "s3:GetObject"),
	})
	cfg["principals"] = tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
		tftypes.NewValue(tftypes.String, "*"),
	})
	cfg["resources"] = tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
		tftypes.NewValue(tftypes.String, "ghost-bucket/*"),
	})

	state := tfsdk.State{
		Raw:    tftypes.NewValue(buildBAPRuleType(), cfg),
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

// TestBucketAccessPolicyRuleResource_Delete verifies DELETE succeeds.
func TestBucketAccessPolicyRuleResource_Delete(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterBucketAccessPolicyHandlers(ms.Mux)

	createTestBAPPolicy(t, ms, "del-rule-bucket")

	r := newTestBAPRuleResource(t, ms)
	s := bapRuleResourceSchema(t).Schema

	// Create.
	plan := bapRulePlanWith(t, "del-rule-bucket",
		[]string{"s3:GetObject"},
		[]string{"*"},
		[]string{"del-rule-bucket/*"},
	)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildBAPRuleType(), nil), Schema: s},
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
}

// TestBucketAccessPolicyRuleResource_ImportState verifies import by "bucketName/ruleName".
func TestBucketAccessPolicyRuleResource_ImportState(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterBucketAccessPolicyHandlers(ms.Mux)

	// Seed a policy with a rule so import can find it.
	store.Seed(&client.BucketAccessPolicy{
		ID:         "bap-imp-rule",
		Name:       "imp-rule-bucket",
		Bucket:     client.NamedReference{Name: "imp-rule-bucket"},
		Enabled:    true,
		IsLocal:    true,
		PolicyType: "s3",
		Rules: []client.BucketAccessPolicyRule{
			{
				Name:    "imp-rule",
				Actions: []string{"s3:GetObject"},
				Effect:  "allow",
				Principals: client.BucketAccessPolicyPrincipals{
					All: []string{"*"},
				},
				Resources: []string{"imp-rule-bucket/*"},
				Policy:    &client.NamedReference{Name: "imp-rule-bucket"},
			},
		},
	})

	r := newTestBAPRuleResource(t, ms)
	s := bapRuleResourceSchema(t).Schema

	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildBAPRuleType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "imp-rule-bucket/imp-rule"}, importResp)

	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState returned error: %s", importResp.Diagnostics)
	}

	var model bucketAccessPolicyRuleModel
	if diags := importResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Name.ValueString() != "imp-rule" {
		t.Errorf("expected name=imp-rule, got %s", model.Name.ValueString())
	}
	if model.BucketName.ValueString() != "imp-rule-bucket" {
		t.Errorf("expected bucket_name=imp-rule-bucket, got %s", model.BucketName.ValueString())
	}
	if model.Effect.ValueString() != "allow" {
		t.Errorf("expected effect=allow, got %s", model.Effect.ValueString())
	}

	var actions []string
	model.Actions.ElementsAs(context.Background(), &actions, false)
	if len(actions) != 1 || actions[0] != "s3:GetObject" {
		t.Errorf("expected actions=[s3:GetObject], got %v", actions)
	}
}
