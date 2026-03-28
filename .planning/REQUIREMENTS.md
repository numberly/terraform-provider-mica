# Requirements: Terraform Provider FlashBlade

**Defined:** 2026-03-28
**Core Value:** Operational teams can reliably create, update, delete, and reconcile drift on FlashBlade storage resources through Terraform with zero surprises

## v1.2 Requirements

Requirements for milestone v1.2 — Code Quality & Robustness.

### Bug Fixes

- [x] **BUG-01**: Account export Delete passes correct short export name (not combined)
- [x] **BUG-02**: File system `writable` field mapped correctly in Read (0 drift on plan)
- [x] **BUG-03**: IsNotFound scoped to actual "not found" messages (not all 400s)
- [x] **BUG-04**: Fix omitempty on nested structs (use pointers or json:"-" where needed)

### Architecture Cleanup

- [x] **ARC-01**: Split models.go into domain files (storage, policies, exports, admin)
- [ ] **ARC-02**: Unified compositeID helper for policy rule import/delete
- [ ] **ARC-03**: Extract stringOrNull to shared helper (used by all rule resources)

### Test Hardening

- [ ] **TST-01**: Idempotence tests — Create → Read → plan shows 0 changes for each resource family
- [ ] **TST-02**: Mock handlers validate query params (reject unknown params, require mandatory ones)
- [ ] **TST-03**: Complete Update lifecycle tests for resources missing them

### Validators

- [ ] **VAL-01**: Name format validators (S3 rule: alphanumeric, virtual host: no dots, etc.)
- [ ] **VAL-02**: Terraform plan-time validation for enum fields (effect, permission, versioning)

## v1.3 Requirements

Deferred to future release.

- **SYSS-01**: Syslog server CA certificate settings
- **AUTH-01**: OAuth2 client_credentials grant type

## Out of Scope

| Feature | Reason |
|---------|--------|
| New resources | v1.2 is quality-only, no new features |
| Pulumi bridge | Deferred, provider structure compatible |
| Terraform Registry | Internal distribution first |
| Hardware management | Out of API scope |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| BUG-01 | Phase 9 | Complete |
| BUG-02 | Phase 9 | Complete |
| BUG-03 | Phase 9 | Complete |
| BUG-04 | Phase 9 | Complete |
| ARC-01 | Phase 10 | Complete |
| ARC-02 | Phase 10 | Pending |
| ARC-03 | Phase 10 | Pending |
| TST-01 | Phase 11 | Pending |
| TST-02 | Phase 11 | Pending |
| TST-03 | Phase 11 | Pending |
| VAL-01 | Phase 11 | Pending |
| VAL-02 | Phase 11 | Pending |

**Coverage:**
- v1.2 requirements: 12 total
- Mapped to phases: 12
- Unmapped: 0

---
*Requirements defined: 2026-03-28*
*Last updated: 2026-03-28 after v1.2 roadmap creation*
