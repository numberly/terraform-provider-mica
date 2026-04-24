# Phase 4: Object/Network/Quota Policies and Array Admin - Context

**Gathered:** 2026-03-27
**Status:** Ready for planning

<domain>
## Phase Boundary

Three remaining policy families (object store access, network access, quota) following Phase 3's parent/child pattern, plus singleton array administration resources (DNS, NTP, SMTP) with a new lifecycle pattern (no create/delete — read/update only with reset-on-delete).

</domain>

<decisions>
## Implementation Decisions

### Array Admin Singletons (DNS, NTP, SMTP)
- **Lifecycle**: Create = read current + PATCH config. Read = GET. Update = PATCH. Delete = PATCH back to empty/default values (reset).
- **Destroy behavior**: Reset to defaults — clear nameservers, NTP servers, SMTP relay. Explicit and clean lifecycle.
- **Import ID**: `default` — there's only one instance per array (`terraform import flashblade_array_dns.x default`)
- **Data sources**: Read-only data sources alongside resources for cross-provider references

### SMTP / Alerts Config
- **Flat attributes**: relay_host, sender_domain as top-level strings — no nested blocks
- **Alert watchers in same resource**: `flashblade_array_smtp` manages relay config + alert recipients together (not separate resources)

### Object Store Access Policy Rules (IAM-style)
- **Conditions**: JSON string — `conditions = jsonencode({...})` — flexible, like AWS IAM policy documents
- **Actions**: List of strings — `actions = ["s3:GetObject", "s3:PutObject"]`
- **Effect**: Simple string — `effect = "allow" | "deny"` — validated at plan time
- **Same parent/child pattern**: `flashblade_object_store_access_policy` + `flashblade_object_store_access_policy_rule`

### Network Access Policy
- **Full API coverage** — all fields: client, interfaces, effect, description
- **Same attachment pattern**: String reference on both bucket and file system resources — already supported from Phase 1/2
- **Same parent/child pattern as Phase 3**

### Quota Policy
- **quota_limit**: Integer input in bytes + computed `quota_limit_display` showing human-readable value (e.g., "1 TB")
- **Enforcement**: Boolean `enforced` attribute
- **Same parent/child pattern as Phase 3**

### Carried Forward from Phase 3 (applies to all 3 policy families)
- Separate resources for policy + rules
- Full explicit naming (`flashblade_object_store_access_policy_rule`, etc.)
- Policy delete guard if attached to resources
- In-place rename on policy (unless API restricts — check per family)
- Composite import ID: `policy_name/rule_index` or `policy_name/rule_name` depending on API
- Independent rule lifecycle
- Full API attribute coverage

### Claude's Discretion
- Per-family rule API differences (dedicated endpoint vs PATCH-based — research determines)
- Import ID format per family (index vs name)
- Policy rename capability per family (check API)
- Mock handler implementation details
- Default values for singleton reset-on-delete

</decisions>

<specifics>
## Specific Ideas

- Phase 4 is the largest phase (32 requirements) — efficient execution matters
- The 3 policy families should follow Phase 3 patterns closely — minimal new design needed
- Array admin singletons are a genuinely new pattern — get it right for DNS first, replicate for NTP/SMTP
- Object store access policy's IAM-style conditions (JSON string) is the most complex schema in the project

</specifics>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/provider/nfs_export_policy_resource.go` — Policy resource template (CRUD, rename, delete guard)
- `internal/provider/nfs_export_policy_rule_resource.go` — Rule resource template (composite import, independent lifecycle)
- `internal/provider/smb_share_policy_rule_resource.go` — Name-based rule import variant
- `internal/provider/snapshot_policy_resource.go` — RequiresReplace name variant (if needed for any Phase 4 policy)
- `internal/client/nfs_export_policies.go` — Client CRUD pattern for policies + rules

### Established Patterns
- 3 distinct rule API patterns proven in Phase 3: dedicated endpoint with index, dedicated endpoint with name, PATCH-based
- Delete guard: check for attached resources before allowing policy deletion
- TDD RED/GREEN with mock server
- 101 tests currently passing

### Integration Points
- `internal/provider/provider.go` — Currently registers 10 resources + 7 data sources. Phase 4 adds ~10 more resources + ~7 data sources.
- `internal/client/models.go` — Add ObjectStoreAccessPolicy, NetworkAccessPolicy, QuotaPolicy, ArrayDns, ArrayNtp, ArraySmtp models
- Bucket and file system resources already have policy reference string attributes

</code_context>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 04-object-network-quota-policies-and-array-admin*
*Context gathered: 2026-03-27*
