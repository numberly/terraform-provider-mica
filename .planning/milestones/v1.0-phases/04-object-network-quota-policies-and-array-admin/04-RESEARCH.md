# Phase 4: Object/Network/Quota Policies and Array Admin - Research

**Researched:** 2026-03-27
**Domain:** FlashBlade REST API v2.22 — policy families + singleton array admin resources
**Confidence:** HIGH (all findings derived from FLASHBLADE_API.md, existing codebase patterns)

## Summary

Phase 4 implements three policy families (Object Store Access, Network Access, and Quota) and three
singleton array admin resources (DNS, NTP, SMTP). The three policy families follow Phase 3's
parent/child resource pattern closely — the primary new design work is for the Object Store Access
Policy rules with their IAM-style `conditions` (JSON object) schema and the singleton lifecycle.

**Critical discovery:** Network Access Policy has **no POST or DELETE** endpoint at the policy
level — only PATCH. It is a system-managed singleton like DNS/NTP/SMTP and must be treated
identically: Create=GET+PATCH, Delete=PATCH to reset/disable. This is a change from the initial
assumption that it would follow the standard parent/child pattern.

**Critical discovery:** There is no `/quota-policies` endpoint in FlashBlade API v2.22. The
closest quota endpoints are `/quotas/groups` and `/quotas/users` (per-file-system user/group
quotas) and `default_group_quota`/`default_user_quota` fields on file systems. The
"Quota Policy" requirements (QTP-*/QTR-*) must map to **quotas/groups and quotas/users** — not
a dedicated policy type. Research is flagged LOW confidence for exact quota policy endpoint
mapping; the planner must decide whether QTP/QTR means file-system-scoped user/group quotas.

**Primary recommendation:** Follow NFS/SMB patterns exactly for OAP rules (index-based import
ID); treat Network Access Policy as a singleton (not parent/child); clarify quota policy scope
before planning QTP/QTR tasks.

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

#### Array Admin Singletons (DNS, NTP, SMTP)
- **Lifecycle**: Create = read current + PATCH config. Read = GET. Update = PATCH. Delete = PATCH back to empty/default values (reset).
- **Destroy behavior**: Reset to defaults — clear nameservers, NTP servers, SMTP relay. Explicit and clean lifecycle.
- **Import ID**: `default` — there's only one instance per array (`terraform import flashblade_array_dns.x default`)
- **Data sources**: Read-only data sources alongside resources for cross-provider references

#### SMTP / Alerts Config
- **Flat attributes**: relay_host, sender_domain as top-level strings — no nested blocks
- **Alert watchers in same resource**: `flashblade_array_smtp` manages relay config + alert recipients together (not separate resources)

#### Object Store Access Policy Rules (IAM-style)
- **Conditions**: JSON string — `conditions = jsonencode({...})` — flexible, like AWS IAM policy documents
- **Actions**: List of strings — `actions = ["s3:GetObject", "s3:PutObject"]`
- **Effect**: Simple string — `effect = "allow" | "deny"` — validated at plan time
- **Same parent/child pattern**: `flashblade_object_store_access_policy` + `flashblade_object_store_access_policy_rule`

#### Network Access Policy
- **Full API coverage** — all fields: client, interfaces, effect, description
- **Same attachment pattern**: String reference on both bucket and file system resources — already supported from Phase 1/2
- **Same parent/child pattern as Phase 3**

#### Quota Policy
- **quota_limit**: Integer input in bytes + computed `quota_limit_display` showing human-readable value (e.g., "1 TB")
- **Enforcement**: Boolean `enforced` attribute
- **Same parent/child pattern as Phase 3**

#### Carried Forward from Phase 3 (applies to all 3 policy families)
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

### Deferred Ideas (OUT OF SCOPE)
None — discussion stayed within phase scope
</user_constraints>

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| OAP-01 | User can create an object store access policy with name and rules | POST /object-store-access-policies?names= (name as query param) + `description` in body |
| OAP-02 | User can update object store access policy attributes | PATCH /object-store-access-policies?names= — only `rules` is writable (description is POST-only) |
| OAP-03 | User can destroy an object store access policy | DELETE /object-store-access-policies?names= |
| OAP-04 | User can import an existing object store access policy into Terraform state by name | ImportState by name, same pattern as NFS policy |
| OAP-05 | Data source returns object store access policy attributes by name or filter | GET /object-store-access-policies with ?names= or ?filter= |
| OAR-01 | User can create object store access policy rules (effect, action, resource, condition) | POST /object-store-access-policies/rules?policy_names=&names= — body: actions(array), conditions(object), effect(string), resources(array) |
| OAR-02 | User can update object store access policy rules | PATCH /object-store-access-policies/rules?policy_names=&names= |
| OAR-03 | User can destroy object store access policy rules | DELETE /object-store-access-policies/rules?policy_names=&names= |
| OAR-04 | User can import object store access policy rules using composite ID | Import ID: policy_name/rule_name (rules have server-assigned name in GET response) |
| NAP-01 | User can create a network access policy with name | CRITICAL: No POST endpoint — singleton policy. Create=GET+PATCH (same as DNS/NTP/SMTP) |
| NAP-02 | User can update network access policy attributes | PATCH /network-access-policies?names= — writable: enabled, name, rules |
| NAP-03 | User can destroy a network access policy | No DELETE endpoint — Delete=PATCH to reset (disable, clear rules) |
| NAP-04 | User can import an existing network access policy into Terraform state by name | Import ID: policy name (or "default" if singleton) |
| NAP-05 | Data source returns network access policy attributes by name or filter | GET /network-access-policies |
| NAR-01 | User can create network access policy rules (client, interfaces, effect) | POST /network-access-policies/rules?policy_names=&names= — body: client, effect, interfaces(array), index |
| NAR-02 | User can update network access policy rules | PATCH /network-access-policies/rules?policy_names=&names= |
| NAR-03 | User can destroy network access policy rules | DELETE /network-access-policies/rules?policy_names=&names= |
| NAR-04 | User can import network access policy rules using composite ID | Import ID: policy_name/rule_index (rules have `index` in GET response, no string name) |
| QTP-01 | User can create a quota policy with name | OPEN QUESTION: No /quota-policies endpoint exists. See quota research below. |
| QTP-02 | User can update quota policy attributes | Depends on endpoint resolution |
| QTP-03 | User can destroy a quota policy | Depends on endpoint resolution |
| QTP-04 | User can import an existing quota policy into Terraform state by name | Depends on endpoint resolution |
| QTP-05 | Data source returns quota policy attributes by name or filter | Depends on endpoint resolution |
| QTR-01 | User can create quota policy rules (quota_limit, enforced) | Likely POST /quotas/groups or /quotas/users per-filesystem |
| QTR-02 | User can update quota policy rules | PATCH /quotas/groups or /quotas/users |
| QTR-03 | User can destroy quota policy rules | DELETE /quotas/groups or /quotas/users |
| QTR-04 | User can import quota policy rules using composite ID | Likely filesystem_name/uid or filesystem_name/gid |
| ADM-01 | User can manage array DNS configuration (nameservers, domain, search) | GET+PATCH /dns — fields: domain, nameservers(array), services(array), sources(array) |
| ADM-02 | User can manage array NTP configuration (servers) | GET+PATCH /arrays — field: ntp_servers(array). No dedicated /ntp endpoint. |
| ADM-03 | User can manage array SMTP configuration (relay host, sender) | GET+PATCH /smtp-servers — fields: relay_host, sender_domain, encryption_mode. Alert watchers managed via /alert-watchers (separate resource in requirements, but CONTEXT says same resource) |
| ADM-04 | Data sources for DNS, NTP, SMTP read-only access | GET /dns, GET /arrays (ntp_servers field), GET /smtp-servers |
| ADM-05 | User can import existing array admin configuration into Terraform state | Import ID: "default" for all three |
</phase_requirements>

---

## Standard Stack

### Core (all confirmed from codebase)
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| terraform-plugin-framework | v1.x | Resource/data source skeleton | Already used across all Phase 1-3 resources |
| terraform-plugin-framework-timeouts | v0.x | Operation timeouts | Already used in all resources |
| terraform-plugin-log/tflog | v0.x | Structured logging | Project standard |
| encoding/json | stdlib | JSON marshaling for conditions field | stdlib, no extra dep |

### No new dependencies for Phase 4.

---

## Architecture Patterns

### Established Patterns (Phase 3 → Phase 4 reuse)

#### Pattern 1: Standard Policy Resource (OAP follows this)
**What:** Policy resource = CRUD with name as query param, enable/disable, delete guard
**When to use:** OAP (has POST + DELETE at policy level)
**Template:** `nfs_export_policy_resource.go`

Key details for OAP:
- POST body: `description`(string), `rules`(array) — name as `?names=` query param
- PATCH body: `rules`(array) only — description is NOT patchable (POST-only field, confirmed from API spec)
- DELETE: `?names=` query param
- Rename: name is in PATCH body — in-place rename supported (same as NFS)
- `exclude_rules=true` query param on GET is available (use for policy-level reads)

#### Pattern 2: Singleton Admin Resource (DNS, NTP, SMTP, NetworkAccessPolicy)
**What:** Resource represents a system singleton. Create=GET+PATCH, Delete=PATCH-to-reset
**When to use:** DNS (`/dns`), NTP (`/arrays` ntp_servers), SMTP (`/smtp-servers`), NetworkAccessPolicy
**Key implementation:**

```go
// Create: read current state, then PATCH desired config
func (r *arrayDnsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    // Step 1: GET current (verify exists / read ID)
    current, err := r.client.GetArrayDns(ctx)
    // Step 2: PATCH with desired config
    err = r.client.PatchArrayDns(ctx, patch)
    // Step 3: Read back and set state
}

// Delete: PATCH back to empty/defaults
func (r *arrayDnsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    err := r.client.PatchArrayDns(ctx, ArrayDnsPatch{
        Nameservers: &[]string{},  // empty slice = clear
        Domain:      strPtr(""),
    })
}

// ImportState: always ID = "default"
func (r *arrayDnsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    // req.ID is "default" — ignore it, just read current
    resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
```

#### Pattern 3: Rule Resource with Index-based Import (NetworkAccessPolicy rules, similar to NFS)
**What:** Rule has numeric `index`, import ID = `policy_name/index`
**When to use:** Network Access Policy rules (confirmed — rules have `index` integer, no string name in GET)
**Template:** `nfs_export_policy_rule_resource.go`

#### Pattern 4: Rule Resource with Name-based Import (OAP rules)
**What:** Rule has server-assigned `name`, import ID = `policy_name/rule_name`
**When to use:** OAP rules (GET response has `name` field, `effect` is `ro string` — confirmed from PolicyRuleObjectAccess model)
**Template:** `smb_share_policy_rule_resource.go`

Key OAP rule details:
- `effect` is READ-ONLY in GET response (`effect`(ro string)) — must be provided on POST, then treated as computed/RequiresReplace on update
- `conditions` is `object` (arbitrary JSON) — implement as `types.String` with `jsonencode` convention
- `actions` is `array` of strings
- `resources` is `array` of strings (ARN-like resource patterns)
- No `index` field — name-based identification

### Recommended New File Structure

```
internal/client/
├── object_store_access_policies.go    # GetOAP, PostOAP, PatchOAP, DeleteOAP, GetOAPRule, PostOAPRule, PatchOAPRule, DeleteOAPRule
├── network_access_policies.go         # GetNAP, PatchNAP (no Post/Delete), GetNAPRule, PostNAPRule, PatchNAPRule, DeleteNAPRule
├── array_admin.go                     # GetArrayDns, PatchArrayDns, GetArray, PatchArrayNtp, GetSmtpServer, PatchSmtpServer, GetAlertWatchers, PostAlertWatcher, PatchAlertWatcher, DeleteAlertWatcher
└── (quota resolution pending)

internal/provider/
├── object_store_access_policy_resource.go
├── object_store_access_policy_resource_test.go
├── object_store_access_policy_data_source.go
├── object_store_access_policy_rule_resource.go
├── object_store_access_policy_rule_resource_test.go
├── network_access_policy_resource.go      # SINGLETON — no parent, no POST/DELETE at policy level
├── network_access_policy_resource_test.go
├── network_access_policy_data_source.go
├── network_access_policy_rule_resource.go
├── network_access_policy_rule_resource_test.go
├── array_dns_resource.go
├── array_dns_resource_test.go
├── array_dns_data_source.go
├── array_ntp_resource.go
├── array_ntp_resource_test.go
├── array_ntp_data_source.go
├── array_smtp_resource.go              # includes alert_watchers nested list
├── array_smtp_resource_test.go
└── array_smtp_data_source.go
```

### Anti-Patterns to Avoid
- **Treating NetworkAccessPolicy as a createable resource**: It has no POST endpoint. Applying the NFS pattern would send POST to a 404 or 405.
- **Treating `effect` as mutable on OAP rules**: The API marks `effect` as `(ro string)` in GET response. It is set on POST and cannot be changed — RequiresReplace.
- **Sending `description` in PATCH for OAP**: The PATCH body only accepts `rules`(array). Description is POST-only.
- **Using `/quotas/groups` as "quota policies"**: These are per-filesystem user/group quota overrides, not a policy abstraction. See open question.

---

## API Endpoint Schemas (Confirmed)

### Object Store Access Policy

**POST** `/api/2.22/object-store-access-policies?names=<name>`
- Params: `enforce_action_restrictions`(boolean, optional)
- Body: `description`(string), `rules`(array of PolicyRuleObjectAccessBulkManage)

**PATCH** `/api/2.22/object-store-access-policies?names=<name>`
- Params: `enforce_action_restrictions`(boolean, optional)
- Body: `rules`(array) only

**DELETE** `/api/2.22/object-store-access-policies?names=<name>`

**GET** `/api/2.22/object-store-access-policies?names=<name>`
- Params: `exclude_rules`(boolean) — useful for policy-level reads

Response model (`ObjectStoreAccessPolicy`):
```
account(object)    — reference, ro
arn(ro string)     — Amazon Resource Name, computed
context(ro)
created(ro int)
description(ro string)
enabled(ro bool)
id(ro string)
is_local(ro bool)
location(object, ro)
name(string)       — writable via PATCH
policy_type(ro)
realms(ro array)
rules(array)       — PolicyRuleObjectAccess objects
updated(ro int)
```

### Object Store Access Policy Rule

**POST** `/api/2.22/object-store-access-policies/rules?policy_names=<policy>&names=<rulename>`
- Body (`PolicyRuleObjectAccessPost`): `actions`(array), `conditions`(object), `effect`(string), `resources`(array)

**PATCH** `/api/2.22/object-store-access-policies/rules?policy_names=<policy>&names=<rulename>`
- Body: same fields as POST

**DELETE** `/api/2.22/object-store-access-policies/rules?policy_names=<policy>&names=<rulename>`

Response model (`PolicyRuleObjectAccess`):
```
actions(array)      — list of strings e.g. "s3:GetObject"
conditions(object)  — arbitrary JSON object (IAM condition map)
context(ro)
effect(ro string)   — "allow" or "deny" — READ ONLY after creation
name(ro string)     — server-assigned rule name
policy(object)      — NamedReference to parent policy
resources(array)    — list of strings (ARN-like resource patterns)
```

**CRITICAL**: `effect` is `ro string` in GET — it is set at POST and is immutable. Terraform schema must use `RequiresReplace` on `effect`.

**Rule name**: The rule `name` is server-assigned (read-only). The `name` used in the POST `?names=` query parameter is what becomes the rule name — it IS the user-specified rule name. Import ID: `policy_name/rule_name`.

### Network Access Policy (SINGLETON — no POST/DELETE at policy level)

**GET** `/api/2.22/network-access-policies?names=<name>`
**PATCH** `/api/2.22/network-access-policies?names=<name>`
- Body: `enabled`(boolean), `name`(string), `rules`(array)
- `version`(ro), `id`(ro), `is_local`(ro), `policy_type`(ro), `realms`(ro)

No POST, No DELETE at policy level. Existing policies are system-provided.

**POST** `/api/2.22/network-access-policies/rules?policy_names=<policy>&names=<rulename>`
- Body: `client`(string), `effect`(string), `index`(integer), `interfaces`(array)
- `name`(ro) — server-assigned

**PATCH** `/api/2.22/network-access-policies/rules?policy_names=<policy>&names=<rulename>`
- Body: same as POST + `policy`(object)

**DELETE** `/api/2.22/network-access-policies/rules?policy_names=<policy>&names=<rulename>`

Response model (`NetworkAccessPolicyRule`):
```
client(string)         — IP/CIDR or "*"
effect(string)         — "allow" or "deny"
id(ro string)
index(integer)         — numeric position, used for import ID
interfaces(array)      — list of interface names (["nfs", "smb", "s3", ...])
name(ro string)        — server-assigned
policy(object)         — NamedReference
policy_version(ro)
```

Import ID for NAP rules: `policy_name/index` (same as NFS export policy rule pattern).

### DNS (Singleton)

**GET** `/api/2.22/dns`
**POST** `/api/2.22/dns` — creates a new DNS config (may be used on arrays with no default DNS)
**PATCH** `/api/2.22/dns`
- Body: `domain`(string), `name`(string), `nameservers`(array), `services`(array), `sources`(array)
**DELETE** `/api/2.22/dns`

Response model (`Dns`):
```
domain(string)       — domain suffix appended by array
id(ro string)
name(string)         — writable
nameservers(array)   — list of DNS server IPs
realms(ro array)
services(array)      — services using this DNS config
sources(array)       — network interfaces for DNS traffic
```

Delete = PATCH to clear nameservers and domain. Because POST exists, the singleton lifecycle (Create=PATCH) needs validation: on most arrays, a default DNS config already exists; use GET first, PATCH if exists, POST if not.

**Default/reset state**: `nameservers: []`, `domain: ""` (empty string)

### NTP (via /arrays — no dedicated endpoint)

**GET** `/api/2.22/arrays` — returns `ntp_servers`(array) among other array fields
**PATCH** `/api/2.22/arrays` — body: `ntp_servers`(array)

The `arrays` endpoint is the general array config endpoint. NTP resource wraps this by reading/writing only the `ntp_servers` field.

Response model (`Array` relevant fields):
```
ntp_servers(array)   — list of NTP server hostnames/IPs
name(string)         — array name
id(ro)
time_zone(string)    — also manageable here (may add to NTP resource)
```

**Default/reset state**: `ntp_servers: []` (empty array)

**Important**: The `flashblade_array_ntp` resource MUST NOT overwrite other Array fields on PATCH — use targeted `ntp_servers` only patch.

### SMTP (Singleton)

**GET** `/api/2.22/smtp-servers`
**PATCH** `/api/2.22/smtp-servers`
- Body: `encryption_mode`(string), `relay_host`(string), `sender_domain`(string)

Response model (`SmtpServer`):
```
encryption_mode(string)   — TLS mode e.g. "none", "tls", "starttls"
id(ro string)
name(ro string)
relay_host(string)        — SMTP relay server
sender_domain(string)     — domain appended to sender email
```

**Default/reset state**: `relay_host: ""`, `sender_domain: ""`, `encryption_mode: "none"`

### Alert Watchers (managed within `flashblade_array_smtp`)

**GET** `/api/2.22/alert-watchers`
**POST** `/api/2.22/alert-watchers`
- Body: `minimum_notification_severity`(string) — name passed as `?names=` (email address is the name)
**PATCH** `/api/2.22/alert-watchers`
- Body: `enabled`(boolean), `minimum_notification_severity`(string)
**DELETE** `/api/2.22/alert-watchers`

Response model (`AlertWatcher`):
```
enabled(boolean)                    — is email notification enabled
id(ro string)
minimum_notification_severity(string) — "info", "warning", "error", "critical"
name(ro string)                     — the email address (name IS the email)
```

**Decision implication**: `flashblade_array_smtp` will manage SMTP relay config + alert watchers as a composite resource. Alert watchers are identified by email address (name). The resource will use a `SetNestedAttribute` for watchers within the SMTP resource.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| JSON conditions encoding | Custom marshaler | `json.Unmarshal` + `types.String` in schema | `jsonencode()` in HCL handles user side, Go handles storage |
| Policy version conflict detection | Custom ETag logic | Pass `policy_version` in PATCH params | NAP rules accept `versions` param for optimistic concurrency |
| Human-readable quota display | Custom byte formatter | `fmt.Sprintf("%.0f %s", ...)` helper | Consistent with `quota_limit_display` computed field |

---

## Common Pitfalls

### Pitfall 1: NetworkAccessPolicy — Sending POST for Create
**What goes wrong:** Resource Create method calls `PostNetworkAccessPolicy()` which 404s — this endpoint does not exist.
**Why it happens:** Developer assumes all policies follow the standard Create/Delete CRUD lifecycle.
**How to avoid:** Implement Create as GET+PATCH. The policy already exists on the array (system-provided).
**Warning signs:** 404 or 405 on first `terraform apply`.

### Pitfall 2: OAP `effect` Field Treated as Mutable
**What goes wrong:** User changes `effect` from "allow" to "deny" in config — API returns error or silently ignores because `effect` is `ro string` in GET.
**Why it happens:** The PATCH body for OAP rules accepts `effect` (it's in the schema) but the GET response marks it read-only.
**How to avoid:** Add `RequiresReplace` plan modifier on `effect` in Terraform schema.
**Warning signs:** Apply succeeds but Read shows old value → perpetual drift.

### Pitfall 3: OAP PATCH Body Includes `description`
**What goes wrong:** PATCH includes `description` field — API rejects or ignores it.
**Why it happens:** Developer mirrors POST schema into PATCH model.
**How to avoid:** `ObjectStoreAccessPolicyPatch` struct must only have `Rules` field — no `Description`.

### Pitfall 4: NAP Rules PATCH Without `policy_version`
**What goes wrong:** Concurrent or sequential NAP rule updates fail with version conflict errors.
**Why it happens:** NAP rules use optimistic concurrency via `policy_version`. The PATCH endpoint accepts `versions` query param.
**How to avoid:** Read `policy_version` from policy GET, pass as `?versions=` query param in rule PATCH/DELETE.

### Pitfall 5: Array NTP PATCH Overwrites Other Array Config
**What goes wrong:** NTP resource PATCH sends the full Array object (including name, timezone, etc.) which causes unintended changes.
**Why it happens:** PATCH model shares the `Array` struct with other fields.
**How to avoid:** `ArrayNtpPatch` struct must ONLY contain `NtpServers *[]string`. Other Array fields must be absent (omitempty).

### Pitfall 6: DNS POST vs PATCH on Create
**What goes wrong:** Always calling PATCH on Create fails if DNS config doesn't exist yet (new array, no default DNS).
**Why it happens:** Assuming singleton means always PATCH.
**How to avoid:** GET first. If 404 → POST. If found → PATCH.

### Pitfall 7: Alert Watcher Name is Email Address
**What goes wrong:** Resource sets `name` as a display label — API uses the email address itself as the name (it's the unique identifier).
**Why it happens:** Other resources have `name` as a user-chosen label.
**How to avoid:** The `name` field for alert watchers IS the recipient email address. It is `ro` in GET but set via `?names=<email>` in POST.

### Pitfall 8: Quota Policy Endpoint Mismatch
**What goes wrong:** Implementation assumes a `/quota-policies` endpoint that doesn't exist.
**Why it happens:** Requirements use "Quota Policy" terminology but v2.22 API has no such endpoint.
**How to avoid:** See Open Questions — resolve before planning QTP/QTR tasks.

---

## Code Examples

### OAP Rule Model Structs
```go
// Source: FLASHBLADE_API.md PolicyRuleObjectAccess / PolicyRuleObjectAccessPost

// ObjectStoreAccessPolicy GET model
type ObjectStoreAccessPolicy struct {
    ID          string                         `json:"id,omitempty"`
    Name        string                         `json:"name"`
    Description string                         `json:"description,omitempty"`
    ARN         string                         `json:"arn,omitempty"`
    Enabled     bool                           `json:"enabled"`
    IsLocal     bool                           `json:"is_local,omitempty"`
    PolicyType  string                         `json:"policy_type,omitempty"`
    Rules       []ObjectStoreAccessPolicyRule  `json:"rules,omitempty"`
    Created     int64                          `json:"created,omitempty"`
    Updated     int64                          `json:"updated,omitempty"`
}

// ObjectStoreAccessPolicyPost — POST body only
type ObjectStoreAccessPolicyPost struct {
    Description string                              `json:"description,omitempty"`
    Rules       []ObjectStoreAccessPolicyRulePost   `json:"rules,omitempty"`
}

// ObjectStoreAccessPolicyPatch — PATCH body (description NOT included)
type ObjectStoreAccessPolicyPatch struct {
    Rules []ObjectStoreAccessPolicyRulePost `json:"rules,omitempty"`
}

// ObjectStoreAccessPolicyRule — GET response rule object
type ObjectStoreAccessPolicyRule struct {
    Name       string          `json:"name,omitempty"`       // server-assigned, ro
    Effect     string          `json:"effect,omitempty"`     // ro after creation
    Actions    []string        `json:"actions,omitempty"`
    Conditions json.RawMessage `json:"conditions,omitempty"` // arbitrary JSON object
    Resources  []string        `json:"resources,omitempty"`
    Policy     *NamedReference `json:"policy,omitempty"`
}

// ObjectStoreAccessPolicyRulePost — POST/PATCH rule body
type ObjectStoreAccessPolicyRulePost struct {
    Effect     string          `json:"effect,omitempty"`
    Actions    []string        `json:"actions,omitempty"`
    Conditions json.RawMessage `json:"conditions,omitempty"`
    Resources  []string        `json:"resources,omitempty"`
}
```

### Singleton Create Pattern (DNS example)
```go
// Source: Phase 4 design — singleton lifecycle
func (r *arrayDnsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var plan arrayDnsModel
    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

    // GET current to verify exists and get ID
    current, err := r.client.GetArrayDns(ctx)
    if err != nil && !client.IsNotFound(err) {
        resp.Diagnostics.AddError("Error reading DNS config", err.Error())
        return
    }

    patch := client.ArrayDnsPatch{}
    if !plan.Domain.IsNull() {
        patch.Domain = plan.Domain.ValueStringPointer()
    }
    if !plan.Nameservers.IsNull() {
        ns := make([]string, 0)
        resp.Diagnostics.Append(plan.Nameservers.ElementsAs(ctx, &ns, false)...)
        patch.Nameservers = &ns
    }

    if current == nil {
        // No existing config — POST
        _, err = r.client.PostArrayDns(ctx, client.ArrayDnsPost{...})
    } else {
        // Existing config — PATCH
        _, err = r.client.PatchArrayDns(ctx, patch)
    }
    // ... read back, set state
}
```

### Singleton Delete Pattern (reset to defaults)
```go
// Source: Phase 4 design
func (r *arrayDnsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    empty := []string{}
    emptyDomain := ""
    err := r.client.PatchArrayDns(ctx, client.ArrayDnsPatch{
        Nameservers: &empty,
        Domain:      &emptyDomain,
    })
    if err != nil {
        resp.Diagnostics.AddError("Error resetting DNS config", err.Error())
    }
    // Terraform removes from state automatically
}
```

### NAP Rule Index-based Import
```go
// Source: Phase 3 NFS rule pattern adapted for NAP
func (r *networkAccessPolicyRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    // ID format: "policy_name/index"
    parts := strings.SplitN(req.ID, "/", 2)
    if len(parts) != 2 {
        resp.Diagnostics.AddError("Invalid import ID", "Expected format: policy_name/index")
        return
    }
    index, err := strconv.Atoi(parts[1])
    if err != nil {
        resp.Diagnostics.AddError("Invalid index in import ID", err.Error())
        return
    }
    rule, err := r.client.GetNetworkAccessPolicyRuleByIndex(ctx, parts[0], index)
    // ... set state
}
```

### OAP conditions as JSON string in Terraform schema
```go
// Source: Phase 4 design — conditions as types.String with jsonencode convention
"conditions": schema.StringAttribute{
    Optional:    true,
    Description: "IAM-style conditions as a JSON string. Use jsonencode({...}) in HCL.",
    // No plan modifier — conditions can be updated in-place
},

// In readIntoState: marshal json.RawMessage back to string
if rule.Conditions != nil {
    plan.Conditions = types.StringValue(string(rule.Conditions))
} else {
    plan.Conditions = types.StringNull()
}
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Snapshot policy rules (PATCH add_rules/remove_rules) | NFS/SMB rules (dedicated rule endpoint) | Phase 3 | Phase 4 OAP and NAP rules have dedicated endpoints — use POST/PATCH/DELETE, not bulk inline |
| Per-resource IDs as import key | policy_name/rule_name or policy_name/index composite | Phase 3 | Phase 4 rules follow same composite import pattern |
| Policy rename RequiresReplace (snapshot) | Policy rename in-place PATCH (NFS, SMB) | Phase 3 | OAP: rename supported (name in GET model is writable); NAP: rename supported (name in PATCH body) |

---

## Open Questions

### OQ-1: Quota Policy Endpoint (CRITICAL — blocks QTP/QTR planning)
- **What we know**: No `/quota-policies` endpoint exists in FlashBlade API v2.22. The `/quotas/groups` and `/quotas/users` endpoints manage per-filesystem user/group space quotas (integer bytes, no `enforced` field). The `default_group_quota`/`default_user_quota` fields on file systems set defaults. There is no named quota policy abstraction.
- **What's unclear**: What the requirements authors intended by "quota policy" — possibly: (a) per-file-system quota settings (not a policy), (b) quotas/groups + quotas/users as the "rule" analogues, or (c) a feature that doesn't exist in v2.22.
- **Recommendation**: Reinterpret QTP/QTR as `flashblade_quota_user` and `flashblade_quota_group` resources targeting `/quotas/users` and `/quotas/groups` respectively. The "policy" framing is dropped; these are direct quota management resources scoped to a file system + user/group. The `enforced` boolean has no equivalent in the API (quotas are always enforced if set). The planner should confirm this reinterpretation before creating QTP tasks.

### OQ-2: OAP `effect` in PATCH — Can It Actually Change?
- **What we know**: GET response marks `effect` as `ro string`. PATCH body allows `effect`. The API schema reference `PolicyRuleObjectAccessPost` shows `effect`(string) as writable.
- **What's unclear**: Whether sending `effect` in PATCH to change it is accepted or silently ignored.
- **Recommendation**: Treat as `RequiresReplace` (conservative, safe). If the requirement to change effect in-place comes up, it can be relaxed. Better to force replace than produce silent drift.

### OQ-3: NAP Policy Import ID Format
- **What we know**: NetworkAccessPolicy has no POST — it exists as a system singleton. There may be only one, or there may be named policies.
- **What's unclear**: Whether NAP follows the same "there's always exactly one" constraint as DNS/NTP, or whether multiple named NAPs can exist on a FlashBlade.
- **Recommendation**: Treat as multi-instance (import by name). The singleton lifecycle (Create=GET+PATCH, Delete=reset) still applies, but the import ID is the policy name, not hardcoded "default".

---

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing + httptest.NewServer mock |
| Config file | none — standard `go test ./...` |
| Quick run command | `go test ./internal/... -run TestObjectStoreAccess -v -count=1` |
| Full suite command | `go test ./internal/... -count=1` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| OAP-01 | Create OAP via POST | unit | `go test ./internal/... -run TestObjectStoreAccessPolicyResource_Create -v` | ❌ Wave 0 |
| OAP-02 | Update OAP via PATCH | unit | `go test ./internal/... -run TestObjectStoreAccessPolicyResource_Update -v` | ❌ Wave 0 |
| OAP-03 | Delete OAP via DELETE | unit | `go test ./internal/... -run TestObjectStoreAccessPolicyResource_Delete -v` | ❌ Wave 0 |
| OAP-04 | Import OAP by name | unit | `go test ./internal/... -run TestObjectStoreAccessPolicyResource_Import -v` | ❌ Wave 0 |
| OAP-05 | Data source OAP | unit | `go test ./internal/... -run TestObjectStoreAccessPolicyDataSource -v` | ❌ Wave 0 |
| OAR-01 | Create OAP rule | unit | `go test ./internal/... -run TestObjectStoreAccessPolicyRuleResource_Create -v` | ❌ Wave 0 |
| OAR-02 | Update OAP rule | unit | included above | ❌ Wave 0 |
| OAR-03 | Delete OAP rule | unit | included above | ❌ Wave 0 |
| OAR-04 | Import OAP rule by name | unit | included above | ❌ Wave 0 |
| NAP-01 | Create NAP (GET+PATCH singleton) | unit | `go test ./internal/... -run TestNetworkAccessPolicyResource_Create -v` | ❌ Wave 0 |
| NAP-02 | Update NAP | unit | included above | ❌ Wave 0 |
| NAP-03 | Delete NAP (reset) | unit | included above | ❌ Wave 0 |
| NAP-04 | Import NAP by name | unit | included above | ❌ Wave 0 |
| NAP-05 | Data source NAP | unit | `go test ./internal/... -run TestNetworkAccessPolicyDataSource -v` | ❌ Wave 0 |
| NAR-01 | Create NAP rule | unit | `go test ./internal/... -run TestNetworkAccessPolicyRuleResource_Create -v` | ❌ Wave 0 |
| NAR-02..04 | Update/Delete/Import NAP rule | unit | included above | ❌ Wave 0 |
| ADM-01 | DNS resource CRUD + reset | unit | `go test ./internal/... -run TestArrayDnsResource -v` | ❌ Wave 0 |
| ADM-02 | NTP resource CRUD + reset | unit | `go test ./internal/... -run TestArrayNtpResource -v` | ❌ Wave 0 |
| ADM-03 | SMTP resource CRUD + alert watchers | unit | `go test ./internal/... -run TestArraySmtpResource -v` | ❌ Wave 0 |
| ADM-04 | DNS/NTP/SMTP data sources | unit | `go test ./internal/... -run TestArray.*DataSource -v` | ❌ Wave 0 |
| ADM-05 | Import DNS/NTP/SMTP with id=default | unit | included above | ❌ Wave 0 |

QTP/QTR tests deferred pending OQ-1 resolution.

### Sampling Rate
- **Per task commit:** `go test ./internal/... -count=1` (full suite, ~101 tests currently)
- **Per wave merge:** `go test ./internal/... -count=1 -race`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/provider/object_store_access_policy_resource_test.go` — OAP-01..05, OAR-01..04
- [ ] `internal/provider/network_access_policy_resource_test.go` — NAP-01..05, NAR-01..04
- [ ] `internal/provider/array_dns_resource_test.go` — ADM-01, ADM-04, ADM-05
- [ ] `internal/provider/array_ntp_resource_test.go` — ADM-02, ADM-04, ADM-05
- [ ] `internal/provider/array_smtp_resource_test.go` — ADM-03, ADM-04, ADM-05
- [ ] Mock handlers in `provider_test.go` for all new endpoints

---

## Sources

### Primary (HIGH confidence)
- `FLASHBLADE_API.md` — Full API v2.22 reference. All endpoint schemas, request/response models confirmed here:
  - Lines 564-580: Object Store Access Policies endpoints + rule endpoints
  - Lines 554-562: Network Access Policies endpoints (confirmed no POST/DELETE)
  - Lines 213-218: DNS endpoints
  - Lines 103-104: Arrays PATCH (NTP via `ntp_servers`)
  - Lines 759-762: SMTP servers endpoint
  - Lines 79-83: Alert watchers endpoint
  - Lines 1163, 1215-1219: ObjectStoreAccessPolicy and PolicyRuleObjectAccess models
  - Lines 1105-1113: NetworkAccessPolicy and NetworkAccessPolicyRule models
  - Lines 1005: Dns model
  - Lines 863: AlertWatcher model
  - Lines 1297: SmtpServer model
  - Lines 705-716: Quotas (groups/users) endpoints — no quota-policies endpoint
- `internal/client/models.go` — All existing Phase 1-3 model structs (structural reference)
- `internal/provider/nfs_export_policy_resource.go` — Phase 3 policy resource pattern
- `internal/provider/nfs_export_policy_rule_resource.go` — Phase 3 index-based rule pattern
- `internal/provider/smb_share_policy_rule_resource.go` — Phase 3 name-based rule pattern
- `internal/provider/snapshot_policy_resource.go` — RequiresReplace name variant

### Secondary (MEDIUM confidence)
- Phase 3 accumulated decisions in `STATE.md` — confirmed rule import formats and API schema differences

### Tertiary (LOW confidence)
- Quota policy interpretation (OQ-1) — no official endpoint, reinterpretation based on available API endpoints only

---

## Metadata

**Confidence breakdown:**
- Object Store Access Policy: HIGH — endpoints, models, all fields confirmed from FLASHBLADE_API.md
- Network Access Policy: HIGH — singleton behavior confirmed (no POST/DELETE), models confirmed
- DNS: HIGH — endpoints and model confirmed
- NTP: HIGH — confirmed lives on /arrays endpoint, not a dedicated endpoint
- SMTP: HIGH — endpoint and model confirmed
- Alert Watchers: HIGH — endpoint and model confirmed (separate from SMTP at API level, combined in resource per CONTEXT decision)
- Quota Policy: LOW — no matching endpoint in API; reinterpretation required

**Research date:** 2026-03-27
**Valid until:** 2026-04-27 (stable API, 30-day window)
