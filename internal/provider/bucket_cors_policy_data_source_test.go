package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/numberly/terraform-provider-mica/internal/client"
	"github.com/numberly/terraform-provider-mica/internal/testmock"
	"github.com/numberly/terraform-provider-mica/internal/testmock/handlers"
)

// ---- helpers ----------------------------------------------------------------

func newTestBucketCorsPolicyDataSource(t *testing.T, ms *testmock.MockServer) *bucketCorsPolicyDataSource {
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
	return &bucketCorsPolicyDataSource{client: c}
}

func bucketCorsPolicyDSSchema(t *testing.T) datasource.SchemaResponse {
	t.Helper()
	d := &bucketCorsPolicyDataSource{}
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	return resp
}

func buildBucketCorsPolicyDSType() tftypes.Object {
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":          tftypes.String,
		"bucket_name": tftypes.String,
		"is_local":    tftypes.Bool,
		"policy_type": tftypes.String,
	}}
}

func nullBucketCorsPolicyDSConfig() map[string]tftypes.Value {
	return map[string]tftypes.Value{
		"id":          tftypes.NewValue(tftypes.String, nil),
		"bucket_name": tftypes.NewValue(tftypes.String, nil),
		"is_local":    tftypes.NewValue(tftypes.Bool, nil),
		"policy_type": tftypes.NewValue(tftypes.String, nil),
	}
}

// ---- tests ------------------------------------------------------------------

func TestUnit_BucketCorsPolicyDataSource_Basic(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterCorsHandlers(ms.Mux)

	store.Seed(&client.CrossOriginResourceSharingPolicy{
		ID:         "cors-ds-1",
		Name:       "ds-bucket-cors",
		Bucket:     client.NamedReference{Name: "ds-bucket"},
		IsLocal:    true,
		PolicyType: "cross-origin-resource-sharing",
		Rules: []client.CorsRule{
			{Name: "corsrule0", AllowedHeaders: []string{"*"}, AllowedMethods: []string{"GET"}, AllowedOrigins: []string{"https://example.com"}},
		},
	})

	d := newTestBucketCorsPolicyDataSource(t, ms)
	s := bucketCorsPolicyDSSchema(t).Schema

	cfg := nullBucketCorsPolicyDSConfig()
	cfg["bucket_name"] = tftypes.NewValue(tftypes.String, "ds-bucket")

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildBucketCorsPolicyDSType(), nil), Schema: s},
	}
	config := tfsdk.Config{Raw: tftypes.NewValue(buildBucketCorsPolicyDSType(), cfg), Schema: s}
	d.Read(context.Background(), datasource.ReadRequest{Config: config}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}

	var model bucketCorsPolicyDataSourceModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}
	if model.ID.ValueString() != "cors-ds-1" {
		t.Errorf("expected id=cors-ds-1, got %s", model.ID.ValueString())
	}
	if !model.IsLocal.ValueBool() {
		t.Error("expected is_local=true")
	}
	if model.BucketName.ValueString() != "ds-bucket" {
		t.Errorf("expected bucket_name=ds-bucket, got %s", model.BucketName.ValueString())
	}
}

func TestUnit_BucketCorsPolicyDataSource_NotFound(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterCorsHandlers(ms.Mux)

	d := newTestBucketCorsPolicyDataSource(t, ms)
	s := bucketCorsPolicyDSSchema(t).Schema

	cfg := nullBucketCorsPolicyDSConfig()
	cfg["bucket_name"] = tftypes.NewValue(tftypes.String, "ghost-bucket")

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildBucketCorsPolicyDSType(), nil), Schema: s},
	}
	config := tfsdk.Config{Raw: tftypes.NewValue(buildBucketCorsPolicyDSType(), cfg), Schema: s}
	d.Read(context.Background(), datasource.ReadRequest{Config: config}, readResp)

	if !readResp.Diagnostics.HasError() {
		t.Error("expected error for not-found CORS policy")
	}
}
