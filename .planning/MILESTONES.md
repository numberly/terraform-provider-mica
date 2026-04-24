# Milestones: Terraform Provider FlashBlade

## Completed Milestones

### v1.0 — Core Provider (completed 2026-03-28)

**Goal:** Full Terraform provider for FlashBlade with CRUD, import, and drift detection for all storage resources.

**Delivered:**

- Provider scaffold + HTTP client (auth, TLS, retry, version negotiation)
- 22 resources: file systems, object store (accounts, buckets, access keys), 6 policy families (NFS export, SMB share, snapshot, OAP, NAP, quota), array admin (DNS, NTP, SMTP), file system exports, account exports, server data source
- 227 unit tests, CI pipeline, documentation auto-generated
- 14 bugs fixed via live FlashBlade acceptance testing

**Phases:** 1–5 (20 plans)
**Last phase number:** 5

### v1.1 — Servers & Exports (completed 2026-03-28)

**Goal:** Complete server lifecycle management and export infrastructure.

**Delivered:**

- 10 new resources: server, S3 export policy + rule, virtual host, SMB client policy + rule, syslog server, file system export, account export (TDD consolidated)
- 268 unit tests (41 new), all passing
- 26 resources tested apply/destroy against live FlashBlade
- Complete S3 tenant onboarding workflow (server → account → export → policies → bucket)

**Phases:** 6–8 (8 plans)
**Last phase number:** 8

### v1.2 — Code Quality & Robustness (completed 2026-03-29)

**Goal:** Fix latent bugs, harden test coverage, add validators, clean up architecture.

**Delivered:**

- 4 bug fixes (account export Delete, filesystem writable drift, IsNotFound scoping, omitempty)
- models.go split into 5 domain files, compositeID + stringOrNull shared helpers
- 9 idempotence tests, 4 mock handlers hardened with query param validation
- 2 custom validators (Alphanumeric, HostnameNoDot) + enum OneOf validators
- 329 unit tests (61 new), all passing

**Phases:** 9–11 (7 plans)
**Last phase number:** 11

### v1.3 — Release Readiness (completed 2026-03-29)

**Goal:** Best practices from major providers: state migration, import docs, transport hardening.

**Delivered:**

- SchemaVersion 0 + UpgradeState on all 28 resources
- int64/float64UseStateForUnknown helpers consolidated
- ±20% jitter in exponential backoff
- 27 import.sh files + tfplugindocs regenerated
- 340 unit tests, all passing

**Phases:** 12–13 (4 plans)
**Last phase number:** 13

### v2.0 — Cross-Array Bucket Replication (completed 2026-03-29)

**Goal:** Enable operators to set up bidirectional S3 bucket replication between two FlashBlade arrays through Terraform.

**Delivered:**

- Access key enhancement: optional secret_access_key input for cross-array credential sharing
- Object store remote credentials resource (CRUD + import)
- Bucket replica link resource (bidirectional replication, CRUD + import)
- Array connection data source (read existing inter-array connection)
- Complete dual-provider replication workflow example
- Unit tests for all new resources + acceptance test HCL

**Phases:** 14–17 (8 plans)
**Last phase number:** 17

### v2.0.1 — Quality & Hardening (completed 2026-03-30)

**Goal:** Harden the codebase post-v2.0 with security fixes, code quality improvements, dead code removal, and test coverage.

**Delivered:**

- OAuth2 error sanitization, context propagation through all auth paths, 30s HTTP timeout
- errors.As() migration, ParseAPIError hardening, fresh-GET bucket delete guard
- 8 shared helpers (spaceAttrTypes, nullTimeoutsValue, getOneByName[T], pollUntilGone[T], etc.)
- Dead code removal (~405 lines), math/rand/v2 modernization
- 16 new tests (5 data source, OAuth2, pagination, 3 HCL acceptance), coverage 68.4%
- Access key name param fix, replica link delete-by-ID fix, volatile attr cleanup

**Phases:** 18–22 (7 plans)
**Last phase number:** 22

### v2.1 — Bucket Advanced Features (completed 2026-03-30)

**Goal:** Complete bucket sub-resource ecosystem — lifecycle rules, access policies, audit filters, QoS policies, and inline config blocks.

**Delivered:**

- Bucket inline attributes: eradication_config, object_lock_config, public_access_config, public_status
- Lifecycle rule resource + data source (prefix, retention, multipart cleanup)
- Bucket access policy + rule resources + data source (IAM-style per-bucket authorization)
- Bucket audit filter resource + data source (S3 operation auditing with prefix filters)
- QoS policy resource + member resource + data source (bandwidth/IOPS limits)
- Testing and documentation for all new resources

**Phases:** 23–27 (10 plans)
**Last phase number:** 27

### v2.1.1 — Network Interfaces (VIPs) (completed 2026-03-31)

**Goal:** Enable operators to manage the full networking stack — subnets, VIPs, and server-to-VIP relationships — through Terraform.

**Delivered:**

- LAG data source (read existing link aggregation groups)
- Subnet resource + data source (CRUD, import, drift detection)
- Network interface resource + data source (VIPs with service/server semantics, validators)
- Server enrichment: computed network_interfaces list, schema v0→v1 migration
- Networking workflow example (LAG → subnet → VIP → server)
- Documentation and import guides for all new resources

**Phases:** 28–31 (6 plans)
**Last phase number:** 31

### v2.1.3 — Code Review Fixes & S3 Users (completed 2026-04-02)

**Goal:** Fix all issues identified by the full codebase code review, and add S3 user management with per-user policy associations.

**Delivered:**

- FreezeLockedObjects typo fix, dead filesystem schema attrs removed, diagnostic severity fix
- OAuth2 context propagation, RetryBaseDelay removal, golangci-lint expansion (gosec/bodyclose/noctx/exhaustive)
- Object store user resource + data source (create/delete named S3 users, full_access support)
- Object store user-policy member resource (associate access policies to users)
- full_access fix: sent as query param (write-only, not returned by API)
- quota_limit PATCH guard on object store account (IsUnknown check)

**Phases:** 32–35 (7 plans)
**Last phase number:** 35

### v2.2 — S3 Target Replication & Security Infrastructure (completed 2026-04-14)

**Goal:** Enable S3 target replication to external endpoints, certificate/TLS management, and array connections through Terraform.

**Delivered:**

- Target resource + data source (external S3 endpoints, CA cert groups)
- Remote credentials enhancement + replica link with target support
- Certificate resource + data source (PEM import, write-only passphrase/private_key)
- TLS policy resource + data source + member resource
- Certificate group resource + data source + member resource
- Array connection resource + data source (connection_key sensitive, throttle, replication_addresses)
- Array connection key ephemeral resource
- Array DNS singleton → named resource transform
- Documentation, import guides, workflow examples

**Phases:** 36–42 (11+ plans)
**Last phase number:** 42

### tools-v1.0 — API Tooling Pipeline (completed 2026-04-14)

**Goal:** Automate swagger-to-reference conversion, API version diffing, and provider upgrade orchestration through Claude Code skills with Python tooling.

**Delivered:**

- Shared Python library (swagger_utils.py) — allOf resolver, path normalizer, schema flattener, 15 tests
- swagger-to-reference skill — parse_swagger.py (226 paths, 538 ops → 1734 lines), browse_api.py (6 subcommands)
- api-diff skill — diff_swagger.py (structured endpoint/schema diff), generate_migration_plan.py (cross-ref ROADMAP.md), known_discrepancies.md
- api-upgrade skill — upgrade_version.py (38 files, dry-run/apply), 5-phase orchestration SKILL.md with review gates
- CLAUDE.md updated with API tools, api_references/ convention
- Full E2E pipeline validated on swagger-2.22.json and swagger-2.23.json

**Phases:** 43–48 (9 plans)
**Last phase number:** 48

### v2.22.1 — Directory Service – Array Management (completed 2026-04-17)

**Goal:** Manage the FlashBlade array management LDAP directory service through Terraform — URIs, bind credentials, CA certificate group, and user-attribute configuration.

**Delivered:**

- `flashblade_directory_service_management` resource (singleton; drift detection on 10 fields; Delete = full-reset PATCH; import by literal `management`; `bind_password` Sensitive write-only)
- `flashblade_directory_service_management` data source (computed-only, nested `ca_certificate` / `ca_certificate_group` objects, no `bind_password`)
- Reusable `LDAPURIValidator()` list validator (rejects non-`ldap://`/`ldaps://` URIs)
- Mock handler with GET+PATCH-only contract and `**NamedReference` clear/set support
- HCL examples (`resource.tf`, `import.sh`, `data-source.tf`), auto-generated docs, ROADMAP + CONVENTIONS updated
- Test baseline 787 → 798 (+11), 0 lint issues

**Phases:** 49 (5 plans)
**Last phase number:** 49

**Archives:** [milestones/v2.22.1-ROADMAP.md](milestones/v2.22.1-ROADMAP.md) · [v2.22.1-REQUIREMENTS.md](milestones/v2.22.1-REQUIREMENTS.md) · [v2.22.1-MILESTONE-AUDIT.md](milestones/v2.22.1-MILESTONE-AUDIT.md)

### v2.22.2 — Directory Service Roles & Role Mappings (completed 2026-04-17)

**Goal:** LDAP group ↔ FlashBlade role mapping through Terraform — map LDAP groups to built-in management roles and associate those roles with management access policies.

**Delivered:**

- `flashblade_directory_service_role` resource + data source (CRUD on LDAP group ↔ role mapping; drift on 4 fields; composite ID not needed; import by name)
- `flashblade_management_access_policy_directory_service_role_membership` resource (additive policy↔role association; composite ID `<role_name>/<policy_name>` with `/` separator per D-05; Update returns error; destroy severs link only)
- Schema v0→v1 upgrader on DSR (name moved from Computed to Required after 50.1 bugfix)
- Mock handlers with `?policy_names=&member_names=` query contract (real-array contract: `member_names` not `role_names`)
- HCL examples, auto-generated docs (3 new), ROADMAP + CONVENTIONS updated (baseline 818)
- Live-array UAT on par5/pa7 confirming HTTP 200 on DSR create
- Test baseline 798 → 818 (+20), 0 lint issues

**Phases:** 50, 50.1 (8 plans total — 5 + 3 decimal insertion for defect fix)
**Last phase number:** 50.1

**Known gaps:** No VALIDATION.md (Nyquist) for either phase — tracked as backlog (same pattern as v2.22.1).

**Archives:** [milestones/v2.22.2-ROADMAP.md](milestones/v2.22.2-ROADMAP.md) · [v2.22.2-REQUIREMENTS.md](milestones/v2.22.2-REQUIREMENTS.md) · [v2.22.2-MILESTONE-AUDIT.md](milestones/v2.22.2-MILESTONE-AUDIT.md)

### v2.22.3 — Convention Compliance (completed 2026-04-20)

**Goal:** Close 27 CONVENTIONS.md violations identified by code audit; clarify convention for legitimate exceptions previously undocumented.

**Delivered:**

- `doublePointerRefForPatch` helper for `**NamedReference` omit/clear/set semantics
- `SubnetPost.VLAN *int64` fix (VLAN=0 now sendable); `SubnetPatch.LinkAggregationGroup **NamedReference` (detach support)
- `**NamedReference` Patch support on FileSystemExport + ObjectStoreAccountExport (server/policy clear)
- `QosPolicyPost` name via `?names=` query param; unlimited quotas (`*int64` zero sendable)
- `ResourceWithUpgradeState` assertion added to 7 resources missing it
- `ResourceWithImportState` shim on access_key with informative error; exception documented in CONVENTIONS.md
- `NfsExportPolicyRulePatch.Security *[]string` (clear semantics); NFS rule schema v0→v1
- Test function renames to `TestUnit_` prefix across 4 files (6 functions)
- PATCH slice fields → `*[]string` on 3 policy rule resources; "always send" exception formalized in CONVENTIONS.md
- Mock handler store canonicalization (`byName`/`nextID` pattern)
- Test delta: 752 → 779 (+27), 0 lint issues

**Phases:** 51–53 (11 plans)
**Last phase number:** 53

**Archives:** [milestones/v2.22.3-ROADMAP.md](milestones/v2.22.3-ROADMAP.md) · [v2.22.3-REQUIREMENTS.md](milestones/v2.22.3-REQUIREMENTS.md)

### pulumi-2.22.3 — Pulumi Bridge Alpha (completed 2026-04-24)

**Goal:** Bridge the existing TF FlashBlade provider to Pulumi via `pulumi-terraform-bridge/v3 pkg/pf/*`, with Python + Go SDKs distributed privately.

**Delivered:**

- Pulumi bridge scaffold under `./pulumi/` (3 go.mod modules, Makefile, goreleaser config)
- `resources.go` with ProviderInfo for 54 resources + 40 data sources (SingleModule tokens, MustApplyAutoAliases)
- 7 `Secret:tfbridge.True()` + 1 `Secret:tfbridge.False()` marks on sensitive fields (belt-and-braces)
- 2 ComputeID closures kept (bucket_access_policy_rule, management_access_policy_dsr_membership); 3 removed after code review (TF data.ID flows through natively)
- `make tfgen` → `schema.json` (8255 lines) + `bridge-metadata.json` committed with CI drift gate
- Python SDK (`.whl` buildable) + Go SDK (lean, no bridge dep leak) generated
- `pulumi-ci.yml` pipeline (schema drift gate, tests, lint) + `pulumi-release.yml` (cosign keyless, sdk/go tag push)
- 8 runnable examples (bucket, target, remote_credentials, s3-replication) × (Python + Go)
- Consumer docs: README, CHANGELOG alpha entry, GOPRIVATE install instructions
- Code review cleanup: 4 Critical + 8 Important + 7 Minor items fixed (14 commits)
- 23 bridge tests + 836 TF provider tests (0 regression from v2.22.3 baseline)
- 2 guardrails added: git commit-msg hook + Claude Code PreToolUse hook blocking Co-Authored-By trailers
- Dependabot: go-git/v5 bumped to 5.18.0 (GHSA-3xc5-wrhm-f963)

**Phases:** 54–58 (15 plans total)
**Last phase number:** 58

**Known gaps (tech debt):**
- `pulumi import` round-trip tests on composite-ID resources: validated statically but not tested live against array (deferred to post-alpha)
- ProgramTest coverage limited to 6 examples; full 54-resource coverage deferred to post-alpha
- TEST-02 examples delivered but live execution on array not run (deferred per VERIFICATION.md)
- RELEASE-03 smoke test (`pulumi plugin install`) deferred until first `pulumi-*` tag pushed

**Archives:** [milestones/pulumi-2.22.3-ROADMAP.md](milestones/pulumi-2.22.3-ROADMAP.md) · [pulumi-2.22.3-REQUIREMENTS.md](milestones/pulumi-2.22.3-REQUIREMENTS.md)
