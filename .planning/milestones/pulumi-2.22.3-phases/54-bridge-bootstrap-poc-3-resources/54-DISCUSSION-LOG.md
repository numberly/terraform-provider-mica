# Phase 54: Bridge Bootstrap + POC (3 Resources) - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-21
**Phase:** 54-bridge-bootstrap-poc-3-resources
**Areas discussed:** Provider config surface, MustComputeTokens scope, Version injection, Test tier

---

## Provider configuration surface

| Option | Description | Selected |
|---|---|---|
| Miroir 1:1 du TF (Recommended) | All TF config keys exposed (endpoint, api_token, oauth2_*, skip_tls_verify, ca_certificate). Auto-map via bridge, no overrides except Secret on tokens/secrets. Zero-surprise for TF-migrant users. | ✓ |
| Simplifié (endpoint + api_token) | Only endpoint + api_token in Phase 54. oauth2_* and ca_certificate added in Phase 55 if requested. Smaller surface but initial feature loss. | |

**User's choice:** Miroir 1:1 du TF
**Notes:** Zero-surprise migration path takes priority. All config keys exposed from Phase 54.

---

## MustComputeTokens scope in Phase 54

| Option | Description | Selected |
|---|---|---|
| Tous les 49 resources/DS (Recommended) | MustComputeTokens + KnownModules for all in Phase 54. make tfgen produces full schema. Phase 55 adds overrides (ComputeID, additional AdditionalSecretOutputs, state upgrader tests). Module assignments locked early. | ✓ |
| Uniquement les 3 POC | Minimalist mapping in Phase 54. Schema has only 3 resources. Phase 55 extends to 49. Less noise in first commit but significant schema shift between 54 and 55. | |

**User's choice:** Tous les 49 resources/DS
**Notes:** Locks module assignment early; avoids schema churn.

---

## Version injection mechanism

| Option | Description | Selected |
|---|---|---|
| git describe via Makefile (Recommended) | No extra dependency. VERSION=$(git describe --tags --dirty --always) in Makefile. LDFLAGS -X ...Version=$(VERSION). Already used by existing TF .goreleaser.yml. | ✓ |
| pulumictl get version | Official Pulumi tool, handles Pulumi-specific semver (dev versions, dirty suffix). Requires pulumictl installed on all dev machines + CI. | |

**User's choice:** git describe via Makefile
**Notes:** Uniformity with existing TF release pipeline. No extra tool to install.

---

## Test tier in Phase 54

| Option | Description | Selected |
|---|---|---|
| Unit-only + 1 smoke E2E (Recommended) | resources_test.go (mapping coverage, secrets, DeleteTimeout asserts) + 1 minimal Go ProgramTest that runs pulumi up/destroy on target. De-risks E2E path early. | |
| Unit-only, ProgramTest en Phase 58 | Phase 54 fast, no real FlashBlade needed. ProgramTest deferred with docs/examples in Phase 58. Risk of discovering E2E problem late. | |
| Full ProgramTest (3 resources × 2 languages) | 6 ProgramTests from Phase 54 (target-py, target-go, remote_credentials-py, remote_credentials-go, bucket-py, bucket-go). Max de-risk but heavy scope for a bootstrap phase. | |
| **Other (user-provided):** "Unit-only, je m'occuperais des tests à la main pour l'instant" | User handles manual E2E verification against real FlashBlade outside the automated suite. Phase 54 automated tests are unit-only. | ✓ |

**User's choice:** Unit-only (user handles manual E2E during Phase 54)
**Notes:** No ProgramTest in Phase 54. Full ProgramTest coverage deferred to Phase 58 as planned. User validates manually during bootstrap.

---

## Claude's Discretion

- Exact directory structure under `./pulumi/` (flat `provider/` + `sdk/` + `examples/` is confirmed; internal grouping inside each is Claude's call).
- Makefile target names.
- `omitTimeoutsOnAll` helper naming and placement.
- Version stamping flow detail (ldflags path, VERSION computation location).
- Order of boilerplate adaptation steps during execution.

## Deferred Ideas

- Write-Only Fields pattern migration — deferred to a follow-up milestone once SDK v3.231.0+ readiness verified on Python + Go.
- ProgramTest coverage — deferred to Phase 58 (6 examples: target/remote_credentials/bucket × python/go).
- Remaining composite IDs (bucket_access_policy_rule, network_access_policy_rule, DSR membership) — Phase 55.
- State upgrader `pulumi refresh` smoke tests — Phase 55.
