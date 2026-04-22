---
phase: 58-release-pipeline-docs
plan: 04
type: execute
subsystem: pulumi-bridge
tags: [pulumi, testing, composite-id, examples, import]
dependency_graph:
  requires:
    - "58-01"
    - "58-02"
  provides:
    - "TEST-03"
    - "RELEASE-03"
  affects:
    - pulumi/provider/resources_test.go
    - pulumi/examples/*/Pulumi.yaml
tech-stack:
  added: []
  patterns:
    - "TestProviderInfo_ImportSyntax_* naming convention for grep-ability"
    - "Pulumi.yaml main: main.go for Go runtime ProgramTest readiness"
key-files:
  created: []
  modified:
    - pulumi/provider/resources_test.go
    - pulumi/examples/target-go/Pulumi.yaml
    - pulumi/examples/remote_credentials-go/Pulumi.yaml
    - pulumi/examples/bucket-go/Pulumi.yaml
decisions: []
metrics:
  duration_minutes: 12
  completed_date: "2026-04-22"
---

# Phase 58 Plan 04: Composite-ID Import Tests + Example Validation Summary

**One-liner:** Added 4 composite-ID import syntax validation tests and fixed Go example metadata for ProgramTest readiness.

## Tasks Completed

| # | Task | Commit | Files |
|---|------|--------|-------|
| 1 | Add import syntax validation tests to resources_test.go | `bdf37f2` | `pulumi/provider/resources_test.go` |
| 2 | Validate example Pulumi.yaml files for ProgramTest readiness | `768d265` | `pulumi/examples/target-go/Pulumi.yaml`, `pulumi/examples/remote_credentials-go/Pulumi.yaml`, `pulumi/examples/bucket-go/Pulumi.yaml` |

## Changes Made

### Task 1: Import Syntax Tests (TEST-03)

Appended 4 new tests to `pulumi/provider/resources_test.go`:

- `TestProviderInfo_ImportSyntax_ObjectStoreAccessPolicyRule` — validates `policyName/name` composite ID
- `TestProviderInfo_ImportSyntax_BucketAccessPolicyRule` — validates `bucketName/name` composite ID
- `TestProviderInfo_ImportSyntax_NetworkAccessPolicyRule` — validates `policyName/name` composite ID
- `TestProviderInfo_ImportSyntax_ManagementAccessPolicyDSRMembership` — validates `role/policy` composite ID with colon edge case

Each test:
- Reuses the existing `ComputeID` closure (no new logic)
- Documents the correct `pulumi import` command in a comment
- Validates the exact ID string format

Test count increased from 19 to 23 (all passing).

### Task 2: Example Pulumi.yaml Validation (RELEASE-03)

Verified all 6 example `Pulumi.yaml` files exist and are valid YAML.

Added `main: main.go` to the 3 Go examples (required by Go runtime for ProgramTest):
- `pulumi/examples/target-go/Pulumi.yaml`
- `pulumi/examples/remote_credentials-go/Pulumi.yaml`
- `pulumi/examples/bucket-go/Pulumi.yaml`

Python examples unchanged (default `__main__.py` is correct).

## Deviations from Plan

None — plan executed exactly as written.

## Auth Gates

None.

## Known Stubs

None.

## Self-Check: PASSED

- [x] `pulumi/provider/resources_test.go` modified — 4 new tests present
- [x] All 23 tests pass (`go test ./... -count=1`)
- [x] `bdf37f2` commit exists
- [x] `768d265` commit exists
- [x] All 6 example `Pulumi.yaml` files are valid YAML
- [x] Go examples include `main: main.go`
