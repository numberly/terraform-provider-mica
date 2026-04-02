# Requirements: Terraform Provider FlashBlade

**Defined:** 2026-04-02
**Core Value:** Operational teams can reliably create, update, delete, and reconcile drift on FlashBlade storage resources through Terraform with zero surprises

## v2.2 Requirements

Requirements for S3 Target Replication. Enables operators to replicate buckets to external S3-compatible endpoints through Terraform.

### Target Management

- [x] **TGT-01**: Operator can create a target with name, address, and optional ca_certificate_group via `terraform apply` — `apply -> plan` shows 0 diff
- [x] **TGT-02**: Operator can update mutable target fields (address, ca_certificate_group) and destroy a target via Terraform
- [x] **TGT-03**: Operator can import an existing target into Terraform state and subsequent `plan` shows 0 diff
- [x] **TGT-04**: Operator can read an existing target by name via `flashblade_target` data source and access its address, status, and status_details attributes
- [x] **TGT-05**: Drift detection logs changes when a target is modified outside Terraform

### Remote Credentials Enhancement

- [ ] **RC-01**: Operator can create remote credentials referencing a target (not just an array connection) via the existing `flashblade_object_store_remote_credentials` resource
- [ ] **RC-02**: Existing remote credentials functionality for array connections is not broken by the enhancement

### Replication Workflow

- [ ] **BRL-01**: Operator can create a bucket replica link using remote credentials that reference a target — replication to external S3 endpoint works end-to-end

### Documentation

- [ ] **DOC-01**: Import documentation (import.sh) exists for `flashblade_target` with correct syntax and realistic identifiers
- [ ] **DOC-02**: A workflow example in `examples/s3-target-replication/` demonstrates the full stack: target creation → remote credentials → bucket replica link to external S3
- [ ] **DOC-03**: `tfplugindocs generate` produces documentation for all new resources and data sources without errors

## v2.1.3 Requirements (completed)

### Code Correctness

- [x] **CC-01**: FreezeLockgedObjects typo renamed to FreezeLockedObjects across all Go files
- [x] **CC-02**: Dead schema attributes nfs_export_policy and smb_share_policy removed from filesystem resource
- [x] **CC-03**: Diagnostic severity preserved when converting mapFSToModel results — warnings remain warnings, errors remain errors

### Test Quality

- [ ] **TQ-01**: Acceptance tests no longer use ExpectNonEmptyPlan: true — plan convergence is verified
- [ ] **TQ-02**: Acceptance test coverage expanded to at least 3 additional high-risk resources

### Client Hardening

- [x] **CH-01**: OAuth2 token refresh uses caller context where possible instead of context.Background()
- [x] **CH-02**: RetryBaseDelay duration heuristic removed — callers must use explicit time.Duration values
- [x] **CH-03**: Unused ctx parameters removed from bucket extract functions

### Code Cleanup

- [x] **CL-01**: mustObjectValue passthrough helper removed — callers use types.ObjectValue directly
- [x] **CL-02**: golangci-lint configuration expanded with gosec, bodyclose, noctx, and exhaustive linters

### Object Store Users

- [x] **OSU-01**: Operator can create a named S3 user under an account via Terraform
- [x] **OSU-02**: Operator can delete an S3 user via Terraform destroy
- [x] **OSU-03**: Operator can read an existing S3 user by name via data source
- [x] **OSU-04**: Operator can import an existing S3 user into Terraform state with no drift
- [x] **OSU-05**: Operator can associate one or more access policies to a user via a member resource
- [x] **OSU-06**: Operator can remove a policy association from a user via Terraform destroy
- [x] **OSU-07**: Drift detection logs changes when user or policy association is modified outside Terraform

## Future Requirements

Deferred to future release. Tracked but not in current roadmap.

### Target Replication Monitoring

- **MON-01**: Data source for target replication performance metrics (`/targets/performance/replication`)

## Out of Scope

| Feature | Reason |
|---------|--------|
| Array connection resource (create/delete) | Separate concern, deferred to future milestone |
| File system replica links to targets | S3 bucket replication only for v2.2 |
| Target certificate management | Handled by existing certificate + TLS policy resources |
| Cascading replication (target -> target) | Not supported by FlashBlade API |
| Realms | Not relevant for current usage |
| Pulumi bridge | Deferred, provider structure compatible |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| TGT-01 | Phase 36 | Complete |
| TGT-02 | Phase 36 | Complete |
| TGT-03 | Phase 36 | Complete |
| TGT-04 | Phase 36 | Complete |
| TGT-05 | Phase 36 | Complete |
| RC-01 | Phase 37 | Pending |
| RC-02 | Phase 37 | Pending |
| BRL-01 | Phase 37 | Pending |
| DOC-01 | Phase 38 | Pending |
| DOC-02 | Phase 38 | Pending |
| DOC-03 | Phase 38 | Pending |

**v2.2 Coverage:**
- v2.2 requirements: 11 total
- Mapped to phases: 11
- Unmapped: 0

---
*Requirements defined: 2026-04-02*
*Last updated: 2026-04-02 after v2.2 roadmap creation*
