# Phase 54: Bridge Bootstrap + POC (3 Resources) - Context

**Gathered:** 2026-04-21
**Status:** Ready for planning

<domain>
## Phase Boundary

Stand up `./pulumi/` sub-directory (3 separate `go.mod` modules, `pulumi-tfgen-flashblade` + `pulumi-resource-flashblade` binaries, `ProviderInfo` in `resources.go`) and prove the full bridge chain (`pf.ShimProvider` → `make tfgen` → embedded schema → compiled runtime plugin) on 3 representative resources: `flashblade_target` (auto-tokenization baseline), `flashblade_object_store_remote_credentials` (write-once secrets), `flashblade_bucket` (soft-delete + 30-min delete timeout).

Mitigates critical bridge pitfalls PB1 (DeleteTimeout truncation), PB2 (wrong composite ID separator), PB3 (secret-ness lost on state update), PB4 (replace SHA coupling), PB7 (timeouts block leaks into SDK).

**In scope:** BRIDGE-01..05, MAPPING-02/03/05, MAPPING-01/04 (extended per D-02), COMPOSITE-01, SECRETS-01/02 (scoped to POC 3 resources), SOFTDELETE-01, TEST-01.

**Out of scope (other phases):** MAPPING-01 overrides for the remaining 46 resources (Phase 55), COMPOSITE-02..04 (Phase 55), full SECRETS-02 on all 6 fields + SECRETS-03 coverage assertion (Phase 55), SOFTDELETE-02/03 on filesystem (Phase 55), UPGRADE-* (Phase 55), SDK generation (Phase 56), CI pipeline (Phase 57), release + ProgramTest + docs (Phase 58).

</domain>

<decisions>
## Implementation Decisions

### Provider configuration surface

- **D-01:** Mirror the TF provider's full config surface 1:1 in `ProviderInfo.Config`. Every key from `FlashBladeProvider.Schema` is exposed through the bridge — `endpoint`, `api_token`, `oauth2_client_id`, `oauth2_client_secret`, `oauth2_token_url`, `skip_tls_verify`, `ca_certificate`. No keys dropped or renamed in Phase 54. Rationale: zero-surprise migration for TF users and no second-pass schema change in Phase 55.

### Mapping scope in Phase 54

- **D-02:** `MustComputeTokens` + `KnownModules` + `MustApplyAutoAliases` are applied to **all 49 TF resources + data sources** in this phase (not just the 3 POC resources). `make tfgen` produces the full schema from Phase 54. Phase 55 only adds targeted overrides (remaining `ComputeID`, additional `AdditionalSecretOutputs`, `pulumi refresh` state-upgrader tests). This locks module assignments (`bucket`, `filesystem`, `policy`, `objectstore`, `array`, `network`, `index`) early and avoids a disruptive schema shift between 54 and 55.

### Version injection

- **D-03:** Version embedded via `git describe --tags --dirty --always` piped through the Makefile into Go `-ldflags "-X <module>/provider/pkg/version.Version=$(VERSION)"`. No `pulumictl` dependency. Pattern aligned with the existing TF `.goreleaser.yml` that already uses `git describe`. This keeps tooling uniform between the TF release pipeline and the Pulumi release pipeline (Phase 58).

### Test tier in Phase 54

- **D-04:** Unit-only in Phase 54. `./pulumi/provider/resources_test.go` asserts: mapping coverage (`len(Resources) == 28`, `len(DataSources) == 21`), every `Sensitive` TF field is promoted to a Pulumi Secret for the 3 POC resources, `bucket.DeleteTimeout >= 25*time.Minute`, no `timeouts` input appears in any resource schema, `api_token` is marked Secret in `ProviderInfo.Config`. No `ProgramTest` / `pulumi up` in Phase 54 — the user handles manual E2E verification against a real FlashBlade outside the automated test suite. Full ProgramTest coverage deferred to Phase 58 as already planned.

### Scoped to POC 3 resources (Phase 54 partial coverage of mapping reqs)

- **D-05:** Phase 54 hits the minimum subset of REQUIREMENTS.md overrides needed to prove the chain:
  - `flashblade_target` — validates auto-tokenization baseline, explicit `Create/Update/DeleteTimeout` matching TF defaults.
  - `flashblade_object_store_remote_credentials` — validates `Secret: tfbridge.True()` + `AdditionalSecretOutputs: ["secretAccessKey"]` pattern. `api_token` in `ProviderInfo.Config` also gets Secret mark.
  - `flashblade_bucket` — validates `Fields["timeouts"].Omit = true` + explicit `DeleteTimeout: 30*time.Minute`.
  - `flashblade_object_store_access_policy_rule` — validates `ComputeID` with `/` separator + string rule name (format verified against `ImportState` at `internal/provider/object_store_access_policy_rule_resource.go:361-387`).

### Claude's Discretion

- Exact directory structure under `./pulumi/` (flat `provider/` + `sdk/` + `examples/` is confirmed; internal grouping inside each is Claude's call).
- Makefile target names (expected conventional: `tfgen`, `provider`, `generate_python`, `generate_go`, `build_sdks`, `test`, `lint`, `clean`).
- `omitTimeoutsOnAll` helper naming and placement (Claude's call, applied before `MustComputeTokens`).
- Version stamping flow detail (ldflags path, where `VERSION` is computed in the Makefile).
- Order of boilerplate adaptation steps during Phase 54 execution.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Phase-scoping research (authoritative for this milestone)

- `.planning/research/SUMMARY.md` — milestone-level synthesis with pinned versions, composite ID corrections, and pitfall mitigations.
- `.planning/research/STACK.md` — pinned bridge `v3.127.0`, SDK `v3.231.0`, replace SHA `v2.0.0-20260318212141-5525259d096b`, three-go.mod layout.
- `.planning/research/ARCHITECTURE.md` — monorepo integration, embed chain, build ordering, Go SDK sub-module tagging.
- `.planning/research/FEATURES.md` — table-stakes behaviors, composite ID table, anti-features.
- `.planning/research/PITFALLS.md` — PB1 (DeleteTimeout), PB2 (composite ID format), PB3 (secrets belt-and-braces), PB4 (replace SHA), PB7 (timeouts omit).
- `pulumi-bridge.md` (repo root) — consolidated pre-research (Section 1 Architecture, Section 3 Wiring PF, Section 4 Mapping Conventions, Section 5 Makefile, Section 10 Pitfalls). **Section 10.3 `policy_name:rule_index` example is INCORRECT** — use `/` separator with string rule name (see COMPOSITE-01 and D-05).

### Milestone contract

- `.planning/REQUIREMENTS.md` — 39 REQ-IDs across BRIDGE/MAPPING/COMPOSITE/SECRETS/SOFTDELETE/UPGRADE/SDK/CI/RELEASE/TEST/DOCS. Phase 54 owns 13 (see Traceability table).
- `.planning/ROADMAP.md` — Phase 54 section (~line 796): goal, depends-on, success criteria.
- `.planning/PROJECT.md` — Current Milestone section (Pulumi Bridge Alpha goal + constraints).
- `.planning/STATE.md` — Current Position, Accumulated Context (locked decisions).

### Provider code references for POC resources

- `internal/provider/provider.go:48-52` — `New(version string) func() provider.Provider` factory. Bridge wiring: `pf.ShimProvider(fb.New(version.Version)())`.
- `internal/provider/target_resource.go` — POC resource #1 (auto-tokenization baseline).
- `internal/provider/object_store_remote_credentials_resource.go` — POC resource #2 (secret_access_key write-once pattern).
- `internal/provider/bucket_resource.go` — POC resource #3 (soft-delete + eradication_config + 30-min timeouts block). See `Delete` + `pollUntilGone[T]` usage.
- `internal/provider/object_store_access_policy_rule_resource.go:361-387` (`ImportState`) — composite ID format `policy_name/rule_name` confirmation (canonical source for `ComputeID` logic).
- `CONVENTIONS.md` — resource convention patterns (POST/Patch structs, `**NamedReference`, schema versioning). Bridge layer must not break these invariants.

### External libraries / guides

- `pulumi/pulumi-terraform-bridge/v3 pkg/pf/*` — `tfgen.Main`, `tfbridge.Main`, `pf.ShimProvider`. Source docs: `pkg/pf/README.md` and `docs/guides/upgrade-sdk-to-pf.md` in the bridge repo.
- `pulumi/pulumi-tf-provider-boilerplate` — reference layout for `provider/cmd/pulumi-tfgen-<name>`, `provider/cmd/pulumi-resource-<name>`, `provider/pkg/version/version.go`.
- `pulumi/pulumi-random` — canonical `pkg/pf` reference provider (closest to our shape: framework-based, small surface).
- `pulumi/pulumi-cloudflare` — reference for Makefile, two-goreleaser split, three-go.mod layout.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets

- **`provider.New(version string) func() provider.Provider`** (`internal/provider/provider.go:48-52`) — already returns the exact `func() provider.Provider` closure required by `pf.ShimProvider`. Wiring line: `pf.ShimProvider(fb.New(version.Version)())`. No changes to the TF provider code needed.
- **`ImportState` implementations** (all `*_resource.go` files) — already parse composite IDs via `parseCompositeID(req.ID, 2)`. `ComputeID` callbacks in the bridge just need to reconstruct the SAME format. Read the resource's `ImportState` to confirm separator; do NOT guess from docs.
- **Bucket soft-delete with `pollUntilGone[T]`** (`internal/provider/bucket_resource.go`) — preserved as-is. Bridge layer only adds `DeleteTimeout: 30*time.Minute` in `ResourceInfo` so Pulumi doesn't kill the poll at 5 min.
- **`.goreleaser.yml`** (repo root) — existing TF release pipeline already uses `git describe`-style versioning + cosign. New `./pulumi/.goreleaser.yml` will mirror the cosign and platform matrix, with its own trigger on `pulumi-*` tags (Phase 58).

### Established Patterns

- **Go module path:** `github.com/numberly/opentofu-provider-flashblade` (confirmed in root `go.mod`). Bridge modules nested under `./pulumi/provider/` and `./pulumi/sdk/go/`.
- **Resource registration:** `FlashBladeProvider.Resources()` returns the closed list (28) + `DataSources()` returns 21. Bridge's `MustComputeTokens` walks `pf.ShimProvider`'s introspection — no need to maintain a parallel list.
- **Framework version:** `terraform-plugin-framework` via `provider.Provider` interface. `pf.ShimProvider` is the correct entry point (NOT `shimv2.NewProvider`).
- **Sensitive fields:** All sensitive TF fields already use `schema.Sensitive: true`. Bridge auto-promotes these; `AdditionalSecretOutputs` is the belt-and-braces layer for write-once fields per PB3.

### Integration Points

- **TF provider ↔ bridge:** `./pulumi/provider/go.mod` depends on `github.com/numberly/opentofu-provider-flashblade` (root module) via `replace ../../`. Root `go.mod` is not modified.
- **Schema files (committed):** `./pulumi/provider/cmd/pulumi-resource-flashblade/schema.json`, `schema-embed.json`, `bridge-metadata.json` are generated by `make tfgen` and committed. Phase 54 produces the first baseline; Phase 57 adds CI drift gate.
- **Existing CI/release:** `.github/workflows/*.yml` and `.goreleaser.yml` are untouched in Phase 54. Pulumi CI (Phase 57) and release (Phase 58) land in separate files.

</code_context>

<specifics>
## Specific Ideas

- **Composite ID verification discipline:** Before writing COMPOSITE-01's `ComputeID`, re-read `internal/provider/object_store_access_policy_rule_resource.go:361-387` to confirm `policy_name/rule_name` format. Do NOT copy `pulumi-bridge.md` Section 10.3 example (shows `policy_name:rule_index` which is wrong).
- **Replace directive SHA:** Pin exactly `v2.0.0-20260318212141-5525259d096b` for `hashicorp/terraform-plugin-sdk/v2 => github.com/pulumi/terraform-plugin-sdk/v2` — coupled to bridge `v3.127.0`. Any bridge bump requires re-verifying this SHA against the bridge's own `go.mod` (PB4 mitigation).
- **Pulumi identifier collisions:** Verify first `make tfgen` output for any `id`/`urn`/`provider` field name collisions + Python token collisions on resources named `target`. Handle via `Fields[].Name` override if collision found.
- **No `SetAutonaming`:** Do NOT call `prov.SetAutonaming(...)` — FlashBlade storage names are operational identifiers; random suffix is wrong for this domain (confirmed in SUMMARY.md and MAPPING-05 requirement).

</specifics>

<deferred>
## Deferred Ideas

- ProgramTest `pulumi up`/`destroy` on real FlashBlade → Phase 58 (user handles manual E2E during Phase 54).
- Additional composite IDs (COMPOSITE-02/03/04) → Phase 55.
- Full `AdditionalSecretOutputs` audit on the remaining 4 sensitive fields (bind_password, connection_key, private_key, private_key_passphrase) → Phase 55 (Phase 54 covers api_token + secret_access_key only).
- State upgrader `pulumi refresh` smoke tests → Phase 55.
- Python wheel + Go SDK sub-module generation → Phase 56.
- CI workflows → Phase 57.
- Release pipeline (goreleaser + cosign + `sdk/go/vX.Y.Z` tag) → Phase 58.
- `PULUMI_CONVERT=1` HCL example auto-conversion → Phase 58.
- Future milestone: migrate from `Secret + AdditionalSecretOutputs` to Write-Only Fields once SDK v3.231.0+ readiness is verified for Python + Go.

</deferred>

---

*Phase: 54-bridge-bootstrap-poc-3-resources*
*Context gathered: 2026-04-21*
