# Requirements: Terraform Provider FlashBlade

**Defined:** 2026-03-29
**Core Value:** Operational teams can reliably create, update, delete, and reconcile drift on FlashBlade storage resources through Terraform with zero surprises

## v1.3 Requirements

Requirements for milestone v1.3 — Release Readiness.

### State Migration

- [x] **MIG-01**: All 28 resources declare SchemaVersion 0 explicitly
- [x] **MIG-02**: UpgradeState framework wired with empty upgrader list (ready for future schema changes)

### Import Documentation

- [ ] **DOC-01**: import.md file for all 27 importable resources with correct syntax
- [ ] **DOC-02**: tfplugindocs regenerated to include import sections in Registry docs

### Helper Consolidation

- [ ] **HLP-01**: int64UseStateForUnknown moved from filesystem_resource.go to helpers.go
- [ ] **HLP-02**: float64UseStateForUnknown added to helpers.go for consistency

### Transport Hardening

- [ ] **TRN-01**: Exponential backoff includes ±20% random jitter to prevent thundering herds

### Sensitive Data

- [ ] **SEC-01**: secret_access_key uses write-only argument pattern (Terraform 1.11+ compatible)

## v1.4 Requirements

Deferred to future release.

- **SYSS-01**: Syslog server CA certificate settings
- **AUTH-01**: OAuth2 client_credentials grant type

## Out of Scope

| Feature | Reason |
|---------|--------|
| New resources | v1.3 is release-readiness only |
| Pulumi bridge | Deferred, provider structure compatible |
| Terraform Registry publishing | This milestone prepares for it, next milestone publishes |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| MIG-01 | Phase 12 | Complete |
| MIG-02 | Phase 12 | Complete |
| DOC-01 | Phase 13 | Pending |
| DOC-02 | Phase 13 | Pending |
| HLP-01 | Phase 12 | Pending |
| HLP-02 | Phase 12 | Pending |
| TRN-01 | Phase 12 | Pending |
| SEC-01 | Phase 13 | Pending |

**Coverage:**
- v1.3 requirements: 8 total
- Mapped to phases: 8
- Unmapped: 0

---
*Requirements defined: 2026-03-29*
*Last updated: 2026-03-29 after roadmap creation*
