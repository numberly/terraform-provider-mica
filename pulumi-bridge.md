# Pulumi Bridge for terraform-provider-mica

> Recherche consolidée — avril 2026. Cible: bridger le provider Terraform FlashBlade®
> (terraform-plugin-framework, Go 1.25+, ~20 ressources) vers Pulumi multi-langages.

---

## 1. Architecture cible

Bridge officiel: [`pulumi/pulumi-terraform-bridge`](https://github.com/pulumi/pulumi-terraform-bridge) v3.

Deux sous-packages — **utiliser `pkg/pf/*`** (terraform-plugin-framework), pas le shim SDK v2:

| Phase | Package | Binaire |
|---|---|---|
| Codegen (build-time) | `pkg/pf/tfgen` | `pulumi-tfgen-flashblade` |
| Runtime plugin | `pkg/pf/tfbridge` | `pulumi-resource-flashblade` |

Flow:
1. `tfgen` introspecte le schéma TF → émet `schema.json` (package schema Pulumi) + `bridge-metadata.json` + SDKs par langage.
2. Au runtime, `tfbridge` traduit RPC Pulumi ↔ CRUD TF via le shim `tfshim`.

### Versions (avril 2026)

| Composant | Version |
|---|---|
| pulumi-terraform-bridge/v3 | v3.126.0 |
| pulumi/pkg/v3, pulumi/sdk/v3 | v3.220.0 |
| terraform-plugin-framework | v1.16+ |
| terraform-plugin-go | v0.29.0 |
| Go toolchain | 1.24.7+ (OK avec notre 1.25) |

---

## 2. Layout du repo `pulumi-flashblade`

```
pulumi-flashblade/
├── provider/
│   ├── go.mod                                         # dépend de bridge + tf-provider-flashblade
│   ├── resources.go                                   # ProviderInfo, mappings, overrides
│   ├── resources_test.go                              # tests de mapping (unmapped, secrets, tokens)
│   ├── pkg/version/version.go                         # ldflags -X ...Version
│   └── cmd/
│       ├── pulumi-tfgen-flashblade/main.go            # pf/tfgen.Main
│       └── pulumi-resource-flashblade/
│           ├── main.go                                # pf/tfbridge.Main
│           ├── schema.json                            # généré, committé
│           ├── schema-embed.json                      # généré, //go:embed
│           ├── bridge-metadata.json                   # généré, //go:embed
│           └── Pulumi.yaml
├── sdk/{nodejs,python,go,dotnet,java}/                # générés
├── examples/{bucket-ts,bucket-py,target-go,...}/      # ProgramTest par langage
├── docs/{_index.md,installation-configuration.md}
├── .ci-mgmt.yaml                                      # drive ci-mgmt templates
├── .github/workflows/                                 # générés par ci-mgmt
├── Makefile
└── .goreleaser.yml
```

Source de vérité: [`pulumi/pulumi-tf-provider-boilerplate`](https://github.com/pulumi/pulumi-tf-provider-boilerplate) (`./setup.sh` bootstrap).

---

## 3. Wiring spécifique plugin-framework

### `provider/cmd/pulumi-tfgen-flashblade/main.go`

```go
package main

import (
    "github.com/pulumi/pulumi-terraform-bridge/v3/pkg/pf/tfgen"
    flashblade "github.com/pure-storage/pulumi-flashblade/provider"
)

func main() {
    // Note: pas de paramètre version (différence vs SDK v2)
    tfgen.Main("flashblade", flashblade.Provider())
}
```

### `provider/cmd/pulumi-resource-flashblade/main.go`

```go
package main

import (
    "context"
    _ "embed"

    pftfbridge "github.com/pulumi/pulumi-terraform-bridge/v3/pkg/pf/tfbridge"
    flashblade "github.com/pure-storage/pulumi-flashblade/provider"
)

//go:embed schema-embed.json
var schema []byte

//go:embed bridge-metadata.json
var metadata []byte

func main() {
    meta := pftfbridge.ProviderMetadata{PackageSchema: schema}
    pftfbridge.Main(context.Background(), "flashblade", flashblade.Provider(), meta)
}
```

Différences PF vs SDK v2: context.Context obligatoire dans `tfbridge.Main`, `Version` + `MetadataInfo` obligatoires dans ProviderInfo, `P: pf.ShimProvider(...)` au lieu de `shimv2.NewProvider(...)`.

### `provider/resources.go`

```go
package flashblade

import (
    _ "embed"

    "github.com/pulumi/pulumi-terraform-bridge/v3/pkg/pf"
    "github.com/pulumi/pulumi-terraform-bridge/v3/pkg/tfbridge"
    "github.com/pulumi/pulumi/sdk/v3/go/common/tokens"

    fb "github.com/pure-storage/terraform-provider-mica/internal/provider"
)

//go:embed cmd/pulumi-resource-flashblade/bridge-metadata.json
var bridgeMetadata []byte

const (
    mainPkg = "flashblade"
    mainMod = "index"
)

func Provider() tfbridge.ProviderInfo {
    prov := tfbridge.ProviderInfo{
        P:            pf.ShimProvider(fb.New(version.Version)()),
        Name:         "flashblade",
        Version:      version.Version,
        MetadataInfo: tfbridge.NewProviderMetadata(bridgeMetadata),
        DisplayName:  "FlashBlade",
        Publisher:    "pure-storage",
        LogoURL:      "https://...",
        PluginDownloadURL: "github://api.github.com/pure-storage",
        Description:  "A Pulumi package for creating/managing Pure Storage FlashBlade resources.",
        Keywords:     []string{"pulumi", "flashblade", "purestorage", "category/storage"},
        License:      "Apache-2.0",
        Homepage:     "https://www.pulumi.com",
        Repository:   "https://github.com/pure-storage/pulumi-flashblade",
        GitHubOrg:    "pure-storage",

        Config: map[string]*tfbridge.SchemaInfo{
            "api_token": {Secret: tfbridge.True()},
        },

        Resources: map[string]*tfbridge.ResourceInfo{
            // Auto-mapping couvre 90% — surcharges ciblées ci-dessous

            // Composite ID (policy_name:rule_index)
            "flashblade_object_store_access_policy_rule": {
                Tok: tfbridge.MakeResource(mainPkg, "policy", "AccessPolicyRule"),
                ComputeID: func(ctx context.Context, state resource.PropertyMap) (resource.ID, error) {
                    name := state["policyName"].StringValue()
                    idx := state["ruleIndex"].NumberValue()
                    return resource.ID(fmt.Sprintf("%s:%d", name, int(idx))), nil
                },
            },

            // Sensitive / write-once
            "flashblade_object_store_access_key": {
                Tok: tfbridge.MakeResource(mainPkg, "objectstore", "AccessKey"),
                Fields: map[string]*tfbridge.SchemaInfo{
                    "secret_access_key": {Secret: tfbridge.True()},
                },
            },

            // Strip timeouts block (Pulumi utilise customTimeouts)
            "flashblade_bucket": {
                Tok: tfbridge.MakeResource(mainPkg, "bucket", "Bucket"),
                Fields: map[string]*tfbridge.SchemaInfo{
                    "timeouts": {Omit: true},
                },
            },
        },

        DataSources: map[string]*tfbridge.DataSourceInfo{
            // idem, auto-map + surcharges
        },

        JavaScript: &tfbridge.JavaScriptInfo{...},
        Python:     &tfbridge.PythonInfo{...},
        Golang:     &tfbridge.GolangInfo{...},
        CSharp:     &tfbridge.CSharpInfo{...},
        Java:       &tfbridge.JavaInfo{...},
    }

    // Auto-tokenization — couvre 90% des cas sans surcharge manuelle
    prov.MustComputeTokens(tokens.KnownModules("flashblade_", "index",
        []string{"bucket", "filesystem", "policy", "objectstore", "array", "network"},
        tokens.MakeStandard(mainPkg)))
    prov.MustApplyAutoAliases()
    prov.SetAutonaming(255, "-")

    return prov
}
```

---

## 4. Conventions de mapping

| TF → Pulumi | Règle |
|---|---|
| `snake_case` (TF) → `camelCase` (SDK) | Auto, surchargeable via `Fields[x].Name` |
| `flashblade_bucket` → `flashblade:bucket:Bucket` | Via `MustComputeTokens` + `KnownModules` |
| `id`, `urn`, `provider` | Réservés Pulumi — renommer |
| `name` | `SetAutonaming(255, "-")` ajoute suffixe aléatoire si omis |
| `timeouts {}` block | **`Omit: true`** — Pulumi gère via `customTimeouts` option |
| `Sensitive: true` (TF) | Auto-promu en secret Pulumi ; doubler avec `AdditionalSecretOutputs` pour imbriqués |
| Composite ID (`name:index`) | `ComputeID` callback dans `ResourceInfo` |
| Import par name | Gratuit si ID Pulumi = ID TF ; sinon override `DocInfo.ImportDetails` |

---

## 5. Makefile (aligné sur pulumi-cloudflare)

```make
PROVIDER        := flashblade
CODEGEN         := pulumi-tfgen-$(PROVIDER)
PROVIDER_PATH   := provider
VERSION         := $(shell pulumictl get version)
LDFLAGS         := -X github.com/pure-storage/pulumi-$(PROVIDER)/provider/pkg/version.Version=$(VERSION)

tfgen:           ## Generate schema + bridge-metadata
	go build -o bin/$(CODEGEN) ./provider/cmd/$(CODEGEN)
	./bin/$(CODEGEN) schema --out $(PROVIDER_PATH)/cmd/pulumi-resource-$(PROVIDER)

provider: tfgen
	go build -o bin/pulumi-resource-$(PROVIDER) -ldflags "$(LDFLAGS)" \
	  ./provider/cmd/pulumi-resource-$(PROVIDER)

generate_nodejs: tfgen ; ./bin/$(CODEGEN) nodejs --out sdk/nodejs
generate_python: tfgen ; ./bin/$(CODEGEN) python --out sdk/python
generate_go:     tfgen ; ./bin/$(CODEGEN) go     --out sdk/go
generate_dotnet: tfgen ; ./bin/$(CODEGEN) dotnet --out sdk/dotnet
generate_java:   tfgen ; ./bin/$(CODEGEN) java   --out sdk/java
generate_sdks:   generate_nodejs generate_python generate_go generate_dotnet generate_java

build_sdks:      build_nodejs build_python build_go build_dotnet build_java
install_sdks:    install_nodejs_sdk install_python_sdk install_go_sdk install_dotnet_sdk install_java_sdk

test:    ; cd examples && go test -v -tags=all -parallel 4 -timeout 2h
lint:    ; golangci-lint run ./...
build:   provider build_sdks install_sdks build_registry_docs
```

Env vars utiles: `PULUMI_CONVERT=1` (conversion HCL→SDK langages), `PULUMI_SKIP_MISSING_MAPPING_ERROR=true` (dev only), `COVERAGE_OUTPUT_DIR` (translation report).

---

## 6. CI/CD — piloté par `pulumi/ci-mgmt`

Chaque repo bridgé a un `.ci-mgmt.yaml`; [`pulumi/ci-mgmt`](https://github.com/pulumi/ci-mgmt) génère les workflows GitHub Actions. Workflows canoniques:

| Workflow | Trigger | Rôle |
|---|---|---|
| `prerequisites.yml` | workflow_call | `make tfgen`, upload artifact `schema-embed.json` (partagé — évite régénérations) |
| `build_provider.yml` | workflow_call | `goreleaser build --snapshot` cross-compile (linux/darwin/windows × amd64/arm64) |
| `build_sdk.yml` | workflow_call, matrix | build 1 SDK par langage |
| `pull-request.yml` | PR | prerequisites → build_provider → build_sdk (matrix) → test |
| `run-acceptance-tests.yml` | PR labeled | `pulumi up` smoke tests sur `examples/` |
| `master.yml` | push main | full build + publish SDKs dev |
| `prerelease.yml` / `release.yml` | tag `v*` | `goreleaser release`, publish npm/PyPI/NuGet/Maven, tag `sdk/go/vX.Y.Z` |
| `upgrade-provider.yml` | cron | bump auto du provider TF upstream + PR |
| `upgrade-bridge.yml` | cron | bump auto du bridge + PR |

Snippet clé — cache de schéma:

```yaml
- run: make tfgen
- uses: actions/upload-artifact@v4
  with:
    name: schema-embed.json
    path: provider/cmd/pulumi-resource-flashblade/schema-embed.json
```

---

## 7. Release

- **Versioning**: indépendant du provider TF upstream. Bump `provider/go.mod` sur tag TF ; embed version via ldflags.
- **Tags**: `vX.Y.Z` sur le repo + sous-tag `sdk/go/vX.Y.Z` pour le module Go.
- **goreleaser**: cross-compile 6 plateformes, signature cosign/GPG, assets release. `pulumi plugin install` récupère depuis `pluginDownloadURL` (`github://api.github.com/<org>`).
- **Breaking changes** (bump `SchemaVersion` TF): ajouter `ResourceInfo.Aliases` ou bumper `MajorVersion: 2`. Documenter CHANGELOG.

---

## 8. Docs

- `examples/resources/flashblade_*/resource.tf` (déjà présents) sont auto-convertis par tfgen (`PULUMI_CONVERT=1`) en TS/Python/Go/C#/Java.
- Échecs de conversion: `tfbridge.ResourceInfo.Docs = &tfbridge.DocInfo{Source: "...", ReplaceExamplesSection: true}` ou snippets manuels sous `docs/`.
- `make build_registry_docs` émet le bundle Pulumi Registry.

---

## 9. Tests

| Niveau | Fichier | Rôle |
|---|---|---|
| Unit — ProviderInfo | `provider/resources_test.go` | Vérifier chaque `flashblade_*` est mappé, secrets flaggés, tokens cohérents |
| Schema snapshot | `schema.json` committé | Diff en CI sur PR |
| Integration — `ProgramTest` | `examples/<res>-<lang>/` | `pulumi up/preview/destroy` réel, gated `FLASHBLADE_*` env |
| Smoke par langage | 1 dir par SDK | Minimal stack par langage via `integration.ProgramTest` |

Les tests unit/mock Go existants du provider TF (818 tests) restent inchangés et couvrent la logique métier.

---

## 10. Pièges connus et mitigations

### 10.1 Maturité `pf` bridge

Le sous-package `pkg/pf` est GA mais plus fin que le shim SDK v2.
Limitations connues (tracking: [#744](https://github.com/pulumi/pulumi-terraform-bridge/issues/744),
[#956](https://github.com/pulumi/pulumi-terraform-bridge/issues/956)):

- `PreCheckCallback` custom, certains edge cases `TransformFromState`
- Muxing SDKv2+pf via `muxer` package (pas notre cas)
- Schema version moins testé ([#1667](https://github.com/pulumi/pulumi-terraform-bridge/issues/1667))

**Mitigation**: pin latest `pulumi-terraform-bridge/pf`, référence = [`pulumi-random`](https://github.com/pulumi/pulumi-random). Budgéter du temps pour filer des issues.

### 10.2 Soft-delete / two-phase destroy (buckets, filesystems)

Le bridge appelle `Delete` TF verbatim — notre `destroyed=true` → eradicate → `pollUntilGone` fonctionne tel quel. Pièges:

- Timeout `delete` Pulumi par défaut = **5min**, notre TF default = 30min. Les utilisateurs doivent set `customTimeouts.delete`. Bug bridge: [#1652](https://github.com/pulumi/pulumi-terraform-bridge/issues/1652).
- Si `Delete` panic pendant le poll, Pulumi garde la ressource en state — workaround: `pulumi state delete`.
- `Update` qui force replace déclenche `Delete` ([pulumi-terraform#362](https://github.com/pulumi/pulumi-terraform/issues/362)) — vérifier idempotence destroy sur états partiels.

**Mitigation**:
- Définir `DeleteTimeout: 30*time.Minute` par ressource dans `ResourceInfo`.
- Documenter `destroy_eradicate_on_delete=false` comme default sûr.
- Stripper le bloc `timeouts {}` TF (`Fields["timeouts"].Omit = true`).

### 10.3 Composite IDs

Les ressources framework n'exposent pas d'`id` unique comme SDKv2. Pour `policy_name:rule_index`:

```go
"flashblade_object_store_access_policy_rule": {
    ComputeID: func(ctx, state resource.PropertyMap) (resource.ID, error) {
        return resource.ID(fmt.Sprintf("%s:%d",
            state["policyName"].StringValue(),
            int(state["ruleIndex"].NumberValue()))), nil
    },
},
```

L'import path doit parser le même format — override `ImportStateInput` si nécessaire. Bug: [#2272](https://github.com/pulumi/pulumi-terraform-bridge/issues/2272) "inputs to import do not match" — tester `pulumi import` en CI pour chaque ressource.

### 10.4 State upgraders (SchemaVersion + UpgradeState)

Les upgraders TF tournent transparent **mais**:

- [#1667](https://github.com/pulumi/pulumi-terraform-bridge/issues/1667): `RawState` arrive transformé par le bridge (pré-application schema-aware).
- [#2428](https://github.com/pulumi/pulumi-terraform-bridge/issues/2428): le bridge écrivait version 0 explicitement — fixé, mais auditer le state.
- Pulumi stocke la version sous `__meta`. Un mismatch V1→V2 avec un binaire stale déserialise en schéma V2 (risque corruption silencieuse).

**Mitigation**: `PriorSchema` byte-identique à la version shippée. Tester les upgraders via `pulumi refresh` round-trip avec snapshots de versions antérieures. Ne JAMAIS re-numéroter les versions.

### 10.5 Secrets / write-once fields

`Sensitive: true` (framework) → auto-promu secret Pulumi ([#10](https://github.com/pulumi/pulumi-terraform-bridge/issues/10)).
Gaps ([#1028](https://github.com/pulumi/pulumi-terraform-bridge/issues/1028)): perte de secret-ness dans structs imbriqués.

Pour `SecretAccessKey` (write-once, Read retourne null): utiliser le modèle [Write-Only Fields](https://www.pulumi.com/docs/iac/concepts/secrets/write-only-fields/). Belt-and-braces:

```go
Fields: map[string]*tfbridge.SchemaInfo{
    "secret_access_key": {Secret: tfbridge.True()},
},
// + dans ProviderInfo.ExtraResourceInfo[...].AdditionalSecretOutputs
```

### 10.6 Timeouts

Pulumi **n'expose pas** le bloc TF `timeouts {}` — le stripper. Users utilisent l'option [`customTimeouts`](https://www.pulumi.com/docs/iac/concepts/resources/options/customtimeouts/) au niveau ressource. Propagation buggée: [#1652](https://github.com/pulumi/pulumi-terraform-bridge/issues/1652), [pulumi#12987](https://github.com/pulumi/pulumi/issues/12987).

**Mitigation**: `Fields: {"timeouts": {Omit: true}}` + defaults raisonnables côté serveur.

### 10.7 Drift detection / `tflog.Debug`

Les lignes `tflog.Debug` ne remontent qu'avec `TF_LOG=DEBUG` ou `PULUMI_DEBUG_GRPC`. Les users `pulumi up` normaux ne les verront jamais. Le diff détaillé Pulumi affiche quand même la drift comme changement de plan.

**Mitigation**: documenter `PULUMI_DEBUG_PROVIDERS` / `logging.LogLevel=DEBUG`. Ne pas s'appuyer sur tflog pour l'UX.

### 10.8 Collisions de noms

Mots réservés Pulumi à renommer (`Fields[].Name` ou `Tok`): `id`, `urn`, `provider`.
Ressources nommées `target`, `admin` etc. peuvent collider avec des tokens SDK par langage (ex: Python `from ... import target`). Vérifier la sortie `make tfgen`.

---

## 11. Top priorités — POC

Ordre recommandé pour un proof-of-concept (2-3 semaines):

1. **Bootstrap** (1-2j): fork `pulumi-tf-provider-boilerplate`, `./setup.sh flashblade`, setup `provider/go.mod` sur notre provider TF, wiring PF (section 3).
2. **POC 3 ressources** (3-4j): `target`, `remote_credentials`, `bucket` — couvre:
   - Auto-tokenization (target)
   - Secrets / write-once (remote_credentials.secret_access_key)
   - Soft-delete + timeouts (bucket)
3. **Tests ProviderInfo + ProgramTest TS** (2-3j): 1 example par langage minimum, `pulumi import` testé.
4. **CI via `pulumi/ci-mgmt`** (1-2j): `.ci-mgmt.yaml`, workflows générés, goreleaser config, caching schema.
5. **Coverage complète** (5-7j): mapper les 17 ressources + data sources restantes, overrides composite IDs, docs.
6. **Premier release alpha** (1j): tag `v0.1.0-alpha.1`, publier SDKs, smoke test depuis projet externe.

### 3 actions critiques à ne pas rater

1. `ComputeID` explicite pour chaque ressource à ID composite ; tester symétrie `pulumi import`.
2. `Secret: tfbridge.True()` + `AdditionalSecretOutputs` belt-and-braces sur tous les champs sensibles ; adopter write-only pour `SecretAccessKey`.
3. `Omit: true` sur le bloc `timeouts` ; défauts par ressource dans `ResourceInfo` ; documenter `customTimeouts` pour l'eradication longue.

---

## 12. Références

### Docs Pulumi
- [Bridging a TF provider (official guide)](https://www.pulumi.com/docs/guides/pulumi-packages/how-to-author/)
- [`pulumi-tf-provider-boilerplate`](https://github.com/pulumi/pulumi-tf-provider-boilerplate)
- [`pulumi/pulumi-terraform-bridge`](https://github.com/pulumi/pulumi-terraform-bridge) — `pkg/pf/` et `docs/guides/`
- [`pulumi/ci-mgmt`](https://github.com/pulumi/ci-mgmt) — templates CI partagés
- [Write-Only Fields](https://www.pulumi.com/docs/iac/concepts/secrets/write-only-fields/)
- [`customTimeouts`](https://www.pulumi.com/docs/iac/concepts/resources/options/customtimeouts/)
- [Destroy failures troubleshooting](https://www.pulumi.com/docs/support/troubleshooting/common-issues/destroy-failures/)

### Providers bridgés framework (références)
- [`pulumi-random`](https://github.com/pulumi/pulumi-random) — canonical pf reference
- [`pulumi-cloudflare`](https://github.com/pulumi/pulumi-cloudflare) — Makefile, workflows, goreleaser
- [`pulumi-gitlab`](https://github.com/pulumi/pulumi-gitlab)
- [`pulumi-vault`](https://github.com/pulumi/pulumi-vault)

### Issues bridge clés (à watcher)
- [#744 pf Epic](https://github.com/pulumi/pulumi-terraform-bridge/issues/744)
- [#956 pf support tracker](https://github.com/pulumi/pulumi-terraform-bridge/issues/956)
- [#1028 secret bits lost in nested](https://github.com/pulumi/pulumi-terraform-bridge/issues/1028)
- [#1652 timeout propagation bugs](https://github.com/pulumi/pulumi-terraform-bridge/issues/1652)
- [#1667 RawState distortion](https://github.com/pulumi/pulumi-terraform-bridge/issues/1667)
- [#2272 import input mismatch](https://github.com/pulumi/pulumi-terraform-bridge/issues/2272)
- [#2428 SchemaVersion handling](https://github.com/pulumi/pulumi-terraform-bridge/issues/2428)

### Code-source bridge
- [`pkg/tfbridge/info.go` — ResourceInfo / SchemaInfo](https://github.com/pulumi/pulumi-terraform-bridge/blob/main/pkg/tfbridge/info.go)
- [`pkg/pf/README.md`](https://github.com/pulumi/pulumi-terraform-bridge/blob/main/pkg/pf/README.md)
- [`docs/guides/upgrade-sdk-to-pf.md`](https://github.com/pulumi/pulumi-terraform-bridge/blob/main/docs/guides/upgrade-sdk-to-pf.md)
- [`docs/guides/upgrade-sdk-to-mux.md`](https://github.com/pulumi/pulumi-terraform-bridge/blob/main/docs/guides/upgrade-sdk-to-mux.md)
