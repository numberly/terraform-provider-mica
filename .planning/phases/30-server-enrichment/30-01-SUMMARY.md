---
phase: 30-server-enrichment
plan: "01"
subsystem: server-enrichment
tags: [server, network-interfaces, vip, schema-migration, enrichment]
dependency_graph:
  requires: [internal/client/network_interfaces.go, internal/client/models_network.go]
  provides: [server resource v1 schema, server data source with network_interfaces]
  affects: [flashblade_server resource, flashblade_server data source]
tech_stack:
  added: []
  patterns: [computed-list enrichment, state-upgrader-v0-to-v1, client-side-filter]
key_files:
  created: []
  modified:
    - internal/provider/server_resource.go
    - internal/provider/server_data_source.go
    - internal/provider/server_resource_test.go
    - internal/provider/server_data_source_test.go
decisions:
  - "StateUpgrader v0->v1 sets network_interfaces to empty list on old state"
  - "VIP enrichment uses warning diagnostic (not error) to avoid blocking CRUD when ListNetworkInterfaces fails"
  - "Use types.ListValueMust(types.StringType, []attr.Value{}) for empty list to prevent spurious drift"
  - "Client-side filter on attached_servers — API does not support filter by attached server"
metrics:
  duration_seconds: 362
  completed_date: "2026-03-31"
  tasks_completed: 2
  files_modified: 4
---

# Phase 30 Plan 01: Server Enrichment — network_interfaces Attribute Summary

**One-liner:** Schema v0->v1 migration with computed network_interfaces list populated via ListNetworkInterfaces + client-side filter by attached_servers server name.

## What Was Built

Added computed `network_interfaces` attribute to both `flashblade_server` resource and data source. Operators can now discover which VIPs are attached to a server directly from Terraform state without any manual API calls.

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Add network_interfaces attribute, schema v0->v1 migration, VIP enrichment | 6a3c354 | server_resource.go, server_data_source.go, server_resource_test.go, server_data_source_test.go |
| 2 | Build validation and lint check | — (no code changes) | — |

## Implementation Details

### server_resource.go

- Schema version bumped from `0` to `1`
- `serverResourceModel` gains `NetworkInterfaces types.List` with `tfsdk:"network_interfaces"`
- `network_interfaces` attribute: `Computed: true`, `ElementType: types.StringType`, `UseStateForUnknown()` plan modifier
- `serverV0StateModel` struct for v0 state deserialization during upgrade
- `UpgradeState` returns a v0 upgrader that copies all v0 fields and sets `NetworkInterfaces = types.ListValueMust(types.StringType, []attr.Value{})`
- `mapServerToModel` accepts `*client.FlashBladeClient` and calls `enrichServerNetworkInterfaces`
- `enrichServerNetworkInterfaces`: calls `ListNetworkInterfaces`, filters by `ni.AttachedServers[*].Name == serverName`, uses warning diagnostic on error
- `ImportState` initializes `NetworkInterfaces` to empty list before calling `mapServerToModel`

### server_data_source.go

- `serverDataSourceModel` gains `NetworkInterfaces types.List` with `tfsdk:"network_interfaces"`
- `network_interfaces` attribute added to schema: `Computed: true`
- `enrichDataSourceNetworkInterfaces` function: same filter logic as resource enrichment
- `Read` method calls enrichment after populating DNS

### Tests

- `TestUnit_Server_StateUpgradeV0ToV1`: confirms upgrade from v0 state without `network_interfaces` produces empty list (not null)
- `TestUnit_Server_VIPEnrichment`: creates server, seeds VIPs with attached_servers, verifies only matching VIP name appears
- `TestUnit_Server_VIPEnrichment_Read`: same for Read path
- `TestUnit_Server_NoVIPs`: confirms empty list (not null) when no VIPs match
- `TestUnit_Server_SchemaVersion`: confirms schema version is 1
- `TestUnit_ServerDataSource_VIPEnrichment`: data source Read with 2 attached VIPs + 1 unattached
- All existing tests updated to include `network_interfaces` in `buildServerType()` / `buildServerDSType()` / `nullServerConfig()` / `nullServerDSConfig()`
- All tests register `RegisterNetworkInterfaceHandlers` to enable enrichment

## Verification Results

```
go test ./internal/provider/ -run "TestUnit_Server" -v -count=1
16 tests passed

go build ./...         # SUCCESS
go vet ./...           # no issues
go test ./internal/provider/ -count=1
417 tests passed
```

## Decisions Made

1. **StateUpgrader sets empty list:** Old state without `network_interfaces` gets `[]` not `null` — consistent with the "no drift" decision for this attribute.

2. **Warning on enrichment failure:** If `ListNetworkInterfaces` fails, CRUD is not blocked. A warning diagnostic is added and `network_interfaces` is set to empty list. Rationale: VIP enrichment is discovery-only, not ownership — a transient API error should not prevent server management.

3. **Client-side filter:** API does not support `?attached_server=` filter param. All VIPs are listed and filtered client-side by `ni.AttachedServers[*].Name`. Confirmed in STATE.md Phase 30 blocker resolution.

4. **`UseStateForUnknown` on resource:** Prevents unnecessary diffs between plan and apply when VIP attachment hasn't changed.

## Deviations from Plan

None — plan executed exactly as written.

## Self-Check: PASSED

- `internal/provider/server_resource.go` — FOUND
- `internal/provider/server_data_source.go` — FOUND
- `internal/provider/server_resource_test.go` — FOUND
- `internal/provider/server_data_source_test.go` — FOUND
- Commit 6a3c354 — FOUND (`git log --oneline -3`)
- `TestUnit_Server_StateUpgradeV0ToV1` in test file — FOUND
- `network_interfaces` in both resource and data source — FOUND
- `Version: 1` in server_resource.go — FOUND
