---
gsd_state_version: 1.0
milestone: v1.0
milestone_name: milestone
status: executing
stopped_at: Completed 05-03-PLAN.md — documentation suite approved, all 225 tests pass
last_updated: "2026-03-28T08:21:07.382Z"
last_activity: 2026-03-27 — NAP singleton resource, rule resource, data source — all tests pass (136 total)
progress:
  total_phases: 5
  completed_phases: 5
  total_plans: 20
  completed_plans: 20
  percent: 81
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-03-26)

**Core value:** Operational teams can reliably create, update, delete, and reconcile drift on FlashBlade storage resources through Terraform with zero surprises — every plan reflects reality, every apply converges
**Current focus:** Phase 1 — Foundation

## Current Position

Phase: 4 of 5 (Object/Network/Quota policies and array admin)
Plan: 3 of 5 in current phase
Status: In progress
Last activity: 2026-03-27 — NAP singleton resource, rule resource, data source — all tests pass (136 total)

Progress: [████████░░] 81%

## Performance Metrics

**Velocity:**
- Total plans completed: 0
- Average duration: -
- Total execution time: -

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| - | - | - | - |

**Recent Trend:**
- Last 5 plans: -
- Trend: -

*Updated after each plan completion*
| Phase 01-foundation P01 | 35 | 2 tasks | 17 files |
| Phase 01-foundation P03 | 25 | 4 tasks | 5 files |
| Phase 01-foundation P02 | 52 | 1 tasks | 2 files |
| Phase 01-foundation P04 | 158 | 2 tasks | 11 files |
| Phase 02-object-store-resources P01 | 576 | 2 tasks | 12 files |
| Phase 02-object-store-resources P03 | 265 | 2 tasks | 7 files |
| Phase 02-object-store-resources P02 | 488 | 2 tasks | 8 files |
| Phase 03-file-based-policy-resources P01 | 27 | 2 tasks | 7 files |
| Phase 03-file-based-policy-resources P02 | 35 | 2 tasks | 6 files |
| Phase 03-file-based-policy-resources P04 | 474 | 2 tasks | 6 files |
| Phase 03-file-based-policy-resources P03 | 30 | 2 tasks | 5 files |
| Phase 04-object-network-quota-policies-and-array-admin P01 | 324 | 2 tasks | 9 files |
| Phase 04-object-network-quota-policies-and-array-admin P03 | 12 | 2 tasks | 6 files |
| Phase 04-object-network-quota-policies-and-array-admin P02 | 67 | 2 tasks | 7 files |
| Phase 04-object-network-quota-policies-and-array-admin P05 | 20 | 3 tasks | 10 files |
| Phase 05-quality-hardening P01 | 25 | 2 tasks | 24 files |
| Phase 05-quality-hardening P02 | 8 | 2 tasks | 32 files |
| Phase 05-quality-hardening P04 | 90 | 2 tasks | 19 files |
| Phase 05-quality-hardening P03 | 20 | 2 tasks | 100 files |
| Phase 05-quality-hardening P03 | 20 | 3 tasks | 100 files |

## Accumulated Context

### Decisions

Decisions are logged in PROJECT.md Key Decisions table.
Recent decisions affecting current work:

- [Roadmap]: terraform-plugin-framework over SDK/v2 — modern API, plan modifiers, diagnostics
- [Roadmap]: Three-tier testing — unit + mocked integration (CI-safe) + acceptance (real array)
- [Roadmap]: All 6 policy families in v1 — avoids click-ops fallback for ops team
- [Phase 01-foundation]: Client layer is pure Go with zero terraform-plugin-framework imports — testable with httptest.NewServer
- [Phase 01-foundation]: OAuth2 uses custom FlashBladeTokenSource (token-exchange grant) not standard clientcredentials.Config
- [Phase 01-foundation]: HTTPClient() exported on FlashBladeClient for transport-layer testing without mocking internals
- [Phase 01-foundation]: GetFileSystem synthesizes 404 APIError on empty items list — FlashBlade returns HTTP 200 with empty items for non-existent resources, not HTTP 404
- [Phase 01-foundation]: testmock PATCH handler uses raw map[string]json.RawMessage for true PATCH semantics without overwriting absent fields
- [Phase 01-foundation]: PollUntilEradicated queries ?destroyed=true to avoid race with same-name file system creation
- [Phase 01-foundation]: auth block uses SingleNestedAttribute not SingleNestedBlock — framework recommendation for typed config access
- [Phase 01-foundation]: Configure validates endpoint and auth before calling NewClient — cleaner error messages than client-level errors
- [Phase 01-foundation]: retry_base_delay parsed as time.Duration string in Configure — decoupled from client's internal Duration type
- [Phase 01-foundation]: timeouts uses Attributes() not Block() — consistent with provider auth pattern; Optional not Required
- [Phase 01-foundation]: destroy_eradicate_on_delete defaults true via booldefault.StaticBool — clean teardown is ops-team default
- [Phase 01-foundation]: ImportState initializes timeouts.Value with types.ObjectNull to satisfy timeouts.Type custom serialization
- [Phase 02-object-store-resources]: Object store account name passed as ?names= query param on POST (not in body) — matches FlashBlade API
- [Phase 02-object-store-resources]: Single-phase DELETE for accounts (no soft-delete) with bucket-existence guard before delete
- [Phase 02-object-store-resources]: All Phase 2 model structs added in plan 01 — Bucket, AccessKey models pre-loaded so plans 02-03 skip models.go
- [Phase 02-object-store-resources]: WriteJSONListResponse/WriteJSONError extracted as generic helpers — all mock handlers use package-level functions
- [Phase 02-object-store-resources]: Access key has no ImportState — secret unavailable after creation; all attributes RequiresReplace; Read does not overwrite SecretAccessKey
- [Phase 02-object-store-resources]: destroy_eradicate_on_delete defaults false for buckets — production S3 data safety, eradication is opt-in
- [Phase 02-object-store-resources]: Bucket name and account have RequiresReplace (ForceNew) — S3 immutability semantics
- [Phase 03-file-based-policy-resources]: NFS rule GET model uses int for anonuid/anongid (API integer), PATCH model uses *string (API schema difference confirmed in FLASHBLADE_API.md)
- [Phase 03-file-based-policy-resources]: SnapshotPolicyPatch omits Name field entirely — structurally enforces read-only name constraint for snapshot policies
- [Phase 03-file-based-policy-resources]: Snapshot mock PATCH processes remove_rules before add_rules for atomic replace semantics via ReplaceSnapshotPolicyRule
- [Phase 03-file-based-policy-resources]: NFS policy name has no RequiresReplace — rename is in-place via PATCH
- [Phase 03-file-based-policy-resources]: Rule import uses composite ID policy_name/rule_index resolved via GetNfsExportPolicyRuleByIndex
- [Phase 03-file-based-policy-resources]: readIntoState returns diag.Diagnostics for clean caller composition in rule resource
- [Phase 03-file-based-policy-resources]: Snapshot policy name has RequiresReplace — API does not support rename via PATCH
- [Phase 03-file-based-policy-resources]: SMB policy has no Version field — omitted from schema (unlike NFS export policy)
- [Phase 03-file-based-policy-resources]: Snapshot rule ID is synthetic {policy_name}/{rule_name} — rules have no server-issued UUID
- [Phase 03-file-based-policy-resources]: SMB rule import uses composite ID policy_name/rule_name (string name) — not numeric index like NFS
- [Phase 03-file-based-policy-resources]: Snapshot rule update uses ReplaceSnapshotPolicyRule (atomic PATCH remove+add) — no dedicated rule PATCH endpoint
- [Phase 04-object-network-quota-policies-and-array-admin]: OAP rule POST uses both policy_names= and names= query params (unlike NFS which server-assigns name)
- [Phase 04-object-network-quota-policies-and-array-admin]: NetworkAccessPolicy has no POST/DELETE at policy level — singletons only (GET+PATCH)
- [Phase 04-object-network-quota-policies-and-array-admin]: ArrayNtpPatch sends only ntp_servers field to avoid unintentional modification of other array settings
- [Phase 04-object-network-quota-policies-and-array-admin P03]: NAP Delete=PATCH(enabled=false) — singleton reset pattern, no DELETE endpoint at policy level
- [Phase 04-object-network-quota-policies-and-array-admin P03]: NAP rule readIntoState returns diag.Diagnostics — composition-friendly, mirrors Phase 3 NFS pattern
- [Phase 04-object-network-quota-policies-and-array-admin]: OAP description RequiresReplace — POST-only field, PATCH rejects it; OAP rule effect RequiresReplace — read-only after creation
- [Phase 04-object-network-quota-policies-and-array-admin]: OAP conditions stored as types.String with jsonencode convention — json.RawMessage round-trip at API boundary
- [Phase 04-object-network-quota-policies-and-array-admin]: OAP delete guard: ListObjectStoreAccessPolicyMembers before DELETE — prevents detach errors on policies attached to buckets
- [Phase 04-object-network-quota-policies-and-array-admin]: SMTP alert_watchers nested within SMTP resource (not separate Terraform resources) — single resource manages composite lifecycle
- [Phase 04-object-network-quota-policies-and-array-admin]: DNS Create uses GET-first then PATCH-or-POST — handles both fresh arrays and arrays with existing DNS config
- [Phase 05-quality-hardening]: IsConflict/IsUnprocessable follow exact pattern of IsNotFound for consistent error helper API
- [Phase 05-quality-hardening]: Validator tests call ValidateString/ValidateInt64 directly without full provider spin-up — fast unit tests
- [Phase 05-quality-hardening]: Only bucket, quota_group, quota_user received new validators — other resources lack clear enum/range fields per research guidance
- [Phase 05-quality-hardening]: Pagination loop uses url.Values params object accumulating continuation_token on each iteration — callers see identical return type
- [Phase 05-quality-hardening]: Error-path tests confirmed production code already handled 409/422/404 via AddError — no production changes needed in 05-02
- [Phase 05-quality-hardening]: AccessKey lifecycle is Create->Read->Delete only (no Update — all fields RequiresReplace; no Import — secret unavailable after creation)
- [Phase 05-quality-hardening]: Lifecycle test single mock server pattern reused across all 19 resources — no per-step server restart needed
- [Phase 05-quality-hardening]: go:generate directive placed in main.go — standard Go convention; tfplugindocs discovers it automatically
- [Phase 05-quality-hardening]: docs-check CI job uses hashicorp/setup-terraform action to ensure tfplugindocs can run terraform init during doc generation
- [Phase 05-quality-hardening]: go:generate directive placed in main.go — standard Go convention; tfplugindocs discovers it automatically
- [Phase 05-quality-hardening]: docs-check CI job uses hashicorp/setup-terraform action to ensure tfplugindocs can run terraform init during doc generation

### Pending Todos

None yet.

### Blockers/Concerns

- [Phase 1]: OAuth2 grant type is non-standard (`urn:ietf:params:oauth:grant-type:token-exchange`) — confirm request body format against live array before auth implementation
- [Phase 1]: Soft-delete eradication polling endpoint and poll interval not confirmed in FLASHBLADE_API.md — validate during Phase 1
- [Phase 3]: SetNestedAttribute + computed sub-field interaction in framework requires validation before first policy rule resource
- [Phase 4]: Object store access policy rule IAM schema (conditions/effects) not fully mapped — requires FLASHBLADE_API.md deep-dive during planning
- [Phase 4]: Array admin singleton DELETE semantics (reset to defaults vs. error) unconfirmed

## Session Continuity

Last session: 2026-03-28T08:16:59.800Z
Stopped at: Completed 05-03-PLAN.md — documentation suite approved, all 225 tests pass
Resume file: None
