# Requirements: Terraform Provider FlashBlade

**Defined:** 2026-03-29
**Core Value:** Operational teams can reliably create, update, delete, and reconcile drift on FlashBlade storage resources through Terraform with zero surprises

## v2.0 Requirements

Requirements for milestone v2.0 — Cross-Array Bucket Replication.

### Access Key Enhancement

- [ ] **AKE-01**: Access key resource accepts optional `secret_access_key` input to import an existing key pair onto a second FlashBlade
- [ ] **AKE-02**: When `secret_access_key` is provided, POST sends it in the body; when omitted, API generates it
- [ ] **AKE-03**: Bucket `versioning` attribute validated as required when bucket is used with replication

### Remote Credentials

- [ ] **RCR-01**: Operator can create remote credentials with access_key_id + secret_access_key referencing a remote array
- [ ] **RCR-02**: Operator can update remote credentials (rotate keys)
- [ ] **RCR-03**: Operator can import existing remote credentials into Terraform state

### Bucket Replica Link

- [ ] **BRL-01**: Operator can create a bucket replica link between a local bucket and a remote bucket
- [ ] **BRL-02**: Operator can pause/resume a replica link via Terraform apply
- [ ] **BRL-03**: Operator can destroy a replica link cleanly
- [ ] **BRL-04**: Operator can import an existing replica link into Terraform state
- [ ] **BRL-05**: Replica link exposes read-only status fields (direction, lag, recovery_point, status)

### Array Connection

- [ ] **ACN-01**: Data source reads an existing array connection by remote name
- [ ] **ACN-02**: Data source exposes id, status, remote name, management_address, replication_addresses

### Workflow

- [ ] **WFL-01**: Complete dual-provider replication workflow example (FB-A + FB-B, same tenant, shared credentials, bidirectional replica links)
- [ ] **WFL-02**: TDD unit tests + mock handlers for all new resources
- [ ] **WFL-03**: Acceptance test for replication lifecycle on live FlashBlade pair

### Documentation

- [ ] **DOC-01**: HCL examples + import.sh for all new resources (remote credentials, replica link, array connection DS)
- [ ] **DOC-02**: tfplugindocs regenerated with new resources included
- [ ] **DOC-03**: README updated with replication resources category + coverage table

## v2.1 Requirements

Deferred to future release.

- Target resource (external S3 replication to AWS/Backblaze/OCI)
- Array connection resource (create/delete, not just data source)
- File system replica links
- Cascading replication (A → B → C)

## Out of Scope

| Feature | Reason |
|---------|--------|
| External S3 target replication | v2.1 — different topology, needs target + credentials model |
| Array connection management | v2.1 — connection key workflow is complex, data source sufficient for now |
| File system replication | v2.1 — different API family (ActiveDR), not S3 replication |
| Cascading replication | v2.1 — advanced topology, requires cascading_enabled support |
| Sync replication | Not supported by FlashBlade for S3 (async only) |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| AKE-01 | — | Pending |
| AKE-02 | — | Pending |
| AKE-03 | — | Pending |
| RCR-01 | — | Pending |
| RCR-02 | — | Pending |
| RCR-03 | — | Pending |
| BRL-01 | — | Pending |
| BRL-02 | — | Pending |
| BRL-03 | — | Pending |
| BRL-04 | — | Pending |
| BRL-05 | — | Pending |
| ACN-01 | — | Pending |
| ACN-02 | — | Pending |
| WFL-01 | — | Pending |
| WFL-02 | — | Pending |
| WFL-03 | — | Pending |
| DOC-01 | — | Pending |
| DOC-02 | — | Pending |
| DOC-03 | — | Pending |

**Coverage:**
- v2.0 requirements: 19 total
- Mapped to phases: 0
- Unmapped: 19

---
*Requirements defined: 2026-03-29*
*Last updated: 2026-03-29 after milestone v2.0 initialization*
