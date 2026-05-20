---
plan: 60-04
tag: v2.23.0
tag_sha: f85ea8dcef6e5c321e46b3aee68a45109e4a0a86
release_url: https://github.com/numberly/terraform-provider-mica/releases/tag/v2.23.0
workflow_run: https://github.com/numberly/terraform-provider-mica/actions/runs/26152091668
published_at: 2026-05-20T08:57:26Z
artefact_count: 10
cosign_verified: true
---

# Release: v2.23.0

URL: https://github.com/numberly/terraform-provider-mica/releases/tag/v2.23.0
Workflow run: https://github.com/numberly/terraform-provider-mica/actions/runs/26152091668

Tag `v2.23.0` is an annotated tag pointing to `f85ea8d` (origin/main HEAD).

## Workflow Summary

Run ID: 26152091668
Conclusion: success
Duration: ~3 minutes total (CI Gate ~48s, Release job ~2m44s)

Jobs:
- CI Gate / Tests: success (48s)
- CI Gate / Lint: success (34s)
- CI Gate / Docs up to date: success (17s)
- Release (GoReleaser + Cosign): success (2m44s)

## Artefacts

| Artefact | Description |
|----------|-------------|
| `terraform-provider-mica_2.23.0_linux_amd64.zip` | Linux amd64 binary |
| `terraform-provider-mica_2.23.0_linux_arm64.zip` | Linux arm64 binary |
| `terraform-provider-mica_2.23.0_darwin_amd64.zip` | macOS amd64 binary |
| `terraform-provider-mica_2.23.0_darwin_arm64.zip` | macOS arm64 binary |
| `terraform-provider-mica_2.23.0_windows_amd64.zip` | Windows amd64 binary |
| `terraform-provider-mica_2.23.0_SHA256SUMS` | Checksum file (SHA256) |
| `terraform-provider-mica_2.23.0_SHA256SUMS.sig` | GPG detached signature |
| `terraform-provider-mica_2.23.0_SHA256SUMS.cosign.sig` | Cosign keyless signature |
| `terraform-provider-mica_2.23.0_SHA256SUMS.cosign.pem` | Cosign certificate |
| `terraform-provider-mica_2.23.0_manifest.json` | Terraform Registry manifest |

Total: 10 artefacts

## Signatures

- GPG (.sig): present (signed with GPG_FINGERPRINT from repo secret)
- Cosign keyless (.cosign.sig + .cosign.pem): verified in-workflow (exit 0)
  - Identity regexp: `https://github.com/numberly/terraform-provider-mica`
  - OIDC issuer: `https://token.actions.githubusercontent.com`

## Handoff to Plan 60-05

Next: bump `TEST_BASELINE` in `GNUmakefile` + archive milestone via `/gsd:complete-milestone`.
