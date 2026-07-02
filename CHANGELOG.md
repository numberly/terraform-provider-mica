## [2.23.3] — 2026-07-02

### Added

- **`flashblade_bucket_cors_policy` resource + data source** — per-bucket wildcard CORS toggle. FlashBlade only supports fully permissive CORS today (origins/headers `*`, all methods), so the resource takes just a `bucket_name`: its presence ensures the bucket's CORS policy (empty-body POST, auto-named) and applies the single wildcard rule via the `/rules` sub-endpoint (no PATCH); destroy removes the policy.

### Changed

- Pulumi bridge: regenerated schema + Go/Python SDKs to expose `mica:index:BucketCorsPolicy`, and synced pre-existing 2.23 SDK drift (workload, resiliency groups, and 2.23 field additions on file_system/exports/policies).

## [2.23.0] — 2026-05-20

### Added

- **`flashblade_workload` resource** — CRUD + drift detection + import. Covers WORKLOAD-01.
- **`flashblade_workload` data source** — read-only lookup by name. Covers WORKLOAD-02.
- **`flashblade_resiliency_group` data source** — read-only. Covers RESILIENCY-01.
- **`flashblade_resiliency_group_member` data source** — read-only. Covers RESILIENCY-02.

### Changed

- **FlashBlade API version bumped from `2.22` to `2.23`** across provider, client, mock handlers, and API references. Covers API-01..03.
- **Schema migrations (v0 → v1)** add a computed `workload` field to the following resources. State is auto-upgraded on first refresh; no user action required:
  - `flashblade_file_system` (SCHEMA-01)
  - `flashblade_file_system_export` (SCHEMA-02)
  - `flashblade_nfs_export_policy` (SCHEMA-03)
  - `flashblade_smb_client_policy` (SCHEMA-04)
  - `flashblade_smb_share_policy` (SCHEMA-05)
  - `flashblade_qos_policy` (SCHEMA-06, plus computed `context` field SCHEMA-07; schema v1 → v2)
- **Pulumi bridge artefacts regenerated** for API 2.23 (`schema.json`, `bridge-metadata.json`, `schema-embed.json`). Covers BRIDGE-01..03.

### Migration

No manual user action required. Schema upgraders run automatically on first `terraform refresh` / `plan` / `apply` after upgrading the provider. The new `workload` attribute is computed (read-only) and surfaces the FlashBlade workload assignment for the underlying object.

### Validation

Validated on FlashBlade arrays `par5` and `pa7` (API 2.23, Purity//FB ≥ 4.6.7). `make test` (807+ tests) + `make lint` + `make docs` + `make tfgen` all clean on the release commit.

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
