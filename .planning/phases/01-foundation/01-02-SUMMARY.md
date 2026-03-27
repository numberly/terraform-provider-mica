---
phase: 01-foundation
plan: 02
subsystem: infra
tags: [go, terraform-provider, terraform-plugin-framework, oauth2, tls, provider-schema, configure]

# Dependency graph
requires:
  - phase: 01-foundation/01-01
    provides: FlashBladeClient with NewClient, NegotiateVersion, Config struct — consumed by Configure
provides:
  - FlashBladeProvider with full HCL config schema (endpoint, auth, TLS, retry)
  - Configure method with env var fallbacks (FLASHBLADE_*), validation, client injection
  - Sensitive field marking for api_token, client_id, key_id
  - NegotiateVersion call at Configure time — fails fast on v2.22 mismatch
  - tflog.Info structured logging for endpoint and auth_mode
  - 7 unit tests covering schema, sensitive fields, env fallback, missing endpoint, missing auth, version mismatch
affects: [03-file-system-resource, all subsequent resource plans]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - Provider Configure reads env vars only when config attribute is null/empty — config always wins
    - auth.SingleNestedAttribute pattern for credential isolation — api_token and oauth2 are siblings inside auth{}
    - Sensitive: true on StringAttribute prevents Terraform from printing credential values in plan output
    - NegotiateVersion called in Configure — fails plan/apply before any resource operation if API version unsupported
    - TDD: RED commit (test) → GREEN commit (feat) per task

key-files:
  created:
    - internal/provider/provider_test.go
  modified:
    - internal/provider/provider.go

key-decisions:
  - "auth block uses SingleNestedAttribute not SingleNestedBlock — framework recommendation for typed config access"
  - "Configure validates endpoint and auth before calling NewClient — cleaner error messages than client-level errors"
  - "max_retries defaults to 3 in Configure if zero/absent — consistent with client defaults"
  - "retry_base_delay parsed as time.Duration string in Configure — decoupled from client's internal Duration type"

patterns-established:
  - "Pattern: env var fallback applied after config.Attr.ValueString() is empty — simple nil-then-env chain"
  - "Pattern: authMode string ('token' vs 'oauth2') derived before tflog call — never log credential values"
  - "Pattern: provider_test.go uses tftypes.Object directly to build config values — avoids heavyweight acceptance test infrastructure for unit tests"

requirements-completed: [PROV-01, PROV-02, PROV-03, PROV-05, PROV-06, PROV-07]

# Metrics
duration: 52min
completed: 2026-03-27
---

# Phase 1, Plan 02: Provider Schema and Configure Summary

**terraform-plugin-framework provider with full HCL schema (endpoint + auth{api_token,oauth2} + TLS + retry), FLASHBLADE_* env var fallbacks, sensitive field marking, NegotiateVersion fail-fast, and 7 unit tests — provider binary compiles and Configure succeeds on terraform init**

## Performance

- **Duration:** ~52 min
- **Started:** 2026-03-27T06:28:46Z
- **Completed:** 2026-03-27T07:20:00Z
- **Tasks:** 1 (TDD: 2 commits — RED test + GREEN implementation)
- **Files modified:** 2

## Accomplishments

- Full provider schema with all config attributes per CONTEXT.md locked decisions — endpoint, auth{api_token, oauth2{client_id,key_id,issuer}}, ca_cert_file, ca_cert, insecure_skip_verify, max_retries, retry_base_delay
- Configure with FLASHBLADE_* env var fallback for all credential and endpoint fields
- Sensitive marking on api_token, client_id, key_id — Terraform suppresses these from plan output
- NegotiateVersion called at Configure time — fails early with clear diagnostic if API v2.22 not supported
- tflog.Info structured logging with endpoint and auth_mode (never logs credentials)
- 7 unit tests: schema structure, sensitive fields, env var fallback, missing endpoint, missing auth, tflog output, version mismatch

## Task Commits

Each task was committed atomically (TDD pattern):

1. **Task 1: TDD RED — failing provider tests** - `dfacf1b` (test)
2. **Task 1: TDD GREEN — provider implementation** - `b31e90a` (feat)

## Files Created/Modified

- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/provider/provider.go` - Full provider implementation: FlashBladeProvider, schema, Configure, Resources/DataSources stubs
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/provider/provider_test.go` - 7 unit tests for schema and Configure behaviors
- `main.go` - Already correct from Plan 01 scaffold; no changes required

## Decisions Made

- Used `schema.SingleNestedAttribute` (not `schema.SingleNestedBlock`) for `auth` and `oauth2` — framework v1.x recommendation for typed struct mapping via `tfsdk` tags.
- Configure validates endpoint and auth presence before constructing `client.Config` — yields cleaner user-facing error messages than raw client errors.
- `max_retries` defaults to 3 in Configure when zero/absent — mirrors client internal default for consistency.
- `retry_base_delay` accepted as Go duration string (e.g. `"1s"`) in provider config, parsed by `time.ParseDuration` — decoupled from client's `time.Duration` field.

## Deviations from Plan

None — plan executed exactly as written. The TDD cycle produced tests that covered all 6 behaviors specified in `<behavior>` plus an additional `TestUnit_Configure_VersionMismatch` test explicitly required by `<action>` step h.

## Issues Encountered

- Initial test file design attempted `tftypes.NewValue` with wrong type signatures for the nested object — fixed by building explicit `tftypes.Object` type variables matching the schema structure exactly. Standard tftypes usage pattern; not a framework bug.
- `main.go` was already fully implemented in Plan 01 (not a stub as the plan's `<action>` implied) — no changes needed.

## Next Phase Readiness

- Provider compiles, all 33 tests pass (client + provider), `go vet` clean, `go build ./...` clean
- Provider Configure fully functional — connects to FlashBlade, negotiates API version, injects client
- Resources() and DataSources() return empty slices — ready for Plan 03/04 resource additions
- No blockers for Phase 1 Plan 03

---
*Phase: 01-foundation*
*Completed: 2026-03-27*
