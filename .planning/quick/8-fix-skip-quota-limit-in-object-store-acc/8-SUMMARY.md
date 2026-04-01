# Quick Task 8: fix: skip quota_limit in object store account PATCH when unchanged

**Date:** 2026-04-01
**Commit:** 8612b5b

## Problem

When patching an object store account (e.g., changing only `hard_limit_enabled`), the provider sent `quota_limit="0"` to the FlashBlade API because the `Update` method lacked an `IsUnknown()` guard. The API rejects `quota_limit=0` with HTTP 400: "must be between 1 byte and 9.22 exabytes".

## Root Cause

In `object_store_account_resource.go:291`, the comparison `!plan.QuotaLimit.Equal(state.QuotaLimit)` returns `true` when the plan value is Unknown (Computed attribute). `ValueInt64()` then returns `0`, which gets serialized as `"0"` in the PATCH body.

The equivalent code in `bucket_resource.go:465` correctly guards with `!plan.QuotaLimit.IsUnknown()`.

## Fix

Added `IsUnknown()` guards to both `QuotaLimit` and `HardLimitEnabled` comparisons in the Update method, matching the pattern from bucket_resource.go.

## Files Changed

- `internal/provider/object_store_account_resource.go` — 2 lines changed (added IsUnknown guards)

## Verification

- `make build` — passes
- `make test` — all tests pass
