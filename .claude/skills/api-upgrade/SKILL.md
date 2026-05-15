---
name: api-upgrade
description: "Orchestrate a FlashBlade Terraform provider upgrade to a new REST API version through 6 review-gated phases: infrastructure version bump, schema updates, new resources, deprecations, documentation, and Pulumi bridge alignment. Consumes api-diff migration plan output and delegates new resource implementation to flashblade-resource-builder."
---

# api-upgrade

## Purpose

Provides a repeatable, review-gated sequence for upgrading the Terraform provider to a new
FlashBlade REST API version without missing steps. Consumes the `migration-plan.md` produced
by the `api-diff` skill and applies changes mechanically across all provider layers. Each phase
ends with an explicit review gate to catch errors before proceeding. The skill combines a
mechanical version bump script with guided orchestration for model changes, new resources,
deprecations, and documentation regeneration.

## When to Use

- New FlashBlade firmware is available with a new REST API version
- After `api-diff` has produced a migration plan (`migration-plan.md` or equivalent)
- When `ROADMAP.md` needs updating for new or deprecated endpoints

## Prerequisites

- `api-diff` skill completed — `/tmp/migration-plan.md` (or equivalent path) available
- `.claude/skills/api-upgrade/scripts/upgrade_version.py` present
- Python 3.10+ available (stdlib only — no pip installs required)

## Workflow

**Working directory.** All commands assume CWD is the project root (`terraform-provider-mica`). Run `pwd` once at the start to confirm; if relative-path commands fail, check CWD before debugging the command.

**Spawn one agent at a time.** Do NOT issue two `Agent()` calls in the same message during Phases 2 and 3. Agents share the working branch and several touch `provider.go`, `ROADMAP.md`, and shared model files — parallel spawns produce merge conflicts. Wait for each agent's commit, run independent verification, then spawn the next.

**Test baseline snapshot.** Before starting Phase 2, capture the current test count and store it as `INITIAL_BASELINE`:

```bash
make test 2>&1 | tee /tmp/api-upgrade-baseline.log
grep -oP 'Test count: \K\d+' /tmp/api-upgrade-baseline.log | tail -1   # captures the final count
```

After every agent commits in Phases 2/3, re-run `make test` independently and compare against the snapshot. Self-reported deltas in agent return summaries are advisory only — the orchestrator's verification is authoritative.

### Phase 1 — Infrastructure

Run the mechanical version bump script to update all hardcoded version strings in the provider:

```bash
python3 .claude/skills/api-upgrade/scripts/upgrade_version.py \
  --from OLD --to NEW --apply
```

Then verify the build and test suite:

```bash
make build
make test
```

Commit: `chore: bump API version from OLD to NEW`

```bash
git commit --no-verify -m "chore: bump API version from OLD to NEW"
```

`--no-verify` is required by project policy (CLAUDE.md) for all subagent and orchestrator commits in this workflow.

#### Review Gate 1

Before proceeding to Phase 2, verify all of the following:

- [ ] `make build` passes with zero errors
- [ ] `make test` passes (count >= previous baseline)
- [ ] `const APIVersion` in `internal/client/client.go` shows NEW version
- [ ] Mock server versions slice contains NEW version
- [ ] All mock handler paths use `/api/NEW/`

Type 'gate-1 passed' to continue.

---

### Phase 2 — Schema Updates

Consume `update_models` items from the migration plan. **Spawn the
`flashblade-resource-modifier` agent** (`.claude/agents/flashblade-resource-modifier.md`)
for each modified resource, in serial. Reliability rationale: schema migrations carry the
worst failure mode in this provider (silent state corruption when a state upgrader is
wrong), so each modification runs in an isolated context with the full
`Modify Existing Resource` checklist applied exhaustively.

**Execution mode: serial, one agent at a time.** Same reasoning as Phase 3 — agents share
the working branch, several touch `provider.go` / `ROADMAP.md` / shared model files.

**Phase 2 pre-flight (do once before the loop):**

1. Snapshot current schema versions for ALL resources in `update_models`:
   ```bash
   for r in <list-of-resources-from-update_models>; do
     # Use mcp__serena__find_symbol on Schema to read Version: line from the AST.
     # Record as: <resource>=<version> in /tmp/phase2-versions.txt
   done
   ```
   This is the source of truth. Pass these versions into agent spawns. Re-read after every commit.

2. Validate every `update_models` entry has a non-empty `changes` list. If any entry has zero changes, skip it (or remove it from the migration plan) — do NOT spawn the agent on an empty change set.

For each item in `update_models` (one at a time):

**Extract the current schema version via Serena** (NOT via grep — `grep` on `.go` files violates the Serena-first project rule and triggers a hook warning):

Use `mcp__serena__find_symbol` with name `Schema` in `internal/provider/<resource>_resource.go` and read the `Version: N` line from the returned snippet. If no `Version:` line is present in the symbol body, the schema is version 0 (implicit). Cross-check against the snapshot in `/tmp/phase2-versions.txt`.

If the snapshot value and the file value diverge (e.g. the snapshot was taken before a previous agent in this Phase 2 committed a bump), use the file value — it's authoritative. Re-snapshot after the spawn.

Then spawn:

```
Agent({
  description: "Modify flashblade_<name> for v<NEW>",
  subagent_type: "flashblade-resource-modifier",
  model: "sonnet",   // schema migration is pattern-following — not Opus
  prompt: `
    resource_name: <snake_case>
    api_endpoint: <e.g. /buckets>
    api_version: <NEW>
    api_reference_path: api_references/<NEW>.md
    domain: <storage|policies|network|...>
    current_schema_version: <N — extracted via the grep above>
    soft_delete: <true|false>
    changes:
      - add: <field_name> (<type>, <required|optional|computed>, <api_default_or_none>)
      - rename: <old_name> -> <new_name>
      - typechange: <field_name> (<old_type> -> <new_type>)
      - remove: <field_name>
  `
})
```

**Building the `changes` block:** parse the migration plan's `update_models` entry for
this resource. The agent will NOT infer changes from the swagger — you must enumerate them
explicitly. If a change is ambiguous (e.g. a typechange that may lose data), expect a
BLOCKED return and resolve the ambiguity before re-spawning.

After each agent returns:
1. Read the **Schema migration** + **Schema diff (HCL)** + **State upgrader behavior** sections — these tell you what end-users will experience.
2. **Trust but verify (mandatory, per agent — not "at least one")**, using the SHA from the agent's `**Commit:** <sha>` field (NOT `HEAD`, which after several serial agents only reflects the last one):
   ```bash
   AGENT_SHA=<sha from agent's Commit: field>

   # 1. PriorSchema verification comment is present and lists the v<N> struct fields:
   git show "$AGENT_SHA" -- internal/provider/<resource>_resource.go | grep -B1 -A20 'PriorSchema verified against'

   # 2. Test count: run make test independently and compare to snapshot.
   make test 2>&1 | grep -oP 'Test count: \K\d+' | tail -1   # must be ≥ snapshot + 1

   # 3. Test file has real assertions (not no-op placeholders):
   git show "$AGENT_SHA" -- internal/provider/<resource>_resource_test.go | grep -cE 't\.(Errorf|Fatalf)'
   #    Must return ≥ 2 (rough heuristic: every assertion-bearing test has at least one)

   # 4. Diff content matches the agent's claimed file list:
   git show --stat "$AGENT_SHA"
   ```
   Update the snapshot: `make test 2>&1 | grep -oP 'Test count: \K\d+' | tail -1 > /tmp/phase2-baseline.txt`.

3. If Status: BLOCKED — the agent will name a specific blocker (typechange data loss, field used in composite import ID, schema version mismatch, etc.). Resolve and re-spawn.

#### Review Gate 2

Before proceeding to Phase 3, verify all of the following:

- [ ] **Prerequisites resolved or deferred**: re-read `prerequisites` from the migration plan. For each item with `type: http_method`, either (a) the missing client infrastructure has been committed (e.g. `putOne[B,R]` generic added, CONVENTIONS updated, implementor agent SKILL extended) OR (b) every Phase 3 candidate using that method is being explicitly skipped this run (track them in ROADMAP under Candidate/tech debt with a note linking back to the prereq). If neither — STOP. Proceeding without addressing prereqs is the failure mode this gate exists for.
- [ ] All `update_models` items from migration plan applied (one agent invocation per resource, all returned SUCCESS)
- [ ] Each modified resource has a new state upgrader entry AND a state upgrader test (`TestUnit_<Resource>Resource_StateUpgrade_V<N>toV<N+1>`)
- [ ] `make test` passes; count delta ≥ +1 per modified resource (the new upgrader test)
- [ ] No compilation errors (`make build` clean)
- [ ] **Spot-check (mandatory — every modified resource, not "at least one")** — for each resource using its agent-returned SHA:
  ```bash
  git show <sha> -- internal/provider/<resource>_resource.go | grep -B1 -A 50 'PriorSchema verified against'
  ```
  Verify by eye that:
  - The `// PriorSchema verified against ...` comment is present and lists every field of the v<N> struct
  - Every attribute in `PriorSchema` was already present in v<N> (none are from the `changes: add:` list)
  - In the upgrader body, NEW fields use `types.*Null()` for null-hydration OR `types.ListValueMust(elemType, []attr.Value{})` for Computed empty-list hydration (per agent's reliability rule on null-vs-empty)
  - Renamed fields: prior side reads the OLD name, new side writes the NEW name
- [ ] **Independent test count check**: `make test` count ≥ snapshot + N (where N = number of resources modified, expecting +1 per resource minimum)
- [ ] `git log` shows one commit per modified resource (atomic, in the order spawned)

Type 'gate-2 passed' to continue.

---

### Phase 3 — New Resources

Consume `new_resources` items from the migration plan. For each new endpoint, **spawn the
`flashblade-resource-implementor` agent** (`.claude/agents/flashblade-resource-implementor.md`)
to implement the full lifecycle in an isolated context window. The agent follows
`flashblade-resource-builder/SKILL.md` internally.

**Execution mode: serial, one agent at a time.** Reason: agents share the working branch
(no worktrees — see CLAUDE.md), and several touch `provider.go` + `ROADMAP.md`. Running
them in parallel would cause merge conflicts on those shared files. Serial execution still
delivers the main benefit: each resource implementation runs in its own context, keeping
the orchestrator's window clean across the whole upgrade.

**Phase 3 baseline initialisation (MANDATORY before the first spawn).** The `≥ snapshot + 9` delta check below requires a Phase-3-specific baseline that reflects the state AFTER Phase 2 commits. Do NOT reuse `/tmp/phase2-baseline.txt` (it was updated after each Phase 2 agent and now represents the post-Phase-2 count, which is exactly what we want for Phase 3's start — but the file is conceptually Phase 2's last value, not Phase 3's start). Initialise explicitly:

```bash
make test 2>&1 | grep -oP 'Test count: \K\d+' | tail -1 > /tmp/phase3-baseline.txt
cat /tmp/phase3-baseline.txt  # confirm non-empty integer
```

If the file is empty or absent, the per-agent delta check below would compute `<count> - <empty> >= 9`, which evaluates as a shell error and silently passes. Verify before spawning the first agent.

For each item in `new_resources` (one at a time):

```
Agent({
  description: "Implement flashblade_<name>",
  subagent_type: "flashblade-resource-implementor",
  model: "sonnet",   // pattern-following work — not Opus (per global CLAUDE.md model rules)
  prompt: `
    resource_name: <snake_case>
    api_endpoint: <e.g. /nfs-export-policies>
    api_version: <NEW>
    api_reference_path: api_references/<NEW>.md
    domain: <e.g. policies, storage, network>
    soft_delete: <true|false>
    reference_resource: <path to closest existing resource as template>
  `
})
```

After each agent returns, verify independently using the SHA from its `**Commit:** <sha>` field (NOT `HEAD`):

```bash
AGENT_SHA=<sha from agent's Commit: field>

# 1. Test count delta — independent re-run vs snapshot:
make test 2>&1 | grep -oP 'Test count: \K\d+' | tail -1   # must be ≥ snapshot + 9

# 2. Tests contain real assertions:
git show "$AGENT_SHA" -- internal/client/<resource>_test.go internal/provider/<resource>_resource_test.go internal/provider/<resource>_data_source_test.go | grep -cE 't\.(Errorf|Fatalf)'
#    Must return ≥ 5 (one per test minimum, more typical)

# 3. Diff matches claimed files:
git show --stat "$AGENT_SHA"

# 4. provider.go registration is unique (no duplicate from re-spawn):
grep -c "New<Resource>Resource" internal/provider/provider.go    # must equal 1
grep -c "New<Resource>DataSource" internal/provider/provider.go  # must equal 1
```

If Status: BLOCKED — read the agent's "Need from orchestrator" field, resolve, then re-spawn. Update the snapshot baseline after each successful spawn: `make test 2>&1 | grep -oP 'Test count: \K\d+' | tail -1 > /tmp/phase3-baseline.txt`.

If a `new_resources` item is too divergent from existing patterns (novel auth, unusual
sub-resources, etc.), implement it inline in the orchestrator instead of delegating —
agents follow patterns, they don't invent them.

#### Review Gate 3

Before proceeding to Phase 4, verify all of the following:

- [ ] All `new_resources` items from migration plan implemented (one agent invocation per item, all returned SUCCESS)
- [ ] Each new resource has >= 9 tests (5 client + 3 resource + 1 data source)
- [ ] `make test` passes; count increased by >= 9 per new resource
- [ ] All new resources registered in `provider.go` (verify with `mcp__serena__find_symbol` on `Resources` and `DataSources` in `provider.go`)
- [ ] `make docs` regenerated `docs/resources/<resource>.md` AND `docs/data-sources/<resource>.md` for every new resource
- [ ] `ROADMAP.md` updated for every new resource — moved from "Not Implemented" to the correct "Implemented" subsection, header counters and `Last updated` date bumped
- [ ] `git log` shows one commit per resource (agent commits are atomic)

Type 'gate-3 passed' to continue.

---

### Phase 4 — Deprecations

Consume `deprecated` items from the migration plan:

- Remove client CRUD methods for fully removed endpoints
- Remove or archive resource and data source files
- Remove registration from `internal/provider/provider.go`
- Run `make build` + `make test`

Commit:
```bash
git commit --no-verify -m "feat(deprecations): remove/stub <list> for vNEW"
```

**Note:** Stubs are acceptable when a resource is removed from the API but still in use by
operators. Leave a stub returning `resp.Diagnostics.AddError("deprecated", "...")` with a
clear deprecation message rather than silently breaking existing configs.

#### Review Gate 4

Before proceeding to Phase 5, verify all of the following:

- [ ] All `deprecated` items handled (removed or stubbed)
- [ ] No dead code referencing removed endpoints
- [ ] `make build` passes
- [ ] `make test` passes

Type 'gate-4 passed' to continue.

---

### Phase 5 — Documentation

Regenerate all documentation using the `swagger-to-reference` skill
(`.claude/skills/swagger-to-reference/SKILL.md`):

```bash
PYTHONPATH=.claude/skills python3 .claude/skills/swagger-to-reference/scripts/parse_swagger.py \
  swagger-NEW.json \
  --version NEW \
  --output api_references/NEW.md
```

Regenerate provider docs:

```bash
make docs
```

Then update `FLASHBLADE_API.md` if any hand-curated sections changed, and verify `ROADMAP.md`:

- Phase 3 already moved each new resource from "Not Implemented" to "Implemented" with `Status: Done`. Phase 5 only **verifies counters and `Last updated` date are current** for the whole document — do NOT attempt to "move from Candidate to Implemented" (no such section exists; "Candidate" is a status value within the Implemented table).
- Confirm: counters in the header match the actual count of `Done` rows; `Last updated` is today's date.

Commit:
```bash
git commit --no-verify -m "docs: update API reference and provider docs for vNEW"
```

#### Review Gate 5

Before proceeding to Phase 6, verify all of the following:

- [ ] `api_references/NEW.md` generated successfully
- [ ] `make docs` regenerated all `docs/` files
- [ ] `ROADMAP.md` updated: new resources moved from Candidate to Implemented, counters updated
- [ ] `FLASHBLADE_API.md` updated if hand-curated sections changed

Type 'gate-5 passed' to continue.

---

### Phase 6 — Pulumi Bridge

The Pulumi bridge (`pulumi/provider/`) wraps the TF provider for Pulumi consumers. Resources and data sources flow through automatically via `tokens.SingleModule("flashblade_", ...)` — there is no manual registration step — BUT three things drift on every TF surface change and the Pulumi CI surfaces them only after the PR is pushed. Fix them inline now to avoid a CI round-trip.

**Step 6.1 — Bump count assertions.**

`pulumi/provider/resources_test.go` hardcodes the expected resource/data-source counts. Compute the new values, then update:

```bash
# Current TF provider counts (authoritative — count entries in Resources() + DataSources() in provider.go)
ACTUAL_RESOURCES=$(grep -cE '^\s+NewFlashblade[A-Z]\w+Resource,' internal/provider/provider.go || true)
ACTUAL_DATASOURCES=$(grep -cE '^\s+NewFlashblade[A-Z]\w+DataSource,' internal/provider/provider.go || true)

# OR more robust — read from prov via Serena `find_symbol` on Resources/DataSources
# and count the slice entries. The grep above is a quick start; if your provider.go
# uses a different constructor name pattern, prefer Serena.

# Replace the consts in pulumi/provider/resources_test.go:
#   expectedResources   = <ACTUAL_RESOURCES>
#   expectedDataSources = <ACTUAL_DATASOURCES>
```

Note: the constructor naming in this repo is `New<Resource>Resource` and `New<Resource>DataSource` without a `Flashblade` prefix. Adjust the grep accordingly, or just count `Resources()` / `DataSources()` entries by reading the function bodies via Serena.

**Step 6.2 — ComputeID overrides for new resources with composite import IDs.**

The Pulumi bridge derives the resource ID from TF's `id` attribute by default. Resources that import via a composite key (e.g. `policyName/ruleName`, `account/user/policy`) instead of a UUID need an explicit `ComputeID` override in `pulumi/provider/resources.go`. Look for prior examples: `s3_export_policy_rule`, `bucket_access_policy_rule`, `management_access_policy_directory_service_role_membership`.

For each new resource from Phase 3:

- Read its `ImportState` function via `mcp__serena__find_symbol`.
- If it parses a composite key (presence of `strings.Split(req.ID, "/")` or similar), add a `ComputeID` block in `resources.go` following the existing pattern. The block reads the relevant state fields and reconstructs the composite ID.
- If import is by name or UUID directly, no override is needed.

**Step 6.3 — Run Pulumi tests.**

```bash
cd pulumi/provider && go test ./... -count=1
cd ../..
```

Must pass. The failure mode `TestProviderInfo_ResourceAndDataSourceCounts: Resources count = X, want Y` means Step 6.1 was missed.

**Step 6.4 — Update `TEST_BASELINE` in `GNUmakefile` and pulumi CHANGELOG (optional).**

- `TEST_BASELINE` in the project `GNUmakefile` is the floor enforced by `make test`. Bump it to the current count so future regressions of more than a couple of tests are caught locally:
  ```bash
  CURRENT=$(make test 2>&1 | grep -oP 'Test count: \K\d+' | tail -1)
  sed -i "s/^TEST_BASELINE=.*/TEST_BASELINE=$CURRENT/" GNUmakefile
  ```
- `pulumi/CHANGELOG.md` is only updated when cutting a new Pulumi release (independent versioning cadence from the TF provider). If this upgrade will trigger a Pulumi release, add a top-level `## [vX.Y.Z-pulumi.beta]` entry with the API version covered; otherwise skip.

Commit:
```bash
git add pulumi/provider/resources_test.go GNUmakefile <pulumi/provider/resources.go if ComputeID added> <pulumi/CHANGELOG.md if updated>
git commit --no-verify -m "chore(pulumi): align bridge expectations with API NEW additions"
```

#### Review Gate 6 (Final)

Before closing the upgrade, verify all of the following:

- [ ] `cd pulumi/provider && go test ./... -count=1` passes
- [ ] `pulumi/provider/resources_test.go` consts match actual TF provider counts
- [ ] Every new Phase 3 resource that imports via a composite key has a corresponding `ComputeID` override in `pulumi/provider/resources.go` (Serena `find_symbol` on ImportState to verify)
- [ ] `GNUmakefile` `TEST_BASELINE` matches current `make test` count (so future regressions are caught locally before CI)
- [ ] All commits pushed, PR opened, CI green (TF Tests/Lint/Docs and Pulumi Prerequisites all SUCCESS)

Type 'gate-6 passed' — upgrade complete.

---

## Output

- Updated provider targeting the NEW API version
- `api_references/NEW.md` — AI-optimized reference for the new version
- `ROADMAP.md` updated with new resource coverage and counters

## Troubleshooting

| Problem | Cause | Fix |
|---------|-------|-----|
| `upgrade_version.py` exits 1 "No replacements found" | Version already applied or wrong `--from` value | Check `const APIVersion` in `internal/client/client.go` for the current version |
| `make test` count decreased after Phase 3 | New resource missing test coverage | Add missing tests per CONVENTIONS.md minimums (≥ 9 per new resource = 5 client + 3 resource + 1 data source) |
| Phase 5 ROADMAP says move from "Candidate" but section doesn't exist | Section "Candidate" never existed — it's a status, not a section | Phase 3 already moved rows to "Implemented"; Phase 5 only verifies counters/date |
| Schema version grep returns empty in Phase 2 pre-flight | `grep` is no longer used — Serena AST lookup replaced it | Use `mcp__serena__find_symbol` on `Schema` and read the `Version: N` line |
| Agent re-spawned after BLOCKED produces duplicate registration in `provider.go` | Implementor was supposed to skip via Serena pre-check but didn't | Manually remove the duplicate, or have agent re-check via `find_symbol` before append |
| `make docs` fails | Schema change without schema version bump | Increment `SchemaVersion` in `Schema()` and add the corresponding `UpgradeState` entry |
| `ImportError` in `parse_swagger.py` | Missing `PYTHONPATH` | Prefix command with `PYTHONPATH=.claude/skills python3 ...` |
