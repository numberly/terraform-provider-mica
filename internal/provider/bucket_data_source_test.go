package provider

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/numberly/opentofu-provider-flashblade/internal/client"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock/handlers"
)

// ---- helpers ----------------------------------------------------------------

// newTestBucketDataSource creates a bucketDataSource wired to the given mock server.
func newTestBucketDataSource(t *testing.T, ms *testmock.MockServer) *bucketDataSource {
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
	return &bucketDataSource{client: c}
}

// bucketDataSourceSchema returns the schema for the bucket data source.
func bucketDataSourceSchema(t *testing.T) datasource.SchemaResponse {
	t.Helper()
	d := &bucketDataSource{}
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	return resp
}

// buildBucketDSType returns the tftypes.Object for the bucket data source schema.
func buildBucketDSType() tftypes.Object {
	spaceType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"data_reduction":      tftypes.Number,
		"snapshots":           tftypes.Number,
		"total_physical":      tftypes.Number,
		"unique":              tftypes.Number,
		"virtual":             tftypes.Number,
		"snapshots_effective": tftypes.Number,
	}}
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":                 tftypes.String,
		"name":               tftypes.String,
		"account":            tftypes.String,
		"created":            tftypes.Number,
		"destroyed":          tftypes.Bool,
		"time_remaining":     tftypes.Number,
		"versioning":         tftypes.String,
		"quota_limit":        tftypes.String,
		"hard_limit_enabled": tftypes.Bool,
		"object_count":       tftypes.Number,
		"bucket_type":        tftypes.String,
		"retention_lock":     tftypes.String,
		"space":              spaceType,
	}}
}

// nullBucketDSConfig returns a base config map with all data source attributes null.
func nullBucketDSConfig() map[string]tftypes.Value {
	spaceType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"data_reduction":      tftypes.Number,
		"snapshots":           tftypes.Number,
		"total_physical":      tftypes.Number,
		"unique":              tftypes.Number,
		"virtual":             tftypes.Number,
		"snapshots_effective": tftypes.Number,
	}}
	return map[string]tftypes.Value{
		"id":                 tftypes.NewValue(tftypes.String, nil),
		"name":               tftypes.NewValue(tftypes.String, nil),
		"account":            tftypes.NewValue(tftypes.String, nil),
		"created":            tftypes.NewValue(tftypes.Number, nil),
		"destroyed":          tftypes.NewValue(tftypes.Bool, nil),
		"time_remaining":     tftypes.NewValue(tftypes.Number, nil),
		"versioning":         tftypes.NewValue(tftypes.String, nil),
		"quota_limit":        tftypes.NewValue(tftypes.String, nil),
		"hard_limit_enabled": tftypes.NewValue(tftypes.Bool, nil),
		"object_count":       tftypes.NewValue(tftypes.Number, nil),
		"bucket_type":        tftypes.NewValue(tftypes.String, nil),
		"retention_lock":     tftypes.NewValue(tftypes.String, nil),
		"space":              tftypes.NewValue(spaceType, nil),
	}
}

// ---- tests ------------------------------------------------------------------

// TestUnit_BucketDataSource verifies the data source reads a bucket by name.
func TestUnit_BucketDataSource(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	accountStore := handlers.RegisterObjectStoreAccountHandlers(ms.Mux)
	handlers.RegisterBucketHandlers(ms.Mux, accountStore)

	// Pre-seed account and bucket.
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
	_, err = c.PostObjectStoreAccount(context.Background(), "ds-account", client.ObjectStoreAccountPost{})
	if err != nil {
		t.Fatalf("PostObjectStoreAccount: %v", err)
	}
	_, err = c.PostBucket(context.Background(), "ds-bucket", client.BucketPost{
		Account:    client.NamedReference{Name: "ds-account"},
		Versioning: "enabled",
		QuotaLimit: "10737418240",
	})
	if err != nil {
		t.Fatalf("PostBucket: %v", err)
	}

	d := newTestBucketDataSource(t, ms)
	s := bucketDataSourceSchema(t).Schema

	cfg := nullBucketDSConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "ds-bucket")

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildBucketDSType(), nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(buildBucketDSType(), cfg), Schema: s},
	}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("DataSource Read returned error: %s", readResp.Diagnostics)
	}

	var model bucketDataSourceModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Name.ValueString() != "ds-bucket" {
		t.Errorf("expected name=ds-bucket, got %s", model.Name.ValueString())
	}
	if model.Account.ValueString() != "ds-account" {
		t.Errorf("expected account=ds-account, got %s", model.Account.ValueString())
	}
	if model.Versioning.ValueString() != "enabled" {
		t.Errorf("expected versioning=enabled, got %s", model.Versioning.ValueString())
	}
	if model.QuotaLimit.ValueString() != "10737418240" {
		t.Errorf("expected quota_limit=10737418240, got %s", model.QuotaLimit.ValueString())
	}
	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected ID to be populated")
	}
}

// TestUnit_BucketDataSource_NotFound verifies that a missing bucket returns an error diagnostic.
func TestUnit_BucketDataSource_NotFound(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	accountStore := handlers.RegisterObjectStoreAccountHandlers(ms.Mux)
	handlers.RegisterBucketHandlers(ms.Mux, accountStore)

	d := newTestBucketDataSource(t, ms)
	s := bucketDataSourceSchema(t).Schema

	cfg := nullBucketDSConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "nonexistent-bucket")

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildBucketDSType(), nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(buildBucketDSType(), cfg), Schema: s},
	}, readResp)

	if !readResp.Diagnostics.HasError() {
		t.Error("expected error diagnostic for not-found bucket, got none")
	}
}
