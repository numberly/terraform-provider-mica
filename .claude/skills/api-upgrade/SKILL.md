---
name: api-upgrade
description: "Orchestrate a FlashBlade Terraform provider upgrade to a new REST API version through 5 review-gated phases: infrastructure version bump, schema updates, new resources, deprecations, and documentation. Consumes api-diff migration plan output and delegates new resource implementation to flashblade-resource-builder."
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

Consume `update_models` items from the migration plan:

- For each modified endpoint: update Post/Patch/Get structs in `internal/client/models_<domain>.go`
- Add new fields following CONVENTIONS.md pointer rules:
  - PATCH structs: every new field is a pointer (`*string`, `**NamedReference`)
  - GET structs: plain types, no pointers on scalars
  - POST structs: no pointers, `omitempty` on optional fields
- Add drift detection logging in the corresponding resource `Read` method for every new mutable or computed field:
  ```go
  if data.NewField.ValueString() != apiObj.NewField {
      tflog.Debug(ctx, "drift detected", map[string]any{
          "resource": name, "field": "new_field",
          "was": data.NewField.ValueString(), "now": apiObj.NewField,
      })
  }
  ```
- Run `make build` + `make test` after each model change

#### Review Gate 2

Before proceeding to Phase 3, verify all of the following:

- [ ] All `update_models` items from migration plan applied
- [ ] No compilation errors (`make build` clean)
- [ ] Drift detection logging added for new fields in `Read` methods
- [ ] `make test` passes

Type 'gate-2 passed' to continue.

---

### Phase 3 — New Resources

Consume `new_resources` items from the migration plan. For each new endpoint, use the
`flashblade-resource-builder` skill (`.claude/skills/flashblade-resource-builder/SKILL.md`)
to implement the full lifecycle:

- Model structs (Get/Post/Patch) in `internal/client/models_<domain>.go`
- Client CRUD methods in `internal/client/<resource>.go` using `getOneByName[T]`
- Mock handler in `internal/testmock/handlers/<resource>.go` with Seed, GET, POST, PATCH, DELETE
- Resource in `internal/provider/<resource>_resource.go` with all 4 interfaces
- Data source in `internal/provider/<resource>_data_source.go`
- Tests: `internal/client/<resource>_test.go` + `internal/provider/<resource>_resource_test.go` + `internal/provider/<resource>_data_source_test.go`
- HCL examples: `examples/resources/flashblade_<resource>/resource.tf`, `import.sh`, and `examples/data-sources/flashblade_<resource>/data-source.tf`
- Register in `internal/provider/provider.go` (Resources + DataSources slices)
- Run `make test` after each new resource

#### Review Gate 3

Before proceeding to Phase 4, verify all of the following:

- [ ] All `new_resources` items from migration plan implemented
- [ ] Each new resource has >= 8 tests (4 client + 3 resource + 1 data source)
- [ ] `make test` passes; count increased by >= 8 per new resource
- [ ] All new resources registered in `provider.go`

Type 'gate-3 passed' to continue.

---

### Phase 4 — Deprecations

Consume `deprecated` items from the migration plan:

- Remove client CRUD methods for fully removed endpoints
- Remove or archive resource and data source files
- Remove registration from `internal/provider/provider.go`
- Run `make build` + `make test`

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

Then update `FLASHBLADE_API.md` if any hand-curated sections changed, and update `ROADMAP.md`
to move new resources from Candidate to Implemented (update counters and `Last updated` date).

Commit: `docs: update API reference and provider docs for vNEW`

#### Review Gate 5 (Final)

Before closing the upgrade, verify all of the following:

- [ ] `api_references/NEW.md` generated successfully
- [ ] `make docs` regenerated all `docs/` files
- [ ] `ROADMAP.md` updated: new resources moved from Candidate to Implemented, counters updated
- [ ] `FLASHBLADE_API.md` updated if hand-curated sections changed
- [ ] All commits pushed, PR opened

Type 'gate-5 passed' — upgrade complete.

---

## Output

- Updated provider targeting the NEW API version
- `api_references/NEW.md` — AI-optimized reference for the new version
- `ROADMAP.md` updated with new resource coverage and counters

## Troubleshooting

| Problem | Cause | Fix |
|---------|-------|-----|
| `upgrade_version.py` exits 1 "No replacements found" | Version already applied or wrong `--from` value | Check `const APIVersion` in `internal/client/client.go` for the current version |
| `make test` count decreased after Phase 3 | New resource missing test coverage | Add missing tests per CONVENTIONS.md minimums (>= 8 per resource) |
| `make docs` fails | Schema change without schema version bump | Increment `SchemaVersion` in `Schema()` and add the corresponding `UpgradeState` entry |
| `ImportError` in `parse_swagger.py` | Missing `PYTHONPATH` | Prefix command with `PYTHONPATH=.claude/skills python3 ...` |
