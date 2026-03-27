# Phase 3: File-Based Policy Resources - Research

**Researched:** 2026-03-26
**Domain:** FlashBlade REST API v2.22 — NFS export policies, SMB share policies, snapshot policies and their rules
**Confidence:** HIGH (all findings sourced directly from FLASHBLADE_API.md)

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- **Separate resources** for policy and rules: `flashblade_nfs_export_policy` + `flashblade_nfs_export_policy_rule` — rules have independent lifecycle from parent
- **Composite import ID** for rules: `policy_name/rule_index` format (e.g., `terraform import flashblade_nfs_export_policy_rule.x "my-policy/0"`)
- Rules are fully independent: create/delete individual rules without touching the parent policy
- Full API attribute coverage on policy objects — consistent with all other resources
- **Policy delete: fail with diagnostic** if policy is attached to file systems — "Policy is in use, detach first." Consistent with account-bucket guard pattern from Phase 2.
- **Policy rename: in-place** via PATCH — consistent with file system (Phase 1), unlike bucket/account (ForceNew)

### Claude's Discretion
- Rule ordering semantics (index handling, Set vs List, drift on reorder)
- Per-family attribute differences and how to handle them
- Mock handler design for policy + rule interactions
- Whether to extract shared policy/rule code into helpers or keep each family self-contained

### Deferred Ideas (OUT OF SCOPE)
None — discussion stayed within phase scope
</user_constraints>

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| NFP-01 | User can create an NFS export policy with name and optional settings | POST `/nfs-export-policies?names=` with `enabled` body field |
| NFP-02 | User can update NFS export policy attributes | PATCH `/nfs-export-policies?names=` — `name` and `enabled` are writable |
| NFP-03 | User can destroy an NFS export policy | DELETE `/nfs-export-policies?names=` — guard against attached file systems first |
| NFP-04 | User can import an existing NFS export policy into Terraform state by name | ImportState by name → Read → populate state |
| NFP-05 | Data source returns NFS export policy attributes by name or filter | GET `/nfs-export-policies?names=` or `?filter=` |
| NFR-01 | User can create NFS export policy rules (client, access, permissions) | POST `/nfs-export-policies/rules` with `policy` ref + rule fields |
| NFR-02 | User can update NFS export policy rules | PATCH `/nfs-export-policies/rules` with `names=` (rule name) + `policy_names=` |
| NFR-03 | User can destroy NFS export policy rules | DELETE `/nfs-export-policies/rules` with `names=` and `policy_names=` |
| NFR-04 | User can import NFS export policy rules using composite ID (policy_name:rule_index) | ImportState parses `policy_name/rule_index`, list rules, find by index |
| SMP-01 | User can create an SMB share policy with name and optional settings | POST `/smb-share-policies?names=` with `enabled` body field |
| SMP-02 | User can update SMB share policy attributes | PATCH `/smb-share-policies?names=` — `name` and `enabled` are writable |
| SMP-03 | User can destroy an SMB share policy | DELETE `/smb-share-policies?names=` — no `versions` param needed |
| SMP-04 | User can import an existing SMB share policy into Terraform state by name | ImportState by name → Read → populate state |
| SMP-05 | Data source returns SMB share policy attributes by name or filter | GET `/smb-share-policies?names=` or `?filter=` |
| SMR-01 | User can create SMB share policy rules | POST `/smb-share-policies/rules` with `policy` ref + rule fields |
| SMR-02 | User can update SMB share policy rules | PATCH `/smb-share-policies/rules` with rule `names=` and `policy_names=` |
| SMR-03 | User can destroy SMB share policy rules | DELETE `/smb-share-policies/rules` |
| SMR-04 | User can import SMB share policy rules using composite ID | ImportState parses composite ID — see SMB rule identity section below |
| SNP-01 | User can create a snapshot policy with name and optional settings | POST `/policies?names=` with `enabled` body field |
| SNP-02 | User can update snapshot policy attributes | PATCH `/policies?names=` — `name` is `ro` in PATCH; `enabled` is writable |
| SNP-03 | User can destroy a snapshot policy | DELETE `/policies?names=` — check `policies/file-systems` first |
| SNP-04 | User can import an existing snapshot policy into Terraform state by name | ImportState by name → Read → populate state |
| SNP-05 | Data source returns snapshot policy attributes by name or filter | GET `/policies?names=` or `?filter=` |
| SNR-01 | User can create snapshot policy rules (schedule, retention) | PATCH `/policies?names=` using `add_rules` array |
| SNR-02 | User can update snapshot policy rules | PATCH `/policies?names=` using `remove_rules` + `add_rules` (replace pattern) |
| SNR-03 | User can destroy snapshot policy rules | PATCH `/policies?names=` using `remove_rules` array |
| SNR-04 | User can import snapshot policy rules using composite ID | Fetch policy, extract rules array, find by index position |
</phase_requirements>

---

## Summary

Phase 3 implements 6 resources and 6 data sources across three policy families. The three families share a common pattern (policy + rules) but differ critically in their rule APIs and identity mechanisms. NFS and SMB rules have dedicated sub-endpoints; snapshot policy rules are managed exclusively through the parent PATCH body. This is the most important architectural asymmetry to plan around.

NFS export policy rules are ordered by `index` (server-assigned, read-only in the InPolicy form). They can be repositioned using `before_rule_id`/`before_rule_name` params on POST/PATCH. The `index` field is writable on the direct rule endpoint but read-only when embedded in the policy object — the server recomputes it after insertions. SMB share policy rules have no `index` — they are unordered and identified by their auto-assigned `name` (read-only). Snapshot policy rules have no dedicated endpoint at all; they are managed inline via `add_rules`/`remove_rules` on PATCH `/api/2.22/policies`.

**Primary recommendation:** NFS rules use index-based import (policy_name/0, /1...), SMB rules use name-based import (policy_name/rule-name), snapshot rules use index (position in the `rules` array) for import. The composite import format `policy_name/rule_index` from the user decision is correct for NFS and snapshot; for SMB the "index" in the import ID should actually be the rule name (since rules have no numeric index). This needs explicit planner decision.

---

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| terraform-plugin-framework | existing | Resource schema, CRUD, import | Already used in phases 1-2 |
| terraform-plugin-framework-timeouts | existing | Timeout blocks | Already used in phases 1-2 |
| terraform-plugin-log/tflog | existing | Drift detection logging | Already used in phases 1-2 |
| encoding/json | stdlib | PATCH raw map semantics | Already used in mock handlers |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| github.com/google/uuid | existing | UUID generation in mock handlers | Mock handler Create operations |
| net/url | stdlib | Query parameter encoding | Client methods |
| strings | stdlib | Join for multi-value params | Client list opts |

### Installation
```bash
# No new dependencies — all libraries already present from Phases 1-2
```

---

## Architecture Patterns

### Recommended Project Structure
```
internal/
├── client/
│   ├── models.go                          # Add 9 new model structs (3 families x 3 types)
│   ├── nfs_export_policies.go             # NFS policy + rule CRUD methods
│   ├── smb_share_policies.go              # SMB policy + rule CRUD methods
│   └── snapshot_policies.go               # Snapshot policy CRUD (rules via PATCH)
├── provider/
│   ├── nfs_export_policy_resource.go      # flashblade_nfs_export_policy resource
│   ├── nfs_export_policy_resource_test.go
│   ├── nfs_export_policy_data_source.go
│   ├── nfs_export_policy_data_source_test.go
│   ├── nfs_export_policy_rule_resource.go # flashblade_nfs_export_policy_rule resource
│   ├── nfs_export_policy_rule_resource_test.go
│   ├── smb_share_policy_resource.go
│   ├── smb_share_policy_resource_test.go
│   ├── smb_share_policy_data_source.go
│   ├── smb_share_policy_data_source_test.go
│   ├── smb_share_policy_rule_resource.go
│   ├── smb_share_policy_rule_resource_test.go
│   ├── snapshot_policy_resource.go
│   ├── snapshot_policy_resource_test.go
│   ├── snapshot_policy_data_source.go
│   ├── snapshot_policy_data_source_test.go
│   ├── snapshot_policy_rule_resource.go
│   └── snapshot_policy_rule_resource_test.go
└── testmock/handlers/
    ├── nfs_export_policies.go             # Mock for both policy + rules endpoints
    ├── smb_share_policies.go
    └── snapshot_policies.go
```

### Pattern 1: Policy Resource (NFS / SMB / Snapshot)
**What:** Standard CRUD resource for policy-level objects. Identical skeleton for all 3 families.
**When to use:** Policy parent resources (6 resources).
**Example:**
```go
// Source: internal/provider/object_store_account_resource.go (existing pattern)
func (r *nfsExportPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    // 1. Read plan
    // 2. POST to /nfs-export-policies?names={name}
    // 3. Read back and set state
}

func (r *nfsExportPolicyResource) Delete(...) {
    // Guard: check if policy is attached to any file systems
    // Use GET /file-systems?filter=nfs.export_policy.name='{name}' OR dedicated endpoint
    // If attached: AddError("Policy is in use, detach first.")
    // If clear: DELETE /nfs-export-policies?names={name}
}
```

### Pattern 2: NFS Rule Resource (index-ordered, dedicated endpoint)
**What:** Rule resource identified by `policy_name` + `index`. Rules are server-ordered.
**When to use:** `flashblade_nfs_export_policy_rule` resource.
**Key behavior:**
- Create: POST `/nfs-export-policies/rules` — body includes `policy` reference; `index` is server-assigned
- Read: GET `/nfs-export-policies/rules?policy_names={policy}&names={rule_name}`
- Update: PATCH `/nfs-export-policies/rules?names={rule_name}&policy_names={policy}`
- Delete: DELETE `/nfs-export-policies/rules?names={rule_name}&policy_names={policy}`
- Import: parse `policy_name/index_str`, list all rules for the policy, find by `index`

```go
// Source: FLASHBLADE_API.md — NfsExportPolicyRule schema
type NfsExportPolicyRule struct {
    ID                       string         `json:"id"`           // ro, UUID
    Name                     string         `json:"name"`         // ro, server-assigned
    Index                    int            `json:"index"`        // ordering key, server-managed
    PolicyVersion            string         `json:"policy_version"` // ro
    Policy                   NamedReference `json:"policy"`
    Access                   string         `json:"access"`       // rw
    Client                   string         `json:"client"`       // rw
    Permission               string         `json:"permission"`   // rw
    Anonuid                  int            `json:"anonuid"`      // rw
    Anongid                  int            `json:"anongid"`      // rw
    Atime                    bool           `json:"atime"`        // rw
    Fileid32bit              bool           `json:"fileid_32bit"` // rw
    Secure                   bool           `json:"secure"`       // rw
    Security                 []string       `json:"security"`     // rw, array of string
    RequiredTransportSecurity string        `json:"required_transport_security"` // rw
}
```

### Pattern 3: SMB Share Rule Resource (unordered, dedicated endpoint, name-identified)
**What:** Rule resource identified by `policy_name` + `rule_name` (the server-assigned `name`). No `index`.
**When to use:** `flashblade_smb_share_policy_rule` resource.
**Key behavior:**
- Create: POST `/smb-share-policies/rules` — body includes `policy` reference; `name` is server-assigned (ro)
- Import composite ID: `policy_name/rule_name` — NOTE: user decision says "rule_index" but SMB has no index.
  The planner must decide: use `rule_name` as the "index" token in the composite ID, or use position-in-list.
  **Recommendation:** use `rule_name` (the server-assigned string name) as the second token for SMB.

```go
// Source: FLASHBLADE_API.md — SmbSharePolicyRule schema
type SmbSharePolicyRule struct {
    ID          string         `json:"id"`           // ro, UUID
    Name        string         `json:"name"`         // ro, server-assigned identifier
    Policy      NamedReference `json:"policy"`
    Principal   string         `json:"principal"`    // rw — user or group
    Change      string         `json:"change"`       // rw — "allow" | "deny"
    FullControl string         `json:"full_control"` // rw — "allow" | "deny"
    Read        string         `json:"read"`         // rw — "allow" | "deny"
}
```

### Pattern 4: Snapshot Rule Resource (managed through parent PATCH, no dedicated endpoint)
**What:** Rule resource that simulates CRUD through PATCH on the parent policy.
**When to use:** `flashblade_snapshot_policy_rule` resource.
**Key behavior:**
- Create: PATCH `/policies?names={policy}` with `add_rules: [{...rule fields}]`
- Read: GET `/policies?names={policy}` → extract from `rules` array by position/name
- Update: PATCH with `remove_rules: [{name: old_name}]` + `add_rules: [{...new fields}]`
- Delete: PATCH with `remove_rules: [{name: rule_name}]`
- Import: GET policy, find rule by index position in `rules` array

```go
// Source: FLASHBLADE_API.md — SnapshotPolicyPatch schema
// Rules are embedded — exact rule field schema not exposed in API reference
// Rule identity: rules have a `name` field (server-assigned) used in remove_rules
// Snapshot rules contain schedule/retention fields (at_time, every_n_*, keep_snapshots_for)
// NOTE: Full snapshot rule field schema requires verification against live array
type SnapshotPolicyRule struct {
    Name          string `json:"name"`         // server-assigned, used to remove
    // schedule fields: at_time, every_n_minutes, every_n_hours, every_n_days
    // retention fields: keep_for, keep_snapshots_for, client_match
    // CONFIDENCE: LOW — not fully enumerated in FLASHBLADE_API.md
}
```

### Anti-Patterns to Avoid
- **Storing index as a stable ID in Terraform state:** NFS rule `index` changes when rules are reordered or deleted externally — the `name` (server-assigned UUID-like string) is the stable identity. Store `name` for API operations; store `index` only as a computed attribute and as the import key.
- **Assuming SMB rules have `index`:** SMB share policy rules have no ordering concept and no index field. The identifier is `name` (server-assigned, read-only).
- **Trying to PATCH snapshot policy name:** The `name` field is marked `ro` in `SnapshotPolicyPatch` — snapshot policies cannot be renamed in-place. This contradicts the user decision for uniform policy rename. The planner must document this exception or verify against a live array.
- **Using `rules` array in POST for inline rule creation:** The user decision requires rules to be separate resources with independent lifecycle. Do not use the `rules` array on policy POST/PATCH.
- **Calling `before_rule_id`/`before_rule_name` without verification:** These params control insertion order. If not used, the API assigns the next available index. For Terraform-managed rules, create at the end (no `before_rule` param) and use `index` as read-only computed state.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Rule CRUD for snapshot policies | Custom rule state tracking | Patch parent with add_rules/remove_rules | No dedicated endpoint exists |
| Parsing composite import ID | Custom parser | `strings.SplitN(id, "/", 2)` | Trivial stdlib, 2-part split |
| Thread-safe in-memory mock state | Custom sync | `sync.Mutex` + maps (existing pattern) | Already proven in Phase 2 handlers |
| Policy attachment check | Custom scan | GET /file-systems with filter or GET /policies/file-systems | API provides membership endpoints |

---

## Common Pitfalls

### Pitfall 1: NFS Rule Index Is Read-Only After Creation (inline form)
**What goes wrong:** `NfsExportPolicyRuleInPolicy.index` is `ro` — the index embedded within the policy object is not settable. Only the direct endpoint (`NfsExportPolicyRule.index`) has a writable `index`, used for repositioning.
**Why it happens:** The API distinguishes between "index in embedded policy context" (read-only) and "index in direct rule context" (writable for repositioning).
**How to avoid:** Mark `index` as `Computed: true` in Terraform schema. Never send `index` in POST body (use `before_rule_id` if ordering is critical). Read `index` from the response and store it as computed state.
**Warning signs:** If plan shows index changing every apply, the implementation is sending `index` in POST.

### Pitfall 2: SMB Share Policy Rules Have No Index
**What goes wrong:** Implementing SMB rule import as `policy_name/0`, `policy_name/1` by position will break when rules are reordered or deleted outside Terraform.
**Why it happens:** SMB rules have no `index` field — they are an unordered collection identified by server-assigned `name`.
**How to avoid:** Use `policy_name/rule_name` as the composite import ID for SMB rules. The `name` field is the stable server-assigned identifier.
**Warning signs:** Import of SMB rules breaks after any external modification.

### Pitfall 3: Snapshot Policy Name Is Read-Only in PATCH
**What goes wrong:** Sending `name` in PATCH `/api/2.22/policies` body to rename a snapshot policy will fail or be silently ignored — `SnapshotPolicyPatch` marks `name` as `ro`.
**Why it happens:** Unlike NFS and SMB policies (`NfsExportPolicyPatch.name` = writable, `SmbSharePolicyPatch.name` = writable), snapshot policy name is immutable after creation.
**How to avoid:** Either use `RequiresReplace` for snapshot policy name (ForceNew), or use a filter GET to verify and document the exception explicitly. Do NOT apply the uniform "rename in-place" pattern to snapshot policies.
**Warning signs:** `terraform plan` after name change shows no-op for snapshot policies when it should show replace.

### Pitfall 4: Snapshot Rules Have No Dedicated API Endpoint
**What goes wrong:** Attempting to GET `/api/2.22/policies/rules` returns 404 — there is no rules sub-endpoint for snapshot policies.
**Why it happens:** The snapshot policy API embeds rules in the policy body and manages them through `add_rules`/`remove_rules` on PATCH, not through a dedicated sub-resource.
**How to avoid:** Snapshot rule Read must GET the parent policy and extract the rule from `rules[]` by index. All mutations go through PATCH on the parent.
**Warning signs:** 404 on any `GET /api/2.22/policies/rules` call.

### Pitfall 5: Policy "In-Use" Guard Requires Correct Membership Endpoint
**What goes wrong:** Trying to DELETE a policy attached to file systems will fail at the API level with an error. Failing to implement the guard means poor UX (API error instead of clear diagnostic).
**Why it happens:** The API refuses to delete policies currently attached to file systems.
**How to avoid:** Before DELETE, call GET `/api/2.22/policies/file-systems?policy_names={name}` (for snapshot) or GET `/api/2.22/file-systems?filter=...` (for NFS/SMB). If result is non-empty, return `AddError("Policy is in use, detach first.")`.
**Warning signs:** Delete returns opaque API error instead of clear Terraform diagnostic.

### Pitfall 6: Rule `name` (server-assigned) vs User-Visible Identity
**What goes wrong:** User-facing Terraform attribute for "rule identity" is confused with the server-assigned `name` field. The server-assigned `name` is an internal identifier, not something users specify.
**Why it happens:** FlashBlade rule resources use `name` as their identity, but for NFS/SMB rules this is auto-generated and not user-settable. The Terraform schema should expose `name` as `Computed` only.
**How to avoid:** NFS rule: expose `name` as Computed (used for PATCH/DELETE targeting). Expose `index` as Computed (used for import). SMB rule: expose `name` as Computed (used as import key). User identifies rules via `policy_name` + resource-level terraform name.

---

## Code Examples

### NFS Export Policy — Client Create Method
```go
// Source: pattern from internal/client/object_store_accounts.go
func (c *FlashBladeClient) PostNfsExportPolicy(ctx context.Context, name string, body NfsExportPolicyPost) (*NfsExportPolicy, error) {
    path := "/nfs-export-policies?names=" + url.QueryEscape(name)
    var resp ListResponse[NfsExportPolicy]
    if err := c.post(ctx, path, body, &resp); err != nil {
        return nil, err
    }
    if len(resp.Items) == 0 {
        return nil, fmt.Errorf("PostNfsExportPolicy: empty response from server")
    }
    return &resp.Items[0], nil
}
```

### NFS Rule — List by Policy
```go
// Source: FLASHBLADE_API.md — GET /nfs-export-policies/rules supports policy_names param
func (c *FlashBladeClient) ListNfsExportPolicyRules(ctx context.Context, policyName string) ([]NfsExportPolicyRule, error) {
    path := "/nfs-export-policies/rules?policy_names=" + url.QueryEscape(policyName)
    var resp ListResponse[NfsExportPolicyRule]
    if err := c.get(ctx, path, &resp); err != nil {
        return nil, err
    }
    return resp.Items, nil
}
```

### NFS Rule — Get by Index (for ImportState)
```go
// Source: FLASHBLADE_API.md — index is a computed attribute on each rule
func (c *FlashBladeClient) GetNfsExportPolicyRuleByIndex(ctx context.Context, policyName string, index int) (*NfsExportPolicyRule, error) {
    rules, err := c.ListNfsExportPolicyRules(ctx, policyName)
    if err != nil {
        return nil, err
    }
    for _, r := range rules {
        if r.Index == index {
            return &r, nil
        }
    }
    return nil, &APIError{StatusCode: 404, Message: fmt.Sprintf("rule at index %d not found in policy %q", index, policyName)}
}
```

### Snapshot Rule — Add via PATCH
```go
// Source: FLASHBLADE_API.md — PATCH /policies uses add_rules/remove_rules
type SnapshotPolicyAddRulesPatch struct {
    AddRules []SnapshotPolicyRulePost `json:"add_rules"`
}

func (c *FlashBladeClient) AddSnapshotPolicyRule(ctx context.Context, policyName string, rule SnapshotPolicyRulePost) (*SnapshotPolicy, error) {
    path := "/policies?names=" + url.QueryEscape(policyName)
    body := SnapshotPolicyAddRulesPatch{AddRules: []SnapshotPolicyRulePost{rule}}
    var resp ListResponse[SnapshotPolicy]
    if err := c.patch(ctx, path, body, &resp); err != nil {
        return nil, err
    }
    if len(resp.Items) == 0 {
        return nil, fmt.Errorf("AddSnapshotPolicyRule: empty response from server")
    }
    return &resp.Items[0], nil
}
```

### SMB Rule Import State Pattern
```go
// Source: composite ID design — SMB uses rule name not index
func (r *smbSharePolicyRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    // ID format: "policy_name/rule_name"
    parts := strings.SplitN(req.ID, "/", 2)
    if len(parts) != 2 {
        resp.Diagnostics.AddError("Invalid import ID", "Expected format: policy_name/rule_name")
        return
    }
    policyName, ruleName := parts[0], parts[1]
    // ... look up rule by name in policy
}
```

### Policy Delete Guard (NFS example)
```go
// Source: pattern from object_store_account_resource.go Delete method
func (r *nfsExportPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    // Check if policy is attached to file systems before deleting
    members, err := r.client.ListNfsExportPolicyMembers(ctx, data.Name.ValueString())
    if err != nil {
        resp.Diagnostics.AddError("Error checking policy attachments before deletion", err.Error())
        return
    }
    if len(members) > 0 {
        resp.Diagnostics.AddError(
            "Cannot delete NFS export policy",
            fmt.Sprintf("Policy %q is attached to %d file system(s). Detach first.", data.Name.ValueString(), len(members)),
        )
        return
    }
    if err := r.client.DeleteNfsExportPolicy(ctx, data.Name.ValueString()); err != nil {
        // ...
    }
}
```

### Mock Handler for Policy + Rules (two-endpoint registration)
```go
// Source: pattern from internal/testmock/handlers/object_store_accounts.go
func RegisterNfsExportPolicyHandlers(mux *http.ServeMux) *nfsExportPolicyStore {
    store := &nfsExportPolicyStore{
        policies: make(map[string]*client.NfsExportPolicy),
        rules:    make(map[string]map[string]*client.NfsExportPolicyRule), // policyName -> ruleName -> rule
        nextIndex: make(map[string]int), // policyName -> next available index
    }
    mux.HandleFunc("/api/2.22/nfs-export-policies", store.handlePolicy)
    mux.HandleFunc("/api/2.22/nfs-export-policies/rules", store.handleRules)
    return store
}
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Inline rules in policy resource (nested blocks) | Separate rule resources with independent lifecycle | User decision (Phase 3) | Rules can be added/removed without touching policy |
| Index as stable ID | Server-assigned `name` as stable ID, `index` as computed ordering | API design | Import uses index to find rule, then persists `name` for targeting |

**Critical API differences across families:**

| Family | Rule Endpoint | Rule Identity | Index | Rename Policy |
|--------|--------------|---------------|-------|---------------|
| NFS export | `/nfs-export-policies/rules` | `name` (ro, server-assigned) | `index` (computed, ordering) | Yes (PATCH name writable) |
| SMB share | `/smb-share-policies/rules` | `name` (ro, server-assigned) | None | Yes (PATCH name writable) |
| Snapshot | No dedicated endpoint (PATCH parent) | `name` (ro, in rules array) | Position in rules array | **No** (PATCH name is `ro`) |

---

## Open Questions

1. **Snapshot policy rule schema fields**
   - What we know: PATCH supports `add_rules` and `remove_rules` arrays; rules have a `name` field (server-assigned)
   - What's unclear: Exact fields in a snapshot policy rule (schedule, retention parameters — `at_time`, `every_n_hours`, `keep_for`, etc.)
   - Recommendation: Look up Pure Storage FlashBlade API documentation for the `PolicyScheduleRule` or `SnapshotRule` schema. The FLASHBLADE_API.md condensed reference omits the rule body detail. Until confirmed, treat snapshot rule schema fields as LOW confidence.

2. **Snapshot policy rename (critical discrepancy)**
   - What we know: `SnapshotPolicyPatch.name` is marked `ro` in FLASHBLADE_API.md
   - What's unclear: Whether this is enforced by the API or just a documentation artifact
   - Recommendation: The planner should flag this as a **known exception** to the user's "uniform rename" policy. Either use `RequiresReplace` for snapshot policy name, or explicitly document that rename is not supported for snapshot policies. Do not silently fail.

3. **"Attached to file systems" check endpoint for NFS/SMB policies**
   - What we know: `/api/2.22/policies/file-systems` exists for snapshot policies; `/api/2.22/file-systems/policies` exists for file system side
   - What's unclear: Whether there is a direct reverse-lookup endpoint for NFS/SMB policies (e.g., `/nfs-export-policies/file-systems` or `/nfs-export-policies/members`)
   - Recommendation: Use `GET /api/2.22/file-systems?filter=nfs.export_policy.name='{policy_name}'` as the guard for NFS, similarly for SMB. Alternatively, attempt DELETE and handle the API error as the guard (less clean UX).

4. **Composite import ID format for SMB rules**
   - What we know: User decision says `policy_name/rule_index` but SMB has no `index`
   - What's unclear: Whether to use rule `name` (server-assigned string) or position in list
   - Recommendation: Use `policy_name/rule_name` for SMB import ID. Document clearly. The user decision's "rule_index" token maps to the rule's identifying attribute for each family — `index` for NFS/snapshot, `name` for SMB.

---

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go standard testing + terraform-plugin-framework test helpers |
| Config file | none (go test standard) |
| Quick run command | `go test ./internal/provider/ -run TestNfsExportPolicy -v -timeout 30s` |
| Full suite command | `go test ./... -timeout 120s` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| NFP-01 | Create NFS export policy | unit | `go test ./internal/provider/ -run TestNfsExportPolicyResource_Create -v` | ❌ Wave 0 |
| NFP-02 | Update NFS export policy | unit | `go test ./internal/provider/ -run TestNfsExportPolicyResource_Update -v` | ❌ Wave 0 |
| NFP-03 | Delete NFS export policy with guard | unit | `go test ./internal/provider/ -run TestNfsExportPolicyResource_Delete -v` | ❌ Wave 0 |
| NFP-04 | Import NFS export policy by name | unit | `go test ./internal/provider/ -run TestNfsExportPolicyResource_Import -v` | ❌ Wave 0 |
| NFP-05 | NFS export policy data source | unit | `go test ./internal/provider/ -run TestNfsExportPolicyDataSource -v` | ❌ Wave 0 |
| NFR-01 | Create NFS export policy rule | unit | `go test ./internal/provider/ -run TestNfsExportPolicyRuleResource_Create -v` | ❌ Wave 0 |
| NFR-02 | Update NFS export policy rule | unit | `go test ./internal/provider/ -run TestNfsExportPolicyRuleResource_Update -v` | ❌ Wave 0 |
| NFR-03 | Delete NFS export policy rule | unit | `go test ./internal/provider/ -run TestNfsExportPolicyRuleResource_Delete -v` | ❌ Wave 0 |
| NFR-04 | Import NFS export policy rule by composite ID | unit | `go test ./internal/provider/ -run TestNfsExportPolicyRuleResource_Import -v` | ❌ Wave 0 |
| SMP-01 | Create SMB share policy | unit | `go test ./internal/provider/ -run TestSmbSharePolicyResource_Create -v` | ❌ Wave 0 |
| SMP-03 | Delete SMB share policy with guard | unit | `go test ./internal/provider/ -run TestSmbSharePolicyResource_Delete -v` | ❌ Wave 0 |
| SMR-01 | Create SMB share policy rule | unit | `go test ./internal/provider/ -run TestSmbSharePolicyRuleResource_Create -v` | ❌ Wave 0 |
| SMR-04 | Import SMB share policy rule by name | unit | `go test ./internal/provider/ -run TestSmbSharePolicyRuleResource_Import -v` | ❌ Wave 0 |
| SNP-01 | Create snapshot policy | unit | `go test ./internal/provider/ -run TestSnapshotPolicyResource_Create -v` | ❌ Wave 0 |
| SNP-02 | Update snapshot policy (no rename) | unit | `go test ./internal/provider/ -run TestSnapshotPolicyResource_Update -v` | ❌ Wave 0 |
| SNR-01 | Create snapshot policy rule via add_rules | unit | `go test ./internal/provider/ -run TestSnapshotPolicyRuleResource_Create -v` | ❌ Wave 0 |
| SNR-04 | Import snapshot policy rule by index | unit | `go test ./internal/provider/ -run TestSnapshotPolicyRuleResource_Import -v` | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./internal/... -run TestNfs -timeout 30s` (family-scoped)
- **Per wave merge:** `go test ./... -timeout 120s`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/testmock/handlers/nfs_export_policies.go` — mock for NFP + NFR tests
- [ ] `internal/testmock/handlers/smb_share_policies.go` — mock for SMP + SMR tests
- [ ] `internal/testmock/handlers/snapshot_policies.go` — mock for SNP + SNR tests
- [ ] All 12 resource/data source test files — none exist yet
- [ ] Models in `internal/client/models.go` — Phase 3 structs not yet defined

---

## Sources

### Primary (HIGH confidence)
- `FLASHBLADE_API.md` — FlashBlade REST API v2.22 condensed reference — all endpoint URLs, HTTP methods, body fields, query params, and type schemas for NFS export policies, SMB share policies, snapshot policies and their rules
- `internal/client/models.go` — existing model patterns for Phase 3 structs
- `internal/client/object_store_accounts.go` — client method pattern to replicate
- `internal/provider/object_store_account_resource.go` — resource pattern to replicate
- `internal/testmock/handlers/object_store_accounts.go` — mock handler pattern to replicate
- `.planning/phases/03-file-based-policy-resources/03-CONTEXT.md` — user decisions

### Secondary (MEDIUM confidence)
- `internal/testmock/handlers/helpers.go` — WriteJSONListResponse/WriteJSONError patterns confirmed usable for Phase 3 mocks

### Tertiary (LOW confidence)
- Snapshot policy rule field schema (at_time, every_n_*, keep_for) — inferred from general FlashBlade knowledge, not found in FLASHBLADE_API.md condensed reference. Requires verification.

---

## Metadata

**Confidence breakdown:**
- NFS export policy + rule API: HIGH — full schema in FLASHBLADE_API.md
- SMB share policy + rule API: HIGH — full schema in FLASHBLADE_API.md; notable absence of `index`
- Snapshot policy API: HIGH for endpoints and PATCH structure; LOW for rule body schema (opaque `rules` array)
- Index semantics: HIGH — `NfsExportPolicyRuleInPolicy.index` explicitly marked `ro`; `NfsExportPolicyRule.index` explicitly writable (for reorder)
- Snapshot rename exception: HIGH — `SnapshotPolicyPatch.name` explicitly marked `ro`
- Policy attachment guard endpoints: MEDIUM — `/policies/file-systems` confirmed for snapshot; NFS/SMB reverse-lookup endpoint not confirmed

**Research date:** 2026-03-26
**Valid until:** 2026-04-26 (API version pinned to 2.22 — stable)
