# Requirements — Milestone v2.22.2 Directory Service Roles & Role Mappings

**Milestone goal:** Ajouter la gestion Terraform des role mappings LDAP ↔ FlashBlade
via deux ressources: `flashblade_directory_service_role` (role standalone + data source)
et `flashblade_management_access_policy_directory_service_role_membership` (association
composite role ↔ management_access_policy), suivant le pattern `_member` déjà établi
dans le provider.

**Numbering:** `CATEGORY-NN`. DSR = Directory Service Role, DSRM = Directory Service
Role Membership, DOC = Documentation, QA = Quality. Numbering restart per milestone.

---

## Active (v2.22.2)

### Directory Service Role (DSR)

- [x] **DSR-01** — User can create a directory service role via a
  `flashblade_directory_service_role` resource, specifying `group` (CN of the LDAP
  group), `group_base` (DN search base), and `role` (NamedReference to a built-in
  FlashBlade role: `array_admin`, `storage_admin`, `ops_admin`, `readonly`). Backed
  by `POST /directory-services/roles`.
- [x] **DSR-02** — User can update `group` and `group_base` via PATCH without
  replacing the resource. Changing `role` triggers `RequiresReplace` (role is
  immutable on the FB side — validated via the API contract).
- [x] **DSR-03** — User can destroy a role cleanly via
  `DELETE /directory-services/roles?names=<name>`. No soft-delete dance required.
- [x] **DSR-04** — User can import an existing role by name with
  `terraform import flashblade_directory_service_role.<alias> <role_name>`. Import
  initialises timeouts to null.
- [x] **DSR-05** — Terraform detects drift on every mutable/computed field
  (`group`, `group_base`, `role.name`, `management_access_policies` list) and
  logs via `tflog.Debug` with the standard `{resource, field, was, now}` shape.
  `management_access_policies` is a computed-only list of policy references
  (populated by the API — membership resources manage the actual associations).
- [x] **DSR-06** — User can read an existing role through a
  `flashblade_directory_service_role` data source (computed-only schema keyed
  by required `name` attribute).

### Directory Service Role Membership (DSRM)

- [x] **DSRM-01** — User can associate a management access policy with a
  directory service role via a
  `flashblade_management_access_policy_directory_service_role_membership` resource.
  Required attributes: `policy` (NamedReference to the management access policy)
  and `role` (NamedReference to the directory service role). Backed by
  `POST /management-access-policies/directory-services/roles`.
- [x] **DSRM-02** — Destroying the membership calls
  `DELETE /management-access-policies/directory-services/roles` with both
  `policy_names` and `role_names` query params — association is removed but
  neither the policy nor the role is deleted.
- [x] **DSRM-03** — User can import an existing membership with composite ID
  `<policy_name>:<role_name>`, matching the pattern used by
  `qos_policy_member_resource.go`, `tls_policy_member_resource.go`, and
  `certificate_group_member_resource.go`.
- [x] **DSRM-04** — Both `policy` and `role` trigger `RequiresReplace` on
  change (memberships are immutable — you create a new pair, not rename an
  existing one).
- [x] **DSRM-05** — Read verifies the membership still exists via
  `GET /management-access-policies/directory-services/roles?policy_names=<p>&role_names=<r>`.
  If the list is empty (association removed outside Terraform), call
  `resp.State.RemoveResource(ctx)`.

### Documentation (DOC)

- [ ] **DOC-01** — Working HCL examples at
  `examples/resources/flashblade_directory_service_role/resource.tf` plus
  `import.sh` with a canonical import command, and
  `examples/data-sources/flashblade_directory_service_role/data-source.tf`.
- [ ] **DOC-02** — Working HCL example at
  `examples/resources/flashblade_management_access_policy_directory_service_role_membership/resource.tf`
  plus `import.sh` using the composite ID format
  `<policy_name>:<role_name>`.
- [ ] **DOC-03** — `make docs` regenerates
  `docs/resources/directory_service_role.md`,
  `docs/data-sources/directory_service_role.md`, and
  `docs/resources/management_access_policy_directory_service_role_membership.md`
  without manual edits.

### Quality (QA)

- [x] **QA-01** — ≥4 client unit tests named `TestUnit_DirectoryServiceRole_*`
  covering `GET` (found + not-found), `POST`, `PATCH` (group/group_base), and
  `DELETE` on the standalone role.
- [x] **QA-02** — ≥3 client unit tests named
  `TestUnit_ManagementAccessPolicyDirectoryServiceRoleMembership_*` covering
  `GET` (exists + not-exists), `POST`, and `DELETE`.
- [x] **QA-03** — ≥3 resource unit tests for the role:
  `TestUnit_DirectoryServiceRoleResource_Lifecycle`, `_Import`, `_DriftDetection`.
- [x] **QA-04** — ≥3 resource unit tests for the membership:
  `TestUnit_ManagementAccessPolicyDirectoryServiceRoleMembershipResource_Lifecycle`,
  `_Import`, `_MissingAssociation` (the last validates the RemoveResource path
  from DSRM-05).
- [x] **QA-05** — ≥1 data source unit test:
  `TestUnit_DirectoryServiceRoleDataSource_Basic`.
- [ ] **QA-06** — `make test` passes with total count ≥ 812 (baseline 798 from
  v2.22.1 + 14 new: 4+3 client + 3+3 resource + 1 data source).
- [x] **QA-07** — `make lint` clean (0 issues). Both resources declare the
  four standard interface assertions (`Resource`, `WithConfigure`,
  `WithImportState`, `WithUpgradeState`) with empty `UpgradeState` map at
  schema version 0.
- [ ] **QA-08** — `ROADMAP.md` (root-level API coverage) updated in the same
  commit as the implementation: entries moved from "Not Implemented" to
  "Implemented → Array Administration", header counters refreshed,
  `Last updated` date bumped to the shipping date.

---

## Future Requirements (deferred)

- `/directory-services/test` ephemeral resource — dry-run LDAP bind validation
  for CI pipelines (pre-apply check). Different resource shape (ephemeral),
  better handled in its own milestone.
- Role scope/locality attributes — if the API later exposes multi-array role
  propagation fields, revisit.
- Bulk role import (HCL `for_each` from LDAP group list) — usability helper,
  not a provider requirement.

## Out of Scope (v2.22.2)

- Active Directory accounts (`/active-directory`) — separate endpoint family
  (Kerberos AD joins for NFS/SMB), unrelated to RBAC role mappings.
- NFS directory service variant — separate milestone.
- Management access policy resource itself (`/management-access-policies`) —
  built-in policies ship with FlashBlade (`pure:policy/array_admin`, etc.)
  and cannot be created or modified by customers; a read-only data source
  could be useful but is not in this milestone's scope.

---

## Traceability

| REQ-ID  | Phase | Status |
|---------|-------|--------|
| DSR-01  | 50    | Complete |
| DSR-02  | 50    | Complete |
| DSR-03  | 50    | Complete |
| DSR-04  | 50    | Complete |
| DSR-05  | 50    | Complete |
| DSR-06  | 50    | Complete |
| DSRM-01 | 50    | Complete |
| DSRM-02 | 50    | Complete |
| DSRM-03 | 50    | Complete |
| DSRM-04 | 50    | Complete |
| DSRM-05 | 50    | Complete |
| DOC-01  | 50    | Pending |
| DOC-02  | 50    | Pending |
| DOC-03  | 50    | Pending |
| QA-01   | 50    | Complete |
| QA-02   | 50    | Complete |
| QA-03   | 50    | Complete |
| QA-04   | 50    | Complete |
| QA-05   | 50    | Complete |
| QA-06   | 50    | Pending |
| QA-07   | 50    | Complete |
| QA-08   | 50    | Pending |

*Filled by gsd-roadmapper.*
