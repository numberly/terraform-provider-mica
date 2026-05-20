---
phase: 60-v2-23-0-release
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - CHANGELOG.md
  - .planning/phases/59-api-2-23-consolidation-validation/59-HUMAN-UAT.md
autonomous: false
requirements:
  - RELEASE-01
must_haves:
  truths:
    - "VALID-04 (live acceptance on par5 + pa7) is resolved before release proceeds"
    - "CHANGELOG.md has a v2.23.0 entry following the [2.22.4] format"
    - "CHANGELOG entry covers: API 2.23 bump, new workload resource, 3 new data sources, 6 schema migrations, Pulumi bridge regen"
  artifacts:
    - path: "CHANGELOG.md"
      provides: "v2.23.0 release notes"
      contains: "## [2.23.0]"
    - path: ".planning/phases/59-api-2-23-consolidation-validation/59-HUMAN-UAT.md"
      provides: "VALID-04 resolution evidence"
      contains: "status: resolved"
  key_links:
    - from: "CHANGELOG.md v2.23.0 entry"
      to: "REQUIREMENTS.md SCHEMA/WORKLOAD/RESILIENCY/BRIDGE rows"
      via: "narrative reference"
      pattern: "workload|resiliency|schema v[01]"
---

<objective>
Open the release pipeline with two mandatory gates: (1) confirm the human-verified live acceptance gap (VALID-04) is resolved, then (2) author the v2.23.0 CHANGELOG entry.

Purpose: VALID-04 is the single blocker carried from Phase 59. No release work proceeds until par5 + pa7 acceptance is signed off. The CHANGELOG must exist on branch BEFORE rebase/PR so reviewers see release notes in the PR diff.

Output: VALID-04 marked resolved in 59-HUMAN-UAT.md (or aborted/escalated if not), and CHANGELOG.md updated with a `[2.23.0]` section.
</objective>

<execution_context>
@$HOME/.claude/get-shit-done/workflows/execute-plan.md
@$HOME/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/PROJECT.md
@.planning/ROADMAP.md
@.planning/REQUIREMENTS.md
@.planning/STATE.md
@.planning/phases/59-api-2-23-consolidation-validation/59-HUMAN-UAT.md
@.planning/phases/59-api-2-23-consolidation-validation/59-VERIFICATION.md
@CHANGELOG.md
@CLAUDE.md
@CONVENTIONS.md
</context>

<tasks>

<task type="checkpoint:human-verify" gate="blocking">
  <name>Task 1: Pre-flight gate — confirm VALID-04 live acceptance is done</name>
  <read_first>
    - .planning/phases/59-api-2-23-consolidation-validation/59-HUMAN-UAT.md
    - .planning/phases/59-api-2-23-consolidation-validation/59-VERIFICATION.md (Human Verification Required section)
    - .planning/REQUIREMENTS.md (VALID-04 row)
  </read_first>
  <what-built>
    Nothing yet — this is a pre-flight human gate. The Phase 59 verification report flags VALID-04 (live acceptance on par5 + pa7) as `pending` in 59-HUMAN-UAT.md. The release MUST NOT proceed until the operator confirms the live cycle has been run (or has explicitly waived it with documented rationale).
  </what-built>
  <how-to-verify>
    1. Read `.planning/phases/59-api-2-23-consolidation-validation/59-HUMAN-UAT.md` — confirm VALID-04 status
    2. If `status: pending`, run the live cycle described in 59-VERIFICATION.md "Human Verification Required #1":
       - For each array in {par5, pa7}, export `FLASHBLADE_ENDPOINT` + `FLASHBLADE_API_TOKEN`
       - For each fixture subdir under `examples/acceptance/api-2-23/{workload,resiliency_group,resiliency_group_member,migrated_schemas}`:
         a. `terraform init`
         b. `terraform apply -auto-approve`
         c. `terraform plan -detailed-exitcode` (expect exit 0)
         d. `terraform destroy -auto-approve`
       - NOTE: fixtures directory may not exist yet — see plan 59-04 task 1 or author manually first
    3. After live runs: update 59-HUMAN-UAT.md:
       - Frontmatter `status: resolved`
       - Test 1 `result: [pass]`
       - Add `## Resolution` section with array + date + summary, and link to `59-04-ACCEPTANCE.md` if created
    4. Confirm 59-HUMAN-UAT.md is committed: `git status .planning/phases/59-*/`
  </how-to-verify>
  <acceptance_criteria>
    - `grep -E '^status: resolved' .planning/phases/59-api-2-23-consolidation-validation/59-HUMAN-UAT.md` returns 1 match
    - `grep -E 'result: \[pass\]' .planning/phases/59-api-2-23-consolidation-validation/59-HUMAN-UAT.md` returns ≥1 match
    - `git status --porcelain .planning/phases/59-api-2-23-consolidation-validation/59-HUMAN-UAT.md` is empty (committed)
  </acceptance_criteria>
  <resume-signal>
    Type "approved" to confirm VALID-04 is resolved (acceptance ran clean on par5 + pa7), or
    type "waived: <reason>" to explicitly waive with rationale (escalates risk — Claude will record waiver in CHANGELOG), or
    describe blockers (e.g. fixture authoring still needed, array credentials missing)
  </resume-signal>
  <done>59-HUMAN-UAT.md frontmatter `status: resolved` and VALID-04 row marked Complete in REQUIREMENTS.md traceability table (Task 3 will update REQUIREMENTS.md).</done>
</task>

<task type="auto">
  <name>Task 2: Author v2.23.0 CHANGELOG entry</name>
  <read_first>
    - CHANGELOG.md (full file — match the [2.22.4] style exactly)
    - .planning/REQUIREMENTS.md (WORKLOAD/RESILIENCY/SCHEMA/BRIDGE rows for accurate scope)
    - .planning/ROADMAP.md (Phase 59 success criteria for tone)
    - internal/client/client.go (confirm `const APIVersion = "2.23"`)
    - GNUmakefile (confirm BINARY_NAME, namespace = numberly/mica)
  </read_first>
  <action>
    Prepend a new `## [2.23.0] — YYYY-MM-DD` section at the TOP of `CHANGELOG.md` (above the existing `## [2.22.4]` section). Use today's date in ISO format.

    Required sections (match [2.22.4] layout — H2 title, then H3 subsections):

    ### Added
    - **`flashblade_workload` resource** — CRUD + drift detection + import. Covers WORKLOAD-01.
    - **`flashblade_workload` data source**. Covers WORKLOAD-02.
    - **`flashblade_resiliency_group` data source** (read-only). Covers RESILIENCY-01.
    - **`flashblade_resiliency_group_member` data source** (read-only). Covers RESILIENCY-02.

    ### Changed
    - **FlashBlade API version bumped from `2.22` to `2.23`** across provider, client, mock handlers, and API references. Covers API-01..03.
    - **Schema migrations (v0 → v1)** add a computed `workload` field to the following resources. State is auto-upgraded on first refresh; no user action required:
      - `flashblade_file_system` (SCHEMA-01)
      - `flashblade_file_system_export` (SCHEMA-02, v1 → v2)
      - `flashblade_nfs_export_policy` (SCHEMA-03)
      - `flashblade_smb_client_policy` (SCHEMA-04)
      - `flashblade_smb_share_policy` (SCHEMA-05)
      - `flashblade_qos_policy` (SCHEMA-06, plus computed `context` field SCHEMA-07, v1 → v2)
    - **Pulumi bridge artefacts regenerated** for API 2.23 (`schema.json`, `bridge-metadata.json`, `schema-embed.json`). Covers BRIDGE-01..03.

    ### Migration
    No manual user action required. Schema upgraders run automatically on first `terraform refresh` / `plan` / `apply` after upgrading the provider. The new `workload` attribute is computed (read-only) and surfaces the FlashBlade workload assignment for the underlying object.

    ### Validation
    Validated on FlashBlade arrays `par5` and `pa7` (API 2.23, Purity//FB ≥ 4.6.7). `make test` (807+ tests) + `make lint` + `make docs` + `make tfgen` all clean on the release commit.

    Rules:
    - Use Markdown only (no HTML)
    - Bullet style: `- **bold key** — description`
    - Do NOT use emojis
    - Do NOT include Co-Authored-By or AI signature lines
    - Keep technical and concise (Cromagnon mode)

    If the operator chose `waived` in Task 1, add a `### Known limitations` block:
    > Live acceptance on par5 + pa7 was waived for this release. Rationale: <reason from operator>. Acceptance will be re-run before v2.24.0.
  </action>
  <verify>
    <automated>head -5 CHANGELOG.md | grep -E '^## \[2\.23\.0\] —' &amp;&amp; grep -E '^### (Added|Changed|Migration|Validation)$' CHANGELOG.md | head -4 &amp;&amp; ! grep -E 'Co-Authored-By' CHANGELOG.md</automated>
  </verify>
  <acceptance_criteria>
    - `head -1 CHANGELOG.md` matches `^## \[2\.23\.0\] — 20\d\d-\d\d-\d\d$`
    - `grep -c '^### ' CHANGELOG.md` ≥ 6 (4 new + ≥2 existing)
    - All 4 new resource/data-source REQ-IDs (WORKLOAD-01, WORKLOAD-02, RESILIENCY-01, RESILIENCY-02) appear in the new section: `awk '/^## \[2\.23\.0\]/,/^## \[2\.22\.4\]/' CHANGELOG.md | grep -E 'WORKLOAD-0[12]|RESILIENCY-0[12]'` returns ≥4 matches
    - All 6 schema migrations cited: `awk '/^## \[2\.23\.0\]/,/^## \[2\.22\.4\]/' CHANGELOG.md | grep -cE 'SCHEMA-0[1-7]'` ≥ 6
    - No Co-Authored-By lines: `! grep -i 'co-authored-by' CHANGELOG.md`
  </acceptance_criteria>
  <done>CHANGELOG.md starts with a complete `[2.23.0]` section covering Added / Changed / Migration / Validation, citing the relevant REQ-IDs and matching the existing layout style.</done>
</task>

<task type="auto">
  <name>Task 3: Mark VALID-04 + RELEASE-01 resolved in REQUIREMENTS.md and commit</name>
  <read_first>
    - .planning/REQUIREMENTS.md (Traceability table at the bottom)
  </read_first>
  <action>
    Update `.planning/REQUIREMENTS.md`:

    1. In the "VALIDATION — finalisation" section, flip `[ ] VALID-04` to `[x] VALID-04` (the inline acceptance line just below the `### VALIDATION` heading)
    2. In the "RELEASE — ship v2.23.0" section, flip `[ ] RELEASE-01` to `[x] RELEASE-01`
    3. In the Traceability table:
       - `| VALID-04 | Phase 59 | Pending |` → `| VALID-04 | Phase 59 | Complete |`
       - `| RELEASE-01 | Phase 60 | Pending |` → `| RELEASE-01 | Phase 60 | Complete |`
    4. Update the coverage footer line to reflect 2 more Complete (e.g. `21 retro / Done · 6 Phase 59 active · 6 Phase 60 active` if applicable — keep counts correct)

    Then commit ALL files in one commit (--no-verify per CLAUDE.md, English commit message):

    ```bash
    git add CHANGELOG.md \
      .planning/phases/59-api-2-23-consolidation-validation/59-HUMAN-UAT.md \
      .planning/REQUIREMENTS.md
    git commit --no-verify -m "docs(release): add v2.23.0 CHANGELOG entry and close VALID-04"
    ```

    Do NOT push yet — push happens in plan 60-02 after rebase.
  </action>
  <verify>
    <automated>git log -1 --pretty=%s | grep -E 'docs\(release\): add v2\.23\.0 CHANGELOG' &amp;&amp; git diff HEAD~1 HEAD --name-only | sort -u | grep -E '(CHANGELOG\.md|59-HUMAN-UAT\.md|REQUIREMENTS\.md)' | wc -l | grep -E '^3$'</automated>
  </verify>
  <acceptance_criteria>
    - HEAD commit subject matches `docs(release): add v2.23.0 CHANGELOG entry and close VALID-04`
    - HEAD commit touches exactly 3 files: CHANGELOG.md, 59-HUMAN-UAT.md, REQUIREMENTS.md
    - `grep -E '^\| VALID-04 .* Complete' .planning/REQUIREMENTS.md` returns 1 match
    - `grep -E '^\| RELEASE-01 .* Complete' .planning/REQUIREMENTS.md` returns 1 match
    - `git log -1 --pretty=%B | grep -i 'co-authored-by'` returns nothing
    - Working tree clean for these three files: `git status --porcelain CHANGELOG.md .planning/REQUIREMENTS.md .planning/phases/59-*/59-HUMAN-UAT.md` is empty
  </acceptance_criteria>
  <done>Single commit on `test/api-upgrade-2.23` containing CHANGELOG entry + UAT resolution + REQUIREMENTS traceability updates. Branch is ready for rebase in plan 60-02.</done>
</task>

</tasks>

<verification>
- VALID-04 is resolved (or explicitly waived with documented rationale)
- CHANGELOG.md `[2.23.0]` section exists with Added/Changed/Migration/Validation subsections
- REQUIREMENTS.md traceability shows VALID-04 + RELEASE-01 as Complete
- All updates landed in a single commit on `test/api-upgrade-2.23` (not pushed yet)
</verification>

<success_criteria>
- `head -1 CHANGELOG.md` matches `^## \[2\.23\.0\]`
- `grep 'status: resolved' .planning/phases/59-*/59-HUMAN-UAT.md` returns 1 match
- HEAD commit subject = `docs(release): add v2.23.0 CHANGELOG entry and close VALID-04`
- Working tree clean
</success_criteria>

<output>
After completion, create `.planning/phases/60-v2-23-0-release/60-01-SUMMARY.md` documenting:
- VALID-04 resolution path (live cycle / waiver)
- Commit hash of the CHANGELOG commit
- Hand-off to plan 60-02 (rebase + PR)
</output>
