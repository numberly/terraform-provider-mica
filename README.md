# Terraform Provider FlashBlade

[![CI](https://github.com/numberly/opentofu-provider-flashblade/actions/workflows/ci.yml/badge.svg)](https://github.com/numberly/opentofu-provider-flashblade/actions/workflows/ci.yml)
![Coverage](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/SoulKyu/59bd98f69a5ecbb7e643402fde956fed/raw/coverage.json)
![Go Version](https://img.shields.io/badge/go-1.25-00ADD8?logo=go&logoColor=white)
![Terraform](https://img.shields.io/badge/terraform-%E2%89%A5_1.0-844FBA?logo=terraform&logoColor=white)
[![Go Report Card](https://goreportcard.com/badge/github.com/numberly/opentofu-provider-flashblade)](https://goreportcard.com/report/github.com/numberly/opentofu-provider-flashblade)
![Latest Release](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/SoulKyu/59bd98f69a5ecbb7e643402fde956fed/raw/release.json)

Terraform provider for [Pure Storage FlashBlade](https://www.purestorage.com/products/file-and-object/flashblade.html), managing storage resources via the FlashBlade REST API v2.22.

## Overview

This provider enables GitOps-driven management of FlashBlade storage: file systems, object store accounts and buckets, access policies, quotas, lifecycle rules, audit filters, QoS policies, cross-array replication, and array-level configuration — all as Terraform resources.

## Requirements

- Terraform >= 1.0
- Go >= 1.25 (for development only)
- FlashBlade array with REST API v2.22+ (Purity//FB 4.6.7+)

## Installation

```hcl
terraform {
  required_providers {
    flashblade = {
      source  = "numberly/flashblade"
      version = "~> 2.1"
    }
  }
}
```

## Provider Configuration

```hcl
provider "flashblade" {
  endpoint = "https://flashblade.example.com"

  # Option A: API token
  auth = {
    api_token = var.flashblade_api_token
  }

  # Option B: OAuth2 token exchange
  # auth = {
  #   oauth2 = {
  #     client_id = var.client_id
  #     key_id    = var.key_id
  #     issuer    = var.issuer
  #   }
  # }
}
```

Environment variables: `FLASHBLADE_HOST`, `FLASHBLADE_API_TOKEN`.

## Resources & Data Sources

### Storage

| Resource | Data Source | Description |
|----------|:----------:|-------------|
| `flashblade_file_system` | ✅ | NFS/SMB file system with soft-delete lifecycle |
| `flashblade_bucket` | ✅ | S3 bucket (versioning, quota, eradication, object lock, public access) |
| `flashblade_object_store_account` | ✅ | Object store account (S3 namespace) |
| `flashblade_object_store_access_key` | ✅ | S3 access key pair (cross-array secret sharing) |

### Bucket Advanced Features

| Resource | Data Source | Description |
|----------|:----------:|-------------|
| `flashblade_lifecycle_rule` | ✅ | Per-bucket lifecycle rule (version retention, multipart cleanup) |
| `flashblade_bucket_access_policy` | ✅ | Per-bucket IAM-style access policy |
| `flashblade_bucket_access_policy_rule` | — | Rule within a bucket access policy (principals format varies by firmware) |
| `flashblade_bucket_audit_filter` | ✅ | Per-bucket S3 audit filter (actions + prefix) |
| `flashblade_qos_policy` | ✅ | QoS policy (bandwidth + IOPS limits) |
| `flashblade_qos_policy_member` | — | Assign QoS policy to file systems or realms (buckets not supported on API v2.22) |

### Servers & Exports

| Resource | Data Source | Description |
|----------|:----------:|-------------|
| `flashblade_server` | ✅ | FlashBlade server with DNS configuration |
| `flashblade_file_system_export` | ✅ | File system export to a server (NFS) |
| `flashblade_object_store_account_export` | ✅ | Object store account export to a server (S3) |
| `flashblade_object_store_virtual_host` | ✅ | S3 virtual-hosted-style endpoint |

### NFS Policies

| Resource | Data Source | Description |
|----------|:----------:|-------------|
| `flashblade_nfs_export_policy` | ✅ | NFS export policy |
| `flashblade_nfs_export_policy_rule` | — | Rule within an NFS export policy |

### SMB Policies

| Resource | Data Source | Description |
|----------|:----------:|-------------|
| `flashblade_smb_share_policy` | ✅ | SMB share policy (file permissions) |
| `flashblade_smb_share_policy_rule` | — | Rule within an SMB share policy |
| `flashblade_smb_client_policy` | ✅ | SMB client policy (auth, encryption) |
| `flashblade_smb_client_policy_rule` | — | Rule within an SMB client policy |

### S3 Policies

| Resource | Data Source | Description |
|----------|:----------:|-------------|
| `flashblade_object_store_access_policy` | ✅ | IAM-style S3 access policy |
| `flashblade_object_store_access_policy_rule` | — | Rule within an S3 access policy |
| `flashblade_s3_export_policy` | ✅ | S3 export transport-level access policy |
| `flashblade_s3_export_policy_rule` | — | Rule within an S3 export policy |

### Snapshot & Network Policies

| Resource | Data Source | Description |
|----------|:----------:|-------------|
| `flashblade_snapshot_policy` | ✅ | Snapshot schedule policy |
| `flashblade_snapshot_policy_rule` | — | Rule within a snapshot policy |
| `flashblade_network_access_policy` | ✅ | Network access policy (singleton) |
| `flashblade_network_access_policy_rule` | — | Rule within a network access policy |

### Quotas

| Resource | Data Source | Description |
|----------|:----------:|-------------|
| `flashblade_quota_user` | ✅ | Per-filesystem user quota |
| `flashblade_quota_group` | ✅ | Per-filesystem group quota |

### Array Administration

| Resource | Data Source | Description |
|----------|:----------:|-------------|
| `flashblade_array_dns` | ✅ | Array DNS configuration (singleton) |
| `flashblade_array_ntp` | ✅ | Array NTP server list (singleton) |
| `flashblade_array_smtp` | ✅ | Array SMTP relay and alert watchers (singleton) |
| `flashblade_syslog_server` | ✅ | Syslog server configuration |

### Replication

| Resource | Data Source | Description |
|----------|:----------:|-------------|
| `flashblade_object_store_remote_credentials` | ✅ | S3 credentials for cross-array replication |
| `flashblade_bucket_replica_link` | ✅ | Bucket-to-bucket replica link (pause/resume) |
| — | `flashblade_array_connection` | Array connection status (read-only) |

**Total: 36 resources, 28 data sources**

## Workflow Examples

Production-ready configurations showing how resources compose together:

| Workflow | Description |
|----------|-------------|
| [Object Store Setup](examples/workflows/object-store-setup/) | S3-compatible storage: account, bucket, access key |
| [NFS File Share](examples/workflows/nfs-file-share/) | Team shared storage with export policy |
| [Multi-Protocol File System](examples/workflows/multi-protocol-file-system/) | Windows + Linux access on same FS |
| [Array Admin Baseline](examples/workflows/array-admin-baseline/) | Day-1 DNS, NTP, SMTP configuration |
| [Secured S3 Bucket](examples/workflows/secured-s3-bucket/) | Bucket with network + access policies |
| [S3 Tenant Full-Stack](examples/workflows/s3-tenant-full-stack/) | Complete S3 onboarding: server → account → export → policies → key → bucket |
| [Vault S3 Onboarding](examples/workflows/vault-s3-onboarding/) | Same as above + Vault for zero-secret credential management |
| [S3 Bucket Replication](examples/workflows/s3-bucket-replication/) | Bidirectional cross-array S3 replication with shared credentials |
| [Bucket Advanced Features](examples/workflows/bucket-advanced-features/) | Lifecycle rules, access policies, audit filters, QoS |

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
