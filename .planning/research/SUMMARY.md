# Project Research Summary

**Project:** terraform-provider-flashblade
**Domain:** Terraform Provider (Go) — REST API wrapping, enterprise storage infrastructure
**Researched:** 2026-03-26
**Confidence:** HIGH

## Executive Summary

Building a Terraform provider for Pure Storage FlashBlade is a well-understood problem class: wrap a REST API using `terraform-plugin-framework` (protocol v6), expose each API resource family as a Terraform resource + data source pair, and maintain strict separation between the HTTP client layer and provider logic. The FlashBlade REST API v2.22 is a JSON-over-HTTPS API with OAuth2 `client_credentials` auth and name-based resource addressing. The recommended approach is to build the shared HTTP client first (auth, retries, TLS, API versioning), establish a single correct CRUD pattern on the first resource (`flashblade_file_system`), then systematically replicate that pattern across all 12 resource families.

The primary risk is the FlashBlade-specific soft-delete model: buckets and file systems are not immediately eradicated on DELETE — a two-phase destroy (PATCH `destroyed=true` then DELETE) is required, and the provider must poll for eradication before returning. Getting this wrong corrupts acceptance test runs and blocks re-creation of same-named resources. A second systemic risk is computed attribute misuse: FlashBlade returns read-only fields on every response that must be modelled as `Computed: true` with `UseStateForUnknown()` plan modifiers; missing this pattern produces "provider produced inconsistent result" framework panics that are time-consuming to debug.

The scope is clearly bounded: one FlashBlade array per provider instance, no multi-array management in a single block, no performance metrics, no session/lock management. The MVP covers full CRUD for all storage resource families (filesystem, bucket, object store account/access key, 6 policy families, array admin singletons), plus import support, drift detection with structured audit logging, and a three-tier test strategy (unit + mocked integration + acceptance). The codebase targets Go 1.25.5 with `terraform-plugin-framework` v1.19.0.

---

## Key Findings

### Recommended Stack

The stack is fully prescribed by HashiCorp's official scaffolding and has no meaningful alternatives for a new provider. `terraform-plugin-framework` v1.19.0 replaces the deprecated SDKv2 and is the only path to protocol v6 features (plan modifiers, write-only attributes, native diagnostics). Authentication uses `golang.org/x/oauth2/clientcredentials` for OAuth2 token exchange and a custom `http.RoundTripper` for API token session injection — both layered in the HTTP transport, not per-resource. All tooling (`goreleaser`, `tfplugindocs`, `golangci-lint v2`, `tfproviderlint`) follows the official release pipeline pattern.

**Core technologies:**
- `Go 1.25.5` — only supported implementation language; minimum version required by framework v1.19.0
- `terraform-plugin-framework v1.19.0` — protocol v6, plan modifiers, diagnostics; SDKv2 is maintenance-only and must not be used for new providers
- `terraform-plugin-testing v1.15.0` — official acceptance test harness; required for `resource.Test` / `TestCase` / `TestStep`
- `terraform-plugin-log v0.10.0` — structured logging via `tflog`; stdlib `log` bypasses Terraform's log routing
- `golang.org/x/oauth2/clientcredentials` — handles token exchange, auto-refresh, and concurrency for OAuth2 auth
- `net/http` + `crypto/tls` (stdlib) — custom `RoundTripper` chain for auth injection, TLS with custom CA, retries; no third-party HTTP client
- `goreleaser` + `ghaction-terraform-provider-release` — mandatory for Terraform Registry publication

See `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/.planning/research/STACK.md` for full version matrix and installation commands.

### Expected Features

The FlashBlade provider must cover all resource families the ops team interacts with daily. The feature set is larger than a typical provider because FlashBlade has 6 distinct policy families, each modelled as parent resource + child rule resource (12 policy resources total). Every resource requires a companion data source. Import support is non-negotiable: the ops team has existing FlashBlade infrastructure that must be adopted.

**Must have (table stakes):**
- Full CRUD + import for all 12 resource families — Terraform's core contract; partial coverage is unusable
- Accurate drift detection (Read populates ALL attributes, including computed) — ops compliance requirement
- `terraform import` for all resources — adoption of existing infrastructure; composite IDs for policy rules (`policy_name:rule_index`)
- Dual authentication: API token (dev) + OAuth2 client_credentials (production CI/CD)
- TLS with custom CA certificate support — enterprise environments with private PKI
- Sensitive attribute flags on all credential fields — security baseline
- Environment variable fallbacks for all provider config — CI/CD pipelines
- Correct `Computed: true` + `UseStateForUnknown()` on all read-only API fields — prevents "inconsistent result" panics
- Structured logging via `tflog` with drift field-by-field detail — audit compliance
- Three-tier test suite: unit (schema/validators) + mocked integration (CI-safe) + acceptance (real array)
- API versioning via `const APIVersion` + startup negotiation against `GET /api/api_version`

**Should have (differentiators):**
- Structured drift audit log output — field-by-field diff in `tflog.Info` when Read detects state change
- Mocked API integration tests using `httptest.NewServer` — fast CI feedback without a real array
- Full policy family coverage in v1 — competitors ship partial policy support, forcing click-ops fallback
- `force_destroy` attribute on buckets/filesystems — explicit opt-in to bypass soft-delete eradication delay
- Composite import IDs for policy rules — rules have no standalone ID; `policy_name:rule_index` convention
- Destroyed state lifecycle (two-phase soft-delete) — naive single DELETE causes data-safety issues and name-collision failures

**Defer (v2+):**
- Bucket/filesystem replica links — complex state machine, DR use case, requires multi-array test infrastructure
- Array connection management — multi-array connectivity, low initial priority
- Active Directory integration — domain join via Terraform, high operational risk
- Pulumi bridge — defer until provider API is stable
- Terraform Registry publication — after internal validation confirms stability (target v1.x)
- Object Lock / WORM configuration — add when compliance/WORM use cases surface (P2 after core)

See `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/.planning/research/FEATURES.md` for full prioritization matrix and anti-features to avoid.

### Architecture Approach

The architecture is a strict three-layer separation: `internal/client/` (pure Go HTTP client, zero framework imports), `internal/provider/` (thin Terraform resource adapters, one file per resource), and `internal/testmock/` (centralized `httptest` mock server for CI). The provider's `Configure` method builds the single `FlashBladeClient` and injects it into all resources via `resp.ResourceData` — resources never construct HTTP clients themselves. Every Create and Update ends with a `Read` call to sync state from the API's authoritative response.

**Major components:**
1. `main.go` — provider binary entry point, `providerserver.Serve()`
2. `internal/provider/provider.go` — schema definition, `Configure` (client injection), `Resources()` / `DataSources()` factory registries
3. `internal/provider/*_resource.go` — one file per resource type, CRUD + Import; thin adapter pattern calling client methods
4. `internal/provider/*_data_source.go` — read-only listing/lookup per resource type
5. `internal/client/client.go` + `auth.go` + `transport.go` — `FlashBladeClient`, auth flow (API token session + OAuth2), retry RoundTripper, X-Request-ID, TLS
6. `internal/client/*_ops.go` — per-domain API call methods with typed Go structs (no Terraform types)
7. `internal/testmock/server.go` + `handlers/` — `httptest` mock server factory for mocked integration tests

See `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/.planning/research/ARCHITECTURE.md` for full project layout, data flow diagrams, and build order.

### Critical Pitfalls

1. **FlashBlade soft-delete is not Terraform destroy** — `Delete` must PATCH `destroyed=true` then DELETE, then poll for eradication completion. Skipping this causes name-collision failures on re-apply and corrupts acceptance test reruns. Add a `force_destroy` attribute. This must be in the first resource implementation, not a follow-up fix.

2. **Computed attribute misuse → "inconsistent result after apply"** — every field marked `(ro ...)` in the FlashBlade API spec must be `Computed: true` in the schema with `UseStateForUnknown()`. Every Create/Update must end with a `Read` call that writes the full API response to state — never copy from the plan. Missing this produces hard-to-debug framework panics.

3. **Unordered policy rules → perpetual false drift** — policy rules returned by the FlashBlade API may not preserve insertion order. Using `ListNestedAttribute` for rules produces infinite plan diffs. Use `SetNestedAttribute` for unordered collections; verify with an idempotency acceptance test (`apply → plan → 0 diff`).

4. **Incomplete import breaks existing infra adoption** — `ImportState` + `Read` must populate every attribute from the API response. Test `import → plan → 0 diff` for every resource from day one. Do not defer import to a later phase.

5. **Sensitive values leaked via error messages or state** — `api_token`, `client_secret`, and all credential fields must be `Sensitive: true`. The HTTP transport layer must strip `x-auth-token` and `Authorization` headers before passing errors to framework diagnostics. Must be addressed in Phase 1 provider configuration, before any other work.

6. **Missing retry logic for transient API errors** — FlashBlade returns 503/429/5xx during maintenance and load. Exponential backoff must be in the shared HTTP transport (not per-resource). Failing to implement this causes non-deterministic `terraform apply` failures in production.

See `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/.planning/research/PITFALLS.md` for full pitfall catalogue, recovery strategies, and the "looks done but isn't" checklist.

---

## Implications for Roadmap

Architecture research defines a clear build order where each layer depends on the previous. The pitfall catalogue maps directly to phases: most critical pitfalls must be resolved in Phase 1 (client scaffold + first resource), before any resource duplication begins.

### Phase 1: Foundation — Client, Provider Scaffold, and First Resource

**Rationale:** All resources depend on the HTTP client (auth, TLS, retries, API versioning) and the provider's `Configure` injection. Establishing the correct CRUD pattern on one resource (`flashblade_file_system`) creates the template for all subsequent resources. Every Phase 1 pitfall — soft-delete, computed attributes, sensitive values, retry logic, API versioning — must be resolved here before the pattern is replicated 20+ times.

**Delivers:**
- `internal/client/`: `FlashBladeClient`, dual auth (API token + OAuth2), custom TLS, exponential backoff transport, `X-Request-ID`, API version constant + startup negotiation
- `internal/provider/provider.go`: schema (endpoint, auth, TLS), `Configure` client injection, empty `Resources()` / `DataSources()` stubs
- `flashblade_file_system` resource + data source: full CRUD, import, drift detection with `tflog` audit output, `force_destroy`, two-phase soft-delete
- `internal/testmock/`: `httptest` mock server factory for CI-safe integration tests
- Unit tests: schema, validators, plan modifiers
- Integration test: mocked filesystem lifecycle (no real array)
- Security baseline: all credential fields `Sensitive: true`, auth headers stripped from error messages

**Addresses from FEATURES.md:** Provider configuration (P1), `flashblade_file_system` (P1), dual auth (P1), TLS (P1), env var fallbacks (P1), drift detection (P1), `tflog` structured logging (P1), mocked integration tests (P1)

**Avoids from PITFALLS.md:** Soft-delete (#1), computed attribute misuse (#2), sensitive value leakage (#5), missing retry (#6/7), hardcoded API version (#8)

**Research flag:** Standard patterns — `terraform-plugin-framework` provider structure is well-documented. The FlashBlade soft-delete two-phase destroy is the one non-standard element; its implementation is fully specified in PITFALLS.md.

---

### Phase 2: Object Store Resources

**Rationale:** Object store resources (`object_store_account` → `bucket` → `object_store_access_key`) have a strict creation dependency chain that must be respected. Building them second validates the client pattern under dependency constraints and covers the second-highest-frequency ops use case after filesystems.

**Delivers:**
- `flashblade_object_store_account` resource + data source (CRUD, import)
- `flashblade_bucket` resource + data source (CRUD, import, `force_destroy`, two-phase soft-delete, `destroyed` state lifecycle)
- `flashblade_object_store_access_key` resource + data source (CRUD, import, `secret_access_key` sensitive)
- Acceptance tests for full account → bucket → access key lifecycle
- Import acceptance tests for all three resources

**Addresses from FEATURES.md:** `flashblade_object_store_account` (P1), `flashblade_bucket` (P1), `flashblade_object_store_access_key` (P1), `force_destroy` (P1)

**Avoids from PITFALLS.md:** Soft-delete two-phase on buckets (#1), sensitive `secret_access_key` (#5), `ro` fields on bucket space/metadata (#2), pagination on list data sources (#9)

**Research flag:** Standard patterns — object store resources follow the filesystem CRUD template established in Phase 1. No novel patterns.

---

### Phase 3: Policy Resources — NFS, SMB, Snapshot

**Rationale:** These three policy families cover the highest-frequency access control use cases for file-based workloads (NFS for Linux, SMB for Windows, snapshot for backup SLAs). They introduce the parent/child policy+rule pattern that all six policy families share. Building all three together validates the pattern before the remaining three policy families.

**Delivers:**
- `flashblade_nfs_export_policy` + `flashblade_nfs_export_policy_rule` (resource + data source, CRUD, import)
- `flashblade_smb_share_policy` + `flashblade_smb_share_policy_rule` (resource + data source, CRUD, import)
- `flashblade_snapshot_policy` + `flashblade_snapshot_policy_rule` (resource + data source, CRUD, import)
- Composite import ID convention: `policy_name:rule_index` — documented in schema `MarkdownDescription`
- Idempotency acceptance tests: `apply → plan → 0 diff` after rule reorder (validates `SetNestedAttribute` choice)

**Addresses from FEATURES.md:** NFS/SMB policy resources (P1), snapshot policy (P1), composite import IDs (differentiator)

**Avoids from PITFALLS.md:** Unordered policy rules → perpetual false drift (#3); use `SetNestedAttribute` for all rule collections; incomplete import (#4)

**Research flag:** Needs attention — the choice between `SetNestedAttribute` and `ListNestedAttribute` for policy rules, and the handling of computed sub-fields (rule `id`) within set elements, requires careful implementation. The framework's `UseStateForUnknown` is insufficient for set elements with computed fields; custom plan modifiers may be needed.

---

### Phase 4: Policy Resources — Object Store Access, Network Access, Quota

**Rationale:** These three policy families complete the policy coverage and are structurally identical to Phase 3. Object store access policy is the S3 IAM equivalent and has higher schema complexity (IAM-style JSON). Network access and quota policies are simpler.

**Delivers:**
- `flashblade_object_store_access_policy` + `flashblade_object_store_access_policy_rule` (resource + data source, CRUD, import)
- `flashblade_network_access_policy` + `flashblade_network_access_policy_rule` (resource + data source, CRUD, import)
- `flashblade_quota_policy` + `flashblade_quota_policy_rule` (resource + data source, CRUD, import; `quota_limit`, `hard_limit_enabled` attributes)
- Full policy coverage for all 6 policy types — differentiator vs. competing providers

**Addresses from FEATURES.md:** Object store access policy (P1), network access policy (P1), quota policy (P1), full policy family coverage (differentiator)

**Avoids from PITFALLS.md:** Same unordered rules pitfall (#3); same incomplete import pitfall (#4); quota values may be normalized by API → `Computed: true` on adjusted fields (#2)

**Research flag:** Object store access policy rule schema (IAM-style conditions/effects) may require careful modelling against the FlashBlade API spec — review `FLASHBLADE_API.md` during planning.

---

### Phase 5: Array Admin Resources

**Rationale:** Array admin resources (DNS, NTP, SMTP, alert watchers) are singleton-style resources — PATCH-only with no CREATE/DELETE for some. They are lower operational priority than storage resources but required for full IaC coverage of the array. Building them last avoids blocking critical path resources.

**Delivers:**
- `flashblade_array_dns` resource + data source (singleton PATCH pattern)
- `flashblade_array_ntp` resource + data source
- `flashblade_array_smtp` resource + data source
- `flashblade_array_alert_watcher` resource + data source
- Singleton import pattern (no ID argument — resource is the array itself)

**Addresses from FEATURES.md:** Array admin resources (P1), read-only array data sources (P1)

**Avoids from PITFALLS.md:** Singleton resources must not implement `Delete` as a no-op silently — return a diagnostic if destroy is called on a non-destroyable resource (#4 import variant)

**Research flag:** Singleton resource pattern (PATCH-only, no lifecycle CREATE/DELETE) is less common in the Terraform framework. Worth verifying the correct approach to `Delete` (no-op vs. error) in official docs during planning.

---

### Phase 6: Test Hardening, Documentation, Release Preparation

**Rationale:** All resources are implemented; now harden acceptance test coverage to production standards, generate documentation, and prepare the release pipeline. This phase must not be skipped — the "looks done but isn't" checklist from PITFALLS.md applies here.

**Delivers:**
- Full acceptance test suite: every resource has `create → read → update → read → destroy` test
- Import acceptance tests for all resources: `import → plan → 0 diff`
- Idempotency tests for all resources: `apply → plan → 0 diff`
- Pagination tests for all data sources against arrays with >100 resources
- `tfplugindocs` generation from schema descriptions + `examples/` directory
- `goreleaser` + `ghaction-terraform-provider-release` CI pipeline
- Security review: `terraform show planfile` confirms all credentials are redacted; state file audit

**Addresses from FEATURES.md:** Acceptance tests for all resources (P1), Terraform Registry publication preparation (P2)

**Avoids from PITFALLS.md:** Runs the full "looks done but isn't" checklist — computed fields, pagination, sensitive attributes, OAuth2 token refresh under long applies, policy rule idempotency (#1 through #8)

**Research flag:** Standard patterns — goreleaser and Registry publishing workflow is well-documented by HashiCorp. No research needed.

---

### Phase Ordering Rationale

- **Client before resources:** No resource can be implemented without the HTTP client. The client must include retry logic and API versioning from day one — these cannot be retrofitted without modifying every resource.
- **First resource as template:** `flashblade_file_system` is the highest-frequency resource and introduces the soft-delete pitfall. Solving it once, correctly, establishes the pattern all other resources follow. The cost of getting it wrong multiplies across 20+ resources if deferred.
- **Object store before policies:** Bucket depends on object_store_account. Policies are independent of storage resources and can be built in either order, but object store resources are higher operational priority.
- **Policy families grouped:** All 6 policy families share the parent/child pattern. Splitting them across non-adjacent phases would require context-switching back to the same structural pattern.
- **Array admin last:** Lower priority; singleton-style resources differ from normal CRUD; leaves the team free to focus on blocking resources first.
- **Test hardening as a dedicated phase:** Experience with provider development shows that import tests and idempotency tests are consistently deferred and never properly implemented when treated as per-resource afterthoughts. Making them a named phase forces completion.

### Research Flags

Phases likely needing deeper research during planning:
- **Phase 3 (Policy resources — NFS/SMB/snapshot):** The `SetNestedAttribute` + computed sub-field interaction in the framework requires careful validation. Research the correct plan modifier approach for set elements with computed `id` fields before writing the first policy rule resource.
- **Phase 4 (Object store access policy):** Object store access policy rules have IAM-style schema (effect, actions, resources, conditions). Review `FLASHBLADE_API.md` policy rule structure in detail during planning to model the schema correctly before implementation.
- **Phase 5 (Array admin singletons):** The correct framework pattern for singleton resources (PATCH-only, no Create/Delete lifecycle) needs verification — particularly whether `Delete` should be a no-op or return an explicit error.

Phases with standard, well-documented patterns (research-phase optional):
- **Phase 1 (Foundation):** terraform-plugin-framework provider structure is extensively documented by HashiCorp tutorials. The FlashBlade-specific elements (soft-delete, auth) are fully specified in PITFALLS.md and STACK.md.
- **Phase 2 (Object store resources):** Follows the filesystem pattern exactly. Dependency ordering (account → bucket → access key) is clear from FEATURES.md.
- **Phase 6 (Release preparation):** goreleaser + Registry workflow is a solved problem with official HashiCorp tooling.

---

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | All versions verified against official scaffolding `go.mod` (March 2026) and `pkg.go.dev`. No ambiguity. |
| Features | HIGH | HashiCorp official docs + FlashBlade API reference. Feature set is well-bounded by `PROJECT.md` scope. |
| Architecture | HIGH | Official HashiCorp scaffolding + production provider reference (`terraform-provider-hcp`). Patterns are proven. |
| Pitfalls | HIGH (framework) / MEDIUM (FlashBlade-specific) | Framework pitfalls from official sources and verified GitHub issues. FlashBlade API quirks (soft-delete, pagination) inferred from `FLASHBLADE_API.md` — not validated against a live array yet. |

**Overall confidence:** HIGH

### Gaps to Address

- **Soft-delete eradication polling behavior:** The exact API endpoint and poll interval for confirming eradication completion on buckets/filesystems is not documented in `FLASHBLADE_API.md`. Verify against a live array or Pure Storage support during Phase 1 implementation.
- **Policy rule ordering guarantees:** Whether the FlashBlade API preserves insertion order for NFS/SMB/snapshot policy rules is not confirmed from the API spec alone. The `SetNestedAttribute` recommendation is conservative (correct even if order is preserved). Validate during Phase 3 acceptance testing.
- **Object store access policy rule schema:** The full structure of IAM-style conditions/actions/resources in the FlashBlade object store access policy rule is not fully mapped in research. Requires `FLASHBLADE_API.md` deep-dive during Phase 4 planning.
- **Array admin singleton semantics:** Whether `DELETE` on DNS/NTP/SMTP resources resets to defaults or is an error is not confirmed. Verify before implementing Phase 5 `Delete` handlers.
- **OAuth2 token exchange grant type:** The FlashBlade OAuth2 grant type is `urn:ietf:params:oauth:grant-type:token-exchange` — this is non-standard. Confirm the exact request body format against a live array before Phase 1 auth implementation.

---

## Sources

### Primary (HIGH confidence)
- [terraform-provider-scaffolding-framework go.mod](https://raw.githubusercontent.com/hashicorp/terraform-provider-scaffolding-framework/main/go.mod) — authoritative dependency versions (March 2026)
- [terraform-plugin-framework v1.19.0 — pkg.go.dev](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-framework) — version and Go 1.25 requirement
- [terraform-plugin-testing v1.15.0 — pkg.go.dev](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-testing) — version and release date
- [HashiCorp Plugin Framework benefits](https://developer.hashicorp.com/terraform/plugin/framework-benefits) — SDKv2 deprecation rationale
- [HashiCorp Provider Best Practices](https://developer.hashicorp.com/terraform/plugin/best-practices)
- [Terraform Plugin Framework — Providers](https://developer.hashicorp.com/terraform/plugin/framework/providers)
- [Terraform Plugin Framework — Resources Configure](https://developer.hashicorp.com/terraform/plugin/framework/resources/configure)
- [Sensitive state best practices](https://developer.hashicorp.com/terraform/plugin/best-practices/sensitive-state)
- [Resource import — terraform-plugin-framework](https://developer.hashicorp.com/terraform/plugin/framework/resources/import)
- [Plan modification — terraform-plugin-framework](https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification)
- [hashicorp/ghaction-terraform-provider-release](https://github.com/hashicorp/ghaction-terraform-provider-release) — official release workflow
- [golang.org/x/oauth2/clientcredentials](https://pkg.go.dev/golang.org/x/oauth2/clientcredentials) — token flow pattern
- FlashBlade REST API 2.22 reference (`FLASHBLADE_API.md` in repo root) — `destroyed` field, `continuation_token`, `(ro ...)` field annotations, `/api/api_version` endpoint

### Secondary (MEDIUM confidence)
- [hashicorp/terraform-provider-hcp](https://github.com/hashicorp/terraform-provider-hcp) — production provider reference for `internal/clients` + `internal/provider` split pattern
- [golangci-lint v2 announcement](https://ldez.github.io/blog/2025/03/23/golangci-lint-v2/) — v2 config format change
- [bflad/tfproviderlint GitHub](https://github.com/bflad/tfproviderlint) — provider-specific linter (last verified active)
- [GitHub: computed field producing spurious plan changes](https://github.com/hashicorp/terraform-plugin-framework/issues/628) — community-verified pattern
- [GitHub: Provider produced unexpected value after apply](https://github.com/hashicorp/terraform-plugin-framework/issues/840) — community-verified pattern
- [HashiCorp Discuss: SetNestedAttribute with Computed attributes](https://discuss.hashicorp.com/t/dealing-with-unordered-sets-of-complex-objects-setnestedattribute-with-computed-attributes/61874) — community pattern

### Tertiary (LOW confidence)
- [How to Handle Terraform API Rate Limiting](https://oneuptime.com/blog/post/2026-02-23-how-to-handle-terraform-api-rate-limiting/view) — single source; retry pattern confirmed by framework documentation

---

*Research completed: 2026-03-26*
*Ready for roadmap: yes*
