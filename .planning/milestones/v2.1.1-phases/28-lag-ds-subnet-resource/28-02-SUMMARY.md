---
phase: 28-lag-ds-subnet-resource
plan: "02"
subsystem: provider-layer
tags: [resource, datasource, subnet, lag, network, terraform]
dependency_graph:
  requires:
    - "client.Subnet, client.SubnetPost, client.SubnetPatch (from 28-01)"
    - "client.LinkAggregationGroup (from 28-01)"
    - "FlashBladeClient.GetSubnet, PostSubnet, PatchSubnet, DeleteSubnet (from 28-01)"
    - "FlashBladeClient.GetLinkAggregationGroup (from 28-01)"
    - "handlers.RegisterSubnetHandlers, handlers.RegisterLinkAggregationGroupHandlers (from 28-01)"
  provides:
    - "flashblade_subnet resource (full CRUD + import + drift detection)"
    - "flashblade_subnet data source"
    - "flashblade_link_aggregation_group data source"
    - "NewSubnetResource, NewSubnetDataSource, NewLinkAggregationGroupDataSource registered in provider"
  affects:
    - "Phase 29 (network interfaces reference subnets by name)"
tech_stack:
  added: []
  patterns:
    - "lag_name as flat types.String (not nested object) — consistent with all reference patterns"
    - "lagNameToRef/refToLagName helpers for NamedReference conversion"
    - "partial PATCH via pointer types — only changed fields sent to API"
    - "mapSubnetToModel shared helper used by resource and data source"
    - "Read logs drift via tflog.Info when key fields change"
    - "ImportState by name with nullTimeoutsValue initialization"
key_files:
  created:
    - internal/provider/subnet_resource.go
    - internal/provider/subnet_resource_test.go
    - internal/provider/subnet_data_source.go
    - internal/provider/subnet_data_source_test.go
    - internal/provider/link_aggregation_group_data_source.go
    - internal/provider/link_aggregation_group_data_source_test.go
  modified:
    - internal/provider/provider.go
decisions:
  - "lag_name exposed as flat types.String — consistent with all other reference attributes in the codebase"
  - "mapSubnetToModel is shared between resource and a separate mapSubnetToDataSourceModel (to avoid circular dependency between model types)"
  - "Services and Interfaces computed lists use types.ListNull(types.StringType) for empty — not emptyStringList() — to signal absent data vs empty"
  - "Drift detection uses tflog.Info (not Warn) for gateway/prefix/mtu changes — informational, not actionable"
metrics:
  duration: "~5 minutes"
  completed_date: "2026-03-31"
  tasks_completed: 2
  tasks_total: 2
  files_created: 6
  files_modified: 1
---

# Phase 28 Plan 02: Provider Layer (Resource + Data Sources) Summary

**One-liner:** flashblade_subnet resource with full CRUD/import/drift detection and flashblade_link_aggregation_group data source backed by separate struct types and mock-server integration tests.

## Objective

Build the Terraform provider layer for Phase 28: subnet resource (CRUD + import + drift), subnet data source, LAG data source, integration tests, and provider registration. This completes Phase 28 and unblocks Phase 29 (network interfaces reference subnets).

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Create subnet resource with CRUD, import, and drift detection | 99b2ff0 | subnet_resource.go, subnet_resource_test.go |
| 2 | Create subnet data source, LAG data source, and provider registration | a1b9f13 | subnet_data_source.go, subnet_data_source_test.go, link_aggregation_group_data_source.go, link_aggregation_group_data_source_test.go, provider.go |

## What Was Built

### Task 1 — Subnet Resource

**`internal/provider/subnet_resource.go`:** Full resource implementation:
- `subnetResource` + `NewSubnetResource()` factory — `Resource`, `ResourceWithConfigure`, `ResourceWithImportState`
- `subnetResourceModel` with all fields: ID (Computed/UseStateForUnknown), Name (Required/RequiresReplace), Prefix (Required), Gateway/LagName (Optional/Computed), MTU/VLAN (Optional/Computed/UseStateForUnknown), Enabled/Services/Interfaces (Computed), Timeouts
- `lagNameToRef` / `refToLagName` helpers for NamedReference ↔ flat string conversion
- `mapSubnetToModel` maps API response to model with services/interfaces list conversion
- **Create:** Builds SubnetPost, calls PostSubnet, maps response
- **Read:** Calls GetSubnet, handles 404 with RemoveResource, logs drift via tflog.Info for prefix/gateway/mtu changes
- **Update:** Partial PATCH — compares plan vs state with `.Equal()`, only includes changed fields via pointer types
- **Delete:** Calls DeleteSubnet, tolerates 404
- **ImportState:** Fetches by name, initializes Timeouts with nullTimeoutsValue

**`internal/provider/subnet_resource_test.go`:** 6 unit tests:
- Create: verifies all attributes populated (ID, name, prefix, gateway, mtu, vlan, lag_name, enabled)
- Update: create then update gateway/mtu, verify new values
- Delete: create then delete, verify 404 from API
- Import: create then import by name, verify all fields
- Drift: create, patch externally via client, Read, verify state reflects new value
- NotFound: Read on missing subnet removes state (null)

### Task 2 — Data Sources and Provider Registration

**`internal/provider/subnet_data_source.go`:** Read-only data source:
- Same field set as resource minus Timeouts
- `mapSubnetToDataSourceModel` (separate from resource mapper to keep types independent)
- Returns "Subnet not found" diagnostic on 404

**`internal/provider/link_aggregation_group_data_source.go`:** Read-only LAG data source:
- `lagDataSourceModel`: Name (Required), ID/Status/MacAddress/PortSpeed/LagSpeed/Ports (Computed)
- GetLinkAggregationGroup + Ports list mapping
- Returns "Link aggregation group not found" diagnostic on 404

**`internal/provider/provider.go`:** Registered 3 new entries:
- `NewSubnetResource` in `Resources()` after `NewQosPolicyMemberResource`
- `NewSubnetDataSource` and `NewLinkAggregationGroupDataSource` in `DataSources()` after `NewQosPolicyDataSource`

## Verification

```
go test ./internal/... -count=1 -timeout 120s
```

Result: **PASSED — 458 tests across 4 packages, 0 regressions**.

New tests by group:
- `TestUnit_SubnetResource_*`: 6 tests (Create, Update, Delete, Import, Drift, NotFound)
- `TestUnit_SubnetDataSource_*`: 3 tests (Read, NotFound, Schema)
- `TestUnit_LagDataSource_*`: 3 tests (Read, NotFound, Schema)

## Deviations from Plan

None — plan executed exactly as written.

## Self-Check

- [x] `internal/provider/subnet_resource.go` exists — 307 lines, exports NewSubnetResource
- [x] `internal/provider/subnet_resource_test.go` exists — 6 TestUnit_SubnetResource_* tests
- [x] `internal/provider/subnet_data_source.go` exists — exports NewSubnetDataSource
- [x] `internal/provider/subnet_data_source_test.go` exists — TestUnit_SubnetDataSource_* tests
- [x] `internal/provider/link_aggregation_group_data_source.go` exists — exports NewLinkAggregationGroupDataSource
- [x] `internal/provider/link_aggregation_group_data_source_test.go` exists — TestUnit_LagDataSource_* tests
- [x] `internal/provider/provider.go` registers all 3 new types
- [x] Commit 99b2ff0: feat(28-02): add flashblade_subnet resource
- [x] Commit a1b9f13: feat(28-02): add subnet and LAG data sources
- [x] All 458 tests pass — no regressions

## Self-Check: PASSED
