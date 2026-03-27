# Phase 2: Object Store Resources - Context

**Gathered:** 2026-03-27
**Status:** Ready for planning

<domain>
## Phase Boundary

Complete object store resource chain: account, bucket, and access key with dependency ordering. Replicates Phase 1 CRUD/import/drift patterns on 3 new resources with resource-specific lifecycle decisions (soft-delete defaults, immutable keys, dependency validation).

</domain>

<decisions>
## Implementation Decisions

### Bucket Attribute Scope
- Full coverage of all API attributes — consistent with Phase 1 decision for file systems
- Versioning exposed as simple string attribute: `versioning = "enabled" | "suspended"` — mirrors API directly
- Quota attributes inline on bucket (`quota_limit`, `hard_limit_enabled`) AND available via quota policy attachment (Phase 4) — both approaches supported
- Space attributes (total, used, virtual) as computed on bucket resource — same pattern as file system
- **Bucket name: ForceNew on rename** — unlike file system (in-place). Rationale: S3 clients hardcode bucket names, rename would break them silently
- **Account reference: RequiresReplace** — account is immutable after creation per API, changing account forces destroy + recreate
- Object lock and retention_lock attributes exposed but not deeply validated — full WORM support deferred to v2 (ESR-01)

### Access Key Lifecycle
- Secret access key stored in state at creation, marked Sensitive, UseStateForUnknown on subsequent reads — write-once secret pattern
- **ForceNew on all attributes** — access keys are immutable API objects, any change = destroy old + create new
- **No import support for access keys** — secret_access_key is unavailable after creation, import would produce incomplete state. Keys should be created via Terraform only.
- Access key references account by name string (`object_store_account` attribute)

### Account-Bucket Dependency
- **Dual enforcement**: Terraform implicit deps handle ordering when user references resources; provider returns clear diagnostic if API returns account-not-found
- Full API attribute coverage on accounts — consistent with all other resources
- **Account deletion fails if buckets exist** — provider returns clear diagnostic: "account has existing buckets, destroy them first." No cascade delete.
- **Account name: ForceNew on rename** — renaming would break bucket references. Consistent with bucket ForceNew-on-rename.
- Access keys reference accounts by name (not by user object)

### Soft-Delete Behavior
- **Bucket**: Same Phase 1 pattern (configurable `destroy_eradicate_on_delete`, sync poll, destroyed visible in drift) BUT **default is `false`** (recoverable by default) — buckets are higher-risk data containers
- **Object store account**: Simple DELETE — accounts don't hold data directly, no two-phase soft-delete needed. API may not support soft-delete for accounts.
- **Non-empty bucket delete**: Fail with clear diagnostic — "bucket contains objects, empty it first." No force_destroy auto-empty.

### Claude's Discretion
- Exact model struct field mapping for ObjectStoreAccount, Bucket, AccessKey
- Mock server handler implementation for the 3 new resource types
- Test structure (follow Phase 1 patterns: TDD RED/GREEN, mock-based unit tests)
- Error message wording for dependency validation diagnostics

</decisions>

<specifics>
## Specific Ideas

- The dependency chain (account → bucket → access key) should work in a single `terraform apply` — Terraform's graph handles ordering
- Access keys are the most critical resource for the ops team — secret exposure must be zero (no plan output, no logs)
- Bucket delete safety is paramount — default to recoverable (`destroy_eradicate_on_delete = false`) because buckets hold production data

</specifics>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/provider/filesystem_resource.go` (750 lines): Full CRUD template with timeouts, soft-delete, drift detection, import — copy and adapt for bucket
- `internal/client/filesystems.go`: Client CRUD methods pattern — replicate for accounts, buckets, access keys
- `internal/client/models.go`: Add ObjectStoreAccount, Bucket, ObjectStoreAccessKey model structs
- `internal/testmock/handlers/filesystems.go`: Mock handler pattern — replicate for 3 new resource types
- `internal/provider/filesystem_data_source.go`: Data source template

### Established Patterns
- TDD: RED (failing tests) → GREEN (implementation) — 2 commits per task
- Client layer: zero terraform-plugin-framework imports, pure Go HTTP
- Resource: Configure → CRUD → Import → drift logging via tflog
- Mock server: thread-safe in-memory state with UUID IDs
- Soft-delete: PATCH destroyed=true → DELETE → PollUntilEradicated

### Integration Points
- `internal/provider/provider.go` — Add NewAccountResource, NewBucketResource, NewAccessKeyResource to Resources() and DataSources()
- `internal/client/client.go` — No changes needed (shared HTTP methods)
- `internal/testmock/server.go` — Register new handlers in NewMockServer or via RegisterXHandlers pattern

</code_context>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 02-object-store-resources*
*Context gathered: 2026-03-27*
