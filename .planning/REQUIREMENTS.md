# Requirements: Terraform Provider FlashBlade

**Defined:** 2026-03-29
**Core Value:** Operational teams can reliably create, update, delete, and reconcile drift on FlashBlade storage resources through Terraform with zero surprises

## v2.0.1 Requirements

Requirements for quality & hardening release. Derived from comprehensive 5-agent audit (code review, security, linting, test coverage, dead code analysis).

### Security

- [x] **SEC-01**: OAuth2 token exchange error sanitizes response body instead of dumping raw content (auth.go:130)
- [x] **SEC-02**: Provider emits tflog.Warn when insecure_skip_verify is enabled for visibility
- [x] **SEC-03**: fetchToken() accepts context parameter for cancellation support (auth.go:112)
- [x] **SEC-04**: NewClient() accepts context parameter instead of using context.Background() (client.go:96,107)
- [x] **SEC-05**: HTTP client has a global safety-net timeout configured (client.go:84)

### Error Handling

- [x] **ERR-01**: IsNotFound, IsConflict, IsUnprocessable use errors.As() instead of direct type assertion (errors.go:67,88,97)
- [x] **ERR-02**: Resource-level error checks use errors.As() pattern (quota_group, quota_user, object_store_account)
- [x] **ERR-03**: ParseAPIError handles io.ReadAll failure gracefully instead of silently ignoring (errors.go:46)
- [x] **ERR-04**: LoginWithAPIToken uses http.NewRequestWithContext directly instead of nil-check workaround (auth.go:26-32)

### Code Quality — Validators

- [x] **VAL-01**: Regex patterns compiled once at package level instead of per-invocation (validators.go:33,66)

### Code Quality — Helpers & Deduplication

- [x] **DUP-01**: Shared spaceAttrTypes() helper replaces 4 duplicated space schema definitions
- [x] **DUP-02**: Shared mapSpaceToObject() helper used by filesystem, bucket, and data sources
- [x] **DUP-03**: nullTimeoutsValue() helper replaces 29 duplicated timeout initialization blocks in ImportState
- [x] **DUP-04**: mustObjectValue() consolidated into single shared helper in helpers.go (used by filesystem, bucket, object_store_account)
- [x] **DUP-05**: DiagnosticReporter named interface type replaces inline interface in readIntoState signatures
- [x] **DUP-06**: Generic getOneByName[T] client helper replaces ~15 identical Get*ByName patterns
- [x] **DUP-07**: Generic pollUntilGone[T] helper unifies PollUntilEradicated and PollBucketUntilEradicated
- [x] **DUP-08**: mapFSToModel shared between filesystem resource and data source instead of duplicated

### Dead Code Removal

- [ ] **DCR-01**: Remove 5 unused List* functions and their List*Opts types from client (nfs_export_policies, smb_share_policies, smb_client_policies, snapshot_policies, s3_export_policies)
- [ ] **DCR-02**: Remove unused IsUnprocessable helper from errors.go
- [ ] **DCR-03**: Replace SourceReference with NamedReference (identical types) in models_storage.go
- [ ] **DCR-04**: Remove 29 empty UpgradeState implementations (add back only when schema version bump needed)

### Modernization

- [ ] **MOD-01**: Replace math/rand with math/rand/v2 for Go 1.25 idiomatic usage (transport.go:101)
- [x] **MOD-02**: mustObjectValue returns diagnostics instead of panic() for safer error handling

### Test Coverage

- [ ] **TST-01**: Unit tests for object_store_virtual_host data source (Read + NotFound)
- [ ] **TST-02**: Unit tests for remote_credentials data source (Read + NotFound)
- [ ] **TST-03**: Unit tests for bucket_replica_link data source (Read + NotFound)
- [ ] **TST-04**: Unit tests for file_system_export data source (Read + NotFound)
- [ ] **TST-05**: Unit tests for object_store_account_export data source (Read + NotFound)
- [ ] **TST-06**: OAuth2 provider configuration test (client_id + key_id + issuer flow)
- [ ] **TST-07**: HCL-based acceptance tests using resource.UnitTest with mock server for full Terraform lifecycle (plan → apply → refresh → import → destroy)
- [ ] **TST-08**: Pagination tests for client methods beyond filesystems (at least buckets + one policy type)

### Code Consistency

- [x] **CON-01**: Bucket delete guard does fresh GET before object count check instead of using stale state (bucket_resource.go:416-423)
- [x] **CON-02**: countItems in test mock helpers uses reflect or param instead of JSON round-trip (testmock/handlers/helpers.go:36-46)

## Future Requirements

### v2.1

- **FUT-01**: FlashBladeClient.Close()/Logout() method for session cleanup
- **FUT-02**: Generic Configure helper to reduce 54 identical Configure method bodies
- **FUT-03**: Acceptance tests on live FlashBlade pair for replication resources
- **FUT-04**: Client package direct unit tests for all 21 untested client files
- **FUT-05**: Concurrent operation testing for thread-safety validation

## Out of Scope

| Feature | Reason |
|---------|--------|
| New resources or data sources | This is a hardening-only milestone |
| API version upgrade | No new API features needed |
| CI/CD pipeline changes | Pipeline already functional from v1.3 |
| Documentation regeneration | No schema changes that affect docs |
| Policy resource CRUD abstraction | High complexity, low ROI given framework limitations — track as tech debt |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| SEC-01 | Phase 18 | Complete |
| SEC-02 | Phase 18 | Complete |
| SEC-03 | Phase 18 | Complete |
| SEC-04 | Phase 18 | Complete |
| SEC-05 | Phase 18 | Complete |
| ERR-01 | Phase 19 | Complete |
| ERR-02 | Phase 19 | Complete |
| ERR-03 | Phase 19 | Complete |
| ERR-04 | Phase 18 | Complete |
| VAL-01 | Phase 20 | Complete |
| DUP-01 | Phase 20 | Complete |
| DUP-02 | Phase 20 | Complete |
| DUP-03 | Phase 20 | Complete |
| DUP-04 | Phase 20 | Complete |
| DUP-05 | Phase 20 | Complete |
| DUP-06 | Phase 20 | Complete |
| DUP-07 | Phase 20 | Complete |
| DUP-08 | Phase 20 | Complete |
| DCR-01 | Phase 21 | Pending |
| DCR-02 | Phase 21 | Pending |
| DCR-03 | Phase 21 | Pending |
| DCR-04 | Phase 21 | Pending |
| MOD-01 | Phase 21 | Pending |
| MOD-02 | Phase 20 | Complete |
| TST-01 | Phase 22 | Pending |
| TST-02 | Phase 22 | Pending |
| TST-03 | Phase 22 | Pending |
| TST-04 | Phase 22 | Pending |
| TST-05 | Phase 22 | Pending |
| TST-06 | Phase 22 | Pending |
| TST-07 | Phase 22 | Pending |
| TST-08 | Phase 22 | Pending |
| CON-01 | Phase 19 | Complete |
| CON-02 | Phase 19 | Complete |

**Coverage:**
- v2.0.1 requirements: 34 total
- Mapped to phases: 34/34
- Unmapped: 0

---
*Requirements defined: 2026-03-29*
*Last updated: 2026-03-29 after roadmap creation (phases 18-22)*
