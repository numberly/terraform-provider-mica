---
phase: 59-api-2-23-consolidation-validation
plan: 03
type: execute
wave: 2
depends_on:
  - 59-02
files_modified:
  - pulumi/provider/cmd/pulumi-resource-flashblade/schema.json
  - pulumi/provider/cmd/pulumi-resource-flashblade/bridge-metadata.json
  - pulumi/provider/cmd/pulumi-resource-flashblade/schema-embed.json
  - .planning/phases/59-api-2-23-consolidation-validation/59-03-BRIDGE.md
autonomous: true
requirements:
  - VALID-05

must_haves:
  truths:
    - "make tfgen runs successfully in pulumi/provider/"
    - "git diff --exit-code on pulumi/ is clean after tfgen"
    - "Pulumi schema.json + bridge-metadata.json reflect API 2.23 resources (workload, resiliency_group, resiliency_group_member)"
  artifacts:
    - path: ".planning/phases/59-api-2-23-consolidation-validation/59-03-BRIDGE.md"
      provides: "tfgen outcome report"
      contains: "make tfgen, git diff --exit-code"
  key_links:
    - from: "make tfgen"
      to: "pulumi/provider/cmd/pulumi-resource-flashblade/schema.json"
      via: "tfgen schema regeneration"
      pattern: "git diff --exit-code clean"
---

<objective>
Run `make tfgen` in `pulumi/provider/` and verify `git diff --exit-code` is clean for the entire `pulumi/` tree. If diff is non-empty, the committed schema artefacts drifted from the source — commit the regen result.

Purpose: Cover VALID-05 — the CI gate that ensures committed Pulumi bridge metadata stays in sync with the underlying Terraform provider.
Output: `.planning/phases/59-api-2-23-consolidation-validation/59-03-BRIDGE.md`
</objective>

<execution_context>
@$HOME/.claude/get-shit-done/workflows/execute-plan.md
@$HOME/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/STATE.md
@.planning/REQUIREMENTS.md
@CLAUDE.md
</context>

<tasks>

<task type="auto">
  <name>Task 1: Run make tfgen and verify clean diff</name>
  <files>pulumi/provider/cmd/pulumi-resource-flashblade/schema.json, pulumi/provider/cmd/pulumi-resource-flashblade/bridge-metadata.json, pulumi/provider/cmd/pulumi-resource-flashblade/schema-embed.json, .planning/phases/59-api-2-23-consolidation-validation/59-03-BRIDGE.md</files>
  <read_first>
    - pulumi/provider/Makefile (or GNUmakefile in pulumi/provider/) to confirm tfgen target
    - .planning/STATE.md (decision: schema.json + bridge-metadata.json committed; CI gate via git diff --exit-code after make tfgen)
  </read_first>
  <action>
    From repo root on branch `test/api-upgrade-2.23`:

    1. Pre-flight: `cd pulumi/provider && ls Makefile GNUmakefile 2>/dev/null` to confirm Makefile target location. Inspect with `grep -n '^tfgen:' pulumi/provider/Makefile pulumi/provider/GNUmakefile 2>/dev/null` to confirm the target exists.

    2. Capture baseline: `git status --porcelain pulumi/ > /tmp/phase59-pulumi-pre.txt`. Should be empty (clean working tree).

    3. Run: `cd pulumi/provider && make tfgen 2>&1 | tee /tmp/phase59-tfgen.log`. Capture exit code.

    4. Capture post-state: `git status --porcelain pulumi/ > /tmp/phase59-pulumi-post.txt` and `git diff --stat pulumi/`.

    5. Write `.planning/phases/59-api-2-23-consolidation-validation/59-03-BRIDGE.md`:

       ```markdown
       # Phase 59 — Pulumi Bridge Check

       ## make tfgen
       - Working dir: pulumi/provider/
       - Exit code: <N>
       - Duration: <approx>
       - Status: PASS|FAIL

       ## git diff --exit-code -- pulumi/ (post-tfgen)
       - Modified files: <list>
       - Diff stat: <output>
       - Status: CLEAN|DRIFTED

       ## Resolution
       <if drifted: commit hash of regen commit; if clean: 'no drift, nothing to commit'>
       ```

    6. If `git diff --exit-code -- pulumi/` returns non-zero after tfgen: stage the modified files and commit:

       ```
       git add pulumi/
       git commit --no-verify -m "chore(pulumi): regenerate bridge artefacts after make tfgen (Phase 59 VALID-05)"
       ```

    7. Final verification: re-run `cd pulumi/provider && make tfgen && cd ../.. && git diff --exit-code -- pulumi/` — must return 0 with no further changes (idempotent).
  </action>
  <verify>
    <automated>cd /home/gule/Workspace/team-infrastructure/terraform-provider-mica/pulumi/provider && make tfgen && cd /home/gule/Workspace/team-infrastructure/terraform-provider-mica && git diff --exit-code -- pulumi/</automated>
  </verify>
  <acceptance_criteria>
    - `make tfgen` exits 0
    - After tfgen, `git diff --exit-code -- pulumi/` returns 0
    - `git status --porcelain pulumi/` is empty
    - Re-running tfgen is idempotent (second run produces no diff)
    - BRIDGE.md exists with PASS + CLEAN status
  </acceptance_criteria>
  <done>Pulumi bridge artefacts are in sync with the Terraform provider. Idempotent tfgen confirmed. BRIDGE.md records outcome.</done>
</task>

</tasks>

<verification>
- `cd pulumi/provider && make tfgen` followed by `git diff --exit-code -- pulumi/` returns 0
- BRIDGE.md reports CLEAN status
</verification>

<success_criteria>
- VALID-05: `make tfgen` is idempotent; no Pulumi schema drift
- Any regen output committed in single chore() commit
</success_criteria>

<output>
After completion, create `.planning/phases/59-api-2-23-consolidation-validation/59-03-SUMMARY.md` with: tfgen exit, diff result, whether a regen commit was needed.
</output>
