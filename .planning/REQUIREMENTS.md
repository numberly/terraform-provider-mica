# Requirements: Terraform Provider FlashBlade

**Defined:** 2026-03-28
**Core Value:** Operational teams can reliably create, update, delete, and reconcile drift on FlashBlade storage resources through Terraform with zero surprises

## v1.1 Requirements

Requirements for milestone v1.1 — Servers & Exports.

### Server Management

- [x] **SRV-01**: Operator can create a FlashBlade server with DNS configuration via Terraform
- [x] **SRV-02**: Operator can update server DNS configuration via Terraform apply
- [x] **SRV-03**: Operator can destroy a server with cascade delete option for dependent exports
- [x] **SRV-04**: Operator can import an existing server into Terraform state
- [x] **SRV-05**: Server data source reads existing server by name (consolidate existing)

### S3 Export Policies

- [x] **S3P-01**: Operator can create an S3 export policy with enable/disable toggle
- [x] **S3P-02**: Operator can create S3 export policy rules with actions/effect/resources (IAM-style)
- [x] **S3P-03**: Operator can update and delete S3 export policy rules independently
- [x] **S3P-04**: Operator can import S3 export policies and rules into Terraform state

### Object Store Virtual Hosts

- [x] **VH-01**: Operator can create a virtual host with hostname and attached servers
- [x] **VH-02**: Operator can update attached servers list on a virtual host
- [x] **VH-03**: Operator can import an existing virtual host into Terraform state

### SMB Client Policies

- [x] **SMC-01**: Operator can create an SMB client policy with enable toggle
- [x] **SMC-02**: Operator can create SMB client policy rules with client/encryption/permission
- [x] **SMC-03**: Operator can update and delete SMB client policy rules independently
- [x] **SMC-04**: Operator can import SMB client policies and rules into Terraform state

### Syslog

- [x] **SYS-01**: Operator can create a syslog server with URI, services, and sources
- [x] **SYS-02**: Operator can update syslog server configuration
- [x] **SYS-03**: Operator can import an existing syslog server into Terraform state

### Export Consolidation

- [x] **EXP-01**: File system export resource has proper TDD unit tests and mock handlers
- [x] **EXP-02**: Account export resource has proper TDD unit tests and mock handlers
- [ ] **EXP-03**: All export resources pass acceptance tests against live FlashBlade

## v1.2 Requirements

Deferred to future release.

### Syslog Settings

- **SYSS-01**: Operator can configure syslog CA certificate settings

### OAuth2 Authentication

- **AUTH-01**: Provider supports OAuth2 client_credentials grant type

## Out of Scope

| Feature | Reason |
|---------|--------|
| Pulumi bridge | Deferred post-v1, provider structure will be compatible |
| Terraform Registry publishing | Internal distribution first, public later |
| Hardware management | Out of API scope |
| Multi-array orchestration | Single FlashBlade target per provider instance |
| Data migration tooling | Provider manages configuration, not data movement |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| SRV-01 | Phase 6 | Complete |
| SRV-02 | Phase 6 | Complete |
| SRV-03 | Phase 6 | Complete |
| SRV-04 | Phase 6 | Complete |
| SRV-05 | Phase 6 | Complete |
| S3P-01 | Phase 7 | Complete |
| S3P-02 | Phase 7 | Complete |
| S3P-03 | Phase 7 | Complete |
| S3P-04 | Phase 7 | Complete |
| VH-01 | Phase 7 | Complete |
| VH-02 | Phase 7 | Complete |
| VH-03 | Phase 7 | Complete |
| SMC-01 | Phase 8 | Complete |
| SMC-02 | Phase 8 | Complete |
| SMC-03 | Phase 8 | Complete |
| SMC-04 | Phase 8 | Complete |
| SYS-01 | Phase 8 | Complete |
| SYS-02 | Phase 8 | Complete |
| SYS-03 | Phase 8 | Complete |
| EXP-01 | Phase 6 | Complete |
| EXP-02 | Phase 6 | Complete |
| EXP-03 | Phase 8 | Pending |

**Coverage:**
- v1.1 requirements: 22 total
- Mapped to phases: 22
- Unmapped: 0

---
*Requirements defined: 2026-03-28*
*Last updated: 2026-03-28 after roadmap v1.1 creation*
