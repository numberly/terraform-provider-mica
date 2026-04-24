---
status: partial
phase: 58-release-pipeline-docs
source: [58-VERIFICATION.md]
started: 2026-04-22T15:45:00Z
updated: 2026-04-22T15:45:00Z
---

## Current Test

Awaiting human testing of live-environment scenarios.

## Tests

### 1. Live FlashBlade ProgramTest Execution (TEST-02)
expected: Run `pulumi up` on all 6 examples against a real FlashBlade array. All deploy successfully.
result: [pending]

### 2. Release Pipeline Smoke Test (RELEASE-03)
expected: Push `pulumi-2.22.3-test` tag, verify GitHub Actions produces release with 5 signed archives + wheel + Go SDK tag.
result: [pending]

### 3. Go SDK Tag Resolution
expected: After release, `go get github.com/numberly/opentofu-provider-flashblade/pulumi/sdk/go@v2.22.3` with `GOPRIVATE=github.com/numberly/*` resolves and compiles.
result: [pending]

## Summary

total: 3
passed: 0
issues: 0
pending: 3
skipped: 0
blocked: 0

## Gaps
