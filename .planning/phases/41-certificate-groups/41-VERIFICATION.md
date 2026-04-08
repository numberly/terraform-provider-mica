---
phase: 41-certificate-groups
verified: 2026-04-08T00:00:00Z
status: passed
score: 6/6 must-haves verified
re_verification: false
---

# Phase 41: Certificate Groups Verification Report

**Phase Goal:** Operators can manage certificate groups and their certificate memberships through Terraform, enabling CA certificate trust bundles for targets, array connections, and directory services
**Verified:** 2026-04-08
**Status:** passed
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|---------|
| 1 | Operator can create a certificate group by name via terraform apply; subsequent plan shows 0 diff | VERIFIED | `certificate_group_resource.go`: Create calls `PostCertificateGroup`, maps all fields to state. `id` has `UseStateForUnknown()`, `realms` uses empty-safe list mapping — no perpetual drift. |
| 2 | Operator can destroy a certificate group via terraform destroy without errors | VERIFIED | Delete calls `DeleteCertificateGroup`, handles `IsNotFound` as no-op. Mock handler deletes group + members, returns 200. |
| 3 | terraform import flashblade_certificate_group.x group-name populates id, name, realms; subsequent plan shows 0 diff | VERIFIED | ImportState fetches by name, maps all fields including empty-safe realms, initialises CRD timeouts with correct create/read/delete keys only. |
| 4 | data.flashblade_certificate_group reads an existing group by name and exposes id, name, realms | VERIFIED | `certificate_group_data_source.go`: Read maps all 3 fields inline, empty-safe realms, `IsNotFound` produces `AddError`. |
| 5 | Operator can add a certificate to a group via flashblade_certificate_group_member and remove it via terraform destroy | VERIFIED | `certificate_group_member_resource.go`: Create calls `PostCertificateGroupMember`, Delete calls `DeleteCertificateGroupMember`. Read uses `ListCertificateGroupMembers` + linear scan; not-found triggers `RemoveResource`. |
| 6 | Drift detection logs field-level changes via tflog when a certificate group is modified outside Terraform | VERIFIED | Read logs `tflog.Debug("drift detected", ...)` with resource/field/was/now keys for both `name` and `realms` fields (lines 166-183 of `certificate_group_resource.go`). |

**Score:** 6/6 truths verified

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/client/models_network.go` | CertificateGroup, CertificateGroupPost, CertificateGroupMember structs | VERIFIED | Lines 185-202: all 3 structs present with correct fields and JSON tags |
| `internal/client/certificate_groups.go` | 6 client methods | VERIFIED | All 6 exported: GetCertificateGroup, PostCertificateGroup, DeleteCertificateGroup, ListCertificateGroupMembers, PostCertificateGroupMember, DeleteCertificateGroupMember |
| `internal/testmock/handlers/certificate_groups.go` | Thread-safe mock with Seed+SeedMember, Facade | VERIFIED | Store uses `sync.Mutex`, separate `groups` and `members` maps, `CertificateGroupStoreFacade` exported, both Seed and SeedMember present |
| `internal/client/certificate_groups_test.go` | 7 TestUnit_ client tests | VERIFIED | 7 tests: Get_Found, Get_NotFound, Post, Delete, Member_Post, Member_List, Member_Delete |
| `internal/provider/certificate_group_resource.go` | Resource with 4 interface assertions, CRD | VERIFIED | All 4 interface assertions present (lines 22-25), Version 0, no Update, correct timeouts |
| `internal/provider/certificate_group_data_source.go` | Data source with 2 interface assertions | VERIFIED | 2 interface assertions (lines 16-17), Read maps id/name/realms |
| `internal/provider/certificate_group_member_resource.go` | Member resource with 4 interface assertions | VERIFIED | All 4 interface assertions (lines 20-23), UpgradeState returns empty map, composite ID import |
| `internal/provider/provider.go` | Registration of all 3 new components | VERIFIED | Lines 303, 304, 363: NewCertificateGroupResource, NewCertificateGroupMemberResource, NewCertificateGroupDataSource |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| `certificate_group_resource.go` | `client.GetCertificateGroup / PostCertificateGroup / DeleteCertificateGroup` | r.client calls in Create/Read/Delete | WIRED | `r.client.PostCertificateGroup` (line 127), `r.client.GetCertificateGroup` (line 154), `r.client.DeleteCertificateGroup` (line 213) |
| `certificate_group_member_resource.go` | `client.PostCertificateGroupMember / DeleteCertificateGroupMember / ListCertificateGroupMembers` | r.client calls in Create/Read/Delete | WIRED | All 3 client methods called and results consumed |
| `provider.go` | `NewCertificateGroupResource, NewCertificateGroupDataSource, NewCertificateGroupMemberResource` | Resources() and DataSources() slices | WIRED | All 3 registrations confirmed at lines 303, 304, 363 |
| `certificate_groups.go` | `getOneByName[CertificateGroup]` | generic helper | WIRED | Line 13: `return getOneByName[CertificateGroup](...)` |
| `certificate_groups.go` | `postOne[struct{}, CertificateGroupMember]` | generic helper for member POST | WIRED | Line 54: `return postOne[struct{}, CertificateGroupMember](...)` |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|--------------|--------|--------------------|--------|
| `certificate_group_resource.go` | `group` (Create), `data` (Read) | `r.client.PostCertificateGroup` / `r.client.GetCertificateGroup` — both backed by mock HTTP responses with real JSON | Yes | FLOWING |
| `certificate_group_data_source.go` | `config` | `d.client.GetCertificateGroup` — HTTP GET to mock, seeded data returned | Yes | FLOWING |
| `certificate_group_member_resource.go` | `members` (Read) | `r.client.ListCertificateGroupMembers` — pagination loop consuming real HTTP response | Yes | FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| All 735 tests pass | `go test ./... -count=1` | 735 passed | PASS |
| Lint clean | `make lint` | 0 issues | PASS |
| 7 client tests for certificate groups | `grep TestUnit_CertificateGroup internal/client/certificate_groups_test.go` | 7 functions | PASS |
| 6 provider tests for certificate groups | `grep TestUnit_CertificateGroup internal/provider/...` | 6 functions | PASS |
| Mock GET returns empty list (not 404) on unknown name | `handleGroupGet`: falls through to `WriteJSONListResponse(w, 200, items)` when filter misses | empty list HTTP 200 | PASS |

### Requirements Coverage

| Requirement | Source Plan | Description (inferred from plan) | Status | Evidence |
|-------------|------------|----------------------------------|--------|---------|
| CERTG-01 | 41-01, 41-02 | Client + resource for creating/reading/deleting certificate groups | SATISFIED | `GetCertificateGroup`, `PostCertificateGroup`, `DeleteCertificateGroup` + resource CRD lifecycle |
| CERTG-02 | 41-01, 41-02 | terraform import by name works | SATISFIED | `ImportState` fetches by name, maps all fields, correct CRD timeouts |
| CERTG-03 | 41-01, 41-02 | Data source reads group by name exposing id/name/realms | SATISFIED | `certificate_group_data_source.go` Read maps all 3 fields |
| CERTG-04 | 41-02 | `flashblade_certificate_group_member` resource for cert/group association | SATISFIED | `certificate_group_member_resource.go` full CRD lifecycle with composite ID import |
| CERTG-05 | 41-01, 41-02 | Client member methods + mock handlers + drift detection | SATISFIED | `PostCertificateGroupMember`, `DeleteCertificateGroupMember`, `ListCertificateGroupMembers` + mock + tflog drift logging |

Note: CERTG-01 through CERTG-05 are phase-local requirement IDs defined in plan frontmatter. They do not appear in `.planning/REQUIREMENTS.md` (which only tracks v2.2 requirements TGT/RC/BRL/DOC series). This is consistent with the REQUIREMENTS.md traceability table which stops at DOC-03/Phase 38.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| None | — | — | — | — |

Scanned all 9 new files for: TODO/FIXME, `return null`, empty implementations, hardcoded empty data passed to renderers. No issues found. `return []client.CertificateGroup{}` and `types.ListValueMust(..., []attr.Value{})` patterns are correct empty-safe initializations, not stubs.

### Human Verification Required

None. All success criteria verified programmatically.

### Gaps Summary

No gaps. All 6 observable truths verified, all 8 artifacts pass all 4 levels (exists, substantive, wired, data flowing), all 5 key links confirmed wired.

## CONVENTIONS.md Compliance

| Check | Status | Detail |
|-------|--------|--------|
| 4 interface assertions per resource | PASS | `certificate_group_resource.go` lines 22-25; `certificate_group_member_resource.go` lines 20-23 |
| 2 interface assertions per data source | PASS | `certificate_group_data_source.go` lines 16-17 |
| Test naming `TestUnit_<Resource>_<Operation>` | PASS | All 13 tests follow convention |
| Mock GET returns empty list HTTP 200 (not 404) | PASS | `handleGroupGet` and `handleMemberGet` both fall through to `WriteJSONListResponse(w, 200, ...)` |
| Generic helpers `getOneByName[T]` and `postOne[T]` used | PASS | `getOneByName[CertificateGroup]` (line 13), `postOne[CertificateGroupPost, CertificateGroup]` (line 19), `postOne[struct{}, CertificateGroupMember]` (line 54) |
| Plan modifiers: `UseStateForUnknown` only on stable computed fields | PASS | `id` has `UseStateForUnknown`; `realms` has none (volatile) |
| `RequiresReplace` on immutable name fields | PASS | `name` in both resources has `RequiresReplace` |
| Drift detection in Read via `tflog.Debug` | PASS | `name` and `realms` both logged with resource/field/was/now keys |
| `UpgradeState` returns empty map at Version 0 | PASS | Both resources return `map[int64]resource.StateUpgrader{}` |
| CRD timeouts (create/read/delete only, no update) | PASS | Schema `timeouts.Opts{Create: true, Read: true, Delete: true}` |
| ImportState uses inline CRD null timeouts | PASS | Inline `timeouts.Value{Object: types.ObjectNull(...)}` with create/read/delete keys |
| Store has `sync.Mutex` | PASS | `certificateGroupStore` has `mu sync.Mutex` |
| Mock registered with API version path | PASS | `/api/2.22/certificate-groups/certificates` and `/api/2.22/certificate-groups` |
| Certificates sub-path registered before parent path | PASS | Line 32 registers `/certificates` before line 33 registers `/certificate-groups` |
| Total test count >= 729 (requirement) | PASS | 735 tests — exceeds requirement |
| CONVENTIONS.md baseline updated | PASS | Updated to 735 |
| ROADMAP.md updated, Certificate Groups marked Done | PASS | Line 89: `| Certificate Groups | flashblade_certificate_group | Yes | Done | ...` |
| HCL examples present | PASS | `resource.tf` + `import.sh` for both resources, `data-source.tf` for data source |
| docs/ regenerated by `make docs` | PASS | `docs/resources/certificate_group.md`, `docs/resources/certificate_group_member.md`, `docs/data-sources/certificate_group.md` all present |
| `make lint` 0 issues | PASS | Confirmed |
| `make test` all pass | PASS | 735/735 |

---

_Verified: 2026-04-08_
_Verifier: Claude (gsd-verifier)_
