# Phase 7: S3 Export Policies & Virtual Hosts - Research

**Researched:** 2026-03-28
**Domain:** FlashBlade S3 export policies (IAM-style rules) and object store virtual host management
**Confidence:** HIGH

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| S3P-01 | Operator can create an S3 export policy with enable/disable toggle | API: POST /s3-export-policies body has `enabled`(boolean). Pattern: identical to NFS export policy structure. |
| S3P-02 | Operator can create S3 export policy rules with actions/effect/resources (IAM-style) | API: POST /s3-export-policies/rules body has `actions`(array), `effect`(string), `resources`(array). Direct analog to OAP rule resource (fully mapped in codebase). |
| S3P-03 | Operator can update and delete S3 export policy rules independently | API: PATCH /s3-export-policies/rules (same fields), DELETE by names+policy_names params. Established pattern in NFS/OAP rule resources. |
| S3P-04 | Operator can import S3 export policies and rules into Terraform state | Policy: import by name. Rules: composite ID `policy_name/rule_name` (OAP pattern) or `policy_name/rule_index` (NFS pattern). API PATCH body has `effect` writable → use `policy_name/rule_name` like OAP. |
| VH-01 | Operator can create a virtual host with hostname and attached servers | API: POST /object-store-virtual-hosts body has `hostname`(string) and `attached_servers`(array of NamedReference). |
| VH-02 | Operator can update attached servers list on a virtual host | API: PATCH body supports `attached_servers`(replace), `add_attached_servers`(additive), `remove_attached_servers`(subtractive). |
| VH-03 | Operator can import an existing virtual host into Terraform state | Virtual host name comes from `name`(ro) in GET response. Import by name. |
</phase_requirements>

---

## Summary

Phase 7 introduces two resource families: S3 export policies with IAM-style rules, and object store virtual hosts. Both are well-covered by existing codebase patterns — no new architectural territory.

S3 export policy and its rules are structurally identical to the Object Store Access Policy (OAP) family (Phase 4). Rules have `actions`, `effect`, and `resources` — the same three fields as OAP rules, with no `conditions` or `name` field in the POST/PATCH body. The key difference from OAP rules: rule name is server-assigned (like NFS rules), not operator-supplied. Import strategy therefore uses `policy_name/rule_index` (NFS pattern) rather than `policy_name/rule_name` (OAP pattern) — unless the API returns a stable server-assigned name on every GET.

Virtual hosts introduce an `attached_servers` list managed with add/remove semantics in PATCH. The API body lists `name`(ro) meaning the virtual host's name is derived at POST time from the `hostname` parameter (passed via `?names=` or implied). The PATCH body also exposes `name`(string) writable, meaning rename is supported. The `realms` and `context` fields are read-only; `hostname` is the primary user-supplied field for both POST and PATCH.

**Primary recommendation:** Model S3 export policy after NFS export policy, model S3 export policy rule after OAP rule (dropping conditions and rule-name). Model virtual host as a flat resource with a `types.List` of server name references and standard add/remove PATCH semantics.

---

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| terraform-plugin-framework | (project existing) | Resource/datasource scaffolding | Already in use across all 22 resources |
| terraform-plugin-framework-timeouts | (project existing) | Timeout blocks on all CRUD ops | Established in all resources since Phase 1 |
| terraform-plugin-framework-validators | (project existing) | `stringvalidator.OneOf` for `effect` | Used in OAP rule for allow/deny validation |
| internal/client | (local) | Pure Go HTTP client, zero framework imports | Project decision v1.0 |
| internal/testmock | (local) | httptest.NewServer mock infrastructure | Project decision v1.0 |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| github.com/google/uuid | (project existing) | UUID generation in mock handlers | All mock handler POST implementations |

---

## Architecture Patterns

### Recommended Project Structure

New files follow the existing naming convention exactly:

```
internal/client/
├── s3_export_policies.go          # S3 export policy + rules client methods
├── object_store_virtual_hosts.go  # Virtual host client methods
internal/provider/
├── s3_export_policy_resource.go
├── s3_export_policy_rule_resource.go
├── object_store_virtual_host_resource.go
internal/testmock/handlers/
├── s3_export_policies.go
├── object_store_virtual_hosts.go
internal/provider/
├── s3_export_policy_resource_test.go
├── s3_export_policy_rule_resource_test.go
├── object_store_virtual_host_resource_test.go
```

### Pattern 1: Policy Resource (S3 Export Policy mirrors NFS Export Policy)

The `flashblade_s3_export_policy` resource is structurally identical to `flashblade_nfs_export_policy`:
- POST: `?names=<name>` query param, body `{ "enabled": bool }`
- PATCH: `?names=<name>`, body supports `name` and `enabled`
- DELETE: `?names=<name>`
- Import: by policy name

Key fields: `id`(computed), `name`(required), `enabled`(optional+computed, default true).

Note: API PATCH body includes `rules`(array) — this is the inline rules path. Do NOT use it; manage rules independently via the rules sub-resource (same approach as NFS/OAP).

**Model struct pattern (mirrors nfsExportPolicyModel):**
```go
// Source: internal/provider/nfs_export_policy_resource.go
type s3ExportPolicyModel struct {
    ID       types.String   `tfsdk:"id"`
    Name     types.String   `tfsdk:"name"`
    Enabled  types.Bool     `tfsdk:"enabled"`
    IsLocal  types.Bool     `tfsdk:"is_local"`
    PolicyType types.String `tfsdk:"policy_type"`
    Version  types.String   `tfsdk:"version"`
    Timeouts timeouts.Value `tfsdk:"timeouts"`
}
```

### Pattern 2: Rule Resource (S3 Export Policy Rule)

S3 export policy rules are structurally similar to OAP rules, but with important differences:

| Attribute | OAP Rule | S3 Export Rule |
|-----------|----------|----------------|
| `name` | Required, user-supplied, RequiresReplace | Server-assigned (computed), like NFS rules |
| `effect` | Required, RequiresReplace | Writable via PATCH (not RequiresReplace) |
| `actions` | Required, list | Required, list |
| `resources` | Required, list | Required, list |
| `conditions` | Optional, JSON string | Not present |
| Import ID | `policy_name/rule_name` | `policy_name/rule_index` (index-based, like NFS) |

API body for POST: `{ "actions": [...], "effect": "allow|deny", "resources": [...] }` — rule name is NOT in POST body, it is server-assigned.

PATCH body: `{ "actions": [...], "effect": "...", "resources": [...] }` — effect IS patchable (unlike OAP where effect has RequiresReplace).

**Model struct pattern:**
```go
type s3ExportPolicyRuleModel struct {
    ID         types.String   `tfsdk:"id"`
    PolicyName types.String   `tfsdk:"policy_name"`
    Name       types.String   `tfsdk:"name"`       // computed, server-assigned
    Index      types.Int64    `tfsdk:"index"`      // computed, for import
    Effect     types.String   `tfsdk:"effect"`     // required, patchable
    Actions    types.List     `tfsdk:"actions"`    // required
    Resources  types.List     `tfsdk:"resources"`  // required
    Timeouts   timeouts.Value `tfsdk:"timeouts"`
}
```

### Pattern 3: Virtual Host Resource

Virtual hosts are a flat resource with a `hostname` and a `[]NamedReference` list for `attached_servers`.

Key API behaviors:
- POST: body fields `hostname`(string), `attached_servers`(array of `{name: string}`)
- PATCH: `?names=<name>`, body supports `name`(rename), `hostname`(change), `attached_servers`(full replace), `add_attached_servers`(additive), `remove_attached_servers`(subtractive)
- `name` in GET is read-only (server-derived from hostname); import by `name`
- `realms` and `context` are read-only — expose as computed if needed but not required by requirements

**Recommended PATCH strategy:** Use full-replace `attached_servers` (not add/remove) for simplicity. The Terraform model holds the desired state as a list; on Update, send the full desired list. This avoids tracking diff state between old/new sets.

**Model struct:**
```go
type objectStoreVirtualHostModel struct {
    ID              types.String   `tfsdk:"id"`
    Name            types.String   `tfsdk:"name"`
    Hostname        types.String   `tfsdk:"hostname"`
    AttachedServers types.List     `tfsdk:"attached_servers"` // list of server names (string)
    Timeouts        timeouts.Value `tfsdk:"timeouts"`
}
```

`attached_servers` stored as `types.List` with `types.StringType` elements (just server names). The NamedReference wiring is an implementation detail of the client layer.

**Virtual host name vs hostname:** The `name` field in the API is read-only (computed from hostname at creation time). It is used for `?names=` query params in PATCH/DELETE/GET. `hostname` is writable. Import uses `name`.

### Anti-Patterns to Avoid

- **Using inline `rules` in POST/PATCH body for policies:** Always manage rules as separate resources. Inline rule creation via POST body creates state management complexity.
- **Using add/remove_attached_servers instead of full replace:** Increases implementation complexity with no benefit in typical Terraform workflows. Full replace is idiomatic.
- **RequiresReplace on effect for S3 rules:** Unlike OAP, effect IS patchable. Do not copy OAP's RequiresReplace on effect.
- **Operator-supplied rule name for S3 rules:** S3 export policy rules have server-assigned names. Do not model as OAP (where name is user-supplied). Model as NFS (server-assigned, index-based import).

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| HTTP PATCH semantics | Custom delta logic | `map[string]json.RawMessage` decode (existing pattern) | Eliminates false updates on untouched fields |
| Pagination | Manual loop | Existing `ListResponse[T]` + continuation_token pattern | Already tested, handles edge cases |
| 404 synthesis | Custom error type | `APIError{StatusCode: 404}` + `client.IsNotFound()` | Established project-wide pattern |
| Test server | httptest setup | `testmock.NewMockServer()` + `handlers.Register*` | One-line test server initialization |
| Timeout blocks | Manual context.WithTimeout | `timeouts.Attributes()` + `timeouts.Value` | Consistent with all 22 existing resources |

---

## Common Pitfalls

### Pitfall 1: S3 Rule Name is Server-Assigned (Not User-Supplied)

**What goes wrong:** Developer copies OAP rule pattern and makes `name` a Required field. POST body then includes `name` which the API rejects (HTTP 400).

**Why it happens:** OAP rules have user-supplied names; S3 export policy rules follow the NFS pattern where names are server-assigned UUIDs.

**How to avoid:** Make `name` Computed + UseStateForUnknown. Read the server-assigned name from POST response. Use it for subsequent PATCH/DELETE calls.

**Warning signs:** API returns 400 on rule POST with name in body.

### Pitfall 2: Virtual Host `name` vs `hostname`

**What goes wrong:** Developer uses `hostname` as the Terraform resource identifier and tries to look up the VH by hostname on Read/Update/Delete.

**Why it happens:** Confusion between the writable `hostname` field and the read-only `name` field derived at creation.

**How to avoid:** Track `name` (server-assigned, read-only) as the Terraform ID and `?names=` query parameter. Track `hostname` as a separate writable attribute. On import, use `name` not `hostname`.

**Warning signs:** GET/PATCH/DELETE return 404 when using hostname as the query parameter.

### Pitfall 3: S3 Rule `effect` Is Patchable (Unlike OAP)

**What goes wrong:** Developer adds RequiresReplace to `effect` (copied from OAP), forcing unnecessary resource replacement on effect changes.

**Why it happens:** Direct copy from OAP rule without checking the S3 export policy API spec.

**How to avoid:** API spec confirms PATCH body accepts `effect`. Do not add RequiresReplace. A simple PATCH is sufficient.

**Warning signs:** `terraform plan` shows destroy/create when only effect changes.

### Pitfall 4: Virtual Host POST — `name` is Read-Only

**What goes wrong:** Developer tries to pass `name` in the POST body. The API marks it as `(ro string)`.

**Why it happens:** POST body spec lists `name`(ro string) which means it appears in GET responses but cannot be set on POST.

**How to avoid:** POST only `hostname` and `attached_servers`. Read `name` from the POST response. Use `name` for subsequent API calls.

**Warning signs:** POST returns the resource with a different `name` value than what was sent, or returns 400.

### Pitfall 5: Virtual Host POST Query Parameter

**What goes wrong:** Virtual host creation uses `?names=` instead of whatever the actual query param is.

**Why it happens:** Most resources use `?names=` for POST but servers use `?create_ds=`. Virtual hosts may follow a different convention.

**How to avoid:** Verify the actual query parameter accepted by POST /object-store-virtual-hosts from a live API. The API spec entry says `POST | Body: attached_servers, hostname` with no explicit param note, suggesting `?names=` is correct — but confirm.

**Warning signs:** POST returns 400 "names query parameter is required".

**Recommendation:** Use `?names=<hostname>` for POST (consistent with policy resources) and read the server-assigned `name` from the response.

### Pitfall 6: S3 Policy `version` Field and PATCH Versioning

**What goes wrong:** The API PATCH body for s3-export-policies lists `rules`(array) — a developer might try to manage rules inline in the policy PATCH, leading to rules being wiped on policy updates.

**Why it happens:** Some FlashBlade policy types use inline rule management via add_rules/remove_rules (snapshot policies). S3 export policies have a separate `/rules` endpoint.

**How to avoid:** Always use the dedicated `/s3-export-policies/rules` endpoint for rule CRUD. Never send `rules` in the policy PATCH body.

---

## Code Examples

### S3 Export Policy POST (client layer pattern)

```go
// Source: internal/client/nfs_export_policies.go (adapted for S3)
func (c *FlashBladeClient) PostS3ExportPolicy(ctx context.Context, name string, body S3ExportPolicyPost) (*S3ExportPolicy, error) {
    path := "/s3-export-policies?names=" + url.QueryEscape(name)
    var resp ListResponse[S3ExportPolicy]
    if err := c.post(ctx, path, body, &resp); err != nil {
        return nil, err
    }
    if len(resp.Items) == 0 {
        return nil, fmt.Errorf("PostS3ExportPolicy: empty response from server")
    }
    return &resp.Items[0], nil
}
```

### S3 Export Policy Rule POST (client layer — policy_names query param)

```go
// Source: internal/client/nfs_export_policies.go (adapted)
func (c *FlashBladeClient) PostS3ExportPolicyRule(ctx context.Context, policyName string, body S3ExportPolicyRulePost) (*S3ExportPolicyRule, error) {
    path := "/s3-export-policies/rules?policy_names=" + url.QueryEscape(policyName)
    var resp ListResponse[S3ExportPolicyRule]
    if err := c.post(ctx, path, body, &resp); err != nil {
        return nil, err
    }
    if len(resp.Items) == 0 {
        return nil, fmt.Errorf("PostS3ExportPolicyRule: empty response from server")
    }
    return &resp.Items[0], nil
}
```

### Virtual Host POST (client layer)

```go
// Proposed pattern: POST with ?names= query param
func (c *FlashBladeClient) PostObjectStoreVirtualHost(ctx context.Context, hostname string, body ObjectStoreVirtualHostPost) (*ObjectStoreVirtualHost, error) {
    // hostname is used as the names param; the API returns a server-assigned name
    path := "/object-store-virtual-hosts?names=" + url.QueryEscape(hostname)
    var resp ListResponse[ObjectStoreVirtualHost]
    if err := c.post(ctx, path, body, &resp); err != nil {
        return nil, err
    }
    if len(resp.Items) == 0 {
        return nil, fmt.Errorf("PostObjectStoreVirtualHost: empty response from server")
    }
    return &resp.Items[0], nil
}
```

### Virtual Host PATCH with full-replace attached_servers

```go
// Source pattern: internal/client/nfs_export_policies.go
type ObjectStoreVirtualHostPatch struct {
    Name            *string          `json:"name,omitempty"`
    Hostname        *string          `json:"hostname,omitempty"`
    AttachedServers []NamedReference `json:"attached_servers,omitempty"` // full replace
}
```

### Mock handler PATCH with raw map for true PATCH semantics

```go
// Source: internal/testmock/handlers/nfs_export_policies.go
var rawPatch map[string]json.RawMessage
if err := json.NewDecoder(r.Body).Decode(&rawPatch); err != nil {
    WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
    return
}
if v, ok := rawPatch["enabled"]; ok {
    var enabled bool
    if err := json.Unmarshal(v, &enabled); err != nil {
        WriteJSONError(w, http.StatusBadRequest, "invalid enabled field")
        return
    }
    policy.Enabled = enabled
}
```

### Import state with composite ID (S3 rule, index-based like NFS)

```go
// Source: internal/provider/nfs_export_policy_rule_resource.go
func (r *s3ExportPolicyRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    parts := strings.SplitN(req.ID, "/", 2)
    if len(parts) != 2 { /* error */ }
    policyName := parts[0]
    index, err := strconv.Atoi(parts[1])
    // ... look up by index, populate state
}
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| SingleNestedBlock for computed nested objects | SingleNestedAttribute (types.Object) | v1.1 (Phase 6 fix) | Required for computed fields; planner must use this |
| Inline rule management via POST/PATCH body | Separate /rules endpoint resources | Architecture decision v1.0 | Never mix inline and dedicated rule resources |

---

## Model Structs to Add to models.go

The following structs belong in `internal/client/models.go`:

```go
// S3ExportPolicy represents a FlashBlade S3 export policy from GET responses.
type S3ExportPolicy struct {
    ID         string `json:"id,omitempty"`
    Name       string `json:"name"`
    Enabled    bool   `json:"enabled"`
    IsLocal    bool   `json:"is_local,omitempty"`
    PolicyType string `json:"policy_type,omitempty"`
    Version    string `json:"version,omitempty"`
}

type S3ExportPolicyPost struct {
    Enabled *bool `json:"enabled,omitempty"`
}

type S3ExportPolicyPatch struct {
    Name    *string `json:"name,omitempty"`
    Enabled *bool   `json:"enabled,omitempty"`
}

// S3ExportPolicyRule represents a rule from GET /s3-export-policies/rules.
type S3ExportPolicyRule struct {
    ID     string         `json:"id,omitempty"`
    Name   string         `json:"name,omitempty"`  // server-assigned
    Index  int            `json:"index"`
    Policy NamedReference `json:"policy,omitempty"`
    Effect string         `json:"effect,omitempty"`
    Actions   []string   `json:"actions,omitempty"`
    Resources []string   `json:"resources,omitempty"`
}

// S3ExportPolicyRulePost - rule name is NOT sent; policy is query param only.
type S3ExportPolicyRulePost struct {
    Effect    string   `json:"effect"`
    Actions   []string `json:"actions"`
    Resources []string `json:"resources"`
}

type S3ExportPolicyRulePatch struct {
    Effect    *string  `json:"effect,omitempty"`
    Actions   []string `json:"actions,omitempty"`
    Resources []string `json:"resources,omitempty"`
}

// ObjectStoreVirtualHost represents a FlashBlade virtual host from GET responses.
type ObjectStoreVirtualHost struct {
    ID              string           `json:"id,omitempty"`
    Name            string           `json:"name"`            // read-only, server-assigned
    Hostname        string           `json:"hostname,omitempty"`
    AttachedServers []NamedReference `json:"attached_servers,omitempty"`
}

// ObjectStoreVirtualHostPost - name is read-only; hostname + attached_servers are writable.
type ObjectStoreVirtualHostPost struct {
    Hostname        string           `json:"hostname"`
    AttachedServers []NamedReference `json:"attached_servers,omitempty"`
}

type ObjectStoreVirtualHostPatch struct {
    Name            *string          `json:"name,omitempty"`
    Hostname        *string          `json:"hostname,omitempty"`
    AttachedServers []NamedReference `json:"attached_servers,omitempty"` // full replace semantics
}
```

---

## Open Questions

1. **Virtual host POST query parameter**
   - What we know: POST body has `hostname` and `attached_servers`; API spec does not show an explicit query param note
   - What's unclear: Does POST use `?names=<hostname>`, `?names=<desired-name>`, or no query param at all?
   - Recommendation: Default to `?names=<hostname>` (consistent with all policy resources). If the API returns 400, try without the query param or with a different param. Validate against live API in Phase 8 acceptance tests.

2. **S3 export policy rule `index` field**
   - What we know: NFS rules have `index` in GET responses; S3 export policy rule API structure mirrors NFS (server-assigned name, separate /rules endpoint)
   - What's unclear: Whether S3 export policy rules expose an `index` field or use a different identifier for ordering/lookup
   - Recommendation: Include `index` in the model struct. If absent in API responses, use name-based lookup only and adjust import to use `policy_name/rule_name`.

3. **S3 export policy `version` field**
   - What we know: NFS export policies expose a `version` field; the S3 API PATCH body lists `rules`(array) but not `version`
   - What's unclear: Whether the S3 export policy GET response includes a `version` field for optimistic locking
   - Recommendation: Include `version` as computed in the model (with UseStateForUnknown). If absent in API, it will come back as empty string and not cause drift.

---

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go standard `testing` package |
| Config file | none (go test ./...) |
| Quick run command | `go test ./internal/provider/... -run TestUnit_ -v` |
| Full suite command | `go test ./...` |

### Phase Requirements to Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| S3P-01 | Create S3 export policy with enabled toggle | unit | `go test ./internal/provider/... -run TestUnit_S3ExportPolicy` | Wave 0 |
| S3P-02 | Create S3 export policy rule with actions/effect/resources | unit | `go test ./internal/provider/... -run TestUnit_S3ExportPolicyRule` | Wave 0 |
| S3P-03 | Update and delete S3 export policy rules independently | unit | `go test ./internal/provider/... -run TestUnit_S3ExportPolicyRule_Lifecycle` | Wave 0 |
| S3P-04 | Import S3 export policies and rules | unit | `go test ./internal/provider/... -run TestUnit_S3ExportPolicy.*Import` | Wave 0 |
| VH-01 | Create virtual host with hostname and attached servers | unit | `go test ./internal/provider/... -run TestUnit_ObjectStoreVirtualHost` | Wave 0 |
| VH-02 | Update attached servers list | unit | `go test ./internal/provider/... -run TestUnit_ObjectStoreVirtualHost_Update` | Wave 0 |
| VH-03 | Import existing virtual host | unit | `go test ./internal/provider/... -run TestUnit_ObjectStoreVirtualHost.*Import` | Wave 0 |

### Sampling Rate

- **Per task commit:** `go test ./internal/client/... ./internal/testmock/... -v`
- **Per wave merge:** `go test ./...`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps

- [ ] `internal/provider/s3_export_policy_resource_test.go` — covers S3P-01, S3P-04 (policy)
- [ ] `internal/provider/s3_export_policy_rule_resource_test.go` — covers S3P-02, S3P-03, S3P-04 (rules)
- [ ] `internal/provider/object_store_virtual_host_resource_test.go` — covers VH-01, VH-02, VH-03
- [ ] `internal/testmock/handlers/s3_export_policies.go` — mock handler for policies and rules
- [ ] `internal/testmock/handlers/object_store_virtual_hosts.go` — mock handler for virtual hosts
- [ ] `internal/client/s3_export_policies.go` — client CRUD methods
- [ ] `internal/client/object_store_virtual_hosts.go` — client CRUD methods

---

## Sources

### Primary (HIGH confidence)

- `FLASHBLADE_API.md` (local) — POST/PATCH/DELETE/GET specs for `/s3-export-policies`, `/s3-export-policies/rules`, `/object-store-virtual-hosts`
- `internal/provider/nfs_export_policy_resource.go` — reference pattern for policy resource (identical structure)
- `internal/provider/nfs_export_policy_rule_resource.go` — reference pattern for rule resource with server-assigned names
- `internal/provider/object_store_access_policy_rule_resource.go` — reference pattern for IAM-style rule resource
- `internal/provider/server_resource.go` — reference pattern for NamedReference list (attached_servers analog is DNS list)
- `internal/client/models.go` — existing model structs
- `internal/testmock/handlers/nfs_export_policies.go` — full mock handler reference pattern

### Secondary (MEDIUM confidence)

- `internal/testmock/handlers/object_store_access_policies.go` — OAP mock handler (user-supplied rule names pattern, contrasts with S3 server-assigned)

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — all libraries already in use, no new dependencies
- Architecture: HIGH — S3 export policy maps 1:1 to NFS export policy; S3 rule maps to hybrid NFS/OAP pattern; virtual host is new but straightforward
- Pitfalls: HIGH — derived from direct API spec inspection and cross-referencing with existing implementation decisions
- Open questions: MEDIUM — virtual host POST param and S3 rule index field need live API confirmation (addressed in Phase 8 acceptance tests)

**Research date:** 2026-03-28
**Valid until:** 2026-06-28 (stable API, 90 days)
