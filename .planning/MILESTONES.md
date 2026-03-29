# Milestones: Terraform Provider FlashBlade

## Completed Milestones

### v1.0 — Core Provider (completed 2026-03-28)

**Goal:** Full Terraform provider for FlashBlade with CRUD, import, and drift detection for all storage resources.

**Delivered:**
- Provider scaffold + HTTP client (auth, TLS, retry, version negotiation)
- 22 resources: file systems, object store (accounts, buckets, access keys), 6 policy families (NFS export, SMB share, snapshot, OAP, NAP, quota), array admin (DNS, NTP, SMTP), file system exports, account exports, server data source
- 227 unit tests, CI pipeline, documentation auto-generated
- 14 bugs fixed via live FlashBlade acceptance testing

**Phases:** 1–5 (20 plans)
**Last phase number:** 5

### v1.1 — Servers & Exports (completed 2026-03-28)

**Goal:** Complete server lifecycle management and export infrastructure.

**Delivered:**
- 10 new resources: server, S3 export policy + rule, virtual host, SMB client policy + rule, syslog server, file system export, account export (TDD consolidated)
- 268 unit tests (41 new), all passing
- 26 resources tested apply/destroy against live FlashBlade
- Complete S3 tenant onboarding workflow (server → account → export → policies → bucket)

**Phases:** 6–8 (8 plans)
**Last phase number:** 8

### v1.2 — Code Quality & Robustness (completed 2026-03-29)

**Goal:** Fix latent bugs, harden test coverage, add validators, clean up architecture.

**Delivered:**
- 4 bug fixes (account export Delete, filesystem writable drift, IsNotFound scoping, omitempty)
- models.go split into 5 domain files, compositeID + stringOrNull shared helpers
- 9 idempotence tests, 4 mock handlers hardened with query param validation
- 2 custom validators (Alphanumeric, HostnameNoDot) + enum OneOf validators
- 329 unit tests (61 new), all passing

**Phases:** 9–11 (7 plans)
**Last phase number:** 11

### v1.3 — Release Readiness (completed 2026-03-29)

**Goal:** Best practices from major providers: state migration, import docs, transport hardening.

**Delivered:**
- SchemaVersion 0 + UpgradeState on all 28 resources
- int64/float64UseStateForUnknown helpers consolidated
- ±20% jitter in exponential backoff
- 27 import.sh files + tfplugindocs regenerated
- 340 unit tests, all passing

**Phases:** 12–13 (4 plans)
**Last phase number:** 13
