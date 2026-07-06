# Phase 61: `flashblade_snmp_manager` Resource & Data Source — Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in `61-CONTEXT.md` — this log preserves the alternatives considered.

**Date:** 2026-05-20
**Phase:** 61-flashblade-snmp-manager
**Areas discussed:** Plan granularity, Schema nesting style, Version switch behavior, Drift detection granularity

---

## Plan Granularity

| Option | Description | Selected |
|---|---|---|
| **1 monolithic plan** | Single `61-01-implement-snmp-manager-PLAN.md` covering all 16 checklist items. Aligns with project's `coarse` granularity default. | ✓ |
| Split in 3 plans | Foundations (models+client+mocks+tests) → Resource+DS → Docs+ROADMAP+verification | |

**User's choice:** 1 monolithic plan (via "Accept all 4 recommendations").
**Notes:** Project `config.json` sets `granularity: coarse`. Recent comparable milestones (v2.22.1, v2.22.2) shipped each resource in a small number of plans, and the volume here fits comfortably in one. Splitting was offered as a fallback for reviewability but rejected.

---

## Schema Nesting Style (for `v2c` and `v3`)

| Option | Description | Selected |
|---|---|---|
| **`schema.SingleNestedAttribute`** | Modern terraform-plugin-framework attribute, `Optional: true, Computed: true`. HCL form: `v3 = { user = "...", auth_protocol = "MD5" }`. Pattern confirmed in `array_connection_resource.go:121-141` (`throttle`). | ✓ |
| `schema.SingleNestedBlock` | Legacy block syntax. HCL form: `v3 { user = "..." }`. Not used anywhere else in this codebase. | |

**User's choice:** `SingleNestedAttribute` (via "Accept all 4 recommendations").
**Notes:** Evidence-based; one-to-one match with the existing `throttle` attribute on `flashblade_array_connection`. No competing pattern.

---

## Version Switch Behavior (v2c ↔ v3)

| Option | Description | Selected |
|---|---|---|
| **Omit unused block** | On Update, send the new block + new `version` and omit the other. Rely on server-side to clear the unused config. No `RequiresReplace`. | ✓ |
| Explicit null on unused block | Send `v2c: null` in PATCH when switching to v3 (and vice-versa). Forces a clean state but adds custom logic. | |
| `RequiresReplace` on `version` | Force resource recreate when `version` changes. Safest but heavy UX (passphrase re-entry, etc.). | |

**User's choice:** Omit unused block (via "Accept all 4 recommendations").
**Notes:** Aligns with `CONVENTIONS.md` directive "let the server validate". If the real API does not clear the old block on its own, drift will surface via the `_DriftDetection` test or live UAT, and we'll either:
- (a) elevate `version` to `RequiresReplace` in a follow-up patch, or
- (b) add explicit-null handling on transition.
The HCL example documents the behaviour and the potential workarounds (taint + apply, or `terraform state rm`).

---

## Drift Detection Granularity

| Option | Description | Selected |
|---|---|---|
| **Per-leaf field** | Log `tflog.Debug` for each leaf: `host`, `notification`, `version`, `v3.user`, `v3.auth_protocol`, `v3.privacy_protocol`. Sensitive write-once fields excluded (API doesn't return them). Total: 6 logs. | ✓ |
| Per-nested-block | One log per top-level field, with the entire nested-block diff folded into `was`/`now`. Less verbose but harder to filter in production logs. | |

**User's choice:** Per-leaf (via "Accept all 4 recommendations").
**Notes:** Matches `CONVENTIONS.md` ("Drift detection on all mutable/computed fields"). Pattern confirmed in `directory_service_management_resource.go` (10 per-leaf drift calls). Sensitive write-once fields (`v2c.community`, `v3.auth_passphrase`, `v3.privacy_passphrase`) are deliberately NOT logged because the API never returns them, so there is no `was`/`now` to compare.

---

## Claude's Discretion

- Exact wording of HCL example comments and drift-log keys.
- Choice of `Seed()` signature (variadic vs slice) — match closest existing handler.
- Whether to include any fields beyond what the swagger defines — **no**, stay strict.

## Deferred Ideas

- `GET /snmp-managers/test` connectivity check (resource-action pattern, future milestone covering all `/{resource}/test`).
- `flashblade_snmp_agent` resource (singleton PATCH-only on `/snmp-agents`; could reuse `SnmpV2c`/`SnmpV3` models — separate milestone).
- Pulumi bridge regen for `flashblade_snmp_manager` (owned by `pulumi-2.23.x`).
- `TEST_BASELINE` bump in `GNUmakefile` (release-only milestones).

## Process

- User explicitly invoked the `flashblade-resource-builder` skill — required by the discussion.
- Pre-check via Serena (`SnmpManager`, `Snmp*`, `snmp_manager`) returned 0 matches → greenfield implementation.
- API schemas (`_snmp_v2c`, `_snmp_v3`, `_snmp_v3_post`) extracted directly from `swagger-2.23.json` because the markdown reference does not expand nested objects.
