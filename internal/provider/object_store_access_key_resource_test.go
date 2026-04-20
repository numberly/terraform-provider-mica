package provider

import (
	"context"
	"strings"
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

// newTestAccessKeyResource creates an objectStoreAccessKeyResource wired to the given mock server.
func newTestAccessKeyResource(t *testing.T, ms *testmock.MockServer) *objectStoreAccessKeyResource {
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
	return &objectStoreAccessKeyResource{client: c}
}

// accessKeyResourceSchema returns the parsed schema for the access key resource.
func accessKeyResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &objectStoreAccessKeyResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

// buildAccessKeyType returns the tftypes.Object for the access key resource schema.
func buildAccessKeyType() tftypes.Object {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"delete": tftypes.String,
	}}
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"name":                  tftypes.String,
		"object_store_account":  tftypes.String,
		"user":                  tftypes.String,
		"access_key_id":         tftypes.String,
		"secret_access_key":     tftypes.String,
		"created":               tftypes.Number,
		"enabled":               tftypes.Bool,
		"timeouts":              timeoutsType,
	}}
}

// nullAccessKeyConfig returns a base config map with all attributes null.
func nullAccessKeyConfig() map[string]tftypes.Value {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"delete": tftypes.String,
	}}
	return map[string]tftypes.Value{
		"name":                  tftypes.NewValue(tftypes.String, nil),
		"object_store_account":  tftypes.NewValue(tftypes.String, nil),
		"user":                  tftypes.NewValue(tftypes.String, nil),
		"access_key_id":         tftypes.NewValue(tftypes.String, nil),
		"secret_access_key":     tftypes.NewValue(tftypes.String, nil),
		"created":               tftypes.NewValue(tftypes.Number, nil),
		"enabled":               tftypes.NewValue(tftypes.Bool, nil),
		"timeouts":              tftypes.NewValue(timeoutsType, nil),
	}
}

// accessKeyPlanWithAccount returns a tfsdk.Plan with the given account name.
func accessKeyPlanWithAccount(t *testing.T, accountName string) tfsdk.Plan {
	t.Helper()
	s := accessKeyResourceSchema(t).Schema
	cfg := nullAccessKeyConfig()
	cfg["object_store_account"] = tftypes.NewValue(tftypes.String, accountName)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildAccessKeyType(), cfg),
		Schema: s,
	}
}

// seedAccount pre-creates an account in the mock so access key tests can reference it.
func seedAccount(t *testing.T, ms *testmock.MockServer, accountName string) {
	t.Helper()
	c, err := client.NewClient(context.Background(), client.Config{
		Endpoint:           ms.URL(),
		APIToken:           "test-token",
		InsecureSkipVerify: true,
		MaxRetries:         1,
	})
	if err != nil {
		t.Fatalf("NewClient for seed: %v", err)
	}
	_, err = c.PostObjectStoreAccount(context.Background(), accountName, client.ObjectStoreAccountPost{})
	if err != nil {
		t.Fatalf("seed account %q: %v", accountName, err)
	}
}

// ---- tests ------------------------------------------------------------------

// TestUnit_AccessKey_Create verifies POST creates a key and non-secret fields are
// populated in state (name, access_key_id, created, enabled).
// NOTE: In unit tests, resp.State.Set() stores all values including write-only fields.
// The framework server layer (fwserver.NullifyWriteOnlyAttributes) strips write-only
// values before persisting to disk — this only applies in the full Terraform pipeline.
// The write-only guarantee is therefore tested via TestUnit_AccessKey_WriteOnly (schema inspection).
func TestUnit_AccessKey_Create(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	accountStore := handlers.RegisterObjectStoreAccountHandlers(ms.Mux)
	handlers.RegisterObjectStoreUserHandlers(ms.Mux, accountStore)
	handlers.RegisterObjectStoreAccessKeyHandlers(ms.Mux, accountStore)

	seedAccount(t, ms, "test-account")

	r := newTestAccessKeyResource(t, ms)
	s := accessKeyResourceSchema(t).Schema

	plan := accessKeyPlanWithAccount(t, "test-account")
	req := resource.CreateRequest{Plan: plan}
	resp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildAccessKeyType(), nil), Schema: s},
	}

	r.Create(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Create returned error: %s", resp.Diagnostics)
	}

	var model objectStoreAccessKeyModel
	if diags := resp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Name.IsNull() || model.Name.ValueString() == "" {
		t.Error("expected non-empty name after Create")
	}
	if model.AccessKeyID.IsNull() || model.AccessKeyID.ValueString() == "" {
		t.Error("expected non-empty access_key_id after Create")
	}
	// The Create method sets SecretAccessKey — it is available during apply for operator capture.
	// The framework server layer strips it from the persisted state file (write-only guarantee).
	// Schema-level enforcement is verified by TestUnit_AccessKey_WriteOnly.
	if model.SecretAccessKey.IsNull() || model.SecretAccessKey.ValueString() == "" {
		t.Error("expected secret_access_key to be set in Create response (available during apply)")
	}
	if model.Created.IsNull() || model.Created.ValueInt64() == 0 {
		t.Error("expected created timestamp to be populated after Create")
	}
}

// TestUnit_AccessKey_Delete verifies DELETE removes the access key from the store.
func TestUnit_AccessKey_Delete(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	accountStore := handlers.RegisterObjectStoreAccountHandlers(ms.Mux)
	handlers.RegisterObjectStoreUserHandlers(ms.Mux, accountStore)
	handlers.RegisterObjectStoreAccessKeyHandlers(ms.Mux, accountStore)

	seedAccount(t, ms, "delete-account")

	r := newTestAccessKeyResource(t, ms)
	s := accessKeyResourceSchema(t).Schema

	// Create first.
	plan := accessKeyPlanWithAccount(t, "delete-account")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildAccessKeyType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	var created objectStoreAccessKeyModel
	if diags := createResp.State.Get(context.Background(), &created); diags.HasError() {
		t.Fatalf("Get created state: %s", diags)
	}

	// Delete.
	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: createResp.State}, deleteResp)

	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete returned error: %s", deleteResp.Diagnostics)
	}

	// Verify it's gone via the client.
	c, _ := client.NewClient(context.Background(), client.Config{
		Endpoint:           ms.URL(),
		APIToken:           "test-token",
		InsecureSkipVerify: true,
		MaxRetries:         1,
	})
	_, err := c.GetObjectStoreAccessKey(context.Background(), created.Name.ValueString())
	if err == nil || !client.IsNotFound(err) {
		t.Errorf("expected access key to be deleted, got: %v", err)
	}
}

// TestUnit_AccessKey_SecretWriteOnly verifies write-only semantics at the unit level:
// - Create sets the secret on the model response (available to operator during apply)
// - Read does NOT set SecretAccessKey (API never returns it — no new value introduced)
//
// NOTE: The actual state file persistence guarantee (secret_access_key null on disk)
// is enforced by fwserver.NullifyWriteOnlyAttributes in the full Terraform pipeline.
// At the unit test layer (calling r.Create / r.Read directly), tfsdk.State.Set()
// stores values without applying write-only nullification. The schema-level contract
// (WriteOnly: true, Sensitive: false) is verified by TestUnit_AccessKey_WriteOnly.
func TestUnit_AccessKey_SecretWriteOnly(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	accountStore := handlers.RegisterObjectStoreAccountHandlers(ms.Mux)
	handlers.RegisterObjectStoreUserHandlers(ms.Mux, accountStore)
	handlers.RegisterObjectStoreAccessKeyHandlers(ms.Mux, accountStore)

	seedAccount(t, ms, "secret-account")

	r := newTestAccessKeyResource(t, ms)
	s := accessKeyResourceSchema(t).Schema

	// Step 1: Create — secret is returned by the API and set on the model for operator capture.
	plan := accessKeyPlanWithAccount(t, "secret-account")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildAccessKeyType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	var afterCreate objectStoreAccessKeyModel
	if diags := createResp.State.Get(context.Background(), &afterCreate); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}

	// Create sets the secret so it is available during apply (operator can capture via output).
	if afterCreate.SecretAccessKey.IsNull() || afterCreate.SecretAccessKey.ValueString() == "" {
		t.Error("secret_access_key must be set in Create response (available to operator during apply)")
	}
	secretFromCreate := afterCreate.SecretAccessKey.ValueString()

	// Step 2: Read — the resource Read method does NOT set SecretAccessKey (API never returns it).
	// This ensures Read never introduces a spurious non-null value into state after the
	// fwserver nullification has cleared it.
	readResp := &resource.ReadResponse{
		State: createResp.State,
	}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}

	var afterRead objectStoreAccessKeyModel
	if diags := readResp.State.Get(context.Background(), &afterRead); diags.HasError() {
		t.Fatalf("Get read state: %s", diags)
	}

	// Read must not overwrite secret_access_key — it leaves the incoming state value as-is.
	// (In production, the state will have null from fwserver nullification, so Read keeps null.)
	if afterRead.SecretAccessKey.ValueString() != secretFromCreate {
		t.Errorf("Read must not modify SecretAccessKey: want %q, got %q",
			secretFromCreate, afterRead.SecretAccessKey.ValueString())
	}
}

// TestUnit_AccessKey_ForceNew verifies that object_store_account has RequiresReplace semantics.
func TestUnit_AccessKey_ForceNew(t *testing.T) {
	schResp := accessKeyResourceSchema(t).Schema

	// Verify object_store_account has RequiresReplace.
	accountAttr, ok := schResp.Attributes["object_store_account"]
	if !ok {
		t.Fatal("object_store_account attribute not found in schema")
	}
	strAttr, ok := accountAttr.(resschema.StringAttribute)
	if !ok {
		t.Fatalf("object_store_account is not a resschema.StringAttribute, got %T", accountAttr)
	}
	if len(strAttr.PlanModifiers) == 0 {
		t.Error("expected RequiresReplace plan modifier on object_store_account attribute")
	}

	// Verify enabled also has RequiresReplace.
	enabledAttr, ok := schResp.Attributes["enabled"]
	if !ok {
		t.Fatal("enabled attribute not found in schema")
	}
	boolAttr, ok := enabledAttr.(resschema.BoolAttribute)
	if !ok {
		t.Fatalf("enabled is not a resschema.BoolAttribute, got %T", enabledAttr)
	}
	if len(boolAttr.PlanModifiers) == 0 {
		t.Error("expected RequiresReplace plan modifier on enabled attribute")
	}
}

// TestUnit_AccessKey_Lifecycle exercises the Create->Read->Delete sequence (no Update — all fields RequiresReplace).
func TestUnit_AccessKey_Lifecycle(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	accountStore := handlers.RegisterObjectStoreAccountHandlers(ms.Mux)
	handlers.RegisterObjectStoreUserHandlers(ms.Mux, accountStore)
	handlers.RegisterObjectStoreAccessKeyHandlers(ms.Mux, accountStore)

	seedAccount(t, ms, "lifecycle-account")

	r := newTestAccessKeyResource(t, ms)
	s := accessKeyResourceSchema(t).Schema

	// Step 1: Create.
	createPlan := accessKeyPlanWithAccount(t, "lifecycle-account")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildAccessKeyType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}
	var createModel objectStoreAccessKeyModel
	if diags := createResp.State.Get(context.Background(), &createModel); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}
	if createModel.AccessKeyID.IsNull() || createModel.AccessKeyID.ValueString() == "" {
		t.Error("Create: expected non-empty access_key_id")
	}
	// Create sets the secret (available during apply). fwserver nullifies it before disk write.
	if createModel.SecretAccessKey.IsNull() || createModel.SecretAccessKey.ValueString() == "" {
		t.Error("Create: expected secret_access_key set in response (available during apply)")
	}

	// Step 2: Read — Read does not set secret_access_key (API never returns it).
	// The value in state is left as-is from the incoming state (in production: null from fwserver).
	readResp := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read: %s", readResp.Diagnostics)
	}
	var readModel objectStoreAccessKeyModel
	if diags := readResp.State.Get(context.Background(), &readModel); diags.HasError() {
		t.Fatalf("Get read state: %s", diags)
	}
	if readModel.SecretAccessKey.ValueString() != createModel.SecretAccessKey.ValueString() {
		t.Error("Read: must not modify SecretAccessKey — left as-is from incoming state")
	}
	if readModel.AccessKeyID.ValueString() != createModel.AccessKeyID.ValueString() {
		t.Errorf("Read: access_key_id changed: create=%s read=%s",
			createModel.AccessKeyID.ValueString(), readModel.AccessKeyID.ValueString())
	}

	// Step 3: Delete (no Update — access key is fully immutable).
	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: readResp.State}, deleteResp)
	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete: %s", deleteResp.Diagnostics)
	}
	_, err := r.client.GetObjectStoreAccessKey(context.Background(), createModel.Name.ValueString())
	if err == nil || !client.IsNotFound(err) {
		t.Errorf("expected access key to be deleted, got: %v", err)
	}
}

// TestUnit_AccessKey_NoImport verifies the resource does NOT implement ResourceWithImportState.
func TestUnit_AccessKey_NoImport(t *testing.T) {
	r := NewObjectStoreAccessKeyResource()
	if _, ok := r.(resource.ResourceWithImportState); ok {
		t.Error("objectStoreAccessKeyResource must NOT implement ResourceWithImportState — secret unavailable after creation")
	}
}

// TestUnit_AccessKey_SecretSensitive verifies that secret_access_key is Sensitive
// (hidden from plan output) and has UseStateForUnknown (preserved across reads).
// NOTE: WriteOnly was reverted because it strips the value from state, breaking
// outputs and cross-resource references (e.g. writing to Vault).
func TestUnit_AccessKey_SecretSensitive(t *testing.T) {
	s := accessKeyResourceSchema(t).Schema

	attr, ok := s.Attributes["secret_access_key"]
	if !ok {
		t.Fatal("secret_access_key attribute not found in schema")
	}
	strAttr, ok := attr.(resschema.StringAttribute)
	if !ok {
		t.Fatalf("secret_access_key is not a resschema.StringAttribute, got %T", attr)
	}
	if !strAttr.Sensitive {
		t.Error("secret_access_key: expected Sensitive=true")
	}
	if len(strAttr.PlanModifiers) == 0 {
		t.Error("secret_access_key: expected UseStateForUnknown PlanModifier")
	}
}

// accessKeyPlanWithSecret returns a tfsdk.Plan with the given account name and explicit secret.
func accessKeyPlanWithSecret(t *testing.T, accountName, secret string) tfsdk.Plan {
	t.Helper()
	s := accessKeyResourceSchema(t).Schema
	cfg := nullAccessKeyConfig()
	cfg["object_store_account"] = tftypes.NewValue(tftypes.String, accountName)
	cfg["secret_access_key"] = tftypes.NewValue(tftypes.String, secret)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildAccessKeyType(), cfg),
		Schema: s,
	}
}

// TestUnit_AccessKey_CreateWithSecret verifies that when an explicit secret is provided,
// the mock/API echoes back the provided value instead of generating a random one.
func TestUnit_AccessKey_CreateWithSecret(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	accountStore := handlers.RegisterObjectStoreAccountHandlers(ms.Mux)
	handlers.RegisterObjectStoreUserHandlers(ms.Mux, accountStore)
	handlers.RegisterObjectStoreAccessKeyHandlers(ms.Mux, accountStore)

	seedAccount(t, ms, "secret-explicit-account")

	r := newTestAccessKeyResource(t, ms)
	s := accessKeyResourceSchema(t).Schema

	const expectedSecret = "my-cross-array-secret-key-value-12345678"
	plan := accessKeyPlanWithSecret(t, "secret-explicit-account", expectedSecret)
	req := resource.CreateRequest{Plan: plan}
	resp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildAccessKeyType(), nil), Schema: s},
	}

	r.Create(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Create returned error: %s", resp.Diagnostics)
	}

	var model objectStoreAccessKeyModel
	if diags := resp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.SecretAccessKey.ValueString() != expectedSecret {
		t.Errorf("expected secret_access_key=%q, got %q", expectedSecret, model.SecretAccessKey.ValueString())
	}
}

// TestUnit_AccessKey_SecretOptionalComputed verifies schema properties of secret_access_key.
func TestUnit_AccessKey_SecretOptionalComputed(t *testing.T) {
	s := accessKeyResourceSchema(t).Schema

	attr, ok := s.Attributes["secret_access_key"]
	if !ok {
		t.Fatal("secret_access_key attribute not found in schema")
	}
	strAttr, ok := attr.(resschema.StringAttribute)
	if !ok {
		t.Fatalf("secret_access_key is not a resschema.StringAttribute, got %T", attr)
	}
	if !strAttr.Optional {
		t.Error("secret_access_key: expected Optional=true")
	}
	if !strAttr.Computed {
		t.Error("secret_access_key: expected Computed=true")
	}
	if !strAttr.Sensitive {
		t.Error("secret_access_key: expected Sensitive=true")
	}
}

// TestUnit_AccessKey_SecretRequiresReplace verifies RequiresReplace modifier on secret_access_key.
func TestUnit_AccessKey_SecretRequiresReplace(t *testing.T) {
	s := accessKeyResourceSchema(t).Schema

	attr, ok := s.Attributes["secret_access_key"]
	if !ok {
		t.Fatal("secret_access_key attribute not found in schema")
	}
	strAttr, ok := attr.(resschema.StringAttribute)
	if !ok {
		t.Fatalf("secret_access_key is not a resschema.StringAttribute, got %T", attr)
	}

	// Must have at least 2 plan modifiers: UseStateForUnknown + RequiresReplace.
	if len(strAttr.PlanModifiers) < 2 {
		t.Fatalf("expected at least 2 plan modifiers on secret_access_key, got %d", len(strAttr.PlanModifiers))
	}

	// Check that RequiresReplace is present by looking at Description().
	found := false
	for _, pm := range strAttr.PlanModifiers {
		desc := pm.Description(context.Background())
		if desc == "If the value of this attribute changes, Terraform will destroy and recreate the resource." {
			found = true
			break
		}
	}
	if !found {
		t.Error("secret_access_key: RequiresReplace plan modifier not found")
	}
}

// TestUnit_AccessKey_PlanModifiers verifies all RequiresReplace and UseStateForUnknown
// plan modifiers in the object_store_access_key resource schema.
func TestUnit_AccessKey_PlanModifiers(t *testing.T) {
	s := accessKeyResourceSchema(t).Schema

	// access_key_id — UseStateForUnknown (write-once)
	akAttr, ok := s.Attributes["access_key_id"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("access_key_id attribute not found or wrong type")
	}
	if len(akAttr.PlanModifiers) == 0 {
		t.Error("expected UseStateForUnknown plan modifier on access_key_id attribute")
	}

	// object_store_account — RequiresReplace
	accountAttr, ok := s.Attributes["object_store_account"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("object_store_account attribute not found or wrong type")
	}
	if len(accountAttr.PlanModifiers) == 0 {
		t.Error("expected RequiresReplace plan modifier on object_store_account attribute")
	}

	// enabled — RequiresReplace
	enabledAttr, ok := s.Attributes["enabled"].(resschema.BoolAttribute)
	if !ok {
		t.Fatal("enabled attribute not found or wrong type")
	}
	if len(enabledAttr.PlanModifiers) == 0 {
		t.Error("expected RequiresReplace plan modifier on enabled attribute")
	}
}

func TestUnit_ObjectStoreAccessKeyResource_Import_Rejected(t *testing.T) {
	r := &objectStoreAccessKeyResource{}
	resp := &resource.ImportStateResponse{}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "any-name"}, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatalf("expected error diagnostic, got none")
	}
	errs := resp.Diagnostics.Errors()
	if len(errs) != 1 {
		t.Fatalf("expected exactly 1 error diagnostic, got %d", len(errs))
	}
	if errs[0].Summary() != "Import not supported" {
		t.Errorf("expected summary %q, got %q", "Import not supported", errs[0].Summary())
	}
	if !strings.Contains(errs[0].Detail(), "secret_access_key") {
		t.Errorf("expected detail to mention secret_access_key, got %q", errs[0].Detail())
	}
}
