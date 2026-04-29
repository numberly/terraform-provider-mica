# Mica — Pulumi provider for Pure Storage FlashBlade®

A Pulumi package for managing resources on Pure Storage FlashBlade® arrays. Bridged from the Mica Terraform provider via `pulumi-terraform-bridge/v3`.

> **Mica is an independent open-source project. It is NOT affiliated with, endorsed by, certified by, or sponsored by Pure Storage, Inc.**
> Pure Storage®, FlashBlade®, and Purity® are registered trademarks of Pure Storage, Inc. and/or its affiliates and are used here only as nominative descriptive references to identify the target system.

## Installation

### Prerequisites

- [Pulumi CLI](https://www.pulumi.com/docs/install/) >= 3.142.0
- Go 1.25+ (for Go SDK consumers)
- Python 3.9+ (for Python SDK consumers)

### Provider Plugin

Install the provider binary from GitHub Releases:

```bash
pulumi plugin install resource mica 2.22.3-pulumi.alpha --server github://api.github.com/numberly
```

Pulumi bridge releases are tagged `v{X.Y.Z}-pulumi[.suffix]` (valid SemVer prerelease). Replace `2.22.3-pulumi.alpha` with any published version, dropping the leading `v`.

### Python SDK

Install the wheel from the GitHub Release asset:

```bash
pip install https://github.com/numberly/terraform-provider-mica/releases/download/v2.22.3-pulumi.alpha/pulumi_mica-2.22.3-pulumi.alpha-py3-none-any.whl
```

Or add to your `requirements.txt`:

```
pulumi_mica @ https://github.com/numberly/terraform-provider-mica/releases/download/v2.22.3-pulumi.alpha/pulumi_mica-2.22.3-pulumi.alpha-py3-none-any.whl
```

### Go SDK

The Go SDK is distributed as a Go module via git tags. It is **versioned
independently** from the provider binary (the provider is at v2.x, but the
Go SDK stays on major v0 so the module path does not need a `/vN` suffix).
The SDK version lives in `pulumi/sdk/go/VERSION`.

Because the repository is private, configure `GOPRIVATE`:

```bash
export GOPRIVATE="github.com/numberly/*"
```

Then fetch the SDK (use the Go SDK tag, not the provider tag):

```bash
go get github.com/numberly/terraform-provider-mica/pulumi/sdk/go@v0.1.0-pulumi.alpha
```

The Go module tag follows the pattern `sdk/go/v{SDK_VERSION}-pulumi[.suffix]`
(e.g., `sdk/go/v0.1.0-pulumi.alpha`). Each provider release tag
`v{X.Y.Z}-pulumi[.suffix]` triggers a matching Go SDK tag that reuses the
same `-pulumi[.suffix]` prerelease portion.

## Provider Configuration

The provider accepts the same configuration as the Mica Terraform provider. Set values via Pulumi config or environment variables:

| Pulumi Config Key | Environment Variable | Description |
|---|---|---|
| `mica:endpoint` | `FLASHBLADE_HOST` | Array management endpoint URL |
| `mica:auth.apiToken` | `FLASHBLADE_API_TOKEN` | API token for authentication |
| `mica:auth.oauth2.clientId` | `FLASHBLADE_OAUTH2_CLIENT_ID` | OAuth2 client ID |
| `mica:auth.oauth2.issuer` | `FLASHBLADE_OAUTH2_ISSUER` | OAuth2 issuer URL |
| `mica:auth.oauth2.keyId` | `FLASHBLADE_OAUTH2_KEY_ID` | OAuth2 key ID |
| `mica:caCert` | — | Inline PEM-encoded CA certificate |
| `mica:caCertFile` | — | Path to a PEM-encoded CA certificate file |
| `mica:insecureSkipVerify` | — | Skip TLS verification (development only) |
| `mica:maxRetries` | — | Retry attempts for 429/5xx (default `3`) |

### Python Example

```python
import pulumi
import pulumi_mica as mica

provider = mica.Provider("mica",
    endpoint="https://array.example.com",
    auth={"api_token": "t.abc123"},
)

target = mica.Target("primary",
    name="s3-replication-target",
    address="s3.us-east-1.amazonaws.com",
    opts=pulumi.ResourceOptions(provider=provider),
)
```

### Go Example

```go
package main

import (
    "github.com/numberly/terraform-provider-mica/pulumi/sdk/go/mica"
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        provider, err := mica.NewProvider(ctx, "mica", &mica.ProviderArgs{
            Endpoint: pulumi.String("https://array.example.com"),
            Auth:     &mica.ProviderAuthArgs{ApiToken: pulumi.String("t.abc123")},
        })
        if err != nil {
            return err
        }

        target, err := mica.NewTarget(ctx, "primary", &mica.TargetArgs{
            Name:    pulumi.String("s3-replication-target"),
            Address: pulumi.String("s3.us-east-1.amazonaws.com"),
        }, pulumi.Provider(provider))
        if err != nil {
            return err
        }

        ctx.Export("target_name", target.Name)
        return nil
    })
}
```

## Resource Naming

Array resource names are operational identifiers. This provider **does not use autonaming** — you must supply an explicit `name` for every resource. Choose names that are stable and meaningful in your infrastructure.

## Soft-Delete Resources

`flashblade_bucket` and `flashblade_file_system` use two-phase soft-delete:

1. `pulumi destroy` marks the resource as destroyed (soft-delete)
2. The provider polls until the resource is fully eradicated

The default delete timeout is 30 minutes. If your array is slow to eradicate, increase the timeout via `customTimeouts`:

### Python

```python
opts=pulumi.ResourceOptions(
    custom_timeouts=pulumi.CustomTimeouts(
        create="20m",
        update="20m",
        delete="30m",
    ),
)
```

### Go

```go
pulumi.Timeouts(&pulumi.CustomTimeouts{
    Create: "20m",
    Update: "20m",
    Delete: "30m",
})
```

## Composite ID Import

Some resources use composite IDs for import. The separator is `/` (forward slash).
This is the composite ID format used by `pulumi import`.

### Object Store Access Policy Rule

```bash
pulumi import flashblade:index:ObjectStoreAccessPolicyRule my-rule mypolicy/myrulename
```

### Bucket Access Policy Rule

```bash
pulumi import flashblade:index:BucketAccessPolicyRule my-rule mybucket/myrulename
```

### Network Access Policy Rule

```bash
pulumi import flashblade:index:NetworkAccessPolicyRule my-rule mypolicy/myrulename
```

### Management Access Policy Directory Service Role Membership

Role comes first (policy names may contain slashes):

```bash
pulumi import flashblade:index:ManagementAccessPolicyDirectoryServiceRoleMembership my-membership myrole/mypolicy
```

## Examples

See [`examples/`](examples/) for working programs:

- `target-py/` / `target-go/` — S3 replication target
- `remote_credentials-py/` / `remote_credentials-go/` — Cross-array credentials
- `bucket-py/` / `bucket-go/` — Object store bucket with soft-delete
- `s3-replication-py/` / `s3-replication-go/` — Full dual-array bidirectional replication: accounts, S3 export policies, IAM access policies, S3 users with access keys, versioned buckets, remote credentials, replica links, lifecycle rules, audit filters, and QoS (the most complete end-to-end example)

## State Upgrades

The following resources have Terraform state upgraders that the bridge delegates automatically:

- `flashblade_server` (v0 -> v1 -> v2)
- `flashblade_directory_service_role` (v0 -> v1)
- `flashblade_object_store_remote_credentials` (v0 -> v1)

No manual action is required. Run `pulumi refresh` after importing existing Terraform-managed resources.

## Sensitive Fields

The following fields are marked as secrets in the Pulumi schema:

- `flashblade_object_store_access_key.secret_access_key`
- `flashblade_object_store_remote_credentials.secret_access_key`
- `flashblade_array_connection.connection_key`
- `flashblade_array_connection_key.connection_key`
- `flashblade_certificate.passphrase`
- `flashblade_certificate.private_key`
- `flashblade_directory_service_management.bind_password`

Secret values are encrypted in Pulumi state and masked in CLI output.

## Known Limitations

These are the known limitations of this provider:

- **No TypeScript, C#, or Java SDKs** — Python and Go only.
- **Not published on the Pulumi Registry** — distributed via GitHub Releases only (`pulumi plugin install ... --server github://...`).
- **No PyPI publication yet** — install the wheel from the release asset URL.
- **Write-once fields** are secret but not write-only at the SDK layer (deferred to a future milestone).
- **Delete timeout** is inherited from the Mica Terraform provider (30m for bucket/filesystem). Bridge-level timeout overrides are not available in bridge v3.127.0.

## Trademarks

Pure Storage®, FlashBlade®, FlashBlade//EXA®, FlashArray®, Evergreen//One®, Purity®, and related marks are registered trademarks of Pure Storage, Inc. and/or its affiliates. All other trademarks are the property of their respective owners.

This project uses these names solely as nominative descriptive references to identify the target system. See [`../NOTICE`](../NOTICE) for full attribution.

## License

Apache-2.0
