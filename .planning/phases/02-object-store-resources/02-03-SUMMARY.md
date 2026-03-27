---
phase: 02-object-store-resources
plan: 03
subsystem: object-store-access-keys
tags: [object-store, access-key, write-once-secret, tdd, immutable-resource]
dependency_graph:
  requires: [02-01]
  provides: [flashblade_object_store_access_key resource, flashblade_object_store_access_key data source]
  affects: [provider.go]
tech_stack:
  added: []
  patterns:
    - UseStateForUnknown for write-once secret preservation
    - Read-does-not-overwrite pattern for immutable fields
    - All-ForceNew resource (RequiresReplace on every attribute)
    - No ImportState by design (secret unavailable after creation)
key_files:
  created:
    - internal/client/object_store_access_keys.go
    - internal/testmock/handlers/object_store_access_keys.go
    - internal/provider/object_store_access_key_resource.go
    - internal/provider/object_store_access_key_data_source.go
    - internal/provider/object_store_access_key_resource_test.go
    - internal/provider/object_store_access_key_data_source_test.go
  modified:
    - internal/provider/provider.go
decisions:
  - "Access key resource has no ImportState — secret_access_key is unavailable after creation (confirmed in plan)"
  - "All access key attributes are RequiresReplace — immutable resource, no Update path"
  - "Read method does not set SecretAccessKey — leaves prior state value intact; UseStateForUnknown propagates it through plan cycles"
  - "Update method stub added to satisfy resource.Resource interface — should never be called since all fields are ForceNew"
  - "Data source object_store_account field derived from key.User.Name (extract account prefix before first /)"
metrics:
  duration: 265s
  completed_date: "2026-03-27"
  tasks_completed: 2
  files_changed: 7
---

# Phase 02 Plan 03: Object Store Access Key Summary

**One-liner:** Write-once access key resource with secret preservation via UseStateForUnknown + Read-does-not-overwrite pattern, no ImportState, all-ForceNew schema.

## Tasks Completed

| # | Task | Commit | Result |
|---|------|--------|--------|
| 1 | Access key client methods + mock handler | `0374295` | GetObjectStoreAccessKey, ListObjectStoreAccessKeys, PostObjectStoreAccessKey, DeleteObjectStoreAccessKey; mock strips secret on GET, generates on POST |
| 2 | Access key resource + data source + provider registration + tests | `7485294` (RED), `b353cce` (GREEN) | 6 tests pass, 65 total no regressions |

## Artifacts Produced

- `/internal/client/object_store_access_keys.go` — 4 client methods, no Update, no terraform imports
- `/internal/testmock/handlers/object_store_access_keys.go` — thread-safe mock, POST generates random key+secret, GET strips secret, validates account exists
- `/internal/provider/object_store_access_key_resource.go` — Create+Read+Delete resource, no Update callable, no ImportState
- `/internal/provider/object_store_access_key_data_source.go` — reads key by name, secret always empty
- `/internal/provider/provider.go` — NewAccessKeyResource + NewAccessKeyDataSource registered

## Key Design Decisions

1. **Secret preservation:** `secret_access_key` uses `UseStateForUnknown` in schema AND Read explicitly does NOT overwrite it. Both mechanisms work together to survive plan/apply/refresh cycles.

2. **No ImportState:** Per plan requirement. `resource.ResourceWithImportState` is intentionally not implemented. The `TestUnit_AccessKey_NoImport` test validates this at the interface level.

3. **Update stub:** `resource.Resource` interface requires an Update method signature. Added a stub that returns an error diagnostic — should never reach it in practice since all attributes are ForceNew.

4. **Mock account validation:** POST handler cross-references the account store to ensure the referenced account exists before generating keys. This prevents orphaned keys in tests.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Added Update stub to satisfy resource.Resource interface**
- **Found during:** Task 2 implementation (compile error)
- **Issue:** `resource.Resource` interface mandates an `Update` method signature even for immutable resources
- **Fix:** Added stub Update that returns diagnostic error — all attributes are RequiresReplace so framework will never call it
- **Files modified:** `internal/provider/object_store_access_key_resource.go`
- **Commit:** `b353cce`

## Self-Check: PASSED

- `internal/client/object_store_access_keys.go` — FOUND
- `internal/testmock/handlers/object_store_access_keys.go` — FOUND
- `internal/provider/object_store_access_key_resource.go` — FOUND
- `internal/provider/object_store_access_key_data_source.go` — FOUND
- Commits `0374295`, `7485294`, `b353cce` — FOUND
- `go test ./internal/... -count=1` — 65 passed
