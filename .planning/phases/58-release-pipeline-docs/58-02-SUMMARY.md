---
phase: 58-release-pipeline-docs
plan: 02
subsystem: pulumi
key-files:
  created:
    - pulumi/examples/target-py/Pulumi.yaml
    - pulumi/examples/target-py/__main__.py
    - pulumi/examples/target-go/Pulumi.yaml
    - pulumi/examples/target-go/main.go
    - pulumi/examples/remote_credentials-py/Pulumi.yaml
    - pulumi/examples/remote_credentials-py/__main__.py
    - pulumi/examples/remote_credentials-go/Pulumi.yaml
    - pulumi/examples/remote_credentials-go/main.go
    - pulumi/examples/bucket-py/Pulumi.yaml
    - pulumi/examples/bucket-py/__main__.py
    - pulumi/examples/bucket-go/Pulumi.yaml
    - pulumi/examples/bucket-go/main.go
tags: [pulumi, examples, python, go, docs, test-fixtures]
decisions: []
tech-stack:
  patterns:
    - Pulumi ProgramTest-style examples with Pulumi.yaml + main program
    - Provider configured via endpoint + auth.api_token with env var fallback comments
    - ResourceOptions(provider=provider) for explicit provider binding
    - CustomTimeouts for soft-delete resources (20m create/update, 30m delete)
metrics:
  duration: "0:08:00"
  completed_date: "2026-04-22"
---

# Phase 58 Plan 02: Pulumi ProgramTest Examples Summary

Created 6 ProgramTest-style examples under `./pulumi/examples/` covering 3 representative resources in Python and Go.

## What Was Built

| Example | Runtime | Resource | Key Feature |
|---------|---------|----------|-------------|
| `target-py` | Python | `flashblade.Target` | Basic resource with provider config |
| `target-go` | Go | `flashblade.Target` | Basic resource with provider config |
| `remote_credentials-py` | Python | `flashblade.ObjectStoreRemoteCredentials` | Composite name format, sensitive field |
| `remote_credentials-go` | Go | `flashblade.ObjectStoreRemoteCredentials` | Composite name format, sensitive field |
| `bucket-py` | Python | `flashblade.Bucket` | Soft-delete, custom timeouts (20m/20m/30m) |
| `bucket-go` | Go | `flashblade.Bucket` | Soft-delete, custom timeouts (20m/20m/30m) |

## Commits

| Task | Commit | Description |
|------|--------|-------------|
| 1 | `6c2a66b` | feat(58-02): add target Pulumi examples in Python and Go |
| 2 | `8440c6f` | feat(58-02): add remote credentials Pulumi examples in Python and Go |
| 3 | `b9c7e3c` | feat(58-02): add bucket Pulumi examples in Python and Go with soft-delete timeouts |

## Verification

- All 6 example directories exist with `Pulumi.yaml` + main program
- All Python examples import `pulumi_flashblade` and construct resources correctly
- All Go examples import `github.com/numberly/opentofu-provider-flashblade/pulumi/sdk/go/flashblade` and use `pulumi.Run`
- Bucket examples demonstrate `custom_timeouts` / `CustomTimeouts` for soft-delete (30m delete)
- Remote credentials examples show sensitive field handling (`secret_access_key` / `SecretAccessKey`)
- `.gitkeep` no longer the only file in `pulumi/examples/`

## Deviations from Plan

None - plan executed exactly as written.

## Known Stubs

None - all examples are fully wired with realistic resource configurations.

## Self-Check: PASSED

- [x] All 12 created files exist on disk
- [x] All 3 commits exist in git history
- [x] Acceptance criteria from all 3 tasks satisfied
