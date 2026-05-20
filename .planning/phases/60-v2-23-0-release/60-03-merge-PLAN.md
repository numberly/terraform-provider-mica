---
phase: 60-v2-23-0-release
plan: 03
type: execute
wave: 3
depends_on: ["60-02"]
files_modified: []
autonomous: false
requirements:
  - RELEASE-04
must_haves:
  truths:
    - "PR test/api-upgrade-2.23 -> main is merged"
    - "main HEAD includes the CHANGELOG v2.23.0 commit"
  artifacts:
    - path: ".planning/phases/60-v2-23-0-release/60-03-MERGE.md"
      provides: "Merge commit hash + timestamp"
      contains: "merge_sha:"
  key_links:
    - from: "PR head test/api-upgrade-2.23"
      to: "main branch tip"
      via: "gh pr merge --squash"
      pattern: "squash"
---

<objective>
Merge the v2.23.0 PR into `main` using the project's standard merge strategy.

Purpose: bring the release branch onto `main` so the v2.23.0 tag can be created on the canonical history. This is the point of no return — once merged, the release goes live as soon as the tag is pushed.

Output: PR merged, merge commit recorded, main pulled locally.
</objective>

<execution_context>
@$HOME/.claude/get-shit-done/workflows/execute-plan.md
@$HOME/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/ROADMAP.md
@.planning/REQUIREMENTS.md
@.planning/phases/60-v2-23-0-release/60-02-SUMMARY.md
@.planning/phases/60-v2-23-0-release/60-02-PR.md
@CLAUDE.md
</context>

<tasks>

<task type="checkpoint:decision" gate="blocking">
  <name>Task 1: Select merge strategy</name>
  <read_first>
    - .planning/phases/60-v2-23-0-release/60-02-PR.md
    - CLAUDE.md (commit conventions)
    - Recent merge commits on main: `git log origin/main --merges --oneline | head -10`
  </read_first>
  <decision>Choose merge strategy for the v2.23.0 PR.</decision>
  <context>
    The branch has ~25+ commits of granular history (per-phase, per-skill, per-resource). Preserve granular history for archaeology, or squash to a single release commit?

    Recent main history pattern should drive this choice — inspect with `git log origin/main --oneline -30` before deciding.
  </context>
  <options>
    <option id="merge">
      <name>Merge commit (--merge)</name>
      <pros>Preserves all granular commit messages on main; bisect-friendly per-feature</pros>
      <cons>Adds a merge commit; main history becomes non-linear</cons>
    </option>
    <option id="squash">
      <name>Squash merge (--squash)</name>
      <pros>One clean release commit on main; aligned with linear-history convention if the project uses it</pros>
      <cons>Loses per-commit attribution on main (still preserved on the feature branch ref / PR)</cons>
    </option>
    <option id="rebase">
      <name>Rebase merge (--rebase)</name>
      <pros>Linear history AND preserves per-commit granularity</pros>
      <cons>Rewrites commit dates; may break GoReleaser changelog grouping if commits aren't conventional-style</cons>
    </option>
  </options>
  <resume-signal>Select: merge, squash, or rebase</resume-signal>
  <done>Operator authorized merge strategy.</done>
</task>

<task type="auto">
  <name>Task 2: Merge PR and pull main locally</name>
  <read_first>
    - .planning/phases/60-v2-23-0-release/60-02-PR.md (PR number / URL)
    - Decision from task 1
  </read_first>
  <action>
    Verify PR still green before merging:
    ```bash
    gh pr checks test/api-upgrade-2.23
    gh pr view test/api-upgrade-2.23 --json mergeable,mergeStateStatus
    ```

    Merge using the strategy chosen in task 1. Replace `<STRATEGY>` with one of `--merge`, `--squash`, `--rebase`:

    ```bash
    gh pr merge test/api-upgrade-2.23 <STRATEGY> --delete-branch=false
    ```

    NOTE: Do NOT delete the remote branch yet. It is harmless to leave it in place; deleting is reversible but premature deletion makes hot-fix workflows harder. The remote feature branch can be deleted manually after the tag is published (Plan 60-04 verifies tag).

    Switch to main, pull, and verify the merge:
    ```bash
    git checkout main
    git pull --ff-only origin main
    git log -1 --pretty='%H %s'
    ```

    Capture merge SHA:
    ```bash
    MERGE_SHA=$(git rev-parse HEAD)
    PR_NUM=$(grep '^pr_number:' .planning/phases/60-v2-23-0-release/60-02-PR.md | awk '{print $2}')
    ```

    Write `.planning/phases/60-v2-23-0-release/60-03-MERGE.md`:
    ```markdown
    ---
    plan: 60-03
    pr_number: <PR_NUM>
    merge_sha: <MERGE_SHA>
    merge_strategy: <STRATEGY>
    merged_at: <ISO timestamp>
    ---

    # Merge: v2.23.0 -> main

    Merge commit: <MERGE_SHA>
    Strategy: <STRATEGY>
    PR: <PR URL>

    main tip:
    ```
    git log -1 --pretty='%H %an %s' on main
    ```
    ```

    Commit on `main` (the planning doc only — release-relevant content already merged):
    ```bash
    git add .planning/phases/60-v2-23-0-release/60-03-MERGE.md
    git commit --no-verify -m "docs(release): record v2.23.0 merge sha"
    git push origin main
    ```
  </action>
  <verify>
    <automated>git rev-parse --abbrev-ref HEAD | grep -E '^main$' &amp;&amp; gh pr view test/api-upgrade-2.23 --json state --jq .state | grep -E '^MERGED$' &amp;&amp; test -f .planning/phases/60-v2-23-0-release/60-03-MERGE.md &amp;&amp; grep -E '^merge_sha: [0-9a-f]{7,}' .planning/phases/60-v2-23-0-release/60-03-MERGE.md</automated>
  </verify>
  <acceptance_criteria>
    - `gh pr view test/api-upgrade-2.23 --json state --jq .state` == `MERGED`
    - Local branch is `main`: `git rev-parse --abbrev-ref HEAD` == `main`
    - main HEAD includes the v2.23.0 work: `git log --oneline | head -50 | grep -qE 'docs\(release\): add v2\.23\.0 CHANGELOG'` (squash: search in merge message instead)
    - `head -1 CHANGELOG.md` matches `^## \[2\.23\.0\]` on main
    - `.planning/phases/60-v2-23-0-release/60-03-MERGE.md` exists with merge_sha frontmatter
    - main is in sync with remote: `git rev-parse main` == `git rev-parse origin/main`
  </acceptance_criteria>
  <done>PR merged to main with chosen strategy, merge SHA recorded, local main checked out and synced with origin.</done>
</task>

</tasks>

<verification>
- PR state is MERGED
- main HEAD contains the v2.23.0 CHANGELOG entry
- 60-03-MERGE.md records merge SHA and strategy
- Local main == origin/main
</verification>

<success_criteria>
- `gh pr view test/api-upgrade-2.23 --json state --jq .state` returns `MERGED`
- `head -1 CHANGELOG.md` on main matches `^## \[2\.23\.0\]`
- Working tree clean on main
</success_criteria>

<output>
After completion, create `.planning/phases/60-v2-23-0-release/60-03-SUMMARY.md` with merge SHA, strategy, and hand-off to plan 60-04 (tag + release).
</output>
