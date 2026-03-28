package provider

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/soulkyu/terraform-provider-flashblade/internal/client"
	"github.com/soulkyu/terraform-provider-flashblade/internal/testmock"
	"github.com/soulkyu/terraform-provider-flashblade/internal/testmock/handlers"
)

// ---- helpers ----------------------------------------------------------------

func newTestArraySmtpResource(t *testing.T, ms *testmock.MockServer) *arraySmtpResource {
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
	return &arraySmtpResource{client: c}
}

func arraySmtpResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &arraySmtpResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

// alertWatcherTFType returns the tftypes.Object for a single alert watcher.
func alertWatcherTFType() tftypes.Object {
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"email":                         tftypes.String,
		"enabled":                       tftypes.Bool,
		"minimum_notification_severity": tftypes.String,
	}}
}

func buildArraySmtpType() tftypes.Object {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":              tftypes.String,
		"relay_host":      tftypes.String,
		"sender_domain":   tftypes.String,
		"encryption_mode": tftypes.String,
		"alert_watchers":  tftypes.Set{ElementType: alertWatcherTFType()},
		"timeouts":        timeoutsType,
	}}
}

func nullArraySmtpConfig() map[string]tftypes.Value {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return map[string]tftypes.Value{
		"id":              tftypes.NewValue(tftypes.String, nil),
		"relay_host":      tftypes.NewValue(tftypes.String, nil),
		"sender_domain":   tftypes.NewValue(tftypes.String, nil),
		"encryption_mode": tftypes.NewValue(tftypes.String, nil),
		"alert_watchers":  tftypes.NewValue(tftypes.Set{ElementType: alertWatcherTFType()}, nil),
		"timeouts":        tftypes.NewValue(timeoutsType, nil),
	}
}

// buildAlertWatcherValue builds a tftypes.Value for a single alert watcher.
func buildAlertWatcherValue(email string, enabled bool, severity string) tftypes.Value {
	return tftypes.NewValue(alertWatcherTFType(), map[string]tftypes.Value{
		"email":                         tftypes.NewValue(tftypes.String, email),
		"enabled":                       tftypes.NewValue(tftypes.Bool, enabled),
		"minimum_notification_severity": tftypes.NewValue(tftypes.String, severity),
	})
}

func arraySmtpPlanWith(t *testing.T, relayHost, senderDomain string, watchers []tftypes.Value) tfsdk.Plan {
	t.Helper()
	s := arraySmtpResourceSchema(t).Schema
	cfg := nullArraySmtpConfig()
	cfg["relay_host"] = tftypes.NewValue(tftypes.String, relayHost)
	cfg["sender_domain"] = tftypes.NewValue(tftypes.String, senderDomain)
	cfg["encryption_mode"] = tftypes.NewValue(tftypes.String, "none")
	cfg["alert_watchers"] = tftypes.NewValue(tftypes.Set{ElementType: alertWatcherTFType()}, watchers)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildArraySmtpType(), cfg),
		Schema: s,
	}
}

// ---- resource tests ---------------------------------------------------------

// TestArraySmtpResource_Create verifies that Create sets SMTP config and creates alert watcher.
func TestArraySmtpResource_Create(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterArrayAdminHandlers(ms.Mux)

	r := newTestArraySmtpResource(t, ms)
	s := arraySmtpResourceSchema(t).Schema

	watchers := []tftypes.Value{
		buildAlertWatcherValue("admin@example.com", true, "warning"),
	}
	plan := arraySmtpPlanWith(t, "smtp.example.com", "example.com", watchers)
	resp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildArraySmtpType(), nil), Schema: s},
	}

	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Create returned error: %s", resp.Diagnostics)
	}

	var model arraySmtpModel
	if diags := resp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected non-empty ID after Create")
	}
	if model.RelayHost.ValueString() != "smtp.example.com" {
		t.Errorf("expected relay_host=smtp.example.com, got %s", model.RelayHost.ValueString())
	}
	if model.SenderDomain.ValueString() != "example.com" {
		t.Errorf("expected sender_domain=example.com, got %s", model.SenderDomain.ValueString())
	}

	var watcherModels []alertWatcherModel
	if diags := model.AlertWatchers.ElementsAs(context.Background(), &watcherModels, false); diags.HasError() {
		t.Fatalf("Get alert_watchers: %s", diags)
	}
	if len(watcherModels) != 1 {
		t.Fatalf("expected 1 alert watcher, got %d", len(watcherModels))
	}
	if watcherModels[0].Email.ValueString() != "admin@example.com" {
		t.Errorf("expected watcher email=admin@example.com, got %s", watcherModels[0].Email.ValueString())
	}
	if watcherModels[0].MinimumNotificationSeverity.ValueString() != "warning" {
		t.Errorf("expected severity=warning, got %s", watcherModels[0].MinimumNotificationSeverity.ValueString())
	}
}

// TestArraySmtpResource_Update verifies adding a second watcher and changing relay_host.
func TestArraySmtpResource_Update(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterArrayAdminHandlers(ms.Mux)

	r := newTestArraySmtpResource(t, ms)
	s := arraySmtpResourceSchema(t).Schema

	// Create with one watcher.
	createWatchers := []tftypes.Value{
		buildAlertWatcherValue("admin@example.com", true, "warning"),
	}
	createPlan := arraySmtpPlanWith(t, "smtp.example.com", "example.com", createWatchers)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildArraySmtpType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Update: change relay_host and add second watcher.
	updateWatchers := []tftypes.Value{
		buildAlertWatcherValue("admin@example.com", true, "warning"),
		buildAlertWatcherValue("ops@example.com", true, "error"),
	}
	updatePlan := arraySmtpPlanWith(t, "smtp2.example.com", "example.com", updateWatchers)
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildArraySmtpType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlan,
		State: createResp.State,
	}, updateResp)

	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update returned error: %s", updateResp.Diagnostics)
	}

	var model arraySmtpModel
	if diags := updateResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.RelayHost.ValueString() != "smtp2.example.com" {
		t.Errorf("expected relay_host=smtp2.example.com after update, got %s", model.RelayHost.ValueString())
	}

	var watcherModels []alertWatcherModel
	if diags := model.AlertWatchers.ElementsAs(context.Background(), &watcherModels, false); diags.HasError() {
		t.Fatalf("Get alert_watchers: %s", diags)
	}
	if len(watcherModels) != 2 {
		t.Errorf("expected 2 alert watchers after update, got %d", len(watcherModels))
	}
}

// TestArraySmtpResource_Delete verifies that Delete resets SMTP config and removes all watchers.
func TestArraySmtpResource_Delete(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterArrayAdminHandlers(ms.Mux)

	r := newTestArraySmtpResource(t, ms)
	s := arraySmtpResourceSchema(t).Schema

	// Create with one watcher.
	createWatchers := []tftypes.Value{
		buildAlertWatcherValue("admin@example.com", true, "warning"),
	}
	createPlan := arraySmtpPlanWith(t, "smtp.example.com", "example.com", createWatchers)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildArraySmtpType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Delete resets config and removes watchers.
	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: createResp.State}, deleteResp)

	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete returned error: %s", deleteResp.Diagnostics)
	}

	// Verify SMTP reset.
	smtp, err := r.client.GetSmtpServer(context.Background())
	if err != nil {
		t.Fatalf("GetSmtpServer after delete: %v", err)
	}
	if smtp.RelayHost != "" {
		t.Errorf("expected relay_host to be reset, got %q", smtp.RelayHost)
	}
	if smtp.SenderDomain != "" {
		t.Errorf("expected sender_domain to be reset, got %q", smtp.SenderDomain)
	}

	// Verify watchers removed.
	watcherList, err := r.client.GetAlertWatchers(context.Background())
	if err != nil {
		t.Fatalf("GetAlertWatchers after delete: %v", err)
	}
	if len(watcherList) != 0 {
		t.Errorf("expected alert watchers to be empty after delete, got %v", watcherList)
	}
}

// TestArraySmtpResource_Import verifies ImportState populates SMTP config and watchers.
func TestArraySmtpResource_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterArrayAdminHandlers(ms.Mux)

	r := newTestArraySmtpResource(t, ms)
	s := arraySmtpResourceSchema(t).Schema

	// Seed state via direct client calls.
	relayHost := "smtp.import.example.com"
	senderDomain := "import.example.com"
	encMode := "tls"
	if _, err := r.client.PatchSmtpServer(context.Background(), client.SmtpServerPatch{
		RelayHost:      &relayHost,
		SenderDomain:   &senderDomain,
		EncryptionMode: &encMode,
	}); err != nil {
		t.Fatalf("PatchSmtpServer: %v", err)
	}
	if _, err := r.client.PostAlertWatcher(context.Background(), "import@example.com", client.AlertWatcherPost{
		MinimumNotificationSeverity: "error",
	}); err != nil {
		t.Fatalf("PostAlertWatcher: %v", err)
	}

	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildArraySmtpType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "default"}, importResp)

	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState returned error: %s", importResp.Diagnostics)
	}

	var model arraySmtpModel
	if diags := importResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected non-empty ID after import")
	}
	if model.RelayHost.ValueString() != "smtp.import.example.com" {
		t.Errorf("expected relay_host=smtp.import.example.com after import, got %s", model.RelayHost.ValueString())
	}
	if model.EncryptionMode.ValueString() != "tls" {
		t.Errorf("expected encryption_mode=tls after import, got %s", model.EncryptionMode.ValueString())
	}

	var watcherModels []alertWatcherModel
	if diags := model.AlertWatchers.ElementsAs(context.Background(), &watcherModels, false); diags.HasError() {
		t.Fatalf("Get alert_watchers: %s", diags)
	}
	if len(watcherModels) != 1 {
		t.Fatalf("expected 1 alert watcher after import, got %d", len(watcherModels))
	}
	if watcherModels[0].Email.ValueString() != "import@example.com" {
		t.Errorf("expected watcher email=import@example.com, got %s", watcherModels[0].Email.ValueString())
	}
}

// TestArraySmtpResource_WatcherRemoval verifies that removing a watcher from the plan deletes it.
func TestArraySmtpResource_WatcherRemoval(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterArrayAdminHandlers(ms.Mux)

	r := newTestArraySmtpResource(t, ms)
	s := arraySmtpResourceSchema(t).Schema

	// Create with 2 watchers.
	createWatchers := []tftypes.Value{
		buildAlertWatcherValue("admin@example.com", true, "warning"),
		buildAlertWatcherValue("ops@example.com", true, "error"),
	}
	createPlan := arraySmtpPlanWith(t, "smtp.example.com", "example.com", createWatchers)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildArraySmtpType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Update to keep only one watcher.
	updateWatchers := []tftypes.Value{
		buildAlertWatcherValue("admin@example.com", true, "warning"),
	}
	updatePlan := arraySmtpPlanWith(t, "smtp.example.com", "example.com", updateWatchers)
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildArraySmtpType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlan,
		State: createResp.State,
	}, updateResp)

	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update (watcher removal) returned error: %s", updateResp.Diagnostics)
	}

	// Verify only one watcher remains via client.
	watcherList, err := r.client.GetAlertWatchers(context.Background())
	if err != nil {
		t.Fatalf("GetAlertWatchers: %v", err)
	}
	if len(watcherList) != 1 {
		t.Errorf("expected 1 alert watcher after removal, got %d", len(watcherList))
	}
	if watcherList[0].Name != "admin@example.com" {
		t.Errorf("expected remaining watcher=admin@example.com, got %s", watcherList[0].Name)
	}
}

// ---- data source tests ------------------------------------------------------

func newTestArraySmtpDataSource(t *testing.T, ms *testmock.MockServer) *arraySmtpDataSource {
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
	return &arraySmtpDataSource{client: c}
}

func arraySmtpDSSchema(t *testing.T) datasource.SchemaResponse {
	t.Helper()
	d := &arraySmtpDataSource{}
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	return resp
}

func buildArraySmtpDSType() tftypes.Object {
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":              tftypes.String,
		"relay_host":      tftypes.String,
		"sender_domain":   tftypes.String,
		"encryption_mode": tftypes.String,
		"alert_watchers":  tftypes.Set{ElementType: alertWatcherTFType()},
	}}
}

// TestArraySmtpDataSource verifies data source reads current SMTP config and watchers.
func TestArraySmtpDataSource(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterArrayAdminHandlers(ms.Mux)

	// Seed state via client.
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
	relayHost := "smtp.ds.example.com"
	senderDomain := "ds.example.com"
	encMode := "starttls"
	if _, err := c.PatchSmtpServer(context.Background(), client.SmtpServerPatch{
		RelayHost:      &relayHost,
		SenderDomain:   &senderDomain,
		EncryptionMode: &encMode,
	}); err != nil {
		t.Fatalf("PatchSmtpServer: %v", err)
	}
	if _, err := c.PostAlertWatcher(context.Background(), "ds@example.com", client.AlertWatcherPost{
		MinimumNotificationSeverity: "info",
	}); err != nil {
		t.Fatalf("PostAlertWatcher: %v", err)
	}

	d := newTestArraySmtpDataSource(t, ms)
	s := arraySmtpDSSchema(t).Schema

	dsType := buildArraySmtpDSType()
	cfg := map[string]tftypes.Value{
		"id":              tftypes.NewValue(tftypes.String, nil),
		"relay_host":      tftypes.NewValue(tftypes.String, nil),
		"sender_domain":   tftypes.NewValue(tftypes.String, nil),
		"encryption_mode": tftypes.NewValue(tftypes.String, nil),
		"alert_watchers":  tftypes.NewValue(tftypes.Set{ElementType: alertWatcherTFType()}, nil),
	}

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(dsType, nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(dsType, cfg), Schema: s},
	}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("DataSource Read returned error: %s", readResp.Diagnostics)
	}

	var model arraySmtpDataSourceModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected non-empty ID")
	}
	if model.RelayHost.ValueString() != "smtp.ds.example.com" {
		t.Errorf("expected relay_host=smtp.ds.example.com, got %s", model.RelayHost.ValueString())
	}
	if model.EncryptionMode.ValueString() != "starttls" {
		t.Errorf("expected encryption_mode=starttls, got %s", model.EncryptionMode.ValueString())
	}

	var watcherModels []alertWatcherDataSourceModel
	if diags := model.AlertWatchers.ElementsAs(context.Background(), &watcherModels, false); diags.HasError() {
		t.Fatalf("Get alert_watchers: %s", diags)
	}
	if len(watcherModels) != 1 {
		t.Fatalf("expected 1 alert watcher in data source, got %d", len(watcherModels))
	}
	if watcherModels[0].Email.ValueString() != "ds@example.com" {
		t.Errorf("expected watcher email=ds@example.com, got %s", watcherModels[0].Email.ValueString())
	}

	// Ensure the unused attr import is referenced.
	_ = attr.Value(nil)
}
