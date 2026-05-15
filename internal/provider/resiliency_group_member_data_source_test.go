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

// newTestResiliencyGroupMemberDataSource creates a resiliencyGroupMemberDataSource
// wired to the given mock server.
func newTestResiliencyGroupMemberDataSource(t *testing.T, ms *testmock.MockServer) *resiliencyGroupMemberDataSource {
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
	return &resiliencyGroupMemberDataSource{client: c}
}

// resiliencyGroupMemberDSSchema returns the schema for the data source.
func resiliencyGroupMemberDSSchema(t *testing.T) datasource.SchemaResponse {
	t.Helper()
	d := &resiliencyGroupMemberDataSource{}
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	return resp
}

// buildResiliencyGroupMemberDSType returns the tftypes.Object matching the schema.
func buildResiliencyGroupMemberDSType() tftypes.Object {
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":                    tftypes.String,
		"resiliency_group_name": tftypes.String,
		"member_name":           tftypes.String,
		"group_id":              tftypes.String,
		"group_resource_type":   tftypes.String,
		"member_id":             tftypes.String,
		"member_resource_type":  tftypes.String,
	}}
}

// nullResiliencyGroupMemberDSConfig returns a base config map with all
// attributes null.
func nullResiliencyGroupMemberDSConfig() map[string]tftypes.Value {
	return map[string]tftypes.Value{
		"id":                    tftypes.NewValue(tftypes.String, nil),
		"resiliency_group_name": tftypes.NewValue(tftypes.String, nil),
		"member_name":           tftypes.NewValue(tftypes.String, nil),
		"group_id":              tftypes.NewValue(tftypes.String, nil),
		"group_resource_type":   tftypes.NewValue(tftypes.String, nil),
		"member_id":             tftypes.NewValue(tftypes.String, nil),
		"member_resource_type":  tftypes.NewValue(tftypes.String, nil),
	}
}

// ---- tests ------------------------------------------------------------------

// TestUnit_ResiliencyGroupMemberDataSource_Basic verifies the data source reads
// a seeded membership row by (group_name, member_name).
func TestUnit_ResiliencyGroupMemberDataSource_Basic(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterResiliencyGroupMemberHandlers(ms.Mux)
	store.Seed(&client.ResiliencyGroupMember{
		Group:  client.ResiliencyReference{ID: "rg-id-1", Name: "rg0", ResourceType: "resiliency-groups"},
		Member: client.ResiliencyReference{ID: "fs-id-2", Name: "fs-beta", ResourceType: "file-systems"},
	})

	d := newTestResiliencyGroupMemberDataSource(t, ms)
	s := resiliencyGroupMemberDSSchema(t).Schema

	cfg := nullResiliencyGroupMemberDSConfig()
	cfg["resiliency_group_name"] = tftypes.NewValue(tftypes.String, "rg0")
	cfg["member_name"] = tftypes.NewValue(tftypes.String, "fs-beta")

	objType := buildResiliencyGroupMemberDSType()
	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(objType, cfg), Schema: s},
	}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("DataSource Read returned error: %s", readResp.Diagnostics)
	}

	var model resiliencyGroupMemberDataSourceModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.ValueString() != "rg0/fs-beta" {
		t.Errorf("expected id=rg0/fs-beta, got %s", model.ID.ValueString())
	}
	if model.ResiliencyGroupName.ValueString() != "rg0" {
		t.Errorf("expected resiliency_group_name=rg0, got %s", model.ResiliencyGroupName.ValueString())
	}
	if model.MemberName.ValueString() != "fs-beta" {
		t.Errorf("expected member_name=fs-beta, got %s", model.MemberName.ValueString())
	}
	if model.GroupID.ValueString() != "rg-id-1" {
		t.Errorf("expected group_id=rg-id-1, got %s", model.GroupID.ValueString())
	}
	if model.GroupResourceType.ValueString() != "resiliency-groups" {
		t.Errorf("expected group_resource_type=resiliency-groups, got %s", model.GroupResourceType.ValueString())
	}
	if model.MemberID.ValueString() != "fs-id-2" {
		t.Errorf("expected member_id=fs-id-2, got %s", model.MemberID.ValueString())
	}
	if model.MemberResourceType.ValueString() != "file-systems" {
		t.Errorf("expected member_resource_type=file-systems, got %s", model.MemberResourceType.ValueString())
	}
}

// TestUnit_ResiliencyGroupMemberDataSource_NotFound verifies the error diagnostic
// when no row matches the (group_name, member_name) pair.
func TestUnit_ResiliencyGroupMemberDataSource_NotFound(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterResiliencyGroupMemberHandlers(ms.Mux)

	d := newTestResiliencyGroupMemberDataSource(t, ms)
	s := resiliencyGroupMemberDSSchema(t).Schema

	cfg := nullResiliencyGroupMemberDSConfig()
	cfg["resiliency_group_name"] = tftypes.NewValue(tftypes.String, "rg0")
	cfg["member_name"] = tftypes.NewValue(tftypes.String, "ghost")

	objType := buildResiliencyGroupMemberDSType()
	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(objType, cfg), Schema: s},
	}, readResp)

	if !readResp.Diagnostics.HasError() {
		t.Fatal("expected error diagnostic for not-found member, got none")
	}

	found := false
	for _, diag := range readResp.Diagnostics {
		if diag.Summary() == "Resiliency group member not found" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'Resiliency group member not found' diagnostic, got: %s", readResp.Diagnostics)
	}
}

// TestUnit_ResiliencyGroupMemberDataSource_Schema verifies that both lookup
// attributes are Required and all other attributes are Computed.
func TestUnit_ResiliencyGroupMemberDataSource_Schema(t *testing.T) {
	resp := resiliencyGroupMemberDSSchema(t)
	s := resp.Schema

	requiredAttrs := []string{"resiliency_group_name", "member_name"}
	for _, name := range requiredAttrs {
		attr, ok := s.Attributes[name]
		if !ok {
			t.Errorf("attribute %q not found in schema", name)
			continue
		}
		if !attr.IsRequired() {
			t.Errorf("attribute %q should be Required", name)
		}
	}

	computedAttrs := []string{"id", "group_id", "group_resource_type", "member_id", "member_resource_type"}
	for _, name := range computedAttrs {
		attr, ok := s.Attributes[name]
		if !ok {
			t.Errorf("attribute %q not found in schema", name)
			continue
		}
		if !attr.IsComputed() {
			t.Errorf("attribute %q should be Computed", name)
		}
	}
}
