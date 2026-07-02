## [2.22.12] — 2026-07-02

### Fixed

- **`flashblade_s3_export_policy_rule`**: fall back to the composite `policy_name/rule_index` id when the array returns an empty rule id (some rules created by older provider builds), so the Pulumi bridge no longer fails with `returned empty resource.ID from create`. The real id is kept when present.

## [2.22.11] — 2026-07-02

### Fixed

- **`flashblade_bucket_cors_policy`** and **`flashblade_s3_export_policy_rule`**: make rule creation idempotent. The array rejects re-POSTing an existing rule with HTTP 400 `Rule already exists.`, which broke apply/retry when a rule was left on the array by an earlier partial apply. CORS now deletes the wildcard rule before recreating it (ensure policy → delete rule → post rule); S3 export policy rule now adopts the existing rule on already-exists instead of erroring.

## [2.22.10] — 2026-07-02

### Fixed

- **`flashblade_bucket_cors_policy`**: use the bucket name as the resource id. The FlashBlade CORS policy GET returns an empty id, which made resource creation fail under the Pulumi bridge (`returned empty resource.ID from create`). The bucket name is unique per CORS policy, stable, and matches import-by-bucket-name.

## [2.22.9] — 2026-07-02

### Added

- **`flashblade_bucket_cors_policy` resource + data source** — per-bucket wildcard CORS toggle. FlashBlade only supports fully permissive CORS today (origins/headers `*`, all methods), so the resource takes just a `bucket_name`: its presence ensures the bucket's CORS policy (empty-body POST, auto-named) and applies the single wildcard rule via the `/rules` sub-endpoint (no PATCH); destroy removes the policy. Backport of the 2.23-line feature.

### Changed

- Pulumi bridge: rebranded the SDK from the legacy `pulumi_flashblade` package to `pulumi_mica`, and regenerated schema + Go/Python SDKs to expose `mica:index:BucketCorsPolicy`.

## [2.22.4] — 2026-04-28

### Project rebrand

This release renames the project to **Mica** for open-source release. The provider continues to target Pure Storage FlashBlade® arrays exactly as before.

### Changed (breaking)

- **Registry source path**: `numberly/flashblade` → `numberly/mica`
- **Go module path**: `github.com/numberly/opentofu-provider-flashblade` → `github.com/numberly/terraform-provider-mica`
- **Pulumi package name**: `pulumi-flashblade` → `pulumi-mica`
- **Pulumi resource tokens**: `flashblade:*:*` → `mica:*:*`
- **License**: now distributed under **GPL v3** (was: unspecified)

### Unchanged

- Terraform resource type names: `flashblade_bucket`, `flashblade_target`, `flashblade_file_system`, etc.
- HCL `provider "flashblade" {}` block syntax (the local alias remains user-controlled)
- Internal Go identifiers (`FlashBladeClient`, package layout, etc.)
- All schema fields, behaviors, and acceptance test fixtures

### Migration

Update the `source` field in `required_providers`:

```hcl
terraform {
  required_providers {
    flashblade = {
      source  = "numberly/mica"   # was: "numberly/flashblade"
      version = "2.22.4"
    }
  }
}
```

Then migrate existing state:

```bash
terraform init
terraform state replace-provider numberly/flashblade numberly/mica
```

`replace-provider` rewrites every resource's recorded provider reference. Without this step, `terraform plan` will fail with a provider mismatch error.

### Versioning note

This project tracks the upstream FlashBlade API version as `MAJOR.MINOR.PATCH`. Despite the patch-level bump in `v2.22.4`, this release contains breaking changes (registry source path, module path, license). Pin exactly with `version = "2.22.4"` rather than `~> 2.22.4` if you want to control migration timing.
