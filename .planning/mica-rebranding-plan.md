# Mica Rebranding Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rebrand `terraform-provider-flashblade` → `terraform-provider-mica` to comply with Pure Storage trademark guidelines for open-source release. Switch license to GPL v3. Tag `v2.22.4`. Generate a personal migration guide for the project owner's existing Terraform/Pulumi state at the end of execution.

**Architecture:** Three-layer change: (1) identity layer renames (repo, Go module, binary, registry slug, Pulumi package name); (2) license switch with NOTICE; (3) documentation rebrand with trademark disclaimers. Resource type prefix `flashblade_*` and internal Go identifiers stay (descriptive nominative references). Provider `Metadata.TypeName` stays `"flashblade"` so HCL resource names remain `flashblade_bucket`, etc. Only the registry source path changes (`numberly/flashblade` → `numberly/mica`). Pulumi `Name` field becomes `mica` (asymmetric vs Terraform — documented in README).

**Tech Stack:** Go 1.25+, terraform-plugin-framework (MPL 2.0, GPL v3 compatible), goreleaser, GitHub Actions, Pulumi tfbridge.

**Reference spec:** `.planning/mica-rebranding-design.md`

**Audience for migration docs:** future open-source users (CHANGELOG/README cover them). The current sole consumer of the project is the project owner, who will run a personalized migration via Task 14 output.

**Conventions enforced from project CLAUDE.md / CONVENTIONS.md:**
- Commits in English, Conventional Commits style (`feat:`, `fix:`, `chore:`, `docs:`)
- No `Co-Authored-By` trailers
- Subagent commits use `--no-verify`
- Inclusive terminology
- Test naming: `TestUnit_<Resource>_<Operation>`
- Test count baseline: ≥ 818

---

## Pre-flight

### Task 0: Baseline validation

**Files:** none (read-only checks)

- [ ] **Step 1: Verify clean working tree**

```bash
git status --short
```

Expected: empty output. If not, stash or commit before proceeding.

- [ ] **Step 2: Capture baseline test count and lint state**

```bash
COUNT=$(go test ./... -count=1 -v 2>&1 | grep -c '^=== RUN  ')
echo "Baseline test count: $COUNT"
test "$COUNT" -ge 818 || { echo "FAIL: baseline test count below 818"; exit 1; }
make lint
```

Expected: count ≥ 818, lint exits 0. Save the exact `COUNT` value — must not decrease throughout execution.

- [ ] **Step 3: Capture baseline build**

```bash
make build && ls -la terraform-provider-flashblade*
```

Expected: binary present, named `terraform-provider-flashblade`.

- [ ] **Step 4: Inspect Pulumi Makefile targets**

```bash
test -f pulumi/Makefile && grep -E '^[a-zA-Z_-]+:' pulumi/Makefile
test -f pulumi/provider/Makefile && grep -E '^[a-zA-Z_-]+:' pulumi/provider/Makefile 2>/dev/null
```

Note the actual target names for schema generation and SDK build. Common names: `tfgen`, `provider`, `build_sdks`, `gen_go_sdk`, `schema`. **Record the discovered target names** — they replace the placeholders `<SCHEMA_TGT>` and `<SDK_BUILD_TGT>` used in Task 4.

- [ ] **Step 5: Inspect Pulumi `cmd/` directory layout**

```bash
ls pulumi/provider/cmd/ 2>/dev/null
ls pulumi/sdk/go/ 2>/dev/null
```

Record the existing subdirectory names (typically `pulumi-resource-flashblade/` and `pulumi/sdk/go/flashblade/`). These will be renamed in Task 4.

- [ ] **Step 6: No commit (read-only baseline)**

---

## Phase 1 — Identity layer

### Task 1: Rename Go module path

**Files:**
- Modify: `go.mod` (line 1)
- Modify: all `.go` files importing `github.com/numberly/opentofu-provider-flashblade/...` (~30 files)
- Modify: `pulumi/provider/go.mod` (line 1)
- Modify: `pulumi/sdk/go/go.mod` if it references the parent module
- Modify: all `.go` files under `pulumi/` importing the old module path

**Why:** Repo will be renamed `terraform-provider-mica`, Go module path must match.

- [ ] **Step 1: Update root `go.mod`**

```bash
sed -i 's|^module github.com/numberly/opentofu-provider-flashblade$|module github.com/numberly/terraform-provider-mica|' go.mod
head -1 go.mod
```

Expected output: `module github.com/numberly/terraform-provider-mica`

- [ ] **Step 2: Update all Go imports under root**

```bash
rg -l 'github.com/numberly/opentofu-provider-flashblade' --type go | \
  xargs sed -i 's|github.com/numberly/opentofu-provider-flashblade|github.com/numberly/terraform-provider-mica|g'
```

- [ ] **Step 3: Update all Pulumi `go.mod` files**

```bash
for f in pulumi/provider/go.mod pulumi/sdk/go/go.mod; do
  [ -f "$f" ] && sed -i 's|github.com/numberly/opentofu-provider-flashblade|github.com/numberly/terraform-provider-mica|g' "$f"
done
```

- [ ] **Step 4: Update Go imports under `pulumi/`**

```bash
rg -l 'github.com/numberly/opentofu-provider-flashblade' pulumi/ | \
  xargs sed -i 's|github.com/numberly/opentofu-provider-flashblade|github.com/numberly/terraform-provider-mica|g'
```

- [ ] **Step 5: Verify no straggling references**

```bash
rg 'numberly/opentofu-provider-flashblade'
```

Expected: empty output.

- [ ] **Step 6: Tidy and rebuild**

```bash
go mod tidy
go build ./...
(cd pulumi/provider && go mod tidy && go build ./...)
[ -f pulumi/sdk/go/go.mod ] && (cd pulumi/sdk/go && go mod tidy && go build ./...)
```

Expected: no errors anywhere.

- [ ] **Step 7: Run tests across both modules**

```bash
make test
(cd pulumi/provider && go test ./... -count=1)
```

Expected: all pass, count ≥ 818 in main module.

- [ ] **Step 8: Commit**

```bash
git add -A
git commit -m "chore(rebrand): rename Go module to terraform-provider-mica"
```

---

### Task 2: Update provider Address & generator hint

**Files:**
- Modify: `main.go` (`go:generate` comment, `Address` field, log line)

**Why:** `Address` is the gRPC binary identity used by Terraform CLI to resolve `source = "numberly/mica"`. Must match new registry slug. `Metadata.TypeName` stays `"flashblade"` so resources remain `flashblade_*`.

- [ ] **Step 1: Update `Address` field**

```bash
sed -i 's|"registry.terraform.io/numberly/flashblade"|"registry.terraform.io/numberly/mica"|' main.go
```

- [ ] **Step 2: Update `go:generate` provider-name flag**

```bash
sed -i 's|--provider-name flashblade|--provider-name mica|' main.go
```

- [ ] **Step 3: Update log line**

Use the `Read` tool to confirm the exact current log line in `main.go`, then use `Edit` tool to replace:

Old: `log.Printf("[INFO] flashblade provider %s (commit=%s, built=%s)", version, commit, date)`

New: `log.Printf("[INFO] mica provider %s (commit=%s, built=%s)", version, commit, date)`

(Use `Edit` rather than sed to avoid escaping pitfalls with quotes/brackets.)

- [ ] **Step 4: Verify `Metadata.TypeName` is unchanged**

```bash
grep 'resp.TypeName' internal/provider/provider.go
```

Expected: `resp.TypeName = "flashblade"` (DO NOT change — this drives the resource prefix).

- [ ] **Step 5: Build and test**

```bash
go build ./...
make test
```

Expected: ≥ 818 tests pass.

- [ ] **Step 6: Commit**

```bash
git add main.go
git commit -m "feat(rebrand): switch provider Address to numberly/mica"
```

---

### Task 3: Update goreleaser binary & project name

**Files:**
- Modify: `.goreleaser.yml`
- Modify: `.goreleaser.pulumi.yml` (if it exists)
- Modify: `GNUmakefile` if it hard-codes `terraform-provider-flashblade`

**Why:** Built binary must be named `terraform-provider-mica` for Terraform CLI to pick it up via `source = "numberly/mica"`.

- [ ] **Step 1: Update `.goreleaser.yml` binary name**

```bash
sed -i 's|"terraform-provider-flashblade_v|"terraform-provider-mica_v|' .goreleaser.yml
```

- [ ] **Step 2: Verify or set `project_name`**

```bash
grep -nE '^project_name:|name_template:' .goreleaser.yml
```

If `project_name:` is missing, prepend it to `.goreleaser.yml`:

```yaml
project_name: terraform-provider-mica
```

If `project_name:` exists, update its value to `terraform-provider-mica` (use `Edit` tool).

- [ ] **Step 3: Update `.goreleaser.pulumi.yml`**

```bash
[ -f .goreleaser.pulumi.yml ] && grep -n 'flashblade\|opentofu-provider' .goreleaser.pulumi.yml
```

For each match, replace appropriately:

```bash
[ -f .goreleaser.pulumi.yml ] && sed -i 's|pulumi-resource-flashblade|pulumi-resource-mica|g; s|opentofu-provider-flashblade|terraform-provider-mica|g' .goreleaser.pulumi.yml
```

- [ ] **Step 4: Update `GNUmakefile` if hard-coded**

```bash
grep -n 'terraform-provider-flashblade\|opentofu-provider' GNUmakefile
```

If matches found:

```bash
sed -i 's|terraform-provider-flashblade|terraform-provider-mica|g; s|opentofu-provider-flashblade|terraform-provider-mica|g' GNUmakefile
```

- [ ] **Step 5: Snapshot build sanity check**

```bash
goreleaser build --snapshot --clean --single-target
find dist -name 'terraform-provider-mica*' -type f
```

Expected: at least one file matching `terraform-provider-mica_v*` exists under `dist/`.

- [ ] **Step 6: Commit**

```bash
git add .goreleaser.yml .goreleaser.pulumi.yml GNUmakefile
git commit -m "chore(rebrand): update goreleaser project and binary names to mica"
```

---

## Phase 2 — Pulumi bridge

### Task 4: Update Pulumi `ProviderInfo.Name` and constants

**Files:**
- Modify: `pulumi/provider/resources.go`
- Modify: `pulumi/provider/resources_test.go` (only `Name` assertions; `ProviderTypeName: "flashblade"` stays)
- Rename: `pulumi/provider/cmd/pulumi-resource-flashblade/` → `pulumi/provider/cmd/pulumi-resource-mica/`
- Regenerate: schema and Go SDK (target names recorded in Task 0 step 4)

**Why:** Pulumi `Name` field controls (1) the published package name `pulumi-mica`, (2) the resource token prefix `mica:module:Resource`. The `ProviderTypeName` used by the bridge to map upstream TF resources stays `"flashblade"` (matches `Metadata.TypeName` in the TF provider).

- [ ] **Step 0: Discover Pulumi Makefile targets**

Self-contained discovery (re-run even if recorded earlier in Task 0 — fresh subagents have no prior context):

```bash
SCHEMA_TGT=$(grep -E '^(tfgen|gen_schema|schema):' pulumi/Makefile pulumi/provider/Makefile 2>/dev/null | head -1 | cut -d: -f2 | tr -d ' ')
SDK_BUILD_TGT=$(grep -E '^(build_sdks|gen_go_sdk|build_go_sdk):' pulumi/Makefile pulumi/provider/Makefile 2>/dev/null | head -1 | cut -d: -f2 | tr -d ' ')
echo "Schema target: $SCHEMA_TGT"
echo "SDK target: $SDK_BUILD_TGT"
test -n "$SCHEMA_TGT" || { echo "FAIL: schema target not found in pulumi Makefiles"; exit 1; }
test -n "$SDK_BUILD_TGT" || { echo "FAIL: SDK build target not found in pulumi Makefiles"; exit 1; }
```

Export `SCHEMA_TGT` and `SDK_BUILD_TGT` for use in steps 7 and 8 (or note them and substitute manually).

- [ ] **Step 1: Update `mainPkg` constant, `Name`, `DisplayName` in `resources.go`**

Use `Read` then `Edit` (not sed) to preserve original whitespace/alignment, since these fields may be aligned with surrounding fields and `gofmt` enforcement is part of `make lint`:

- Locate `const mainPkg = "flashblade"` → replace with `const mainPkg = "mica"`
- Locate the `Name:` line in the `tfbridge.ProviderInfo{...}` struct literal (whatever its exact spacing is) → replace value `"flashblade"` with `"mica"`, preserve column alignment
- Locate `DisplayName: "FlashBlade",` → replace with `DisplayName: "Mica",`, preserve column alignment

Then enforce formatting:

```bash
gofmt -w pulumi/provider/resources.go
```

- [ ] **Step 2: Update `Keywords`**

Use `Read` and `Edit` on `pulumi/provider/resources.go` to update the Keywords slice. Keep `"flashblade"` and `"pure-storage"` as descriptive search keywords (nominative use), add `"mica"`:

Old: `Keywords:    []string{"pulumi", "flashblade", "pure-storage", "category/infrastructure"},`
New: `Keywords:    []string{"pulumi", "mica", "flashblade", "pure-storage", "category/infrastructure"},`

- [ ] **Step 3: Update `Name` assertions in tests**

Find every test that asserts `Name == "flashblade"` (the Pulumi package name) and update to `"mica"`. Find every test that uses `ProviderTypeName: "flashblade"` (the upstream TF type name) and **leave unchanged**.

```bash
grep -n '"flashblade"' pulumi/provider/resources_test.go
```

For each match, inspect context:

```bash
grep -n -B 1 -A 1 '"flashblade"' pulumi/provider/resources_test.go
```

- If preceded by `Name:` or compares to `prov.Name` → change to `"mica"`.
- If part of `ProviderTypeName: "flashblade"` → keep.
- If a string in Keywords or descriptive text → keep.

Apply each change with `Edit` rather than blanket sed (semantic differentiation required).

- [ ] **Step 4: Build Pulumi provider**

```bash
(cd pulumi/provider && go build ./...)
```

Expected: no errors.

- [ ] **Step 5: Run Pulumi tests**

```bash
(cd pulumi/provider && go test ./... -count=1)
```

Expected: all pass. If any test fails on `Name` assertion, return to step 3.

- [ ] **Step 6: Rename Pulumi command directory**

```bash
if [ -d pulumi/provider/cmd/pulumi-resource-flashblade ]; then
  git mv pulumi/provider/cmd/pulumi-resource-flashblade pulumi/provider/cmd/pulumi-resource-mica
fi
```

- [ ] **Step 7: Regenerate Pulumi schema**

Use the `SCHEMA_TGT` discovered in step 0:

```bash
(cd pulumi && make "$SCHEMA_TGT") 2>&1 | tail -30
```

Expected: schema files regenerated under `pulumi/provider/cmd/pulumi-resource-mica/`. Inspect a generated file:

```bash
find pulumi/provider/cmd/pulumi-resource-mica -name 'schema.json' -exec head -5 {} \;
```

The token prefix should be `mica:`.

- [ ] **Step 8: Regenerate Go SDK**

Use the `SDK_BUILD_TGT` discovered in step 0:

```bash
(cd pulumi && make "$SDK_BUILD_TGT") 2>&1 | tail -30
```

Expected: `pulumi/sdk/go/mica/` directory generated.

- [ ] **Step 9: Remove old Pulumi SDK directory**

```bash
if [ -d pulumi/sdk/go/flashblade ]; then
  git rm -r pulumi/sdk/go/flashblade
fi
```

- [ ] **Step 10: Final Pulumi build & test**

```bash
(cd pulumi/provider && go build ./... && go test ./... -count=1)
[ -d pulumi/sdk/go/mica ] && (cd pulumi/sdk/go && go build ./...)
```

Expected: clean build, tests pass.

- [ ] **Step 11: Commit**

```bash
git add -A pulumi/
git commit -m "feat(rebrand): rename Pulumi provider to mica, regen schema and Go SDK"
```

---

## Phase 3 — License & legal

### Task 5: Add GPL v3 LICENSE and NOTICE

**Files:**
- Create: `LICENSE` (full GPL v3 text)
- Create: `NOTICE` (trademark attributions)

**Why:** GPL v3 strong copyleft prevents closed-source forks. NOTICE documents trademark non-affiliation per Pure Storage guidelines.

- [ ] **Step 1: Download GPL v3 official text**

```bash
curl -sSL https://www.gnu.org/licenses/gpl-3.0.txt -o LICENSE
head -3 LICENSE
wc -l LICENSE
```

Expected: starts with `GNU GENERAL PUBLIC LICENSE` then `Version 3, 29 June 2007`. Line count: ~675.

- [ ] **Step 2: Create `NOTICE`**

Use `Write` tool to create `NOTICE` with the following content:

```
Mica
Copyright (C) 2026 Numberly

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, version 3 of the License.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.

------------------------------------------------------------------------

TRADEMARK NOTICE

Pure Storage(R), FlashBlade(R), and Purity(R) are registered trademarks
of Pure Storage, Inc. and/or its affiliates.

This project, "Mica", is independent and is NOT affiliated with,
endorsed by, sponsored by, or otherwise associated with Pure Storage, Inc.

References to Pure Storage products in this project's documentation are
nominative descriptive references identifying the target system that
this software interacts with, and do not imply any endorsement,
certification, or official status.

For Pure Storage trademark guidelines, see:
  https://www.purestorage.com/legal/productenduserinfo.html

------------------------------------------------------------------------

THIRD-PARTY COMPONENTS

This software uses third-party libraries, each governed by its own
license. See go.mod, go.sum, and pulumi/provider/go.mod for the full
dependency list. Notable upstream components:

- terraform-plugin-framework (Mozilla Public License 2.0)
- terraform-plugin-docs (Mozilla Public License 2.0)
- pulumi-terraform-bridge (Apache License 2.0)

The combination of GPL v3 with MPL 2.0 components is permitted per
Mozilla's official compatibility statement and the Free Software
Foundation's license list.
```

- [ ] **Step 3: Commit**

```bash
git add LICENSE NOTICE
git commit -m "feat(rebrand): adopt GPL v3 license and add trademark NOTICE"
```

---

## Phase 4 — Documentation rebrand

### Task 6: Rebrand README

**Files:**
- Modify: `README.md`

**Why:** README is the public face of the project. Must comply with trademark guidelines (™/®, disclaimer) and reflect new identity.

This task uses the `Edit` tool primarily, since multi-line and complex replacements are unsafe with sed.

- [ ] **Step 1: Read existing README**

```bash
wc -l README.md
```

Then use `Read` tool on `README.md` to capture exact current title and structure.

- [ ] **Step 2: Replace title**

Use `Edit` tool with the exact current first heading (whatever it is — likely `# Terraform Provider FlashBlade`) as `old_string`, replace with:

```markdown
# Mica — Terraform/OpenTofu provider for Pure Storage FlashBlade®
```

- [ ] **Step 3: Insert tagline + trademark notice block**

Use `Edit` to insert immediately after the title (find the line break after the title and append):

```markdown

> Mica is an open-source Terraform and OpenTofu provider for Pure Storage FlashBlade® scale-out storage arrays.
>
> **Mica is independent and is NOT affiliated with, endorsed by, or sponsored by Pure Storage, Inc.**

## Trademarks

Pure Storage®, FlashBlade®, and Purity® are registered trademarks of Pure Storage, Inc. and/or its affiliates. This project uses these names only as nominative descriptive references to identify the target system. See [`NOTICE`](./NOTICE) for full attribution.

## Why the asymmetric naming?

Mica deliberately uses different prefixes between Terraform and Pulumi:

- **Terraform / OpenTofu**: resources are named `flashblade_bucket`, `flashblade_target`, etc. — the prefix describes the target system, following the convention of `aws_*`, `google_*`, `vsphere_*` providers.
- **Pulumi**: resources are exposed under the `mica:` namespace (`mica.NewBucket(...)` in Go, `mica.Bucket(...)` in Python/TypeScript) — the Pulumi package name is `pulumi-mica`.

This asymmetry exists because the Pulumi package name is itself a published artifact (subject to trademark rules), while the Terraform resource type is a code-internal identifier (descriptive nominative use).
```

- [ ] **Step 4: Update HCL `source =` references**

```bash
sed -i 's|source[[:space:]]*=[[:space:]]*"numberly/flashblade"|source = "numberly/mica"|g' README.md
```

Verify:

```bash
grep -n 'numberly/' README.md
```

Expected: all `source = "numberly/mica"`. Repository URLs handled in next step.

- [ ] **Step 5: Update repository URLs and badges**

```bash
sed -i 's|github.com/numberly/opentofu-provider-flashblade|github.com/numberly/terraform-provider-mica|g' README.md
sed -i 's|registry.terraform.io/providers/numberly/flashblade|registry.terraform.io/providers/numberly/mica|g' README.md
sed -i 's|registry.terraform.io/numberly/flashblade|registry.terraform.io/numberly/mica|g' README.md
```

- [ ] **Step 6: Replace branding strings**

```bash
sed -i 's|Terraform Provider FlashBlade|Mica|g' README.md
sed -i 's|terraform-provider-flashblade|terraform-provider-mica|g' README.md
sed -i 's|opentofu-provider-flashblade|terraform-provider-mica|g' README.md
```

- [ ] **Step 7: Verify resource examples still use `flashblade_*` prefix**

```bash
grep -n '^resource "flashblade_' README.md
```

Expected: `resource "flashblade_bucket"`, etc. — these stay.

- [ ] **Step 8: Verify ® on first FlashBlade mention**

```bash
grep -n 'FlashBlade' README.md | head -3
```

The first `FlashBlade` occurrence in the file must include `®`. If not, use `Edit` to add it (`FlashBlade` → `FlashBlade®` for that single first occurrence only — subsequent mentions stay plain per typographic convention).

- [ ] **Step 9: Append License section**

Use `Edit` to append at end of `README.md`:

```markdown

## License

Mica is licensed under the [GNU General Public License v3.0](./LICENSE).

The provider is invoked by Terraform and OpenTofu via gRPC IPC. Your Terraform configurations and infrastructure-as-code do not become subject to GPL v3 simply by using Mica — the IPC boundary is the license boundary, the same way the Linux kernel does not impose GPL on userspace programs.

If you redistribute Mica (binaries or source), you must comply with GPL v3: provide source code or a written offer to provide it, and preserve the LICENSE and NOTICE files.
```

- [ ] **Step 10: Commit**

```bash
git add README.md
git commit -m "docs(rebrand): rebrand README with Mica identity and trademark notices"
```

---

### Task 7: Update internal project documentation

**Files:**
- Modify: `CLAUDE.md`
- Modify: `CONVENTIONS.md`
- Modify: `ROADMAP.md`
- Modify: `pulumi-bridge.md`
- Modify: `pulumi/README.md`
- Modify: `pulumi/CHANGELOG.md`
- Modify: `.planning/PROJECT.md`
- Modify: `.planning/MILESTONES.md`

**Why:** Internal docs must reflect the new module path and repo URL. Trademark mentions get ™/® on first occurrence.

- [ ] **Step 1: Update repository URL references**

```bash
rg -l 'numberly/opentofu-provider-flashblade' --type md | \
  xargs sed -i 's|numberly/opentofu-provider-flashblade|numberly/terraform-provider-mica|g'
```

- [ ] **Step 2: Update binary name references**

```bash
rg -l 'terraform-provider-flashblade' --type md | \
  xargs sed -i 's|terraform-provider-flashblade|terraform-provider-mica|g'
```

- [ ] **Step 3: Update HCL `source =` in markdown**

```bash
rg -l 'numberly/flashblade' --type md | \
  xargs sed -i 's|numberly/flashblade|numberly/mica|g'
```

- [ ] **Step 4: Identify docs lacking ® on first FlashBlade mention**

```bash
for f in $(rg -l 'FlashBlade' --type md); do
  first=$(grep -n 'FlashBlade' "$f" | head -1)
  ln=$(echo "$first" | cut -d: -f1)
  content=$(echo "$first" | cut -d: -f2-)
  if ! echo "$content" | grep -qE 'FlashBlade®|FlashBlade™'; then
    echo "NEEDS FIX: $f:$ln"
    echo "  $content"
  fi
done
```

For each `NEEDS FIX` listing, use `Edit` to add `®` to the first **prose** occurrence in that file.

**Caveat — do NOT add `®` inside code blocks.** Markdown files often contain `FlashBlade` inside fenced code blocks (` ```go `, ` ```text `, ` ```hcl `, etc.) where it represents a Go identifier (e.g. `FlashBladeClient`), an HCL token (`flashblade_bucket`), or copy-pasted log output. Adding `®` there corrupts the snippet.

Workflow per file flagged by the loop above:
1. Open the file with `Read`.
2. Find the first `FlashBlade` occurrence that is in **prose** (between paragraphs, in headings, in tagline) — skip occurrences between ` ``` ` fences.
3. Use `Edit` with sufficient context (≥ 1 line before and after) so `old_string` is unique.
4. Apply the single-character addition `FlashBlade` → `FlashBlade®`.

- [ ] **Step 5: Run tests to ensure no breakage**

```bash
make test
make lint
```

Expected: ≥ 818 tests pass, lint clean.

- [ ] **Step 6: Commit**

```bash
git add -A
git commit -m "docs(rebrand): update internal docs with new repo path and ® symbols"
```

---

### Task 8: Update HCL examples

**Files:**
- Modify: `examples/resources/<name>/resource.tf` (~30 files)
- Modify: `examples/data-sources/<name>/data-source.tf` (~30 files)
- Modify: `examples/provider/provider.tf` if it exists

**Why:** Example HCL files demonstrate `source = "numberly/<slug>"`. The slug must be `mica`. Resource types stay `flashblade_*`.

- [ ] **Step 1: List example files referencing the old source**

```bash
rg -l 'numberly/flashblade' examples/
```

- [ ] **Step 2: Update all example files**

```bash
rg -l 'numberly/flashblade' examples/ | xargs sed -i 's|numberly/flashblade|numberly/mica|g'
```

- [ ] **Step 3: Verify resource types unchanged**

```bash
rg '^resource "flashblade_' examples/ | wc -l
rg '^data "flashblade_' examples/ | wc -l
```

Expected: counts > 0 (unchanged from before).

- [ ] **Step 4: Format**

```bash
terraform fmt -recursive examples/
```

- [ ] **Step 5: Commit**

```bash
git add examples/
git commit -m "docs(rebrand): update example HCL source paths to numberly/mica"
```

---

### Task 9: Regenerate Terraform documentation

**Files:**
- Modify: `docs/index.md`, `docs/resources/*.md`, `docs/data-sources/*.md` (auto-generated)

**Why:** `make docs` reads provider schema metadata (now `mica`) and emits updated `docs/`.

- [ ] **Step 1: Run `make docs`**

```bash
make docs
```

- [ ] **Step 2: Sanity-check regenerated docs**

```bash
head -5 docs/index.md
grep -l 'numberly/mica' docs/index.md
RES_PAGES=$(ls docs/resources/*.md | wc -l)
RES_HEADERS=$(grep -lE '^# *flashblade_' docs/resources/*.md | wc -l)
echo "Resource pages: $RES_PAGES, with flashblade_ heading: $RES_HEADERS"
test "$RES_PAGES" -eq "$RES_HEADERS" || echo "WARN: some resource pages missing flashblade_ heading"
```

Expected:
- `docs/index.md` mentions `numberly/mica` in install snippet
- Every resource page has heading starting with `flashblade_`

- [ ] **Step 3: Confirm idempotence**

```bash
make docs
git status docs/
```

Second run must produce no diff.

- [ ] **Step 4: Commit**

```bash
git add docs/
git commit -m "docs(rebrand): regenerate tfplugindocs with new provider identity"
```

---

## Phase 5 — CI/CD

### Task 10: Update GitHub Actions workflows

**Files:**
- Modify: `.github/workflows/ci.yml`
- Modify: `.github/workflows/release.yml`
- Modify: `.github/workflows/pulumi-ci.yml`
- Modify: `.github/workflows/pulumi-prerequisites.yml`
- Modify: `.github/workflows/pulumi-release.yml`

**Why:** Workflows reference repo paths, artifact names, and provider name strings.

- [ ] **Step 1: Find references**

```bash
rg -n 'flashblade|opentofu-provider' .github/workflows/
```

- [ ] **Step 2: Update repo paths**

```bash
rg -l 'numberly/opentofu-provider-flashblade' .github/workflows/ | \
  xargs sed -i 's|numberly/opentofu-provider-flashblade|numberly/terraform-provider-mica|g'
```

- [ ] **Step 3: Update binary/artifact name references**

```bash
rg -l 'terraform-provider-flashblade' .github/workflows/ | \
  xargs sed -i 's|terraform-provider-flashblade|terraform-provider-mica|g'

rg -l 'pulumi-resource-flashblade' .github/workflows/ | \
  xargs sed -i 's|pulumi-resource-flashblade|pulumi-resource-mica|g'
```

- [ ] **Step 4: Inspect remaining `flashblade` matches case-by-case**

```bash
rg -n 'flashblade' .github/workflows/
```

For each match:
- SDK directory references (`pulumi/sdk/go/flashblade`): replace with `mica`
- Tag/release naming patterns (`v*-flashblade-*`): replace with `mica`
- Comments / descriptive text: leave as-is (nominative)

Apply with `Edit` per file (semantic differentiation required).

- [ ] **Step 5: Validate workflow YAML**

PyYAML is third-party and may not be installed. Try in order: PyYAML, then `yq`, then skip with a warning.

```bash
if python3 -c "import yaml" 2>/dev/null; then
  for f in .github/workflows/*.yml; do
    python3 -c "import yaml; yaml.safe_load(open('$f'))" && echo "OK: $f" || echo "BROKEN: $f"
  done
elif command -v yq >/dev/null 2>&1; then
  for f in .github/workflows/*.yml; do
    yq eval '.' "$f" >/dev/null && echo "OK: $f" || echo "BROKEN: $f"
  done
else
  echo "WARN: no YAML validator available (install PyYAML via 'pip install pyyaml' or yq)"
  echo "      skipping local validation; GitHub Actions will report syntax errors on push"
fi
```

Expected: all `OK` if a validator is present, otherwise the warning is acceptable for this single workflow change.

- [ ] **Step 6: Commit**

```bash
git add .github/workflows/
git commit -m "ci(rebrand): align GitHub Actions workflows with terraform-provider-mica"
```

---

## Phase 6 — Validation & release prep

### Task 11: Final validation pass

**Files:** none (read-only validation)

- [ ] **Step 1: Lint clean**

```bash
make lint
```

Expected: exit 0.

- [ ] **Step 2: Test count meets baseline**

```bash
COUNT=$(go test ./... -count=1 -v 2>&1 | grep -c '^=== RUN  ')
echo "Final test count: $COUNT"
test "$COUNT" -ge 818 || { echo "FAIL: test count regressed below 818"; exit 1; }
```

- [ ] **Step 3: Goreleaser produces correctly-named binary**

```bash
goreleaser build --snapshot --clean --single-target
find dist -name 'terraform-provider-mica_*' -type f
file $(find dist -name 'terraform-provider-mica_*' -type f | head -1)
```

Expected:
- At least one matching file
- `file` reports it as an executable for the current platform

- [ ] **Step 4: Documentation regen idempotent**

```bash
make docs
git status docs/
```

Expected: empty diff.

- [ ] **Step 5: Pulumi bridge build & test**

```bash
(cd pulumi/provider && go build ./... && go test ./... -count=1)
[ -d pulumi/sdk/go/mica ] && (cd pulumi/sdk/go && go build ./...)
```

Expected: clean build, tests pass.

- [ ] **Step 6: Verify zero straggling old-name references in code/config/CI**

```bash
echo "=== Go imports (must be empty) ==="
rg 'numberly/opentofu-provider-flashblade' --type go

echo "=== Markdown URLs (must be empty) ==="
rg 'numberly/opentofu-provider-flashblade' --type md

echo "=== HCL sources (must be empty) ==="
rg 'numberly/flashblade' -g '*.tf'

echo "=== Workflows (must be empty) ==="
rg 'numberly/opentofu-provider-flashblade' .github/workflows/

echo "=== Old SDK directories (must be empty) ==="
ls pulumi/sdk/go/flashblade 2>&1
ls pulumi/provider/cmd/pulumi-resource-flashblade 2>&1
```

Each section must report no matches / "No such file or directory". Note: `flashblade_*` resource type names ARE expected to remain — those are the descriptive nominative prefix.

- [ ] **Step 7: No commit**

If any check fails, create a fixup commit before proceeding.

---

### Task 12: CHANGELOG entry and tag preparation

**Files:**
- Modify: `CHANGELOG.md`
- Modify: `pulumi/CHANGELOG.md`

**Why:** Documents the rebrand release for future open-source users; tag `v2.22.4` will reference these entries.

- [ ] **Step 1: Inspect existing CHANGELOG format**

```bash
head -30 CHANGELOG.md 2>/dev/null || echo "no CHANGELOG"
```

- [ ] **Step 2: Prepend new entry to root `CHANGELOG.md`**

Use `Edit` to insert at the top (after any unchanged header section):

```markdown
## [2.22.4] — 2026-04-28

### Project rebrand

This release renames the project to **Mica** for open-source release. The provider continues to target Pure Storage FlashBlade® arrays exactly as before.

### Changed (breaking)

- **Registry source path**: `numberly/flashblade` → `numberly/mica`
- **Go module path**: `github.com/numberly/opentofu-provider-flashblade` → `github.com/numberly/terraform-provider-mica`
- **Pulumi package name**: `pulumi-flashblade` → `pulumi-mica`
- **Pulumi resource tokens**: `flashblade:*:*` → `mica:*:*`
- **License**: now distributed under **GPL v3** (was: unspecified)

### Unchanged

- Terraform resource type names: `flashblade_bucket`, `flashblade_target`, `flashblade_filesystem`, etc.
- HCL `provider "flashblade" {}` block syntax (the local alias remains user-controlled)
- Internal Go identifiers (`FlashBladeClient`, package layout, etc.)
- All schema fields, behaviors, and acceptance test fixtures

### Migration

Update the `source` field in `required_providers`:

```hcl
terraform {
  required_providers {
    flashblade = {
      source  = "numberly/mica"   # was: "numberly/flashblade"
      version = "2.22.4"
    }
  }
}
```

Then migrate existing state:

```bash
terraform init
terraform state replace-provider numberly/flashblade numberly/mica
```

`replace-provider` rewrites every resource's recorded provider reference. Without this step, `terraform plan` will fail with a provider mismatch error.

### Versioning note

This project tracks the upstream FlashBlade API version as `MAJOR.MINOR.PATCH`. Despite the patch-level bump in `v2.22.4`, this release contains breaking changes (registry source path, module path, license). Pin exactly with `version = "2.22.4"` rather than `~> 2.22.4` if you want to control migration timing.
```

- [ ] **Step 3: Prepend new entry to `pulumi/CHANGELOG.md`**

Use `Edit` to insert at the top:

```markdown
## [2.22.4] — 2026-04-28

### Project rebrand

The Pulumi provider for Pure Storage FlashBlade® has been renamed from `pulumi-flashblade` to `pulumi-mica`.

### Changed (breaking)

- Pulumi package name: `pulumi-flashblade` → `pulumi-mica`
- Resource token namespace: `flashblade:*:*` → `mica:*:*`
- Go SDK import path: `github.com/numberly/opentofu-provider-flashblade/pulumi/sdk/go/flashblade` → `github.com/numberly/terraform-provider-mica/pulumi/sdk/go/mica`
- License: now distributed under **GPL v3**

### Migration

Pulumi does not provide a built-in `replace-provider` equivalent for renamed type tokens. Existing stacks reference `flashblade:*:*` resources by type token in state, and a fresh import is the safest path.

For each existing stack:

1. Export current stack state: `pulumi stack export --file old-state.json`
2. In `old-state.json`, search-and-replace `"flashblade:` with `"mica:` (this rewrites the resource URN type token).
3. Also rewrite the Go SDK import paths in your IaC code.
4. Update Pulumi.yaml or package.json to depend on `pulumi-mica` instead of `pulumi-flashblade`.
5. Import: `pulumi stack import --file old-state.json`
6. Run `pulumi preview` to verify no diffs are detected.

If diffs appear, the rename was incomplete — investigate before applying.
```

- [ ] **Step 4: Commit**

```bash
git add CHANGELOG.md pulumi/CHANGELOG.md
git commit -m "docs(rebrand): add v2.22.4 CHANGELOG entries with migration steps"
```

---

### Task 13: Tag preparation

**Files:** none (tag created locally; user pushes manually)

**Why:** Per agreement, the user (Guillaume) handles publication and registry submission outside this plan.

- [ ] **Step 1: Verify clean state**

```bash
git status --short
git log --oneline -15
```

Expected: clean working tree, last ~12 commits show the rebrand sequence.

- [ ] **Step 2: Create annotated tag (DO NOT PUSH)**

```bash
git tag -a v2.22.4 -m "Mica rebrand: rename to terraform-provider-mica, switch to GPL v3"
git tag -l v2.22.4
```

- [ ] **Step 3: Heads-up for user (out-of-band)**

Print to the user during execution:

> ⚠️ **Logo required for Terraform Registry submission.** The Terraform Registry requires a logo SVG/PNG for provider listings. This plan does not generate one. Before submitting to <https://registry.terraform.io/publish/provider>, create a Mica logo (avoid Pure Storage "P" iconography or look-alike styles) and include it in the repo per Registry submission guidelines.

---

### Task 14: Generate personal migration commands for the project owner

**Files:** none (output only — do not write to disk unless requested)

**Why:** The project owner is currently the sole user. They will migrate their personal Terraform state and Pulumi stacks separately, outside this rebrand workflow. This task gathers the exact commands they need into a single output, so they don't have to reverse-engineer them from the CHANGELOG.

- [ ] **Step 1: Detect owner's likely Terraform state references**

Scope the search to common workspace locations to avoid scanning the entire home directory:

```bash
echo "=== Likely Terraform workspace candidates ==="
for d in "$HOME/Workspace" "$HOME/projects" "$HOME/code" "$HOME/dev" "$PWD/.."; do
  [ -d "$d" ] && find "$d" -maxdepth 5 -name '*.tfstate' -type f 2>/dev/null
done | sort -u | head -20

echo ""
echo "=== Likely Terraform configs referencing flashblade ==="
for d in "$HOME/Workspace" "$HOME/projects" "$HOME/code" "$HOME/dev"; do
  [ -d "$d" ] && rg -l --type-add 'tf:*.tf' --type tf 'numberly/flashblade' "$d" 2>/dev/null
done | sort -u | head -20
```

The output is informational — do not modify any user files. The point is to surface candidate paths so the owner can run the migration commands in the right places.

- [ ] **Step 2: Print Terraform migration commands**

Output to the conversation (verbatim — the user will copy/paste):

```
=================================================================
PERSONAL MIGRATION GUIDE — Mica rebrand (v2.22.4)
=================================================================

For each Terraform/OpenTofu workspace where you used numberly/flashblade:

# 1. Update the required_providers block
sed -i 's|"numberly/flashblade"|"numberly/mica"|g' *.tf

# 2. Re-initialize with the new source
terraform init -upgrade
# or: tofu init -upgrade

# 3. Rewrite state to point at the new provider address
terraform state replace-provider \
  registry.terraform.io/numberly/flashblade \
  registry.terraform.io/numberly/mica
# or: tofu state replace-provider ...

# 4. Verify no drift
terraform plan
# Expected: "No changes. Your infrastructure matches the configuration."

If `terraform plan` still complains about provider mismatch, check
that no nested module declares its own required_providers block with
the old source path.
```

- [ ] **Step 3: Print Pulumi migration commands**

Output to the conversation:

```
=================================================================
For each Pulumi stack where you used pulumi-flashblade:
=================================================================

# 1. Update your Pulumi project dependencies
#    - Go: replace import paths in your .go files
#         github.com/numberly/opentofu-provider-flashblade/pulumi/sdk/go/flashblade
#       → github.com/numberly/terraform-provider-mica/pulumi/sdk/go/mica
#    - Python/TS: update package.json/requirements.txt to depend on pulumi-mica
#                 instead of pulumi-flashblade

# 2. Export current stack state
pulumi stack export --file mica-migration.json

# 3a. Rewrite resource type tokens
sed -i 's|"flashblade:|"mica:|g' mica-migration.json

# 3b. Rewrite provider URNs (separate prefix, safe to sed in one pass)
#     Pulumi state stores provider refs as URNs containing the segment
#     `pulumi:providers:<name>::`. The prefix `pulumi:providers:` is
#     unique enough to sed without collateral damage:
grep -c 'pulumi:providers:flashblade' mica-migration.json   # confirm count > 0
sed -i 's|pulumi:providers:flashblade|pulumi:providers:mica|g' mica-migration.json

# 4. Import the rewritten state
pulumi stack import --file mica-migration.json

# 5. Verify no diffs
pulumi preview
# Expected: "no changes" — same infrastructure, new package name.

If `pulumi preview` shows resource recreations, do NOT apply.
The rename was incomplete; inspect the URNs in mica-migration.json.

=================================================================
END OF PERSONAL MIGRATION GUIDE
=================================================================
```

- [ ] **Step 4: No commit (output is conversational, not source)**

---

## Self-review checklist

After completing all tasks, verify:

- [ ] No `opentofu-provider-flashblade` Go import paths remain (Task 11 step 6).
- [ ] No `numberly/flashblade` source paths in HCL/docs (Task 11 step 6).
- [ ] No `pulumi-resource-flashblade` directory; `pulumi-resource-mica` exists.
- [ ] No `pulumi/sdk/go/flashblade` directory; `pulumi/sdk/go/mica` exists.
- [ ] `make lint` exits 0 (Task 11 step 1).
- [ ] Test count ≥ 818 baseline (Task 11 step 2).
- [ ] Goreleaser produces `terraform-provider-mica_*` binary (Task 11 step 3).
- [ ] `make docs` is idempotent (Task 11 step 4).
- [ ] Pulumi bridge builds & tests pass (Task 11 step 5).
- [ ] `LICENSE` is GPL v3 full text (Task 5 step 1).
- [ ] `NOTICE` documents Pure Storage trademark non-affiliation (Task 5 step 2).
- [ ] README has tagline + trademark section + asymmetric naming explainer + ® on first FlashBlade mention + License section (Task 6).
- [ ] CHANGELOG.md has v2.22.4 entry with `terraform state replace-provider` migration command (Task 12 step 2).
- [ ] pulumi/CHANGELOG.md has v2.22.4 entry with Pulumi state migration steps (Task 12 step 3).
- [ ] Tag `v2.22.4` created locally, NOT pushed (Task 13 step 2).
- [ ] Logo warning printed to user (Task 13 step 3).
- [ ] Personal migration commands printed for the owner (Task 14 steps 2-3).

## Out of scope (per design spec)

The following are explicitly not part of this plan:

- Renaming internal Go types (`FlashBladeClient`, `FlashBladeClientConfig`, etc.) or `internal/provider/<resource>_resource.go` filenames.
- Renaming the `flashblade_*` resource/data source type prefix.
- Creating a logo asset.
- Pushing the tag, renaming the GitHub repo, submitting to Terraform/OpenTofu Registry.
- Generating user-facing release notes beyond CHANGELOG entries.
