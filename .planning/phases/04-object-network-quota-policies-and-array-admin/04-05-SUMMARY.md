---
phase: 04-object-network-quota-policies-and-array-admin
plan: "05"
subsystem: infra
tags: [terraform-plugin-framework, singleton, smtp, ntp, dns, alert-watchers, flashblade]

requires:
  - phase: 04-01
    provides: client layer — GetArrayDns/PatchArrayDns, GetArrayNtp/PatchArrayNtp, GetSmtpServer/PatchSmtpServer, AlertWatcher CRUD, testmock handlers

provides:
  - flashblade_array_dns singleton resource (Create=GET+PATCH/POST, Delete resets to defaults)
  - flashblade_array_ntp singleton resource (wraps /arrays ntp_servers field only)
  - flashblade_array_smtp composite singleton resource with nested alert_watchers set
  - Data sources: flashblade_array_dns, flashblade_array_ntp, flashblade_array_smtp
  - All three import with id="default"
  - Full test coverage: 16 tests for DNS/NTP/SMTP resources and data sources

affects:
  - phase-05 (end-to-end acceptance tests)

tech-stack:
  added: []
  patterns:
    - Singleton lifecycle — Create=GET+PATCH/POST fallback, Delete=PATCH-to-reset (not DELETE)
    - Composite singleton — SMTP resource owns alert_watchers nested set (not separate resources)
    - Alert watcher diff in Update — planMap vs stateMap by email key
    - ImportState initializes timeouts.Value with types.ObjectNull to avoid serialization errors
    - helpers.go — shared listToStrings/emptyStringList utilities for list conversion

key-files:
  created:
    - internal/provider/array_dns_resource.go
    - internal/provider/array_dns_data_source.go
    - internal/provider/array_dns_resource_test.go
    - internal/provider/array_ntp_resource.go
    - internal/provider/array_ntp_data_source.go
    - internal/provider/array_ntp_resource_test.go
    - internal/provider/array_smtp_resource.go
    - internal/provider/array_smtp_data_source.go
    - internal/provider/array_smtp_resource_test.go
    - internal/provider/helpers.go
  modified:
    - internal/provider/provider.go (resources + data sources registered)

key-decisions:
  - "SMTP alert_watchers nested within SMTP resource (not separate Terraform resources) per user decision — single resource manages composite lifecycle"
  - "DNS Create uses GET-first then PATCH-or-POST pattern — handles both fresh arrays (no DNS) and arrays with existing config"
  - "NTP PATCH sends only ntp_servers field via ArrayNtpPatch struct — structurally enforces non-interference with other array settings"
  - "Alert watcher email is the API name field — watcher identity is the email address, not a UUID"
  - "SMTP Delete resets to defaults (relay_host='', sender_domain='', encryption_mode='none') plus deletes all watchers — singleton pattern"

patterns-established:
  - "Singleton resource pattern: Create=GET+PATCH, Delete=PATCH-to-defaults, ImportState accepts 'default' as ID"
  - "Composite singleton: readIntoState helper fetches multiple API endpoints and merges into single model"
  - "Alert watcher diff: planMap/stateMap keyed by email enables O(n) add/remove/update reconciliation"

requirements-completed:
  - ADM-01
  - ADM-02
  - ADM-03
  - ADM-04
  - ADM-05

duration: 20min
completed: 2026-03-28
---

# Phase 04 Plan 05: Array Admin (DNS, NTP, SMTP) Summary

**Three singleton admin resources (DNS, NTP, SMTP) with composite alert watcher management, data sources, and full test coverage — 16 tests, 147 total suite green**

## Performance

- **Duration:** ~20 min (continuation from interrupted session)
- **Started:** 2026-03-28
- **Completed:** 2026-03-28
- **Tasks:** 3 (2 from prior session + test files completed)
- **Files modified:** 10 created + 1 modified (provider.go)

## Accomplishments

- DNS singleton resource with POST-if-404/PATCH-if-exists Create pattern and reset-on-delete
- NTP singleton resource wrapping /arrays ntp_servers field with strict struct isolation
- SMTP composite singleton owning nested alert_watchers with full add/remove/update diff logic
- All three data sources (read-only, no input required)
- 16 tests covering Create/Update/Delete/Import/DataSource for all three resources plus SMTP watcher removal

## Task Commits

1. **Tasks 1+2: DNS/NTP/SMTP resources and data sources** - `a34747b` (feat)
2. **Task 3: NTP and SMTP tests** - `1726ea6` (test)

## Files Created/Modified

- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/provider/array_dns_resource.go` — DNS singleton resource, POST fallback on 404
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/provider/array_dns_data_source.go` — DNS data source
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/provider/array_dns_resource_test.go` — DNS lifecycle tests
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/provider/array_ntp_resource.go` — NTP singleton wrapping /arrays
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/provider/array_ntp_data_source.go` — NTP data source
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/provider/array_ntp_resource_test.go` — NTP lifecycle tests
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/provider/array_smtp_resource.go` — SMTP composite singleton with watcher diff
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/provider/array_smtp_data_source.go` — SMTP data source with nested watchers
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/provider/array_smtp_resource_test.go` — SMTP + watcher lifecycle tests
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/provider/helpers.go` — listToStrings, emptyStringList shared utilities

## Decisions Made

- SMTP alert watchers managed within SMTP resource (not as separate Terraform resources) — single lifecycle per user's architectural decision
- DNS Create does GET first to detect existing config, then POST (404) or PATCH (found) — handles both fresh and pre-configured arrays
- NTP ArrayNtpPatch struct has only NtpServers field by design — prevents unintentional modification of other array settings
- Alert watcher email IS the API name — email is the unique identity key in the diff logic

## Deviations from Plan

None — plan executed exactly as written. The test files were the only missing pieces from the interrupted session; all resource implementations were already complete.

## Issues Encountered

None — the interruption left resource files complete, only test files were missing. Tests were created following the array_dns_resource_test.go pattern exactly.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- All Phase 04 resources complete (OAP, NAP, Quota, Array Admin DNS/NTP/SMTP)
- Full test suite at 147 tests, all green
- Ready for Phase 05 (acceptance tests / end-to-end)

---
*Phase: 04-object-network-quota-policies-and-array-admin*
*Completed: 2026-03-28*
