---
phase: 60-v2-23-0-release
plan: 05
type: execute
wave: 5
depends_on: ["60-04"]
files_modified:
  - GNUmakefile
autonomous: false
requirements:
  - RELEASE-06
  - RELEASE-07
must_haves:
  truths:
    - "TEST_BASELINE in GNUmakefile reflects the post-merge actual test count on main"
    - "make test passes on main with the new baseline"
    - "Milestone v2.23.0 is archived under .planning/milestones/"
    - "STATE.md and ROADMAP.md reflect milestone completion"
  artifacts:
    - path: "GNUmakefile"
      provides: "Updated TEST_BASELINE"
      contains: "TEST_BASELINE="
    - path: ".planning/milestones/v2.23.0-ROADMAP.md"
      provides: "Archived milestone roadmap snapshot"
      contains: "v2.23.0"
  key_links:
    - from: "GNUmakefile TEST_BASELINE"
      to: "actual go test count on main"
      via: "go test ./internal/... -list '.*' | grep -c '^Test'"
      pattern: "TEST_BASELINE=[0-9]+"
    - from: "STATE.md milestone status"
      to: ".planning/milestones/v2.23.0-ROADMAP.md"
      via: "/gsd:complete-milestone command"
      pattern: "shipped"
---

<objective>
Close out the milestone: bump `TEST_BASELINE` in `GNUmakefile` to the post-merge actual test count, then archive the v2.23.0 milestone via `/gsd:complete-milestone`.

Purpose: the test baseline guards against silent test loss in future PRs and must reflect reality. Archiving the milestone marks v2.23.0 as shipped in STATE.md and ROADMAP.md, moves the snapshot under `.planning/milestones/`, and frees the working slot for the next milestone.

Output: GNUmakefile baseline bumped + committed on main; milestone archived; STATE.md updated.
</objective>

<execution_context>
@$HOME/.claude/get-shit-done/workflows/execute-plan.md
@$HOME/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/STATE.md
@.planning/ROADMAP.md
@.planning/REQUIREMENTS.md
@.planning/phases/60-v2-23-0-release/60-04-SUMMARY.md
@GNUmakefile
@CLAUDE.md
@CONVENTIONS.md
</context>

<tasks>

<task type="auto">
  <name>Task 1: Measure actual test count on main and bump TEST_BASELINE</name>
  <read_first>
    - GNUmakefile (current value TEST_BASELINE=807, line 19)
    - CONVENTIONS.md (Test Coverage section: "Baseline ... must not decrease")
  </read_first>
  <action>
    Ensure local main is clean and up to date:
    ```bash
    git checkout main
    git pull --ff-only origin main
    git status --porcelain    # must be empty
    ```

    Run the test list to get the current authoritative count (same command the Makefile uses internally):
    ```bash
    ACTUAL=$(go test ./internal/... -list '.*' 2>/dev/null | grep -c '^Test')
    echo "Actual: $ACTUAL"
    CURRENT=$(awk -F= '/^TEST_BASELINE=/{print $2}' GNUmakefile)
    echo "Current baseline: $CURRENT"
    ```

    Decision rule:
    - If `ACTUAL > CURRENT`: bump to `ACTUAL`
    - If `ACTUAL == CURRENT`: no bump needed (skip to task 2 with a note in summary)
    - If `ACTUAL < CURRENT`: STOP — `make test` would fail; investigate the regression before continuing

    Update GNUmakefile (in-place edit, preserve everything else):
    - Find line `TEST_BASELINE=<old>`
    - Replace `<old>` with `$ACTUAL`

    Verify the change:
    ```bash
    grep -E '^TEST_BASELINE=' GNUmakefile     # exactly one line, new value
    make test                                  # must exit 0, report "Test count: $ACTUAL (baseline $ACTUAL)"
    ```

    Commit on main (--no-verify per CLAUDE.md):
    ```bash
    git add GNUmakefile
    git commit --no-verify -m "chore(release): bump TEST_BASELINE to $ACTUAL after v2.23.0"
    git push origin main
    ```
  </action>
  <verify>
    <automated>NEW=$(awk -F= '/^TEST_BASELINE=/{print $2}' GNUmakefile); ACTUAL=$(go test ./internal/... -list '.*' 2>/dev/null | grep -c '^Test'); [ "$NEW" = "$ACTUAL" ] &amp;&amp; make test &gt;/dev/null 2&gt;&amp;1 &amp;&amp; git log -1 --pretty=%s | grep -E '^chore\(release\): bump TEST_BASELINE'</automated>
  </verify>
  <acceptance_criteria>
    - `awk -F= '/^TEST_BASELINE=/{print $2}' GNUmakefile` == `go test ./internal/... -list '.*' | grep -c '^Test'`
    - `make test` exit 0 with new baseline
    - HEAD commit subject matches `chore(release): bump TEST_BASELINE to [0-9]+ after v2.23.0`
    - HEAD commit touches only GNUmakefile: `git diff HEAD~1 HEAD --name-only` == `GNUmakefile`
    - Pushed: `git rev-parse main` == `git rev-parse origin/main`
    - No Co-Authored-By: `git log -1 --pretty=%B | ! grep -i 'co-authored-by'`
  </acceptance_criteria>
  <done>GNUmakefile TEST_BASELINE matches the actual go test count on main; commit pushed.</done>
</task>

<task type="checkpoint:human-action" gate="blocking">
  <name>Task 2: Archive milestone via /gsd:complete-milestone</name>
  <read_first>
    - .planning/STATE.md
    - .planning/ROADMAP.md (Active Milestone section)
    - .planning/REQUIREMENTS.md (all 33 rows for sanity)
    - .planning/phases/60-v2-23-0-release/60-04-SUMMARY.md (release URL for archive narrative)
  </read_first>
  <what-built>
    The `/gsd:complete-milestone` slash command must be invoked by the operator in the Claude Code session. It moves the milestone roadmap snapshot to `.planning/milestones/v2.23.0-ROADMAP.md`, updates STATE.md (status → shipped, last_activity, Recent Milestones list), updates the ROADMAP.md Milestones list (✅ shipped marker + date), and clears the Active Milestone slot.

    This is a slash-command-driven workflow, not a bash sequence — that is why it is `checkpoint:human-action`.
  </what-built>
  <how-to-verify>
    1. In the Claude Code session, run:
       ```
       /gsd:complete-milestone v2.23.0
       ```
    2. Follow the prompts. Provide the release URL from `.planning/phases/60-v2-23-0-release/60-04-RELEASE.md` when asked.
    3. After the command completes, verify on disk:
       ```bash
       # Archive file exists
       test -f .planning/milestones/v2.23.0-ROADMAP.md && echo OK

       # STATE.md no longer points to Phase 60 as current focus
       grep -E '^milestone: v2\.23\.0' .planning/STATE.md
       grep -E '^status: shipped|^status: complete' .planning/STATE.md

       # ROADMAP.md Milestones list shows shipped marker
       grep -E '✅ \*\*v2\.23\.0\*\*' .planning/ROADMAP.md

       # All changes committed
       git status --porcelain .planning/
       ```
    4. Push any commits the command made:
       ```bash
       git push origin main
       ```
  </how-to-verify>
  <acceptance_criteria>
    - `.planning/milestones/v2.23.0-ROADMAP.md` exists
    - `.planning/STATE.md` either changes milestone status to shipped/complete OR rotates to a new active milestone
    - `.planning/ROADMAP.md` Milestones list shows `✅ **v2.23.0**` with a `shipped <date>` annotation
    - REQUIREMENTS.md (or its archived snapshot) shows RELEASE-06 and RELEASE-07 as Complete
    - Working tree under `.planning/` is clean
    - `git rev-parse main` == `git rev-parse origin/main`
  </acceptance_criteria>
  <resume-signal>Type "archived" once `/gsd:complete-milestone v2.23.0` completed and the verification commands all return expected output.</resume-signal>
  <done>Milestone archived: snapshot under `.planning/milestones/`, STATE.md + ROADMAP.md reflect shipped status, all planning changes pushed to main.</done>
</task>

</tasks>

<verification>
- TEST_BASELINE in GNUmakefile matches actual on-main test count
- `make test` on main exit 0 with new baseline
- `.planning/milestones/v2.23.0-ROADMAP.md` exists
- STATE.md + ROADMAP.md reflect milestone shipped
- All commits pushed to origin/main
</verification>

<success_criteria>
- `grep -E '^TEST_BASELINE=[0-9]+$' GNUmakefile` returns the bumped value
- `test -f .planning/milestones/v2.23.0-ROADMAP.md` is true
- `git rev-parse main` == `git rev-parse origin/main`
- All 7 RELEASE-* requirements are Complete in the archived requirements snapshot
</success_criteria>

<output>
After completion, create `.planning/phases/60-v2-23-0-release/60-05-SUMMARY.md` with:
- Old baseline → new baseline
- Archive path
- Final commit hashes
- Phase 60 closure declaration
</output>
