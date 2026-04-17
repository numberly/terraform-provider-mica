# Phase 50: Directory Service Roles & Role Mappings - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-17
**Phase:** 50-directory-service-roles-role-mappings
**Mode:** `--auto` — user approved auto-advance chain; all gray areas resolved with recommended defaults and swagger-informed corrections.

---

## name attribute on DSR

| Option | Selected |
|--------|----------|
| Required string (user-supplied) | |
| Computed + UseStateForUnknown (server-generated) | ✓ |
| Optional + Computed with default | |

**Rationale:** Swagger explicitly marks `name` as `readOnly: true` on `DirectoryServiceRole`. POST body has no `name` field and no `names` query param. Server generates the name (likely derived from the associated policy). Overrides REQ DSR-01 wording that implied user-named roles.

---

## role attribute handling on DSR

| Option | Selected |
|--------|----------|
| Expose `role` as the primary writable attribute | |
| Drop `role` entirely, expose only `management_access_policies` | |
| Expose `role` as Computed-only (API-populated legacy backfill) | ✓ |

**Rationale:** Swagger marks `role` as Deprecated with note _"deprecated in favor of management_access_policies, but remains for backwards compatibility"_. Using it as a writable input would lock users into the deprecated path. Surfacing it as Computed lets operators see the legacy-compatibility backfill when present, without diff noise.

---

## management_access_policies handling on DSR

| Option | Selected |
|--------|----------|
| Optional + Computed list | |
| Required list + RequiresReplace on change | ✓ |
| Computed-only list (users must use DSRM for all associations) | |

**Rationale:** Writable on POST but **readonly on PATCH**. Cannot mutate via the role endpoint after creation. RequiresReplace forces re-creation if the user changes the HCL list, and documentation steers users to the DSRM membership resource for post-creation changes.

---

## Composite ID format for DSRM

| Option | Selected |
|--------|----------|
| `policy_name:role_name` (as suggested in REQ DSRM-03) | |
| `role_name/policy_name` using existing `compositeID()` / `parseCompositeID()` helper | ✓ |
| Custom escape/URL-encode policy name | |

**Rationale:** Built-in policy names contain both `:` and `/` (e.g. `pure:policy/array_admin`). Putting role_name **first** with `/` separator and using `SplitN(id, "/", 2)` lets the policy name retain all its embedded characters. Consistent with existing member resources. Corrects REQ DSRM-03 wording.

---

## DSR `role` / `management_access_policies` marshalling in HCL schema

| Option | Selected |
|--------|----------|
| Nested object shape `{ name = "..." }` | |
| Flat string / list-of-strings (mapped internally to NamedReference) | ✓ |

**Rationale:** Matches `target_resource.go` `ca_certificate_group = "corp-ca"` pattern — flat HCL UX, internal `**NamedReference` conversion in the Patch struct.

---

## DSRM attribute shape

| Option | Selected |
|--------|----------|
| Nested `policy { name = "..." }` + `role { name = "..." }` | |
| Flat `policy = "..."` + `role = "..."` both RequiresReplace | ✓ |

**Rationale:** Mirrors existing `qos_policy_member`, `tls_policy_member`, `certificate_group_member` — flat string attrs, all RequiresReplace. User explicitly requested the `_member` pattern.

---

## Interface assertions on member resource

| Option | Selected |
|--------|----------|
| 3 assertions (match existing qos/tls/object_store member pattern) | |
| 4 assertions incl. WithUpgradeState + empty map (match CONVENTIONS.md) | ✓ |

**Rationale:** CONVENTIONS.md mandates all 4. Existing `_member` resources predate the convention — new resources comply.

---

## Mock handler cross-store validation

| Option | Selected |
|--------|----------|
| Simple set of `(policy, role)` pairs | ✓ |
| Validate policy exists in management_access_policies store | |
| Validate role exists in directory_service_roles store | |

**Rationale:** Simpler mock; tests seed independently. No need to replicate cross-resource FK in the mock layer.

---

## Claude's Discretion

- Exact choice of separator character in mock handler store key (`|`, `\0`, or composite string) — implementation detail.
- Wording of tflog drift messages — follow `{resource, field, was, now}` shape from CONVENTIONS.md.
- Exact helper file placement for `_referenceWritable` ↔ flat-string conversion.
- Test seed values — prefer realistic policy names like `pure:policy/array_admin`.

## Open Questions forwarded to Planner

- Q1: Server-side name derivation for POST `/directory-services/roles` responses.
- Q2: Whether empty `management_access_policies` list is a valid POST payload.
- Q3: Whether DSRM POST is idempotent when the association already exists.

## Deferred Ideas

- `/directory-services/test` ephemeral resource.
- Management access policy data source.
- Active Directory account family.
- NFS directory service variant.
