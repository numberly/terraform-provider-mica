package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/numberly/terraform-provider-mica/internal/client"
	"github.com/numberly/terraform-provider-mica/internal/testmock"
	"github.com/numberly/terraform-provider-mica/internal/testmock/handlers"
)

// ---- helpers ----------------------------------------------------------------

func newTestSnmpManagerResource(t *testing.T, ms *testmock.MockServer) *snmpManagerResource {
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
	return &snmpManagerResource{client: c}
}

func snmpManagerResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &snmpManagerResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

func snmpManagerV2cType() tftypes.Object {
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"community": tftypes.String,
	}}
}

func snmpManagerV3Type() tftypes.Object {
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"user":               tftypes.String,
		"auth_protocol":      tftypes.String,
		"auth_passphrase":    tftypes.String,
		"privacy_protocol":   tftypes.String,
		"privacy_passphrase": tftypes.String,
	}}
}

func buildSnmpManagerType() tftypes.Object {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":           tftypes.String,
		"name":         tftypes.String,
		"host":         tftypes.String,
		"notification": tftypes.String,
		"version":      tftypes.String,
		"v2c":          snmpManagerV2cType(),
		"v3":           snmpManagerV3Type(),
		"timeouts":     timeoutsType,
	}}
}

func nullSnmpManagerConfig() map[string]tftypes.Value {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return map[string]tftypes.Value{
		"id":           tftypes.NewValue(tftypes.String, nil),
		"name":         tftypes.NewValue(tftypes.String, nil),
		"host":         tftypes.NewValue(tftypes.String, nil),
		"notification": tftypes.NewValue(tftypes.String, nil),
		"version":      tftypes.NewValue(tftypes.String, nil),
		"v2c":          tftypes.NewValue(snmpManagerV2cType(), nil),
		"v3":           tftypes.NewValue(snmpManagerV3Type(), nil),
		"timeouts":     tftypes.NewValue(timeoutsType, nil),
	}
}

// snmpManagerPlanV3 builds a v3 plan tfsdk.Plan for the snmp_manager resource.
func snmpManagerPlanV3(t *testing.T, name, host, notification, user, authProto, authPass, privProto, privPass string) tfsdk.Plan {
	t.Helper()
	s := snmpManagerResourceSchema(t).Schema
	cfg := nullSnmpManagerConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, name)
	cfg["host"] = tftypes.NewValue(tftypes.String, host)
	cfg["notification"] = tftypes.NewValue(tftypes.String, notification)
	cfg["version"] = tftypes.NewValue(tftypes.String, "v3")
	cfg["v3"] = tftypes.NewValue(snmpManagerV3Type(), map[string]tftypes.Value{
		"user":               tftypes.NewValue(tftypes.String, user),
		"auth_protocol":      tftypes.NewValue(tftypes.String, authProto),
		"auth_passphrase":    tftypes.NewValue(tftypes.String, authPass),
		"privacy_protocol":   tftypes.NewValue(tftypes.String, privProto),
		"privacy_passphrase": tftypes.NewValue(tftypes.String, privPass),
	})
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildSnmpManagerType(), cfg),
		Schema: s,
	}
}

// ---- tests ------------------------------------------------------------------

// TestUnit_SnmpManagerResource_Lifecycle: create v3 → update host → update notification →
// update auth_protocol (v3 block) → delete.
func TestUnit_SnmpManagerResource_Lifecycle(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterSnmpManagerHandlers(ms.Mux)

	r := newTestSnmpManagerResource(t, ms)
	s := snmpManagerResourceSchema(t).Schema

	// Step 1: Create with full v3 block.
	plan := snmpManagerPlanV3(t, "mgr-life", "snmp.example.com", "trap",
		"purity_user", "MD5", "auth-secret-32max", "AES", "priv-secret-min8")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSnmpManagerType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create returned error: %s", createResp.Diagnostics)
	}

	var afterCreate snmpManagerModel
	if diags := createResp.State.Get(context.Background(), &afterCreate); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}
	if afterCreate.ID.IsNull() || afterCreate.ID.ValueString() == "" {
		t.Error("expected non-empty id after Create")
	}
	if afterCreate.Host.ValueString() != "snmp.example.com" {
		t.Errorf("expected host=snmp.example.com, got %s", afterCreate.Host.ValueString())
	}
	if afterCreate.Version.ValueString() != "v3" {
		t.Errorf("expected version=v3, got %s", afterCreate.Version.ValueString())
	}
	// Sensitive fields must be preserved in state after Create even though API
	// never echoes them.
	var v3 snmpV3Model
	if diags := afterCreate.V3.As(context.Background(), &v3, basetypes.ObjectAsOptions{}); diags.HasError() {
		t.Fatalf("decode v3 after create: %s", diags)
	}
	if v3.AuthPassphrase.ValueString() != "auth-secret-32max" {
		t.Errorf("expected v3.auth_passphrase preserved, got %q", v3.AuthPassphrase.ValueString())
	}
	if v3.PrivacyPassphrase.ValueString() != "priv-secret-min8" {
		t.Errorf("expected v3.privacy_passphrase preserved, got %q", v3.PrivacyPassphrase.ValueString())
	}

	// Step 2: Update host only.
	updatePlanHost := snmpManagerPlanV3(t, "mgr-life", "new.example.com", "trap",
		"purity_user", "MD5", "auth-secret-32max", "AES", "priv-secret-min8")
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSnmpManagerType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlanHost,
		State: createResp.State,
	}, updateResp)
	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update host: %s", updateResp.Diagnostics)
	}
	var afterHost snmpManagerModel
	if diags := updateResp.State.Get(context.Background(), &afterHost); diags.HasError() {
		t.Fatalf("Get update state: %s", diags)
	}
	if afterHost.Host.ValueString() != "new.example.com" {
		t.Errorf("expected host=new.example.com, got %s", afterHost.Host.ValueString())
	}

	// Step 3: Update notification (trap -> inform).
	updatePlanNotif := snmpManagerPlanV3(t, "mgr-life", "new.example.com", "inform",
		"purity_user", "MD5", "auth-secret-32max", "AES", "priv-secret-min8")
	updateResp2 := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSnmpManagerType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlanNotif,
		State: updateResp.State,
	}, updateResp2)
	if updateResp2.Diagnostics.HasError() {
		t.Fatalf("Update notification: %s", updateResp2.Diagnostics)
	}
	var afterNotif snmpManagerModel
	if diags := updateResp2.State.Get(context.Background(), &afterNotif); diags.HasError() {
		t.Fatalf("Get update state: %s", diags)
	}
	if afterNotif.Notification.ValueString() != "inform" {
		t.Errorf("expected notification=inform, got %s", afterNotif.Notification.ValueString())
	}

	// Step 4: Update v3 inner field (auth_protocol MD5 -> SHA).
	updatePlanAuth := snmpManagerPlanV3(t, "mgr-life", "new.example.com", "inform",
		"purity_user", "SHA", "auth-secret-32max", "AES", "priv-secret-min8")
	updateResp3 := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSnmpManagerType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlanAuth,
		State: updateResp2.State,
	}, updateResp3)
	if updateResp3.Diagnostics.HasError() {
		t.Fatalf("Update v3 auth_protocol: %s", updateResp3.Diagnostics)
	}
	// Confirm the mock store received the full v3 block (including sensitive fields).
	stored, ok := store.Get("mgr-life")
	if !ok {
		t.Fatal("manager missing from store after auth_protocol update")
	}
	if stored.V3 == nil || stored.V3.AuthProtocol != "SHA" {
		t.Errorf("expected stored v3.auth_protocol=SHA, got %+v", stored.V3)
	}
	if stored.V3 == nil || stored.V3.AuthPassphrase != "auth-secret-32max" {
		t.Errorf("expected v3 block to be sent atomically including passphrase, got %+v", stored.V3)
	}

	// Step 5: Delete.
	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: updateResp3.State}, deleteResp)
	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete returned error: %s", deleteResp.Diagnostics)
	}
	if _, ok := store.Get("mgr-life"); ok {
		t.Error("expected manager deleted from store")
	}
}

// TestUnit_SnmpManagerResource_Import: seed v3 manager → import by name → confirm
// sensitive fields null and non-sensitive populated.
func TestUnit_SnmpManagerResource_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterSnmpManagerHandlers(ms.Mux)

	store.Seed(&client.SnmpManager{
		Name:         "mgr-import",
		Host:         "snmp.example.com",
		Notification: "trap",
		Version:      "v3",
		V3: &client.SnmpV3{
			User:              "purity_user",
			AuthProtocol:      "SHA",
			AuthPassphrase:    "should-be-stripped",
			PrivacyProtocol:   "AES",
			PrivacyPassphrase: "should-also-be-stripped",
		},
	})

	r := newTestSnmpManagerResource(t, ms)
	s := snmpManagerResourceSchema(t).Schema

	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSnmpManagerType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "mgr-import"}, importResp)
	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState: %s", importResp.Diagnostics)
	}

	var model snmpManagerModel
	if diags := importResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get import state: %s", diags)
	}

	if model.Name.ValueString() != "mgr-import" {
		t.Errorf("expected name=mgr-import, got %s", model.Name.ValueString())
	}
	if model.Version.ValueString() != "v3" {
		t.Errorf("expected version=v3, got %s", model.Version.ValueString())
	}
	if !model.Timeouts.Object.IsNull() {
		t.Error("expected timeouts to be null after import")
	}

	var v3 snmpV3Model
	if diags := model.V3.As(context.Background(), &v3, basetypes.ObjectAsOptions{}); diags.HasError() {
		t.Fatalf("decode v3 after import: %s", diags)
	}
	if v3.User.ValueString() != "purity_user" {
		t.Errorf("expected v3.user=purity_user, got %q", v3.User.ValueString())
	}
	if v3.AuthProtocol.ValueString() != "SHA" {
		t.Errorf("expected v3.auth_protocol=SHA, got %q", v3.AuthProtocol.ValueString())
	}
	if v3.PrivacyProtocol.ValueString() != "AES" {
		t.Errorf("expected v3.privacy_protocol=AES, got %q", v3.PrivacyProtocol.ValueString())
	}
	if !v3.AuthPassphrase.IsNull() {
		t.Errorf("expected v3.auth_passphrase null after import, got %q", v3.AuthPassphrase.ValueString())
	}
	if !v3.PrivacyPassphrase.IsNull() {
		t.Errorf("expected v3.privacy_passphrase null after import, got %q", v3.PrivacyPassphrase.ValueString())
	}
}

// TestUnit_SnmpManagerResource_DriftDetection: create v3 manager → mutate stored
// manager on the array side → Read → confirm state reflects new values and
// sensitive fields are preserved.
func TestUnit_SnmpManagerResource_DriftDetection(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterSnmpManagerHandlers(ms.Mux)

	r := newTestSnmpManagerResource(t, ms)
	s := snmpManagerResourceSchema(t).Schema

	// Create.
	plan := snmpManagerPlanV3(t, "mgr-drift", "snmp.example.com", "trap",
		"purity_user", "MD5", "auth-secret-32max", "AES", "priv-secret-min8")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSnmpManagerType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Mutate every non-sensitive leaf in the mock store (simulate out-of-band
	// drift on all 6 fields the resource watches).
	ok := store.Mutate("mgr-drift", func(m *client.SnmpManager) {
		m.Host = "drifted.example.com"
		m.Notification = "inform"
		m.Version = "v3"
		if m.V3 != nil {
			m.V3.User = "drifted_user"
			m.V3.AuthProtocol = "SHA"
			m.V3.PrivacyProtocol = "DES"
		}
	})
	if !ok {
		t.Fatal("Mutate returned false; manager not in store")
	}

	// Read to detect drift.
	readResp := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read drift detection: %s", readResp.Diagnostics)
	}

	var afterDrift snmpManagerModel
	if diags := readResp.State.Get(context.Background(), &afterDrift); diags.HasError() {
		t.Fatalf("Get drift state: %s", diags)
	}

	// State must reflect drifted API values.
	if afterDrift.Host.ValueString() != "drifted.example.com" {
		t.Errorf("expected host=drifted.example.com after drift, got %s", afterDrift.Host.ValueString())
	}
	if afterDrift.Notification.ValueString() != "inform" {
		t.Errorf("expected notification=inform after drift, got %s", afterDrift.Notification.ValueString())
	}

	var v3 snmpV3Model
	if diags := afterDrift.V3.As(context.Background(), &v3, basetypes.ObjectAsOptions{}); diags.HasError() {
		t.Fatalf("decode v3: %s", diags)
	}
	if v3.User.ValueString() != "drifted_user" {
		t.Errorf("expected v3.user=drifted_user, got %q", v3.User.ValueString())
	}
	if v3.AuthProtocol.ValueString() != "SHA" {
		t.Errorf("expected v3.auth_protocol=SHA, got %q", v3.AuthProtocol.ValueString())
	}
	if v3.PrivacyProtocol.ValueString() != "DES" {
		t.Errorf("expected v3.privacy_protocol=DES, got %q", v3.PrivacyProtocol.ValueString())
	}

	// Sensitive fields must be preserved (write-once contract).
	if v3.AuthPassphrase.ValueString() != "auth-secret-32max" {
		t.Errorf("expected v3.auth_passphrase preserved across drift, got %q", v3.AuthPassphrase.ValueString())
	}
	if v3.PrivacyPassphrase.ValueString() != "priv-secret-min8" {
		t.Errorf("expected v3.privacy_passphrase preserved across drift, got %q", v3.PrivacyPassphrase.ValueString())
	}
}
