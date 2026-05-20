---
phase: 59-api-2-23-consolidation-validation
plan: 02
type: execute
wave: 1
depends_on: []
files_modified:
  - GNUmakefile
  - docs/
  - .planning/phases/59-api-2-23-consolidation-validation/59-02-CHECKS.md
autonomous: true
requirements:
  - VALID-01
  - VALID-02
  - VALID-03

must_haves:
  truths:
    - "make test exits 0 on test/api-upgrade-2.23"
    - "make lint reports 0 issues"
    - "make docs produces no uncommitted changes in docs/"
    - "TEST_BASELINE in GNUmakefile is consistent with current test count (≥ baseline)"
  artifacts:
    - path: ".planning/phases/59-api-2-23-consolidation-validation/59-02-CHECKS.md"
      provides: "Captured outputs of make test / lint / docs"
      contains: "make test, make lint, make docs"
  key_links:
    - from: "make test"
      to: "TEST_BASELINE in GNUmakefile"
      via: "shell guard in GNUmakefile:23-27"
      pattern: "actual >= TEST_BASELINE"
---

<objective>
Run `make test`, `make lint`, and `make docs` on branch `test/api-upgrade-2.23`. Capture outputs in CHECKS.md. If any check fails or `make docs` produces uncommitted changes under `docs/`, fix the offending source and re-run. Do NOT bump TEST_BASELINE to the new count (reserved for Phase 60 RELEASE-06); only ensure current test count ≥ existing baseline (807).

Purpose: Cover VALID-01..03 — the green-build acceptance gates for the upgrade.
Output: `.planning/phases/59-api-2-23-consolidation-validation/59-02-CHECKS.md`, fixed source (if needed), regenerated `docs/` (if needed, committed).
</objective>

<execution_context>
@$HOME/.claude/get-shit-done/workflows/execute-plan.md
@$HOME/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/STATE.md
@.planning/REQUIREMENTS.md
@GNUmakefile
@CLAUDE.md
@CONVENTIONS.md
</context>

<tasks>

<task type="auto">
  <name>Task 1: Run make test and make lint, capture outputs</name>
  <files>.planning/phases/59-api-2-23-consolidation-validation/59-02-CHECKS.md</files>
  <read_first>
    - GNUmakefile (TEST_BASELINE value, lint target)
    - CONVENTIONS.md (test naming, baseline policy)
  </read_first>
  <action>
    From repo root on branch `test/api-upgrade-2.23`:

    1. Run `make test 2>&1 | tee /tmp/phase59-make-test.log`. Capture exit code.
    2. Extract test count from output line `Test count: N (baseline 807)`.
    3. Run `make lint 2>&1 | tee /tmp/phase59-make-lint.log`. Capture exit code.

    Write to `.planning/phases/59-api-2-23-consolidation-validation/59-02-CHECKS.md`:

    ```markdown
    # Phase 59 — Make Checks

    ## make test
    - Exit code: <N>
    - Test count: <N>
    - TEST_BASELINE (GNUmakefile): 807
    - Delta vs baseline: +<N>
    - Status: PASS|FAIL

    ## make lint
    - Exit code: <N>
    - Issues: <count or 'none'>
    - Status: PASS|FAIL
    ```

    If `make test` fails: investigate using Serena MCP (`find_symbol`, `get_references`) on the failing test's referenced symbols. Fix root cause in source — do NOT lower TEST_BASELINE. Re-run until green.

    If `make lint` fails: fix lint issues in flagged files. Re-run until 0 issues.

    Do NOT modify TEST_BASELINE in this task.
  </action>
  <verify>
    <automated>cd /home/gule/Workspace/team-infrastructure/terraform-provider-mica && make test 2>&1 | tail -5 && make lint 2>&1 | tail -5</automated>
  </verify>
  <acceptance_criteria>
    - `make test` exit code is 0
    - `make lint` exit code is 0
    - Output lines `Test count: N (baseline 807)` with N ≥ 807
    - CHECKS.md contains exit codes + test count + lint issue count
  </acceptance_criteria>
  <done>make test and make lint both green on test/api-upgrade-2.23. CHECKS.md records outcomes.</done>
</task>

<task type="auto">
  <name>Task 2: Run make docs and commit any regenerated docs</name>
  <files>docs/, .planning/phases/59-api-2-23-consolidation-validation/59-02-CHECKS.md</files>
  <read_first>
    - GNUmakefile (docs target = go generate ./...)
    - CLAUDE.md (Do NOT edit docs/ manually rule)
  </read_first>
  <action>
    From repo root:

    1. Run `git status --porcelain docs/` — record baseline (should be clean).
    2. Run `make docs 2>&1 | tee /tmp/phase59-make-docs.log`. Capture exit code.
    3. Run `git status --porcelain docs/` — capture any modifications.
    4. Run `git diff --stat docs/` to summarize.

    Append to CHECKS.md:

    ```markdown
    ## make docs
    - Exit code: <N>
    - Files modified by regen: <list from git status>
    - Diff stat: <git diff --stat output>
    - Status: PASS|FAIL
    ```

    If `docs/` has uncommitted modifications after `make docs`: this is expected for an API upgrade. Stage with `git add docs/` and commit:

    ```
    git commit --no-verify -m "docs(api-2.23): regenerate tfplugindocs after make docs"
    ```

    If `make docs` exits non-zero: investigate go generate / tfplugindocs failure; fix template or example files (NEVER edit docs/ directly).

    Final state must satisfy: `git diff --exit-code -- docs/` returns 0 AND `git status --porcelain docs/` is empty.
  </action>
  <verify>
    <automated>cd /home/gule/Workspace/team-infrastructure/terraform-provider-mica && make docs && git diff --exit-code -- docs/ && test -z "$(git status --porcelain docs/)"</automated>
  </verify>
  <acceptance_criteria>
    - `make docs` exits 0
    - `git diff --exit-code -- docs/` returns 0 (clean)
    - `git status --porcelain docs/` is empty
    - Any regen changes are committed (no manual docs/ edits)
    - CHECKS.md updated with docs section
  </acceptance_criteria>
  <done>docs/ is clean and committed. CHECKS.md reflects all three make targets green.</done>
</task>

</tasks>

<verification>
- `make test` exit code 0, test count ≥ 807
- `make lint` exit code 0, 0 issues
- `make docs` followed by `git diff --exit-code -- docs/` returns 0
- CHECKS.md exists with PASS status for all three targets
</verification>

<success_criteria>
- VALID-01: `make test` passes with count ≥ TEST_BASELINE (807)
- VALID-02: `make lint` clean
- VALID-03: `make docs` regenerated, docs/ diff (if any) committed
- No manual edits to docs/
- TEST_BASELINE unchanged (Phase 60 RELEASE-06 will bump)
</success_criteria>

<output>
After completion, create `.planning/phases/59-api-2-23-consolidation-validation/59-02-SUMMARY.md` with: test count, lint result, docs regen outcome, any source fixes applied.
</output>
