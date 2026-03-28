# Terraform Provider FlashBlade

## What This Is

A Terraform provider for Pure Storage FlashBlade that enables operational teams to manage storage infrastructure as code — file systems, object stores, buckets, policies, and array administration. Built with terraform-plugin-framework for maximum reliability, targeting the FlashBlade REST API v2.22 (Purity//FB 4.6.7). Designed for high-frequency CRUD operations with robust drift detection and audit logging.

## Core Value

Operational teams can reliably create, update, delete, and reconcile drift on FlashBlade storage resources (buckets, file systems, policies) through Terraform with zero surprises — every plan reflects reality, every apply converges.

## Current Milestone: v1.2 — Code Quality & Robustness

**Goal:** Fix latent bugs, harden test coverage with idempotence checks and strict mock validation, add input validators, and clean up architectural inconsistencies across the provider.

**Target features:**
- Fix confirmed bugs (account export Delete, filesystem writable drift, IsNotFound too broad)
- Split monolithic models.go by domain
- Unified composite ID helper for policy rule import/delete
- Idempotence tests (Create → Read → plan = 0 changes)
- Mock handlers that validate query params (catch API mismatches in CI)
- Terraform validators for resource names (alphanumeric rules, hostname formats)

## Requirements

### Validated

- ✓ Provider authenticates via API token — v1.0
- ✓ 22 core resources with CRUD, import, drift detection — v1.0
- ✓ 10 server/export resources with TDD tests — v1.1
- ✓ 268 unit tests, CI pipeline, documentation — v1.1
- ✓ 26 resources acceptance-tested against live FlashBlade — v1.1

### Active

- [ ] Fix account export Delete bug (combined name vs short name)
- [ ] Fix filesystem writable drift (permanent 1-change on plan)
- [ ] Tighten IsNotFound to avoid masking real 400 errors
- [ ] Split models.go by domain (storage, policies, exports, admin)
- [ ] Unified compositeID helper for policy rule import/delete
- [ ] Extract stringOrNull helper to shared location
- [ ] Idempotence tests for all resource families
- [ ] Mock query param validation (reject invalid params like real API)
- [ ] Complete Update tests for resources missing them
- [ ] Terraform validators (name format, hostname, S3 rule name)
- [ ] Fix omitempty on nested structs (prevent sending empty objects)

### Out of Scope

- Pulumi bridge — deferred, provider structure compatible
- Terraform Registry publishing — internal distribution first
- Hardware management — out of API scope
- Multi-array orchestration — single target per provider instance
- Syslog server CA cert settings — defer to v1.3
- New resources — this milestone is quality-only, no new features

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
*Last updated: 2026-03-28 after milestone v1.2 initialization*
