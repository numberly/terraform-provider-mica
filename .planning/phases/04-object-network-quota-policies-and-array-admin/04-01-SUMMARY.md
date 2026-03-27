---
phase: 04-object-network-quota-policies-and-array-admin
plan: "01"
subsystem: client-layer
tags: [models, client, mock, object-store-access-policy, network-access-policy, quota, dns, ntp, smtp, alert-watcher]
dependency_graph:
  requires: []
  provides:
    - ObjectStoreAccessPolicy client CRUD + rule CRUD
    - NetworkAccessPolicy singleton + rule CRUD
    - QuotaUser and QuotaGroup CRUD
    - ArrayDns, ArrayNtp, SmtpServer, AlertWatcher client methods
    - Mock HTTP handlers for all 4 resource families
  affects:
    - internal/client/models.go
    - internal/testmock/handlers/*
tech_stack:
  added: []
  patterns:
    - name-as-query-param convention (POST/PATCH/DELETE use ?names=)
    - policy_names+names dual query param for rule endpoints
    - file_system_names+uids/gids for quota endpoints
    - singleton pattern for DNS, NTP, SMTP (no POST/DELETE at resource level)
    - in-memory sync.Mutex store with composite keys for quotas
key_files:
  created:
    - internal/client/object_store_access_policies.go
    - internal/client/network_access_policies.go
    - internal/client/quotas.go
    - internal/client/array_admin.go
    - internal/testmock/handlers/object_store_access_policies.go
    - internal/testmock/handlers/network_access_policies.go
    - internal/testmock/handlers/quotas.go
    - internal/testmock/handlers/array_admin.go
  modified:
    - internal/client/models.go
decisions:
  - "OAP rule POST uses both policy_names= and names= query params (unlike NFS which server-assigns name)"
  - "NetworkAccessPolicy has no POST/DELETE at policy level — singletons only (GET+PATCH)"
  - "NAP mock pre-seeds a 'default' policy so tests can immediately add rules without creating the policy first"
  - "Quota keys are composite fileSystemName/uid or fileSystemName/gid strings — avoids nested map complexity"
  - "ArrayNtpPatch sends only ntp_servers field — prevents unintentional modification of other array settings"
  - "DNS/NTP/SMTP mock handlers treat singleton as pre-seeded — POST upserts rather than creating duplicate"
metrics:
  duration_seconds: 324
  completed_date: "2026-03-26"
  tasks_completed: 2
  files_created: 8
  files_modified: 1
---

# Phase 4 Plan 01: Phase 4 Client Layer Foundation Summary

**One-liner:** Full CRUD client layer for OAP/NAP/quota/array-admin with singleton, rule, and composite-key patterns plus in-memory mock handlers for all resource families.

## Tasks Completed

| Task | Name | Commit | Key Files |
|------|------|--------|-----------|
| 1 | Phase 4 model structs + OAP/NAP/quota client methods | 3f53f36 | models.go, object_store_access_policies.go, network_access_policies.go, quotas.go |
| 2 | Array admin client methods + all 4 mock handlers | 76ff28e | array_admin.go, handlers/object_store_access_policies.go, handlers/network_access_policies.go, handlers/quotas.go, handlers/array_admin.go |

## What Was Built

### Model Structs (models.go additions)

- **ObjectStoreAccessPolicy family:** GET/Post/Patch models + Rule GET/Post/Patch. Rule conditions use `json.RawMessage` for flexible IAM condition objects.
- **NetworkAccessPolicy family:** GET/Patch models (no Post — singletons) + Rule GET/Post/Patch.
- **Quota family:** QuotaUser and QuotaGroup GET/Post/Patch models. Filesystem and user/group are `*NamedReference` in GET, passed as query params on write.
- **Array admin family:** ArrayDns GET/Post/Patch, ArrayInfo (NTP-only view), ArrayNtpPatch (ntp_servers only), SmtpServer GET/Patch, AlertWatcher GET/Post/Patch.

### Client Methods

- `object_store_access_policies.go`: 10 methods — full policy CRUD + rule CRUD + member list for delete guard
- `network_access_policies.go`: 9 methods — policy GET+List+Patch (no POST/DELETE) + rule full CRUD + index-based lookup
- `quotas.go`: 12 methods — 6 for user quotas (Get/List/Post/Patch/Delete) + 6 for group quotas
- `array_admin.go`: 11 methods — DNS GET/Post/Patch, NTP GET/Patch, SMTP GET/Patch, AlertWatcher GET/Post/Patch/Delete

### Mock Handlers

All 4 handlers follow the established pattern: `sync.Mutex` in-memory store, `WriteJSONListResponse`/`WriteJSONError` helpers, raw map PATCH semantics.

Special behaviors:
- **OAP rules:** User provides rule name via `?names=` on POST (unlike NFS where server assigns it)
- **NAP:** Pre-seeded "default" singleton policy; no POST/DELETE at policy level
- **Quotas:** Composite key `fileSystemName/uid` and `fileSystemName/gid` for O(1) lookup
- **Array admin:** DNS/NTP/SMTP pre-seeded as singletons; AlertWatcher is full CRUD keyed by email

## Deviations from Plan

### Auto-fixed Issues

None.

### Scope Adjustments

**server.go not modified:** The plan mentioned adding Register calls to server.go, but examining the existing codebase revealed that `server.go` only provides auth endpoints (`/api/login`, `/api/api_version`). All resource handlers are registered per-test by calling `handlers.RegisterXxxHandlers(ms.Mux)` explicitly — this is the established pattern across all existing test files. No change to server.go was necessary; the mock handlers are functional as-is.

## Verification

```
go build ./...    -> Success
go test ./...     -> 101 passed in 5 packages (no regressions)
```

## Self-Check

Files created/exist:
- internal/client/object_store_access_policies.go: FOUND
- internal/client/network_access_policies.go: FOUND
- internal/client/quotas.go: FOUND
- internal/client/array_admin.go: FOUND
- internal/testmock/handlers/object_store_access_policies.go: FOUND
- internal/testmock/handlers/network_access_policies.go: FOUND
- internal/testmock/handlers/quotas.go: FOUND
- internal/testmock/handlers/array_admin.go: FOUND

Commits verified:
- 3f53f36: feat(04-01): add Phase 4 model structs and OAP/NAP/quota client CRUD methods
- 76ff28e: feat(04-01): add array admin client methods and all Phase 4 mock handlers

## Self-Check: PASSED
