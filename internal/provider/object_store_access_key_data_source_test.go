package provider

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/numberly/opentofu-provider-flashblade/internal/client"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock/handlers"
)

// newTestAccessKeyDataSource creates an objectStoreAccessKeyDataSource wired to the given mock server.
func newTestAccessKeyDataSource(t *testing.T, ms *testmock.MockServer) *objectStoreAccessKeyDataSource {
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
	return &objectStoreAccessKeyDataSource{client: c}
}

// accessKeyDSSchema returns the parsed schema for the access key data source.
func accessKeyDSSchema(t *testing.T) datasource.SchemaResponse {
	t.Helper()
	ds := &objectStoreAccessKeyDataSource{}
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	return resp
}

// buildAccessKeyDSType returns the tftypes.Object for the access key data source schema.
func buildAccessKeyDSType() tftypes.Object {
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"name":              tftypes.String,
		"access_key_id":     tftypes.String,
		"secret_access_key": tftypes.String,
		"created":           tftypes.Number,
		"enabled":           tftypes.Bool,
		"object_store_account": tftypes.String,
	}}
}

// TestUnit_AccessKeyDataSource verifies the data source reads a key by name.
// secret_access_key is always empty (API does not return it on GET).
func TestUnit_AccessKeyDataSource(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	accountStore := handlers.RegisterObjectStoreAccountHandlers(ms.Mux)
	handlers.RegisterObjectStoreUserHandlers(ms.Mux, accountStore)
	keyStore := handlers.RegisterObjectStoreAccessKeyHandlers(ms.Mux, accountStore)

	// Seed an account and then create a key using the resource helper.
	seedAccount(t, ms, "ds-account")

	// Create an access key via the resource so we know its name.
	keyResource := newTestAccessKeyResource(t, ms)
	s := accessKeyResourceSchema(t).Schema
	plan := accessKeyPlanWithAccount(t, "ds-account")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildAccessKeyType(), nil), Schema: s},
	}
	keyResource.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create access key: %s", createResp.Diagnostics)
	}

	var created objectStoreAccessKeyModel
	if diags := createResp.State.Get(context.Background(), &created); diags.HasError() {
		t.Fatalf("Get created state: %s", diags)
	}

	_ = keyStore // referenced indirectly via mock server

	// Now read via data source.
	ds := newTestAccessKeyDataSource(t, ms)
	dsS := accessKeyDSSchema(t).Schema

	dsType := buildAccessKeyDSType()
	cfg := map[string]tftypes.Value{
		"name":                 tftypes.NewValue(tftypes.String, created.Name.ValueString()),
		"access_key_id":        tftypes.NewValue(tftypes.String, nil),
		"secret_access_key":    tftypes.NewValue(tftypes.String, nil),
		"created":              tftypes.NewValue(tftypes.Number, nil),
		"enabled":              tftypes.NewValue(tftypes.Bool, nil),
		"object_store_account": tftypes.NewValue(tftypes.String, nil),
	}
	config := tfsdk.Config{
		Raw:    tftypes.NewValue(dsType, cfg),
		Schema: dsS,
	}

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(dsType, nil), Schema: dsS},
	}
	ds.Read(context.Background(), datasource.ReadRequest{Config: config}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("DataSource Read returned error: %s", readResp.Diagnostics)
	}

	var dsModel objectStoreAccessKeyDataSourceModel
	if diags := readResp.State.Get(context.Background(), &dsModel); diags.HasError() {
		t.Fatalf("Get DS state: %s", diags)
	}

	if dsModel.Name.ValueString() != created.Name.ValueString() {
		t.Errorf("expected name=%q, got %q", created.Name.ValueString(), dsModel.Name.ValueString())
	}
	if dsModel.AccessKeyID.IsNull() || dsModel.AccessKeyID.ValueString() == "" {
		t.Error("expected access_key_id to be populated in data source")
	}
	// Secret is always empty from GET — that is expected behavior.
	if dsModel.SecretAccessKey.ValueString() != "" {
		t.Errorf("expected secret_access_key to be empty from data source GET, got non-empty value")
	}
}
