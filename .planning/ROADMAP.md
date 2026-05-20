# Roadmap: Terraform Provider FlashBlade

## Milestones

- ✅ **v1.0** — Core Provider (shipped 2026-03-28)
- ✅ **v1.1** — Servers & Exports (shipped 2026-03-28)
- ✅ **v1.2** — Code Quality & Robustness (shipped 2026-03-29)
- ✅ **v1.3** — Release Readiness (shipped 2026-03-29)
- ✅ **v2.0** — Cross-Array Bucket Replication (shipped 2026-03-29)
- ✅ **v2.0.1** — Quality & Hardening (shipped 2026-03-30)
- ✅ **v2.1** — Bucket Advanced Features (shipped 2026-03-30)
- ✅ **v2.1.1** — Network Interfaces (VIPs) (shipped 2026-03-31)
- ✅ **v2.1.3** — Code Review Fixes & S3 Users (shipped 2026-04-02)
- ✅ **v2.2** — S3 Target Replication & Security (shipped 2026-04-14)
- ✅ **tools-v1.0** — API Tooling Pipeline (shipped 2026-04-14)
- ✅ **v2.22.1** — Directory Service Array Management (shipped 2026-04-17)
- ✅ **v2.22.2** — DS Roles & Role Mappings (shipped 2026-04-17)
- ✅ **v2.22.3** — Convention Compliance (shipped 2026-04-20)
- ✅ **pulumi-2.22.3** — Pulumi Bridge Alpha (shipped 2026-04-24)
- ✅ **v2.23.0** — FlashBlade API 2.23 Upgrade (shipped 2026-05-20)
- 🚧 **v2.23.1** — `flashblade_snmp_manager` resource & data source (in planning, started 2026-05-20)

See `.planning/MILESTONES.md` for milestone details and `.planning/milestones/` for per-milestone roadmap + requirements archives.

---

## Current State — v2.23.1

**Goal:** Ship `flashblade_snmp_manager` resource + data source (full CRUD on `/api/2.23/snmp-managers`) driven by the `flashblade-resource-builder` skill with zero deviation from `CONVENTIONS.md`.

**Branch:** `implem-snmp-managers` (from clean `main`)
**Requirements:** see `.planning/REQUIREMENTS.md` (SNMP-01..13, 13 active)

### Phase Map

| #  | Phase | Goal | Requirements | Success Criteria |
|----|-------|------|--------------|------------------|
| 61 | `flashblade_snmp_manager` Resource & Data Source | Deliver resource + DS satisfying the *New Resource* 16-item checklist | SNMP-01..13 | 5 (see below) |

### Phase 61: `flashblade_snmp_manager` Resource & Data Source

**Goal:** Implement Terraform resource `flashblade_snmp_manager` and matching data source against `/api/2.23/snmp-managers`, including 3 model structs (Get/Post/Patch + nested `v2c`/`v3`), client CRUD via `getOneByName[T]`, mock handler with empty-list GET=200, ≥9 new `TestUnit_` tests, HCL examples, regenerated docs, and the repo-level `ROADMAP.md` row move — all in the strict order of the *New Resource* checklist in `CONVENTIONS.md`.

**Requirements covered:** SNMP-01, SNMP-02, SNMP-03, SNMP-04, SNMP-05, SNMP-06, SNMP-07, SNMP-08, SNMP-09, SNMP-10, SNMP-11, SNMP-12, SNMP-13 (13 total).

**Success criteria:**
1. Provider compiles (`make build`), all linters pass (`make lint`), full test suite passes (`make test`) with total count ≥ `TEST_BASELINE + 9` (≥ 816).
2. `make docs` regenerates `docs/resources/snmp_manager.md` and `docs/data-sources/snmp_manager.md` with no manual edits and no diff on re-run.
3. Local Terraform plan/apply against the mock array (via `internal/testmock`) succeeds for the Create / Read / Update / Delete / Import path.
4. Repo-level `ROADMAP.md` row `SNMP Managers` moved from *Medium Priority — Not Implemented* to *Array Administration / Implemented* (status `Done`, version `v2.23.1`), counters and footer date refreshed, change present in the same commit as the implementation.
5. Sensitive fields (`v2c.community`, `v3.auth_passphrase`, `v3.privacy_passphrase`) confirmed never logged (review `tflog.*` calls) and never overwritten in `Read()` from a missing API field (write-once verified by `_Import` and `_DriftDetection` tests).

**Out of scope (carried from milestone scope):**
- `GET /snmp-managers/test` endpoint (deferred to future `/{resource}/test` resource-action milestone).
- Pulumi bridge regen.
- `TEST_BASELINE` bump in `GNUmakefile` (release-only).

<details>
<summary>Archived: v2.23.0 — FlashBlade API 2.23 Upgrade (shipped 2026-05-20)</summary>

Full archive: [`milestones/v2.23.0-ROADMAP.md`](milestones/v2.23.0-ROADMAP.md) · [`milestones/v2.23.0-REQUIREMENTS.md`](milestones/v2.23.0-REQUIREMENTS.md)

2 phases (59 + 60), 10 plans, 33/33 requirements satisfied. Squash commit `3fd485d`, tag `v2.23.0`, [GitHub Release](https://github.com/numberly/terraform-provider-mica/releases/tag/v2.23.0).

</details>

---

*Last updated: 2026-05-20 — milestone v2.23.1 started (`flashblade_snmp_manager` resource + data source, 1 phase, 13 requirements). Run `/gsd:plan-phase 61` (or `/gsd:discuss-phase 61` first) to begin.*
