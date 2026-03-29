# Requirements: Terraform Provider FlashBlade

**Defined:** 2026-03-29
**Core Value:** Operational teams can reliably create, update, delete, and reconcile drift on FlashBlade storage resources through Terraform with zero surprises

## v2.0.1 Requirements

Requirements for quality & hardening release. Derived from comprehensive 5-agent audit (code review, security, linting, test coverage, dead code analysis).

### Security

- [ ] **SEC-01**: OAuth2 token exchange error sanitizes response body instead of dumping raw content (auth.go:130)
- [ ] **SEC-02**: Provider emits tflog.Warn when insecure_skip_verify is enabled for visibility
- [ ] **SEC-03**: fetchToken() accepts context parameter for cancellation support (auth.go:112)
- [ ] **SEC-04**: NewClient() accepts context parameter instead of using context.Background() (client.go:96,107)
- [ ] **SEC-05**: HTTP client has a global safety-net timeout configured (client.go:84)

### Error Handling

- [ ] **ERR-01**: IsNotFound, IsConflict, IsUnprocessable use errors.As() instead of direct type assertion (errors.go:67,88,97)
- [ ] **ERR-02**: Resource-level error checks use errors.As() pattern (quota_group, quota_user, object_store_account)
- [ ] **ERR-03**: ParseAPIError handles io.ReadAll failure gracefully instead of silently ignoring (errors.go:46)
- [ ] **ERR-04**: LoginWithAPIToken uses http.NewRequestWithContext directly instead of nil-check workaround (auth.go:26-32)

### Code Quality — Validators

- [ ] **VAL-01**: Regex patterns compiled once at package level instead of per-invocation (validators.go:33,66)

### Code Quality — Helpers & Deduplication

- [ ] **DUP-01**: Shared spaceAttrTypes() helper replaces 4 duplicated space schema definitions
- [ ] **DUP-02**: Shared mapSpaceToObject() helper used by filesystem, bucket, and data sources
- [ ] **DUP-03**: nullTimeoutsValue() helper replaces 29 duplicated timeout initialization blocks in ImportState
- [ ] **DUP-04**: mustObjectValue() consolidated into single shared helper in helpers.go (used by filesystem, bucket, object_store_account)
- [ ] **DUP-05**: DiagnosticReporter named interface type replaces inline interface in readIntoState signatures
- [ ] **DUP-06**: Generic getOneByName[T] client helper replaces ~15 identical Get*ByName patterns
- [ ] **DUP-07**: Generic pollUntilGone[T] helper unifies PollUntilEradicated and PollBucketUntilEradicated
- [ ] **DUP-08**: mapFSToModel shared between filesystem resource and data source instead of duplicated

### Dead Code Removal

- [ ] **DCR-01**: Remove 5 unused List* functions and their List*Opts types from client (nfs_export_policies, smb_share_policies, smb_client_policies, snapshot_policies, s3_export_policies)
- [ ] **DCR-02**: Remove unused IsUnprocessable helper from errors.go
- [ ] **DCR-03**: Replace SourceReference with NamedReference (identical types) in models_storage.go
- [ ] **DCR-04**: Remove 29 empty UpgradeState implementations (add back only when schema version bump needed)

### Modernization

- [ ] **MOD-01**: Replace math/rand with math/rand/v2 for Go 1.25 idiomatic usage (transport.go:101)
- [ ] **MOD-02**: mustObjectValue returns diagnostics instead of panic() for safer error handling

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

- [ ] **CON-01**: Bucket delete guard does fresh GET before object count check instead of using stale state (bucket_resource.go:416-423)
- [ ] **CON-02**: countItems in test mock helpers uses reflect or param instead of JSON round-trip (testmock/handlers/helpers.go:36-46)

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
| SEC-01 | — | Pending |
| SEC-02 | — | Pending |
| SEC-03 | — | Pending |
| SEC-04 | — | Pending |
| SEC-05 | — | Pending |
| ERR-01 | — | Pending |
| ERR-02 | — | Pending |
| ERR-03 | — | Pending |
| ERR-04 | — | Pending |
| VAL-01 | — | Pending |
| DUP-01 | — | Pending |
| DUP-02 | — | Pending |
| DUP-03 | — | Pending |
| DUP-04 | — | Pending |
| DUP-05 | — | Pending |
| DUP-06 | — | Pending |
| DUP-07 | — | Pending |
| DUP-08 | — | Pending |
| DCR-01 | — | Pending |
| DCR-02 | — | Pending |
| DCR-03 | — | Pending |
| DCR-04 | — | Pending |
| MOD-01 | — | Pending |
| MOD-02 | — | Pending |
| TST-01 | — | Pending |
| TST-02 | — | Pending |
| TST-03 | — | Pending |
| TST-04 | — | Pending |
| TST-05 | — | Pending |
| TST-06 | — | Pending |
| TST-07 | — | Pending |
| TST-08 | — | Pending |
| CON-01 | — | Pending |
| CON-02 | — | Pending |

**Coverage:**
- v2.0.1 requirements: 34 total
- Mapped to phases: 0
- Unmapped: 34 ⚠️

---
*Requirements defined: 2026-03-29*
*Last updated: 2026-03-29 after quality audit analysis*
