---
phase: 61-flashblade-snmp-manager
plan: 01
subsystem: infra
tags: [terraform-provider, flashblade, snmp, api-2.23, sensitive-write-once, nested-attributes]

# Dependency graph
requires:
  - phase: 59-api-2.23-consolidation
    provides: FlashBlade API 2.23 client infrastructure, getOneByName[T] / postOne / patchOne generics, mock helper conventions
provides:
  - flashblade_snmp_manager resource (CRUD on /api/2.23/snmp-managers)
  - flashblade_snmp_manager data source (lookup by name)
  - Reusable atomic-nested-block pattern for resources with sensitive write-once fields
  - Mock handler precedent for stripping sensitive fields on GET parity with real API
affects: [future-snmp-test-action-resource, future-resources-with-sensitive-write-once-fields]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Sensitive write-once fields: never read from API GET, preserved verbatim in state, nulled on ImportState"
    - "Atomic nested block (v2c/v3) following ArrayConnectionPatch.Throttle template"
    - "Per-leaf drift detection: 6 explicit tflog.Debug calls covering host, notification, version, v3.user, v3.auth_protocol, v3.privacy_protocol"
    - "In-place version switch (v2c <-> v3) with no RequiresReplace"

key-files:
  created:
    - internal/client/snmp_managers.go
    - internal/client/snmp_managers_test.go
    - internal/testmock/handlers/snmp_managers.go
    - internal/provider/snmp_manager_resource.go
    - internal/provider/snmp_manager_resource_test.go
    - internal/provider/snmp_manager_data_source.go
    - internal/provider/snmp_manager_data_source_test.go
    - examples/resources/flashblade_snmp_manager/resource.tf
    - examples/resources/flashblade_snmp_manager/import.sh
    - examples/data-sources/flashblade_snmp_manager/data-source.tf
    - docs/resources/snmp_manager.md
    - docs/data-sources/snmp_manager.md
  modified:
    - internal/client/models_admin.go
    - internal/provider/provider.go
    - ROADMAP.md

key-decisions:
  - "Sensitive write-once: community / auth_passphrase / privacy_passphrase preserved from state in Read; nulled in ImportState (operator re-supplies via apply)."
  - "Used Pattern A from CONVENTIONS api_contracts: SnmpManagerPost uses *SnmpV3Post (stricter validators) while SnmpManagerPatch reuses *SnmpV3."
  - "In-place version switch supported (no RequiresReplace on `version`) per D-06; ImportState surfaces the API-reported version + corresponding block."
  - "Mock server.go NOT touched — codebase pattern is per-test handler registration via ms.Mux (not centralized in NewMockServer); follows existing target/array_admin precedent."
  - "Stricter SnmpV3Post validators applied at provider schema level (LengthAtMost(32) on auth_passphrase, LengthBetween(8,63) on privacy_passphrase) so PATCH never sends values the array would reject on POST."

patterns-established:
  - "Atomic nested *Patch block sent verbatim (no per-leaf pointer): copy of ArrayConnectionPatch.Throttle shape"
  - "stripSensitive() helper in mock handler to mirror real API GET behaviour (community/passphrases blanked)"
  - "snmpV2cAttrTypes / snmpV3AttrTypes shared between resource and data source for consistent types.Object construction"

requirements-completed: [SNMP-01, SNMP-02, SNMP-03, SNMP-04, SNMP-05, SNMP-06, SNMP-07, SNMP-08, SNMP-09, SNMP-10, SNMP-11, SNMP-12, SNMP-13]

# Metrics
duration: 15min
completed: 2026-05-20
---

# Phase 61 Plan 01: Implement SNMP Manager Summary

**flashblade_snmp_manager resource + data source (full CRUD on /api/2.23/snmp-managers) with atomic v2c/v3 nested blocks, sensitive write-once secrets (community + 2 passphrases), per-leaf drift detection across 6 leaves, and in-place v2c<->v3 switch support.**

## Performance

- **Duration:** ~15 min
- **Started:** 2026-05-20T12:22:05Z
- **Completed:** 2026-05-20T12:37:08Z
- **Tasks:** 13
- **Files modified/created:** 15 (12 created, 3 modified)
- **Tests:** 816 total (807 baseline + 9 new — exactly meets `TEST_BASELINE + 9` floor per CONVENTIONS.md New Resource checklist item 14)

## Accomplishments

- **flashblade_snmp_manager resource**: full CRUD on `/api/2.23/snmp-managers`; 4 framework interfaces (`Resource`, `ResourceWithConfigure`, `ResourceWithImportState`, `ResourceWithUpgradeState`); schema `Version: 0`; all 4 timeouts (20m/5m/20m/30m); v2c and v3 as `schema.SingleNestedAttribute`; in-place version switch supported.
- **flashblade_snmp_manager data source**: 2 framework interfaces (`DataSource`, `DataSourceWithConfigure`); `name` Required, all other fields Computed; not-found via `AddError`.
- **Sensitive write-once contract**: `community`, `auth_passphrase`, `privacy_passphrase` are never overwritten from API responses in `Read()` (3 `// skip sensitive write-once` markers); preserved verbatim in state; nulled on `ImportState`.
- **6-leaf drift detection**: 6 literal `tflog.Debug(ctx, "drift detected", ...)` calls covering `host`, `notification`, `version`, `v3.user`, `v3.auth_protocol`, `v3.privacy_protocol`. Zero sensitive fields ever surface in logs.
- **Mock handler**: thread-safe `snmpManagerStore` with `Seed`, `Get`, `Mutate` test helpers; GET no-match returns HTTP 200 + `{"items": []}` (mirrors real API); `stripSensitive()` blanks community/passphrases on every response.
- **9 new TestUnit_ tests**: 5 client + 3 resource (Lifecycle, Import, DriftDetection) + 1 data source (Basic).
- **HCL examples + generated docs**: v3 primary + commented v2c snippet + in-place switch note; import by name; data source output; `make docs` idempotent on second run.
- **ROADMAP.md row move**: from `Medium Priority -- Admin and security` to `Array Administration / Implemented` with status `Done` and `v2.23.1` note; coverage counters bumped to 56 resources / 44 data sources, provider version `v2.23.1`.

## Task Commits

Each task was committed atomically with `--no-verify` and no `Co-Authored-By` trailer (per project rules):

1. **T01: SnmpManager client model structs** — `a241ec1` (feat)
2. **T02: SnmpManager client CRUD methods** — `0c83d83` (feat)
3. **T03: Mock handler /snmp-managers** — `cbf400c` (feat)
4. **T04: 5 client tests** — `3d9c980` (test)
5. **T05: flashblade_snmp_manager resource** — `5a1a55e` (feat)
6. **T06: 3 resource tests** — `6430079` (test)
7. **T07: flashblade_snmp_manager data source** — `08e4227` (feat)
8. **T08: 1 data source test** — `7ea1f7d` (test)
9. **T09: Register in provider.go** — `03789e1` (feat)
10. **T10: HCL examples** — `565ad2b` (docs)
11. **T11-T13: Generated docs + ROADMAP row move + staticcheck QF1008 fix** — `24098d1` (docs)

All 11 commits land on branch `implem-snmp-managers` (branched from clean `main`).

## Files Created/Modified

**Client layer:**
- `internal/client/models_admin.go` — appended 6 struct types (`SnmpV2c`, `SnmpV3`, `SnmpV3Post`, `SnmpManager`, `SnmpManagerPost`, `SnmpManagerPatch`)
- `internal/client/snmp_managers.go` — Get/List/Post/Patch/Delete via `getOneByName[SnmpManager]` / `postOne` / `patchOne` / `c.delete`
- `internal/client/snmp_managers_test.go` — 5 `TestUnit_SnmpManager_*` tests against `httptest.NewServer` + the mock handler

**Mock layer:**
- `internal/testmock/handlers/snmp_managers.go` — `snmpManagerStore`, `RegisterSnmpManagerHandlers`, GET/POST/PATCH/DELETE for `/api/2.23/snmp-managers`, `stripSensitive()` helper

**Provider layer:**
- `internal/provider/snmp_manager_resource.go` — 4 interfaces, schema v0, 6 drift logs, sensitive write-once Read mapping, atomic v2c/v3 PATCH blocks, in-place version switch
- `internal/provider/snmp_manager_resource_test.go` — Lifecycle (create/update host/update notification/update auth_protocol/delete) + Import + DriftDetection
- `internal/provider/snmp_manager_data_source.go` — 2 interfaces, `name` Required + all-Computed schema, not-found via `AddError`
- `internal/provider/snmp_manager_data_source_test.go` — Basic seed + read + null-sensitive-fields assertion + not-found path
- `internal/provider/provider.go` — `NewSnmpManagerResource` and `NewSnmpManagerDataSource` registered under `Array administration`

**Examples + docs + roadmap:**
- `examples/resources/flashblade_snmp_manager/{resource.tf,import.sh}` — v3 primary + commented v2c + in-place switch note + import-by-name
- `examples/data-sources/flashblade_snmp_manager/data-source.tf` — DS by name + output
- `docs/resources/snmp_manager.md` + `docs/data-sources/snmp_manager.md` — generated by `make docs` (idempotent on second run)
- `ROADMAP.md` — row moved to `Array Administration / Implemented`, header counters bumped (`Provider version: v2.23.1`, 56 resources / 44 data sources)

## Decisions Made

- **`server.go` not touched**: the existing codebase registers handlers per-test via `ms.Mux` (e.g. `handlers.RegisterTargetHandlers(ms.Mux)`), not centrally in `NewMockServer()`. The plan's T03 Step 2 ("wire into server.go") was satisfied by following the convention — provider tests call `handlers.RegisterSnmpManagerHandlers(ms.Mux)` directly. See *Deviations* below.
- **Validator severity**: applied the stricter `SnmpV3Post` constraints (`LengthAtMost(32)` on `auth_passphrase`, `LengthBetween(8, 63)` on `privacy_passphrase`) at the **provider schema** level for both Create AND Update, even though the API only enforces them on POST. This gives operators predictable validation before PATCH, matching D-04.
- **Drift inlined, not helper-routed**: T05's first implementation routed drift through a `logDrift()` helper (single tflog call). Refactored to 6 inline `tflog.Debug` calls to satisfy the plan's exactly-6 requirement and to make each leaf trivially grep-able. The 3 v3 leaves are gated on `prevV3 != nil` / `mgr.V3 != nil` and degrade gracefully (`""` placeholders) when the block flips.
- **Pre-existing staticcheck QF1008 in test**: `model.Timeouts.Object.IsNull()` was flagged by `golangci-lint`; corrected to `model.Timeouts.IsNull()`. Bundled into the final docs commit (T13) rather than a fresh test commit since the fix is trivial.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Mock handler wiring follows codebase pattern, not plan literal text**

- **Found during:** T03 (mock handler implementation)
- **Issue:** Plan's T03 Step 2 said "Wire into `internal/testmock/server.go`". Inspection showed the existing codebase never wires resource handlers in `server.go` — `NewMockServer` only registers `/api/login` and `/api/api_version`. Per-resource handlers are registered by the calling test (see `target_resource_test.go:101`, `array_connection_resource_test.go`).
- **Fix:** Followed the established convention: tests call `handlers.RegisterSnmpManagerHandlers(ms.Mux)` themselves. The exported `RegisterSnmpManagerHandlers` returns the store pointer for `Seed`/`Get`/`Mutate` helpers — same shape as `RegisterTargetHandlers`.
- **Files modified:** none (server.go untouched)
- **Verification:** All 9 new tests pass; `make test` reports 816 ok.
- **Committed in:** `cbf400c` (T03 commit)

**2. [Rule 1 - Bug / Lint] staticcheck QF1008: embedded field selector**

- **Found during:** T13 (`make lint`)
- **Issue:** `model.Timeouts.Object.IsNull()` in `snmp_manager_resource_test.go:286` — `Object` is the embedded `basetypes.ObjectValue` and the selector can be removed per QF1008.
- **Fix:** Replaced with `model.Timeouts.IsNull()`.
- **Files modified:** `internal/provider/snmp_manager_resource_test.go`
- **Verification:** `make lint` reports `0 issues.`; tests still pass.
- **Committed in:** `24098d1` (T13 commit, bundled with docs/ROADMAP)

---

**Total deviations:** 2 auto-fixed (1 blocking convention-mismatch, 1 lint-correctness).
**Impact on plan:** Neither deviation altered the contract delivered. Pattern alignment with the rest of the codebase is preserved.

## Issues Encountered

- `basetypes` import missing on first compile of `snmp_manager_resource_test.go` — added `github.com/hashicorp/terraform-plugin-framework/types/basetypes` import; immediate fix.
- First implementation of drift used a helper function (1 literal `tflog.Debug` site) — refactored to 6 inline calls to match the plan's "exactly 6" verification regex. Caught before the T05 commit.
- `make docs` produces a noisy log (template generation for every existing resource) but is fully deterministic — second run leaves only the 2 new files untouched (`git diff --quiet docs/` of pre-existing files passes).

## User Setup Required

None — no external services involved. The resource is exercised end-to-end via mocked unit tests.

## Next Phase Readiness

- v2.23.1 milestone is **ready for tag + merge**: branch `implem-snmp-managers` is clean, 816 tests green, `make lint` zero issues, `make docs` idempotent, ROADMAP.md reflects new state.
- Future SNMP work (resource-action for `GET /snmp-managers/test`) tracked separately; do NOT confuse with this plan.
- Pattern established here (atomic *Patch nested block + sensitive write-once + per-leaf drift) is reusable for similar resources where the API never echoes secrets back (e.g. potential `KMIP` resource in `Medium Priority -- Admin and security`).

---
*Phase: 61-flashblade-snmp-manager*
*Completed: 2026-05-20*

## Self-Check: PASSED

- All 13 expected files present on disk.
- All 11 task commit hashes resolve in `git log`.
- `make build`, `make lint`, `make test` (816 ok), `make docs` (idempotent) all green at completion.
