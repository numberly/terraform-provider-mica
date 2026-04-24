---
phase: 49-directory-service-management
plan: "04"
subsystem: provider-datasource
tags: [go, terraform-provider, flashblade, directory-service, ldap, data-source]

# Dependency graph
requires:
  - phase: 49-directory-service-management-plan-01
    provides: DirectoryService struct + GetDirectoryServiceManagement client method
  - phase: 49-directory-service-management-plan-02
    provides: RegisterDirectoryServicesHandlers mock handler with Seed()

provides:
  - flashblade_directory_service_management data source (computed-only, singleton)
  - NewDirectoryServiceManagementDataSource factory function
  - TestUnit_DirectoryServiceManagementDataSource_Basic test

affects:
  - 49-05 (provider registration must include NewDirectoryServiceManagementDataSource)

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Singleton data source: no name argument, always reads management"
    - "Computed-only schema: no bind_password, no name attribute (D-06)"
    - "SingleNestedAttribute for NamedReference objects (ca_certificate, ca_certificate_group)"
    - "namedRefAttrTypes/namedRefObjectValue helpers for nested object conversion"

key-files:
  created:
    - internal/provider/directory_service_management_data_source.go
    - internal/provider/directory_service_management_data_source_test.go
  modified: []

key-decisions:
  - "namedRefAttrTypes and namedRefObjectValue defined in data source file (package-level helpers, no naming conflict with 49-03 resource file)"
  - "bind_password excluded from schema per D-06 (write-only, never returned by API)"
  - "No name attribute in schema — always reads singleton management internally"
  - "2 interface assertions only: DataSource + DataSourceWithConfigure (no ImportState, no UpgradeState)"

requirements-completed: [DSM-08, QA-03, QA-05]

# Metrics
duration: 3min
completed: 2026-04-17
---

# Phase 49 Plan 04: Directory Service Management Data Source Summary

**Computed-only singleton data source for FlashBlade management LDAP config: no name/bind_password, ca_certificate and ca_certificate_group as nested objects with name sub-attribute, 1 TestUnit_DirectoryServiceManagementDataSource_Basic test passing**

## Performance

- **Duration:** ~3 min
- **Started:** 2026-04-17T08:01:28Z
- **Completed:** 2026-04-17T08:03:52Z
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments

- Created `directory_service_management_data_source.go` (215 lines): 2 interface assertions, computed-only schema (11 attributes), namedRefAttrTypes/namedRefObjectValue helpers, Read calling GetDirectoryServiceManagement(ctx, "management") with IsNotFound handling
- Created `directory_service_management_data_source_test.go` (156 lines): helpers (newTestDirectoryServiceManagementDataSource, directoryServiceManagementDSSchema, buildDirectoryServiceManagementDSType, nullDirectoryServiceManagementDSConfig) + TestUnit_DirectoryServiceManagementDataSource_Basic asserting 9 fields

## Task Commits

1. **Task 1: Create directoryServiceManagementDataSource with computed-only schema** - `db13b87` (feat)
2. **Task 2: Create TestUnit_DirectoryServiceManagementDataSource_Basic** - `f12ca35` (test)

## Files Created/Modified

- `internal/provider/directory_service_management_data_source.go` — Data source: 2 interfaces, schema with SingleNestedAttribute for ca_certificate/ca_certificate_group, Read with inline mapping
- `internal/provider/directory_service_management_data_source_test.go` — Basic test: seed + read + 9-field assertions

## Schema Fields

| Attribute | Type | Notes |
|-----------|------|-------|
| `id` | String (Computed) | Unique identifier |
| `enabled` | Bool (Computed) | Directory service enabled state |
| `uris` | List(String) (Computed) | LDAP server URIs |
| `base_dn` | String (Computed) | Base Distinguished Name |
| `bind_user` | String (Computed) | Bind user DN |
| `ca_certificate` | SingleNestedAttribute (Computed) | Nested object with `name` sub-attribute |
| `ca_certificate_group` | SingleNestedAttribute (Computed) | Nested object with `name` sub-attribute |
| `user_login_attribute` | String (Computed) | LDAP login attribute |
| `user_object_class` | String (Computed) | LDAP object class |
| `ssh_public_key_attribute` | String (Computed) | SSH key LDAP attribute |
| `services` | List(String) (Computed) | Services using this config |

**Excluded per D-06:** `name` (singleton, always management), `bind_password` (write-only)

## Decisions Made

- `namedRefAttrTypes` and `namedRefObjectValue` defined as package-level helpers in the data source file — no collision with 49-03 resource file helpers since both live in the `provider` package but operate on different types
- Schema uses `schema.SingleNestedAttribute` with `Attributes: map[string]schema.Attribute{"name": schema.StringAttribute{Computed: true}}` for both TLS reference fields
- IsNotFound error produces `AddError` (not `RemoveResource`) per CONVENTIONS.md §"Data Source Implementation"

## Deviations from Plan

None — plan executed exactly as written.

## Known Stubs

None. All schema fields populated from live API response in Read; test seeds verify end-to-end mapping.

## Self-Check: PASSED

- `internal/provider/directory_service_management_data_source.go` — FOUND
- `internal/provider/directory_service_management_data_source_test.go` — FOUND
- Commit `db13b87` — FOUND
- Commit `f12ca35` — FOUND
- `go build ./internal/provider/...` — passes
- `go test ./internal/provider/...` — 475 passed
- `TestUnit_DirectoryServiceManagementDataSource_Basic` — PASS

---
*Phase: 49-directory-service-management*
*Completed: 2026-04-17*
