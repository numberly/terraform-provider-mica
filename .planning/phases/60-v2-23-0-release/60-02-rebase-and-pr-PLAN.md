---
phase: 60-v2-23-0-release
plan: 02
type: execute
wave: 2
depends_on: ["60-01"]
files_modified: []
autonomous: false
requirements:
  - RELEASE-02
  - RELEASE-03
must_haves:
  truths:
    - "test/api-upgrade-2.23 is rebased onto latest origin/main with no conflicts"
    - "Branch is force-pushed to origin with --force-with-lease (no unrelated overwrites)"
    - "PR test/api-upgrade-2.23 -> main exists and CI is green"
  artifacts:
    - path: ".planning/phases/60-v2-23-0-release/60-02-PR.md"
      provides: "PR URL + CI status snapshot"
      contains: "https://github.com/numberly/terraform-provider-mica/pull/"
  key_links:
    - from: "local test/api-upgrade-2.23"
      to: "origin/test/api-upgrade-2.23"
      via: "git push --force-with-lease"
      pattern: "force-with-lease"
    - from: "PR head branch"
      to: ".github/workflows/ci.yml"
      via: "GitHub Actions trigger on PR"
      pattern: "CI Gate"
---

<objective>
Rebase `test/api-upgrade-2.23` onto the latest `main`, force-push, open the release PR, and wait for CI green.

Purpose: deliver the release branch to `main` cleanly. CI must run on the rebased branch (not the stale tip) so the release tag gate downstream is honest.

Output: PR URL recorded, CI green, branch ready to merge.
</objective>

<execution_context>
@$HOME/.claude/get-shit-done/workflows/execute-plan.md
@$HOME/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/ROADMAP.md
@.planning/REQUIREMENTS.md
@CLAUDE.md
@CONVENTIONS.md
@.github/workflows/release.yml
@.planning/phases/60-v2-23-0-release/60-01-SUMMARY.md
</context>

<tasks>

<task type="checkpoint:decision" gate="blocking">
  <name>Task 1: Confirm rebase strategy and authorize force-push</name>
  <read_first>
    - CLAUDE.md (rule: "always rebase the branch on the target branch before pushing")
    - .planning/phases/60-v2-23-0-release/60-01-SUMMARY.md (confirm CHANGELOG commit landed)
  </read_first>
  <decision>Authorize rebase of `test/api-upgrade-2.23` onto `origin/main` followed by `git push --force-with-lease`.</decision>
  <context>
    The branch is ~25+ commits ahead of `main`. Rebase will rewrite all of those commits' hashes. `--force-with-lease` protects against overwriting unseen changes on the remote. This is a non-destructive operation in principle, but it does rewrite published history on a feature branch — confirm there are no other collaborators with local checkouts to avoid surprise.
  </context>
  <options>
    <option id="rebase">
      <name>Rebase + force-with-lease (recommended)</name>
      <pros>Linear history, clean PR diff, matches CLAUDE.md convention</pros>
      <cons>Rewrites commit hashes on the published feature branch</cons>
    </option>
    <option id="merge-main">
      <name>Merge main into the feature branch instead</name>
      <pros>No history rewrite</pros>
      <cons>Adds merge commit noise; violates CLAUDE.md "always rebase before pushing"</cons>
    </option>
  </options>
  <resume-signal>Select: rebase or merge-main</resume-signal>
  <done>Operator authorized rebase strategy.</done>
</task>

<task type="auto">
  <name>Task 2: Rebase onto origin/main and force-push</name>
  <read_first>
    - .planning/phases/60-v2-23-0-release/60-01-SUMMARY.md
  </read_first>
  <action>
    Verify current branch and clean tree:
    ```bash
    git rev-parse --abbrev-ref HEAD          # must print: test/api-upgrade-2.23
    git status --porcelain                   # must be empty
    ```

    Fetch latest and rebase (per CLAUDE.md — git rebase, never -i, never --no-edit):
    ```bash
    git fetch origin main
    git rebase origin/main
    ```

    If conflicts arise: STOP, report conflicts to operator, do NOT auto-resolve. Conflict resolution requires human decision (the merge content is product-relevant).

    On clean rebase, force-push:
    ```bash
    git push --force-with-lease origin test/api-upgrade-2.23
    ```

    Then verify the remote tip matches local:
    ```bash
    git fetch origin test/api-upgrade-2.23
    [ "$(git rev-parse HEAD)" = "$(git rev-parse origin/test/api-upgrade-2.23)" ] && echo OK
    ```
  </action>
  <verify>
    <automated>git rev-parse HEAD &gt; /tmp/local_head &amp;&amp; git fetch origin test/api-upgrade-2.23 &amp;&amp; git rev-parse origin/test/api-upgrade-2.23 &gt; /tmp/remote_head &amp;&amp; diff /tmp/local_head /tmp/remote_head &amp;&amp; git log origin/main..HEAD --oneline | head -1</automated>
  </verify>
  <acceptance_criteria>
    - `git rev-parse HEAD` == `git rev-parse origin/test/api-upgrade-2.23`
    - `git log origin/main..HEAD --oneline | wc -l` ≥ 1 (branch has commits ahead of main)
    - `git log HEAD..origin/main --oneline | wc -l` == 0 (no commits on main not on branch — i.e. rebased)
    - Working tree clean: `git status --porcelain` empty
    - CHANGELOG commit present: `git log origin/main..HEAD --oneline | grep -q 'docs(release): add v2.23.0 CHANGELOG'`
  </acceptance_criteria>
  <done>`test/api-upgrade-2.23` is rebased on `origin/main` and the rebased tip is pushed to origin via --force-with-lease.</done>
</task>

<task type="auto">
  <name>Task 3: Open PR and record URL</name>
  <read_first>
    - CHANGELOG.md (use the [2.23.0] section as the PR body source)
    - .planning/REQUIREMENTS.md (RELEASE/VALID rows for body checklist)
  </read_first>
  <action>
    Open the PR via `gh` CLI. Title and body MUST be in English. NO Co-Authored-By, NO AI signature.

    ```bash
    gh pr create \
      --base main \
      --head test/api-upgrade-2.23 \
      --title "release: v2.23.0 — FlashBlade API 2.23 upgrade" \
      --body "$(cat <<'EOF'
    ## Summary

    Ships v2.23.0: FlashBlade API 2.23 support — new `flashblade_workload` resource, three new data sources (`workload`, `resiliency_group`, `resiliency_group_member`), and `workload` field schema v0→v1 migrations on six existing resources. Pulumi bridge artefacts regenerated.

    Full release notes: see `CHANGELOG.md` `[2.23.0]` section.

    ## Scope

    - **New resource:** `flashblade_workload`
    - **New data sources:** `flashblade_workload`, `flashblade_resiliency_group`, `flashblade_resiliency_group_member`
    - **Schema migrations (v0→v1):** `file_system`, `file_system_export`, `nfs_export_policy`, `smb_client_policy`, `smb_share_policy`, `qos_policy` (qos_policy also gets a v1→v2 for the `context` field)
    - **Pulumi bridge:** `schema.json`, `bridge-metadata.json`, `schema-embed.json` regenerated for API 2.23

    ## Validation

    - `make test` — clean, 807+ tests
    - `make lint` — 0 issues
    - `make docs` — idempotent (`git diff --exit-code docs/` clean)
    - `make tfgen` — idempotent (`git diff --exit-code pulumi/` clean)
    - Live acceptance on FlashBlade arrays `par5` and `pa7` (see `.planning/phases/59-*/59-HUMAN-UAT.md`)

    ## Test plan

    - [ ] CI Gate (release.yml `ci` job) green on this PR
    - [ ] Reviewer confirms CHANGELOG entry is accurate
    - [ ] After merge: tag `v2.23.0` triggers GoReleaser (handled in milestone Phase 60-04)

    ## Out of scope

    Pulumi SDK regen + publish — handled in a separate `pulumi-2.23.0` milestone.
    EOF
    )"
    ```

    Capture the URL:
    ```bash
    PR_URL=$(gh pr view test/api-upgrade-2.23 --json url --jq .url)
    PR_NUMBER=$(gh pr view test/api-upgrade-2.23 --json number --jq .number)
    ```

    Write `.planning/phases/60-v2-23-0-release/60-02-PR.md`:
    ```markdown
    ---
    plan: 60-02
    pr_url: <PR_URL>
    pr_number: <PR_NUMBER>
    head: test/api-upgrade-2.23
    base: main
    opened: <ISO date>
    ---

    # PR: release v2.23.0

    URL: <PR_URL>
    Number: <PR_NUMBER>

    ## CI status snapshot

    (updated by task 4)
    ```

    Commit the PR record file:
    ```bash
    git add .planning/phases/60-v2-23-0-release/60-02-PR.md
    git commit --no-verify -m "docs(release): record v2.23.0 PR URL"
    git push origin test/api-upgrade-2.23
    ```
  </action>
  <verify>
    <automated>test -f .planning/phases/60-v2-23-0-release/60-02-PR.md &amp;&amp; grep -E 'https://github.com/numberly/terraform-provider-mica/pull/[0-9]+' .planning/phases/60-v2-23-0-release/60-02-PR.md &amp;&amp; gh pr view test/api-upgrade-2.23 --json state --jq .state | grep -E '^OPEN$'</automated>
  </verify>
  <acceptance_criteria>
    - PR exists and is OPEN: `gh pr view test/api-upgrade-2.23 --json state --jq .state` == `OPEN`
    - PR base is `main`: `gh pr view test/api-upgrade-2.23 --json baseRefName --jq .baseRefName` == `main`
    - `.planning/phases/60-v2-23-0-release/60-02-PR.md` exists with a valid GitHub PR URL
    - Last commit subject == `docs(release): record v2.23.0 PR URL`
  </acceptance_criteria>
  <done>PR opened against main, URL recorded in 60-02-PR.md, file committed and pushed.</done>
</task>

<task type="checkpoint:human-verify" gate="blocking">
  <name>Task 4: Wait for CI green</name>
  <read_first>
    - .planning/phases/60-v2-23-0-release/60-02-PR.md (for PR number)
    - .github/workflows/ci.yml (understand what CI checks run)
  </read_first>
  <what-built>
    PR opened in task 3. CI runs automatically on push. This checkpoint waits for the CI Gate workflow (and any other required checks) to succeed before authorizing merge.
  </what-built>
  <how-to-verify>
    Poll CI status:
    ```bash
    gh pr checks test/api-upgrade-2.23
    ```
    Or watch:
    ```bash
    gh pr checks test/api-upgrade-2.23 --watch
    ```

    All required checks MUST be green (state: SUCCESS or CONCLUSION: success). If any fail:
    1. Inspect failure: `gh run view <run-id> --log-failed`
    2. STOP this plan — fix on branch (separate commit), push, restart task 4
    3. Do NOT proceed to merge with a red CI

    On all-green, append to `.planning/phases/60-v2-23-0-release/60-02-PR.md`:
    ```
    ## CI status snapshot (final)

    | Check | Status | URL |
    |-------|--------|-----|
    | <name> | success | <run url> |

    Captured: <ISO timestamp>
    ```

    Commit:
    ```bash
    git add .planning/phases/60-v2-23-0-release/60-02-PR.md
    git commit --no-verify -m "docs(release): record CI green on v2.23.0 PR"
    git push origin test/api-upgrade-2.23
    ```
  </how-to-verify>
  <acceptance_criteria>
    - `gh pr checks test/api-upgrade-2.23 --json state,conclusion --jq '.[] | select(.state != "COMPLETED" or .conclusion != "SUCCESS")'` returns nothing (all checks completed and successful)
    - 60-02-PR.md contains a "CI status snapshot (final)" section
    - HEAD commit subject == `docs(release): record CI green on v2.23.0 PR`
  </acceptance_criteria>
  <resume-signal>Type "ci-green" to confirm all required checks passed, or "ci-failed: <run-url>" to halt and triage.</resume-signal>
  <done>CI is green on the PR head commit. Branch is mergeable.</done>
</task>

</tasks>

<verification>
- Branch is rebased on origin/main with no merge commits ahead of main
- PR is OPEN, targets main, head is test/api-upgrade-2.23
- All CI required checks are SUCCESS
- 60-02-PR.md records PR URL + CI snapshot
</verification>

<success_criteria>
- `gh pr view test/api-upgrade-2.23 --json state,mergeable --jq '.state + "/" + (.mergeable // "?")'` returns `OPEN/MERGEABLE`
- 60-02-PR.md exists and contains both PR URL and CI snapshot sections
</success_criteria>

<output>
After completion, create `.planning/phases/60-v2-23-0-release/60-02-SUMMARY.md` with:
- PR URL + number
- Rebase outcome (any conflicts encountered)
- CI run URLs
- Hand-off to plan 60-03 (merge)
</output>
