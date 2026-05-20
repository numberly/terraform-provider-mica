# Milestone v2.23.1 Requirements — `flashblade_snmp_manager`

**Status:** 🚧 Active (planning)
**Started:** 2026-05-20
**Branch:** `implem-snmp-managers`
**API source:** `api_references/2.23.md` + `swagger-2.23.json` (`SnmpManager`, `SnmpManagerPost`, `_snmp_v2c`, `_snmp_v3`, `_snmp_v3_post`)

## Scope

Implement Terraform resource `flashblade_snmp_manager` (full CRUD) and matching data source against `/api/2.23/snmp-managers`, following the *New Resource* 16-item checklist in `CONVENTIONS.md` with **zero deviation**, driven by the `flashblade-resource-builder` skill.

## Out of Scope

- `GET /snmp-managers/test` — connectivity test (resource-action pattern, deferred to a dedicated milestone covering all `/{resource}/test` endpoints).
- Pulumi bridge regen / SDK publish — owned by a separate `pulumi-2.23.x` milestone.
- Bumping `TEST_BASELINE` in `GNUmakefile` — reserved for release milestones.

## Active Requirements

### Resource & Data Source

- [x] **SNMP-01** — Resource `flashblade_snmp_manager` implements full CRUD (Create / Read / Update / Delete) against `/api/2.23/snmp-managers` via the `flashblade-resource-builder` skill.
- [x] **SNMP-02** — Resource supports both SNMP protocol versions (`v2c`, `v3`) through nested config blocks, with enum validators on `notification` (`inform`|`trap`), `version` (`v2c`|`v3`), `v3.auth_protocol` (`MD5`|`SHA`), `v3.privacy_protocol` (`AES`|`DES`).
- [x] **SNMP-04** — Data source `flashblade_snmp_manager` (2 interfaces only: `DataSource`, `DataSourceWithConfigure`); `name` Required, all others Computed; not-found → `AddError`.

### Security & State

- [x] **SNMP-03** — Sensitive fields `v2c.community`, `v3.auth_passphrase`, `v3.privacy_passphrase` are marked `Sensitive: true`, treated **write-once** (API never returns them on GET → Read must not overwrite state), and null in ImportState.
- [x] **SNMP-05** — ImportState by `name` (`?names=`-based; never by UUID), uses `nullTimeoutsValue()`, sets all sensitive/write-once fields to null.
- [x] **SNMP-06** — Drift detection via `tflog.Debug(ctx, "drift detected", {"resource", "field", "was", "now"})` on every mutable/computed field (`host`, `notification`, `version`, `v2c.community`-presence, `v3.user`, `v3.auth_protocol`, `v3.privacy_protocol`).

### Test Infrastructure

- [x] **SNMP-07** — Mock handler `internal/testmock/handlers/snmp_managers.go` with `snmpManagerStore` (mutex + byName + nextID), `RegisterSnmpManagerHandlers(mux)` returning `*snmpManagerStore` for `Seed()`, and GET-with-no-match returning HTTP 200 + empty list (NOT 404). Uses shared helpers (`ValidateQueryParams`, `RequireQueryParam`, `WriteJSONListResponse`, `WriteJSONError`).
- [x] **SNMP-08** — At least **9 new** unit tests prefixed `TestUnit_`:
  - 5 client tests (`TestUnit_SnmpManager_Get_Found`, `_Get_NotFound`, `_Post`, `_Patch`, `_Delete`)
  - 3 resource tests (`TestUnit_SnmpManagerResource_Lifecycle`, `_Import`, `_DriftDetection`)
  - 1 data source test (`TestUnit_SnmpManagerDataSource_Basic`)

### Wiring & Documentation

- [x] **SNMP-09** — Resource and data source registered in `internal/provider/provider.go` (`NewSnmpManagerResource` in `Resources()`, `NewSnmpManagerDataSource` in `DataSources()`).
- [x] **SNMP-10** — Documentation regenerated via `make docs`; HCL examples present at `examples/resources/flashblade_snmp_manager/{resource.tf,import.sh}` and `examples/data-sources/flashblade_snmp_manager/data-source.tf`; `import.sh` uses `name` (not UUID).
- [x] **SNMP-11** — Repo-level `ROADMAP.md` row for `SNMP Managers` moved from *Medium Priority — Not Implemented* to *Array Administration / Implemented* (status `Done`, notes mention `v2.23.1; full CRUD; sensitive write-once community/passphrases`), counters + footer date/version refreshed, all in the **same commit** as the implementation.

### Quality Gates

- [x] **SNMP-12** — `make build && make test && make lint && make docs` all clean; total test count ≥ `TEST_BASELINE + 9` (≥ 816).
- [x] **SNMP-13** — Out-of-scope endpoints (`/snmp-managers/test`) documented in PROJECT.md as explicit deferral; no provider code references them in v2.23.1.

## Traceability

| Req ID    | Description (short)                                            | Phase | Status     |
| --------- | -------------------------------------------------------------- | ----- | ---------- |
| SNMP-01   | Resource + full CRUD via skill                                  | 61    | 🚧 planned |
| SNMP-02   | v2c/v3 support + enum validators                                | 61    | 🚧 planned |
| SNMP-03   | Sensitive write-once (community, passphrases)                   | 61    | 🚧 planned |
| SNMP-04   | Data source (lookup by name)                                    | 61    | 🚧 planned |
| SNMP-05   | ImportState by name                                             | 61    | 🚧 planned |
| SNMP-06   | Drift detection on all mutable/computed fields                  | 61    | 🚧 planned |
| SNMP-07   | Mock handler (Seed, empty-list GET=200, shared helpers)         | 61    | 🚧 planned |
| SNMP-08   | ≥ 9 new TestUnit_ tests (5 client + 3 resource + 1 DS)          | 61    | 🚧 planned |
| SNMP-09   | Registration in provider.go                                     | 61    | 🚧 planned |
| SNMP-10   | HCL examples + `make docs` regen                                | 61    | 🚧 planned |
| SNMP-11   | Repo-level ROADMAP.md row moved to Done (same commit as impl.)  | 61    | 🚧 planned |
| SNMP-12   | `make build && test && lint && docs` clean; count ≥ 816         | 61    | 🚧 planned |
| SNMP-13   | `/snmp-managers/test` explicitly OOS                            | 61    | 🚧 planned |

**Totals:** 13 active · 0 satisfied · 0 deferred.

## Notes

- Requirement IDs use a single prefix `SNMP-` for resource scope plus quality gates (no separate `QA-`/`DOC-` split — milestone is small and cohesive).
- Validators choose the **stricter** `_snmp_v3_post` constraints (privacy_passphrase 8-63, auth_passphrase ≤ 32, community ≤ 32) for safer UX on both POST and PATCH paths.
- No cross-field validator between `version` and the presence of `v2c`/`v3` blocks — defer to server-side validation to stay aligned with `CONVENTIONS.md` ("let the server validate").
- Domain placement in `internal/client/models_admin.go` (alongside `SmtpServer`, `SyslogServer`, `AlertWatcher`) confirmed via `mcp__serena__get_symbols_overview`.

---

_Updated 2026-05-20 on milestone v2.23.1 start._
