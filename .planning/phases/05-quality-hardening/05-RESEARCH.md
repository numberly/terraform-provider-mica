# Phase 5: Quality Hardening - Research

**Researched:** 2026-03-28
**Domain:** Go testing patterns, terraform-plugin-framework schema testing, terraform-plugin-docs, GitHub Actions CI
**Confidence:** HIGH

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- **Full lifecycle mocked tests**: Every resource gets Create→Read→Update→Read→Delete mocked integration test + import→plan→0-diff test
- **Schema/validator unit tests**: Every resource gets schema correctness tests (required fields, plan modifiers, validators) — catch config regressions
- **Both are equally critical** — no prioritization compromise
- **HTTP retry tests**: Mock server returns 429/503 on first attempt, succeeds on retry. Verify client handles transparently.
- **API error handling**: Every API error (404, 409 conflict, 422 validation) produces a clear Terraform diagnostic — test all error paths
- **Pagination**: List operations with continuation_token return all results — test mock server pagination behavior
- NOT testing concurrent modifications (out of scope for unit tests — would need real array)
- **GitHub Actions workflow** for automated testing on push/PR
- **terraform-plugin-docs** for auto-generation from schema + examples/ directory
- **Comprehensive examples**: Multiple HCL examples per resource — basic, advanced, import, and common patterns
- **Standard provider README**: Installation, configuration, quick start, resource list, contributing guide

### Claude's Discretion
- Which specific tests are missing vs already covered from Phases 1-4
- Test file organization (one file per resource vs grouped)
- terraform-plugin-docs template customization
- GitHub Actions workflow structure (matrix, caching, etc.)
- README structure and content depth

### Deferred Ideas (OUT OF SCOPE)
None — discussion stayed within phase scope
</user_constraints>

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| QUA-01 | All resources implement correct plan modifiers (UseStateForUnknown for stable computed, RequiresReplace for immutable) | Audit confirms 16/19 resources have no plan-modifier assertions in tests; production code has modifiers, but correctness is unverified by tests |
| QUA-02 | All resources validate input at plan time (invalid quota values, enum fields, required references) | Only `object_store_access_policy_rule` has a `stringvalidator.OneOf("allow","deny")` — no other resources have plan-time validators; no test verifies rejection of invalid values |
| QUA-03 | Unit tests cover all schema definitions, validators, and plan modifiers | Schema helper functions exist in all test files; but explicit plan-modifier/validator assertions exist only for 3/19 resources; no 409/422 diagnostic tests exist |
| QUA-04 | Mocked API integration tests cover CRUD lifecycle for all resource families | CRUD lifecycle tests exist for all 19 resources; no standardized "plan→apply→0-diff" idempotency test across all resources; filesystem has 13 granular tests, others have 4-6 basic ones |
| QUA-05 | HTTP client implements retry with exponential backoff for transient API errors | ALREADY COMPLETE — `transport.go` implements exponential backoff; 3 tests in `transport_test.go` cover 429, 503, and max-retries |
| QUA-06 | Provider documentation generated via terraform-plugin-docs for all resources and data sources | `terraform-plugin-docs v0.24.0` already in `go.mod`; `tools/tools.go` pins the binary; `docs/` directory does not exist; `examples/` has only 3 files (provider + file_system resource + file_system data source); 18 resources and 13 data sources need example files; `.github/workflows/` does not exist |
</phase_requirements>

---

## Summary

Phase 5 is a gap-filling phase on a codebase with 147 passing tests across 4 packages. The retry infrastructure (QUA-05) is already complete and tested. The primary work falls into four buckets: schema/plan-modifier/validator unit tests (QUA-01/02/03), error-path diagnostic tests (409/422 coverage), documentation scaffolding (QUA-06), and CI workflow creation.

The most significant gap is QUA-03: every resource test file has a schema-loading helper (`*ResourceSchema(t)`) but only 3 of 19 resources assert plan-modifier correctness. No test in the codebase verifies that invalid validator inputs are rejected at plan time. The only plan-time validator in the codebase is `stringvalidator.OneOf("allow","deny")` on the OAP rule `effect` field — it is untested.

For documentation (QUA-06), the tooling is fully wired: `terraform-plugin-docs v0.24.0` is pinned, `go generate ./...` is the entry point, but no `docs/` directory exists and `examples/` is nearly empty (only `flashblade_file_system`). The tool needs a complete `examples/` tree before it can produce meaningful output.

**Primary recommendation:** QUA-05 is done — skip it. Work in order: QUA-01/02/03 (schema+validator tests per resource family), QUA-04 (lifecycle completeness audit per family), error-path tests (409/422), pagination test, then QUA-06 (examples + docs generation + README + CI).

---

## Current State Audit

### Test Count Breakdown (147 total, all passing)

| Package | Tests | Notes |
|---------|-------|-------|
| `internal/provider` | 118 | All resource + data source mocked tests |
| `internal/client` | 26 | Client, filesystems, auth, transport |
| `internal/testmock` | 3 | Mock server lifecycle, login/version |

### QUA-05: Retry — ALREADY COMPLETE

| Test | Status |
|------|--------|
| `TestUnit_RetryTransport_429` | PASS — 429→retry→200 |
| `TestUnit_RetryTransport_503` | PASS — 503→retry→200 |
| `TestUnit_RetryTransport_MaxRetries` | PASS — exhausts retries, returns final 503 |

`transport.go` implements exponential backoff capped at 30 seconds. `IsRetryable()` covers 429 and all 5xx. Body replay is implemented for POST/PATCH retries. **No work needed for QUA-05.**

### QUA-03: Plan Modifier Tests — 16/19 Resources Missing

Resources with explicit `RequiresReplace`/`UseStateForUnknown` assertions in tests:

| Resource | What is tested |
|----------|----------------|
| `object_store_access_key` | `RequiresReplace` on `object_store_account` and `enabled` |
| `object_store_account` | `RequiresReplace` on `name` |
| `snapshot_policy` | Comment only ("Name change is RequiresReplace") — no assertion |

Resources with `RequiresReplace` in production code but NO test assertion:

| Resource | Immutable fields (RequiresReplace in schema) |
|----------|---------------------------------------------|
| `bucket` | `name`, `account` |
| `object_store_access_policy_rule` | `effect`, `policy_name`, `name` |
| `network_access_policy_rule` | `policy_name` |
| `smb_share_policy_rule` | `policy_name` |
| `snapshot_policy` | `name` (comment only, no assertion) |
| `snapshot_policy_rule` | `policy_name` |
| `nfs_export_policy_rule` | (no explicit RequiresReplace found; `policy_name` likely required) |
| `quota_group` | `file_system_name`, `gid` |
| `quota_user` | `file_system_name`, `uid` |

Resources with no `RequiresReplace` in schema (singletons or rename-capable):
`filesystem`, `nfs_export_policy`, `smb_share_policy`, `array_dns`, `array_ntp`, `array_smtp`, `network_access_policy`, `object_store_access_policy`

### QUA-02: Validators — Only 1 Exists, Not Tested for Rejection

| Resource | Field | Validator | Invalid Test |
|----------|-------|-----------|-------------|
| `object_store_access_policy_rule` | `effect` | `stringvalidator.OneOf("allow","deny")` | None — valid values only tested |

No other resource has `Validators:` defined in schema. This means Phase 5 must both:
1. Add missing validators where input can be invalid (quota values, enum fields)
2. Add tests that verify invalid inputs are rejected

### QUA-04: Lifecycle Coverage — Exists But Uneven

All 19 resources have Create/Read/Delete mocked tests. The `filesystem` resource is the gold standard with 13 granular tests including drift-log, idempotency, and soft-only-delete. Other resources follow a simpler 4-test pattern (Create, Update, Delete, Import).

**Missing across all resources:**
- No standardized "apply → read back → 0-diff" idempotency test
- No "Read removes state on 404" test (exists only for `filesystem` and `network_access_policy`)

### QUA-04: Error Path Coverage — 409/422 Missing from Provider Layer

The mock handlers emit 404 (not found), 409 (conflict/already exists), 422 (unprocessable), and 400 (bad request) responses. The provider tests currently do NOT exercise 409 or 422 error paths. 404 is tested only for `filesystem` (removes state) and `object_store_account` data source.

| Error Code | Mock Handler Emits | Provider Test Covers |
|------------|-------------------|---------------------|
| 404 | All handlers | `filesystem` Read, `object_store_account` DS, `network_access_policy` Create |
| 409 Conflict | `buckets`, `nfs_export_policies`, `object_store_access_policies`, `quotas` | None |
| 422 / 400 | `filesystems` (delete-before-destroy) | None in provider layer |

### QUA-04: Pagination — Not Tested at Any Layer

`ListFileSystems` accepts `ContinuationToken` and `Limit` parameters. The client passes them through but does **not** auto-paginate (it's a single-page call). Data sources that call list endpoints use single-page results. No test simulates a paginated response (page 1 returns items + token, page 2 returns remaining items).

The CONTEXT.md decision is: "List operations with continuation_token return all results — test mock server pagination behavior." This implies the client list function needs auto-pagination logic, or the data source layer needs to implement it.

### QUA-06: Documentation — Tooling Ready, Content Missing

| Item | Status |
|------|--------|
| `terraform-plugin-docs v0.24.0` in `go.mod` | PRESENT |
| `tools/tools.go` pins `tfplugindocs` binary | PRESENT |
| `GNUmakefile` `generate` target: `go generate ./...` | PRESENT |
| `go:generate` directive in main or provider package | MISSING — must add `//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs` |
| `docs/` directory | MISSING — created by `tfplugindocs` on first run |
| `templates/` directory | MISSING — optional; default templates used if absent |
| `examples/provider/provider.tf` | PRESENT |
| `examples/resources/flashblade_file_system/resource.tf` | PRESENT |
| `examples/data-sources/flashblade_file_system/data-source.tf` | PRESENT |
| All other resource examples (18 resources) | MISSING |
| All other data source examples (13 data sources, excl. file_system) | MISSING |
| `README.md` | MISSING — not present at repo root |
| `.github/workflows/` | MISSING |

---

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `terraform-plugin-framework` | v1.19.0 (in go.mod) | Schema, resource, provider interfaces | Already used throughout |
| `terraform-plugin-framework-validators` | v0.19.0 | `stringvalidator.OneOf`, `int64validator.AtLeast`, etc. | Already imported in OAP rule |
| `terraform-plugin-docs` | v0.24.0 | Documentation generation from schema + examples | Already pinned in tools.go |
| `net/http/httptest` | stdlib | Mock server for tests | Already used via testmock package |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `terraform-plugin-go/tftypes` | v0.31.0 | Build `tftypes.Object` for test state construction | All provider unit tests |
| `terraform-plugin-framework/tfsdk` | v1.19.0 | `tfsdk.State`, `tfsdk.Plan` for direct resource method testing | All provider unit tests |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Direct resource method testing (current) | `resource/testing` helper framework | Current approach is lighter and faster; testing framework adds acceptance test overhead |
| `stringvalidator.OneOf` | Custom validator | OneOf handles enum fields cleanly; custom only for complex cross-field validation |

**Installation:** Nothing new needed — all dependencies already in `go.mod`.

---

## Architecture Patterns

### Recommended Project Structure

No changes to existing layout. Phase 5 adds files within existing directories:

```
internal/
├── provider/                          # Existing — add tests to existing _test.go files
│   └── *_resource_test.go            # Add plan-modifier, validator, error-path tests here
├── client/
│   └── filesystems.go                # May need auto-pagination logic
examples/
├── provider/provider.tf              # EXISTS
├── resources/
│   ├── flashblade_file_system/       # EXISTS
│   ├── flashblade_bucket/            # ADD resource.tf
│   ├── flashblade_object_store_account/ # ADD resource.tf
│   └── ... (one dir per resource)
└── data-sources/
    ├── flashblade_file_system/        # EXISTS
    └── ... (one dir per data source)
docs/                                  # GENERATED by tfplugindocs — do not hand-edit
.github/
└── workflows/
    └── ci.yml                         # ADD
README.md                              # ADD
```

### Pattern 1: Plan Modifier Assertion in Schema Test

**What:** Inspect `schema.Attribute.PlanModifiers` slice directly in unit test
**When to use:** QUA-01 — verify immutable fields have `RequiresReplace`, computed stable fields have `UseStateForUnknown`

```go
// Source: existing pattern in internal/provider/object_store_access_key_resource_test.go
func TestUnit_BucketResource_PlanModifiers(t *testing.T) {
    s := bucketResourceSchema(t).Schema

    // name must have RequiresReplace
    nameAttr, ok := s.Attributes["name"].(schema.StringAttribute)
    if !ok {
        t.Fatal("name attribute not found or wrong type")
    }
    hasRR := false
    for _, pm := range nameAttr.PlanModifiers {
        if _, ok := pm.(stringplanmodifier.RequiresReplaceModifier); ok {
            hasRR = true
        }
    }
    if !hasRR {
        t.Error("expected RequiresReplace plan modifier on name attribute")
    }

    // id must have UseStateForUnknown
    idAttr, ok := s.Attributes["id"].(schema.StringAttribute)
    // ...same pattern...
}
```

### Pattern 2: Validator Rejection Test

**What:** Call resource `Create` with an invalid enum value, assert diagnostic error
**When to use:** QUA-02 — verify `stringvalidator.OneOf` rejects invalid input at plan time

```go
// Source: pattern derived from existing Create tests + framework validator behavior
// NOTE: Validators run during framework plan/apply, not directly in Create method.
// Test approach: build a plan with invalid value, call ValidateConfig if available,
// OR test that the Create fails with a diagnostic error.
func TestUnit_OAPRule_InvalidEffect(t *testing.T) {
    ms := testmock.NewMockServer()
    defer ms.Close()
    handlers.RegisterObjectStoreAccessPolicyHandlers(ms.Mux)

    r := newTestOAPRuleResource(t, ms)
    plan := oapRulePlan(t, "test-policy", "test-rule", "invalid-effect", []string{"s3:GetObject"}, []string{"arn:aws:s3:::*"})
    resp := &resource.CreateResponse{State: tfsdk.State{Schema: oapRuleResourceSchema(t).Schema}}
    r.Create(context.Background(), resource.CreateRequest{Plan: plan}, resp)

    if !resp.Diagnostics.HasError() {
        t.Error("expected error diagnostic for invalid effect value")
    }
}
```

**Important caveat:** `stringvalidator.OneOf` runs at the framework plan stage before `Create` is called. In direct method tests, validators are bypassed. The test should instead check the schema attribute has `Validators` populated and test the validator's `ValidateString` method directly.

### Pattern 3: 404 Removes State from Read

**What:** Mock server returns 404; `Read` should set `resp.State.RemoveResource(ctx)`
**When to use:** QUA-04 — verify all resources handle "disappeared" resources gracefully

```go
// Source: filesystem_resource_test.go TestUnit_FileSystem_Read_NotFound (line 452)
func TestUnit_BucketResource_Read_NotFound(t *testing.T) {
    ms := testmock.NewMockServer()
    defer ms.Close()
    // Register handler that returns 404 for bucket GET
    ms.RegisterHandler("/api/2.22/buckets", func(w http.ResponseWriter, r *http.Request) {
        if r.Method == http.MethodGet {
            handlers.WriteJSONError(w, http.StatusNotFound, "bucket not found")
            return
        }
    })

    r := newTestBucketResource(t, ms)
    s := bucketResourceSchema(t).Schema
    state := tfsdk.State{Raw: tftypes.NewValue(buildBucketType(), nullBucketConfig()), Schema: s}
    resp := &resource.ReadResponse{State: state}
    r.Read(ctx, resource.ReadRequest{State: state}, resp)

    if !resp.State.Raw.IsNull() {
        t.Error("expected state to be removed (null) when resource not found")
    }
}
```

### Pattern 4: 409 Conflict → Clear Diagnostic

**What:** Mock server returns 409 on Create; assert diagnostic has meaningful error summary
**When to use:** QUA-04 — verify create-conflict produces usable error, not panic

```go
// New pattern — no existing example in codebase
func TestUnit_BucketResource_Create_Conflict(t *testing.T) {
    ms := testmock.NewMockServer()
    defer ms.Close()
    ms.RegisterHandler("/api/2.22/buckets", func(w http.ResponseWriter, r *http.Request) {
        handlers.WriteJSONError(w, http.StatusConflict, "bucket already exists")
    })

    r := newTestBucketResource(t, ms)
    plan := bucketPlanWithNameAndAccount(t, "existing-bucket", "myaccount")
    resp := &resource.CreateResponse{}
    r.Create(ctx, resource.CreateRequest{Plan: plan}, resp)

    if !resp.Diagnostics.HasError() {
        t.Error("expected error diagnostic on 409 conflict")
    }
}
```

### Pattern 5: Pagination in List Operations

**What:** Mock server returns page 1 (items + continuation_token), page 2 (remaining items); verify all items are returned
**When to use:** QUA-04 — continuation_token must collect all pages

```go
// New pattern — client-level test
func TestUnit_FileSystem_List_Paginated(t *testing.T) {
    page := 0
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        page++
        if page == 1 {
            // Return 2 items + continuation_token
            writeJSON(w, ListResponse{
                Items: []FileSystem{{Name: "fs-1"}, {Name: "fs-2"}},
                ContinuationToken: "token-page-2",
            })
        } else {
            // Return final page with no token
            writeJSON(w, ListResponse{Items: []FileSystem{{Name: "fs-3"}}})
        }
    }))
    // ... verify client returns all 3 items
}
```

**Key decision needed:** `ListFileSystems` currently passes `ContinuationToken` as input but does NOT auto-paginate (it returns a single page). The CONTEXT.md decision "List operations with continuation_token return all results" implies the client or data source layer must auto-paginate. Phase 5 must implement this before testing it.

### Pattern 6: terraform-plugin-docs Generation

**What:** `go:generate` directive + `examples/` directory → `docs/` output
**When to use:** QUA-06 — one-time setup, then `make generate` produces docs

```bash
# In main.go or provider.go, add:
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-name flashblade

# In GNUmakefile, docs target:
docs:
    go generate ./...
```

Required `examples/` structure (tfplugindocs convention):
```
examples/
├── provider/provider.tf                          # Provider configuration example
├── resources/<resource_type>/resource.tf         # Resource usage example
└── data-sources/<data_source_type>/data-source.tf # Data source usage example
```

`tfplugindocs` reads `Description` fields from schema attributes to generate attribute tables. Descriptions are already populated for all resources. Templates in `templates/` are optional (default templates are good enough).

### Anti-Patterns to Avoid

- **Testing validators via Create method directly:** Framework validators run before `Create` — they won't fire in direct method calls. Test validator behavior by calling the validator's `ValidateString` method directly, or add the validator to `schema.Attribute.Validators` and inspect the slice.
- **Hand-editing `docs/`:** docs/ is generated — never commit manual edits. Regenerate with `make generate`.
- **Separate test file per gap:** Add plan-modifier tests to existing `*_resource_test.go` files, not new files. One file per resource stays maintainable.
- **Skipping the `go:generate` directive:** `make generate` only works if a `//go:generate` comment exists somewhere in the package. Add it to `main.go`.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Enum validation | Custom validator struct | `stringvalidator.OneOf(...)` from `terraform-plugin-framework-validators` | Already imported, handles all edge cases |
| Integer range validation | Custom validator | `int64validator.AtLeast(0)`, `int64validator.Between(a, b)` | Same package, already in go.mod |
| Doc generation | Manual markdown | `terraform-plugin-docs` via `go generate` | Reads schema descriptions automatically |
| GitHub Actions Go setup | Custom setup | `actions/setup-go@v5` + `actions/cache@v4` | Standard, handles go module cache |

**Key insight:** The validator and plan-modifier framework infrastructure is already imported and working. The gap is purely test assertions and missing `Validators:` declarations — not implementation complexity.

---

## Common Pitfalls

### Pitfall 1: Testing Framework Validators via Direct Method Calls

**What goes wrong:** `r.Create(ctx, req, resp)` is called in tests; framework validators registered in `schema.Attribute.Validators` are NOT invoked — they run at the plan stage in the full framework pipeline.

**Why it happens:** Provider unit tests bypass the full terraform-plugin-framework execution pipeline.

**How to avoid:** To verify a validator is present: inspect `schema.Attribute.Validators` slice directly. To verify it rejects invalid input: call the validator's method directly (e.g., `v.ValidateString(ctx, req, resp)`).

**Warning signs:** Test passes even with invalid value — validator appears to be no-op.

### Pitfall 2: Forgetting go:generate Directive for tfplugindocs

**What goes wrong:** `make generate` runs `go generate ./...` but produces nothing because no `//go:generate` comment exists.

**Why it happens:** `tools/tools.go` pins the binary as a dependency but `go generate` needs an explicit comment in a package file.

**How to avoid:** Add `//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-name flashblade` to `main.go`.

### Pitfall 3: Pagination Is NOT Auto-Applied

**What goes wrong:** Tests pass for small lists; in production with 100+ file systems, data sources silently return only the first page.

**Why it happens:** `ListFileSystems` passes `ContinuationToken` as a parameter but does not loop until `ContinuationToken` in the response is empty.

**How to avoid:** Before writing the pagination test, verify whether auto-pagination should live in the client layer or data source layer. The `ListResponse` model already has `ContinuationToken string` in `models.go`. Implement a `listAllFileSystems` loop in the client or data source, then test with a two-page mock response.

### Pitfall 4: tfplugindocs Fails on Missing Example Files

**What goes wrong:** `tfplugindocs` generates empty or broken docs for resources without a matching `examples/resources/<type>/resource.tf`.

**Why it happens:** The tool uses examples as usage snippets in generated docs. Missing files result in empty "Example Usage" sections or tool errors.

**How to avoid:** Create all `examples/resources/*/resource.tf` and `examples/data-sources/*/data-source.tf` files before running `make generate`. The file content can be minimal but must be valid HCL.

### Pitfall 5: 409 and 422 Not Mapped to Diagnostic in Provider

**What goes wrong:** A 409 Conflict from the API surfaces as an opaque error string, not a meaningful Terraform diagnostic.

**Why it happens:** The current provider code uses `resp.Diagnostics.AddError(...)` generically — it does not distinguish 409 from other errors.

**How to avoid:** In Create methods, detect `client.IsConflict(err)` (needs implementation — only `IsNotFound` exists) and add a targeted diagnostic. The test drives this: write the failing test first, then add `IsConflict` to `errors.go`.

---

## Code Examples

### Schema Plan Modifier Type Assertion

```go
// Source: internal/provider/object_store_access_key_resource_test.go lines 262-285
nameAttr, ok := schResp.Attributes["name"].(schema.StringAttribute)
if !ok {
    t.Fatal("name attribute not found or wrong type")
}
hasRR := false
for _, pm := range nameAttr.PlanModifiers {
    if _, ok := pm.(stringplanmodifier.RequiresReplaceModifier); ok {
        hasRR = true
    }
}
if !hasRR {
    t.Error("expected RequiresReplace plan modifier on name attribute")
}
```

### Direct Validator Test Pattern

```go
// New pattern — directly tests the validator function
import "github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"

func TestUnit_OAPRule_EffectValidator(t *testing.T) {
    s := oapRuleResourceSchema(t).Schema
    effectAttr, ok := s.Attributes["effect"].(schema.StringAttribute)
    if !ok {
        t.Fatal("effect attribute not found")
    }
    if len(effectAttr.Validators) == 0 {
        t.Fatal("effect attribute has no validators")
    }
    // Verify OneOf is registered
    var hasOneOf bool
    for _, v := range effectAttr.Validators {
        req := validator.StringRequest{ConfigValue: types.StringValue("invalid")}
        resp := &validator.StringResponse{}
        v.ValidateString(context.Background(), req, resp)
        if resp.Diagnostics.HasError() {
            hasOneOf = true
        }
    }
    if !hasOneOf {
        t.Error("expected validator to reject 'invalid' effect value")
    }
}
```

### WriteJSONError Usage in Error-Path Tests

```go
// Source: internal/testmock/handlers/helpers.go (WriteJSONError exists)
// Usage in test-specific handlers:
ms.RegisterHandler("/api/2.22/buckets", func(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodPost {
        handlers.WriteJSONError(w, http.StatusConflict, "bucket \"my-bucket\" already exists")
    }
})
```

### tfplugindocs go:generate Directive

```go
// Source: terraform-plugin-docs official docs — place in main.go
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-name flashblade
```

### GitHub Actions CI Workflow Pattern

```yaml
# .github/workflows/ci.yml
# Source: standard Go provider pattern
name: CI
on:
  push:
    branches: [main]
  pull_request:
    branches: [main]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
          cache: true
      - run: go test ./internal/... -count=1 -timeout 5m
      - run: golangci-lint run ./...
        if: success()
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Manual markdown docs | `terraform-plugin-docs` auto-generation | ~2021 (framework era) | Docs stay in sync with schema; no manual drift |
| SDK v2 resource testing with `resource.UnitTest` | Direct method testing with `tfsdk.State`/`tfsdk.Plan` | framework v1.x | No external binary needed; fast; already used in this codebase |
| Separate `_schema_test.go` files | Tests colocated in `*_resource_test.go` | Project convention | Already established in this codebase |

**Deprecated/outdated:**
- `terraform-plugin-sdk/v2/helper/resource.UnitTest`: Not applicable — this provider uses terraform-plugin-framework, not SDK v2.

---

## Open Questions

1. **Auto-pagination ownership: client layer vs data source layer?**
   - What we know: `ListFileSystems` accepts a `ContinuationToken` parameter but returns a single page. `ListResponse` has a `ContinuationToken` field in `models.go` but the response token is never read.
   - What's unclear: Should the client loop internally (transparent to callers) or should each data source implement a loop? Client-layer is simpler and DRY; data-source-layer allows callers to control page size for streaming use cases.
   - Recommendation: Implement auto-pagination in the client layer as a `listAllFileSystems`-style internal helper. This is consistent with the project pattern of keeping HTTP logic in the client package. **Must be decided before planning QUA-04 pagination tasks.**

2. **IsConflict / Is422 helper: add to errors.go?**
   - What we know: `errors.go` has `IsNotFound` (404) and `IsRetryable`. No `IsConflict` (409) or `IsUnprocessable` (422) helpers exist.
   - What's unclear: Whether provider resources currently handle 409 by falling through to a generic error, or whether they return a cryptic diagnostic.
   - Recommendation: Add `IsConflict(err error) bool` and `IsUnprocessable(err error) bool` to `errors.go` as part of the error-path test wave. These are 3-line functions following the `IsNotFound` pattern.

3. **Missing QTP-04/05/QTR-04 (Quota import + data source)**
   - What we know: REQUIREMENTS.md marks QTP-04, QTP-05, QTR-04 as Pending from Phase 4.
   - What's unclear: Are `quota_group` and `quota_user` import + data source tests missing, or was the implementation deferred?
   - Recommendation: Audit `quota_group_resource.go` and `quota_user_resource.go` for `ImportState` implementation and data source registration before assuming Phase 5 must add them.

---

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go testing stdlib (`testing` package) |
| Config file | None — no `pytest.ini` / `jest.config` equivalent; tests run with `go test` |
| Quick run command | `go test ./internal/... -count=1 -timeout 5m` |
| Full suite command | `go test ./internal/... -count=1 -race -timeout 10m` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| QUA-01 | Plan modifiers correct on all resources | unit | `go test ./internal/provider/... -run TestUnit.*PlanModifier -count=1` | ❌ Wave 0 — add to existing `*_resource_test.go` |
| QUA-02 | Invalid enum/value rejected at plan time | unit | `go test ./internal/provider/... -run TestUnit.*Validator -count=1` | ❌ Wave 0 — add to existing `*_resource_test.go` |
| QUA-03 | All schema defs, validators, plan modifiers tested | unit | `go test ./internal/provider/... -count=1` | Partial — schema helpers exist; assertions missing |
| QUA-04 | CRUD lifecycle for all families + error paths | unit | `go test ./internal/... -count=1` | Partial — CRUD exists; 409/422/pagination missing |
| QUA-05 | Retry backoff on 429/503 | unit | `go test ./internal/client/... -run TestUnit_RetryTransport -count=1` | ✅ `internal/client/transport_test.go` |
| QUA-06 | Docs generated, examples complete, CI workflow | manual + generate | `make generate && go test ./internal/... -count=1` | ❌ Wave 0 — examples/, docs/, .github/workflows/, README.md |

### Sampling Rate

- **Per task commit:** `go test ./internal/... -count=1 -timeout 5m`
- **Per wave merge:** `go test ./internal/... -count=1 -race -timeout 10m`
- **Phase gate:** Full suite green + `make generate` produces non-empty `docs/` + `.github/workflows/ci.yml` present before `/gsd:verify-work`

### Wave 0 Gaps

- [ ] Plan modifier test functions — add to each `*_resource_test.go` where `RequiresReplace` or `UseStateForUnknown` exists in schema
- [ ] Validator test functions — add to `object_store_access_policy_rule_resource_test.go` (direct validator call); add validators to other resources where enum/range validation needed
- [ ] `IsConflict` and `IsUnprocessable` helpers — add to `internal/client/errors.go`
- [ ] Auto-pagination — add to relevant `internal/client/*.go` list functions
- [ ] `examples/resources/*/resource.tf` — 18 new files (one per resource type)
- [ ] `examples/data-sources/*/data-source.tf` — 13 new files (one per data source type, excl. file_system which exists)
- [ ] `//go:generate` directive — add to `main.go`
- [ ] `.github/workflows/ci.yml` — new file
- [ ] `README.md` — new file at repo root

*(No new test infrastructure needed — existing `testmock` package covers all resource families)*

---

## Sources

### Primary (HIGH confidence)
- Direct codebase inspection — all findings from `grep`, `ls`, and `Read` tool calls against the actual repository
- `internal/client/transport.go` — confirmed retry implementation (exponential backoff, 429/503)
- `internal/client/transport_test.go` — confirmed 3 retry tests exist and pass
- `go.mod` — confirmed `terraform-plugin-docs v0.24.0` present
- `tools/tools.go` — confirmed `tfplugindocs` binary pinned

### Secondary (MEDIUM confidence)
- terraform-plugin-docs convention (examples/ structure) — based on `tools/tools.go` import path and standard provider scaffold pattern; consistent with v0.24.0 behavior

### Tertiary (LOW confidence)
- GitHub Actions workflow snippet — based on standard Go community patterns; exact syntax should be verified against current `actions/setup-go@v5` docs before finalizing

---

## Metadata

**Confidence breakdown:**
- Current state audit (what exists): HIGH — direct file inspection
- Standard stack: HIGH — already in go.mod, no version uncertainty
- Architecture patterns: HIGH — derived from existing test code in this repo
- Pitfalls: HIGH — derived from direct code analysis (pagination gap confirmed, validator bypass confirmed)
- CI workflow: MEDIUM — standard pattern but exact action versions not verified

**Research date:** 2026-03-28
**Valid until:** 2026-04-28 (stable Go testing patterns; terraform-plugin-docs API stable at v0.24.0)
