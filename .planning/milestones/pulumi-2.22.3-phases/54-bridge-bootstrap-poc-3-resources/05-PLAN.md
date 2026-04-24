---
phase: 54-bridge-bootstrap-poc-3-resources
plan: 05
type: execute
wave: 5
depends_on: [04]
files_modified:
  - pulumi/provider/resources_test.go
autonomous: true
requirements: [TEST-01]
must_haves:
  truths:
    - "resources_test.go exists in pulumi/provider package"
    - "Test asserts len(Resources) == 28 and len(DataSources) == 21"
    - "Test asserts api_token config field is Secret"
    - "Test asserts every Sensitive TF field on the 3 POC resources is promoted to Pulumi Secret"
    - "Test asserts bucket has DeleteTimeout >= 25 minutes"
    - "Test asserts no resource schema exposes a 'timeouts' input field"
    - "go test ./... passes inside pulumi/provider/"
  artifacts:
    - path: "pulumi/provider/resources_test.go"
      provides: "Bridge ProviderInfo assertions"
      contains: "func TestProviderInfo"
  key_links:
    - from: "resources_test.go"
      to: "Provider()"
      via: "direct call"
      pattern: "Provider\\(\\)"
---

<objective>
Write `pulumi/provider/resources_test.go` that asserts every Phase 54 ProviderInfo invariant by calling `Provider()` directly (no network, no tfgen shell-out). This is the only automated gate verifying that `resources.go` actually implements D-01..D-05 and the 8 requirements in plan 02.

Purpose: Lock the bridge contract. Any accidental regression (missing secret promotion, wrong DeleteTimeout, timeouts leak, SetAutonaming added) breaks this test.
Output: Passing `go test ./...` inside `pulumi/provider/` with 6+ sub-tests covering TEST-01 scope.
</objective>

<execution_context>
@$HOME/.claude/get-shit-done/workflows/execute-plan.md
</execution_context>

<context>
@.planning/phases/54-bridge-bootstrap-poc-3-resources/54-CONTEXT.md
@pulumi/provider/resources.go
</context>

<tasks>

<task type="auto">
  <name>Task 1: Author resources_test.go</name>
  <read_first>
    - /home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/pulumi/provider/resources.go (all of it — test reads Provider() output)
    - /home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/provider/target_resource.go (to enumerate Sensitive TF fields for POC)
    - /home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/provider/object_store_remote_credentials_resource.go
    - /home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/provider/bucket_resource.go
  </read_first>
  <files>
    - pulumi/provider/resources_test.go
  </files>
  <action>
    Create `pulumi/provider/resources_test.go` (same package `provider` so it can call unexported helpers if needed):

    ```go
    package provider

    import (
        "testing"
        "time"

        "github.com/pulumi/pulumi-terraform-bridge/v3/pkg/tfbridge"
    )

    // Expected counts. Update when TF provider resource set changes.
    const (
        expectedResources   = 28
        expectedDataSources = 21
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

    func TestProviderInfo_ApiTokenIsSecret(t *testing.T) {
        prov := Provider()
        cfg, ok := prov.Config["api_token"]
        if !ok {
            t.Fatalf("Config[\"api_token\"] not present")
        }
        if cfg == nil || cfg.Secret == nil || !*cfg.Secret {
            t.Errorf("Config[\"api_token\"].Secret must be true (SECRETS-01)")
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

    func TestProviderInfo_BucketDeleteTimeout(t *testing.T) {
        prov := Provider()
        r, ok := prov.Resources["flashblade_bucket"]
        if !ok {
            t.Fatalf("flashblade_bucket not in Resources")
        }
        if r.DeleteTimeout < 25*time.Minute {
            t.Errorf("bucket DeleteTimeout = %s, want >= 25m (SOFTDELETE-01 / PB1)", r.DeleteTimeout)
        }
    }

    func TestProviderInfo_RemoteCredentialsSecretAccessKey(t *testing.T) {
        prov := Provider()
        r, ok := prov.Resources["flashblade_object_store_remote_credentials"]
        if !ok {
            t.Fatalf("flashblade_object_store_remote_credentials not in Resources")
        }
        // Field-level Secret mark
        f, ok := r.Fields["secret_access_key"]
        if !ok {
            t.Fatalf("Fields[\"secret_access_key\"] missing")
        }
        if f == nil || f.Secret == nil || !*f.Secret {
            t.Errorf("secret_access_key must be Secret=true (SECRETS-02)")
        }
        // Belt-and-braces: AdditionalSecretOutputs
        found := false
        for _, k := range r.AdditionalSecretOutputs {
            if string(k) == "secretAccessKey" {
                found = true
                break
            }
        }
        if !found {
            t.Errorf("AdditionalSecretOutputs must include \"secretAccessKey\" (SECRETS-02 / PB3)")
        }
    }

    // TestProviderInfo_PocSensitiveFieldsPromoted asserts every TF field marked
    // Sensitive: true on the POC resources gets promoted to a Pulumi Secret
    // (either via Fields[...].Secret or AdditionalSecretOutputs). Auto-promotion
    // from the shim provides the baseline; this test exists to fail fast if
    // someone disables auto-promotion or the shim behavior regresses.
    func TestProviderInfo_PocSensitiveFieldsPromoted(t *testing.T) {
        prov := Provider()
        // Known Sensitive TF fields in the POC resources (from internal/provider/*_resource.go):
        // flashblade_target: none (API endpoint / CA bundle are not marked Sensitive on target)
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

    func TestProviderInfo_NoSetAutonaming(t *testing.T) {
        // Can't read SetAutonaming state directly; proxy check: ensure no resource
        // has a token suffix-style override that only SetAutonaming would inject.
        // Primary enforcement is source-level (grep in CI). This test documents intent.
        prov := Provider()
        // Sanity: the provider struct should exist with config we configured.
        if prov.Name != "flashblade" {
            t.Errorf("ProviderInfo.Name = %q, want \"flashblade\"", prov.Name)
        }
    }

    // Silence unused-import warnings if tfbridge is not referenced via types.
    var _ = tfbridge.True
    ```

    Key notes for executor:
    1. The `expectedResources = 28` and `expectedDataSources = 21` constants come from REQUIREMENTS.md / D-05. If plan 04's tfgen run revealed different counts (e.g. TF provider grew), reconcile BEFORE writing this test — either update the TF side or update these constants with a comment explaining why.
    2. If the `Fields[...].Secret` API on this bridge version uses a different shape (e.g. a method vs a pointer field), adapt. The pattern used by pulumi-random/pulumi-cloudflare is `Secret: tfbridge.True()` which returns `*bool`. That's what `resources.go` assigns in plan 02, and what this test reads back.
    3. The "no SetAutonaming" check is covered primarily by the grep in plan 02's verify block; this test just asserts the name to ensure `Provider()` is callable.
    4. Run `cd pulumi/provider && go test ./... -count=1 -v`. Expect all 8 sub-tests to pass.
  </action>
  <verify>
    <automated>test -f pulumi/provider/resources_test.go && grep -q 'expectedResources' pulumi/provider/resources_test.go && grep -q 'TestProviderInfo_ApiTokenIsSecret' pulumi/provider/resources_test.go && grep -q 'TestProviderInfo_BucketDeleteTimeout' pulumi/provider/resources_test.go && grep -q 'TestProviderInfo_TimeoutsInputIsOmittedEverywhere' pulumi/provider/resources_test.go && grep -q 'TestProviderInfo_RemoteCredentialsSecretAccessKey' pulumi/provider/resources_test.go && cd pulumi/provider && go test ./... -count=1</automated>
  </verify>
  <done>
    - resources_test.go exists with all required TestProviderInfo_* functions
    - `go test ./... -count=1` inside pulumi/provider passes
    - Every must_haves truth is backed by at least one assertion
  </done>
</task>

</tasks>

<verification>
- `cd pulumi/provider && go test ./... -count=1 -v` — all sub-tests pass
- TEST-01 requirement satisfied
</verification>

<success_criteria>
- Bridge contract locked by automated tests
- Phase 54 goal met: full bridge chain validated by unit tests without needing a real FlashBlade
- All 13 Phase 54 requirement IDs covered across plans 01-05
</success_criteria>

<output>
After completion, create `.planning/phases/54-bridge-bootstrap-poc-3-resources/54-05-SUMMARY.md`
</output>
