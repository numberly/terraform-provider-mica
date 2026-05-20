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
- 🚧 **v2.23.0** — FlashBlade API 2.23 Upgrade (in progress)

See `.planning/MILESTONES.md` for milestone details and `.planning/milestones/` for per-milestone roadmap + requirements archives.

---

## Active Milestone: v2.23.0 — FlashBlade API 2.23 Upgrade

**Working branch:** `test/api-upgrade-2.23`
**Granularity:** coarse
**Total requirements:** 33 (19 retro / already implemented + 14 active)
**Coverage:** 33/33 ✓

### Phases

- [x] **Phase 59: API 2.23 Consolidation & Validation** — Retro-document the API 2.23 work already on branch, then validate the whole upgrade clean (test/lint/docs/acceptance/bridge). (completed 2026-05-20)
- [ ] **Phase 60: v2.23.0 Release** — Rebase, merge, tag, publish artefacts, archive milestone.

### Phase Details

#### Phase 59: API 2.23 Consolidation & Validation

**Goal**: API 2.23 upgrade (version bump, workload + resiliency resources, schema v1 migrations, Pulumi bridge regen) is fully documented, traceable, and validated end-to-end against real arrays — provider is release-ready on `test/api-upgrade-2.23`.

**Depends on**: Nothing (branch already carries the implementation)

**Requirements** (retro / already implemented — 19):
- API-01, API-02, API-03
- WORKLOAD-01, WORKLOAD-02, WORKLOAD-03
- RESILIENCY-01, RESILIENCY-02, RESILIENCY-03
- SCHEMA-01, SCHEMA-02, SCHEMA-03, SCHEMA-04, SCHEMA-05, SCHEMA-06, SCHEMA-07, SCHEMA-08
- BRIDGE-01, BRIDGE-02, BRIDGE-03

**Requirements** (active work — 6):
- VALID-01, VALID-02, VALID-03, VALID-04, VALID-05, VALID-06

**Success Criteria** (what must be TRUE):
  1. `make test` passes on `test/api-upgrade-2.23` with count ≥ updated `TEST_BASELINE` (covers VALID-01)
  2. `make lint` returns 0 issues and `make docs` regenerates without manual edits (covers VALID-02, VALID-03)
  3. Acceptance tests run live against par5 + pa7 on the new resources (workload, resiliency_group, resiliency_group_member) and all 6 migrated schemas, with no drift on re-apply (covers VALID-04)
  4. `make tfgen` followed by `git diff --exit-code` is clean — no Pulumi schema / bridge-metadata drift (covers VALID-05)
  5. ROADMAP.md coverage counters and footer reflect API 2.23 deltas and `v2.23.0` as the provider version (covers VALID-06)

**Plans:** 4/5 plans complete
- [x] 59-01-retro-traceability-PLAN.md — Trace 19 retro REQ-IDs against branch state (RETRO.md)
- [x] 59-02-make-checks-PLAN.md — Run make test/lint/docs and capture outputs (CHECKS.md)
- [x] 59-03-pulumi-bridge-check-PLAN.md — Run make tfgen, verify clean diff on pulumi/
- [ ] 59-04-acceptance-live-PLAN.md — Live acceptance against par5 + pa7 (HCL fixtures + ACCEPTANCE.md, human-gated)
- [x] 59-05-roadmap-counters-PLAN.md — Refresh root ROADMAP.md counters / version / API version

#### Phase 60: v2.23.0 Release

**Goal**: `v2.23.0` is shipped — branch merged to `main`, tagged, GoReleaser artefacts published, baseline bumped, milestone archived.

**Depends on**: Phase 59

**Requirements** (active work — 7):
- RELEASE-01, RELEASE-02, RELEASE-03, RELEASE-04, RELEASE-05, RELEASE-06, RELEASE-07

**Success Criteria** (what must be TRUE):
  1. CHANGELOG.md has a `v2.23.0` entry covering new resources, schema migrations, and Pulumi bridge regen (covers RELEASE-01)
  2. `test/api-upgrade-2.23` is rebased on `main`, PR opens green, and is merged (covers RELEASE-02, RELEASE-03, RELEASE-04)
  3. Tag `v2.23.0` exists on `main` and GoReleaser has published signed release artefacts (covers RELEASE-05)
  4. `TEST_BASELINE` in `GNUmakefile` is bumped to the new post-merge test count (covers RELEASE-06)
  5. Milestone v2.23.0 is archived under `.planning/milestones/` via `/gsd:complete-milestone` (covers RELEASE-07)

**Plans**: 5 plans
- [x] 60-01-preflight-and-changelog-PLAN.md — Pre-flight gate (VALID-04 resolution) + v2.23.0 CHANGELOG entry
- [x] 60-02-rebase-and-pr-PLAN.md — Rebase test/api-upgrade-2.23 on main, push, open PR, wait for CI green
- [ ] 60-03-merge-PLAN.md — Merge PR into main
- [ ] 60-04-tag-and-release-PLAN.md — Tag v2.23.0, push, verify GoReleaser publication
- [ ] 60-05-baseline-and-archive-PLAN.md — Bump TEST_BASELINE + archive milestone via /gsd:complete-milestone

### Progress Table

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 59. API 2.23 Consolidation & Validation | 4/5 | Complete    | 2026-05-20 |
| 60. v2.23.0 Release | 2/5 | In Progress|  |

---

*Last updated: 2026-05-20 — Phase 60 plans authored (5 plans, linear waves 1-5).*
