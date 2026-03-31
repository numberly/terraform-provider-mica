# Stack Research

**Domain:** Terraform Provider (Go) — REST API wrapping, storage infrastructure
**Researched:** 2026-03-26 (base stack) / 2026-03-30 (milestone v2.1.1 — network interfaces)
**Confidence:** HIGH — all versions verified against official scaffolding go.mod (March 2026) and pkg.go.dev

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
*Researched: 2026-03-26 (base) / 2026-03-30 (v2.1.1 network interfaces addendum)*
