# Phase 61: `flashblade_snmp_manager` Resource & Data Source — Context

**Gathered:** 2026-05-20
**Status:** Ready for planning

<domain>
## Phase Boundary

Deliver Terraform resource `flashblade_snmp_manager` (full CRUD) and matching data source against `/api/2.23/snmp-managers`, plus mock handler, ≥9 new `TestUnit_` tests, HCL examples, regenerated docs, and the repo-level `ROADMAP.md` row move — all following the *New Resource* 16-item checklist of `CONVENTIONS.md` with zero deviation, driven by the `flashblade-resource-builder` skill.

**In scope:** GET / POST / PATCH / DELETE on `/api/2.23/snmp-managers`, with nested `v2c` (community) and `v3` (user, auth_protocol/auth_passphrase, privacy_protocol/privacy_passphrase) config blocks.

**Out of scope:** `GET /snmp-managers/test` connectivity check (resource-action pattern, deferred), Pulumi bridge regen, `TEST_BASELINE` bump in `GNUmakefile`.

</domain>

<decisions>
## Implementation Decisions

### Plan Granularity

- **D-01:** **One monolithic plan** `61-01-implement-snmp-manager-PLAN.md` covering the 16-item *New Resource* checklist. Aligns with `coarse` granularity from `.planning/config.json` and matches the volume-per-plan baseline of v2.22.1 and v2.22.2 (each comparable resource shipped as a small handful of plans, but in this case the total work fits cleanly in one plan with clear sequential sub-steps).

### Schema Shape

- **D-02:** Resource exposes `v2c` and `v3` as `schema.SingleNestedAttribute{ Optional: true, Computed: true, Attributes: ... }`. Pattern confirmed via `internal/provider/array_connection_resource.go:121-141` (`throttle` nested attribute on `flashblade_array_connection`). HCL form: `v3 = { user = "...", auth_protocol = "MD5", ... }`.
- **D-03:** Three client model structs in `internal/client/models_admin.go`:
  - `SnmpManager` (GET) — `ID`, `Name`, `Host`, `Notification`, `Version`, `V2c *SnmpV2c`, `V3 *SnmpV3`
  - `SnmpManagerPost` (POST) — same fields minus `ID`/`Name` (Name carried via `?names=`), with `V3 *SnmpV3Post` (stricter `_snmp_v3_post` constraints)
  - `SnmpManagerPatch` (PATCH) — pointers everywhere; nested blocks atomic via `*SnmpV2c` / `*SnmpV3` (pattern from `ArrayConnectionPatch.Throttle *ArrayConnectionThrottle` in `models_admin.go:141-146`)
- **D-04:** Enum validators from `stringvalidator`:
  - `notification`: `OneOf("inform", "trap")`
  - `version`: `OneOf("v2c", "v3")` — Required
  - `v3.auth_protocol`: `OneOf("MD5", "SHA")`
  - `v3.privacy_protocol`: `OneOf("AES", "DES")`
  - `v2c.community`: `LengthAtMost(32)`
  - `v3.auth_passphrase`: `LengthAtMost(32)` (POST constraint)
  - `v3.privacy_passphrase`: `LengthBetween(8, 63)` (POST constraint applied on both POST and PATCH paths — stricter UX)
- **D-05:** No cross-field validator between `version` and the presence of `v2c`/`v3` blocks. Server-side validation only (aligns with CONVENTIONS "let the server validate").

### Version Switch Behavior (v2c ↔ v3)

- **D-06:** Provider does NOT force replace on `version` change. On Update, send the new block + new `version` and **omit** the unused block (no explicit `null`). If the real API does not clear the previously-active block on its own, the resulting drift will surface in `_DriftDetection` test or live UAT and will be addressed as a follow-up (potentially elevating `version` to `RequiresReplace` in a later patch). Plan modifier on `version`: **none** (mutable in-place by default).
- **D-07:** HCL `resource.tf` example documents this behaviour explicitly: a comment block notes that switching SNMP versions in-place is permitted; if drift appears on the unused block, document the workaround (taint + apply or update `version` block via `terraform state rm`).

### Sensitive / Write-Once

- **D-08:** `v2c.community`, `v3.auth_passphrase`, `v3.privacy_passphrase` marked `Sensitive: true`. `Read()` never assigns to these from API response (API does not return them); state value is preserved verbatim (pattern from `mapDirectoryServiceToModel` in `internal/provider/directory_service_management_resource.go:467-517` which skips `BindPassword`).
- **D-09:** `ImportState` sets the three sensitive fields to null (`types.StringNull()`) inside their nested blocks (`v2c = { community = null }`, `v3 = { auth_passphrase = null, privacy_passphrase = null }`). User must re-supply them on next apply or accept the drift (documented in `import.sh`).

### Drift Detection

- **D-10:** Per-leaf `tflog.Debug(ctx, "drift detected", { resource, field, was, now })` calls in `Read()` for **6 fields**:
  1. `host`
  2. `notification`
  3. `version`
  4. `v3.user`
  5. `v3.auth_protocol`
  6. `v3.privacy_protocol`
  The three sensitive write-once fields are excluded (API never returns them → no value to compare). Pattern from `directory_service_management_resource.go` (10 per-leaf drift calls).

### Mock Handler

- **D-11:** `internal/testmock/handlers/snmp_managers.go` with `snmpManagerStore` (mutex + `byName map[string]*client.SnmpManager` + `nextID int`). `RegisterSnmpManagerHandlers(mux *http.ServeMux) *snmpManagerStore` returns the store so tests can call `Seed(...)`. GET-with-no-`?names=`-match → HTTP 200 + empty list (CRITICAL — not 404). Shared helpers `ValidateQueryParams`, `RequireQueryParam`, `WriteJSONListResponse`, `WriteJSONError`.
- **D-12:** Sensitive fields in the mock store: passphrases and community are NOT echoed in GET responses (mirror real API); they ARE accepted on POST/PATCH so client tests can verify the request body went out correctly.

### Tests (≥ 9 new, prefix `TestUnit_`)

- **D-13:** Client (5 tests): `TestUnit_SnmpManager_Get_Found`, `_Get_NotFound`, `_Post`, `_Patch`, `_Delete`.
- **D-14:** Resource (3 tests): `TestUnit_SnmpManagerResource_Lifecycle`, `_Import`, `_DriftDetection`. `_Lifecycle` covers Create with v3 → Update host → Update notification → Delete. `_DriftDetection` verifies the 6 leaf drift logs fire.
- **D-15:** Data source (1 test): `TestUnit_SnmpManagerDataSource_Basic`.

### Wiring, Docs, Roadmap

- **D-16:** Register `NewSnmpManagerResource` in `provider.go` `Resources()` and `NewSnmpManagerDataSource` in `DataSources()`.
- **D-17:** HCL examples cover both v2c and v3 variants. `examples/resources/flashblade_snmp_manager/resource.tf` shows v3 (richer); a commented snippet shows v2c. `import.sh` imports by name.
- **D-18:** Repo-level `ROADMAP.md` (project root, not `.planning/ROADMAP.md`) row `SNMP Managers` moved from *Medium Priority — Not Implemented* (line 145) to *Array Administration / Implemented* with `Done`, `Yes` data source, notes: `v2.23.1; full CRUD; sensitive write-once community/passphrases; /test endpoint deferred`. Counters + footer date + provider version refreshed in the **same commit** as the implementation.

### Process

- **D-19:** All commits use `git commit --no-verify`. No `Co-Authored-By` trailer. Per the project `CLAUDE.md`.
- **D-20:** Branch: `implem-snmp-managers` from clean `main`. Create at the start of plan execution, not during this discussion.

### Claude's Discretion

- Exact wording of HCL example comments and drift-log keys (just match the `{ resource, field, was, now }` map convention).
- Choice of `Seed()` signature (variadic vs slice) — match the closest existing handler (`alert_watchers` or `syslog_servers`).
- Whether to include a `display_name` / `description` field beyond what the swagger defines — **no**, stick to the swagger.

### Folded Todos

_None — `gsd-tools todo match-phase 61` returned 0 matches._

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project Conventions (mandatory)

- `CONVENTIONS.md` — full file, especially the *New Resource* checklist (16 items) and the *Test Coverage* / *Test Conventions* tables. **Authoritative — zero deviation.**
- `CLAUDE.md` — Project instructions (commit policy, Serena requirement, `--no-verify`).

### API Source

- `api_references/2.23.md` §`Snmp managers` (lines 1010-1016) — endpoint list (the body params inlined there do NOT detail `v2c`/`v3`; fall back to swagger).
- `swagger-2.23.json` — schemas `SnmpManager`, `SnmpManagerPost`, `_snmp_v2c`, `_snmp_v3`, `_snmp_v3_post` (authoritative for nested fields).

### Code Patterns to Reuse

- `internal/client/targets.go` — canonical example of `getOneByName[T]` usage.
- `internal/client/models_admin.go:104-146` — `ArrayConnection` / `ArrayConnectionPatch` / `ArrayConnectionThrottle` — pattern for **atomic nested config block** in Patch (`*ArrayConnectionThrottle`).
- `internal/provider/array_connection_resource.go:76-166` — pattern for **`SingleNestedAttribute`** with `Optional: true, Computed: true`.
- `internal/provider/directory_service_management_resource.go:467-517` (`mapDirectoryServiceToModel`) — pattern for **never touching sensitive write-once fields in `Read()`**.
- `internal/testmock/handlers/targets.go` — canonical example of a mock handler with Seed + empty-list GET=200.

### Skill

- `.claude/skills/flashblade-resource-builder/` — must be loaded and followed for the lifecycle (models → client → mocks → tests → resource → DS → docs).

### Roadmap (where to update)

- `ROADMAP.md` (project root) §*Array Administration / Implemented* (table line ~94-104) and §*Medium Priority — Not Implemented* (line ~145 — remove SNMP row).

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets

- **`getOneByName[T]` generic** (`internal/client/client.go`) — used for every GET-single in this codebase; do NOT hand-roll list-then-filter logic.
- **`*ArrayConnectionThrottle` atomic nested Patch pattern** (`models_admin.go:141-146`) — direct template for `SnmpManagerPatch.V2c *SnmpV2c` and `SnmpManagerPatch.V3 *SnmpV3`.
- **`schema.SingleNestedAttribute` with `Optional+Computed`** (`array_connection_resource.go:121-141`) — direct template for the `v2c` and `v3` attributes in the resource schema.
- **`mapDirectoryServiceToModel` write-once skipping pattern** (`directory_service_management_resource.go:467-517`) — direct template for how `Read()` must avoid touching `community` / `auth_passphrase` / `privacy_passphrase`.
- **Shared mock helpers** in `internal/testmock/handlers/helpers.go` — `ValidateQueryParams`, `RequireQueryParam`, `WriteJSONListResponse`, `WriteJSONError`.

### Established Patterns

- **Models domain placement**: SNMP belongs in `models_admin.go` alongside `SmtpServer`, `SyslogServer`, `AlertWatcher` (notifications / array admin domain). Confirmed via `mcp__serena__get_symbols_overview`.
- **Schema versioning**: Start at `Version: 0`. No `UpgradeState` migration entries yet (new resource).
- **Plan modifiers**: `id` → `UseStateForUnknown()`; `name` → `RequiresReplace()`; everything else → none (especially nothing on `version`, `host`, `notification` per **D-06**).
- **Timeouts**: 20m Create, 5m Read, 20m Update, 30m Delete (defaults).
- **Drift detection**: log-only (`tflog.Debug`), never error.

### Integration Points

- **`internal/provider/provider.go`** — append `NewSnmpManagerResource` to `Resources()` and `NewSnmpManagerDataSource` to `DataSources()`.
- **`internal/testmock/server.go`** — register the new handler set; ensure provider tests can call `testNewMockedProvider()` and reach the handler.
- **`examples/`** — new dirs `examples/resources/flashblade_snmp_manager/` and `examples/data-sources/flashblade_snmp_manager/`; `make docs` will pick these up.
- **`docs/`** — auto-generated by `tfplugindocs` via `make docs`; never edit by hand.

</code_context>

<specifics>
## Specific Ideas

- User explicitly invoked the `flashblade-resource-builder` skill for this work — it must orchestrate the model → client → mocks → tests → resource → DS → docs chain.
- User explicitly invoked the `swagger-to-reference` skill to validate the API. The actual schemas for `_snmp_v2c` / `_snmp_v3` / `_snmp_v3_post` came from the raw `swagger-2.23.json` because the markdown reference does not expand nested objects.
- "Zero déviation" from `CONVENTIONS.md` was explicit — this CONTEXT.md is a refinement of the locked plan, not a re-negotiation.
- Pre-check performed via Serena: zero `Snmp*` / `snmp_manager` collision in the codebase. Implementation is greenfield.

</specifics>

<deferred>
## Deferred Ideas

### From this milestone scope

- **`GET /snmp-managers/test` connectivity check** — pattern *resource action*. Belongs to a dedicated milestone covering all `/{resource}/test` endpoints (`/dns/test`, `/smtp/test`, `/array/eula`, etc.). Not started.
- **`flashblade_snmp_agent` resource** — `/api/2.23/snmp-agents` is a singleton PATCH-only (GET + PATCH only). Similar shape (`v2c`, `v3` blocks) but different lifecycle. Worth a separate milestone right after this one (could reuse `SnmpV2c` / `SnmpV3` models).
- **Pulumi bridge regen for `flashblade_snmp_manager`** — owned by `pulumi-2.23.x` milestone.
- **`TEST_BASELINE` bump** in `GNUmakefile` — reserved for release milestones, not feature milestones.

### Reviewed Todos (not folded)

_None — no project todos matched Phase 61._

</deferred>

---

*Phase: 61-flashblade-snmp-manager*
*Context gathered: 2026-05-20*
