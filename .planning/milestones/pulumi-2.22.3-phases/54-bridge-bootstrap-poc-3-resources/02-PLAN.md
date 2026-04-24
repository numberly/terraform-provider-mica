---
phase: 54-bridge-bootstrap-poc-3-resources
plan: 02
type: execute
wave: 2
depends_on: [01]
files_modified:
  - pulumi/provider/resources.go
autonomous: true
requirements: [BRIDGE-05, MAPPING-02, MAPPING-03, MAPPING-05, COMPOSITE-01, SECRETS-01, SECRETS-02, SOFTDELETE-01]
must_haves:
  truths:
    - "ProviderInfo exposes all 7 TF provider config keys (endpoint, api_token, oauth2_client_id, oauth2_client_secret, oauth2_token_url, skip_tls_verify, ca_certificate) — D-01"
    - "api_token in ProviderInfo.Config is Secret"
    - "ProviderInfo.PluginDownloadURL is set to github://api.github.com/numberly — BRIDGE-05"
    - "MustComputeTokens + KnownModules + MustApplyAutoAliases applied to all 49 TF resources/DS — D-02"
    - "No SetAutonaming call present — MAPPING-05"
    - "timeouts input field is omitted on every resource — MAPPING-02"
    - "bucket and filesystem resources have DeleteTimeout 30min — SOFTDELETE-01"
    - "target, remote_credentials, bucket have explicit Create/Update/DeleteTimeout per TF defaults — MAPPING-03"
    - "object_store_remote_credentials has secret_access_key marked Secret + AdditionalSecretOutputs — SECRETS-02"
    - "object_store_access_policy_rule has ComputeID producing policyName + \"/\" + ruleName — COMPOSITE-01"
  artifacts:
    - path: "pulumi/provider/resources.go"
      provides: "tfbridge.ProviderInfo with all overrides"
      contains: "tfbridge.ProviderInfo"
  key_links:
    - from: "pulumi/provider/resources.go"
      to: "internal/provider.New(version.Version)"
      via: "pf.ShimProvider(fb.New(version.Version)())"
      pattern: "pf\\.ShimProvider"
    - from: "pulumi/provider/resources.go"
      to: "pulumi/provider/pkg/version"
      via: "Go import"
      pattern: "provider/pkg/version"
---

<objective>
Populate `pulumi/provider/resources.go` with the full `tfbridge.ProviderInfo` definition: config surface mirroring TF 1:1 (D-01), auto-tokenization for all 49 resources+DS (D-02), `omitTimeoutsOnAll` helper (MAPPING-02), POC overrides (target timeouts, remote_credentials secret, bucket soft-delete timeout, policy-rule ComputeID), and a `Provider()` accessor used by both binaries.

Purpose: This file is the single source of truth for bridge behavior. Every POC invariant (secrets, timeouts, composite IDs) lives here, verified by resources_test.go in plan 05.
Output: `go build ./...` inside `pulumi/provider/` compiles successfully once plan 03 adds the cmd packages (this task alone may fail to build because `Provider()` is not yet consumed — that's fine; plan 03 closes the loop).
</objective>

<execution_context>
@$HOME/.claude/get-shit-done/workflows/execute-plan.md
</execution_context>

<context>
@.planning/phases/54-bridge-bootstrap-poc-3-resources/54-CONTEXT.md
@.planning/research/STACK.md
@.planning/research/PITFALLS.md
@internal/provider/provider.go
@internal/provider/object_store_access_policy_rule_resource.go
</context>

<tasks>

<task type="auto">
  <name>Task 1: Author pulumi/provider/resources.go</name>
  <read_first>
    - /home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/provider/provider.go (lines 1-80, confirm `New(version string) func() provider.Provider` signature)
    - /home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/provider/object_store_access_policy_rule_resource.go (lines 361-387 — confirmed composite ID format `policy_name/rule_name`)
    - /home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/.planning/phases/54-bridge-bootstrap-poc-3-resources/54-CONTEXT.md (D-01..D-05)
  </read_first>
  <files>
    - pulumi/provider/resources.go
  </files>
  <action>
    Create `pulumi/provider/resources.go` with the following structure:

    ```go
    package provider

    import (
        "context"
        "fmt"
        "time"

        "github.com/pulumi/pulumi-terraform-bridge/v3/pkg/pf/tfbridge"
        shimpf "github.com/pulumi/pulumi-terraform-bridge/v3/pkg/pf"
        "github.com/pulumi/pulumi-terraform-bridge/v3/pkg/tfbridge/tokens"
        "github.com/pulumi/pulumi/sdk/v3/go/common/resource"

        fb "github.com/numberly/opentofu-provider-flashblade/internal/provider"
        "github.com/numberly/opentofu-provider-flashblade/pulumi/provider/pkg/version"
    )

    // mainPkg is the package prefix used by tfgen for generated SDKs.
    const mainPkg = "flashblade"

    // Provider builds the full tfbridge.ProviderInfo with ShimProvider + all overrides.
    // Consumed by both cmd/pulumi-tfgen-flashblade (build time) and
    // cmd/pulumi-resource-flashblade (runtime).
    func Provider() tfbridge.ProviderInfo {
        prov := tfbridge.ProviderInfo{
            Name:    "flashblade",
            Version: version.Version,
            P:       shimpf.ShimProvider(fb.New(version.Version)()),

            DisplayName: "FlashBlade",
            Publisher:   "numberly",
            Description: "A Pulumi package for managing Pure Storage FlashBlade resources.",
            Keywords:    []string{"pulumi", "flashblade", "pure-storage", "category/infrastructure"},
            License:     "Apache-2.0",
            Homepage:    "https://github.com/numberly/opentofu-provider-flashblade",
            Repository:  "https://github.com/numberly/opentofu-provider-flashblade",
            // PluginDownloadURL — required for `pulumi plugin install --server
            // github://api.github.com/numberly ...` to resolve the runtime binary
            // from GitHub Releases (BRIDGE-05).
            PluginDownloadURL: "github://api.github.com/numberly",

            // Config — mirrors the TF provider schema 1:1 (D-01).
            Config: map[string]*tfbridge.SchemaInfo{
                "api_token": {
                    Secret: tfbridge.True(),
                },
                "oauth2_client_secret": {
                    Secret: tfbridge.True(),
                },
                "ca_certificate": {
                    Secret: tfbridge.True(),
                },
            },

            MetadataInfo: tfbridge.NewProviderMetadata(nil),
        }

        // ---- Auto-tokenization across all 49 TF resources + data sources (D-02) ----
        // Module assignment driven by resource name prefix.
        prov.MustComputeTokens(tokens.KnownModules(
            "flashblade_", // TF resource prefix
            "index",       // default module
            []string{
                "bucket",
                "filesystem",
                "policy",
                "objectstore",
                "array",
                "network",
            },
            tokens.MakeStandard(mainPkg),
        ))

        // ---- Shared helper: omit the `timeouts` input block on every resource (MAPPING-02) ----
        // The `timeouts` block is a TF ergonomic; Pulumi expresses these via CustomTimeouts.
        //
        // ORDERING NOTE: `omitTimeoutsOnAll` runs AFTER `MustComputeTokens` populates
        // `prov.Resources`. Iterating an empty map before tokenization would be a no-op.
        // REQUIREMENTS.md MAPPING-02 was amended in commit ebc4eb8 to match this ordering,
        // and CONTEXT.md "Claude's Discretion" note about ordering is resolved by the same
        // rule: MustComputeTokens first, then omitTimeoutsOnAll, then per-resource overrides.
        // Do NOT reorder based on the original (now-amended) wording.
        omitTimeoutsOnAll(&prov)

        // ---- POC overrides (D-05) ----

        // flashblade_target — auto-tokenization baseline; explicit TF-default timeouts (MAPPING-03).
        if r, ok := prov.Resources["flashblade_target"]; ok {
            r.CreateTimeout = 20 * time.Minute
            r.UpdateTimeout = 20 * time.Minute
            r.DeleteTimeout = 20 * time.Minute
        } else {
            panic("flashblade_target resource not found after MustComputeTokens")
        }

        // flashblade_object_store_remote_credentials — write-once secret (SECRETS-02, PB3).
        if r, ok := prov.Resources["flashblade_object_store_remote_credentials"]; ok {
            if r.Fields == nil {
                r.Fields = map[string]*tfbridge.SchemaInfo{}
            }
            r.Fields["secret_access_key"] = &tfbridge.SchemaInfo{
                Secret: tfbridge.True(),
            }
            r.AdditionalSecretOutputs = append(r.AdditionalSecretOutputs, resource.PropertyKey("secretAccessKey"))
            r.CreateTimeout = 20 * time.Minute
            r.UpdateTimeout = 20 * time.Minute
            r.DeleteTimeout = 20 * time.Minute
        } else {
            panic("flashblade_object_store_remote_credentials resource not found after MustComputeTokens")
        }

        // flashblade_bucket — soft-delete + eradication polls, needs 30-min Delete (SOFTDELETE-01, PB1).
        if r, ok := prov.Resources["flashblade_bucket"]; ok {
            r.CreateTimeout = 20 * time.Minute
            r.UpdateTimeout = 20 * time.Minute
            r.DeleteTimeout = 30 * time.Minute
        } else {
            panic("flashblade_bucket resource not found after MustComputeTokens")
        }

        // flashblade_filesystem — same soft-delete pattern as bucket (SOFTDELETE-01 pre-applied
        // here per D-02 scope; filesystem resource overrides otherwise deferred to Phase 55).
        if r, ok := prov.Resources["flashblade_filesystem"]; ok {
            r.DeleteTimeout = 30 * time.Minute
        }

        // flashblade_object_store_access_policy_rule — composite ID with "/" separator,
        // string rule name (COMPOSITE-01, verified against
        // internal/provider/object_store_access_policy_rule_resource.go:361-387).
        if r, ok := prov.Resources["flashblade_object_store_access_policy_rule"]; ok {
            r.ComputeID = func(
                ctx context.Context,
                state resource.PropertyMap,
            ) (resource.ID, error) {
                policyName, ok1 := state["policyName"]
                ruleName, ok2 := state["name"]
                if !ok1 || !ok2 {
                    return "", fmt.Errorf(
                        "object_store_access_policy_rule: missing policyName or name in state (got keys %v)",
                        mapKeys(state),
                    )
                }
                ps, psOk := policyName.V.(string)
                rs, rsOk := ruleName.V.(string)
                if !psOk || !rsOk {
                    return "", fmt.Errorf(
                        "object_store_access_policy_rule: policyName and name must be strings",
                    )
                }
                return resource.ID(ps + "/" + rs), nil
            }
        } else {
            panic("flashblade_object_store_access_policy_rule resource not found after MustComputeTokens")
        }

        // ---- MAPPING-05: Do NOT call prov.SetAutonaming. FlashBlade names are operational. ----

        // Apply auto-aliases (D-02).
        prov.MustApplyAutoAliases()

        return prov
    }

    // omitTimeoutsOnAll hides the TF `timeouts` input block from the generated Pulumi schema
    // across every resource (MAPPING-02 / PB7). Must run AFTER MustComputeTokens (so
    // prov.Resources is populated) and BEFORE per-resource overrides that touch Fields.
    func omitTimeoutsOnAll(prov *tfbridge.ProviderInfo) {
        for _, r := range prov.Resources {
            if r == nil {
                continue
            }
            if r.Fields == nil {
                r.Fields = map[string]*tfbridge.SchemaInfo{}
            }
            r.Fields["timeouts"] = &tfbridge.SchemaInfo{Omit: true}
        }
    }

    func mapKeys(m resource.PropertyMap) []string {
        out := make([]string, 0, len(m))
        for k := range m {
            out = append(out, string(k))
        }
        return out
    }
    ```

    Critical requirements (executor must verify each):
    1. Import path for internal provider: `github.com/numberly/opentofu-provider-flashblade/internal/provider` aliased as `fb`.
    2. `pf.ShimProvider(fb.New(version.Version)())` — note double-invocation (`New(...)()` returns `provider.Provider`).
    3. MustComputeTokens called BEFORE per-resource overrides (overrides read `prov.Resources[...]` which MustComputeTokens populates).
    4. `omitTimeoutsOnAll` runs AFTER MustComputeTokens (iterating an empty map before tokenization is a no-op) and BEFORE POC overrides. REQUIREMENTS.md MAPPING-02 was amended in commit ebc4eb8 to match this ordering; do not second-guess based on the original spec wording. The same clarification resolves the CONTEXT.md "Claude's Discretion" note about placement.
    5. `PluginDownloadURL: "github://api.github.com/numberly"` is REQUIRED (BRIDGE-05) — without it `pulumi plugin install --server github://api.github.com/numberly ...` cannot resolve the runtime binary.
    6. ComputeID uses `"/"` separator, reads `policyName` and `name` (Pulumi camelCase of TF `policy_name` and `name`).
    7. No `SetAutonaming` call anywhere.
    8. `api_token` is the config secret under Config key — NOT under any other key.
    9. MetadataInfo initialized via `tfbridge.NewProviderMetadata(nil)` — runtime embedded bytes plumbed in plan 03's main.go.

    After writing the file, run `cd pulumi/provider && go mod tidy` to populate go.sum. Expect "no Go files" errors from cmd/ subdirs (plan 03 adds those) but the root package should tidy cleanly once this file is in place.
  </action>
  <verify>
    <automated>test -f pulumi/provider/resources.go && grep -q 'pf\.ShimProvider\|shimpf\.ShimProvider' pulumi/provider/resources.go && grep -q 'MustComputeTokens' pulumi/provider/resources.go && grep -q 'MustApplyAutoAliases' pulumi/provider/resources.go && grep -q 'omitTimeoutsOnAll' pulumi/provider/resources.go && grep -q 'PluginDownloadURL' pulumi/provider/resources.go && grep -q '"api_token"' pulumi/provider/resources.go && grep -q 'flashblade_bucket' pulumi/provider/resources.go && grep -q '30 \* time.Minute' pulumi/provider/resources.go && grep -q 'AdditionalSecretOutputs' pulumi/provider/resources.go && grep -q 'secretAccessKey' pulumi/provider/resources.go && grep -q 'ps + "/" + rs\|"/"' pulumi/provider/resources.go && ! grep -q 'SetAutonaming' pulumi/provider/resources.go && cd pulumi/provider && go mod tidy 2>&1 | grep -v "no Go files in .*/cmd/" | { ! grep -q "error"; }</automated>
  </verify>
  <done>
    - `pulumi/provider/resources.go` defines `func Provider() tfbridge.ProviderInfo`
    - All 7 config keys referenced (3 Secret, 4 plain) — D-01
    - `PluginDownloadURL` set to `github://api.github.com/numberly` — BRIDGE-05
    - MustComputeTokens called with all 6 KnownModules + `index` default — D-02
    - `omitTimeoutsOnAll` helper present and invoked AFTER MustComputeTokens
    - POC resource overrides in place for target, remote_credentials, bucket, policy-rule
    - No `SetAutonaming` anywhere in file
    - `go mod tidy` completes without errors for the provider package (cmd/ subdirs may warn)
  </done>
</task>

</tasks>

<verification>
- File exists
- `go vet ./...` from pulumi/provider clean for the root package (cmd/ not yet present — skip via `go vet .`)
- All grep assertions pass (see automated block)
</verification>

<success_criteria>
- resources.go compiles standalone (`cd pulumi/provider && go build .`)
- All 8 requirements covered: BRIDGE-05 (ProviderInfo wiring + PluginDownloadURL), MAPPING-02 (timeouts omit), MAPPING-03 (explicit timeouts), MAPPING-05 (no autonaming), COMPOSITE-01 (policy rule `/` separator), SECRETS-01 (api_token secret), SECRETS-02 (remote_credentials), SOFTDELETE-01 (bucket 30min delete)
- Every locked decision D-01, D-02, D-05 has a corresponding line in the file
</success_criteria>

<output>
After completion, create `.planning/phases/54-bridge-bootstrap-poc-3-resources/54-02-SUMMARY.md`
</output>
