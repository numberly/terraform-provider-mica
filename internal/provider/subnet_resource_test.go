package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/numberly/terraform-provider-mica/internal/client"
	"github.com/numberly/terraform-provider-mica/internal/testmock"
	"github.com/numberly/terraform-provider-mica/internal/testmock/handlers"
)

// ---- helpers ----------------------------------------------------------------

// newTestSubnetResource creates a subnetResource wired to the given mock server.
func newTestSubnetResource(t *testing.T, ms *testmock.MockServer) *subnetResource {
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
	return &subnetResource{client: c}
}

// subnetResourceSchema returns the parsed schema for the subnet resource.
func subnetResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &subnetResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

// buildSubnetType returns the tftypes.Object for the subnet resource schema.
func buildSubnetType() tftypes.Object {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":         tftypes.String,
		"name":       tftypes.String,
		"prefix":     tftypes.String,
		"gateway":    tftypes.String,
		"mtu":        tftypes.Number,
		"vlan":       tftypes.Number,
		"lag_name":   tftypes.String,
		"enabled":    tftypes.Bool,
		"services":   tftypes.List{ElementType: tftypes.String},
		"interfaces": tftypes.List{ElementType: tftypes.String},
		"timeouts":   timeoutsType,
	}}
}

// nullSubnetConfig returns a base config map with all resource attributes null.
func nullSubnetConfig() map[string]tftypes.Value {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return map[string]tftypes.Value{
		"id":         tftypes.NewValue(tftypes.String, nil),
		"name":       tftypes.NewValue(tftypes.String, nil),
		"prefix":     tftypes.NewValue(tftypes.String, nil),
		"gateway":    tftypes.NewValue(tftypes.String, nil),
		"mtu":        tftypes.NewValue(tftypes.Number, nil),
		"vlan":       tftypes.NewValue(tftypes.Number, nil),
		"lag_name":   tftypes.NewValue(tftypes.String, nil),
		"enabled":    tftypes.NewValue(tftypes.Bool, nil),
		"services":   tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"interfaces": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"timeouts":   tftypes.NewValue(timeoutsType, nil),
	}
}

// subnetPlanWith returns a tfsdk.Plan for the subnet resource.
func subnetPlanWith(t *testing.T, fields map[string]tftypes.Value) tfsdk.Plan {
	t.Helper()
	s := subnetResourceSchema(t).Schema
	cfg := nullSubnetConfig()
	for k, v := range fields {
		cfg[k] = v
	}
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildSubnetType(), cfg),
		Schema: s,
	}
}

// ---- tests ------------------------------------------------------------------

// TestUnit_SubnetResource_Create verifies Create populates all attributes from API response.
func TestUnit_SubnetResource_Create(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	lagStore := handlers.RegisterLinkAggregationGroupHandlers(ms.Mux)
	lagStore.Seed(&client.LinkAggregationGroup{
		ID:   "lag-001",
		Name: "lag0",
	})
	handlers.RegisterSubnetHandlers(ms.Mux)

	r := newTestSubnetResource(t, ms)
	s := subnetResourceSchema(t).Schema

	plan := subnetPlanWith(t, map[string]tftypes.Value{
		"name":     tftypes.NewValue(tftypes.String, "test-subnet"),
		"prefix":   tftypes.NewValue(tftypes.String, "10.21.200.0/24"),
		"gateway":  tftypes.NewValue(tftypes.String, "10.21.200.1"),
		"mtu":      tftypes.NewValue(tftypes.Number, 9000),
		"vlan":     tftypes.NewValue(tftypes.Number, 100),
		"lag_name": tftypes.NewValue(tftypes.String, "lag0"),
	})

	resp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSubnetType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Create returned error: %s", resp.Diagnostics)
	}

	var model subnetResourceModel
	if diags := resp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected non-empty ID after Create")
	}
	if model.Name.ValueString() != "test-subnet" {
		t.Errorf("expected name=test-subnet, got %s", model.Name.ValueString())
	}
	if model.Prefix.ValueString() != "10.21.200.0/24" {
		t.Errorf("expected prefix=10.21.200.0/24, got %s", model.Prefix.ValueString())
	}
	if model.Gateway.ValueString() != "10.21.200.1" {
		t.Errorf("expected gateway=10.21.200.1, got %s", model.Gateway.ValueString())
	}
	if model.MTU.ValueInt64() != 9000 {
		t.Errorf("expected mtu=9000, got %d", model.MTU.ValueInt64())
	}
	if model.VLAN.ValueInt64() != 100 {
		t.Errorf("expected vlan=100, got %d", model.VLAN.ValueInt64())
	}
	if model.LagName.ValueString() != "lag0" {
		t.Errorf("expected lag_name=lag0, got %s", model.LagName.ValueString())
	}
	if model.Enabled.IsNull() {
		t.Error("expected enabled to be populated")
	}
}

// TestUnit_SubnetResource_Update verifies Update applies partial PATCH and updates state.
func TestUnit_SubnetResource_Update(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterSubnetHandlers(ms.Mux)

	r := newTestSubnetResource(t, ms)
	s := subnetResourceSchema(t).Schema

	// Create first.
	createPlan := subnetPlanWith(t, map[string]tftypes.Value{
		"name":    tftypes.NewValue(tftypes.String, "update-subnet"),
		"prefix":  tftypes.NewValue(tftypes.String, "10.0.0.0/24"),
		"gateway": tftypes.NewValue(tftypes.String, "10.0.0.1"),
		"mtu":     tftypes.NewValue(tftypes.Number, 1500),
	})
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSubnetType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Update gateway and mtu.
	updatePlan := subnetPlanWith(t, map[string]tftypes.Value{
		"name":    tftypes.NewValue(tftypes.String, "update-subnet"),
		"prefix":  tftypes.NewValue(tftypes.String, "10.0.0.0/24"),
		"gateway": tftypes.NewValue(tftypes.String, "10.0.0.254"),
		"mtu":     tftypes.NewValue(tftypes.Number, 9000),
	})
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSubnetType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlan,
		State: createResp.State,
	}, updateResp)

	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update returned error: %s", updateResp.Diagnostics)
	}

	var model subnetResourceModel
	if diags := updateResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Gateway.ValueString() != "10.0.0.254" {
		t.Errorf("expected gateway=10.0.0.254 after update, got %s", model.Gateway.ValueString())
	}
	if model.MTU.ValueInt64() != 9000 {
		t.Errorf("expected mtu=9000 after update, got %d", model.MTU.ValueInt64())
	}
}

// TestUnit_SubnetResource_Delete verifies Delete removes the subnet.
func TestUnit_SubnetResource_Delete(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterSubnetHandlers(ms.Mux)

	r := newTestSubnetResource(t, ms)
	s := subnetResourceSchema(t).Schema

	plan := subnetPlanWith(t, map[string]tftypes.Value{
		"name":   tftypes.NewValue(tftypes.String, "delete-subnet"),
		"prefix": tftypes.NewValue(tftypes.String, "10.10.0.0/24"),
	})
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSubnetType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: createResp.State}, deleteResp)

	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete returned error: %s", deleteResp.Diagnostics)
	}

	// Verify subnet is gone.
	_, err := r.client.GetSubnet(context.Background(), "delete-subnet")
	if err == nil || !client.IsNotFound(err) {
		t.Errorf("expected subnet to be deleted, got: %v", err)
	}
}

// TestUnit_SubnetResource_Import verifies ImportState populates all attributes.
func TestUnit_SubnetResource_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterSubnetHandlers(ms.Mux)

	r := newTestSubnetResource(t, ms)
	s := subnetResourceSchema(t).Schema

	// Create first.
	plan := subnetPlanWith(t, map[string]tftypes.Value{
		"name":   tftypes.NewValue(tftypes.String, "import-subnet"),
		"prefix": tftypes.NewValue(tftypes.String, "10.20.0.0/24"),
	})
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSubnetType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Import by name.
	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSubnetType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "import-subnet"}, importResp)

	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState returned error: %s", importResp.Diagnostics)
	}

	var model subnetResourceModel
	if diags := importResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Name.ValueString() != "import-subnet" {
		t.Errorf("expected name=import-subnet after import, got %s", model.Name.ValueString())
	}
	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected ID to be populated after import")
	}
	if model.Prefix.ValueString() != "10.20.0.0/24" {
		t.Errorf("expected prefix=10.20.0.0/24 after import, got %s", model.Prefix.ValueString())
	}
}

// TestUnit_SubnetResource_Drift verifies Read detects drift when subnet is modified externally.
func TestUnit_SubnetResource_Drift(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterSubnetHandlers(ms.Mux)

	r := newTestSubnetResource(t, ms)
	s := subnetResourceSchema(t).Schema

	// Create via resource.
	plan := subnetPlanWith(t, map[string]tftypes.Value{
		"name":    tftypes.NewValue(tftypes.String, "drift-subnet"),
		"prefix":  tftypes.NewValue(tftypes.String, "10.30.0.0/24"),
		"gateway": tftypes.NewValue(tftypes.String, "10.30.0.1"),
	})
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSubnetType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Simulate external change: modify subnet directly in mock store.
	subnet := store.AddSubnet("drift-subnet-external", "10.30.0.0/24", "")
	// Directly modify the created subnet's gateway via AddSubnet trick —
	// instead seed a modified version via the store.
	_ = subnet // ignore the new one

	// Direct modify: create a new subnet to simulate the seeded one being modified.
	// Use PatchSubnet directly on the client to simulate external change.
	newGateway := "10.30.0.254"
	_, err := r.client.PatchSubnet(context.Background(), "drift-subnet", client.SubnetPatch{
		Gateway: &newGateway,
	})
	if err != nil {
		t.Fatalf("PatchSubnet (external change simulation): %v", err)
	}

	// Read should reflect the new state (drift detected).
	readResp := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}

	var model subnetResourceModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	// State should now reflect the drifted value.
	if model.Gateway.ValueString() != "10.30.0.254" {
		t.Errorf("expected gateway=10.30.0.254 after drift Read, got %s", model.Gateway.ValueString())
	}
}

// TestUnit_SubnetResource_NotFound verifies Read removes resource from state on 404.
func TestUnit_SubnetResource_NotFound(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterSubnetHandlers(ms.Mux)

	r := newTestSubnetResource(t, ms)
	s := subnetResourceSchema(t).Schema

	cfg := nullSubnetConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "nonexistent-subnet")
	cfg["id"] = tftypes.NewValue(tftypes.String, "some-id")
	state := tfsdk.State{Raw: tftypes.NewValue(buildSubnetType(), cfg), Schema: s}

	readResp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), resource.ReadRequest{State: state}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}
	if !readResp.State.Raw.IsNull() {
		t.Error("expected state to be removed (null) when subnet not found")
	}
}

// TestUnit_Subnet_StateUpgrade_V0toV1 verifies that the v0->v1 upgrader is a
// no-op identity: every attribute present in v0 state lands in v1 state unchanged.
func TestUnit_Subnet_StateUpgrade_V0toV1(t *testing.T) {
	r := &subnetResource{}
	upgraders := r.UpgradeState(context.Background())

	upgrader, ok := upgraders[0]
	if !ok {
		t.Fatal("expected v0->v1 upgrader at key 0")
	}
	if upgrader.PriorSchema == nil {
		t.Fatal("expected PriorSchema to be set for v0->v1 upgrader")
	}

	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	stringList := tftypes.List{ElementType: tftypes.String}

	v0Type := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":         tftypes.String,
		"name":       tftypes.String,
		"prefix":     tftypes.String,
		"gateway":    tftypes.String,
		"mtu":        tftypes.Number,
		"vlan":       tftypes.Number,
		"lag_name":   tftypes.String,
		"enabled":    tftypes.Bool,
		"services":   stringList,
		"interfaces": stringList,
		"timeouts":   timeoutsType,
	}}

	v0Val := tftypes.NewValue(v0Type, map[string]tftypes.Value{
		"id":         tftypes.NewValue(tftypes.String, "sub-001"),
		"name":       tftypes.NewValue(tftypes.String, "my-subnet"),
		"prefix":     tftypes.NewValue(tftypes.String, "10.0.0.0/24"),
		"gateway":    tftypes.NewValue(tftypes.String, "10.0.0.1"),
		"mtu":        tftypes.NewValue(tftypes.Number, 1500),
		"vlan":       tftypes.NewValue(tftypes.Number, 0),
		"lag_name":   tftypes.NewValue(tftypes.String, "lag0"),
		"enabled":    tftypes.NewValue(tftypes.Bool, true),
		"services":   tftypes.NewValue(stringList, nil),
		"interfaces": tftypes.NewValue(stringList, nil),
		"timeouts":   tftypes.NewValue(timeoutsType, nil),
	})

	priorState := tfsdk.State{
		Raw:    v0Val,
		Schema: *upgrader.PriorSchema,
	}

	currentSchema := subnetResourceSchema(t).Schema
	resp := &resource.UpgradeStateResponse{
		State: tfsdk.State{
			Raw:    tftypes.NewValue(buildSubnetType(), nil),
			Schema: currentSchema,
		},
	}
	req := resource.UpgradeStateRequest{State: &priorState}

	upgrader.StateUpgrader(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("StateUpgrader returned error: %s", resp.Diagnostics)
	}

	var model subnetResourceModel
	if diags := resp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get upgraded state: %s", diags)
	}

	if model.Name.ValueString() != "my-subnet" {
		t.Errorf("expected name=my-subnet, got %s", model.Name.ValueString())
	}
	if model.Prefix.ValueString() != "10.0.0.0/24" {
		t.Errorf("expected prefix=10.0.0.0/24, got %s", model.Prefix.ValueString())
	}
	if model.VLAN.ValueInt64() != 0 {
		t.Errorf("expected vlan=0, got %d", model.VLAN.ValueInt64())
	}
	if model.LagName.ValueString() != "lag0" {
		t.Errorf("expected lag_name=lag0, got %s", model.LagName.ValueString())
	}
	if model.MTU.ValueInt64() != 1500 {
		t.Errorf("expected mtu=1500, got %d", model.MTU.ValueInt64())
	}
}

// subnetBodyCaptor is a tiny http.Handler wrapper that records the last request
// body for POST/PATCH to /api/2.22/subnets and then delegates to the real mock.
type subnetBodyCaptor struct {
	inner     http.Handler
	lastPOST  []byte
	lastPATCH []byte
}

func (c *subnetBodyCaptor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/api/2.22/subnets" && (r.Method == http.MethodPost || r.Method == http.MethodPatch) {
		buf, _ := io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewReader(buf))
		if r.Method == http.MethodPost {
			c.lastPOST = buf
		} else {
			c.lastPATCH = buf
		}
	}
	c.inner.ServeHTTP(w, r)
}

// newCaptorClient builds a FlashBladeClient pointed at a captor-wrapped mock mux.
func newCaptorClient(t *testing.T, captor *subnetBodyCaptor) (*client.FlashBladeClient, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(captor)
	c, err := client.NewClient(context.Background(), client.Config{
		Endpoint:           srv.URL,
		APIToken:           "test-token",
		InsecureSkipVerify: true,
		MaxRetries:         1,
	})
	if err != nil {
		srv.Close()
		t.Fatalf("NewClient: %v", err)
	}
	return c, srv
}

// TestUnit_Subnet_Create_VLANZero verifies that when the plan sets vlan=0 explicitly,
// the POST body contains "vlan":0 (not omitted). This is the R-001 regression guard.
func TestUnit_Subnet_Create_VLANZero(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterSubnetHandlers(ms.Mux)

	captor := &subnetBodyCaptor{inner: ms.Mux}
	c, captureSrv := newCaptorClient(t, captor)
	defer captureSrv.Close()

	zero := int64(0)
	_, err := c.PostSubnet(context.Background(), "test-subnet", client.SubnetPost{
		Prefix: "10.0.0.0/24",
		VLAN:   &zero,
	})
	if err != nil {
		t.Fatalf("PostSubnet: %v", err)
	}

	if captor.lastPOST == nil {
		t.Fatal("expected POST body to be captured")
	}
	var body map[string]any
	if err := json.Unmarshal(captor.lastPOST, &body); err != nil {
		t.Fatalf("decode POST body: %v", err)
	}
	v, ok := body["vlan"]
	if !ok {
		t.Fatalf("expected 'vlan' in POST body, got: %s", string(captor.lastPOST))
	}
	vf, _ := v.(float64)
	if vf != 0 {
		t.Errorf("expected vlan=0 in POST body, got %v", v)
	}
}

// TestUnit_Subnet_Patch_ClearLag verifies that when lag_name transitions from
// set to null, the PATCH body contains "link_aggregation_group":null (CLEAR).
// This is the R-002 regression guard.
func TestUnit_Subnet_Patch_ClearLag(t *testing.T) {
	state := types.StringValue("lag0")
	plan := types.StringNull()

	patch := client.SubnetPatch{}
	patch.LinkAggregationGroup = doublePointerRefForPatch(state, plan)

	raw, err := json.Marshal(patch)
	if err != nil {
		t.Fatalf("marshal patch: %v", err)
	}

	var body map[string]json.RawMessage
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatalf("decode patch: %v", err)
	}
	v, ok := body["link_aggregation_group"]
	if !ok {
		t.Fatalf("expected 'link_aggregation_group' key in PATCH body, got %s", string(raw))
	}
	if string(v) != "null" {
		t.Errorf("expected link_aggregation_group=null in PATCH body, got %s", string(v))
	}
}
