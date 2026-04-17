# Terraform Provider FlashBlade

## What This Is

A Terraform provider for Pure Storage FlashBlade that enables operational teams to manage storage infrastructure as code — file systems, object stores, buckets, policies, and array administration. Built with terraform-plugin-framework for maximum reliability, targeting the FlashBlade REST API v2.22 (Purity//FB 4.6.7). Designed for high-frequency CRUD operations with robust drift detection and audit logging.

## Core Value

Operational teams can reliably create, update, delete, and reconcile drift on FlashBlade storage resources (buckets, file systems, policies) through Terraform with zero surprises — every plan reflects reality, every apply converges.

## Current Milestone: v2.22.2 Directory Service Roles & Role Mappings

**Goal:** Ajouter la gestion Terraform des role mappings LDAP ↔ FlashBlade via deux ressources suivant le pattern `_member` déjà établi dans le provider.

**Target features:**
- Ressource `flashblade_directory_service_role` — maps un groupe LDAP à un rôle FlashBlade (array_admin, storage_admin, etc.) via `/directory-services/roles`
- Data source `flashblade_directory_service_role` — lecture d'un mapping existant par name
- Ressource `flashblade_management_access_policy_directory_service_role_membership` — association séparée role ↔ management_access_policy (composite ID `policy_name:role_name`) via `/management-access-policies/directory-services/roles`
- Suit le pattern des 5 ressources `_member` existantes (qos_policy_member, tls_policy_member, certificate_group_member, audit_object_store_policy_member, object_store_user_policy)
- Examples HCL + import.sh + `make docs` régénéré

**Last shipped:** v2.22.1 — Directory Service Management (2026-04-17, 798 tests, 0 lint issues)

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

### Active

- [ ] `flashblade_directory_service_role` resource — LDAP group → FB role mapping (DSR-01)
- [ ] `flashblade_directory_service_role` data source — read existing mapping (DSR-02)
- [ ] `flashblade_management_access_policy_directory_service_role_membership` — composite ID membership resource (DSRM-01)
- [ ] HCL examples + import.sh for both resources (DOC-01, DOC-02)

### Out of Scope

- Pulumi bridge — deferred, provider structure compatible
- Array connection resource (create/delete) — defer to future
- File system replica links — defer to v2.1
- Cascading replication — defer to v2.1
- Directory Service NFS/SMB variants — defer to future milestones
- Directory Service roles / role mappings — defer (separate endpoint family)
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
*Last updated: 2026-04-17 after starting milestone v2.22.2 (Directory Service Roles & Role Mappings)*
