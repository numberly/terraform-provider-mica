# Phase 53 Context — Cosmetic Hygiene

**Milestone:** v2.22.3 convention-compliance
**Goal:** PATCH slice fields across policy rules use `*[]string`, and mock handler stores follow the canonical `byName`/`nextID` pattern.
**Requirements:** R-011, R-012

---

## Canonical Refs

- `CONVENTIONS.md` §Pointer rules (PATCH: every field pointer; slice pointer guidance)
- `CONVENTIONS.md` §Mock Handlers (Store pattern: `mu sync.Mutex`, `byName map`, `nextID int`; `Seed` method; synthetic ID format `fmt.Sprintf("xxx-%d", s.nextID)`)
- R-011 targets (model files):
  - `internal/client/models_policies.go` — `ObjectStoreAccessPolicyRulePatch`, `S3ExportPolicyRulePatch`, `NetworkAccessPolicyRulePatch`
  - `internal/client/models_network.go` — `NetworkInterfacePatch.Services` and `.AttachedServers`
- R-012 targets (handler files):
  - `internal/testmock/handlers/qos_policies.go` — `policies` → `byName`
  - `internal/testmock/handlers/subnets.go` — `uuid.New()` → `nextID` pattern
  - `internal/testmock/handlers/network_interfaces.go` — `uuid.New()` → `nextID` pattern

---

## Prior Decisions (inherited)

- **Phase 51 / 52** established pattern for PATCH pointer migration with state upgrader. However for R-011 we are adding `*[]string`, not `**T` — state upgrader is NOT required (no Terraform attribute type change).
- Tests baseline after Phase 52: **832**. Target after Phase 53: **≥ 832** (no net decrease; any added clear tests count as bonus).

---

## Gray Areas — Decisions

### GA-1 — Slice migration strategy

**Decision:** For each PATCH struct, migrate list fields from `[]string` / `[]NamedReference` to `*[]string` / `*[]NamedReference` with `omitempty`. Resource `Update()` passes `&values` when the attribute is present in plan and intended to be set, `nil` when untouched.

**Exception explicitly kept:** `NetworkInterfacePatch.Services` and `.AttachedServers` already carry a file-level comment "do NOT use omitempty — clearing them requires sending `[]` in JSON." This is a legitimate API-forced deviation. Rather than migrate, **document this exception in CONVENTIONS.md §Mock Handlers or §Pointer rules** as a formalized carve-out ("for API endpoints where clearing a list requires explicit `[]` payload, keep plain slices with no omitempty"). No code change for NetworkInterfacePatch.

**Rationale:** Don't break working code with a spurious migration; formalize the real-world constraint instead.

### GA-2 — Scope of slice migrations needing tests

**Decision:** For `ObjectStoreAccessPolicyRulePatch` / `S3ExportPolicyRulePatch` / `NetworkAccessPolicyRulePatch`, add one "PATCH clear list" test per resource: `TestUnit_<Resource>_Patch_Clear<Field>` that sends an empty-pointer slice (`&[]string{}`) and asserts the mock received empty list. Keep existing tests working (rename internal usage to pointer construction).

**Rationale:** Matches Phase 52 R-009 NFS pattern exactly; regression-proof.

### GA-3 — Handler migrations

**Decision:**
- `qos_policies.go`: mechanical rename of the `policies` field to `byName`. No API change to the store type; callers (`Seed`, handler methods) rename field accesses. `members` map (compound resource) stays — it's a legitimate secondary map.
- `subnets.go` + `network_interfaces.go`: replace `uuid.New().String()` with a `nextID int` counter; ID format `fmt.Sprintf("subnet-%d", s.nextID)` / `fmt.Sprintf("nic-%d", s.nextID)`. Remove the `byID` map if unused after this migration (grep for usages). Drop `github.com/google/uuid` import from these files; keep it if other handlers still need it.

### GA-4 — Commit strategy

**Decision:** Two atomic commits:
- `fix(53-01): PATCH slice fields use *[]string in 3 policy rule models` (R-011)
- `fix(53-02): mock handler hygiene — byName + nextID pattern` (R-012)

Skip test-dense migrations if any PATCH model slice has no current test coverage — don't invent tests that weren't part of the violation.

### GA-5 — Convention doc updates

**Decision:** Add formalized carve-out paragraph in `CONVENTIONS.md` §Pointer rules for NetworkInterfacePatch-style "always send" slices. Ship in the R-011 commit.

---

## Decisions Locked

| ID | Decision | Applies to |
|---|---|---|
| D-53-01 | `ObjectStoreAccessPolicyRulePatch`, `S3ExportPolicyRulePatch`, `NetworkAccessPolicyRulePatch` slice fields → `*[]string` | R-011 |
| D-53-02 | `NetworkInterfacePatch.Services`/`.AttachedServers` stay plain slices; CONVENTIONS.md formalizes "always send" exception | R-011 |
| D-53-03 | `qos_policies.go` handler: rename `policies` → `byName` | R-012 |
| D-53-04 | `subnets.go` + `network_interfaces.go` handlers: `uuid.New()` → `nextID` counter with `fmt.Sprintf` | R-012 |
| D-53-05 | 2 commits total (R-011 grouped, R-012 grouped) | phase-wide |

---

## Scope for planner

- 2 plans total
- Net test delta: +0 to +3 (new PATCH clear tests per model migrated, if existing tests don't already cover)
- No schema version bumps (PATCH-body-only changes; Terraform attribute types unchanged)
- CONVENTIONS.md gets one new paragraph about "always send" slice exception
