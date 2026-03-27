# Phase 2: Object Store Resources - Research

**Researched:** 2026-03-27
**Domain:** FlashBlade REST API v2.22 — Object Store Accounts, Buckets, Access Keys; terraform-plugin-framework resource patterns
**Confidence:** HIGH — all findings derived from the existing codebase, FLASHBLADE_API.md (authoritative), and Phase 1 established patterns

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**Bucket Attribute Scope**
- Full coverage of all API attributes — consistent with Phase 1 decision for file systems
- Versioning exposed as simple string attribute: `versioning = "enabled" | "suspended"` — mirrors API directly
- Quota attributes inline on bucket (`quota_limit`, `hard_limit_enabled`) AND available via quota policy attachment (Phase 4) — both approaches supported
- Space attributes (total, used, virtual) as computed on bucket resource — same pattern as file system
- **Bucket name: ForceNew on rename** — unlike file system (in-place). Rationale: S3 clients hardcode bucket names, rename would break them silently
- **Account reference: RequiresReplace** — account is immutable after creation per API, changing account forces destroy + recreate
- Object lock and retention_lock attributes exposed but not deeply validated — full WORM support deferred to v2 (ESR-01)

**Access Key Lifecycle**
- Secret access key stored in state at creation, marked Sensitive, UseStateForUnknown on subsequent reads — write-once secret pattern
- **ForceNew on all attributes** — access keys are immutable API objects, any change = destroy old + create new
- **No import support for access keys** — secret_access_key is unavailable after creation, import would produce incomplete state. Keys should be created via Terraform only.
- Access key references account by name string (`object_store_account` attribute)

**Account-Bucket Dependency**
- **Dual enforcement**: Terraform implicit deps handle ordering when user references resources; provider returns clear diagnostic if API returns account-not-found
- Full API attribute coverage on accounts — consistent with all other resources
- **Account deletion fails if buckets exist** — provider returns clear diagnostic: "account has existing buckets, destroy them first." No cascade delete.
- **Account name: ForceNew on rename** — renaming would break bucket references. Consistent with bucket ForceNew-on-rename.
- Access keys reference accounts by name (not by user object)

**Soft-Delete Behavior**
- **Bucket**: Same Phase 1 pattern (configurable `destroy_eradicate_on_delete`, sync poll, destroyed visible in drift) BUT **default is `false`** (recoverable by default) — buckets are higher-risk data containers
- **Object store account**: Simple DELETE — accounts don't hold data directly, no two-phase soft-delete needed.
- **Non-empty bucket delete**: Fail with clear diagnostic — "bucket contains objects, empty it first." No force_destroy auto-empty.

### Claude's Discretion
- Exact model struct field mapping for ObjectStoreAccount, Bucket, AccessKey
- Mock server handler implementation for the 3 new resource types
- Test structure (follow Phase 1 patterns: TDD RED/GREEN, mock-based unit tests)
- Error message wording for dependency validation diagnostics

### Deferred Ideas (OUT OF SCOPE)

None — discussion stayed within phase scope
</user_constraints>

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| OSA-01 | User can create an object store account with name | POST /api/2.22/object-store-accounts — name provided via query param `names`, body carries quota/bucket_defaults |
| OSA-02 | User can update object store account attributes | PATCH /api/2.22/object-store-accounts?names= — mutable fields: quota_limit, hard_limit_enabled, bucket_defaults, public_access_config |
| OSA-03 | User can destroy an object store account | DELETE /api/2.22/object-store-accounts — simple single-phase; provider validates no buckets exist before calling |
| OSA-04 | User can import an existing account into Terraform state by name | Same ImportState pattern as filesystem_resource.go; name is the import ID |
| OSA-05 | Data source returns object store account attributes by name or filter | GET /api/2.22/object-store-accounts?names= — same pattern as filesystem_data_source.go |
| BKT-01 | User can create a bucket with name, account reference, and optional settings | POST /api/2.22/buckets — body includes account object; name via query param |
| BKT-02 | User can update bucket attributes (quotas, versioning, policies) | PATCH /api/2.22/buckets?names= — mutable: quota_limit, hard_limit_enabled, versioning, public_access_config, eradication_config |
| BKT-03 | User can destroy a bucket (two-phase: mark destroyed, then eradicate) | PATCH destroyed=true then DELETE — same two-phase as file system; default destroy_eradicate_on_delete=false |
| BKT-04 | User can import an existing bucket into Terraform state by name | ImportState by name; account ref and all attributes populated from GET response |
| BKT-05 | Data source returns bucket attributes by name or filter | GET /api/2.22/buckets?names= with optional ?destroyed= filter |
| BKT-06 | Drift detection logs field-level diffs via tflog when Read finds state divergence | Same tflog.Info pattern as filesystem Read; check quota_limit, versioning, hard_limit_enabled |
| OAK-01 | User can create an object store access key for a given account | POST /api/2.22/object-store-access-keys — body: user={name: "<account>/admin"} |
| OAK-02 | User can delete an object store access key | DELETE /api/2.22/object-store-access-keys?names= |
| OAK-03 | Secret access key is marked Sensitive and only available at creation time | secret_access_key: Sensitive:true + UseStateForUnknown; write-once from POST response |
| OAK-04 | User can import an existing access key into Terraform state by name | NOTE: CONTEXT.md locks this as NO import support — see constraint above |
| OAK-05 | Data source returns access key attributes by name or filter | GET /api/2.22/object-store-access-keys?names= — secret_access_key will be empty in data source responses |
</phase_requirements>

---

## Summary

Phase 2 implements three tightly coupled resources — object store account, bucket, and access key — that form the complete S3 object storage provisioning chain on FlashBlade. The pattern is a direct extension of the Phase 1 file system resource: client CRUD methods in `internal/client/`, provider resource in `internal/provider/`, mock handlers in `internal/testmock/handlers/`.

The key differences from Phase 1 are: (1) ForceNew on name/account rather than in-place rename, (2) bucket soft-delete defaults to `false` instead of `true`, (3) access keys are write-once immutable objects with no Update path, (4) a three-resource dependency chain where account → bucket → access key must be created in order and destroyed in reverse order, and (5) the access key's `secret_access_key` field requires special handling — it is returned only at creation time and must never be overwritten on subsequent reads.

The FlashBlade API uses a user object as the target for access key creation. An access key belongs to an `/object-store-users` object, which is implicitly created per account. The `user` field on POST `/object-store-access-keys` takes a name in the format `<account-name>/admin` (the default user within the account).

**Primary recommendation:** Implement resources in dependency order: Account → Bucket → Access Key. Follow the filesystem_resource.go template exactly for Account and Bucket; Access Key is a stripped-down variant with Create + Delete + Read only (no Update, no Import).

---

## Standard Stack

### Core (unchanged from Phase 1)

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| terraform-plugin-framework | v1.19.0 | Schema, CRUD, plan modifiers | Project standard from Phase 1 |
| terraform-plugin-framework-timeouts | (existing) | Per-resource timeouts | Used in filesystem_resource.go |
| terraform-plugin-log/tflog | v0.10.0 | Structured drift logging | Project standard |
| internal/client | (project) | HTTP client layer | Zero framework imports — project pattern |

### No New Dependencies Required

All required libraries are already in go.mod from Phase 1. Phase 2 adds files only — no `go get` needed.

---

## Architecture Patterns

### Recommended File Layout (Phase 2 additions)

```
internal/
├── client/
│   ├── object_store_accounts.go       # GetAccount, ListAccounts, PostAccount, PatchAccount, DeleteAccount
│   ├── buckets.go                     # GetBucket, ListBuckets, PostBucket, PatchBucket, PollBucketUntilEradicated
│   ├── object_store_access_keys.go    # GetAccessKey, ListAccessKeys, PostAccessKey, DeleteAccessKey
│   └── models.go                      # ADD: ObjectStoreAccount, ObjectStoreAccountPost/Patch, Bucket, BucketPost, BucketPatch, ObjectStoreAccessKey, ObjectStoreAccessKeyPost
├── provider/
│   ├── object_store_account_resource.go       # flashblade_object_store_account CRUD + Import
│   ├── object_store_account_resource_test.go
│   ├── object_store_account_data_source.go    # flashblade_object_store_account data source
│   ├── object_store_account_data_source_test.go
│   ├── bucket_resource.go                     # flashblade_bucket CRUD + Import
│   ├── bucket_resource_test.go
│   ├── bucket_data_source.go                  # flashblade_bucket data source
│   ├── bucket_data_source_test.go
│   ├── object_store_access_key_resource.go    # flashblade_object_store_access_key Create + Delete + Read
│   ├── object_store_access_key_resource_test.go
│   ├── object_store_access_key_data_source.go # flashblade_object_store_access_key data source
│   └── object_store_access_key_data_source_test.go
└── testmock/handlers/
    ├── object_store_accounts.go   # In-memory account store, CRUD handlers
    ├── buckets.go                 # In-memory bucket store, CRUD handlers (account-aware)
    └── object_store_access_keys.go # In-memory access key store, Create + Delete handlers
```

### Pattern 1: Account Resource (Simple DELETE, ForceNew on Name)

Account differs from file system in two ways: no soft-delete (single-phase DELETE), and name is ForceNew (not updatable in-place).

```go
// internal/provider/object_store_account_resource.go
// Schema excerpt — key plan modifiers
"name": schema.StringAttribute{
    Required:    true,
    Description: "The name of the object store account. Changing this forces a new resource.",
    PlanModifiers: []planmodifier.String{
        stringplanmodifier.RequiresReplace(), // ForceNew on rename
    },
},
"id": schema.StringAttribute{
    Computed: true,
    PlanModifiers: []planmodifier.String{
        stringplanmodifier.UseStateForUnknown(),
    },
},
```

Delete function — single-phase (no soft-delete):
```go
func (r *objectStoreAccountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    var data objectStoreAccountModel
    resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
    // ... timeout setup ...

    // Validate no buckets exist before deleting.
    buckets, err := r.client.ListBuckets(ctx, ListBucketsOpts{AccountNames: []string{data.Name.ValueString()}})
    if err != nil && !client.IsNotFound(err) {
        resp.Diagnostics.AddError("Error checking account buckets", err.Error())
        return
    }
    if len(buckets) > 0 {
        resp.Diagnostics.AddError(
            "Account has existing buckets",
            fmt.Sprintf("Object store account %q has %d bucket(s). Destroy all buckets before deleting the account.", data.Name.ValueString(), len(buckets)),
        )
        return
    }

    if err := r.client.DeleteObjectStoreAccount(ctx, data.Name.ValueString()); err != nil {
        if client.IsNotFound(err) {
            return
        }
        resp.Diagnostics.AddError("Error deleting object store account", err.Error())
    }
}
```

### Pattern 2: Bucket Resource (ForceNew on Name AND Account, Default destroy_eradicate_on_delete=false)

```go
// Schema excerpt — key differences from file system
"name": schema.StringAttribute{
    Required: true,
    PlanModifiers: []planmodifier.String{
        stringplanmodifier.RequiresReplace(), // ForceNew — S3 clients hardcode names
    },
},
"account": schema.StringAttribute{
    Required:    true,
    Description: "The name of the object store account that owns this bucket.",
    PlanModifiers: []planmodifier.String{
        stringplanmodifier.RequiresReplace(), // RequiresReplace — immutable post-creation
    },
},
"destroy_eradicate_on_delete": schema.BoolAttribute{
    Optional: true,
    Computed: true,
    Default:  booldefault.StaticBool(false), // FALSE default — buckets hold production data
    Description: "When true, Terraform eradicates the bucket on destroy. Default false (recoverable).",
},
"versioning": schema.StringAttribute{
    Optional:    true,
    Computed:    true,
    Description: "Versioning state: 'none', 'enabled', or 'suspended'.",
},
```

Bucket POST body — account is an object with name field:
```go
// internal/client/models.go
type BucketPost struct {
    Account          NamedReference `json:"account"`              // {name: "<account-name>"}
    BucketType       string         `json:"bucket_type,omitempty"`
    HardLimitEnabled bool           `json:"hard_limit_enabled,omitempty"`
    QuotaLimit       string         `json:"quota_limit,omitempty"`
    Versioning       string         `json:"versioning,omitempty"`
    RetentionLock    string         `json:"retention_lock,omitempty"`
    ObjectLockConfig *ObjectLockConfig `json:"object_lock_config,omitempty"`
    EradicationConfig *EradicationConfig `json:"eradication_config,omitempty"`
}

type NamedReference struct {
    Name string `json:"name"`
    ID   string `json:"id,omitempty"`
}
```

Bucket DELETE — check non-empty before soft-delete:
```go
func (r *bucketResource) Delete(ctx context.Context, ...) {
    // ... get state ...

    // Validate bucket is empty (API will reject non-empty buckets anyway,
    // but we surface a clear diagnostic rather than letting the API error bubble up).
    bkt, err := r.client.GetBucket(ctx, data.Name.ValueString())
    if err == nil && bkt.ObjectCount > 0 {
        resp.Diagnostics.AddError(
            "Bucket is not empty",
            fmt.Sprintf("Bucket %q contains %d object(s). Empty the bucket before destroying.", data.Name.ValueString(), bkt.ObjectCount),
        )
        return
    }

    // Phase 1: soft-delete
    destroyed := true
    if _, err := r.client.PatchBucket(ctx, data.ID.ValueString(), BucketPatch{Destroyed: &destroyed}); err != nil {
        // ...
    }

    // Phase 2: eradicate only if destroy_eradicate_on_delete=true
    eradicate := data.DestroyEradicateOnDelete.ValueBool()
    if eradicate {
        // DeleteBucket + PollBucketUntilEradicated
    }
}
```

### Pattern 3: Access Key Resource (Create + Delete Only, Write-Once Secret)

Access key has no Update path (all attributes ForceNew) and no ImportState (secret unavailable after creation).

```go
// Ensure interface — no ResourceWithImportState
var _ resource.Resource = &objectStoreAccessKeyResource{}
var _ resource.ResourceWithConfigure = &objectStoreAccessKeyResource{}
// NOTE: NO resource.ResourceWithImportState — intentional per CONTEXT.md

// Schema excerpt
"secret_access_key": schema.StringAttribute{
    Computed:  true,
    Sensitive: true,  // Never shown in plan/apply output
    Description: "The secret access key. Only available at creation time; stored in state.",
    PlanModifiers: []planmodifier.String{
        stringplanmodifier.UseStateForUnknown(), // Preserves value on subsequent reads
    },
},
"access_key_id": schema.StringAttribute{
    Computed: true,
    PlanModifiers: []planmodifier.String{
        stringplanmodifier.UseStateForUnknown(),
    },
},
"object_store_account": schema.StringAttribute{
    Required: true,
    Description: "The name of the object store account to create the access key for.",
    PlanModifiers: []planmodifier.String{
        stringplanmodifier.RequiresReplace(), // ForceNew — immutable
    },
},
"enabled": schema.BoolAttribute{
    Optional: true,
    Computed: true,
    PlanModifiers: []planmodifier.Bool{
        boolplanmodifier.RequiresReplace(), // ForceNew — any change recreates key
    },
},
```

Access key POST — user field format:
```go
// The API POST body for access keys: user is an object store user, not an account.
// The default user within account "myaccount" is named "myaccount/admin".
type ObjectStoreAccessKeyPost struct {
    User            NamedReference `json:"user"`              // {name: "<account>/admin"}
    SecretAccessKey string         `json:"secret_access_key,omitempty"` // optional: bring-your-own secret
}
```

Create function — capture secret immediately:
```go
func (r *objectStoreAccessKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var data objectStoreAccessKeyModel
    resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

    // The user name for the account's default access is "<account>/admin".
    userName := data.ObjectStoreAccount.ValueString() + "/admin"

    key, err := r.client.PostObjectStoreAccessKey(ctx, client.ObjectStoreAccessKeyPost{
        User: client.NamedReference{Name: userName},
    })
    if err != nil {
        resp.Diagnostics.AddError("Error creating access key", err.Error())
        return
    }

    // Capture secret NOW — it is only returned at creation time.
    data.SecretAccessKey = types.StringValue(key.SecretAccessKey)
    data.AccessKeyID = types.StringValue(key.AccessKeyID)
    data.Name = types.StringValue(key.Name)
    data.Created = types.Int64Value(key.Created)
    data.Enabled = types.BoolValue(key.Enabled)

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
```

Read function — preserve secret from state:
```go
func (r *objectStoreAccessKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    var data objectStoreAccessKeyModel
    resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

    key, err := r.client.GetObjectStoreAccessKey(ctx, data.Name.ValueString())
    if err != nil {
        if client.IsNotFound(err) {
            resp.State.RemoveResource(ctx)
            return
        }
        resp.Diagnostics.AddError("Error reading access key", err.Error())
        return
    }

    // IMPORTANT: secret_access_key is NOT returned by GET — preserve state value.
    // UseStateForUnknown plan modifier handles this, but we must not overwrite with empty.
    data.AccessKeyID = types.StringValue(key.AccessKeyID)
    data.Name = types.StringValue(key.Name)
    data.Created = types.Int64Value(key.Created)
    data.Enabled = types.BoolValue(key.Enabled)
    // data.SecretAccessKey is NOT updated here — preserved from state.

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
```

### Pattern 4: Mock Handler for Account-Bucket Cross-Reference

Bucket mock handler must validate account existence (simulates API 400 on unknown account):

```go
// internal/testmock/handlers/buckets.go
type bucketStore struct {
    mu      sync.Mutex
    byName  map[string]*client.Bucket
    byID    map[string]*client.Bucket
    accounts *objectStoreAccountStore // cross-reference for account validation
}

func (s *bucketStore) handlePost(w http.ResponseWriter, r *http.Request) {
    // ... decode body ...
    // Validate account exists.
    if _, ok := s.accounts.byName[body.Account.Name]; !ok {
        writeError(w, http.StatusBadRequest, fmt.Sprintf("account %q not found", body.Account.Name))
        return
    }
    // ... create bucket ...
}
```

### Anti-Patterns to Avoid

- **Overwriting secret_access_key on Read:** The GET response returns `secret_access_key` as empty — do NOT set `data.SecretAccessKey = types.StringValue("")`. The field must remain as read from state.
- **Calling Update on access keys:** Access keys have no mutable fields via the provider (CONTEXT.md: ForceNew on all). Do not implement an `Update` method.
- **Using ID (not name) for access key DELETE:** FlashBlade access key DELETE uses `?names=` like most resources, not `?ids=`. Verify against API.
- **Forgetting the /admin user suffix:** Access keys are created for an `object-store-users` entity. The default user for account `myaccount` is `myaccount/admin`. Sending only the account name will fail.
- **Soft-deleting an account:** Accounts do not support `PATCH destroyed=true`. DELETE is direct. Provider must not attempt two-phase delete for accounts.

---

## API Field Mapping

### ObjectStoreAccount (from FLASHBLADE_API.md)

| API Field | Type | Writable | Provider Attribute | Plan Modifier |
|-----------|------|----------|--------------------|---------------|
| `id` | ro string | no | `id` | UseStateForUnknown |
| `name` | ro string | no (set at POST via `names` param) | `name` | RequiresReplace |
| `created` | ro integer | no | `created` | UseStateForUnknown |
| `quota_limit` | integer | yes | `quota_limit` | none |
| `hard_limit_enabled` | boolean | yes | `hard_limit_enabled` | none |
| `bucket_defaults` | object | yes | `bucket_defaults` block | none |
| `public_access_config` | object | yes | `public_access_config` block | none |
| `space` | object | no | `space` block | Computed-only |
| `object_count` | ro integer | no | `object_count` | Computed |
| `realms` | ro array | no | omit in v1 (Phase 4+) | — |
| `context` | ro object | no | omit in v1 | — |

**POST params:** `names` is a query parameter (not body field) on POST `/object-store-accounts`. Body carries the optional fields.

**Note:** The `name` field is `ro` in the API schema (cannot be changed via PATCH). POST uses the `names` query parameter to set the name — consistent with how file systems work but slightly different from the struct shape.

### Bucket (from FLASHBLADE_API.md)

| API Field | Type | Writable | Provider Attribute | Plan Modifier |
|-----------|------|----------|--------------------|---------------|
| `id` | ro string | no | `id` | UseStateForUnknown |
| `name` | ro string | no | `name` | RequiresReplace |
| `account` | object | no (POST only) | `account` string (name) | RequiresReplace |
| `created` | ro integer | no | `created` | UseStateForUnknown |
| `destroyed` | boolean | yes | `destroyed` | Computed |
| `time_remaining` | ro integer | no | `time_remaining` | Computed |
| `versioning` | string | yes (PATCH) | `versioning` | none |
| `quota_limit` | integer | yes | `quota_limit` | none |
| `hard_limit_enabled` | boolean | yes | `hard_limit_enabled` | none |
| `object_count` | ro integer | no | `object_count` | Computed |
| `bucket_type` | ro string | no | `bucket_type` | UseStateForUnknown |
| `retention_lock` | string | yes | `retention_lock` | none |
| `object_lock_config` | object | yes | `object_lock_config` block | none |
| `eradication_config` | object | yes | `eradication_config` block | none |
| `public_access_config` | object | yes | `public_access_config` block | none |
| `space` | object | no | `space` block | Computed-only |
| `qos_policy` | object | yes | omit v1 (ESR-02) | — |
| `storage_class` | object | yes | omit v1 | — |

**PATCH endpoint:** Uses `?names=` or `?ids=`. Using ID (like file system) is preferred for rename-safety — but bucket name is ForceNew so names or IDs both work. Use IDs for consistency with filesystem pattern.

**Account field on POST:** `account` is an object `{name: "<account-name>"}` in the POST body, not a query param.

### ObjectStoreAccessKey (from FLASHBLADE_API.md)

| API Field | Type | Writable | Provider Attribute | Plan Modifier |
|-----------|------|----------|--------------------|---------------|
| `name` | ro string | no | `name` | UseStateForUnknown (Computed) |
| `access_key_id` | ro string | no | `access_key_id` | UseStateForUnknown |
| `secret_access_key` | ro string | no (POST response only) | `secret_access_key` | UseStateForUnknown + Sensitive |
| `created` | ro integer | no | `created` | UseStateForUnknown |
| `enabled` | boolean | yes (PATCH) | `enabled` | RequiresReplace (ForceNew) |
| `user` | ro object | no | `object_store_account` (maps to user name) | RequiresReplace |
| `context` | ro object | no | omit | — |

**POST body:** `user` is `{name: "<account>/admin"}`. The `secret_access_key` field in POST body is optional (bring-your-own). The response includes `secret_access_key` only at creation.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead |
|---------|-------------|-------------|
| Polling for bucket eradication | Custom wait loop | Copy `PollUntilEradicated` from `filesystems.go`, rename to `PollBucketUntilEradicated` |
| PATCH semantics with absent fields | Custom JSON omitempty logic | Pointer types in PatchXxx structs (same as `FileSystemPatch`) |
| Thread-safe in-memory mock state | Custom sync wrapper | `sync.Mutex` + `byName`/`byID` maps (same pattern as `fileSystemStore`) |
| Read-at-end-of-write | Custom state sync | `readIntoState` helper (same pattern as `filesystemResource.readIntoState`) |
| Timeouts per operation | Custom context deadline | `terraform-plugin-framework-timeouts` (same as filesystem) |

---

## Common Pitfalls

### Pitfall 1: Access Key Secret Overwritten on Read

**What goes wrong:** The Read function calls `GetObjectStoreAccessKey` and sets `data.SecretAccessKey = types.StringValue(key.SecretAccessKey)`. The GET endpoint returns `secret_access_key` as empty string. This overwrites the value stored at creation time with `""`, making the secret permanently lost in state.

**How to avoid:** In the Read function, do NOT update `SecretAccessKey` from the API response. Only update non-secret fields. The `UseStateForUnknown` plan modifier prevents plan-time unknown values, but the Read function itself must not overwrite the field.

**Warning signs:** `secret_access_key` becomes `""` in state after the first `terraform refresh`. Test: `terraform apply` then `terraform refresh` — secret must remain non-empty.

### Pitfall 2: Account Deletion with Buckets Still Existing

**What goes wrong:** Provider calls DELETE on an account that still has buckets. The FlashBlade API will return an error, but it will be a raw API error without context. The operator sees a confusing API message instead of a clear "destroy buckets first" instruction.

**How to avoid:** Before calling DELETE, call `ListBuckets` with `account_names` filter. If any buckets are returned, return a diagnostic error: `"Object store account \"<name>\" has <N> bucket(s). Destroy all buckets before deleting the account."`. This runs the check in the provider before the API call.

**Warning signs:** Acceptance test for destroy ordering fails with API error instead of provider diagnostic.

### Pitfall 3: Bucket Name vs ID for PATCH

**What goes wrong:** `PatchBucket` uses `?names=<name>` instead of `?ids=<id>`. Since bucket name is ForceNew (not renameable), this works — but it is inconsistent with the `PatchFileSystem` pattern which uses IDs.

**How to avoid:** Store the bucket `ID` in state (Computed, UseStateForUnknown) and use `?ids=<id>` in PATCH calls, exactly like `PatchFileSystem`. This is the defensive choice: if a name ever becomes mutable in a future API version, ID-based updates are more stable.

### Pitfall 4: User Name Format for Access Key POST

**What goes wrong:** The POST body for access key creation sends `user: {name: "myaccount"}` (account name only). The API expects `user: {name: "myaccount/admin"}` (the default user within the account). The API returns 404 or 400 on wrong format.

**How to avoid:** Construct the user name as `data.ObjectStoreAccount.ValueString() + "/admin"` in the Create function. The mock handler must also validate this format.

### Pitfall 5: Bucket soft-delete Default Diverges from File System

**What goes wrong:** Developer copies `filesystem_resource.go` and forgets to change `booldefault.StaticBool(true)` to `booldefault.StaticBool(false)` for `destroy_eradicate_on_delete` on the bucket resource.

**How to avoid:** Explicitly set `Default: booldefault.StaticBool(false)` and add a unit test asserting the default is `false`. The difference is intentional: buckets hold production data and recovery should be the default.

### Pitfall 6: Account POST name is a Query Parameter, Not Body Field

**What goes wrong:** Developer puts `name` in the POST body struct (`ObjectStoreAccountPost`). For object store accounts (and most FlashBlade resources), `name` is set via the `?names=<name>` query parameter on POST, not in the JSON body. Sending `name` in the body may be silently ignored or cause a 400.

**How to avoid:** In `PostObjectStoreAccount(ctx, name string, body ObjectStoreAccountPost)`, pass name as a query parameter: `POST /api/2.22/object-store-accounts?names=<name>`. Check the pattern in `PostFileSystem` — file systems also use `?names=`. Verify in FLASHBLADE_API.md: the POST body schema for accounts lists `account_exports`, `bucket_defaults`, `hard_limit_enabled`, `quota_limit` — no `name` field.

### Pitfall 7: Non-Empty Bucket Error from API vs Provider Diagnostic

**What goes wrong:** Provider calls PATCH `destroyed=true` on a non-empty bucket. The API will reject this with an error. Without a pre-check, the operator sees a raw API error message instead of a clear Terraform diagnostic.

**How to avoid:** In the Delete function, before soft-deleting, check `bucket.ObjectCount > 0` from the Read response and return a clear diagnostic. The mock handler should simulate the API's rejection for the eradication test.

---

## Code Examples

### Client Method: PostObjectStoreAccount

```go
// internal/client/object_store_accounts.go
// Source: filesystems.go pattern + FLASHBLADE_API.md ObjectStoreAccountPost schema

func (c *FlashBladeClient) PostObjectStoreAccount(ctx context.Context, name string, body ObjectStoreAccountPost) (*ObjectStoreAccount, error) {
    path := "/object-store-accounts?names=" + url.QueryEscape(name)
    var resp ListResponse[ObjectStoreAccount]
    if err := c.post(ctx, path, body, &resp); err != nil {
        return nil, err
    }
    if len(resp.Items) == 0 {
        return nil, fmt.Errorf("PostObjectStoreAccount: empty response from server")
    }
    return &resp.Items[0], nil
}
```

### Client Method: PostObjectStoreAccessKey (captures secret)

```go
// internal/client/object_store_access_keys.go
func (c *FlashBladeClient) PostObjectStoreAccessKey(ctx context.Context, body ObjectStoreAccessKeyPost) (*ObjectStoreAccessKey, error) {
    var resp ListResponse[ObjectStoreAccessKey]
    if err := c.post(ctx, "/object-store-access-keys", body, &resp); err != nil {
        return nil, err
    }
    if len(resp.Items) == 0 {
        return nil, fmt.Errorf("PostObjectStoreAccessKey: empty response from server")
    }
    return &resp.Items[0], nil
    // IMPORTANT: resp.Items[0].SecretAccessKey is only populated here at creation time.
    // Subsequent GETs will return SecretAccessKey as "".
}
```

### Model Structs for models.go (additions)

```go
// internal/client/models.go — additions for Phase 2

// NamedReference is a generic {id, name} reference object used in multiple resources.
type NamedReference struct {
    ID   string `json:"id,omitempty"`
    Name string `json:"name,omitempty"`
}

// ObjectStoreAccount represents a FlashBlade object store account.
type ObjectStoreAccount struct {
    ID               string          `json:"id"`
    Name             string          `json:"name"`
    Created          int64           `json:"created,omitempty"`
    QuotaLimit       int64           `json:"quota_limit,omitempty"`
    HardLimitEnabled bool            `json:"hard_limit_enabled"`
    ObjectCount      int64           `json:"object_count,omitempty"`
    Space            Space           `json:"space,omitempty"`
    // bucket_defaults and public_access_config are objects — define sub-structs as needed
}

// ObjectStoreAccountPost is the POST body for object store accounts.
// NOTE: name is passed as a query parameter, not in the body.
type ObjectStoreAccountPost struct {
    QuotaLimit       string `json:"quota_limit,omitempty"`
    HardLimitEnabled *bool  `json:"hard_limit_enabled,omitempty"`
}

// ObjectStoreAccountPatch is the PATCH body.
type ObjectStoreAccountPatch struct {
    QuotaLimit       *string `json:"quota_limit,omitempty"`
    HardLimitEnabled *bool   `json:"hard_limit_enabled,omitempty"`
}

// Bucket represents a FlashBlade S3 bucket.
type Bucket struct {
    ID               string         `json:"id"`
    Name             string         `json:"name"`
    Account          NamedReference `json:"account"`
    Created          int64          `json:"created,omitempty"`
    Destroyed        bool           `json:"destroyed"`
    TimeRemaining    int64          `json:"time_remaining,omitempty"`
    Versioning       string         `json:"versioning,omitempty"`
    QuotaLimit       int64          `json:"quota_limit,omitempty"`
    HardLimitEnabled bool           `json:"hard_limit_enabled"`
    ObjectCount      int64          `json:"object_count,omitempty"`
    BucketType       string         `json:"bucket_type,omitempty"`
    RetentionLock    string         `json:"retention_lock,omitempty"`
    Space            Space          `json:"space,omitempty"`
}

// BucketPost is the POST body for bucket creation.
type BucketPost struct {
    Account          NamedReference `json:"account"`
    Versioning       string         `json:"versioning,omitempty"`
    QuotaLimit       string         `json:"quota_limit,omitempty"`
    HardLimitEnabled *bool          `json:"hard_limit_enabled,omitempty"`
    RetentionLock    string         `json:"retention_lock,omitempty"`
}

// BucketPatch is the PATCH body for bucket updates.
type BucketPatch struct {
    Destroyed        *bool   `json:"destroyed,omitempty"`
    Versioning       *string `json:"versioning,omitempty"`
    QuotaLimit       *string `json:"quota_limit,omitempty"`
    HardLimitEnabled *bool   `json:"hard_limit_enabled,omitempty"`
}

// ObjectStoreAccessKey represents a FlashBlade S3 access key.
type ObjectStoreAccessKey struct {
    Name            string         `json:"name"`
    AccessKeyID     string         `json:"access_key_id"`
    SecretAccessKey string         `json:"secret_access_key"` // Only populated at creation time
    Created         int64          `json:"created,omitempty"`
    Enabled         bool           `json:"enabled"`
    User            NamedReference `json:"user"`
}

// ObjectStoreAccessKeyPost is the POST body for access key creation.
type ObjectStoreAccessKeyPost struct {
    User NamedReference `json:"user"` // {name: "<account>/admin"}
}
```

### Provider Registration (provider.go additions)

```go
// internal/provider/provider.go — Resources() addition
func (p *FlashBladeProvider) Resources(_ context.Context) []func() resource.Resource {
    return []func() resource.Resource{
        NewFilesystemResource,
        NewObjectStoreAccountResource, // Phase 2
        NewBucketResource,             // Phase 2
        NewObjectStoreAccessKeyResource, // Phase 2
    }
}

func (p *FlashBladeProvider) DataSources(_ context.Context) []func() datasource.DataSource {
    return []func() datasource.DataSource{
        NewFilesystemDataSource,
        NewObjectStoreAccountDataSource, // Phase 2
        NewBucketDataSource,             // Phase 2
        NewObjectStoreAccessKeyDataSource, // Phase 2
    }
}
```

---

## State of the Art

| Phase 1 Pattern | Phase 2 Variation | Reason |
|-----------------|------------------|--------|
| `name` is Optional+updatable in-place | `name` is Required+RequiresReplace | Account and bucket names are immutable post-creation in FlashBlade API; S3 clients hardcode bucket names |
| `destroy_eradicate_on_delete` defaults `true` | Bucket defaults `false` | Buckets hold production S3 data; recoverable-by-default is safer |
| Full CRUD + Update | Access Key: Create + Delete + Read only | Access keys are immutable; ForceNew on all attributes means no Update is ever called |
| `ImportState` always implemented | Access Key: NO ImportState | `secret_access_key` is unavailable after creation; import would produce permanently incomplete state |
| Simple DELETE (file system: two-phase) | Account: simple DELETE; Bucket: two-phase | Accounts hold no data directly; buckets hold S3 objects (same risk as file systems) |

---

## Open Questions

1. **Account name is passed as query param on POST — confirmation needed**
   - What we know: FLASHBLADE_API.md POST body schema for accounts does not include a `name` field; the `names` query parameter pattern is used for all FlashBlade POST endpoints
   - What's unclear: Whether POST `/object-store-accounts` accepts `name` in the body as an alternative
   - Recommendation: Use `?names=<name>` on POST (consistent with filesystems.go pattern); verify in mock test by checking that omitting the query param causes a name-related error from the mock

2. **Access key user name format: `<account>/admin` assumption**
   - What we know: `ObjectStoreUser.name` is of form `<account>/<username>`; FLASHBLADE_API.md shows POST `/object-store-users` creates users under an account
   - What's unclear: Whether the default user is exactly `<account>/admin` or some other format; whether the user must be pre-created
   - Recommendation: Mock handler accepts `<account>/admin` as the user name; document in resource description that access keys are created for the account's default admin user. Validate against real array in acceptance test.

3. **ListBuckets by account name filter for pre-delete validation**
   - What we know: GET `/api/2.22/buckets` supports `names`, `ids`, `destroyed` params per FLASHBLADE_API.md
   - What's unclear: Whether a filter like `filter=account.name='myaccount'` works, or if there is a dedicated `account_names` query param
   - Recommendation: Use `filter=account.name='<account-name>'` which follows the FlashBlade filter syntax documented in FLASHBLADE_API.md Common Parameters; test in mock handler with filter parsing

---

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go test (stdlib) + terraform-plugin-testing v1.15.0 |
| Config file | none (go test native) |
| Quick run command | `go test ./internal/... -run TestUnit -count=1 -timeout 60s` |
| Full suite command | `go test ./internal/... -count=1 -timeout 300s` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| OSA-01 | Create account via POST | unit (mock) | `go test ./internal/provider/ -run TestObjectStoreAccount_Create -count=1` | ❌ Wave 0 |
| OSA-02 | Update account attributes | unit (mock) | `go test ./internal/provider/ -run TestObjectStoreAccount_Update -count=1` | ❌ Wave 0 |
| OSA-03 | Delete account (simple DELETE) | unit (mock) | `go test ./internal/provider/ -run TestObjectStoreAccount_Delete -count=1` | ❌ Wave 0 |
| OSA-03 | Delete with buckets fails gracefully | unit (mock) | `go test ./internal/provider/ -run TestObjectStoreAccount_DeleteWithBuckets -count=1` | ❌ Wave 0 |
| OSA-04 | Import account by name | unit (mock) | `go test ./internal/provider/ -run TestObjectStoreAccount_Import -count=1` | ❌ Wave 0 |
| OSA-05 | Data source reads account | unit (mock) | `go test ./internal/provider/ -run TestObjectStoreAccountDataSource -count=1` | ❌ Wave 0 |
| BKT-01 | Create bucket with account reference | unit (mock) | `go test ./internal/provider/ -run TestBucket_Create -count=1` | ❌ Wave 0 |
| BKT-02 | Update bucket versioning/quota | unit (mock) | `go test ./internal/provider/ -run TestBucket_Update -count=1` | ❌ Wave 0 |
| BKT-03 | Destroy bucket (two-phase, default recoverable) | unit (mock) | `go test ./internal/provider/ -run TestBucket_Delete -count=1` | ❌ Wave 0 |
| BKT-03 | Destroy non-empty bucket fails | unit (mock) | `go test ./internal/provider/ -run TestBucket_DeleteNonEmpty -count=1` | ❌ Wave 0 |
| BKT-04 | Import bucket by name | unit (mock) | `go test ./internal/provider/ -run TestBucket_Import -count=1` | ❌ Wave 0 |
| BKT-05 | Data source reads bucket | unit (mock) | `go test ./internal/provider/ -run TestBucketDataSource -count=1` | ❌ Wave 0 |
| BKT-06 | Drift detection logs diffs | unit (mock) | `go test ./internal/provider/ -run TestBucket_Drift -count=1` | ❌ Wave 0 |
| OAK-01 | Create access key, capture secret | unit (mock) | `go test ./internal/provider/ -run TestAccessKey_Create -count=1` | ❌ Wave 0 |
| OAK-02 | Delete access key | unit (mock) | `go test ./internal/provider/ -run TestAccessKey_Delete -count=1` | ❌ Wave 0 |
| OAK-03 | Secret is Sensitive, UseStateForUnknown | unit (mock) | `go test ./internal/provider/ -run TestAccessKey_SecretPreserved -count=1` | ❌ Wave 0 |
| OAK-04 | No import support (intentional) | n/a — documented omission | — | n/a |
| OAK-05 | Data source reads access key | unit (mock) | `go test ./internal/provider/ -run TestAccessKeyDataSource -count=1` | ❌ Wave 0 |

### Sampling Rate

- **Per task commit:** `go test ./internal/... -run TestUnit -count=1 -timeout 60s`
- **Per wave merge:** `go test ./internal/... -count=1 -timeout 300s`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps

All test files for Phase 2 resources are new (no existing infrastructure covers them):

- [ ] `internal/provider/object_store_account_resource_test.go` — covers OSA-01 through OSA-04
- [ ] `internal/provider/object_store_account_data_source_test.go` — covers OSA-05
- [ ] `internal/provider/bucket_resource_test.go` — covers BKT-01 through BKT-06
- [ ] `internal/provider/bucket_data_source_test.go` — covers BKT-05
- [ ] `internal/provider/object_store_access_key_resource_test.go` — covers OAK-01 through OAK-03, OAK-05
- [ ] `internal/provider/object_store_access_key_data_source_test.go` — covers OAK-05
- [ ] `internal/testmock/handlers/object_store_accounts.go` — mock account CRUD
- [ ] `internal/testmock/handlers/buckets.go` — mock bucket CRUD with account cross-ref
- [ ] `internal/testmock/handlers/object_store_access_keys.go` — mock access key create/delete

Framework already installed (`go test ./internal/...` works from Phase 1).

---

## Sources

### Primary (HIGH confidence)

- `FLASHBLADE_API.md` (repo root) — authoritative FlashBlade REST API v2.22 reference; all endpoint paths, body fields, and ro annotations verified directly
- `internal/provider/filesystem_resource.go` — Phase 1 template; all patterns replicated from here
- `internal/client/filesystems.go` — client method patterns (PollUntilEradicated, PATCH with IDs, etc.)
- `internal/client/models.go` — existing model conventions; Phase 2 models extend same file
- `internal/testmock/handlers/filesystems.go` — mock handler template (store struct, byName/byID, raw PATCH)
- `internal/testmock/server.go` — mock server registration pattern
- `.planning/phases/02-object-store-resources/02-CONTEXT.md` — locked user decisions

### Secondary (MEDIUM confidence)

- `.planning/research/PITFALLS.md` — project-scoped pitfall catalog for soft-delete, computed fields, sensitive values
- `.planning/research/ARCHITECTURE.md` — project architecture decisions; provider registration patterns
- `.planning/research/STACK.md` — dependency versions; all Phase 2 deps already present

---

## Metadata

**Confidence breakdown:**

- Standard stack: HIGH — no new dependencies; all existing patterns apply
- Architecture: HIGH — direct replication of filesystem_resource.go with documented variations
- API field mapping: HIGH — derived from FLASHBLADE_API.md authoritative schema
- Access key user name format: MEDIUM — `<account>/admin` is the standard pattern; needs acceptance test confirmation
- Account filter by account for ListBuckets pre-delete check: MEDIUM — `filter=` syntax is documented but untested against the real API

**Research date:** 2026-03-27
**Valid until:** 2026-05-27 (60 days — stable API, no fast-moving ecosystem concerns)
