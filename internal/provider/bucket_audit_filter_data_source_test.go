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

func newTestBucketAuditFilterDataSource(t *testing.T, ms *testmock.MockServer) *bucketAuditFilterDataSource {
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
	return &bucketAuditFilterDataSource{client: c}
}

func bafDSSchema(t *testing.T) datasource.SchemaResponse {
	t.Helper()
	d := &bucketAuditFilterDataSource{}
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	return resp
}

func buildBAFDSType() tftypes.Object {
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"bucket_name": tftypes.String,
		"actions":     tftypes.Set{ElementType: tftypes.String},
		"s3_prefixes": tftypes.Set{ElementType: tftypes.String},
	}}
}

func nullBAFDSConfig() map[string]tftypes.Value {
	return map[string]tftypes.Value{
		"bucket_name": tftypes.NewValue(tftypes.String, nil),
		"actions":     tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, nil),
		"s3_prefixes": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, nil),
	}
}

func TestBucketAuditFilterDataSource_Read(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterBucketAuditFilterHandlers(ms.Mux)
	store.Seed(&client.BucketAuditFilter{
		Name:       "baf-ds-1",
		Bucket:     client.NamedReference{Name: "ds-bucket"},
		Actions:    []string{"s3:GetObject", "s3:PutObject"},
		S3Prefixes: []string{"logs/", "data/"},
	})

	d := newTestBucketAuditFilterDataSource(t, ms)
	s := bafDSSchema(t).Schema
	objType := buildBAFDSType()

	cfg := nullBAFDSConfig()
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

	var model bucketAuditFilterDataSourceModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.BucketName.ValueString() != "ds-bucket" {
		t.Errorf("expected bucket_name=ds-bucket, got %s", model.BucketName.ValueString())
	}

	var actions []string
	if diags := model.Actions.ElementsAs(context.Background(), &actions, false); diags.HasError() {
		t.Fatalf("ElementsAs actions: %s", diags)
	}
	if len(actions) != 2 || actions[0] != "s3:GetObject" || actions[1] != "s3:PutObject" {
		t.Errorf("expected actions=[s3:GetObject, s3:PutObject], got %v", actions)
	}

	var prefixes []string
	if diags := model.S3Prefixes.ElementsAs(context.Background(), &prefixes, false); diags.HasError() {
		t.Fatalf("ElementsAs s3_prefixes: %s", diags)
	}
	if len(prefixes) != 2 || prefixes[0] != "logs/" || prefixes[1] != "data/" {
		t.Errorf("expected s3_prefixes=[logs/, data/], got %v", prefixes)
	}
}

func TestBucketAuditFilterDataSource_Read_NotFound(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterBucketAuditFilterHandlers(ms.Mux)

	d := newTestBucketAuditFilterDataSource(t, ms)
	s := bafDSSchema(t).Schema
	objType := buildBAFDSType()

	cfg := nullBAFDSConfig()
	cfg["bucket_name"] = tftypes.NewValue(tftypes.String, "nope")

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(objType, cfg), Schema: s},
	}, readResp)

	if !readResp.Diagnostics.HasError() {
		t.Error("expected error diagnostic for not-found bucket audit filter, got none")
	}

	found := false
	for _, diag := range readResp.Diagnostics {
		if diag.Summary() == "Bucket audit filter not found" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'Bucket audit filter not found' diagnostic, got: %s", readResp.Diagnostics)
	}
}
