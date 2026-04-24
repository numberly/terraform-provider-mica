---
phase: 54-bridge-bootstrap-poc-3-resources
plan: 03
type: execute
wave: 3
depends_on: [02]
files_modified:
  - pulumi/provider/cmd/pulumi-tfgen-flashblade/main.go
  - pulumi/provider/cmd/pulumi-resource-flashblade/main.go
autonomous: true
requirements: [BRIDGE-04]
must_haves:
  truths:
    - "pulumi-tfgen-flashblade binary builds and entry-points pf/tfgen.Main"
    - "pulumi-resource-flashblade binary builds and entry-points pf/tfbridge.Main with embedded schema + metadata"
    - "Runtime main.go uses //go:embed for both schema-embed.json and bridge-metadata.json"
  artifacts:
    - path: "pulumi/provider/cmd/pulumi-tfgen-flashblade/main.go"
      provides: "Build-time schema generator entry"
      contains: "tfgen.Main"
    - path: "pulumi/provider/cmd/pulumi-resource-flashblade/main.go"
      provides: "Runtime gRPC provider entry"
      contains: "pftfbridge.Main"
  key_links:
    - from: "cmd/pulumi-resource-flashblade/main.go"
      to: "schema-embed.json + bridge-metadata.json"
      via: "//go:embed"
      pattern: "//go:embed schema-embed.json"
    - from: "cmd/main.go files"
      to: "pulumi/provider.Provider()"
      via: "Go import"
      pattern: "flashblade\\.Provider\\(\\)"
---

<objective>
Create the two binary entry points: `pulumi-tfgen-flashblade` (build-time schema generator) and `pulumi-resource-flashblade` (runtime Pulumi plugin). Both import `pulumi/provider.Provider()` defined in plan 02.

Purpose: Until these exist, `make tfgen` has no binary to execute and schema generation cannot happen. Plan 04 depends on a buildable tfgen binary.
Output: `cd pulumi/provider && go build ./cmd/pulumi-tfgen-flashblade` succeeds. Runtime binary build expects embed targets to exist — provide placeholder files so compilation succeeds before first tfgen run.
</objective>

<execution_context>
@$HOME/.claude/get-shit-done/workflows/execute-plan.md
</execution_context>

<context>
@.planning/phases/54-bridge-bootstrap-poc-3-resources/54-CONTEXT.md
@.planning/research/STACK.md
@pulumi/provider/resources.go
</context>

<tasks>

<task type="auto">
  <name>Task 1: Create pulumi-tfgen-flashblade main.go</name>
  <read_first>
    - /home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/pulumi/provider/resources.go (confirm `Provider()` signature and package name `provider`)
    - /home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/.planning/research/STACK.md (lines ~233-246 — PF tfgen signature)
  </read_first>
  <files>
    - pulumi/provider/cmd/pulumi-tfgen-flashblade/main.go
  </files>
  <action>
    Create `pulumi/provider/cmd/pulumi-tfgen-flashblade/main.go` with:

    ```go
    package main

    import (
        "github.com/pulumi/pulumi-terraform-bridge/v3/pkg/pf/tfgen"

        flashblade "github.com/numberly/opentofu-provider-flashblade/pulumi/provider"
    )

    func main() {
        // PF tfgen does not take a version parameter (differs from SDK v2 bridge).
        // Version is injected into the runtime binary via ldflags -X.
        tfgen.Main("flashblade", flashblade.Provider())
    }
    ```

    Verify build: `cd pulumi/provider && go build ./cmd/pulumi-tfgen-flashblade`. Expect success.
  </action>
  <verify>
    <automated>test -f pulumi/provider/cmd/pulumi-tfgen-flashblade/main.go && grep -q 'tfgen.Main("flashblade"' pulumi/provider/cmd/pulumi-tfgen-flashblade/main.go && grep -q 'numberly/opentofu-provider-flashblade/pulumi/provider' pulumi/provider/cmd/pulumi-tfgen-flashblade/main.go && cd pulumi/provider && go build ./cmd/pulumi-tfgen-flashblade</automated>
  </verify>
  <done>
    - File exists with `tfgen.Main("flashblade", flashblade.Provider())`
    - `go build ./cmd/pulumi-tfgen-flashblade` succeeds
    - Binary `pulumi/provider/cmd/pulumi-tfgen-flashblade/pulumi-tfgen-flashblade` produced
  </done>
</task>

<task type="auto">
  <name>Task 2: Create pulumi-resource-flashblade main.go with //go:embed placeholders</name>
  <read_first>
    - /home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/pulumi/provider/resources.go
    - /home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/.planning/research/STACK.md (lines ~248-269 — PF tfbridge.Main signature with ProviderMetadata)
  </read_first>
  <files>
    - pulumi/provider/cmd/pulumi-resource-flashblade/main.go
    - pulumi/provider/cmd/pulumi-resource-flashblade/schema-embed.json
    - pulumi/provider/cmd/pulumi-resource-flashblade/bridge-metadata.json
  </files>
  <action>
    Create `pulumi/provider/cmd/pulumi-resource-flashblade/main.go` with:

    ```go
    package main

    import (
        "context"
        _ "embed"

        pftfbridge "github.com/pulumi/pulumi-terraform-bridge/v3/pkg/pf/tfbridge"

        flashblade "github.com/numberly/opentofu-provider-flashblade/pulumi/provider"
    )

    //go:embed schema-embed.json
    var pulumiSchema []byte

    //go:embed bridge-metadata.json
    var bridgeMetadata []byte

    func main() {
        meta := pftfbridge.ProviderMetadata{
            PackageSchema:  pulumiSchema,
            BridgeMetadata: bridgeMetadata,
        }
        if err := pftfbridge.Main(
            context.Background(),
            "flashblade",
            flashblade.Provider(),
            meta,
        ); err != nil {
            panic(err)
        }
    }
    ```

    Create placeholder embed files so `go build` succeeds before the first `make tfgen` run (real content is generated in plan 04):

    `pulumi/provider/cmd/pulumi-resource-flashblade/schema-embed.json`:
    ```
    {}
    ```

    `pulumi/provider/cmd/pulumi-resource-flashblade/bridge-metadata.json`:
    ```
    {}
    ```

    These placeholders are overwritten by `make tfgen` in plan 04.

    Verify build: `cd pulumi/provider && go build ./cmd/pulumi-resource-flashblade`. Expect success (the binary will not function correctly until real schema is generated, but compilation must pass).

    NOTE on pftfbridge.Main signature: the bridge pf/tfbridge API has seen minor param-shape changes across v3.x. If `pftfbridge.Main(ctx, name, providerInfo, meta)` does not compile against v3.127.0, fall back to the signature the compiler reports (likely one of: with/without error return, named struct args). Adjust the call accordingly but keep the overall wiring (embedded PackageSchema + BridgeMetadata fed to Main).
  </action>
  <verify>
    <automated>test -f pulumi/provider/cmd/pulumi-resource-flashblade/main.go && test -f pulumi/provider/cmd/pulumi-resource-flashblade/schema-embed.json && test -f pulumi/provider/cmd/pulumi-resource-flashblade/bridge-metadata.json && grep -q '//go:embed schema-embed.json' pulumi/provider/cmd/pulumi-resource-flashblade/main.go && grep -q '//go:embed bridge-metadata.json' pulumi/provider/cmd/pulumi-resource-flashblade/main.go && grep -q 'pftfbridge.Main' pulumi/provider/cmd/pulumi-resource-flashblade/main.go && cd pulumi/provider && go build ./cmd/pulumi-resource-flashblade</automated>
  </verify>
  <done>
    - Runtime main.go exists with both //go:embed directives
    - Placeholder `{}` JSON files exist for both embed targets
    - `go build ./cmd/pulumi-resource-flashblade` succeeds
  </done>
</task>

</tasks>

<verification>
- Both main.go files exist
- Both `go build ./cmd/...` succeed
- Binaries produced under `pulumi/provider/cmd/*/pulumi-*-flashblade`
</verification>

<success_criteria>
- BRIDGE-04 satisfied: two entry points exist and compile
- Runtime binary embeds placeholder schema (ready to be replaced by plan 04's `make tfgen`)
</success_criteria>

<output>
After completion, create `.planning/phases/54-bridge-bootstrap-poc-3-resources/54-03-SUMMARY.md`
</output>
