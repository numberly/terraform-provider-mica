# Stack Research

**Domain:** Terraform Provider (Go) — REST API wrapping, storage infrastructure
**Researched:** 2026-03-26 (base stack) / 2026-03-30 (milestone v2.1.1 — network interfaces) / 2026-04-21 (milestone pulumi-2.22.3 — Pulumi bridge)
**Confidence:** HIGH — all versions verified against official scaffolding go.mod (March 2026) and pkg.go.dev

---

## Milestone pulumi-2.22.3 Addendum: Pulumi Bridge Layer

> This section covers **only the new `./pulumi/` sub-directory stack**.
> The base provider stack (terraform-plugin-framework, golangci-lint, GoReleaser) is unchanged.
> Scope: Python + Go SDKs only. TypeScript, C#, Java — explicitly out of scope.

### Core Bridge Libraries

The bridge layer lives in `./pulumi/provider/go.mod` — a separate Go module from the root provider.

| Technology | Pinned Version | Purpose | Why |
|------------|---------------|---------|-----|
| `github.com/pulumi/pulumi-terraform-bridge/v3` | `v3.127.0` | Core bridge engine: `pkg/pf/tfgen` (build-time schema introspection) + `pkg/pf/tfbridge` (runtime gRPC server) | Only official bridge for terraform-plugin-framework providers. Must use `pkg/pf/*` — not the SDK v2 shim path. v3.127.0 verified on pkg.go.dev 2026-04-21. Canonical reference `pulumi-random` uses this version. |
| `github.com/pulumi/pulumi/sdk/v3` | `v3.231.0` | Pulumi Go SDK — `resource.PropertyMap`, `resource.ID`, tokens, secrets | Direct dependency of bridge; must be co-versioned with the bridge. v3.231.0 latest as of 2026-04-16. `pulumi-random` uses v3.228.0; use v3.231.0 for freshest release. |
| `github.com/pulumi/pulumi/pkg/v3` | `v3.231.0` | Pulumi package schema types (used by tfgen for `schema.json` generation) | Co-versioned with sdk/v3 — always bump both together on bridge upgrades. |
| `github.com/hashicorp/terraform-plugin-go` | `v0.31.0` | Low-level TF plugin protocol (indirect dep of bridge) | Pulled in transitively; explicit pin avoids surprises on `go mod tidy`. Go 1.25 required — matches our toolchain. |

**Version note:** Pre-existing research in `pulumi-bridge.md` cited bridge v3.126.0 and sdk v3.220.0. Live verification on 2026-04-21 shows bridge at v3.127.0 (published same day) and sdk at v3.231.0. Always match against `pulumi-random/provider/go.mod` as the canonical source.

### Mandatory Replace Directives

The bridge requires a `replace` directive for the Pulumi-maintained fork of the HashiCorp plugin SDK. Without it, the build fails — the standard HashiCorp SDK and the fork diverge at the protocol level.

```
# ./pulumi/provider/go.mod
require (
    github.com/hashicorp/terraform-plugin-sdk/v2 v2.0.0-20260318212141-5525259d096b
    github.com/pulumi/pulumi-terraform-bridge/v3 v3.127.0
    github.com/pulumi/pulumi/sdk/v3 v3.231.0
    github.com/pulumi/pulumi/pkg/v3 v3.231.0
)

replace (
    # Pulumi fork of the TF plugin SDK — mandatory, matches bridge go.mod
    github.com/hashicorp/terraform-plugin-sdk/v2 => github.com/pulumi/terraform-plugin-sdk/v2 v2.0.0-20260318212141-5525259d096b

    # Local path to the upstream TF provider (avoids publishing a new version on every change)
    github.com/soulkyu/terraform-provider-flashblade => ../

    # Local path to generated Go SDK (needed during development / Makefile build)
    github.com/soulkyu/pulumi-flashblade/sdk/go => ../sdk/go
)
```

**SHA hygiene:** The fork SHA `20260318212141-5525259d096b` was verified against both `pulumi-random` and `pulumi-cloudflare` go.mod files on 2026-04-21. This SHA changes with each bridge release. After every `go get github.com/pulumi/pulumi-terraform-bridge/v3@latest`, run:
```bash
grep "terraform-plugin-sdk" pulumi/provider/go.mod
```
and update the replace SHA to match what the bridge declares in its own go.mod.

### Python SDK Tooling

| Tool | Version | Purpose | Why Needed |
|------|---------|---------|-----------|
| `pulumictl` | `v0.0.50` | Computes semantic version string for LDFLAGS and Python package metadata (`pulumictl get version`) | All canonical Pulumi providers use this in Makefile. Without it, `$(VERSION)` in LDFLAGS and `setup.py`'s version field must be manually scripted. Binary: download from GitHub releases. |
| `build` (Python) | `1.2.1` | `python -m build` packages `sdk/python/` into a `.whl` | PEP 517-compliant build frontend. Used verbatim by `pulumi-cloudflare`'s `build_python` target. Replaces `setup.py bdist_wheel`. Produces the installable private `.whl` artifact. |

**pulumictl installation (CI + local):**
```bash
# Linux amd64
curl -sL https://github.com/pulumi/pulumictl/releases/download/v0.0.50/pulumictl-v0.0.50-linux-amd64.tar.gz \
  | tar xz -C ~/.local/bin
```

**Python build (called by `make build_python`):**
```bash
cd sdk/python && python3 -m venv venv && venv/bin/pip install build==1.2.1 && venv/bin/python -m build
```

**What `pulumictl get version` returns:** A semver string derived from the nearest git tag — e.g., `0.1.0-alpha.1+dev` for dev builds, `0.1.0` for release tags. This becomes the Python package version in `setup.py` and the Go `-X ...Version=` ldflags value.

### Go SDK Distribution (Private)

Go SDK distribution uses Git tags — no package registry required.

| Concern | Mechanism | Detail |
|---------|-----------|--------|
| Module path | `github.com/soulkyu/pulumi-flashblade/sdk/go` | Declared in `sdk/go/go.mod`; consumers import this path |
| Tag convention | `sdk/go/v0.1.0` | Separate tag from provider tag `v0.1.0` — required by Go module system for sub-directory modules |
| Private access | `GOPRIVATE=github.com/soulkyu/*` | Set in consumer environments; bypasses GOPROXY/GOSUM for this org |
| Sum DB bypass | `GONOSUMCHECK=github.com/soulkyu/*` | Needed alongside GOPRIVATE when the sum DB cannot reach private repos |
| CI build | `replace` directive in `pulumi/provider/go.mod` | Local path `../sdk/go` during build; no remote fetch needed in CI |

**Tag workflow (after goreleaser release):**
```bash
git tag sdk/go/v0.1.0
git push origin sdk/go/v0.1.0
```

**`sdk/go/go.mod`** (minimal — consumers import only pulumi/sdk):
```
module github.com/soulkyu/pulumi-flashblade/sdk/go

go 1.25

require (
    github.com/pulumi/pulumi/sdk/v3 v3.231.0
)
```

**Consumer go.mod** (for internal teams using the Go SDK):
```
require github.com/soulkyu/pulumi-flashblade/sdk/go v0.1.0

# Plus in environment:
# export GOPRIVATE=github.com/soulkyu/*
# export GONOSUMCHECK=github.com/soulkyu/*
```

### Python SDK Distribution (Private)

No PyPI required. Private distribution via GitHub release assets.

| Concern | Mechanism |
|---------|-----------|
| Package format | `.whl` built by `python -m build` from `sdk/python/` |
| Distribution | Attached to GitHub release as asset by goreleaser `extra_files` |
| Consumer install | `pip install https://github.com/soulkyu/pulumi-flashblade/releases/download/v0.1.0/pulumi_flashblade-0.1.0-py3-none-any.whl` |
| Package name | `pulumi-flashblade` (generated by tfgen from `ProviderInfo.Name = "flashblade"`) |
| No Twine | Not needed for private distribution |

### GoReleaser Configuration for Bridge

The bridge needs its own goreleaser config at `./pulumi/.goreleaser.yml`. The existing root `.goreleaser.yml` handles the Terraform provider binary — keep them separate.

**Archive naming convention** (Pulumi CLI requirement for `pulumi plugin install`):
```
pulumi-resource-{provider}-v{VERSION}-{os}-{arch}.tar.gz
```
Example: `pulumi-resource-flashblade-v0.1.0-linux-amd64.tar.gz`

**Minimal `./pulumi/.goreleaser.yml`:**
```yaml
project_name: pulumi-flashblade

before:
  hooks:
    - make -C . tfgen   # schema + bridge-metadata generated before binary build

builds:
  - id: pulumi-resource-flashblade
    dir: provider
    main: ./cmd/pulumi-resource-flashblade
    binary: pulumi-resource-flashblade
    ldflags:
      - -X github.com/soulkyu/pulumi-flashblade/provider/pkg/version.Version={{.Version}}
    goos: [linux, darwin, windows]
    goarch: [amd64, arm64]
    ignore:
      - goos: windows
        goarch: arm64

archives:
  - id: archive
    builds: [pulumi-resource-flashblade]
    name_template: "{{ .Binary }}-v{{ .Version }}-{{ .Os }}-{{ .Arch }}"
    format: tar.gz
    format_overrides:
      - goos: windows
        format: zip

signs:
  - cmd: cosign
    artifacts: all
    args:
      - sign-blob
      - --output-certificate=${certificate}
      - --output-signature=${signature}
      - ${artifact}
      - --yes

release:
  github:
    owner: soulkyu
    name: pulumi-flashblade
  draft: false
  prerelease: auto
  extra_files:
    - glob: sdk/python/dist/*.whl   # attach Python wheel to release
```

**`pluginDownloadURL` in `resources.go`:**
```go
PluginDownloadURL: "github://api.github.com/soulkyu",
```
Pulumi CLI 3.56.0+ understands this `github://` scheme and fetches the binary from GitHub releases of the named org. This is the correct private distribution mechanism — no Pulumi Registry needed.

### Go Module Layout for `./pulumi/`

```
./pulumi/
├── provider/
│   ├── go.mod          # module github.com/soulkyu/pulumi-flashblade/provider
│   ├── go.sum
│   ├── resources.go    # ProviderInfo, ShimProvider wiring, mappings, overrides
│   ├── pkg/version/version.go   # ldflags -X injection target
│   └── cmd/
│       ├── pulumi-tfgen-flashblade/main.go        # pf/tfgen.Main entry
│       └── pulumi-resource-flashblade/
│           ├── main.go                             # pf/tfbridge.Main entry
│           ├── schema.json                         # generated + committed (for CI diff)
│           ├── schema-embed.json                   # generated + committed (//go:embed)
│           └── bridge-metadata.json                # generated + committed (//go:embed)
└── sdk/
    ├── go/
    │   ├── go.mod      # module github.com/soulkyu/pulumi-flashblade/sdk/go
    │   └── flashblade/ # generated by tfgen go subcommand
    └── python/
        ├── setup.py    # generated by tfgen python subcommand
        ├── pyproject.toml
        └── pulumi_flashblade/
```

**Why `schema.json` AND `schema-embed.json`:** `schema.json` is committed for human-readable CI diffs on PR. `schema-embed.json` is embedded via `//go:embed` into the runtime binary — this enables `pulumi plugin install` to introspect the schema without running tfgen. Both must be committed; both are regenerated by `make tfgen`.

**Why three separate `go.mod` files:**
- Root `go.mod` — the Terraform provider (unchanged)
- `pulumi/provider/go.mod` — bridge + binaries (depends on bridge/v3, our TF provider via replace)
- `pulumi/sdk/go/go.mod` — consumer-facing Go SDK (depends only on pulumi/sdk/v3, no bridge dep)

Keeping the SDK module lean avoids forcing bridge transitive deps on end users.

### Wiring: PF Bridge Entry Points

`provider/cmd/pulumi-tfgen-flashblade/main.go`:
```go
package main

import (
    "github.com/pulumi/pulumi-terraform-bridge/v3/pkg/pf/tfgen"
    flashblade "github.com/soulkyu/pulumi-flashblade/provider"
)

func main() {
    // Note: no version parameter — PF tfgen differs from SDK v2 tfgen here
    tfgen.Main("flashblade", flashblade.Provider())
}
```

`provider/cmd/pulumi-resource-flashblade/main.go`:
```go
package main

import (
    "context"
    _ "embed"

    pftfbridge "github.com/pulumi/pulumi-terraform-bridge/v3/pkg/pf/tfbridge"
    flashblade "github.com/soulkyu/pulumi-flashblade/provider"
)

//go:embed schema-embed.json
var schema []byte

//go:embed bridge-metadata.json
var metadata []byte

func main() {
    meta := pftfbridge.ProviderMetadata{PackageSchema: schema}
    pftfbridge.Main(context.Background(), "flashblade", flashblade.Provider(), meta)
}
```

Key differences from SDK v2 bridge: `context.Context` is required in `tfbridge.Main`, `Version` and `MetadataInfo` are mandatory in `ProviderInfo`, and `pf.ShimProvider(...)` replaces `shimv2.NewProvider(...)`.

### Tools Summary

| Tool | Version | Purpose | Notes |
|------|---------|---------|-------|
| `pulumictl` | v0.0.50 | Version injection | Download binary from GitHub releases |
| `python build` | 1.2.1 | Python wheel packaging | Install inside virtualenv per `make build_python` |
| `golangci-lint` | existing (reuse) | Lint `./pulumi/provider/...` | Same binary, run separately from root |
| `pulumi` CLI | compatible with sdk v3.231.0 | ProgramTest + smoke tests | Install in CI alongside Go toolchain |
| `tfplugindocs` | N/A — NOT used | Terraform docs tool | Pulumi uses tfgen for its docs, not tfplugindocs |

### Alternatives Considered

| Recommended | Alternative | Why Not |
|-------------|-------------|---------|
| `pkg/pf/tfgen` + `pkg/pf/tfbridge` | `pkg/tfgen` + `pkg/tfbridge` (SDK v2 shim) | Our provider is terraform-plugin-framework. SDK v2 shim would require `muxer` wrapping — extra complexity, no benefit. `pkg/pf/*` is the correct and supported path. |
| `pulumictl` for version | Manual `git describe --tags` | `pulumictl` handles pre-release semver suffixes (`+dev`, `-alpha.1`) consistently across all Makefile targets. Manual scripting is fragile at edge cases (no tags, dirty tree). |
| `./pulumi/` monorepo sub-dir | Separate repository (`github.com/soulkyu/pulumi-flashblade`) | Avoids maintaining two repos; `replace` directive can point to `../` for the TF provider instead of requiring a published release. Correct for private/alpha stage. |
| GitHub releases (private) | Pulumi Registry (public) | Out of scope per milestone. Registry requires `pulumi/pulumi-package-publisher` action and public schema publication. |
| Single `.whl` attached to release | PyPI / TestPyPI | Private distribution requirement — no public index needed. |

### What NOT to Use (Bridge-Specific)

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| `pkg/tfgen` (non-pf path) | SDK v2 shim — wrong for terraform-plugin-framework providers | `pkg/pf/tfgen` |
| `shimv2.NewProvider(...)` | SDK v2 constructor | `pf.ShimProvider(fb.New(version.Version)())` |
| TypeScript/C#/Java SDK targets in Makefile | Out of scope; adds build time and CI complexity with no consumer | Python + Go only; comment out or omit `generate_nodejs`, `generate_dotnet`, `generate_java` |
| `GONOSUMDB` (wrong env var) | Different from GONOSUMCHECK | Use `GONOSUMCHECK=github.com/soulkyu/*` (or `GONOSUMDB` depending on Go version — check `go env` output) |
| Attaching the Go SDK as a binary artifact | Go SDK is consumed via Git tag + GOPRIVATE, not as a .tar.gz | Tag `sdk/go/vX.Y.Z` and push; consumers `go get` directly |

### Version Compatibility Matrix

| Component | Version | Compatible With |
|-----------|---------|-----------------|
| `pulumi-terraform-bridge/v3` | v3.127.0 | terraform-plugin-framework v1.19.0, pulumi/sdk v3.228+ |
| `pulumi/sdk/v3` + `pulumi/pkg/v3` | v3.231.0 | bridge v3.127.0 (bridge declares min SDK in its go.mod) |
| `terraform-plugin-framework` | v1.19.0 (indirect) | Provider already on v1.19.0 (from base stack) — no upgrade needed |
| `terraform-plugin-go` | v0.31.0 | Go 1.25 required — matches our toolchain exactly |
| `pulumictl` | v0.0.50 | Stateless CLI; no Go/framework coupling |
| Go toolchain | 1.25 (existing) | Bridge v3.127.0 requires Go 1.22+ — Go 1.25 fully compatible |

### Sources (Bridge Addendum)

- `pkg.go.dev/github.com/pulumi/pulumi-terraform-bridge/v3` — v3.127.0 verified 2026-04-21 (HIGH)
- `pkg.go.dev/github.com/pulumi/pulumi/sdk/v3` — v3.231.0 verified 2026-04-16 (HIGH)
- `pkg.go.dev/github.com/pulumi/pulumi/pkg/v3` — v3.231.0 verified 2026-04-16 (HIGH)
- `pkg.go.dev/github.com/hashicorp/terraform-plugin-go` — v0.31.0 verified 2026-03-10 (HIGH)
- `pkg.go.dev/github.com/pulumi/pulumi-terraform-bridge/v3/pkg/pf` — pf package GA, ShimProvider interface marked unstable (HIGH)
- `github.com/pulumi/pulumictl/releases` — v0.0.50 latest (HIGH)
- `github.com/pulumi/pulumi-random/provider/go.mod` — canonical PF reference provider; bridge v3.127.0, sdk v3.228.0, tfp-framework v1.19.0, replace SHA `20260318212141-5525259d096b` (HIGH)
- `github.com/pulumi/pulumi-cloudflare/provider/go.mod` — bridge v3.125.0, sdk v3.226.0, same replace SHA + local replace pattern (HIGH)
- `github.com/pulumi/pulumi-cloudflare/Makefile` — `build_python` target using `build==1.2.1` (HIGH)
- `pulumi-bridge.md` (repo root) — pre-existing consolidated research; versions updated, architecture validated (MEDIUM-HIGH; now upgraded with live verification)
- WebSearch/WebFetch: archive naming `pulumi-resource-{name}-v{version}-{os}-{arch}.tar.gz` — pattern consistent across pulumi-consul, pulumi-random observed releases (MEDIUM)
- `github.com/pulumi/pulumi/issues/8944`, `#9007` — `github://api.github.com/<org>` pluginDownloadURL support for private releases (HIGH — merged, available since Pulumi CLI 3.56.0)

---

## Milestone v2.1.1 Addendum: Network Interface (VIP) Stack

> This section covers **only what is new or different** for the network interface milestone.
> The base stack (framework, oauth2, testing, tooling) is unchanged — see section below.

### No New Library Dependencies

The VIP resource requires zero new `go.mod` entries. All necessary primitives already exist:

| Capability Needed | Already Available Via |
|-------------------|----------------------|
| HTTP CRUD to `/api/2.22/network-interfaces` | `FlashBladeClient.post/patch/delete/getOneByName` — same HTTP client used by all 29+ resources |
| `?names=` query parameter pattern | Identical to `GetServer`, `PatchServer`, `DeleteServer` in `internal/client/servers.go` |
| `attached_servers []NamedReference` field | `NamedReference` struct in `models_common.go`, already used in `ObjectStoreVirtualHost` |
| `services []string` field | Same pattern as `SyslogServer.Services`, `ArrayDns.Services` |
| Nested `subnet` object (read-only) | `types.ObjectType` with `attr.Type` map — same pattern as `serverDNSObjectType()` in `server_resource.go` |
| Read-only fields (`Computed: true` only) | `stringplanmodifier.UseStateForUnknown()` — already used on `id`, `name` fields across all resources |
| `types.List` for `realms` (read-only array) | `types.ListValueFrom(ctx, types.StringType, ...)` — used in `mapServerToModel` for DNS |
| Drift detection logging | `tflog.Info(ctx, "drift detected ...", map[string]any{...})` — pattern from `object_store_virtual_host_resource.go` |
| Import state by name | `resource.ResourceWithImportState` — all existing resources implement this |
| Three-tier testing | Unit + `httptest.NewServer` mock + acceptance — established pattern, no new tooling |

### New Models to Create

Add to `internal/client/models_network.go` (new file, mirrors `models_admin.go` / `models_exports.go` naming convention):

```go
// NetworkInterfaceSubnet is the read-only subnet reference embedded in NetworkInterface.
// The API returns it as an object; only the name is useful for consumers.
type NetworkInterfaceSubnet struct {
    Name string `json:"name,omitempty"`
    ID   string `json:"id,omitempty"`
}

// NetworkInterface represents a FlashBlade network interface from GET /network-interfaces.
// Fields marked ro are returned by the API but cannot be set on POST/PATCH.
type NetworkInterface struct {
    ID              string                  `json:"id,omitempty"`
    Name            string                  `json:"name,omitempty"`       // ro — assigned by API
    Type            string                  `json:"type,omitempty"`       // "vip"
    Address         string                  `json:"address,omitempty"`
    Subnet          *NetworkInterfaceSubnet `json:"subnet,omitempty"`
    Services        []string                `json:"services,omitempty"`
    AttachedServers []NamedReference        `json:"attached_servers,omitempty"`
    Enabled         bool                    `json:"enabled,omitempty"`    // ro
    Gateway         string                  `json:"gateway,omitempty"`    // ro
    MTU             int64                   `json:"mtu,omitempty"`        // ro
    Netmask         string                  `json:"netmask,omitempty"`    // ro
    VLAN            int64                   `json:"vlan,omitempty"`       // ro
    Realms          []NamedReference        `json:"realms,omitempty"`     // ro
}

// NetworkInterfacePost contains fields accepted on POST /network-interfaces.
// Name is passed via ?names= query parameter, not in the body.
// type="vip" is always set; subnet is required on creation.
type NetworkInterfacePost struct {
    Type            string                  `json:"type"`
    Address         string                  `json:"address,omitempty"`
    Subnet          *NetworkInterfaceSubnet `json:"subnet,omitempty"`
    Services        []string                `json:"services,omitempty"`
    AttachedServers []NamedReference        `json:"attached_servers,omitempty"`
}

// NetworkInterfacePatch contains only the mutable fields for PATCH /network-interfaces.
// The API accepts: address, attached_servers, services (verified from FLASHBLADE_API.md).
// All other fields in the response are read-only.
type NetworkInterfacePatch struct {
    Address         *string          `json:"address,omitempty"`
    Services        *[]string        `json:"services,omitempty"`
    AttachedServers []NamedReference `json:"attached_servers,omitempty"`
}
```

### New Client Functions to Create

Add to `internal/client/network_interfaces.go` (new file, mirrors `servers.go`):

```go
func (c *FlashBladeClient) GetNetworkInterface(ctx context.Context, name string) (*NetworkInterface, error)
func (c *FlashBladeClient) PostNetworkInterface(ctx context.Context, name string, body NetworkInterfacePost) (*NetworkInterface, error)
func (c *FlashBladeClient) PatchNetworkInterface(ctx context.Context, name string, body NetworkInterfacePatch) (*NetworkInterface, error)
func (c *FlashBladeClient) DeleteNetworkInterface(ctx context.Context, name string) error
```

All four follow the exact same `?names=` / `ListResponse[T]` pattern as `servers.go`.

### Schema Design Decisions for VIP Fields

| Field | Schema Type | Rationale |
|-------|------------|-----------|
| `address` | `StringAttribute{Required: true}` | User-supplied IP address, mutable via PATCH |
| `subnet` | `StringAttribute{Required: true}` | User supplies subnet name on create; read back as computed object — expose only the name as a string, not a nested object (simpler DX, consistent with `server` field patterns on exports) |
| `services` | `ListAttribute{ElementType: StringType, Optional: true, Computed: true}` | Mutable via PATCH, API may set defaults |
| `attached_servers` | `ListAttribute{ElementType: StringType, Optional: true, Computed: true}` | Full-replace semantics on PATCH — same as `ObjectStoreVirtualHost.AttachedServers` |
| `type` | `StringAttribute{Computed: true}` with hardcoded `"vip"` on POST | Always "vip"; Computed avoids user confusion, UseStateForUnknown for stability |
| `enabled` | `BoolAttribute{Computed: true}` | Read-only from API, UseStateForUnknown |
| `gateway` | `StringAttribute{Computed: true}` | Read-only, derived from subnet |
| `mtu` | `Int64Attribute{Computed: true}` | Read-only, derived from subnet |
| `netmask` | `StringAttribute{Computed: true}` | Read-only, derived from subnet |
| `vlan` | `Int64Attribute{Computed: true}` | Read-only, derived from subnet |
| `realms` | `ListAttribute{ElementType: StringType, Computed: true}` | Read-only array of realm names |

**Key decision on `subnet`:** The API accepts `subnet: {name: "...", id: "..."}` on POST but returns the full object on GET. Expose as `subnet_name` (string) on the Terraform resource — same pattern used for `server` references on exports (`ObjectStoreAccountExport.Server` → Terraform `server_name` string). This avoids a nested object block that users cannot modify.

### Server Resource/Data Source Enrichment

For v2.1.1, the server data source needs a `network_interfaces` computed list showing VIPs associated to a server. The FlashBlade API does not provide a direct `/servers/{id}/network-interfaces` endpoint. The only approach is:

1. `GET /network-interfaces?filter=name='...'` (if API supports filter on attached_servers) — verify during implementation
2. List all network interfaces and filter client-side by `attached_servers[].name == server_name`

**Flag:** This filtering approach must be validated against API capabilities before implementation. If `GET /network-interfaces` does not support filtering by `attached_servers`, client-side filtering on a full list scan is acceptable for small environments but could be slow on large arrays. Mark as a phase-specific research item.

---

## Base Stack (Unchanged from v1.0)

### Core Technologies

| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| Go | 1.25.x | Implementation language | Only supported language for Terraform providers. v1.25 is current LTS and required minimum for all terraform-plugin-* modules as of March 2026. The official scaffolding go.mod pins `go 1.25.5`. |
| terraform-plugin-framework | v1.19.0 | Provider CRUD scaffold, schema, plan modifiers | HashiCorp's current recommended SDK. Protocol v6, better type safety than SDKv2, native plan modifiers, built-in diagnostics, required for all new providers. SDKv2 is in maintenance-only mode. |
| terraform-plugin-go | v0.31.0 | Low-level Terraform protocol bindings | Indirect dependency of framework; occasionally needed directly for custom type implementations or protocol-level hooks. Listed explicitly for version pinning. |
| terraform-plugin-testing | v1.15.0 | Acceptance and integration test runner | The official test harness that runs real `terraform` CLI against provider code. Required for acceptance tests. Provides `resource.Test`, `TestCase`, `TestStep` constructs. |
| terraform-plugin-log | v0.10.0 | Structured logging inside provider | HashiCorp's structured logging for providers. Integrates with Terraform's log streaming. Use `tflog` package, not `log` stdlib. |

### Supporting Libraries

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| golang.org/x/oauth2 | v0.27+ | OAuth2 client_credentials token flow | FlashBlade `POST /oauth2/1.0/token` token exchange. `oauth2/clientcredentials.Config.Client()` returns an auto-refreshing `*http.Client`. Use for production auth. |
| golang.org/x/oauth2/clientcredentials | (same module) | Two-legged OAuth2 subpackage | Specifically for client_credentials grant. Avoids hand-rolling token refresh logic. |
| net/http (stdlib) | — | HTTP transport base | All API calls. Build a custom `http.RoundTripper` chain: TLS config → auth injection → request tracing. Do NOT use third-party HTTP clients. |
| crypto/tls (stdlib) | — | Custom CA certificate support | Enterprise environments with private PKI. `tls.Config{RootCAs: pool}` injected into `http.Transport`. |
| crypto/x509 (stdlib) | — | CA cert pool construction | `x509.SystemCertPool()` + `pool.AppendCertsFromPEM()` for custom CA. |
| encoding/json (stdlib) | — | JSON serialisation/deserialisation | FlashBlade API speaks JSON throughout. Keep stdlib; no need for third-party JSON library. |
| github.com/hashicorp/terraform-plugin-docs | v0.21+ | Registry documentation generation | `tfplugindocs generate` produces Terraform Registry-compatible Markdown from schema descriptions and templates. Run as part of CI/pre-release. |
| net/http/httptest (stdlib) | — | Integration test mock HTTP server | `httptest.NewServer()` for mocked-API integration tests (the "middle tier" without a real FlashBlade). Zero external dependencies. |

### Development Tools

| Tool | Purpose | Notes |
|------|---------|-------|
| goreleaser | Cross-platform binary release | HashiCorp-mandated release pipeline. Builds linux/darwin/windows amd64/arm64 binaries, signs checksums with GPG. Copy `.goreleaser.yml` from `hashicorp/terraform-provider-scaffolding-framework`. |
| tfplugindocs (`github.com/hashicorp/terraform-plugin-docs`) | Docs generation CLI | `go generate ./...` with `//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs`. Install via `tools.go` pattern. |
| golangci-lint v2 | Go linting | 40+ linters in one binary. v2 released March 2025 with revamped config (`.golangci.yml`). Use `linters.default: standard` then enable extras. |
| tfproviderlint (`github.com/bflad/tfproviderlint`) | Terraform-specific Go lint | Catches provider-specific mistakes: missing Read calls, import state patterns, schema issues. Complements golangci-lint. Run `tfproviderlintx` for extended checks. |
| Terraform CLI | Acceptance test runner | `terraform-plugin-testing` shells out to a real `terraform` binary. Set `TF_ACC=1` env var to enable acceptance tests. Pin version via `TF_CLI_PATH` or GitHub Actions setup step. |
| GitHub Actions (`hashicorp/ghaction-terraform-provider-release`) | CI/CD | HashiCorp's reusable workflow for provider release. Handles GoReleaser, GPG signing, Registry publish. |

## Installation

```bash
# No new dependencies for v2.1.1 milestone — all packages already in go.mod.

# Verify current versions are pinned:
go list -m github.com/hashicorp/terraform-plugin-framework   # expect v1.19.0
go list -m github.com/hashicorp/terraform-plugin-testing     # expect v1.15.0
go list -m golang.org/x/oauth2                               # expect v0.34.0+

# New files to create (no go get needed):
# internal/client/models_network.go
# internal/client/network_interfaces.go
# internal/provider/network_interface_resource.go
# internal/provider/network_interface_data_source.go
```

## Alternatives Considered

| Recommended | Alternative | When to Use Alternative |
|-------------|-------------|-------------------------|
| terraform-plugin-framework v1.19 | terraform-plugin-sdk/v2 | Never for new providers. SDKv2 is feature-frozen, maintenance-only. Only valid when maintaining a legacy provider that cannot be migrated. |
| stdlib net/http | resty, go-resty, cleanhttp | Only if the API requires complex retry logic not worth handrolling. hashicorp/go-retryablehttp is an acceptable complement for retry-with-backoff on transient errors. |
| httptest.NewServer (stdlib) | WireMock, httpmock jarcoal | WireMock is Java-native, adds container overhead. jarcoal/httpmock is reasonable but httptest.NewServer is zero-dep and idiomatic Go for this use case. |
| golang.org/x/oauth2 | Hand-rolled token refresh | Never hand-roll. x/oauth2 handles refresh, concurrency, and expiry correctly. Its `clientcredentials` sub-package covers the FlashBlade token exchange exactly. |
| goreleaser | Manual Makefile multi-platform build | Only for internal-only providers that never publish to Registry. If Registry publication is planned (it is here), use goreleaser from day one. |
| `subnet_name` string attribute | Nested `subnet` object block | The subnet object is read-only after creation; nesting adds schema complexity with zero benefit. A string attribute `subnet_name` is consistent with how all other resource references are handled in this provider. |

## What NOT to Use

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| terraform-plugin-sdk/v2 | Feature-frozen, no new framework features (plan modifiers, write-only attrs, identity). HashiCorp's official position: migrate away. | terraform-plugin-framework v1.19 |
| terraform-plugin-mux | Only needed when combining framework + SDKv2 providers in one binary. Adds complexity with no benefit for pure-framework providers. | Direct framework provider, no mux needed |
| Third-party HTTP clients (resty, fiber, etc.) | Adds deps, hides transport layer — breaks custom CA/auth RoundTripper chain. stdlib net/http is sufficient and transparent. | stdlib net/http with RoundTripper chain |
| `log` stdlib inside provider | Terraform has its own structured log capture. stdlib `log` writes bypass Terraform's log routing and appear as raw stderr noise. | `github.com/hashicorp/terraform-plugin-log/tflog` |
| Hardcoded API version paths | FlashBlade exposes `/api/2.22/...` — hardcoding across 538 endpoints creates drift risk when upgrading. | Const for API version, assembled in HTTP client constructor |
| Nested `subnet` object block in schema | The subnet ref is write-once on POST and read-only thereafter. A `schema.SingleNestedAttribute` block would require the user to specify `subnet { name = "..." }` but never be able to modify it. | `subnet_name` as a `StringAttribute` with `RequiresReplace()` plan modifier |
| Filtering network interfaces via `attached_servers` at read time | The API may not support filtering network interfaces by attached server names. Doing a full list scan in `Read` on every refresh adds latency and is fragile. | Only look up by name in the resource Read; expose VIPs on the server data source via a dedicated computed attribute populated from a scoped list call |

## Stack Patterns by Variant

**For API token auth (dev/local):**
- Build an `http.RoundTripper` that injects `x-auth-token` header
- Configured via provider block `api_token` attribute
- Use `stringplanmodifier.UseStateForUnknown()` on computed-only fields

**For OAuth2 client_credentials auth (production):**
- Use `golang.org/x/oauth2/clientcredentials.Config{...}.TokenSource(ctx)`
- Wrap with `oauth2.NewClient(ctx, tokenSource)` which auto-refreshes
- FlashBlade token endpoint: `POST /oauth2/1.0/token` with `grant_type=urn:ietf:params:oauth:grant-type:token-exchange`

**For custom CA certificate (enterprise TLS):**
```go
pool, _ := x509.SystemCertPool()
pool.AppendCertsFromPEM(customCAPEM)
transport := &http.Transport{
    TLSClientConfig: &tls.Config{RootCAs: pool},
}
```
Layer this transport under the auth RoundTripper via composition.

**For mocked integration tests (CI without FlashBlade):**
- Use `httptest.NewServer(mux)` where `mux` is an `http.ServeMux` recording expected calls
- Provider configured to point at `httptest.Server.URL`
- Run with `go test ./... -run TestIntegration` (no `TF_ACC=1`)

**For acceptance tests (real FlashBlade required):**
- Set `TF_ACC=1`, `FLASHBLADE_ENDPOINT`, `FLASHBLADE_API_TOKEN`
- Use `resource.Test(t, resource.TestCase{...})` from `terraform-plugin-testing`
- Gate in CI: run only on manual trigger or dedicated FlashBlade environment

**For VIP `subnet_name` with RequiresReplace:**
- Mark `subnet_name` with `stringplanmodifier.RequiresReplace()` — changing the subnet on a VIP requires destroy+create
- Mark `type` with `stringplanmodifier.UseStateForUnknown()` and hard-code `"vip"` in `NetworkInterfacePost.Type` — the provider only manages VIP type interfaces
- All read-only fields (gateway, mtu, netmask, vlan, enabled, realms) get `UseStateForUnknown()` to avoid perpetual plan diffs on the first apply

**For `attached_servers` full-replace semantics:**
- On PATCH, always send the complete desired list — same behavior as `ObjectStoreVirtualHost`
- Empty slice `[]NamedReference{}` means detach all servers
- `nil` in the PATCH body means "don't change" — use pointer-to-slice if omit-vs-empty matters; for this field the list-replace approach is simpler

## Version Compatibility

| Package | Compatible With | Notes |
|---------|-----------------|-------|
| terraform-plugin-framework v1.19.0 | Go 1.25+, terraform-plugin-go v0.31.0 | Sourced from official scaffolding go.mod (March 10, 2026) |
| terraform-plugin-testing v1.15.0 | Go 1.25+, Terraform CLI 1.0+ | Same release date as framework v1.19.0 — released together |
| terraform-plugin-log v0.10.0 | Go 1.25+ | Indirect via framework; pin explicitly to avoid transitive bumps |
| golang.org/x/oauth2 | Go 1.21+ | Follows x/ library policy; no breaking changes expected |
| golangci-lint v2 | Go 1.22+ for targets | v2 config format incompatible with v1 `.golangci.yml` — write fresh config |

## Sources

- [terraform-plugin-framework v1.19.0 — pkg.go.dev](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-framework) — version and Go 1.25 requirement confirmed HIGH confidence
- [terraform-plugin-testing v1.15.0 — pkg.go.dev](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-testing) — version and release date confirmed HIGH confidence
- [terraform-provider-scaffolding-framework go.mod](https://raw.githubusercontent.com/hashicorp/terraform-provider-scaffolding-framework/main/go.mod) — authoritative dependency pinning, March 2026 HIGH confidence
- [HashiCorp Plugin Framework benefits](https://developer.hashicorp.com/terraform/plugin/framework-benefits) — SDKv2 deprecation rationale HIGH confidence
- [golang.org/x/oauth2/clientcredentials](https://pkg.go.dev/golang.org/x/oauth2/clientcredentials) — token flow pattern HIGH confidence
- [terraform-plugin-docs GitHub](https://github.com/hashicorp/terraform-plugin-docs) — tfplugindocs CLI usage MEDIUM confidence (version unverified)
- [bflad/tfproviderlint GitHub](https://github.com/bflad/tfproviderlint) — provider-specific linter MEDIUM confidence (last verified active)
- [golangci-lint v2 announcement](https://ldez.github.io/blog/2025/03/23/golangci-lint-v2/) — v2 config format change MEDIUM confidence
- [hashicorp/ghaction-terraform-provider-release](https://github.com/hashicorp/ghaction-terraform-provider-release) — official release workflow HIGH confidence
- FLASHBLADE_API.md lines 372-377 — network-interfaces CRUD endpoints, mutable vs read-only fields confirmed HIGH confidence (authoritative API reference)
- `internal/client/models_exports.go` — NamedReference, AttachedServers pattern confirmed HIGH confidence (codebase)
- `internal/provider/object_store_virtual_host_resource.go` — attached_servers full-replace pattern confirmed HIGH confidence (codebase)
- `internal/provider/server_resource.go` — nested object (serverDNSObjectType) pattern confirmed HIGH confidence (codebase)

---
*Stack research for: Terraform Provider for Pure Storage FlashBlade*
*Researched: 2026-03-26 (base) / 2026-03-30 (v2.1.1 network interfaces addendum) / 2026-04-21 (pulumi-2.22.3 bridge addendum)*
