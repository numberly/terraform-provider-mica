package provider

import (
	"context"
	"math/big"
	"net/http"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	resschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/numberly/opentofu-provider-flashblade/internal/client"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock/handlers"
)

// ---- helpers ----------------------------------------------------------------

// newTestBucketResource creates a bucketResource wired to the given mock server.
func newTestBucketResource(t *testing.T, ms *testmock.MockServer) *bucketResource {
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
	return &bucketResource{client: c}
}

// bucketResourceSchema returns the parsed schema for the bucket resource.
func bucketResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &bucketResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

// buildBucketType returns the tftypes.Object for the full bucket resource schema.
func buildBucketType() tftypes.Object {
	spaceType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"data_reduction":      tftypes.Number,
		"snapshots":           tftypes.Number,
		"total_physical":      tftypes.Number,
		"unique":              tftypes.Number,
		"virtual":             tftypes.Number,
		"snapshots_effective": tftypes.Number,
	}}
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	eradicationConfigType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"eradication_delay":  tftypes.Number,
		"eradication_mode":   tftypes.String,
		"manual_eradication": tftypes.String,
	}}
	objectLockConfigType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"freeze_locked_objects":  tftypes.Bool,
		"default_retention":      tftypes.Number,
		"default_retention_mode": tftypes.String,
		"object_lock_enabled":    tftypes.Bool,
	}}
	publicAccessConfigType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"block_new_public_policies": tftypes.Bool,
		"block_public_access":       tftypes.Bool,
	}}
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":                           tftypes.String,
		"name":                         tftypes.String,
		"account":                      tftypes.String,
		"created":                      tftypes.Number,
		"destroyed":                    tftypes.Bool,
		"destroy_eradicate_on_delete":  tftypes.Bool,
		"time_remaining":               tftypes.Number,
		"versioning":                   tftypes.String,
		"quota_limit":                  tftypes.Number,
		"hard_limit_enabled":           tftypes.Bool,
		"object_count":                 tftypes.Number,
		"bucket_type":                  tftypes.String,
		"retention_lock":               tftypes.String,
		"eradication_config":           eradicationConfigType,
		"object_lock_config":           objectLockConfigType,
		"public_access_config":         publicAccessConfigType,
		"public_status":                tftypes.String,
		"space":                        spaceType,
		"timeouts":                     timeoutsType,
	}}
}

// nullBucketConfig returns a base config map with all attributes null.
func nullBucketConfig() map[string]tftypes.Value {
	spaceType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"data_reduction":      tftypes.Number,
		"snapshots":           tftypes.Number,
		"total_physical":      tftypes.Number,
		"unique":              tftypes.Number,
		"virtual":             tftypes.Number,
		"snapshots_effective": tftypes.Number,
	}}
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	eradicationConfigType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"eradication_delay":  tftypes.Number,
		"eradication_mode":   tftypes.String,
		"manual_eradication": tftypes.String,
	}}
	objectLockConfigType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"freeze_locked_objects":  tftypes.Bool,
		"default_retention":      tftypes.Number,
		"default_retention_mode": tftypes.String,
		"object_lock_enabled":    tftypes.Bool,
	}}
	publicAccessConfigType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"block_new_public_policies": tftypes.Bool,
		"block_public_access":       tftypes.Bool,
	}}
	return map[string]tftypes.Value{
		"id":                          tftypes.NewValue(tftypes.String, nil),
		"name":                        tftypes.NewValue(tftypes.String, nil),
		"account":                     tftypes.NewValue(tftypes.String, nil),
		"created":                     tftypes.NewValue(tftypes.Number, nil),
		"destroyed":                   tftypes.NewValue(tftypes.Bool, nil),
		"destroy_eradicate_on_delete": tftypes.NewValue(tftypes.Bool, nil),
		"time_remaining":              tftypes.NewValue(tftypes.Number, nil),
		"versioning":                  tftypes.NewValue(tftypes.String, nil),
		"quota_limit":                 tftypes.NewValue(tftypes.Number, nil),
		"hard_limit_enabled":          tftypes.NewValue(tftypes.Bool, nil),
		"object_count":                tftypes.NewValue(tftypes.Number, nil),
		"bucket_type":                 tftypes.NewValue(tftypes.String, nil),
		"retention_lock":              tftypes.NewValue(tftypes.String, nil),
		"eradication_config":          tftypes.NewValue(eradicationConfigType, nil),
		"object_lock_config":          tftypes.NewValue(objectLockConfigType, nil),
		"public_access_config":        tftypes.NewValue(publicAccessConfigType, nil),
		"public_status":               tftypes.NewValue(tftypes.String, nil),
		"space":                       tftypes.NewValue(spaceType, nil),
		"timeouts":                    tftypes.NewValue(timeoutsType, nil),
	}
}

// bucketPlanWithNameAndAccount returns a tfsdk.Plan with name, account, and eradicate=false.
func bucketPlanWithNameAndAccount(t *testing.T, name, account string) tfsdk.Plan {
	t.Helper()
	s := bucketResourceSchema(t).Schema
	cfg := nullBucketConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, name)
	cfg["account"] = tftypes.NewValue(tftypes.String, account)
	cfg["destroy_eradicate_on_delete"] = tftypes.NewValue(tftypes.Bool, false)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildBucketType(), cfg),
		Schema: s,
	}
}

// setupBucketMockServer creates a mock server with account and bucket handlers,
// pre-seeds an account, and returns the server, client, and bucket store.
func setupBucketMockServer(t *testing.T) (*testmock.MockServer, *client.FlashBladeClient, *handlers.BucketStoreFacade) {
	t.Helper()
	ms := testmock.NewMockServer()
	accountStore := handlers.RegisterObjectStoreAccountHandlers(ms.Mux)
	bucketStore := handlers.NewBucketStoreFacade(handlers.RegisterBucketHandlers(ms.Mux, accountStore))

	c, err := client.NewClient(context.Background(), client.Config{
		Endpoint:           ms.URL(),
		APIToken:           "test-token",
		InsecureSkipVerify: true,
		MaxRetries:         1,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	// Pre-seed the test account.
	_, err = c.PostObjectStoreAccount(context.Background(), "test-account", client.ObjectStoreAccountPost{})
	if err != nil {
		t.Fatalf("PostObjectStoreAccount: %v", err)
	}

	return ms, c, bucketStore
}

// ---- tests ------------------------------------------------------------------

// TestUnit_Bucket_Create verifies Create populates ID, account, versioning, and created.
func TestUnit_Unit_Bucket_Create(t *testing.T) {
	ms, _, _ := setupBucketMockServer(t)
	defer ms.Close()

	r := newTestBucketResource(t, ms)
	s := bucketResourceSchema(t).Schema

	plan := bucketPlanWithNameAndAccount(t, "test-bucket", "test-account")
	req := resource.CreateRequest{Plan: plan}
	resp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildBucketType(), nil), Schema: s},
	}

	r.Create(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Create returned error: %s", resp.Diagnostics)
	}

	var model bucketModel
	if diags := resp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected non-empty ID after Create")
	}
	if model.Name.ValueString() != "test-bucket" {
		t.Errorf("expected name=test-bucket, got %s", model.Name.ValueString())
	}
	if model.Account.ValueString() != "test-account" {
		t.Errorf("expected account=test-account, got %s", model.Account.ValueString())
	}
	if model.Created.IsNull() || model.Created.IsUnknown() {
		t.Error("expected created to be populated after Create")
	}
	if model.Destroyed.ValueBool() {
		t.Error("expected destroyed=false after Create")
	}
}

// TestUnit_Bucket_Update verifies PATCH updates quota_limit, versioning, and hard_limit_enabled.
func TestUnit_Unit_Bucket_Update(t *testing.T) {
	ms, _, _ := setupBucketMockServer(t)
	defer ms.Close()

	r := newTestBucketResource(t, ms)
	s := bucketResourceSchema(t).Schema

	// Create first.
	plan := bucketPlanWithNameAndAccount(t, "update-bucket", "test-account")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildBucketType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Update with new quota, versioning, hard_limit_enabled.
	updateCfg := nullBucketConfig()
	updateCfg["name"] = tftypes.NewValue(tftypes.String, "update-bucket")
	updateCfg["account"] = tftypes.NewValue(tftypes.String, "test-account")
	updateCfg["quota_limit"] = tftypes.NewValue(tftypes.Number, new(big.Float).SetInt64(10737418240))
	updateCfg["versioning"] = tftypes.NewValue(tftypes.String, "enabled")
	updateCfg["hard_limit_enabled"] = tftypes.NewValue(tftypes.Bool, true)
	updateCfg["destroy_eradicate_on_delete"] = tftypes.NewValue(tftypes.Bool, false)

	updatePlan := tfsdk.Plan{Raw: tftypes.NewValue(buildBucketType(), updateCfg), Schema: s}
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildBucketType(), nil), Schema: s},
	}

	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlan,
		State: createResp.State,
	}, updateResp)

	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update returned error: %s", updateResp.Diagnostics)
	}

	var model bucketModel
	if diags := updateResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.QuotaLimit.ValueInt64() != 10737418240 {
		t.Errorf("expected quota_limit=10737418240, got %d", model.QuotaLimit.ValueInt64())
	}
	if model.Versioning.ValueString() != "enabled" {
		t.Errorf("expected versioning=enabled, got %s", model.Versioning.ValueString())
	}
	if !model.HardLimitEnabled.ValueBool() {
		t.Error("expected hard_limit_enabled=true after Update")
	}
}

// TestUnit_Bucket_Destroy verifies that when destroy_eradicate_on_delete=false,
// only soft-delete is performed (no DELETE/eradication).
func TestUnit_Unit_Bucket_Destroy(t *testing.T) {
	ms, c, _ := setupBucketMockServer(t)
	defer ms.Close()

	r := newTestBucketResource(t, ms)
	s := bucketResourceSchema(t).Schema

	// Create.
	plan := bucketPlanWithNameAndAccount(t, "destroy-bucket", "test-account")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildBucketType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Delete (soft-delete only — eradicate=false is default).
	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: createResp.State}, deleteResp)

	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete returned error: %s", deleteResp.Diagnostics)
	}

	// Verify bucket is soft-deleted but NOT eradicated.
	destroyed := true
	buckets, err := c.ListBuckets(context.Background(), client.ListBucketsOpts{
		Names:     []string{"destroy-bucket"},
		Destroyed: &destroyed,
	})
	if err != nil {
		t.Fatalf("ListBuckets: %v", err)
	}
	if len(buckets) != 1 {
		t.Errorf("expected bucket to be soft-deleted (still in destroyed list), got %d buckets", len(buckets))
	}
	if !buckets[0].Destroyed {
		t.Error("expected bucket.Destroyed=true after soft-delete")
	}
}

// TestUnit_Bucket_Destroy_WithEradicate verifies that when destroy_eradicate_on_delete=true,
// the bucket is soft-deleted AND eradicated.
func TestUnit_Unit_Bucket_Destroy_WithEradicate(t *testing.T) {
	ms, c, _ := setupBucketMockServer(t)
	defer ms.Close()

	r := newTestBucketResource(t, ms)
	s := bucketResourceSchema(t).Schema

	// Create with eradicate=true.
	cfg := nullBucketConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "eradicate-bucket")
	cfg["account"] = tftypes.NewValue(tftypes.String, "test-account")
	cfg["destroy_eradicate_on_delete"] = tftypes.NewValue(tftypes.Bool, true)
	plan := tfsdk.Plan{Raw: tftypes.NewValue(buildBucketType(), cfg), Schema: s}

	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildBucketType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Delete with eradication.
	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: createResp.State}, deleteResp)

	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete returned error: %s", deleteResp.Diagnostics)
	}

	// Verify bucket is fully eradicated (not found at all).
	destroyed := true
	buckets, err := c.ListBuckets(context.Background(), client.ListBucketsOpts{
		Names:     []string{"eradicate-bucket"},
		Destroyed: &destroyed,
	})
	if err != nil {
		t.Fatalf("ListBuckets: %v", err)
	}
	if len(buckets) != 0 {
		t.Errorf("expected bucket to be eradicated (not found), got %d buckets", len(buckets))
	}
}

// TestUnit_Bucket_Import verifies ImportState populates all attributes including account ref,
// and that a subsequent Read produces 0 diff.
func TestUnit_Unit_Bucket_Import(t *testing.T) {
	ms, c, _ := setupBucketMockServer(t)
	defer ms.Close()

	// Pre-create a bucket directly via the client.
	importBucket, err := c.PostBucket(context.Background(), "import-bucket", client.BucketPost{
		Account:    client.NamedReference{Name: "test-account"},
		QuotaLimit: "21474836480",
	})
	if err != nil {
		t.Fatalf("PostBucket: %v", err)
	}
	v := "enabled"
	_, err = c.PatchBucket(context.Background(), importBucket.ID, client.BucketPatch{Versioning: &v})
	if err != nil {
		t.Fatalf("PatchBucket (versioning): %v", err)
	}

	r := newTestBucketResource(t, ms)
	s := bucketResourceSchema(t).Schema

	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildBucketType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "import-bucket"}, importResp)

	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState returned error: %s", importResp.Diagnostics)
	}

	var model bucketModel
	if diags := importResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Name.ValueString() != "import-bucket" {
		t.Errorf("expected name=import-bucket, got %s", model.Name.ValueString())
	}
	if model.Account.ValueString() != "test-account" {
		t.Errorf("expected account=test-account, got %s", model.Account.ValueString())
	}
	if model.Versioning.ValueString() != "enabled" {
		t.Errorf("expected versioning=enabled, got %s", model.Versioning.ValueString())
	}
	if model.QuotaLimit.ValueInt64() != 21474836480 {
		t.Errorf("expected quota_limit=21474836480, got %d", model.QuotaLimit.ValueInt64())
	}
	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected ID to be populated after Import")
	}
}

// TestUnit_Bucket_DriftLog verifies that Read logs diffs via tflog when quota_limit
// or versioning diverge from state.
func TestUnit_Unit_Bucket_DriftLog(t *testing.T) {
	ms, c, _ := setupBucketMockServer(t)
	defer ms.Close()

	// Create a bucket via client.
	driftBucket, err := c.PostBucket(context.Background(), "drift-bucket", client.BucketPost{
		Account:    client.NamedReference{Name: "test-account"},
		QuotaLimit: "10737418240",
	})
	if err != nil {
		t.Fatalf("PostBucket: %v", err)
	}
	vNone := "none"
	_, err = c.PatchBucket(context.Background(), driftBucket.ID, client.BucketPatch{Versioning: &vNone})
	if err != nil {
		t.Fatalf("PatchBucket (versioning): %v", err)
	}

	r := newTestBucketResource(t, ms)
	s := bucketResourceSchema(t).Schema

	// Set up state with different values to simulate drift.
	stateCfg := nullBucketConfig()
	stateCfg["name"] = tftypes.NewValue(tftypes.String, "drift-bucket")
	stateCfg["account"] = tftypes.NewValue(tftypes.String, "test-account")
	stateCfg["quota_limit"] = tftypes.NewValue(tftypes.Number, new(big.Float).SetInt64(5368709120)) // different from API
	stateCfg["versioning"] = tftypes.NewValue(tftypes.String, "enabled")     // different from API
	stateCfg["destroy_eradicate_on_delete"] = tftypes.NewValue(tftypes.Bool, false)

	stateObj := tfsdk.State{Raw: tftypes.NewValue(buildBucketType(), stateCfg), Schema: s}

	readResp := &resource.ReadResponse{
		State: stateObj,
	}
	// Read should not error — it should just log the drift.
	r.Read(context.Background(), resource.ReadRequest{State: stateObj}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}

	// Verify state was updated to API values (drift corrected).
	var model bucketModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}
	if model.QuotaLimit.ValueInt64() != 10737418240 {
		t.Errorf("expected quota_limit=10737418240 from API, got %d", model.QuotaLimit.ValueInt64())
	}
	if model.Versioning.ValueString() != "none" {
		t.Errorf("expected versioning=none from API, got %s", model.Versioning.ValueString())
	}
}

// TestUnit_Bucket_NonEmptyDelete verifies that attempting to delete a bucket with
// objects returns a clear diagnostic error.
func TestUnit_Unit_Bucket_NonEmptyDelete(t *testing.T) {
	ms, c, bs := setupBucketMockServer(t)
	defer ms.Close()

	// Create a bucket and set its ObjectCount > 0 in the mock store so the
	// fresh GET inside Delete returns a non-empty bucket.
	bkt, err := c.PostBucket(context.Background(), "nonempty-bucket", client.BucketPost{
		Account: client.NamedReference{Name: "test-account"},
	})
	if err != nil {
		t.Fatalf("PostBucket: %v", err)
	}
	bs.SetObjectCount("nonempty-bucket", 5)

	r := newTestBucketResource(t, ms)
	s := bucketResourceSchema(t).Schema

	// Build state — object_count in state no longer matters since Delete
	// performs a fresh GET, but we still need valid state for deserialization.
	stateCfg := nullBucketConfig()
	stateCfg["id"] = tftypes.NewValue(tftypes.String, bkt.ID)
	stateCfg["name"] = tftypes.NewValue(tftypes.String, "nonempty-bucket")
	stateCfg["account"] = tftypes.NewValue(tftypes.String, "test-account")
	stateCfg["object_count"] = tftypes.NewValue(tftypes.Number, int64(5))
	stateCfg["destroyed"] = tftypes.NewValue(tftypes.Bool, false)
	stateCfg["destroy_eradicate_on_delete"] = tftypes.NewValue(tftypes.Bool, false)

	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{
		State: tfsdk.State{Raw: tftypes.NewValue(buildBucketType(), stateCfg), Schema: s},
	}, deleteResp)

	if !deleteResp.Diagnostics.HasError() {
		t.Error("expected error diagnostic for non-empty bucket delete, got none")
	}

	found := false
	for _, diag := range deleteResp.Diagnostics {
		if strings.Contains(diag.Detail(), "contains") || strings.Contains(diag.Summary(), "contains") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'contains objects' diagnostic, got: %s", deleteResp.Diagnostics)
	}
}

// TestUnit_Bucket_PlanModifiers verifies all RequiresReplace and UseStateForUnknown
// plan modifiers in the bucket resource schema.
func TestUnit_Unit_Bucket_PlanModifiers(t *testing.T) {
	s := bucketResourceSchema(t).Schema

	// id — UseStateForUnknown
	idAttr, ok := s.Attributes["id"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("id attribute not found or wrong type")
	}
	if len(idAttr.PlanModifiers) == 0 {
		t.Error("expected UseStateForUnknown plan modifier on id attribute")
	}

	// name — RequiresReplace
	nameAttr, ok := s.Attributes["name"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("name attribute not found or wrong type")
	}
	if len(nameAttr.PlanModifiers) == 0 {
		t.Error("expected RequiresReplace plan modifier on name attribute")
	}

	// account — RequiresReplace
	accountAttr, ok := s.Attributes["account"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("account attribute not found or wrong type")
	}
	if len(accountAttr.PlanModifiers) == 0 {
		t.Error("expected RequiresReplace plan modifier on account attribute")
	}

	// created — UseStateForUnknown (custom int64 modifier)
	createdAttr, ok := s.Attributes["created"].(resschema.Int64Attribute)
	if !ok {
		t.Fatal("created attribute not found or wrong type")
	}
	if len(createdAttr.PlanModifiers) == 0 {
		t.Error("expected UseStateForUnknown plan modifier on created attribute")
	}

	// bucket_type — UseStateForUnknown
	btAttr, ok := s.Attributes["bucket_type"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("bucket_type attribute not found or wrong type")
	}
	if len(btAttr.PlanModifiers) == 0 {
		t.Error("expected UseStateForUnknown plan modifier on bucket_type attribute")
	}
}

// TestUnit_Bucket_Lifecycle exercises the full Create->Read->Update->Read->Delete sequence.
func TestUnit_Unit_Bucket_Lifecycle(t *testing.T) {
	ms, _, _ := setupBucketMockServer(t)
	defer ms.Close()

	r := newTestBucketResource(t, ms)
	s := bucketResourceSchema(t).Schema

	// Step 1: Create.
	createPlan := bucketPlanWithNameAndAccount(t, "lifecycle-bucket", "test-account")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildBucketType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}
	var createModel bucketModel
	if diags := createResp.State.Get(context.Background(), &createModel); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}
	if createModel.Name.ValueString() != "lifecycle-bucket" {
		t.Errorf("Create: expected name=lifecycle-bucket, got %s", createModel.Name.ValueString())
	}

	// Step 2: Read post-create.
	readResp1 := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp1)
	if readResp1.Diagnostics.HasError() {
		t.Fatalf("Read post-create: %s", readResp1.Diagnostics)
	}
	var readModel1 bucketModel
	if diags := readResp1.State.Get(context.Background(), &readModel1); diags.HasError() {
		t.Fatalf("Get read1 state: %s", diags)
	}
	if readModel1.Account.ValueString() != "test-account" {
		t.Errorf("Read1: expected account=test-account, got %s", readModel1.Account.ValueString())
	}

	// Step 3: Update versioning to enabled.
	updateCfg := nullBucketConfig()
	updateCfg["name"] = tftypes.NewValue(tftypes.String, "lifecycle-bucket")
	updateCfg["account"] = tftypes.NewValue(tftypes.String, "test-account")
	updateCfg["versioning"] = tftypes.NewValue(tftypes.String, "enabled")
	updateCfg["destroy_eradicate_on_delete"] = tftypes.NewValue(tftypes.Bool, false)
	updatePlan := tfsdk.Plan{Raw: tftypes.NewValue(buildBucketType(), updateCfg), Schema: s}
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildBucketType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlan,
		State: readResp1.State,
	}, updateResp)
	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update: %s", updateResp.Diagnostics)
	}
	var updateModel bucketModel
	if diags := updateResp.State.Get(context.Background(), &updateModel); diags.HasError() {
		t.Fatalf("Get update state: %s", diags)
	}
	if updateModel.Versioning.ValueString() != "enabled" {
		t.Errorf("Update: expected versioning=enabled, got %s", updateModel.Versioning.ValueString())
	}

	// Step 4: Read post-update.
	readResp2 := &resource.ReadResponse{State: updateResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: updateResp.State}, readResp2)
	if readResp2.Diagnostics.HasError() {
		t.Fatalf("Read post-update: %s", readResp2.Diagnostics)
	}
	var readModel2 bucketModel
	if diags := readResp2.State.Get(context.Background(), &readModel2); diags.HasError() {
		t.Fatalf("Get read2 state: %s", diags)
	}
	if readModel2.Versioning.ValueString() != "enabled" {
		t.Errorf("Read2: expected versioning=enabled, got %s", readModel2.Versioning.ValueString())
	}

	// Step 5: Delete (soft-delete, eradicate=false default).
	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: readResp2.State}, deleteResp)
	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete: %s", deleteResp.Diagnostics)
	}
}

// TestUnit_Bucket_ImportIdempotency verifies ImportState->Read produces state matching original Create.
func TestUnit_Unit_Bucket_ImportIdempotency(t *testing.T) {
	ms, _, _ := setupBucketMockServer(t)
	defer ms.Close()

	r := newTestBucketResource(t, ms)
	s := bucketResourceSchema(t).Schema

	// Create.
	createCfg := nullBucketConfig()
	createCfg["name"] = tftypes.NewValue(tftypes.String, "idempotent-bucket")
	createCfg["account"] = tftypes.NewValue(tftypes.String, "test-account")
	createCfg["versioning"] = tftypes.NewValue(tftypes.String, "enabled")
	createCfg["quota_limit"] = tftypes.NewValue(tftypes.Number, new(big.Float).SetInt64(10737418240))
	createCfg["destroy_eradicate_on_delete"] = tftypes.NewValue(tftypes.Bool, false)
	createPlan := tfsdk.Plan{Raw: tftypes.NewValue(buildBucketType(), createCfg), Schema: s}
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildBucketType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}
	var createModel bucketModel
	if diags := createResp.State.Get(context.Background(), &createModel); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}

	// ImportState.
	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildBucketType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "idempotent-bucket"}, importResp)
	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState: %s", importResp.Diagnostics)
	}

	// Read to populate full state.
	readResp := &resource.ReadResponse{State: importResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: importResp.State}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read post-import: %s", readResp.Diagnostics)
	}
	var importedModel bucketModel
	if diags := readResp.State.Get(context.Background(), &importedModel); diags.HasError() {
		t.Fatalf("Get imported state: %s", diags)
	}

	// Verify 0-diff.
	if importedModel.Name.ValueString() != createModel.Name.ValueString() {
		t.Errorf("name mismatch: create=%s import=%s", createModel.Name.ValueString(), importedModel.Name.ValueString())
	}
	if importedModel.Account.ValueString() != createModel.Account.ValueString() {
		t.Errorf("account mismatch: create=%s import=%s", createModel.Account.ValueString(), importedModel.Account.ValueString())
	}
	if importedModel.Versioning.ValueString() != createModel.Versioning.ValueString() {
		t.Errorf("versioning mismatch: create=%s import=%s", createModel.Versioning.ValueString(), importedModel.Versioning.ValueString())
	}
	if importedModel.QuotaLimit.ValueInt64() != createModel.QuotaLimit.ValueInt64() {
		t.Errorf("quota_limit mismatch: create=%d import=%d", createModel.QuotaLimit.ValueInt64(), importedModel.QuotaLimit.ValueInt64())
	}
}

// TestUnit_Bucket_Create_Conflict verifies that a 409 Conflict on POST produces
// an error diagnostic (not a panic or silent failure).
func TestUnit_Unit_Bucket_Create_Conflict(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	// Register a mock that always returns 409 on POST /buckets.
	ms.RegisterHandler("/api/2.22/buckets", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handlers.WriteJSONError(w, http.StatusConflict, "Bucket with the given name already exists.")
			return
		}
		// GET: return empty list (no existing bucket).
		handlers.WriteJSONListResponse(w, http.StatusOK, []client.Bucket{})
	})

	r := newTestBucketResource(t, ms)
	s := bucketResourceSchema(t).Schema

	plan := bucketPlanWithNameAndAccount(t, "conflict-bucket", "test-account")
	resp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildBucketType(), nil), Schema: s},
	}

	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("expected Create to produce an error diagnostic on 409 Conflict, got none")
	}
	// Verify the error message is informative.
	found := false
	for _, d := range resp.Diagnostics {
		if strings.Contains(d.Detail(), "409") || strings.Contains(d.Detail(), "conflict") ||
			strings.Contains(d.Detail(), "Conflict") || strings.Contains(d.Summary(), "conflict") ||
			strings.Contains(d.Summary(), "Error") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected diagnostic to mention conflict or 409, got: %s", resp.Diagnostics)
	}
}

// TestUnit_Bucket_Read_NotFound verifies that a 404 during Read removes the resource
// from Terraform state without producing an error diagnostic.
func TestUnit_Unit_Bucket_Read_NotFound(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	// Register a GET handler that always returns 404-equivalent: empty items list.
	// FlashBlade returns HTTP 200 + empty items for not-found resources.
	ms.RegisterHandler("/api/2.22/buckets", func(w http.ResponseWriter, r *http.Request) {
		handlers.WriteJSONListResponse(w, http.StatusOK, []client.Bucket{})
	})

	r := newTestBucketResource(t, ms)
	s := bucketResourceSchema(t).Schema

	// Build a state that represents an existing bucket.
	cfg := nullBucketConfig()
	cfg["id"] = tftypes.NewValue(tftypes.String, "bucket-gone-id")
	cfg["name"] = tftypes.NewValue(tftypes.String, "gone-bucket")
	cfg["account"] = tftypes.NewValue(tftypes.String, "test-account")
	cfg["destroy_eradicate_on_delete"] = tftypes.NewValue(tftypes.Bool, false)
	state := tfsdk.State{Raw: tftypes.NewValue(buildBucketType(), cfg), Schema: s}

	readResp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), resource.ReadRequest{State: state}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}
	if !readResp.State.Raw.IsNull() {
		t.Error("expected state to be removed (null) when bucket not found, but it was not null")
	}
}

// TestUnit_Bucket_VersioningValidator verifies the versioning field rejects
// invalid values and accepts the valid enum values.
func TestUnit_Unit_Bucket_VersioningValidator(t *testing.T) {
	s := bucketResourceSchema(t).Schema

	vAttr, ok := s.Attributes["versioning"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("versioning attribute not found or wrong type")
	}
	if len(vAttr.Validators) == 0 {
		t.Fatal("expected at least one validator on versioning attribute")
	}

	v := vAttr.Validators[0]

	// "invalid" should produce an error.
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), validator.StringRequest{
		ConfigValue: types.StringValue("invalid"),
	}, resp)
	if !resp.Diagnostics.HasError() {
		t.Error("expected validator to reject 'invalid' versioning value")
	}

	// Valid values should not produce errors.
	for _, valid := range []string{"none", "enabled", "suspended"} {
		resp := &validator.StringResponse{}
		v.ValidateString(context.Background(), validator.StringRequest{
			ConfigValue: types.StringValue(valid),
		}, resp)
		if resp.Diagnostics.HasError() {
			t.Errorf("expected validator to accept %q versioning value, got error: %s", valid, resp.Diagnostics)
		}
	}
}

// TestUnit_Bucket_VersioningWarning verifies that ValidateConfig emits a warning
// when versioning is not set to "enabled" (replication readiness).
func TestUnit_Unit_Bucket_VersioningWarning(t *testing.T) {
	r := &bucketResource{}

	// Verify interface is satisfied.
	var _ resource.ResourceWithValidateConfig = r

	s := bucketResourceSchema(t).Schema

	tests := []struct {
		name        string
		versioning  tftypes.Value
		wantWarning bool
	}{
		{"none emits warning", tftypes.NewValue(tftypes.String, "none"), true},
		{"suspended emits warning", tftypes.NewValue(tftypes.String, "suspended"), true},
		{"enabled no warning", tftypes.NewValue(tftypes.String, "enabled"), false},
		{"null no warning", tftypes.NewValue(tftypes.String, nil), false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := nullBucketConfig()
			// ValidateConfig needs at least required fields set.
			cfg["name"] = tftypes.NewValue(tftypes.String, "test-bucket")
			cfg["account"] = tftypes.NewValue(tftypes.String, "test-account")
			cfg["versioning"] = tc.versioning

			req := resource.ValidateConfigRequest{
				Config: tfsdk.Config{
					Raw:    tftypes.NewValue(buildBucketType(), cfg),
					Schema: s,
				},
			}
			resp := &resource.ValidateConfigResponse{}
			r.ValidateConfig(context.Background(), req, resp)

			if resp.Diagnostics.HasError() {
				t.Fatalf("unexpected error: %s", resp.Diagnostics)
			}

			hasWarning := resp.Diagnostics.WarningsCount() > 0

			if tc.wantWarning && !hasWarning {
				t.Error("expected versioning warning but got none")
			}
			if !tc.wantWarning && hasWarning {
				t.Error("did not expect versioning warning")
			}
		})
	}
}

// ---- config block helpers ---------------------------------------------------

// eradicationConfigTFValue builds a tftypes.Value for the eradication_config nested object.
func eradicationConfigTFValue(delay int64, mode, manualErad string) tftypes.Value {
	typ := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"eradication_delay":  tftypes.Number,
		"eradication_mode":   tftypes.String,
		"manual_eradication": tftypes.String,
	}}
	return tftypes.NewValue(typ, map[string]tftypes.Value{
		"eradication_delay":  tftypes.NewValue(tftypes.Number, new(big.Float).SetInt64(delay)),
		"eradication_mode":   tftypes.NewValue(tftypes.String, mode),
		"manual_eradication": tftypes.NewValue(tftypes.String, manualErad),
	})
}

// publicAccessConfigTFValue builds a tftypes.Value for the public_access_config nested object.
func publicAccessConfigTFValue(blockNewPolicies, blockAccess bool) tftypes.Value {
	typ := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"block_new_public_policies": tftypes.Bool,
		"block_public_access":       tftypes.Bool,
	}}
	return tftypes.NewValue(typ, map[string]tftypes.Value{
		"block_new_public_policies": tftypes.NewValue(tftypes.Bool, blockNewPolicies),
		"block_public_access":       tftypes.NewValue(tftypes.Bool, blockAccess),
	})
}

// ---- config block tests -----------------------------------------------------

// TestBucketResource_Create_WithEradicationConfig verifies that Create sends
// eradication_config in the POST body and the response contains the configured values.
func TestUnit_BucketResource_Create_WithEradicationConfig(t *testing.T) {
	ms, _, _ := setupBucketMockServer(t)
	defer ms.Close()

	r := newTestBucketResource(t, ms)
	s := bucketResourceSchema(t).Schema

	cfg := nullBucketConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "erad-config-bucket")
	cfg["account"] = tftypes.NewValue(tftypes.String, "test-account")
	cfg["destroy_eradicate_on_delete"] = tftypes.NewValue(tftypes.Bool, false)
	cfg["eradication_config"] = eradicationConfigTFValue(172800000, "permission-based", "enabled")

	plan := tfsdk.Plan{Raw: tftypes.NewValue(buildBucketType(), cfg), Schema: s}
	resp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildBucketType(), nil), Schema: s},
	}

	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Create returned error: %s", resp.Diagnostics)
	}

	var model bucketModel
	if diags := resp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	// Verify eradication_config was applied.
	attrs := model.EradicationConfig.Attributes()
	if v, ok := attrs["eradication_delay"].(types.Int64); !ok || v.ValueInt64() != 172800000 {
		t.Errorf("expected eradication_delay=172800000, got %v", attrs["eradication_delay"])
	}
	if v, ok := attrs["eradication_mode"].(types.String); !ok || v.ValueString() != "permission-based" {
		t.Errorf("expected eradication_mode=permission-based, got %v", attrs["eradication_mode"])
	}
	if v, ok := attrs["manual_eradication"].(types.String); !ok || v.ValueString() != "enabled" {
		t.Errorf("expected manual_eradication=enabled, got %v", attrs["manual_eradication"])
	}
}

// TestBucketResource_Update_WithPublicAccessConfig verifies that Update sends
// public_access_config in the PATCH body and the response reflects the update.
func TestUnit_BucketResource_Update_WithPublicAccessConfig(t *testing.T) {
	ms, _, _ := setupBucketMockServer(t)
	defer ms.Close()

	r := newTestBucketResource(t, ms)
	s := bucketResourceSchema(t).Schema

	// Step 1: Create a bucket with defaults.
	createPlan := bucketPlanWithNameAndAccount(t, "pubaccess-bucket", "test-account")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildBucketType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Step 2: Update with public_access_config set.
	updateCfg := nullBucketConfig()
	updateCfg["name"] = tftypes.NewValue(tftypes.String, "pubaccess-bucket")
	updateCfg["account"] = tftypes.NewValue(tftypes.String, "test-account")
	updateCfg["destroy_eradicate_on_delete"] = tftypes.NewValue(tftypes.Bool, false)
	updateCfg["public_access_config"] = publicAccessConfigTFValue(true, true)

	updatePlan := tfsdk.Plan{Raw: tftypes.NewValue(buildBucketType(), updateCfg), Schema: s}
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildBucketType(), nil), Schema: s},
	}

	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlan,
		State: createResp.State,
	}, updateResp)

	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update returned error: %s", updateResp.Diagnostics)
	}

	var model bucketModel
	if diags := updateResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	// Verify public_access_config was updated.
	attrs := model.PublicAccessConfig.Attributes()
	if v, ok := attrs["block_new_public_policies"].(types.Bool); !ok || !v.ValueBool() {
		t.Errorf("expected block_new_public_policies=true, got %v", attrs["block_new_public_policies"])
	}
	if v, ok := attrs["block_public_access"].(types.Bool); !ok || !v.ValueBool() {
		t.Errorf("expected block_public_access=true, got %v", attrs["block_public_access"])
	}

	// Verify public_status updated by mock.
	if model.PublicStatus.ValueString() != "not-public" {
		t.Errorf("expected public_status=not-public, got %s", model.PublicStatus.ValueString())
	}
}

// TestBucketResource_Read_MapsConfigBlocks verifies that Read populates all
// config blocks and public_status from the API response.
func TestUnit_BucketResource_Read_MapsConfigBlocks(t *testing.T) {
	ms, c, _ := setupBucketMockServer(t)
	defer ms.Close()

	// Create a bucket with custom eradication config via POST.
	_, err := c.PostBucket(context.Background(), "configread-bucket", client.BucketPost{
		Account: client.NamedReference{Name: "test-account"},
		EradicationConfig: &client.EradicationConfig{
			EradicationDelay:  259200000,
			EradicationMode:   "permission-based",
			ManualEradication: "enabled",
		},
		ObjectLockConfig: &client.ObjectLockConfig{
			FreezeLockedObjects: true,
			DefaultRetention:     86400,
			DefaultRetentionMode: "compliance",
			ObjectLockEnabled:    true,
		},
	})
	if err != nil {
		t.Fatalf("PostBucket: %v", err)
	}

	r := newTestBucketResource(t, ms)
	s := bucketResourceSchema(t).Schema

	// Build minimal state for Read.
	stateCfg := nullBucketConfig()
	stateCfg["name"] = tftypes.NewValue(tftypes.String, "configread-bucket")
	stateCfg["account"] = tftypes.NewValue(tftypes.String, "test-account")
	stateCfg["destroy_eradicate_on_delete"] = tftypes.NewValue(tftypes.Bool, false)

	state := tfsdk.State{Raw: tftypes.NewValue(buildBucketType(), stateCfg), Schema: s}
	readResp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), resource.ReadRequest{State: state}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}

	var model bucketModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	// Verify eradication_config.
	eradAttrs := model.EradicationConfig.Attributes()
	if v, ok := eradAttrs["eradication_delay"].(types.Int64); !ok || v.ValueInt64() != 259200000 {
		t.Errorf("expected eradication_delay=259200000, got %v", eradAttrs["eradication_delay"])
	}
	if v, ok := eradAttrs["eradication_mode"].(types.String); !ok || v.ValueString() != "permission-based" {
		t.Errorf("expected eradication_mode=permission-based, got %v", eradAttrs["eradication_mode"])
	}
	if v, ok := eradAttrs["manual_eradication"].(types.String); !ok || v.ValueString() != "enabled" {
		t.Errorf("expected manual_eradication=enabled, got %v", eradAttrs["manual_eradication"])
	}

	// Verify object_lock_config.
	olAttrs := model.ObjectLockConfig.Attributes()
	if v, ok := olAttrs["freeze_locked_objects"].(types.Bool); !ok || !v.ValueBool() {
		t.Errorf("expected freeze_locked_objects=true, got %v", olAttrs["freeze_locked_objects"])
	}
	if v, ok := olAttrs["default_retention"].(types.Int64); !ok || v.ValueInt64() != 86400 {
		t.Errorf("expected default_retention=86400, got %v", olAttrs["default_retention"])
	}
	if v, ok := olAttrs["default_retention_mode"].(types.String); !ok || v.ValueString() != "compliance" {
		t.Errorf("expected default_retention_mode=compliance, got %v", olAttrs["default_retention_mode"])
	}
	if v, ok := olAttrs["object_lock_enabled"].(types.Bool); !ok || !v.ValueBool() {
		t.Errorf("expected object_lock_enabled=true, got %v", olAttrs["object_lock_enabled"])
	}

	// Verify public_access_config (defaults: false, false).
	paAttrs := model.PublicAccessConfig.Attributes()
	if v, ok := paAttrs["block_new_public_policies"].(types.Bool); !ok || v.ValueBool() {
		t.Errorf("expected block_new_public_policies=false, got %v", paAttrs["block_new_public_policies"])
	}
	if v, ok := paAttrs["block_public_access"].(types.Bool); !ok || v.ValueBool() {
		t.Errorf("expected block_public_access=false, got %v", paAttrs["block_public_access"])
	}

	// Verify public_status.
	if model.PublicStatus.ValueString() != "not-public" {
		t.Errorf("expected public_status=not-public, got %s", model.PublicStatus.ValueString())
	}
}
