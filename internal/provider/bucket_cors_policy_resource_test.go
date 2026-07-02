package provider

import (
	"context"
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

// newTestBucketCorsPolicyResource creates a bucketCorsPolicyResource wired to the mock server.
func newTestBucketCorsPolicyResource(t *testing.T, ms *testmock.MockServer) *bucketCorsPolicyResource {
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
	return &bucketCorsPolicyResource{client: c}
}

// bucketCorsPolicyResourceSchema returns the parsed schema for the resource.
func bucketCorsPolicyResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &bucketCorsPolicyResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

// buildBucketCorsPolicyType returns the tftypes.Object for the resource schema.
func buildBucketCorsPolicyType() tftypes.Object {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":          tftypes.String,
		"bucket_name": tftypes.String,
		"is_local":    tftypes.Bool,
		"policy_type": tftypes.String,
		"timeouts":    timeoutsType,
	}}
}

// nullBucketCorsPolicyConfig returns a base config map with all attributes null.
func nullBucketCorsPolicyConfig() map[string]tftypes.Value {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return map[string]tftypes.Value{
		"id":          tftypes.NewValue(tftypes.String, nil),
		"bucket_name": tftypes.NewValue(tftypes.String, nil),
		"is_local":    tftypes.NewValue(tftypes.Bool, nil),
		"policy_type": tftypes.NewValue(tftypes.String, nil),
		"timeouts":    tftypes.NewValue(timeoutsType, nil),
	}
}

// bucketCorsPolicyPlanWith returns a plan targeting the given bucket name.
func bucketCorsPolicyPlanWith(t *testing.T, bucketName string) tfsdk.Plan {
	t.Helper()
	s := bucketCorsPolicyResourceSchema(t).Schema
	cfg := nullBucketCorsPolicyConfig()
	cfg["bucket_name"] = tftypes.NewValue(tftypes.String, bucketName)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildBucketCorsPolicyType(), cfg),
		Schema: s,
	}
}

// ---- tests ------------------------------------------------------------------

func TestUnit_BucketCorsPolicyResource_Metadata(t *testing.T) {
	r := &bucketCorsPolicyResource{}
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{}, &resp)
	if resp.TypeName != "flashblade_bucket_cors_policy" {
		t.Errorf("expected type name flashblade_bucket_cors_policy, got %s", resp.TypeName)
	}
}

func TestUnit_BucketCorsPolicyResource_Schema(t *testing.T) {
	s := bucketCorsPolicyResourceSchema(t).Schema

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

	// The wildcard-only toggle exposes no rules attribute.
	if _, exists := s.Attributes["rules"]; exists {
		t.Error("did not expect a rules attribute on the wildcard toggle resource")
	}

	if attr, ok := s.Attributes["is_local"].(resschema.BoolAttribute); !ok || !attr.Computed {
		t.Error("is_local: expected Computed BoolAttribute")
	}
}

func TestUnit_BucketCorsPolicyResource_Lifecycle(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterCorsHandlers(ms.Mux)

	r := newTestBucketCorsPolicyResource(t, ms)
	s := bucketCorsPolicyResourceSchema(t).Schema

	plan := bucketCorsPolicyPlanWith(t, "cors-bucket")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildBucketCorsPolicyType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create returned error: %s", createResp.Diagnostics)
	}

	var model bucketCorsPolicyModel
	if diags := createResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}
	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected non-empty id after Create")
	}

	// Create must have applied the single wildcard rule.
	policy, err := r.client.GetCorsPolicy(context.Background(), "cors-bucket")
	if err != nil {
		t.Fatalf("Get after create: %v", err)
	}
	if len(policy.Rules) != 1 {
		t.Fatalf("expected 1 wildcard rule after Create, got %d: %v", len(policy.Rules), policy.Rules)
	}
	rule := policy.Rules[0]
	if rule.Name != corsWildcardRuleName {
		t.Errorf("expected rule name %q, got %q", corsWildcardRuleName, rule.Name)
	}
	if len(rule.AllowedOrigins) != 1 || rule.AllowedOrigins[0] != "*" {
		t.Errorf("expected wildcard origins, got %v", rule.AllowedOrigins)
	}
	if len(rule.AllowedHeaders) != 1 || rule.AllowedHeaders[0] != "*" {
		t.Errorf("expected wildcard headers, got %v", rule.AllowedHeaders)
	}

	// Read.
	readResp := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}

	// Update is idempotent (re-applies the wildcard rule).
	updateResp := &resource.UpdateResponse{State: readResp.State}
	r.Update(context.Background(), resource.UpdateRequest{Plan: bucketCorsPolicyPlanWith(t, "cors-bucket"), State: readResp.State}, updateResp)
	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update returned error: %s", updateResp.Diagnostics)
	}

	// Delete.
	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: updateResp.State}, deleteResp)
	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete returned error: %s", deleteResp.Diagnostics)
	}
	if _, err := r.client.GetCorsPolicy(context.Background(), "cors-bucket"); err == nil || !client.IsNotFound(err) {
		t.Errorf("expected policy deleted, got: %v", err)
	}
}

func TestUnit_BucketCorsPolicyResource_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterCorsHandlers(ms.Mux)

	store.Seed(&client.CrossOriginResourceSharingPolicy{
		ID:         "cors-imp-1",
		Name:       "imp-bucket-cors",
		Bucket:     client.NamedReference{Name: "imp-bucket"},
		IsLocal:    true,
		PolicyType: "cross-origin-resource-sharing",
		Rules: []client.CorsRule{
			{Name: corsWildcardRuleName, AllowedMethods: corsWildcardMethods, AllowedOrigins: []string{"*"}, AllowedHeaders: []string{"*"}},
		},
	})

	r := newTestBucketCorsPolicyResource(t, ms)
	s := bucketCorsPolicyResourceSchema(t).Schema

	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildBucketCorsPolicyType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "imp-bucket"}, importResp)
	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState returned error: %s", importResp.Diagnostics)
	}

	var model bucketCorsPolicyModel
	if diags := importResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}
	if model.ID.ValueString() != "cors-imp-1" {
		t.Errorf("expected id=cors-imp-1, got %s", model.ID.ValueString())
	}
	if model.BucketName.ValueString() != "imp-bucket" {
		t.Errorf("expected bucket_name=imp-bucket, got %s", model.BucketName.ValueString())
	}
}

func TestUnit_BucketCorsPolicyResource_DriftDetection(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterCorsHandlers(ms.Mux)

	r := newTestBucketCorsPolicyResource(t, ms)
	s := bucketCorsPolicyResourceSchema(t).Schema

	// Create.
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildBucketCorsPolicyType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: bucketCorsPolicyPlanWith(t, "drift-bucket")}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Mutate the store out-of-band to simulate drift on a computed field.
	store.Seed(&client.CrossOriginResourceSharingPolicy{
		ID:         "cors-drift-1",
		Name:       "drift-bucket-cors",
		Bucket:     client.NamedReference{Name: "drift-bucket"},
		IsLocal:    false,
		PolicyType: "changed-type",
		Rules: []client.CorsRule{
			{Name: corsWildcardRuleName, AllowedMethods: corsWildcardMethods, AllowedOrigins: []string{"*"}, AllowedHeaders: []string{"*"}},
		},
	})

	readResp := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}

	var model bucketCorsPolicyModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}
	if model.PolicyType.ValueString() != "changed-type" {
		t.Errorf("expected policy_type reconciled to changed-type after drift Read, got %s", model.PolicyType.ValueString())
	}
}
