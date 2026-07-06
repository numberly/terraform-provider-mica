# Mica — OpenWiki Quickstart

**Mica** is an open-source **Terraform / OpenTofu provider** for Pure Storage
FlashBlade® scale-out storage. It manages FlashBlade objects (file systems,
S3 buckets, accounts, access/network/snapshot policies, quotas, networking,
cross-array replication, array-level config) as declarative IaC resources via
the FlashBlade REST API.

> Independent project — **not** affiliated with or endorsed by Pure Storage, Inc.
> `FlashBlade®`/`Purity®` are used nominatively to name the target system.

- **Module**: `github.com/numberly/terraform-provider-mica` (Go 1.25+)
- **Framework**: `terraform-plugin-framework` (never SDK v2) served over gRPC
- **Target API**: FlashBlade REST **v2.23** (Purity//FB 4.6.7+); a **v2.22**
  maintenance line ships in parallel — see [API versioning](api-versioning.md)
- **Surface** (current, per `internal/provider/provider.go`): **57 resources**,
  **45 data sources**. The `README.md` catalog (40/32) reflects an older release.
- **Also ships as a Pulumi package** (`mica:` namespace) via a Terraform bridge —
  see [Pulumi bridge](pulumi-bridge.md).

## What this repo is (and why it's shaped this way)

FlashBlade's REST API distinguishes create/update/read payloads per resource, is
versioned, and has quirks (soft-delete + eradication, composite policy-rule IDs,
firmware-dependent field shapes, swagger inaccuracies). The provider absorbs all
of that behind a strict, repeatable per-resource pattern so that adding or
upgrading a resource is mechanical and reviewable. Two things enforce that
discipline and are required reading before changing code:

- **[CONVENTIONS.md](../CONVENTIONS.md)** — the single source of truth for code
  layout, struct/pointer rules, plan modifiers, tests. Summarized in
  [conventions.md](conventions.md).
- **[CLAUDE.md](../CLAUDE.md)** — project instructions + skills index.

## Entry points

| Path | Role |
|------|------|
| `main.go` | Provider binary; serves `registry.terraform.io/numberly/mica`; hosts the `//go:generate tfplugindocs` directive |
| `internal/provider/provider.go` | Provider schema, `Configure`, `Resources()`/`DataSources()` registration |
| `internal/client/client.go` | HTTP client, generics, API-version prefixing |
| `pulumi/provider/resources.go` | Pulumi bridge `ProviderInfo` |
| `.planning/` | GSD structured-development state (milestones/phases) |

## Install & configure (users)

```hcl
terraform {
  required_providers {
    flashblade = { source = "numberly/mica" }
  }
}

provider "flashblade" {
  endpoint = "https://flashblade.example.com"   # or FLASHBLADE_HOST
  auth = { api_token = var.flashblade_api_token } # or FLASHBLADE_API_TOKEN
  # auth = { oauth2 = { client_id = ..., key_id = ..., issuer = ... } }
}
```

Note the naming asymmetry: Terraform resource types keep the descriptive
`flashblade_*` prefix (like `aws_*`); the Pulumi package is `mica`.

## Develop (contributors)

```bash
make build     # go build -trimpath -o terraform-provider-mica
make test      # go test ./internal/... (TEST_BASELINE regression gate)
make lint      # golangci-lint run ./...
make docs      # go generate ./... (tfplugindocs — never hand-edit docs/)
make install   # into ~/.terraform.d/plugins/.../numberly/mica/dev/
```

The root build file is `GNUmakefile`. `make install-hooks` wires the
`commit-msg` hook that rejects `Co-Authored-By` trailers.

## Sections

- **[Architecture](architecture.md)** — client/auth/transport, generic CRUD
  helpers, model structs, provider registration, resource & data source patterns,
  soft-delete.
- **[Conventions](conventions.md)** — the strict rules every resource follows
  (file layout, pointer rules, plan modifiers, drift detection, schema upgraders).
- **[Testing](testing.md)** — three-tier strategy, mock HTTP server, baseline gate.
- **[API versioning](api-versioning.md)** — swagger→reference tooling, api-diff,
  the 6-phase api-upgrade, and the dual 2.22/2.23 release lines.
- **[Pulumi bridge](pulumi-bridge.md)** — how the TF provider is republished as
  the `mica:` Pulumi package.
- **[Workflow & release](workflow-and-release.md)** — the GSD `.planning`
  workflow, build/CI/release pipelines, and docs generation.

## Where to start when changing code

| Task | Start here |
|------|-----------|
| Add a resource/data source | [conventions.md](conventions.md) + `flashblade-resource-builder` skill; mirror `internal/provider/bucket_resource.go` |
| Modify a resource's schema | Bump `SchemaVersion` + write a state upgrader — [conventions.md](conventions.md) |
| Upgrade to a new API version | [api-versioning.md](api-versioning.md) (`api-upgrade` skill, 6 phases) |
| Touch client HTTP/auth | [architecture.md](architecture.md) → `internal/client/{client,auth,transport,errors}.go` |
| Update the Pulumi surface | [pulumi-bridge.md](pulumi-bridge.md) → `pulumi/provider/resources.go` |
