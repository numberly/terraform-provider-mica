# Phase 57: CI Pipeline - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-22
**Phase:** 57-ci-pipeline
**Areas discussed:** Job structure, Trigger scope, Linting strategy, TF root validation

---

## Job structure

| Option | Description | Selected |
|--------|-------------|----------|
| Multi-jobs avec artifacts | prerequisites → build + SDKs parallèles, aligné CI-01 | ✓ |
| Job monolithique séquentiel | Un seul job, plus simple mais séquentiel | |
| Matrix unique | Factorise YAML mais masque dépendance tfgen | |

**User's choice:** Multi-jobs avec artifacts (prerequisites → build + SDKs parallèles)
**Notes:** Aligné avec CI-01 qui spécifie explicitement cette structure. Parallélisme SDK Python/Go pertinent.

---

## Trigger scope

| Option | Description | Selected |
|--------|-------------|----------|
| ./pulumi/** + internal/** + go.mod root | Tout changement pouvant affecter le schema | ✓ |
| ./pulumi/** uniquement | Strict mais risque breaks silencieux | |
| ./pulumi/** + go.mod root | Moyen terme | |

**User's choice:** `./pulumi/**` + `internal/**` + `go.mod` root (recommandé)
**Notes:** Défense contre les breaks silencieux — `pf.ShimProvider` introspecte le provider TF, donc un changement dans `internal/provider/*` modifie le schema généré.

---

## Linting strategy

| Option | Description | Selected |
|--------|-------------|----------|
| Linter uniquement pulumi/provider/... | Code maintenu manuellement uniquement | ✓ |
| Linter tout avec exclusions | Uniforme mais config plus complexe | |
| Linter tout sans exclusion | Strict mais faux positifs sur code généré | |

**User's choice:** Linter uniquement `pulumi/provider/...` (recommandé)
**Notes:** SDK Go validé par `go build`. La pratique standard pour les projets avec génération de code est de linter uniquement le code maintenu.

---

## TF root validation

| Option | Description | Selected |
|--------|-------------|----------|
| Valider uniquement pulumi/... | CI TF valide déjà le root | ✓ |
| Inclure go build ./... root | Défense en profondeur mais redondant | |
| Inclure go test ./internal/... | Maximum défense mais +5min | |

**User's choice:** Valider uniquement `pulumi/...` (recommandé)
**Notes:** Le `replace` dans le bridge fera échouer le build si le root est cassé. Pas besoin de duplication avec CI TF.

---

## Corrections Made

No corrections — all assumptions confirmed.

---

## Claude's Discretion

- Exact job naming convention within the YAML
- Artifact retention duration
- `workflow_call` for Phase 58 release pipeline reuse
- Specific `goreleaser build --snapshot` flags

---

## Deferred Ideas

- `workflow_call` reuse for Phase 58 release pipeline gate — deferred to Phase 58
- Multi-platform build matrix in CI — release-only, CI does snapshot
- Caching `go build` artifacts between runs — possible optimization, not MVP
