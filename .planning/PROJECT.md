# Terraform Provider FlashBlade

## Current State

**Latest shipped:** v2.23.0 (FlashBlade API 2.23 Upgrade) — 2026-05-20
**Active milestone:** v2.23.1 — `flashblade_snmp_manager` resource & data source (phase 61 complete, ready to ship)

**Shipped to date:** 16 milestones, 61 phases
**TF Provider:** v2.23.0 (55 resources + 43 data sources, 807 tests, [GitHub Release](https://github.com/numberly/terraform-provider-mica/releases/tag/v2.23.0)) — pending v2.23.1: 56 resources + 44 data sources, 816 tests on branch `implem-snmp-managers`
**Pulumi Bridge:** pulumi-2.22.3 alpha (private distribution via GitHub Releases, Python + Go SDKs) — bridge schema regen'd for API 2.23 but no new Pulumi release yet

## Current Milestone: v2.23.1 `flashblade_snmp_manager`

**Goal:** Add a `flashblade_snmp_manager` Terraform resource + data source for `/api/2.23/snmp-managers` (full CRUD), driven by the `flashblade-resource-builder` skill and zero deviation from `CONVENTIONS.md`.

**Target features:**
- `flashblade_snmp_manager` resource — full CRUD (Create/Read/Update/Delete) against `/api/2.23/snmp-managers`
- `flashblade_snmp_manager` data source — lookup by `name`
- Nested config for SNMP `v2c` (community) and `v3` (user, auth_protocol/auth_passphrase, privacy_protocol/privacy_passphrase) with enum validators
- Sensitive write-once handling on `community`, `auth_passphrase`, `privacy_passphrase` (API never returns them on GET)
- Mock handler with Seed + empty-list GET=200, ≥9 new tests prefixed `TestUnit_`
- ImportState by `name` (never by UUID)
- HCL examples + `make docs` regenerated
- Repo-level `ROADMAP.md` row moved from *Medium Priority — Not Implemented* to *Array Administration / Implemented* in the same commit as the implementation

**Key context:**
- Branch: `implem-snmp-managers` (from clean `main`)
- API source: `api_references/2.23.md` + `swagger-2.23.json` (`SnmpManager`, `SnmpManagerPost`, `_snmp_v2c`, `_snmp_v3`, `_snmp_v3_post`)
- Pre-check (Serena `find_symbol`): no existing `Snmp*` / `snmp_manager` collision
- Domain home: `internal/client/models_admin.go` (alongside `SmtpServer`, `SyslogServer`, `AlertWatcher`)
- `TEST_BASELINE` (GNUmakefile) = 807 → expected ≥ 816 post-implementation (do **not** bump baseline in this milestone — reserved for release milestones)
- Out of scope: `GET /snmp-managers/test` (resource-action pattern, deferred), Pulumi bridge regen (separate `pulumi-2.23.x` milestone)

## Last Completed Milestone: v2.23.0 — FlashBlade API 2.23 Upgrade (shipped 2026-05-20)

**Goal:** Aligner le provider sur l'API FlashBlade 2.23, ajouter le support des Workloads et Resiliency Groups, puis livrer la release (validation, docs, tag, merge).

**Target features (déjà implémentés sur `test/api-upgrade-2.23`):**
- API version bump 2.22 → 2.23 (provider, client, mock, examples, references)
- `flashblade_workload` resource + data source
- `flashblade_resiliency_group` data source (DS-only)
- `flashblade_resiliency_group_member` data source (DS-only)
- Schéma v1 (workload field) sur 6 ressources : file_system, file_system_export, nfs_export_policy, smb_client_policy, smb_share_policy, qos_policy
- qos_policy : computed `context` field
- Pulumi bridge regen (schema.json, bridge-metadata.json, schema-embed.json)
- api-diff / api-upgrade skills enhancements (per-field variants, codebase scan, BLOCKING method detection)

**Target features (à livrer dans la finalisation):**
- Validation `make test` + `make lint` + `make docs` clean sur la branche
- Acceptance tests live FlashBlade (par5, pa7) sur les nouvelles ressources et les schémas migrés
- CHANGELOG + release notes v2.23.0
- ROADMAP.md fix-up (coverage counters, version footer)
- Tag `v2.23.0` + merge `test/api-upgrade-2.23` → `main`

**Key context:**
- Travail rétro : ~167 fichiers / ~7000 insertions déjà sur la branche, piloté par les skills `api-diff` et `api-upgrade`
- 2 phases prévues : Phase 59 (consolidation rétro + validation), Phase 60 (release & merge)
- TEST_BASELINE actuel (GNUmakefile) : 807 — à mettre à jour quand v2.23.0 ship

## What This Is

A Terraform provider for Pure Storage FlashBlade that enables operational teams to manage storage infrastructure as code — file systems, object stores, buckets, policies, and array administration. Built with terraform-plugin-framework for maximum reliability, targeting the FlashBlade REST API v2.22 (Purity//FB 4.6.7). Designed for high-frequency CRUD operations with robust drift detection and audit logging.

## Core Value

Operational teams can reliably create, update, delete, and reconcile drift on FlashBlade storage resources (buckets, file systems, policies) through Terraform with zero surprises — every plan reflects reality, every apply converges.

## Last Completed Milestone: pulumi-2.22.3 — Pulumi Bridge Alpha (shipped 2026-04-24)

**Goal:** Expose the FlashBlade Terraform provider to Pulumi users (Python + Go) via the official `pulumi/pulumi-terraform-bridge` (`pkg/pf/*` for terraform-plugin-framework), in a new `./pulumi/` sub-directory with its own `go.mod`, distributed privately through GitHub releases.

**Target features:**
- Pulumi bridge scaffold in `./pulumi/` (tfgen + runtime binaries, ProviderInfo, embedded schema)
- Mapping of all 28 resources + 21 data sources (auto-tokenization + targeted overrides for composite IDs, secrets, timeouts)
- Python and Go SDK generation (embedded `schema.json` + `bridge-metadata.json`)
- ProgramTest coverage on 3 representative resources (target, remote_credentials, bucket)
- Private release pipeline: GitHub Actions build + goreleaser + cosign, tag `pulumi-2.22.3`
- Auto-converted HCL examples (`PULUMI_CONVERT=1`) + 2 hand-written examples (bucket-py, bucket-go)

**Key context:** Research already consolidated in `pulumi-bridge.md` (12 sections, 8 pitfalls, 6-step POC plan). Bridges the existing v2.22.3 provider (28 resources, 21 DS, 779 tests) without rewriting anything.

**Last shipped:** v2.22.3 — Convention Compliance (2026-04-20, 779 tests, 0 lint issues, 12/12 requirements satisfied) — [archive](milestones/v2.22.3-ROADMAP.md)

## Requirements

### Validated

- ✓ 28 resources + 21 data sources with CRUD, import, drift detection — v1.0+v1.1
- ✓ 340 unit tests, state migration framework, validators — v1.2+v1.3
- ✓ CI/CD with GoReleaser + Cosign, import docs, jitter backoff — v1.3
- ✓ S3 tenant onboarding workflow (server → account → export → policies → bucket)
- ✓ Cross-array bucket replication (remote credentials, replica links, array connection DS) — v2.0
- ✓ 368 unit tests, dual-provider replication workflow example — v2.0
- ✓ Security hardening, code quality, 8 shared helpers, dead code removal — v2.0.1
- ✓ 394 unit tests, HCL acceptance tests, 68.4% coverage — v2.0.1
- ✓ Bucket advanced features (lifecycle rules, access policies, audit filters, QoS) — v2.1
- ✓ Network interfaces/VIPs, subnets, LAG data source — v2.1.1
- ✓ Code review fixes, S3 users + user-policy associations, full_access fix — v2.1.3
- ✓ S3 Target replication, certificates, TLS policies, array connections — v2.2
- ✓ API tooling pipeline (swagger parser, diff, upgrade) — tools-v1.0
- ✓ Directory Service Management resource + data source (LDAP admin auth, singleton PATCH-only) — v2.22.1
- ✓ 798 unit tests, LDAPURIValidator, `**NamedReference` Patch pattern validated — v2.22.1
- ✓ Directory Service Role resource + data source + membership composite-ID resource — v2.22.2
- ✓ 814 unit tests, role_name/policy_name composite ID (role FIRST per policy-contains-colon constraint) — v2.22.2
- ✓ Directory Service Role POST `?names=` bug fix + schema v1 (`name` Required + RequiresReplace) + upgrader — Phase 50.1
- ✓ 818 unit tests, end-to-end validated against real FlashBlades (par5, pa7) — Phase 50.1

### Active

- **v2.23.1** — `flashblade_snmp_manager` resource + data source (SNMP-01..13, see `.planning/REQUIREMENTS.md`)

### Known Follow-up Defects

_(none — `member_names` query-param fix landed in commit `05faac1`, certificate regression fixed in `d67484c`)_

### Out of Scope

- Array connection resource (create/delete) — defer to future
- File system replica links — defer to v2.1
- Cascading replication — defer to v2.1
- Directory Service NFS/SMB variants — defer to future milestones
- Active Directory accounts — defer (separate endpoint family)

## Context

- **API reference**: `FLASHBLADE_API.md` in repo root — AI-optimized documentation covering 226 paths, 538 operations
- **Target environment**: Purity//FB 4.6.7, REST API v2.22
- **Consumers**: Operational team doing frequent bucket/filesystem lifecycle management
- **Go module**: `github.com/soulkyu/terraform-provider-flashblade`
- **Framework**: terraform-plugin-framework (modern, recommended by HashiCorp)
- **Future**: Pulumi compatibility planned, Terraform Registry publication planned

## Constraints

- **Framework**: terraform-plugin-framework only — no SDK v2 mixing
- **Go version**: 1.22+ (required by terraform-plugin-framework latest)
- **API version**: v2.22 — provider must handle API versioning headers
- **Auth**: Must support both API token (dev) and OAuth2 client_credentials (prod)
- **TLS**: Must support custom CA certificates (enterprise environments)
- **Testing**: Three-tier — unit, integration (mocked), acceptance (real FlashBlade)
- **Reliability**: Every resource must implement full CRUD + Read (drift) + Import
- **Naming**: Terraform conventions — `flashblade_` prefix, snake_case resources

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| terraform-plugin-framework over SDK/v2 | Modern API, better type safety, diagnostics, plan modifiers — HashiCorp recommended path | — Pending |
| Three-tier testing strategy | Max reliability: unit for logic, mocked for CI, acceptance for real validation | — Pending |
| All policies in v1 scope | Ops team needs full policy control, not partial — avoids click-ops fallback | — Pending |
| Drift detection with audit logging | Ops team needs visibility into what changed outside Terraform for compliance | — Pending |
| Import support for all resources | Team has existing FlashBlade infra to adopt into Terraform state | — Pending |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `/gsd:transition`):
1. Requirements invalidated? → Move to Out of Scope with reason
2. Requirements validated? → Move to Validated with phase reference
3. New requirements emerged? → Add to Active
4. Decisions to log? → Add to Key Decisions
5. "What This Is" still accurate? → Update if drifted

**After each milestone** (via `/gsd:complete-milestone`):
1. Full review of all sections
2. Core Value check — still the right priority?
3. Audit Out of Scope — reasons still valid?
4. Update Context with current state

---
*Last updated: 2026-05-20 — phase 61 complete (`flashblade_snmp_manager` resource + data source shipped, 9/9 must-haves verified, 816 tests on branch `implem-snmp-managers`).*
