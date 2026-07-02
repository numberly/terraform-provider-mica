# bug: `object_lock_config` is sent in `POST /buckets` and breaks bucket creation (HTTP 400)

## Summary

Creating a `flashblade_bucket` with `object_lock_config` set fails at apply time:

```
Error: Error creating bucket
  with flashblade_bucket.par5,
  FlashBlade API error (HTTP 400): Invalid body parameter: object_lock_enabled
```

`object_lock_config` is a **PATCH-only** field on the FlashBlade REST API (exactly like
`versioning` and `public_access_config`), but the provider includes it in the `POST /buckets`
create body. The FlashBlade API rejects the create request.

The bug is invisible in CI because the in-repo test mock **accepts** `object_lock_config` on
POST, unlike the real array. So tests pass, real applies fail.

Impact: any bucket created with `object_lock_config` set cannot be created at all. Buckets
**without** `object_lock_config` are unaffected (the field is omitted from POST when null).

- Provider: `numberly/mica` (observed on `2.22.6`)
- Affected resource: `flashblade_bucket`

## Reproduction

```hcl
resource "flashblade_bucket" "example" {
  account    = "my-account"
  name       = "object-lock-repro"
  versioning = "enabled"

  object_lock_config = {
    object_lock_enabled = true
  }
}
```

`terraform apply` → `HTTP 400: Invalid body parameter: object_lock_enabled`.

Conversely, the same resource **without** the `object_lock_config` block applies fine.

## Root cause

The create path puts `object_lock_config` into the POST body. The FlashBlade `POST /buckets`
endpoint does not accept `object_lock_config` (its sub-field `object_lock_enabled` is reported
as an invalid body parameter). The provider already handles two sibling PATCH-only fields
correctly — `versioning` and `public_access_config` — by skipping them on POST and applying
them via PATCH after creation. `object_lock_config` was left in the POST path by mistake.

### Evidence (exact locations)

1. **POST body includes object_lock_config** — `internal/provider/bucket_resource.go:339-341`
   ```go
   if cfg := extractObjectLockConfig(data.ObjectLockConfig); cfg != nil {
       post.ObjectLockConfig = cfg
   }
   // public_access_config is NOT valid on POST — skip   <-- note: object_lock should be skipped too
   ```

2. **The struct still carries it on POST** — `internal/client/models_storage.go:157-168`
   ```go
   // NOTE: versioning is NOT a valid POST parameter — use PATCH after creation.
   // NOTE: public_access_config is NOT valid on POST — PATCH only.
   type BucketPost struct {
       Account           NamedReference     `json:"account"`
       QuotaLimit        string             `json:"quota_limit,omitempty"`
       HardLimitEnabled  bool               `json:"hard_limit_enabled,omitempty"`
       RetentionLock     string             `json:"retention_lock,omitempty"`
       EradicationConfig *EradicationConfig `json:"eradication_config,omitempty"`
       ObjectLockConfig  *ObjectLockConfig  `json:"object_lock_config,omitempty"`  // <-- must be removed
   }
   ```

3. **The correct sibling pattern already exists** — `internal/provider/bucket_resource.go:350-360`
   ```go
   // Versioning is not a valid POST parameter — apply via PATCH after creation.
   if !data.Versioning.IsNull() && !data.Versioning.IsUnknown() {
       v := data.Versioning.ValueString()
       _, err := r.client.PatchBucket(ctx, bucket.ID, client.BucketPatch{Versioning: &v})
       ...
   }
   ```

4. **Update already does the right thing via PATCH** — `internal/provider/bucket_resource.go:476-478`
   (no change needed; this is why *adding* object lock to an existing bucket works, while
   *creating* a new bucket with it fails).

5. **Why CI didn't catch it** — the test mock accepts object_lock_config on POST:
   `internal/testmock/handlers/buckets.go:196-197`
   ```go
   if body.ObjectLockConfig != nil {
       b.ObjectLockConfig = *body.ObjectLockConfig
   }
   ```
   The real array returns HTTP 400 instead. The mock must mirror the array.

## Fix

Mirror the existing `versioning` handling: drop `object_lock_config` from POST, apply it via a
post-create PATCH. Object lock requires versioning, so the object-lock PATCH must run **after**
the versioning PATCH.

### 1. `internal/client/models_storage.go`

Remove `ObjectLockConfig` from `BucketPost` and extend the doc comment. Keep it on
`BucketPatch` (unchanged).

```go
// NOTE: versioning is NOT a valid POST parameter — use PATCH after creation.
// NOTE: public_access_config is NOT valid on POST — PATCH only.
// NOTE: object_lock_config is NOT valid on POST — PATCH only (and requires versioning).
type BucketPost struct {
    Account           NamedReference     `json:"account"`
    QuotaLimit        string             `json:"quota_limit,omitempty"`
    HardLimitEnabled  bool               `json:"hard_limit_enabled,omitempty"`
    RetentionLock     string             `json:"retention_lock,omitempty"`
    EradicationConfig *EradicationConfig `json:"eradication_config,omitempty"`
}
```

### 2. `internal/provider/bucket_resource.go` — Create

Delete the POST block at lines 339-341:

```go
// DELETE these lines:
if cfg := extractObjectLockConfig(data.ObjectLockConfig); cfg != nil {
    post.ObjectLockConfig = cfg
}
```

Then, right **after** the existing versioning PATCH block (currently ending around line 360),
add an object-lock PATCH:

```go
// object_lock_config is not a valid POST parameter — apply via PATCH after
// creation, and after versioning (object lock requires versioning enabled).
if cfg := extractObjectLockConfig(data.ObjectLockConfig); cfg != nil {
    if _, err := r.client.PatchBucket(ctx, bucket.ID, client.BucketPatch{ObjectLockConfig: cfg}); err != nil {
        resp.Diagnostics.AddError("Error setting bucket object lock config", err.Error())
        return
    }
}
```

No change to `extractObjectLockConfig` (`bucket_resource.go:700-721`) or to Update
(`bucket_resource.go:476-478`).

### 3. `internal/testmock/handlers/buckets.go` — make the mock match the real array

In `handlePost`, reject `object_lock_config` in the POST body so the mock fails the same way
the array does. Replace the lenient apply (lines ~196-197) with a 400:

```go
if body.ObjectLockConfig != nil {
    WriteJSONError(w, http.StatusBadRequest, "Invalid body parameter: object_lock_enabled")
    return
}
```

`handlePatch` already supports `object_lock_config` (`buckets.go:295-301`) — leave it.

## Tests

- Add an acceptance/unit test: create a bucket with `object_lock_config = { object_lock_enabled = true, default_retention = 3600, default_retention_mode = "compliance" }` and assert the final state reflects it (proves create → PATCH path works). The existing test fixtures already reference `object_lock_config` (`bucket_resource_test.go:93,147,1061-1065,1106-1107`).
- Add a negative mock test (or rely on the updated mock) asserting that a POST body containing `object_lock_config` returns 400 — this is the regression guard that would have caught the bug.
- Confirm the "no object lock" path still creates with a clean POST (no `object_lock_config` key).

## Acceptance criteria

- [ ] `terraform apply` of the reproduction config succeeds; object lock is set on the bucket.
- [ ] POST body sent to `/buckets` no longer contains `object_lock_config`.
- [ ] Object lock is applied via PATCH after creation, after versioning.
- [ ] Test mock returns HTTP 400 when `object_lock_config` is present in a POST body.
- [ ] New test covers create-with-object-lock end to end; `go test ./...` passes.
- [ ] Buckets created without `object_lock_config` are unchanged (no PATCH, no diff).

## Notes / caveats

- `ObjectLockConfig.ObjectLockEnabled` is `bool` with `json:"...,omitempty"`
  (`models_storage.go:123-128`), so an explicit `object_lock_enabled = false` serializes to
  nothing. That's fine for the enable case, but if explicit-false support is ever needed the
  field should become `*bool`. Out of scope for this fix.
- Downstream: once released, bump the version constraint in the `terraform-flashblade-s3-bucket`
  module (currently pinned `numberly/mica` `2.22.6`) to the fixed release.
