# Quick Task 8: fix: skip quota_limit in object store account PATCH when unchanged

**Created:** 2026-04-01
**Mode:** quick

## Task 1: Add IsUnknown guards to Update method

**Files:** `internal/provider/object_store_account_resource.go`
**Action:** Add `!plan.QuotaLimit.IsUnknown()` and `!plan.HardLimitEnabled.IsUnknown()` guards to the PATCH field comparisons in Update, matching the pattern used in bucket_resource.go.
**Verify:** `make build && make test`
**Done:** Guards prevent sending zero-value quota_limit/hard_limit_enabled when plan values are unknown.
