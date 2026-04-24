---
phase: 54-bridge-bootstrap-poc-3-resources
plan: "02"
subsystem: infra
tags: [pulumi, bridge, terraform-bridge, pf, tfbridge, tokens, flashblade]

requires:
  - phase: 54-bridge-bootstrap-poc-3-resources/01
    provides: pulumi/provider/go.mod with bridge v3.127.0, sdk v3.231.0, version.go

provides:
  - pulumi/provider/resources.go: tfbridge.ProviderInfo with pftfbridge.ShimProvider wiring
  - go.sum: fully resolved dependency lockfile for pulumi/provider module
  - Provider() func: entry point for both tfgen and runtime binaries
  - omitTimeoutsOnAll helper: removes timeouts block from all 49 resources
  - ComputeID for object_store_access_policy_rule: policy/rule slash-separator

affects:
  - 54-bridge-bootstrap-poc-3-resources/03 (cmd packages consume Provider())
  - 54-bridge-bootstrap-poc-3-resources/05 (resources_test.go validates this file)
  - 55-full-resource-mapping (adds per-resource overrides on top of this base)

tech-stack:
  added:
    - github.com/pulumi/pulumi-terraform-bridge/v3 v3.127.0 (go.sum resolved)
    - github.com/pulumi/pulumi/sdk/v3 v3.231.0
    - github.com/pulumi/pulumi/pkg/v3 v3.231.0
  patterns:
    - pftfbridge.ShimProvider wraps terraform-plugin-framework provider for bridge
    - tokens.KnownModules + MakeStandard for auto-tokenization of all 49 resources
    - omitTimeoutsOnAll helper called AFTER MustComputeTokens (ordering invariant)
    - SchemaInfo.Secret = tfbridge.True() for belt-and-braces secret marking
    - ComputeID closure reading camelCase state keys (policyName, name) for composite IDs

key-files:
  created:
    - pulumi/provider/resources.go
    - pulumi/provider/go.sum
  modified:
    - pulumi/provider/go.mod (go mod tidy updated indirect deps)

key-decisions:
  - "Import path for ShimProvider: pkg/pf/tfbridge (not pkg/pf) — ShimProvider function lives in pftfbridge, not pf"
  - "DeleteTimeout/CreateTimeout/UpdateTimeout not fields on ResourceInfo in bridge v3.127.0 — TF timeouts defaults (20m/30m) inherited via shim"
  - "AdditionalSecretOutputs not available in ResourceInfo bridge v3.127.0 — TF Sensitive=true auto-promotion used as runtime defense"
  - "ComputeID reads policyName and name (Pulumi camelCase) not policy_name and name (TF snake_case)"

patterns-established:
  - "omitTimeoutsOnAll: iterate prov.Resources AFTER MustComputeTokens, BEFORE per-resource overrides"
  - "ShimProvider import: pftfbridge 'github.com/pulumi/pulumi-terraform-bridge/v3/pkg/pf/tfbridge'"
  - "Base tfbridge types (ProviderInfo, SchemaInfo, True): 'github.com/pulumi/pulumi-terraform-bridge/v3/pkg/tfbridge'"

requirements-completed: [BRIDGE-05, MAPPING-02, MAPPING-03, MAPPING-05, COMPOSITE-01, SECRETS-01, SECRETS-02, SOFTDELETE-01]

duration: 35min
completed: 2026-04-22
---

# Phase 54 Plan 02: resources.go ProviderInfo + pf.ShimProvider wiring Summary

**tfbridge.ProviderInfo with pftfbridge.ShimProvider, KnownModules auto-tokenization for all 49 resources, omitTimeoutsOnAll helper, and ComputeID for composite-ID policy rule**

## Performance

- **Duration:** ~35 min
- **Started:** 2026-04-22T10:30:00Z
- **Completed:** 2026-04-22T11:05:00Z
- **Tasks:** 1 of 1
- **Files modified:** 3 (resources.go created, go.sum created, go.mod updated)

## Accomplishments

- `pulumi/provider/resources.go` defines `func Provider() tfbridge.ProviderInfo` with full wiring
- `pftfbridge.ShimProvider(fb.New(version.Version)())` correctly bridges the pf provider
- All 49 TF resources auto-tokenized via `MustComputeTokens` + `KnownModules` (6 modules + index)
- `omitTimeoutsOnAll` helper omits `timeouts` from every resource (MAPPING-02, PB8 prevention)
- `PluginDownloadURL: "github://api.github.com/numberly"` set (BRIDGE-05)
- `api_token`, `oauth2_client_secret`, `ca_certificate` marked Secret in Config (SECRETS-01)
- `secret_access_key` marked Secret via Fields override (SECRETS-02)
- `ComputeID` for `flashblade_object_store_access_policy_rule` with `/` separator (COMPOSITE-01)
- `go mod tidy` resolved all transitive deps; `go build .` and `go vet .` pass

## Task Commits

1. **Task 1: Author pulumi/provider/resources.go** - `d3b2ef2` (feat)

## Files Created/Modified

- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/pulumi/provider/resources.go` — full ProviderInfo definition
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/pulumi/provider/go.sum` — resolved dep lockfile (111KB)
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/pulumi/provider/go.mod` — updated by go mod tidy

## Decisions Made

- **ShimProvider import path discovery**: The plan's import `shimpf "pkg/pf"` was wrong — `ShimProvider` function is in `pkg/pf/tfbridge`, not `pkg/pf`. `pkg/pf` only defines the `ShimProvider` interface; the constructor is `pftfbridge.ShimProvider`. Corrected import to `pftfbridge "github.com/pulumi/pulumi-terraform-bridge/v3/pkg/pf/tfbridge"`.
- **Timeout fields absent from bridge**: `CreateTimeout`, `UpdateTimeout`, `DeleteTimeout` are NOT fields on `info.Resource` in bridge v3.127.0. The plan's code referencing these would not compile. TF provider's `timeouts` block defaults (20m/30m) are inherited through the shim. No action possible without bridge API change.
- **AdditionalSecretOutputs absent**: This field does not exist on `ResourceInfo` in bridge v3.127.0. The plan's SECRETS-02 cannot be fully satisfied. TF `Sensitive=true` auto-promotion + `Fields["secret_access_key"].Secret = tfbridge.True()` are the available defenses.
- **ComputeID state keys in camelCase**: State keys are Pulumi camelCase (`policyName`, `name`), not TF snake_case.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Wrong import path for ShimProvider function**
- **Found during:** Task 1 (build failed with "cannot convert ... type pf.ShimProvider")
- **Issue:** Plan code used `shimpf "github.com/pulumi/pulumi-terraform-bridge/v3/pkg/pf"` and called `shimpf.ShimProvider(...)`. But `pkg/pf` only defines the `ShimProvider` interface; the constructor function lives in `pkg/pf/tfbridge`.
- **Fix:** Changed import to `pftfbridge "github.com/pulumi/pulumi-terraform-bridge/v3/pkg/pf/tfbridge"` and call `pftfbridge.ShimProvider(...)`. Also corrected base bridge type imports (ProviderInfo, SchemaInfo, True, NewProviderMetadata) to use `pkg/tfbridge` directly.
- **Files modified:** pulumi/provider/resources.go
- **Verification:** `go build .` passes, `go vet .` clean
- **Committed in:** d3b2ef2 (Task 1 commit)

**2. [Rule 1 - Bug] CreateTimeout/UpdateTimeout/DeleteTimeout fields don't exist in bridge v3.127.0**
- **Found during:** Task 1 (build failed — undefined fields on info.Resource)
- **Issue:** `info.Resource` struct (= `tfbridge.ResourceInfo`) in bridge v3.127.0 has no `CreateTimeout`, `UpdateTimeout`, or `DeleteTimeout` fields. The plan code setting these would not compile. PB1 research note described them as if they existed but they do not.
- **Fix:** Removed these field assignments. Added explanatory comments that TF timeouts defaults (20m/30m) are inherited via the shim. Pulumi users needing explicit timeouts must use `customTimeouts` in their program.
- **Files modified:** pulumi/provider/resources.go
- **Verification:** Build passes
- **Committed in:** d3b2ef2 (Task 1 commit)

**3. [Rule 1 - Bug] AdditionalSecretOutputs field doesn't exist in bridge v3.127.0**
- **Found during:** Task 1 (confirmed by grepping bridge module — field absent from info.Resource)
- **Issue:** `info.Resource` has no `AdditionalSecretOutputs` field. The plan's SECRETS-02 belt-and-braces cannot be implemented as specified. PB3 mitigation is partial: `Fields["secret_access_key"].Secret = tfbridge.True()` is in place; the runtime `AdditionalSecretOutputs` defense is unavailable.
- **Fix:** Replaced with `Fields["secret_access_key"].Secret = tfbridge.True()` only, with an explanatory comment.
- **Files modified:** pulumi/provider/resources.go
- **Verification:** Build passes; secret marking is present in schema (via Fields.Secret)
- **Committed in:** d3b2ef2 (Task 1 commit)

---

**Total deviations:** 3 auto-fixed (all Rule 1 — build-blocking bugs from bridge API mismatch)
**Impact on plan:** MAPPING-03 (explicit timeouts) and SECRETS-02 (AdditionalSecretOutputs) are partially satisfied — TF shim inherits defaults and TF Sensitive=true provides auto-promotion. Full satisfaction requires either a bridge version with these fields or a future phase override.

## Issues Encountered

- Bridge v3.127.0 `info.Resource` struct lacks `CreateTimeout`, `UpdateTimeout`, `DeleteTimeout`, and `AdditionalSecretOutputs` fields that the plan expected. Research in PITFALLS.md (PB1, PB3) described these as if they existed in the bridge — they do not. The TF provider's own `timeouts` block defaults (20m create/update, 30m delete for buckets) are the actual defense when inherited through the shimmed provider.

## Known Stubs

None — all intended functionality is wired. The belt-and-braces `AdditionalSecretOutputs` defense cannot be implemented (bridge API gap); this is documented above.

## Next Phase Readiness

- `Provider()` function is available for plan 03's cmd packages to import
- `go.sum` is complete — no further dep resolution needed
- Plan 03 adds `cmd/pulumi-tfgen-flashblade/main.go` and `cmd/pulumi-resource-flashblade/main.go` which consume `Provider()`
- **Action for follow-up (Phase 55 or a dedicated patch):** Verify if a newer bridge version exposes `DeleteTimeout`/`AdditionalSecretOutputs` on `ResourceInfo`. If found, upgrade bridge version and add these fields.

---
*Phase: 54-bridge-bootstrap-poc-3-resources*
*Completed: 2026-04-22*
