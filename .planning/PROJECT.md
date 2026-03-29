# Terraform Provider FlashBlade

## What This Is

A Terraform provider for Pure Storage FlashBlade that enables operational teams to manage storage infrastructure as code — file systems, object stores, buckets, policies, and array administration. Built with terraform-plugin-framework for maximum reliability, targeting the FlashBlade REST API v2.22 (Purity//FB 4.6.7). Designed for high-frequency CRUD operations with robust drift detection and audit logging.

## Core Value

Operational teams can reliably create, update, delete, and reconcile drift on FlashBlade storage resources (buckets, file systems, policies) through Terraform with zero surprises — every plan reflects reality, every apply converges.

## Current Milestone: v2.0.1 — Quality & Hardening

**Goal:** Harden the codebase post-v2.0 release with security fixes, code quality improvements, dead code removal, duplication reduction, and test coverage gap closure identified by comprehensive 5-agent audit.

**Target features:**
- Fix OAuth2 error body leak and add context propagation to auth paths
- Migrate error handling to errors.As() for wrapped error resilience
- Extract shared helpers (spaceAttrTypes, nullTimeoutsValue, mustObjectValue, getOneByName[T])
- Remove dead code (unused List* functions, SourceReference, empty UpgradeState)
- Compile regex at package level in validators
- Add tests for 5 uncovered data sources + OAuth2 provider config test
- Add HCL-based acceptance tests using mock server

## Requirements

### Validated

- ✓ 28 resources + 21 data sources with CRUD, import, drift detection — v1.0+v1.1
- ✓ 340 unit tests, state migration framework, validators — v1.2+v1.3
- ✓ CI/CD with GoReleaser + Cosign, import docs, jitter backoff — v1.3
- ✓ S3 tenant onboarding workflow (server → account → export → policies → bucket)
- ✓ Cross-array bucket replication (remote credentials, replica links, array connection DS) — v2.0
- ✓ 368 unit tests, dual-provider replication workflow example — v2.0

### Active

- [ ] OAuth2 error body sanitization and context propagation
- [ ] errors.As() migration for wrapped error resilience
- [ ] Shared helper extraction (space schema, timeouts, mustObjectValue, getOneByName)
- [ ] Dead code removal (unused List*, SourceReference, empty UpgradeState)
- [ ] Regex pre-compilation in validators
- [ ] Test coverage for 5 uncovered data sources
- [ ] HCL-based acceptance tests with mock server
- [ ] OAuth2 provider configuration test

### Out of Scope

- Pulumi bridge — deferred, provider structure compatible
- Target resource (external S3 replication) — defer to v2.1
- Array connection resource (create/delete) — defer to v2.1
- File system replica links — defer to v2.1
- Cascading replication — defer to v2.1

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
*Last updated: 2026-03-29 after milestone v2.0.1 initialization*
