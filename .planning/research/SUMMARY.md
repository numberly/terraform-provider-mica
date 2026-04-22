# Research Summary — pulumi-2.22.3 (Pulumi Bridge Alpha)

*Synthesized 2026-04-21 from STACK.md, FEATURES.md, ARCHITECTURE.md, PITFALLS.md, and `pulumi-bridge.md`.*

## Executive Summary

Milestone `pulumi-2.22.3` adds a Pulumi bridge layer (`./pulumi/` sub-directory) on top of the existing Terraform FlashBlade provider (terraform-plugin-framework, Go 1.25, 28 resources + 21 data sources, 779 tests). The bridge uses `pulumi/pulumi-terraform-bridge/v3 pkg/pf/*` — the correct path for `terraform-plugin-framework`-based providers (not `pkg/tfgen`/`shimv2`).

The implementation lives in `./pulumi/` with **three separate `go.mod` files** (bridge binaries, runtime plugin, consumer Go SDK). The root `go.mod` and all `internal/*` code are untouched — the TF provider is wired via `replace ../`. Scope is strictly **Python + Go SDKs**, private distribution via **GitHub Releases** (no PyPI, no npm, no Pulumi Registry, no `ci-mgmt` automation).

Tokenization via `MustComputeTokens` + `KnownModules` + `MustApplyAutoAliases` handles ~90% of the 49 resources/data-sources automatically. Manual overrides are required on: 4 composite-ID resources (`ComputeID`), 6 sensitive/write-once fields (`Secret` + `AdditionalSecretOutputs`), all 28 resources (`timeouts.Omit = true`), and 2 soft-delete resources (`DeleteTimeout: 30*time.Minute`).

**3 open questions block Phase 1 start** — must be resolved before any code is written.

## Pinned Stack (verified 2026-04-21)

| Component | Version | Source |
|---|---|---|
| `pulumi-terraform-bridge/v3` | **v3.127.0** | pkg.go.dev 2026-04-21 (pulumi-bridge.md's v3.126.0 is stale) |
| `pulumi/sdk/v3` + `pkg/v3` | **v3.231.0** | pkg.go.dev 2026-04-16 (pulumi-bridge.md's v3.220.0 is stale) |
| `terraform-plugin-framework` | v1.19.0 | pulumi-random canonical reference |
| `terraform-plugin-sdk/v2` replace SHA | `v2.0.0-20260318212141-5525259d096b` | Required fork for bridge — re-verify on every bridge bump |
| `pulumictl` | v0.0.50 | Version injection (ldflags + Python metadata) |
| `python build` | 1.2.1 | Wheel packaging |
| Go toolchain | 1.25 (existing) | Bridge v3.127.0 requires Go 1.22+ |

**Three `go.mod` files required:**
1. `./go.mod` — TF provider, **unchanged**
2. `./pulumi/provider/go.mod` — bridge + binaries + `replace ../../` for TF provider
3. `./pulumi/sdk/go/go.mod` — lean consumer SDK (only `pulumi/sdk/v3`)

**Anti-patterns to avoid:** using `pkg/tfgen` instead of `pkg/pf/tfgen`; running `go mod tidy` from root; using `shimv2.NewProvider(...)` instead of `pf.ShimProvider(fb.New(version)())`.

## Mapping Scope (49 TF resources + data sources)

| Concern | Scope |
|---|---|
| Auto-tokenization (`MustComputeTokens` + `KnownModules`) | ~90% of 49 resources/DS — one-liner per module group (bucket, filesystem, objectstore, policy, network, array, index) |
| Composite ID `ComputeID` | **4 resources** — see correction table below |
| Secrets belt-and-braces (`Secret: tfbridge.True()` + `AdditionalSecretOutputs`) | **6 fields** — `api_token`, `secret_access_key`, `bind_password`, `connection_key`, certificate `private_key`, `private_key_passphrase` |
| `Fields["timeouts"].Omit = true` | **All 28 resources** — applied via helper |
| `DeleteTimeout: 30*time.Minute` explicit override | **2 resources** — `flashblade_bucket`, `flashblade_filesystem` (soft-delete poll) |
| `Create/Update/DeleteTimeout` explicit | All 28 resources, matching TF timeouts block defaults (Create 20min, Update 20min, Delete 30min) |
| State upgraders requiring `pulumi refresh` smoke test | `flashblade_server` (v0→v1→v2), `flashblade_directory_service_role` (v0→v1), `flashblade_remote_credentials` (v0→v1) |

### Composite ID correction (CRITICAL — do NOT copy `pulumi-bridge.md` Section 10.3)

`pulumi-bridge.md` Section 10.3 shows `policy_name:rule_index` (colon + integer). **This is incorrect.** Verified against `internal/provider/*_resource.go` `readIntoState`:

| Resource | Actual composite ID | Separator | Notes |
|---|---|---|---|
| `flashblade_object_store_access_policy_rule` | `policyName/ruleName` (string) | `/` | Not integer index |
| `flashblade_bucket_access_policy_rule` | `policyName/ruleName` | `/` | Same pattern |
| `flashblade_network_access_policy_rule` | `policyName/ruleName` | `/` | Same pattern |
| `flashblade_management_access_policy_directory_service_role_membership` | `roleName/policyName` (role FIRST) | `/` | Policy names contain `:` (e.g. `pure:policy/array_admin`) so `/` is mandatory |

**Rule:** Before writing any `ComputeID`, read the resource's `readIntoState` in `internal/provider/` to confirm the exact format.

## Distribution — Private GitHub Releases

| Artifact | Mechanism |
|---|---|
| `pulumi-resource-flashblade` (runtime plugin binary) | goreleaser → GitHub Release assets → `pulumi plugin install resource flashblade vX.Y.Z --server github://api.github.com/<org>` |
| Python SDK (`.whl`) | `python -m build` → attach via goreleaser `extra_files` → install from GitHub Release URL |
| Go SDK | Consumer `go get` with `GOPRIVATE=github.com/<org>/*` — **requires git tag `sdk/go/vX.Y.Z`** in addition to release tag `pulumi-X.Y.Z` |
| Signing | cosign (same pattern as existing TF goreleaser) |

**Separate pipelines:** existing `.goreleaser.yml` stays unchanged (triggered on `v*` tags). New `./pulumi/.goreleaser.yml` triggered on `pulumi-*` tags. Only `pulumi-resource-flashblade` is released — `pulumi-tfgen-flashblade` is build-time only.

**Out of scope (do not implement):** `ci-mgmt` template generation, PyPI publish, npm publish, NuGet publish, Maven publish, registry.pulumi.com, Pulumi AI docs portal.

## Critical Pitfalls Requiring Phase 1 Mitigation

| ID | Severity | Pitfall | Mitigation |
|---|---|---|---|
| **PB1** | CRITICAL | Pulumi default `DeleteTimeout` = 5 min kills `pollUntilGone` on buckets/filesystems (30-min polling loop). Orphaned bucket + next `pulumi up` → 409 name collision. Bridge issue #1652. | `DeleteTimeout: 30*time.Minute` in `ResourceInfo` for `flashblade_bucket` + `flashblade_filesystem`. `resources_test.go` asserts `DeleteTimeout >= 25*time.Minute`. |
| **PB2** | CRITICAL | Writing `ComputeID` from `pulumi-bridge.md` Section 10.3 example (`policyName:ruleIndex`) causes `pulumi import` to always fail with bridge issue #2272. | Before any `ComputeID`, read `readIntoState` in `internal/provider/<resource>_resource.go`. Use `/` separator + string `ruleName`. |
| **PB3** | HIGH | Bridge issue #1028 — secret-ness lost on state update for write-once fields (API returns null on Read). `Sensitive: true` auto-promotion alone is insufficient. | `Secret: tfbridge.True()` + `AdditionalSecretOutputs: []string{"secretAccessKey", ...}` on all 6 sensitive fields. Verify via `pulumi stack export`. |
| **PB4** | MEDIUM | Replace directive SHA is coupled to bridge version. `go get pulumi-terraform-bridge@latest` without updating SHA breaks build silently. | After every bridge bump, verify `pulumi/provider/go.mod` replace SHA matches the bridge's own `go.mod`. Document in upgrade runbook. |
| **PB5** | HIGH | Tag `pulumi-0.1.0` alone is insufficient — Go modules require `sdk/go/v0.1.0` tag for the sub-module. Missing causes `go get` 404. | Post-goreleaser step in `pulumi-release.yml`: `git tag sdk/go/vX.Y.Z && git push origin sdk/go/vX.Y.Z`. |
| **PB6** | MEDIUM | `flashblade_server` has SchemaVersion 2 + 2 upgraders + `types.List` attrs → highest risk for bridge issue #1667 (`RawState` distortion). | Phase 2 `pulumi refresh` smoke test with snapshots from v0 and v1 states. |
| **PB7** | MEDIUM | `timeouts {}` block leaks into every SDK if not stripped before first `make tfgen`. Re-running is cheap; cleaning a committed schema is noisy. | Helper `omitTimeoutsOnAll(Resources)` applied in `resources.go` BEFORE first `make tfgen`. |

## Phase Breakdown (feeds roadmapper)

**Phase 1 — Bootstrap + ProviderInfo POC (3 resources).** Gate on `pf.ShimProvider` + 3 `go.mod` compiling. Validate full chain on `target` (auto-token baseline), `remote_credentials` (secrets), `bucket` (soft-delete + 30min timeout). Deliverables: binaries buildable, schema emitted, `resources_test.go` green, ProgramTest passing on 3 resources. Mitigates PB1, PB2, PB3, PB4, PB7.

**Phase 2 — Full mapping (28 resources + 21 data sources).** Module assignment, 4 `ComputeID` callbacks (verified against `readIntoState`), full secrets coverage, state upgrader `pulumi refresh` smoke tests. Schema snapshot committed + CI diff gate.

**Phase 3 — SDK generation (Python + Go).** `make generate_python generate_go`, embed schemas, Python wheel build, Go SDK `sdk/go/go.mod`. CI job chain: `prerequisites` (tfgen + upload-artifact) → `build_provider` + `generate_sdk_*` (download-artifact).

**Phase 4 — Private release pipeline.** `./pulumi/.goreleaser.yml`, `pulumi-ci.yml`, `pulumi-release.yml`, cosign, post-release `sdk/go/vX.Y.Z` tagging. Smoke test: consumer project installs plugin via `github://api.github.com/<org>` + imports SDK.

**Phase 5 — Docs + onboarding.** Auto-conversion `PULUMI_CONVERT=1` of existing HCL examples, hand-written `bucket-py` + `bucket-go` ProgramTest examples, `docs/pulumi-installation.md` (GOPRIVATE setup, plugin install URL, wheel install URL).

## Blocking Open Questions (must resolve before Phase 1)

1. **Go module path** — `github.com/soulkyu/*` (current TF provider) vs `github.com/pure-storage/*` (pulumi-bridge.md example)? Determines all generated paths; post-Phase-1 change = mass renaming.
2. **Schema commit policy** — commit `schema.json` + `schema-embed.json` + `bridge-metadata.json` (canonical, CI diff gate), or `.gitignore` them (forces `make tfgen` on every checkout)? Recommendation: **commit** (matches pulumi-random, pulumi-cloudflare).
3. **Write-Only Fields SDK readiness** — is the pattern stable in `pulumi/sdk/v3 v3.231.0` for Python AND Go? If yes, apply to `secret_access_key`, `bind_password`, etc. If no, fallback to `Secret` + `AdditionalSecretOutputs` only. Requires checking SDK changelog.

## Non-Blocking Open Questions

- `SetAutonaming` scope — disable entirely or leave opt-in? Storage names are operational identifiers, random suffix is likely wrong. **Recommendation: omit `SetAutonaming` call.**
- Token module for `index` fallback resources — `target`, `server`, `remote_credentials`, `certificate`, `directory_service_*` default to `index`. Acceptable for private provider; validate with first `make tfgen` output.
- `GONOSUMCHECK` vs `GONOSUMDB` — env var name differs across Go versions. Validate with `go env` on Go 1.25 before documenting consumer setup.

## Confidence Assessment

| Area | Confidence | Basis |
|---|---|---|
| Stack versions | HIGH | Live pkg.go.dev verification 2026-04-21; cross-checked against pulumi-random + pulumi-cloudflare |
| Architecture (monorepo + replace) | HIGH | Verified against boilerplate + 4 canonical bridged providers |
| Features + mapping coverage | HIGH | Derived from bridge source (`info.go`) + actual `internal/provider/` inspection |
| Pitfalls (bridge-specific) | HIGH | Each PB cross-referenced to specific bridge issue number + actual provider code |
| Write-Only Fields maturity | MEDIUM | Pattern documented; SDK readiness for Python+Go on v3.231.0 not verified |
| Scale/performance | MEDIUM | Not validated against real FlashBlade at scale |

**Overall: HIGH — but Phase 1 gated on 3 open questions above.**

---
*Sources synthesized:* `.planning/research/STACK.md`, `.planning/research/FEATURES.md`, `.planning/research/ARCHITECTURE.md`, `.planning/research/PITFALLS.md`, `pulumi-bridge.md`
