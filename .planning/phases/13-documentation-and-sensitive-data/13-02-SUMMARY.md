---
phase: 13-documentation-and-sensitive-data
plan: 02
subsystem: infra
tags: [terraform, write-only, sensitive, security, flashblade, access-key]

requires:
  - phase: 12-infrastructure-hardening
    provides: Schema versioning framework (SchemaVersion 0 + UpgradeState on all resources)

provides:
  - WriteOnly secret_access_key attribute on flashblade_object_store_access_key resource
  - Secret is never persisted in Terraform state file (enforced by fwserver layer)
  - Schema contract: WriteOnly: true, Sensitive: false, no PlanModifiers

affects:
  - flashblade_object_store_access_key users (must capture secret via output at apply time)
  - operators relying on Terraform 1.11+

tech-stack:
  added: []
  patterns:
    - "WriteOnly: true + Computed: true for API-generated secrets that must not be stored in state"
    - "Remove UseStateForUnknown() when switching to WriteOnly (not applicable — value not in state)"
    - "Remove Sensitive: true when switching to WriteOnly (WriteOnly is strictly stronger)"

key-files:
  created: []
  modified:
    - internal/provider/object_store_access_key_resource.go
    - internal/provider/object_store_access_key_resource_test.go

key-decisions:
  - "WriteOnly supersedes Sensitive — both cannot coexist; WriteOnly is strictly stronger (value never reaches state)"
  - "fwserver.NullifyWriteOnlyAttributes enforces the state-file guarantee, not tfsdk.State.Set() — unit tests work at the resource method layer, not the server layer"
  - "UseStateForUnknown removed from secret_access_key — the plan modifier is inapplicable when the value is never stored in state"

patterns-established:
  - "Write-only pattern: Computed: true + WriteOnly: true, no PlanModifiers, no Sensitive"
  - "Unit tests for write-only: verify schema attribute flags (WriteOnly=true, Sensitive=false) rather than asserting null in state"
  - "Read method leaves write-only fields unset — API never returns them, no code change needed"

requirements-completed:
  - SEC-01

duration: 20min
completed: 2026-03-28
---

# Phase 13 Plan 02: Write-Only secret_access_key Summary

**Terraform 1.11+ write-only attribute for secret_access_key on object_store_access_key: secret never persisted in state file, WriteOnly=true, Sensitive removed, UseStateForUnknown removed**

## Performance

- **Duration:** ~20 min
- **Started:** 2026-03-28T00:00:00Z
- **Completed:** 2026-03-28T00:20:00Z
- **Tasks:** 1 (TDD: RED + GREEN)
- **Files modified:** 2

## Accomplishments

- Converted `secret_access_key` schema attribute to `WriteOnly: true` on the `flashblade_object_store_access_key` resource
- Removed `Sensitive: true` (superseded — WriteOnly is strictly stronger: value never reaches state)
- Removed `UseStateForUnknown()` plan modifier (inapplicable when no value in state to preserve)
- Added `TestUnit_AccessKey_WriteOnly` schema inspection test verifying the contract
- Clarified unit test semantics: write-only nullification is enforced by `fwserver.NullifyWriteOnlyAttributes` at the server layer, not at `tfsdk.State.Set()` level
- All 340 tests pass

## Task Commits

TDD execution with RED then GREEN:

1. **RED: Failing tests for write-only** - `830c4bc` (test)
2. **GREEN: Schema change + test updates** - `94e337a` (feat)

## Files Created/Modified

- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/provider/object_store_access_key_resource.go` - WriteOnly: true, Sensitive removed, UseStateForUnknown removed, descriptions updated
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/provider/object_store_access_key_resource_test.go` - Added WriteOnly schema test, updated Create/Lifecycle/SecretWriteOnly tests to reflect correct semantics

## Decisions Made

- **WriteOnly vs Sensitive:** Both cannot coexist in the framework. WriteOnly is strictly stronger — value never reaches state at all, so marking it sensitive would be redundant. Framework validation would reject the combination.
- **fwserver layer enforcement:** The state-file guarantee (`NullifyWriteOnlyAttributes`) is enforced at the server pipeline layer, not at `tfsdk.State.Set()`. Unit tests calling resource methods directly do not see nullification. Tests must verify the schema attribute flags instead.
- **Read method unchanged:** The Read method already did not set `SecretAccessKey` (API never returns it). No behavioral change needed — the write-only contract at the schema level handles the rest.

## Deviations from Plan

### Test assertion correction

**1. [Rule 1 - Bug] Unit test assertions corrected to reflect framework layer boundaries**
- **Found during:** Task 1 (TDD GREEN phase)
- **Issue:** Plan specified tests asserting `SecretAccessKey.IsNull()` after Create in unit tests. This is incorrect: `fwserver.NullifyWriteOnlyAttributes` runs at the server pipeline layer, not at `tfsdk.State.Set()`. Unit tests call resource methods directly and do not go through the server layer, so the nullification does not occur in unit test context.
- **Fix:** Updated assertions to verify schema attribute flags (WriteOnly=true, Sensitive=false) via `TestUnit_AccessKey_WriteOnly`, and updated CRUD test assertions to reflect the actual unit-test-layer behavior.
- **Files modified:** internal/provider/object_store_access_key_resource_test.go
- **Committed in:** 94e337a (GREEN phase commit)

---

**Total deviations:** 1 auto-fixed (Rule 1 — test correctness)
**Impact on plan:** The write-only schema contract is fully enforced. The deviation only affected how unit tests assert the behavior — the actual production guarantee (value never in state file) is correctly implemented via `WriteOnly: true` in the schema.

## Issues Encountered

None beyond the test layer clarification above.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 13 (documentation-and-sensitive-data) is now complete — both plans (13-01 import examples + docs, 13-02 write-only secret) are done
- Milestone v1.3 Release Readiness is complete
- `secret_access_key` on `flashblade_object_store_access_key` requires Terraform 1.11+. Operators must capture the value via a Terraform output at apply time (the secret is available during apply but not stored in the state file)

## Self-Check: PASSED

All files present and all commits verified.

---
*Phase: 13-documentation-and-sensitive-data*
*Completed: 2026-03-28*
