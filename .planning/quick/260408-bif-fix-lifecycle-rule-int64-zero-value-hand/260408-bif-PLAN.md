# Quick Task 260408-bif: Fix lifecycle rule int64 zero-value handling

**Created:** 2026-04-08
**Status:** Ready for execution

## Problem

The 4 duration fields in lifecycle rules use plain `int64` making `0` indistinguishable from "not set":
1. POST/PATCH `omitempty` silently drops `0` values
2. `mapLifecycleRuleToModel` uses `!= 0` sentinel — masks drift when API returns `0`

## Solution

Change `int64` → `*int64` for duration fields in GET and POST structs. PATCH already uses `*int64`.

## Tasks

### Task 1: Model structs + client layer

**Files:** `internal/client/models_storage.go`
**Action:** Change 4 fields to `*int64` in `LifecycleRule` (GET) and `LifecycleRulePost` structs

### Task 2: Resource + data source mapping

**Files:** `internal/provider/lifecycle_rule_resource.go`, `internal/provider/lifecycle_rule_data_source.go`
**Action:**
- `mapLifecycleRuleToModel`: `!= 0` → `!= nil`, dereference with `*`
- Data source Read: same fix
- Resource Create: pass `&v` for pointer fields

### Task 3: Mock handler

**Files:** `internal/testmock/handlers/lifecycle_rules.go`
**Action:** handlePost copies `*int64` directly (same type now). handlePatch: assign `&val`.

### Task 4: Tests

**Files:** `internal/client/lifecycle_rules_test.go`, `internal/provider/lifecycle_rule_resource_test.go`, `internal/provider/lifecycle_rule_data_source_test.go`
**Action:** Update seeds and assertions to use `*int64` (inline `&val` or helper).

### Task 5: Validate

**Action:** `make test && make lint`
