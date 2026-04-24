---
phase: 54-bridge-bootstrap-poc-3-resources
plan: 04
type: execute
wave: 4
depends_on: [03]
files_modified:
  - pulumi/Makefile
  - pulumi/provider/cmd/pulumi-resource-flashblade/schema.json
  - pulumi/provider/cmd/pulumi-resource-flashblade/schema-embed.json
  - pulumi/provider/cmd/pulumi-resource-flashblade/bridge-metadata.json
autonomous: true
requirements: [BRIDGE-01]
must_haves:
  truths:
    - "make -C pulumi tfgen produces schema.json, schema-embed.json, bridge-metadata.json"
    - "make -C pulumi provider builds runtime binary with embedded real schema"
    - "Schema files are committed to git"
    - "./pulumi/ directory layout finalized per research/ARCHITECTURE.md"
    - "All 7 TF provider config keys appear in schema.json config.variables (D-01 at schema level)"
  artifacts:
    - path: "pulumi/provider/cmd/pulumi-resource-flashblade/schema.json"
      provides: "Human-readable Pulumi schema (for CI diff)"
    - path: "pulumi/provider/cmd/pulumi-resource-flashblade/schema-embed.json"
      provides: "Embedded runtime schema"
    - path: "pulumi/provider/cmd/pulumi-resource-flashblade/bridge-metadata.json"
      provides: "Embedded bridge metadata"
  key_links:
    - from: "Makefile tfgen target"
      to: "pulumi-tfgen-flashblade binary"
      via: "shell invocation with schema subcommand"
      pattern: "pulumi-tfgen-flashblade schema"
---

<objective>
Fill in the real `tfgen`, `provider`, and `clean` targets in `pulumi/Makefile`, then run `make tfgen` to produce the three committed schema files. Rebuild the runtime binary with the real embedded schema.

Purpose: Completes BRIDGE-01 (directory layout) — the last missing artifacts are the committed schema files. These become the baseline for Phase 57's CI drift gate.
Output: Schema files generated, committed, and the runtime binary rebuilt with real embed payload.
</objective>

<execution_context>
@$HOME/.claude/get-shit-done/workflows/execute-plan.md
</execution_context>

<context>
@.planning/phases/54-bridge-bootstrap-poc-3-resources/54-CONTEXT.md
@.planning/research/ARCHITECTURE.md
@pulumi/Makefile
</context>

<tasks>

<task type="auto">
  <name>Task 1: Replace Makefile stubs with real tfgen/provider/test targets</name>
  <read_first>
    - /home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/pulumi/Makefile (existing skeleton from plan 01)
    - /home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/.planning/research/ARCHITECTURE.md (tfgen schema subcommand usage)
  </read_first>
  <files>
    - pulumi/Makefile
  </files>
  <action>
    Edit `pulumi/Makefile`. Replace the stubbed `tfgen`, `provider`, `test`, and `clean` targets with real implementations while keeping the top-of-file variable block from plan 01 (PROJECT_NAME, PROVIDER_PKG, VERSION, LDFLAGS, PROVIDER_DIR, TFGEN_BIN, RUNTIME_BIN, SCHEMA_DIR):

    ```makefile
    .PHONY: version tfgen tfgen_build provider test clean

    version:
    	@echo $(VERSION)

    # Build tfgen binary (no ldflags version injection — tfgen is a build tool).
    tfgen_build:
    	cd $(PROVIDER_DIR) && go build -o cmd/pulumi-tfgen-flashblade/pulumi-tfgen-flashblade ./cmd/pulumi-tfgen-flashblade

    # Generate schema artifacts. tfgen's `schema` subcommand emits schema.json + bridge-metadata.json
    # + schema-embed.json into the --out directory.
    tfgen: tfgen_build
    	$(TFGEN_BIN) schema --out $(SCHEMA_DIR)
    	@test -s $(SCHEMA_DIR)/schema.json        || (echo "schema.json empty" && exit 1)
    	@test -s $(SCHEMA_DIR)/schema-embed.json  || (echo "schema-embed.json empty" && exit 1)
    	@test -s $(SCHEMA_DIR)/bridge-metadata.json || (echo "bridge-metadata.json empty" && exit 1)

    # Build runtime provider binary with ldflags-injected VERSION and real embedded schema.
    provider: tfgen
    	cd $(PROVIDER_DIR) && go build $(LDFLAGS) -o cmd/pulumi-resource-flashblade/pulumi-resource-flashblade ./cmd/pulumi-resource-flashblade

    # Unit tests for the bridge ProviderInfo (plan 05).
    test:
    	cd $(PROVIDER_DIR) && go test ./... -count=1

    clean:
    	rm -f $(TFGEN_BIN) $(RUNTIME_BIN)
    	rm -f $(SCHEMA_DIR)/schema.json $(SCHEMA_DIR)/schema-embed.json $(SCHEMA_DIR)/bridge-metadata.json
    ```

    Note on tfgen subcommand name: Pulumi PF tfgen supports `schema` as a subcommand that writes `schema.json`, `schema-embed.json`, and `bridge-metadata.json` alongside each other when given `--out`. If v3.127.0 uses a different subcommand spelling (e.g. `--only schema-embed`), adapt the invocation. The key invariant: after `make tfgen`, all three files exist AND are non-empty in `$(SCHEMA_DIR)`.
  </action>
  <verify>
    <automated>grep -q 'tfgen_build' pulumi/Makefile && grep -q 'pulumi-tfgen-flashblade schema' pulumi/Makefile && grep -q '^provider: tfgen' pulumi/Makefile</automated>
  </verify>
  <done>
    - Makefile targets `tfgen_build`, `tfgen`, `provider`, `test`, `clean` are all real (no "not yet implemented" exit-1 stubs)
    - `tfgen` depends on `tfgen_build`
    - `provider` depends on `tfgen` (ensures schema is present before embed)
  </done>
</task>

<task type="auto">
  <name>Task 2: Run make tfgen and commit schema files</name>
  <read_first>
    - /home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/pulumi/Makefile (post-task-1)
  </read_first>
  <files>
    - pulumi/provider/cmd/pulumi-resource-flashblade/schema.json
    - pulumi/provider/cmd/pulumi-resource-flashblade/schema-embed.json
    - pulumi/provider/cmd/pulumi-resource-flashblade/bridge-metadata.json
  </files>
  <action>
    Execute:

    ```bash
    make -C pulumi tfgen
    ```

    This must succeed with a clean exit. All three files must be non-empty JSON.

    Then verify the runtime binary rebuilds with the real schema:

    ```bash
    make -C pulumi provider
    ```

    The runtime binary may emit warnings on stdout (e.g. "no Pulumi.yaml found") when run directly without arguments — that's fine, it only runs in a Pulumi program context. The important signal is that `go build` completes and the binary exists.

    Spot-check schema correctness:
    1. `jq '.name' pulumi/provider/cmd/pulumi-resource-flashblade/schema.json` → `"flashblade"`
    2. `jq '.resources | length' pulumi/provider/cmd/pulumi-resource-flashblade/schema.json` → 28 (count of TF resources)
    3. `jq '.functions | length' pulumi/provider/cmd/pulumi-resource-flashblade/schema.json` → 21 (count of data sources)
    4. `jq '.config.variables.apiToken.secret' pulumi/provider/cmd/pulumi-resource-flashblade/schema.json` → `true`
    5. `jq '.resources["flashblade:bucket:Bucket"]' pulumi/provider/cmd/pulumi-resource-flashblade/schema.json` → non-null (token mapping worked via KnownModules)
    6. D-01 schema-level coverage: all 7 TF provider config keys must appear in `schema.json` under `config.variables` (camelCased by the bridge). Run:

       ```bash
       for k in endpoint apiToken oauth2ClientId oauth2ClientSecret oauth2TokenUrl skipTlsVerify caCertificate; do
         jq -e ".config.variables.$k" pulumi/provider/cmd/pulumi-resource-flashblade/schema.json >/dev/null \
           || { echo "MISSING config.variables.$k (D-01 breach)"; exit 1; }
       done
       ```

       This validates D-01 (1:1 config mirror) at the schema level. Only 3 keys are explicitly named in `ProviderInfo.Config` (the Secret-marked ones); the other 4 (`endpoint`, `oauth2ClientId`, `oauth2TokenUrl`, `skipTlsVerify`) must be auto-exposed via TF schema introspection. If any are missing, the auto-exposure is broken and `ProviderInfo.Config` needs explicit entries for them.

    If counts diverge from 28/21, investigate (likely cause: TF provider added/removed a resource since REQUIREMENTS.md was written). Update counts in `resources_test.go` (plan 05) accordingly OR adjust KnownModules mapping.

    If `apiToken` is not under `config.variables`, the Pulumi key naming may differ (snake→camel). Inspect `jq '.config.variables | keys' schema.json` to confirm.

    The three schema files are NOT in `.gitignore` (see plan 01 — we only ignored binaries). Let `git add` pick them up in the final phase commit.
  </action>
  <verify>
    <automated>make -C pulumi tfgen && test -s pulumi/provider/cmd/pulumi-resource-flashblade/schema.json && test -s pulumi/provider/cmd/pulumi-resource-flashblade/schema-embed.json && test -s pulumi/provider/cmd/pulumi-resource-flashblade/bridge-metadata.json && make -C pulumi provider && test -x pulumi/provider/cmd/pulumi-resource-flashblade/pulumi-resource-flashblade && for k in endpoint apiToken oauth2ClientId oauth2ClientSecret oauth2TokenUrl skipTlsVerify caCertificate; do jq -e ".config.variables.$k" pulumi/provider/cmd/pulumi-resource-flashblade/schema.json >/dev/null || exit 1; done</automated>
  </verify>
  <done>
    - `make tfgen` succeeds cleanly
    - Three schema files are non-empty JSON
    - `make provider` builds the runtime binary
    - Binary contains real schema (not the `{}` placeholder from plan 03)
    - Spot-check: schema.json has `"name": "flashblade"` and non-zero resources count
    - All 7 TF config keys (endpoint, apiToken, oauth2ClientId, oauth2ClientSecret, oauth2TokenUrl, skipTlsVerify, caCertificate) present in `schema.json` `config.variables` (D-01 at schema level)
  </done>
</task>

</tasks>

<verification>
- `make -C pulumi tfgen` exits 0
- `make -C pulumi provider` exits 0
- `jq '.name' schema.json` returns `"flashblade"`
- Schema files are tracked by git (not ignored)
</verification>

<success_criteria>
- BRIDGE-01 satisfied: directory layout is now complete — provider/, sdk/go/, Makefile, schema artifacts all in place
- `make tfgen && make provider` end-to-end flow verified
- Runtime binary embeds real schema payload (not placeholder)
- D-01 verified at schema level: all 7 config keys exposed
</success_criteria>

<output>
After completion, create `.planning/phases/54-bridge-bootstrap-poc-3-resources/54-04-SUMMARY.md`
</output>
