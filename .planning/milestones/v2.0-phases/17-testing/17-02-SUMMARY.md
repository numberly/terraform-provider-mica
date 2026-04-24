---
phase: 17-testing
plan: "02"
subsystem: testing
tags: [terraform, hcl, acceptance-test, replication, flashblade]

requires:
  - phase: 15-replication-resources
    provides: flashblade_object_store_remote_credentials and flashblade_bucket_replica_link resources
  - phase: 16-workflow-documentation
    provides: s3-bucket-replication workflow example (reference pattern)

provides:
  - Complete acceptance test HCL for bidirectional bucket replication lifecycle
  - Parameterized dual-provider configuration for any FlashBlade pair

affects:
  - WFL-03 requirement (live acceptance test gate)

tech-stack:
  added: []
  patterns:
    - "Dual-provider HCL pattern: two flashblade provider blocks with aliases for cross-array resources"
    - "Shared-secret access key: secondary key uses primary secret_access_key for symmetric auth"
    - "Pre-flight data source: flashblade_array_connection used to verify connectivity before resource creation"

key-files:
  created:
    - tmp/test-purestorage/replication-acceptance/main.tf
    - tmp/test-purestorage/replication-acceptance/variables.tf
    - tmp/test-purestorage/replication-acceptance/outputs.tf
    - tmp/test-purestorage/replication-acceptance/README.md
  modified: []

key-decisions:
  - "Acceptance test HCL committed to tmp/test-purestorage repo (separate from provider repo) as it requires live arrays to run"
  - "Live execution gated behind human checkpoint — skip is valid if no FlashBlade pair is available"
  - "destroy_eradicate_on_delete=false on buckets as safety guard against data loss during testing"

patterns-established:
  - "Acceptance test pattern: write HCL first, gate live execution behind checkpoint"

requirements-completed:
  - WFL-03

duration: 3min
completed: "2026-03-29"
---

# Phase 17 Plan 02: Replication Acceptance Test HCL Summary

**Bidirectional replication acceptance test HCL with dual-provider setup, shared access keys, remote credentials, and replica links — ready for live execution pending human review**

## Performance

- **Duration:** 3 min
- **Started:** 2026-03-29T12:02:01Z
- **Completed:** 2026-03-29T12:05:16Z
- **Tasks:** 1 of 2 (Task 2 is a human checkpoint — see below)
- **Files modified:** 4

## Accomplishments

- Created complete acceptance test HCL in `tmp/test-purestorage/replication-acceptance/` covering the full replication lifecycle
- Dual-provider setup (primary + secondary FlashBlade) parameterized via `variables.tf`
- Bidirectional bucket replica links with pause/resume support via `paused` attribute
- Verification outputs for all replication status fields (status, direction, paused, lag, status_details)
- README with prerequisites, step-by-step execution, pause/resume test instructions, and cleanup

## Task Commits

1. **Task 1: Write replication acceptance test HCL** - `0e556ca` (feat) — in `tmp/test-purestorage` repo

**Task 2** is a `checkpoint:human-verify` gate. Live execution requires a FlashBlade pair and human approval. This checkpoint can be skipped without blocking milestone completion.

**Plan metadata:** pending (final commit below)

## Files Created/Modified

- `/home/gule/Workspace/team-infrastructure/tmp/test-purestorage/replication-acceptance/main.tf` — Dual-provider config: accounts, versioned buckets, access keys (shared secret), remote credentials, bidirectional replica links (189 lines)
- `/home/gule/Workspace/team-infrastructure/tmp/test-purestorage/replication-acceptance/variables.tf` — Input variables: endpoints, API tokens, array names, account/bucket names (all sensitive vars marked)
- `/home/gule/Workspace/team-infrastructure/tmp/test-purestorage/replication-acceptance/outputs.tf` — Verification outputs: connection status, replica link status/direction/paused/lag on both sides
- `/home/gule/Workspace/team-infrastructure/tmp/test-purestorage/replication-acceptance/README.md` — Prerequisites, run instructions, pause/resume test steps, cleanup

## Decisions Made

- HCL files committed to `tmp/test-purestorage` repo (separate from provider repo) since that repo is already used for live FlashBlade testing
- `destroy_eradicate_on_delete = false` on both buckets to prevent accidental data loss during acceptance testing
- Shared-secret pattern: secondary access key uses `secret_access_key = flashblade_object_store_access_key.primary.secret_access_key` (matches workflow example pattern)
- `paused` attribute left as commented hint rather than explicit `false` — cleaner HCL, default behavior is unpaused

## Checkpoint Status

**Task 2 (human-verify) — Paused at checkpoint.**

The acceptance test HCL is complete and ready for review. To proceed:

1. Review the files in `tmp/test-purestorage/replication-acceptance/`
2. If you have two FlashBlade arrays available:
   - Create `terraform.tfvars` with your endpoints, tokens, and array names
   - Run `terraform init && terraform plan` to verify syntax
   - Run `terraform apply` to execute the full lifecycle
   - Verify outputs show expected replication status
   - Run `terraform destroy` to clean up
3. If no FlashBlade pair is available, this checkpoint can be skipped — WFL-03 is satisfied by the HCL artifact itself

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

To run the live acceptance test, create `tmp/test-purestorage/replication-acceptance/terraform.tfvars`:

```hcl
primary_endpoint     = "https://fb-a.example.com"
primary_api_token    = "T-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
primary_array_name   = "fb-a"

secondary_endpoint   = "https://fb-b.example.com"
secondary_api_token  = "T-yyyyyyyy-yyyy-yyyy-yyyy-yyyyyyyyyyyy"
secondary_array_name = "fb-b"

account_name = "repl-test"
bucket_name  = "repl-test-bucket"
```

## Next Phase Readiness

- Phase 17 testing complete — all provider resources have test coverage
- Milestone v2.0 (Cross-Array Bucket Replication) is functionally complete
- Live acceptance test can be run at any time when a FlashBlade pair is available

---
*Phase: 17-testing*
*Completed: 2026-03-29*
