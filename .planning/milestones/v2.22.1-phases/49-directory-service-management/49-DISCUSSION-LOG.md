# Phase 49: Directory Service Management - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-17
**Phase:** 49-directory-service-management
**Areas discussed:** name attribute shape, Delete reset strategy, management sub-object modelling, uris validation strictness

---

## name attribute shape

| Option | Description | Selected |
|--------|-------------|----------|
| Hardcoded internally | No `name` attribute in schema. Resource always targets `management`. Simplest UX, prevents misuse. Future nfs variant is a separate resource with its own hardcoded name. | ✓ |
| Required + validator = `management` | User writes `name = "management"` in HCL. Explicit but redundant for a service-locked resource. Forces RequiresReplace dance if anyone tries to change it. | |
| Optional + Computed, default `management` | Implicit `management` if omitted. Allows override but meaningless since resource is locked to management endpoint. Risk: user sets wrong value and API rejects. | |

**User's choice:** Hardcoded internally (recommended).
**Notes:** Import literal remains `management`. Future `flashblade_directory_service_nfs` will follow the same pattern as a sibling resource.

---

## Delete reset strategy

| Option | Description | Selected |
|--------|-------------|----------|
| Full reset | PATCH with `enabled=false`, empty `uris`, empty `bind_user`/`base_dn`, null `ca_certificate`/`ca_certificate_group`, empty management sub-object. Clean slate — matches "delete" semantics. Next apply starts fresh. | ✓ |
| Disable only | PATCH `enabled=false`; keep uris/bind_user/base_dn/refs in place. Faster re-enable but risks state drift if config is recreated with different values. | |
| Disable + clear sensitive | PATCH `enabled=false`; `bind_password` cleared implicitly. Keeps non-sensitive config but loses auth. | |

**User's choice:** Full reset (recommended).
**Notes:** Clear semantic contract — `terraform destroy` leaves the array ready for a fresh `terraform apply`.

---

## management sub-object modelling

| Option | Description | Selected |
|--------|-------------|----------|
| Optional + Computed | Unset → API defaults (sAMAccountName / uid / User / posixAccount / person). Set → value wins. Drift detected when API diverges from state. Matches computed-with-default pattern. | ✓ |
| Optional only | Unset → null in state; API default causes false drift. Requires drift suppression or forced PATCH on every apply. Bad UX. | |
| Required | Force users to set all three attributes. Most explicit but noisy — most users want API defaults. | |

**User's choice:** Optional + Computed (recommended).
**Notes:** Standard pattern for fields with API-populated defaults. No plan modifiers — let the framework surface drift.

---

## uris validation strictness

| Option | Description | Selected |
|--------|-------------|----------|
| Scheme validator | Reject URIs not starting with `ldap://` or `ldaps://`. Catches typos at plan time. Consistent with existing Alphanumeric/HostnameNoDot validators. | ✓ |
| Permissive (no validator) | Accept any strings; API rejects at apply time. Worse UX for typos. | |
| Full URI validator | Validate full URI form (scheme+host+port). More thorough but risks rejecting valid but unusual LDAP URI forms. Higher maintenance. | |

**User's choice:** Scheme validator (recommended).
**Notes:** Add alongside existing validators in `internal/provider/validators/` package. Error message format: `"uris[N] must start with ldap:// or ldaps://"`.

---

## Claude's Discretion

- Exact location of the URI scheme validator (dedicated package file vs inline helper) — choose based on existing layout.
- Wording of `tflog.Debug` drift messages — follow `{resource, field, was, now}` shape from CONVENTIONS.md.
- Ordering of fields inside the schema (logical grouping: identity, connection, TLS, management sub-object, state).
- Test seed values — use `ldaps://ldap.example.com:636`-style entries for readability.

## Deferred Ideas

- `flashblade_directory_service_nfs` sibling resource — future milestone.
- Directory service roles / role mappings (`/directory-services/roles`) — future milestone.
- Active Directory accounts (`/active-directory`) — future milestone.
- `/directory-services/test` ephemeral action — better as future CLI helper.
- SMB sub-object surface — deferred permanently (DEPRECATED in v2.22).
