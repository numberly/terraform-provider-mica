---
phase: 54-bridge-bootstrap-poc-3-resources
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - pulumi/.gitignore
  - pulumi/provider/go.mod
  - pulumi/provider/pkg/version/version.go
  - pulumi/sdk/go/go.mod
  - pulumi/Makefile
autonomous: true
requirements: [BRIDGE-02, BRIDGE-03]
must_haves:
  truths:
    - "./pulumi/ sub-directory exists with three-module layout (provider/, sdk/go/, Makefile)"
    - "pulumi/provider/go.mod declares correct module path and pins bridge v3.127.0, sdk/pkg v3.231.0, replace SHA v2.0.0-20260318212141-5525259d096b"
    - "pulumi/sdk/go/go.mod is lean (only pulumi/sdk/v3 dep)"
    - "version.go exposes a Version var injectable via -ldflags"
    - "Makefile defines VERSION via git describe --tags --dirty --always (no pulumictl)"
  artifacts:
    - path: "pulumi/provider/go.mod"
      provides: "Bridge module manifest"
      contains: "github.com/numberly/opentofu-provider-flashblade/pulumi/provider"
    - path: "pulumi/provider/pkg/version/version.go"
      provides: "Version var for ldflags injection"
      contains: "var Version"
    - path: "pulumi/sdk/go/go.mod"
      provides: "Lean consumer Go SDK module"
      contains: "github.com/numberly/opentofu-provider-flashblade/pulumi/sdk/go"
    - path: "pulumi/Makefile"
      provides: "Build orchestration skeleton with VERSION computation"
      contains: "git describe"
  key_links:
    - from: "pulumi/provider/go.mod"
      to: "root go.mod (github.com/numberly/opentofu-provider-flashblade)"
      via: "replace ../../"
      pattern: "replace.*numberly/opentofu-provider-flashblade.*=>.*\\.\\./\\.\\."
---

<objective>
Create the `./pulumi/` subtree with three-module Go layout, version injection skeleton, and Makefile VERSION target. No Go source (beyond version.go) or bridge code yet — just the build/dependency skeleton that later plans fill in.

Purpose: Establish dependency pins and module paths exactly once so downstream plans (resources.go, binaries, tests) can `go build` incrementally without chasing version drift (PB4 mitigation).
Output: Buildable but empty module skeleton — `cd pulumi/provider && go mod tidy` must succeed after plan 02 adds resources.go, but this plan's go.mod + go.sum may remain minimal (require blocks only) until plan 02.
</objective>

<execution_context>
@$HOME/.claude/get-shit-done/workflows/execute-plan.md
</execution_context>

<context>
@.planning/phases/54-bridge-bootstrap-poc-3-resources/54-CONTEXT.md
@.planning/research/STACK.md
@go.mod
</context>

<tasks>

<task type="auto">
  <name>Task 1: Create pulumi/provider/go.mod and version.go</name>
  <read_first>
    - /home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/go.mod (to confirm module path `github.com/numberly/opentofu-provider-flashblade` and Go toolchain version)
    - /home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/.planning/phases/54-bridge-bootstrap-poc-3-resources/54-CONTEXT.md (locked decisions D-03, pinned versions)
  </read_first>
  <files>
    - pulumi/provider/go.mod
    - pulumi/provider/pkg/version/version.go
    - pulumi/.gitignore
  </files>
  <action>
    Create `pulumi/provider/go.mod` with exactly:

    ```
    module github.com/numberly/opentofu-provider-flashblade/pulumi/provider

    go 1.25

    require (
        github.com/hashicorp/terraform-plugin-sdk/v2 v2.0.0-20260318212141-5525259d096b
        github.com/pulumi/pulumi-terraform-bridge/v3 v3.127.0
        github.com/pulumi/pulumi/pkg/v3 v3.231.0
        github.com/pulumi/pulumi/sdk/v3 v3.231.0
        github.com/numberly/opentofu-provider-flashblade v0.0.0
    )

    replace (
        github.com/hashicorp/terraform-plugin-sdk/v2 => github.com/pulumi/terraform-plugin-sdk/v2 v2.0.0-20260318212141-5525259d096b
        github.com/numberly/opentofu-provider-flashblade => ../../
    )
    ```

    Note: `v0.0.0` pseudo-version on the local require is required so `go mod` accepts the replace — real resolution comes from `../../`.

    Create `pulumi/provider/pkg/version/version.go`:

    ```go
    // Package version exposes the provider build version.
    // Value is injected at build time via -ldflags "-X .../pkg/version.Version=<git describe>".
    package version

    var Version = "dev"
    ```

    Create `pulumi/.gitignore`:

    ```
    # Build artifacts
    /provider/cmd/pulumi-tfgen-flashblade/pulumi-tfgen-flashblade
    /provider/cmd/pulumi-resource-flashblade/pulumi-resource-flashblade

    # Python/TS/Go SDK generated output (Phase 56)
    /sdk/go/flashblade/
    /sdk/python/bin/
    /sdk/python/pulumi_flashblade.egg-info/
    /sdk/python/pulumi_flashblade/
    /sdk/nodejs/
    /sdk/dotnet/

    # Tooling
    .pulumi/
    ```

    Per D-03: `-X` target path is `github.com/numberly/opentofu-provider-flashblade/pulumi/provider/pkg/version.Version`.

    DO NOT run `go mod tidy` in this task — `go mod tidy` requires at least one .go file that imports the declared deps, which plan 02 provides.
  </action>
  <verify>
    <automated>test -f pulumi/provider/go.mod && test -f pulumi/provider/pkg/version/version.go && test -f pulumi/.gitignore && grep -q "v3.127.0" pulumi/provider/go.mod && grep -q "v3.231.0" pulumi/provider/go.mod && grep -q "5525259d096b" pulumi/provider/go.mod && grep -q "numberly/opentofu-provider-flashblade" pulumi/provider/go.mod && grep -q "=> ../../" pulumi/provider/go.mod && grep -q 'var Version' pulumi/provider/pkg/version/version.go</automated>
  </verify>
  <done>
    - `pulumi/provider/go.mod` exists with bridge v3.127.0, sdk/pkg v3.231.0, replace SHA 20260318212141-5525259d096b, replace `../../` for root module
    - `pulumi/provider/pkg/version/version.go` exports `var Version = "dev"`
    - `pulumi/.gitignore` excludes build binaries and generated SDK content
  </done>
</task>

<task type="auto">
  <name>Task 2: Create pulumi/sdk/go/go.mod (lean consumer module)</name>
  <read_first>
    - /home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/.planning/research/STACK.md (section "Go SDK Distribution (Private)" — lean go.mod shape)
  </read_first>
  <files>
    - pulumi/sdk/go/go.mod
  </files>
  <action>
    Create `pulumi/sdk/go/go.mod` with exactly:

    ```
    module github.com/numberly/opentofu-provider-flashblade/pulumi/sdk/go

    go 1.25

    require github.com/pulumi/pulumi/sdk/v3 v3.231.0
    ```

    This is a placeholder — Phase 56 generates the actual SDK Go source. Keeping only the `pulumi/sdk/v3` dependency ensures consumers don't transitively pull the bridge or TF framework.

    No go.sum is generated until source files exist.
  </action>
  <verify>
    <automated>test -f pulumi/sdk/go/go.mod && grep -q "numberly/opentofu-provider-flashblade/pulumi/sdk/go" pulumi/sdk/go/go.mod && grep -q "pulumi/sdk/v3 v3.231.0" pulumi/sdk/go/go.mod && ! grep -q "pulumi-terraform-bridge" pulumi/sdk/go/go.mod</automated>
  </verify>
  <done>
    - `pulumi/sdk/go/go.mod` declares module path `github.com/numberly/opentofu-provider-flashblade/pulumi/sdk/go`
    - Contains only `pulumi/sdk/v3 v3.231.0` dep (no bridge, no TF framework)
  </done>
</task>

<task type="auto">
  <name>Task 3: Create pulumi/Makefile skeleton with VERSION target</name>
  <read_first>
    - /home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/.planning/phases/54-bridge-bootstrap-poc-3-resources/54-CONTEXT.md (D-03: version via `git describe --tags --dirty --always`, no pulumictl)
    - /home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/.goreleaser.yml (existing git-describe pattern to mirror)
  </read_first>
  <files>
    - pulumi/Makefile
  </files>
  <action>
    Create `pulumi/Makefile` with the following skeleton (build/tfgen targets are stubbed — plan 04 fills them in):

    ```makefile
    PROJECT_NAME := pulumi-flashblade
    PROVIDER_PKG := github.com/numberly/opentofu-provider-flashblade/pulumi/provider
    VERSION := $(shell git describe --tags --dirty --always 2>/dev/null || echo dev)
    LDFLAGS := -ldflags "-X $(PROVIDER_PKG)/pkg/version.Version=$(VERSION)"

    PROVIDER_DIR := provider
    TFGEN_BIN   := $(PROVIDER_DIR)/cmd/pulumi-tfgen-flashblade/pulumi-tfgen-flashblade
    RUNTIME_BIN := $(PROVIDER_DIR)/cmd/pulumi-resource-flashblade/pulumi-resource-flashblade
    SCHEMA_DIR  := $(PROVIDER_DIR)/cmd/pulumi-resource-flashblade

    .PHONY: version tfgen provider test clean

    version:
    	@echo $(VERSION)

    # tfgen, provider, test, clean implementations are added by plan 04 + plan 05.
    tfgen:
    	@echo "tfgen target not yet implemented — see plan 04"
    	@exit 1

    provider:
    	@echo "provider target not yet implemented — see plan 04"
    	@exit 1

    test:
    	@echo "test target not yet implemented — see plan 05"
    	@exit 1

    clean:
    	rm -f $(TFGEN_BIN) $(RUNTIME_BIN)
    	rm -f $(SCHEMA_DIR)/schema.json $(SCHEMA_DIR)/schema-embed.json $(SCHEMA_DIR)/bridge-metadata.json
    ```

    Per D-03: the `-X` target must match the package path declared in version.go, i.e. `github.com/numberly/opentofu-provider-flashblade/pulumi/provider/pkg/version.Version`.

    Verify `make -C pulumi version` prints something (either a git-describe output or `dev`).
  </action>
  <verify>
    <automated>test -f pulumi/Makefile && grep -q "git describe --tags --dirty --always" pulumi/Makefile && grep -q "pulumi/provider/pkg/version.Version" pulumi/Makefile && make -C pulumi version</automated>
  </verify>
  <done>
    - `pulumi/Makefile` exists
    - `VERSION` computed from `git describe --tags --dirty --always` with `dev` fallback
    - `LDFLAGS` targets `github.com/numberly/opentofu-provider-flashblade/pulumi/provider/pkg/version.Version`
    - `make -C pulumi version` prints a non-empty string
    - `tfgen`, `provider`, `test` targets are stubbed with clear "see plan N" messages
  </done>
</task>

</tasks>

<verification>
- Directory layout: `pulumi/{provider,sdk/go}` exists with go.mod in each
- `grep -q v3.127.0 pulumi/provider/go.mod` passes
- `grep -q v3.231.0 pulumi/provider/go.mod` passes
- `grep -q 5525259d096b pulumi/provider/go.mod` passes
- `make -C pulumi version` prints a non-empty VERSION string
- No `pulumictl` reference anywhere in pulumi/
</verification>

<success_criteria>
- All 3 go.mod/Makefile/version.go files in place
- Module paths consistent (`github.com/numberly/opentofu-provider-flashblade/pulumi/provider` and `.../pulumi/sdk/go`)
- Replace directive pins the mandatory Pulumi fork SHA
- Downstream plan 02 can add `resources.go` and run `go mod tidy` without go.mod edits
</success_criteria>

<output>
After completion, create `.planning/phases/54-bridge-bootstrap-poc-3-resources/54-01-SUMMARY.md`
</output>
