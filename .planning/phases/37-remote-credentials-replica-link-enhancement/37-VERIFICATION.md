---
phase: 37-remote-credentials-replica-link-enhancement
verified: 2026-04-01T00:00:00Z
status: passed
score: 4/4 must-haves verified
re_verification: false
gaps: []
human_verification: []
---

# Phase 37: Remote Credentials Target Support + BRL Validation — Verification Report

**Phase Goal:** Operators can create remote credentials referencing a target (not just an array connection), and existing bucket replica links work end-to-end against external S3 endpoints with no regression on array-to-array replication
**Verified:** 2026-04-01
**Status:** passed
**Re-verification:** No — initial verification

---

## Goal Achievement

### Observable Truths

| #  | Truth                                                                                                                                         | Status     | Evidence                                                                                                                          |
|----|-----------------------------------------------------------------------------------------------------------------------------------------------|------------|-----------------------------------------------------------------------------------------------------------------------------------|
| 1  | Operator can create `flashblade_object_store_remote_credentials` with `target_name` set and `remote_name` omitted                            | VERIFIED   | `Create()` routes to `?target_names=` when `TargetName` is non-empty; `target_name` preserved from plan; `TestUnit_RemoteCredentials_Create_WithTarget` passes |
| 2  | Operator can create `flashblade_object_store_remote_credentials` with `remote_name` set and `target_name` omitted — existing configs unchanged | VERIFIED   | `remote_name` is Optional+Computed+RequiresReplace; `Create()` falls through to `?remote_names=` when TargetName is null; `TestUnit_RemoteCredentials_Post_WithRemote` and full lifecycle tests pass |
| 3  | Bucket replica link resource creates successfully when `remote_credentials_name` references credentials tied to a target                     | VERIFIED   | `TestUnit_BRL_WithTargetCredentials` seeds a target-backed credential and creates a BRL — passes without error, state has `remote_credentials_name=target-cred` |
| 4  | `make test` passes with no regressions                                                                                                        | VERIFIED   | 692 tests pass across 5 packages (7 new tests: 3 client-level + 3 provider RC + 1 BRL); `make build` and `make lint` both clean   |

**Score:** 4/4 truths verified

---

## Required Artifacts

| Artifact                                                  | Provides                                                                                                    | Status     | Details                                                                                   |
|-----------------------------------------------------------|-------------------------------------------------------------------------------------------------------------|------------|-------------------------------------------------------------------------------------------|
| `internal/client/remote_credentials.go`                  | `PostRemoteCredentials` updated with `targetName` param; routes to `?target_names=` or `?remote_names=`     | VERIFIED   | Lines 42-57: signature has `remoteName string, targetName string`; conditional routing confirmed |
| `internal/provider/remote_credentials_resource.go`        | Schema v1 with `target_name` Optional+RequiresReplace; `remote_name` Optional+Computed+RequiresReplace; v0->v1 upgrader | VERIFIED   | `Version: 1` at line 67; both attributes present; `UpgradeState` returns upgrader at key 0 with `remoteCredentialsV0Model` |
| `internal/testmock/handlers/remote_credentials.go`       | POST handler accepts `?remote_names=` XOR `?target_names=`; rejects both/neither with 400                  | VERIFIED   | Lines 94-114: XOR logic enforced; `refName` set to whichever param was non-empty          |

---

## Key Link Verification

| From                                                         | To                                     | Via                                                       | Status  | Details                                                                                     |
|--------------------------------------------------------------|----------------------------------------|-----------------------------------------------------------|---------|---------------------------------------------------------------------------------------------|
| `provider/remote_credentials_resource.go Create()`          | `client.PostRemoteCredentials()`       | `targetName` from `data.TargetName` when set, else empty  | WIRED   | Lines 215-224: conditional extraction of `targetName`/`remoteName`, passed as 3rd/4th args |
| `client/remote_credentials.go PostRemoteCredentials()`      | `/object-store-remote-credentials`     | `?target_names=` if non-empty, else `?remote_names=`      | WIRED   | Lines 44-47: conditional path construction confirmed                                        |
| `provider/remote_credentials_resource.go mapRemoteCredentialsToModel()` | `remoteCredentialsModel.RemoteName` | `cred.Remote.Name` → `data.RemoteName`; `TargetName` preserved from plan | WIRED   | Line 391: `data.RemoteName = types.StringValue(cred.Remote.Name)`; `data.TargetName` preserved at lines 231-238 |

---

## Requirements Coverage

| Requirement | Source Plan | Description                                                                                  | Status    | Evidence                                                                                            |
|-------------|-------------|----------------------------------------------------------------------------------------------|-----------|-----------------------------------------------------------------------------------------------------|
| RC-01       | 37-01-PLAN  | Operator can create remote credentials referencing a target via `flashblade_object_store_remote_credentials` | SATISFIED | `target_name` schema attribute + `Create()` routing + `TestUnit_RemoteCredentials_Create_WithTarget` |
| RC-02       | 37-01-PLAN  | Existing remote credentials functionality for array connections is not broken                | SATISFIED | `remote_name` Optional+Computed preserves backward compat; v0->v1 upgrader sets `target_name=null`; `TestUnit_RemoteCredentials_StateUpgrade_V0toV1` + full lifecycle suite pass |
| BRL-01      | 37-01-PLAN  | Operator can create a bucket replica link using remote credentials that reference a target   | SATISFIED | `TestUnit_BRL_WithTargetCredentials` seeds target-backed credential, creates BRL — passes           |

No orphaned requirements. All three IDs declared in the plan and checked in REQUIREMENTS.md.

---

## Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| — | — | — | — | None found |

No TODOs, FIXMEs, placeholder returns, or stub implementations detected in the modified files.

---

## Human Verification Required

None. All observable behaviors are covered by unit tests using the mock server. Acceptance tests against a real FlashBlade array are out of scope for this phase.

---

## Gaps Summary

No gaps. All must-haves verified against the actual codebase:

- `PostRemoteCredentials` signature updated (4 params), conditional routing implemented and exercised by 3 client-level tests.
- Schema v1 implemented with `target_name` (Optional+RequiresReplace) and `remote_name` changed to Optional+Computed+RequiresReplace.
- State upgrader v0→v1 present and passing `TestUnit_RemoteCredentials_StateUpgrade_V0toV1`.
- `Create()` wiring verified: routes to `?target_names=` when `TargetName` is set, falls back to `?remote_names=`.
- `mapRemoteCredentialsToModel` does not set `TargetName` (correct — preserved from plan/state).
- Mock handler enforces XOR on `remote_names` / `target_names`.
- BRL resource unmodified but tested via `TestUnit_BRL_WithTargetCredentials` — passes.
- 692 total tests pass; `make build` clean; `make lint` clean (0 issues).

---

_Verified: 2026-04-01_
_Verifier: Claude (gsd-verifier)_
