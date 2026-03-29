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

func newTestRemoteCredentialsDataSource(t *testing.T, ms *testmock.MockServer) *remoteCredentialsDataSource {
	t.Helper()
	c, err := client.NewClient(context.Background(), client.Config{
		Endpoint:           ms.URL(),
		APIToken:           "test-token",
		InsecureSkipVerify: true,
		MaxRetries:         1,
		RetryBaseDelay:     1 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return &remoteCredentialsDataSource{client: c}
}

func remoteCredentialsDSSchema(t *testing.T) datasource.SchemaResponse {
	t.Helper()
	d := &remoteCredentialsDataSource{}
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	return resp
}

func buildRemoteCredentialsDSType() tftypes.Object {
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":            tftypes.String,
		"name":          tftypes.String,
		"access_key_id": tftypes.String,
		"remote_name":   tftypes.String,
	}}
}

func nullRemoteCredentialsDSConfig() map[string]tftypes.Value {
	return map[string]tftypes.Value{
		"id":            tftypes.NewValue(tftypes.String, nil),
		"name":          tftypes.NewValue(tftypes.String, nil),
		"access_key_id": tftypes.NewValue(tftypes.String, nil),
		"remote_name":   tftypes.NewValue(tftypes.String, nil),
	}
}

func TestUnit_RemoteCredentialsDataSource_Read(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterRemoteCredentialsHandlers(ms.Mux)
	store.Seed(&client.ObjectStoreRemoteCredentials{
		ID:          "rc-001",
		Name:        "remote-cred-1",
		AccessKeyID: "AKTEST123",
		Remote:      client.NamedReference{Name: "remote-array-1"},
	})

	d := newTestRemoteCredentialsDataSource(t, ms)
	s := remoteCredentialsDSSchema(t).Schema
	objType := buildRemoteCredentialsDSType()

	cfg := nullRemoteCredentialsDSConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "remote-cred-1")

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(objType, cfg), Schema: s},
	}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}

	var model remoteCredentialsDataSourceModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.ValueString() != "rc-001" {
		t.Errorf("expected id=rc-001, got %s", model.ID.ValueString())
	}
	if model.AccessKeyID.ValueString() != "AKTEST123" {
		t.Errorf("expected access_key_id=AKTEST123, got %s", model.AccessKeyID.ValueString())
	}
	if model.RemoteName.ValueString() != "remote-array-1" {
		t.Errorf("expected remote_name=remote-array-1, got %s", model.RemoteName.ValueString())
	}
}

func TestUnit_RemoteCredentialsDataSource_NotFound(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterRemoteCredentialsHandlers(ms.Mux)

	d := newTestRemoteCredentialsDataSource(t, ms)
	s := remoteCredentialsDSSchema(t).Schema
	objType := buildRemoteCredentialsDSType()

	cfg := nullRemoteCredentialsDSConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "nonexistent")

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(objType, cfg), Schema: s},
	}, readResp)

	if !readResp.Diagnostics.HasError() {
		t.Error("expected error diagnostic for not-found remote credentials, got none")
	}

	found := false
	for _, diag := range readResp.Diagnostics {
		if diag.Summary() == "Remote credentials not found" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'Remote credentials not found' diagnostic, got: %s", readResp.Diagnostics)
	}
}
