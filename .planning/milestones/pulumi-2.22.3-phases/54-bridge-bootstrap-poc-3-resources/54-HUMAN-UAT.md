---
status: partial
phase: 54-bridge-bootstrap-poc-3-resources
source: [54-VERIFICATION.md]
started: 2026-04-22T14:30:00Z
updated: 2026-04-22T14:30:00Z
---

## Current Test

[awaiting human testing]

## Tests

### 1. Pulumi Plugin Install Resolution
expected: `pulumi plugin install resource flashblade v0.0.1 --server github://api.github.com/numberly` downloads and installs without error
result: [pending]

### 2. ComputeID Import Round-Trip
expected: `pulumi import flashblade:index:ObjectStoreAccessPolicyRule rule "my-policy/0"` captures policyName and name correctly; `pulumi refresh` shows no drift
result: [pending]

### 3. Soft-Delete Timeout Adequacy
expected: Destroying a bucket resource with Pulumi completes delete + eradication polling within inherited 30m TF timeout
result: [pending]

## Summary

total: 3
passed: 0
issues: 0
pending: 3
skipped: 0
blocked: 0

## Gaps
