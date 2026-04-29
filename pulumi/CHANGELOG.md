# Changelog

## [2.22.6-pulumi.beta] — 2026-04-29

First post-rebrand Pulumi release. Distributed via GitHub Releases (no Pulumi Registry publication).

### Added

- Plugin install command: `pulumi plugin install resource mica 2.22.6-pulumi.beta --server github://api.github.com/numberly`
- Python wheel: `pulumi_mica-2.22.6-py3-none-any.whl` published as a release asset.
- Go SDK companion tag: `sdk/go/v0.1.0-pulumi.beta`.

### Notes

- Builds against the same provider sources as the Terraform `v2.22.6` release (registry-published manifest fix).
- No schema, behavior, or token changes vs `[2.22.4]` rebrand baseline — see that entry for the breaking changes from the `pulumi-flashblade` era.

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

## v2.22.3-pulumi.alpha

**Status:** Alpha — functional but not production-hardened.

### Features

- **Bridge scaffold** — Full `pulumi-terraform-bridge/v3` integration with `pkg/pf/tfbridge` and `pkg/pf/tfgen`.
- **54 resources + 40 data sources** — All Terraform resources and data sources bridged with auto-tokenization via `SingleModule` (token form: `flashblade:index:*`).
- **Python SDK** — Generated `pulumi_flashblade` package, installable as a wheel from GitHub Releases.
- **Go SDK** — Generated Go module at `github.com/numberly/terraform-provider-mica/pulumi/sdk/go`, fetchable via `go get` with `GOPRIVATE`. Versioned independently on major v0 (see `pulumi/sdk/go/VERSION`); the first alpha tag is `sdk/go/v0.1.0-pulumi.alpha`.
- **Cosign-signed binaries** — Multi-platform plugin archives (`linux/darwin/windows` x `amd64/arm64`) signed with keyless Sigstore.
- **Composite ID support** — Import works for all 4 composite-ID resources using `/` separator:
  - `flashblade_object_store_access_policy_rule`
  - `flashblade_bucket_access_policy_rule`
  - `flashblade_network_access_policy_rule`
  - `flashblade_management_access_policy_directory_service_role_membership`
- **Sensitive field promotion** — 7 sensitive fields are marked as Pulumi secrets (auto-promoted from TF schema + explicit overrides).
- **Soft-delete defense** — Bucket and filesystem resources inherit 30-minute delete timeout from the TF provider for two-phase destroy + eradication polling.
- **State upgrader delegation** — TF state upgraders for `flashblade_server`, `flashblade_directory_service_role`, and `flashblade_object_store_remote_credentials` are delegated through the bridge.
- **Schema drift gate** — CI enforces that `schema.json` and `bridge-metadata.json` are committed and unchanged after `make tfgen`.
- **No autonaming** — Resource names are operational identifiers; consumers must supply explicit `name` values.

### Upgrade Notes

These upgrade notes cover moving from "no Pulumi" to alpha:

1. Install the provider plugin:
   ```bash
   pulumi plugin install resource flashblade 2.22.3-pulumi.alpha --server github://api.github.com/numberly
   ```

2. Install the Python SDK:
   ```bash
   pip install https://github.com/numberly/terraform-provider-mica/releases/download/v2.22.3-pulumi.alpha/pulumi_flashblade-2.22.3-pulumi.alpha-py3-none-any.whl
   ```

3. Or fetch the Go SDK (independent v0 versioning — use the SDK tag, not the provider tag):
   ```bash
   export GOPRIVATE="github.com/numberly/*"
   go get github.com/numberly/terraform-provider-mica/pulumi/sdk/go@v0.1.0-pulumi.alpha
   ```

4. Configure the provider in your Pulumi program (see `README.md` for examples).

### Known Limitations

These are the known limitations of this alpha release:

- **Write-once fields are not write-only at the SDK layer** — Secret values are encrypted in state but may still appear in SDK structs. The Write-Only Fields pattern is deferred pending SDK maturity verification.
- **No TypeScript, C#, or Java SDKs** — Scope is Python + Go only.
- **No Pulumi Registry publication** — Install via GitHub Releases only.
- **No PyPI publication** — Install the wheel from the release asset URL.
- **Bridge-level timeout overrides unavailable** — `ResourceInfo` in bridge v3.127.0 does not expose `CreateTimeout`/`UpdateTimeout`/`DeleteTimeout`. Timeouts are inherited from the TF provider shim. Use `customTimeouts` in your Pulumi program if needed.
- **ProgramTest coverage is limited to 3 resources** — `target`, `remote_credentials`, and `bucket` have examples. Full 54-resource coverage is deferred.
- **`pulumi import` round-trip tests are manual** — Composite ID ComputeID closures are unit-tested, but full `pulumi import` + `pulumi refresh` + drift-assertion tests are deferred.
- **State upgrader smoke tests are manual** — TF state upgraders are registered and delegated, but full `pulumi refresh` with pre-captured state snapshots is deferred.
