# Phase 1: Foundation - Research

**Researched:** 2026-03-26
**Domain:** Terraform Provider (Go) — terraform-plugin-framework, FlashBlade REST API v2.22, greenfield project
**Confidence:** HIGH

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

#### HTTP Client Design
- Configurable retries: user sets `max_retries` and `base_delay` in provider config block (with sensible defaults)
- Per-resource Terraform timeouts via terraform-plugin-framework timeouts block (create/read/update/delete individually)
- No global HTTP client timeout — resource-level timeouts govern operation duration
- Retry on transient errors (429, 503, 5xx) with exponential backoff

#### Provider Config Schema
- Nested blocks for auth separation:
  ```hcl
  provider "flashblade" {
    endpoint = "https://flashblade.example.com"
    auth {
      api_token = "..."
    }
    # OR
    auth {
      oauth2 {
        client_id  = "..."
        key_id     = "..."
        issuer     = "..."
      }
    }
  }
  ```
- Environment variable fallbacks with `FLASHBLADE_` prefix: `FLASHBLADE_HOST`, `FLASHBLADE_API_TOKEN`, `FLASHBLADE_OAUTH2_CLIENT_ID`, etc.
- TLS configuration: `ca_cert_file` (path), `ca_cert` (inline PEM string), and `insecure_skip_verify` (boolean, for dev/testing, with warning in docs)
- Retry configuration: `max_retries` and `retry_base_delay` in provider config

#### File System Resource Scope
- Full coverage of all API attributes — expose everything the API provides
- Full NFS/SMB protocol blocks inline on the resource
- Space attributes (total, used, virtual, unique, snapshots) exposed as computed
- Snapshot directory hidden by default (follow Pure Storage defaults, don't expose as configurable)
- ID as primary identifier internally — API ID used for all CRUD calls, stable across renames
- Name is user-facing key — used for import, display, and user references
- Explicit defaults in schema matching API defaults — user sees them in plan
- In-place rename supported via PATCH (name change does not force recreation)
- Multi-protocol allowed freely (NFS + SMB simultaneously) — no provider-level restriction
- Policy references accept plain strings

#### Soft-Delete Behavior
- Configurable `destroy_eradicate_on_delete` boolean attribute on the resource (default: `true`)
  - `true` = PATCH destroyed=true then DELETE (full removal, name reusable)
  - `false` = PATCH destroyed=true only (recoverable within eradication window)
- If file system is soft-deleted outside Terraform: keep in state with `destroyed = true` attribute visible
- Synchronous eradication: when `destroy_eradicate_on_delete = true`, poll until fully eradicated before returning
- This pattern becomes the template for buckets and object store accounts in Phase 2

### Claude's Discretion
- HTTP log levels and verbosity (request/response at DEBUG vs TRACE)
- Exact retry defaults (max_retries count, base_delay duration)
- Go project scaffolding details (Makefile targets, CI config, linting setup)
- Internal error classification logic (retryable vs terminal)

### Deferred Ideas (OUT OF SCOPE)
None — discussion stayed within phase scope.
</user_constraints>

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| PROV-01 | Provider accepts endpoint URL, API token, and TLS CA certificate via config block | Provider schema pattern + TLS RoundTripper chain documented below |
| PROV-02 | Provider accepts OAuth2 client_id, key_id, and issuer for client_credentials auth | OAuth2 token exchange endpoint confirmed in FLASHBLADE_API.md; non-standard grant type documented |
| PROV-03 | Provider falls back to FLASHBLADE_ENDPOINT, FLASHBLADE_API_TOKEN, FLASHBLADE_OAUTH2_* env vars | terraform-plugin-framework `os.Getenv` pattern in Configure documented |
| PROV-04 | Provider negotiates API version on startup via GET /api/api_version and targets v2.22 | Endpoint confirmed in FLASHBLADE_API.md; version negotiation pattern documented |
| PROV-05 | Provider marks api_token, oauth2 private key, and access key secrets as Sensitive in schema | `schema.StringAttribute{Sensitive: true}` pattern documented; pitfalls section covers this |
| PROV-06 | Provider logs all operations with structured tflog output | terraform-plugin-log v0.10.0 confirmed; tflog usage documented |
| PROV-07 | Provider supports custom CA certificate for TLS verification | `crypto/tls` + `crypto/x509` stdlib pattern documented with code example |
| FS-01 | User can create a file system with name, provisioned size, and optional policy attachments | POST /api/2.22/file-systems body documented; FileSystemPost schema extracted |
| FS-02 | User can update file system attributes (size, policies, NFS settings, SMB settings) | PATCH /api/2.22/file-systems body documented; in-place rename confirmed via `name` field in FileSystemPatch |
| FS-03 | User can destroy a file system (two-phase: mark destroyed, then eradicate) | `destroyed` field on FileSystemPatch confirmed; DELETE endpoint confirmed; `time_remaining` polling field identified |
| FS-04 | User can read file system state including all computed attributes (space, created timestamp) | FileSystem schema extracted: `created`(ro), `space`(object ro), `id`(ro), `time_remaining`(ro), `promotion_status`(ro) |
| FS-05 | User can import an existing file system into Terraform state by name | Import-by-name pattern documented; `ImportState` + `Read` flow documented |
| FS-06 | Data source returns file system attributes by name or filter | GET /api/2.22/file-systems with `?names=` param confirmed; data source pattern documented |
| FS-07 | Drift detection logs field-level diffs via tflog when Read finds state divergence | tflog structured logging pattern documented; drift detection flow documented |
</phase_requirements>

---

## Summary

Phase 1 builds the entire foundation from scratch: Go module, project scaffold, shared HTTP client (auth + TLS + retry), provider schema, and the first full resource (`flashblade_file_system`). Every pattern established here — client injection, Read-at-end-of-write, soft-delete handling, schema conventions, drift logging — will be copy-paste templates for 12+ resources in phases 2–4.

The FlashBlade API is clean REST JSON with two auth modes (session token and OAuth2 token exchange) and a documented soft-delete lifecycle. The two confirmed blockers from STATE.md — OAuth2 non-standard grant type and eradication polling mechanism — are now clarified: the `urn:ietf:params:oauth:grant-type:token-exchange` grant type requires a custom `oauth2.TokenSource` (not `clientcredentials.Config`), and eradication polling uses `GET /api/2.22/file-systems?names=<name>&destroyed=true` until 404 (resource fully eradicated) while `time_remaining` is a read-only field on the `FileSystem` object (milliseconds remaining before auto-eradication).

**Primary recommendation:** Build the client layer first (`internal/client/`), then the provider scaffold, then the file system resource. The client layer must have zero terraform-plugin-framework imports — it is a pure Go HTTP library. Every test-critical function in the client layer must be unit-testable with `httptest.NewServer`.

---

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go | 1.25.x | Implementation language | Required minimum for all terraform-plugin-* modules as of March 2026; official scaffolding pins 1.25.5 |
| terraform-plugin-framework | v1.19.0 | Provider scaffold, schema, CRUD, plan modifiers | HashiCorp's current recommended SDK; protocol v6; SDKv2 is maintenance-only |
| terraform-plugin-log | v0.10.0 | Structured tflog logging inside provider | Required for Terraform log routing; stdlib `log` bypasses Terraform's log capture |
| terraform-plugin-go | v0.31.0 | Low-level protocol bindings | Indirect dep; pin explicitly to prevent transitive version surprises |
| terraform-plugin-testing | v1.15.0 | Acceptance test runner | Official harness for running real Terraform CLI against provider |
| golang.org/x/oauth2 | v0.27+ | OAuth2 token handling | Handles token refresh and concurrency correctly; use for session-scoped token source |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| net/http (stdlib) | — | HTTP transport base | All API calls — build RoundTripper chain over this |
| crypto/tls (stdlib) | — | Custom CA cert TLS config | For `ca_cert_file` / `ca_cert` provider attributes |
| crypto/x509 (stdlib) | — | CA cert pool construction | `x509.SystemCertPool()` + `AppendCertsFromPEM()` |
| net/http/httptest (stdlib) | — | Mock HTTP server for integration tests | CI tests without a real FlashBlade |
| github.com/hashicorp/terraform-plugin-docs | v0.21+ | Registry documentation generation | Run via `go:generate` — required for every resource |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `golang.org/x/oauth2` with custom TokenSource | `golang.org/x/oauth2/clientcredentials.Config` | `clientcredentials` uses the standard `client_credentials` grant; FlashBlade uses the non-standard `token-exchange` grant — requires custom TokenSource |
| Custom retry in transport.go | `github.com/hashicorp/go-retryablehttp` | go-retryablehttp is acceptable but adds a dependency; for a single-provider library, a small custom retry loop in `transport.go` is simpler and fully controllable |
| `httptest.NewServer` (stdlib) | `github.com/jarcoal/httpmock` | httpmock is reasonable but adds a dependency; stdlib httptest is zero-dep and idiomatic Go |

**Installation:**
```bash
go mod init github.com/soulkyu/terraform-provider-flashblade
go get github.com/hashicorp/terraform-plugin-framework@v1.19.0
go get github.com/hashicorp/terraform-plugin-testing@v1.15.0
go get github.com/hashicorp/terraform-plugin-log@v0.10.0
go get github.com/hashicorp/terraform-plugin-go@v0.31.0
go get golang.org/x/oauth2@latest
go get github.com/hashicorp/terraform-plugin-docs@latest
```

---

## Architecture Patterns

### Recommended Project Structure
```
terraform-provider-flashblade/
├── main.go
├── go.mod
├── go.sum
├── GNUmakefile
├── .goreleaser.yml
├── .golangci.yml
├── terraform-registry-manifest.json
│
├── internal/
│   ├── provider/
│   │   ├── provider.go                     # FlashBladeProvider schema + Configure
│   │   ├── provider_test.go
│   │   ├── filesystem_resource.go          # flashblade_file_system CRUD + Import
│   │   ├── filesystem_resource_test.go
│   │   ├── filesystem_data_source.go       # flashblade_file_system data source
│   │   └── filesystem_data_source_test.go
│   │
│   ├── client/
│   │   ├── client.go        # FlashBladeClient struct + NewClient
│   │   ├── auth.go          # API token session login + OAuth2 custom TokenSource
│   │   ├── transport.go     # RoundTripper: retries, X-Request-ID, header logging
│   │   ├── errors.go        # API error type, HTTP status classification
│   │   ├── models.go        # Go structs for FlashBlade API JSON (plain Go types only)
│   │   └── filesystems.go   # GetFileSystem, PostFileSystem, PatchFileSystem, DeleteFileSystem
│   │
│   └── testmock/
│       ├── server.go        # httptest server factory
│       └── handlers/
│           └── filesystems.go
│
├── examples/
│   ├── provider/provider.tf
│   └── resources/flashblade_file_system/resource.tf
│
├── docs/                    # Generated by tfplugindocs — never edit manually
└── tools/
    └── tools.go             # blank imports: tfplugindocs
```

### Pattern 1: Provider-Configured Client Injection
**What:** `FlashBladeProvider.Configure` builds the single `*FlashBladeClient` and stores it in `resp.ResourceData`. Each resource receives it via its own `Configure` call.
**When to use:** Always — the only sanctioned pattern in terraform-plugin-framework.

```go
// Source: https://developer.hashicorp.com/terraform/plugin/framework/providers
// internal/provider/provider.go
func (p *FlashBladeProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
    var config flashBladeProviderModel
    resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
    if resp.Diagnostics.HasError() {
        return
    }
    // Resolve env var fallbacks
    endpoint := config.Endpoint.ValueString()
    if endpoint == "" {
        endpoint = os.Getenv("FLASHBLADE_HOST")
    }
    c, err := client.NewClient(client.Config{
        Endpoint:         endpoint,
        APIToken:         config.Auth.APIToken.ValueString(),
        MaxRetries:       int(config.MaxRetries.ValueInt64()),
        RetryBaseDelay:   config.RetryBaseDelay.ValueString(),
        CACertFile:       config.CACertFile.ValueString(),
        CACert:           config.CACert.ValueString(),
        InsecureSkipVerify: config.InsecureSkipVerify.ValueBool(),
    })
    if err != nil {
        resp.Diagnostics.AddError("FlashBlade client initialization failed", err.Error())
        return
    }
    resp.ResourceData = c
    resp.DataSourceData = c
}

// internal/provider/filesystem_resource.go
func (r *filesystemResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
    if req.ProviderData == nil {
        return
    }
    c, ok := req.ProviderData.(*client.FlashBladeClient)
    if !ok {
        resp.Diagnostics.AddError("Unexpected provider data type",
            fmt.Sprintf("expected *client.FlashBladeClient, got %T", req.ProviderData))
        return
    }
    r.client = c
}
```

### Pattern 2: Read-at-End-of-Write (Canonical State Sync)
**What:** After every Create and Update, call the resource's `Read` logic to populate state from the API response — never copy the plan directly.
**When to use:** Every Create and Update, without exception.

```go
// internal/provider/filesystem_resource.go
func (r *filesystemResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var plan filesystemModel
    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    if resp.Diagnostics.HasError() {
        return
    }
    _, err := r.client.PostFileSystem(ctx, client.FileSystemPost{
        Name:        plan.Name.ValueString(),
        Provisioned: plan.Provisioned.ValueInt64(),
        // ... other fields
    })
    if err != nil {
        resp.Diagnostics.AddError("Create file system failed", err.Error())
        return
    }
    // Always re-read to populate computed fields: id, created, space.*
    r.readIntoState(ctx, plan.Name.ValueString(), &resp.State, &resp.Diagnostics)
}
```

### Pattern 3: Soft-Delete Two-Phase Destroy
**What:** `Delete` sends `PATCH destroyed=true`, then (if `destroy_eradicate_on_delete=true`) sends `DELETE`, then polls until `GET ?names=<name>&destroyed=true` returns 404.
**When to use:** For every resource that uses FlashBlade's soft-delete model (file systems, buckets — Phase 1 establishes the template).

```go
// internal/provider/filesystem_resource.go
func (r *filesystemResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    var state filesystemModel
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

    // Phase 1: soft-delete
    err := r.client.PatchFileSystem(ctx, state.ID.ValueString(), client.FileSystemPatch{
        Destroyed: boolPtr(true),
    })
    if err != nil {
        resp.Diagnostics.AddError("Soft-delete failed", err.Error())
        return
    }

    if !state.DestroyEradicateOnDelete.ValueBool() {
        return // leave soft-deleted, name still reserved
    }

    // Phase 2: eradicate
    err = r.client.DeleteFileSystem(ctx, state.ID.ValueString())
    if err != nil {
        resp.Diagnostics.AddError("Eradication DELETE failed", err.Error())
        return
    }

    // Phase 3: poll until name is free (404 on destroyed list)
    if err := r.client.PollUntilEradicated(ctx, state.Name.ValueString()); err != nil {
        resp.Diagnostics.AddError("Eradication polling timed out", err.Error())
    }
}
```

### Pattern 4: OAuth2 Non-Standard Token Exchange
**What:** FlashBlade's OAuth2 endpoint uses `grant_type=urn:ietf:params:oauth:grant-type:token-exchange` — not the standard `client_credentials` grant. This requires a custom `oauth2.TokenSource`.
**When to use:** When provider is configured with OAuth2 (`auth.oauth2` block).

```go
// Source: FLASHBLADE_API.md Auth section
// internal/client/auth.go

// FlashBladeTokenSource implements oauth2.TokenSource for the non-standard
// token-exchange grant type used by FlashBlade.
type FlashBladeTokenSource struct {
    endpoint    string
    apiToken    string
    httpClient  *http.Client
    mu          sync.Mutex
    cachedToken *oauth2.Token
}

func (ts *FlashBladeTokenSource) Token() (*oauth2.Token, error) {
    ts.mu.Lock()
    defer ts.mu.Unlock()
    if ts.cachedToken != nil && ts.cachedToken.Valid() {
        return ts.cachedToken, nil
    }
    // POST /oauth2/1.0/token
    // Content-Type: application/x-www-form-urlencoded
    // grant_type=urn:ietf:params:oauth:grant-type:token-exchange
    // &subject_token=<API_TOKEN>
    // &subject_token_type=urn:ietf:params:oauth:token-type:jwt
    form := url.Values{
        "grant_type":         {"urn:ietf:params:oauth:grant-type:token-exchange"},
        "subject_token":      {ts.apiToken},
        "subject_token_type": {"urn:ietf:params:oauth:token-type:jwt"},
    }
    resp, err := ts.httpClient.PostForm(ts.endpoint+"/oauth2/1.0/token", form)
    // ... parse {access_token, expires_in} → *oauth2.Token
}
```

**Note (from STATE.md blocker):** The exact `subject_token` semantics need confirmation against a live array — the API docs show `subject_token=<API_TOKEN>` but the flow may require a JWT assertion, not the raw API token. Implement with this shape but validate during Phase 1 execution.

### Pattern 5: API Version Negotiation at Startup
**What:** On `Configure`, call `GET /api/api_version` and verify `"2.22"` appears in the `versions` array.
**When to use:** Provider Configure, before any resource operations.

```go
// internal/client/client.go
func (c *FlashBladeClient) NegotiateVersion(ctx context.Context) error {
    var result struct {
        Versions []string `json:"versions"`
    }
    if err := c.getUnversioned(ctx, "/api/api_version", &result); err != nil {
        return fmt.Errorf("version negotiation failed: %w", err)
    }
    for _, v := range result.Versions {
        if v == APIVersion {
            return nil
        }
    }
    return fmt.Errorf("FlashBlade does not support API version %s; available: %v", APIVersion, result.Versions)
}
```

### Pattern 6: Drift Detection with Structured tflog
**What:** In `Read`, after fetching API state, compare each field against current state and emit structured `tflog.Info` for any divergence.
**When to use:** In every resource `Read` function that detects drift.

```go
// internal/provider/filesystem_resource.go
func (r *filesystemResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    var state filesystemModel
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

    fs, err := r.client.GetFileSystem(ctx, state.Name.ValueString())
    if err != nil {
        if isNotFound(err) {
            resp.State.RemoveResource(ctx) // deleted outside Terraform
            return
        }
        resp.Diagnostics.AddError("Read file system failed", err.Error())
        return
    }

    // FS-07: log field-level drift
    if fs.Provisioned != state.Provisioned.ValueInt64() {
        tflog.Info(ctx, "drift detected on file_system", map[string]any{
            "resource":         state.Name.ValueString(),
            "field":            "provisioned",
            "state_value":      state.Provisioned.ValueInt64(),
            "api_value":        fs.Provisioned,
        })
    }
    // ... map all fields from fs → state model
    resp.Diagnostics.Append(resp.State.Set(ctx, mapFSToModel(fs))...)
}
```

### Pattern 7: TLS Transport Chain
**What:** Build the `http.Transport` with custom CA cert pool, then layer auth on top via a `RoundTripper` wrapper.
**When to use:** Always when `ca_cert_file`, `ca_cert`, or `insecure_skip_verify` is set.

```go
// Source: https://pkg.go.dev/crypto/tls
// internal/client/client.go
func buildTransport(cfg Config) (http.RoundTripper, error) {
    pool, err := x509.SystemCertPool()
    if err != nil {
        pool = x509.NewCertPool()
    }
    if cfg.CACertFile != "" {
        pem, err := os.ReadFile(cfg.CACertFile)
        if err != nil {
            return nil, fmt.Errorf("reading ca_cert_file: %w", err)
        }
        pool.AppendCertsFromPEM(pem)
    }
    if cfg.CACert != "" {
        pool.AppendCertsFromPEM([]byte(cfg.CACert))
    }
    base := &http.Transport{
        TLSClientConfig: &tls.Config{
            RootCAs:            pool,
            InsecureSkipVerify: cfg.InsecureSkipVerify,
        },
    }
    return &retryTransport{
        base:       base,
        maxRetries: cfg.MaxRetries,
        baseDelay:  cfg.RetryBaseDelay,
    }, nil
}
```

### Anti-Patterns to Avoid

- **Copying plan to state:** Never `resp.State.Set(ctx, plan)` — always call `Read` or `readIntoState` at end of Create/Update.
- **Terraform types in client layer:** Never use `types.String`, `types.Bool` etc. in `internal/client/` — use plain Go types only.
- **Inline API version strings:** Never hardcode `"/api/2.22/"` in individual method calls — use `const APIVersion = "2.22"` and a `baseURL` on the client.
- **Skipping import tests:** Write the acceptance test for `import → plan → 0 diff` at the same time as the resource, not later.
- **Sending `ro` fields in PATCH body:** The FlashBlade API may return 422 if read-only fields (marked `(ro ...)` in FLASHBLADE_API.md) are included in PATCH requests — only send writable fields.
- **`log` stdlib inside provider:** Use `tflog` exclusively — stdlib `log` bypasses Terraform's log routing.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| OAuth2 token lifecycle (refresh, expiry, concurrency) | Custom token struct with manual refresh | Custom `oauth2.TokenSource` wrapping `golang.org/x/oauth2` primitives | x/oauth2 handles concurrent Token() calls, expiry windows, and retry semantics correctly |
| TLS cert pool management | Manual PEM parsing | `x509.SystemCertPool()` + `pool.AppendCertsFromPEM()` | Handles platform cert store differences (Linux/macOS/Windows); AppendCertsFromPEM is battle-tested |
| Acceptance test runner | Custom Terraform subprocess runner | `terraform-plugin-testing` `resource.Test` | Handles provider registration, state cleanup, import testing lifecycle |
| Provider documentation | Manual Markdown in `docs/` | `tfplugindocs generate` | Generates Terraform Registry-compatible Markdown from schema; keeping hand-written docs in sync is error-prone |
| Structured log output | `fmt.Printf` or `log.Printf` | `tflog.Info`, `tflog.Debug`, `tflog.Warn` | tflog routes to Terraform's log capture; `fmt` output appears as raw stderr noise |

---

## Common Pitfalls

### Pitfall 1: FlashBlade Soft-Delete Is Not Terraform Destroy
**What goes wrong:** Calling `DELETE` once and returning leaves a tombstoned object on the array. A subsequent `terraform apply` with the same name will fail with a name-collision 409 until the eradication timer expires (default: 24h).
**Why it happens:** Developers model destroy as a single API call; FlashBlade uses a data-safety two-phase delete.
**How to avoid:** Always PATCH `destroyed=true` first, then DELETE, then poll `GET /api/2.22/file-systems?names=<name>&destroyed=true` until 404.
**Warning signs:** Acceptance test passes on first run, then fails on second run because the array still holds the tombstoned name.

### Pitfall 2: Computed Attribute Misuse → "Provider produced inconsistent result after apply"
**What goes wrong:** Attributes returned computed by the API (marked `(ro ...)` in FLASHBLADE_API.md) that are not declared `Computed: true` in the schema cause framework consistency errors on every apply.
**Why it happens:** Developers forget that `id`, `created`, `space`, `time_remaining`, `promotion_status`, `context`, `realms` are all read-only server-assigned fields.
**How to avoid:** All `(ro ...)` fields in FileSystem model → `Computed: true` in schema. Add `UseStateForUnknown()` plan modifier to stable fields (`id`, `created`).
**Warning signs:** Acceptance test fails with "inconsistent result" on update; second plan shows perpetual diff on computed fields.

### Pitfall 3: Sensitive Values Leaked via Error Messages or State
**What goes wrong:** `api_token` stored in state as plaintext; auth headers included in error message strings; CI logs show `x-auth-token: ...`.
**How to avoid:** Mark `api_token` and OAuth2 credential fields `Sensitive: true` in schema. In `transport.go`, strip `x-auth-token` and `Authorization` headers before constructing error messages.

### Pitfall 4: Missing Read-After-Write → Stale State
**What goes wrong:** State file shows wrong values for `id`, `created`, `space.*` immediately after create.
**How to avoid:** Every `Create` and `Update` must end with a full `Read` call that overwrites state from the API response.

### Pitfall 5: OAuth2 Token Exchange — Non-Standard Grant Type
**What goes wrong:** Using `clientcredentials.Config` with `GrantType` override — the standard package may not support the `token-exchange` form correctly.
**How to avoid:** Implement a custom `oauth2.TokenSource`. The token endpoint is `POST /oauth2/1.0/token` with form body `grant_type=urn:ietf:params:oauth:grant-type:token-exchange&subject_token=<TOKEN>&subject_token_type=urn:ietf:params:oauth:token-type:jwt`.
**Open question:** Whether `subject_token` expects the raw API token string or a JWT assertion — must be validated against a live array.

### Pitfall 6: Sending Read-Only Fields in PATCH Body
**What goes wrong:** Including `(ro ...)` fields like `id`, `created`, `context`, `realms`, `promotion_status` in a PATCH body may cause 422 errors from the FlashBlade API.
**How to avoid:** In `FileSystemPatch` Go struct, only include writable fields from the `FileSystemPatch` schema in FLASHBLADE_API.md. Separate `FileSystem` (GET response) from `FileSystemPatch` (PATCH request) Go types.

### Pitfall 7: No Retry on Transient Errors
**What goes wrong:** `terraform apply` fails permanently on 429/503 errors during array maintenance or firmware updates.
**How to avoid:** Implement exponential backoff in `transport.go` for 429, 503, and 500 responses. Respect `context.Context` deadline from resource-level timeouts.

### Pitfall 8: Incomplete Import → Perpetual Diff After Import
**What goes wrong:** `terraform import` succeeds but subsequent `terraform plan` shows diffs because `Read` leaves optional fields null.
**How to avoid:** `Read` must populate every schema attribute from the API response. Test `import → plan → 0 diff` as part of the Phase 1 acceptance test suite.

---

## Code Examples

### FileSystem API Schema (from FLASHBLADE_API.md)

**FileSystem (GET response) — key fields:**
- `id` (ro string) — stable UUID, use for all CRUD operations
- `name` (ro string on GET, mutable on PATCH) — user-facing name
- `provisioned` (integer) — size in bytes
- `created` (ro integer) — Unix timestamp ms
- `destroyed` (boolean) — soft-delete flag
- `time_remaining` (ro integer) — ms before auto-eradication (null if not destroyed)
- `nfs` (object) — NFS configuration
- `smb` (object) — SMB configuration
- `http` (object) — HTTP configuration
- `multi_protocol` (object) — multi-protocol config
- `space` (object) — usage stats (all computed)
- `snapshot_directory_enabled` (boolean)
- `hard_limit_enabled` (boolean)
- `fast_remove_directory_enabled` (boolean)
- `writable` (boolean)
- `default_group_quota` (integer)
- `default_user_quota` (integer)
- `qos_policy` (object) — reference
- `promotion_status` (ro string) — `promoted` or `demoted`
- `requested_promotion_state` (string)
- `context` (ro object)
- `realms` (ro array)
- `eradication_config` (object) — eradication settings
- `group_ownership` (string)
- `node_group` (object) — reference
- `source` (object) — source snapshot reference
- `storage_class` (object)

**FileSystemPost (POST body) — writable at creation:**
`default_group_quota`, `default_user_quota`, `eradication_config`, `fast_remove_directory_enabled`, `group_ownership`, `hard_limit_enabled`, `http`, `multi_protocol`, `nfs`, `node_group`, `provisioned`, `qos_policy`, `smb`, `snapshot_directory_enabled`, `source`, `writable`

**FileSystemPatch (PATCH body) — writable after creation:**
`created`(ro — do NOT send), `default_group_quota`, `default_user_quota`, `destroyed`, `fast_remove_directory_enabled`, `group_ownership`, `hard_limit_enabled`, `http`, `id`(ro — do NOT send), `multi_protocol`, `name` (mutable — in-place rename), `nfs`, `provisioned`, `qos_policy`, `requested_promotion_state`, `smb`, `snapshot_directory_enabled`, `source`, `storage_class`, `time_remaining`(ro — do NOT send), `writable`

### Eradication Polling Pattern

```go
// internal/client/filesystems.go
func (c *FlashBladeClient) PollUntilEradicated(ctx context.Context, name string) error {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    for {
        select {
        case <-ctx.Done():
            return fmt.Errorf("context cancelled while waiting for eradication of %s", name)
        case <-ticker.C:
            tflog.Debug(ctx, "polling for eradication completion", map[string]any{"name": name})
            _, err := c.GetFileSystem(ctx, name, withDestroyed(true))
            if isNotFound(err) {
                return nil // fully eradicated
            }
            if err != nil {
                return fmt.Errorf("polling eradication status: %w", err)
            }
            // still exists (destroyed=true state), keep polling
        }
    }
}
```

### API Query Pattern (names param, not path param)
```go
// Source: FLASHBLADE_API.md Common Parameters
// GET /api/2.22/file-systems?names=fs-01  (NOT /api/2.22/file-systems/fs-01)
func (c *FlashBladeClient) GetFileSystem(ctx context.Context, name string, opts ...QueryOption) (*FileSystem, error) {
    u := c.baseURL + "/file-systems"
    req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
    q := req.URL.Query()
    q.Set("names", name)
    req.URL.RawQuery = q.Encode()
    // ... execute, parse items[0]
}
```

### Schema Attribute Conventions

```go
// internal/provider/filesystem_resource.go
// Required writable field
"name": schema.StringAttribute{
    Required:    true,
    Description: "Name of the file system. Used as the import ID.",
    PlanModifiers: []planmodifier.String{
        // In-place rename is allowed — do NOT add RequiresReplace
    },
},

// Optional+Computed field (API may adjust value)
"provisioned": schema.Int64Attribute{
    Optional:    true,
    Computed:    true,
    Description: "Provisioned size in bytes.",
},

// Computed-only stable field (set once at creation)
"id": schema.StringAttribute{
    Computed:    true,
    Description: "Unique ID assigned by FlashBlade. Used internally for all CRUD calls.",
    PlanModifiers: []planmodifier.String{
        stringplanmodifier.UseStateForUnknown(),
    },
},

// Sensitive credential field
"api_token": schema.StringAttribute{
    Optional:    true,
    Sensitive:   true,
    Description: "FlashBlade API token. Set via FLASHBLADE_API_TOKEN env var if absent.",
},

// Computed-only read-only field (never set by user)
"created": schema.Int64Attribute{
    Computed:    true,
    Description: "Creation timestamp in milliseconds since Unix epoch.",
    PlanModifiers: []planmodifier.Int64{
        int64planmodifier.UseStateForUnknown(),
    },
},
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| terraform-plugin-sdk/v2 | terraform-plugin-framework v1.19 | Ongoing migration; SDKv2 feature-frozen as of 2023 | Framework is mandatory for new providers; plan modifiers, write-only attrs, identity support |
| Manual Makefile multi-platform build | goreleaser + hashicorp/ghaction-terraform-provider-release | 2022 | Required for Terraform Registry publication |
| golangci-lint v1 `.golangci.yml` | golangci-lint v2 (March 2025) | v2 released March 2025 | v2 config format is incompatible with v1 — write fresh config |

**Deprecated / outdated:**
- `terraform-plugin-sdk/v2`: feature-frozen, use framework instead
- `terraform-plugin-mux`: only needed when combining framework + SDKv2 providers; not needed for pure-framework

---

## Open Questions

1. **OAuth2 `subject_token` exact semantics**
   - What we know: FLASHBLADE_API.md shows `subject_token=<API_TOKEN>` in the token-exchange form post
   - What's unclear: Whether the value is the raw API token string, or whether it must be a signed JWT assertion (as the `subject_token_type=jwt` hint suggests)
   - Recommendation: Implement with raw API token first; add a `// TODO: validate subject_token format against live array` comment; test during Phase 1 execution before merging

2. **Eradication polling endpoint and interval**
   - What we know: `time_remaining` (ro integer ms) field exists on `FileSystem`; `GET /api/2.22/file-systems?destroyed=true&names=<name>` returns the resource if it's still tombstoned
   - What's unclear: Whether `time_remaining` becomes 0 before the GET returns 404, or whether it stays nonzero until the resource disappears entirely
   - Recommendation: Poll `GET ?destroyed=true&names=<name>` until 404; use 5s poll interval with context deadline; `time_remaining` is informational only in logs

3. **NFS/SMB sub-object exact fields**
   - What we know: `nfs` and `smb` are object fields on FileSystem; FLASHBLADE_API.md abbreviates object contents
   - What's unclear: The complete field list inside `nfs` (v3_enabled, v4_1_enabled, rules, etc.) and `smb` (enabled, access_based_enum, etc.) objects
   - Recommendation: Use the CONTEXT.md schema as the source of truth for the HCL shape; verify exact API field names by calling `GET /api/2.22/file-systems` on a test array during Phase 1

---

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | terraform-plugin-testing v1.15.0 (acceptance) + `testing` stdlib (unit) |
| Config file | none — greenfield, Wave 0 must create |
| Quick run command | `go test ./internal/... -run TestUnit -count=1` |
| Full suite command | `go test ./... -count=1` |
| Acceptance test command | `TF_ACC=1 go test ./internal/provider/... -run TestAcc -v -timeout 120m` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| PROV-01 | Provider accepts endpoint + api_token + CA cert config | unit | `go test ./internal/provider/... -run TestUnit_ProviderSchema` | ❌ Wave 0 |
| PROV-02 | Provider accepts OAuth2 auth block | unit | `go test ./internal/provider/... -run TestUnit_ProviderSchema` | ❌ Wave 0 |
| PROV-03 | Provider falls back to env vars | unit | `go test ./internal/client/... -run TestUnit_EnvVarFallback` | ❌ Wave 0 |
| PROV-04 | API version negotiation on startup | unit (mock) | `go test ./internal/client/... -run TestUnit_NegotiateVersion` | ❌ Wave 0 |
| PROV-05 | Sensitive fields not leaked | unit | `go test ./internal/provider/... -run TestUnit_SensitiveAttributes` | ❌ Wave 0 |
| PROV-06 | Structured tflog output | unit (mock) | `go test ./internal/provider/... -run TestUnit_TflogOutput` | ❌ Wave 0 |
| PROV-07 | Custom CA cert TLS transport | unit | `go test ./internal/client/... -run TestUnit_CustomCATLS` | ❌ Wave 0 |
| FS-01 | Create file system CRUD lifecycle | acceptance | `TF_ACC=1 go test ./internal/provider/... -run TestAcc_FileSystem_Create` | ❌ Wave 0 |
| FS-02 | Update file system attributes + in-place rename | acceptance | `TF_ACC=1 go test ./internal/provider/... -run TestAcc_FileSystem_Update` | ❌ Wave 0 |
| FS-03 | Destroy → two-phase soft-delete + eradication | acceptance | `TF_ACC=1 go test ./internal/provider/... -run TestAcc_FileSystem_Destroy` | ❌ Wave 0 |
| FS-04 | Read: all computed attributes populated | unit (mock) | `go test ./internal/provider/... -run TestUnit_FileSystem_Read` | ❌ Wave 0 |
| FS-05 | Import by name → plan 0 diff | acceptance | `TF_ACC=1 go test ./internal/provider/... -run TestAcc_FileSystem_Import` | ❌ Wave 0 |
| FS-06 | Data source returns attributes by name | unit (mock) | `go test ./internal/provider/... -run TestUnit_FileSystemDataSource` | ❌ Wave 0 |
| FS-07 | Drift detection emits structured tflog | unit (mock) | `go test ./internal/provider/... -run TestUnit_FileSystem_DriftLog` | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./internal/... -run TestUnit -count=1`
- **Per wave merge:** `go test ./... -count=1` (includes integration tests with mock server)
- **Phase gate:** Full suite green (unit + mock integration) before `/gsd:verify-work`; acceptance tests require real FlashBlade array + `TF_ACC=1`

### Wave 0 Gaps
- [ ] `go.mod` / `go.sum` — module initialization: `go mod init github.com/soulkyu/terraform-provider-flashblade`
- [ ] `main.go` — provider binary entry point
- [ ] `internal/provider/provider.go` — FlashBladeProvider skeleton
- [ ] `internal/client/client.go` + `auth.go` + `transport.go` + `errors.go` + `models.go` — client layer
- [ ] `internal/client/filesystems.go` — file system CRUD methods
- [ ] `internal/provider/filesystem_resource.go` — flashblade_file_system resource
- [ ] `internal/provider/filesystem_data_source.go` — flashblade_file_system data source
- [ ] `internal/testmock/server.go` + `handlers/filesystems.go` — mock HTTP server
- [ ] `GNUmakefile` — `make build`, `make test`, `make testacc`, `make lint`
- [ ] `.golangci.yml` — linter config (golangci-lint v2 format)
- [ ] `examples/provider/provider.tf` — required for tfplugindocs
- [ ] `examples/resources/flashblade_file_system/resource.tf` — required for tfplugindocs
- [ ] All `*_test.go` files listed in the test map above

---

## Sources

### Primary (HIGH confidence)
- FLASHBLADE_API.md (in repo root) — FlashBlade REST API v2.22: file system endpoints, auth flows, schema definitions, `(ro ...)` field annotations, eradication fields, `continuation_token` pagination, `/api/api_version` endpoint
- `.planning/research/STACK.md` — Go version, terraform-plugin-framework v1.19.0, supporting library versions, all verified against official scaffolding go.mod (March 2026)
- `.planning/research/ARCHITECTURE.md` — project structure, component responsibilities, data flow, build order, anti-patterns (verified against HashiCorp official docs)
- `.planning/research/PITFALLS.md` — pitfall catalogue with phase mapping (HIGH confidence for framework pitfalls from official docs; MEDIUM for FlashBlade-specific API quirks)

### Secondary (MEDIUM confidence)
- `.planning/phases/01-foundation/01-CONTEXT.md` — user locked decisions, verified during discussion phase
- [HashiCorp Plugin Framework — Resources Configure](https://developer.hashicorp.com/terraform/plugin/framework/resources/configure) — client injection pattern
- [Terraform Plugin Framework — Plan Modification](https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification) — UseStateForUnknown, RequiresReplace

### Tertiary (LOW confidence)
- OAuth2 token exchange `subject_token` semantics — inferred from FLASHBLADE_API.md; must be validated against live array

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — all versions verified against official scaffolding go.mod (March 2026) and pkg.go.dev
- Architecture: HIGH — based on ARCHITECTURE.md (verified against HashiCorp official docs + production provider references)
- FlashBlade API endpoints: HIGH — sourced from FLASHBLADE_API.md in repo root
- OAuth2 token exchange detail: MEDIUM — grant type confirmed; `subject_token` exact value LOW (unvalidated against live array)
- Eradication polling: MEDIUM — `time_remaining` field confirmed; exact polling stop condition LOW (inferred from 404 behavior)
- Pitfalls: HIGH for framework; MEDIUM for FlashBlade-specific

**Research date:** 2026-03-26
**Valid until:** 2026-04-25 (stable stack; terraform-plugin-framework releases quarterly)
