---
phase: 59-api-2-23-consolidation-validation
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - .planning/phases/59-api-2-23-consolidation-validation/59-01-RETRO.md
autonomous: true
requirements:
  - API-01
  - API-02
  - API-03
  - WORKLOAD-01
  - WORKLOAD-02
  - WORKLOAD-03
  - RESILIENCY-01
  - RESILIENCY-02
  - RESILIENCY-03
  - SCHEMA-01
  - SCHEMA-02
  - SCHEMA-03
  - SCHEMA-04
  - SCHEMA-05
  - SCHEMA-06
  - SCHEMA-07
  - SCHEMA-08
  - BRIDGE-01
  - BRIDGE-02
  - BRIDGE-03

must_haves:
  truths:
    - "Every retro REQ-ID can be traced to a concrete artifact (file, symbol, test, or commit) on branch test/api-upgrade-2.23"
    - "RETRO.md report exists and lists per-REQ-ID verification evidence"
    - "No retro REQ-ID is missing its expected artifact"
  artifacts:
    - path: ".planning/phases/59-api-2-23-consolidation-validation/59-01-RETRO.md"
      provides: "Traceability matrix for the 19 retro REQ-IDs"
      contains: "API-01, WORKLOAD-01, RESILIENCY-01, SCHEMA-01, BRIDGE-01"
  key_links:
    - from: "RETRO.md"
      to: "branch test/api-upgrade-2.23"
      via: "git log + file existence + symbol existence"
      pattern: "REQ-ID -> evidence command -> result"
---

<objective>
Walk the 19 retro REQ-IDs from Phase 59 against the actual branch state. For each REQ-ID, run a verification command (file exists, symbol exists, test exists, commit referenced is on branch) and record evidence in RETRO.md. No code changes — pure traceability.

Purpose: Produce the audit trail required by the milestone to formalize already-shipped work into GSD.
Output: `.planning/phases/59-api-2-23-consolidation-validation/59-01-RETRO.md`
</objective>

<execution_context>
@$HOME/.claude/get-shit-done/workflows/execute-plan.md
@$HOME/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/STATE.md
@.planning/ROADMAP.md
@.planning/REQUIREMENTS.md
@CLAUDE.md
@CONVENTIONS.md
</context>

<tasks>

<task type="auto">
  <name>Task 1: Verify API + BRIDGE + RESILIENCY retro REQ-IDs</name>
  <files>.planning/phases/59-api-2-23-consolidation-validation/59-01-RETRO.md</files>
  <read_first>
    - .planning/REQUIREMENTS.md (REQ-ID → commit / artifact mapping)
    - CLAUDE.md (Serena MCP mandate for symbol lookups)
  </read_first>
  <action>
    Verify on branch `test/api-upgrade-2.23` (do NOT switch branches; just run from current checkout):

    For API-01..03 (version bump):
    - API-01: `git log test/api-upgrade-2.23 --oneline | grep -E '9a9715b|API 2\.23'` returns non-empty; `rg -c '2\.23' internal/client/ | head -5` shows occurrences; `grep -c '2\.22' FLASHBLADE_API.md` may be > 0 (legacy file ok) but `rg "API.*2\.22" internal/client/transport.go` returns 0.
    - API-02: `git log test/api-upgrade-2.23 --oneline | grep bfe0ad0` non-empty; `rg "2\.22" internal/testmock/ internal/provider/*_test.go | wc -l` recorded as evidence (expected: stale strings removed).
    - API-03: `git log test/api-upgrade-2.23 --oneline | grep 19ea8dc` non-empty; `grep -n 'internal/\*\*/\*\.go' .claude/skills/api-upgrade/scripts/upgrade_version.py` returns the scan glob.

    For RESILIENCY-01..03:
    - RESILIENCY-01: Use Serena `find_symbol` for `NewResiliencyGroupDataSource` in `internal/provider/`; record file path.
    - RESILIENCY-02: Use Serena `find_symbol` for `NewResiliencyGroupMemberDataSource`; record file path.
    - RESILIENCY-03: `ls internal/testmock/handlers/resiliency_group*.go` + `rg 'TestUnit_ResiliencyGroup' internal/ -l`.

    For BRIDGE-01..03:
    - BRIDGE-01: `git log test/api-upgrade-2.23 --oneline | grep c65d063` non-empty; `ls pulumi/provider/cmd/pulumi-resource-flashblade/schema.json pulumi/provider/cmd/pulumi-resource-flashblade/bridge-metadata.json pulumi/provider/cmd/pulumi-resource-flashblade/schema-embed.json` all exist.
    - BRIDGE-02: `git log test/api-upgrade-2.23 --oneline | grep d216d24` non-empty; `rg 'TestProviderInfo_' pulumi/provider/ -l`.
    - BRIDGE-03: `git log test/api-upgrade-2.23 --oneline | grep -E '907623d|de2697a'` non-empty; `grep -n 'make tfgen' .claude/skills/api-upgrade/SKILL.md` returns hits.

    Append per-REQ-ID rows to `.planning/phases/59-api-2-23-consolidation-validation/59-01-RETRO.md` in this format:

    ```
    | REQ-ID | Evidence command | Result | Status |
    |--------|------------------|--------|--------|
    | API-01 | git log ... grep 9a9715b | <commit hash + subject> | OK |
    ```

    If any check fails, set Status = `MISSING` and add a `## Gaps` section listing the failure. Do NOT attempt to fix code in this task.
  </action>
  <verify>
    <automated>test -f .planning/phases/59-api-2-23-consolidation-validation/59-01-RETRO.md && grep -E '^\| (API|RESILIENCY|BRIDGE)-0[1-9]' .planning/phases/59-api-2-23-consolidation-validation/59-01-RETRO.md | wc -l | grep -E '^9$'</automated>
  </verify>
  <acceptance_criteria>
    - File `.planning/phases/59-api-2-23-consolidation-validation/59-01-RETRO.md` exists
    - Contains exactly 9 rows for API-01..03, RESILIENCY-01..03, BRIDGE-01..03
    - Each row has an Evidence command and a Result
    - If any row is `MISSING`, a `## Gaps` section exists
  </acceptance_criteria>
  <done>9 retro REQ-IDs (API+RESILIENCY+BRIDGE) are documented in RETRO.md with concrete evidence commands and results.</done>
</task>

<task type="auto">
  <name>Task 2: Verify WORKLOAD + SCHEMA retro REQ-IDs</name>
  <files>.planning/phases/59-api-2-23-consolidation-validation/59-01-RETRO.md</files>
  <read_first>
    - .planning/REQUIREMENTS.md
    - .planning/phases/59-api-2-23-consolidation-validation/59-01-RETRO.md (append, don't overwrite)
    - CONVENTIONS.md (state upgrader pattern, test naming)
  </read_first>
  <action>
    Append rows to the existing RETRO.md (do NOT recreate).

    For WORKLOAD-01..03:
    - WORKLOAD-01: Use Serena `find_symbol` for `NewWorkloadResource` in `internal/provider/`. Confirm resource file: `ls internal/provider/workload_resource.go internal/provider/workload_resource_test.go`. Check `rg 'TestUnit_WorkloadResource_(Lifecycle|Import|DriftDetection)' internal/provider/ -l` returns the test file.
    - WORKLOAD-02: Use Serena `find_symbol` for `NewWorkloadDataSource`. Confirm `ls internal/provider/workload_data_source.go internal/provider/workload_data_source_test.go`.
    - WORKLOAD-03: `ls internal/client/workload.go internal/client/workload_test.go internal/testmock/handlers/workload.go` all exist; `rg 'TestUnit_Workload_(Get|Post|Patch|Delete)' internal/client/ -c`.

    For SCHEMA-01..08:
    - SCHEMA-01 (file_system): `rg 'SchemaVersion:\s*1' internal/provider/file_system_resource.go` returns a hit; `rg 'TestUnit_FileSystemResource_StateUpgrade_V0toV1' internal/provider/ -l`.
    - SCHEMA-02 (file_system_export): same checks on `file_system_export_resource.go` + upgrade test.
    - SCHEMA-03 (nfs_export_policy): same on `nfs_export_policy_resource.go`.
    - SCHEMA-04 (smb_client_policy): same on `smb_client_policy_resource.go`.
    - SCHEMA-05 (smb_share_policy): same on `smb_share_policy_resource.go`.
    - SCHEMA-06 (qos_policy v0→v1 workload field): `rg 'TestUnit_QosPolicyResource_StateUpgrade_V0toV1' internal/provider/ -l`.
    - SCHEMA-07 (qos_policy v1→v2 context field): `rg 'TestUnit_QosPolicyResource_StateUpgrade_V1toV2' internal/provider/ -l`; `rg 'SchemaVersion:\s*2' internal/provider/qos_policy_resource.go`.
    - SCHEMA-08: `git log test/api-upgrade-2.23 --oneline | grep 5ab258a` non-empty; spot-check qos_policy upgrader uses `Pattern B` (`xxxV<N>Model`) per CONVENTIONS.md state upgrader section.

    Each row format identical to Task 1. Mark MISSING and add to `## Gaps` section on failure.

    Finally, write a `## Summary` section at top of RETRO.md with: `Total REQ-IDs verified: 19/19` OR `Total REQ-IDs verified: N/19, gaps: <list>`.
  </action>
  <verify>
    <automated>grep -E '^\| (WORKLOAD|SCHEMA)-0[1-9]' .planning/phases/59-api-2-23-consolidation-validation/59-01-RETRO.md | wc -l | grep -E '^11$'</automated>
  </verify>
  <acceptance_criteria>
    - RETRO.md has 11 additional rows (WORKLOAD-01..03 + SCHEMA-01..08)
    - Total rows for all 19 retro REQ-IDs present
    - `## Summary` section reports total verified count
    - No code files modified (only RETRO.md written)
  </acceptance_criteria>
  <done>All 19 retro REQ-IDs traced in RETRO.md. Summary section reports verified count and any gaps.</done>
</task>

</tasks>

<verification>
- `wc -l .planning/phases/59-api-2-23-consolidation-validation/59-01-RETRO.md` returns ≥ 25 lines (header + summary + 19 rows + gaps)
- `git status` shows only RETRO.md as new file (no code modifications)
</verification>

<success_criteria>
- All 19 retro REQ-IDs documented in RETRO.md with evidence commands + results
- Any gaps surfaced in dedicated `## Gaps` section for Wave 2 plans to address
- No source code modified
</success_criteria>

<output>
After completion, create `.planning/phases/59-api-2-23-consolidation-validation/59-01-SUMMARY.md` summarizing: total REQ-IDs verified, any gaps found, location of RETRO.md.
</output>
