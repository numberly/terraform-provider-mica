# Requirements: Terraform Provider FlashBlade

**Defined:** 2026-03-28
**Core Value:** Operational teams can reliably create, update, delete, and reconcile drift on FlashBlade storage resources through Terraform with zero surprises

## v1.1 Requirements

Requirements for milestone v1.1 — Servers & Exports.

### Server Management

- [ ] **SRV-01**: Operator can create a FlashBlade server with DNS configuration via Terraform
- [ ] **SRV-02**: Operator can update server DNS configuration via Terraform apply
- [ ] **SRV-03**: Operator can destroy a server with cascade delete option for dependent exports
- [ ] **SRV-04**: Operator can import an existing server into Terraform state
- [ ] **SRV-05**: Server data source reads existing server by name (consolidate existing)

### S3 Export Policies

- [ ] **S3P-01**: Operator can create an S3 export policy with enable/disable toggle
- [ ] **S3P-02**: Operator can create S3 export policy rules with actions/effect/resources (IAM-style)
- [ ] **S3P-03**: Operator can update and delete S3 export policy rules independently
- [ ] **S3P-04**: Operator can import S3 export policies and rules into Terraform state

### Object Store Virtual Hosts

- [ ] **VH-01**: Operator can create a virtual host with hostname and attached servers
- [ ] **VH-02**: Operator can update attached servers list on a virtual host
- [ ] **VH-03**: Operator can import an existing virtual host into Terraform state

### SMB Client Policies

- [ ] **SMC-01**: Operator can create an SMB client policy with enable toggle
- [ ] **SMC-02**: Operator can create SMB client policy rules with client/encryption/permission
- [ ] **SMC-03**: Operator can update and delete SMB client policy rules independently
- [ ] **SMC-04**: Operator can import SMB client policies and rules into Terraform state

### Syslog

- [ ] **SYS-01**: Operator can create a syslog server with URI, services, and sources
- [ ] **SYS-02**: Operator can update syslog server configuration
- [ ] **SYS-03**: Operator can import an existing syslog server into Terraform state

### Export Consolidation

- [ ] **EXP-01**: File system export resource has proper TDD unit tests and mock handlers
- [ ] **EXP-02**: Account export resource has proper TDD unit tests and mock handlers
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
| SRV-01 | — | Pending |
| SRV-02 | — | Pending |
| SRV-03 | — | Pending |
| SRV-04 | — | Pending |
| SRV-05 | — | Pending |
| S3P-01 | — | Pending |
| S3P-02 | — | Pending |
| S3P-03 | — | Pending |
| S3P-04 | — | Pending |
| VH-01 | — | Pending |
| VH-02 | — | Pending |
| VH-03 | — | Pending |
| SMC-01 | — | Pending |
| SMC-02 | — | Pending |
| SMC-03 | — | Pending |
| SMC-04 | — | Pending |
| SYS-01 | — | Pending |
| SYS-02 | — | Pending |
| SYS-03 | — | Pending |
| EXP-01 | — | Pending |
| EXP-02 | — | Pending |
| EXP-03 | — | Pending |

**Coverage:**
- v1.1 requirements: 22 total
- Mapped to phases: 0
- Unmapped: 22

---
*Requirements defined: 2026-03-28*
*Last updated: 2026-03-28 after milestone v1.1 initialization*
