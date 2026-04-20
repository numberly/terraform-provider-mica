# 52-03 Summary — ImportState shim on object_store_access_key (R-008)

**Commits:**
- `be99a39` — feat(52-03): add ResourceWithImportState shim on object_store_access_key
- `1841434` — fix(52-03): update TestUnit_AccessKey_ImportRejected for new shim

## What shipped

- `internal/provider/object_store_access_key_resource.go` now asserts `resource.ResourceWithImportState` (4th mandatory interface per CONVENTIONS.md §Resource Implementation).
- `ImportState` method is implemented as an explicit reject-shim that emits `resp.Diagnostics.AddError("Import not supported", "flashblade_object_store_access_key cannot be imported because secret_access_key is only returned at creation time and is never retrievable afterwards. Recreate the resource via terraform apply instead.")`.
- Pre-existing `TestUnit_AccessKey_NoImport` (asserted the resource must NOT implement the interface) renamed to `TestUnit_AccessKey_ImportRejected` and updated to assert the new contract: interface is present, but any import emits an error diagnostic.
- CONVENTIONS.md §Resource Implementation formally documents the exception rationale (write-once secret).

## Verification

- `make test` — green, no regression.
- `make lint` — 0 issues.

## Deviation

Initial commit from the executor agent left a failing test (`TestUnit_AccessKey_NoImport` asserted the old invariant and started failing after the interface was added). Follow-up commit `1841434` rewrote the test to assert the new contract. One extra commit not in the plan template, but content is aligned with the plan's intent.

## Requirements closed

- R-008 — access key 4th interface + documented exception
