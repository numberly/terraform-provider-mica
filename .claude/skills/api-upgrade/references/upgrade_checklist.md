# API Version Upgrade Checklist

**Upgrading from:** `___` **To:** `___`
**Date:** `___`

## Phase 1 — Infrastructure

- [ ] Run `upgrade_version.py --from OLD --to NEW --apply`
- [ ] `make build` passes
- [ ] `make test` passes (count >= baseline)
- [ ] `const APIVersion` in `internal/client/client.go` shows NEW
- [ ] Mock server versions slice contains NEW
- [ ] All handler paths use `/api/NEW/`
- [ ] Committed: `chore: bump API version from OLD to NEW`
- [ ] **Gate 1 confirmed**

## Phase 2 — Schema Updates

- [ ] `update_models` items from migration plan reviewed
- [ ] Each modified struct updated (Post/Patch/Get fields per CONVENTIONS.md pointer rules)
- [ ] PATCH struct fields are pointers (`*string`, `**NamedReference`)
- [ ] GET struct fields are plain types (no pointers on scalars)
- [ ] Drift detection logging added for new computed/mutable fields in `Read` methods
- [ ] `make build` passes
- [ ] `make test` passes
- [ ] **Gate 2 confirmed**

## Phase 3 — New Resources

- [ ] `new_resources` items from migration plan reviewed
- [ ] `flashblade-resource-builder` workflow followed for each new resource
- [ ] Model structs (Get/Post/Patch) created in `models_<domain>.go`
- [ ] Client CRUD methods created using `getOneByName[T]`
- [ ] Mock handler created with Seed, GET, POST, PATCH, DELETE
- [ ] Resource file created with all 4 interfaces
- [ ] Data source file created
- [ ] Each new resource registered in `provider.go` (Resources + DataSources)
- [ ] >= 8 tests per new resource (4 client + 3 resource + 1 data source)
- [ ] `make test` count increased by >= 8 per new resource
- [ ] HCL examples created (`resource.tf`, `import.sh`, `data-source.tf`)
- [ ] **Gate 3 confirmed**

## Phase 4 — Deprecations

- [ ] `deprecated` items from migration plan reviewed
- [ ] Removed or stubbed in client, resource/data source files, and `provider.go`
- [ ] No orphaned references to removed endpoints in codebase
- [ ] `make build` passes
- [ ] `make test` passes
- [ ] **Gate 4 confirmed**

## Phase 5 — Documentation

- [ ] `parse_swagger.py` run → `api_references/NEW.md` generated
- [ ] Path count in generated file matches `len(swagger["paths"])` from swagger-NEW.json
- [ ] `make docs` run → `docs/` regenerated
- [ ] `ROADMAP.md` updated (new resources: Candidate → Implemented, counters updated, `Last updated` date updated)
- [ ] `FLASHBLADE_API.md` updated if hand-curated sections changed
- [ ] Committed: `docs: update API reference and provider docs for vNEW`
- [ ] PR opened
- [ ] **Gate 5 confirmed — upgrade complete**
