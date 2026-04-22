# Pulumi FlashBlade Provider

A Pulumi package for managing Pure Storage FlashBlade resources. This provider is bridged from the existing Terraform provider using `pulumi-terraform-bridge/v3`.

## Installation

### Prerequisites

- [Pulumi CLI](https://www.pulumi.com/docs/install/) >= 3.142.0
- Go 1.25+ (for Go SDK consumers)
- Python 3.9+ (for Python SDK consumers)

### Provider Plugin

Install the provider binary from GitHub Releases:

```bash
pulumi plugin install resource flashblade v2.22.3 --server github://api.github.com/numberly
```

For a specific version, replace `v2.22.3` with the desired tag (without the `pulumi-` prefix).

### Python SDK

Install the wheel from the GitHub Release asset:

```bash
pip install https://github.com/numberly/opentofu-provider-flashblade/releases/download/pulumi-2.22.3/pulumi_flashblade-2.22.3-py3-none-any.whl
```

Or add to your `requirements.txt`:

```
pulumi_flashblade @ https://github.com/numberly/opentofu-provider-flashblade/releases/download/pulumi-2.22.3/pulumi_flashblade-2.22.3-py3-none-any.whl
```

### Go SDK

The Go SDK is distributed as a Go module via git tags. Because the repository is private, configure `GOPRIVATE`:

```bash
export GOPRIVATE="github.com/numberly/*"
```

Then fetch the SDK:

```bash
go get github.com/numberly/opentofu-provider-flashblade/pulumi/sdk/go@v2.22.3
```

The Go module tag follows the pattern `sdk/go/vX.Y.Z` (e.g., `sdk/go/v2.22.3`).

## Provider Configuration

The provider accepts the same configuration as the Terraform provider. Set values via Pulumi config or environment variables:

| Pulumi Config Key | Environment Variable | Description |
|---|---|---|
| `flashblade:endpoint` | `FLASHBLADE_ENDPOINT` | FlashBlade management IP/hostname |
| `flashblade:auth.apiToken` | `FLASHBLADE_AUTH_API_TOKEN` | API token for authentication |
| `flashblade:auth.oauth2.clientId` | `FLASHBLADE_AUTH_OAUTH2_CLIENT_ID` | OAuth2 client ID |
| `flashblade:auth.oauth2.keyId` | `FLASHBLADE_AUTH_OAUTH2_KEY_ID` | OAuth2 key ID |
| `flashblade:caCert` | `FLASHBLADE_CA_CERT` | CA certificate content |
| `flashblade:insecureSkipVerify` | `FLASHBLADE_INSECURE_SKIP_VERIFY` | Skip TLS verification |

### Python Example

```python
import pulumi
import pulumi_flashblade as flashblade

provider = flashblade.Provider("flashblade",
    endpoint="https://flashblade.example.com",
    auth={"api_token": "t.abc123"},
)

target = flashblade.Target("primary",
    name="s3-replication-target",
    address="s3.us-east-1.amazonaws.com",
    opts=pulumi.ResourceOptions(provider=provider),
)
```

### Go Example

```go
package main

import (
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
    "github.com/numberly/opentofu-provider-flashblade/pulumi/sdk/go/flashblade"
)

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        provider, err := flashblade.NewProvider(ctx, "flashblade", &flashblade.ProviderArgs{
            Endpoint: pulumi.String("https://flashblade.example.com"),
            Auth:     pulumi.StringMap{"api_token": pulumi.String("t.abc123")},
        })
        if err != nil {
            return err
        }

        target, err := flashblade.NewTarget(ctx, "primary", &flashblade.TargetArgs{
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

FlashBlade resource names are operational identifiers. This provider **does not use autonaming** — you must supply an explicit `name` for every resource. Choose names that are stable and meaningful in your infrastructure.

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
- **No Pulumi Registry publication** — install via GitHub Releases only.
- **No PyPI publication** — install the wheel from the release asset URL.
- **Write-once fields** are secret but not write-only at the SDK layer (deferred to a future milestone).
- **Delete timeout** is inherited from the Terraform provider (30m for bucket/filesystem). Bridge-level timeout overrides are not available in bridge v3.127.0.

## License

Apache-2.0
