---
plan: 60-03
pr_number: 16
pr_url: https://github.com/numberly/terraform-provider-mica/pull/16
merge_sha: 3fd485d1a987725661dd28fd9b6431f2d5e62210
merge_strategy: squash
merged_at: 2026-05-20T00:00:00Z
---

# Merge: v2.23.0 -> main

Merge commit: 3fd485d1a987725661dd28fd9b6431f2d5e62210
Strategy: squash
PR: https://github.com/numberly/terraform-provider-mica/pull/16

## main tip

```
3fd485d1a987725661dd28fd9b6431f2d5e62210 Guillaume LEGRAIN feat(v2.23.0): FlashBlade API 2.23 upgrade — workload + resiliency groups (#16)
```

## Notes

- Local main had 3 divergent commits (`e3be4fe`, `6d5667e`, `2751103`) in `.claude/` tooling files.
- Rebased local main onto `origin/main`; conflicts in `.claude/` files resolved by keeping the squash commit (origin/main) version.
- The dropped commit (`2751103 chore: update serena project config`) was detected as already-upstream by git.
- `git rev-parse main == git rev-parse origin/main` after rebase: no push needed.
- Branch `test/api-upgrade-2.23` left in place (not deleted); cleanup deferred post-tag.

## Handoff to plan 60-04

Next: create tag `v2.23.0` on `3fd485d` and push to trigger GoReleaser.
