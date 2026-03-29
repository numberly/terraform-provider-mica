# Terraform Provider FlashBlade

## What This Is

A Terraform provider for Pure Storage FlashBlade that enables operational teams to manage storage infrastructure as code — file systems, object stores, buckets, policies, and array administration. Built with terraform-plugin-framework for maximum reliability, targeting the FlashBlade REST API v2.22 (Purity//FB 4.6.7). Designed for high-frequency CRUD operations with robust drift detection and audit logging.

## Core Value

Operational teams can reliably create, update, delete, and reconcile drift on FlashBlade storage resources (buckets, file systems, policies) through Terraform with zero surprises — every plan reflects reality, every apply converges.

## Current Milestone: v1.3 — Release Readiness

**Goal:** Implement best practices from major providers (AWS, Cloudflare) to prepare for public release: state migration framework, import documentation, transport hardening, and write-only sensitive fields.

**Target features:**
- State migration framework (SchemaVersion + UpgradeState on all resources)
- Import documentation (import.md for every resource)
- Move int64UseStateForUnknown to shared helpers
- Jitter in exponential backoff (±20%)
- Write-only argument for secret_access_key (Terraform 1.11+)

## Requirements

### Validated

- ✓ Provider authenticates via API token — v1.0
- ✓ 28 resources with CRUD, import, drift detection — v1.0+v1.1
- ✓ 329 unit tests with idempotence + mock param validation — v1.2
- ✓ 4 bug fixes, architecture cleanup, validators — v1.2
- ✓ CI/CD with GoReleaser + Cosign keyless signing

### Active

- [ ] SchemaVersion 0 + UpgradeState framework on all 28 resources
- [ ] import.md files for all 27 importable resources
- [ ] Move int64UseStateForUnknown to helpers.go
- [ ] Jitter ±20% in exponential backoff
- [ ] Write-only argument for secret_access_key

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
*Last updated: 2026-03-29 after milestone v1.3 initialization*
