# Phase 49: Directory Service Management - Context

**Gathered:** 2026-04-17
**Status:** Ready for planning

<domain>
## Phase Boundary

Deliver a Terraform resource `flashblade_directory_service_management` and matching
read-only data source that manage the FlashBlade LDAP **management** directory
service (admin authentication) through the singleton PATCH-only endpoint
`PATCH /directory-services?names=management`.

Scope includes: resource CRUD surface (Create/Read/Update/Delete semantics
mapped onto PATCH), data source, mock handler, client CRUD, unit tests,
HCL examples, generated docs, ROADMAP update.

Out of scope (see REQUIREMENTS.md):
- NFS directory service variant (`flashblade_directory_service_nfs`)
- Directory service roles / role mappings
- Active Directory accounts
- SMB sub-object (DEPRECATED in v2.22)
- `/directory-services/test` ephemeral endpoint
- Multi-service generic resource with name selector (rejected)

</domain>

<decisions>
## Implementation Decisions

### Carried Forward (from STATE.md + REQUIREMENTS.md)
- **D-00-a:** Singleton PATCH-only. No POST, no DELETE. Create and Update both
  call `PatchDirectoryServiceManagement`. Delete sends a reset PATCH.
- **D-00-b:** `bind_password` is sensitive + write-only. Never returned by
  API. Stored sensitive in state. Not surfaced in plan diff. Left empty on
  import.
- **D-00-c:** `ca_certificate` and `ca_certificate_group` use the
  `**NamedReference` pattern: outer nil = omit from PATCH, outer non-nil +
  inner nil = explicit null (clear reference), outer non-nil + inner non-nil
  = set to named reference.
- **D-00-d:** `management` sub-object holds exactly three fields in this
  phase: `user_login_attribute`, `user_object_class`,
  `ssh_public_key_attribute`.
- **D-00-e:** Import key is the literal string `management`. ImportState
  calls `nullTimeoutsValue()` and leaves `bind_password` empty.
- **D-00-f:** `SchemaVersion: 0` with an empty `UpgradeState` map.
- **D-00-g:** Drift detection via `tflog.Debug` on every mutable/computed
  field with `{resource, field, was, now}` shape (DSM-07).

### Resource Schema Shape
- **D-01:** The `name` attribute is **not exposed** in the resource schema.
  The resource is hardcoded to target `management` internally. HCL
  configurations do not (and cannot) set a name. The import literal is still
  `management` (e.g. `terraform import flashblade_directory_service_management.example management`).
  - Rationale: prevents user misuse, enforces the singleton-per-service model,
    keeps the NFS variant as a clean sibling resource rather than a selector.

### Delete Semantics
- **D-02:** Delete sends a **full-reset PATCH** with:
  - `enabled = false`
  - `uris = []`
  - `bind_user = ""`
  - `base_dn = ""`
  - `ca_certificate = { name: null }` (explicit clear via `**NamedReference`)
  - `ca_certificate_group = { name: null }` (explicit clear via `**NamedReference`)
  - `management = { user_login_attribute: "", user_object_class: "", ssh_public_key_attribute: "" }`
  - `bind_password` omitted (implicitly cleared since never re-sent)
  - Rationale: matches the "destroy" semantic contract — next apply starts
    from a clean slate; avoids lingering config leaking between recreates.

### Management Sub-Object Modelling
- **D-03:** `user_login_attribute`, `user_object_class`, and
  `ssh_public_key_attribute` are **Optional + Computed** string attributes.
  - Unset in HCL → do not send in PATCH → API populates defaults (e.g.
    `sAMAccountName` for AD, `uid` otherwise, `User`/`posixAccount`/`person`
    depending on server type).
  - Set in HCL → value wins.
  - Read compares last-known state vs API value; logs drift via `tflog.Debug`
    when the API has diverged from state.
  - No plan modifiers on these fields (computed-with-default pattern — let
    the framework show diffs when API drifts).

### URIs Validation
- **D-04:** `uris` is a required list of strings. A **list validator** rejects
  any entry not matching `^ldaps?://` at plan time. Error message:
  `"uris[N] must start with ldap:// or ldaps://"`.
  - Implementation pattern: add a custom validator under
    `internal/provider/validators/` alongside existing `Alphanumeric` and
    `HostnameNoDot` validators, or inline list validator if simple enough.

### Plan Modifiers
- **D-05:** Standard plan-modifier placement:
  - `id`, `created`, `services` (read-only array) → `UseStateForUnknown()`
    only when stable (`id` is stable; `services` changes based on enabled
    state so NO modifier).
  - No `name` attribute → no `RequiresReplace` needed.
  - `bind_password`: `UseStateForUnknown()` (write-once sensitive — state
    preserves value across plans).

### Data Source Shape
- **D-06:** Data source mirrors the resource schema but is **computed-only**:
  - No `name` attribute (always reads `management`).
  - No `bind_password` (never readable).
  - `ca_certificate` and `ca_certificate_group` exposed as nested objects with
    a `name` string attribute (matches resource shape; consistent with other
    data sources like `flashblade_target`).
  - `services` exposed as a computed list of strings.
  - No timeouts block (data sources don't use timeouts).

### Claude's Discretion
- Exact placement of the URI validator (`internal/provider/validators/`
  package vs inline) — choose consistent with current repo layout.
- Wording of tflog.Debug drift messages (follow `{resource, field, was, now}`
  shape from CONVENTIONS.md).
- Ordering of fields in schema definition (logical grouping: identity,
  connection, TLS, management sub-object, state).
- Exact test seed values (pick a realistic `ldaps://ldap.example.com:636`
  style to aid readability).

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project conventions
- `CLAUDE.md` — Provider architecture, workflow, do/don't list, patterns to follow
- `CONVENTIONS.md` — Full coding conventions (model structs, client CRUD, mock
  handlers, resource patterns, tests, state upgraders) — §"Model Structs",
  §"Resource Implementation", §"Data Source Implementation", §"Test Conventions"

### Milestone context
- `.planning/PROJECT.md` — Current milestone vision and constraints
- `.planning/REQUIREMENTS.md` — v2.22.1 requirements DSM-01 through QA-06 (all 17 items)
- `.planning/STATE.md` — Accumulated decisions from milestone kickoff
- `.planning/ROADMAP.md` §"Phase 49" — Goal, success criteria, depends-on chain

### API reference
- `api_references/2.22.md` lines 430-437 — directory-services endpoints summary
- `api_references/2.22.md` line 1249 — `DirectoryService` schema (all fields)
- `swagger-2.22.json` §`_directoryServiceManagement` (lines 39848-39868) — management sub-object schema with defaults and examples

### Existing patterns to mirror
- `internal/client/models_storage.go` — Target structs show Get/Post/Patch
  separation with `**NamedReference` for `ca_certificate_group`
- `internal/client/models_admin.go` line 140 — Patch pointer semantics doc
  (outer nil = omit, outer non-nil + inner nil = set null)
- `internal/provider/target_resource.go` — Reference resource with
  `ca_certificate_group` NamedReference handling (schema + Patch build)
- `internal/provider/array_dns_resource.go` — Closest singleton-style
  resource (PATCH-driven reset on delete; since Phase 42 named but the
  reset pattern still applies)
- `internal/provider/array_smtp_resource.go` — Singleton with sensitive
  password handling
- `internal/testmock/handlers/targets.go` — Mock handler for `**NamedReference`
  fields
- `internal/provider/validators/` — Existing custom validators
  (`Alphanumeric`, `HostnameNoDot`, CIDR) — pattern to follow for URI scheme
  validator

### API tooling (for schema introspection during planning)
- `PYTHONPATH=.claude/skills python3 .claude/skills/swagger-to-reference/scripts/browse_api.py api_references/2.22.md schema DirectoryService`
- `PYTHONPATH=.claude/skills python3 .claude/skills/swagger-to-reference/scripts/browse_api.py api_references/2.22.md schema _directoryServiceManagement` (if supported)

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `getOneByName[T]` (internal/client helpers) — use for the GET half of the
  client. The endpoint returns `{"items": [...]}` filtered by `?names=management`.
- `**NamedReference` Patch pattern (internal/client/models_admin.go:140) —
  ready to reuse verbatim for `ca_certificate` + `ca_certificate_group`.
- `nullTimeoutsValue()` helper — already used by every resource for
  ImportState timeout initialisation.
- `ValidateQueryParams`, `RequireQueryParam`, `WriteJSONListResponse`,
  `WriteJSONError` (internal/testmock/handlers/helpers.go) — mock handler
  scaffolding.
- Existing validators (Alphanumeric, HostnameNoDot, CIDR) in
  `internal/provider/validators/` — add URI scheme validator alongside.
- `testNewMockedProvider()` test harness — wires client → mock server URL.

### Established Patterns
- **Three model structs per resource** — `DirectoryService` (GET),
  `DirectoryServicePost` (unused here — no POST endpoint), and
  `DirectoryServicePatch` (all-pointer). Since there is no POST, the client
  exposes only `Get` and `Patch` methods.
- **Singleton + filter-by-name** — `PATCH /directory-services?names=management`
  mirrors `array_dns` / `array_smtp` / `array_ntp` endpoint shape. Mock
  handler paths use `/api/2.22/directory-services`.
- **Drift detection in Read** — every mutable field compares state vs API
  and logs with `tflog.Debug`.
- **Interface assertions** — resource must declare `Resource`,
  `WithConfigure`, `WithImportState`, `WithUpgradeState` (even with empty
  UpgradeState map at schema version 0).
- **Write-once sensitive fields** — `bind_password` follows the same
  contract as `remote_credentials.secret_access_key`,
  `access_key.secret_access_key`, `array_connection.connection_key`,
  `certificate.private_key`.

### Integration Points
- Register factory in `internal/provider/provider.go`:
  - `Resources()` slice → append `NewDirectoryServiceManagementResource`
  - `DataSources()` slice → append `NewDirectoryServiceManagementDataSource`
- Test baseline in CONVENTIONS.md must bump from 779 → ≥ 787 (+8 new tests
  per QA-04).
- ROADMAP.md entry must move from "Not Implemented → High Priority" to
  "Implemented → Array Administration" with counters refreshed (QA-06).
- HCL examples live at
  `examples/resources/flashblade_directory_service_management/` and
  `examples/data-sources/flashblade_directory_service_management/`.

</code_context>

<specifics>
## Specific Ideas

- The API-documented defaults (`sAMAccountName`, `uid`, `User`,
  `posixAccount`, `person`) are worth noting in the resource description /
  attribute docs so operators understand when/why the API populates values
  they didn't set.
- Example HCL should use `ldaps://` URIs to match enterprise expectations
  (TLS-validated LDAP).
- Drift messages should use field names matching HCL attribute names (snake
  case, e.g. `user_login_attribute`) not Go field names (CamelCase) so the
  logs are operator-friendly.

</specifics>

<deferred>
## Deferred Ideas

- **NFS directory service variant** — `flashblade_directory_service_nfs` as a
  separate resource (same endpoint, different nested sub-object: `nis_domains`,
  `nis_servers`). Future milestone.
- **Directory service roles / role mappings** — `/directory-services/roles`
  endpoint family, maps LDAP groups to FlashBlade roles. Future milestone.
- **Active Directory accounts** — `/active-directory` endpoint family
  (Kerberos joins used by NFS/SMB). Future milestone.
- **Directory service test endpoint** — `/directory-services/test` ephemeral
  validation action. Better suited to a future CLI helper, not a Terraform
  resource.
- **SMB sub-object surface** — Out of scope permanently because the `smb`
  sub-object is DEPRECATED in v2.22.

### Reviewed Todos (not folded)
_None — `todo match-phase 49` returned no matches._

</deferred>

---

*Phase: 49-directory-service-management*
*Context gathered: 2026-04-17*
