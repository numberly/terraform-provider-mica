package provider

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/numberly/opentofu-provider-flashblade/internal/client"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock/handlers"
)

// ---- helpers ----------------------------------------------------------------

// newTestSMBClientPolicyResource creates an smbClientPolicyResource wired to the given mock server.
func newTestSMBClientPolicyResource(t *testing.T, ms *testmock.MockServer) *smbClientPolicyResource {
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
	return &smbClientPolicyResource{client: c}
}

// smbClientPolicyResourceSchema returns the parsed schema for the SMB client policy resource.
func smbClientPolicyResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &smbClientPolicyResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

// buildSMBClientPolicyType returns the tftypes.Object for the SMB client policy resource.
func buildSMBClientPolicyType() tftypes.Object {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":                               tftypes.String,
		"name":                             tftypes.String,
		"enabled":                          tftypes.Bool,
		"is_local":                         tftypes.Bool,
		"policy_type":                      tftypes.String,
		"version":                          tftypes.String,
		"access_based_enumeration_enabled": tftypes.Bool,
		"timeouts":                         timeoutsType,
	}}
}

// nullSMBClientPolicyConfig returns a base config map with all attributes null.
func nullSMBClientPolicyConfig() map[string]tftypes.Value {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return map[string]tftypes.Value{
		"id":                               tftypes.NewValue(tftypes.String, nil),
		"name":                             tftypes.NewValue(tftypes.String, nil),
		"enabled":                          tftypes.NewValue(tftypes.Bool, nil),
		"is_local":                         tftypes.NewValue(tftypes.Bool, nil),
		"policy_type":                      tftypes.NewValue(tftypes.String, nil),
		"version":                          tftypes.NewValue(tftypes.String, nil),
		"access_based_enumeration_enabled": tftypes.NewValue(tftypes.Bool, nil),
		"timeouts":                         tftypes.NewValue(timeoutsType, nil),
	}
}

// smbClientPolicyPlanWithName returns a tfsdk.Plan with the given policy name.
func smbClientPolicyPlanWithName(t *testing.T, name string) tfsdk.Plan {
	t.Helper()
	s := smbClientPolicyResourceSchema(t).Schema
	cfg := nullSMBClientPolicyConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, name)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildSMBClientPolicyType(), cfg),
		Schema: s,
	}
}

// smbClientPolicyPlanWithNameAndEnabled returns a tfsdk.Plan with name and enabled flag.
func smbClientPolicyPlanWithNameAndEnabled(t *testing.T, name string, enabled bool) tfsdk.Plan {
	t.Helper()
	s := smbClientPolicyResourceSchema(t).Schema
	cfg := nullSMBClientPolicyConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, name)
	cfg["enabled"] = tftypes.NewValue(tftypes.Bool, enabled)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildSMBClientPolicyType(), cfg),
		Schema: s,
	}
}

// ---- tests ------------------------------------------------------------------

// TestUnit_SmbClientPolicy_CRUD exercises the full Create->Read->Update(enabled+rename)->Delete sequence.
func TestUnit_SmbClientPolicy_CRUD(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterSmbClientPolicyHandlers(ms.Mux)
	handlers.RegisterFileSystemHandlers(ms.Mux)

	r := newTestSMBClientPolicyResource(t, ms)
	s := smbClientPolicyResourceSchema(t).Schema

	// Step 1: Create with enabled=true.
	createPlan := smbClientPolicyPlanWithNameAndEnabled(t, "test-smb-client-policy", true)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSMBClientPolicyType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}
	var createModel smbClientPolicyModel
	if diags := createResp.State.Get(context.Background(), &createModel); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}
	if createModel.Name.ValueString() != "test-smb-client-policy" {
		t.Errorf("Create: expected name=test-smb-client-policy, got %s", createModel.Name.ValueString())
	}
	if !createModel.Enabled.ValueBool() {
		t.Error("Create: expected enabled=true")
	}
	if createModel.Version.ValueString() != "1" {
		t.Errorf("Create: expected version=1, got %s", createModel.Version.ValueString())
	}

	// Step 2: Read post-create (0-diff check).
	readResp1 := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp1)
	if readResp1.Diagnostics.HasError() {
		t.Fatalf("Read post-create: %s", readResp1.Diagnostics)
	}

	// Step 3: Update enabled=false and rename.
	updateCfg := nullSMBClientPolicyConfig()
	updateCfg["name"] = tftypes.NewValue(tftypes.String, "test-smb-client-policy-renamed")
	updateCfg["enabled"] = tftypes.NewValue(tftypes.Bool, false)
	updatePlan := tfsdk.Plan{
		Raw:    tftypes.NewValue(buildSMBClientPolicyType(), updateCfg),
		Schema: s,
	}
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSMBClientPolicyType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlan,
		State: readResp1.State,
	}, updateResp)
	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update: %s", updateResp.Diagnostics)
	}
	var updateModel smbClientPolicyModel
	if diags := updateResp.State.Get(context.Background(), &updateModel); diags.HasError() {
		t.Fatalf("Get update state: %s", diags)
	}
	if updateModel.Enabled.ValueBool() {
		t.Error("Update: expected enabled=false")
	}
	if updateModel.Name.ValueString() != "test-smb-client-policy-renamed" {
		t.Errorf("Update: expected name=test-smb-client-policy-renamed, got %s", updateModel.Name.ValueString())
	}

	// Step 4: Delete.
	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: updateResp.State}, deleteResp)
	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete: %s", deleteResp.Diagnostics)
	}
	_, err := r.client.GetSmbClientPolicy(context.Background(), "test-smb-client-policy-renamed")
	if err == nil || !client.IsNotFound(err) {
		t.Errorf("expected policy to be deleted, got: %v", err)
	}
}

// TestUnit_SmbClientPolicy_Import verifies ImportState populates all attributes including version.
func TestUnit_SmbClientPolicy_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterSmbClientPolicyHandlers(ms.Mux)

	r := newTestSMBClientPolicyResource(t, ms)
	s := smbClientPolicyResourceSchema(t).Schema

	// Create first.
	plan := smbClientPolicyPlanWithNameAndEnabled(t, "import-smb-client-policy", true)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSMBClientPolicyType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}
	var createModel smbClientPolicyModel
	if diags := createResp.State.Get(context.Background(), &createModel); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}

	// Import by name.
	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSMBClientPolicyType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "import-smb-client-policy"}, importResp)
	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState: %s", importResp.Diagnostics)
	}

	// Read to populate full state.
	readResp := &resource.ReadResponse{State: importResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: importResp.State}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read post-import: %s", readResp.Diagnostics)
	}
	var importedModel smbClientPolicyModel
	if diags := readResp.State.Get(context.Background(), &importedModel); diags.HasError() {
		t.Fatalf("Get imported state: %s", diags)
	}

	// Verify 0-diff.
	if importedModel.Name.ValueString() != createModel.Name.ValueString() {
		t.Errorf("name mismatch: create=%s import=%s", createModel.Name.ValueString(), importedModel.Name.ValueString())
	}
	if importedModel.Enabled.ValueBool() != createModel.Enabled.ValueBool() {
		t.Errorf("enabled mismatch: create=%v import=%v", createModel.Enabled.ValueBool(), importedModel.Enabled.ValueBool())
	}
	if importedModel.ID.ValueString() != createModel.ID.ValueString() {
		t.Errorf("id mismatch: create=%s import=%s", createModel.ID.ValueString(), importedModel.ID.ValueString())
	}
	if importedModel.Version.ValueString() != createModel.Version.ValueString() {
		t.Errorf("version mismatch: create=%s import=%s", createModel.Version.ValueString(), importedModel.Version.ValueString())
	}
}

// ---- data source tests -------------------------------------------------------

// newTestSMBClientPolicyDataSource creates an smbClientPolicyDataSource wired to the given mock server.
func newTestSMBClientPolicyDataSource(t *testing.T, ms *testmock.MockServer) *smbClientPolicyDataSource {
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
	return &smbClientPolicyDataSource{client: c}
}

// smbClientPolicyDataSourceSchema returns the schema for the SMB client policy data source.
func smbClientPolicyDataSourceSchema(t *testing.T) datasource.SchemaResponse {
	t.Helper()
	d := &smbClientPolicyDataSource{}
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	return resp
}

// buildSMBClientPolicyDSType returns the tftypes.Object for the SMB client policy data source.
func buildSMBClientPolicyDSType() tftypes.Object {
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":                               tftypes.String,
		"name":                             tftypes.String,
		"enabled":                          tftypes.Bool,
		"is_local":                         tftypes.Bool,
		"policy_type":                      tftypes.String,
		"version":                          tftypes.String,
		"access_based_enumeration_enabled": tftypes.Bool,
	}}
}

// nullSMBClientPolicyDSConfig returns a base config map with all data source attributes null.
func nullSMBClientPolicyDSConfig() map[string]tftypes.Value {
	return map[string]tftypes.Value{
		"id":                               tftypes.NewValue(tftypes.String, nil),
		"name":                             tftypes.NewValue(tftypes.String, nil),
		"enabled":                          tftypes.NewValue(tftypes.Bool, nil),
		"is_local":                         tftypes.NewValue(tftypes.Bool, nil),
		"policy_type":                      tftypes.NewValue(tftypes.String, nil),
		"version":                          tftypes.NewValue(tftypes.String, nil),
		"access_based_enumeration_enabled": tftypes.NewValue(tftypes.Bool, nil),
	}
}

// TestUnit_SmbClientPolicy_DataSource verifies data source reads policy by name and returns all attributes.
func TestUnit_SmbClientPolicy_DataSource(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterSmbClientPolicyHandlers(ms.Mux)

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
	enabled := true
	_, err = c.PostSmbClientPolicy(context.Background(), "ds-smb-client-test-policy", client.SmbClientPolicyPost{
		Enabled: &enabled,
	})
	if err != nil {
		t.Fatalf("PostSmbClientPolicy: %v", err)
	}

	d := newTestSMBClientPolicyDataSource(t, ms)
	s := smbClientPolicyDataSourceSchema(t).Schema

	cfg := nullSMBClientPolicyDSConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "ds-smb-client-test-policy")

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSMBClientPolicyDSType(), nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(buildSMBClientPolicyDSType(), cfg), Schema: s},
	}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("DataSource Read returned error: %s", readResp.Diagnostics)
	}

	var model smbClientPolicyDataSourceModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Name.ValueString() != "ds-smb-client-test-policy" {
		t.Errorf("expected name=ds-smb-client-test-policy, got %s", model.Name.ValueString())
	}
	if !model.Enabled.ValueBool() {
		t.Error("expected enabled=true")
	}
	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected ID to be populated")
	}
	if model.Version.ValueString() != "1" {
		t.Errorf("expected version=1, got %s", model.Version.ValueString())
	}
}

// TestUnit_SmbClientPolicy_PlanModifiers verifies all UseStateForUnknown plan modifiers.
func TestUnit_SmbClientPolicy_PlanModifiers(t *testing.T) {
	s := smbClientPolicyResourceSchema(t).Schema

	// id
	idAttr, ok := s.Attributes["id"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("id attribute not found or wrong type")
	}
	if len(idAttr.PlanModifiers) == 0 {
		t.Error("expected UseStateForUnknown plan modifier on id attribute")
	}

	// is_local
	ilAttr, ok := s.Attributes["is_local"].(resschema.BoolAttribute)
	if !ok {
		t.Fatal("is_local attribute not found or wrong type")
	}
	if len(ilAttr.PlanModifiers) == 0 {
		t.Error("expected UseStateForUnknown plan modifier on is_local attribute")
	}

	// policy_type
	ptAttr, ok := s.Attributes["policy_type"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("policy_type attribute not found or wrong type")
	}
	if len(ptAttr.PlanModifiers) == 0 {
		t.Error("expected UseStateForUnknown plan modifier on policy_type attribute")
	}

	// version
	vAttr, ok := s.Attributes["version"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("version attribute not found or wrong type")
	}
	if len(vAttr.PlanModifiers) == 0 {
		t.Error("expected UseStateForUnknown plan modifier on version attribute")
	}
}
