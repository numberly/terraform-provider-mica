package provider

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	resschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/numberly/opentofu-provider-flashblade/internal/client"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock/handlers"
)

// ---- helpers ----------------------------------------------------------------

// newTestBucketReplicaLinkResource creates a bucketReplicaLinkResource wired to the given mock server.
func newTestBucketReplicaLinkResource(t *testing.T, ms *testmock.MockServer) *bucketReplicaLinkResource {
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
	return &bucketReplicaLinkResource{client: c}
}

// bucketReplicaLinkResourceSchema returns the parsed schema for the bucket replica link resource.
func bucketReplicaLinkResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &bucketReplicaLinkResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

// buildBucketReplicaLinkType returns the tftypes.Object for the bucket replica link resource schema.
func buildBucketReplicaLinkType() tftypes.Object {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":                        tftypes.String,
		"local_bucket_name":         tftypes.String,
		"remote_bucket_name":        tftypes.String,
		"remote_credentials_name":   tftypes.String,
		"remote_name":               tftypes.String,
		"paused":                    tftypes.Bool,
		"cascading_enabled":         tftypes.Bool,
		"direction":                 tftypes.String,
		"status":                    tftypes.String,
		"status_details":            tftypes.String,
		"lag":                       tftypes.Number,
		"recovery_point":            tftypes.Number,
		"object_backlog_count":      tftypes.Number,
		"object_backlog_total_size": tftypes.Number,
		"timeouts":                  timeoutsType,
	}}
}

// nullBucketReplicaLinkConfig returns a base config map with all attributes null.
func nullBucketReplicaLinkConfig() map[string]tftypes.Value {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return map[string]tftypes.Value{
		"id":                        tftypes.NewValue(tftypes.String, nil),
		"local_bucket_name":         tftypes.NewValue(tftypes.String, nil),
		"remote_bucket_name":        tftypes.NewValue(tftypes.String, nil),
		"remote_credentials_name":   tftypes.NewValue(tftypes.String, nil),
		"remote_name":               tftypes.NewValue(tftypes.String, nil),
		"paused":                    tftypes.NewValue(tftypes.Bool, nil),
		"cascading_enabled":         tftypes.NewValue(tftypes.Bool, nil),
		"direction":                 tftypes.NewValue(tftypes.String, nil),
		"status":                    tftypes.NewValue(tftypes.String, nil),
		"status_details":            tftypes.NewValue(tftypes.String, nil),
		"lag":                       tftypes.NewValue(tftypes.Number, nil),
		"recovery_point":            tftypes.NewValue(tftypes.Number, nil),
		"object_backlog_count":      tftypes.NewValue(tftypes.Number, nil),
		"object_backlog_total_size": tftypes.NewValue(tftypes.Number, nil),
		"timeouts":                  tftypes.NewValue(timeoutsType, nil),
	}
}

// bucketReplicaLinkPlanWith returns a tfsdk.Plan with the given field values.
func bucketReplicaLinkPlanWith(t *testing.T, localBucket, remoteBucket, remoteCredentials string, paused bool) tfsdk.Plan {
	t.Helper()
	s := bucketReplicaLinkResourceSchema(t).Schema
	cfg := nullBucketReplicaLinkConfig()
	cfg["local_bucket_name"] = tftypes.NewValue(tftypes.String, localBucket)
	cfg["remote_bucket_name"] = tftypes.NewValue(tftypes.String, remoteBucket)
	if remoteCredentials != "" {
		cfg["remote_credentials_name"] = tftypes.NewValue(tftypes.String, remoteCredentials)
	}
	cfg["paused"] = tftypes.NewValue(tftypes.Bool, paused)
	cfg["cascading_enabled"] = tftypes.NewValue(tftypes.Bool, false)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildBucketReplicaLinkType(), cfg),
		Schema: s,
	}
}

// ---- tests ------------------------------------------------------------------

// TestUnit_BucketReplicaLink_Create verifies POST creates a bucket replica link;
// state has id, direction=outbound, status=replicating, paused=false.
func TestUnit_BucketReplicaLink_Create(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterBucketReplicaLinkHandlers(ms.Mux)

	r := newTestBucketReplicaLinkResource(t, ms)
	s := bucketReplicaLinkResourceSchema(t).Schema

	plan := bucketReplicaLinkPlanWith(t, "local-bucket", "remote-bucket", "test-creds", false)
	req := resource.CreateRequest{Plan: plan}
	resp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildBucketReplicaLinkType(), nil), Schema: s},
	}

	r.Create(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Create returned error: %s", resp.Diagnostics)
	}

	var model bucketReplicaLinkModel
	if diags := resp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected non-empty id after Create")
	}
	if model.LocalBucketName.ValueString() != "local-bucket" {
		t.Errorf("expected local_bucket_name=local-bucket, got %s", model.LocalBucketName.ValueString())
	}
	if model.RemoteBucketName.ValueString() != "remote-bucket" {
		t.Errorf("expected remote_bucket_name=remote-bucket, got %s", model.RemoteBucketName.ValueString())
	}
	if model.Direction.ValueString() != "outbound" {
		t.Errorf("expected direction=outbound, got %s", model.Direction.ValueString())
	}
	if model.Status.ValueString() != "replicating" {
		t.Errorf("expected status=replicating, got %s", model.Status.ValueString())
	}
	if model.Paused.ValueBool() != false {
		t.Error("expected paused=false after Create")
	}
	if model.RemoteCredentialsName.ValueString() != "test-creds" {
		t.Errorf("expected remote_credentials_name=test-creds, got %s", model.RemoteCredentialsName.ValueString())
	}
}

// TestUnit_BucketReplicaLink_Read verifies GET populates all computed fields
// (direction, status, lag, recovery_point, object_backlog_*).
func TestUnit_BucketReplicaLink_Read(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterBucketReplicaLinkHandlers(ms.Mux)

	r := newTestBucketReplicaLinkResource(t, ms)
	s := bucketReplicaLinkResourceSchema(t).Schema

	// Create first.
	plan := bucketReplicaLinkPlanWith(t, "read-local", "read-remote", "read-creds", false)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildBucketReplicaLinkType(), nil), Schema: s},
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

	var model bucketReplicaLinkModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Direction.ValueString() != "outbound" {
		t.Errorf("expected direction=outbound, got %s", model.Direction.ValueString())
	}
	if model.Status.ValueString() != "replicating" {
		t.Errorf("expected status=replicating, got %s", model.Status.ValueString())
	}
	if model.LocalBucketName.ValueString() != "read-local" {
		t.Errorf("expected local_bucket_name=read-local, got %s", model.LocalBucketName.ValueString())
	}
	if model.RemoteBucketName.ValueString() != "read-remote" {
		t.Errorf("expected remote_bucket_name=read-remote, got %s", model.RemoteBucketName.ValueString())
	}
	// Lag, recovery_point, object_backlog_* should be populated (default 0 from mock).
	if model.Lag.IsNull() || model.Lag.IsUnknown() {
		t.Error("expected lag to be populated after Read")
	}
	if model.RecoveryPoint.IsNull() || model.RecoveryPoint.IsUnknown() {
		t.Error("expected recovery_point to be populated after Read")
	}
}

// TestUnit_BucketReplicaLink_Update_Pause verifies PATCH with paused=true pauses the link.
func TestUnit_BucketReplicaLink_Update_Pause(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterBucketReplicaLinkHandlers(ms.Mux)

	r := newTestBucketReplicaLinkResource(t, ms)
	s := bucketReplicaLinkResourceSchema(t).Schema

	// Create (unpaused).
	createPlan := bucketReplicaLinkPlanWith(t, "pause-local", "pause-remote", "pause-creds", false)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildBucketReplicaLinkType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Update: pause.
	pausePlan := bucketReplicaLinkPlanWith(t, "pause-local", "pause-remote", "pause-creds", true)
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildBucketReplicaLinkType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  pausePlan,
		State: createResp.State,
	}, updateResp)

	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update returned error: %s", updateResp.Diagnostics)
	}

	var model bucketReplicaLinkModel
	if diags := updateResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Paused.ValueBool() != true {
		t.Error("expected paused=true after pause update")
	}
}

// TestUnit_BucketReplicaLink_Update_Resume verifies creating paused, then PATCH with paused=false resumes.
func TestUnit_BucketReplicaLink_Update_Resume(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterBucketReplicaLinkHandlers(ms.Mux)

	r := newTestBucketReplicaLinkResource(t, ms)
	s := bucketReplicaLinkResourceSchema(t).Schema

	// Create (paused).
	createPlan := bucketReplicaLinkPlanWith(t, "resume-local", "resume-remote", "resume-creds", true)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildBucketReplicaLinkType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	var createModel bucketReplicaLinkModel
	if diags := createResp.State.Get(context.Background(), &createModel); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}
	if createModel.Paused.ValueBool() != true {
		t.Error("expected paused=true after Create with paused=true")
	}

	// Update: resume.
	resumePlan := bucketReplicaLinkPlanWith(t, "resume-local", "resume-remote", "resume-creds", false)
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildBucketReplicaLinkType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  resumePlan,
		State: createResp.State,
	}, updateResp)

	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update returned error: %s", updateResp.Diagnostics)
	}

	var model bucketReplicaLinkModel
	if diags := updateResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Paused.ValueBool() != false {
		t.Error("expected paused=false after resume update")
	}
}

// TestUnit_BucketReplicaLink_Delete verifies DELETE removes the link;
// subsequent GET returns not-found.
func TestUnit_BucketReplicaLink_Delete(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterBucketReplicaLinkHandlers(ms.Mux)

	r := newTestBucketReplicaLinkResource(t, ms)
	s := bucketReplicaLinkResourceSchema(t).Schema

	// Create.
	plan := bucketReplicaLinkPlanWith(t, "del-local", "del-remote", "del-creds", false)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildBucketReplicaLinkType(), nil), Schema: s},
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
	_, err := r.client.GetBucketReplicaLink(context.Background(), "del-local", "del-remote")
	if err == nil || !client.IsNotFound(err) {
		t.Errorf("expected bucket replica link to be deleted, got: %v", err)
	}
}

// TestUnit_BucketReplicaLink_Import verifies ImportState by composite ID
// "localBucket/remoteBucket" populates all fields.
func TestUnit_BucketReplicaLink_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterBucketReplicaLinkHandlers(ms.Mux)

	r := newTestBucketReplicaLinkResource(t, ms)
	s := bucketReplicaLinkResourceSchema(t).Schema

	// Create so import can find it.
	plan := bucketReplicaLinkPlanWith(t, "imp-local", "imp-remote", "imp-creds", false)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildBucketReplicaLinkType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Import by composite ID.
	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildBucketReplicaLinkType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "imp-local/imp-remote"}, importResp)

	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState returned error: %s", importResp.Diagnostics)
	}

	var model bucketReplicaLinkModel
	if diags := importResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected non-empty id after import")
	}
	if model.LocalBucketName.ValueString() != "imp-local" {
		t.Errorf("expected local_bucket_name=imp-local, got %s", model.LocalBucketName.ValueString())
	}
	if model.RemoteBucketName.ValueString() != "imp-remote" {
		t.Errorf("expected remote_bucket_name=imp-remote, got %s", model.RemoteBucketName.ValueString())
	}
	if model.Direction.ValueString() != "outbound" {
		t.Errorf("expected direction=outbound, got %s", model.Direction.ValueString())
	}
	if model.Status.ValueString() != "replicating" {
		t.Errorf("expected status=replicating, got %s", model.Status.ValueString())
	}
}

// TestUnit_BucketReplicaLink_Idempotence verifies Create -> Read -> all attributes unchanged (no false drift).
func TestUnit_BucketReplicaLink_Idempotence(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterBucketReplicaLinkHandlers(ms.Mux)

	r := newTestBucketReplicaLinkResource(t, ms)
	s := bucketReplicaLinkResourceSchema(t).Schema

	// Create.
	plan := bucketReplicaLinkPlanWith(t, "idemp-local", "idemp-remote", "idemp-creds", false)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildBucketReplicaLinkType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	var afterCreate bucketReplicaLinkModel
	if diags := createResp.State.Get(context.Background(), &afterCreate); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}

	// Read.
	readResp := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read: %s", readResp.Diagnostics)
	}

	var afterRead bucketReplicaLinkModel
	if diags := readResp.State.Get(context.Background(), &afterRead); diags.HasError() {
		t.Fatalf("Get read state: %s", diags)
	}

	// Compare all fields.
	if afterRead.ID.ValueString() != afterCreate.ID.ValueString() {
		t.Errorf("id drift: create=%s read=%s", afterCreate.ID.ValueString(), afterRead.ID.ValueString())
	}
	if afterRead.LocalBucketName.ValueString() != afterCreate.LocalBucketName.ValueString() {
		t.Errorf("local_bucket_name drift: create=%s read=%s", afterCreate.LocalBucketName.ValueString(), afterRead.LocalBucketName.ValueString())
	}
	if afterRead.RemoteBucketName.ValueString() != afterCreate.RemoteBucketName.ValueString() {
		t.Errorf("remote_bucket_name drift: create=%s read=%s", afterCreate.RemoteBucketName.ValueString(), afterRead.RemoteBucketName.ValueString())
	}
	if afterRead.Paused.ValueBool() != afterCreate.Paused.ValueBool() {
		t.Errorf("paused drift: create=%v read=%v", afterCreate.Paused.ValueBool(), afterRead.Paused.ValueBool())
	}
	if afterRead.Direction.ValueString() != afterCreate.Direction.ValueString() {
		t.Errorf("direction drift: create=%s read=%s", afterCreate.Direction.ValueString(), afterRead.Direction.ValueString())
	}
	if afterRead.Status.ValueString() != afterCreate.Status.ValueString() {
		t.Errorf("status drift: create=%s read=%s", afterCreate.Status.ValueString(), afterRead.Status.ValueString())
	}
}

// TestUnit_BucketReplicaLink_Lifecycle exercises Create -> Read -> Update(pause) ->
// Read -> Update(resume) -> Read -> Delete.
func TestUnit_BucketReplicaLink_Lifecycle(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterBucketReplicaLinkHandlers(ms.Mux)

	r := newTestBucketReplicaLinkResource(t, ms)
	s := bucketReplicaLinkResourceSchema(t).Schema

	// Step 1: Create (unpaused).
	createPlan := bucketReplicaLinkPlanWith(t, "lc-local", "lc-remote", "lc-creds", false)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildBucketReplicaLinkType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	var createModel bucketReplicaLinkModel
	if diags := createResp.State.Get(context.Background(), &createModel); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}
	if createModel.Paused.ValueBool() != false {
		t.Error("Create: expected paused=false")
	}

	// Step 2: Read.
	readResp1 := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp1)
	if readResp1.Diagnostics.HasError() {
		t.Fatalf("Read 1: %s", readResp1.Diagnostics)
	}

	// Step 3: Update (pause).
	pausePlan := bucketReplicaLinkPlanWith(t, "lc-local", "lc-remote", "lc-creds", true)
	updateResp1 := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildBucketReplicaLinkType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  pausePlan,
		State: readResp1.State,
	}, updateResp1)
	if updateResp1.Diagnostics.HasError() {
		t.Fatalf("Update (pause): %s", updateResp1.Diagnostics)
	}

	var pauseModel bucketReplicaLinkModel
	if diags := updateResp1.State.Get(context.Background(), &pauseModel); diags.HasError() {
		t.Fatalf("Get pause state: %s", diags)
	}
	if pauseModel.Paused.ValueBool() != true {
		t.Error("Update: expected paused=true")
	}

	// Step 4: Read.
	readResp2 := &resource.ReadResponse{State: updateResp1.State}
	r.Read(context.Background(), resource.ReadRequest{State: updateResp1.State}, readResp2)
	if readResp2.Diagnostics.HasError() {
		t.Fatalf("Read 2: %s", readResp2.Diagnostics)
	}

	// Step 5: Update (resume).
	resumePlan := bucketReplicaLinkPlanWith(t, "lc-local", "lc-remote", "lc-creds", false)
	updateResp2 := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildBucketReplicaLinkType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  resumePlan,
		State: readResp2.State,
	}, updateResp2)
	if updateResp2.Diagnostics.HasError() {
		t.Fatalf("Update (resume): %s", updateResp2.Diagnostics)
	}

	var resumeModel bucketReplicaLinkModel
	if diags := updateResp2.State.Get(context.Background(), &resumeModel); diags.HasError() {
		t.Fatalf("Get resume state: %s", diags)
	}
	if resumeModel.Paused.ValueBool() != false {
		t.Error("Update: expected paused=false after resume")
	}

	// Step 6: Read.
	readResp3 := &resource.ReadResponse{State: updateResp2.State}
	r.Read(context.Background(), resource.ReadRequest{State: updateResp2.State}, readResp3)
	if readResp3.Diagnostics.HasError() {
		t.Fatalf("Read 3: %s", readResp3.Diagnostics)
	}

	// Step 7: Delete.
	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: readResp3.State}, deleteResp)
	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete: %s", deleteResp.Diagnostics)
	}

	_, err := r.client.GetBucketReplicaLink(context.Background(), "lc-local", "lc-remote")
	if err == nil || !client.IsNotFound(err) {
		t.Errorf("expected bucket replica link to be deleted, got: %v", err)
	}
}

// TestUnit_BucketReplicaLink_Schema verifies schema properties:
// - local_bucket_name and remote_bucket_name have RequiresReplace
// - paused is Optional+Computed
// - direction, status, lag are Computed
func TestUnit_BucketReplicaLink_Schema(t *testing.T) {
	s := bucketReplicaLinkResourceSchema(t).Schema

	// local_bucket_name: Required + RequiresReplace.
	localAttr, ok := s.Attributes["local_bucket_name"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("local_bucket_name attribute not found or wrong type")
	}
	if !localAttr.Required {
		t.Error("local_bucket_name: expected Required=true")
	}
	if len(localAttr.PlanModifiers) == 0 {
		t.Error("local_bucket_name: expected RequiresReplace plan modifier")
	}

	// remote_bucket_name: Required + RequiresReplace.
	remoteAttr, ok := s.Attributes["remote_bucket_name"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("remote_bucket_name attribute not found or wrong type")
	}
	if !remoteAttr.Required {
		t.Error("remote_bucket_name: expected Required=true")
	}
	if len(remoteAttr.PlanModifiers) == 0 {
		t.Error("remote_bucket_name: expected RequiresReplace plan modifier")
	}

	// paused: Optional + Computed.
	pausedAttr, ok := s.Attributes["paused"].(resschema.BoolAttribute)
	if !ok {
		t.Fatal("paused attribute not found or wrong type")
	}
	if !pausedAttr.Optional {
		t.Error("paused: expected Optional=true")
	}
	if !pausedAttr.Computed {
		t.Error("paused: expected Computed=true")
	}

	// direction: Computed.
	directionAttr, ok := s.Attributes["direction"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("direction attribute not found or wrong type")
	}
	if !directionAttr.Computed {
		t.Error("direction: expected Computed=true")
	}

	// status: Computed.
	statusAttr, ok := s.Attributes["status"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("status attribute not found or wrong type")
	}
	if !statusAttr.Computed {
		t.Error("status: expected Computed=true")
	}

	// lag: Computed.
	lagAttr, ok := s.Attributes["lag"].(resschema.Int64Attribute)
	if !ok {
		t.Fatal("lag attribute not found or wrong type")
	}
	if !lagAttr.Computed {
		t.Error("lag: expected Computed=true")
	}
}
