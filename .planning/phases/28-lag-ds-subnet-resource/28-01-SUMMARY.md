---
phase: 28-lag-ds-subnet-resource
plan: "01"
subsystem: client-layer
tags: [models, client, mock-handlers, subnet, lag, network]
dependency_graph:
  requires: []
  provides:
    - "client.Subnet, client.SubnetPost, client.SubnetPatch structs"
    - "client.LinkAggregationGroup struct"
    - "FlashBladeClient.GetSubnet, PostSubnet, PatchSubnet, DeleteSubnet, ListSubnets"
    - "FlashBladeClient.GetLinkAggregationGroup, ListLinkAggregationGroups"
    - "handlers.RegisterSubnetHandlers (full CRUD mock with ?names= POST)"
    - "handlers.RegisterLinkAggregationGroupHandlers (GET-only mock with Seed)"
  affects: []
tech_stack:
  added: []
  patterns:
    - "?names= query param at POST for user-provided subnet names (not ?create_ds=)"
    - "SubnetPatch uses *int64 for MTU/VLAN to handle zero-value (VLAN=0 = untagged)"
    - "raw map[string]json.RawMessage for true PATCH semantics in mock handler"
    - "Seed() method on read-only mock stores for test data injection"
key_files:
  created:
    - internal/client/models_network.go
    - internal/client/subnets.go
    - internal/client/link_aggregation_groups.go
    - internal/testmock/handlers/subnets.go
    - internal/testmock/handlers/link_aggregation_groups.go
  modified: []
decisions:
  - "SubnetPost does not include Name field — passed via ?names= query param only"
  - "LAG client is GET-only — no POST/PATCH/DELETE (hardware-managed)"
  - "SubnetPatch uses *int64 for MTU and VLAN for zero-value safety"
  - "LAG mock requires explicit Seed() before any GET request in tests"
metrics:
  duration: "~15 minutes"
  completed_date: "2026-03-31"
  tasks_completed: 2
  tasks_total: 2
  files_created: 5
  files_modified: 0
---

# Phase 28 Plan 01: Client Layer (Models + CRUD + Mock Handlers) Summary

**One-liner:** Subnet and LAG client layer with separate Post/Patch/Get structs, ?names=-based CRUD methods, and full CRUD + read-only mock handlers for integration test infrastructure.

## Objective

Establish the data layer and test infrastructure for Phase 28. Downstream plan (28-02) uses these to build the Terraform resource, data sources, and integration tests.

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Create network model structs and client CRUD methods | da9be4b | models_network.go, subnets.go, link_aggregation_groups.go |
| 2 | Create subnet and LAG mock handlers | d6c276f | handlers/subnets.go, handlers/link_aggregation_groups.go |

## What Was Built

### Task 1 — Client Layer

**`internal/client/models_network.go`:** Three subnet structs following the established separate-struct pattern:
- `Subnet` (full GET response, all fields)
- `SubnetPost` (5 writable fields only — no Name, no Enabled, no Interfaces, no Services)
- `SubnetPatch` (pointer types for all 5 writable fields; `*int64` for MTU/VLAN ensures VLAN=0 is serializable)
- `LinkAggregationGroup` (read-only GET struct, all fields)

**`internal/client/subnets.go`:** 5 CRUD methods, all using `?names=` query param:
- `GetSubnet` via `getOneByName[Subnet]`
- `ListSubnets` via generic `c.get`
- `PostSubnet` — name passed via `?names=`, not request body
- `PatchSubnet` — partial updates via pointer-type struct
- `DeleteSubnet` — no cascade parameter needed

**`internal/client/link_aggregation_groups.go`:** 2 GET-only methods:
- `GetLinkAggregationGroup` via `getOneByName[LinkAggregationGroup]`
- `ListLinkAggregationGroups`

### Task 2 — Mock Handlers

**`internal/testmock/handlers/subnets.go`:** Full CRUD mock:
- `RegisterSubnetHandlers` registers at `/api/2.22/subnets`
- `AddSubnet(name, prefix, lagName)` seeder with sane defaults (Enabled=true, MTU=1500, VLAN=0)
- POST reads `?names=` query param (not `?create_ds=`)
- PATCH uses `map[string]json.RawMessage` for true field-level semantics
- DELETE cleans both `byName` and `byID` maps
- Thread-safe via `sync.Mutex`

**`internal/testmock/handlers/link_aggregation_groups.go`:** Read-only mock:
- `RegisterLinkAggregationGroupHandlers` registers at `/api/2.22/link-aggregation-groups`
- `Seed(*client.LinkAggregationGroup)` for test data injection (required before any GET)
- Non-GET methods return 405 Method Not Allowed

## Verification

```
go build ./internal/client/ && go build ./internal/testmock/... && echo "Plan 01 verified"
```

Result: PASSED — all 5 files compile cleanly.

## Deviations from Plan

None — plan executed exactly as written.

## Self-Check

- [x] `internal/client/models_network.go` exists and contains 3 subnet structs + LAG struct
- [x] `internal/client/subnets.go` exists with 5 exported methods
- [x] `internal/client/link_aggregation_groups.go` exists with 2 exported GET-only methods
- [x] `internal/testmock/handlers/subnets.go` exists with full CRUD + AddSubnet seeder
- [x] `internal/testmock/handlers/link_aggregation_groups.go` exists with Seed + GET-only
- [x] Commit da9be4b: feat(28-01): add Subnet/LAG model structs and client CRUD methods
- [x] Commit d6c276f: feat(28-01): add subnet and LAG mock handlers for integration tests

## Self-Check: PASSED
