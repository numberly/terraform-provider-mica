# Requirements — Milestone v2.22.1 Directory Service – Array Management

**Milestone goal:** Ajouter la gestion Terraform du directory service LDAP `management`
(authentification admin) via une ressource singleton PATCH-only, avec data source,
drift detection, import, et documentation complète.

**Numbering:** `CATEGORY-NN`. DSM = Directory Service Management, DOC = Documentation,
QA = Quality. Numbering restart per milestone.

---

## Active (v2.22.1)

### Directory Service Management (DSM)

- [ ] **DSM-01** — User can configure the LDAP `management` directory service
  (`uris`, `base_dn`, `bind_user`) via a `flashblade_directory_service_management`
  resource backed by `PATCH /directory-services?names=management`.
- [ ] **DSM-02** — User can provide a sensitive, write-only `bind_password` that
  is never returned by the API nor surfaced in plan/state diffs.
- [ ] **DSM-03** — User can reference a `ca_certificate` and/or
  `ca_certificate_group` by name (`NamedReference` pattern) for LDAPS TLS
  validation, and clear either reference by omitting the attribute.
- [ ] **DSM-04** — User can set management-specific LDAP attributes:
  `user_login_attribute`, `user_object_class`, `ssh_public_key_attribute`
  (nested under the `management` object in the PATCH body).
- [ ] **DSM-05** — User can enable or disable the management directory service
  via the `enabled` boolean.
- [ ] **DSM-06** — User can import an existing configuration by name (always
  `"management"`) with `terraform import`. Import initialises timeouts to null
  and leaves `bind_password` empty (write-once).
- [ ] **DSM-07** — Terraform detects drift on every mutable/computed field
  (`enabled`, `uris`, `base_dn`, `bind_user`, `ca_certificate.name`,
  `ca_certificate_group.name`, `user_login_attribute`, `user_object_class`,
  `ssh_public_key_attribute`, `services`) and logs via `tflog.Debug` with the
  standard `{resource, field, was, now}` shape.
- [ ] **DSM-08** — User can read the current management configuration through
  a `flashblade_directory_service_management` data source (computed-only
  schema, no `bind_password`).

### Documentation (DOC)

- [ ] **DOC-01** — Working HCL example at
  `examples/resources/flashblade_directory_service_management/resource.tf` plus
  `import.sh` with the canonical `terraform import … management` command.
- [ ] **DOC-02** — Working data source example at
  `examples/data-sources/flashblade_directory_service_management/data-source.tf`.
- [ ] **DOC-03** — `make docs` regenerates
  `docs/resources/directory_service_management.md` and the matching data
  source doc without manual edits.

### Quality (QA)

- [ ] **QA-01** — ≥4 client unit tests named `TestUnit_DirectoryServiceManagement_*`
  covering `GET` (found + not-found) and `PATCH` (at least one field variant +
  one reference variant).
- [ ] **QA-02** — ≥3 resource unit tests: `Lifecycle`, `Import`,
  `DriftDetection` with `TestUnit_DirectoryServiceManagementResource_*` naming.
- [ ] **QA-03** — ≥1 data source unit test: `TestUnit_DirectoryServiceManagementDataSource_Basic`.
- [ ] **QA-04** — `make test` passes with total count ≥ 787 (baseline 779 + 8 new).
- [ ] **QA-05** — `make lint` clean (0 issues). Resource declares all four
  interface assertions (`Resource`, `WithConfigure`, `WithImportState`,
  `WithUpgradeState`) and an empty `UpgradeState` map at schema version 0.
- [ ] **QA-06** — `ROADMAP.md` updated in the same commit as the implementation:
  entry moved from "Not Implemented → High Priority" to "Implemented → Array
  Administration", header counters refreshed, `Last updated` date bumped.

---

## Future Requirements (deferred)

- Directory Service NFS variant (`flashblade_directory_service_nfs`) — same
  endpoint, different nested object (`nis_domains`, `nis_servers`).
- Directory Service roles / role mappings (`/directory-services/roles`) —
  maps LDAP groups to FlashBlade roles.
- Active Directory accounts (`/active-directory`) — separate endpoint family
  for Kerberos AD joins used by NFS/SMB.

## Out of Scope (v2.22.1)

- `flashblade_directory_service_smb` — the `smb` sub-object is DEPRECATED in
  the v2.22 API, tracked only for completeness; no Terraform surface.
- Directory service `test` endpoint (`/directory-services/test`) — ephemeral
  validation action, not stateful, better suited as a future CLI helper than a
  Terraform resource.
- Multi-service singleton (`flashblade_directory_service` with `name` selector)
  — rejected in favour of one resource per service to keep attribute sets
  disjoint.

---

## Traceability

| REQ-ID | Phase | Status |
|--------|-------|--------|
| DSM-01 | Phase 49 | Pending |
| DSM-02 | Phase 49 | Pending |
| DSM-03 | Phase 49 | Pending |
| DSM-04 | Phase 49 | Pending |
| DSM-05 | Phase 49 | Pending |
| DSM-06 | Phase 49 | Pending |
| DSM-07 | Phase 49 | Pending |
| DSM-08 | Phase 49 | Pending |
| DOC-01 | Phase 49 | Pending |
| DOC-02 | Phase 49 | Pending |
| DOC-03 | Phase 49 | Pending |
| QA-01  | Phase 49 | Pending |
| QA-02  | Phase 49 | Pending |
| QA-03  | Phase 49 | Pending |
| QA-04  | Phase 49 | Pending |
| QA-05  | Phase 49 | Pending |
| QA-06  | Phase 49 | Pending |

*Filled by gsd-roadmapper.*
