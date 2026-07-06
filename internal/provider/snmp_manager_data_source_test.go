package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/numberly/terraform-provider-mica/internal/client"
	"github.com/numberly/terraform-provider-mica/internal/testmock"
	"github.com/numberly/terraform-provider-mica/internal/testmock/handlers"
)

func newTestSnmpManagerDataSource(t *testing.T, ms *testmock.MockServer) *snmpManagerDataSource {
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
	return &snmpManagerDataSource{client: c}
}

func snmpManagerDSSchema(t *testing.T) datasource.SchemaResponse {
	t.Helper()
	d := &snmpManagerDataSource{}
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	return resp
}

func buildSnmpManagerDSType() tftypes.Object {
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":           tftypes.String,
		"name":         tftypes.String,
		"host":         tftypes.String,
		"notification": tftypes.String,
		"version":      tftypes.String,
		"v2c":          snmpManagerV2cType(),
		"v3":           snmpManagerV3Type(),
	}}
}

func nullSnmpManagerDSConfig() map[string]tftypes.Value {
	return map[string]tftypes.Value{
		"id":           tftypes.NewValue(tftypes.String, nil),
		"name":         tftypes.NewValue(tftypes.String, nil),
		"host":         tftypes.NewValue(tftypes.String, nil),
		"notification": tftypes.NewValue(tftypes.String, nil),
		"version":      tftypes.NewValue(tftypes.String, nil),
		"v2c":          tftypes.NewValue(snmpManagerV2cType(), nil),
		"v3":           tftypes.NewValue(snmpManagerV3Type(), nil),
	}
}

// TestUnit_SnmpManagerDataSource_Basic seeds a v3 manager and reads it via the
// data source. Asserts non-sensitive fields populated, sensitive fields null,
// and not-found path surfaces an AddError diagnostic.
func TestUnit_SnmpManagerDataSource_Basic(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterSnmpManagerHandlers(ms.Mux)

	store.Seed(&client.SnmpManager{
		Name:         "ds-mgr",
		Host:         "snmp.example.com",
		Notification: "trap",
		Version:      "v3",
		V3: &client.SnmpV3{
			User:              "purity_user",
			AuthProtocol:      "SHA",
			AuthPassphrase:    "stripped",
			PrivacyProtocol:   "AES",
			PrivacyPassphrase: "stripped",
		},
	})

	d := newTestSnmpManagerDataSource(t, ms)
	s := snmpManagerDSSchema(t).Schema
	objType := buildSnmpManagerDSType()

	cfg := nullSnmpManagerDSConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "ds-mgr")

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(objType, cfg), Schema: s},
	}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}

	var model snmpManagerDataSourceModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Host.ValueString() != "snmp.example.com" {
		t.Errorf("expected host=snmp.example.com, got %s", model.Host.ValueString())
	}
	if model.Notification.ValueString() != "trap" {
		t.Errorf("expected notification=trap, got %s", model.Notification.ValueString())
	}
	if model.Version.ValueString() != "v3" {
		t.Errorf("expected version=v3, got %s", model.Version.ValueString())
	}

	var v3 snmpV3Model
	if diags := model.V3.As(context.Background(), &v3, basetypes.ObjectAsOptions{}); diags.HasError() {
		t.Fatalf("decode v3: %s", diags)
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
		t.Errorf("expected v3.auth_passphrase=null, got %q", v3.AuthPassphrase.ValueString())
	}
	if !v3.PrivacyPassphrase.IsNull() {
		t.Errorf("expected v3.privacy_passphrase=null, got %q", v3.PrivacyPassphrase.ValueString())
	}

	// Not-found path: should surface an AddError diagnostic.
	cfg2 := nullSnmpManagerDSConfig()
	cfg2["name"] = tftypes.NewValue(tftypes.String, "missing")
	notFoundResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(objType, cfg2), Schema: s},
	}, notFoundResp)
	if !notFoundResp.Diagnostics.HasError() {
		t.Error("expected diagnostics error for missing manager, got none")
	}
}
