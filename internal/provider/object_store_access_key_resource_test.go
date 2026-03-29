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

// newTestAccessKeyResource creates an objectStoreAccessKeyResource wired to the given mock server.
func newTestAccessKeyResource(t *testing.T, ms *testmock.MockServer) *objectStoreAccessKeyResource {
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
	c, err := client.NewClient(client.Config{
		Endpoint:           ms.URL(),
		APIToken:           "test-token",
		InsecureSkipVerify: true,
		MaxRetries:         1,
		RetryBaseDelay:     1 * time.Millisecond,
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

// TestUnit_AccessKey_Create verifies POST creates a key and state contains
// access_key_id (and other non-secret fields). With WriteOnly, secret_access_key
// is NOT stored in state but the Create call itself does not error.
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
	// With WriteOnly: true, secret_access_key is NOT persisted in state.
	// The framework strips write-only values from state after resp.State.Set().
	if !model.SecretAccessKey.IsNull() {
		t.Errorf("expected secret_access_key to be null in state after Create (write-only), got %q",
			model.SecretAccessKey.ValueString())
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
	c, _ := client.NewClient(client.Config{
		Endpoint:           ms.URL(),
		APIToken:           "test-token",
		InsecureSkipVerify: true,
		MaxRetries:         1,
		RetryBaseDelay:     1 * time.Millisecond,
	})
	_, err := c.GetObjectStoreAccessKey(context.Background(), created.Name.ValueString())
	if err == nil || !client.IsNotFound(err) {
		t.Errorf("expected access key to be deleted, got: %v", err)
	}
}

// TestUnit_AccessKey_SecretWriteOnly verifies write-only behavior:
// - After Create, secret_access_key is NOT stored in Terraform state (write-only)
// - After Read, secret_access_key is null in state (never returned by API, never stored)
// - No plan diff is generated for secret_access_key (it is always null in state)
func TestUnit_AccessKey_SecretWriteOnly(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	accountStore := handlers.RegisterObjectStoreAccountHandlers(ms.Mux)
	handlers.RegisterObjectStoreUserHandlers(ms.Mux, accountStore)
	handlers.RegisterObjectStoreAccessKeyHandlers(ms.Mux, accountStore)

	seedAccount(t, ms, "secret-account")

	r := newTestAccessKeyResource(t, ms)
	s := accessKeyResourceSchema(t).Schema

	// Step 1: Create — with WriteOnly, the framework does NOT persist secret_access_key in state.
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

	// With WriteOnly: true, the framework strips the value from state after Set().
	// The state value must be null (not persisted).
	if !afterCreate.SecretAccessKey.IsNull() {
		t.Errorf("secret_access_key must be null in state after Create (write-only — not persisted), got %q",
			afterCreate.SecretAccessKey.ValueString())
	}

	// Step 2: Read — state is refreshed from API. Secret remains null (API never returns it, Write-only never stores it).
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

	// After Read, secret_access_key must still be null — no diff, no plan change.
	if !afterRead.SecretAccessKey.IsNull() {
		t.Errorf("secret_access_key must be null in state after Read (write-only), got %q",
			afterRead.SecretAccessKey.ValueString())
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
	// With WriteOnly: true, secret_access_key is NOT persisted in state.
	if !createModel.SecretAccessKey.IsNull() {
		t.Errorf("Create: expected secret_access_key null in state (write-only), got %q",
			createModel.SecretAccessKey.ValueString())
	}

	// Step 2: Read — secret_access_key remains null in state (write-only, API never returns it).
	readResp := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read: %s", readResp.Diagnostics)
	}
	var readModel objectStoreAccessKeyModel
	if diags := readResp.State.Get(context.Background(), &readModel); diags.HasError() {
		t.Fatalf("Get read state: %s", diags)
	}
	if !readModel.SecretAccessKey.IsNull() {
		t.Errorf("Read: secret_access_key must remain null in state (write-only), got %q",
			readModel.SecretAccessKey.ValueString())
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
	r := NewAccessKeyResource()
	if _, ok := r.(resource.ResourceWithImportState); ok {
		t.Error("objectStoreAccessKeyResource must NOT implement ResourceWithImportState — secret unavailable after creation")
	}
}

// TestUnit_AccessKey_WriteOnly verifies that secret_access_key is a write-only attribute
// (not stored in Terraform state) and that Sensitive is false (superseded by WriteOnly).
func TestUnit_AccessKey_WriteOnly(t *testing.T) {
	s := accessKeyResourceSchema(t).Schema

	attr, ok := s.Attributes["secret_access_key"]
	if !ok {
		t.Fatal("secret_access_key attribute not found in schema")
	}
	strAttr, ok := attr.(resschema.StringAttribute)
	if !ok {
		t.Fatalf("secret_access_key is not a resschema.StringAttribute, got %T", attr)
	}
	if !strAttr.WriteOnly {
		t.Error("secret_access_key: expected WriteOnly=true, got false")
	}
	if strAttr.Sensitive {
		t.Error("secret_access_key: expected Sensitive=false (superseded by WriteOnly), got true")
	}
	if len(strAttr.PlanModifiers) != 0 {
		t.Errorf("secret_access_key: expected no PlanModifiers (UseStateForUnknown removed), got %d", len(strAttr.PlanModifiers))
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
