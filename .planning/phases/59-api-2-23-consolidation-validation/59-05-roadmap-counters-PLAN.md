---
phase: 59-api-2-23-consolidation-validation
plan: 05
type: execute
wave: 3
depends_on:
  - 59-01
  - 59-02
  - 59-03
  - 59-04
files_modified:
  - ROADMAP.md
autonomous: true
requirements:
  - VALID-06

must_haves:
  truths:
    - "ROADMAP.md (repo root) lists flashblade_workload as a Done resource with Data Source = Yes"
    - "ROADMAP.md lists flashblade_resiliency_group and flashblade_resiliency_group_member as DS-only Done entries"
    - "Header counters (Covered, Coverage %) reflect API 2.23 deltas"
    - "Last updated date = current session date (2026-05-20 or later)"
    - "Provider version footer = v2.23.0"
    - "API version in header = 2.23"
  artifacts:
    - path: "ROADMAP.md"
      provides: "Root API coverage roadmap, refreshed for v2.23.0"
      contains: "Provider version: v2.23.0, 2.23"
  key_links:
    - from: "ROADMAP.md counters"
      to: "Implemented section rows"
      via: "manual count of Done resources/data sources"
      pattern: "Covered: N | Coverage of IaC-relevant CRUD: M%"
---

<objective>
Refresh the root `ROADMAP.md` (API coverage tracker, NOT `.planning/ROADMAP.md`) to reflect API 2.23 deltas: updated counters, "Last updated" date, "Provider version: v2.23.0", "API version: 2.23". Cross-check that workload, resiliency_group, resiliency_group_member are listed in the right Implemented subsections.

Purpose: Cover VALID-06.
Output: Updated `ROADMAP.md` committed.
</objective>

<execution_context>
@$HOME/.claude/get-shit-done/workflows/execute-plan.md
@$HOME/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@ROADMAP.md
@.planning/REQUIREMENTS.md
@.planning/phases/59-api-2-23-consolidation-validation/59-01-RETRO.md
@CLAUDE.md
</context>

<tasks>

<task type="auto">
  <name>Task 1: Recount Implemented resources/data sources and refresh ROADMAP.md header + footer</name>
  <files>ROADMAP.md</files>
  <read_first>
    - ROADMAP.md (current header, all Implemented subsections, footer)
    - .planning/REQUIREMENTS.md (confirm v2.23.0 deltas: +1 resource workload, +3 data sources workload/resiliency_group/resiliency_group_member)
    - .planning/STATE.md (Performance Metrics line: "Resources / Data sources: 54 / 40 pre-API-2.23. Expected delta on merge: +1 resource (workload), +3 data sources")
  </read_first>
  <action>
    1. Audit current ROADMAP.md state:
       - Confirm `flashblade_workload` row exists in `## Implemented > Storage` with Data Source = Yes, Status = Done.
       - Confirm `flashblade_resiliency_group` row exists in `## Implemented > Networking` with Resource = No, Data Source = Yes, Status = DS-only.
       - Confirm `flashblade_resiliency_group_member` row exists similarly.
       - If any missing, add the row (use existing rows as template).

    2. Recount totals using grep:
       - Resources count: `grep -cE '^\| `flashblade_' ROADMAP.md` lines where Resource column is a `flashblade_*` name (i.e. not "No").
       - Data sources count: rows where Data Source column = `Yes`.
       - Verify post-2.23 expected counts: 55 resources / 43 data sources (54+1, 40+3). Adjust counters in header line `**Total API sections:** 84 | **Covered:** ~XX | **Coverage of IaC-relevant CRUD:** ~XX%`.

    3. Update header line 6:
       - `**Last updated:** 2026-05-20 (API 2.23 upgrade: workload resource + resiliency_group/member data sources, schema v1 migrations on file_system, file_system_export, nfs/smb/qos policies)`

    4. Confirm line 7: `**Provider version:** v2.23.0` (already present per current file — keep).

    5. Confirm line 3: `FlashBlade® REST API v2.22 (Purity//FB 4.6.7) coverage status...` — UPDATE to `FlashBlade® REST API v2.23 (Purity//FB 4.6.7+) coverage status for terraform-provider-mica.` since API version bumped (API-01).

    6. Spot-check: in `## Implemented > Storage` row for `flashblade_file_system`, ensure Notes mentions `gained workload field (API 2.23, schema v1)` — already present. Same for `file_system_export` (workload back-reference (API 2.23, schema v2)) — verify. Same for qos_policy (schema v2 adds computed context field (API 2.23)) — verify.

    7. Verify no remaining `v2.22` strings in version-bumped contexts: `grep -n 'API v2\.22' ROADMAP.md` — fix any stragglers in the file header (legacy "shipped v2.22.x" milestone strings are fine, only API version line must reflect 2.23).

    Commit:
    ```
    git add ROADMAP.md
    git commit --no-verify -m "docs(roadmap): refresh API coverage counters for v2.23.0 (Phase 59 VALID-06)"
    ```
  </action>
  <verify>
    <automated>cd /home/gule/Workspace/team-infrastructure/terraform-provider-mica && grep -c '^\*\*Provider version:\*\* v2.23.0' ROADMAP.md | grep -E '^1$' && grep -c 'FlashBlade® REST API v2.23' ROADMAP.md | grep -E '^1$' && grep -c 'flashblade_workload' ROADMAP.md && grep -c 'flashblade_resiliency_group' ROADMAP.md && git diff --exit-code -- ROADMAP.md</automated>
  </verify>
  <acceptance_criteria>
    - `grep -c '^\*\*Provider version:\*\* v2.23.0' ROADMAP.md` returns 1
    - `grep -c 'FlashBlade® REST API v2.23' ROADMAP.md` returns ≥ 1
    - `grep 'flashblade_workload' ROADMAP.md` returns hits in `## Implemented > Storage`
    - `grep 'flashblade_resiliency_group' ROADMAP.md` returns 2 entries (resiliency_group + resiliency_group_member) in `## Implemented > Networking`
    - `Last updated` line includes API 2.23 mention
    - `git diff --exit-code -- ROADMAP.md` returns 0 (changes committed)
  </acceptance_criteria>
  <done>ROADMAP.md reflects API 2.23 state. Counters, version, date all updated and committed.</done>
</task>

</tasks>

<verification>
- ROADMAP.md header reports API 2.23 and Provider v2.23.0
- New resources/data sources listed in Implemented sections
- Counters consistent with: 55 resources (54+1), 43 data sources (40+3)
- Changes committed in single docs(roadmap) commit
</verification>

<success_criteria>
- VALID-06: ROADMAP.md counters refreshed, footer date + provider version bumped to v2.23.0, API version = 2.23
</success_criteria>

<output>
After completion, create `.planning/phases/59-api-2-23-consolidation-validation/59-05-SUMMARY.md` with: counter deltas applied, lines changed, commit hash.
</output>
