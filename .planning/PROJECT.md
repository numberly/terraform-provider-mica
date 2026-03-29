# Terraform Provider FlashBlade

## What This Is

A Terraform provider for Pure Storage FlashBlade that enables operational teams to manage storage infrastructure as code — file systems, object stores, buckets, policies, and array administration. Built with terraform-plugin-framework for maximum reliability, targeting the FlashBlade REST API v2.22 (Purity//FB 4.6.7). Designed for high-frequency CRUD operations with robust drift detection and audit logging.

## Core Value

Operational teams can reliably create, update, delete, and reconcile drift on FlashBlade storage resources (buckets, file systems, policies) through Terraform with zero surprises — every plan reflects reality, every apply converges.

## Current Milestone: v2.0 — Cross-Array Bucket Replication

**Goal:** Enable operators to set up bidirectional S3 bucket replication between two FlashBlade arrays through Terraform, with shared credentials for transparent failover.

**Target features:**
- Access key resource accepts optional secret_access_key input (import existing key to second array)
- Object store remote credentials resource (each FB authenticates to the other)
- Bucket replica link resource (bidirectional replication between two FB arrays)
- Array connection data source (read existing inter-array connection)
- Complete replication workflow example (dual-provider, same tenant on both arrays)

## Requirements

### Validated

- ✓ 28 resources + 21 data sources with CRUD, import, drift detection — v1.0+v1.1
- ✓ 340 unit tests, state migration framework, validators — v1.2+v1.3
- ✓ CI/CD with GoReleaser + Cosign, import docs, jitter backoff — v1.3
- ✓ S3 tenant onboarding workflow (server → account → export → policies → bucket)

### Active

- [ ] Access key accepts optional secret_access_key input for cross-array credential sharing
- [ ] Object store remote credentials resource (CRUD)
- [ ] Bucket replica link resource (bidirectional, CRUD)
- [ ] Array connection data source (read existing connection)
- [ ] Bucket versioning must be enforced/validated for replication
- [ ] Complete dual-provider replication workflow example
- [ ] TDD tests for all new resources + acceptance test on live FlashBlade pair

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
*Last updated: 2026-03-29 after milestone v2.0 initialization*
