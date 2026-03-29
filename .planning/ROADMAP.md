# Roadmap: Terraform Provider FlashBlade

## Milestones

- v1.0 Core Provider (Phases 1-5) -- shipped 2026-03-28
- v1.1 Servers & Exports (Phases 6-8) -- shipped 2026-03-28
- v1.2 Code Quality & Robustness (Phases 9-11) -- shipped 2026-03-29
- v1.3 Release Readiness (Phases 12-13) -- shipped 2026-03-29
- v2.0 Cross-Array Bucket Replication (Phases 14-17) -- shipped 2026-03-29
- v2.0.1 Quality & Hardening (Phases 18-22) -- in progress

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

### v2.0.1 Quality & Hardening (In Progress)

**Milestone Goal:** Harden the codebase post-v2.0 release with security fixes, code quality improvements, dead code removal, duplication reduction, and test coverage gap closure identified by comprehensive 5-agent audit.

- [x] **Phase 18: Security & Auth Hardening** - Sanitize OAuth2 errors, add context propagation, set HTTP safety-net timeout (completed 2026-03-29)
- [x] **Phase 19: Error Handling & Consistency** - Migrate to errors.As(), fix ParseAPIError, harden bucket delete and test helpers (completed 2026-03-29)
- [x] **Phase 20: Code Quality -- Validators & Deduplication** - Compile regex at init, extract 8 shared helpers, safer mustObjectValue (completed 2026-03-29)
- [x] **Phase 21: Dead Code Removal & Modernization** - Remove unused List* functions, SourceReference, empty UpgradeState; update math/rand (completed 2026-03-29)
- [ ] **Phase 22: Test Coverage** - Unit tests for 5 uncovered data sources, OAuth2 config test, HCL acceptance tests, pagination tests

## Phase Details

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
**Plans**: 1 plan

Plans:
- [ ] 18-01-PLAN.md — OAuth2 error sanitization, TLS warning, context propagation, HTTP timeout, API token request cleanup

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
**Plans**: 1 plan

Plans:
- [ ] 19-01-PLAN.md — errors.As() migration, ParseAPIError hardening, bucket delete guard, countItems fix

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
**Plans**: 2 plans

Plans:
- [ ] 20-01-PLAN.md — Regex pre-compilation, space schema helpers, nullTimeoutsValue, DiagnosticReporter interface
- [ ] 20-02-PLAN.md — Generic getOneByName[T], pollUntilGone[T], mapFSToModel sharing, mustObjectValue diagnostics

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
**Plans**: 1 plan

Plans:
- [ ] 21-01-PLAN.md — Remove unused List* functions, IsUnprocessable, SourceReference, empty UpgradeState, update math/rand

### Phase 22: Test Coverage
**Goal**: All data sources and auth paths have unit tests, and HCL-based acceptance tests validate the full Terraform lifecycle through the provider
**Depends on**: Phase 21 (all code changes landed so tests validate final state)
**Requirements**: TST-01, TST-02, TST-03, TST-04, TST-05, TST-06, TST-07, TST-08
**Success Criteria** (what must be TRUE):
  1. Unit tests exist for all 5 previously uncovered data sources (virtual host, remote credentials, bucket replica link, file system export, account export) covering Read success and NotFound error paths
  2. OAuth2 provider configuration test verifies that `client_id` + `key_id` + `issuer` flow initializes the provider without errors (mock token endpoint)
  3. At least 3 resources have HCL-based acceptance tests using `resource.UnitTest` with a mock server that exercise plan, apply, refresh, import, and destroy
  4. Pagination tests exist for at least buckets and one policy type (in addition to existing filesystem pagination tests)
**Plans**: 1 plan

Plans:
- [ ] 22-01-PLAN.md — Unit tests for 5 uncovered data sources + OAuth2 provider config test
- [ ] 22-02-PLAN.md — HCL-based acceptance tests with mock server + pagination tests for buckets and policies

## Progress

**Execution Order:**
Phases execute in numeric order: 18 -> 19 -> 20 -> 21 -> 22

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
| 18. Security & Auth Hardening | 1/1 | Complete    | 2026-03-29 | - |
| 19. Error Handling & Consistency | 1/1 | Complete    | 2026-03-29 | - |
| 20. Code Quality -- Validators & Dedup | 2/2 | Complete    | 2026-03-29 | - |
| 21. Dead Code Removal & Modernization | 1/1 | Complete    | 2026-03-29 | - |
| 22. Test Coverage | v2.0.1 | 0/2 | Not started | - |
