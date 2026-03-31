package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/numberly/opentofu-provider-flashblade/internal/client"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock/handlers"
)

func newTestAccountExportDataSource(t *testing.T, ms *testmock.MockServer) *objectStoreAccountExportDataSource {
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
	return &objectStoreAccountExportDataSource{client: c}
}

func accountExportDSSchema(t *testing.T) datasource.SchemaResponse {
	t.Helper()
	d := &objectStoreAccountExportDataSource{}
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	return resp
}

func buildAccountExportDSType() tftypes.Object {
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":           tftypes.String,
		"name":         tftypes.String,
		"account_name": tftypes.String,
		"server_name":  tftypes.String,
		"enabled":      tftypes.Bool,
		"policy_name":  tftypes.String,
	}}
}

func nullAccountExportDSConfig() map[string]tftypes.Value {
	return map[string]tftypes.Value{
		"id":           tftypes.NewValue(tftypes.String, nil),
		"name":         tftypes.NewValue(tftypes.String, nil),
		"account_name": tftypes.NewValue(tftypes.String, nil),
		"server_name":  tftypes.NewValue(tftypes.String, nil),
		"enabled":      tftypes.NewValue(tftypes.Bool, nil),
		"policy_name":  tftypes.NewValue(tftypes.String, nil),
	}
}

func TestUnit_AccountExportDataSource_Read(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterObjectStoreAccountExportHandlers(ms.Mux)
	store.AddObjectStoreAccountExport("myaccount", "s3-policy", "srv-s3")

	d := newTestAccountExportDataSource(t, ms)
	s := accountExportDSSchema(t).Schema
	objType := buildAccountExportDSType()

	cfg := nullAccountExportDSConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "myaccount/myaccount")

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(objType, cfg), Schema: s},
	}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}

	var model objectStoreAccountExportDataSourceModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Name.ValueString() != "myaccount/myaccount" {
		t.Errorf("expected name=myaccount/myaccount, got %s", model.Name.ValueString())
	}
	if model.AccountName.ValueString() != "myaccount" {
		t.Errorf("expected account_name=myaccount, got %s", model.AccountName.ValueString())
	}
	if model.ServerName.ValueString() != "srv-s3" {
		t.Errorf("expected server_name=srv-s3, got %s", model.ServerName.ValueString())
	}
	if model.PolicyName.ValueString() != "s3-policy" {
		t.Errorf("expected policy_name=s3-policy, got %s", model.PolicyName.ValueString())
	}
}

func TestUnit_AccountExportDataSource_NotFound(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterObjectStoreAccountExportHandlers(ms.Mux)

	d := newTestAccountExportDataSource(t, ms)
	s := accountExportDSSchema(t).Schema
	objType := buildAccountExportDSType()

	cfg := nullAccountExportDSConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "nonexistent/nonexistent")

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(objType, cfg), Schema: s},
	}, readResp)

	if !readResp.Diagnostics.HasError() {
		t.Error("expected error diagnostic for not-found account export, got none")
	}

	found := false
	for _, diag := range readResp.Diagnostics {
		if diag.Summary() == "Object store account export not found" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'Object store account export not found' diagnostic, got: %s", readResp.Diagnostics)
	}
}
