# Terraform Provider FlashBlade

## What This Is

A Terraform provider for Pure Storage FlashBlade that enables operational teams to manage storage infrastructure as code — file systems, object stores, buckets, policies, and array administration. Built with terraform-plugin-framework for maximum reliability, targeting the FlashBlade REST API v2.22 (Purity//FB 4.6.7). Designed for high-frequency CRUD operations with robust drift detection and audit logging.

## Core Value

Operational teams can reliably create, update, delete, and reconcile drift on FlashBlade storage resources (buckets, file systems, policies) through Terraform with zero surprises — every plan reflects reality, every apply converges.

## Requirements

### Validated

(None yet — ship to validate)

### Active

- [ ] Provider authenticates via API token and OAuth2 client_credentials
- [ ] Provider configuration (endpoint, auth, TLS settings)
- [ ] Resource & data source: file systems
- [ ] Resource & data source: object store accounts
- [ ] Resource & data source: object store access keys
- [ ] Resource & data source: buckets
- [ ] Resource & data source: NFS export policies & rules
- [ ] Resource & data source: SMB share policies & rules
- [ ] Resource & data source: snapshot policies & rules
- [ ] Resource & data source: object store access policies & rules
- [ ] Resource & data source: network access policies & rules
- [ ] Resource & data source: quota policies & rules
- [ ] Resource & data source: array administration (DNS, NTP, SMTP, alerts)
- [ ] Accurate drift detection with detailed diff logging for audit
- [ ] Import support for all resources (adopt existing infra into state)
- [ ] Comprehensive acceptance tests against real FlashBlade
- [ ] Unit tests for all logic (schema, planmodifiers, validators)
- [ ] Integration tests with mocked API for CI without FlashBlade

### Out of Scope

- Pulumi bridge — deferred post-v1, provider structure will be compatible
- Terraform Registry publishing — internal distribution first, public later
- Hardware management (drive replacement, shelf operations) — out of API scope
- Multi-array orchestration — single FlashBlade target per provider instance
- Data migration tooling — provider manages configuration, not data movement

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
*Last updated: 2026-03-26 after initialization*
