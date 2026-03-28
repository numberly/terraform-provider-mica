# Terraform Provider FlashBlade

[![CI](https://github.com/soulkyu/terraform-provider-flashblade/actions/workflows/ci.yml/badge.svg)](https://github.com/soulkyu/terraform-provider-flashblade/actions/workflows/ci.yml)

Terraform provider for [Pure Storage FlashBlade](https://www.purestorage.com/products/file-and-object/flashblade.html), managing storage resources via the FlashBlade REST API v2.22.

## Overview

This provider enables GitOps-driven management of FlashBlade storage: file systems, object store accounts and buckets, access policies, quotas, and array-level configuration — all as Terraform resources.

## Requirements

- Terraform >= 1.0
- Go >= 1.22 (for development only)
- FlashBlade array with REST API v2.22+

## Installation

```hcl
terraform {
  required_providers {
    flashblade = {
      source  = "soulkyu/flashblade"
      version = "~> 1.0"
    }
  }
}
```

## Provider Configuration

```hcl
provider "flashblade" {
  endpoint = "https://flashblade.example.com"

  # Option A: API token
  auth {
    api_token = var.flashblade_api_token
  }

  # Option B: OAuth2 token exchange
  # auth {
  #   oauth2_client_id     = var.client_id
  #   oauth2_client_secret = var.client_secret
  # }
}
```

Environment variable: `FLASHBLADE_ENDPOINT`, `FLASHBLADE_API_TOKEN`.

## Resources

| Resource | Description |
|----------|-------------|
| `flashblade_file_system` | NFS/SMB file system with soft-delete lifecycle |
| `flashblade_bucket` | S3 bucket (account-scoped, versioning, quota) |
| `flashblade_object_store_account` | Object store account (S3 namespace) |
| `flashblade_object_store_access_key` | S3 access key pair (create-only, no import) |
| `flashblade_nfs_export_policy` | NFS export policy |
| `flashblade_nfs_export_policy_rule` | Rule within an NFS export policy |
| `flashblade_smb_share_policy` | SMB share policy |
| `flashblade_smb_share_policy_rule` | Rule within an SMB share policy |
| `flashblade_snapshot_policy` | Snapshot schedule policy |
| `flashblade_snapshot_policy_rule` | Rule within a snapshot policy |
| `flashblade_object_store_access_policy` | IAM-style S3 access policy |
| `flashblade_object_store_access_policy_rule` | Rule within an S3 access policy |
| `flashblade_network_access_policy` | Network access policy (singleton, adopt existing) |
| `flashblade_network_access_policy_rule` | Rule within a network access policy |
| `flashblade_quota_group` | Per-filesystem group quota |
| `flashblade_quota_user` | Per-filesystem user quota |
| `flashblade_array_dns` | Array DNS configuration (singleton) |
| `flashblade_array_ntp` | Array NTP server list (singleton) |
| `flashblade_array_smtp` | Array SMTP relay and alert watchers (singleton) |

## Data Sources

| Data Source | Description |
|-------------|-------------|
| `flashblade_file_system` | Look up an existing file system |
| `flashblade_bucket` | Look up an existing bucket |
| `flashblade_object_store_account` | Look up an existing object store account |
| `flashblade_object_store_access_key` | Look up an existing access key |
| `flashblade_nfs_export_policy` | Look up an existing NFS export policy |
| `flashblade_smb_share_policy` | Look up an existing SMB share policy |
| `flashblade_snapshot_policy` | Look up an existing snapshot policy |
| `flashblade_object_store_access_policy` | Look up an existing S3 access policy |
| `flashblade_network_access_policy` | Look up an existing network access policy |
| `flashblade_quota_group` | Look up an existing group quota |
| `flashblade_quota_user` | Look up an existing user quota |
| `flashblade_array_dns` | Read current array DNS configuration |
| `flashblade_array_ntp` | Read current array NTP configuration |
| `flashblade_array_smtp` | Read current array SMTP configuration |

## Development

```bash
# Build
make build

# Run unit tests
make test

# Run linter
make lint

# Regenerate docs/
make docs

# Install locally for manual testing
make install
```

## Documentation

Generated docs are in the `docs/` directory and published to the [Terraform Registry](https://registry.terraform.io/providers/soulkyu/flashblade/latest).

## Contributing

1. Fork the repository
2. Create a feature branch
3. Run `make test` and `make lint` before opening a PR
4. Ensure `make docs` produces no diff

## License

See [LICENSE](LICENSE).
