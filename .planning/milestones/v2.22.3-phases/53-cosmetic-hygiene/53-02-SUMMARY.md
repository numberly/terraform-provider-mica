---
phase: 53-cosmetic-hygiene
plan: 02
subsystem: testmock-handlers
tags: [conventions, mock-handlers, R-012]
requires:
  - CONVENTIONS.md §Mock Handlers (byName/nextID canonical pattern)
provides:
  - canonical byName field naming in qosPolicyStore
  - nextID-based ID generation in subnetStore and networkInterfaceStore
  - google/uuid import dropped from subnets.go and network_interfaces.go
affects:
  - internal/testmock/handlers/qos_policies.go
  - internal/testmock/handlers/subnets.go
  - internal/testmock/handlers/network_interfaces.go
tech-stack:
  added: []
  patterns:
    - "Mock store IDs: fmt.Sprintf(\"<resource>-%d\", s.nextID) under existing mutex"
    - "Canonical byName field naming for single-name-keyed stores"
key-files:
  created:
    - .planning/phases/53-cosmetic-hygiene/53-02-SUMMARY.md
  modified:
    - internal/testmock/handlers/qos_policies.go
    - internal/testmock/handlers/subnets.go
    - internal/testmock/handlers/network_interfaces.go
decisions:
  - "byID retained in subnets.go and network_interfaces.go — still used by handleGet no-filter branch and handleDelete (planner note confirmed)."
metrics:
  duration_seconds: ~180
  completed_date: 2026-04-20
  test_count_before: 778
  test_count_after: 778
  tests_added: 0
  commits: 1
---

# Phase 53 Plan 02: Mock handler hygiene — byName + nextID (R-012)

## One-liner

Rename `qosPolicyStore.policies` → `byName` and replace `uuid.New().String()` with a `nextID` counter in `subnetStore` and `networkInterfaceStore`, producing `subnet-N` / `nic-N` IDs and dropping the `google/uuid` import from both files.

## Files touched

| File | Change |
|------|--------|
| `internal/testmock/handlers/qos_policies.go` | Renamed struct field `policies` → `byName`; updated constructor + 10 internal references. `members` map unchanged. |
| `internal/testmock/handlers/subnets.go` | Added `nextID int` to store; replaced 2 `uuid.New().String()` calls with `fmt.Sprintf("subnet-%d", s.nextID)` after `s.nextID++`; dropped `github.com/google/uuid` import. |
| `internal/testmock/handlers/network_interfaces.go` | Added `nextID int` to store; replaced 2 `uuid.New().String()` calls with `fmt.Sprintf("nic-%d", s.nextID)` after `s.nextID++`; dropped `github.com/google/uuid` import. |

## Pre-flight grep outputs

- `rg 's\.policies\b' internal/ --type go` → matches scattered across 11 unrelated handler files (each file's own `policies` field in its own store); only `qos_policies.go` touched in scope.
- `rg 'subnetStore\b|networkInterfaceStore\b' internal/ --type go -g '!subnets.go' -g '!network_interfaces.go'` → **0 matches** (stores unexported, no external ref).
- `rg 'google/uuid' internal/testmock/handlers/subnets.go internal/testmock/handlers/network_interfaces.go` → 1 match each (both dropped).

## Test deltas

| Metric | Value |
|--------|-------|
| Before | 778 (top-level `Test*` count per GNUmakefile) |
| After | 778 |
| Added | 0 |
| Removed | 0 |

Note: GNUmakefile `TEST_BASELINE=752`; actual count 778 well above baseline. Plan frontmatter referenced 832 from SUMMARY 53-01 which used a different counting lens (all subtests); current make logic counts top-level `Test*` via `go test -list`.

## Verification

- `rg 's\.policies\b' internal/testmock/handlers/qos_policies.go` → 0 matches.
- `rg 'uuid\.New' internal/testmock/handlers/subnets.go internal/testmock/handlers/network_interfaces.go` → 0 matches.
- `rg 'google/uuid' internal/testmock/handlers/subnets.go internal/testmock/handlers/network_interfaces.go` → 0 matches.
- `rg 'nextID\s+int' internal/testmock/handlers/subnets.go internal/testmock/handlers/network_interfaces.go` → 1 match each.
- `rg '\.byID' internal/testmock/handlers/subnets.go` → 4 matches (retained).
- `rg '\.byID' internal/testmock/handlers/network_interfaces.go` → 4 matches (retained).
- `go build ./...` → clean.
- `make test` → all packages `ok`, count 778 ≥ baseline 752.
- `make lint` → `0 issues.`

## Commit

- `f153ee2 fix(53-02): mock handler hygiene — byName + nextID pattern`

No `Co-Authored-By` trailer. `--no-verify` used per project policy.

## Deviations from Plan

None. Plan executed exactly as written.

## Known Stubs

None.

## Self-Check: PASSED

- qos_policies.go: `byName` field present; no `s.policies` remain (verified via rg).
- subnets.go: `nextID` field present; `fmt.Sprintf("subnet-%d"...)` present; no `uuid.New` or `google/uuid`.
- network_interfaces.go: `nextID` field present; `fmt.Sprintf("nic-%d"...)` present; no `uuid.New` or `google/uuid`.
- `byID` retained in both files (4 refs each).
- Commit `f153ee2` exists on current branch.
- `make test` and `make lint` both clean.
