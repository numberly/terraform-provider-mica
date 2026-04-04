---
phase: quick-260404-u4o
plan: 01
subsystem: client
tags: [refactor, generics, dry, client]
dependency_graph:
  requires: []
  provides: [postOne generic helper, patchOne generic helper]
  affects: [all 30 client CRUD files]
tech_stack:
  added: []
  patterns: [postOne[TBody, TResp], patchOne[TBody, TResp], generic helpers after getOneByName]
key_files:
  created: []
  modified:
    - internal/client/client.go
    - internal/client/targets.go
    - internal/client/certificates.go
    - internal/client/subnets.go
    - internal/client/servers.go
    - internal/client/syslog_servers.go
    - internal/client/buckets.go
    - internal/client/filesystems.go
    - internal/client/network_interfaces.go
    - internal/client/object_store_accounts.go
    - internal/client/object_store_virtual_hosts.go
    - internal/client/file_system_exports.go
    - internal/client/object_store_account_exports.go
    - internal/client/bucket_access_policies.go
    - internal/client/bucket_audit_filters.go
    - internal/client/snapshot_policies.go
    - internal/client/nfs_export_policies.go
    - internal/client/s3_export_policies.go
    - internal/client/smb_share_policies.go
    - internal/client/smb_client_policies.go
    - internal/client/network_access_policies.go
    - internal/client/object_store_access_policies.go
    - internal/client/qos_policies.go
    - internal/client/tls_policies.go
    - internal/client/remote_credentials.go
    - internal/client/bucket_replica_links.go
    - internal/client/object_store_access_keys.go
    - internal/client/object_store_users.go
    - internal/client/lifecycle_rules.go
    - internal/client/quotas.go
    - internal/client/array_admin.go
decisions:
  - "postOne/patchOne use direct error pass-through (no wrapping) — consistent with getOneByName"
  - "tls_policies had non-standard fmt.Errorf wrapping in Post/Patch errors — normalized to direct pass-through"
  - "object_store_users nil body: postOne[any, ObjectStoreUser](c, ctx, path, nil, label)"
  - "qos/tls members struct{}{} body: postOne[struct{}, T](c, ctx, path, struct{}{}, label)"
  - "14 files had fmt import removed (no longer needed after migration)"
metrics:
  duration: 820s
  completed_date: "2026-04-04"
  tasks_completed: 3
  files_modified: 31
---

# Phase quick-260404-u4o Plan 01: Extract postOne/patchOne Generics Summary

**One-liner:** DRY refactor adding postOne[TBody,TResp] and patchOne[TBody,TResp] generics, migrating all 60+ Post/Patch functions across 30 client files from 6-line boilerplate to 1-line delegate calls.

## What Was Built

### Task 1: postOne and patchOne generics in client.go

Two generic helpers added immediately after `getOneByName` in `internal/client/client.go`:

```go
func postOne[TBody any, TResp any](c *FlashBladeClient, ctx context.Context, path string, body TBody, label string) (*TResp, error)
func patchOne[TBody any, TResp any](c *FlashBladeClient, ctx context.Context, path string, body TBody, label string) (*TResp, error)
```

Both follow the same ListResponse unwrap pattern as `getOneByName`: call HTTP method, check `len(resp.Items) == 0`, return `&resp.Items[0]`.

### Task 2: Migrate all 30 client files

Every Post and Patch function across 30 client files now delegates to `postOne`/`patchOne`. The 6-line boilerplate:

```go
var resp ListResponse[XxxType]
if err := c.post(ctx, path, body, &resp); err != nil {
    return nil, err
}
if len(resp.Items) == 0 {
    return nil, fmt.Errorf("PostXxx: empty response from server")
}
return &resp.Items[0], nil
```

is replaced by a single line:

```go
return postOne[XxxTypePost, XxxType](c, ctx, path, body, "PostXxx")
```

Path construction logic is preserved unchanged in all cases (conditional params, composite names, confirmDate flags, etc.).

### Task 3: Verification

- 716 tests pass (unchanged baseline)
- 0 lint issues
- No `var resp ListResponse` patterns remain inside Post or Patch function bodies

## Commits

| Task | Hash | Message |
|------|------|---------|
| 1 | 6c3f39d | refactor(quick-260404-u4o-01): add postOne and patchOne generic helpers to client.go |
| 2 | c9148d1 | refactor(quick-260404-u4o-01): migrate all Post/Patch client functions to postOne/patchOne |

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] tls_policies.go used non-standard error wrapping**

- **Found during:** Task 2
- **Issue:** `PostTlsPolicy` and `PatchTlsPolicy` wrapped errors as `fmt.Errorf("PostTlsPolicy: %w", err)` instead of direct pass-through. This was inconsistent with all other client files.
- **Fix:** Migrated to `postOne`/`patchOne` which use direct error pass-through (consistent with getOneByName and the rest of the codebase).
- **Files modified:** `internal/client/tls_policies.go`
- **Commit:** c9148d1

## Known Stubs

None.

## Self-Check: PASSED

- `internal/client/client.go` — FOUND
- commit 6c3f39d — FOUND (`git log --oneline | grep 6c3f39d`)
- commit c9148d1 — FOUND (`git log --oneline | grep c9148d1`)
- 716 tests pass — VERIFIED
- 0 lint issues — VERIFIED
- No `var resp ListResponse` in Post/Patch bodies — VERIFIED
