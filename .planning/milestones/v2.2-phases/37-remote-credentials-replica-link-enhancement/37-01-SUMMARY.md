---
phase: 37-remote-credentials-replica-link-enhancement
plan: "01"
subsystem: remote-credentials
tags: [remote-credentials, bucket-replica-link, schema-migration, target-replication]
dependency_graph:
  requires: [36-02-SUMMARY.md]
  provides: [RC-01, RC-02, BRL-01]
  affects: [flashblade_object_store_remote_credentials, client.PostRemoteCredentials, mock-handler]
tech_stack:
  added: []
  patterns:
    - "Optional+Computed on remote_name (API-populated field)"
    - "TargetName preserved from plan/state like SecretAccessKey (not returned by GET)"
    - "Conditional query routing: ?target_names= vs ?remote_names= based on plan attribute"
    - "v0->v1 state upgrader using intermediate model struct (remoteCredentialsV0Model)"
key_files:
  created: []
  modified:
    - internal/client/remote_credentials.go
    - internal/client/remote_credentials_test.go
    - internal/testmock/handlers/remote_credentials.go
    - internal/provider/remote_credentials_resource.go
    - internal/provider/remote_credentials_resource_test.go
    - internal/provider/bucket_replica_link_resource_test.go
decisions:
  - "remote_name changed to Optional+Computed: API populates it from Remote.Name on POST/GET; target-backed creds get target name there too — acceptable"
  - "target_name preserved from plan/state like SecretAccessKey (same pattern — not returned by GET)"
  - "v0->v1 upgrader uses remoteCredentialsV0Model intermediate struct (same pattern as server_resource.go)"
  - "Mock handler uses ValidateQueryParams with all three params [names, remote_names, target_names] then enforces XOR logic manually"
metrics:
  duration: 388s
  completed_date: "2026-04-01"
  tasks: 2
  files: 6
---

# Phase 37 Plan 01: Remote Credentials Target Support + BRL Validation Summary

Extended `flashblade_object_store_remote_credentials` to accept either a target reference (`target_name`) or an array connection reference (`remote_name`), and validated the bucket replica link resource works end-to-end when remote credentials reference a target via S3 target replication.

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Client + mock handler — target_names support on PostRemoteCredentials | 3cdecc4 | remote_credentials.go, remote_credentials_test.go, handlers/remote_credentials.go, provider/remote_credentials_resource.go |
| 2 | Schema v1 + state upgrader + target-backed RC + BRL tests | 1b274d2 | provider/remote_credentials_resource.go, provider/remote_credentials_resource_test.go, provider/bucket_replica_link_resource_test.go |

## Decisions Made

- **remote_name becomes Optional+Computed**: API always returns `Remote.Name` (whether it's a remote array or target name). This means existing configs with `remote_name = "array-name"` work unchanged; target-backed creds get the target name in `remote_name` on API response — acceptable and consistent.
- **target_name preservation pattern**: Same as `SecretAccessKey` — preserved from plan in Create(), from state in Read(). API never returns this field.
- **v0->v1 upgrader**: Uses `remoteCredentialsV0Model` intermediate struct following the established pattern from `server_resource.go`. Sets `target_name = null` for all existing resources.
- **Mock handler XOR logic**: `ValidateQueryParams` accepts all three params but the handler enforces mutual exclusivity with explicit error messages after extraction.

## Deviations from Plan

None — plan executed exactly as written.

## Test Results

- 692 total tests pass (7 new: 3 client + 3 provider RC + 1 BRL)
- `make build` clean
- `make test` clean
- `make lint` clean (0 issues)

## Requirements Satisfied

- RC-01: `flashblade_object_store_remote_credentials` can be created with `target_name` set (TestUnit_RemoteCredentials_Create_WithTarget)
- RC-02: Existing configs with `remote_name` work unchanged; v0->v1 state upgrade migrates without data loss (TestUnit_RemoteCredentials_StateUpgrade_V0toV1)
- BRL-01: BRL create with `remote_credentials_name` referencing a target-backed credential succeeds (TestUnit_BRL_WithTargetCredentials)

## Self-Check: PASSED
