---
phase: quick-260408-kbr
plan: 01
subsystem: array-connections
tags: [resource, ephemeral, sensitive, no-importstate]
dependency_graph:
  requires: [internal/client/models_admin.go, internal/testmock/handlers/array_connection_key.go]
  provides: [flashblade_array_connection_key resource]
  affects: [internal/provider/provider.go]
tech_stack:
  added: []
  patterns: [singleton-mock-handler, plain-json-response, no-importstate-ephemeral, no-op-delete]
key_files:
  created:
    - internal/client/array_connection_key.go
    - internal/client/array_connection_key_test.go
    - internal/testmock/handlers/array_connection_key.go
    - internal/provider/array_connection_key_resource.go
    - internal/provider/array_connection_key_resource_test.go
    - examples/resources/flashblade_array_connection_key/resource.tf
  modified:
    - internal/client/models_admin.go
    - internal/provider/provider.go
decisions:
  - ArrayConnectionKey endpoint returns plain JSON object (not ListResponse wrapper) — use c.get/c.post directly
  - Delete is a no-op: keys expire automatically, no API call needed
  - No ImportState: key is ephemeral with no stable import identifier
  - id set to connection_key value (synthetic stable ID per creation)
  - UseStateForUnknown on all computed fields (connection_key, created, expires, id) — stable after POST
metrics:
  duration: ~15min
  completed: "2026-04-08T12:51:25Z"
  tasks: 2
  files: 8
---

# Quick 260408-kbr: Add flashblade_array_connection_key Resource

## One-liner

POST-only ephemeral key generator backed by `/array-connections/connection-key` with Sensitive `connection_key`, no ImportState, and no-op Delete.

## What Was Built

### Task 1: Model + Client + Mock Handler (63f86b2)

- Appended `ArrayConnectionKey` struct to `internal/client/models_admin.go`
- Created `GetArrayConnectionKey` and `PostArrayConnectionKey` client methods using low-level `c.get`/`c.post` (endpoint returns plain object, not ListResponse)
- Created singleton mock handler for `/api/2.22/array-connections/connection-key` — GET returns current key as plain JSON, POST generates synthetic key and overwrites current
- 3 client tests: `TestUnit_ArrayConnectionKey_Get`, `TestUnit_ArrayConnectionKey_Post`, `TestUnit_ArrayConnectionKey_Get_AfterPost`

### Task 2: Resource + Tests + Registration (99e1512)

- Created `flashblade_array_connection_key` resource: Create (POST), Read (GET), Delete (no-op)
- Only 2 interfaces: `resource.Resource` + `resource.ResourceWithConfigure` (no ImportState, no UpgradeState)
- `connection_key` is `Sensitive: true` + `Computed`; `UseStateForUnknown` on all computed fields
- Read: removes resource from state if API returns empty key (expired/reset)
- Registered `NewArrayConnectionKeyResource` in provider.go under Replication group
- 3 resource tests: `TestUnit_ArrayConnectionKeyResource_Lifecycle`, `TestUnit_ArrayConnectionKeyResource_DriftDetection`, `TestUnit_ArrayConnectionKeyResource_DeleteNoOp`
- HCL example at `examples/resources/flashblade_array_connection_key/resource.tf`

## Verification

- `make build`: 0 errors
- `make test`: 751 tests pass (745 baseline + 6 new)
- `make lint`: 0 issues

## Deviations from Plan

None — plan executed exactly as written.

## Known Stubs

None.

## Self-Check: PASSED

- `internal/client/array_connection_key.go`: exists
- `internal/client/array_connection_key_test.go`: exists
- `internal/testmock/handlers/array_connection_key.go`: exists
- `internal/provider/array_connection_key_resource.go`: exists
- `internal/provider/array_connection_key_resource_test.go`: exists
- `examples/resources/flashblade_array_connection_key/resource.tf`: exists
- Commits 63f86b2 and 99e1512: verified
