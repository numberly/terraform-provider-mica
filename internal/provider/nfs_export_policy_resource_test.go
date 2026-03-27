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

// newTestNFSPolicyResource creates an nfsExportPolicyResource wired to the given mock server.
func newTestNFSPolicyResource(t *testing.T, ms *testmock.MockServer) *nfsExportPolicyResource {
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
	return &nfsExportPolicyResource{client: c}
}

// nfsPolicyResourceSchema returns the parsed schema for the NFS export policy resource.
func nfsPolicyResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &nfsExportPolicyResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

// buildNFSPolicyType returns the tftypes.Object for the NFS export policy resource.
func buildNFSPolicyType() tftypes.Object {
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

// nullNFSPolicyConfig returns a base config map with all attributes null.
func nullNFSPolicyConfig() map[string]tftypes.Value {
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

// nfsPolicyPlanWithName returns a tfsdk.Plan with the given policy name.
func nfsPolicyPlanWithName(t *testing.T, name string) tfsdk.Plan {
	t.Helper()
	s := nfsPolicyResourceSchema(t).Schema
	cfg := nullNFSPolicyConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, name)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildNFSPolicyType(), cfg),
		Schema: s,
	}
}

// nfsPolicyPlanWithNameAndEnabled returns a tfsdk.Plan with name and enabled flag.
func nfsPolicyPlanWithNameAndEnabled(t *testing.T, name string, enabled bool) tfsdk.Plan {
	t.Helper()
	s := nfsPolicyResourceSchema(t).Schema
	cfg := nullNFSPolicyConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, name)
	cfg["enabled"] = tftypes.NewValue(tftypes.Bool, enabled)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildNFSPolicyType(), cfg),
		Schema: s,
	}
}

// ---- tests ------------------------------------------------------------------

// TestNfsExportPolicyResource_Create verifies Create populates ID, name, and enabled.
func TestNfsExportPolicyResource_Create(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterNfsExportPolicyHandlers(ms.Mux)

	r := newTestNFSPolicyResource(t, ms)
	s := nfsPolicyResourceSchema(t).Schema

	plan := nfsPolicyPlanWithNameAndEnabled(t, "test-policy", true)
	resp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNFSPolicyType(), nil), Schema: s},
	}

	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Create returned error: %s", resp.Diagnostics)
	}

	var model nfsExportPolicyModel
	if diags := resp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected non-empty ID after Create")
	}
	if model.Name.ValueString() != "test-policy" {
		t.Errorf("expected name=test-policy, got %s", model.Name.ValueString())
	}
	if !model.Enabled.ValueBool() {
		t.Error("expected enabled=true after Create")
	}
}

// TestNfsExportPolicyResource_Update verifies PATCH updates enabled flag and supports rename.
func TestNfsExportPolicyResource_Update(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterNfsExportPolicyHandlers(ms.Mux)

	r := newTestNFSPolicyResource(t, ms)
	s := nfsPolicyResourceSchema(t).Schema

	// Create first.
	createPlan := nfsPolicyPlanWithNameAndEnabled(t, "update-policy", true)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNFSPolicyType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Update enabled=false.
	newPlan := nfsPolicyPlanWithNameAndEnabled(t, "update-policy", false)
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNFSPolicyType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  newPlan,
		State: createResp.State,
	}, updateResp)

	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update returned error: %s", updateResp.Diagnostics)
	}

	var model nfsExportPolicyModel
	if diags := updateResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}
	if model.Enabled.ValueBool() {
		t.Error("expected enabled=false after update")
	}

	// Now rename the policy in-place.
	renamePlan := nfsPolicyPlanWithNameAndEnabled(t, "update-policy-renamed", false)
	renameResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNFSPolicyType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  renamePlan,
		State: updateResp.State,
	}, renameResp)

	if renameResp.Diagnostics.HasError() {
		t.Fatalf("Rename update returned error: %s", renameResp.Diagnostics)
	}

	var renamedModel nfsExportPolicyModel
	if diags := renameResp.State.Get(context.Background(), &renamedModel); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}
	if renamedModel.Name.ValueString() != "update-policy-renamed" {
		t.Errorf("expected name=update-policy-renamed after rename, got %s", renamedModel.Name.ValueString())
	}
}

// TestNfsExportPolicyResource_Delete verifies DELETE removes the policy.
func TestNfsExportPolicyResource_Delete(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterNfsExportPolicyHandlers(ms.Mux)
	// Register file-systems handler for the delete-guard (ListNfsExportPolicyMembers).
	handlers.RegisterFileSystemHandlers(ms.Mux)

	r := newTestNFSPolicyResource(t, ms)
	s := nfsPolicyResourceSchema(t).Schema

	// Create first.
	plan := nfsPolicyPlanWithName(t, "delete-policy")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNFSPolicyType(), nil), Schema: s},
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
	_, err := r.client.GetNfsExportPolicy(context.Background(), "delete-policy")
	if err == nil || !client.IsNotFound(err) {
		t.Errorf("expected policy to be deleted, got: %v", err)
	}
}

// TestNfsExportPolicyResource_Import verifies ImportState populates all attributes.
func TestNfsExportPolicyResource_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterNfsExportPolicyHandlers(ms.Mux)

	r := newTestNFSPolicyResource(t, ms)
	s := nfsPolicyResourceSchema(t).Schema

	// Create first.
	plan := nfsPolicyPlanWithName(t, "import-policy")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNFSPolicyType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Import by name.
	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNFSPolicyType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "import-policy"}, importResp)

	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState returned error: %s", importResp.Diagnostics)
	}

	var model nfsExportPolicyModel
	if diags := importResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Name.ValueString() != "import-policy" {
		t.Errorf("expected name=import-policy after import, got %s", model.Name.ValueString())
	}
	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected ID to be populated after import")
	}
}

// ---- data source tests -------------------------------------------------------

// newTestNFSPolicyDataSource creates an nfsExportPolicyDataSource wired to the given mock server.
func newTestNFSPolicyDataSource(t *testing.T, ms *testmock.MockServer) *nfsExportPolicyDataSource {
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
	return &nfsExportPolicyDataSource{client: c}
}

// nfsPolicyDataSourceSchema returns the schema for the NFS export policy data source.
func nfsPolicyDataSourceSchema(t *testing.T) datasource.SchemaResponse {
	t.Helper()
	d := &nfsExportPolicyDataSource{}
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	return resp
}

// buildNFSPolicyDSType returns the tftypes.Object for the NFS export policy data source.
func buildNFSPolicyDSType() tftypes.Object {
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":          tftypes.String,
		"name":        tftypes.String,
		"enabled":     tftypes.Bool,
		"is_local":    tftypes.Bool,
		"policy_type": tftypes.String,
		"version":     tftypes.String,
	}}
}

// nullNFSPolicyDSConfig returns a base config map with all data source attributes null.
func nullNFSPolicyDSConfig() map[string]tftypes.Value {
	return map[string]tftypes.Value{
		"id":          tftypes.NewValue(tftypes.String, nil),
		"name":        tftypes.NewValue(tftypes.String, nil),
		"enabled":     tftypes.NewValue(tftypes.Bool, nil),
		"is_local":    tftypes.NewValue(tftypes.Bool, nil),
		"policy_type": tftypes.NewValue(tftypes.String, nil),
		"version":     tftypes.NewValue(tftypes.String, nil),
	}
}

// TestNfsExportPolicyDataSource verifies data source reads policy by name and returns all attributes.
func TestNfsExportPolicyDataSource(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterNfsExportPolicyHandlers(ms.Mux)

	// Create a policy via the resource client so the data source can find it.
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
	_, err = c.PostNfsExportPolicy(context.Background(), "ds-test-policy", client.NfsExportPolicyPost{
		Enabled: &enabled,
	})
	if err != nil {
		t.Fatalf("PostNfsExportPolicy: %v", err)
	}

	d := newTestNFSPolicyDataSource(t, ms)
	s := nfsPolicyDataSourceSchema(t).Schema

	cfg := nullNFSPolicyDSConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "ds-test-policy")

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNFSPolicyDSType(), nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(buildNFSPolicyDSType(), cfg), Schema: s},
	}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("DataSource Read returned error: %s", readResp.Diagnostics)
	}

	var model nfsExportPolicyDataSourceModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Name.ValueString() != "ds-test-policy" {
		t.Errorf("expected name=ds-test-policy, got %s", model.Name.ValueString())
	}
	if !model.Enabled.ValueBool() {
		t.Error("expected enabled=true")
	}
	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected ID to be populated")
	}
}
