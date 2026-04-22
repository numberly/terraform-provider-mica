package provider

import (
	"context"
	"testing"

	tfbridge "github.com/pulumi/pulumi-terraform-bridge/v3/pkg/tfbridge"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
)

// Expected counts. Matches TF provider registrations (54 resources, 40 data sources).
// Update when TF provider resource set changes.
const (
	expectedResources   = 54
	expectedDataSources = 40
)

// POC resources under test (D-05).
var pocResources = []string{
	"flashblade_target",
	"flashblade_object_store_remote_credentials",
	"flashblade_bucket",
	"flashblade_object_store_access_policy_rule",
}

func TestProviderInfo_ResourceAndDataSourceCounts(t *testing.T) {
	prov := Provider()
	if got := len(prov.Resources); got != expectedResources {
		t.Errorf("Resources count = %d, want %d", got, expectedResources)
	}
	if got := len(prov.DataSources); got != expectedDataSources {
		t.Errorf("DataSources count = %d, want %d", got, expectedDataSources)
	}
}

// TestProviderInfo_ConfigHasNoManualOverrides verifies that Config is empty.
// The TF provider uses a nested auth.* block; Sensitive fields are auto-promoted
// by the bridge from TF schema introspection. No manual Config overrides are needed
// (SECRETS-01). If this changes, the test must be updated.
func TestProviderInfo_ConfigHasNoManualOverrides(t *testing.T) {
	prov := Provider()
	if len(prov.Config) != 0 {
		t.Errorf("Config should be empty (nested auth.* handled by shim auto-promotion), got %d keys", len(prov.Config))
	}
}

func TestProviderInfo_TimeoutsInputIsOmittedEverywhere(t *testing.T) {
	prov := Provider()
	for name, r := range prov.Resources {
		if r == nil {
			continue
		}
		f, ok := r.Fields["timeouts"]
		if !ok {
			t.Errorf("resource %q: Fields[\"timeouts\"] missing — omitTimeoutsOnAll did not run (MAPPING-02)", name)
			continue
		}
		if f == nil || !f.Omit {
			t.Errorf("resource %q: timeouts field must have Omit=true (MAPPING-02)", name)
		}
	}
}

// TestProviderInfo_BucketSoftDeleteRegistered verifies that flashblade_bucket is registered
// (SOFTDELETE-01). DeleteTimeout does not exist on ResourceInfo in bridge v3.127.0 —
// the TF provider's timeouts block default (Delete 30m) is inherited via the shim.
// Pulumi users needing extended timeouts should use customTimeouts in their program.
func TestProviderInfo_BucketSoftDeleteRegistered(t *testing.T) {
	prov := Provider()
	_, ok := prov.Resources["flashblade_bucket"]
	if !ok {
		t.Fatalf("flashblade_bucket not in Resources (SOFTDELETE-01)")
	}
}

// TestProviderInfo_FileSystemSoftDeleteRegistered verifies flashblade_file_system is registered.
func TestProviderInfo_FileSystemSoftDeleteRegistered(t *testing.T) {
	prov := Provider()
	_, ok := prov.Resources["flashblade_file_system"]
	if !ok {
		t.Fatalf("flashblade_file_system not in Resources (SOFTDELETE-01)")
	}
}

func TestProviderInfo_RemoteCredentialsSecretAccessKey(t *testing.T) {
	prov := Provider()
	r, ok := prov.Resources["flashblade_object_store_remote_credentials"]
	if !ok {
		t.Fatalf("flashblade_object_store_remote_credentials not in Resources")
	}
	// Field-level Secret mark (SECRETS-02, PB3).
	// AdditionalSecretOutputs does not exist on ResourceInfo in bridge v3.127.0;
	// TF Sensitive=true auto-promotion is the runtime defense.
	f, ok := r.Fields["secret_access_key"]
	if !ok {
		t.Fatalf("Fields[\"secret_access_key\"] missing")
	}
	if f == nil || f.Secret == nil || !*f.Secret {
		t.Errorf("secret_access_key must be Secret=true (SECRETS-02)")
	}
}

// TestProviderInfo_PocSensitiveFieldsPromoted asserts every TF field marked
// Sensitive: true on the POC resources gets promoted to a Pulumi Secret
// via Fields[...].Secret. Auto-promotion from the shim provides the baseline;
// this test exists to fail fast if auto-promotion regresses or someone removes
// the explicit Secret override added in resources.go (SECRETS-02).
func TestProviderInfo_PocSensitiveFieldsPromoted(t *testing.T) {
	prov := Provider()
	// Known Sensitive TF fields in the POC resources (from internal/provider/*_resource.go):
	// flashblade_target: none (endpoint / CA bundle are not marked Sensitive in TF schema)
	// flashblade_object_store_remote_credentials: secret_access_key
	// flashblade_bucket: none
	// flashblade_object_store_access_policy_rule: none
	// Adjust this map if the TF provider adds Sensitive fields.
	expectedSecrets := map[string][]string{
		"flashblade_object_store_remote_credentials": {"secret_access_key"},
	}
	for resName, fields := range expectedSecrets {
		r, ok := prov.Resources[resName]
		if !ok {
			t.Errorf("resource %q not in Resources", resName)
			continue
		}
		for _, f := range fields {
			info, ok := r.Fields[f]
			if !ok || info == nil || info.Secret == nil || !*info.Secret {
				t.Errorf("resource %q field %q must be Secret=true", resName, f)
			}
		}
	}
}

func TestProviderInfo_PolicyRuleComputeIDPresent(t *testing.T) {
	prov := Provider()
	r, ok := prov.Resources["flashblade_object_store_access_policy_rule"]
	if !ok {
		t.Fatalf("flashblade_object_store_access_policy_rule not in Resources")
	}
	if r.ComputeID == nil {
		t.Errorf("flashblade_object_store_access_policy_rule.ComputeID must be set (COMPOSITE-01)")
	}
}

// TestProviderInfo_ArrayConnectionKeySensitiveIDFalse verifies that the sensitive ID
// override on flashblade_array_connection_key uses tfbridge.False() (not True()).
// The bridge requires explicit acknowledgment that a sensitive ID will be exposed in state.
func TestProviderInfo_ArrayConnectionKeySensitiveIDFalse(t *testing.T) {
	prov := Provider()
	r, ok := prov.Resources["flashblade_array_connection_key"]
	if !ok {
		t.Fatalf("flashblade_array_connection_key not in Resources")
	}
	f, ok := r.Fields["id"]
	if !ok {
		t.Fatalf("Fields[\"id\"] missing on flashblade_array_connection_key")
	}
	if f == nil || f.Secret == nil || *f.Secret {
		t.Errorf("flashblade_array_connection_key.id must have Secret=false (tfbridge.False()) — not true, not nil")
	}
}

func TestProviderInfo_NoSetAutonaming(t *testing.T) {
	// FlashBlade names are operational identifiers — no random suffix (MAPPING-05).
	// SetAutonaming was deliberately omitted. Enforcement is source-level.
	// This test documents intent by asserting the provider Name is correct.
	prov := Provider()
	if prov.Name != "flashblade" {
		t.Errorf("ProviderInfo.Name = %q, want \"flashblade\"", prov.Name)
	}
}

func TestProviderInfo_AllPocResourcesPresent(t *testing.T) {
	prov := Provider()
	for _, name := range pocResources {
		if _, ok := prov.Resources[name]; !ok {
			t.Errorf("POC resource %q not in Resources (D-05)", name)
		}
	}
}

// ---- SECRETS-03: All sensitive fields promoted ----

// TestProviderInfo_AllSensitiveFieldsPromoted verifies every TF field with Sensitive:true
// has an explicit Secret:tfbridge.True() mark in resources.go (SECRETS-03).
// Exception: flashblade_array_connection_key.id uses False() — acknowledged ID exposure.
func TestProviderInfo_AllSensitiveFieldsPromoted(t *testing.T) {
	prov := Provider()
	sensitiveFields := map[string][]string{
		"flashblade_array_connection":                {"connection_key"},
		"flashblade_array_connection_key":            {"connection_key"},
		"flashblade_certificate":                     {"passphrase", "private_key"},
		"flashblade_object_store_access_key":         {"secret_access_key"},
		"flashblade_object_store_remote_credentials": {"secret_access_key"},
		"flashblade_directory_service_management":    {"bind_password"},
	}
	for resName, fields := range sensitiveFields {
		r, ok := prov.Resources[resName]
		if !ok {
			t.Errorf("resource %q not in Resources", resName)
			continue
		}
		for _, f := range fields {
			info, ok := r.Fields[f]
			if !ok || info == nil || info.Secret == nil || !*info.Secret {
				t.Errorf("resource %q field %q must have Secret=true (SECRETS-03)", resName, f)
			}
		}
	}
}

// TestProviderInfo_ArrayConnectionKeySecrets verifies the dual-override on array_connection_key:
// id must be False() (not True()) and connection_key must be True().
func TestProviderInfo_ArrayConnectionKeySecrets(t *testing.T) {
	prov := Provider()
	r, ok := prov.Resources["flashblade_array_connection_key"]
	if !ok {
		t.Fatalf("flashblade_array_connection_key not in Resources")
	}
	idInfo, ok := r.Fields["id"]
	if !ok || idInfo == nil || idInfo.Secret == nil || *idInfo.Secret {
		t.Errorf("flashblade_array_connection_key.id must have Secret=false")
	}
	ckInfo, ok := r.Fields["connection_key"]
	if !ok || ckInfo == nil || ckInfo.Secret == nil || !*ckInfo.Secret {
		t.Errorf("flashblade_array_connection_key.connection_key must have Secret=true")
	}
}

// ---- SOFTDELETE-03: Soft-delete resources registered ----

// TestProviderInfo_SoftDeleteResourcesRegistered verifies both soft-delete resources are registered.
// Bridge v3.127.0 has no DeleteTimeout on ResourceInfo; TF timeouts defaults (30m) are inherited via shim.
func TestProviderInfo_SoftDeleteResourcesRegistered(t *testing.T) {
	prov := Provider()
	softDeleteResources := []string{
		"flashblade_bucket",
		"flashblade_file_system",
	}
	for _, name := range softDeleteResources {
		if _, ok := prov.Resources[name]; !ok {
			t.Errorf("soft-delete resource %q not in Resources (SOFTDELETE-03)", name)
		}
	}
}

// ---- COMPOSITE-02/03/04: ComputeID closures ----

// TestProviderInfo_AllCompositeIDsPresent verifies all 4 composite-ID resources have ComputeID set.
func TestProviderInfo_AllCompositeIDsPresent(t *testing.T) {
	prov := Provider()
	compositeResources := []string{
		"flashblade_object_store_access_policy_rule",
		"flashblade_bucket_access_policy_rule",
		"flashblade_network_access_policy_rule",
		"flashblade_management_access_policy_directory_service_role_membership",
	}
	for _, name := range compositeResources {
		r, ok := prov.Resources[name]
		if !ok {
			t.Errorf("composite-ID resource %q not in Resources", name)
			continue
		}
		if r.ComputeID == nil {
			t.Errorf("composite-ID resource %q must have ComputeID set", name)
		}
	}
}

// TestProviderInfo_ComputeID_BucketAccessPolicyRule invokes the COMPOSITE-02 closure
// with sample PropertyMap data and asserts the returned ID (COMPOSITE-02).
func TestProviderInfo_ComputeID_BucketAccessPolicyRule(t *testing.T) {
	prov := Provider()
	r := prov.Resources["flashblade_bucket_access_policy_rule"]
	if r == nil || r.ComputeID == nil {
		t.Fatalf("ComputeID not set on flashblade_bucket_access_policy_rule")
	}
	state := resource.PropertyMap{
		"bucketName": resource.NewStringProperty("my-bucket"),
		"name":       resource.NewStringProperty("rule1"),
	}
	id, err := r.ComputeID(context.Background(), state)
	if err != nil {
		t.Fatalf("ComputeID error: %v", err)
	}
	if string(id) != "my-bucket/rule1" {
		t.Errorf("expected 'my-bucket/rule1', got %q", id)
	}
}

// TestProviderInfo_ComputeID_NetworkAccessPolicyRule invokes the COMPOSITE-03 closure
// with sample PropertyMap data and asserts the returned ID (COMPOSITE-03).
func TestProviderInfo_ComputeID_NetworkAccessPolicyRule(t *testing.T) {
	prov := Provider()
	r := prov.Resources["flashblade_network_access_policy_rule"]
	if r == nil || r.ComputeID == nil {
		t.Fatalf("ComputeID not set on flashblade_network_access_policy_rule")
	}
	state := resource.PropertyMap{
		"policyName": resource.NewStringProperty("nap-policy"),
		"name":       resource.NewStringProperty("rule2"),
	}
	id, err := r.ComputeID(context.Background(), state)
	if err != nil {
		t.Fatalf("ComputeID error: %v", err)
	}
	if string(id) != "nap-policy/rule2" {
		t.Errorf("expected 'nap-policy/rule2', got %q", id)
	}
}

// TestProviderInfo_ComputeID_ManagementAccessPolicyDSRMembership invokes the COMPOSITE-04
// closure including the colon edge case for built-in policy names like pure:policy/array_admin (COMPOSITE-04).
func TestProviderInfo_ComputeID_ManagementAccessPolicyDSRMembership(t *testing.T) {
	prov := Provider()
	r := prov.Resources["flashblade_management_access_policy_directory_service_role_membership"]
	if r == nil || r.ComputeID == nil {
		t.Fatalf("ComputeID not set on flashblade_management_access_policy_directory_service_role_membership")
	}

	// Normal case.
	state := resource.PropertyMap{
		"role":   resource.NewStringProperty("ops-admin"),
		"policy": resource.NewStringProperty("custom-policy"),
	}
	id, err := r.ComputeID(context.Background(), state)
	if err != nil {
		t.Fatalf("ComputeID error: %v", err)
	}
	if string(id) != "ops-admin/custom-policy" {
		t.Errorf("expected 'ops-admin/custom-policy', got %q", id)
	}

	// COMPOSITE-04: colon-containing built-in policy name.
	stateColon := resource.PropertyMap{
		"role":   resource.NewStringProperty("array-admin-role"),
		"policy": resource.NewStringProperty("pure:policy/array_admin"),
	}
	idColon, err := r.ComputeID(context.Background(), stateColon)
	if err != nil {
		t.Fatalf("ComputeID error with colon policy: %v", err)
	}
	if string(idColon) != "array-admin-role/pure:policy/array_admin" {
		t.Errorf("expected 'array-admin-role/pure:policy/array_admin', got %q", idColon)
	}
}

// ---- UPGRADE-01/02/03: State upgrader resource registration ----

// TestProviderInfo_StateUpgraderResourcesRegistered verifies the 3 resources with TF state
// upgraders are registered in the bridge (UPGRADE-01/02/03). Bridge delegates state upgrades
// to the TF provider via the shim. Full pulumi refresh smoke tests are deferred to Phase 58.
func TestProviderInfo_StateUpgraderResourcesRegistered(t *testing.T) {
	prov := Provider()
	upgraderResources := []string{
		"flashblade_server",                           // v0→v1→v2 (UPGRADE-01)
		"flashblade_directory_service_role",            // v0→v1 (UPGRADE-02)
		"flashblade_object_store_remote_credentials",   // v0→v1 (UPGRADE-03)
	}
	for _, name := range upgraderResources {
		if _, ok := prov.Resources[name]; !ok {
			t.Errorf("state-upgrader resource %q not in Resources (UPGRADE)", name)
		}
	}
}

// Silence unused-import warning if tfbridge types are not referenced directly.
var _ = tfbridge.True
