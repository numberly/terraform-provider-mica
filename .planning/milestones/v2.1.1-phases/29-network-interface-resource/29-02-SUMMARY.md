---
phase: 29-network-interface-resource
plan: 02
subsystem: provider
tags: [resource, data-source, network-interface, crud, validators, tdd]
dependency_graph:
  requires: ["29-01"]
  provides: ["flashblade_network_interface resource", "flashblade_network_interface data source"]
  affects: ["provider.go registration"]
tech_stack:
  added: []
  patterns:
    - "serviceTypeValidator: inline validator.String enum pattern"
    - "networkInterfaceServicesValidator: resource.ConfigValidator cross-field validation"
    - "AttachedServers uses empty list (not null) to prevent spurious drift"
    - "services: []string collapsed to single types.String in schema"
    - "RequiresReplace on name, subnet_name, type"
key_files:
  created:
    - internal/provider/network_interface_resource.go
    - internal/provider/network_interface_resource_test.go
    - internal/provider/network_interface_data_source.go
    - internal/provider/network_interface_data_source_test.go
  modified:
    - internal/provider/provider.go
decisions:
  - "AttachedServers uses empty list (not null) when API returns no servers — prevents spurious drift on next plan"
  - "services collapsed from []string to single types.String — API enforces single service per NI"
  - "ConfigValidator defers when Services or AttachedServers is Unknown — supports plan-time deferral"
  - "niServersToNamedRefs returns nil (not []) for null/empty list — PATCH struct handles zero-value via Always-include semantics"
metrics:
  duration_minutes: 8
  completed_date: "2026-03-31"
  tasks_completed: 2
  files_created: 4
  files_modified: 1
---

# Phase 29 Plan 02: Network Interface Resource & Data Source Summary

**One-liner:** flashblade_network_interface resource (CRUD, cross-field ConfigValidator, RequiresReplace, drift detection, import) and data source registered in provider, with 22 unit tests.

## Tasks Completed

| # | Task | Commit | Key Files |
|---|------|--------|-----------|
| 1 | Create network interface resource with CRUD, validators, RequiresReplace, drift detection, import | d814d75 | network_interface_resource.go, network_interface_resource_test.go |
| 2 | Create network interface data source, register both in provider.go | b191781 | network_interface_data_source.go, network_interface_data_source_test.go, provider.go |

## What Was Built

### Resource: `flashblade_network_interface`

- **CRUD:** Create/Read/Update/Delete via `PostNetworkInterface`, `GetNetworkInterface`, `PatchNetworkInterface`, `DeleteNetworkInterface`
- **Schema:** `name`, `subnet_name`, `type` have `RequiresReplace`; `enabled`, `gateway`, `mtu`, `netmask`, `vlan`, `realms` are Computed with `UseStateForUnknown`
- **services:** Single `types.String` (not list) — the API accepts one service per network interface; collapsed from `[]string`
- **attached_servers:** `Optional+Computed`, `types.List` of StringType; empty list (not null) when API returns no servers
- **Validators:**
  - `serviceTypeValidator()`: enum validator accepting `data`, `sts`, `egress-only`, `replication`
  - `networkInterfaceServicesValidator` (ConfigValidator): `data`/`sts` require at least 1 `attached_server`; `egress-only`/`replication` forbid `attached_servers`
- **Drift detection:** Read logs `tflog.Info` on address or services change
- **ImportState:** by name, populates all fields with no drift on subsequent plan

### Data Source: `flashblade_network_interface`

- Reads NI by name; all fields Computed except `name` (Required)
- Returns error diagnostic on NotFound (not state removal — data source semantics)
- Same mapping logic as resource (`mapNetworkInterfaceToDataSourceModel`)

### Provider Registration

- `NewNetworkInterfaceResource` added to `Resources()` after `NewSubnetResource`
- `NewNetworkInterfaceDataSource` added to `DataSources()` after `NewLinkAggregationGroupDataSource`

## Test Coverage

22 unit tests across resource and data source:

**Resource (18 tests):**
- `TestUnit_NetworkInterface_Create` — POST, all fields populated
- `TestUnit_NetworkInterface_Update` — PATCH address and attached_servers
- `TestUnit_NetworkInterface_Delete` — DELETE, verify 404 after
- `TestUnit_NetworkInterface_Schema` — RequiresReplace, Computed, Optional+Computed
- `TestUnit_NetworkInterface_ServicesValidator` — valid/invalid enum values
- `TestUnit_NetworkInterface_ConfigValidator` — 8 cross-field combinations
- `TestUnit_NetworkInterface_Import` — ImportState by name, all fields match
- `TestUnit_NetworkInterface_Drift` — Read after external PATCH reflects new state
- `TestUnit_NetworkInterface_NotFound` — Read returns null state on 404
- `TestUnit_NetworkInterface_AttachedServersEmptyList` — empty list (not null) for no servers

**Data source (4 tests):**
- `TestUnit_NetworkInterfaceDataSource_Read` — all fields populated
- `TestUnit_NetworkInterfaceDataSource_NotFound` — error diagnostic
- `TestUnit_NetworkInterfaceDataSource_Schema` — name Required, rest Computed
- `TestUnit_NetworkInterfaceDataSource_WithServers` — attached_servers list populated

## Verification Results

```
go test ./internal/provider/ -run TestUnit_NetworkInterface -count=1 -v  → 22 tests PASS
go build ./...                                                            → SUCCESS
go test ./internal/... -count=1                                           → 480 tests PASS (no regressions)
go vet ./internal/...                                                     → no issues
```

## Deviations from Plan

None — plan executed exactly as written.

## Requirements Met

- NI-01: `flashblade_network_interface` resource exists and is registered
- NI-02: Create with name, address, subnet, type, services, attached_servers
- NI-03: Update address, services, attached_servers; subnet/type force replacement
- NI-04: Delete network interface
- NI-05: services enum validator (data, sts, egress-only, replication)
- NI-06: Cross-field validator (ConfigValidator) enforcing attached_servers rules
- NI-07: Data source reads NI by name
- NI-08: ImportState by name with no drift on subsequent plan
- NI-09: Drift detection logs changes to address, services, attached_servers
- NI-10: Computed fields (enabled, gateway, mtu, netmask, vlan, realms) populated after apply
