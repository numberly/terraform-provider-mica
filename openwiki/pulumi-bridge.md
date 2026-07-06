# Pulumi bridge

The same Terraform provider is republished as a **Pulumi package** named `mica`,
so users can consume FlashBlade resources from Go and Python (Pulumi programs)
instead of HCL. This is done with the official
[`pulumi-terraform-bridge`](https://github.com/pulumi/pulumi-terraform-bridge)
v3, using the **`pkg/pf/*`** path (for `terraform-plugin-framework`, not the SDK
v2 shim). Everything lives under `pulumi/`.

Design notes and version pins: `pulumi-bridge.md` (repo root) and
`pulumi/README.md`.

## Two binaries (`pulumi/provider/cmd/`)

| Binary | Phase | Role |
|--------|-------|------|
| `pulumi-tfgen-mica` | build-time | Introspects the TF schema → emits `schema.json`, `bridge-metadata.json`, and per-language SDKs. `make tfgen`. |
| `pulumi-resource-mica` | runtime | The Pulumi provider plugin; embeds the generated schema and routes RPCs through the bridge. `make provider`. |

> Older docs (and the root `pulumi-bridge.md`) call these
> `pulumi-tfgen-flashblade` / `pulumi-resource-flashblade`; after the `mica`
> rebrand the actual binaries are `-mica`.

## `ProviderInfo` (`pulumi/provider/resources.go`)

`Provider()` builds `tfbridge.ProviderInfo`:

- `P: pftfbridge.ShimProvider(fb.New(version)())` — wraps the **parent TF
  provider** (`internal/provider`). No CRUD logic is reimplemented; the bridge
  introspects the parent schema and translates Pulumi RPC ↔ TF CRUD.
- Package `mica`, publisher `numberly`, license `GPL-3.0-only`.
- `PluginDownloadURL: github://api.github.com/numberly/terraform-provider-mica`
  (two-segment form so `pulumi plugin install` resolves the binary from GitHub
  Releases).
- **Token mapping**:
  `prov.MustComputeTokens(tokens.SingleModule("flashblade_", "index", …mica…))`
  maps every TF `flashblade_*` type into a single `index` module → tokens like
  `mica:index:Bucket`, `mica:index:FileSystem`. `SingleModule` avoids the
  edge case where a resource name equals a module prefix.
- `Config: {}` empty — sensitive auth fields auto-promote to Pulumi **Secrets**
  from the TF `Sensitive` flag.

**Per-resource overrides** applied after tokenization:

- `omitTimeoutsOnAll()` hides the TF `timeouts` block (Pulumi uses
  `customTimeouts`).
- Explicit `Secret: True()` on credential fields (access keys, array-connection
  keys, certificate passphrase/private key, directory-service bind password,
  remote credentials).
- `ComputeID` closures for composite-ID resources whose API returns no `id`
  (e.g. `s3_export_policy_rule` = `policy/index`,
  `bucket_access_policy_rule` = `bucket/rule`).
- `prov.MustApplyAutoAliases()`.

## SDKs

`tfgen` drives **Python + Go** SDKs (`make generate_python` / `generate_go`);
NodeJS/.NET/Java are explicitly out of scope. SDKs consume the committed schema;
`SKIP_TFGEN=1` / `_from_schema` variants let CI reuse pre-fetched schema
artifacts. The Go SDK is versioned independently (stays on major v0, tag pattern
`sdk/go/v{VERSION}-pulumi[.suffix]`).

## Why the asymmetric naming

TF resource types keep the historical, descriptive `flashblade_*` prefix (HCL:
`flashblade_bucket`). `SingleModule` rewrites the *token namespace* to the
rebranded package, so the same resource surfaces as `mica:index:Bucket` in
Pulumi. The Pulumi package name is a published, trademark-sensitive artifact;
the TF resource type is a code-internal descriptive identifier. See the README's
"Why the asymmetric naming?" section.

## Release

Pulumi artifacts release on `v*-pulumi*` tags via a separate GoReleaser config
and workflow — details in
[workflow-and-release.md](workflow-and-release.md#build--release). A schema-drift
gate in CI fails if `schema.json`/`schema-embed.json`/`bridge-metadata.json`
aren't regenerated after a provider change.
