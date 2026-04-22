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
- tools-v1.0 API Tooling Pipeline (Phases 43-48) -- shipped 2026-04-14
- ✅ v2.22.1 Directory Service – Array Management (Phases 49-49) -- shipped 2026-04-17 — [archive](milestones/v2.22.1-ROADMAP.md)
- ✅ v2.22.2 Directory Service Roles & Role Mappings (Phases 50, 50.1) -- shipped 2026-04-17 — [archive](milestones/v2.22.2-ROADMAP.md)
- ✅ v2.22.3 Convention Compliance (Phases 51-53) -- shipped 2026-04-20 — [archive](milestones/v2.22.3-ROADMAP.md)
- pulumi-2.22.3 Pulumi Bridge Alpha (Phases 54-58) -- in progress

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

### Phase 3: Policy Resources
**Goal**: Operators can attach and manage all six policy families (NFS export, SMB share, snapshot, OAP, NAP, quota) through Terraform, completing the full storage policy matrix
**Depends on**: Phase 2
**Requirements**: NFX-01, NFX-02, NFX-03, SMB-01, SMB-02, SMB-03, SNAP-01, SNAP-02, SNAP-03, OAP-01, OAP-02, OAP-03, NAP-01, NAP-02, NAP-03, QTA-01, QTA-02, QTA-03
**Success Criteria** (what must be TRUE):
  1. Operator can create a policy, add rules to it, and attach it to a file system or bucket in a single `terraform apply`
  2. Policy rule deletion and re-creation does not leave orphaned rules; `terraform plan` after apply shows 0 diff
  3. `terraform import` of a policy rule uses the composite ID format `policy_name:rule_index`
**Plans:** 3/3 plans complete

Plans:
- [x] 03-01-PLAN.md — NFS export policy + SMB share policy: models, client, mock, resource, data source
- [x] 03-02-PLAN.md — Snapshot policy + OAP (object access policy) + NAP (network access policy): models, client, mock, resource, data source
- [x] 03-03-PLAN.md — Quota policy: models, client, mock, resource, data source + policy rule pattern consolidation

### Phase 4: Array Administration
**Goal**: Operators can manage array-level configuration (DNS, NTP, SMTP) and generate full documentation for all resources
**Depends on**: Phase 3
**Requirements**: DNS-01, DNS-02, NTP-01, NTP-02, SMTP-01, SMTP-02, SMTP-03, SMTP-04
**Success Criteria** (what must be TRUE):
  1. Operator can configure DNS, NTP, and SMTP in a single `terraform apply`; subsequent plan shows 0 diff
  2. `terraform import` works for all three array resources
  3. `make docs` regenerates all resource + data source docs without manual edits, matching the live resource schemas
**Plans:** 2/2 plans complete

Plans:
- [x] 04-01-PLAN.md — DNS, NTP, SMTP singleton resources: models, client, mock, resource, data source
- [x] 04-02-PLAN.md — Documentation generation, HCL examples, import.sh for all 22 resources

### Phase 5: Stabilization & Live Testing
**Goal**: All 22 resources pass acceptance tests against a real FlashBlade, with 14 identified bugs fixed and CI pipeline green
**Depends on**: Phase 4
**Requirements**: TEST-01, TEST-02, TEST-03, TEST-04, TEST-05, TEST-06, TEST-07, TEST-08, TEST-09, TEST-10, TEST-11, TEST-12, TEST-13, TEST-14
**Success Criteria** (what must be TRUE):
  1. `make test` passes all 227 unit tests with 0 failures
  2. Acceptance test HCL (`examples/`) applies cleanly against a live FlashBlade with 0 errors
  3. All 14 identified bugs are closed and their fix commits reference the bug ID
**Plans:** 5/5 plans complete

Plans:
- [x] 05-01-PLAN.md — Bug fixes: DNS singleton race, bucket quota drift, access key re-read
- [x] 05-02-PLAN.md — Bug fixes: NFS rule index collision, SMB share policy attach
- [x] 05-03-PLAN.md — Bug fixes: OAP/NAP member resource ordering, snapshot rule time parse
- [x] 05-04-PLAN.md — Bug fixes: SMTP auth fields, quota policy scope validation
- [x] 05-05-PLAN.md — CI pipeline: GitHub Actions, golangci-lint, tfplugindocs generate check

</details>

<details>
<summary>v1.1 Servers & Exports (Phases 6-8) - SHIPPED 2026-03-28</summary>

### Phase 6: Server Resource
**Goal**: Operators can register and manage FlashBlade servers (compute nodes/clients) through Terraform with full CRUD, import, and drift detection
**Depends on**: Phase 5
**Requirements**: SRV-01, SRV-02, SRV-03, SRV-04, SRV-05
**Success Criteria** (what must be TRUE):
  1. Operator can create a server with IQN/WWN/NFS client list and it appears in FlashBlade after apply
  2. `terraform import flashblade_server.x name` populates all attributes; subsequent plan shows 0 diff
  3. Updating client list triggers PATCH with only changed fields; plan shows accurate diff
**Plans:** 2/2 plans complete

Plans:
- [x] 06-01-PLAN.md — Server: models, client CRUD, mock handler, resource with full CRUD/import/drift
- [x] 06-02-PLAN.md — Server data source + acceptance test HCL

### Phase 7: Export Infrastructure
**Goal**: Operators can manage S3 export policies, virtual hosts, and SMB client policies through Terraform — completing the export layer for both object and file workloads
**Depends on**: Phase 6
**Requirements**: S3EX-01, S3EX-02, S3EX-03, VH-01, VH-02, VH-03, SMBCL-01, SMBCL-02, SMBCL-03
**Success Criteria** (what must be TRUE):
  1. Operator can create an S3 export policy, add a rule, and attach a virtual host in a single apply
  2. SMB client policy rules are ordered correctly; import uses composite ID `policy_name:rule_index`
  3. `terraform destroy` removes all resources cleanly without dependency ordering errors
**Plans:** 3/3 plans complete

Plans:
- [x] 07-01-PLAN.md — S3 export policy + rule: models, client, mock, resource, data source
- [x] 07-02-PLAN.md — Virtual host: models, client, mock, resource, data source
- [x] 07-03-PLAN.md — SMB client policy + rule: models, client, mock, resource, data source

### Phase 8: Syslog & File Export
**Goal**: Operators can manage syslog servers and file system / account exports through Terraform, completing the operations and export surface
**Depends on**: Phase 7
**Requirements**: SYSLOG-01, SYSLOG-02, FSEX-01, FSEX-02, ACEX-01, ACEX-02
**Success Criteria** (what must be TRUE):
  1. Operator can add a syslog server and it receives log events after apply
  2. File system export and account export apply/destroy cleanly; import works by name
  3. All 268 unit tests pass; 26 resources verified against live FlashBlade
**Plans:** 3/3 plans complete

Plans:
- [x] 08-01-PLAN.md — Syslog server: models, client, mock, resource, data source
- [x] 08-02-PLAN.md — File system export: models, client, mock, resource (TDD consolidated)
- [x] 08-03-PLAN.md — Account export: models, client, mock, resource, data source

</details>

<details>
<summary>v1.2 Code Quality & Robustness (Phases 9-11) - SHIPPED 2026-03-29</summary>

### Phase 9: Bug Fixes
**Goal**: All latent bugs identified post-v1.1 are fixed and regression-tested
**Depends on**: Phase 8
**Requirements**: BUG-01, BUG-02, BUG-03, BUG-04
**Success Criteria** (what must be TRUE):
  1. Account export Delete no longer errors on 404 (idempotent destroy)
  2. File system `writable` drift is detected correctly in Read
  3. `IsNotFound` only matches HTTP 404 (not 400/500); `omitempty` applied to optional POST fields
**Plans:** 2/2 plans complete

Plans:
- [x] 09-01-PLAN.md — Account export delete fix + filesystem writable drift fix
- [x] 09-02-PLAN.md — IsNotFound scope fix + omitempty audit

### Phase 10: Architecture Consolidation
**Goal**: Shared helpers are extracted, models.go is split into domain files, and all mock handlers are hardened with query param validation
**Depends on**: Phase 9
**Requirements**: ARCH-01, ARCH-02, ARCH-03, ARCH-04, ARCH-05
**Success Criteria** (what must be TRUE):
  1. `models.go` is replaced by 5 domain files; no compilation errors
  2. `compositeID` and `stringOrNull` helpers used by all resources that need them
  3. All mock handlers reject unknown query params with 400; `?filter=` validated where applicable
**Plans:** 2/2 plans complete

Plans:
- [x] 10-01-PLAN.md — models.go split + shared helpers (compositeID, stringOrNull)
- [x] 10-02-PLAN.md — Mock handler hardening with query param validation

### Phase 11: Validators & Test Coverage
**Goal**: Custom validators are in place, idempotence tests cover all stateful operations, and test count reaches 329
**Depends on**: Phase 10
**Requirements**: VAL-01, VAL-02, VAL-03, COV-01, COV-02
**Success Criteria** (what must be TRUE):
  1. `Alphanumeric` and `HostnameNoDot` validators reject invalid input with clear error messages
  2. 9 idempotence tests cover apply → plan → 0 diff for the most complex resources
  3. `make test` reports 329 tests, 0 failures
**Plans:** 3/3 plans complete

Plans:
- [x] 11-01-PLAN.md — Alphanumeric + HostnameNoDot validators + enum OneOf validators
- [x] 11-02-PLAN.md — Idempotence tests for file system, bucket, policies
- [x] 11-03-PLAN.md — Idempotence tests for server, exports, syslog + test count audit

</details>

<details>
<summary>v1.3 Release Readiness (Phases 12-13) - SHIPPED 2026-03-29</summary>

### Phase 12: State Migration Framework
**Goal**: All 28 resources have SchemaVersion 0 + empty UpgradeState, establishing the migration chain foundation for future schema changes
**Depends on**: Phase 11
**Requirements**: MIGR-01, MIGR-02, MIGR-03
**Success Criteria** (what must be TRUE):
  1. Every resource file declares `SchemaVersion: 0` and an empty `UpgradeState` map
  2. `make test` still passes 340 tests (no regressions from schema changes)
  3. `int64UseStateForUnknown` and `float64UseStateForUnknown` helpers extracted and used consistently
**Plans:** 2/2 plans complete

Plans:
- [x] 12-01-PLAN.md — SchemaVersion 0 + UpgradeState on all 28 resources
- [x] 12-02-PLAN.md — UseStateForUnknown helpers consolidation

### Phase 13: Release Infrastructure
**Goal**: The provider is ready for Terraform Registry publication with import docs, jitter backoff, and GoReleaser + Cosign pipeline
**Depends on**: Phase 12
**Requirements**: REL-01, REL-02, REL-03, REL-04
**Success Criteria** (what must be TRUE):
  1. `make docs` regenerates all 28 resource docs + 27 import.sh files without errors
  2. GoReleaser + Cosign pipeline produces signed binaries for linux/darwin amd64/arm64
  3. Exponential backoff includes ±20% jitter; retry behavior validated in unit test
**Plans:** 2/2 plans complete

Plans:
- [x] 13-01-PLAN.md — GoReleaser + Cosign pipeline + jitter backoff
- [x] 13-02-PLAN.md — 27 import.sh files + tfplugindocs regeneration

</details>

<details>
<summary>v2.0 Cross-Array Bucket Replication (Phases 14-17) - SHIPPED 2026-03-29</summary>

### Phase 14: Access Key Enhancement
**Goal**: Operators can provide an existing secret access key during access key creation, enabling cross-array credential sharing for replication setup
**Depends on**: Phase 13
**Requirements**: OAK-06
**Success Criteria** (what must be TRUE):
  1. `flashblade_object_store_access_key` accepts optional `secret_access_key` input; when provided, it is sent in the POST body and stored sensitive in state
  2. When not provided, behavior is unchanged (API generates the key)
  3. `terraform plan` never shows `secret_access_key` in plain text
**Plans:** 1/1 plans complete

Plans:
- [x] 14-01-PLAN.md — Access key optional secret_access_key input field

### Phase 15: Remote Credentials Resource
**Goal**: Operators can create and manage object store remote credentials (cross-array access key pairs) through Terraform
**Depends on**: Phase 14
**Requirements**: RC-01, RC-02, RC-03, RC-04, RC-05
**Success Criteria** (what must be TRUE):
  1. Operator can create remote credentials referencing an access key ID + secret; credentials appear in FlashBlade after apply
  2. `secret_access_key` is write-only — never returned by API, not in plan diff, sensitive in state
  3. `terraform import flashblade_object_store_remote_credentials.x name` populates all non-secret fields; subsequent plan shows 0 diff
**Plans:** 2/2 plans complete

Plans:
- [x] 15-01-PLAN.md — Remote credentials: models, client CRUD, mock handler
- [x] 15-02-PLAN.md — Remote credentials resource + data source + import

### Phase 16: Bucket Replica Link Resource
**Goal**: Operators can establish bidirectional S3 bucket replication between two FlashBlade arrays through Terraform
**Depends on**: Phase 15
**Requirements**: RL-01, RL-02, RL-03, RL-04, RL-05
**Success Criteria** (what must be TRUE):
  1. Operator can create a replica link specifying local bucket, remote bucket, remote credentials, and direction; link appears active after apply
  2. `terraform destroy` cleanly removes the replica link
  3. `terraform import flashblade_bucket_replica_link.x name` populates all attributes; subsequent plan shows 0 diff
**Plans:** 2/2 plans complete

Plans:
- [x] 16-01-PLAN.md — Bucket replica link: models, client CRUD, mock handler
- [x] 16-02-PLAN.md — Bucket replica link resource + data source + import

### Phase 17: Array Connection Data Source & Workflow
**Goal**: Operators can read existing array connections and complete the full replication workflow via Terraform
**Depends on**: Phase 16
**Requirements**: AC-01, AC-02, WF-01
**Success Criteria** (what must be TRUE):
  1. `flashblade_array_connection` data source reads an existing inter-array connection by name
  2. Complete dual-provider replication workflow example applies cleanly end-to-end
  3. All 368 unit tests pass; `make docs` regenerated
**Plans:** 2/2 plans complete

Plans:
- [x] 17-01-PLAN.md — Array connection data source: models, client read, mock handler, data source
- [x] 17-02-PLAN.md — Dual-provider replication workflow example + test count audit

</details>

<details>
<summary>v2.0.1 Quality & Hardening (Phases 18-22) - SHIPPED 2026-03-30</summary>

### Phase 18: Security Hardening
**Goal**: Auth paths are secure, context-propagated, and time-bounded; error messages sanitize sensitive values
**Depends on**: Phase 17
**Requirements**: SEC-01, SEC-02, SEC-03, SEC-04
**Success Criteria** (what must be TRUE):
  1. OAuth2 error messages never include raw credentials or token values
  2. All auth paths propagate caller context (no `context.Background()` in auth code)
  3. HTTP client has a 30s timeout; timeout is validated in a unit test
**Plans:** 2/2 plans complete

Plans:
- [x] 18-01-PLAN.md — OAuth2 error sanitization + context propagation
- [x] 18-02-PLAN.md — 30s HTTP timeout + ParseAPIError hardening

### Phase 19: Code Quality
**Goal**: Error handling uses modern Go patterns, dead code is removed, and math/rand is modernized
**Depends on**: Phase 18
**Requirements**: QUAL-01, QUAL-02, QUAL-03, QUAL-04
**Success Criteria** (what must be TRUE):
  1. All error type assertions use `errors.As()` (not type assertions)
  2. ~405 lines of dead code removed; `make build` still passes
  3. `math/rand/v2` used throughout; no deprecated rand functions
**Plans:** 2/2 plans complete

Plans:
- [x] 19-01-PLAN.md — errors.As migration + dead code removal
- [x] 19-02-PLAN.md — math/rand/v2 modernization + fresh-GET bucket delete guard

### Phase 20: Shared Helpers
**Goal**: Eight shared helper functions are extracted and used consistently across all resources, eliminating duplication
**Depends on**: Phase 19
**Requirements**: HELP-01, HELP-02, HELP-03, HELP-04, HELP-05, HELP-06, HELP-07, HELP-08
**Success Criteria** (what must be TRUE):
  1. `getOneByName[T]`, `pollUntilGone[T]`, `nullTimeoutsValue()`, and `spaceAttrTypes` used by all applicable resources
  2. No resource hand-rolls list+filter logic that `getOneByName[T]` could replace
  3. `make test` passes 394 tests (16 new) with 0 failures
**Plans:** 2/2 plans complete

Plans:
- [x] 20-01-PLAN.md — getOneByName[T] + pollUntilGone[T] + spaceAttrTypes helpers
- [x] 20-02-PLAN.md — nullTimeoutsValue + 5 additional shared helpers

### Phase 21: Bug Fixes
**Goal**: Access key name param, replica link delete-by-ID, and volatile attr issues are fixed
**Depends on**: Phase 20
**Requirements**: FIX-01, FIX-02, FIX-03
**Success Criteria** (what must be TRUE):
  1. Access key PATCH sends name via query param (not body); existing resources unaffected
  2. Replica link delete uses ID (not name); no 404 on destroy
  3. `UseStateForUnknown` removed from volatile fields; plan shows accurate diff on lag/backlog changes
**Plans:** 1/1 plans complete

Plans:
- [x] 21-01-PLAN.md — Access key name param fix + replica link delete-by-ID + volatile attr cleanup

### Phase 22: Test Coverage
**Goal**: Coverage reaches 68.4% with OAuth2, pagination, and HCL acceptance tests added
**Depends on**: Phase 21
**Requirements**: COV-03, COV-04, COV-05
**Success Criteria** (what must be TRUE):
  1. 5 data source tests, OAuth2 test, pagination test, and 3 HCL acceptance tests added
  2. `go test -cover` reports ≥ 68% coverage
  3. `make lint` clean with expanded golangci-lint rules
**Plans:** 1/1 plans complete

Plans:
- [x] 22-01-PLAN.md — Data source tests + OAuth2 + pagination + HCL acceptance tests

</details>

<details>
<summary>v2.1 Bucket Advanced Features (Phases 23-27) - SHIPPED 2026-03-30</summary>

### Phase 23: Bucket Inline Attributes
**Goal**: Operators can manage bucket-level eradication, object lock, public access, and public status configuration inline in the bucket resource
**Depends on**: Phase 22
**Requirements**: BKT-07, BKT-08, BKT-09, BKT-10
**Success Criteria** (what must be TRUE):
  1. `eradication_config`, `object_lock_config`, `public_access_config` blocks apply and converge with 0-diff second plan
  2. `public_status` is computed-only; drift is detected and logged
  3. `make test` passes with 0 failures
**Plans:** 2/2 plans complete

Plans:
- [x] 23-01-PLAN.md — Bucket inline blocks: eradication_config, object_lock_config, public_access_config
- [x] 23-02-PLAN.md — public_status computed field + bucket schema v0→v1 upgrader

### Phase 24: Lifecycle Rules
**Goal**: Operators can manage bucket lifecycle rules (expiration, transition, multipart cleanup) through Terraform
**Depends on**: Phase 23
**Requirements**: LCR-01, LCR-02, LCR-03, LCR-04
**Success Criteria** (what must be TRUE):
  1. Operator can create a lifecycle rule with prefix, retention days, and multipart cleanup; rule appears in FlashBlade after apply
  2. `terraform import` uses composite ID `bucket_name:rule_index`; subsequent plan shows 0 diff
  3. Deleting all rules in a bucket applies cleanly (empty list PATCH)
**Plans:** 2/2 plans complete

Plans:
- [x] 24-01-PLAN.md — Lifecycle rule: models, client CRUD, mock handler
- [x] 24-02-PLAN.md — Lifecycle rule resource + data source + import

### Phase 25: Bucket Access Policies
**Goal**: Operators can manage per-bucket IAM-style access policies and rules through Terraform
**Depends on**: Phase 24
**Requirements**: BAP-01, BAP-02, BAP-03, BAP-04, BAP-05
**Success Criteria** (what must be TRUE):
  1. Operator can create a bucket access policy, add a rule with principal/action/resource, and attach it to a bucket
  2. Policy rule import uses composite ID; subsequent plan shows 0 diff
  3. `terraform destroy` on policy cascades cleanly
**Plans:** 2/2 plans complete

Plans:
- [x] 25-01-PLAN.md — Bucket access policy + rule: models, client, mock
- [x] 25-02-PLAN.md — Bucket access policy resource + rule resource + data source

### Phase 26: Audit Filters
**Goal**: Operators can manage S3 operation audit filters (event capture per prefix) through Terraform
**Depends on**: Phase 25
**Requirements**: AUD-01, AUD-02, AUD-03
**Success Criteria** (what must be TRUE):
  1. Operator can create an audit filter specifying S3 operations and prefix patterns; filter appears after apply
  2. `terraform import` works by name; subsequent plan shows 0 diff
  3. Drift on operation list is detected and logged
**Plans:** 2/2 plans complete

Plans:
- [x] 26-01-PLAN.md — Audit filter: models, client, mock handler
- [x] 26-02-PLAN.md — Audit filter resource + data source + import

### Phase 27: QoS Policies
**Goal**: Operators can manage QoS policies (bandwidth and IOPS limits) and their bucket memberships through Terraform
**Depends on**: Phase 26
**Requirements**: QOS-01, QOS-02, QOS-03, QOS-04
**Success Criteria** (what must be TRUE):
  1. Operator can create a QoS policy with limits, create a member resource attaching a bucket, and the limits are enforced
  2. Removing a member resource detaches the bucket from the policy cleanly
  3. All bucket advanced feature tests pass; `make docs` regenerated
**Plans:** 2/2 plans complete

Plans:
- [x] 27-01-PLAN.md — QoS policy: models, client, mock handler, resource, data source
- [x] 27-02-PLAN.md — QoS policy member resource + documentation + test audit

</details>

<details>
<summary>v2.1.1 Network Interfaces (VIPs) (Phases 28-31) - SHIPPED 2026-03-31</summary>

### Phase 28: LAG Data Source
**Goal**: Operators can read existing link aggregation group configurations through a Terraform data source
**Depends on**: Phase 27
**Requirements**: LAG-01, LAG-02
**Success Criteria** (what must be TRUE):
  1. `flashblade_lag` data source reads an existing LAG by name with all attributes populated
  2. Not-found produces a clear error (not panic); data source has no write operations
**Plans:** 1/1 plans complete

Plans:
- [x] 28-01-PLAN.md — LAG data source: models, client read, mock handler, data source

### Phase 29: Subnet Resource
**Goal**: Operators can manage FlashBlade subnets through Terraform with full CRUD, import, and drift detection
**Depends on**: Phase 28
**Requirements**: SUBNET-01, SUBNET-02, SUBNET-03, SUBNET-04
**Success Criteria** (what must be TRUE):
  1. Operator can create a subnet with gateway, prefix length, and VLAN; subnet appears after apply
  2. `terraform import flashblade_subnet.x name` populates all fields; subsequent plan shows 0 diff
  3. CIDR validator rejects malformed input with a clear error message
**Plans:** 2/2 plans complete

Plans:
- [x] 29-01-PLAN.md — Subnet: models, client CRUD, mock handler
- [x] 29-02-PLAN.md — Subnet resource + data source + CIDR validator + import

### Phase 30: Network Interface Resource
**Goal**: Operators can manage VIPs (virtual IP addresses) attached to servers and subnets through Terraform
**Depends on**: Phase 29
**Requirements**: NI-01, NI-02, NI-03, NI-04, NI-05
**Success Criteria** (what must be TRUE):
  1. Operator can create a VIP specifying address, subnet, services, and server; VIP appears after apply
  2. Service list changes trigger PATCH with accurate diff; plan shows updated services
  3. `terraform import` works by name; subsequent plan shows 0 diff
**Plans:** 2/2 plans complete

Plans:
- [x] 30-01-PLAN.md — Network interface: models, client CRUD, mock handler
- [x] 30-02-PLAN.md — Network interface resource + validators + data source + import

### Phase 31: Server Enrichment & Workflow
**Goal**: Server resource exposes computed network_interfaces list, and a complete networking workflow example is documented
**Depends on**: Phase 30
**Requirements**: SRV-06, SRV-07, WF-02
**Success Criteria** (what must be TRUE):
  1. `flashblade_server` resource shows `network_interfaces` as computed list after apply
  2. Schema v0→v1 migration runs without errors on existing state
  3. Networking workflow example (LAG → subnet → VIP → server) applies cleanly end-to-end
**Plans:** 2/2 plans complete

Plans:
- [x] 31-01-PLAN.md — Server schema v0→v1: add network_interfaces computed list + state upgrader
- [x] 31-02-PLAN.md — Networking workflow example + documentation audit

</details>

<details>
<summary>v2.1.3 Code Review Fixes & S3 Users (Phases 32-35) - SHIPPED 2026-04-02</summary>

### Phase 32: Code Review Bug Fixes
**Goal**: All critical issues from the full codebase review are fixed — typos, dead schema attrs, diagnostic severity, context propagation
**Depends on**: Phase 31
**Requirements**: CR-01, CR-02, CR-03, CR-04
**Success Criteria** (what must be TRUE):
  1. FreezeLockedObjects typo fixed; `make build` passes
  2. Dead filesystem schema attributes removed; schema version bumped + upgrader added
  3. Diagnostic severity levels correct (Error vs Warning); OAuth2 context propagated
**Plans:** 2/2 plans complete

Plans:
- [x] 32-01-PLAN.md — FreezeLockedObjects fix + dead schema attrs removal + schema upgrader
- [x] 32-02-PLAN.md — Diagnostic severity fix + OAuth2 context propagation

### Phase 33: Lint Expansion
**Goal**: golangci-lint is expanded with gosec, bodyclose, noctx, exhaustive; RetryBaseDelay removed; all lint issues resolved
**Depends on**: Phase 32
**Requirements**: LINT-01, LINT-02, LINT-03
**Success Criteria** (what must be TRUE):
  1. `.golangci.yml` enables gosec, bodyclose, noctx, exhaustive linters
  2. `RetryBaseDelay` and similar heuristic duration fields removed from all structs
  3. `make lint` exits clean (0 issues) with expanded rule set
**Plans:** 1/1 plans complete

Plans:
- [x] 33-01-PLAN.md — .golangci.yml expansion + RetryBaseDelay removal + lint resolution

### Phase 34: S3 User Management
**Goal**: Operators can create and delete named S3 users and control their admin access level through Terraform
**Depends on**: Phase 33
**Requirements**: OSU-01, OSU-02, OSU-03, OSU-04
**Success Criteria** (what must be TRUE):
  1. Operator can create an object store user with `full_access=true`; user appears in FlashBlade after apply
  2. `full_access` is sent as query param (write-only); no drift detected on Read
  3. `terraform import` works by name; subsequent plan shows 0 diff
**Plans:** 2/2 plans complete

Plans:
- [x] 34-01-PLAN.md — Object store user: models, client CRUD, mock handler (full_access as query param)
- [x] 34-02-PLAN.md — Object store user resource + data source + import

### Phase 35: User-Policy Associations & Account Fix
**Goal**: Operators can associate access policies to S3 users, and the quota_limit PATCH guard is fixed for object store accounts
**Depends on**: Phase 34
**Requirements**: OSU-05, OAK-FIX-01
**Success Criteria** (what must be TRUE):
  1. Operator can create a `flashblade_object_store_user_policy` member resource linking a user to an access policy
  2. Removing the member resource detaches cleanly; policy is not deleted
  3. `quota_limit` is only sent in PATCH when it is known (not Unknown); no spurious 400 errors
**Plans:** 2/2 plans complete

Plans:
- [x] 35-01-PLAN.md — Object store user-policy member resource
- [x] 35-02-PLAN.md — quota_limit IsUnknown guard + test audit (test baseline update)

</details>

<details>
<summary>v2.2 S3 Target Replication & TLS (Phases 36-42) - SHIPPED 2026-04-14</summary>

### Phase 36: Target Resource
**Goal**: Operators can manage external S3 target endpoints (for target replication) through Terraform with full CRUD, import, and drift detection
**Depends on**: Phase 35
**Requirements**: TGT-01, TGT-02, TGT-03, TGT-04, TGT-05
**Success Criteria** (what must be TRUE):
  1. Operator can create a target specifying address, CA cert group, and type; target appears after apply
  2. `terraform import flashblade_target.x name` populates all fields; subsequent plan shows 0 diff
  3. CA cert group reference cleared by omitting the attribute; next plan shows 0 diff
**Plans:** 2/2 plans complete

Plans:
- [x] 36-01-PLAN.md — Target: models, client CRUD, mock handler
- [x] 36-02-PLAN.md — Target resource + data source + import

### Phase 37: Remote Credentials & Replica Link Enhancement
**Goal**: Remote credentials support CA cert groups and replica links support target references for target replication
**Depends on**: Phase 36
**Requirements**: RC-06, RL-06, RL-07
**Success Criteria** (what must be TRUE):
  1. Remote credentials accept `ca_certificate_group` NamedReference; existing credentials unaffected
  2. Replica link accepts `target` NamedReference; existing links unaffected
  3. Schema version bumped + upgrader added for both resources; `make test` passes
**Plans:** 2/2 plans complete

Plans:
- [x] 37-01-PLAN.md — Remote credentials schema v1→v2: ca_certificate_group NamedReference
- [x] 37-02-PLAN.md — Replica link schema v0→v1: target NamedReference

### Phase 38: Certificate Resource
**Goal**: Operators can import PEM certificates (with optional passphrase/private key) into FlashBlade through Terraform
**Depends on**: Phase 37
**Requirements**: CERT-01, CERT-02, CERT-03, CERT-04
**Success Criteria** (what must be TRUE):
  1. Operator can create a certificate by providing PEM content; certificate appears after apply
  2. `passphrase` and `private_key` are write-only — never returned by API, not in plan diff
  3. `terraform import flashblade_certificate.x name` works; subsequent plan shows 0 diff
**Plans:** 2/2 plans complete

Plans:
- [x] 38-01-PLAN.md — Certificate: models, client CRUD, mock handler
- [x] 38-02-PLAN.md — Certificate resource + data source + write-only fields + import

### Phase 39: TLS Policy Resource
**Goal**: Operators can manage TLS policies (cipher suites, certificate assignments) through Terraform
**Depends on**: Phase 38
**Requirements**: TLS-01, TLS-02, TLS-03, TLS-04, TLS-05
**Success Criteria** (what must be TRUE):
  1. Operator can create a TLS policy with cipher list; policy applies after apply
  2. TLS policy member resource attaches a certificate to a policy; member import uses composite ID
  3. `terraform destroy` on policy + members cascades cleanly
**Plans:** 2/2 plans complete

Plans:
- [x] 39-01-PLAN.md — TLS policy: models, client, mock, resource, data source
- [x] 39-02-PLAN.md — TLS policy member resource + import

### Phase 40: Certificate Group Resource
**Goal**: Operators can manage certificate groups (CA bundles) and their members through Terraform
**Depends on**: Phase 39
**Requirements**: CG-01, CG-02, CG-03, CG-04
**Success Criteria** (what must be TRUE):
  1. Operator can create a certificate group, add member certificates, and attach the group to a target
  2. Certificate group member import uses composite ID; subsequent plan shows 0 diff
  3. `terraform destroy` removes group + members cleanly
**Plans:** 2/2 plans complete

Plans:
- [x] 40-01-PLAN.md — Certificate group: models, client, mock, resource, data source
- [x] 40-02-PLAN.md — Certificate group member resource + import

### Phase 41: Array Connection Resource
**Goal**: Operators can manage array connections (inter-array links) through Terraform including sensitive connection keys
**Depends on**: Phase 40
**Requirements**: AC-03, AC-04, AC-05, AC-06, AC-07
**Success Criteria** (what must be TRUE):
  1. Operator can create an array connection specifying management address, connection key, and throttle settings
  2. `connection_key` is sensitive write-only; not in plan diff, not in state after creation
  3. Array connection key ephemeral resource generates a one-time key; subsequent plan shows 0 diff
**Plans:** 2/2 plans complete

Plans:
- [x] 41-01-PLAN.md — Array connection: models, client CRUD, mock handler + ephemeral key resource
- [x] 41-02-PLAN.md — Array connection resource + data source + import

### Phase 42: Array DNS Transform & v2.2 Close
**Goal**: Array DNS singleton is transformed to a named resource, and v2.2 is closed with documentation and import guides for all new resources
**Depends on**: Phase 41
**Requirements**: DNS-03, DNS-04, DOC-v22-01, DOC-v22-02
**Success Criteria** (what must be TRUE):
  1. `flashblade_array_dns` resource accepts a `name` attribute and applies cleanly (backward-compatible)
  2. All v2.2 resources have import.sh files; `make docs` regenerated
  3. `make test` passes all tests; `make lint` clean
**Plans:** 2/2 plans complete

Plans:
- [x] 42-01-PLAN.md — Array DNS singleton→named transform + schema upgrader
- [x] 42-02-PLAN.md — v2.2 documentation audit + import guides + make docs

</details>

<details>
<summary>tools-v1.0 API Tooling Pipeline (Phases 43-48) - SHIPPED 2026-04-14</summary>

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
**Plans**: 2 plans
**UI hint**: no

Plans:
- [x] 44-01-PLAN.md — parse_swagger.py converter (swagger.json to AI-optimized markdown)
- [x] 44-02-PLAN.md — swagger-to-reference SKILL.md with version-prompt workflow

### Phase 45: API Browsing Tools
**Goal**: Claude can search, inspect, and compare API endpoints and schemas interactively from a generated reference file
**Depends on**: Phase 44 (reference file must exist to browse)
**Requirements**: BRWS-01, BRWS-02, BRWS-03, BRWS-04
**Success Criteria** (what must be TRUE):
  1. `python3 browse_api.py --tag buckets` lists all endpoints tagged `buckets` with their HTTP methods and summary lines
  2. `python3 browse_api.py --schema BucketPost` displays all fields with types, readOnly annotations, and required flags
  3. `python3 browse_api.py --compare BucketPost BucketPatch` shows a side-by-side diff table highlighting fields present in one but not the other, and type mismatches
  4. `python3 browse_api.py --stats` outputs path count, schema count, and method distribution (GET/POST/PATCH/DELETE counts)
**Plans**: 1 plan

Plans:
- [x] 45-01-PLAN.md — browse_api.py CLI tool (markdown parser + subcommands: --tag, --schema, --compare, --stats, --method, --search)

### Phase 46: api-diff Skill
**Goal**: Claude can produce a structured diff between two swagger versions, annotate discrepancies, and generate a migration plan cross-referenced with ROADMAP.md
**Depends on**: Phase 43 (shared library for path normalization)
**Requirements**: DIFF-01, DIFF-02, DIFF-03, DIFF-04, INTG-02 (partial)
**Success Criteria** (what must be TRUE):
  1. Running the skill on swagger-2.22.json and swagger-2.23.json produces a diff listing new endpoints, removed endpoints, and modified schemas — no duplicates caused by version prefix differences
  2. Each diff item is annotated as `real_change`, `swagger_artifact`, or `needs_verification` based on `known_discrepancies.md` lookup
  3. The migration plan output cross-references ROADMAP.md entries — new endpoints that match a "Not Implemented" roadmap entry are flagged as candidates
  4. api-diff SKILL.md exists with valid YAML frontmatter and structured activation, steps, and output sections
**Plans**: 2 plans

### Phase 47: api-upgrade Skill
**Goal**: Claude can mechanically update API version references across the codebase and orchestrate the provider upgrade sequence with explicit review gates
**Depends on**: Phase 46 (diff output informs upgrade scope)
**Requirements**: UPGR-01, UPGR-02, UPGR-03, INTG-02 (partial)
**Success Criteria** (what must be TRUE):
  1. `python3 upgrade_version.py --from 2.22 --to 2.23 --dry-run` lists every file and line that would change (APIVersion const, mock server version strings, mock handler paths) without modifying any file
  2. `python3 upgrade_version.py --from 2.22 --to 2.23 --apply` applies all changes and `make build` passes with no compilation errors
  3. api-upgrade SKILL.md exists with 5 named phases (infra, schemas, new resources, deprecations, docs), each with explicit review gate instructions and acceptance criteria
**Plans**: 2 plans
Plans:
- [x] 47-01-PLAN.md — upgrade_version.py script with dry-run and apply modes
- [x] 47-02-PLAN.md — SKILL.md 5-phase workflow and upgrade_checklist.md

### Phase 48: Integration & Validation
**Goal**: All three skills are documented in CLAUDE.md, the `api_references/` convention is established, and the full pipeline is validated end-to-end on real swagger files
**Depends on**: Phase 47 (all skills must exist before integration)
**Requirements**: INTG-01, INTG-02 (finalize)
**Success Criteria** (what must be TRUE):
  1. CLAUDE.md references the three skills (swagger-to-reference, api-diff, api-upgrade) with one-line descriptions and the `api_references/` output convention
  2. Running the full pipeline on swagger-2.22.json and swagger-2.23.json end-to-end (convert → browse → diff → migration plan) completes without errors and produces readable artifacts
  3. All three SKILL.md files pass YAML frontmatter validation and follow the skill-creator format consistently
**Plans**: 2 plans

</details>

---

<details>
<summary>✅ v2.22.1 Directory Service – Array Management (Phase 49) — SHIPPED 2026-04-17</summary>

Full details archived at [milestones/v2.22.1-ROADMAP.md](milestones/v2.22.1-ROADMAP.md).

- [x] Phase 49: Directory Service Management (5/5 plans) — completed 2026-04-17
</details>

---

<details>
<summary>✅ v2.22.2 Directory Service Roles & Role Mappings (Phases 50, 50.1) — SHIPPED 2026-04-17</summary>

Full details archived at [milestones/v2.22.2-ROADMAP.md](milestones/v2.22.2-ROADMAP.md).

- [x] Phase 50: Directory Service Roles & Role Mappings (5/5 plans) — completed 2026-04-17
- [x] Phase 50.1: Fix directory_service_role POST missing ?names= query param (3/3 plans, INSERTED) — completed 2026-04-17
</details>

---

<details>
<summary>✅ v2.22.3 Convention Compliance (Phases 51-53) — SHIPPED 2026-04-20</summary>

Full details archived at [milestones/v2.22.3-ROADMAP.md](milestones/v2.22.3-ROADMAP.md).

- [x] Phase 51: Critical Pointer & Schema Fixes (completed 2026-04-20)
- [x] Phase 52: Important Conformance (completed 2026-04-20)
- [x] Phase 53: Cosmetic Hygiene (completed 2026-04-20)
</details>

---

## 🚧 pulumi-2.22.3 Pulumi Bridge Alpha (Phases 54-58) — IN PROGRESS

- [ ] Phase 54: Bridge Bootstrap + POC (3 Resources)
- [ ] Phase 55: Full Mapping — 28 Resources + 21 Data Sources
- [ ] Phase 56: SDK Generation — Python + Go
- [ ] Phase 57: CI Pipeline
- [ ] Phase 58: Release Pipeline + Docs

## Phase Details

### Phase 54: Bridge Bootstrap + POC (3 Resources)
**Goal**: The Pulumi bridge scaffold compiles and the full chain — `pf.ShimProvider` → `make tfgen` → schema emission → `resources_test.go` green — is validated on 3 representative resources (target, remote_credentials, bucket)
**Depends on**: Phase 53 (v2.22.3 complete, 779 tests baseline)
**Requirements**: BRIDGE-01, BRIDGE-02, BRIDGE-03, BRIDGE-04, BRIDGE-05, COMPOSITE-01, SECRETS-01, SECRETS-02 (partial: 3 POC resources), SOFTDELETE-01, MAPPING-02, MAPPING-03, MAPPING-05, TEST-01
**Success Criteria** (what must be TRUE):
  1. `cd pulumi/provider && go build ./...` and `cd pulumi/sdk/go && go build ./...` exit 0 — both modules compile with pinned bridge v3.127.0 + replace directives
  2. `make tfgen` runs to completion and emits `schema.json`, `schema-embed.json`, `bridge-metadata.json` in `pulumi/provider/cmd/pulumi-resource-flashblade/`; no `MISSING` token warnings for target, remote_credentials, or bucket
  3. `resources_test.go` assertions pass: bucket `DeleteTimeout >= 25*time.Minute`, no `timeouts` input in any resource schema, `api_token` is secret in provider config
  4. `pulumi import flashblade:index:ObjectStoreAccessPolicyRule my-rule 'mypolicy/myrulename'` round-trip succeeds (COMPOSITE-01 validation)
  5. `pulumi stack export` for a remote_credentials resource shows `secret_access_key` value is secret (not plaintext)
**Plans**: TBD
**UI hint**: no

### Phase 55: Full Mapping — 28 Resources + 21 Data Sources
**Goal**: All 49 TF resources and data sources have valid Pulumi tokens, all 4 composite-ID overrides are in place, full secrets coverage is applied, and state-upgrader smoke tests pass for the 3 affected resources
**Depends on**: Phase 54 (bridge compiles, POC pattern validated)
**Requirements**: MAPPING-01, MAPPING-04, COMPOSITE-02, COMPOSITE-03, COMPOSITE-04, SECRETS-02 (complete), SECRETS-03, SOFTDELETE-02, SOFTDELETE-03, UPGRADE-01, UPGRADE-02, UPGRADE-03
**Success Criteria** (what must be TRUE):
  1. `make tfgen` reports zero `MISSING` tokens across all 28 resources + 21 data sources; `MustComputeTokens` + `KnownModules` covers ~90% automatically
  2. `resources_test.go` asserts `len(Resources) == 28` and `len(DataSources) == 21`; every field with `Sensitive: true` in the TF schema maps to a Pulumi Secret (SECRETS-03 auto-mapping test)
  3. `pulumi import` round-trip succeeds for all 4 composite-ID resources using `/`-separated IDs, including a test with a colon-containing policy name (`pure:policy/array_admin`)
  4. `pulumi refresh` smoke tests pass for `flashblade_server` (v0→v1→v2), `flashblade_directory_service_role` (v0→v1), and `flashblade_remote_credentials` (v0→v1) using pre-captured v0/v1 state snapshots
  5. `schema.json`, `schema-embed.json`, `bridge-metadata.json` are committed; `git diff --exit-code` on these 3 files exits 0 after `make tfgen`
**Plans**: TBD
**UI hint**: no

### Phase 56: SDK Generation — Python + Go
**Goal**: A working Python wheel and a compilable Go SDK module are generated from the committed schema, with CI artifact flow wired
**Depends on**: Phase 55 (all tokens valid, schema committed)
**Requirements**: SDK-01, SDK-02, SDK-03, SDK-04
**Success Criteria** (what must be TRUE):
  1. `make generate_python` produces `pulumi/sdk/python/` with `pulumi_flashblade` import path; `python -m build` exits 0 and emits a `.whl` file
  2. `make generate_go` produces `pulumi/sdk/go/` under `github.com/numberly/opentofu-provider-flashblade/pulumi/sdk/go`; `cd pulumi/sdk/go && go build ./...` exits 0
  3. `schema.json` and `schema-embed.json` are the committed files — SDK gen does not regenerate them (CI diff gate confirms 0 changes)
  4. No `generate_nodejs`, `generate_dotnet`, or `generate_java` Makefile targets exist (TypeScript/C#/Java explicitly out of scope)
**Plans**: TBD
**UI hint**: no

### Phase 57: CI Pipeline
**Goal**: A `pulumi-ci.yml` workflow runs automatically on PRs touching `./pulumi/**`, enforces the schema drift gate, and does not touch any existing TF provider workflows
**Depends on**: Phase 56 (SDK gen targets exist and work)
**Requirements**: CI-01, CI-02, CI-03
**Success Criteria** (what must be TRUE):
  1. A PR modifying any file under `./pulumi/` triggers `pulumi-ci.yml`; the workflow runs `make tfgen`, uploads schema-embed.json as artifact, then builds provider + SDKs in parallel jobs
  2. Introducing a manual edit to `schema.json` in a PR causes `pulumi-ci.yml` to fail with a non-zero exit on the `git diff --exit-code` step (schema drift gate enforced)
  3. Merging a TF-provider-only PR (touching `./internal/**` only) does NOT trigger `pulumi-ci.yml`; the existing TF CI workflow still passes unchanged
**Plans**: TBD
**UI hint**: no

### Phase 58: Release Pipeline + Docs
**Goal**: The `pulumi-2.22.3` tag produces cosign-signed plugin binaries and a Python wheel on GitHub Releases; consumer onboarding is documented; 6 ProgramTest examples exist
**Depends on**: Phase 57 (CI pipeline green)
**Requirements**: RELEASE-01, RELEASE-02, RELEASE-03, TEST-02, TEST-03, DOCS-01, DOCS-02, DOCS-03, DOCS-04
**Success Criteria** (what must be TRUE):
  1. Pushing tag `pulumi-2.22.3` triggers `pulumi-release.yml`; GitHub Release `pulumi-2.22.3` contains 5 signed platform archives (`pulumi-resource-flashblade-v2.22.3-{os}-{arch}.tar.gz`) and a `.whl` file; cosign signatures are attached
  2. `pulumi plugin install resource flashblade v2.22.3 --server github://api.github.com/numberly` succeeds from a consumer machine with no local build; `pulumi up` on a Python consumer program referencing `pulumi_flashblade` imports and runs
  3. `git tag sdk/go/v2.22.3` exists and `go get github.com/numberly/opentofu-provider-flashblade/pulumi/sdk/go@v2.22.3` resolves (with `GOPRIVATE=github.com/numberly/*`)
  4. `./pulumi/examples/` contains 6 working directories (`target-py/`, `target-go/`, `remote_credentials-py/`, `remote_credentials-go/`, `bucket-py/`, `bucket-go/`), each with a valid `Pulumi.yaml` and main program
  5. `./pulumi/README.md` documents GOPRIVATE setup, plugin install URL, wheel install URL, `customTimeouts` for soft-delete, and composite ID import syntax
**Plans**: TBD
**UI hint**: no

## Progress Table

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 54. Bridge Bootstrap + POC (3 Resources) | 2/5 | In Progress|  |
| 55. Full Mapping — 28 Resources + 21 Data Sources | 0/? | Not started | - |
| 56. SDK Generation — Python + Go | 0/? | Not started | - |
| 57. CI Pipeline | 0/? | Not started | - |
| 58. Release Pipeline + Docs | 0/? | Not started | - |

---

_Last updated: 2026-04-21 — milestone pulumi-2.22.3 roadmap created (Phases 54-58, 39 requirements)_
