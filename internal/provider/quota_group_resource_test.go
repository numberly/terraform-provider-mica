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

// newTestQuotaGroupResource creates a quotaGroupResource wired to the given mock server.
func newTestQuotaGroupResource(t *testing.T, ms *testmock.MockServer) *quotaGroupResource {
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
	return &quotaGroupResource{client: c}
}

// quotaGroupResourceSchema returns the parsed schema for the quota group resource.
func quotaGroupResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &quotaGroupResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

// buildQuotaGroupType returns the tftypes.Object for the quota group resource.
func buildQuotaGroupType() tftypes.Object {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":               tftypes.String,
		"file_system_name": tftypes.String,
		"gid":              tftypes.String,
		"quota":            tftypes.Number,
		"usage":            tftypes.Number,
		"timeouts":         timeoutsType,
	}}
}

// nullQuotaGroupConfig returns a base config map with all attributes null.
func nullQuotaGroupConfig() map[string]tftypes.Value {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, nil),
		"file_system_name": tftypes.NewValue(tftypes.String, nil),
		"gid":              tftypes.NewValue(tftypes.String, nil),
		"quota":            tftypes.NewValue(tftypes.Number, nil),
		"usage":            tftypes.NewValue(tftypes.Number, nil),
		"timeouts":         tftypes.NewValue(timeoutsType, nil),
	}
}

// quotaGroupPlan returns a tfsdk.Plan with the given fs name, gid, and quota.
func quotaGroupPlan(t *testing.T, fsName, gid string, quota int64) tfsdk.Plan {
	t.Helper()
	s := quotaGroupResourceSchema(t).Schema
	cfg := nullQuotaGroupConfig()
	cfg["file_system_name"] = tftypes.NewValue(tftypes.String, fsName)
	cfg["gid"] = tftypes.NewValue(tftypes.String, gid)
	cfg["quota"] = tftypes.NewValue(tftypes.Number, quota)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildQuotaGroupType(), cfg),
		Schema: s,
	}
}

// ---- tests ------------------------------------------------------------------

// TestQuotaGroupResource_Create verifies Create populates ID and quota.
func TestQuotaGroupResource_Create(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterQuotaHandlers(ms.Mux)

	r := newTestQuotaGroupResource(t, ms)
	s := quotaGroupResourceSchema(t).Schema

	plan := quotaGroupPlan(t, "testfs", "2000", 1073741824)
	resp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildQuotaGroupType(), nil), Schema: s},
	}

	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Create returned error: %s", resp.Diagnostics)
	}

	var model quotaGroupModel
	if diags := resp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.ValueString() != "testfs/2000" {
		t.Errorf("expected ID=testfs/2000, got %s", model.ID.ValueString())
	}
	if model.FileSystemName.ValueString() != "testfs" {
		t.Errorf("expected file_system_name=testfs, got %s", model.FileSystemName.ValueString())
	}
	if model.GID.ValueString() != "2000" {
		t.Errorf("expected gid=2000, got %s", model.GID.ValueString())
	}
	if model.Quota.ValueInt64() != 1073741824 {
		t.Errorf("expected quota=1073741824, got %d", model.Quota.ValueInt64())
	}
}

// TestQuotaGroupResource_Update verifies PATCH updates the quota limit.
func TestQuotaGroupResource_Update(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterQuotaHandlers(ms.Mux)

	r := newTestQuotaGroupResource(t, ms)
	s := quotaGroupResourceSchema(t).Schema

	// Create first.
	createPlan := quotaGroupPlan(t, "testfs", "2000", 1073741824)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildQuotaGroupType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Update quota to 2GB.
	newPlan := quotaGroupPlan(t, "testfs", "2000", 2147483648)
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildQuotaGroupType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  newPlan,
		State: createResp.State,
	}, updateResp)

	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update returned error: %s", updateResp.Diagnostics)
	}

	var model quotaGroupModel
	if diags := updateResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}
	if model.Quota.ValueInt64() != 2147483648 {
		t.Errorf("expected quota=2147483648 after update, got %d", model.Quota.ValueInt64())
	}
}

// TestQuotaGroupResource_Delete verifies Delete removes the quota.
func TestQuotaGroupResource_Delete(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterQuotaHandlers(ms.Mux)

	r := newTestQuotaGroupResource(t, ms)
	s := quotaGroupResourceSchema(t).Schema

	// Create first.
	plan := quotaGroupPlan(t, "testfs", "2000", 1073741824)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildQuotaGroupType(), nil), Schema: s},
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

	// Verify quota is gone.
	_, err := r.client.GetQuotaGroup(context.Background(), "testfs", "2000")
	if err == nil || !client.IsNotFound(err) {
		t.Errorf("expected quota to be deleted, got: %v", err)
	}
}

// TestQuotaGroupResource_Import verifies ImportState populates all attributes from composite ID.
func TestQuotaGroupResource_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterQuotaHandlers(ms.Mux)

	r := newTestQuotaGroupResource(t, ms)
	s := quotaGroupResourceSchema(t).Schema

	// Create first via client directly.
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
	_, err = c.PostQuotaGroup(context.Background(), "testfs", "2000", client.QuotaGroupPost{Quota: 1073741824})
	if err != nil {
		t.Fatalf("PostQuotaGroup: %v", err)
	}

	// Import by composite ID "testfs/2000".
	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildQuotaGroupType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "testfs/2000"}, importResp)

	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState returned error: %s", importResp.Diagnostics)
	}

	var model quotaGroupModel
	if diags := importResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.ValueString() != "testfs/2000" {
		t.Errorf("expected ID=testfs/2000 after import, got %s", model.ID.ValueString())
	}
	if model.FileSystemName.ValueString() != "testfs" {
		t.Errorf("expected file_system_name=testfs after import, got %s", model.FileSystemName.ValueString())
	}
	if model.GID.ValueString() != "2000" {
		t.Errorf("expected gid=2000 after import, got %s", model.GID.ValueString())
	}
	if model.Quota.ValueInt64() != 1073741824 {
		t.Errorf("expected quota=1073741824 after import, got %d", model.Quota.ValueInt64())
	}
}

// ---- data source tests -------------------------------------------------------

// newTestQuotaGroupDataSource creates a quotaGroupDataSource wired to the given mock server.
func newTestQuotaGroupDataSource(t *testing.T, ms *testmock.MockServer) *quotaGroupDataSource {
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
	return &quotaGroupDataSource{client: c}
}

// quotaGroupDataSourceSchema returns the schema for the quota group data source.
func quotaGroupDataSourceSchema(t *testing.T) datasource.SchemaResponse {
	t.Helper()
	d := &quotaGroupDataSource{}
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	return resp
}

// buildQuotaGroupDSType returns the tftypes.Object for the quota group data source.
func buildQuotaGroupDSType() tftypes.Object {
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":               tftypes.String,
		"file_system_name": tftypes.String,
		"gid":              tftypes.String,
		"quota":            tftypes.Number,
		"usage":            tftypes.Number,
	}}
}

// nullQuotaGroupDSConfig returns a base config map with all data source attributes null.
func nullQuotaGroupDSConfig() map[string]tftypes.Value {
	return map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, nil),
		"file_system_name": tftypes.NewValue(tftypes.String, nil),
		"gid":              tftypes.NewValue(tftypes.String, nil),
		"quota":            tftypes.NewValue(tftypes.Number, nil),
		"usage":            tftypes.NewValue(tftypes.Number, nil),
	}
}

// TestQuotaGroupDataSource verifies data source reads quota by file_system_name+gid and returns all attributes.
func TestQuotaGroupDataSource(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterQuotaHandlers(ms.Mux)

	// Create a quota via the client so the data source can find it.
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
	_, err = c.PostQuotaGroup(context.Background(), "testfs", "2000", client.QuotaGroupPost{Quota: 1073741824})
	if err != nil {
		t.Fatalf("PostQuotaGroup: %v", err)
	}

	d := newTestQuotaGroupDataSource(t, ms)
	s := quotaGroupDataSourceSchema(t).Schema

	cfg := nullQuotaGroupDSConfig()
	cfg["file_system_name"] = tftypes.NewValue(tftypes.String, "testfs")
	cfg["gid"] = tftypes.NewValue(tftypes.String, "2000")

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildQuotaGroupDSType(), nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(buildQuotaGroupDSType(), cfg), Schema: s},
	}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("DataSource Read returned error: %s", readResp.Diagnostics)
	}

	var model quotaGroupDataSourceModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.ValueString() != "testfs/2000" {
		t.Errorf("expected ID=testfs/2000, got %s", model.ID.ValueString())
	}
	if model.FileSystemName.ValueString() != "testfs" {
		t.Errorf("expected file_system_name=testfs, got %s", model.FileSystemName.ValueString())
	}
	if model.GID.ValueString() != "2000" {
		t.Errorf("expected gid=2000, got %s", model.GID.ValueString())
	}
	if model.Quota.ValueInt64() != 1073741824 {
		t.Errorf("expected quota=1073741824, got %d", model.Quota.ValueInt64())
	}
}
