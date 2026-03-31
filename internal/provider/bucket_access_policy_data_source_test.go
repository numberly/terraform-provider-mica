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

func newTestBucketAccessPolicyDataSource(t *testing.T, ms *testmock.MockServer) *bucketAccessPolicyDataSource {
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
	return &bucketAccessPolicyDataSource{client: c}
}

func bapDSSchema(t *testing.T) datasource.SchemaResponse {
	t.Helper()
	d := &bucketAccessPolicyDataSource{}
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	return resp
}

func buildBAPDSType() tftypes.Object {
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"bucket_name": tftypes.String,
		"id":          tftypes.String,
		"enabled":     tftypes.Bool,
		"rule_count":  tftypes.Number,
	}}
}

func nullBAPDSConfig() map[string]tftypes.Value {
	return map[string]tftypes.Value{
		"bucket_name": tftypes.NewValue(tftypes.String, nil),
		"id":          tftypes.NewValue(tftypes.String, nil),
		"enabled":     tftypes.NewValue(tftypes.Bool, nil),
		"rule_count":  tftypes.NewValue(tftypes.Number, nil),
	}
}

func TestBucketAccessPolicyDataSource_Read(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterBucketAccessPolicyHandlers(ms.Mux)
	store.Seed(&client.BucketAccessPolicy{
		ID:      "bap-1",
		Bucket:  client.NamedReference{Name: "ds-bucket"},
		Enabled: true,
		Rules:   make([]client.BucketAccessPolicyRule, 2),
	})

	d := newTestBucketAccessPolicyDataSource(t, ms)
	s := bapDSSchema(t).Schema
	objType := buildBAPDSType()

	cfg := nullBAPDSConfig()
	cfg["bucket_name"] = tftypes.NewValue(tftypes.String, "ds-bucket")

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(objType, cfg), Schema: s},
	}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}

	var model bucketAccessPolicyDataSourceModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.BucketName.ValueString() != "ds-bucket" {
		t.Errorf("expected bucket_name=ds-bucket, got %s", model.BucketName.ValueString())
	}

	if !model.Enabled.ValueBool() {
		t.Errorf("expected enabled=true, got %v", model.Enabled.ValueBool())
	}

	if model.RuleCount.ValueInt64() != 2 {
		t.Errorf("expected rule_count=2, got %d", model.RuleCount.ValueInt64())
	}
}

func TestBucketAccessPolicyDataSource_Read_NotFound(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterBucketAccessPolicyHandlers(ms.Mux)

	d := newTestBucketAccessPolicyDataSource(t, ms)
	s := bapDSSchema(t).Schema
	objType := buildBAPDSType()

	cfg := nullBAPDSConfig()
	cfg["bucket_name"] = tftypes.NewValue(tftypes.String, "nope")

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(objType, cfg), Schema: s},
	}, readResp)

	if !readResp.Diagnostics.HasError() {
		t.Error("expected error diagnostic for not-found bucket access policy, got none")
	}

	found := false
	for _, diag := range readResp.Diagnostics {
		if diag.Summary() == "Bucket access policy not found" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'Bucket access policy not found' diagnostic, got: %s", readResp.Diagnostics)
	}
}
