---
phase: 59-api-2-23-consolidation-validation
plan: 04
type: execute
wave: 2
depends_on:
  - 59-02
files_modified:
  - .planning/phases/59-api-2-23-consolidation-validation/59-04-ACCEPTANCE.md
  - examples/acceptance/api-2-23/
autonomous: false
requirements:
  - VALID-04

must_haves:
  truths:
    - "Acceptance tests run against par5 and pa7 for new resources (workload, resiliency_group, resiliency_group_member)"
    - "Acceptance tests run against par5 and pa7 for the 6 migrated schemas (file_system, file_system_export, nfs_export_policy, smb_client_policy, smb_share_policy, qos_policy)"
    - "Apply -> re-apply produces no drift (terraform plan after apply is clean)"
    - "Destroy is clean (no orphan resources)"
    - "Results documented in ACCEPTANCE.md per array + per resource"
  artifacts:
    - path: ".planning/phases/59-api-2-23-consolidation-validation/59-04-ACCEPTANCE.md"
      provides: "Live acceptance test results report"
      contains: "par5, pa7, apply, plan, destroy"
    - path: "examples/acceptance/api-2-23/"
      provides: "HCL fixtures used for acceptance runs (per-resource subdirs)"
  key_links:
    - from: "terraform apply"
      to: "terraform plan (drift check)"
      via: "re-run plan after apply"
      pattern: "No changes. Your infrastructure matches the configuration."
---

<objective>
Run live acceptance tests against real FlashBlade arrays par5 and pa7 for the 3 new API 2.23 resources/data sources and the 6 migrated schemas. Validate apply -> plan-no-drift -> destroy cycle on each array. Document all outcomes in ACCEPTANCE.md.

This plan is NOT autonomous: it requires the human operator to provide credentials and array access (par5, pa7). Claude orchestrates HCL fixtures and verification steps; the operator runs `terraform apply/plan/destroy` and pastes outputs.

Purpose: Cover VALID-04 — final end-to-end gate before release.
Output: `.planning/phases/59-api-2-23-consolidation-validation/59-04-ACCEPTANCE.md`, HCL fixtures under `examples/acceptance/api-2-23/`.
</objective>

<execution_context>
@$HOME/.claude/get-shit-done/workflows/execute-plan.md
@$HOME/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/STATE.md
@.planning/REQUIREMENTS.md
@CONVENTIONS.md
@examples/
</context>

<tasks>

<task type="auto">
  <name>Task 1: Author HCL fixtures for the 3 new resources and 6 migrated schemas</name>
  <files>examples/acceptance/api-2-23/workload/main.tf, examples/acceptance/api-2-23/resiliency_group/main.tf, examples/acceptance/api-2-23/resiliency_group_member/main.tf, examples/acceptance/api-2-23/migrated_schemas/main.tf</files>
  <read_first>
    - examples/resources/flashblade_workload/resource.tf (existing example, copy as base)
    - examples/data-sources/flashblade_resiliency_group/data-source.tf
    - examples/data-sources/flashblade_resiliency_group_member/data-source.tf
    - examples/resources/flashblade_file_system/resource.tf
    - examples/resources/flashblade_qos_policy/resource.tf
    - CLAUDE.md (Serena MCP for symbol lookup if any schema confusion)
  </read_first>
  <action>
    Create `examples/acceptance/api-2-23/` with one subdir per scope. Each subdir has a self-contained `main.tf` + `variables.tf` + `terraform.tfvars.example`.

    1. `examples/acceptance/api-2-23/workload/main.tf` — provider block (var-driven endpoint + token), `flashblade_workload` resource, matching data source by name. Names prefixed `acc-p59-workload-<rand>` to avoid collisions.

    2. `examples/acceptance/api-2-23/resiliency_group/main.tf` — data source `flashblade_resiliency_group` (well-known group name from the array; use `data.flashblade_resiliency_group.all` if list mode exists, otherwise document required name in tfvars).

    3. `examples/acceptance/api-2-23/resiliency_group_member/main.tf` — data source for one known member.

    4. `examples/acceptance/api-2-23/migrated_schemas/main.tf` — exercises one resource per migrated schema: `flashblade_file_system`, `flashblade_file_system_export`, `flashblade_nfs_export_policy`, `flashblade_smb_client_policy`, `flashblade_smb_share_policy`, `flashblade_qos_policy`. Each has a `workload` reference where applicable (depends on the workload fixture). The qos_policy also exercises the new `context` computed field via `output "qos_context" { value = flashblade_qos_policy.test.context }`.

    5. Add `examples/acceptance/api-2-23/README.md` documenting:
       - Required env vars: `FLASHBLADE_ENDPOINT`, `FLASHBLADE_API_TOKEN`
       - Per-array run command: `cd <subdir> && terraform init && terraform apply -auto-approve && terraform plan -detailed-exitcode && terraform destroy -auto-approve`
       - Expected `terraform plan -detailed-exitcode` after apply: exit 0 (no drift)

    Do NOT add these fixtures to `make docs` paths — they live under `examples/acceptance/` not `examples/resources/`. Verify with `grep -r 'examples/acceptance' tools/ scripts/ tfplugindocs.yaml 2>/dev/null` returns empty.
  </action>
  <verify>
    <automated>cd /home/gule/Workspace/team-infrastructure/terraform-provider-mica && test -f examples/acceptance/api-2-23/workload/main.tf && test -f examples/acceptance/api-2-23/resiliency_group/main.tf && test -f examples/acceptance/api-2-23/resiliency_group_member/main.tf && test -f examples/acceptance/api-2-23/migrated_schemas/main.tf && test -f examples/acceptance/api-2-23/README.md && terraform -chdir=examples/acceptance/api-2-23/workload fmt -check && terraform -chdir=examples/acceptance/api-2-23/migrated_schemas fmt -check</automated>
  </verify>
  <acceptance_criteria>
    - All 4 main.tf files exist under examples/acceptance/api-2-23/
    - README.md documents env vars and per-subdir run command
    - `terraform fmt -check` passes on every subdir
    - `terraform validate` (after `terraform init` with dev-override) passes on every subdir
  </acceptance_criteria>
  <done>HCL fixtures ready for live acceptance runs. README documents operator workflow.</done>
</task>

<task type="checkpoint:human-action" gate="blocking">
  <name>Task 2: Operator runs acceptance fixtures against par5 + pa7</name>
  <what-built>HCL fixtures in examples/acceptance/api-2-23/ exercising 3 new resources/DS and 6 migrated schemas.</what-built>
  <how-to-verify>
    Required: dev-override built (`make build && make dev-override`), env vars `FLASHBLADE_ENDPOINT` + `FLASHBLADE_API_TOKEN` set per array.

    For EACH array in {par5, pa7}:
      For EACH subdir in {workload, resiliency_group, resiliency_group_member, migrated_schemas}:
        1. Set env vars to the array's endpoint + token.
        2. `cd examples/acceptance/api-2-23/<subdir>`
        3. `terraform init`
        4. `terraform apply -auto-approve` — expect: succeeds.
        5. `terraform plan -detailed-exitcode` — expect: exit code 0 (no drift). If exit code 2 (drift detected), capture full plan output and report.
        6. `terraform destroy -auto-approve` — expect: succeeds.

    Record per-(array, subdir) result in `.planning/phases/59-api-2-23-consolidation-validation/59-04-ACCEPTANCE.md`:

    ```markdown
    # Phase 59 — Live Acceptance Results

    ## par5

    | Subdir | apply | plan-no-drift | destroy | Notes |
    |--------|-------|----------------|---------|-------|
    | workload | PASS | PASS | PASS | - |
    | resiliency_group | PASS | PASS | PASS | DS-only, no destroy |
    | ... | | | | |

    ## pa7
    <same table>
    ```

    If any cell is FAIL, include a `## Failures` section with the terraform output and root-cause hypothesis.
  </how-to-verify>
  <resume-signal>Type "approved" with link to populated ACCEPTANCE.md (all cells PASS), or paste failure details for triage.</resume-signal>
  <acceptance_criteria>
    - ACCEPTANCE.md exists and lists outcomes for all 8 combinations (2 arrays × 4 subdirs)
    - All `apply` and `destroy` cells are PASS
    - All `plan-no-drift` cells are PASS (no drift on re-plan after apply)
  </acceptance_criteria>
</task>

</tasks>

<verification>
- ACCEPTANCE.md populated with par5 + pa7 sections
- All apply / plan-no-drift / destroy cells = PASS
- HCL fixtures present and `terraform fmt -check` clean
</verification>

<success_criteria>
- VALID-04: Live acceptance against par5 + pa7 on workload, resiliency_group, resiliency_group_member, and 6 migrated schemas
- No drift on re-plan after apply (terraform plan -detailed-exitcode = 0)
- Destroy is clean on both arrays
</success_criteria>

<output>
After completion, create `.planning/phases/59-api-2-23-consolidation-validation/59-04-SUMMARY.md` with: arrays tested, subdirs covered, pass/fail matrix, any drift findings.
</output>
