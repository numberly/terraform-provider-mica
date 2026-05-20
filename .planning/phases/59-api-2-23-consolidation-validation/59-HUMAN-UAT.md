---
status: partial
phase: 59-api-2-23-consolidation-validation
source: [59-VERIFICATION.md]
started: 2026-05-20T00:00:00Z
updated: 2026-05-20T00:00:00Z
---

## Current Test

awaiting operator action

## Tests

### 1. VALID-04 — Live acceptance against par5 + pa7

expected: Run `terraform apply / plan-no-drift / destroy` cycle for each of the following on FlashBlade arrays `par5` AND `pa7`:
- `flashblade_workload` (new resource)
- `flashblade_resiliency_group` (new data source, read-only)
- `flashblade_resiliency_group_member` (new data source, read-only)
- All 6 migrated schemas (file_system, file_system_export, nfs_export_policy, smb_client_policy, smb_share_policy, qos_policy) — verify workload field reads cleanly and no drift on re-apply

Acceptance criteria:
- 2 × N matrix all PASS (no apply errors, no drift on second plan, clean destroy)
- Results recorded in `.planning/phases/59-api-2-23-consolidation-validation/59-04-ACCEPTANCE.md`
- HCL fixtures committed under `examples/acceptance/api-2-23/` (plan 59-04 task 1 is `type: auto` — can be re-invoked separately to generate them, or author manually)

result: [pending]

## Summary

total: 1
passed: 0
issues: 0
pending: 1
skipped: 0
blocked: 0

## Gaps

_None automated — single item awaits operator action with FlashBlade credentials._
