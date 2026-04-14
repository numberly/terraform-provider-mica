# Terraform Provider FlashBlade

## What This Is

A Terraform provider for Pure Storage FlashBlade that enables operational teams to manage storage infrastructure as code — file systems, object stores, buckets, policies, and array administration. Built with terraform-plugin-framework for maximum reliability, targeting the FlashBlade REST API v2.22 (Purity//FB 4.6.7). Designed for high-frequency CRUD operations with robust drift detection and audit logging.

## Core Value

Operational teams can reliably create, update, delete, and reconcile drift on FlashBlade storage resources (buckets, file systems, policies) through Terraform with zero surprises — every plan reflects reality, every apply converges.

## Current Milestone: tools-v1.0 — API Tooling Pipeline

**Goal:** Automate swagger-to-reference conversion, API version diffing, and provider upgrade orchestration through Claude Code skills with Python tooling.

**Target features:**
- Skill swagger-to-reference: swagger.json → AI-optimized markdown in api_references/ + Python browsing tools
- Skill api-diff: structured diff between swagger versions, swagger inaccuracy handling, migration plan cross-ref ROADMAP.md
- Skill api-upgrade: mechanical version update scripts + Claude orchestration with review gates
- Shared Python lib (_shared/swagger_utils.py) for allOf resolution, path normalization, schema flattening
- CLAUDE.md update with API tools reference and api_references/ convention

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

### Active

- [ ] Target resource — CRUD for external S3 endpoints (TGT-01)
- [ ] Target data source — read existing targets (TGT-02)
- [ ] Remote credentials support for S3 targets (RC-01)
- [ ] Bucket replica links to S3 targets (BRL-01)
- [ ] Workflow examples and documentation (DOC-01)

### Out of Scope

- Pulumi bridge — deferred, provider structure compatible
- Array connection resource (create/delete) — defer to future
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
*Last updated: 2026-04-02 after starting milestone v2.2 (S3 Target Replication)*
