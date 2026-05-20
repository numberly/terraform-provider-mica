---
phase: 60-v2-23-0-release
plan: 04
type: execute
wave: 4
depends_on: ["60-03"]
files_modified: []
autonomous: false
requirements:
  - RELEASE-05
must_haves:
  truths:
    - "Tag v2.23.0 exists on the main branch tip"
    - "Tag is pushed to origin"
    - "release.yml workflow runs and succeeds (CI gate + GoReleaser + Cosign signing)"
    - "GitHub Release v2.23.0 exists with signed artefacts (SHA256SUMS, .sig, .cosign.sig, .pem, zips, manifest)"
  artifacts:
    - path: ".planning/phases/60-v2-23-0-release/60-04-RELEASE.md"
      provides: "Release URL + artefact inventory"
      contains: "release_url:"
  key_links:
    - from: "tag v2.23.0"
      to: ".github/workflows/release.yml"
      via: "push tag event"
      pattern: "on.push.tags"
    - from: "release.yml"
      to: ".goreleaser.yml"
      via: "goreleaser/goreleaser-action@v6"
      pattern: "goreleaser-action"
---

<objective>
Tag the merge commit as `v2.23.0`, push the tag, and verify GoReleaser publishes signed artefacts.

Purpose: this is the publication step. Once the tag is pushed, `release.yml` triggers, runs the CI gate again, builds binaries for linux/darwin/windows × amd64/arm64, signs with GPG + Cosign, and publishes a GitHub Release. The Terraform Registry will ingest the release on its next sync.

Output: tag pushed, release.yml run green, GitHub Release page populated, signatures verified.
</objective>

<execution_context>
@$HOME/.claude/get-shit-done/workflows/execute-plan.md
@$HOME/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/ROADMAP.md
@.planning/REQUIREMENTS.md
@.planning/phases/60-v2-23-0-release/60-03-SUMMARY.md
@.planning/phases/60-v2-23-0-release/60-03-MERGE.md
@.github/workflows/release.yml
@.goreleaser.yml
@CHANGELOG.md
</context>

<tasks>

<task type="checkpoint:human-verify" gate="blocking">
  <name>Task 1: Pre-tag sanity check on main</name>
  <read_first>
    - .planning/phases/60-v2-23-0-release/60-03-MERGE.md
    - CHANGELOG.md (top of file on main)
    - .github/workflows/release.yml (confirm trigger pattern `v*` excluding `v*-pulumi*`)
  </read_first>
  <what-built>Merge from plan 60-03. Local main should match origin/main. This checkpoint verifies main is the canonical state before tagging.</what-built>
  <how-to-verify>
    Run these checks and confirm each:
    ```bash
    # On main, in sync with origin
    git rev-parse --abbrev-ref HEAD                       # main
    git fetch origin main
    [ "$(git rev-parse main)" = "$(git rev-parse origin/main)" ] && echo SYNCED

    # CHANGELOG v2.23.0 is at the top
    head -1 CHANGELOG.md                                   # ## [2.23.0] — ...

    # No existing v2.23.0 tag (locally or remote)
    git tag -l v2.23.0                                     # empty
    git ls-remote --tags origin v2.23.0                    # empty

    # Build still works on the release commit
    make build
    ls -la terraform-provider-mica
    ```
    Confirm:
    1. main is synced with origin
    2. CHANGELOG starts with `## [2.23.0]`
    3. No prior v2.23.0 tag exists
    4. `make build` succeeds
  </how-to-verify>
  <acceptance_criteria>
    - `git rev-parse main` == `git rev-parse origin/main`
    - `head -1 CHANGELOG.md` matches `^## \[2\.23\.0\]`
    - `git tag -l v2.23.0` is empty
    - `git ls-remote --tags origin refs/tags/v2.23.0` is empty
    - `make build` exit 0; `terraform-provider-mica` binary exists
  </acceptance_criteria>
  <resume-signal>Type "ready-to-tag" to proceed, or describe blocker.</resume-signal>
  <done>main is verified release-ready. No prior v2.23.0 tag. Build succeeds.</done>
</task>

<task type="auto">
  <name>Task 2: Create and push annotated tag v2.23.0</name>
  <read_first>
    - CHANGELOG.md (top section content — used in tag annotation)
  </read_first>
  <action>
    Create an annotated tag with a concise English message. Use the CHANGELOG header date in the annotation for traceability.

    ```bash
    git checkout main
    git pull --ff-only origin main

    TAG=v2.23.0
    git tag -a "$TAG" -m "release: $TAG — FlashBlade API 2.23 upgrade

    New: flashblade_workload resource + workload/resiliency_group/resiliency_group_member data sources.
    Changed: API bump 2.22 -> 2.23; six schema v0->v1 migrations adding the workload field.
    Bridge: Pulumi schema artefacts regenerated.

    See CHANGELOG.md for the full release notes."

    # Inspect before push
    git show "$TAG" --stat | head -20

    # Push the tag — this triggers release.yml
    git push origin "$TAG"
    ```

    Notes:
    - Tag pattern in release.yml: `v*` excluding `v*-pulumi*` — `v2.23.0` matches the include rule
    - The `before` hook in `.goreleaser.yml` runs `go test ./internal/...` again — keep main green
  </action>
  <verify>
    <automated>git tag -l v2.23.0 | grep -E '^v2\.23\.0$' &amp;&amp; git ls-remote --tags origin refs/tags/v2.23.0 | grep -E 'refs/tags/v2\.23\.0$' &amp;&amp; git tag -l v2.23.0 -n1 | grep -E 'API 2\.23 upgrade'</automated>
  </verify>
  <acceptance_criteria>
    - Local annotated tag exists: `git cat-file -t v2.23.0` == `tag` (annotated, not lightweight)
    - Tag is pushed: `git ls-remote --tags origin refs/tags/v2.23.0` returns a line
    - Tag annotation references API 2.23 upgrade: `git tag -l v2.23.0 -n5 | grep -qE 'API 2\.23'`
    - Tag points to main HEAD: `[ "$(git rev-parse v2.23.0^{commit})" = "$(git rev-parse main)" ]`
  </acceptance_criteria>
  <done>Annotated tag `v2.23.0` exists locally and on origin, pointing to the main release commit.</done>
</task>

<task type="checkpoint:human-verify" gate="blocking">
  <name>Task 3: Verify release.yml run + GoReleaser publication</name>
  <read_first>
    - .github/workflows/release.yml
    - .goreleaser.yml (artefact list, signatures, manifest)
  </read_first>
  <what-built>Tag push from task 2 triggered the `Release` workflow. This checkpoint waits for the workflow to succeed and verifies the published artefacts.</what-built>
  <how-to-verify>
    Watch the workflow:
    ```bash
    gh run list --workflow=release.yml --limit 5
    gh run watch $(gh run list --workflow=release.yml --branch=main --limit 1 --json databaseId --jq '.[0].databaseId')
    ```

    Once the run is `completed` + `success`, verify the release:
    ```bash
    gh release view v2.23.0
    gh release view v2.23.0 --json assets --jq '.assets[].name' | sort
    ```

    Expected artefacts (from `.goreleaser.yml` + `release.yml`):
    - `terraform-provider-mica_2.23.0_linux_amd64.zip`
    - `terraform-provider-mica_2.23.0_linux_arm64.zip`
    - `terraform-provider-mica_2.23.0_darwin_amd64.zip`
    - `terraform-provider-mica_2.23.0_darwin_arm64.zip`
    - `terraform-provider-mica_2.23.0_windows_amd64.zip`
    - `terraform-provider-mica_2.23.0_SHA256SUMS`
    - `terraform-provider-mica_2.23.0_SHA256SUMS.sig` (GPG)
    - `terraform-provider-mica_2.23.0_SHA256SUMS.cosign.sig`
    - `terraform-provider-mica_2.23.0_SHA256SUMS.cosign.pem`
    - `terraform-provider-mica_2.23.0_manifest.json`

    Verify Cosign signature locally:
    ```bash
    mkdir -p /tmp/v2.23.0-verify && cd /tmp/v2.23.0-verify
    gh release download v2.23.0 -p '*SHA256SUMS*'
    cosign verify-blob \
      --signature terraform-provider-mica_2.23.0_SHA256SUMS.cosign.sig \
      --certificate terraform-provider-mica_2.23.0_SHA256SUMS.cosign.pem \
      --certificate-identity-regexp "https://github.com/numberly/terraform-provider-mica" \
      --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
      terraform-provider-mica_2.23.0_SHA256SUMS
    ```

    Record outcome in `.planning/phases/60-v2-23-0-release/60-04-RELEASE.md`:
    ```markdown
    ---
    plan: 60-04
    tag: v2.23.0
    release_url: https://github.com/numberly/terraform-provider-mica/releases/tag/v2.23.0
    workflow_run: <gh run URL>
    published_at: <ISO timestamp>
    artefact_count: <N>
    cosign_verified: true
    ---

    # Release: v2.23.0

    URL: <release URL>
    Workflow run: <URL>

    ## Artefacts

    <list from `gh release view`>

    ## Signatures

    - GPG (.sig): <fingerprint or "present">
    - Cosign keyless (.cosign.sig + .pem): verified
    ```

    Commit:
    ```bash
    cd <repo>
    git add .planning/phases/60-v2-23-0-release/60-04-RELEASE.md
    git commit --no-verify -m "docs(release): record v2.23.0 GitHub Release artefacts"
    git push origin main
    ```
  </how-to-verify>
  <acceptance_criteria>
    - Most recent release.yml run: `gh run list --workflow=release.yml --limit 1 --json conclusion --jq '.[0].conclusion'` == `success`
    - `gh release view v2.23.0 --json isDraft,isPrerelease --jq '.isDraft + "/" + (.isPrerelease | tostring)'` == `false/false`
    - At least 5 zip artefacts: `gh release view v2.23.0 --json assets --jq '[.assets[].name | select(endswith(".zip"))] | length'` ≥ 5
    - SHA256SUMS + GPG sig + Cosign sig + Cosign cert + manifest all present in `gh release view v2.23.0 --json assets --jq '.assets[].name'`
    - Cosign verification command exit 0
    - `60-04-RELEASE.md` exists and is committed
  </acceptance_criteria>
  <resume-signal>Type "release-published" once all checks pass, or "release-failed: <run-url>" to halt.</resume-signal>
  <done>v2.23.0 is published on GitHub Releases with all expected signed artefacts. Cosign verification clean.</done>
</task>

</tasks>

<verification>
- Tag v2.23.0 exists locally and on origin, pointing to main
- release.yml run on tag push concluded `success`
- GitHub Release v2.23.0 exists with full artefact set
- Cosign signature on SHA256SUMS verifies
- 60-04-RELEASE.md documents the publication
</verification>

<success_criteria>
- `gh release view v2.23.0 --json isDraft --jq .isDraft` == `false`
- `gh release view v2.23.0 --json assets --jq '.assets | length'` ≥ 9
- Cosign verify-blob succeeds locally
</success_criteria>

<output>
After completion, create `.planning/phases/60-v2-23-0-release/60-04-SUMMARY.md` with:
- Release URL
- Workflow run URL
- Artefact inventory
- Hand-off to plan 60-05 (baseline bump + milestone archive)
</output>
