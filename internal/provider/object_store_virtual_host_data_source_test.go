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

func newTestVirtualHostDataSource(t *testing.T, ms *testmock.MockServer) *objectStoreVirtualHostDataSource {
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
	return &objectStoreVirtualHostDataSource{client: c}
}

func virtualHostDSSchema(t *testing.T) datasource.SchemaResponse {
	t.Helper()
	d := &objectStoreVirtualHostDataSource{}
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	return resp
}

func buildVirtualHostDSType() tftypes.Object {
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":               tftypes.String,
		"name":             tftypes.String,
		"filter":           tftypes.String,
		"hostname":         tftypes.String,
		"attached_servers": tftypes.List{ElementType: tftypes.String},
	}}
}

func nullVirtualHostDSConfig() map[string]tftypes.Value {
	return map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, nil),
		"name":             tftypes.NewValue(tftypes.String, nil),
		"filter":           tftypes.NewValue(tftypes.String, nil),
		"hostname":         tftypes.NewValue(tftypes.String, nil),
		"attached_servers": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
	}
}

func TestUnit_VirtualHostDataSource_Read(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterObjectStoreVirtualHostHandlers(ms.Mux)
	store.Seed(&client.ObjectStoreVirtualHost{
		ID:       "vh-123",
		Name:     "s3-vh",
		Hostname: "s3.example.com",
		AttachedServers: []client.NamedReference{
			{Name: "srv-1"},
			{Name: "srv-2"},
		},
	})

	d := newTestVirtualHostDataSource(t, ms)
	s := virtualHostDSSchema(t).Schema
	objType := buildVirtualHostDSType()

	cfg := nullVirtualHostDSConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "s3-vh")

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(objType, cfg), Schema: s},
	}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}

	var model objectStoreVirtualHostDataSourceModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.ValueString() != "vh-123" {
		t.Errorf("expected id=vh-123, got %s", model.ID.ValueString())
	}
	if model.Hostname.ValueString() != "s3.example.com" {
		t.Errorf("expected hostname=s3.example.com, got %s", model.Hostname.ValueString())
	}
}

func TestUnit_VirtualHostDataSource_NotFound(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterObjectStoreVirtualHostHandlers(ms.Mux)

	d := newTestVirtualHostDataSource(t, ms)
	s := virtualHostDSSchema(t).Schema
	objType := buildVirtualHostDSType()

	cfg := nullVirtualHostDSConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "nonexistent")

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(objType, cfg), Schema: s},
	}, readResp)

	if !readResp.Diagnostics.HasError() {
		t.Error("expected error diagnostic for not-found virtual host, got none")
	}

	found := false
	for _, diag := range readResp.Diagnostics {
		if diag.Summary() == "Object store virtual host not found" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'Object store virtual host not found' diagnostic, got: %s", readResp.Diagnostics)
	}
}
