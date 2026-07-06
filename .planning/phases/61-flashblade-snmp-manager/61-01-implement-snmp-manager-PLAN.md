---
phase: 61-flashblade-snmp-manager
plan: 01
slug: implement-snmp-manager
type: execute
wave: 1
depends_on: []
autonomous: true
requirements:
  - SNMP-01
  - SNMP-02
  - SNMP-03
  - SNMP-04
  - SNMP-05
  - SNMP-06
  - SNMP-07
  - SNMP-08
  - SNMP-09
  - SNMP-10
  - SNMP-11
  - SNMP-12
  - SNMP-13
files_modified:
  - internal/client/models_admin.go
  - internal/client/snmp_managers.go
  - internal/client/snmp_managers_test.go
  - internal/testmock/handlers/snmp_managers.go
  - internal/testmock/server.go
  - internal/provider/snmp_manager_resource.go
  - internal/provider/snmp_manager_resource_test.go
  - internal/provider/snmp_manager_data_source.go
  - internal/provider/snmp_manager_data_source_test.go
  - internal/provider/provider.go
  - examples/resources/flashblade_snmp_manager/resource.tf
  - examples/resources/flashblade_snmp_manager/import.sh
  - examples/data-sources/flashblade_snmp_manager/data-source.tf
  - docs/resources/snmp_manager.md
  - docs/data-sources/snmp_manager.md
  - ROADMAP.md
must_haves:
  truths:
    - "Operator can `terraform apply` a `flashblade_snmp_manager` v3 config and it is created on the array."
    - "Operator can `terraform apply` a `flashblade_snmp_manager` v2c config and it is created on the array."
    - "Operator can mutate `host`, `notification`, `v3.user`, `v3.auth_protocol`, `v3.privacy_protocol` via `terraform apply` and the PATCH body carries only the changed fields."
    - "Operator can `terraform destroy` and the resource disappears from the array."
    - "Operator can `terraform import flashblade_snmp_manager.<name> <name>` and the next plan is clean except for the three sensitive fields, which are null."
    - "Operator can `terraform plan` against unchanged state and see no diff (sensitive fields stay in state, never re-fetched)."
    - "Operator can read a single manager via `data \"flashblade_snmp_manager\"` by name; not-found surfaces a clear error."
    - "Drift on any of 6 leaves (`host`, `notification`, `version`, `v3.user`, `v3.auth_protocol`, `v3.privacy_protocol`) is logged via `tflog.Debug` with key `\"drift detected\"`; sensitive fields are NEVER logged."
    - "`make build && make test && make lint && make docs` are all clean; total test count >= 816."
  artifacts:
    - path: "internal/client/models_admin.go"
      provides: "SnmpManager, SnmpV2c, SnmpV3, SnmpV3Post, SnmpManagerPost, SnmpManagerPatch structs"
      contains: "type SnmpManager struct"
    - path: "internal/client/snmp_managers.go"
      provides: "Get/List/Post/Patch/Delete CRUD via getOneByName[SnmpManager]"
      exports: ["GetSnmpManager", "ListSnmpManagers", "PostSnmpManager", "PatchSnmpManager", "DeleteSnmpManager"]
    - path: "internal/client/snmp_managers_test.go"
      provides: "5 TestUnit_SnmpManager_* client tests"
      contains: "TestUnit_SnmpManager_Get_Found"
    - path: "internal/testmock/handlers/snmp_managers.go"
      provides: "snmpManagerStore + RegisterSnmpManagerHandlers; GET no-match => 200 + empty list"
      exports: ["RegisterSnmpManagerHandlers"]
    - path: "internal/provider/snmp_manager_resource.go"
      provides: "Resource with 4 interfaces (Resource, WithConfigure, WithImportState, WithUpgradeState), v2c/v3 SingleNestedAttribute, 6 per-leaf drift logs, sensitive write-once Read mapping"
      exports: ["NewSnmpManagerResource"]
    - path: "internal/provider/snmp_manager_resource_test.go"
      provides: "3 TestUnit_SnmpManagerResource_* tests (Lifecycle, Import, DriftDetection)"
      contains: "TestUnit_SnmpManagerResource_Lifecycle"
    - path: "internal/provider/snmp_manager_data_source.go"
      provides: "Data source with DataSource + DataSourceWithConfigure"
      exports: ["NewSnmpManagerDataSource"]
    - path: "internal/provider/snmp_manager_data_source_test.go"
      provides: "1 TestUnit_SnmpManagerDataSource_Basic test"
      contains: "TestUnit_SnmpManagerDataSource_Basic"
    - path: "examples/resources/flashblade_snmp_manager/resource.tf"
      provides: "v3 example + commented v2c snippet + in-place version switch note"
    - path: "examples/resources/flashblade_snmp_manager/import.sh"
      provides: "terraform import by name (NOT UUID)"
    - path: "examples/data-sources/flashblade_snmp_manager/data-source.tf"
      provides: "data source HCL example"
    - path: "docs/resources/snmp_manager.md"
      provides: "tfplugindocs-generated resource page"
    - path: "docs/data-sources/snmp_manager.md"
      provides: "tfplugindocs-generated data source page"
    - path: "ROADMAP.md"
      provides: "SNMP Managers row in Array Administration / Implemented with Done + v2.23.1 note"
      contains: "SNMP Managers"
  key_links:
    - from: "internal/provider/snmp_manager_resource.go"
      to: "internal/client/snmp_managers.go"
      via: "GetSnmpManager / PostSnmpManager / PatchSnmpManager / DeleteSnmpManager"
      pattern: "GetSnmpManager\\("
    - from: "internal/provider/snmp_manager_resource.go"
      to: "drift logging contract"
      via: "tflog.Debug per leaf"
      pattern: "drift detected"
    - from: "internal/provider/snmp_manager_resource.go"
      to: "write-once skip in mapping function"
      via: "Read mapping never assigns community / auth_passphrase / privacy_passphrase"
      pattern: "// skip sensitive write-once"
    - from: "internal/testmock/handlers/snmp_managers.go"
      to: "real-API GET behaviour parity"
      via: "GET ?names= no match returns 200 + empty list"
      pattern: "WriteJSONListResponse.*\\[\\]"
    - from: "internal/provider/provider.go"
      to: "resource & data source registration"
      via: "NewSnmpManagerResource / NewSnmpManagerDataSource"
      pattern: "NewSnmpManagerResource"
    - from: "ROADMAP.md"
      to: "implementation commit"
      via: "row move in same commit as code"
      pattern: "SNMP Managers.*Done"
---

<objective>
Deliver the `flashblade_snmp_manager` Terraform **resource** + **data source** against `/api/2.23/snmp-managers` with full CRUD, both `v2c` and `v3` nested config blocks, write-once sensitive fields, per-leaf drift detection, mock handler, >= 9 new `TestUnit_` tests, HCL examples, regenerated docs, and the repo-level `ROADMAP.md` row move — all in the strict order of the *New Resource* 16-item checklist in `CONVENTIONS.md`, driven by the `flashblade-resource-builder` skill, with **zero deviation**.

Purpose: ship v2.23.1 with one cohesive, atomic change. SNMP trap destinations are operationally critical; the implementation must mirror the conventions already proven across 55 resources.

Output:
- 1 new resource (`flashblade_snmp_manager`) + 1 new data source.
- 9+ new `TestUnit_` unit tests; total `make test` count >= 816.
- Updated `provider.go`, generated docs, examples, and root `ROADMAP.md`.
- 1 logical implementation commit (and optional intermediate commits, all `--no-verify`).
</objective>

<execution_context>
@$HOME/.claude/get-shit-done/workflows/execute-plan.md
@$HOME/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/STATE.md
@.planning/ROADMAP.md
@.planning/REQUIREMENTS.md
@.planning/phases/61-flashblade-snmp-manager/61-CONTEXT.md
@CONVENTIONS.md
@CLAUDE.md
@.claude/skills/flashblade-resource-builder/SKILL.md
@api_references/2.23.md
@swagger-2.23.json
@ROADMAP.md
@GNUmakefile

<orchestrating_skill>
**MANDATORY:** This phase is driven by the `flashblade-resource-builder` skill at `.claude/skills/flashblade-resource-builder/`. Before starting, the executor MUST:

1. Read `.claude/skills/flashblade-resource-builder/SKILL.md` (lightweight index).
2. Load specific `rules/*.md` on demand (model structs, client CRUD, mocks, resource, data source, tests, docs).
3. Use the skill's lifecycle order as the spine: **models -> client -> mocks -> client tests -> resource -> resource tests -> data source -> ds tests -> registration -> examples -> make docs -> ROADMAP.md**.

The tasks below mirror this lifecycle one-to-one against the 16-item *New Resource* checklist in `CONVENTIONS.md`.
</orchestrating_skill>

<navigation_rule>
**Code navigation in `.go` files MUST use Serena MCP** (`mcp__serena__find_symbol`, `mcp__serena__get_symbols_overview`, `mcp__serena__find_referencing_symbols`).

`Read` is allowed ONLY when the executor is about to edit the file content. `Grep`/`Glob` on `.go` files are BLOCKED by a project hook. This is non-negotiable per `CLAUDE.md`.

Before starting any task, run `mcp__serena__initial_instructions` (Serena instruction manual). At each task's `<read_first>` step that names a `.go` symbol, use `find_symbol` (not `Read`).
</navigation_rule>

<commit_policy>
**MANDATORY:** Every commit in this phase uses:

```bash
git commit --no-verify -m "feat(snmp): <subject>"
```

- **No** `Co-Authored-By` trailer (project rule).
- Conventional Commits prefixes: `feat:`, `test:`, `docs:`, `chore:`.
- The final implementation commit MUST bundle:
  - All code changes (`internal/client/`, `internal/testmock/`, `internal/provider/`, `examples/`, `docs/`)
  - `ROADMAP.md` row move (D-18, SNMP-11)
  - Generated `docs/resources/snmp_manager.md` + `docs/data-sources/snmp_manager.md`

Per D-19 + project `CLAUDE.md`.
</commit_policy>

<branching>
Branch `implem-snmp-managers` MUST be created from a clean `main` at the START of T01 (D-20):

```bash
git checkout main
git pull --ff-only
git status   # must be clean
git checkout -b implem-snmp-managers
```

All subsequent commits land on this branch. No worktrees (project `CLAUDE.md` `Do NOT`).
</branching>

<api_contracts>
Authoritative schemas from `swagger-2.23.json` (validated 2026-05-20):

```text
_snmp_v2c:
  community: string, maxLength=32

_snmp_v3:                   # used in GET (response) and PATCH (request)
  user: string
  auth_protocol: string  (MD5|SHA)
  auth_passphrase: string         # NEVER returned on GET
  privacy_protocol: string  (AES|DES)
  privacy_passphrase: string      # NEVER returned on GET

_snmp_v3_post:              # stricter constraints on POST
  user: string
  auth_protocol: string  (MD5|SHA)
  auth_passphrase: string  maxLength=32           # NEVER returned on GET
  privacy_protocol: string  (AES|DES)
  privacy_passphrase: string  minLength=8 maxLength=63   # NEVER returned on GET

SnmpManager (GET response):
  id: string (ro)
  name: string (ro, supplied via ?names= on POST)
  host: string
  notification: string  (inform|trap)
  version: string  (v2c|v3)
  v2c: _snmp_v2c (community omitted)
  v3:  _snmp_v3  (passphrases omitted)

SnmpManagerPost (POST body, name via ?names=):
  host, notification, version, v2c, v3 (using _snmp_v3_post constraints)

SnmpManagerPatch (PATCH body):
  every field optional (pointer); nested v2c/v3 are *atomic* blocks (same pattern as ArrayConnectionPatch.Throttle)
```

Endpoints:
- `GET    /api/2.23/snmp-managers`           (filter by `?names=`)
- `POST   /api/2.23/snmp-managers?names=NAME`
- `PATCH  /api/2.23/snmp-managers?names=NAME`
- `DELETE /api/2.23/snmp-managers?names=NAME`

OUT of scope: `GET /api/2.23/snmp-managers/test` (SNMP-13, D-X). No code references it.
</api_contracts>

<interfaces>
Reference patterns the executor will lean on (extracted from existing code, do not re-discover):

```go
// internal/client/models_admin.go:104-146 — ArrayConnection / ArrayConnectionThrottle / ArrayConnectionPatch
//   Canonical template for the SnmpManager + SnmpV2c + SnmpV3 + SnmpManagerPatch shape.
//   In particular ArrayConnectionPatch.Throttle *ArrayConnectionThrottle is the atomic-nested-block pattern
//   that SnmpManagerPatch.V2c *SnmpV2c and SnmpManagerPatch.V3 *SnmpV3 MUST copy.

// internal/client/client.go — getOneByName[T] generic
//   func getOneByName[T any](ctx context.Context, c *FlashBladeClient, path, name string) (*T, error)
//   ALL Get<X> implementations in this codebase use this. Never hand-roll list-then-filter.

// internal/testmock/handlers/targets.go — RegisterTargetHandlers + targetStore
//   Canonical mock handler: mutex+byName+nextID, Seed(...), shared helpers, empty-list GET=200.

// internal/provider/array_connection_resource.go:121-141 — Throttle SingleNestedAttribute
//   schema.SingleNestedAttribute{ Optional: true, Computed: true, Attributes: map[string]schema.Attribute{...} }
//   Template for both `v2c` and `v3` attributes.

// internal/provider/directory_service_management_resource.go:467-517 — mapDirectoryServiceToModel
//   Skips BindPassword in Read mapping (API never returns it). Template for skipping community / auth_passphrase / privacy_passphrase.

// internal/provider/target_resource.go — base lifecycle template (Configure, Schema, Create, Read, Update, Delete, ImportState, UpgradeState)
//   4 mandatory interfaces; SchemaVersion = 0; nullTimeoutsValue() on Import.

// internal/provider/target_data_source.go — DS template (2 interfaces, no timeouts)
```
</interfaces>
</context>

<tasks>

<task id="T01" type="auto">
  <name>T01: Create branch, generate model structs in models_admin.go</name>
  <files>internal/client/models_admin.go</files>
  <read_first>
    - .planning/phases/61-flashblade-snmp-manager/61-CONTEXT.md (D-02, D-03, D-20)
    - CONVENTIONS.md §Model Structs
    - swagger-2.23.json (already extracted in `<api_contracts>` above; do not re-read)
    - mcp__serena__get_symbols_overview file=internal/client/models_admin.go (locate insertion point after AlertWatcher block)
    - mcp__serena__find_symbol name=ArrayConnectionThrottle relative_path=internal/client/models_admin.go include_body=true
    - mcp__serena__find_symbol name=ArrayConnectionPatch relative_path=internal/client/models_admin.go include_body=true
  </read_first>
  <action>
    **Step 0 — Branch (D-20):**
    ```bash
    git checkout main && git pull --ff-only && git status   # must be clean
    git checkout -b implem-snmp-managers
    ```

    **Step 1 — Append to `internal/client/models_admin.go` (after the AlertWatcher block; do NOT touch existing structs):**

    Add the following 6 types (exact JSON tags + pointer rules per CONVENTIONS.md §Model Structs):

    ```go
    // SnmpV2c holds the v2c configuration of an SNMP manager.
    // Returned on GET (community omitted by the API).
    // Sent atomically on POST and PATCH.
    type SnmpV2c struct {
        Community string `json:"community,omitempty"`
    }

    // SnmpV3 holds the v3 configuration of an SNMP manager on GET and PATCH.
    // Passphrases are never returned on GET.
    type SnmpV3 struct {
        User             string `json:"user,omitempty"`
        AuthProtocol     string `json:"auth_protocol,omitempty"`
        AuthPassphrase   string `json:"auth_passphrase,omitempty"`
        PrivacyProtocol  string `json:"privacy_protocol,omitempty"`
        PrivacyPassphrase string `json:"privacy_passphrase,omitempty"`
    }

    // SnmpV3Post mirrors SnmpV3 but encodes the stricter POST-time constraints
    // (auth_passphrase <= 32, privacy_passphrase 8..63). Used ONLY in SnmpManagerPost.
    type SnmpV3Post struct {
        User             string `json:"user,omitempty"`
        AuthProtocol     string `json:"auth_protocol,omitempty"`
        AuthPassphrase   string `json:"auth_passphrase,omitempty"`
        PrivacyProtocol  string `json:"privacy_protocol,omitempty"`
        PrivacyPassphrase string `json:"privacy_passphrase,omitempty"`
    }

    // SnmpManager represents the GET /snmp-managers response.
    type SnmpManager struct {
        ID           string    `json:"id"`
        Name         string    `json:"name"`
        Host         string    `json:"host,omitempty"`
        Notification string    `json:"notification,omitempty"`
        Version      string    `json:"version,omitempty"`
        V2c          *SnmpV2c  `json:"v2c,omitempty"`
        V3           *SnmpV3   `json:"v3,omitempty"`
    }

    // SnmpManagerPost is the POST body. Name is supplied via ?names= and excluded.
    // V3 uses the stricter SnmpV3Post constraint set (per D-04).
    type SnmpManagerPost struct {
        Host         string      `json:"host,omitempty"`
        Notification string      `json:"notification,omitempty"`
        Version      string      `json:"version,omitempty"`
        V2c          *SnmpV2c    `json:"v2c,omitempty"`
        V3           *SnmpV3Post `json:"v3,omitempty"`
    }

    // SnmpManagerPatch is the PATCH body. Every field is a pointer.
    // V2c/V3 are atomic nested blocks (template: ArrayConnectionPatch.Throttle).
    type SnmpManagerPatch struct {
        Host         *string  `json:"host,omitempty"`
        Notification *string  `json:"notification,omitempty"`
        Version      *string  `json:"version,omitempty"`
        V2c          *SnmpV2c `json:"v2c,omitempty"`
        V3           *SnmpV3  `json:"v3,omitempty"`
    }
    ```

    **Why these exact shapes:** GET response uses plain types per CONVENTIONS.md (no scalar pointers); POST uses `SnmpV3Post` for stricter validators per D-04; PATCH uses atomic nested blocks (D-03), mirroring `ArrayConnectionPatch.Throttle`. Sensitive fields are inside the nested structs so the Patch-omitting-the-block pattern works cleanly for in-place v2c<->v3 switch (D-06).

    **Step 2 — Optional intermediate commit:**
    ```bash
    git add internal/client/models_admin.go
    git commit --no-verify -m "feat(snmp): add SnmpManager client model structs"
    ```
  </action>
  <verify>
    <automated>cd /home/gule/Workspace/team-infrastructure/terraform-provider-mica &amp;&amp; go build ./internal/client/... &amp;&amp; rg -n "^type SnmpManager " internal/client/models_admin.go &amp;&amp; rg -n "^type SnmpManagerPost " internal/client/models_admin.go &amp;&amp; rg -n "^type SnmpManagerPatch " internal/client/models_admin.go &amp;&amp; rg -n "^type SnmpV2c " internal/client/models_admin.go &amp;&amp; rg -n "^type SnmpV3 " internal/client/models_admin.go &amp;&amp; rg -n "^type SnmpV3Post " internal/client/models_admin.go</automated>
  </verify>
  <acceptance_criteria>
    - Branch `implem-snmp-managers` exists and is checked out (`git rev-parse --abbrev-ref HEAD` returns `implem-snmp-managers`).
    - All 6 struct declarations exist in `internal/client/models_admin.go`: `SnmpManager`, `SnmpManagerPost`, `SnmpManagerPatch`, `SnmpV2c`, `SnmpV3`, `SnmpV3Post`.
    - `SnmpManagerPost.V3` field type is `*SnmpV3Post` (NOT `*SnmpV3`).
    - `SnmpManagerPatch` has every field as a pointer (`*string`, `*SnmpV2c`, `*SnmpV3`).
    - `go build ./internal/client/...` exits 0.
    - No edits to any existing struct (validated by `git diff main -- internal/client/models_admin.go` showing only additions).
  </acceptance_criteria>
  <done>
    Models compile, structs match CONVENTIONS.md §Model Structs rules (pointer policy, JSON tags, name excluded), atomic nested block pattern mirrors `ArrayConnectionPatch.Throttle`.
  </done>
</task>

<task id="T02" type="auto">
  <name>T02: Implement client CRUD using getOneByName[T]</name>
  <files>internal/client/snmp_managers.go</files>
  <read_first>
    - CONVENTIONS.md §Client CRUD Methods
    - mcp__serena__find_symbol name=getOneByName relative_path=internal/client/client.go include_body=true
    - mcp__serena__find_symbol name=GetTarget relative_path=internal/client/targets.go include_body=true
    - mcp__serena__find_symbol name=PostTarget relative_path=internal/client/targets.go include_body=true
    - mcp__serena__find_symbol name=PatchTarget relative_path=internal/client/targets.go include_body=true
    - mcp__serena__find_symbol name=DeleteTarget relative_path=internal/client/targets.go include_body=true
    - mcp__serena__find_symbol name=ListTargets relative_path=internal/client/targets.go include_body=true
  </read_first>
  <action>
    Create `internal/client/snmp_managers.go` with this exact layout (mirrors `targets.go`):

    ```go
    package client

    import (
        "context"
        "fmt"
        "net/http"
        "net/url"
    )

    const snmpManagersPath = "/snmp-managers"

    // GetSnmpManager fetches a single SNMP manager by name.
    // Empty list (no match) is converted to a not-found error by getOneByName.
    func (c *FlashBladeClient) GetSnmpManager(ctx context.Context, name string) (*SnmpManager, error) {
        return getOneByName[SnmpManager](ctx, c, snmpManagersPath, name)
    }

    // ListSnmpManagers returns all SNMP managers. No filters in API 2.23 beyond ?names=.
    func (c *FlashBladeClient) ListSnmpManagers(ctx context.Context) ([]SnmpManager, error) {
        var resp struct{ Items []SnmpManager `json:"items"` }
        if err := c.do(ctx, http.MethodGet, snmpManagersPath, nil, nil, &resp); err != nil {
            return nil, err
        }
        return resp.Items, nil
    }

    // PostSnmpManager creates an SNMP manager. Name is carried in ?names=.
    func (c *FlashBladeClient) PostSnmpManager(ctx context.Context, name string, body SnmpManagerPost) (*SnmpManager, error) {
        q := url.Values{"names": []string{name}}
        var resp struct{ Items []SnmpManager `json:"items"` }
        if err := c.do(ctx, http.MethodPost, snmpManagersPath, q, body, &resp); err != nil {
            return nil, err
        }
        if len(resp.Items) == 0 {
            return nil, fmt.Errorf("snmp_manager %q: empty POST response", name)
        }
        return &resp.Items[0], nil
    }

    // PatchSnmpManager updates an SNMP manager by name.
    func (c *FlashBladeClient) PatchSnmpManager(ctx context.Context, name string, body SnmpManagerPatch) (*SnmpManager, error) {
        q := url.Values{"names": []string{name}}
        var resp struct{ Items []SnmpManager `json:"items"` }
        if err := c.do(ctx, http.MethodPatch, snmpManagersPath, q, body, &resp); err != nil {
            return nil, err
        }
        if len(resp.Items) == 0 {
            return nil, fmt.Errorf("snmp_manager %q: empty PATCH response", name)
        }
        return &resp.Items[0], nil
    }

    // DeleteSnmpManager deletes an SNMP manager by name.
    func (c *FlashBladeClient) DeleteSnmpManager(ctx context.Context, name string) error {
        q := url.Values{"names": []string{name}}
        return c.do(ctx, http.MethodDelete, snmpManagersPath, q, nil, nil)
    }
    ```

    **Critical rules (CONVENTIONS.md §Client CRUD Methods):**
    - Path does NOT include API version (`/snmp-managers`, not `/api/2.23/snmp-managers`). `c.do()` adds the version prefix.
    - GET-single uses `getOneByName[SnmpManager]` (never hand-roll).
    - Errors from `c.do` are returned directly (no `fmt.Errorf` wrap; preserves `APIError`).
    - `?names=` value goes through `url.Values` (which encodes properly).
    - List shape: "No args beyond ctx" (global flat set) per CONVENTIONS.md table.

    If the exact `url.Values{"names": ...}` form differs from what `targets.go` uses, MATCH `targets.go` verbatim. Verify by inspecting `PostTarget` body returned by Serena.
  </action>
  <verify>
    <automated>cd /home/gule/Workspace/team-infrastructure/terraform-provider-mica &amp;&amp; go build ./internal/client/... &amp;&amp; rg -n "func \(c \*FlashBladeClient\) GetSnmpManager\(" internal/client/snmp_managers.go &amp;&amp; rg -n "func \(c \*FlashBladeClient\) ListSnmpManagers\(" internal/client/snmp_managers.go &amp;&amp; rg -n "func \(c \*FlashBladeClient\) PostSnmpManager\(" internal/client/snmp_managers.go &amp;&amp; rg -n "func \(c \*FlashBladeClient\) PatchSnmpManager\(" internal/client/snmp_managers.go &amp;&amp; rg -n "func \(c \*FlashBladeClient\) DeleteSnmpManager\(" internal/client/snmp_managers.go &amp;&amp; rg -n "getOneByName\[SnmpManager\]" internal/client/snmp_managers.go</automated>
  </verify>
  <acceptance_criteria>
    - File exists at `internal/client/snmp_managers.go`.
    - All 5 methods present and exported: `GetSnmpManager`, `ListSnmpManagers`, `PostSnmpManager`, `PatchSnmpManager`, `DeleteSnmpManager`.
    - `GetSnmpManager` body uses `getOneByName[SnmpManager]` (grep shows the literal call).
    - Path constant `snmpManagersPath = "/snmp-managers"` — no `/api/2.23` prefix.
    - `go build ./internal/client/...` exits 0.
    - No use of `fmt.Errorf("...%w", err)` to wrap `c.do` errors.
  </acceptance_criteria>
  <done>
    CRUD layer compiles; signatures, return types, and `getOneByName` use match `targets.go`.
  </done>
</task>

<task id="T03" type="auto">
  <name>T03: Implement mock handler with Seed, empty-list GET=200, shared helpers</name>
  <files>internal/testmock/handlers/snmp_managers.go, internal/testmock/server.go</files>
  <read_first>
    - CONVENTIONS.md §Mock Handlers (the entire section including the table)
    - .planning/phases/61-flashblade-snmp-manager/61-CONTEXT.md (D-11, D-12)
    - mcp__serena__get_symbols_overview file=internal/testmock/handlers/targets.go
    - mcp__serena__find_symbol name=RegisterTargetHandlers relative_path=internal/testmock/handlers/targets.go include_body=true
    - mcp__serena__find_symbol name=targetStore relative_path=internal/testmock/handlers/targets.go include_body=true
    - mcp__serena__get_symbols_overview file=internal/testmock/handlers/helpers.go
    - mcp__serena__get_symbols_overview file=internal/testmock/server.go
  </read_first>
  <action>
    **Step 1 — Create `internal/testmock/handlers/snmp_managers.go`** modeled exactly on `targets.go`:

    ```go
    package handlers

    import (
        "encoding/json"
        "net/http"
        "strconv"
        "sync"

        "github.com/numberly/terraform-provider-mica/internal/client"
    )

    type snmpManagerStore struct {
        mu     sync.Mutex
        byName map[string]*client.SnmpManager
        nextID int
    }

    // Seed inserts pre-existing managers into the store (sensitive fields are stripped from GET responses).
    func (s *snmpManagerStore) Seed(items ...*client.SnmpManager) {
        s.mu.Lock()
        defer s.mu.Unlock()
        for _, it := range items {
            s.nextID++
            if it.ID == "" {
                it.ID = "snmpmgr-" + strconv.Itoa(s.nextID)
            }
            s.byName[it.Name] = it
        }
    }

    func RegisterSnmpManagerHandlers(mux *http.ServeMux) *snmpManagerStore {
        store := &snmpManagerStore{byName: map[string]*client.SnmpManager{}}

        mux.HandleFunc("/api/2.23/snmp-managers", func(w http.ResponseWriter, r *http.Request) {
            switch r.Method {
            case http.MethodGet:
                handleGetSnmpManagers(store, w, r)
            case http.MethodPost:
                handlePostSnmpManager(store, w, r)
            case http.MethodPatch:
                handlePatchSnmpManager(store, w, r)
            case http.MethodDelete:
                handleDeleteSnmpManager(store, w, r)
            default:
                WriteJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
            }
        })

        return store
    }

    // GET: ?names=<name> -> single match or empty list with HTTP 200 (NEVER 404 on no match).
    // No ?names= -> return all.
    func handleGetSnmpManagers(store *snmpManagerStore, w http.ResponseWriter, r *http.Request) {
        if err := ValidateQueryParams(r, "names"); err != nil {
            WriteJSONError(w, http.StatusBadRequest, err.Error())
            return
        }
        store.mu.Lock()
        defer store.mu.Unlock()

        names := r.URL.Query()["names"]
        var items []*client.SnmpManager
        if len(names) == 0 {
            for _, it := range store.byName {
                items = append(items, stripSensitive(it))
            }
        } else {
            for _, n := range names {
                if it, ok := store.byName[n]; ok {
                    items = append(items, stripSensitive(it))
                }
            }
        }
        WriteJSONListResponse(w, items)  // empty -> 200 + {"items": []}, per CONVENTIONS.md
    }

    func handlePostSnmpManager(store *snmpManagerStore, w http.ResponseWriter, r *http.Request) {
        name, err := RequireQueryParam(r, "names")
        if err != nil {
            WriteJSONError(w, http.StatusBadRequest, err.Error())
            return
        }
        var body client.SnmpManagerPost
        if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
            WriteJSONError(w, http.StatusBadRequest, "invalid body")
            return
        }

        store.mu.Lock()
        defer store.mu.Unlock()

        if _, exists := store.byName[name]; exists {
            WriteJSONError(w, http.StatusConflict, "snmp manager already exists")
            return
        }
        store.nextID++
        mgr := &client.SnmpManager{
            ID:           "snmpmgr-" + strconv.Itoa(store.nextID),
            Name:         name,
            Host:         body.Host,
            Notification: body.Notification,
            Version:      body.Version,
        }
        if body.V2c != nil {
            mgr.V2c = &client.SnmpV2c{Community: body.V2c.Community}
        }
        if body.V3 != nil {
            mgr.V3 = &client.SnmpV3{
                User:              body.V3.User,
                AuthProtocol:      body.V3.AuthProtocol,
                AuthPassphrase:    body.V3.AuthPassphrase,
                PrivacyProtocol:   body.V3.PrivacyProtocol,
                PrivacyPassphrase: body.V3.PrivacyPassphrase,
            }
        }
        store.byName[name] = mgr
        WriteJSONListResponse(w, []*client.SnmpManager{stripSensitive(mgr)})
    }

    func handlePatchSnmpManager(store *snmpManagerStore, w http.ResponseWriter, r *http.Request) {
        name, err := RequireQueryParam(r, "names")
        if err != nil {
            WriteJSONError(w, http.StatusBadRequest, err.Error())
            return
        }
        var body client.SnmpManagerPatch
        if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
            WriteJSONError(w, http.StatusBadRequest, "invalid body")
            return
        }

        store.mu.Lock()
        defer store.mu.Unlock()

        mgr, ok := store.byName[name]
        if !ok {
            WriteJSONError(w, http.StatusNotFound, "snmp manager not found")
            return
        }
        if body.Host != nil {
            mgr.Host = *body.Host
        }
        if body.Notification != nil {
            mgr.Notification = *body.Notification
        }
        if body.Version != nil {
            mgr.Version = *body.Version
        }
        if body.V2c != nil {
            mgr.V2c = body.V2c
        }
        if body.V3 != nil {
            mgr.V3 = body.V3
        }
        WriteJSONListResponse(w, []*client.SnmpManager{stripSensitive(mgr)})
    }

    func handleDeleteSnmpManager(store *snmpManagerStore, w http.ResponseWriter, r *http.Request) {
        name, err := RequireQueryParam(r, "names")
        if err != nil {
            WriteJSONError(w, http.StatusBadRequest, err.Error())
            return
        }
        store.mu.Lock()
        defer store.mu.Unlock()
        if _, ok := store.byName[name]; !ok {
            WriteJSONError(w, http.StatusNotFound, "snmp manager not found")
            return
        }
        delete(store.byName, name)
        w.WriteHeader(http.StatusOK)
    }

    // stripSensitive returns a shallow copy with community / auth_passphrase / privacy_passphrase blanked,
    // mirroring real API GET behaviour (D-12).
    func stripSensitive(in *client.SnmpManager) *client.SnmpManager {
        out := *in
        if in.V2c != nil {
            v := *in.V2c
            v.Community = ""
            out.V2c = &v
        }
        if in.V3 != nil {
            v := *in.V3
            v.AuthPassphrase = ""
            v.PrivacyPassphrase = ""
            out.V3 = &v
        }
        return &out
    }
    ```

    **Adjust helper signatures** (`ValidateQueryParams`, `RequireQueryParam`, `WriteJSONListResponse`, `WriteJSONError`) to whatever `helpers.go` actually exports. Match `targets.go` calls 1:1 to be safe.

    **Step 2 — Wire into `internal/testmock/server.go`:**

    Inspect the existing wiring (find where `RegisterTargetHandlers` is called) and append a call to `handlers.RegisterSnmpManagerHandlers(mux)`. Expose the returned store via the test bootstrap so provider tests can `Seed(...)`.
  </action>
  <verify>
    <automated>cd /home/gule/Workspace/team-infrastructure/terraform-provider-mica &amp;&amp; go build ./internal/testmock/... &amp;&amp; rg -n "func RegisterSnmpManagerHandlers\(" internal/testmock/handlers/snmp_managers.go &amp;&amp; rg -n "WriteJSONListResponse" internal/testmock/handlers/snmp_managers.go &amp;&amp; rg -n "RegisterSnmpManagerHandlers" internal/testmock/server.go &amp;&amp; rg -n "stripSensitive" internal/testmock/handlers/snmp_managers.go</automated>
  </verify>
  <acceptance_criteria>
    - `internal/testmock/handlers/snmp_managers.go` exists.
    - `RegisterSnmpManagerHandlers(mux *http.ServeMux) *snmpManagerStore` returns the store (so `Seed` is callable from tests).
    - GET handler uses `WriteJSONListResponse` for the no-match path (NOT `http.Error(... 404 ...)`).
    - Sensitive fields are stripped on every GET response (`stripSensitive` is invoked in GET, POST response, PATCH response).
    - `internal/testmock/server.go` calls `RegisterSnmpManagerHandlers(mux)` from the bootstrap function.
    - `go build ./internal/testmock/...` exits 0.
  </acceptance_criteria>
  <done>
    Mock handler compiles; GET-no-match returns 200 + `{"items":[]}`; passphrases / community are never echoed in GET responses; store is reachable from provider tests via the bootstrap return.
  </done>
</task>

<task id="T04" type="auto">
  <name>T04: Write 5 client unit tests (TestUnit_SnmpManager_*)</name>
  <files>internal/client/snmp_managers_test.go</files>
  <read_first>
    - CONVENTIONS.md §Test Conventions, §Test Coverage
    - mcp__serena__find_symbol name=TestUnit_Target_Get_Found relative_path=internal/client/targets_test.go include_body=true
    - mcp__serena__find_symbol name=TestUnit_Target_Post relative_path=internal/client/targets_test.go include_body=true
    - mcp__serena__find_symbol name=newTestClient relative_path=internal/client/client_test.go include_body=true
  </read_first>
  <action>
    Create `internal/client/snmp_managers_test.go` (package `client_test`) with exactly these 5 tests:

    1. `TestUnit_SnmpManager_Get_Found` — `httptest.NewServer` returns one manager when `?names=mgr1`; assert `Name`, `Host`, `Version`, `V3.User`. Confirm sensitive fields (`V3.AuthPassphrase`, `V3.PrivacyPassphrase`, `V2c.Community`) come back as empty string.
    2. `TestUnit_SnmpManager_Get_NotFound` — `httptest.NewServer` returns `{"items":[]}` with HTTP **200** on `?names=missing`. Assert `getOneByName[SnmpManager]` surfaces a not-found error (the canonical error type used elsewhere in this codebase — find it via Serena on `Get_NotFound` tests in `targets_test.go`).
    3. `TestUnit_SnmpManager_Post` — Posts `SnmpManagerPost{Host:"snmp.example", Notification:"trap", Version:"v3", V3:&SnmpV3Post{User:"u",AuthProtocol:"SHA",AuthPassphrase:"secret",PrivacyProtocol:"AES",PrivacyPassphrase:"longpriv8"}}` against a mock; assert the request body JSON contains `"v3":{"user":"u","auth_protocol":"SHA","auth_passphrase":"secret",...}` and the response is decoded into a `*SnmpManager`. Also assert query param `?names=mgr1` is sent.
    4. `TestUnit_SnmpManager_Patch` — PATCH with `SnmpManagerPatch{Host: stringPtr("new.host")}`; assert request body is `{"host":"new.host"}` (no other fields, thanks to `omitempty`).
    5. `TestUnit_SnmpManager_Delete` — DELETE returns 200; assert the request method is DELETE and the URL contains `names=mgr1`.

    Use `newTestClient(t, srv)` (from `client_test` helpers). Use `t.Fatalf` for setup errors, `t.Errorf` for assertion failures.

    No `Get_Found` test should reach a non-mock `httptest.NewServer` — keep all 5 tests offline.
  </action>
  <verify>
    <automated>cd /home/gule/Workspace/team-infrastructure/terraform-provider-mica &amp;&amp; go test -count=1 -run "TestUnit_SnmpManager_(Get_Found|Get_NotFound|Post|Patch|Delete)$" ./internal/client/... &amp;&amp; rg -n "func TestUnit_SnmpManager_Get_Found\(" internal/client/snmp_managers_test.go &amp;&amp; rg -n "func TestUnit_SnmpManager_Get_NotFound\(" internal/client/snmp_managers_test.go &amp;&amp; rg -n "func TestUnit_SnmpManager_Post\(" internal/client/snmp_managers_test.go &amp;&amp; rg -n "func TestUnit_SnmpManager_Patch\(" internal/client/snmp_managers_test.go &amp;&amp; rg -n "func TestUnit_SnmpManager_Delete\(" internal/client/snmp_managers_test.go</automated>
  </verify>
  <acceptance_criteria>
    - All 5 tests exist with the literal names listed above.
    - `go test -count=1 -run "TestUnit_SnmpManager_(Get_Found|Get_NotFound|Post|Patch|Delete)$" ./internal/client/...` exits 0.
    - `TestUnit_SnmpManager_Get_NotFound` provokes a 200+empty-list response (NOT a 404) and asserts the resulting client error.
    - `TestUnit_SnmpManager_Patch` JSON body assertion confirms only the changed field is sent (validates `omitempty` correctness).
  </acceptance_criteria>
  <done>
    5 client tests green; 200+empty-list contract validated; PATCH body proves `omitempty` works.
  </done>
</task>

<task id="T05" type="auto">
  <name>T05: Implement resource (4 interfaces, SingleNestedAttribute v2c/v3, write-once Read, 6 drift logs)</name>
  <files>internal/provider/snmp_manager_resource.go</files>
  <read_first>
    - CONVENTIONS.md §Resource Implementation
    - .planning/phases/61-flashblade-snmp-manager/61-CONTEXT.md (D-02, D-04, D-06, D-08, D-09, D-10)
    - mcp__serena__find_symbol name=ArrayConnectionResource relative_path=internal/provider/array_connection_resource.go include_body=true
    - mcp__serena__find_symbol name=mapDirectoryServiceToModel relative_path=internal/provider/directory_service_management_resource.go include_body=true
    - mcp__serena__find_symbol name=NewTargetResource relative_path=internal/provider/target_resource.go include_body=true
    - mcp__serena__find_symbol name=targetResource relative_path=internal/provider/target_resource.go include_body=true
    - mcp__serena__find_symbol name=nullTimeoutsValue relative_path=internal/provider/helpers.go include_body=true (or wherever it lives — find via Serena)
  </read_first>
  <action>
    Create `internal/provider/snmp_manager_resource.go`. Required interfaces (ALL 4):
    - `resource.Resource`
    - `resource.ResourceWithConfigure`
    - `resource.ResourceWithImportState`
    - `resource.ResourceWithUpgradeState` (Schema `Version: 0`, empty `UpgradeState` map per CONVENTIONS.md "Empty map when version is 0")

    **Schema:**

    Top-level attributes:
    - `id` — Computed string, `UseStateForUnknown()`.
    - `name` — Required string, `RequiresReplace()`.
    - `host` — Required string. **No** plan modifier.
    - `notification` — Required string, validator `OneOf("inform", "trap")`. No plan modifier.
    - `version` — Required string, validator `OneOf("v2c", "v3")`. **No** plan modifier (D-06: in-place switch allowed).
    - `v2c` — `schema.SingleNestedAttribute{ Optional: true, Computed: true, Attributes: ... }`. Attributes:
        - `community` — Optional string, `Sensitive: true`, validator `LengthAtMost(32)`.
    - `v3` — `schema.SingleNestedAttribute{ Optional: true, Computed: true, Attributes: ... }`. Attributes:
        - `user` — Optional+Computed string.
        - `auth_protocol` — Optional+Computed string, validator `OneOf("MD5", "SHA")`.
        - `auth_passphrase` — Optional string, `Sensitive: true`, validator `LengthAtMost(32)`.
        - `privacy_protocol` — Optional+Computed string, validator `OneOf("AES", "DES")`.
        - `privacy_passphrase` — Optional string, `Sensitive: true`, validator `LengthBetween(8, 63)`.
    - `timeouts` — all 4 (Create 20m, Read 5m, Update 20m, Delete 30m).

    **Create:** Build `SnmpManagerPost` (use `SnmpV3Post` for the v3 block), call `client.PostSnmpManager(ctx, name, body)`, map response into state. **Preserve user-supplied sensitive fields** in state (they were just sent, mock/API will not echo them back).

    **Read:** Call `client.GetSnmpManager(ctx, name)`. On not-found, `resp.State.RemoveResource(ctx)` and return. Build `mapSnmpManagerToModel(ctx, &state, mgr, &resp.Diagnostics)`:
    - Compare each of the 6 leaf fields (`host`, `notification`, `version`, `v3.user`, `v3.auth_protocol`, `v3.privacy_protocol`) against current state; on mismatch emit `tflog.Debug(ctx, "drift detected", map[string]any{"resource":"flashblade_snmp_manager", "field":"<leaf>", "was": <oldVal>, "now": <newVal>})`.
    - **NEVER** read or assign `v2c.community`, `v3.auth_passphrase`, `v3.privacy_passphrase` from the API response. The state value is preserved as-is. Use a `// skip sensitive write-once` comment line above each skip site for grep-ability.

    **Update:** Build `SnmpManagerPatch` with pointers to ONLY the changed fields (compare plan vs. state). For nested blocks, send the entire `*SnmpV2c` / `*SnmpV3` if any leaf changed inside. When `version` changes from v3 -> v2c, send `Version` + `V2c` and OMIT `V3` (D-06). Call `client.PatchSnmpManager(...)`. Preserve sensitive fields in state.

    **Delete:** Call `client.DeleteSnmpManager(ctx, name)`.

    **ImportState:** Import by name. Use `nullTimeoutsValue()`. Set the three sensitive fields to `types.StringNull()` inside their nested objects:
    - `v2c = { community = null }` (only if `version == "v2c"` on the read-back; otherwise `v2c = null`)
    - `v3  = { user, auth_protocol, privacy_protocol from API; auth_passphrase = null; privacy_passphrase = null }` (only if `version == "v3"`)

    Match the SingleNestedAttribute pattern in `array_connection_resource.go` for object construction (`types.ObjectValue` directly, no passthrough wrappers — CONVENTIONS.md §Patterns to Follow).
  </action>
  <verify>
    <automated>cd /home/gule/Workspace/team-infrastructure/terraform-provider-mica &amp;&amp; go build ./internal/provider/... &amp;&amp; rg -n "func NewSnmpManagerResource\(" internal/provider/snmp_manager_resource.go &amp;&amp; rg -c "drift detected" internal/provider/snmp_manager_resource.go | xargs -I{} sh -c 'test {} -ge 6' &amp;&amp; rg -n "// skip sensitive write-once" internal/provider/snmp_manager_resource.go | wc -l | xargs -I{} sh -c 'test {} -ge 3' &amp;&amp; rg -n "resource.ResourceWithUpgradeState" internal/provider/snmp_manager_resource.go &amp;&amp; rg -n "resource.ResourceWithImportState" internal/provider/snmp_manager_resource.go &amp;&amp; rg -n "SingleNestedAttribute" internal/provider/snmp_manager_resource.go | wc -l | xargs -I{} sh -c 'test {} -ge 2'</automated>
  </verify>
  <acceptance_criteria>
    - File `internal/provider/snmp_manager_resource.go` exists.
    - `NewSnmpManagerResource` is exported.
    - All 4 interfaces wired: `Resource`, `ResourceWithConfigure`, `ResourceWithImportState`, `ResourceWithUpgradeState` (grep all four interface assertions or method signatures).
    - Schema `Version: 0`.
    - `v2c` and `v3` are `schema.SingleNestedAttribute` (grep: at least 2 occurrences of `SingleNestedAttribute`).
    - **No** `RequiresReplace` on `version` (grep: `version` field block should NOT contain `RequiresReplace`).
    - Exactly **6** `tflog.Debug(ctx, "drift detected"` calls (one per leaf). The sensitive fields MUST NOT appear in any drift log call.
    - At least **3** `// skip sensitive write-once` markers in Read mapping (community, auth_passphrase, privacy_passphrase).
    - All 4 timeouts present (20m / 5m / 20m / 30m).
    - `ImportState` calls `nullTimeoutsValue()` and sets sensitive fields to `types.StringNull()`.
    - `go build ./internal/provider/...` exits 0.
  </acceptance_criteria>
  <done>
    Resource compiles; sensitive fields are write-once (never assigned from API in Read); 6 drift logs cover the 6 non-sensitive leaves; v2c<->v3 in-place switch is permitted (no `RequiresReplace` on `version`).
  </done>
</task>

<task id="T06" type="auto">
  <name>T06: Write 3 resource tests (Lifecycle, Import, DriftDetection)</name>
  <files>internal/provider/snmp_manager_resource_test.go</files>
  <read_first>
    - CONVENTIONS.md §Test Conventions, §Test Coverage (Resource minimums)
    - mcp__serena__find_symbol name=TestUnit_TargetResource_Lifecycle relative_path=internal/provider/target_resource_test.go include_body=true
    - mcp__serena__find_symbol name=TestUnit_TargetResource_Import relative_path=internal/provider/target_resource_test.go include_body=true
    - mcp__serena__find_symbol name=TestUnit_TargetResource_DriftDetection relative_path=internal/provider/target_resource_test.go include_body=true
    - mcp__serena__find_symbol name=testNewMockedProvider relative_path=internal/provider/provider_test.go include_body=true (or wherever the helper lives; locate via Serena)
  </read_first>
  <action>
    Create `internal/provider/snmp_manager_resource_test.go`. Use the **5 mandatory provider-test helpers** convention (`newTestSnmpManagerResource`, `snmpManagerResourceSchema`, `buildSnmpManagerType`, `nullSnmpManagerConfig`, `snmpManagerPlanWith`).

    Tests:

    1. **`TestUnit_SnmpManagerResource_Lifecycle`** — Use `testNewMockedProvider()` + the returned `snmpManagerStore`. Steps:
       a. Create with `version="v3"`, `notification="trap"`, full v3 block (user, MD5, secret, AES, longpriv8). Assert state has all fields including sensitive ones (state-preserved values).
       b. Update `host`. Assert PATCH body contains ONLY `{"host": "..."}`.
       c. Update `notification` from `trap` to `inform`.
       d. Update v3 inner field (`auth_protocol` MD5 -> SHA). Assert PATCH body contains the full `v3` atomic block.
       e. Delete. Assert manager is gone from store.

    2. **`TestUnit_SnmpManagerResource_Import`** — Seed store with a v3 manager. Import by `name`. Assert the resulting state has `auth_passphrase = null`, `privacy_passphrase = null` (D-09), and `user`/`auth_protocol`/`privacy_protocol` populated from the API. Assert `timeouts` are null (via `nullTimeoutsValue()`).

    3. **`TestUnit_SnmpManagerResource_DriftDetection`** — Seed store with a manager, mutate the stored entry directly (`store.byName["mgr1"].Host = "drifted"`), trigger Read, and assert via `tflog` capture (or by structure of the resulting state diff) that all 6 leaf drift logs fire when each leaf is mutated. At minimum, assert the 6 specific log entries by `field` key: `host`, `notification`, `version`, `v3.user`, `v3.auth_protocol`, `v3.privacy_protocol`. Sensitive fields MUST NOT appear in any captured drift log.

    Use the `acctest` framework's `resource.UnitTest` (NOT `resource.Test`, which requires `TF_ACC=1`).
  </action>
  <verify>
    <automated>cd /home/gule/Workspace/team-infrastructure/terraform-provider-mica &amp;&amp; go test -count=1 -run "TestUnit_SnmpManagerResource_(Lifecycle|Import|DriftDetection)$" ./internal/provider/... &amp;&amp; rg -n "func TestUnit_SnmpManagerResource_Lifecycle\(" internal/provider/snmp_manager_resource_test.go &amp;&amp; rg -n "func TestUnit_SnmpManagerResource_Import\(" internal/provider/snmp_manager_resource_test.go &amp;&amp; rg -n "func TestUnit_SnmpManagerResource_DriftDetection\(" internal/provider/snmp_manager_resource_test.go</automated>
  </verify>
  <acceptance_criteria>
    - All 3 tests exist with the literal names above.
    - `go test -count=1 -run "TestUnit_SnmpManagerResource_(Lifecycle|Import|DriftDetection)$" ./internal/provider/...` exits 0.
    - Lifecycle test exercises Create + at least 2 updates + Delete.
    - Import test asserts `auth_passphrase` and `privacy_passphrase` are null in imported state.
    - DriftDetection test verifies exactly the 6 leaves listed in D-10 and confirms sensitive fields are absent from captured logs.
  </acceptance_criteria>
  <done>
    3 resource tests green; write-once import behaviour confirmed by test; 6-leaf drift contract enforced by test.
  </done>
</task>

<task id="T07" type="auto">
  <name>T07: Implement data source (DataSource + DataSourceWithConfigure)</name>
  <files>internal/provider/snmp_manager_data_source.go</files>
  <read_first>
    - CONVENTIONS.md §Data Source Implementation
    - mcp__serena__find_symbol name=NewTargetDataSource relative_path=internal/provider/target_data_source.go include_body=true
    - mcp__serena__find_symbol name=targetDataSource relative_path=internal/provider/target_data_source.go include_body=true
  </read_first>
  <action>
    Create `internal/provider/snmp_manager_data_source.go`. Implements **2 interfaces only**: `datasource.DataSource`, `datasource.DataSourceWithConfigure`. No timeouts. No plan modifiers.

    Schema attributes (mirror resource schema shape):
    - `name` — **Required** string.
    - `id`, `host`, `notification`, `version` — **Computed** strings.
    - `v2c` — `schema.SingleNestedAttribute{ Computed: true, Attributes: { community: { Computed: true, Sensitive: true } } }`.
    - `v3` — `schema.SingleNestedAttribute{ Computed: true, Attributes: { user, auth_protocol, auth_passphrase (Sensitive), privacy_protocol, privacy_passphrase (Sensitive), all Computed } }`.

    Read: call `client.GetSnmpManager(ctx, name)`. On not-found: `resp.Diagnostics.AddError(...)` (NOT `RemoveResource` — that's resource-only per CONVENTIONS.md). Inline field mapping is fine; sensitive fields end up as empty strings (since API doesn't return them) — convert empty string to `types.StringNull()` for cleaner downstream consumption.

    `NewSnmpManagerDataSource` exported.
  </action>
  <verify>
    <automated>cd /home/gule/Workspace/team-infrastructure/terraform-provider-mica &amp;&amp; go build ./internal/provider/... &amp;&amp; rg -n "func NewSnmpManagerDataSource\(" internal/provider/snmp_manager_data_source.go &amp;&amp; rg -n "datasource.DataSourceWithConfigure" internal/provider/snmp_manager_data_source.go &amp;&amp; rg -n "AddError" internal/provider/snmp_manager_data_source.go</automated>
  </verify>
  <acceptance_criteria>
    - `NewSnmpManagerDataSource` exported.
    - Exactly 2 datasource interfaces wired (`DataSource`, `DataSourceWithConfigure`); no `WithImportState`, no `WithUpgradeState`, no timeouts.
    - `name` is `Required`; everything else is `Computed`.
    - Not-found path calls `AddError` (not `RemoveResource`).
    - `go build ./internal/provider/...` exits 0.
  </acceptance_criteria>
  <done>
    Data source compiles; matches CONVENTIONS.md §Data Source Implementation rules.
  </done>
</task>

<task id="T08" type="auto">
  <name>T08: Write 1 data source test (TestUnit_SnmpManagerDataSource_Basic)</name>
  <files>internal/provider/snmp_manager_data_source_test.go</files>
  <read_first>
    - mcp__serena__find_symbol name=TestUnit_TargetDataSource_Basic relative_path=internal/provider/target_data_source_test.go include_body=true
  </read_first>
  <action>
    Create `internal/provider/snmp_manager_data_source_test.go`. Use the 4 mandatory DS test helpers (`newTestSnmpManagerDataSource`, `snmpManagerDSSchema`, `buildSnmpManagerDSType`, `nullSnmpManagerDSConfig`).

    `TestUnit_SnmpManagerDataSource_Basic`:
    - Seed a v3 manager into the store.
    - Read it via the data source by `name`.
    - Assert `host`, `notification`, `version`, `v3.user`, `v3.auth_protocol`, `v3.privacy_protocol` are populated.
    - Assert `v3.auth_passphrase` and `v3.privacy_passphrase` are `null` (API doesn't return them; DS mapping converts empty to null).
    - Assert not-found path triggers a diagnostic via `AddError`.
  </action>
  <verify>
    <automated>cd /home/gule/Workspace/team-infrastructure/terraform-provider-mica &amp;&amp; go test -count=1 -run "TestUnit_SnmpManagerDataSource_Basic$" ./internal/provider/... &amp;&amp; rg -n "func TestUnit_SnmpManagerDataSource_Basic\(" internal/provider/snmp_manager_data_source_test.go</automated>
  </verify>
  <acceptance_criteria>
    - Test exists with the literal name.
    - `go test -count=1 -run "TestUnit_SnmpManagerDataSource_Basic$" ./internal/provider/...` exits 0.
    - Test asserts sensitive fields are `null` in DS state.
  </acceptance_criteria>
  <done>
    1 DS test green; sensitive-fields-null contract validated.
  </done>
</task>

<task id="T09" type="auto">
  <name>T09: Register resource + data source in provider.go</name>
  <files>internal/provider/provider.go</files>
  <read_first>
    - mcp__serena__find_symbol name=Resources relative_path=internal/provider/provider.go include_body=true
    - mcp__serena__find_symbol name=DataSources relative_path=internal/provider/provider.go include_body=true
  </read_first>
  <action>
    Append `NewSnmpManagerResource` to the slice returned by `Resources()` and `NewSnmpManagerDataSource` to the slice returned by `DataSources()`. Maintain alphabetical order if that's the existing convention (verify by inspecting the slice content returned by Serena).

    Do not touch any other entry.
  </action>
  <verify>
    <automated>cd /home/gule/Workspace/team-infrastructure/terraform-provider-mica &amp;&amp; go build ./... &amp;&amp; rg -n "NewSnmpManagerResource" internal/provider/provider.go &amp;&amp; rg -n "NewSnmpManagerDataSource" internal/provider/provider.go</automated>
  </verify>
  <acceptance_criteria>
    - Both `NewSnmpManagerResource` and `NewSnmpManagerDataSource` appear exactly once in `internal/provider/provider.go`.
    - `go build ./...` exits 0.
  </acceptance_criteria>
  <done>
    Provider knows about the new resource + data source.
  </done>
</task>

<task id="T10" type="auto">
  <name>T10: Write HCL examples (resource.tf, import.sh, data-source.tf)</name>
  <files>examples/resources/flashblade_snmp_manager/resource.tf, examples/resources/flashblade_snmp_manager/import.sh, examples/data-sources/flashblade_snmp_manager/data-source.tf</files>
  <read_first>
    - .planning/phases/61-flashblade-snmp-manager/61-CONTEXT.md (D-07, D-09, D-17)
    - mcp__serena__get_symbols_overview file=examples/resources/flashblade_target/resource.tf  # for shape only — small file, Read is fine here since it's not a .go file
    - examples/resources/flashblade_target/import.sh
    - examples/data-sources/flashblade_target/data-source.tf
  </read_first>
  <action>
    **resource.tf** (primary example = v3, plus commented v2c block):

    ```hcl
    # Primary example: SNMPv3 trap destination.
    resource "flashblade_snmp_manager" "prod_traps" {
      name         = "prod-snmp"
      host         = "snmp.example.com"
      notification = "trap"
      version      = "v3"

      v3 = {
        user               = "purity_user"
        auth_protocol      = "SHA"
        auth_passphrase    = "auth-secret-32max"
        privacy_protocol   = "AES"
        privacy_passphrase = "priv-secret-min8-max63"
      }
    }

    # Alternative: SNMPv2c (commented).
    # resource "flashblade_snmp_manager" "v2c_example" {
    #   name         = "legacy-snmp"
    #   host         = "snmp-old.example.com"
    #   notification = "inform"
    #   version      = "v2c"
    #
    #   v2c = {
    #     community = "public"
    #   }
    # }

    # NOTE: switching `version` in place is permitted (no RequiresReplace). If you observe
    # drift on the unused block after a switch, remove it via `terraform state rm` or
    # taint+apply. See provider docs for details.
    ```

    **import.sh** (import by NAME, not UUID):

    ```bash
    # Import by SNMP manager name. After import, sensitive fields
    # (community, auth_passphrase, privacy_passphrase) are null in state.
    # Set them in your HCL and `terraform apply` to materialise them.
    terraform import flashblade_snmp_manager.prod_traps prod-snmp
    ```

    **data-source.tf**:

    ```hcl
    data "flashblade_snmp_manager" "prod_traps" {
      name = "prod-snmp"
    }

    output "snmp_host" {
      value = data.flashblade_snmp_manager.prod_traps.host
    }
    ```
  </action>
  <verify>
    <automated>cd /home/gule/Workspace/team-infrastructure/terraform-provider-mica &amp;&amp; test -f examples/resources/flashblade_snmp_manager/resource.tf &amp;&amp; test -f examples/resources/flashblade_snmp_manager/import.sh &amp;&amp; test -f examples/data-sources/flashblade_snmp_manager/data-source.tf &amp;&amp; rg -n "terraform import flashblade_snmp_manager" examples/resources/flashblade_snmp_manager/import.sh &amp;&amp; rg -n "version *= *\"v3\"" examples/resources/flashblade_snmp_manager/resource.tf &amp;&amp; rg -n "auth_protocol *= *\"SHA\"" examples/resources/flashblade_snmp_manager/resource.tf</automated>
  </verify>
  <acceptance_criteria>
    - All 3 example files exist at the exact paths in `files_modified`.
    - `resource.tf` includes a v3 block AND a commented v2c snippet AND a comment about the in-place version switch.
    - `import.sh` imports by name `prod-snmp` (NOT a UUID).
    - `data-source.tf` exposes at least one output.
  </acceptance_criteria>
  <done>
    HCL examples present and self-explanatory; doc generation in T11 will consume them.
  </done>
</task>

<task id="T11" type="auto">
  <name>T11: Regenerate docs via `make docs`</name>
  <files>docs/resources/snmp_manager.md, docs/data-sources/snmp_manager.md</files>
  <read_first>
    - CONVENTIONS.md §Documentation
    - GNUmakefile
  </read_first>
  <action>
    Run `make docs`. Verify the generator produced:
    - `docs/resources/snmp_manager.md`
    - `docs/data-sources/snmp_manager.md`

    Do NOT edit either file by hand. If the generator complains (missing example, malformed HCL), fix the example and re-run `make docs`.

    Re-run `make docs` a second time — confirm `git status` shows no further changes (idempotency check).
  </action>
  <verify>
    <automated>cd /home/gule/Workspace/team-infrastructure/terraform-provider-mica &amp;&amp; make docs &amp;&amp; test -f docs/resources/snmp_manager.md &amp;&amp; test -f docs/data-sources/snmp_manager.md &amp;&amp; make docs &amp;&amp; git diff --quiet docs/resources/snmp_manager.md docs/data-sources/snmp_manager.md</automated>
  </verify>
  <acceptance_criteria>
    - Both doc files exist.
    - Running `make docs` twice produces no diff on the second run (idempotent).
    - Neither file was manually edited (no human comments / TODOs in the generated output).
  </acceptance_criteria>
  <done>
    Docs are generated and stable.
  </done>
</task>

<task id="T12" type="auto">
  <name>T12: Move ROADMAP.md row to Implemented + refresh footer (same-commit rule)</name>
  <files>ROADMAP.md</files>
  <read_first>
    - .planning/phases/61-flashblade-snmp-manager/61-CONTEXT.md (D-18)
    - ROADMAP.md (lines 85-160 already loaded — Array Administration table + Medium Priority Not Implemented table)
  </read_first>
  <action>
    **Remove** the `SNMP Managers` row from the *Medium Priority -- Admin and security* table (currently ~line 145).

    **Append** a new row to the *Array Administration / Implemented* table (after `Management Access Policy DS Role Membership`):

    ```
    | SNMP Managers | `flashblade_snmp_manager` | Yes | Done | v2.23.1; full CRUD; sensitive write-once community/passphrases; /test endpoint deferred |
    ```

    **Refresh** the footer / counters: increment the implemented count, bump the provider version to `v2.23.1`, set "Last updated" to today's date (2026-05-20). If the file uses a header counter like "55/X covered", increment to 56.

    **DO NOT** edit `.planning/ROADMAP.md` for this — D-18 explicitly targets the repo-level `ROADMAP.md` at the project root.

    This change MUST be committed in the SAME commit as the implementation (T13's quality-gate commit), not a separate commit.
  </action>
  <verify>
    <automated>cd /home/gule/Workspace/team-infrastructure/terraform-provider-mica &amp;&amp; rg -n "SNMP Managers \| \`flashblade_snmp_manager\`" ROADMAP.md &amp;&amp; rg -nP "SNMP Managers.*Done.*v2\.23\.1" ROADMAP.md &amp;&amp; ! rg -nP "^\| SNMP Managers \| Resource \|" ROADMAP.md</automated>
  </verify>
  <acceptance_criteria>
    - `SNMP Managers` row exists in *Array Administration / Implemented* with `Done` status and the prescribed notes.
    - `SNMP Managers` row is **gone** from *Medium Priority -- Admin and security*.
    - Footer date and provider version reflect `v2.23.1` / `2026-05-20`.
    - The change is staged for inclusion in the implementation commit (do not commit yet — T13 commits everything together).
  </acceptance_criteria>
  <done>
    Repo-level ROADMAP.md row moved; counters refreshed; ready for atomic commit with the implementation.
  </done>
</task>

<task id="T13" type="auto">
  <name>T13: Quality gates (build, test count >= 816, lint) + final commit</name>
  <files>(no new files — runs verification across the tree and produces the commit)</files>
  <read_first>
    - CONVENTIONS.md §Test Coverage (TEST_BASELINE rule)
    - GNUmakefile (TEST_BASELINE = 807, must NOT be bumped)
  </read_first>
  <action>
    **Step 1 — Run the full quality gate suite (in order, fail-fast):**

    ```bash
    make build
    make lint
    make test
    make docs   # second-run idempotency check
    git diff --quiet docs/   # must exit 0
    ```

    `make test` MUST report a total count >= 816 (`TEST_BASELINE` 807 + 9 new tests). DO NOT modify `TEST_BASELINE` in `GNUmakefile` (release-only per `.planning/STATE.md` and CONVENTIONS.md).

    **Step 2 — Assemble the implementation commit:**

    ```bash
    git add internal/client/snmp_managers.go internal/client/snmp_managers_test.go internal/client/models_admin.go \
            internal/testmock/handlers/snmp_managers.go internal/testmock/server.go \
            internal/provider/snmp_manager_resource.go internal/provider/snmp_manager_resource_test.go \
            internal/provider/snmp_manager_data_source.go internal/provider/snmp_manager_data_source_test.go \
            internal/provider/provider.go \
            examples/resources/flashblade_snmp_manager/ examples/data-sources/flashblade_snmp_manager/ \
            docs/resources/snmp_manager.md docs/data-sources/snmp_manager.md \
            ROADMAP.md

    git commit --no-verify -m "feat(snmp): add flashblade_snmp_manager resource and data source

    Implements full CRUD on /api/2.23/snmp-managers with v2c and v3 nested
    config blocks. Sensitive fields (community, auth_passphrase,
    privacy_passphrase) are write-once: API never returns them on GET, so
    Read() preserves state values verbatim and ImportState nulls them.

    - 9 new TestUnit_ tests (5 client + 3 resource + 1 data source)
    - Mock handler with empty-list GET=200 (matches real API)
    - Per-leaf drift detection on 6 non-sensitive fields
    - In-place v2c<->v3 switch permitted (no RequiresReplace on version)
    - Out of scope: /snmp-managers/test endpoint (deferred milestone)

    Closes SNMP-01..13."
    ```

    **Step 3 — Post-commit sanity:**

    ```bash
    git log -1 --stat | head -40
    make test | tail -5     # confirm count >= 816 on a clean tree
    ```

    **NEVER:**
    - Include a `Co-Authored-By` trailer.
    - Drop `--no-verify`.
    - Commit `ROADMAP.md` separately from the code.
    - Bump `TEST_BASELINE` in `GNUmakefile`.
  </action>
  <verify>
    <automated>cd /home/gule/Workspace/team-infrastructure/terraform-provider-mica &amp;&amp; make build &amp;&amp; make lint &amp;&amp; ACTUAL=$(make test 2>&amp;1 | grep -oP 'actual=\K[0-9]+' | tail -1) &amp;&amp; test -n "$ACTUAL" &amp;&amp; test "$ACTUAL" -ge 816 &amp;&amp; ! rg -q "Co-Authored-By" $(git log -1 --format=%H) &amp;&amp; git log -1 --format=%s | rg -q "feat\(snmp\)" &amp;&amp; rg -n "^TEST_BASELINE=807$" GNUmakefile</automated>
  </verify>
  <acceptance_criteria>
    - `make build` exits 0.
    - `make lint` exits 0.
    - `make test` exits 0 AND reports a total count >= 816.
    - `make docs` is idempotent (no diff on second run).
    - Last commit on `implem-snmp-managers` has subject starting with `feat(snmp)` and contains all 16 modified files listed in `files_modified` (plus the two doc files).
    - Last commit message contains NO `Co-Authored-By` line.
    - `git log -1 --pretty=%B` mentions ROADMAP.md row move (implicit via the commit's file list).
    - `GNUmakefile` still has `TEST_BASELINE=807` (NOT bumped).
  </acceptance_criteria>
  <done>
    Build green, lint clean, tests >= 816, docs idempotent, one atomic commit on `implem-snmp-managers` containing code + tests + examples + docs + ROADMAP.md, committed with `--no-verify` and no `Co-Authored-By`.
  </done>
</task>

</tasks>

<verification>
**Phase-level checks** (run after T13):

1. **Build / lint / docs / tests:**
   ```bash
   make build && make lint && make test && make docs
   git diff --quiet docs/
   ```
   All exit 0; total test count >= 816.

2. **Naming convention:**
   ```bash
   rg -n "^func Test" internal/{client,provider}/snmp_manager*_test.go | rg -v "TestUnit_Snmp"
   ```
   Must produce **no** output (every test starts with `TestUnit_Snmp`).

3. **CONVENTIONS.md *New Resource* checklist** — 16 items, verified explicitly:
   | Item | Verification command |
   |------|----------------------|
   | 1. Model structs (Get/Post/Patch) | `rg -c "^type Snmp(Manager\|V2c\|V3\|V3Post\|ManagerPost\|ManagerPatch) " internal/client/models_admin.go` returns 6 |
   | 2. Client CRUD using `getOneByName[T]` | `rg -n "getOneByName\[SnmpManager\]" internal/client/snmp_managers.go` |
   | 3. Mock handler with Seed + empty-list GET=200 | `rg -n "WriteJSONListResponse" internal/testmock/handlers/snmp_managers.go` |
   | 4. Client tests (>=5) with `TestUnit_` prefix | `go test -run "TestUnit_SnmpManager_(Get_Found\|Get_NotFound\|Post\|Patch\|Delete)$" ./internal/client/...` |
   | 5. Resource with all 4 interfaces, schema v0, correct plan modifiers | `rg -c "resource.ResourceWith(Configure\|ImportState\|UpgradeState)" internal/provider/snmp_manager_resource.go` >= 3 |
   | 6. Drift detection on 6 fields | `rg -c "drift detected" internal/provider/snmp_manager_resource.go` >= 6 |
   | 7. ImportState with `nullTimeoutsValue()` | `rg -n "nullTimeoutsValue\(\)" internal/provider/snmp_manager_resource.go` |
   | 8. Resource tests (>=3) | `go test -run "TestUnit_SnmpManagerResource_(Lifecycle\|Import\|DriftDetection)$" ./internal/provider/...` |
   | 9. Data source with Configure + Read | `rg -n "DataSourceWithConfigure" internal/provider/snmp_manager_data_source.go` |
   | 10. Data source test (>=1) | `go test -run "TestUnit_SnmpManagerDataSource_Basic$" ./internal/provider/...` |
   | 11. Registration in `provider.go` | `rg -n "NewSnmpManager(Resource\|DataSource)" internal/provider/provider.go` returns 2 |
   | 12. HCL examples | `test -f examples/resources/flashblade_snmp_manager/resource.tf && test -f examples/resources/flashblade_snmp_manager/import.sh && test -f examples/data-sources/flashblade_snmp_manager/data-source.tf` |
   | 13. `make docs` regenerated | `test -f docs/resources/snmp_manager.md && test -f docs/data-sources/snmp_manager.md` |
   | 14. `make test` count >= TEST_BASELINE + 9 (816) | parsed in T13 verify block |
   | 15. `make lint` clean | `make lint` exit 0 |
   | 16. ROADMAP.md updated | `rg -n "SNMP Managers.*Done.*v2\.23\.1" ROADMAP.md` |

4. **Sensitive-field safety audit:**
   ```bash
   rg -n "(community|auth_passphrase|privacy_passphrase)" internal/provider/snmp_manager_resource.go | rg "tflog"
   ```
   Must produce **no** output (no sensitive value ever appears in a tflog call).

5. **GNUmakefile baseline NOT bumped:**
   ```bash
   rg -n "^TEST_BASELINE=807$" GNUmakefile
   ```
   Must match.

6. **Commit policy:**
   ```bash
   git log implem-snmp-managers..main || true
   git log --oneline main..implem-snmp-managers
   git log main..implem-snmp-managers --pretty=%B | rg "Co-Authored-By"   # must produce nothing
   ```
</verification>

<success_criteria>
1. `make build && make test && make lint && make docs` all clean; total test count >= 816.
2. `make docs` is idempotent (no diff on second run).
3. Mocked Terraform Create / Read / Update / Delete / Import flow passes end-to-end via the unit-test driver.
4. Repo-level `ROADMAP.md` SNMP Managers row in *Array Administration / Implemented*, status `Done`, notes include `v2.23.1`, in the **same commit** as the implementation.
5. `tflog` audit: zero occurrences of `community`, `auth_passphrase`, `privacy_passphrase` in any `tflog.*` call across `internal/provider/snmp_manager_resource.go`. Write-once behaviour verified by `TestUnit_SnmpManagerResource_Import` (null after import) and `TestUnit_SnmpManagerResource_DriftDetection` (no drift log on these fields).
6. Branch `implem-snmp-managers` on top of clean `main`, all commits used `--no-verify`, no `Co-Authored-By` trailer anywhere.
7. `GNUmakefile` `TEST_BASELINE=807` unchanged (release-only bump).
</success_criteria>

<output>
After completion, create `.planning/phases/61-flashblade-snmp-manager/61-01-implement-snmp-manager-SUMMARY.md` documenting:
- Final test count (must be >= 816)
- Commit SHA(s) on `implem-snmp-managers`
- ROADMAP.md row diff (before/after)
- Any deviations from this plan (should be NONE; if any, justify in the summary)
- Confirmation that `flashblade-resource-builder` skill was followed end-to-end
- Confirmation that all 16 *New Resource* checklist items are satisfied
</output>
