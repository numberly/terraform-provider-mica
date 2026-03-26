# Feature Research

**Domain:** Terraform provider for enterprise storage (Pure Storage FlashBlade REST API v2.22)
**Researched:** 2026-03-26
**Confidence:** HIGH (HashiCorp official docs + API reference + ecosystem analysis)

---

## Feature Landscape

### Table Stakes (Users Expect These)

Features users assume exist. Missing these = product feels incomplete or unusable.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Full CRUD for every resource | Terraform's core contract — resources without full lifecycle are broken | MEDIUM | Create/Read/Update/Delete + destroy for all 12 resource families in PROJECT.md |
| Accurate drift detection (Read) | Users run `terraform plan` to detect out-of-band changes; stale Read = silent corruption | HIGH | Must capture ALL attributes in Read, including computed/backend-assigned ones. Call Read after Create/Update to sync state. |
| `terraform import` for all resources | Ops team has existing FlashBlade infra; can't manage without import | MEDIUM | Implement `ResourceWithImportState`; use `resource.ImportStatePassthroughID()` where name is the natural key. Policy rules need composite ID (`policy_name/rule_index`). |
| Data sources for every resource | Users need to reference existing infra (not Terraform-managed) in configs | MEDIUM | Parallel to each resource: `flashblade_file_system`, `flashblade_bucket`, etc. List data sources for multi-result lookups. |
| Dual authentication: API token + OAuth2 | API token = dev/local; OAuth2 client_credentials = production CI/CD | MEDIUM | API token via session login (`POST /api/login` → `x-auth-token`). OAuth2 via `POST /oauth2/1.0/token` with `token-exchange` grant type. |
| TLS support with custom CA certificates | Enterprise environments use internal CAs; hard failure without this | LOW | Configurable `ca_cert_file` or `ca_cert` in provider schema. Standard Go `tls.Config`. |
| Sensitive attribute flags on secrets | API tokens, passwords, access keys must not appear in plan/apply output | LOW | Mark `api_token`, `secret_access_key`, `password` fields as `Sensitive: true` in schema. |
| `sensitive` flag on object store access keys | Access keys stored in state must be redacted from CLI output | LOW | `flashblade_object_store_access_key` — secret_access_key is sensitive |
| Provider-level environment variable configuration | CI/CD pipelines pass credentials via env vars; hardcoding is anti-pattern | LOW | `FLASHBLADE_ENDPOINT`, `FLASHBLADE_API_TOKEN`, `FLASHBLADE_OAUTH2_*` env vars as fallback to config block |
| Correct plan modifiers on computed attributes | Without `UseStateForUnknown` on stable computed attrs, every plan shows "(known after apply)" noise | MEDIUM | All `id`, `eui`, `wwn`, `created`, `space` fields need `UseStateForUnknown`. Immutable fields (account on bucket, nfs_v3 on create) need `RequiresReplace`. |
| Resource validators | Invalid input (negative quota, invalid policy rule syntax) must fail at plan time, not at apply | MEDIUM | Custom `validator.String`/`validator.Int64` for quota values, policy effect enum, retention period ranges |
| Structured logging via `tflog` | Required for debugging provider issues in production; standard across ecosystem | LOW | Use `tflog.Info/Debug/Error` with structured fields (resource name, operation, API path). Critical for the audit logging requirement. |
| Unit tests for schema and validators | Schema regressions and validator bugs are common; catch before acceptance tests | MEDIUM | Test plan modifiers, validators, schema defaults in isolation — no API required |
| Acceptance tests against real FlashBlade | The only way to verify full CRUD behavior including API responses and state convergence | HIGH | Use `resource.Test` + `resource.TestStep` sequences (Create → Read → Update → Read → Destroy) |
| API versioning header | FlashBlade requires `/api/2.22/` prefix; future-proof version handling | LOW | Pin to `2.22` in HTTP client. Version negotiation via `GET /api/api_version` on startup. |

### Differentiators (Competitive Advantage)

Features that set the product apart from a naive provider implementation.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Drift detection with structured audit log output | Ops compliance requirement: when Read detects a diff, log exactly which attributes changed from what to what at INFO level via tflog | MEDIUM | Log field-by-field diffs in Read when prior state differs from API response. Use `tflog.Info(ctx, "drift detected", "resource", name, "attribute", field, "was", old, "now", new)`. |
| Mocked API integration tests for CI (no FlashBlade required) | Enables fast feedback in CI pipelines where real FlashBlade access is unavailable or expensive | HIGH | HTTP mock server (e.g., `net/http/httptest`) implementing FlashBlade API responses. Acceptance tests pass `TF_ACC=1` gate; mocked tests run always. |
| Full policy family coverage in v1 | Competitors (community FlashArray provider) ship partial policy support, forcing click-ops fallback | HIGH | All 6 policy types: NFS export, SMB share, snapshot, object store access, network access, quota. Each has policy + rules as separate sub-resources. |
| Composite import IDs for policy rules | Policy rules have no standalone ID — they belong to a parent policy; naive import breaks | MEDIUM | Convention: `policy_name:rule_index` or `policy_name:rule_name`. Document in resource description. |
| Quota policy resource with hard/soft limits | Storage quota enforcement is a top ops requirement; few providers model this correctly | MEDIUM | Separate `flashblade_quota_policy` and `flashblade_quota_policy_rule` with `quota_limit`, `hard_limit_enabled` attributes |
| Object Lock configuration on buckets | Compliance/WORM requirements are increasingly common; model `object_lock_config` and `retention_lock` | HIGH | Map `retention_lock` enum (`ratcheted`, `unlocked`), `object_lock_config.default_retention_mode`, `object_lock_config.default_retention_period` |
| QoS policy attachment on filesystems and buckets | Storage cost control; ops teams need to assign `qos_policy` to control IOPS/bandwidth | MEDIUM | `qos_policy` as a reference attribute (object with `name`) on both `flashblade_file_system` and `flashblade_bucket` |
| Eradication config management | Controls how quickly destroyed resources are permanently deleted; critical for compliance | MEDIUM | `eradication_config.eradication_delay` on filesystem/bucket resources. Distinguish `destroyed=true` (soft delete) from actual DELETE. |
| Destroyed state lifecycle (soft delete) | FlashBlade buckets/filesystems support `destroyed=true` before permanent deletion; naive delete = data loss | HIGH | Implement two-phase delete: PATCH `destroyed=true`, then DELETE. On Read, if `destroyed=true` in API response, surface in state or remove from state based on user intent. |
| Array admin data sources (DNS, NTP, SMTP) | Ops teams need to read array configuration for cross-provider dependencies without managing it | LOW | Read-only data sources for `flashblade_array_dns`, `flashblade_array_ntp` — complementing the resource versions |

### Anti-Features (Commonly Requested, Often Problematic)

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| Performance metrics resources (`/arrays/performance`, `/buckets/performance`) | "We want dashboards in Terraform" | Performance data is time-series, not configuration state. Terraform is a state machine, not a monitoring tool. Every plan would show spurious diffs. | Use Datadog FlashBlade integration or Prometheus purestorage-exporter for metrics. |
| Snapshot management as resources | "We want Terraform to create on-demand snapshots" | Snapshots are operational artifacts, not declarative infrastructure. Lifecycle is incompatible (snapshots accumulate, Terraform would want to destroy old ones). | Snapshot policies (`flashblade_snapshot_policy`) declare the schedule; execution is API-driven. |
| Multi-array management in one provider block | "We have 5 FlashBlades, why configure 5 providers?" | Provider is a single API client; one endpoint, one auth context. Multi-array = multiple provider aliases. Mixing breaks state isolation. | Use Terraform's `provider` alias pattern (`provider "flashblade" { alias = "prod" }`). Already out-of-scope per PROJECT.md. |
| Automatic resource name generation | "Let the provider generate unique names" | Generated names = unstable state. Terraform re-creates resources if names change between runs. Forces random suffix anti-pattern. | User supplies names. Use `random_id` resource from Terraform's random provider if uniqueness is needed. |
| Hardware management (blades, drives) | "We can see `/api/2.22/blades` and `/api/2.22/drives` in the API" | Hardware state is read-only observation, not declarative configuration. Physical hardware can't be created or destroyed via API. | Read-only data sources at most — but hardware topology is outside platform engineering scope. Already out-of-scope per PROJECT.md. |
| Audit log target resources (`/log-targets/file-systems`, `/log-targets/object-store`) | "We want audit logs configured as code" | Audit log targets reference file system and bucket resources, creating circular dependencies in state. | Manage audit log targets via separate operational scripts post-resource-creation, or as a v1.x add-on after dependency ordering is solved. |
| Session/client management (`/file-systems/sessions`, `/file-systems/locks`) | "We want to manage active sessions" | Sessions are ephemeral runtime state. Terraform managing sessions would terminate user connections on `terraform destroy`. | Operational runbooks, not IaC. |

---

## Feature Dependencies

```
flashblade_object_store_account
    └──required by──> flashblade_bucket
                          └──required by──> flashblade_object_store_access_key (via account)
                          └──optional──> flashblade_object_store_access_policy

flashblade_nfs_export_policy
    └──optional attachment──> flashblade_file_system (via policy reference)

flashblade_smb_share_policy
    └──optional attachment──> flashblade_file_system (via policy reference)

flashblade_snapshot_policy
    └──optional attachment──> flashblade_file_system (via policy attachment)

flashblade_quota_policy
    └──optional attachment──> flashblade_file_system (via qos_policy reference)

flashblade_network_access_policy
    └──optional attachment──> flashblade_bucket (via policy reference)
    └──optional attachment──> flashblade_file_system (via policy reference)

flashblade_array_dns / flashblade_array_ntp / flashblade_array_smtp
    └──singleton resources──> no dependencies, managed independently
```

### Dependency Notes

- **`flashblade_bucket` requires `flashblade_object_store_account`:** The `account` field on bucket creation is mandatory. The account must exist before the bucket can be created.
- **`flashblade_object_store_access_key` references account, not bucket:** Access keys are scoped to an object store user/account pair — not directly to a bucket.
- **Policy attachment is optional for resources:** File systems and buckets can be created without any policy attached; policies are attached post-creation via PATCH. This means policy resources are independent — they can be created before or after the resource they attach to.
- **Policy rules require parent policy:** NFS export policy rules, SMB share policy rules, quota rules, etc., cannot exist without their parent policy. Parent policy must be created first.
- **Destroyed state precedes eradication:** Calling DELETE on a bucket/filesystem that has `destroyed=false` is an API error on some resources. Provider must first PATCH `destroyed=true`, then DELETE — or check current state in Delete.

---

## MVP Definition

### Launch With (v1)

Minimum viable set to cover the ops team's high-frequency CRUD use cases.

- [ ] Provider configuration (endpoint, api_token, OAuth2, TLS CA cert, env var fallbacks) — foundation for everything
- [ ] `flashblade_file_system` resource + data source — highest-frequency ops resource
- [ ] `flashblade_object_store_account` resource + data source — required before buckets
- [ ] `flashblade_bucket` resource + data source — second-highest-frequency ops resource
- [ ] `flashblade_object_store_access_key` resource + data source — access management critical path
- [ ] `flashblade_nfs_export_policy` + `flashblade_nfs_export_policy_rule` resource + data source — NFS filesystems are unusable without exports
- [ ] `flashblade_smb_share_policy` + `flashblade_smb_share_policy_rule` resource + data source — Windows workload access
- [ ] `flashblade_snapshot_policy` + `flashblade_snapshot_policy_rule` resource + data source — backup/recovery SLA
- [ ] `flashblade_object_store_access_policy` + `flashblade_object_store_access_policy_rule` resource + data source — S3 IAM equivalent
- [ ] `flashblade_network_access_policy` + `flashblade_network_access_policy_rule` resource + data source — security boundary
- [ ] `flashblade_quota_policy` + `flashblade_quota_policy_rule` resource + data source — storage cost control
- [ ] Array admin resources (DNS, NTP, SMTP, alerts) — `flashblade_array_dns`, `flashblade_array_ntp`, `flashblade_array_smtp`
- [ ] Import support for all resources — adoption of existing infrastructure
- [ ] Drift detection with structured tflog output — compliance requirement per PROJECT.md
- [ ] Unit tests for schema, validators, plan modifiers
- [ ] Mocked API integration tests (CI-safe)
- [ ] Acceptance tests for core resource families (filesystem, bucket, policies)

### Add After Validation (v1.x)

- [ ] Object Lock and WORM bucket configuration — add when compliance/WORM use cases surface
- [ ] QoS policy attachment — add when storage performance management is requested
- [ ] Eradication config management — add when retention/compliance teams engage
- [ ] Additional array admin data sources (read-only array info) — low effort, add on demand
- [ ] Terraform Registry publication — after internal validation proves stability

### Future Consideration (v2+)

- [ ] Pulumi bridge — provider structure intentionally compatible; defer until provider API is stable
- [ ] Bucket replica links (`flashblade_bucket_replica_link`) — DR automation use case, complex state machine
- [ ] File system replica links (`flashblade_file_system_replica_link`) — same DR complexity
- [ ] Array connection management (`flashblade_array_connection`) — multi-array connectivity, requires test infrastructure
- [ ] API client management (`flashblade_api_client`) — security automation, low priority initially
- [ ] Active Directory integration (`flashblade_active_directory`) — domain join via Terraform, high risk

---

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority |
|---------|------------|---------------------|----------|
| Provider authentication + config | HIGH | LOW | P1 |
| `flashblade_file_system` resource | HIGH | MEDIUM | P1 |
| `flashblade_bucket` resource | HIGH | MEDIUM | P1 |
| `flashblade_object_store_account` resource | HIGH | LOW | P1 |
| `flashblade_object_store_access_key` resource | HIGH | MEDIUM | P1 |
| NFS/SMB policy resources | HIGH | MEDIUM | P1 |
| Snapshot policy resource | HIGH | MEDIUM | P1 |
| Object store access policy resource | HIGH | HIGH | P1 |
| Network access policy resource | MEDIUM | MEDIUM | P1 |
| Quota policy resource | MEDIUM | MEDIUM | P1 |
| Import support (all resources) | HIGH | MEDIUM | P1 |
| Drift detection + audit logging | HIGH | MEDIUM | P1 |
| Array admin resources (DNS/NTP/SMTP) | MEDIUM | LOW | P1 |
| Mocked integration tests | HIGH | HIGH | P1 |
| Acceptance tests (core resources) | HIGH | HIGH | P1 |
| Object Lock / WORM config | MEDIUM | HIGH | P2 |
| QoS policy attachment | MEDIUM | MEDIUM | P2 |
| Eradication config management | MEDIUM | MEDIUM | P2 |
| Bucket/FS replica links | LOW | HIGH | P3 |
| Array connection management | LOW | HIGH | P3 |
| API client management | LOW | MEDIUM | P3 |
| Active Directory resource | LOW | HIGH | P3 |

**Priority key:**
- P1: Must have for launch
- P2: Should have, add when possible
- P3: Nice to have, future consideration

---

## Competitor Feature Analysis

| Feature | devans10/terraform-provider-flash (FlashArray) | PureStorage-OpenConnect/terraform-provider-cbs (Cloud Block Store) | Our Approach |
|---------|----------------------------------------------|----------------------------------------------------------------------|--------------|
| Auth methods | API token + username/password | API token | API token + OAuth2 client_credentials (production-grade) |
| Framework | SDK v2 | SDK v2 | terraform-plugin-framework (protocol v6, modern) |
| Import support | Partial (some resources) | Unknown | All resources — non-negotiable |
| Policy management | None (FlashArray doesn't have FlashBlade policy model) | None | Full coverage: 6 policy types + rules as sub-resources |
| Drift detection | Read-based (passive) | Read-based (passive) | Read-based + structured tflog diff output for audit compliance |
| Testing | Unit + acceptance | Unknown | Three-tier: unit + mocked integration + acceptance |
| Sensitive fields | `sensitive` flag on tokens | Unknown | `sensitive` flag + ephemeral resources for auth tokens |
| Data sources | Full coverage per resource | Limited | Full coverage: data source for every resource type |
| Registry status | Published (registry.terraform.io) | Published | Internal first, Registry v1.x |

---

## Sources

- [HashiCorp Terraform Provider Best Practices](https://developer.hashicorp.com/terraform/plugin/best-practices)
- [HashiCorp Provider Design Principles](https://developer.hashicorp.com/terraform/plugin/best-practices/hashicorp-provider-design-principles)
- [terraform-plugin-framework Resources](https://developer.hashicorp.com/terraform/plugin/framework/resources)
- [terraform-plugin-framework Data Sources](https://developer.hashicorp.com/terraform/plugin/framework/data-sources)
- [Resource Import — terraform-plugin-framework](https://developer.hashicorp.com/terraform/plugin/framework/migrating/resources/import)
- [Sensitive State Best Practices](https://developer.hashicorp.com/terraform/plugin/best-practices/sensitive-state)
- [Detecting Drift — SDKv2 (patterns apply to framework)](https://developer.hashicorp.com/terraform/plugin/sdkv2/best-practices/detecting-drift)
- [Plan Modification — terraform-plugin-framework](https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification)
- [Timeouts — terraform-plugin-framework](https://developer.hashicorp.com/terraform/plugin/framework/resources/timeouts)
- [Acceptance Tests — terraform-plugin-framework](https://developer.hashicorp.com/terraform/plugin/framework/acctests)
- [Writing Log Output — tflog](https://developer.hashicorp.com/terraform/plugin/log/writing)
- [devans10/terraform-provider-flash (FlashArray community provider)](https://github.com/devans10/terraform-provider-flash)
- FlashBlade REST API 2.22 reference (`FLASHBLADE_API.md` in repo root)

---
*Feature research for: Terraform provider for Pure Storage FlashBlade*
*Researched: 2026-03-26*
