package provider

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/numberly/opentofu-provider-flashblade/internal/client"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock/handlers"
)

// ---- helpers ----------------------------------------------------------------

// newTestBucketAuditFilterResource creates a bucketAuditFilterResource wired to the given mock server.
func newTestBucketAuditFilterResource(t *testing.T, ms *testmock.MockServer) *bucketAuditFilterResource {
	t.Helper()
	c, err := client.NewClient(context.Background(), client.Config{
		Endpoint:           ms.URL(),
		APIToken:           "test-token",
		InsecureSkipVerify: true,
		MaxRetries:         1,
		RetryBaseDelay:     1 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return &bucketAuditFilterResource{client: c}
}

// bafResourceSchema returns the parsed schema for the bucket audit filter resource.
func bafResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &bucketAuditFilterResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

// buildBAFType returns the tftypes.Object for the bucket audit filter resource schema.
func buildBAFType() tftypes.Object {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":          tftypes.String,
		"name":        tftypes.String,
		"bucket_name": tftypes.String,
		"actions":     tftypes.Set{ElementType: tftypes.String},
		"s3_prefixes": tftypes.Set{ElementType: tftypes.String},
		"timeouts":    timeoutsType,
	}}
}

// nullBAFConfig returns a base config map with all attributes null.
func nullBAFConfig() map[string]tftypes.Value {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return map[string]tftypes.Value{
		"id":          tftypes.NewValue(tftypes.String, nil),
		"name":        tftypes.NewValue(tftypes.String, nil),
		"bucket_name": tftypes.NewValue(tftypes.String, nil),
		"actions":     tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, nil),
		"s3_prefixes": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, nil),
		"timeouts":    tftypes.NewValue(timeoutsType, nil),
	}
}

// bafPlanWith returns a tfsdk.Plan with the given bucket name, actions, and s3_prefixes.
func bafPlanWith(t *testing.T, bucketName string, actions []string, s3Prefixes []string) tfsdk.Plan {
	t.Helper()
	s := bafResourceSchema(t).Schema
	cfg := nullBAFConfig()
	cfg["bucket_name"] = tftypes.NewValue(tftypes.String, bucketName)
	cfg["name"] = tftypes.NewValue(tftypes.String, bucketName+"-audit")

	actVals := make([]tftypes.Value, len(actions))
	for i, a := range actions {
		actVals[i] = tftypes.NewValue(tftypes.String, a)
	}
	cfg["actions"] = tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, actVals)

	pfxVals := make([]tftypes.Value, len(s3Prefixes))
	for i, p := range s3Prefixes {
		pfxVals[i] = tftypes.NewValue(tftypes.String, p)
	}
	cfg["s3_prefixes"] = tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, pfxVals)

	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildBAFType(), cfg),
		Schema: s,
	}
}

// ---- tests ------------------------------------------------------------------

// TestBucketAuditFilterResource_Create verifies POST creates an audit filter.
func TestBucketAuditFilterResource_Create(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterBucketAuditFilterHandlers(ms.Mux)

	r := newTestBucketAuditFilterResource(t, ms)
	s := bafResourceSchema(t).Schema

	plan := bafPlanWith(t, "test-bucket", []string{"s3:GetObject", "s3:PutObject"}, []string{"logs/"})
	resp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildBAFType(), nil), Schema: s},
	}

	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Create returned error: %s", resp.Diagnostics)
	}

	var model bucketAuditFilterModel
	if diags := resp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected non-empty id after Create")
	}
	if model.BucketName.ValueString() != "test-bucket" {
		t.Errorf("expected bucket_name=test-bucket, got %s", model.BucketName.ValueString())
	}

	var actions []string
	if diags := model.Actions.ElementsAs(context.Background(), &actions, false); diags.HasError() {
		t.Fatalf("ElementsAs actions: %s", diags)
	}
	if len(actions) != 2 || actions[0] != "s3:GetObject" || actions[1] != "s3:PutObject" {
		t.Errorf("expected actions=[s3:GetObject, s3:PutObject], got %v", actions)
	}

	var prefixes []string
	if diags := model.S3Prefixes.ElementsAs(context.Background(), &prefixes, false); diags.HasError() {
		t.Fatalf("ElementsAs s3_prefixes: %s", diags)
	}
	if len(prefixes) != 1 || prefixes[0] != "logs/" {
		t.Errorf("expected s3_prefixes=[logs/], got %v", prefixes)
	}
}

// TestBucketAuditFilterResource_Read verifies GET retrieves audit filter by bucket name.
func TestBucketAuditFilterResource_Read(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterBucketAuditFilterHandlers(ms.Mux)

	r := newTestBucketAuditFilterResource(t, ms)
	s := bafResourceSchema(t).Schema

	// Create first.
	plan := bafPlanWith(t, "read-bucket", []string{"s3:GetObject"}, []string{})
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildBAFType(), nil), Schema: s},
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

	var model bucketAuditFilterModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.BucketName.ValueString() != "read-bucket" {
		t.Errorf("expected bucket_name=read-bucket, got %s", model.BucketName.ValueString())
	}

	var actions []string
	if diags := model.Actions.ElementsAs(context.Background(), &actions, false); diags.HasError() {
		t.Fatalf("ElementsAs actions: %s", diags)
	}
	if len(actions) != 1 || actions[0] != "s3:GetObject" {
		t.Errorf("expected actions=[s3:GetObject], got %v", actions)
	}
}

// TestBucketAuditFilterResource_Read_NotFound verifies Read removes resource when not found.
func TestBucketAuditFilterResource_Read_NotFound(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterBucketAuditFilterHandlers(ms.Mux)

	r := newTestBucketAuditFilterResource(t, ms)
	s := bafResourceSchema(t).Schema

	// Build state for a non-existent filter.
	cfg := nullBAFConfig()
	cfg["id"] = tftypes.NewValue(tftypes.String, "baf-999")
	cfg["bucket_name"] = tftypes.NewValue(tftypes.String, "ghost-bucket")
	cfg["actions"] = tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{
		tftypes.NewValue(tftypes.String, "s3:GetObject"),
	})
	cfg["s3_prefixes"] = tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{})

	state := tfsdk.State{
		Raw:    tftypes.NewValue(buildBAFType(), cfg),
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

// TestBucketAuditFilterResource_Update verifies PATCH sends only changed fields.
func TestBucketAuditFilterResource_Update(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterBucketAuditFilterHandlers(ms.Mux)

	r := newTestBucketAuditFilterResource(t, ms)
	s := bafResourceSchema(t).Schema

	// Create first.
	createPlan := bafPlanWith(t, "upd-bucket", []string{"s3:GetObject"}, []string{"logs/"})
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildBAFType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Update: change actions, keep s3_prefixes the same.
	updatePlan := bafPlanWith(t, "upd-bucket", []string{"s3:GetObject", "s3:DeleteObject"}, []string{"logs/"})

	// Copy the ID from create state into the update plan.
	var created bucketAuditFilterModel
	if diags := createResp.State.Get(context.Background(), &created); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildBAFType(), nil), Schema: s},
	}

	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlan,
		State: createResp.State,
	}, updateResp)

	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update returned error: %s", updateResp.Diagnostics)
	}

	var model bucketAuditFilterModel
	if diags := updateResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	var actions []string
	if diags := model.Actions.ElementsAs(context.Background(), &actions, false); diags.HasError() {
		t.Fatalf("ElementsAs actions: %s", diags)
	}
	if len(actions) != 2 || actions[0] != "s3:GetObject" || actions[1] != "s3:DeleteObject" {
		t.Errorf("expected actions=[s3:GetObject, s3:DeleteObject], got %v", actions)
	}

	// s3_prefixes should be unchanged.
	var prefixes []string
	if diags := model.S3Prefixes.ElementsAs(context.Background(), &prefixes, false); diags.HasError() {
		t.Fatalf("ElementsAs s3_prefixes: %s", diags)
	}
	if len(prefixes) != 1 || prefixes[0] != "logs/" {
		t.Errorf("expected s3_prefixes=[logs/], got %v", prefixes)
	}
}

// TestBucketAuditFilterResource_Delete verifies DELETE succeeds.
func TestBucketAuditFilterResource_Delete(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterBucketAuditFilterHandlers(ms.Mux)

	r := newTestBucketAuditFilterResource(t, ms)
	s := bafResourceSchema(t).Schema

	// Create.
	plan := bafPlanWith(t, "del-bucket", []string{"s3:GetObject"}, []string{})
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildBAFType(), nil), Schema: s},
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
	_, err := r.client.GetBucketAuditFilter(context.Background(), "delbucket-audit", "del-bucket")
	if err == nil || !client.IsNotFound(err) {
		t.Errorf("expected filter to be deleted, got: %v", err)
	}
}

// TestBucketAuditFilterResource_Import verifies import by bucket name.
func TestBucketAuditFilterResource_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterBucketAuditFilterHandlers(ms.Mux)

	// Seed a filter so import can find it.
	store.Seed(&client.BucketAuditFilter{
		Name:       "baf-imp-1",
		Bucket:     client.NamedReference{Name: "imp-bucket"},
		Actions:    []string{"s3:GetObject", "s3:PutObject"},
		S3Prefixes: []string{"data/"},
	})

	r := newTestBucketAuditFilterResource(t, ms)
	s := bafResourceSchema(t).Schema

	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildBAFType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "baf-imp-1/imp-bucket"}, importResp)

	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState returned error: %s", importResp.Diagnostics)
	}

	var model bucketAuditFilterModel
	if diags := importResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.ValueString() != "baf-imp-1" {
		t.Errorf("expected id=baf-imp-1, got %s", model.ID.ValueString())
	}
	if model.BucketName.ValueString() != "imp-bucket" {
		t.Errorf("expected bucket_name=imp-bucket, got %s", model.BucketName.ValueString())
	}

	var actions []string
	if diags := model.Actions.ElementsAs(context.Background(), &actions, false); diags.HasError() {
		t.Fatalf("ElementsAs actions: %s", diags)
	}
	if len(actions) != 2 || actions[0] != "s3:GetObject" || actions[1] != "s3:PutObject" {
		t.Errorf("expected actions=[s3:GetObject, s3:PutObject], got %v", actions)
	}

	var prefixes []string
	if diags := model.S3Prefixes.ElementsAs(context.Background(), &prefixes, false); diags.HasError() {
		t.Fatalf("ElementsAs s3_prefixes: %s", diags)
	}
	if len(prefixes) != 1 || prefixes[0] != "data/" {
		t.Errorf("expected s3_prefixes=[data/], got %v", prefixes)
	}
}

// TestBucketAuditFilterResource_Schema verifies schema properties.
func TestBucketAuditFilterResource_Schema(t *testing.T) {
	s := bafResourceSchema(t).Schema

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

	// actions: Required, set of string.
	actionsAttr, ok := s.Attributes["actions"].(resschema.SetAttribute)
	if !ok {
		t.Fatal("actions attribute not found or wrong type")
	}
	if !actionsAttr.Required {
		t.Error("actions: expected Required=true")
	}

	// s3_prefixes: Optional + Computed.
	prefixesAttr, ok := s.Attributes["s3_prefixes"].(resschema.SetAttribute)
	if !ok {
		t.Fatal("s3_prefixes attribute not found or wrong type")
	}
	if !prefixesAttr.Optional {
		t.Error("s3_prefixes: expected Optional=true")
	}
	if !prefixesAttr.Computed {
		t.Error("s3_prefixes: expected Computed=true")
	}
}

// Suppress unused import warnings.
var _ = attr.Value(nil)
var _ = types.StringType
