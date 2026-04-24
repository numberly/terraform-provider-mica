# Phase 1: Foundation - Context

**Gathered:** 2026-03-26
**Status:** Ready for planning

<domain>
## Phase Boundary

Shared HTTP client (auth, retry, TLS, logging), provider scaffold (config schema, registration), and first resource (`flashblade_file_system`) with full CRUD, import, drift detection, and soft-delete — establishing all patterns for replication in later phases.

</domain>

<decisions>
## Implementation Decisions

### HTTP Client Design
- Configurable retries: user sets `max_retries` and `base_delay` in provider config block (with sensible defaults)
- Per-resource Terraform timeouts via terraform-plugin-framework timeouts block (create/read/update/delete individually)
- No global HTTP client timeout — resource-level timeouts govern operation duration
- Retry on transient errors (429, 503, 5xx) with exponential backoff

### Provider Config Schema
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

### File System Resource Scope
- Full coverage of all API attributes — expose everything the API provides
- Full NFS/SMB protocol blocks inline on the resource:
  ```hcl
  nfs {
    enabled     = true
    v3_enabled  = true
    v4_1_enabled = true
    rules       = "..."
  }
  smb {
    enabled              = true
    access_based_enum    = false
  }
  ```
- Space attributes (total, used, virtual, unique, snapshots) exposed as computed — useful for monitoring in Terraform output
- Snapshot directory hidden by default (follow Pure Storage defaults, don't expose as configurable)
- ID as primary identifier internally — API ID used for all CRUD calls, stable across renames
- Name is user-facing key — used for import, display, and user references
- Explicit defaults in schema matching API defaults — user sees them in plan
- In-place rename supported via PATCH (name change does not force recreation)
- Multi-protocol allowed freely (NFS + SMB simultaneously) — no provider-level restriction
- Policy references accept plain strings — users can pass literal names or resource references (Terraform handles either)

### Soft-Delete Behavior
- Configurable `destroy_eradicate_on_delete` boolean attribute on the resource (default: `true`)
  - `true` = PATCH destroyed=true then DELETE (full removal, name reusable)
  - `false` = PATCH destroyed=true only (recoverable within eradication window)
- If file system is soft-deleted outside Terraform: keep in state with `destroyed = true` attribute visible — drift is shown, user decides
- Synchronous eradication: when `destroy_eradicate_on_delete = true`, poll until fully eradicated before returning — guarantees name is reusable immediately after `terraform destroy`
- This pattern (configurable eradicate, sync polling, drift visibility) becomes the template for buckets and object store accounts in Phase 2

### Claude's Discretion
- HTTP log levels and verbosity (request/response at DEBUG vs TRACE)
- Exact retry defaults (max_retries count, base_delay duration)
- Go project scaffolding details (Makefile targets, CI config, linting setup)
- Internal error classification logic (retryable vs terminal)

</decisions>

<specifics>
## Specific Ideas

- Provider should feel like a first-party HashiCorp provider — follow terraform-plugin-framework conventions exactly
- Ops team does high-frequency CRUD — every `apply → plan` cycle must be clean (0 changes if nothing drifted)
- The file system resource is the template — every pattern decision here carries forward to 12+ resources
- Research flagged: OAuth2 uses non-standard `urn:ietf:params:oauth:grant-type:token-exchange` grant type — may need custom TokenSource instead of standard golang.org/x/oauth2

</specifics>

<code_context>
## Existing Code Insights

### Reusable Assets
- None — greenfield project, no existing code

### Established Patterns
- None — Phase 1 establishes all patterns

### Integration Points
- `FLASHBLADE_API.md` in repo root — AI-optimized API reference covering all 538 operations
- Go module path: `github.com/soulkyu/terraform-provider-flashblade`

</code_context>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 01-foundation*
*Context gathered: 2026-03-26*
