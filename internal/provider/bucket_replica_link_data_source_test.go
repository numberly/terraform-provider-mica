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

func newTestBucketReplicaLinkDataSource(t *testing.T, ms *testmock.MockServer) *bucketReplicaLinkDataSource {
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
	return &bucketReplicaLinkDataSource{client: c}
}

func bucketReplicaLinkDSSchema(t *testing.T) datasource.SchemaResponse {
	t.Helper()
	d := &bucketReplicaLinkDataSource{}
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	return resp
}

func buildBucketReplicaLinkDSType() tftypes.Object {
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":                        tftypes.String,
		"local_bucket_name":         tftypes.String,
		"remote_bucket_name":        tftypes.String,
		"remote_credentials_name":   tftypes.String,
		"remote_name":               tftypes.String,
		"paused":                    tftypes.Bool,
		"cascading_enabled":         tftypes.Bool,
		"direction":                 tftypes.String,
		"status":                    tftypes.String,
		"status_details":            tftypes.String,
		"lag":                       tftypes.Number,
		"recovery_point":            tftypes.Number,
		"object_backlog_count":      tftypes.Number,
		"object_backlog_total_size": tftypes.Number,
	}}
}

func nullBucketReplicaLinkDSConfig() map[string]tftypes.Value {
	return map[string]tftypes.Value{
		"id":                        tftypes.NewValue(tftypes.String, nil),
		"local_bucket_name":         tftypes.NewValue(tftypes.String, nil),
		"remote_bucket_name":        tftypes.NewValue(tftypes.String, nil),
		"remote_credentials_name":   tftypes.NewValue(tftypes.String, nil),
		"remote_name":               tftypes.NewValue(tftypes.String, nil),
		"paused":                    tftypes.NewValue(tftypes.Bool, nil),
		"cascading_enabled":         tftypes.NewValue(tftypes.Bool, nil),
		"direction":                 tftypes.NewValue(tftypes.String, nil),
		"status":                    tftypes.NewValue(tftypes.String, nil),
		"status_details":            tftypes.NewValue(tftypes.String, nil),
		"lag":                       tftypes.NewValue(tftypes.Number, nil),
		"recovery_point":            tftypes.NewValue(tftypes.Number, nil),
		"object_backlog_count":      tftypes.NewValue(tftypes.Number, nil),
		"object_backlog_total_size": tftypes.NewValue(tftypes.Number, nil),
	}
}

func TestUnit_BucketReplicaLinkDataSource_Read(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterBucketReplicaLinkHandlers(ms.Mux)
	store.Seed(&client.BucketReplicaLink{
		ID:                "brl-001",
		LocalBucket:       client.NamedReference{Name: "local-bkt"},
		RemoteBucket:      client.NamedReference{Name: "remote-bkt"},
		Remote:            client.NamedReference{Name: "remote-array"},
		RemoteCredentials: &client.NamedReference{Name: "cred-1"},
		Paused:            false,
		CascadingEnabled:  true,
		Direction:         "outbound",
		Status:            "replicating",
		StatusDetails:     "healthy",
		Lag:               1000,
		RecoveryPoint:     1700000000000,
	})

	d := newTestBucketReplicaLinkDataSource(t, ms)
	s := bucketReplicaLinkDSSchema(t).Schema
	objType := buildBucketReplicaLinkDSType()

	cfg := nullBucketReplicaLinkDSConfig()
	cfg["local_bucket_name"] = tftypes.NewValue(tftypes.String, "local-bkt")
	cfg["remote_bucket_name"] = tftypes.NewValue(tftypes.String, "remote-bkt")

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(objType, cfg), Schema: s},
	}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}

	var model bucketReplicaLinkDataSourceModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.ValueString() != "brl-001" {
		t.Errorf("expected id=brl-001, got %s", model.ID.ValueString())
	}
	if model.Direction.ValueString() != "outbound" {
		t.Errorf("expected direction=outbound, got %s", model.Direction.ValueString())
	}
	if model.Status.ValueString() != "replicating" {
		t.Errorf("expected status=replicating, got %s", model.Status.ValueString())
	}
}

func TestUnit_BucketReplicaLinkDataSource_NotFound(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterBucketReplicaLinkHandlers(ms.Mux)

	d := newTestBucketReplicaLinkDataSource(t, ms)
	s := bucketReplicaLinkDSSchema(t).Schema
	objType := buildBucketReplicaLinkDSType()

	cfg := nullBucketReplicaLinkDSConfig()
	cfg["local_bucket_name"] = tftypes.NewValue(tftypes.String, "nope")
	cfg["remote_bucket_name"] = tftypes.NewValue(tftypes.String, "nope")

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(objType, cfg), Schema: s},
	}, readResp)

	if !readResp.Diagnostics.HasError() {
		t.Error("expected error diagnostic for not-found bucket replica link, got none")
	}

	found := false
	for _, diag := range readResp.Diagnostics {
		if diag.Summary() == "Bucket replica link not found" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'Bucket replica link not found' diagnostic, got: %s", readResp.Diagnostics)
	}
}
