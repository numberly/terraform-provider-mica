# Phase 8: SMB Client Policies, Syslog & Acceptance Tests - Research

**Researched:** 2026-03-28
**Domain:** Terraform provider Go — SMB client policy CRUD, syslog server CRUD, live acceptance tests
**Confidence:** HIGH

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| SMC-01 | Operator can create an SMB client policy with enable toggle | API: POST /smb-client-policies — same pattern as SMB share policy. `version` field key differentiator. |
| SMC-02 | Operator can create SMB client policy rules with client/encryption/permission | API: POST /smb-client-policies/rules with `client`, `encryption`, `permission`, optional `index`, optional `before_rule_id`. |
| SMC-03 | Operator can update and delete SMB client policy rules independently | API: PATCH /smb-client-policies/rules with `versions` param for optimistic concurrency. Rule addressed by `names` + `policy_names`. |
| SMC-04 | Operator can import SMB client policies and rules into Terraform state | Import pattern: policy by name, rule by `policy_name/rule_name` composite ID (same as SMB share). |
| SYS-01 | Operator can create a syslog server with URI, services, and sources | API: POST /syslog-servers — `uri` (PROTOCOL://HOST:PORT), `services` (array), `sources` (array). Name via ?names= param. |
| SYS-02 | Operator can update syslog server configuration | API: PATCH /syslog-servers — all three fields patchable in-place. No rename observed (name is in URI already). |
| SYS-03 | Operator can import an existing syslog server into Terraform state | Import by name. Resource is NOT a singleton — multiple syslog servers allowed. |
| EXP-03 | All export resources pass acceptance tests against live FlashBlade | Live HCL tests using `srv-numberly-backup-pr`. Covers server/export/policy resources only. No blade admin. |
</phase_requirements>

---

## Summary

Phase 8 implements two new resource families — SMB client policies and syslog servers — then validates all v1.1 resources against a live FlashBlade array. All patterns are well-established in the codebase.

The **SMB client policy** is structurally identical to the existing `flashblade_smb_share_policy` resource with one important difference: it has a `version` field used for **optimistic concurrency on rule operations**. The PATCH /smb-client-policies/rules endpoint accepts an optional `versions` query parameter. Rule fields are `client`, `encryption`, `permission`, and `index` — simpler than SMB share rules (no principal/change/full_control). The `before_rule_id` / `before_rule_name` params allow ordering but are not required for basic Terraform use.

The **syslog server** resource is a standard named resource (not a singleton). It has three writable fields — `uri` (format: `PROTOCOL://HOST:PORT`), `services` (array, valid values `data-audit` and `management`), and `sources` (array of network interface names). The `name` field is user-specified, making it straightforward to address by name in PATCH/DELETE.

**Acceptance tests** (EXP-03) extend the existing `tmp/test-purestorage/` project with new `.tf` files covering the server, S3 export policy, virtual host, SMB client policy, and syslog server resources. The existing server `srv-numberly-backup-pr` is used for all server-dependent resources. Blade administration resources are not touched.

**Primary recommendation:** Follow the SMB share policy pattern verbatim for SMB client policy. For syslog, follow the server resource pattern (named, non-singleton, simple flat schema). Acceptance tests live in `tmp/test-purestorage/` as new `.tf` files alongside existing ones.

---

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| terraform-plugin-framework | (existing) | Schema, CRUD, import | Established across all 25 resources |
| terraform-plugin-framework-timeouts | (existing) | Timeout attributes | All resources use this |
| github.com/google/uuid | (existing) | Rule name generation in mocks | Already used in mock handlers |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| internal/testmock | (internal) | Mock HTTP server for unit tests | All resource tests |
| internal/testmock/handlers | (internal) | Per-resource handler registration | One file per resource family |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| versions param for rule PATCH | Ignore optimistic concurrency | API supports it; skip for simplicity in Terraform (no user-visible version tracking needed) |
| before_rule_id for ordering | Index field only | Terraform has no stable ordering mechanism; expose `index` in schema, let API assign |

**Installation:** No new dependencies required. All needed packages already in go.mod.

---

## Architecture Patterns

### Recommended Project Structure

New files to create, mirroring existing Phase 3 SMB share policy files:

```
internal/client/
  smb_client_policies.go        # CRUD client methods
  models.go                     # add SmbClientPolicy*, SyslogServer* structs

internal/testmock/handlers/
  smb_client_policies.go        # mock handler (mirrors smb_share_policies.go)
  syslog_servers.go             # mock handler

internal/provider/
  smb_client_policy_resource.go           # policy resource
  smb_client_policy_rule_resource.go      # rule resource
  smb_client_policy_data_source.go        # data source
  smb_client_policy_resource_test.go      # unit tests
  smb_client_policy_rule_resource_test.go
  syslog_server_resource.go               # syslog resource
  syslog_server_resource_test.go
  syslog_server_data_source.go

tmp/test-purestorage/
  test_smb_client_policies.tf   # acceptance test HCL
  test_syslog_servers.tf        # acceptance test HCL
  test_v11_server_exports.tf    # acceptance tests for server/S3 policy/virtual host
```

### Pattern 1: SMB Client Policy — SMB Share Policy Mirror

**What:** The SMB client policy resource is structurally identical to `flashblade_smb_share_policy` with added `version` (Computed, UseStateForUnknown) and `access_based_enumeration_enabled` (Optional bool) fields.
**When to use:** For `flashblade_smb_client_policy` resource.

Key schema differences vs SMB share policy:
- Add `version` — Computed, UseStateForUnknown (read from API after writes)
- Add `access_based_enumeration_enabled` — Optional bool with default false
- No `version` is needed in Terraform PATCH body (it is read-only per API spec: `version(ro string)`)

```go
// Source: FLASHBLADE_API.md — SmbClientPolicy model
type SmbClientPolicy struct {
    ID                           string                          `json:"id,omitempty"`
    Name                         string                          `json:"name"`
    Enabled                      bool                            `json:"enabled"`
    IsLocal                      bool                            `json:"is_local,omitempty"`
    PolicyType                   string                          `json:"policy_type,omitempty"`
    Version                      string                          `json:"version,omitempty"`
    AccessBasedEnumerationEnabled bool                           `json:"access_based_enumeration_enabled,omitempty"`
    Rules                        []SmbClientPolicyRuleInPolicy   `json:"rules,omitempty"`
}
```

### Pattern 2: SMB Client Policy Rule — Index-Based, Not Name-Lookup

**What:** SMB client policy rules have `client`, `encryption`, `permission`, and `index` fields. The rule `name` is server-assigned (read-only). Import uses `policy_name/rule_name` (same as SMB share). Rules are addressed by `names=` + `policy_names=` in PATCH/DELETE.

**Critical difference vs SMB share rules:** The rule `index` is an integer position within the policy. For Terraform, expose `index` as Computed (server-assigns, can change on reorder by other means). Do NOT expose `before_rule_id` / `before_rule_name` — these are ordering hints for API consumers, not stable Terraform attributes.

```go
// Source: FLASHBLADE_API.md — SmbClientPolicyRule model
type SmbClientPolicyRule struct {
    ID            string         `json:"id,omitempty"`
    Name          string         `json:"name,omitempty"`
    Index         int            `json:"index"`
    Policy        NamedReference `json:"policy,omitempty"`
    PolicyVersion string         `json:"policy_version,omitempty"`
    Client        string         `json:"client,omitempty"`
    Encryption    string         `json:"encryption,omitempty"`
    Permission    string         `json:"permission,omitempty"`
}

type SmbClientPolicyRulePost struct {
    Client     string `json:"client,omitempty"`
    Encryption string `json:"encryption,omitempty"`
    Permission string `json:"permission,omitempty"`
    Index      *int   `json:"index,omitempty"`
}

type SmbClientPolicyRulePatch struct {
    Client     *string `json:"client,omitempty"`
    Encryption *string `json:"encryption,omitempty"`
    Permission *string `json:"permission,omitempty"`
    Index      *int    `json:"index,omitempty"`
}
```

### Pattern 3: Syslog Server — Flat Named Resource

**What:** Simple named resource with `uri`, `services` (list of strings), `sources` (list of strings). Follows the server resource pattern for list attributes but is simpler (no nested objects).

```go
// Source: FLASHBLADE_API.md — SyslogServer model
type SyslogServer struct {
    ID       string   `json:"id,omitempty"`
    Name     string   `json:"name,omitempty"`
    URI      string   `json:"uri,omitempty"`
    Services []string `json:"services,omitempty"`
    Sources  []string `json:"sources,omitempty"`
}

type SyslogServerPost struct {
    URI      string   `json:"uri,omitempty"`
    Services []string `json:"services,omitempty"`
    Sources  []string `json:"sources,omitempty"`
}

type SyslogServerPatch struct {
    URI      *string   `json:"uri,omitempty"`
    Services *[]string `json:"services,omitempty"`
    Sources  *[]string `json:"sources,omitempty"`
}
```

Schema attributes:
- `id` — Computed, UseStateForUnknown
- `name` — Required, RequiresReplace (name is part of URI scheme — not renameable in FlashBlade)
- `uri` — Required (format: `PROTOCOL://HOST:PORT` e.g. `udp://syslog.example.com:514`)
- `services` — Optional, Computed, list of strings, `listdefault.StaticValue` with empty list to avoid null-vs-empty drift
- `sources` — Optional, Computed, list of strings, same treatment

### Pattern 4: Acceptance Tests in `tmp/test-purestorage/`

**What:** New `.tf` files added to the existing workspace. Each file is self-contained. Uses the existing `data "flashblade_server" "backup"` from `test_exports.tf` via data source reference.

**Scope** (per user instruction — no blade admin resources):
- `flashblade_server` (data source read of `srv-numberly-backup-pr`)
- `flashblade_s3_export_policy` + rule
- `flashblade_object_store_virtual_host`
- `flashblade_smb_client_policy` + rule
- `flashblade_syslog_server`

**Pattern for each acceptance test file:**
```hcl
# test_smb_client_policies.tf
resource "flashblade_smb_client_policy" "test" {
  name    = "test-gule-smb-client"
  enabled = true
}

resource "flashblade_smb_client_policy_rule" "allow_all" {
  policy_name = flashblade_smb_client_policy.test.name
  client      = "*"
  permission  = "rw"
  encryption  = "optional"
}
```

### Anti-Patterns to Avoid

- **Exposing `before_rule_id` in schema:** This is an API ordering hint, not a stable Terraform attribute. Index reordering by third parties would cause constant plan diffs.
- **Sending `version` in PATCH body for rules:** The `versions` query param for optimistic concurrency is optional — skip it. The API field `policy_version` in rule responses is read-only. Adding it to Terraform PATCH would cause 400 errors.
- **Null vs empty list drift on services/sources:** Use `listdefault.StaticValue` with empty list (same pattern as `virtual_host.attached_servers` from Phase 7).
- **Singleton assumption for syslog:** Unlike array DNS/NTP/SMTP, syslog servers are regular named resources with full POST/DELETE. Do NOT use the GET-first-then-PATCH-or-POST singleton pattern.
- **Member guard on SMB client policy delete:** The SMB share policy has a member guard (checks file systems using it). SMB client policy may have a similar guard — check if file systems reference client policies, or implement a best-effort delete (if 409 is returned, surface as clear error).

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Pagination | Custom loop | `c.get()` + continuation_token loop (already in all client files) | Existing pattern in every list endpoint |
| Mock in-memory state | Custom sync.Map | `sync.Mutex` + Go map pattern (see smb_share_policies.go handler) | Proven pattern, thread-safe |
| List attribute null-vs-empty | Manual nil check | `listdefault.StaticValue(types.ListValueMust(...))` | Phase 7 established this for virtual host |
| UUID rule names in mock | Custom generator | `uuid.New().String()[:8]` prefix (see smb_share_policies.go mock) | Consistent mock naming |

---

## Common Pitfalls

### Pitfall 1: SMB Client Policy `version` Field Is Read-Only in PATCH Body
**What goes wrong:** `version` appears in `SmbClientPolicyPatch` API spec but it is marked `ro` (read-only). Sending it in the PATCH body will be rejected by the API.
**Why it happens:** The API schema lists it but marks it `ro`. The `version` value is only used as a `versions` query parameter on rules endpoints, not in the policy body.
**How to avoid:** Do not include `version` in `SmbClientPolicyPatch` struct. Model it as Computed + UseStateForUnknown in the Terraform schema.
**Warning signs:** 400 errors from PATCH /smb-client-policies with `invalid field: version`.

### Pitfall 2: SMB Client Policy Rules Endpoint Uses `versions` Query Param (Optional)
**What goes wrong:** If `versions` is sent with wrong value, the API may return 409/412 (optimistic concurrency conflict).
**Why it happens:** FlashBlade uses the policy `version` hash to detect concurrent rule modifications.
**How to avoid:** Do not pass `versions` in rule PATCH/DELETE — it is optional. Without it, the API proceeds without optimistic locking. This is the simplest and safest approach for Terraform.
**Warning signs:** 409 Conflict on rule PATCH unexpectedly.

### Pitfall 3: Syslog URI Format Strictness
**What goes wrong:** FlashBlade validates the URI format strictly: `PROTOCOL://HOST:PORT` (e.g. `udp://syslog.example.com:514`). Missing protocol, missing port, or wrong format causes 400.
**Why it happens:** API validates URI on POST/PATCH.
**How to avoid:** Add a ValidateFunc or validator on the `uri` attribute. Document the expected format in the schema description. Consider a regex validator.
**Warning signs:** 400 errors with `invalid uri format` on POST.

### Pitfall 4: Syslog `services` Valid Values
**What goes wrong:** Only `data-audit` and `management` are valid values for `services`. Sending any other value returns 400.
**Why it happens:** The API enforces an enum.
**How to avoid:** Add a `stringvalidator.OneOf("data-audit", "management")` on each list element in the schema. Or document it clearly.
**Warning signs:** 400 errors on POST.

### Pitfall 5: Null-vs-Empty Drift on `services` and `sources`
**What goes wrong:** API may return empty array `[]` while Terraform state has `null`, causing perpetual plan diff.
**Why it happens:** Same problem as `attached_servers` on virtual host (Phase 7 decision).
**How to avoid:** Use `listdefault.StaticValue` with empty list. Existing precedent: `virtual_host.attached_servers`.
**Warning signs:** `apply -> plan` shows non-empty diff on services/sources even when no change was made.

### Pitfall 6: Acceptance Test Naming Collision
**What goes wrong:** If acceptance test resources use names that already exist on the live FlashBlade, apply fails with 409 Conflict.
**Why it happens:** Live array retains state between test runs.
**How to avoid:** Use unique prefixed names (`test-gule-smb-client-*`, `test-gule-syslog-*`). Destroy before re-running. Add `depends_on` where needed to enforce teardown order.
**Warning signs:** 409 on second test run without destroy.

### Pitfall 7: SMB Client Policy Delete Guard
**What goes wrong:** If a file system is attached to an SMB client policy and you try to delete the policy, the API returns an error.
**Why it happens:** Same concept as SMB share policy member guard (which checks `smb.share_policy.name`). SMB client policies may bind to file systems via `smb.client_policy.name` filter.
**How to avoid:** Implement a member guard using `/file-systems?filter=smb.client_policy.name='...'` before DELETE, mirroring `ListSmbSharePolicyMembers`. If no file systems use it in the acceptance test scenario, this is safe to skip initially and add if live testing reveals 409s.
**Warning signs:** 409 on DELETE /smb-client-policies.

---

## Code Examples

### SMB Client Policy Client Methods (verified against existing smb_share_policies.go pattern)

```go
// Source: internal/client/smb_share_policies.go — mirror this structure
func (c *FlashBladeClient) GetSmbClientPolicy(ctx context.Context, name string) (*SmbClientPolicy, error) {
    path := "/smb-client-policies?names=" + url.QueryEscape(name)
    var resp ListResponse[SmbClientPolicy]
    if err := c.get(ctx, path, &resp); err != nil {
        return nil, err
    }
    if len(resp.Items) == 0 {
        return nil, &APIError{StatusCode: 404, Message: fmt.Sprintf("SMB client policy %q not found", name)}
    }
    return &resp.Items[0], nil
}

// Rule POST — policy_names in query param, body has writable fields only
func (c *FlashBladeClient) PostSmbClientPolicyRule(ctx context.Context, policyName string, body SmbClientPolicyRulePost) (*SmbClientPolicyRule, error) {
    path := "/smb-client-policies/rules?policy_names=" + url.QueryEscape(policyName)
    var resp ListResponse[SmbClientPolicyRule]
    if err := c.post(ctx, path, body, &resp); err != nil {
        return nil, err
    }
    if len(resp.Items) == 0 {
        return nil, fmt.Errorf("PostSmbClientPolicyRule: empty response from server")
    }
    return &resp.Items[0], nil
}
```

### Syslog Server Client Methods

```go
// Source: FLASHBLADE_API.md — POST /syslog-servers, GET /syslog-servers?names=
func (c *FlashBladeClient) PostSyslogServer(ctx context.Context, name string, body SyslogServerPost) (*SyslogServer, error) {
    path := "/syslog-servers?names=" + url.QueryEscape(name)
    var resp ListResponse[SyslogServer]
    if err := c.post(ctx, path, body, &resp); err != nil {
        return nil, err
    }
    if len(resp.Items) == 0 {
        return nil, fmt.Errorf("PostSyslogServer: empty response from server")
    }
    return &resp.Items[0], nil
}
```

### Mock Handler Registration Pattern (from smb_share_policies.go)

```go
// Source: internal/testmock/handlers/smb_share_policies.go — RegisterSmbSharePolicyHandlers
func RegisterSmbClientPolicyHandlers(mux *http.ServeMux) *smbClientPolicyStore {
    store := &smbClientPolicyStore{
        policies: make(map[string]*client.SmbClientPolicy),
        rules:    make(map[string]map[string]*client.SmbClientPolicyRule),
    }
    mux.HandleFunc("/api/2.22/smb-client-policies", store.handlePolicy)
    mux.HandleFunc("/api/2.22/smb-client-policies/rules", store.handleRules)
    return store
}
```

### Provider Registration (add to provider.go Resources() and DataSources())

```go
// Resources() additions:
NewSmbClientPolicyResource,
NewSmbClientPolicyRuleResource,
NewSyslogServerResource,

// DataSources() additions:
NewSmbClientPolicyDataSource,
NewSyslogServerDataSource,
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| ad-hoc acceptance tests | Terraform workspace in `tmp/test-purestorage/` | Phase 6/7 | All acceptance tests co-located, rerunnable |
| Inline rules in policy | Separate rule resources | v1.0 Phase 3 | Enables independent rule lifecycle |
| Singleton detection via list | Named resource with POST/DELETE | All v1.1 non-admin resources | Straightforward CRUD, no special singleton pattern |

**Deprecated/outdated:**
- The `versions` query param for optimistic locking: not needed for Terraform use — skip it.

---

## Open Questions

1. **SMB Client Policy Member Guard**
   - What we know: SMB share policies check `smb.share_policy.name` filter on file systems. SMB client policies likely bind to file systems via `smb.client_policy.name`.
   - What's unclear: The exact filter field name (`smb.client_policy.name` vs something else) — the API spec does not surface this directly in the FLASHBLADE_API.md.
   - Recommendation: Implement `ListSmbClientPolicyMembers` using the filter `smb.client_policy.name='...'`. If the filter field is wrong, the live test will reveal a 409. Add a best-effort guard and fall back gracefully.

2. **Syslog server `sources` field semantics**
   - What we know: `sources` is described as "the network interfaces used for communication with the syslog server". It is an array of strings.
   - What's unclear: Whether it accepts interface names (e.g. `eth0`, `vip0`) or IP addresses. The API spec just says "network interfaces".
   - Recommendation: Model as `types.List` of `types.StringType`. Document that values should be interface names. Let live testing validate.

3. **Acceptance test for virtual host — server attachment**
   - What we know: `flashblade_object_store_virtual_host` requires `attached_servers`. The existing server `srv-numberly-backup-pr` exists.
   - What's unclear: Whether the live FlashBlade will allow creating a virtual host with that server, and whether there's already one.
   - Recommendation: Use a unique name `test-gule-vh` and destroy at end. Use `attached_servers = []` if server attachment is optional.

---

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing (stdlib) + terraform-plugin-framework |
| Config file | none — `go test ./...` from repo root |
| Quick run command | `go test ./internal/provider/ -run TestUnit_SmbClientPolicy -v` |
| Full suite command | `go test ./...` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| SMC-01 | Create/read/update/delete SMB client policy | unit | `go test ./internal/provider/ -run TestUnit_SmbClientPolicy -v` | ❌ Wave 0 |
| SMC-02 | Create SMB client policy rule with client/encryption/permission | unit | `go test ./internal/provider/ -run TestUnit_SmbClientPolicyRule -v` | ❌ Wave 0 |
| SMC-03 | Update and delete rules independently | unit | `go test ./internal/provider/ -run TestUnit_SmbClientPolicyRule -v` | ❌ Wave 0 |
| SMC-04 | Import policy and rule | unit | `go test ./internal/provider/ -run TestUnit_SmbClientPolicy.*Import -v` | ❌ Wave 0 |
| SYS-01 | Create syslog server with URI/services/sources | unit | `go test ./internal/provider/ -run TestUnit_SyslogServer -v` | ❌ Wave 0 |
| SYS-02 | Update syslog server | unit | `go test ./internal/provider/ -run TestUnit_SyslogServer -v` | ❌ Wave 0 |
| SYS-03 | Import syslog server | unit | `go test ./internal/provider/ -run TestUnit_SyslogServer.*Import -v` | ❌ Wave 0 |
| EXP-03 | All v1.1 resources pass live FlashBlade tests | manual/acceptance | `cd tmp/test-purestorage && tofu apply && tofu destroy` | ❌ Wave 0 (new .tf files) |

### Sampling Rate
- **Per task commit:** `go test ./internal/provider/ -run TestUnit_Smb -v && go test ./internal/provider/ -run TestUnit_Syslog -v`
- **Per wave merge:** `go test ./...`
- **Phase gate:** Full suite green + manual acceptance test run against live FlashBlade before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/provider/smb_client_policy_resource_test.go` — covers SMC-01, SMC-04
- [ ] `internal/provider/smb_client_policy_rule_resource_test.go` — covers SMC-02, SMC-03
- [ ] `internal/provider/syslog_server_resource_test.go` — covers SYS-01, SYS-02, SYS-03
- [ ] `tmp/test-purestorage/test_smb_client_policies.tf` — acceptance EXP-03
- [ ] `tmp/test-purestorage/test_syslog_servers.tf` — acceptance EXP-03
- [ ] `tmp/test-purestorage/test_v11_server_exports.tf` — acceptance EXP-03 (server, S3 export policy, virtual host)
- [ ] Model structs in `internal/client/models.go`: `SmbClientPolicy`, `SmbClientPolicyPost`, `SmbClientPolicyPatch`, `SmbClientPolicyRule`, `SmbClientPolicyRulePost`, `SmbClientPolicyRulePatch`, `SmbClientPolicyRuleInPolicy`, `SyslogServer`, `SyslogServerPost`, `SyslogServerPatch`
- [ ] `internal/client/smb_client_policies.go` — client methods
- [ ] `internal/client/syslog_servers.go` — client methods
- [ ] `internal/testmock/handlers/smb_client_policies.go` — mock handler
- [ ] `internal/testmock/handlers/syslog_servers.go` — mock handler

---

## Sources

### Primary (HIGH confidence)
- `FLASHBLADE_API.md` — SMB client policy endpoints (lines 610–619), syslog endpoints (lines 813–821), SmbClientPolicy* data models (lines 1267–1283), SyslogServer* data models (lines 1329–1337)
- `internal/client/smb_share_policies.go` — reference client pattern for smb-client-policies implementation
- `internal/testmock/handlers/smb_share_policies.go` — reference mock handler pattern
- `internal/provider/smb_share_policy_resource.go` — reference resource pattern
- `internal/provider/smb_share_policy_rule_resource.go` — reference rule resource pattern
- `internal/provider/server_resource.go` — reference for list attribute handling (services/sources pattern)
- `internal/provider/provider.go` — registration pattern, confirmed module path

### Secondary (MEDIUM confidence)
- `tmp/test-purestorage/test_exports.tf` — confirms acceptance test workspace structure and `srv-numberly-backup-pr` server name
- `tmp/test-purestorage/test_filesystems.tf` — confirms naming convention (`test-gule-*`) for acceptance tests
- `.planning/STATE.md` decisions table — confirms `listdefault.StaticValue` for null-vs-empty, `policy_names` query param pattern

### Tertiary (LOW confidence)
- Member guard filter field for SMB client policy (`smb.client_policy.name`) — inferred from SMB share policy pattern, not confirmed by API spec

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — all patterns are established in the codebase
- Architecture: HIGH — direct mirror of existing SMB share policy + server resource patterns
- API fields: HIGH — verified against FLASHBLADE_API.md data models
- Acceptance test scope: HIGH — explicitly confirmed by user instruction and existing test workspace
- Pitfalls: MEDIUM — most inferred from API behavior patterns; member guard filter field name is LOW

**Research date:** 2026-03-28
**Valid until:** 2026-04-28 (stable API, 30-day window)
