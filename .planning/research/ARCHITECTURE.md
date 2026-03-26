# Architecture Research

**Domain:** Terraform provider for REST API (Pure Storage FlashBlade)
**Researched:** 2026-03-26
**Confidence:** HIGH

## Standard Architecture

### System Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                      Terraform Core Process                      │
│  (plan / apply / destroy / refresh / import)                    │
└─────────────────────┬───────────────────────────────────────────┘
                       │ gRPC (terraform-plugin-go protocol)
┌─────────────────────▼───────────────────────────────────────────┐
│                   Provider Server (main.go)                      │
│              providerserver.Serve(factory)                       │
├─────────────────────────────────────────────────────────────────┤
│                   Provider Layer (internal/provider/)            │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  FlashBladeProvider (provider.Provider)                  │   │
│  │  - Schema: endpoint, api_token, oauth2_*, tls_*          │   │
│  │  - Configure: build FlashBladeClient, inject to all      │   │
│  │  - Resources(): []resource.Resource factories            │   │
│  │  - DataSources(): []datasource.DataSource factories      │   │
│  └──────────────────────────────────────────────────────────┘   │
├──────────────────┬──────────────────────────────────────────────┤
│  Resources       │  Data Sources                                 │
│  ┌─────────────┐ │ ┌─────────────────┐                          │
│  │ filesystem  │ │ │ filesystem (ro) │                          │
│  │ bucket      │ │ │ bucket (ro)     │                          │
│  │ nfs_policy  │ │ │ nfs_policy (ro) │   (one file per         │
│  │ smb_policy  │ │ │ ...             │    resource type)        │
│  │ ...         │ │ └─────────────────┘                          │
│  └─────────────┘ │                                               │
├─────────────────────────────────────────────────────────────────┤
│                   Client Layer (internal/client/)                │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  FlashBladeClient                                         │   │
│  │  - Auth: API token session + OAuth2 token exchange        │   │
│  │  - HTTP: retries, TLS (custom CA), request-id headers    │   │
│  │  - API versioning: /api/2.22/ prefix on all calls        │   │
│  │  - Resource methods: FileSystems, Buckets, Policies, ... │   │
│  └──────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
                       │ HTTPS REST (JSON)
┌─────────────────────▼───────────────────────────────────────────┐
│                   FlashBlade REST API v2.22                      │
│          https://{array}/api/2.22/{resource}                    │
└─────────────────────────────────────────────────────────────────┘
```

### Component Responsibilities

| Component | Responsibility | Typical Implementation |
|-----------|----------------|------------------------|
| `main.go` | Provider binary entry point, gRPC server setup | `providerserver.Serve()` with optional debug flag |
| `internal/provider/provider.go` | Schema, Configure, resource/datasource registry | Implements `provider.Provider` interface |
| `internal/provider/*_resource.go` | CRUD + Import for one resource type | Implements `resource.Resource` interface |
| `internal/provider/*_data_source.go` | Read-only listing/lookup for one resource type | Implements `datasource.DataSource` interface |
| `internal/client/client.go` | HTTP client, auth, retries, base transport | Thin, stateless; accepts context |
| `internal/client/*_ops.go` | Per-domain API call methods (typed request/response) | Methods on `FlashBladeClient` |
| `examples/` | HCL usage examples (used by tfplugindocs) | Real `.tf` files exercising each resource |
| `docs/` | Generated provider documentation | Built by `tfplugindocs`, never edited manually |
| `tools/` | `go:generate` tooling (tfplugindocs, mockgen) | `tools.go` with blank imports |

## Recommended Project Structure

```
terraform-provider-flashblade/
├── main.go                          # Binary entry point
├── go.mod                           # module github.com/soulkyu/terraform-provider-flashblade
├── go.sum
├── GNUmakefile                      # make build, test, testacc, docs, lint
├── .goreleaser.yml                  # Release binary packaging
├── .golangci.yml                    # Linter config
├── terraform-registry-manifest.json # Registry metadata
│
├── internal/
│   ├── provider/
│   │   ├── provider.go              # FlashBladeProvider: Schema, Configure, Resources, DataSources
│   │   ├── provider_test.go         # Provider acceptance test bootstrap
│   │   │
│   │   ├── filesystem_resource.go        # flashblade_filesystem CRUD + Import
│   │   ├── filesystem_resource_test.go
│   │   ├── filesystem_data_source.go     # flashblade_filesystem data source (list/lookup)
│   │   ├── filesystem_data_source_test.go
│   │   │
│   │   ├── bucket_resource.go
│   │   ├── bucket_data_source.go
│   │   ├── object_store_account_resource.go
│   │   ├── object_store_access_key_resource.go
│   │   │
│   │   ├── nfs_export_policy_resource.go        # Policy parent resource
│   │   ├── nfs_export_policy_rule_resource.go   # Policy rule as child resource
│   │   ├── smb_share_policy_resource.go
│   │   ├── smb_share_policy_rule_resource.go
│   │   ├── snapshot_policy_resource.go
│   │   ├── snapshot_policy_rule_resource.go
│   │   ├── object_store_access_policy_resource.go
│   │   ├── object_store_access_policy_rule_resource.go
│   │   ├── network_access_policy_resource.go
│   │   ├── network_access_policy_rule_resource.go
│   │   ├── quota_policy_resource.go
│   │   ├── quota_policy_rule_resource.go
│   │   │
│   │   ├── array_dns_resource.go            # Array admin: DNS (singleton-style)
│   │   ├── array_ntp_resource.go
│   │   ├── array_smtp_resource.go
│   │   ├── array_alert_watcher_resource.go
│   │   │
│   │   └── testutils_test.go        # Shared acceptance test helpers
│   │
│   ├── client/
│   │   ├── client.go                # FlashBladeClient struct, NewClient, auth flow
│   │   ├── auth.go                  # API token session + OAuth2 token exchange
│   │   ├── transport.go             # http.RoundTripper: retries, X-Request-ID, logging
│   │   ├── errors.go                # API error type, HTTP status mapping
│   │   ├── filesystems.go           # FileSystems CRUD methods
│   │   ├── buckets.go               # Buckets CRUD methods
│   │   ├── object_store_accounts.go
│   │   ├── object_store_access_keys.go
│   │   ├── policies_nfs.go          # NFS export policy + rule methods
│   │   ├── policies_smb.go
│   │   ├── policies_snapshot.go
│   │   ├── policies_object_store_access.go
│   │   ├── policies_network_access.go
│   │   ├── policies_quota.go
│   │   ├── array_admin.go           # DNS, NTP, SMTP, alerts
│   │   └── models.go                # Shared Go structs mapping FlashBlade API JSON
│   │
│   └── testmock/
│       ├── server.go                # httptest server factory for integration tests
│       └── handlers/
│           ├── filesystems.go       # Handlers per endpoint group
│           ├── buckets.go
│           └── ...
│
├── examples/
│   ├── provider/
│   │   └── provider.tf              # Provider block example
│   ├── resources/
│   │   ├── flashblade_filesystem/
│   │   │   └── resource.tf
│   │   ├── flashblade_bucket/
│   │   │   └── resource.tf
│   │   └── ...
│   └── data-sources/
│       └── flashblade_filesystem/
│           └── data-source.tf
│
├── docs/                            # Generated by tfplugindocs — do not edit
│   ├── index.md
│   ├── resources/
│   └── data-sources/
│
└── tools/
    └── tools.go                     # blank imports: tfplugindocs, mockgen
```

### Structure Rationale

- **`internal/client/`:** Keeps all HTTP/API concerns out of provider logic. Resources are thin adapters that call client methods — they contain no raw `http.Client` calls. This boundary is critical for testability (mock the client, not HTTP).
- **`internal/provider/`:** One file per resource type. Keeps each resource self-contained; avoids a 3000-line `provider.go`. Tests co-located as `*_test.go`.
- **`internal/testmock/`:** Centralized mock HTTP server for integration tests. Shared across resource tests; handlers return realistic FlashBlade JSON payloads.
- **`examples/`:** Required input for `tfplugindocs`. Every resource and data source needs a matching `.tf` file or docs generation fails.
- **`tools/`:** `go:generate` isolation; keeps dev tooling from polluting the main module's compiled deps.

## Architectural Patterns

### Pattern 1: Provider-Configured Client Injection

**What:** The `FlashBladeProvider.Configure` method builds the single `FlashBladeClient` (including auth, TLS, retries) and stores it in `resp.ResourceData` and `resp.DataSourceData`. Each resource/data source receives it via its own `Configure(req.ProviderData)` call.

**When to use:** Always — this is the only sanctioned pattern in terraform-plugin-framework. Never construct clients inside resource CRUD methods.

**Trade-offs:** Client is instantiated once per provider lifetime; no per-resource client customization (acceptable for single-array target).

**Example:**
```go
// internal/provider/provider.go
func (p *FlashBladeProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
    var config flashBladeProviderModel
    resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
    if resp.Diagnostics.HasError() {
        return
    }
    client, err := client.NewClient(client.Config{
        Endpoint: config.Endpoint.ValueString(),
        APIToken: config.APIToken.ValueString(),
        // ... TLS, OAuth2 fields
    })
    if err != nil {
        resp.Diagnostics.AddError("Client init failed", err.Error())
        return
    }
    resp.ResourceData = client
    resp.DataSourceData = client
}

// internal/provider/filesystem_resource.go
func (r *filesystemResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
    if req.ProviderData == nil {
        return
    }
    c, ok := req.ProviderData.(*client.FlashBladeClient)
    if !ok {
        resp.Diagnostics.AddError("Unexpected provider data type", fmt.Sprintf("got %T", req.ProviderData))
        return
    }
    r.client = c
}
```

### Pattern 2: Read-at-End-of-Write (Canonical State Sync)

**What:** After every Create and Update, call the resource's own `Read` method (or the underlying client Read) to refresh state from the API. This ensures Terraform state reflects the server's authoritative view, not the request payload.

**When to use:** All resources. FlashBlade API sets computed fields (IDs, timestamps, derived attributes) on creation that are not in the POST body.

**Trade-offs:** One extra GET per write operation — negligible for storage provisioning frequencies.

**Example:**
```go
func (r *filesystemResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var plan filesystemModel
    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    // ... POST to FlashBlade ...
    // Re-read to populate computed fields (id, created, space.*)
    r.readIntoState(ctx, plan.Name.ValueString(), &resp.State, &resp.Diagnostics)
}
```

### Pattern 3: Separate Client Library from Provider Logic

**What:** `internal/client/` has zero terraform-plugin-framework imports. It is a pure Go HTTP client for the FlashBlade API. Provider resources import the client; the client does not know about Terraform types.

**When to use:** Mandatory for testability. Unit-testable without Terraform process; mockable via interface.

**Trade-offs:** Extra abstraction layer. Worth it: client can be tested with standard `net/http/httptest`; resources can be tested with mock client interface.

### Pattern 4: Policy-as-Parent + Rule-as-Child Resources

**What:** Each FlashBlade policy (NFS export, SMB share, snapshot, quota, network access, object store access) has two Terraform resources: the policy container (`flashblade_nfs_export_policy`) and the rule (`flashblade_nfs_export_policy_rule`). Rules reference their policy by name.

**When to use:** FlashBlade API models rules as sub-collections under a policy. Mirroring this in Terraform avoids overly complex nested blocks while preserving independent lifecycle management of rules.

**Trade-offs:** More resources to document; practitioners must manage ordering. Alternative (nested blocks in policy resource) is a valid simplification if rule set is small and static — but loses independent import/drift-detection per rule.

## Data Flow

### Terraform Plan/Apply Flow

```
Practitioner: terraform plan
        |
Terraform Core: reads .tf config
        |
        v
Provider gRPC: ConfigureProvider RPC
        |
FlashBladeProvider.Configure()
  - reads endpoint, api_token / oauth2 from config + env vars
  - calls client.NewClient() → builds http.Client with auth transport
  - stores *FlashBladeClient in ResourceData / DataSourceData
        |
        v
For each resource in plan:
  resource.Configure() → receives *FlashBladeClient, stores as r.client
        |
        v
Terraform Core: ReadResource RPC (refresh)
  resource.Read()
    → r.client.GetFileSystem(ctx, name)
      → HTTP GET /api/2.22/file-systems?names=<name>
      ← JSON response → Go struct
    → map struct fields → tfsdk state types
    → resp.State.Set(ctx, model)
        |
        v
Terraform Core: PlanResourceChange RPC
  diffs planned config against refreshed state
  → generates plan with attribute-level diff
        |
        v
Practitioner: terraform apply
        |
Terraform Core: ApplyResourceChange RPC
  resource.Create() / Update() / Delete()
    → r.client.PostFileSystem / PatchFileSystem / DeleteFileSystem
      → HTTP POST/PATCH/DELETE /api/2.22/file-systems
    → re-read via Read() to populate computed fields
    → resp.State.Set(ctx, model)
```

### Auth Flow

```
client.NewClient(config)
        |
        v
if config.APIToken set:
  POST /api/login (header: api-token: <token>)
  ← x-auth-token header → stored in http.Transport
  all subsequent requests: header x-auth-token: <session_token>

if config.OAuth2 (client_id + private_key):
  POST /oauth2/1.0/token (grant_type=token-exchange, subject_token=<api_token>)
  ← {access_token} → stored in http.Transport
  all subsequent requests: Authorization: Bearer <access_token>

Token refresh: re-authenticate on 401 response (retry once)
```

### Drift Detection Flow

```
terraform refresh / plan -refresh-only
        |
resource.Read()
  → r.client.GetFileSystem(ctx, name)
  → if 404: resp.State.RemoveResource(ctx) ← marks resource as deleted externally
  → map all API fields → model
  → resp.State.Set(ctx, model)
        |
Terraform Core: compares new state to last known state
  → any difference = drift, shown in plan output
```

## Scaling Considerations

This provider targets a single FlashBlade array per provider instance. Scaling is not about users — it is about breadth of resources and API call volume.

| Concern | Approach |
|---------|----------|
| Many concurrent resources (large state) | No special handling needed; Terraform parallelism is controlled by `-parallelism` flag, not provider |
| API rate limiting from FlashBlade | Implement exponential backoff in `transport.go`; FlashBlade REST API does not publish rate limits but will return 429 or 503 under load |
| Large policy rule sets (100+ rules) | Use pagination: FlashBlade list endpoints support `limit`/`continuation_token`; client must paginate until exhausted |
| Long-running operations | FlashBlade REST is synchronous for most operations; no async job polling needed at v2.22 |
| Multiple arrays | Use provider aliasing (`provider "flashblade" { alias = "array2" }`); no code changes needed |

## Anti-Patterns

### Anti-Pattern 1: Terraform Types in Client Layer

**What people do:** Use `types.String`, `types.Bool` etc. from terraform-plugin-framework in `internal/client/` structs.

**Why it's wrong:** Creates a hard dependency on the Terraform framework in a layer that should be a pure HTTP client. Makes unit testing of the client require importing the entire framework. Breaks separation of concerns.

**Do this instead:** Use plain Go types (`string`, `bool`, `int64`, `*string`) in client models. Map to Terraform types in the provider resource layer.

### Anti-Pattern 2: Skipping Read at End of Create/Update

**What people do:** After a successful POST/PATCH, directly set `resp.State` from the plan values, skipping a GET to refresh.

**Why it's wrong:** FlashBlade API populates computed fields (ID, creation timestamps, effective space values, derived permissions) that are not in the request body. Skipping the Read causes permanent plan-time diffs on computed attributes.

**Do this instead:** Always call `Read` (or an internal `readIntoState` helper) at the end of Create and Update.

### Anti-Pattern 3: One Giant Provider File

**What people do:** Put all resources as `Resource1{}, Resource2{}...` with their full implementations in `provider.go`.

**Why it's wrong:** Becomes unmaintainable. 20+ resources means thousands of lines in a single file.

**Do this instead:** One file per resource: `filesystem_resource.go`, `bucket_resource.go`. Provider file only contains `Resources()` and `DataSources()` factory slices.

### Anti-Pattern 4: Hardcoding API Version in Every Request

**What people do:** Hardcode `/api/2.22/` in individual client method calls.

**Why it's wrong:** Future version bumps require changing every file. FlashBlade API also exposes version negotiation via `GET /api/api_version`.

**Do this instead:** Store `baseURL = "https://{endpoint}/api/2.22"` once in `FlashBladeClient`. All methods call `c.get(ctx, "/file-systems", ...)` using the client's base URL.

### Anti-Pattern 5: Ignoring Policy/Rule Ordering

**What people do:** Store NFS export policy rules as a plain list, ignoring the rule index/order returned by the API.

**Why it's wrong:** FlashBlade applies rules in order; reordering causes persistent drift on `terraform plan` as the API returns rules in insertion order.

**Do this instead:** Use `schema.ListNestedAttribute` (ordered) for rules, not `schema.SetNestedAttribute`. Store and compare the rule index field.

## Integration Points

### External Services

| Service | Integration Pattern | Notes |
|---------|---------------------|-------|
| FlashBlade REST API v2.22 | HTTPS REST, JSON, `x-auth-token` or `Bearer` auth | TLS: must support custom CA cert via `tls_ca_cert` config |
| FlashBlade OAuth2 endpoint | `POST /oauth2/1.0/token` token exchange | Use for prod; API token for dev |
| CI/CD (no real array) | `httptest.Server` mock in `internal/testmock/` | `TF_ACC` env var guards acceptance tests |
| Real FlashBlade (acceptance) | Live array, `TF_ACC=1`, `FLASHBLADE_ENDPOINT` + `FLASHBLADE_API_TOKEN` env vars | Must clean up resources in test teardown |

### Internal Boundaries

| Boundary | Communication | Notes |
|----------|---------------|-------|
| `provider` -> `client` | Direct Go method calls on `*FlashBladeClient` | Provider owns no HTTP logic |
| `client` -> FlashBlade API | `http.Client` with custom `RoundTripper` | Auth, retries, `X-Request-ID` header all in transport layer |
| `provider` -> `testmock` | Only in `_test.go` files; mock server URL injected via provider config | Test resources point to `httptest.Server.URL` |
| `resource` -> `resource` | No direct coupling; Terraform handles dependency graph via `depends_on` or attribute references | e.g., `flashblade_bucket` depends on `flashblade_object_store_account` |

## Build Order (Phase Dependency Map)

Components must be built in this order because each layer depends on the previous:

```
1. internal/client/ (core)
   ├── client.go + auth.go + transport.go + errors.go + models.go
   └── Required before any resource can be implemented

2. internal/provider/provider.go (skeleton)
   ├── Schema (endpoint, auth fields)
   ├── Configure (wires client injection)
   └── Empty Resources() / DataSources() stubs

3. First resource: flashblade_filesystem + flashblade_bucket
   ├── Validates end-to-end flow: client -> provider -> resource -> API
   ├── Establishes CRUD pattern all other resources will follow
   └── Acceptance test confirms real FlashBlade connectivity

4. Object store resources (account, access_key, bucket)
   └── Depends on filesystem pattern being proven

5. Policy resources (NFS, SMB, snapshot, quota, network, object store access)
   ├── More complex due to policy + rule parent/child pattern
   └── All follow identical structural pattern once first policy is built

6. Array admin resources (DNS, NTP, SMTP, alerts)
   ├── Often singleton-style (PATCH only, no CREATE/DELETE for some)
   └── Build last; lower operational priority than storage resources

7. Documentation + examples
   └── Run tfplugindocs after all resources have schema + examples/
```

## Sources

- [Terraform Plugin Framework — Providers](https://developer.hashicorp.com/terraform/plugin/framework/providers)
- [Terraform Plugin Framework — Resources Configure](https://developer.hashicorp.com/terraform/plugin/framework/resources/configure)
- [Tutorial: Configure Provider Client](https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-provider-configure)
- [hashicorp/terraform-provider-scaffolding-framework](https://github.com/hashicorp/terraform-provider-scaffolding-framework) — official scaffolding template
- [hashicorp/terraform-provider-hcp](https://github.com/hashicorp/terraform-provider-hcp) — production provider reference (internal/clients, internal/provider split)
- [HashiCorp Provider Best Practices — Design Principles](https://developer.hashicorp.com/terraform/plugin/best-practices/hashicorp-provider-design-principles)
- [Plugin Development — Testing](https://developer.hashicorp.com/terraform/plugin/testing)
- [SDKv2 Detecting Drift](https://developer.hashicorp.com/terraform/plugin/sdkv2/best-practices/detecting-drift) — pattern carries forward to framework

---
*Architecture research for: Terraform provider — Pure Storage FlashBlade REST API*
*Researched: 2026-03-26*
