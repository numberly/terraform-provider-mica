# Terraform Provider FlashBlade

## What This Is

A Terraform provider for Pure Storage FlashBlade that enables operational teams to manage storage infrastructure as code — file systems, object stores, buckets, policies, and array administration. Built with terraform-plugin-framework for maximum reliability, targeting the FlashBlade REST API v2.22 (Purity//FB 4.6.7). Designed for high-frequency CRUD operations with robust drift detection and audit logging.

## Core Value

Operational teams can reliably create, update, delete, and reconcile drift on FlashBlade storage resources (buckets, file systems, policies) through Terraform with zero surprises — every plan reflects reality, every apply converges.

## Current Milestone: v1.1 — Servers & Exports

**Goal:** Complete server lifecycle management and export infrastructure so operators can fully manage FlashBlade server topology, S3 export policies, virtual hosts, and SMB client policies through Terraform.

**Target features:**
- Server resource (full CRUD, not just data source)
- S3 export policy + rules (referenced by account exports)
- Object store virtual hosts (S3 virtual-hosted-style access)
- SMB client policy + rules (parallel to existing SMB share policy)
- Syslog server configuration
- Consolidate existing export resources with proper TDD tests

## Requirements

### Validated

- ✓ Provider authenticates via API token — v1.0
- ✓ Provider configuration (endpoint, auth, TLS settings) — v1.0
- ✓ 22 resources: file systems, object store chain, 6 policy families, array admin, exports — v1.0
- ✓ 227 unit tests, CI pipeline, documentation — v1.0
- ✓ 14 bugs fixed via live FlashBlade acceptance testing — v1.0

### Active

- [ ] Server resource with full CRUD (create, DNS config, cascade delete)
- [ ] S3 export policy resource + rules (IAM-style allow/deny)
- [ ] Object store virtual host resource (hostname + server attachment)
- [ ] SMB client policy resource + rules (client auth, encryption)
- [ ] Syslog server resource (URI, services, sources)
- [ ] Proper TDD unit + mock tests for all new and existing export resources
- [ ] Acceptance tests for server/export lifecycle against live FlashBlade

### Out of Scope

- Pulumi bridge — deferred post-v1, provider structure will be compatible
- Terraform Registry publishing — internal distribution first, public later
- Hardware management (drive replacement, shelf operations) — out of API scope
- Multi-array orchestration — single FlashBlade target per provider instance
- Data migration tooling — provider manages configuration, not data movement
- Syslog server settings (CA cert config) — low priority, defer to v1.2

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

---
*Last updated: 2026-03-28 after milestone v1.1 initialization*
