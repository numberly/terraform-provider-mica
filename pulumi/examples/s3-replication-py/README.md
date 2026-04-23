# FlashBlade S3 Replication Example

Dual-array bidirectional S3 bucket replication — mirrors the `terraform-flashblade-s3-bucket` Terraform module.

## What it creates

- Object store accounts on both arrays
- S3 export policies + account exports (link accounts to servers)
- IAM-style access policies (global rw + per-user)
- Named S3 users with per-user policies and access keys
- Versioned buckets with optional quotas
- Bidirectional replication via remote credentials + replica links
- Optional S3 target replication (instead of array-to-array)
- Lifecycle rules, audit filters, and QoS policy

## Prerequisites

- Array connection configured between both arrays
- Servers pre-provisioned on both arrays
- Pulumi CLI + FlashBlade provider SDK built (`make build_python` or `make build_go`)

## Python

```bash
cd s3-replication-py
pip install ../../../sdk/python/dist/pulumi_flashblade-*.whl
pulumi stack init dev
pulumi config set par5_endpoint "https://par5.flashblade.example.com"
pulumi config set pa7_endpoint "https://pa7.flashblade.example.com"
pulumi config set --secret par5_api_token "t.abc123"
pulumi config set --secret pa7_api_token "t.xyz789"
pulumi config set par5_array_name "par5"
pulumi config set pa7_array_name "pa7"
pulumi config set par5_server_name "server-par5"
pulumi config set pa7_server_name "server-pa7"
pulumi config set account_name "my-account"
pulumi config set bucket_name "my-replicated-bucket"
pulumi config set fqdn "s3.example.com"
pulumi up
```

## Go

```bash
cd s3-replication-go
go mod tidy
pulumi stack init dev
pulumi config set par5Endpoint "https://par5.flashblade.example.com"
pulumi config set pa7Endpoint "https://pa7.flashblade.example.com"
pulumi config set --secret par5ApiToken "t.abc123"
pulumi config set --secret pa7ApiToken "t.xyz789"
pulumi config set par5ArrayName "par5"
pulumi config set pa7ArrayName "pa7"
pulumi config set par5ServerName "server-par5"
pulumi config set pa7ServerName "server-pa7"
pulumi config set accountName "my-account"
pulumi config set bucketName "my-replicated-bucket"
pulumi config set fqdn "s3.example.com"
pulumi up
```

## Optional config

```bash
# Bucket quota (bytes)
pulumi config set bucket_quota_bytes 107374182400

# S3 target replication (instead of array-to-array)
pulumi config set enable_s3_target_replication true
pulumi config set flashblade_target_par5_sees_pa7 "s3-target-par5-to-pa7"
pulumi config set flashblade_target_pa7_sees_par5 "s3-target-pa7-to-par5"

# Audit filters
pulumi config set audit_enabled true
pulumi config set --path audit_prefixes '["logs/", "audit/"]'

# QoS policy
pulumi config set --path qos.max_total_bytes_per_sec 1073741824
pulumi config set --path qos.max_total_ops_per_sec 10000

# Per-user S3 credentials
pulumi config set --path users.app1-rw.vault_secret_path "secrets/team/app1/bucket"
pulumi config set --path users.app1-rw.s3_actions '["s3:GetObject", "s3:PutObject", "s3:DeleteObject", "s3:ListBucket"]'
pulumi config set --path users.app1-rw.s3_resources '["*"]'
pulumi config set --path users.app1-rw.effect "allow"
pulumi config set --path users.app1-rw.full_access false
```
