---
description: Run the full FlashBlade API version upgrade workflow (5 phases, review-gated)
argument-hint: <new_version> (e.g. 2.23)
---

You are about to execute the FlashBlade Terraform provider API version upgrade workflow. Target version: **$ARGUMENTS**.

This command enforces the exact, repeatable sequence so every upgrade goes through the same hardened path. Deviating from this sequence is not allowed — if any step fails, STOP and report; do not invent alternatives.

## Pre-flight (do these IN ORDER before any phase)

1. **Working tree must be clean.**
   ```bash
   git status --short
   ```
   If any file is modified or untracked, STOP and ask the user to commit/stash. Do NOT proceed.

2. **Confirm prerequisite files exist** for both old and new versions:
   ```bash
   ls swagger-2.*.json api_references/2.*.md
   ```
   The new version (`$ARGUMENTS`) needs `swagger-$ARGUMENTS.json` and `api_references/$ARGUMENTS.md`. Identify the current OLD version from `internal/client/client.go` (the `APIVersion` constant).

3. **Determine OLD version**:
   ```bash
   grep -oP 'const APIVersion = "\K[^"]+' internal/client/client.go
   ```
   Store as `OLD`. Confirm `OLD != $ARGUMENTS` (otherwise nothing to upgrade).

4. **Create the upgrade branch**:
   ```bash
   git checkout -b test/api-upgrade-$ARGUMENTS
   ```
   If the branch already exists, ask the user whether to reuse, delete, or pick a different suffix. Never silently overwrite.

5. **Capture the test count baseline** (used as `INITIAL_BASELINE` throughout the upgrade):
   ```bash
   make test 2>&1 | tee /tmp/api-upgrade-baseline.log
   grep -oP 'Test count: \K\d+' /tmp/api-upgrade-baseline.log | tail -1 > /tmp/api-upgrade-baseline.txt
   cat /tmp/api-upgrade-baseline.txt
   ```
   The file must contain a single integer. If empty, STOP — `make test` output format changed and the regex needs updating.

6. **Generate the migration plan** using the `api-diff` skill:
   ```bash
   PYTHONPATH=.claude/skills python3 .claude/skills/api-diff/scripts/diff_swagger.py \
     swagger-OLD.json swagger-$ARGUMENTS.json \
     --format json \
     --discrepancies .claude/skills/api-diff/references/known_discrepancies.json \
     --output /tmp/diff-OLD-$ARGUMENTS.json

   python3 .claude/skills/api-diff/scripts/generate_migration_plan.py \
     /tmp/diff-OLD-$ARGUMENTS.json \
     ROADMAP.md \
     --format markdown \
     --output /tmp/migration-plan-$ARGUMENTS.md
   ```
   (Replace `OLD` with the value from step 3.)

7. **Show the user the migration plan summary** before launching Phase 1. Specifically:
   - How many `update_models` items (Phase 2 work)
   - How many `new_resources` items (Phase 3 work)
   - How many `deprecated` items (Phase 4 work)
   - Any `roadmap_gaps` flagged

   Ask the user: "Migration plan summary above. Proceed to Phase 1 (version bump)?" Wait for confirmation. Do NOT skip this checkpoint.

## Workflow execution

Once the user confirms, invoke the `api-upgrade` skill at `.claude/skills/api-upgrade/SKILL.md` and follow its 5 phases strictly:

- **Phase 1** — Infrastructure version bump (mechanical script)
- **Phase 2** — Schema updates (spawn `flashblade-resource-modifier` agent **once per `update_models` item, serially, never parallel**)
- **Phase 3** — New resources (spawn `flashblade-resource-implementor` agent **once per `new_resources` item, serially, never parallel**)
- **Phase 4** — Deprecations (inline)
- **Phase 5** — Documentation regeneration

**Hard rules — non-negotiable:**

- STOP at every `gate-N passed` checkpoint. Wait for the user to confirm before proceeding to the next phase.
- NEVER spawn two `Agent()` calls in a single message during Phases 2 and 3. Agents share the working branch; parallel spawns produce merge conflicts on `provider.go`, `ROADMAP.md`, and shared model files.
- After every agent commit, verify INDEPENDENTLY using the SHA from the agent's `**Commit:** <sha>` return field (NOT `HEAD`):
  - `make test` count delta against the snapshot (use `grep -oP 'Test count: \K\d+'`)
  - `git show <sha>` content matches the agent's claimed file list
  - Tests contain real assertions (`grep -cE 't\.(Errorf|Fatalf)'` on the test files at `<sha>`)
- If an agent returns Status: BLOCKED, resolve the blocker explicitly (the agent will name it) and re-spawn. Do NOT skip the resource.
- Use `sonnet` model for both sub-agent types per the global CLAUDE.md cost rules (pattern-following, not Opus complexity).
- Every commit (orchestrator or agent) MUST use `git commit --no-verify`.

## Failure recovery

If something goes wrong mid-upgrade:

- Each agent commits atomically — the branch state at any HEAD is consistent with what has been validated so far.
- To abort the upgrade entirely: `git checkout main && git branch -D test/api-upgrade-$ARGUMENTS`. The main branch is untouched.
- To roll back one phase: identify the commit range with `git log --oneline` and `git reset --hard <commit-before-phase>`. Re-run the phase from a clean state.
- Do NOT use `git reset --hard` while agents are still running.

## Post-upgrade

After Gate 5 passes:

- The branch `test/api-upgrade-$ARGUMENTS` contains the full upgrade.
- Run a final full validation: `make build && make test && make lint && make docs`.
- Open a PR for review — do NOT merge to main directly.
- After PR is merged: delete the upgrade branch.
