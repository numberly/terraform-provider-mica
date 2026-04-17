# Phase 50: Directory Service Roles & Role Mappings - Context

**Gathered:** 2026-04-17
**Status:** Ready for planning
**Mode:** Auto (gray areas resolved with recommended defaults)

<domain>
## Phase Boundary

Deliver two Terraform resources for LDAP → FlashBlade RBAC wiring:

1. **`flashblade_directory_service_role`** — a mapping from an LDAP group (CN + base DN) to one or more FlashBlade **management access policies**. Backed by `POST/GET/PATCH/DELETE /directory-services/roles`. Plus a read-only data source keyed by `name`.

2. **`flashblade_management_access_policy_directory_service_role_membership`** — an additive association linking an existing role mapping to an additional management access policy. Backed by `GET/POST/DELETE /management-access-policies/directory-services/roles`. Composite ID import.

Out of scope (see REQUIREMENTS.md §"Out of Scope"):
- `/directory-services/test` ephemeral endpoint — future milestone
- Active Directory accounts (`/active-directory`) — separate endpoint family
- NFS directory service variant
- Management access policy resource (`/management-access-policies`) — built-in policies only
- SMB sub-object (DEPRECATED in v2.22)

</domain>

<decisions>
## Implementation Decisions

### Carried Forward (from v2.22.1 + STATE.md + CONVENTIONS.md)
- **D-00-a:** Three-tier tests (unit/mocked/acceptance); `TestUnit_` prefix mandatory.
- **D-00-b:** All 4 interface assertions on every resource (`Resource`, `WithConfigure`, `WithImportState`, `WithUpgradeState`) with empty `UpgradeState` map at schema version 0 — **even though existing `_member` resources skip `WithUpgradeState`**, new resources MUST comply with CONVENTIONS.md.
- **D-00-c:** Drift detection via `tflog.Debug` on every mutable/computed field with `{resource, field, was, now}` shape.
- **D-00-d:** `**NamedReference` Patch pointer semantics: outer nil = omit, outer non-nil + inner nil = set null, both non-nil = set value.
- **D-00-e:** Commits use `--no-verify`; no `Co-Authored-By` trailers.

### API Reality Check (critical for planner)

Swagger inspection revealed deviations from the REQUIREMENTS.md assumptions — the planner MUST honor these:

- **The `role` attribute on `DirectoryServiceRole` is DEPRECATED per swagger.** Quote: _"role is deprecated in favor of management_access_policies, but remains for backwards compatibility. If a directory service role has exactly one access policy, which corresponds to a valid legacy role of the same name, role will be a reference to that role. Otherwise, it will be null."_
  - Consequence: the resource MUST use `management_access_policies` as the user-facing input (not the deprecated `role`).
- **`name` is read-only on `DirectoryServiceRole`** — server-generated, never user-supplied on POST. REQ DSR-01 wording implied user-named role; corrected below.
- **`management_access_policies` is writable on POST but READONLY on PATCH.** Consequence: initial list is set at creation; post-creation mutations must go through the separate `/management-access-policies/directory-services/roles` endpoint (i.e. the DSRM membership resource).
- **POST body takes no `names` query param.** Name derivation mechanism needs verification in mock/acceptance testing — likely server-generated from the associated policy when there is exactly one, otherwise synthetic.

### Resource Schema: `flashblade_directory_service_role`

- **D-01:** User-facing attributes:
  - `group` — Required string, Mutable via PATCH (CN of the LDAP group).
  - `group_base` — Required string, Mutable via PATCH (DN search base).
  - `management_access_policies` — Required list of strings (names of built-in policies like `pure:policy/array_admin`). Writable on POST only; **list change triggers `RequiresReplace()`** because PATCH rejects it. HCL shape: flat list of strings (internally mapped to `[]NamedReference` for the POST body).
  - `name` — Computed string (`UseStateForUnknown()` — server-generated, stable after creation).
  - `id` — Computed string (`UseStateForUnknown()`).
  - `role` — Computed-only nested object `{ name: string }` — API-populated legacy backfill, may be `null` when the role has multiple policies or non-legacy names.
- **D-02:** The **deprecated `role` attribute is NOT exposed as a writable input**. It surfaces as a computed nested object so operators can see the legacy backfill reported by the API. Drift on `role.name` is logged via `tflog.Debug` but never diff'd (it's derived from `management_access_policies`).
- **D-03:** `management_access_policies` in the **POST body** = `[]*_referenceWritable` (array of `{name: string}`). In the **Terraform schema** = `types.ListAttribute{ElementType: types.StringType}` for flat UX. Conversion helpers live inline in the resource file.

### Resource Schema: `flashblade_management_access_policy_directory_service_role_membership`

- **D-04:** User-facing attributes:
  - `policy` — Required string, `RequiresReplace()` (name of the management access policy).
  - `role` — Required string, `RequiresReplace()` (server-generated name of the directory service role, obtained from the DSR resource's `name` output or a data source lookup).
  - `id` — Computed composite string.
  - Follows the existing `qos_policy_member_resource.go` / `tls_policy_member_resource.go` / `certificate_group_member_resource.go` pattern: flat string attributes, all RequiresReplace, no PATCH semantics.

### Composite ID Format (DSRM)

- **D-05:** Composite ID uses existing `compositeID(parts...)` helper (`strings.Join(parts, "/")`) with format **`role_name/policy_name`**.
  - **CORRECTION from REQUIREMENTS.md DSRM-03** which suggested `policy_name:role_name`. Reason: built-in policy names contain BOTH `:` and `/` (e.g. `pure:policy/array_admin`). Using `:` or `/` as separator with `policy_name` first breaks `strings.SplitN(id, sep, 2)`. Putting **role_name first** (simple identifier, no special chars) lets `SplitN("role/pure:policy/array_admin", "/", 2)` correctly return `["role", "pure:policy/array_admin"]`.
  - Import error message when malformed: `"expected 2 parts separated by '/', got N in 'ID'"` — matches existing `parseCompositeID` behavior.

### Delete Semantics

- **D-06:** DSR Delete: straightforward `DELETE /directory-services/roles?names=<name>`. No soft-delete, no PATCH reset. Standard idempotent 404-tolerant delete.
- **D-07:** DSRM Delete: `DELETE /management-access-policies/directory-services/roles?policy_names=<p>&role_names=<r>`. Dissociates only — leaves both policy and role untouched. Tested by post-delete `terraform plan` on the role showing 0 diff.

### Read Semantics — Missing Association Handling

- **D-08:** DSRM Read: calls `GET /management-access-policies/directory-services/roles?policy_names=<p>&role_names=<r>`. If the response `items` list is empty (association removed outside Terraform), call `resp.State.RemoveResource(ctx)` — standard "resource gone" handling.
- **D-09:** DSR Read: calls `GET /directory-services/roles?names=<name>`. If not found, `RemoveResource(ctx)`. If found, drift check on `group`, `group_base`, `management_access_policies` (list), `role.name` (computed).

### Mock Handler Strategy

- **D-10:** Two new mock handlers:
  1. `internal/testmock/handlers/directory_service_roles.go` — GET/POST/PATCH/DELETE for `/api/2.22/directory-services/roles`. Store: `byName map[string]*client.DirectoryServiceRole` with synthetic ID generator. POST auto-assigns `name` derived from the first policy in `management_access_policies` (e.g. `pure:policy/array_admin` → role name `array_admin`) OR from the deprecated `role.name` if provided OR a sequential fallback `role-<N>`.
  2. `internal/testmock/handlers/management_access_policy_directory_service_role_memberships.go` — GET/POST/DELETE for `/api/2.22/management-access-policies/directory-services/roles`. Store: simple `set map[string]struct{}` keyed by `policy+"|"+role` pairs (**simple set, no cross-store validation** — tests seed independently).
- **D-11:** Mock GET returns empty list + 200 on filter miss (matches real API behavior; `getOneByName[T]` detects not-found).

### Plan Modifiers Summary

| Field | Modifier | Reason |
|-------|----------|--------|
| DSR.id | `UseStateForUnknown()` | Stable computed |
| DSR.name | `UseStateForUnknown()` | Server-generated, stable after create |
| DSR.group | None | Mutable via PATCH; drift detection only |
| DSR.group_base | None | Mutable via PATCH; drift detection only |
| DSR.management_access_policies | `RequiresReplace()` | Readonly on PATCH — cannot mutate, must recreate |
| DSR.role (computed obj) | None | API-derived, may drift |
| DSRM.policy | `RequiresReplace()` | Standard member pattern |
| DSRM.role | `RequiresReplace()` | Standard member pattern |
| DSRM.id | `UseStateForUnknown()` | Stable composite |

### Data Source — `flashblade_directory_service_role`

- **D-12:** Schema: `name` Required (lookup key), all other attributes Computed. No `timeouts`, no plan modifiers, no bind_password-equivalent. `management_access_policies` exposed as computed list of strings. `role` exposed as nested object with computed `name`. Only 2 interface assertions (`DataSource`, `DataSourceWithConfigure`).

### Open Questions for Planner / Researcher

These are not blockers — the planner should verify or pick the safer path:

- **Q1:** Exact behavior of server-generated `name` on POST `/directory-services/roles`. Does the server always derive from the first policy? Is there a `names` query param hidden from the api_references summary? Verify against the swagger POST body + response.
- **Q2:** Whether an empty `management_access_policies` list is a valid POST (edge case: user wants to create a "bare" role mapping then attach policies via DSRM only). If rejected, the resource must enforce at least one element via schema validator.
- **Q3:** Whether the DSRM POST is idempotent when the association already exists (409 vs 200). Needed to decide Read-before-Create or catch-409 strategy in the resource.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project conventions
- `CLAUDE.md` — Provider architecture, workflow, do/don't list, patterns (Serena MCP mandatory, `--no-verify` commits)
- `CONVENTIONS.md` — Full coding conventions; baseline **798 tests** → target ≥ 812

### Milestone context
- `.planning/PROJECT.md` — v2.22.2 milestone scope
- `.planning/REQUIREMENTS.md` — 22 REQ-IDs (DSR-01..06, DSRM-01..05, DOC-01..03, QA-01..08); **note: DSR-01 and DSRM-03 wording needs reconciliation against D-02 and D-05 above**
- `.planning/STATE.md` — accumulated decisions including Phase 49 NamedReference learnings
- `.planning/ROADMAP.md` §"Phase 50" — goal, success criteria, depends-on

### API reference
- `api_references/2.22.md` lines 432-435 — `/directory-services/roles` endpoints
- `api_references/2.22.md` lines 720-722 — `/directory-services/roles/management-access-policies` endpoints
- `api_references/2.22.md` lines 730-732 — `/management-access-policies/directory-services/roles` endpoints (same associations, inverse URL)
- `api_references/2.22.md` lines 1250-1251 — `DirectoryServiceRole` + `DirectoryServiceRolePost` schemas
- `swagger-2.22.json` §`DirectoryServiceRole`, `DirectoryServiceRolePost`, `_referenceWritable`, `_fixedReference`

### Pattern references — member resources (EXACTLY what to mirror)
- `internal/provider/qos_policy_member_resource.go` — simplest composite-ID member (3 interface assertions only; new resources bump to 4 per D-00-b)
- `internal/provider/qos_policy_member_resource_test.go` — test pattern
- `internal/provider/tls_policy_member_resource.go` — composite ID + NamedReference field
- `internal/provider/certificate_group_member_resource.go` — member with UpgradeState already declared (closest to v2.22.1 convention)
- `internal/provider/object_store_user_policy_resource.go` — user-policy membership (semantic analog: subject-policy link)
- `internal/provider/helpers.go` §`compositeID` lines 34-48 — the exact helper to use

### Pattern references — full-CRUD resource
- `internal/provider/target_resource.go` — full CRUD with `**NamedReference` for `ca_certificate_group` (pattern for `role` deprecated attribute marshalling)
- `internal/provider/array_dns_resource.go` — singleton pattern, NOT relevant here (DSR is multi-named)
- `internal/provider/certificate_resource.go` — resource with POST containing arrays

### Phase 49 deliverables worth re-reading
- `internal/client/directory_service.go` — the `patchOne[T,R]` / `getOneByName[T]` usage for directory-services-family endpoints
- `internal/client/models_admin.go` — `DirectoryService*` structs; add new `DirectoryServiceRole*` alongside
- `internal/testmock/handlers/directory_services.go` — closest mock handler to copy from

### Skills
- `.claude/skills/flashblade-resource-builder/SKILL.md` — mandatory checklist (models, client, mock, resource, tests, data source, examples, docs, ROADMAP update, test count ≥ baseline)

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `getOneByName[T]` — GET by name with empty-list-handled not-found detection
- `patchOne[T,R]` — typed PATCH helper (see `internal/client/directory_service.go` for reference usage)
- `compositeID(parts ...string)` / `parseCompositeID(id, n)` — composite import ID helpers (separator `/`)
- `nullTimeoutsValue()` — ImportState timeout initialiser
- `stringOrNull(s string)` — empty-string → null mapping
- Mock helpers: `ValidateQueryParams`, `RequireQueryParam`, `WriteJSONListResponse`, `WriteJSONError`
- Existing member resources follow a disciplined pattern — copy structure verbatim for both new resources

### Established Patterns
- **Model structs**: `DirectoryServiceRole` (GET), `DirectoryServiceRolePost` (POST), `DirectoryServiceRolePatch` (all pointer fields per `**NamedReference` discipline). Plus membership struct and association list response.
- **Flat schema attributes for name references**: `target_resource.go` maps `ca_certificate_group = "corp-ca"` (flat string in HCL) to `CACertificateGroup **NamedReference` (pointer in body). Replicate for `management_access_policies` list and DSRM `policy`/`role`.
- **Mock handler append pattern**: both new handlers register in `internal/testmock/mockserver.go` next to existing `RegisterDirectoryServicesHandlers`.
- **Drift detection shape**: `tflog.Debug(ctx, "drift detected", map[string]any{"resource": name, "field": "group", "was": oldVal, "now": newVal})` — copy verbatim from `array_dns_resource.go` / `directory_service_management_resource.go`.

### Integration Points
- `internal/provider/provider.go` `Resources()` — append `NewDirectoryServiceRoleResource` and `NewManagementAccessPolicyDirectoryServiceRoleMembershipResource`
- `internal/provider/provider.go` `DataSources()` — append `NewDirectoryServiceRoleDataSource`
- `internal/testmock/mockserver.go` — register the two new handler sets alongside existing `RegisterDirectoryServicesHandlers` (from Phase 49)
- `CONVENTIONS.md` §"Test Coverage" — bump baseline from 798 → 812 in the same commit as the final wave
- `ROADMAP.md` (root-level API coverage doc) — move two entries from "Not Implemented" to "Implemented → Array Administration", refresh counters, bump `Last updated` to shipping date

</code_context>

<specifics>
## Specific Ideas

- Example HCL should demonstrate the recommended wiring: create a role with the initial policy, then use one membership resource for each **additional** policy:
  ```hcl
  resource "flashblade_directory_service_role" "admins" {
    group                      = "cn=fb-admins,ou=groups,dc=corp"
    group_base                 = "ou=groups,dc=corp"
    management_access_policies = ["pure:policy/array_admin"]
  }

  resource "flashblade_management_access_policy_directory_service_role_membership" "admins_storage" {
    policy = "pure:policy/storage_admin"
    role   = flashblade_directory_service_role.admins.name
  }
  ```
- Import docs should make the composite-ID format explicit: `terraform import flashblade_management_access_policy_directory_service_role_membership.admins_storage array_admin/pure:policy/storage_admin` (note the FORWARD SLASH separator after role name).
- Drift messages should use HCL attribute names (snake_case), not Go field names.

</specifics>

<deferred>
## Deferred Ideas

- **Ephemeral `/directory-services/test` endpoint** — dry-run LDAP bind validation for CI pre-apply checks. Future milestone.
- **Management access policy data source** — useful to reference built-in policies by name with validation, not in scope.
- **Bulk role import via `for_each` from LDAP** — user-facing helper; not a provider concern.
- **Active Directory account family** — separate endpoint hierarchy.
- **NFS directory service variant** — separate milestone.

### Reviewed Todos (not folded)
_None — `todo match-phase 50` not run; no matches expected._

</deferred>

---

*Phase: 50-directory-service-roles-role-mappings*
*Context gathered: 2026-04-17 (auto mode)*
