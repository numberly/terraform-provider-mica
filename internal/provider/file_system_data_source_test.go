package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/numberly/terraform-provider-mica/internal/client"
	"github.com/numberly/terraform-provider-mica/internal/testmock"
	"github.com/numberly/terraform-provider-mica/internal/testmock/handlers"
)

// ---- helpers ----------------------------------------------------------------

// newTestDataSource creates a filesystemDataSource wired to the given mock server.
func newTestDataSource(t *testing.T, ms *testmock.MockServer) *filesystemDataSource {
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
	return &filesystemDataSource{client: c}
}

// dataSourceSchema returns the schema for the filesystem data source.
func dataSourceSchema(t *testing.T) datasource.SchemaResponse {
	t.Helper()
	d := &filesystemDataSource{}
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	return resp
}

// buildFileSystemDSType returns the tftypes.Object for the filesystem data source schema.
// It mirrors the resource schema except all attributes are Computed (no timeouts block).
func buildFileSystemDSType() tftypes.Object {
	spaceType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"data_reduction":      tftypes.Number,
		"snapshots":           tftypes.Number,
		"total_physical":      tftypes.Number,
		"unique":              tftypes.Number,
		"virtual":             tftypes.Number,
		"snapshots_effective": tftypes.Number,
	}}
	nfsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"enabled":      tftypes.Bool,
		"v3_enabled":   tftypes.Bool,
		"v4_1_enabled": tftypes.Bool,
		"rules":        tftypes.String,
		"transport":    tftypes.String,
	}}
	smbType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"enabled":                         tftypes.Bool,
		"access_based_enumeration_enabled": tftypes.Bool,
		"continuous_availability_enabled":  tftypes.Bool,
		"smb_encryption_enabled":           tftypes.Bool,
	}}
	httpType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"enabled": tftypes.Bool,
	}}
	multiProtocolType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"access_control_style": tftypes.String,
		"safeguard_acls":       tftypes.Bool,
	}}
	defaultQuotasType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"group_quota": tftypes.Number,
		"user_quota":  tftypes.Number,
	}}
	sourceType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":   tftypes.String,
		"name": tftypes.String,
	}}
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":               tftypes.String,
		"name":             tftypes.String,
		"provisioned":      tftypes.Number,
		"destroyed":        tftypes.Bool,
		"time_remaining":   tftypes.Number,
		"created":          tftypes.Number,
		"promotion_status": tftypes.String,
		"writable":         tftypes.Bool,
		"space":            spaceType,
		"nfs":              nfsType,
		"smb":              smbType,
		"http":             httpType,
		"multi_protocol":   multiProtocolType,
		"default_quotas":   defaultQuotasType,
		"source":           sourceType,
	}}
}

// nullDSConfig returns a base config map with all data source attributes null.
func nullDSConfig() map[string]tftypes.Value {
	spaceType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"data_reduction":      tftypes.Number,
		"snapshots":           tftypes.Number,
		"total_physical":      tftypes.Number,
		"unique":              tftypes.Number,
		"virtual":             tftypes.Number,
		"snapshots_effective": tftypes.Number,
	}}
	nfsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"enabled":      tftypes.Bool,
		"v3_enabled":   tftypes.Bool,
		"v4_1_enabled": tftypes.Bool,
		"rules":        tftypes.String,
		"transport":    tftypes.String,
	}}
	smbType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"enabled":                         tftypes.Bool,
		"access_based_enumeration_enabled": tftypes.Bool,
		"continuous_availability_enabled":  tftypes.Bool,
		"smb_encryption_enabled":           tftypes.Bool,
	}}
	httpType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"enabled": tftypes.Bool,
	}}
	multiProtocolType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"access_control_style": tftypes.String,
		"safeguard_acls":       tftypes.Bool,
	}}
	defaultQuotasType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"group_quota": tftypes.Number,
		"user_quota":  tftypes.Number,
	}}
	sourceType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":   tftypes.String,
		"name": tftypes.String,
	}}
	return map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, nil),
		"name":             tftypes.NewValue(tftypes.String, nil),
		"provisioned":      tftypes.NewValue(tftypes.Number, nil),
		"destroyed":        tftypes.NewValue(tftypes.Bool, nil),
		"time_remaining":   tftypes.NewValue(tftypes.Number, nil),
		"created":          tftypes.NewValue(tftypes.Number, nil),
		"promotion_status": tftypes.NewValue(tftypes.String, nil),
		"writable":         tftypes.NewValue(tftypes.Bool, nil),
		"space":            tftypes.NewValue(spaceType, nil),
		"nfs":              tftypes.NewValue(nfsType, nil),
		"smb":              tftypes.NewValue(smbType, nil),
		"http":             tftypes.NewValue(httpType, nil),
		"multi_protocol":   tftypes.NewValue(multiProtocolType, nil),
		"default_quotas":   tftypes.NewValue(defaultQuotasType, nil),
		"source":           tftypes.NewValue(sourceType, nil),
	}
}

// ---- tests ------------------------------------------------------------------

// TestUnit_FileSystemDataSource verifies data source reads file system by name and returns all attributes.
func TestUnit_FileSystemDataSource(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterFileSystemHandlers(ms.Mux)

	// Create a file system via the resource client so the data source can find it.
	c, err := client.NewClient(context.Background(), client.Config{
		Endpoint:           ms.URL(),
		APIToken:           "test-token",
		InsecureSkipVerify: true,
		MaxRetries:         1,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	_, err = c.PostFileSystem(context.Background(), client.FileSystemPost{
		Name:        "ds-test-fs",
		Provisioned: 2147483648,
		NFS:         &client.NFSConfig{Enabled: true},
	})
	if err != nil {
		t.Fatalf("PostFileSystem: %v", err)
	}

	d := newTestDataSource(t, ms)
	s := dataSourceSchema(t).Schema

	cfg := nullDSConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "ds-test-fs")

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildFileSystemDSType(), nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(buildFileSystemDSType(), cfg), Schema: s},
	}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("DataSource Read returned error: %s", readResp.Diagnostics)
	}

	var model filesystemDataSourceModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Name.ValueString() != "ds-test-fs" {
		t.Errorf("expected name=ds-test-fs, got %s", model.Name.ValueString())
	}
	if model.Provisioned.ValueInt64() != 2147483648 {
		t.Errorf("expected provisioned=2147483648, got %d", model.Provisioned.ValueInt64())
	}
	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected ID to be populated")
	}
	if model.NFS.IsNull() || model.NFS.IsUnknown() {
		t.Fatal("expected NFS block to be populated")
	}
	var nfsModel filesystemNFSModel
	if diags := model.NFS.As(context.Background(), &nfsModel, basetypes.ObjectAsOptions{}); diags.HasError() {
		t.Fatalf("NFS.As: %s", diags)
	}
	if !nfsModel.Enabled.ValueBool() {
		t.Error("expected NFS.enabled=true")
	}
}

// TestUnit_FileSystemDataSource_NotFound verifies that a missing file system returns an error diagnostic.
func TestUnit_FileSystemDataSource_NotFound(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterFileSystemHandlers(ms.Mux)

	d := newTestDataSource(t, ms)
	s := dataSourceSchema(t).Schema

	cfg := nullDSConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "nonexistent-fs")

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildFileSystemDSType(), nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(buildFileSystemDSType(), cfg), Schema: s},
	}, readResp)

	if !readResp.Diagnostics.HasError() {
		t.Error("expected error diagnostic for not-found file system, got none")
	}

	// Check the error message contains the file system name.
	found := false
	for _, diag := range readResp.Diagnostics {
		if diag.Summary() == "File system not found" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'File system not found' diagnostic, got: %s", readResp.Diagnostics)
	}
}
