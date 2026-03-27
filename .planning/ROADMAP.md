# Roadmap: Terraform Provider FlashBlade

## Overview

Building the provider in five phases: establish the shared HTTP client and correct CRUD pattern on the first resource (Phase 1), extend to the object store resource chain (Phase 2), implement all six policy families across two phases (Phases 3 and 4, grouped by domain), then harden tests and documentation to production standard (Phase 5). Every FlashBlade-specific pitfall — soft-delete, computed attribute misuse, unordered policy rules, singleton admin resources — is resolved once in the earliest applicable phase before the pattern replicates.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

- [x] **Phase 1: Foundation** - Shared HTTP client, provider scaffold, and first resource (`flashblade_file_system`) establishing all CRUD patterns (completed 2026-03-27)
- [ ] **Phase 2: Object Store Resources** - Complete object store resource chain: account, bucket, and access key with dependency ordering
- [ ] **Phase 3: File-Based Policy Resources** - NFS export, SMB share, and snapshot policy families with parent/child rule pattern
- [ ] **Phase 4: Object/Network/Quota Policies and Array Admin** - Remaining three policy families plus singleton array administration resources
- [ ] **Phase 5: Quality Hardening** - Unit tests, mocked integration tests, documentation generation, and release pipeline

## Phase Details

### Phase 1: Foundation
**Goal**: Operators can configure the provider and manage file systems via Terraform with full CRUD, import, and drift detection — all shared infrastructure patterns established for replication
**Depends on**: Nothing (first phase)
**Requirements**: PROV-01, PROV-02, PROV-03, PROV-04, PROV-05, PROV-06, PROV-07, FS-01, FS-02, FS-03, FS-04, FS-05, FS-06, FS-07
**Success Criteria** (what must be TRUE):
  1. Provider connects to a FlashBlade using either API token or OAuth2 client_credentials, respecting environment variable fallbacks and custom CA certificate
  2. `terraform plan` on a file system shows accurate diff with zero false positives after `apply` (idempotency: apply → plan → 0 changes)
  3. `terraform destroy` on a file system completes the two-phase soft-delete without name-collision failures on re-creation
  4. `terraform import flashblade_file_system.x name` populates all attributes; subsequent `plan` shows 0 diff
  5. Drift detection produces structured `tflog` output listing changed fields when `terraform refresh` detects API-side divergence
**Plans:** 4/4 plans complete

Plans:
- [ ] 01-01-PLAN.md — Project scaffold, build tooling, and shared HTTP client layer (auth, TLS, retry, version negotiation)
- [ ] 01-02-PLAN.md — Provider schema, Configure with env var fallback, and client injection
- [ ] 01-03-PLAN.md — File system client CRUD methods and reusable mock HTTP server
- [ ] 01-04-PLAN.md — flashblade_file_system resource (CRUD, import, soft-delete, drift) and data source

### Phase 2: Object Store Resources
**Goal**: Operators can manage the complete object store resource chain — accounts, buckets, and access keys — through Terraform with full lifecycle and dependency ordering
**Depends on**: Phase 1
**Requirements**: OSA-01, OSA-02, OSA-03, OSA-04, OSA-05, BKT-01, BKT-02, BKT-03, BKT-04, BKT-05, BKT-06, OAK-01, OAK-02, OAK-03, OAK-04, OAK-05
**Success Criteria** (what must be TRUE):
  1. Operator can create an account, then a bucket referencing that account, then generate an access key — all in a single `terraform apply`
  2. Secret access key value is available in state only at creation time, marked Sensitive, and does not appear in plan output
  3. `terraform destroy` on a bucket completes two-phase soft-delete; same bucket name can be recreated immediately after
  4. `terraform import` works for account, bucket, and access key; subsequent `plan` shows 0 diff for each
**Plans:** 2/3 plans executed

Plans:
- [ ] 02-01-PLAN.md — Object store account: models, client CRUD, mock handler, resource, data source
- [ ] 02-02-PLAN.md — Bucket: client CRUD, mock handler with account cross-ref, resource with soft-delete, data source
- [ ] 02-03-PLAN.md — Access key: client methods, mock handler, resource with write-once secret, data source

### Phase 3: File-Based Policy Resources
**Goal**: Operators can manage NFS export, SMB share, and snapshot policies — including rules — through Terraform with no false drift on rule reorder
**Depends on**: Phase 2
**Requirements**: NFP-01, NFP-02, NFP-03, NFP-04, NFP-05, NFR-01, NFR-02, NFR-03, NFR-04, SMP-01, SMP-02, SMP-03, SMP-04, SMP-05, SMR-01, SMR-02, SMR-03, SMR-04, SNP-01, SNP-02, SNP-03, SNP-04, SNP-05, SNR-01, SNR-02, SNR-03, SNR-04
**Success Criteria** (what must be TRUE):
  1. Operator can create an NFS export policy with rules; `apply → plan` shows 0 diff regardless of API rule return order
  2. Operator can import NFS, SMB, and snapshot policy rules using composite ID (`policy_name:rule_index`); subsequent `plan` shows 0 diff
  3. Operator can create, update, and destroy SMB share policy and snapshot policy rules independently of the parent policy lifecycle
  4. All three policy data sources return attributes by name or filter without provider errors
**Plans:** 4 plans

Plans:
- [ ] 03-01-PLAN.md — All Phase 3 model structs, client CRUD methods, and mock handlers for NFS/SMB/Snapshot
- [ ] 03-02-PLAN.md — NFS export policy resource, rule resource, data source with tests
- [ ] 03-03-PLAN.md — SMB share policy resource, rule resource, data source with tests
- [ ] 03-04-PLAN.md — Snapshot policy resource, rule resource (PATCH-based), data source with tests

### Phase 4: Object/Network/Quota Policies and Array Admin
**Goal**: Operators have full policy coverage (object store access, network access, quota) and can manage array-level DNS, NTP, and SMTP configuration through Terraform
**Depends on**: Phase 3
**Requirements**: OAP-01, OAP-02, OAP-03, OAP-04, OAP-05, OAR-01, OAR-02, OAR-03, OAR-04, NAP-01, NAP-02, NAP-03, NAP-04, NAP-05, NAR-01, NAR-02, NAR-03, NAR-04, QTP-01, QTP-02, QTP-03, QTP-04, QTP-05, QTR-01, QTR-02, QTR-03, QTR-04, ADM-01, ADM-02, ADM-03, ADM-04, ADM-05
**Success Criteria** (what must be TRUE):
  1. Operator can create an object store access policy with IAM-style rules (effect, action, resource, condition); `apply → plan` shows 0 diff
  2. Operator can create network access and quota policies with rules; composite import IDs work for all rule types
  3. Operator can manage array DNS, NTP, and SMTP configuration; `apply → plan` shows 0 diff on singleton resources
  4. `terraform destroy` on a singleton array admin resource returns a clear diagnostic (not a silent no-op or panic)
**Plans**: TBD

### Phase 5: Quality Hardening
**Goal**: All resources are covered by unit tests, mocked integration tests, and auto-generated documentation; release pipeline is operational
**Depends on**: Phase 4
**Requirements**: QUA-01, QUA-02, QUA-03, QUA-04, QUA-05, QUA-06
**Success Criteria** (what must be TRUE):
  1. `go test ./...` passes with unit coverage for all schema definitions, validators, and plan modifiers
  2. Mocked integration tests cover the full CRUD lifecycle for all resource families without a live FlashBlade (CI-safe)
  3. HTTP client retries transparently on 429/503/5xx responses; operator sees no transient failures during `terraform apply`
  4. `terraform-plugin-docs` generates complete documentation for every resource and data source without manual editing
**Plans**: TBD

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 2 → 3 → 4 → 5

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Foundation | 4/4 | Complete   | 2026-03-27 |
| 2. Object Store Resources | 2/3 | In Progress|  |
| 3. File-Based Policy Resources | 0/4 | Not started | - |
| 4. Object/Network/Quota Policies and Array Admin | 0/TBD | Not started | - |
| 5. Quality Hardening | 0/TBD | Not started | - |
