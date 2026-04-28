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

- Terraform resource type names: `flashblade_bucket`, `flashblade_target`, `flashblade_filesystem`, etc.
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
