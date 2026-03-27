# Phase 3: File-Based Policy Resources - Context

**Gathered:** 2026-03-27
**Status:** Ready for planning

<domain>
## Phase Boundary

NFS export, SMB share, and snapshot policy families. Each family has a policy resource and a separate rules resource (6 resources total + 6 data sources). Establishes the parent/child policy pattern reused in Phase 4 for remaining policy families.

</domain>

<decisions>
## Implementation Decisions

### Policy-Rule Modeling
- **Separate resources** for policy and rules: `flashblade_nfs_export_policy` + `flashblade_nfs_export_policy_rule` — rules have independent lifecycle from parent
- **Composite import ID** for rules: `policy_name/rule_index` format (e.g., `terraform import flashblade_nfs_export_policy_rule.x "my-policy/0"`)
- Rules are fully independent: create/delete individual rules without touching the parent policy
- Full API attribute coverage on policy objects — consistent with all other resources
- **Policy delete: fail with diagnostic** if policy is attached to file systems — "Policy is in use, detach first." Consistent with account-bucket guard pattern from Phase 2.
- **Policy rename: in-place** via PATCH — consistent with file system (Phase 1), unlike bucket/account (ForceNew)

### Rule Ordering
- **Claude's discretion** — Claude checks API behavior for index semantics (ordered vs unordered), index assignment (user-specified vs auto), and drift handling. Research should determine:
  - Whether NFS/SMB rule index is mutable (can user reorder?)
  - Whether snapshot rules use the same index pattern
  - Whether to use `SetNestedAttribute` (unordered) or `ListNestedAttribute` (ordered) — or neither since rules are separate resources

### Cross-Policy Consistency
- **Full explicit resource names**: `flashblade_nfs_export_policy`, `flashblade_nfs_export_policy_rule`, `flashblade_smb_share_policy`, `flashblade_smb_share_policy_rule`, `flashblade_snapshot_policy`, `flashblade_snapshot_policy_rule`
- **Same resource pattern** for all 3 families — same skeleton for policy CRUD + rule CRUD, only schema attributes differ
- Snapshot rules (schedule/retention) use the same parent/child pattern as NFS/SMB rules (access) — consistent API surface despite different domain semantics
- Claude decides on per-family tuning based on actual API differences

### Claude's Discretion
- Rule ordering semantics (index handling, Set vs List, drift on reorder)
- Per-family attribute differences and how to handle them
- Mock handler design for policy + rule interactions
- Whether to extract shared policy/rule code into helpers or keep each family self-contained

</decisions>

<specifics>
## Specific Ideas

- This phase establishes the policy pattern that Phase 4 replicates for 3 more families — get it right here
- The ops team manages policies frequently — `apply → plan` must be clean (0 changes if nothing drifted), especially with rule ordering
- Research flagged `SetNestedAttribute` for unordered rules — but since rules are separate resources (not nested), this may not apply. Claude should verify.
- Composite import ID (`policy_name/rule_index`) must work: `terraform import → plan → 0 diff`

</specifics>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/provider/object_store_account_resource.go` — Simple CRUD resource template (no soft-delete), good base for policy resource
- `internal/provider/filesystem_resource.go` — Full CRUD with drift detection, timeouts — for rule resources if they need timeouts
- `internal/client/object_store_accounts.go` — Client CRUD pattern to replicate for policy + rule endpoints
- `internal/testmock/handlers/helpers.go` — Generic mock helpers (WriteJSONListResponse, WriteJSONError)
- `internal/testmock/handlers/object_store_accounts.go` — Mock with cross-reference pattern (accounts check for buckets) — reuse for policies checking for attached file systems

### Established Patterns
- TDD: RED → GREEN, 2 commits per task
- Client: zero terraform-plugin-framework imports
- Mock: thread-safe in-memory state, UUID IDs, cross-reference guards
- Resource: Configure → CRUD → Import → drift logging
- Phase 2 introduced the "fail if in use" guard pattern (account can't delete if buckets exist)

### Integration Points
- `internal/provider/provider.go` — Register 6 new resources + 6 data sources
- `internal/client/models.go` — Add NfsExportPolicy, NfsExportPolicyRule, SmbSharePolicy, etc.
- File system resource already has `nfs_export_policy` and `smb_share_policy` string attributes — policies will be referenced by name

</code_context>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 03-file-based-policy-resources*
*Context gathered: 2026-03-27*
