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

func newTestFileSystemExportDataSource(t *testing.T, ms *testmock.MockServer) *fileSystemExportDataSource {
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
	return &fileSystemExportDataSource{client: c}
}

func fileSystemExportDSSchema(t *testing.T) datasource.SchemaResponse {
	t.Helper()
	d := &fileSystemExportDataSource{}
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	return resp
}

func buildFileSystemExportDSType() tftypes.Object {
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":                tftypes.String,
		"name":              tftypes.String,
		"export_name":       tftypes.String,
		"file_system_name":  tftypes.String,
		"server_name":       tftypes.String,
		"share_policy_name": tftypes.String,
		"enabled":           tftypes.Bool,
		"policy_type":       tftypes.String,
		"status":            tftypes.String,
	}}
}

func nullFileSystemExportDSConfig() map[string]tftypes.Value {
	return map[string]tftypes.Value{
		"id":                tftypes.NewValue(tftypes.String, nil),
		"name":              tftypes.NewValue(tftypes.String, nil),
		"export_name":       tftypes.NewValue(tftypes.String, nil),
		"file_system_name":  tftypes.NewValue(tftypes.String, nil),
		"server_name":       tftypes.NewValue(tftypes.String, nil),
		"share_policy_name": tftypes.NewValue(tftypes.String, nil),
		"enabled":           tftypes.NewValue(tftypes.Bool, nil),
		"policy_type":       tftypes.NewValue(tftypes.String, nil),
		"status":            tftypes.NewValue(tftypes.String, nil),
	}
}

func TestUnit_FileSystemExportDataSource_Read(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterFileSystemExportHandlers(ms.Mux)
	store.AddFileSystemExport("myfs", "default-nfs-policy", "srv-main")

	d := newTestFileSystemExportDataSource(t, ms)
	s := fileSystemExportDSSchema(t).Schema
	objType := buildFileSystemExportDSType()

	cfg := nullFileSystemExportDSConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "myfs/myfs")

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(objType, cfg), Schema: s},
	}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}

	var model fileSystemExportDataSourceModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Name.ValueString() != "myfs/myfs" {
		t.Errorf("expected name=myfs/myfs, got %s", model.Name.ValueString())
	}
	if model.ExportName.ValueString() != "myfs" {
		t.Errorf("expected export_name=myfs, got %s", model.ExportName.ValueString())
	}
	if model.ServerName.ValueString() != "srv-main" {
		t.Errorf("expected server_name=srv-main, got %s", model.ServerName.ValueString())
	}
	if model.PolicyType.ValueString() != "nfs" {
		t.Errorf("expected policy_type=nfs, got %s", model.PolicyType.ValueString())
	}
}

func TestUnit_FileSystemExportDataSource_NotFound(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterFileSystemExportHandlers(ms.Mux)

	d := newTestFileSystemExportDataSource(t, ms)
	s := fileSystemExportDSSchema(t).Schema
	objType := buildFileSystemExportDSType()

	cfg := nullFileSystemExportDSConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "nonexistent/nonexistent")

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(objType, cfg), Schema: s},
	}, readResp)

	if !readResp.Diagnostics.HasError() {
		t.Error("expected error diagnostic for not-found file system export, got none")
	}

	found := false
	for _, diag := range readResp.Diagnostics {
		if diag.Summary() == "File system export not found" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'File system export not found' diagnostic, got: %s", readResp.Diagnostics)
	}
}
