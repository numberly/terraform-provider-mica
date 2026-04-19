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

// newTestBAPResource creates a bucketAccessPolicyResource wired to the given mock server.
func newTestBAPResource(t *testing.T, ms *testmock.MockServer) *bucketAccessPolicyResource {
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
	return &bucketAccessPolicyResource{client: c}
}

// bapResourceSchema returns the parsed schema for the bucket access policy resource.
func bapResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &bucketAccessPolicyResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

// buildBucketAccessPolicyType returns the tftypes.Object for the bucket access policy resource schema.
func buildBucketAccessPolicyType() tftypes.Object {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"delete": tftypes.String,
	}}
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":          tftypes.String,
		"bucket_name": tftypes.String,
		"enabled":     tftypes.Bool,
		"timeouts":    timeoutsType,
	}}
}

// nullBAPConfig returns a base config map with all attributes null.
func nullBAPConfig() map[string]tftypes.Value {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"delete": tftypes.String,
	}}
	return map[string]tftypes.Value{
		"id":          tftypes.NewValue(tftypes.String, nil),
		"bucket_name": tftypes.NewValue(tftypes.String, nil),
		"enabled":     tftypes.NewValue(tftypes.Bool, nil),
		"timeouts":    tftypes.NewValue(timeoutsType, nil),
	}
}

// bapPlanWith returns a tfsdk.Plan with the given bucket name.
func bapPlanWith(t *testing.T, bucketName string) tfsdk.Plan {
	t.Helper()
	s := bapResourceSchema(t).Schema
	cfg := nullBAPConfig()
	cfg["bucket_name"] = tftypes.NewValue(tftypes.String, bucketName)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildBucketAccessPolicyType(), cfg),
		Schema: s,
	}
}

// ---- tests ------------------------------------------------------------------

// TestBucketAccessPolicyResource_Metadata verifies type name is "flashblade_bucket_access_policy".
func TestUnit_BucketAccessPolicyResource_Metadata(t *testing.T) {
	r := &bucketAccessPolicyResource{}
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{}, &resp)
	if resp.TypeName != "flashblade_bucket_access_policy" {
		t.Errorf("expected type name flashblade_bucket_access_policy, got %s", resp.TypeName)
	}
}

// TestBucketAccessPolicyResource_Schema verifies schema properties.
func TestUnit_BucketAccessPolicyResource_Schema(t *testing.T) {
	s := bapResourceSchema(t).Schema

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

	// id: Computed.
	idAttr, ok := s.Attributes["id"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("id attribute not found or wrong type")
	}
	if !idAttr.Computed {
		t.Error("id: expected Computed=true")
	}

	// enabled: Computed.
	enabledAttr, ok := s.Attributes["enabled"].(resschema.BoolAttribute)
	if !ok {
		t.Fatal("enabled attribute not found or wrong type")
	}
	if !enabledAttr.Computed {
		t.Error("enabled: expected Computed=true")
	}
}

// TestBucketAccessPolicyResource_Create verifies POST creates a policy.
func TestUnit_BucketAccessPolicyResource_Create(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterBucketAccessPolicyHandlers(ms.Mux)

	r := newTestBAPResource(t, ms)
	s := bapResourceSchema(t).Schema

	plan := bapPlanWith(t, "test-bucket")
	resp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildBucketAccessPolicyType(), nil), Schema: s},
	}

	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Create returned error: %s", resp.Diagnostics)
	}

	var model bucketAccessPolicyModel
	if diags := resp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected non-empty id after Create")
	}
	if model.BucketName.ValueString() != "test-bucket" {
		t.Errorf("expected bucket_name=test-bucket, got %s", model.BucketName.ValueString())
	}
	if model.Enabled.ValueBool() != true {
		t.Error("expected enabled=true after Create")
	}
}

// TestBucketAccessPolicyResource_Read verifies GET retrieves policy by bucket name.
func TestUnit_BucketAccessPolicyResource_Read(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterBucketAccessPolicyHandlers(ms.Mux)

	r := newTestBAPResource(t, ms)
	s := bapResourceSchema(t).Schema

	// Create first.
	plan := bapPlanWith(t, "read-bucket")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildBucketAccessPolicyType(), nil), Schema: s},
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

	var model bucketAccessPolicyModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.BucketName.ValueString() != "read-bucket" {
		t.Errorf("expected bucket_name=read-bucket, got %s", model.BucketName.ValueString())
	}
	if model.Enabled.ValueBool() != true {
		t.Error("expected enabled=true after Read")
	}
}

// TestBucketAccessPolicyResource_ReadNotFound verifies Read removes resource when not found.
func TestUnit_BucketAccessPolicyResource_ReadNotFound(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterBucketAccessPolicyHandlers(ms.Mux)

	r := newTestBAPResource(t, ms)
	s := bapResourceSchema(t).Schema

	// Build state for a non-existent policy.
	cfg := nullBAPConfig()
	cfg["id"] = tftypes.NewValue(tftypes.String, "bap-999")
	cfg["bucket_name"] = tftypes.NewValue(tftypes.String, "ghost-bucket")
	cfg["enabled"] = tftypes.NewValue(tftypes.Bool, true)

	state := tfsdk.State{
		Raw:    tftypes.NewValue(buildBucketAccessPolicyType(), cfg),
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

// TestBucketAccessPolicyResource_Delete verifies DELETE succeeds.
func TestUnit_BucketAccessPolicyResource_Delete(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterBucketAccessPolicyHandlers(ms.Mux)

	r := newTestBAPResource(t, ms)
	s := bapResourceSchema(t).Schema

	// Create.
	plan := bapPlanWith(t, "del-bucket")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildBucketAccessPolicyType(), nil), Schema: s},
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
	_, err := r.client.GetBucketAccessPolicy(context.Background(), "del-bucket")
	if err == nil || !client.IsNotFound(err) {
		t.Errorf("expected policy to be deleted, got: %v", err)
	}
}

// TestBucketAccessPolicyResource_ImportState verifies import by bucket name.
func TestUnit_BucketAccessPolicyResource_ImportState(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterBucketAccessPolicyHandlers(ms.Mux)

	// Seed a policy so import can find it.
	store.Seed(&client.BucketAccessPolicy{
		ID:         "bap-imp-1",
		Name:       "imp-bucket",
		Bucket:     client.NamedReference{Name: "imp-bucket"},
		Enabled:    true,
		IsLocal:    true,
		PolicyType: "s3",
	})

	r := newTestBAPResource(t, ms)
	s := bapResourceSchema(t).Schema

	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildBucketAccessPolicyType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "imp-bucket"}, importResp)

	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState returned error: %s", importResp.Diagnostics)
	}

	var model bucketAccessPolicyModel
	if diags := importResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.ValueString() != "bap-imp-1" {
		t.Errorf("expected id=bap-imp-1, got %s", model.ID.ValueString())
	}
	if model.BucketName.ValueString() != "imp-bucket" {
		t.Errorf("expected bucket_name=imp-bucket, got %s", model.BucketName.ValueString())
	}
	if model.Enabled.ValueBool() != true {
		t.Error("expected enabled=true after import")
	}
}
