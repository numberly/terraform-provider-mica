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
