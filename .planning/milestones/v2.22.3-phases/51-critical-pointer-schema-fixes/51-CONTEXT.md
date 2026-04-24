# Phase 51 Context — Critical Pointer & Schema Fixes

**Milestone:** v2.22.3 convention-compliance
**Goal (from ROADMAP):** Users can create untagged subnets (VLAN=0), detach subnets from LAGs, and clear optional refs on export/account-export PATCH — no silent data loss.
**Requirements:** R-001, R-002, R-003, R-004, R-005

---

## Canonical Refs (mandatory reading for downstream agents)

- `CONVENTIONS.md` §Pointer rules (updated 2026-04-20) — authoritative for POST/PATCH pointer semantics and legitimate exceptions
- `CONVENTIONS.md` §State Upgraders — schema version bump + `PriorSchema` + intermediate `xxxV0Model` pattern
- `.planning/REQUIREMENTS.md` §R-001..R-005 — acceptance criteria for this phase
- `internal/provider/server_resource.go` — reference implementation for v0→v1→v2 upgrader chain
- `internal/provider/remote_credentials_resource.go` — reference for v0→v1 upgrader
- `internal/client/models_network.go` (SubnetPost, SubnetPatch) — lines 21–42
- `internal/client/models_exports.go` (FileSystemExportPatch, ObjectStoreAccountExportPatch) — lines 46–71
- `internal/provider/subnet_resource.go:62` — Version: 0 (to bump to 1)
- `internal/provider/file_system_export_resource.go:60` — Version: 0 (to bump to 1)
- `internal/provider/object_store_account_export_resource.go:57` — Version: 0 (to bump to 1)
- `swagger-2.22.json` — source of truth for field semantics (VLAN=0 untagged, reference clearability)

---

## Prior Decisions (inherited, do not re-ask)

- **CONVENTIONS.md §Pointer rules is authoritative** — `**NamedReference` for clearable PATCH refs, `*int64`/`*string` for semantic-zero POST scalars (2026-04-20 clarification).
- **Schema migration pattern** — Every resource with a model change bumps `Version` and adds a `PriorSchema` upgrader keyed by the prior version (v0→v1 in this phase). Pattern is locked: see `server_resource.go`.
- **Test baseline** — `make test` count must stay ≥ 818 and increase by ≥ 1 per added state upgrader (v2.22.2 baseline).
- **Commit granularity** — one atomic commit per resource migration (client model + resource + upgrader + tests), not one giant commit.
- **No soft-delete changes** — only file systems and buckets have soft-delete; none of the 3 resources in scope are soft-delete.

---

## Gray Areas — Decisions

### GA-1 — PATCH clear semantics exposure in Terraform schema

**Question:** When a Terraform attribute for a nested ref (e.g., `link_aggregation_group`, `share_policy`, `policy`) moves from unset in state to unset in config, should the provider send the double-pointer as `non-nil outer + nil inner` (clear) or `nil outer` (omit)?

**Decision:** When the attribute is explicitly removed from config (was set, now null in plan), send CLEAR (outer non-nil + inner nil). When the attribute was never in state and stays null, send OMIT (outer nil). The Terraform plan semantics already distinguishes these two cases via `plan.Xxx.IsNull()` vs state comparison; the resource Update code must compare `state.Xxx` and `plan.Xxx`:
- `state.Xxx` set and `plan.Xxx` null → CLEAR (`patch.Xxx = new(*NamedReference)` i.e. `**NamedReference{nil}`)
- `state.Xxx == plan.Xxx` → OMIT (`patch.Xxx = nil`)
- Any other change (set → new value or null → value) → SET (`patch.Xxx = &(&NamedReference{Name: v})`)

**Rationale:** This mirrors the convention's three-state semantics exactly and is the only way to expose CLEAR to users without breaking Terraform's own diff model. Reference helper can be added to `internal/provider/helpers.go`:

```go
func doublePointerRefForPatch(state, plan types.String) **client.NamedReference {
    if state.Equal(plan) { return nil }
    if plan.IsNull() {
        var null *client.NamedReference
        return &null
    }
    ref := &client.NamedReference{Name: plan.ValueString()}
    return &ref
}
```

This helper becomes a convention artifact — referenced in future PATCH migrations.

### GA-2 — VLAN=0 Terraform schema encoding

**Question:** Should `vlan` be `Optional + Computed` (so null in HCL means "whatever API returns") or `Optional` with explicit `Default: 0`?

**Decision:** Keep `vlan` as `Optional` with `Int64` type. Null in HCL means "don't send" (API default applies, currently interpreted as 0/untagged by API). Explicit `vlan = 0` means "send vlan=0" (user explicitly untagged). The `*int64` pointer in `SubnetPost` distinguishes these:
- `data.VLAN.IsNull()` → `body.VLAN = nil` (omit from JSON)
- `!data.VLAN.IsNull()` → `body.VLAN = ptr(data.VLAN.ValueInt64())` (send, even if 0)

**Rationale:** Matches user intent. Preserves existing behavior when VLAN is omitted. Fixes the bug for explicit `vlan=0`.

### GA-3 — State upgrader payload when model shape doesn't change on disk

**Question:** Do `R-001..R-004` require a state upgrader, since the Terraform attribute types in the schema aren't changing — only the wire format to the API?

**Decision:** Yes, bump SchemaVersion per CONVENTIONS.md §State Upgraders: "Increment `Version` in `Schema()` when adding a new attribute, changing an attribute type, or renaming an attribute." Even though the attribute types don't change, the **semantics** of `link_aggregation_group`, `share_policy`, `policy`, and `vlan` change for users (they can now clear / set to zero meaningfully). Bumping the version forces existing states to pass through a no-op upgrader that preserves current behavior — a defensive move, not strictly required by the framework but required by project convention to keep the migration chain intact.

**Rationale:** Convention is stricter than the framework requires; we follow convention. The upgraders will be no-op copies (`v0 state → v1 state` identity), but the chain stays in place for future bumps.

### GA-4 — Drift detection impact

**Question:** Do the new `**NamedReference`/`*int64` fields require updated drift detection in `Read`?

**Decision:** No changes to `Read`. Drift detection already compares the Terraform state attribute to the GET response value — neither the GET struct nor the state attribute type changes. The pointer change is PATCH-body-only.

**Rationale:** GET responses do not carry these PATCH-semantic distinctions; the server always reports the current state as a concrete value.

### GA-5 — Commit strategy and test isolation

**Decision:**
- One commit per resource: `fix(51-01): SubnetPost.VLAN *int64 + state upgrader v0→v1`, `fix(51-02): SubnetPatch.LinkAggregationGroup **NamedReference + state upgrader v1→v2`, etc.
- Merge R-001 and R-002 into a single subnet commit with a `v0→v1` upgrader (single version bump covering both changes to avoid v0→v1→v2 chain for one resource).
- Each commit must ship its state upgrader test in the same commit.

**Rationale:** Atomic, reviewable, bisect-friendly. Subnet has two model changes; combining them into one bump avoids chain complexity.

---

## Deferred Ideas (out of scope, capture for backlog)

- Adding a lint rule or `go vet` pass that rejects `*NamedReference` in PATCH structs (would prevent regressions). Candidate for a future phase in milestone v2.23+.
- Introducing a generic `doublePointerRefForPatch` helper in `internal/client/generic_helpers.go` once more PATCH resources migrate (Phase 52 slices, Phase 53 policy rules will inform the shape).

---

## Decisions Locked (summary for planner)

| ID | Decision | Applies to |
|---|---|---|
| D-51-01 | Provide `doublePointerRefForPatch(state, plan types.String) **client.NamedReference` helper in `internal/provider/helpers.go` | R-002, R-003, R-004 |
| D-51-02 | Keep `vlan` Optional Int64; null = omit, value (incl. 0) = send explicitly | R-001 |
| D-51-03 | Bump SchemaVersion to 1 on subnet, file_system_export, object_store_account_export | R-005 |
| D-51-04 | State upgraders are no-op identity transforms; existence required by convention, not by framework | R-005 |
| D-51-05 | One atomic commit per resource, each bundling model + client + resource + upgrader + tests | milestone-wide |
| D-51-06 | Merge R-001 + R-002 into one subnet v0→v1 bump (two model changes, one version jump) | R-001, R-002 |

---

## Scope for planner

Three resources, four model-level fixes, three state upgraders, plus test additions. Estimated task breakdown:
- **T-01** Model updates (`models_network.go` SubnetPost/Patch, `models_exports.go` FileSystemExportPatch + ObjectStoreAccountExportPatch)
- **T-02** Client call-site updates (no changes needed if signatures stay, just struct field handling)
- **T-03** `doublePointerRefForPatch` helper in `internal/provider/helpers.go`
- **T-04** `subnet_resource.go`: schema v1, state upgrader, Update() using helper + VLAN pointer logic
- **T-05** `file_system_export_resource.go`: schema v1, state upgrader, Update() using helper
- **T-06** `object_store_account_export_resource.go`: schema v1, state upgrader, Update() using helper
- **T-07** Tests: `TestUnit_Subnet_StateUpgrade_V0toV1`, `TestUnit_FileSystemExport_StateUpgrade_V0toV1`, `TestUnit_ObjectStoreAccountExport_StateUpgrade_V0toV1`, plus PATCH-clear unit tests for each of the 3 resources

Expected: +7 to +10 tests. Net test count target: 825+.
