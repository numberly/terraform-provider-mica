---
phase: quick-260408-kbr
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - internal/client/models_admin.go
  - internal/client/array_connection_key.go
  - internal/client/array_connection_key_test.go
  - internal/testmock/handlers/array_connection_key.go
  - internal/provider/array_connection_key_resource.go
  - internal/provider/array_connection_key_resource_test.go
  - internal/provider/provider.go
  - examples/resources/flashblade_array_connection_key/resource.tf
autonomous: true
requirements: []

must_haves:
  truths:
    - "Apply generates a connection key via POST and stores it in state"
    - "Refresh reads the current key via GET and detects drift"
    - "connection_key attribute is marked Sensitive in schema"
    - "Destroy is a no-op (key expires automatically, just removes from state)"
    - "No ImportState (ephemeral resource, no stable identifier)"
  artifacts:
    - path: "internal/client/models_admin.go"
      provides: "ArrayConnectionKey struct (appended)"
      contains: "type ArrayConnectionKey struct"
    - path: "internal/client/array_connection_key.go"
      provides: "GetArrayConnectionKey, PostArrayConnectionKey client methods"
      exports: ["GetArrayConnectionKey", "PostArrayConnectionKey"]
    - path: "internal/testmock/handlers/array_connection_key.go"
      provides: "Mock handler for GET/POST /api/2.22/array-connections/connection-key"
    - path: "internal/provider/array_connection_key_resource.go"
      provides: "flashblade_array_connection_key resource"
  key_links:
    - from: "array_connection_key_resource.go"
      to: "client.PostArrayConnectionKey"
      via: "Create method"
    - from: "array_connection_key_resource.go"
      to: "client.GetArrayConnectionKey"
      via: "Read method"
---

<objective>
Implement `flashblade_array_connection_key` resource: a POST/GET-only ephemeral
key generator backed by `/array-connections/connection-key`.

Purpose: Operators can declaratively generate a connection key from Terraform
and consume it as a sensitive output to wire up the remote array.
Output: Client methods, mock handler, resource, tests, HCL example.
</objective>

<execution_context>
@$HOME/.claude/get-shit-done/workflows/execute-plan.md
@$HOME/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/STATE.md
@CONVENTIONS.md

## Key interfaces

From `internal/client/client.go`:
```go
func (c *FlashBladeClient) get(ctx context.Context, path string, result any) error
func (c *FlashBladeClient) post(ctx context.Context, path string, body, result any) error
```

These are low-level methods. Use them directly — the connection-key endpoint
returns a single object, NOT a ListResponse, so `postOne` / `getOneByName`
generics do NOT apply here.

From `internal/client/models_admin.go` (ArrayConnectionKey model — append here):
```go
// (does not exist yet — create it)
type ArrayConnectionKey struct {
    ConnectionKey string `json:"connection_key"`
    Created       int64  `json:"created"`
    Expires       int64  `json:"expires"`
}
```

From `internal/testmock/handlers/helpers.go` (shared mock helpers):
```go
func ValidateQueryParams(w, r, allowed []string) bool
func WriteJSONError(w, statusCode, message)
// Note: no WriteJSONListResponse here — response is a plain object, not a list
```

From `internal/provider/array_connection_resource.go` (neighboring resource
with same domain — use as style reference for Configure, timeouts, tflog.Debug).
</context>

<tasks>

<task type="auto">
  <name>Task 1: Model + Client + Mock handler</name>
  <files>
    internal/client/models_admin.go
    internal/client/array_connection_key.go
    internal/client/array_connection_key_test.go
    internal/testmock/handlers/array_connection_key.go
  </files>
  <action>
**1a. Append to `internal/client/models_admin.go`** (after existing ArrayConnection structs):

```go
// ArrayConnectionKey represents the response from GET/POST /array-connections/connection-key.
// There is only one connection key per array at a time. All fields are read-only.
type ArrayConnectionKey struct {
    ConnectionKey string `json:"connection_key"`
    Created       int64  `json:"created"`
    Expires       int64  `json:"expires"`
}
```

No Post/Patch structs — POST takes no body, no PATCH exists.

**1b. Create `internal/client/array_connection_key.go`**:

```go
package client

import "context"

// GetArrayConnectionKey retrieves the current connection key.
// Returns the key or an error. The endpoint returns a single object, not a list.
func (c *FlashBladeClient) GetArrayConnectionKey(ctx context.Context) (*ArrayConnectionKey, error) {
    var result ArrayConnectionKey
    if err := c.get(ctx, "/array-connections/connection-key", &result); err != nil {
        return nil, fmt.Errorf("GetArrayConnectionKey: %w", err)
    }
    return &result, nil
}

// PostArrayConnectionKey generates a new connection key, replacing any existing one.
// The POST takes no body. Returns the newly generated key.
func (c *FlashBladeClient) PostArrayConnectionKey(ctx context.Context) (*ArrayConnectionKey, error) {
    var result ArrayConnectionKey
    if err := c.post(ctx, "/array-connections/connection-key", nil, &result); err != nil {
        return nil, fmt.Errorf("PostArrayConnectionKey: %w", err)
    }
    return &result, nil
}
```

**1c. Create `internal/testmock/handlers/array_connection_key.go`**:

The mock stores at most one key (singleton). The store holds a `*client.ArrayConnectionKey`.
- GET: return current key as plain JSON object (HTTP 200); if none seeded, return `{}` with 200
- POST: generate a new key (synthetic values), overwrite current, return as plain JSON object

CRITICAL: Response is a plain JSON object, NOT `{"items": [...]}`. Use `json.NewEncoder(w).Encode(key)` directly, not `WriteJSONListResponse`.

Store struct:
```go
type arrayConnectionKeyStore struct {
    mu      sync.Mutex
    current *client.ArrayConnectionKey
    nextID  int
}

func RegisterArrayConnectionKeyHandlers(mux *http.ServeMux) *arrayConnectionKeyStore {
    store := &arrayConnectionKeyStore{nextID: 1}
    mux.HandleFunc("/api/2.22/array-connections/connection-key", store.handle)
    return store
}

func (s *arrayConnectionKeyStore) Seed(key *client.ArrayConnectionKey) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.current = key
}
```

GET handler: validate no unexpected query params, return current key or a zero-value key.
POST handler: generate synthetic key using `fmt.Sprintf("conn-key-%d", s.nextID)`, set Created/Expires to reasonable epoch values (e.g., 1000000000000 and 1000003600000), increment nextID.

**1d. Create `internal/client/array_connection_key_test.go`** (minimum 3 tests):

- `TestUnit_ArrayConnectionKey_Get` — seed key in mock, GET, verify fields match
- `TestUnit_ArrayConnectionKey_Post` — POST, verify returned key has non-empty ConnectionKey
- `TestUnit_ArrayConnectionKey_Get_AfterPost` — POST then GET, verify GET returns same key

Use `httptest.NewServer` with `RegisterArrayConnectionKeyHandlers`. Use existing `newTestClient(t, srv)` helper (same package).

Note: No Get_NotFound test — the API always returns a key (or empty object). No Delete test — no DELETE endpoint.
  </action>
  <verify>
    <automated>cd /home/gule/Workspace/team-infrastructure/terraform-provider-flashblade && go build ./internal/client/... && go test ./internal/client/... -run TestUnit_ArrayConnectionKey -v -count=1</automated>
  </verify>
  <done>3 client tests pass; `go build ./internal/client/...` is clean; mock handler compiles</done>
</task>

<task type="auto">
  <name>Task 2: Resource + tests + registration</name>
  <files>
    internal/provider/array_connection_key_resource.go
    internal/provider/array_connection_key_resource_test.go
    internal/provider/provider.go
    examples/resources/flashblade_array_connection_key/resource.tf
  </files>
  <action>
**2a. Create `internal/provider/array_connection_key_resource.go`**:

This is an UNUSUAL resource — deviates from the standard 4-interface pattern:
- Implements: `resource.Resource`, `resource.ResourceWithConfigure` ONLY
- NO `ResourceWithImportState` (key is ephemeral, no stable import identifier)
- NO `ResourceWithUpgradeState` (schema version 0, no migrations)

Interface assertions (top of file):
```go
var _ resource.Resource = &arrayConnectionKeyResource{}
var _ resource.ResourceWithConfigure = &arrayConnectionKeyResource{}
```

Model struct:
```go
type arrayConnectionKeyModel struct {
    ID            types.String   `tfsdk:"id"`
    ConnectionKey types.String   `tfsdk:"connection_key"`
    Created       types.Int64    `tfsdk:"created"`
    Expires       types.Int64    `tfsdk:"expires"`
    Timeouts      timeouts.Value `tfsdk:"timeouts"`
}
```

Schema (Version: 0):
- `id`: Computed, `UseStateForUnknown()` — stable synthetic ID
- `connection_key`: Computed, **Sensitive: true** — the key itself; `UseStateForUnknown()` since it never changes after POST
- `created`: Computed, `UseStateForUnknown()` — stable after creation
- `expires`: Computed, `UseStateForUnknown()` — stable after creation
- `timeouts`: Standard block (Create + Read only — no Update, no Delete)

Timeouts defaults: Create=20min, Read=5min.

**Create**: call `r.client.PostArrayConnectionKey(ctx)`, set all fields from result, set `id` to the `connection_key` value (use it as synthetic stable ID).

**Read**: call `r.client.GetArrayConnectionKey(ctx)`, log drift for all three fields via `tflog.Debug`. If the API returns an empty key string (key expired or array reset), call `resp.State.RemoveResource(ctx)` and return.

**Update**: should never be called (all attributes are Computed, no Required/Optional to change). Add a stub that calls `resp.Diagnostics.AddError("Update not supported", "All attributes are computed. Use -replace to regenerate the key.")`.

**Delete**: no-op. Keys expire automatically. Just return without error (do NOT call any API). Comment: "// Key expires automatically. No API call needed."

**UpgradeState**: return empty map `map[int64]resource.StateUpgrader{}`.

**2b. Create `internal/provider/array_connection_key_resource_test.go`** (minimum 3 tests):

- `TestUnit_ArrayConnectionKeyResource_Lifecycle` — Create (POST key), Read (GET), Delete (no-op). Verify `connection_key` in state is non-empty, `sensitive` attribute is not logged.
- `TestUnit_ArrayConnectionKeyResource_DriftDetection` — Seed a key, create resource, modify the mock key, Read again, verify drift is logged via `tflog`.
- `TestUnit_ArrayConnectionKeyResource_DeleteNoOp` — Verify Delete does not call the mock (counter stays at 0 or track POST call count).

Use `testmock.NewMockServer()` + `handlers.RegisterArrayConnectionKeyHandlers(ms.Mux)`.
Use `testNewMockedProvider()` helper pattern from neighboring resource tests (see `array_connection_resource_test.go` for reference).

**2c. Register in `internal/provider/provider.go`**:

In `Resources()` method, add `NewArrayConnectionKeyResource` near `NewArrayConnectionResource` (same domain group). No data source needed.

**2d. Create `examples/resources/flashblade_array_connection_key/resource.tf`**:

```hcl
# Generate a connection key for use by the remote array.
# The key is ephemeral — each apply regenerates it.
resource "flashblade_array_connection_key" "key" {}

output "connection_key" {
  value     = flashblade_array_connection_key.key.connection_key
  sensitive = true
}
```

No `import.sh` — ImportState not implemented (key is ephemeral).
  </action>
  <verify>
    <automated>cd /home/gule/Workspace/team-infrastructure/terraform-provider-flashblade && make build && go test ./internal/provider/... -run TestUnit_ArrayConnectionKey -v -count=1</automated>
  </verify>
  <done>
- `make build` passes (0 errors)
- 3 resource tests pass with `TestUnit_ArrayConnectionKey` prefix
- `make test` total count >= 753 (745 + 3 client + 3 resource + 2 registration)
- `make lint` clean
  </done>
</task>

</tasks>

<verification>
```bash
cd /home/gule/Workspace/team-infrastructure/terraform-provider-flashblade
make build   # 0 errors
make test    # >= 753 tests pass
make lint    # 0 issues
```

Functional checks:
- `flashblade_array_connection_key` registered in provider
- `connection_key` attribute has `Sensitive: true`
- No ImportState registered (intentional)
- Delete is a no-op (no API call)
</verification>

<success_criteria>
- `make build` passes
- `make test` passes with total >= 753
- `make lint` reports 0 issues
- `flashblade_array_connection_key` resource functional: Create (POST), Read (GET), Delete (no-op)
- `connection_key` is Sensitive in schema
- No `make docs` needed (no new schema requiring doc regeneration — but run `make docs` to confirm no diff)
</success_criteria>

<output>
After completion, create `.planning/quick/260408-kbr-add-flashblade-array-connection-key-reso/260408-kbr-SUMMARY.md`
</output>
