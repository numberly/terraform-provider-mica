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

// newTestRemoteCredentialsResource creates a remoteCredentialsResource wired to the given mock server.
func newTestRemoteCredentialsResource(t *testing.T, ms *testmock.MockServer) *remoteCredentialsResource {
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
	return &remoteCredentialsResource{client: c}
}

// remoteCredentialsResourceSchema returns the parsed schema for the remote credentials resource.
func remoteCredentialsResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &remoteCredentialsResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

// buildRemoteCredentialsType returns the tftypes.Object for the remote credentials resource schema.
func buildRemoteCredentialsType() tftypes.Object {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":                tftypes.String,
		"name":              tftypes.String,
		"access_key_id":     tftypes.String,
		"secret_access_key": tftypes.String,
		"remote_name":       tftypes.String,
		"timeouts":          timeoutsType,
	}}
}

// nullRemoteCredentialsConfig returns a base config map with all attributes null.
func nullRemoteCredentialsConfig() map[string]tftypes.Value {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return map[string]tftypes.Value{
		"id":                tftypes.NewValue(tftypes.String, nil),
		"name":              tftypes.NewValue(tftypes.String, nil),
		"access_key_id":     tftypes.NewValue(tftypes.String, nil),
		"secret_access_key": tftypes.NewValue(tftypes.String, nil),
		"remote_name":       tftypes.NewValue(tftypes.String, nil),
		"timeouts":          tftypes.NewValue(timeoutsType, nil),
	}
}

// remoteCredentialsPlanWith returns a tfsdk.Plan with the given field values.
func remoteCredentialsPlanWith(t *testing.T, name, remoteName, accessKeyID, secretAccessKey string) tfsdk.Plan {
	t.Helper()
	s := remoteCredentialsResourceSchema(t).Schema
	cfg := nullRemoteCredentialsConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, name)
	cfg["remote_name"] = tftypes.NewValue(tftypes.String, remoteName)
	cfg["access_key_id"] = tftypes.NewValue(tftypes.String, accessKeyID)
	cfg["secret_access_key"] = tftypes.NewValue(tftypes.String, secretAccessKey)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildRemoteCredentialsType(), cfg),
		Schema: s,
	}
}

// ---- tests ------------------------------------------------------------------

// TestUnit_RemoteCredentials_Create verifies POST creates remote credentials and
// state contains id, name, remote_name, and preserves user-provided secrets.
func TestUnit_RemoteCredentials_Create(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterRemoteCredentialsHandlers(ms.Mux)

	r := newTestRemoteCredentialsResource(t, ms)
	s := remoteCredentialsResourceSchema(t).Schema

	plan := remoteCredentialsPlanWith(t, "test-cred", "remote-array", "AKID123", "SECRET456")
	req := resource.CreateRequest{Plan: plan}
	resp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildRemoteCredentialsType(), nil), Schema: s},
	}

	r.Create(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Create returned error: %s", resp.Diagnostics)
	}

	var model remoteCredentialsModel
	if diags := resp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected non-empty id after Create")
	}
	if model.Name.ValueString() != "test-cred" {
		t.Errorf("expected name=test-cred, got %s", model.Name.ValueString())
	}
	if model.RemoteName.ValueString() != "remote-array" {
		t.Errorf("expected remote_name=remote-array, got %s", model.RemoteName.ValueString())
	}
	if model.AccessKeyID.ValueString() != "AKID123" {
		t.Errorf("expected access_key_id=AKID123, got %s", model.AccessKeyID.ValueString())
	}
	if model.SecretAccessKey.ValueString() != "SECRET456" {
		t.Errorf("expected secret_access_key=SECRET456, got %s", model.SecretAccessKey.ValueString())
	}
}

// TestUnit_RemoteCredentials_Read verifies GET populates id, name, access_key_id, remote_name
// and preserves secret_access_key from prior state (GET does not return it).
func TestUnit_RemoteCredentials_Read(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterRemoteCredentialsHandlers(ms.Mux)

	r := newTestRemoteCredentialsResource(t, ms)
	s := remoteCredentialsResourceSchema(t).Schema

	// Create first.
	plan := remoteCredentialsPlanWith(t, "read-cred", "remote-array", "AKID-READ", "SECRET-READ")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildRemoteCredentialsType(), nil), Schema: s},
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

	var model remoteCredentialsModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Name.ValueString() != "read-cred" {
		t.Errorf("expected name=read-cred, got %s", model.Name.ValueString())
	}
	if model.AccessKeyID.ValueString() != "AKID-READ" {
		t.Errorf("expected access_key_id=AKID-READ, got %s", model.AccessKeyID.ValueString())
	}
	// secret_access_key must be preserved from prior state (GET returns empty string).
	if model.SecretAccessKey.ValueString() != "SECRET-READ" {
		t.Errorf("expected secret_access_key preserved as SECRET-READ, got %s", model.SecretAccessKey.ValueString())
	}
}

// TestUnit_RemoteCredentials_Update verifies PATCH with changed access_key_id and/or
// secret_access_key rotates keys; state reflects new values.
func TestUnit_RemoteCredentials_Update(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterRemoteCredentialsHandlers(ms.Mux)

	r := newTestRemoteCredentialsResource(t, ms)
	s := remoteCredentialsResourceSchema(t).Schema

	// Create first.
	createPlan := remoteCredentialsPlanWith(t, "update-cred", "remote-array", "OLD-AKID", "OLD-SECRET")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildRemoteCredentialsType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Update with new keys.
	updatePlan := remoteCredentialsPlanWith(t, "update-cred", "remote-array", "NEW-AKID", "NEW-SECRET")
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildRemoteCredentialsType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlan,
		State: createResp.State,
	}, updateResp)

	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update returned error: %s", updateResp.Diagnostics)
	}

	var model remoteCredentialsModel
	if diags := updateResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.AccessKeyID.ValueString() != "NEW-AKID" {
		t.Errorf("expected access_key_id=NEW-AKID after update, got %s", model.AccessKeyID.ValueString())
	}
	if model.SecretAccessKey.ValueString() != "NEW-SECRET" {
		t.Errorf("expected secret_access_key=NEW-SECRET after update, got %s", model.SecretAccessKey.ValueString())
	}
}

// TestUnit_RemoteCredentials_Delete verifies DELETE removes the resource;
// subsequent GET returns not-found.
func TestUnit_RemoteCredentials_Delete(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterRemoteCredentialsHandlers(ms.Mux)

	r := newTestRemoteCredentialsResource(t, ms)
	s := remoteCredentialsResourceSchema(t).Schema

	// Create first.
	plan := remoteCredentialsPlanWith(t, "delete-cred", "remote-array", "AKID-DEL", "SECRET-DEL")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildRemoteCredentialsType(), nil), Schema: s},
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
	_, err := r.client.GetRemoteCredentials(context.Background(), "delete-cred")
	if err == nil || !client.IsNotFound(err) {
		t.Errorf("expected remote credentials to be deleted, got: %v", err)
	}
}

// TestUnit_RemoteCredentials_Import verifies ImportState by name populates all fields;
// secret_access_key is empty string (user must provide in config or use ignore_changes).
func TestUnit_RemoteCredentials_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterRemoteCredentialsHandlers(ms.Mux)

	r := newTestRemoteCredentialsResource(t, ms)
	s := remoteCredentialsResourceSchema(t).Schema

	// Create a credential first so import can find it.
	plan := remoteCredentialsPlanWith(t, "import-cred", "remote-array", "AKID-IMP", "SECRET-IMP")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildRemoteCredentialsType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Import by name.
	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildRemoteCredentialsType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "import-cred"}, importResp)

	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState returned error: %s", importResp.Diagnostics)
	}

	var model remoteCredentialsModel
	if diags := importResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected non-empty id after import")
	}
	if model.Name.ValueString() != "import-cred" {
		t.Errorf("expected name=import-cred, got %s", model.Name.ValueString())
	}
	if model.RemoteName.ValueString() != "remote-array" {
		t.Errorf("expected remote_name=remote-array, got %s", model.RemoteName.ValueString())
	}
	// secret_access_key must be empty after import.
	if model.SecretAccessKey.ValueString() != "" {
		t.Errorf("expected empty secret_access_key after import, got %s", model.SecretAccessKey.ValueString())
	}
}

// TestUnit_RemoteCredentials_Idempotence verifies Create -> Read -> compare
// all non-secret fields are identical (no drift).
func TestUnit_RemoteCredentials_Idempotence(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterRemoteCredentialsHandlers(ms.Mux)

	r := newTestRemoteCredentialsResource(t, ms)
	s := remoteCredentialsResourceSchema(t).Schema

	// Create.
	plan := remoteCredentialsPlanWith(t, "idemp-cred", "remote-array", "AKID-IDEMP", "SECRET-IDEMP")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildRemoteCredentialsType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	var afterCreate remoteCredentialsModel
	if diags := createResp.State.Get(context.Background(), &afterCreate); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}

	// Read.
	readResp := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read: %s", readResp.Diagnostics)
	}

	var afterRead remoteCredentialsModel
	if diags := readResp.State.Get(context.Background(), &afterRead); diags.HasError() {
		t.Fatalf("Get read state: %s", diags)
	}

	// Compare non-secret fields.
	if afterRead.ID.ValueString() != afterCreate.ID.ValueString() {
		t.Errorf("id drift: create=%s read=%s", afterCreate.ID.ValueString(), afterRead.ID.ValueString())
	}
	if afterRead.Name.ValueString() != afterCreate.Name.ValueString() {
		t.Errorf("name drift: create=%s read=%s", afterCreate.Name.ValueString(), afterRead.Name.ValueString())
	}
	if afterRead.RemoteName.ValueString() != afterCreate.RemoteName.ValueString() {
		t.Errorf("remote_name drift: create=%s read=%s", afterCreate.RemoteName.ValueString(), afterRead.RemoteName.ValueString())
	}
	if afterRead.AccessKeyID.ValueString() != afterCreate.AccessKeyID.ValueString() {
		t.Errorf("access_key_id drift: create=%s read=%s", afterCreate.AccessKeyID.ValueString(), afterRead.AccessKeyID.ValueString())
	}
	// secret_access_key is also preserved from state (not from GET).
	if afterRead.SecretAccessKey.ValueString() != afterCreate.SecretAccessKey.ValueString() {
		t.Errorf("secret_access_key drift: create=%s read=%s", afterCreate.SecretAccessKey.ValueString(), afterRead.SecretAccessKey.ValueString())
	}
}

// TestUnit_RemoteCredentials_Lifecycle exercises the full Create -> Read -> Update -> Read -> Delete sequence.
func TestUnit_RemoteCredentials_Lifecycle(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterRemoteCredentialsHandlers(ms.Mux)

	r := newTestRemoteCredentialsResource(t, ms)
	s := remoteCredentialsResourceSchema(t).Schema

	// Step 1: Create.
	createPlan := remoteCredentialsPlanWith(t, "lifecycle-cred", "remote-array", "AKID-V1", "SECRET-V1")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildRemoteCredentialsType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	var createModel remoteCredentialsModel
	if diags := createResp.State.Get(context.Background(), &createModel); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}
	if createModel.AccessKeyID.ValueString() != "AKID-V1" {
		t.Errorf("Create: expected access_key_id=AKID-V1, got %s", createModel.AccessKeyID.ValueString())
	}

	// Step 2: Read.
	readResp := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read after Create: %s", readResp.Diagnostics)
	}

	// Step 3: Update (key rotation).
	updatePlan := remoteCredentialsPlanWith(t, "lifecycle-cred", "remote-array", "AKID-V2", "SECRET-V2")
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildRemoteCredentialsType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlan,
		State: readResp.State,
	}, updateResp)
	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update: %s", updateResp.Diagnostics)
	}

	var updateModel remoteCredentialsModel
	if diags := updateResp.State.Get(context.Background(), &updateModel); diags.HasError() {
		t.Fatalf("Get update state: %s", diags)
	}
	if updateModel.AccessKeyID.ValueString() != "AKID-V2" {
		t.Errorf("Update: expected access_key_id=AKID-V2, got %s", updateModel.AccessKeyID.ValueString())
	}

	// Step 4: Read after update.
	readResp2 := &resource.ReadResponse{State: updateResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: updateResp.State}, readResp2)
	if readResp2.Diagnostics.HasError() {
		t.Fatalf("Read after Update: %s", readResp2.Diagnostics)
	}

	// Step 5: Delete.
	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: readResp2.State}, deleteResp)
	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete: %s", deleteResp.Diagnostics)
	}

	_, err := r.client.GetRemoteCredentials(context.Background(), "lifecycle-cred")
	if err == nil || !client.IsNotFound(err) {
		t.Errorf("expected remote credentials to be deleted, got: %v", err)
	}
}

// TestUnit_RemoteCredentials_Schema verifies schema properties:
// - name and remote_name have RequiresReplace
// - access_key_id and secret_access_key are Required+Sensitive
func TestUnit_RemoteCredentials_Schema(t *testing.T) {
	s := remoteCredentialsResourceSchema(t).Schema

	// name: Required + RequiresReplace.
	nameAttr, ok := s.Attributes["name"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("name attribute not found or wrong type")
	}
	if !nameAttr.Required {
		t.Error("name: expected Required=true")
	}
	if len(nameAttr.PlanModifiers) == 0 {
		t.Error("name: expected RequiresReplace plan modifier")
	}

	// remote_name: Required + RequiresReplace.
	remoteNameAttr, ok := s.Attributes["remote_name"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("remote_name attribute not found or wrong type")
	}
	if !remoteNameAttr.Required {
		t.Error("remote_name: expected Required=true")
	}
	if len(remoteNameAttr.PlanModifiers) == 0 {
		t.Error("remote_name: expected RequiresReplace plan modifier")
	}

	// access_key_id: Required + Sensitive.
	akidAttr, ok := s.Attributes["access_key_id"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("access_key_id attribute not found or wrong type")
	}
	if !akidAttr.Required {
		t.Error("access_key_id: expected Required=true")
	}
	if !akidAttr.Sensitive {
		t.Error("access_key_id: expected Sensitive=true")
	}

	// secret_access_key: Required + Sensitive.
	secretAttr, ok := s.Attributes["secret_access_key"].(resschema.StringAttribute)
	if !ok {
		t.Fatal("secret_access_key attribute not found or wrong type")
	}
	if !secretAttr.Required {
		t.Error("secret_access_key: expected Required=true")
	}
	if !secretAttr.Sensitive {
		t.Error("secret_access_key: expected Sensitive=true")
	}
}
