# Terraform Provider FlashBlade

## What This Is

A Terraform provider for Pure Storage FlashBlade that enables operational teams to manage storage infrastructure as code — file systems, object stores, buckets, policies, and array administration. Built with terraform-plugin-framework for maximum reliability, targeting the FlashBlade REST API v2.22 (Purity//FB 4.6.7). Designed for high-frequency CRUD operations with robust drift detection and audit logging.

## Core Value

Operational teams can reliably create, update, delete, and reconcile drift on FlashBlade storage resources (buckets, file systems, policies) through Terraform with zero surprises — every plan reflects reality, every apply converges.

## Current Milestone: v2.1.3 — Code Review Fixes

**Goal:** Fix all issues identified by the full codebase code review — critical typos, dead schema attributes, test masking, client hardening, and linting improvements.

**Target features:**
- Fix FreezeLockgedObjects typo in Go struct field name (C1)
- Remove dead nfs_export_policy/smb_share_policy attributes from filesystem schema (I1)
- Fix ExpectNonEmptyPlan masking in acceptance tests (I2)
- Improve OAuth2 token refresh context propagation (I3)
- Expand acceptance test coverage beyond 3 resources (I4)
- Remove fragile RetryBaseDelay heuristic (I5)
- Remove mustObjectValue passthrough helper (M1)
- Remove unused ctx parameters from bucket extract functions (M2)
- Fix diagnostic severity loss in readIntoState (M3)
- Expand golangci-lint configuration (M4)

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

### Active

- [ ] Fix FreezeLockgedObjects typo → FreezeLockedObjects (C1)
- [ ] Remove dead filesystem schema attributes nfs_export_policy/smb_share_policy (I1)
- [ ] Fix ExpectNonEmptyPlan masking in acceptance tests (I2)
- [ ] Improve OAuth2 token refresh context handling (I3)
- [ ] Expand acceptance test coverage to more resources (I4)
- [ ] Remove fragile RetryBaseDelay duration heuristic (I5)
- [ ] Remove mustObjectValue passthrough helper (M1)
- [ ] Remove unused ctx params from bucket extract functions (M2)
- [ ] Fix diagnostic severity loss in readIntoState (M3)
- [ ] Expand golangci-lint linter configuration (M4)

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
*Last updated: 2026-03-31 after milestone v2.1.3 initialization*
