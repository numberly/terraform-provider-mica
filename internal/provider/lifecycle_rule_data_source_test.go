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

func ptrInt64LRDS(v int64) *int64 { return &v }

func newTestLifecycleRuleDataSource(t *testing.T, ms *testmock.MockServer) *lifecycleRuleDataSource {
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
	return &lifecycleRuleDataSource{client: c}
}

func lifecycleRuleDSSchema(t *testing.T) datasource.SchemaResponse {
	t.Helper()
	d := &lifecycleRuleDataSource{}
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	return resp
}

func buildLifecycleRuleDSType() tftypes.Object {
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":                                      tftypes.String,
		"bucket_name":                              tftypes.String,
		"rule_id":                                  tftypes.String,
		"prefix":                                   tftypes.String,
		"enabled":                                  tftypes.Bool,
		"abort_incomplete_multipart_uploads_after": tftypes.Number,
		"keep_current_version_for":                 tftypes.Number,
		"keep_current_version_until":               tftypes.Number,
		"keep_previous_version_for":                tftypes.Number,
		"cleanup_expired_object_delete_marker":     tftypes.Bool,
	}}
}

func nullLifecycleRuleDSConfig() map[string]tftypes.Value {
	return map[string]tftypes.Value{
		"id":                                      tftypes.NewValue(tftypes.String, nil),
		"bucket_name":                              tftypes.NewValue(tftypes.String, nil),
		"rule_id":                                  tftypes.NewValue(tftypes.String, nil),
		"prefix":                                   tftypes.NewValue(tftypes.String, nil),
		"enabled":                                  tftypes.NewValue(tftypes.Bool, nil),
		"abort_incomplete_multipart_uploads_after": tftypes.NewValue(tftypes.Number, nil),
		"keep_current_version_for":                 tftypes.NewValue(tftypes.Number, nil),
		"keep_current_version_until":               tftypes.NewValue(tftypes.Number, nil),
		"keep_previous_version_for":                tftypes.NewValue(tftypes.Number, nil),
		"cleanup_expired_object_delete_marker":     tftypes.NewValue(tftypes.Bool, nil),
	}
}

func TestUnit_LifecycleRuleDataSource_Read(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterLifecycleRuleHandlers(ms.Mux)
	store.Seed(&client.LifecycleRule{
		ID:                    "lcr-ds-1",
		Name:                  "ds-bucket/ds-rule",
		Bucket:                client.NamedReference{Name: "ds-bucket"},
		RuleID:                "ds-rule",
		Prefix:                "logs/",
		Enabled:               true,
		KeepCurrentVersionFor: ptrInt64LRDS(86400000),
		CleanupExpiredObjectDeleteMarker: true,
	})

	d := newTestLifecycleRuleDataSource(t, ms)
	s := lifecycleRuleDSSchema(t).Schema
	objType := buildLifecycleRuleDSType()

	cfg := nullLifecycleRuleDSConfig()
	cfg["bucket_name"] = tftypes.NewValue(tftypes.String, "ds-bucket")
	cfg["rule_id"] = tftypes.NewValue(tftypes.String, "ds-rule")

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(objType, cfg), Schema: s},
	}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}

	var model lifecycleRuleDataSourceModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.ValueString() != "lcr-ds-1" {
		t.Errorf("expected id=lcr-ds-1, got %s", model.ID.ValueString())
	}
	if model.BucketName.ValueString() != "ds-bucket" {
		t.Errorf("expected bucket_name=ds-bucket, got %s", model.BucketName.ValueString())
	}
	if model.RuleID.ValueString() != "ds-rule" {
		t.Errorf("expected rule_id=ds-rule, got %s", model.RuleID.ValueString())
	}
	if model.Prefix.ValueString() != "logs/" {
		t.Errorf("expected prefix=logs/, got %s", model.Prefix.ValueString())
	}
	if model.Enabled.ValueBool() != true {
		t.Error("expected enabled=true")
	}
	if model.KeepCurrentVersionFor.ValueInt64() != 86400000 {
		t.Errorf("expected keep_current_version_for=86400000, got %d", model.KeepCurrentVersionFor.ValueInt64())
	}
	if model.CleanupExpiredObjectDeleteMarker.ValueBool() != true {
		t.Error("expected cleanup_expired_object_delete_marker=true")
	}
}

func TestUnit_LifecycleRuleDataSource_Read_NotFound(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterLifecycleRuleHandlers(ms.Mux)

	d := newTestLifecycleRuleDataSource(t, ms)
	s := lifecycleRuleDSSchema(t).Schema
	objType := buildLifecycleRuleDSType()

	cfg := nullLifecycleRuleDSConfig()
	cfg["bucket_name"] = tftypes.NewValue(tftypes.String, "nope")
	cfg["rule_id"] = tftypes.NewValue(tftypes.String, "nope")

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(objType, cfg), Schema: s},
	}, readResp)

	if !readResp.Diagnostics.HasError() {
		t.Error("expected error diagnostic for not-found lifecycle rule, got none")
	}

	found := false
	for _, diag := range readResp.Diagnostics {
		if diag.Summary() == "Lifecycle rule not found" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'Lifecycle rule not found' diagnostic, got: %s", readResp.Diagnostics)
	}
}
