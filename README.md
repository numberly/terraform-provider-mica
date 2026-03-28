# Terraform Provider FlashBlade

[![CI](https://github.com/numberly/opentofu-provider-flashblade/actions/workflows/ci.yml/badge.svg)](https://github.com/numberly/opentofu-provider-flashblade/actions/workflows/ci.yml)

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
      source  = "numberly/flashblade"
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

## Workflow Examples

Production-ready configurations showing how resources compose together:

| Workflow | Description | Resources Used |
|----------|-------------|----------------|
| [Object Store Setup](examples/workflows/object-store-setup/) | S3-compatible storage: account, bucket, access key | account, bucket, access_key |
| [NFS File Share](examples/workflows/nfs-file-share/) | Team shared storage with export policy | file_system, nfs_export_policy, nfs_export_policy_rule |
| [Multi-Protocol File System](examples/workflows/multi-protocol-file-system/) | Windows + Linux access on same FS | file_system, nfs_export_policy, smb_share_policy |
| [Array Admin Baseline](examples/workflows/array-admin-baseline/) | Day-1 DNS, NTP, SMTP configuration | array_dns, array_ntp, array_smtp |
| [Secured S3 Bucket](examples/workflows/secured-s3-bucket/) | Bucket with network + access policies | bucket, network_access_policy, object_store_access_policy |

## Resource Coverage Roadmap

### v1.0 — Current Release

| Resource | Create | Read | Update | Delete | Import | Data Source | Notes |
|----------|:------:|:----:|:------:|:------:|:------:|:-----------:|-------|
| **Storage** |
| `flashblade_file_system` | ✅ | ✅ | ✅ | ✅ soft-delete | ✅ | ✅ | Two-phase destroy, in-place rename |
| `flashblade_bucket` | ✅ | ✅ | ✅ | ✅ soft-delete | ✅ | ✅ | Default recoverable, ForceNew on rename |
| `flashblade_object_store_account` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | Fails if buckets exist |
| `flashblade_object_store_access_key` | ✅ | ✅ | — | ✅ | — | ✅ | Immutable, write-once secret |
| **NFS Export Policy** |
| `flashblade_nfs_export_policy` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | Guard: fails if attached to FS |
| `flashblade_nfs_export_policy_rule` | ✅ | ✅ | ✅ | ✅ | ✅ `name/index` | — | Index-based ordering |
| **SMB Share Policy** |
| `flashblade_smb_share_policy` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | Guard: fails if attached to FS |
| `flashblade_smb_share_policy_rule` | ✅ | ✅ | ✅ | ✅ | ✅ `name/rule_name` | — | Name-based identity |
| **Snapshot Policy** |
| `flashblade_snapshot_policy` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ForceNew on rename (API read-only) |
| `flashblade_snapshot_policy_rule` | ✅ | ✅ | ✅ | ✅ | ✅ `name/index` | — | Via parent PATCH add/remove_rules |
| **Object Store Access Policy** |
| `flashblade_object_store_access_policy` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | IAM-style, guard if attached |
| `flashblade_object_store_access_policy_rule` | ✅ | ✅ | ✅ | ✅ | ✅ `name/rule_name` | — | JSON conditions, effect RequiresReplace |
| **Network Access Policy** |
| `flashblade_network_access_policy` | ✅ singleton | ✅ | ✅ | ✅ reset | ✅ | ✅ | No POST/DELETE — GET+PATCH only |
| `flashblade_network_access_policy_rule` | ✅ | ✅ | ✅ | ✅ | ✅ `name/index` | — | Index-based ordering |
| **Quota** |
| `flashblade_quota_user` | ✅ | ✅ | ✅ | ✅ | ✅ `fs/uid` | ✅ | Per-filesystem user quota |
| `flashblade_quota_group` | ✅ | ✅ | ✅ | ✅ | ✅ `fs/gid` | ✅ | Per-filesystem group quota |
| **Array Administration** |
| `flashblade_array_dns` | ✅ singleton | ✅ | ✅ | ✅ reset | ✅ `default` | ✅ | Destroy clears config |
| `flashblade_array_ntp` | ✅ singleton | ✅ | ✅ | ✅ reset | ✅ `default` | ✅ | Destroy clears NTP servers |
| `flashblade_array_smtp` | ✅ singleton | ✅ | ✅ | ✅ reset | ✅ `default` | ✅ | Includes alert watchers |

**Legend:** ✅ = supported | — = intentionally not supported | `soft-delete` = two-phase destroy + eradicate | `singleton` = adopt existing via GET+PATCH | `reset` = destroy resets to defaults

### v1.x — Planned

| Resource | Description | Priority |
|----------|-------------|----------|
| Object Lock / WORM config | `retention_lock`, `object_lock_config` on buckets | P2 |
| QoS policy attachment | Bandwidth/IOPS control on file systems and buckets | P2 |
| Eradication config | Custom eradication delay per resource | P2 |
| Acceptance tests | Tests against real FlashBlade array (`TF_ACC=1`) | P2 |
| Terraform Registry | Public publication on registry.terraform.io | P2 |

### v2+ — Future

| Resource | Description | Complexity |
|----------|-------------|------------|
| Bucket replica links | DR automation between arrays | High |
| File system replica links | Cross-array replication | High |
| Array connections | Multi-array connectivity management | High |
| API client management | Service account provisioning | Medium |
| Active Directory | Domain join via Terraform | High |
| Pulumi bridge | Pulumi SDK from Terraform provider | Medium |

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

Generated docs are in the `docs/` directory and published to the [Terraform Registry](https://registry.terraform.io/providers/numberly/flashblade/latest).

## Contributing

1. Fork the repository
2. Create a feature branch
3. Run `make test` and `make lint` before opening a PR
4. Ensure `make docs` produces no diff

## License

See [LICENSE](LICENSE).
