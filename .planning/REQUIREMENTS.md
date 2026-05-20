# Requirements — Milestone v2.23.0

**Milestone:** v2.23.0 — FlashBlade API 2.23 Upgrade
**Status:** In progress (retro + finalisation)
**Working branch:** `test/api-upgrade-2.23`

## Context

Most of the implementation has already landed on `test/api-upgrade-2.23` (≈ 167 fichiers, ~7000 insertions vs `main`). This milestone formalizes that work into GSD and adds the validation + release tail to ship `v2.23.0`.

## v2.23.0 Requirements

### API — version bump

- [x] **API-01** — Provider, client, mock handlers, examples and API references reference API version `2.23` (commit `9a9715b` and follow-ups)
- [x] **API-02** — Stale `2.22` strings replaced in test messages and handler comments (commit `bfe0ad0`)
- [x] **API-03** — `upgrade_version.py` scans all `internal/**/*.go` (commit `19ea8dc`)

### WORKLOAD — new resource + data source

- [x] **WORKLOAD-01** — `flashblade_workload` resource (CRUD + drift + import) implemented
- [x] **WORKLOAD-02** — `flashblade_workload` data source implemented
- [x] **WORKLOAD-03** — Client CRUD, mock handler, tests for workload

### RESILIENCY — read-only data sources

- [x] **RESILIENCY-01** — `flashblade_resiliency_group` data source (DS-only) implemented
- [x] **RESILIENCY-02** — `flashblade_resiliency_group_member` data source (DS-only) implemented
- [x] **RESILIENCY-03** — Mock handlers + tests for resiliency_groups and resiliency_group_members

### SCHEMA — workload field on existing resources (schema v1)

- [x] **SCHEMA-01** — `flashblade_file_system` workload field + schema v0→v1 upgrader + test
- [x] **SCHEMA-02** — `flashblade_file_system_export` computed workload field + schema v0→v1 upgrader + test
- [x] **SCHEMA-03** — `flashblade_nfs_export_policy` computed workload field + schema v0→v1 upgrader + test
- [x] **SCHEMA-04** — `flashblade_smb_client_policy` computed workload field + schema v0→v1 upgrader + test
- [x] **SCHEMA-05** — `flashblade_smb_share_policy` computed workload field + schema v0→v1 upgrader + test
- [x] **SCHEMA-06** — `flashblade_qos_policy` computed workload field + schema v0→v1 upgrader + test
- [x] **SCHEMA-07** — `flashblade_qos_policy` computed `context` field for API 2.23
- [x] **SCHEMA-08** — qos_policy upgrader aligned on canonical chain pattern (commit `5ab258a`)

### BRIDGE — Pulumi bridge alignment

- [x] **BRIDGE-01** — Pulumi `schema.json`, `bridge-metadata.json`, `schema-embed.json` regenerated for API 2.23 additions (commit `c65d063`)
- [x] **BRIDGE-02** — `resources_test.go` ProviderInfo expectations bumped (Resources / DataSources / Schema hash) (commit `d216d24`)
- [x] **BRIDGE-03** — api-upgrade Phase 6 documents `make tfgen` workflow (commits `907623d`, `de2697a`)

### VALIDATION — finalisation (to be done in Phase 59)

- [ ] **VALID-01** — `make test` passes on `test/api-upgrade-2.23`, test count ≥ updated `TEST_BASELINE`
- [ ] **VALID-02** — `make lint` clean
- [ ] **VALID-03** — `make docs` regenerated; `docs/` diff committed
- [ ] **VALID-04** — Acceptance tests run live against real FlashBlade (par5 + pa7) on the new resources (workload, resiliency_group, resiliency_group_member) and on the migrated schemas
- [ ] **VALID-05** — `git diff --exit-code` clean after `make tfgen` (no Pulumi schema drift)
- [ ] **VALID-06** — ROADMAP.md counters (covered / coverage %) refreshed; footer date + provider version bumped to `v2.23.0`

### RELEASE — ship v2.23.0 (to be done in Phase 60)

- [ ] **RELEASE-01** — CHANGELOG entry for `v2.23.0` (added resources, schema migrations, bridge regen)
- [ ] **RELEASE-02** — Rebase `test/api-upgrade-2.23` on `main` (cleanup)
- [ ] **RELEASE-03** — Open PR `test/api-upgrade-2.23` → `main`, pass CI
- [ ] **RELEASE-04** — Merge PR
- [ ] **RELEASE-05** — Tag `v2.23.0` + GoReleaser release artifacts
- [ ] **RELEASE-06** — Update `TEST_BASELINE` in `GNUmakefile` to the new count
- [ ] **RELEASE-07** — Archive milestone via `/gsd:complete-milestone`

## Future Requirements

Deferred to later milestones:

- Workload metadata, label, and tag management resources (if/when API exposes them as CRUD)
- Resiliency group CRUD (currently read-only — no POST/PATCH/DELETE in API 2.23)
- Acceptance tests integrated into CI (currently manual run against par5 / pa7)
- Pulumi `pulumi-2.23.0` SDK regeneration + private release (separate milestone)

## Out of Scope

- Pulumi SDK publication for v2.23.0 — handled in a dedicated `pulumi-2.23.0` milestone
- API 2.24+ features — out of scope for v2.23.0
- Refactors not driven by API 2.23 deltas (e.g. cross-resource helper consolidation)

## Traceability

_Filled by gsd-roadmapper after roadmap creation._
