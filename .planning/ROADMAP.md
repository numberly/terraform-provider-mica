# Roadmap: Terraform Provider FlashBlade

## Milestones

- v1.0 Core Provider (Phases 1-5) -- shipped 2026-03-28
- v1.1 Servers & Exports (Phases 6-8) -- shipped 2026-03-28
- v1.2 Code Quality & Robustness (Phases 9-11) -- shipped 2026-03-29
- v1.3 Release Readiness (Phases 12-13) -- shipped 2026-03-29
- v2.0 Cross-Array Bucket Replication (Phases 14-17) -- shipped 2026-03-29
- v2.0.1 Quality & Hardening (Phases 18-22) -- shipped 2026-03-30
- v2.1 Bucket Advanced Features (Phases 23-27) -- shipped 2026-03-30
- v2.1.1 Network Interfaces (VIPs) (Phases 28-31) -- shipped 2026-03-31
- v2.1.3 Code Review Fixes & S3 Users (Phases 32-35) -- shipped 2026-04-02
- v2.2 S3 Target Replication & TLS (Phases 36-42) -- shipped 2026-04-14
- tools-v1.0 API Tooling Pipeline (Phases 43-48) -- in progress

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

<details>
<summary>v1.0 Core Provider (Phases 1-5) - SHIPPED 2026-03-28</summary>

### Phase 1: Foundation
**Goal**: Operators can configure the provider and manage file systems via Terraform with full CRUD, import, and drift detection — all shared infrastructure patterns established for replication
**Depends on**: Nothing (first phase)
**Requirements**: PROV-01, PROV-02, PROV-03, PROV-04, PROV-05, PROV-06, PROV-07, FS-01, FS-02, FS-03, FS-04, FS-05, FS-06, FS-07
**Success Criteria** (what must be TRUE):
  1. Provider connects to a FlashBlade using either API token or OAuth2 client_credentials, respecting environment variable fallbacks and custom CA certificate
  2. `terraform plan` on a file system shows accurate diff with zero false positives after `apply` (idempotency: apply -> plan -> 0 changes)
  3. `terraform destroy` on a file system completes the two-phase soft-delete without name-collision failures on re-creation
  4. `terraform import flashblade_file_system.x name` populates all attributes; subsequent `plan` shows 0 diff
  5. Drift detection produces structured `tflog` output listing changed fields when `terraform refresh` detects API-side divergence
**Plans:** 4/4 plans complete

Plans:
- [x] 01-01-PLAN.md — Project scaffold, build tooling, and shared HTTP client layer
- [x] 01-02-PLAN.md — Provider schema, Configure with env var fallback, and client injection
- [x] 01-03-PLAN.md — File system client CRUD methods and reusable mock HTTP server
- [x] 01-04-PLAN.md — flashblade_file_system resource (CRUD, import, soft-delete, drift) and data source

### Phase 2: Object Store Resources
**Goal**: Operators can manage the complete object store resource chain — accounts, buckets, and access keys — through Terraform with full lifecycle and dependency ordering
**Depends on**: Phase 1
**Requirements**: OSA-01, OSA-02, OSA-03, OSA-04, OSA-05, BKT-01, BKT-02, BKT-03, BKT-04, BKT-05, BKT-06, OAK-01, OAK-02, OAK-03, OAK-04, OAK-05
**Success Criteria** (what must be TRUE):
  1. Operator can create an account, then a bucket referencing that account, then generate an access key — all in a single `terraform apply`
  2. Secret access key value is available in state only at creation time, marked Sensitive, and does not appear in plan output
  3. `terraform destroy` on a bucket completes two-phase soft-delete; same bucket name can be recreated immediately after
  4. `terraform import` works for account, bucket, and access key; subsequent `plan` shows 0 diff for each
**Plans:** 3/3 plans complete

Plans:
- [x] 02-01-PLAN.md — Object store account: models, client CRUD, mock handler, resource, data source
- [x] 02-02-PLAN.md — Bucket: client CRUD, mock handler with account cross-ref, resource with soft-delete, data source
- [x] 02-03-PLAN.md — Access key: client methods, mock handler, resource with write-once secret, data source

### Phase 3: File-Based Policy Resources
**Goal**: Operators can manage NFS export, SMB share, and snapshot policies — including rules — through Terraform with no false drift on rule reorder
**Depends on**: Phase 2
**Requirements**: NFP-01, NFP-02, NFP-03, NFP-04, NFP-05, NFR-01, NFR-02, NFR-03, NFR-04, SMP-01, SMP-02, SMP-03, SMP-04, SMP-05, SMR-01, SMR-02, SMR-03, SMR-04, SNP-01, SNP-02, SNP-03, SNP-04, SNP-05, SNR-01, SNR-02, SNR-03, SNR-04
**Success Criteria** (what must be TRUE):
  1. Operator can create an NFS export policy with rules; `apply -> plan` shows 0 diff regardless of API rule return order
  2. Operator can import NFS, SMB, and snapshot policy rules using composite ID (`policy_name:rule_index`); subsequent `plan` shows 0 diff
  3. Operator can create, update, and destroy SMB share policy and snapshot policy rules independently of the parent policy lifecycle
  4. All three policy data sources return attributes by name or filter without provider errors
**Plans:** 4/4 plans complete

Plans:
- [x] 03-01-PLAN.md — All Phase 3 model structs, client CRUD methods, and mock handlers for NFS/SMB/Snapshot
- [x] 03-02-PLAN.md — NFS export policy resource, rule resource, data source with tests
- [x] 03-03-PLAN.md — SMB share policy resource, rule resource, data source with tests
- [x] 03-04-PLAN.md — Snapshot policy resource, rule resource (PATCH-based), data source with tests

### Phase 4: Object/Network/Quota Policies and Array Admin
**Goal**: Operators have full policy coverage (object store access, network access, quota) and can manage array-level DNS, NTP, and SMTP configuration through Terraform
**Depends on**: Phase 3
**Requirements**: OAP-01, OAP-02, OAP-03, OAP-04, OAP-05, OAR-01, OAR-02, OAR-03, OAR-04, NAP-01, NAP-02, NAP-03, NAP-04, NAP-05, NAR-01, NAR-02, NAR-03, NAR-04, QTP-01, QTP-02, QTP-03, QTP-04, QTP-05, QTR-01, QTR-02, QTR-03, QTR-04, ADM-01, ADM-02, ADM-03, ADM-04, ADM-05
**Success Criteria** (what must be TRUE):
  1. Operator can create an object store access policy with IAM-style rules (effect, action, resource, condition); `apply -> plan` shows 0 diff
  2. Operator can create network access and quota policies with rules; composite import IDs work for all rule types
  3. Operator can manage array DNS, NTP, and SMTP configuration; `apply -> plan` shows 0 diff on singleton resources
  4. `terraform destroy` on a singleton array admin resource returns a clear diagnostic (not a silent no-op or panic)
**Plans:** 5/5 plans complete

Plans:
- [x] 04-01-PLAN.md — All Phase 4 model structs, client CRUD methods, and mock handlers
- [x] 04-02-PLAN.md — Object store access policy resource, rule resource (IAM-style), data source
- [x] 04-03-PLAN.md — Network access policy singleton resource, rule resource, data source
- [x] 04-04-PLAN.md — Quota user and quota group resources, data sources
- [x] 04-05-PLAN.md — Array admin DNS, NTP, SMTP singleton resources with data sources

### Phase 5: Quality Hardening
**Goal**: All resources are covered by unit tests, mocked integration tests, and auto-generated documentation; release pipeline is operational
**Depends on**: Phase 4
**Requirements**: QUA-01, QUA-02, QUA-03, QUA-04, QUA-05, QUA-06
**Success Criteria** (what must be TRUE):
  1. `go test ./...` passes with unit coverage for all schema definitions, validators, and plan modifiers
  2. Mocked integration tests cover the full CRUD lifecycle for all resource families without a live FlashBlade (CI-safe)
  3. HTTP client retries transparently on 429/503/5xx responses; operator sees no transient failures during `terraform apply`
  4. `terraform-plugin-docs` generates complete documentation for every resource and data source without manual editing
**Plans:** 4/4 plans complete

Plans:
- [x] 05-01-PLAN.md — Error helpers, validators, plan modifier assertions, and validator tests
- [x] 05-02-PLAN.md — Auto-pagination, error-path tests (409/422/404), and retry verification
- [x] 05-03-PLAN.md — HCL examples, terraform-plugin-docs generation, GitHub Actions CI, and README
- [x] 05-04-PLAN.md — Full lifecycle tests and import idempotency tests for all 19 resources

</details>

<details>
<summary>v1.1 Servers & Exports (Phases 6-8) - SHIPPED 2026-03-28</summary>

### Phase 6: Server Resource & Export Consolidation
**Goal**: Operators can manage FlashBlade servers through Terraform and existing export resources have proper TDD test coverage
**Depends on**: Phase 5 (v1.0 complete)
**Requirements**: SRV-01, SRV-02, SRV-03, SRV-04, SRV-05, EXP-01, EXP-02
**Success Criteria** (what must be TRUE):
  1. Operator can create a server with DNS configuration, update its DNS settings, and destroy it (with cascade delete) via Terraform plan/apply
  2. Operator can import an existing FlashBlade server into Terraform state and subsequent plan shows no drift
  3. Server data source reads an existing server by name and exposes its attributes for reference in other resources
  4. File system export and account export resources each have mock handlers and unit tests following TDD patterns established in v1.0
**Plans:** 2/2 plans complete

Plans:
- [x] 06-01-PLAN.md — Server model extension, client CRUD, mock handler, resource, data source update, and tests
- [x] 06-02-PLAN.md — Export mock handlers and TDD unit tests for file system and account exports

### Phase 7: S3 Export Policies & Virtual Hosts
**Goal**: Operators can manage S3 export access policies and virtual-hosted-style S3 endpoints through Terraform
**Depends on**: Phase 6 (servers must exist for virtual host attachment)
**Requirements**: S3P-01, S3P-02, S3P-03, S3P-04, VH-01, VH-02, VH-03
**Success Criteria** (what must be TRUE):
  1. Operator can create an S3 export policy, toggle its enabled state, and add IAM-style rules (actions/effect/resources) via Terraform
  2. Operator can update and delete individual S3 export policy rules without affecting sibling rules
  3. Operator can create a virtual host with a hostname, attach servers to it, and update the server list via Terraform apply
  4. Operator can import existing S3 export policies, rules, and virtual hosts into Terraform state with no drift on subsequent plan
**Plans:** 3/3 plans complete

Plans:
- [x] 07-01-PLAN.md — Model structs, client CRUD methods, and mock handlers for S3 export policies and virtual hosts
- [x] 07-02-PLAN.md — S3 export policy resource, rule resource, data source, and unit tests
- [x] 07-03-PLAN.md — Object store virtual host resource, data source, and unit tests

### Phase 8: SMB Client Policies, Syslog & Acceptance Tests
**Goal**: Remaining resource types are implemented and all v1.1 resources pass acceptance tests against a live FlashBlade
**Depends on**: Phase 7
**Requirements**: SMC-01, SMC-02, SMC-03, SMC-04, SYS-01, SYS-02, SYS-03, EXP-03
**Success Criteria** (what must be TRUE):
  1. Operator can create an SMB client policy with enable toggle, add rules with client/encryption/permission settings, and update or delete rules independently
  2. Operator can create a syslog server with URI, services, and sources, update its configuration, and import it into Terraform state
  3. Operator can import SMB client policies and rules into Terraform state with no drift on subsequent plan
  4. All v1.1 resources (server, S3 export policy/rules, virtual hosts, SMB client policy/rules, syslog server) pass acceptance tests against a live FlashBlade array
**Plans:** 3/3 plans complete

Plans:
- [x] 08-01-PLAN.md — SMB client policy resource, rule resource, data source, client CRUD, mock handler, and unit tests
- [x] 08-02-PLAN.md — Syslog server resource, data source, client CRUD, mock handler, and unit tests
- [x] 08-03-PLAN.md — Acceptance test HCL configs and live FlashBlade validation for all v1.1 resources

</details>

<details>
<summary>v1.2 Code Quality & Robustness (Phases 9-11) - SHIPPED 2026-03-29</summary>

### Phase 9: Bug Fixes
**Goal**: All confirmed bugs are fixed so the provider produces correct plans and correct API calls for every existing resource
**Depends on**: Phase 8 (v1.1 complete)
**Requirements**: BUG-01, BUG-02, BUG-03, BUG-04
**Success Criteria** (what must be TRUE):
  1. `terraform destroy` on an account export sends the short export name to the API (not the combined account/export name) and deletes successfully
  2. `terraform plan` on an existing file system with `writable = true` shows 0 changes (no permanent 1-change drift)
  3. A 400 error from the API that is not "does not exist" propagates as a real error to the operator (IsNotFound no longer masks non-404 failures)
  4. PATCH/POST requests for resources with nested structs do not send empty `{}` objects for unset fields (omitempty works correctly with pointer types)
**Plans:** 2/2 plans complete

Plans:
- [x] 09-01-PLAN.md — Fix account export Delete name extraction and filesystem writable drift
- [x] 09-02-PLAN.md — Scope IsNotFound matching and audit omitempty on struct fields

### Phase 10: Architecture Cleanup
**Goal**: Codebase is organized by domain with shared helpers so that future development is faster and less error-prone
**Depends on**: Phase 9 (fixes applied before refactoring)
**Requirements**: ARC-01, ARC-02, ARC-03
**Success Criteria** (what must be TRUE):
  1. `models.go` is split into domain files (storage, policies, exports, admin) and `go build ./...` compiles without errors
  2. All policy rule resources use a single `compositeID` helper for import parsing and delete ID construction (no duplicated split/join logic)
  3. All rule resources that convert nullable strings use a shared `stringOrNull` helper from a common package (no inline duplicates)
**Plans:** 2/2 plans complete

Plans:
- [x] 10-01-PLAN.md — Split models.go into 5 domain files (storage, policies, exports, admin, common)
- [x] 10-02-PLAN.md — Shared compositeID/parseCompositeID helpers and stringOrNull extraction

### Phase 11: Test Hardening & Validators
**Goal**: Test suite catches API mismatches and regressions; operators get clear plan-time errors for invalid inputs instead of API-time failures
**Depends on**: Phase 10 (tests validate refactored code, validators reference clean models)
**Requirements**: TST-01, TST-02, TST-03, VAL-01, VAL-02
**Success Criteria** (what must be TRUE):
  1. Every resource family has an idempotence test: Create -> Read -> plan shows 0 changes (catches writable-style drift bugs before release)
  2. Mock handlers reject unknown query params and require mandatory ones (catches client-side API mismatches in CI, not production)
  3. Resources that were missing Update lifecycle tests now have them, covering at least one mutable field per resource
  4. `terraform validate` rejects invalid resource names (e.g., dots in virtual host names, non-alphanumeric S3 rule names) with a clear error before any API call
  5. `terraform validate` rejects invalid enum values (e.g., invalid effect, permission, versioning) with the set of allowed values in the error message
**Plans:** 3/3 plans complete

Plans:
- [x] 11-01-PLAN.md — Custom name format validators and enum OneOf validators for 6 resource schemas
- [x] 11-02-PLAN.md — Shared query param validation helper and mock handler hardening for 4 handlers
- [x] 11-03-PLAN.md — Idempotence tests for 9 v1.1 resources and standalone Update tests for 3 resources

</details>

<details>
<summary>v1.3 Release Readiness (Phases 12-13) - SHIPPED 2026-03-29</summary>

### Phase 12: Infrastructure Hardening
**Goal**: All 28 resources have a state migration framework ready for future schema changes, shared plan modifier helpers are consolidated, and retry logic prevents thundering herds
**Depends on**: Phase 11 (v1.2 complete)
**Requirements**: MIG-01, MIG-02, HLP-01, HLP-02, TRN-01
**Success Criteria** (what must be TRUE):
  1. Every resource schema declares `SchemaVersion: 0` and wires an `UpgradeState` method with an empty upgrader slice -- `go build ./...` compiles and `go test ./...` passes
  2. `int64UseStateForUnknown` and `float64UseStateForUnknown` plan modifiers live in `helpers.go` and are referenced from all resource files that need them (no inline definitions remain)
  3. Retry backoff intervals vary by at least 20% between consecutive identical requests (jitter prevents synchronized retries from multiple provider instances)
**Plans**: 2/2 plans complete

Plans:
- [x] 12-01-PLAN.md — State migration framework (SchemaVersion 0 + UpgradeState) on all 28 resources
- [x] 12-02-PLAN.md — Helper consolidation (int64/float64 UseStateForUnknown) and transport jitter

### Phase 13: Documentation & Sensitive Data
**Goal**: Every importable resource has Registry-ready import documentation, and the access key secret uses the write-only argument pattern for Terraform 1.11+ compatibility
**Depends on**: Phase 12 (infrastructure changes landed before docs generation)
**Requirements**: DOC-01, DOC-02, SEC-01
**Success Criteria** (what must be TRUE):
  1. Every importable resource (27 of 28) has an `import.sh` file with correct `terraform import` syntax and an example using realistic identifiers
  2. `tfplugindocs generate` produces documentation that includes import sections for all importable resources without errors or manual edits
  3. `secret_access_key` on `flashblade_object_store_access_key` uses the write-only attribute pattern: the value is never stored in state, never appears in plan diff, and is only sent to the API on create
**Plans**: 2/2 plans complete

Plans:
- [x] 13-01-PLAN.md — Import documentation (import.sh) for 10 remaining importable resources + tfplugindocs regeneration
- [x] 13-02-PLAN.md — Write-only argument pattern for secret_access_key (WriteOnly: true, framework v1.19.0)

</details>

<details>
<summary>v2.0 Cross-Array Bucket Replication (Phases 14-17) - SHIPPED 2026-03-29</summary>

### Phase 14: Access Key Enhancement & Array Connection
**Goal**: Operators can share an S3 access key pair across two FlashBlade arrays and query existing array connections -- the foundation for cross-array replication
**Depends on**: Phase 13 (v1.3 complete)
**Requirements**: AKE-01, AKE-02, AKE-03, ACN-01, ACN-02
**Success Criteria** (what must be TRUE):
  1. Operator can create an access key on FB-A (API generates secret), then create the same key on FB-B by providing `secret_access_key` in HCL -- both keys have identical credentials
  2. When `secret_access_key` is omitted, existing behavior is unchanged (API generates the secret); when provided, the POST body includes it and the API accepts it
  3. `flashblade_array_connection` data source reads an existing connection by remote array name and exposes id, status, management_address, and replication_addresses
  4. Bucket resource validates that `versioning` is set to `"enabled"` when the bucket participates in replication (plan-time error, not API-time)
**Plans:** 2/2 plans complete

Plans:
- [x] 14-01-PLAN.md — Access key resource enhancement (optional secret_access_key input) + bucket versioning warning
- [x] 14-02-PLAN.md — Array connection data source (model, client, mock handler, data source, tests)

### Phase 15: Replication Resources
**Goal**: Operators can create the credential and link infrastructure required for bidirectional bucket replication between two FlashBlade arrays
**Depends on**: Phase 14 (access keys with shared secrets and array connection data must exist)
**Requirements**: RCR-01, RCR-02, RCR-03, BRL-01, BRL-02, BRL-03, BRL-04, BRL-05
**Success Criteria** (what must be TRUE):
  1. Operator can create remote credentials on each FlashBlade pointing to the other array's access key -- `terraform apply` succeeds on both providers
  2. Operator can rotate remote credentials (update access_key_id + secret_access_key) via `terraform apply` without destroying the credential resource
  3. Operator can create a bucket replica link between a local and remote bucket, pause/resume it via attribute change, and destroy it cleanly
  4. Bucket replica link exposes read-only status fields (direction, lag, recovery_point, status) that reflect current replication state after `terraform refresh`
  5. Operator can import existing remote credentials and bucket replica links into Terraform state; subsequent `plan` shows 0 diff
**Plans**: 3/3 plans complete

Plans:
- [x] 15-01-PLAN.md — Remote credentials resource (model, client CRUD, mock handler, resource, data source, tests)
- [x] 15-02-PLAN.md — Bucket replica link resource (model, client CRUD, mock handler, resource, data source, tests)

### Phase 16: Workflow & Documentation
**Goal**: Operators have a complete, copy-pasteable dual-provider replication example and all new resources are documented for the Terraform Registry
**Depends on**: Phase 15 (all replication resources must exist before the workflow references them)
**Requirements**: WFL-01, DOC-01, DOC-02, DOC-03
**Success Criteria** (what must be TRUE):
  1. A complete HCL example in `examples/replication/` demonstrates the full workflow: dual provider config, account, bucket with versioning, access keys (shared secret), remote credentials, and bidirectional replica links
  2. Every new resource and data source (remote credentials, bucket replica link, array connection DS) has an `import.sh` file with realistic identifiers
  3. `tfplugindocs generate` produces documentation including all new resources without errors or manual edits
  4. README coverage table includes the replication resources category with correct resource counts
**Plans:** 1/1 plans complete

Plans:
- [x] 16-01-PLAN.md — HCL examples, replication workflow, docs regeneration, README update

### Phase 17: Testing
**Goal**: All new replication resources have comprehensive test coverage and pass validation against a live FlashBlade pair
**Depends on**: Phase 16 (workflow example informs acceptance test structure)
**Requirements**: WFL-02, WFL-03
**Success Criteria** (what must be TRUE):
  1. TDD unit tests exist for all new resources (remote credentials, bucket replica link) and the enhanced access key resource -- including mock handlers that validate API request bodies
  2. Array connection data source has unit tests covering read-by-name, not-found error, and attribute mapping
  3. Acceptance tests execute the full replication lifecycle on a live FlashBlade pair: create credentials, create replica link, verify status fields, pause/resume, destroy cleanly
**Plans:** 2/2 plans complete

Plans:
- [x] 17-01-PLAN.md — Unit tests for remote credentials and bucket replica link resources (Create, Read, Update, Delete, Import, Idempotence)
- [x] 17-02-PLAN.md — Acceptance test HCL for replication lifecycle on live FlashBlade pair (checkpoint-gated)

</details>

<details>
<summary>v2.0.1 Quality & Hardening (Phases 18-22) - SHIPPED 2026-03-30</summary>

### Phase 18: Security & Auth Hardening
**Goal**: Auth paths are hardened against information leaks and support proper context propagation for cancellation and tracing
**Depends on**: Phase 17 (v2.0 complete)
**Requirements**: SEC-01, SEC-02, SEC-03, SEC-04, SEC-05, ERR-04
**Success Criteria** (what must be TRUE):
  1. OAuth2 token exchange failure logs a sanitized error (no raw response body in error message or logs) and returns a diagnostic the operator can act on
  2. Provider emits a `tflog.Warn` when `insecure_skip_verify` is enabled so operators see the security risk in plan output
  3. `fetchToken()` and `NewClient()` accept a `context.Context` parameter -- no `context.Background()` calls remain in auth.go or client.go initialization paths
  4. HTTP client has a global safety-net timeout (e.g., 30s) so a hung FlashBlade API does not block `terraform apply` indefinitely
  5. `LoginWithAPIToken` constructs the HTTP request with `http.NewRequestWithContext` directly (no nil-check workaround)
**Plans**: 1/1 plans complete

Plans:
- [x] 18-01-PLAN.md — OAuth2 error sanitization, TLS warning, context propagation, HTTP timeout, API token request cleanup

### Phase 19: Error Handling & Consistency
**Goal**: Error classification is resilient to wrapped errors and edge cases across the entire codebase
**Depends on**: Phase 18 (auth changes landed so error patterns are consistent)
**Requirements**: ERR-01, ERR-02, ERR-03, CON-01, CON-02
**Success Criteria** (what must be TRUE):
  1. `IsNotFound`, `IsConflict`, and `IsUnprocessable` use `errors.As()` so they correctly classify errors even when wrapped with `fmt.Errorf("%w", ...)`
  2. Resource-level error checks in quota_group, quota_user, and object_store_account use the same `errors.As()` pattern (no direct type assertions remain)
  3. `ParseAPIError` returns a meaningful error when `io.ReadAll` fails instead of silently returning an empty error
  4. Bucket delete performs a fresh GET before the object-count safety check instead of relying on potentially stale state data
  5. `countItems` in test mock helpers uses `reflect` or a typed parameter instead of JSON marshal/unmarshal round-trip
**Plans**: 1/1 plans complete

Plans:
- [x] 19-01-PLAN.md — errors.As() migration, ParseAPIError hardening, bucket delete guard, countItems fix

### Phase 20: Code Quality -- Validators & Deduplication
**Goal**: Shared helpers eliminate duplicated code across resources, and validators execute without per-call regex compilation overhead
**Depends on**: Phase 19 (error patterns stable before extracting helpers that may reference them)
**Requirements**: VAL-01, DUP-01, DUP-02, DUP-03, DUP-04, DUP-05, DUP-06, DUP-07, DUP-08, MOD-02
**Success Criteria** (what must be TRUE):
  1. Regex patterns in validators.go are compiled once at package level (`var` block with `regexp.MustCompile`) -- no `regexp.Compile` or `regexp.MustCompile` calls inside function bodies
  2. `spaceAttrTypes()` and `mapSpaceToObject()` helpers exist in a shared location and are called by filesystem, bucket, and data source files (no duplicated space schema definitions remain)
  3. `nullTimeoutsValue()` helper replaces all 29 inline timeout initialization blocks in ImportState methods -- verified by searching for the old pattern
  4. Generic `getOneByName[T]` client helper exists and is called by at least 10 Get*ByName methods (reducing ~15 identical implementations to thin wrappers)
  5. `mustObjectValue` returns diagnostics instead of calling `panic()` -- callers check the returned diagnostics and append to response diagnostics
**Plans**: 2/2 plans complete

Plans:
- [x] 20-01-PLAN.md — Regex pre-compilation, space schema helpers, nullTimeoutsValue, DiagnosticReporter interface
- [x] 20-02-PLAN.md — Generic getOneByName[T], pollUntilGone[T], mapFSToModel sharing, mustObjectValue diagnostics

### Phase 21: Dead Code Removal & Modernization
**Goal**: Unused code is removed and deprecated stdlib usage is updated so the codebase is clean for future development
**Depends on**: Phase 20 (shared helpers must exist before removing code that may be replaced by them)
**Requirements**: DCR-01, DCR-02, DCR-03, DCR-04, MOD-01
**Success Criteria** (what must be TRUE):
  1. The 5 unused `List*` functions and their `List*Opts` types are removed from the client package -- `go build ./...` compiles and no test references them
  2. `IsUnprocessable` is removed from errors.go (confirmed unused by all resources)
  3. `SourceReference` type is removed and all references use `NamedReference` instead -- `go build ./...` compiles
  4. The 29 empty `UpgradeState` implementations are removed from resources (only resources with actual schema version bumps keep them)
  5. `math/rand` import is replaced with `math/rand/v2` in transport.go -- jitter behavior unchanged, `go vet ./...` shows no deprecation warnings
**Plans**: 1/1 plans complete

Plans:
- [x] 21-01-PLAN.md — Remove unused List* functions, IsUnprocessable, SourceReference, empty UpgradeState, update math/rand

### Phase 22: Test Coverage
**Goal**: All data sources and auth paths have unit tests, and HCL-based acceptance tests validate the full Terraform lifecycle through the provider
**Depends on**: Phase 21 (all code changes landed so tests validate final state)
**Requirements**: TST-01, TST-02, TST-03, TST-04, TST-05, TST-06, TST-07, TST-08
**Success Criteria** (what must be TRUE):
  1. Unit tests exist for all 5 previously uncovered data sources (virtual host, remote credentials, bucket replica link, file system export, account export) covering Read success and NotFound error paths
  2. OAuth2 provider configuration test verifies that `client_id` + `key_id` + `issuer` flow initializes the provider without errors (mock token endpoint)
  3. At least 3 resources have HCL-based acceptance tests using `resource.UnitTest` with a mock server that exercise plan, apply, refresh, import, and destroy
  4. Pagination tests exist for at least buckets and one policy type (in addition to existing filesystem pagination tests)
**Plans**: 2/2 plans complete

Plans:
- [x] 22-01-PLAN.md — Unit tests for 5 uncovered data sources + OAuth2 provider config test
- [x] 22-02-PLAN.md — HCL-based acceptance tests with mock server + pagination tests for buckets and policies

</details>

### v2.1 Bucket Advanced Features (Shipped 2026-03-30)

**Milestone Goal:** Complete bucket management with lifecycle rules, bucket access policies, audit filters, eradication config, object lock, public access config, and QoS policy support.

- [x] **Phase 23: Bucket Inline Attributes** - Extend bucket resource with eradication, object lock, and public access config (completed 2026-03-30)
- [x] **Phase 24: Lifecycle Rules** - New resource and data source for bucket lifecycle rule management (completed 2026-03-30)
- [x] **Phase 25: Bucket Access Policies** - New resource for IAM-style bucket access policies and rules (completed 2026-03-30)
- [x] **Phase 26: Audit Filters & QoS Policies** - New resources for audit filtering and bandwidth/IOPS limiting (completed 2026-03-30)
- [x] **Phase 27: Testing & Documentation** - Unit tests, mock handlers, import docs, and workflow example (completed 2026-03-30)

### v2.1.1 Network Interfaces (VIPs) (Shipped 2026-03-31)

**Milestone Goal:** Add network interface (Virtual IP) management as a resource and data source, and expose VIPs on the server data source/resource for consumer endpoint discovery.

- [x] **Phase 28: LAG Data Source & Subnet Resource** - LAG read-only data source and full subnet CRUD resource with LAG reference (completed 2026-03-31)
- [x] **Phase 29: Network Interface Resource & Data Source** - Full VIP CRUD resource with service/server semantics and read-only data source (completed 2026-03-31)
- [x] **Phase 30: Server Enrichment & Provider Registration** - Computed VIP list on server resource/data source, schema migration, provider registration (completed 2026-03-31)

## Phase Details

### Phase 23: Bucket Inline Attributes
**Goal**: Operators can configure bucket eradication behavior, object lock settings, and public access restrictions directly on the bucket resource
**Depends on**: Phase 22 (v2.0.1 complete)
**Requirements**: BKT-01, BKT-02, BKT-03, BKT-04
**Success Criteria** (what must be TRUE):
  1. Operator can create a bucket with `eradication_config` (eradication_delay, eradication_mode, manual_eradication) and update these settings via `terraform apply` -- subsequent `plan` shows 0 diff
  2. Operator can create a bucket with `object_lock_config` (freeze_locked_objects, default_retention, default_retention_mode, object_lock_enabled) and update retention settings -- subsequent `plan` shows 0 diff
  3. Operator can update `public_access_config` (block_new_public_policies, block_public_access) on an existing bucket -- subsequent `plan` shows 0 diff
  4. Bucket resource exposes `public_status` as a computed read-only attribute that reflects the current public access state after `terraform refresh`
**Plans**: 1 plan

Plans:
- [ ] 23-01-PLAN.md — Client models + bucket resource schema, mapping, and CRUD for eradication, object lock, public access configs
- [ ] 23-02-PLAN.md — Mock handler updates + test type maps + config block lifecycle tests

### Phase 24: Lifecycle Rules
**Goal**: Operators can manage per-bucket lifecycle rules for version retention and multipart upload cleanup through Terraform
**Depends on**: Phase 23 (bucket attributes must be stable before adding sub-resources that reference buckets)
**Requirements**: LCR-01, LCR-02, LCR-03, LCR-04, LCR-05
**Success Criteria** (what must be TRUE):
  1. Operator can create a lifecycle rule on a bucket with prefix, version retention days, and multipart upload cleanup days via `terraform apply` -- `apply -> plan` shows 0 diff
  2. Operator can update lifecycle rule settings (enabled, retention periods, prefix) and destroy rules independently -- no orphaned rules remain on the API
  3. Operator can import an existing lifecycle rule into Terraform state and subsequent `plan` shows 0 diff
  4. Lifecycle rule data source reads existing rules by bucket name and exposes all rule attributes for reference in other resources
**Plans**: 1 plan

Plans:
- [ ] 24-01-PLAN.md — Client models, CRUD methods, and mock handler for lifecycle rules
- [ ] 24-02-PLAN.md — Lifecycle rule resource, data source, unit tests, and provider registration

### Phase 25: Bucket Access Policies
**Goal**: Operators can manage IAM-style per-bucket authorization with policies and rules through Terraform
**Depends on**: Phase 23 (bucket attributes must be stable before adding sub-resources that reference buckets)
**Requirements**: BAP-01, BAP-02, BAP-03, BAP-04, BAP-05
**Success Criteria** (what must be TRUE):
  1. Operator can create a bucket access policy with rules defining actions, effect, principals, and resources via `terraform apply` -- `apply -> plan` shows 0 diff
  2. Operator can create and delete individual bucket access policy rules independently of the parent policy lifecycle
  3. Operator can import existing bucket access policies into Terraform state and subsequent `plan` shows 0 diff
  4. Bucket access policy data source reads existing policies by bucket name and exposes policy and rule attributes
**Plans**: 1 plan

Plans:
- [ ] 25-01-PLAN.md — Client models, CRUD methods, and mock handler for bucket access policies and rules
- [ ] 25-02-PLAN.md — Bucket access policy resource, rule resource, data source, unit tests, and provider registration

### Phase 26: Audit Filters & QoS Policies
**Goal**: Operators can audit S3 operations per bucket and enforce bandwidth/IOPS limits on buckets and file systems through Terraform
**Depends on**: Phase 23 (bucket attributes must be stable before adding sub-resources that reference buckets)
**Requirements**: BAF-01, BAF-02, BAF-03, BAF-04, QOS-01, QOS-02, QOS-03, QOS-04, QOS-05, QOS-06
**Success Criteria** (what must be TRUE):
  1. Operator can create a bucket audit filter with actions and S3 prefix filtering, update filter settings, and destroy filters via Terraform -- `apply -> plan` shows 0 diff
  2. Operator can import an existing audit filter into Terraform state and subsequent `plan` shows 0 diff
  3. Operator can create a QoS policy with max_total_bytes_per_sec and max_total_ops_per_sec, update limits, and destroy the policy via Terraform -- `apply -> plan` shows 0 diff
  4. Operator can assign a QoS policy to buckets and file systems as members and import existing QoS policies into Terraform state
  5. QoS policy data source reads existing policies by name and exposes all attributes including current member assignments
**Plans**: 3 plans

Plans:
- [ ] 26-01-PLAN.md — Client models, CRUD methods, and mock handlers for audit filters and QoS policies
- [ ] 26-02-PLAN.md — Bucket audit filter resource, data source, unit tests, and provider registration
- [ ] 26-03-PLAN.md — QoS policy resource, member resource, data source, unit tests, and provider registration

### Phase 27: Testing & Documentation
**Goal**: All new v2.1 resources have unit tests with mock handlers, import documentation, and a workflow example showing the complete bucket advanced features stack
**Depends on**: Phase 26 (all resources must exist before comprehensive testing and documentation)
**Requirements**: TST-01, TST-02, DOC-01, DOC-02
**Success Criteria** (what must be TRUE):
  1. Unit tests exist for all new resources (lifecycle rules, bucket access policies, audit filters, QoS policies) and bucket attribute additions -- covering Create, Read, Update, Delete, NotFound paths
  2. Mock handlers exist for all new API endpoints (lifecycle rules, bucket access policies, audit filters, QoS policies) with query param validation
  3. Import documentation (import.sh) exists for all new importable resources with correct syntax and realistic identifiers
  4. A workflow example demonstrates a bucket with lifecycle rules, access policy, audit filter, and QoS policy in a single coherent HCL configuration
**Plans**: 1 plan

Plans:
- [ ] 27-01-PLAN.md — Missing data source test, example HCL + import docs for all v2.1 resources, workflow example, tfplugindocs regeneration


### Phase 28: LAG Data Source & Subnet Resource
**Goal**: Operators can read existing LAG configurations and manage subnets referencing LAGs through Terraform with full CRUD, import, and drift detection
**Depends on**: Phase 27 (v2.1 complete)
**Requirements**: LAG-01, SUB-01, SUB-02, SUB-03, SUB-04, SUB-05, SUB-06
**Success Criteria** (what must be TRUE):
  1. Operator can read an existing LAG by name via `flashblade_lag` data source and access its ports, speed, mac address, and status attributes
  2. Operator can create a subnet with name, prefix, gateway, mtu, vlan, and link_aggregation_group via `terraform apply` -- `apply -> plan` shows 0 diff
  3. Operator can update mutable subnet settings (gateway, prefix, mtu, vlan, link_aggregation_group) and destroy a subnet via Terraform
  4. Operator can import an existing subnet into Terraform state with no drift on subsequent `plan`
  5. Drift detection logs changes when a subnet is modified outside Terraform
**Plans**: 1 plan

Plans:
- [ ] 28-01-PLAN.md — Client models (LAG + Subnet), client CRUD methods, mock handlers
- [ ] 28-02-PLAN.md — LAG data source, Subnet resource, Subnet data source, unit tests, provider registration

### Phase 29: Network Interface Resource & Data Source
**Goal**: Operators can create and manage Virtual IP (VIP) network interfaces through Terraform with full CRUD, import, drift detection, and correct service/server semantics
**Depends on**: Phase 28 (subnet client in place -- network interface references subnet)
**Requirements**: NI-01, NI-02, NI-03, NI-04, NI-05, NI-06, NI-07, NI-08, NI-09, NI-10
**Success Criteria** (what must be TRUE):
  1. Operator can create a network interface with address, subnet, type, services, and attached_servers via `terraform apply` -- `apply -> plan` shows 0 diff
  2. Operator can update address, services, and attached_servers on an existing network interface; subnet and type changes force replacement (`RequiresReplace`)
  3. `terraform validate` rejects invalid service values and rejects attached_servers when service is egress-only or replication (plan-time, not API-time)
  4. Operator can import an existing network interface by its auto-assigned name (e.g., `vip0`) and subsequent `plan` shows 0 diff
  5. All computed read-only fields (enabled, gateway, mtu, netmask, vlan, realms) are populated after `terraform apply` and `terraform refresh`
**Plans**: 1 plan

Plans:
- [ ] 29-01-PLAN.md — Client models (NetworkInterfacePost/Patch/Get), client CRUD methods, mock handler
- [ ] 29-02-PLAN.md — Network interface resource (CRUD, validators, RequiresReplace, drift detection), data source, unit tests, provider registration

### Phase 30: Server Enrichment & Provider Registration
**Goal**: Operators can discover which VIPs are attached to a server directly from the server resource or data source, with correct schema migration on upgrade
**Depends on**: Phase 29 (network interface client in place for ListNetworkInterfaces call)
**Requirements**: SRV-01, SRV-02
**Success Criteria** (what must be TRUE):
  1. `flashblade_server` resource and data source expose a computed `network_interfaces` list populated from VIPs whose `attached_servers` includes that server
  2. Existing users upgrading the provider do not see a state deserialization error -- schema version bump 0->1 with StateUpgrader migrates old state by setting `network_interfaces` to an empty list
  3. `flashblade_network_interface` resource and `flashblade_subnet` resource are registered in `provider.go` and appear in `terraform providers` output
**Plans**: 1 plan

Plans:
- [ ] 30-01-PLAN.md — Server resource/data source enrichment (network_interfaces computed list, schema v0->v1 StateUpgrader, VIP client-side join, tests)

### Phase 31: Documentation & Workflow Examples
**Goal**: All new v2.1.1 resources have complete documentation, import guides, workflow examples, and the README reflects the expanded networking capabilities
**Depends on**: Phase 30 (all resources must exist and be registered before documentation)
**Requirements**: DOC-01, DOC-02, DOC-03, DOC-04
**Success Criteria** (what must be TRUE):
  1. Import documentation (import.sh) exists for all new importable resources (subnet, network interface) with correct syntax and realistic identifiers
  2. A workflow example in `examples/networking/` demonstrates the full stack: LAG data source -> subnet creation -> VIP creation -> server data source reading VIPs
  3. `tfplugindocs generate` produces documentation for all new resources and data sources without errors
  4. README coverage table includes the networking resources category with correct resource and data source counts
**Plans**: 1 plan

Plans:
- [ ] 31-01-PLAN.md — Import docs, networking workflow example, tfplugindocs regeneration, README update

## Progress

**Execution Order:**
Phases execute in numeric order: 23 -> 24 -> 25 -> 26 -> 27 -> 28 -> 29 -> 30 -> 31 -> 32 -> 33 -> 34 -> 35 -> 36 -> 37 -> 38 -> 39 -> 40 -> 41 -> 42 -> 43 -> 44 -> 45 -> 46 -> 47 -> 48

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 1. Foundation | v1.0 | 4/4 | Complete | 2026-03-27 |
| 2. Object Store Resources | v1.0 | 3/3 | Complete | 2026-03-27 |
| 3. File-Based Policy Resources | v1.0 | 4/4 | Complete | 2026-03-27 |
| 4. Object/Network/Quota & Array Admin | v1.0 | 5/5 | Complete | 2026-03-28 |
| 5. Quality Hardening | v1.0 | 4/4 | Complete | 2026-03-28 |
| 6. Server Resource & Export Consolidation | v1.1 | 2/2 | Complete | 2026-03-28 |
| 7. S3 Export Policies & Virtual Hosts | v1.1 | 3/3 | Complete | 2026-03-28 |
| 8. SMB Client Policies, Syslog & Acceptance Tests | v1.1 | 3/3 | Complete | 2026-03-28 |
| 9. Bug Fixes | v1.2 | 2/2 | Complete | 2026-03-28 |
| 10. Architecture Cleanup | v1.2 | 2/2 | Complete | 2026-03-28 |
| 11. Test Hardening & Validators | v1.2 | 3/3 | Complete | 2026-03-29 |
| 12. Infrastructure Hardening | v1.3 | 2/2 | Complete | 2026-03-29 |
| 13. Documentation & Sensitive Data | v1.3 | 2/2 | Complete | 2026-03-29 |
| 14. Access Key Enhancement & Array Connection | v2.0 | 2/2 | Complete | 2026-03-29 |
| 15. Replication Resources | v2.0 | 3/3 | Complete | 2026-03-29 |
| 16. Workflow & Documentation | v2.0 | 1/1 | Complete | 2026-03-29 |
| 17. Testing | v2.0 | 2/2 | Complete | 2026-03-29 |
| 18. Security & Auth Hardening | v2.0.1 | 1/1 | Complete | 2026-03-29 |
| 19. Error Handling & Consistency | v2.0.1 | 1/1 | Complete | 2026-03-29 |
| 20. Code Quality -- Validators & Dedup | v2.0.1 | 2/2 | Complete | 2026-03-29 |
| 21. Dead Code Removal & Modernization | v2.0.1 | 1/1 | Complete | 2026-03-29 |
| 22. Test Coverage | v2.0.1 | 2/2 | Complete | 2026-03-29 |
| 23. Bucket Inline Attributes | v2.1 | 2/2 | Complete | 2026-03-30 |
| 24. Lifecycle Rules | v2.1 | 2/2 | Complete | 2026-03-30 |
| 25. Bucket Access Policies | v2.1 | 2/2 | Complete | 2026-03-30 |
| 26. Audit Filters & QoS Policies | v2.1 | 3/3 | Complete | 2026-03-30 |
| 27. Testing & Documentation | v2.1 | 1/1 | Complete | 2026-03-30 |
| 28. LAG Data Source & Subnet Resource | v2.1.1 | 2/2 | Complete | 2026-03-31 |
| 29. Network Interface Resource & Data Source | v2.1.1 | 2/2 | Complete | 2026-03-31 |
| 30. Server Enrichment & Provider Registration | v2.1.1 | 1/1 | Complete | 2026-03-31 |
| 31. Documentation & Workflow Examples | v2.1.1 | 1/1 | Complete | 2026-03-31 |
| 32. Code Correctness Fixes | v2.1.3 | 1/1 | Complete | 2026-03-31 |
| 33. Client Hardening | v2.1.3 | 2/2 | Complete | 2026-03-31 |
| 34. Test Quality | v2.1.3 | 0/1 | Not started | - |
| 35. Object Store Users | v2.1.3 | 4/4 | Complete | 2026-04-01 |
| 36. Target Resource | v2.2 | 2/2 | Complete | 2026-04-02 |
| 37. Remote Credentials & Replica Link Enhancement | v2.2 | 1/1 | Complete | 2026-04-02 |
| 38. Documentation & Workflow | v2.2 | 1/1 | Complete | 2026-04-02 |
| 39. Certificates | v2.2 | 2/2 | Complete | 2026-04-03 |
| 40. TLS Policies | v2.2 | 2/2 | Complete | 2026-04-03 |
| 41. Certificate Groups | v2.2 | 2/2 | Complete | 2026-04-14 |
| 42. Array Connections | v2.2 | 2/2 | Complete | 2026-04-14 |
| 43. Shared Library | tools-v1.0 | 1/1 | Complete   | 2026-04-14 |
| 44. swagger-to-reference Skill | tools-v1.0 | 0/1 | Not started | - |
| 45. API Browsing Tools | tools-v1.0 | 0/1 | Not started | - |
| 46. api-diff Skill | tools-v1.0 | 0/1 | Not started | - |
| 47. api-upgrade Skill | tools-v1.0 | 0/1 | Not started | - |
| 48. Integration & Validation | tools-v1.0 | 0/1 | Not started | - |

---

## v2.1.3 Code Review Fixes & S3 Users (Phases 32-35)

**Milestone Goal:** Fix all issues identified by the full codebase code review — critical typos, dead schema attributes, diagnostic severity loss, client hardening, linting improvements, and acceptance test quality.

- [x] **Phase 32: Code Correctness Fixes** - Typo rename, dead schema removal, diagnostic severity, and dead helper/param cleanup (completed 2026-03-31)
- [x] **Phase 33: Client Hardening** - OAuth2 context propagation, RetryBaseDelay removal, and golangci-lint expansion (completed 2026-03-31)
- [ ] **Phase 34: Test Quality** - Fix ExpectNonEmptyPlan masking and expand acceptance test coverage
- [x] **Phase 35: Object Store Users** - S3 user resource, data source, and user-policy member resource (completed 2026-04-01)

### Phase 32: Code Correctness Fixes
**Goal**: All correctness issues found in review are resolved — the codebase compiles with no typos, carries no dead schema attributes, propagates diagnostic severity faithfully, and has no unused parameters or passthrough helpers
**Depends on**: Phase 31 (v2.1.1 complete)
**Requirements**: CC-01, CC-02, CC-03, CH-03, CL-01
**Success Criteria** (what must be TRUE):
  1. grep for FreezeLockgedObjects returns zero results — the correct spelling FreezeLockedObjects is used everywhere across structs, tests, and examples
  2. `terraform plan` on an existing filesystem resource shows 0 diff for nfs_export_policy and smb_share_policy — those attributes no longer exist in the schema
  3. When `readIntoState` encounters a warning diagnostic from `mapFSToModel`, the resulting Terraform diagnostic is a warning (not silently promoted to error or dropped)
  4. grep for extractEradicationConfig/extractObjectLockConfig/extractPublicAccessConfig shows ctx parameter removed from all three function signatures
  5. grep for mustObjectValue returns zero results — all callers use `types.ObjectValue` directly
**Plans**: 1 plan

Plans:
- [ ] 32-01-PLAN.md — FreezeLockgedObjects rename, dead schema removal, diagnostic severity fix, unused ctx removal, mustObjectValue elimination

### Phase 33: Client Hardening
**Goal**: The HTTP client and retry logic are free of fragile heuristics and context shortcuts — OAuth2 token refresh respects caller cancellation, retry delay is explicit, and the linter catches new categories of issues
**Depends on**: Phase 32 (code correctness fixes landed first)
**Requirements**: CH-01, CH-02, CL-02
**Success Criteria** (what must be TRUE):
  1. OAuth2 token refresh passes caller context to the HTTP token request — cancelling the Terraform context cancels the in-flight token refresh (not just the resource operation)
  2. RetryBaseDelay identifier is removed from the codebase — all callers pass explicit `time.Duration` values; `go build ./...` confirms no compilation errors
  3. `golangci-lint run ./...` passes with gosec, bodyclose, noctx, and exhaustive linters active — zero new violations introduced by this milestone
**Plans**: 2 plans

Plans:
- [ ] 33-01-PLAN.md — OAuth2 context propagation fix and RetryBaseDelay removal (CH-01, CH-02)
- [ ] 33-02-PLAN.md — golangci-lint config with gosec, bodyclose, noctx, exhaustive and violation fixes (CL-02)

### Phase 34: Test Quality
**Goal**: Acceptance tests accurately verify plan convergence and cover a broader set of high-risk resources — no test silently hides a drift bug, and critical resources beyond the initial 3 are exercised end-to-end
**Depends on**: Phase 33 (client changes landed so tests validate final behaviour)
**Requirements**: TQ-01, TQ-02
**Success Criteria** (what must be TRUE):
  1. grep for ExpectNonEmptyPlan in acceptance test files returns zero results — all tests either assert `ExpectNonEmptyPlan: false` explicitly or omit the field (defaulting to false)
  2. At least 3 additional resources (from: server, bucket replica link, network interface, or any policy family) have acceptance tests exercising plan, apply, refresh, import, and destroy via `resource.UnitTest` with a mock server
  3. All acceptance tests pass (go test ./... -run TestAcc) with zero failures after the convergence fix is applied
**Plans**: 1 plan

Plans:
- [ ] 34-01-PLAN.md — Remove ExpectNonEmptyPlan from existing acceptance tests, add acceptance tests for 3+ additional high-risk resources

### Phase 35: Object Store Users
**Goal**: Operators can create named S3 users, associate access policies to them, and manage per-user credentials — enabling multi-tenant S3 workflows with fine-grained access control
**Depends on**: Phase 33 (client hardening landed)
**Requirements**: OSU-01, OSU-02, OSU-03, OSU-04, OSU-05, OSU-06, OSU-07
**Success Criteria** (what must be TRUE):
  1. Operator can apply to create an S3 user `account/myuser` and destroy to delete it — full lifecycle works
  2. Operator can read an existing S3 user by name via `data.flashblade_object_store_user` with all attributes populated
  3. `terraform import flashblade_object_store_user.x account/username` populates state; subsequent plan shows 0 diff
  4. Operator can create a `flashblade_object_store_user_policy` resource associating a user to an access policy; destroy removes only the association
  5. Drift detection logs when a user or user-policy association is modified outside Terraform
**Plans**: 4 plans

Plans:
- [ ] 35-01-PLAN.md — Client layer: typed models, user-policy association methods, mock handlers, unit tests
- [ ] 35-02-PLAN.md — flashblade_object_store_user resource and data source, provider registration, examples
- [ ] 35-03-PLAN.md — flashblade_object_store_user_policy member resource, provider registration, examples
- [ ] 35-04-PLAN.md — Mocked provider tests for all three resources/data sources, ROADMAP.md update

## v2.2 S3 Target Replication & TLS (Phases 36-42)

**Milestone Goal:** Enable operators to replicate buckets to external S3-compatible endpoints (non-FlashBlade targets) and manage TLS certificates and policies for network interfaces through Terraform.

- [x] **Phase 36: Target Resource** - New resource and data source for managing external S3 endpoint targets (completed 2026-04-02)
- [x] **Phase 37: Remote Credentials & Replica Link Enhancement** - Extend existing resources for target references (completed 2026-04-02)
- [x] **Phase 38: Documentation & Workflow** - Import docs, workflow example, tfplugindocs (completed 2026-04-02)
- [x] **Phase 39: Certificates** - TLS certificate resource and data source with write-only sensitive fields (completed 2026-04-03)
- [x] **Phase 40: TLS Policies** - TLS policy resource, data source, and member resource (completed 2026-04-03)
- [x] **Phase 41: Certificate Groups** - Certificate group resource, data source, and member resource (completed 2026-04-14)
- [x] **Phase 42: Array Connections** - Array connection resource and data source with sensitive connection_key (completed 2026-04-14)

- [x] **Phase 36: Target Resource** - New resource and data source for managing external S3 endpoint targets (CRUD, import, drift detection) (completed 2026-04-02)
- [x] **Phase 37: Remote Credentials & Replica Link Enhancement** - Extend existing resources to support target references, enabling end-to-end S3 target replication (completed 2026-04-02)
- [x] **Phase 38: Documentation & Workflow** - Import docs, workflow example, and tfplugindocs generation for all new resources (completed 2026-04-02)
- [x] **Phase 39: Certificates** - Import and manage TLS certificates (appliance certs with PEM + private key); resource + data source (completed 2026-04-03)
- [x] **Phase 40: TLS Policies** - TLS policy CRUD, TLS policy member association (policy ↔ network interface); resource + data source + member resource (completed 2026-04-03)

### Phase 36: Target Resource
**Goal**: Operators can manage external S3 endpoint targets through Terraform with full CRUD, import, and drift detection
**Depends on**: Phase 35 (v2.1.3 complete)
**Requirements**: TGT-01, TGT-02, TGT-03, TGT-04, TGT-05
**Success Criteria** (what must be TRUE):
  1. Operator can create a target with name, address, and optional ca_certificate_group via `terraform apply` -- subsequent `plan` shows 0 diff
  2. Operator can update mutable target fields (address, ca_certificate_group) and destroy a target via `terraform apply` and `terraform destroy` without errors
  3. `terraform import flashblade_target.x target-name` populates all attributes; subsequent `plan` shows 0 diff
  4. `data.flashblade_target` data source reads an existing target by name and exposes address, status, and status_details attributes
  5. Drift detection logs field-level changes via tflog when a target is modified outside Terraform
**Plans**: 2 plans

Plans:
- [ ] 36-01-PLAN.md — Client models (TargetPost/Patch/Get), client CRUD methods, mock handler, unit tests
- [ ] 36-02-PLAN.md — flashblade_target resource (CRUD, import, drift detection), data source, provider registration

### Phase 37: Remote Credentials & Replica Link Enhancement
**Goal**: Operators can create remote credentials referencing a target (not just an array connection), and existing bucket replica links work end-to-end against external S3 endpoints with no regression on array-to-array replication
**Depends on**: Phase 36 (target resource must exist for remote credentials to reference)
**Requirements**: RC-01, RC-02, BRL-01
**Success Criteria** (what must be TRUE):
  1. Operator can create `flashblade_object_store_remote_credentials` with a `target` reference (instead of an array connection) -- `apply -> plan` shows 0 diff
  2. Existing remote credentials referencing array connections continue to work unchanged after the enhancement -- `apply -> plan` shows 0 diff for pre-existing configs
  3. Operator can create a bucket replica link using remote credentials that reference a target -- end-to-end replication to an external S3 endpoint completes without provider errors
**Plans**: 1 plan

Plans:
- [ ] 37-01-PLAN.md — Remote credentials target support (schema extension, client update, mock handler, unit tests) + BRL-01 validation and tests

### Phase 38: Documentation & Workflow
**Goal**: All new v2.2 resources have complete import documentation, a workflow example demonstrates the full S3 target replication stack, and tfplugindocs generates without errors
**Depends on**: Phase 37 (all resources must exist before documentation and examples)
**Requirements**: DOC-01, DOC-02, DOC-03
**Success Criteria** (what must be TRUE):
  1. Import documentation (import.sh) exists for `flashblade_target` with correct `terraform import` syntax and a realistic identifier
  2. A workflow example in `examples/s3-target-replication/` demonstrates the full stack: target creation, remote credentials referencing the target, and a bucket replica link to the external S3 endpoint
  3. `tfplugindocs generate` produces documentation for all new resources and data sources without errors and without manual edits to the docs/ directory
**Plans**: 1 plan

Plans:
- [ ] 38-01-PLAN.md — import.sh for flashblade_target, s3-target-replication workflow example, tfplugindocs regeneration

### Phase 39: Certificates
**Goal**: Operators can import and manage TLS certificates on a FlashBlade through Terraform with full CRUD, import, and drift detection
**Depends on**: Phase 38 (v2.2 target replication complete)
**Requirements**: CERT-01, CERT-02, CERT-03, CERT-04, CERT-05
**Success Criteria** (what must be TRUE):
  1. Operator can import a certificate with name, PEM certificate body, private key, and optional intermediate certificate via `terraform apply` -- subsequent `plan` shows 0 diff
  2. Operator can update mutable certificate fields and destroy a certificate via `terraform apply` and `terraform destroy` without errors
  3. `terraform import flashblade_certificate.x cert-name` populates all non-sensitive attributes; subsequent `plan` shows 0 diff
  4. `data.flashblade_certificate` data source reads an existing certificate by name and exposes type, status, issuer, validity, and SAN attributes
  5. Drift detection logs field-level changes via tflog when a certificate is modified outside Terraform
  6. Private key and passphrase are marked Sensitive and never appear in plan output or logs
**Plans**: 2 plans

Plans:
- [x] 39-01-PLAN.md — Client models (Certificate/Post/Patch), client CRUD methods, mock handler, unit tests
- [x] 39-02-PLAN.md — flashblade_certificate resource (CRUD, import, drift detection), data source, provider registration, examples, docs

### Phase 40: TLS Policies
**Goal**: Operators can manage TLS policies and assign them to network interfaces through Terraform, controlling cipher suites, minimum TLS version, mutual TLS settings, and appliance certificate selection
**Depends on**: Phase 39 (certificates must exist for TLS policy to reference)
**Requirements**: TLSP-01, TLSP-02, TLSP-03, TLSP-04, TLSP-05, TLSP-06
**Success Criteria** (what must be TRUE):
  1. Operator can create a TLS policy with name, appliance_certificate, min_tls_version, cipher lists, and mTLS settings via `terraform apply` -- subsequent `plan` shows 0 diff
  2. Operator can update all mutable TLS policy fields and destroy a policy via `terraform apply` and `terraform destroy` without errors
  3. `terraform import flashblade_tls_policy.x policy-name` populates all attributes; subsequent `plan` shows 0 diff
  4. `data.flashblade_tls_policy` data source reads an existing TLS policy by name and exposes all configuration attributes
  5. Operator can assign a TLS policy to a network interface via `flashblade_tls_policy_member` and remove the assignment via `terraform destroy`
  6. Drift detection logs field-level changes via tflog when a TLS policy is modified outside Terraform
**Plans**: 2 plans

Plans:
- [x] 40-01-PLAN.md — Client models (TlsPolicy/TlsPolicyPost/TlsPolicyPatch/TlsPolicyMember), client CRUD + member methods, mock handler, unit tests
- [ ] 40-02-PLAN.md — flashblade_tls_policy resource (CRUD, import, drift detection), data source, flashblade_tls_policy_member resource, provider registration, examples, docs

### Phase 41: Certificate Groups
**Goal**: Operators can manage certificate groups and their certificate memberships through Terraform, enabling CA certificate trust bundles for targets, array connections, and directory services
**Depends on**: Phase 39 (certificates must exist for group membership)
**Requirements**: CERTG-01, CERTG-02, CERTG-03, CERTG-04, CERTG-05
**Success Criteria** (what must be TRUE):
  1. Operator can create a certificate group by name via `terraform apply` -- subsequent `plan` shows 0 diff
  2. Operator can destroy a certificate group via `terraform destroy` without errors
  3. `terraform import flashblade_certificate_group.x group-name` populates all attributes; subsequent `plan` shows 0 diff
  4. `data.flashblade_certificate_group` data source reads an existing group by name and exposes id, name, and realms
  5. Operator can add a certificate to a group via `flashblade_certificate_group_member` and remove it via `terraform destroy`
  6. Drift detection logs field-level changes via tflog when a certificate group is modified outside Terraform
**Plans**: 2 plans

Plans:
- [x] 41-01-PLAN.md — Client layer: CertificateGroup/CertificateGroupPost/CertificateGroupMember models, client CRUD + member methods, mock handler + facade, 7 unit tests
- [x] 41-02-PLAN.md — Provider layer: flashblade_certificate_group resource, data source, flashblade_certificate_group_member resource, tests, registration, HCL examples, make docs

### Phase 42: Array Connections
**Goal**: Operators can manage FlashBlade array connections through Terraform with full CRUD, enabling inter-array replication with connection key exchange, encryption, CA certificate group assignment, replication addresses, and bandwidth throttling
**Depends on**: Phase 41 (certificate groups for ca_certificate_group reference)
**Requirements**: ARRC-01, ARRC-02, ARRC-03, ARRC-04, ARRC-05
**Success Criteria** (what must be TRUE):
  1. Operator can create an array connection with management_address, connection_key, encrypted, and optional ca_certificate_group/replication_addresses via `terraform apply` -- subsequent `plan` shows 0 diff
  2. Operator can update mutable fields (management_address, encrypted, ca_certificate_group, replication_addresses, throttle) and destroy a connection via `terraform apply` and `terraform destroy` without errors
  3. `terraform import flashblade_array_connection.x remote-name` populates all non-sensitive attributes; subsequent `plan` shows 0 diff
  4. `data.flashblade_array_connection` data source reads an existing connection by remote name and exposes all configuration and status attributes
  5. connection_key is marked Sensitive and never appears in plan output or logs; it is write-only (POST only, not returned by GET)
  6. Drift detection logs field-level changes via tflog when an array connection is modified outside Terraform
**Plans**: 2 plans

Plans:
- [x] 42-01-PLAN.md — Client layer: ArrayConnection models extended, Post/Patch/Delete client methods, full CRUD mock handler, 7 unit tests
- [x] 42-02-PLAN.md — Provider layer: flashblade_array_connection resource (full CRUD, import, sensitive connection_key), data source extended, registration, HCL examples, make docs


## tools-v1.0 API Tooling Pipeline (Phases 43-48)

**Milestone Goal:** Automate swagger-to-reference conversion, API version diffing, and provider upgrade orchestration through Claude Code skills with Python tooling.

- [x] **Phase 43: Shared Library** - `_shared/swagger_utils.py` with allOf resolver, path normalizer, schema flattener (completed 2026-04-14)
- [ ] **Phase 44: swagger-to-reference Skill** - `parse_swagger.py` + SKILL.md converting swagger.json to AI-optimized markdown
- [ ] **Phase 45: API Browsing Tools** - `browse_api.py` with search, schema display, comparison, and statistics
- [ ] **Phase 46: api-diff Skill** - `diff_swagger.py` + migration plan generator + SKILL.md
- [ ] **Phase 47: api-upgrade Skill** - `upgrade_version.py` + orchestration SKILL.md with review gates
- [ ] **Phase 48: Integration & Validation** - CLAUDE.md update, SKILL.md finalization, end-to-end validation on swagger-2.22.json and swagger-2.23.json

### Phase 43: Shared Library
**Goal**: All Python tooling shares a single, well-tested utility library that resolves allOf schemas, normalizes API paths, and flattens nested schemas
**Depends on**: Phase 42 (v2.2 complete)
**Requirements**: SLIB-01, SLIB-02
**Success Criteria** (what must be TRUE):
  1. `python3 -c "from _shared.swagger_utils import resolve_all_of, normalize_path, flatten_schema"` runs with no import errors using stdlib only
  2. `resolve_all_of` correctly flattens allOf chains found in swagger-2.22.json (404/709 schemas use allOf) — output contains no unresolved `$ref` or `allOf` keys
  3. `normalize_path("/api/2.22/buckets")` returns `"buckets"` — version prefix stripped deterministically
**Plans**: 1 plan

Plans:
- [x] 43-01-PLAN.md — swagger_utils.py implementation and test suite

### Phase 44: swagger-to-reference Skill
**Goal**: Claude can convert any FlashBlade swagger.json into the AI-optimized markdown format matching FLASHBLADE_API.md, with correct allOf resolution and versioned output
**Depends on**: Phase 43 (shared library must exist)
**Requirements**: CONV-01, CONV-02, CONV-03, CONV-04, INTG-02 (partial)
**Success Criteria** (what must be TRUE):
  1. Running the skill on swagger-2.22.json produces `api_references/2.22.md` with a path count matching FLASHBLADE_API.md (226 paths) and no unresolved `$ref` or `allOf` keys in the output
  2. The skill prompts for the API version string before processing and uses it as the output filename
  3. The generated markdown format matches FLASHBLADE_API.md structure (headers, method grouping, schema tables) — a diff between the generated file and the original shows only cosmetic whitespace, not structural differences
  4. swagger-to-reference SKILL.md exists with valid YAML frontmatter and structured activation, steps, and output sections
**Plans**: TBD
**UI hint**: no

### Phase 45: API Browsing Tools
**Goal**: Claude can search, inspect, and compare API endpoints and schemas interactively from a generated reference file
**Depends on**: Phase 44 (reference file must exist to browse)
**Requirements**: BRWS-01, BRWS-02, BRWS-03, BRWS-04
**Success Criteria** (what must be TRUE):
  1. `python3 browse_api.py --tag buckets` lists all endpoints tagged `buckets` with their HTTP methods and summary lines
  2. `python3 browse_api.py --schema BucketPost` displays all fields with types, readOnly annotations, and required flags
  3. `python3 browse_api.py --compare BucketPost BucketPatch` shows a side-by-side diff table highlighting fields present in one but not the other, and type mismatches
  4. `python3 browse_api.py --stats` outputs path count, schema count, and method distribution (GET/POST/PATCH/DELETE counts)
**Plans**: TBD

### Phase 46: api-diff Skill
**Goal**: Claude can produce a structured diff between two swagger versions, annotate discrepancies, and generate a migration plan cross-referenced with ROADMAP.md
**Depends on**: Phase 43 (shared library for path normalization)
**Requirements**: DIFF-01, DIFF-02, DIFF-03, DIFF-04, INTG-02 (partial)
**Success Criteria** (what must be TRUE):
  1. Running the skill on swagger-2.22.json and swagger-2.23.json produces a diff listing new endpoints, removed endpoints, and modified schemas — no duplicates caused by version prefix differences
  2. Each diff item is annotated as `real_change`, `swagger_artifact`, or `needs_verification` based on `known_discrepancies.md` lookup
  3. The migration plan output cross-references ROADMAP.md entries — new endpoints that match a "Not Implemented" roadmap entry are flagged as candidates
  4. api-diff SKILL.md exists with valid YAML frontmatter and structured activation, steps, and output sections
**Plans**: TBD

### Phase 47: api-upgrade Skill
**Goal**: Claude can mechanically update API version references across the codebase and orchestrate the provider upgrade sequence with explicit review gates
**Depends on**: Phase 46 (diff output informs upgrade scope)
**Requirements**: UPGR-01, UPGR-02, UPGR-03, INTG-02 (partial)
**Success Criteria** (what must be TRUE):
  1. `python3 upgrade_version.py --from 2.22 --to 2.23 --dry-run` lists every file and line that would change (APIVersion const, mock server version strings, mock handler paths) without modifying any file
  2. `python3 upgrade_version.py --from 2.22 --to 2.23 --apply` applies all changes and `make build` passes with no compilation errors
  3. api-upgrade SKILL.md exists with 5 named phases (infra, schemas, new resources, deprecations, docs), each with explicit review gate instructions and acceptance criteria
**Plans**: TBD

### Phase 48: Integration & Validation
**Goal**: All three skills are documented in CLAUDE.md, the `api_references/` convention is established, and the full pipeline is validated end-to-end on real swagger files
**Depends on**: Phase 47 (all skills must exist before integration)
**Requirements**: INTG-01, INTG-02 (finalize)
**Success Criteria** (what must be TRUE):
  1. CLAUDE.md references the three skills (swagger-to-reference, api-diff, api-upgrade) with one-line descriptions and the `api_references/` output convention
  2. Running the full pipeline on swagger-2.22.json and swagger-2.23.json end-to-end (convert → browse → diff → migration plan) completes without errors and produces readable artifacts
  3. All three SKILL.md files pass YAML frontmatter validation and follow the skill-creator format consistently
**Plans**: TBD
