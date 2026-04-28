package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/numberly/terraform-provider-mica/internal/client"
	"github.com/numberly/terraform-provider-mica/internal/testmock"
	"github.com/numberly/terraform-provider-mica/internal/testmock/handlers"
)

// ---- helpers ----------------------------------------------------------------

// newTestLogTargetObjectStoreResource creates a logTargetObjectStoreResource wired to the given mock server.
func newTestLogTargetObjectStoreResource(t *testing.T, ms *testmock.MockServer) *logTargetObjectStoreResource {
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
	return &logTargetObjectStoreResource{client: c}
}

// logTargetObjectStoreResourceSchema returns the parsed schema for the log target object store resource.
func logTargetObjectStoreResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &logTargetObjectStoreResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

// buildLogTargetObjectStoreType returns the tftypes.Object for the log target object store resource schema.
func buildLogTargetObjectStoreType() tftypes.Object {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":                  tftypes.String,
		"name":                tftypes.String,
		"bucket_name":         tftypes.String,
		"log_name_prefix":     tftypes.String,
		"log_rotate_duration": tftypes.Number,
		"timeouts":            timeoutsType,
	}}
}

// nullLogTargetObjectStoreConfig returns a base config map with all attributes null.
func nullLogTargetObjectStoreConfig() map[string]tftypes.Value {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return map[string]tftypes.Value{
		"id":                  tftypes.NewValue(tftypes.String, nil),
		"name":                tftypes.NewValue(tftypes.String, nil),
		"bucket_name":         tftypes.NewValue(tftypes.String, nil),
		"log_name_prefix":     tftypes.NewValue(tftypes.String, nil),
		"log_rotate_duration": tftypes.NewValue(tftypes.Number, nil),
		"timeouts":            tftypes.NewValue(timeoutsType, nil),
	}
}

// logTargetObjectStorePlanWith builds a tfsdk.Plan for the log target object store resource.
func logTargetObjectStorePlanWith(t *testing.T, name, bucketName, logNamePrefix string, logRotateDuration *int64) tfsdk.Plan {
	t.Helper()
	s := logTargetObjectStoreResourceSchema(t).Schema
	cfg := nullLogTargetObjectStoreConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, name)
	cfg["bucket_name"] = tftypes.NewValue(tftypes.String, bucketName)
	if logNamePrefix != "" {
		cfg["log_name_prefix"] = tftypes.NewValue(tftypes.String, logNamePrefix)
	}
	if logRotateDuration != nil {
		cfg["log_rotate_duration"] = tftypes.NewValue(tftypes.Number, *logRotateDuration)
	}
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildLogTargetObjectStoreType(), cfg),
		Schema: s,
	}
}

// ---- tests ------------------------------------------------------------------

// TestUnit_LogTargetObjectStoreResource_Lifecycle exercises Create->Read->Update->Delete.
func TestUnit_LogTargetObjectStoreResource_Lifecycle(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterLogTargetObjectStoreHandlers(ms.Mux)

	r := newTestLogTargetObjectStoreResource(t, ms)
	s := logTargetObjectStoreResourceSchema(t).Schema

	dur := int64(3600000)
	createPlan := logTargetObjectStorePlanWith(t, "test-ltos", "audit-bucket", "logs/", &dur)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildLogTargetObjectStoreType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	var createModel logTargetObjectStoreModel
	if diags := createResp.State.Get(context.Background(), &createModel); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}
	if createModel.Name.ValueString() != "test-ltos" {
		t.Errorf("Create: expected name=test-ltos, got %s", createModel.Name.ValueString())
	}
	if createModel.BucketName.ValueString() != "audit-bucket" {
		t.Errorf("Create: expected bucket_name=audit-bucket, got %s", createModel.BucketName.ValueString())
	}
	if createModel.LogNamePrefix.ValueString() != "logs/" {
		t.Errorf("Create: expected log_name_prefix=logs/, got %s", createModel.LogNamePrefix.ValueString())
	}
	if createModel.LogRotateDuration.ValueInt64() != 3600000 {
		t.Errorf("Create: expected log_rotate_duration=3600000, got %d", createModel.LogRotateDuration.ValueInt64())
	}
	if createModel.ID.IsNull() || createModel.ID.ValueString() == "" {
		t.Error("Create: expected non-empty ID")
	}

	// Read post-create (0-diff check).
	readResp := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read post-create: %s", readResp.Diagnostics)
	}

	// Update bucket_name.
	dur2 := int64(7200000)
	updatePlan := logTargetObjectStorePlanWith(t, "test-ltos", "new-audit-bucket", "logs/", &dur2)
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildLogTargetObjectStoreType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlan,
		State: readResp.State,
	}, updateResp)
	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update: %s", updateResp.Diagnostics)
	}

	var updateModel logTargetObjectStoreModel
	if diags := updateResp.State.Get(context.Background(), &updateModel); diags.HasError() {
		t.Fatalf("Get update state: %s", diags)
	}
	if updateModel.BucketName.ValueString() != "new-audit-bucket" {
		t.Errorf("Update: expected bucket_name=new-audit-bucket, got %s", updateModel.BucketName.ValueString())
	}
	if updateModel.LogRotateDuration.ValueInt64() != 7200000 {
		t.Errorf("Update: expected log_rotate_duration=7200000, got %d", updateModel.LogRotateDuration.ValueInt64())
	}

	// Delete.
	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: updateResp.State}, deleteResp)
	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete: %s", deleteResp.Diagnostics)
	}
	_, err := r.client.GetLogTargetObjectStore(context.Background(), "test-ltos")
	if err == nil || !client.IsNotFound(err) {
		t.Errorf("expected log target object store to be deleted, got: %v", err)
	}
}

// TestUnit_LogTargetObjectStoreResource_Import verifies ImportState populates all attributes correctly.
func TestUnit_LogTargetObjectStoreResource_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterLogTargetObjectStoreHandlers(ms.Mux)

	r := newTestLogTargetObjectStoreResource(t, ms)
	s := logTargetObjectStoreResourceSchema(t).Schema

	dur := int64(3600000)
	createPlan := logTargetObjectStorePlanWith(t, "import-ltos", "audit-bucket", "audit/", &dur)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildLogTargetObjectStoreType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	var createModel logTargetObjectStoreModel
	if diags := createResp.State.Get(context.Background(), &createModel); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}

	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildLogTargetObjectStoreType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "import-ltos"}, importResp)
	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState: %s", importResp.Diagnostics)
	}

	var importedModel logTargetObjectStoreModel
	if diags := importResp.State.Get(context.Background(), &importedModel); diags.HasError() {
		t.Fatalf("Get imported state: %s", diags)
	}

	if importedModel.Name.ValueString() != createModel.Name.ValueString() {
		t.Errorf("name mismatch: create=%s import=%s", createModel.Name.ValueString(), importedModel.Name.ValueString())
	}
	if importedModel.ID.ValueString() != createModel.ID.ValueString() {
		t.Errorf("id mismatch: create=%s import=%s", createModel.ID.ValueString(), importedModel.ID.ValueString())
	}
	if importedModel.BucketName.ValueString() != createModel.BucketName.ValueString() {
		t.Errorf("bucket_name mismatch: create=%s import=%s", createModel.BucketName.ValueString(), importedModel.BucketName.ValueString())
	}
	if importedModel.LogNamePrefix.ValueString() != createModel.LogNamePrefix.ValueString() {
		t.Errorf("log_name_prefix mismatch: create=%s import=%s", createModel.LogNamePrefix.ValueString(), importedModel.LogNamePrefix.ValueString())
	}
	if importedModel.LogRotateDuration.ValueInt64() != createModel.LogRotateDuration.ValueInt64() {
		t.Errorf("log_rotate_duration mismatch: create=%d import=%d", createModel.LogRotateDuration.ValueInt64(), importedModel.LogRotateDuration.ValueInt64())
	}
}

// TestUnit_LogTargetObjectStoreResource_DriftDetection verifies Read after Create shows no attribute drift.
func TestUnit_LogTargetObjectStoreResource_DriftDetection(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterLogTargetObjectStoreHandlers(ms.Mux)

	r := newTestLogTargetObjectStoreResource(t, ms)
	s := logTargetObjectStoreResourceSchema(t).Schema

	dur := int64(3600000)
	plan := logTargetObjectStorePlanWith(t, "drift-ltos", "audit-bucket", "logs/", &dur)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildLogTargetObjectStoreType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	readResp := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}

	var beforeModel, afterModel logTargetObjectStoreModel
	if diags := createResp.State.Get(context.Background(), &beforeModel); diags.HasError() {
		t.Fatalf("Get before state: %s", diags)
	}
	if diags := readResp.State.Get(context.Background(), &afterModel); diags.HasError() {
		t.Fatalf("Get after state: %s", diags)
	}

	if beforeModel.ID.ValueString() != afterModel.ID.ValueString() {
		t.Errorf("ID changed after Read: %s -> %s", beforeModel.ID.ValueString(), afterModel.ID.ValueString())
	}
	if beforeModel.Name.ValueString() != afterModel.Name.ValueString() {
		t.Errorf("Name changed after Read: %s -> %s", beforeModel.Name.ValueString(), afterModel.Name.ValueString())
	}
	if beforeModel.BucketName.ValueString() != afterModel.BucketName.ValueString() {
		t.Errorf("BucketName changed after Read: %s -> %s", beforeModel.BucketName.ValueString(), afterModel.BucketName.ValueString())
	}
	if beforeModel.LogNamePrefix.ValueString() != afterModel.LogNamePrefix.ValueString() {
		t.Errorf("LogNamePrefix changed after Read: %s -> %s", beforeModel.LogNamePrefix.ValueString(), afterModel.LogNamePrefix.ValueString())
	}
	if beforeModel.LogRotateDuration.ValueInt64() != afterModel.LogRotateDuration.ValueInt64() {
		t.Errorf("LogRotateDuration changed after Read: %d -> %d", beforeModel.LogRotateDuration.ValueInt64(), afterModel.LogRotateDuration.ValueInt64())
	}
}

// ---- data source tests -------------------------------------------------------

// newTestLogTargetObjectStoreDataSource creates a logTargetObjectStoreDataSource wired to the given mock server.
func newTestLogTargetObjectStoreDataSource(t *testing.T, ms *testmock.MockServer) *logTargetObjectStoreDataSource {
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
	return &logTargetObjectStoreDataSource{client: c}
}

// logTargetObjectStoreDSSchema returns the schema for the log target object store data source.
func logTargetObjectStoreDSSchema(t *testing.T) datasource.SchemaResponse {
	t.Helper()
	d := &logTargetObjectStoreDataSource{}
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	return resp
}

// buildLogTargetObjectStoreDSType returns the tftypes.Object for the log target object store data source.
func buildLogTargetObjectStoreDSType() tftypes.Object {
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":                  tftypes.String,
		"name":                tftypes.String,
		"bucket_name":         tftypes.String,
		"log_name_prefix":     tftypes.String,
		"log_rotate_duration": tftypes.Number,
	}}
}

// nullLogTargetObjectStoreDSConfig returns a base config map with all data source attributes null.
func nullLogTargetObjectStoreDSConfig() map[string]tftypes.Value {
	return map[string]tftypes.Value{
		"id":                  tftypes.NewValue(tftypes.String, nil),
		"name":                tftypes.NewValue(tftypes.String, nil),
		"bucket_name":         tftypes.NewValue(tftypes.String, nil),
		"log_name_prefix":     tftypes.NewValue(tftypes.String, nil),
		"log_rotate_duration": tftypes.NewValue(tftypes.Number, nil),
	}
}

// TestUnit_LogTargetObjectStoreDataSource_Basic verifies data source reads log target object store by name.
func TestUnit_LogTargetObjectStoreDataSource_Basic(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterLogTargetObjectStoreHandlers(ms.Mux)

	store.Seed(&client.LogTargetObjectStore{
		ID:            "ltos-ds-001",
		Name:          "ds-ltos-test",
		Bucket:        client.NamedReference{Name: "ds-audit-bucket"},
		LogNamePrefix: client.AuditLogNamePrefix{Prefix: "audit/"},
		LogRotate:     client.AuditLogRotate{Duration: 3600000},
	})

	d := newTestLogTargetObjectStoreDataSource(t, ms)
	s := logTargetObjectStoreDSSchema(t).Schema

	cfg := nullLogTargetObjectStoreDSConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "ds-ltos-test")

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildLogTargetObjectStoreDSType(), nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(buildLogTargetObjectStoreDSType(), cfg), Schema: s},
	}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("DataSource Read returned error: %s", readResp.Diagnostics)
	}

	var model logTargetObjectStoreDataSourceModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Name.ValueString() != "ds-ltos-test" {
		t.Errorf("expected name=ds-ltos-test, got %s", model.Name.ValueString())
	}
	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected ID to be populated")
	}
	if model.BucketName.ValueString() != "ds-audit-bucket" {
		t.Errorf("expected bucket_name=ds-audit-bucket, got %s", model.BucketName.ValueString())
	}
	if model.LogNamePrefix.ValueString() != "audit/" {
		t.Errorf("expected log_name_prefix=audit/, got %s", model.LogNamePrefix.ValueString())
	}
	if model.LogRotateDuration.ValueInt64() != 3600000 {
		t.Errorf("expected log_rotate_duration=3600000, got %d", model.LogRotateDuration.ValueInt64())
	}
}
