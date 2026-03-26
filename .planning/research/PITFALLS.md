# Pitfalls Research

**Domain:** Terraform provider for storage appliance (Pure Storage FlashBlade, REST API v2.22)
**Researched:** 2026-03-26
**Confidence:** HIGH (framework pitfalls from official sources + MEDIUM on FlashBlade-specific API quirks)

---

## Critical Pitfalls

### Pitfall 1: FlashBlade Soft-Delete Is Not Terraform Destroy

**What goes wrong:**
The FlashBlade API uses a two-phase deletion model for buckets and file systems. `DELETE` on the endpoint does not immediately eradicate the resource — it sets `destroyed=true` (soft-delete state). The resource remains in the array and is visible to `GET` calls with `?destroyed=true`. Actual eradication requires a second operation (either via `eradication_config` TTL or explicit POST to an eradication endpoint). If the provider's Delete function calls `DELETE` and exits, Terraform considers the resource gone, but the array still holds a tombstoned object that blocks recreation of a same-named resource until the eradication delay expires.

**Why it happens:**
Developers model Terraform's destroy lifecycle as a single API call. The FlashBlade API is designed for data-safety, not for infrastructure-as-code semantics. The soft-delete exists to prevent accidental data loss.

**How to avoid:**
In the provider's `Delete` function:
1. Send `PATCH` with `destroyed=true` to soft-delete.
2. Then send `DELETE` to trigger eradication immediately (or send `PATCH` with `eradication_config` set to an immediate TTL if supported).
3. Poll until the resource no longer appears in `GET ?destroyed=true` before returning — otherwise a fast subsequent `terraform apply` will fail with a name-collision conflict.
Also expose a `force_destroy` attribute in the schema so operators explicitly opt-in to immediate eradication of buckets containing objects.

**Warning signs:**
- Acceptance tests pass on first run, then fail on second run because the array still holds a tombstoned resource with the same name.
- `POST /api/2.22/buckets` returns 400/409 with "resource name already in use" after a destroy.
- `terraform destroy && terraform apply` fails on file systems or buckets.

**Phase to address:**
Phase implementing file system and bucket resources (CRUD core). Must be part of the initial resource scaffolding, not a follow-up patch.

---

### Pitfall 2: "Provider produced inconsistent result after apply" from Computed Attribute Misuse

**What goes wrong:**
The terraform-plugin-framework enforces strict state/plan consistency. If a resource's `Update` function reads the new state from the API and writes it to `resp.State`, but some computed attribute (e.g., `id`, `created`, provisioned space after rounding) differs from what was in `req.Plan`, Terraform raises `Provider produced inconsistent result after apply`. This is one of the most common framework-specific errors and stops the apply completely.

**Why it happens:**
Developers forget that:
- Every attribute returned by the API must either match the plan exactly or be marked `Computed: true` with `UseStateForUnknown()` in the plan modifier.
- FlashBlade returns read-only fields (marked `ro` in the API spec) that are not in the user's config — these must be modelled as `Computed: true`.
- The API may round or normalize values (e.g., `provisioned` storage size in bytes might be rounded to a block boundary).

**How to avoid:**
- All `(ro ...)` fields in the FlashBlade API spec → `Computed: true` in schema, never `Required` or `Optional`.
- Add `UseStateForUnknown()` plan modifier to stable computed fields (IDs, creation timestamps).
- For mutable `Optional+Computed` attributes (e.g., quota values the API may adjust), mark as `Computed: true` and always re-read from the API response after create/update.
- End every `Create` and `Update` with a `Read` call that populates the full state from the API response — never copy from the plan directly.

**Warning signs:**
- `terraform apply` succeeds once, then the second plan shows perpetual diff on a computed field.
- Unit tests pass, but acceptance tests fail with "inconsistent result" on update.
- Fields like `id`, `created`, `space`, `time_remaining` appearing in a diff when unchanged.

**Phase to address:**
Phase 1 (provider scaffold and first resource). Establish a per-resource `flattenXxx`/`expandXxx` pattern from day one that always round-trips through the API response.

---

### Pitfall 3: List-Ordered Attributes Causing False Drift on Policy Rules

**What goes wrong:**
Terraform list attributes are ordered by index. The FlashBlade API for policy rules (NFS export rules, SMB share rules, snapshot policy rules, object store access rules) may return rules in a different order than they were created, or the API may reorder rules internally. If policy rules are modelled as `ListNestedAttribute`, every plan will show a diff even when nothing changed, because the order differs.

**Why it happens:**
The provider developer uses `ListNestedAttribute` because rules feel like "an ordered list". But the API doesn't guarantee order, making the state non-idempotent.

**How to avoid:**
- Use `SetNestedAttribute` for policy rules where the API does not guarantee order.
- Use `ListNestedAttribute` only when the API explicitly preserves insertion order AND the user's configuration order is semantically significant.
- When `SetNestedAttribute` has computed sub-fields (e.g., rule `id`), use custom plan modifiers to reconcile them — the framework's `UseStateForUnknown` alone is insufficient for set elements.
- Test by running `terraform plan` twice after an initial apply: the second plan should show no diff ("no changes").

**Warning signs:**
- Idempotency test (`plan → apply → plan → 0 diff`) fails on any policy resource.
- The API returns rules in a different order than they were POST'd.
- Perpetual diffs on `rules` blocks after clean apply.

**Phase to address:**
Phase implementing policy resources (NFS, SMB, snapshot, object store access policies). Design the attribute type before writing the first policy resource.

---

### Pitfall 4: Incomplete Import Support Breaks Adoption of Existing Infrastructure

**What goes wrong:**
`ImportState` is implemented as a stub (just sets the ID), but `Read` does not populate all required attributes. On `terraform import`, Terraform calls `ImportState` then `Read`. If `Read` skips optional attributes or leaves them null, the state file is incomplete, and subsequent `terraform plan` shows unwanted diffs or errors.

**Why it happens:**
Developers implement `ImportState` late, after the resource is "done", and don't test it as part of normal acceptance testing. The `Read` function silently leaves optional attributes at their zero value if the API doesn't return them.

**How to avoid:**
- Write an acceptance test for `terraform import` for every resource from day one (not as a future task).
- `Read` must populate every schema attribute, including optional ones, from the API response. Null/unknown means the API returned nothing — don't silently skip.
- For the FlashBlade provider, the import ID is the resource name (not a UUID). Validate in `ImportState` that the name format is correct.
- Test that `import → plan` produces zero diff.

**Warning signs:**
- `terraform import` succeeds but `terraform plan` immediately shows a forced replacement.
- Imported resources show `(known after apply)` for attributes that should be known.
- Optional attributes default to zero values instead of API-returned values after import.

**Phase to address:**
Each resource phase must include import acceptance tests. Do not defer to a dedicated "import phase."

---

### Pitfall 5: Sensitive Values Leaked via Error Messages or State

**What goes wrong:**
The provider stores API tokens or OAuth2 client secrets in provider configuration. If these are not marked `Sensitive: true` in the schema, they appear in plan output, apply output, and state files in plaintext. Additionally, if error handling constructs messages that include the HTTP request (which contains `x-auth-token` or `Authorization: Bearer ...` headers), the token is printed to the terminal or CI logs.

**Why it happens:**
- The `api_token` field is `StringAttribute` without `Sensitive: true`.
- HTTP client error wrapping includes the full request.
- State is stored as JSON — any string attribute is stored verbatim.

**How to avoid:**
- Mark `api_token`, `client_secret`, and any credential-related fields as `Sensitive: true` in the schema. This prevents CLI display but does not encrypt state.
- Implement a custom HTTP transport that strips `x-auth-token` and `Authorization` headers from error messages before they are returned to the framework.
- For `api_token` specifically: consider not storing it in state at all by using an ephemeral pattern (read from config, never write to state).
- Document that the remote state backend must use encryption at rest.

**Warning signs:**
- CI logs contain `x-auth-token:` or `Bearer ` strings.
- `terraform show` or `terraform state show` reveals the API token in plaintext.
- Security scanner flags state file for containing credentials.

**Phase to address:**
Phase 1 (provider configuration). Must be addressed before any other work.

---

### Pitfall 6: Not Calling Read at End of Create/Update → Stale State

**What goes wrong:**
The Create or Update function populates `resp.State` from the plan values instead of from the API response. The array may have adjusted the values (e.g., normalised provisioned size, assigned an internal `id`, computed `created` timestamp). The state drifts from reality on first apply and stays wrong until the next refresh.

**Why it happens:**
It's faster to write `resp.State.Set(ctx, plan)` than to add another API GET call. Developers assume the API accepted exactly what was sent.

**How to avoid:**
- Mandatorily end every `Create` and `Update` with a call to the private `read` function that fetches the canonical API state and writes it to `resp.State`.
- Never copy `req.Plan` directly to `resp.State` for any attribute that the API may modify.
- This is a framework-level rule, not a per-resource guideline.

**Warning signs:**
- `provisioned` in state shows a different value than the array admin sees.
- `id` field stays empty or incorrect in state.
- Drift is detected on first `terraform plan` after a clean `terraform apply`.

**Phase to address:**
Phase 1 — establish the Create/Update pattern in the first resource as a template for all others.

---

### Pitfall 7: Missing Retry Logic for Transient API Errors

**What goes wrong:**
The FlashBlade REST API may return 503, 429, or temporary 5xx errors under load, during firmware updates, or during HA failover events. An ops team doing high-frequency CRUD will hit these. Without retry logic, `terraform apply` fails permanently, leaving the user to manually re-run or repair state.

**Why it happens:**
HTTP client is created with default settings and no retry middleware. The provider treats every non-2xx response as a terminal error.

**How to avoid:**
- Wrap all API calls with exponential backoff + jitter for 429 (rate limit), 503 (service unavailable), and 500 (transient) responses.
- Use `context.Context` deadlines (from `terraform-plugin-framework-timeouts`) so retries respect user-configured operation timeouts.
- Implement a single HTTP client factory in the provider so retry logic is centralized, not duplicated per resource.
- Log retries at `WARN` level via `tflog` so operators see them in `TF_LOG=WARN`.

**Warning signs:**
- Acceptance tests fail non-deterministically during array maintenance windows.
- Users report that `terraform apply` fails with "connection reset" or "503" on large plans.
- No retry or backoff code anywhere in the HTTP client.

**Phase to address:**
Phase 1 (provider scaffold and HTTP client). Retries must be in the shared client, not per-resource.

---

### Pitfall 8: Hardcoded API Version in URL Path

**What goes wrong:**
Every endpoint is coded as `/api/2.22/...`. When the FlashBlade firmware is upgraded to a new API version, or when the provider is used against an older Purity release that exposes an earlier version, all requests fail with 404 or version-mismatch errors.

**Why it happens:**
The developer hardcodes the version string at each API call site rather than using a constant or a negotiated version.

**How to avoid:**
- Define a single `const APIVersion = "2.22"` in one file and construct all paths via a helper: `apiPath(p.apiVersion, "/buckets")`.
- During provider initialization, call `GET /api/api_version` to verify the target array supports v2.22. Surface a clear error if not — fail fast rather than failing on the first CRUD call.
- Make `api_version` a provider schema attribute (Optional, default `"2.22"`) for forward compatibility.

**Warning signs:**
- URL construction is scattered across resource files with inlined `"2.22"` strings.
- A firmware upgrade causes all resources to return 404.
- No version negotiation at provider startup.

**Phase to address:**
Phase 1 (provider configuration and HTTP client scaffold).

---

## Technical Debt Patterns

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Copy `req.Plan` → `resp.State` on Update | Faster to write | State diverges from reality; drift immediately detected by operators | Never |
| Use `ListNestedAttribute` for all policy rules | Simpler code | Perpetual false-positive diffs in every plan | Never for unordered API collections |
| Skip `ImportState` for now | Saves time per resource | Existing FlashBlade infra cannot be adopted; ops team must destroy and recreate | Never — import is a stated requirement |
| Inline API version string | Trivial to write | Version upgrade requires grep-and-replace across all resources | Never — use a constant from day one |
| Skip retry logic for V1 | Faster initial build | First production apply during array maintenance fails permanently | Never for ops-facing provider |
| Model all computed fields as `Optional` | Avoids thinking about plan modifiers | "Inconsistent result" errors surface unpredictably in production | Never |
| Store `api_token` in state as plain string | No extra code | Token visible in plaintext state file — security violation | Never — mark `Sensitive: true` at minimum |

---

## Integration Gotchas

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| FlashBlade soft-delete | Call DELETE once, assume resource is gone | PATCH `destroyed=true`, then DELETE; poll for eradication completion before returning |
| FlashBlade pagination | Read first page only; miss resources > 100 items | Always follow `continuation_token` until exhausted on all GET calls |
| OAuth2 token exchange | Treat the access token as long-lived | OAuth2 tokens have a TTL (`access_token_ttl_in_ms`); refresh before expiry; handle 401 by refreshing and retrying once |
| Session-based auth (`x-auth-token`) | Store session token in state | Use session tokens only in-memory within a provider configure call; never persist to state |
| TLS with custom CA | Use default Go TLS config | Build HTTP client with `tls.Config{RootCAs: certPool}` from the provider's `ca_certificate` attribute |
| API `ro` fields in PATCH body | Include read-only fields in PATCH request body | Only send writable fields; sending `ro` fields may cause 422 validation errors |
| FlashBlade `names` query param | Use path params for name-based lookups | The API uses query params (`?names=fs1`) not path params (`/file-systems/fs1`) for most resources |

---

## Performance Traps

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| Full resource list to find one resource | Slow plans when array has hundreds of file systems | Always use `?names=<name>` or `?ids=<id>` filter param; never fetch all and filter in Go | At ~50+ resources in the array |
| Missing pagination on data sources | Data source silently returns partial list | Follow `continuation_token` on every list call in data sources | At ~100 resources (default API page size) |
| No connection pooling | New TLS handshake per resource operation | Reuse a single `*http.Client` per provider instance — the framework creates one provider per workspace | At parallelism > 5 resources |
| Polling without context cancellation | `terraform destroy -target=...` hangs indefinitely | All polling loops must check `ctx.Done()` and respect the context deadline | Any apply with a user interrupt (Ctrl-C) |

---

## Security Mistakes

| Mistake | Risk | Prevention |
|---------|------|------------|
| `api_token` not marked `Sensitive: true` | Token appears in plan output, CI logs, state file | Mark all credential attributes `Sensitive: true` in schema |
| Auth token included in error messages | Token logged to terminal and CI system | Strip `x-auth-token` and `Authorization` headers in error-wrapping layer before passing to framework diagnostics |
| TLS verification disabled by default | Man-in-the-middle against the FlashBlade management interface | `insecure` attribute must default to `false`; emit a warning diagnostic when set to `true` |
| `expose_api_token` param on `GET /admins/api-tokens` | Accidentally fetching and logging tokens during Read | Never call `?expose_api_token=true` in the provider; that is for the UI only |
| Wildcard resource permissions in acceptance tests | Test credentials have over-privileged access | Document minimum required FlashBlade role for each operation; acceptance test setup uses a dedicated low-privilege test user |

---

## UX Pitfalls (Operator Experience)

| Pitfall | Operator Impact | Better Approach |
|---------|-----------------|-----------------|
| Import ID format undocumented | Operator guesses UUID when the ID is a resource name | Document import syntax in every resource's schema `MarkdownDescription`; use `flashblade_bucket.example "bucket-name"` in examples |
| Vague "resource not found" on Read | Operator doesn't know if it's a permissions issue or the resource truly missing | Return `resp.Diagnostics.AddError` with a message that includes the resource name and the HTTP status code |
| No diff detail for drift in audit logging | Compliance team can't trace what changed outside Terraform | Use `tflog.Info` with structured fields (`{resource: "fs-01", old_provisioned: X, new_provisioned: Y}`) in the Read function when drift is detected |
| Perpetual diff on policy rules | Every `terraform plan` shows changes, operator stops trusting plans | Ensure set-based modelling for rules; idempotency test must be in CI |
| `terraform destroy` takes minutes with no output | Operator thinks the apply is hung during eradication polling | Log polling progress via `tflog.Debug` with expected wait time |

---

## "Looks Done But Isn't" Checklist

- [ ] **Soft-delete resources (buckets, file systems):** `Delete` must handle two-phase destroy — verify eradication, not just the PATCH/DELETE call.
- [ ] **Import:** Every resource must have an acceptance test that runs `import → plan → 0 diff`. Import is not done until this passes.
- [ ] **Idempotency:** Every resource must have an acceptance test that runs `apply → plan → 0 diff`. A clean plan after apply is the definition of "correct."
- [ ] **Computed fields:** All `(ro ...)` fields in the API spec are `Computed: true` in the schema — verify by grepping the API spec against the schema.
- [ ] **Sensitive attributes:** `api_token` and `client_secret` have `Sensitive: true` — verify with `terraform show` after a plan that they are redacted.
- [ ] **Pagination:** Any data source or Read that calls a list endpoint follows `continuation_token` — verify by testing against an array with > 100 resources of that type.
- [ ] **Policy rules type:** No `ListNestedAttribute` used where the API returns unordered collections — verify by reordering rules in config and confirming `plan` shows no diff.
- [ ] **API version header:** Provider startup calls `GET /api/api_version` and surfaces an error if v2.22 is not in the list.
- [ ] **OAuth2 token refresh:** If the provider runs a long apply (> `access_token_ttl_in_ms`), it refreshes the token and retries — verify with a synthetic short-TTL in integration tests.

---

## Recovery Strategies

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| Soft-delete not handled, resource recreated with same name fails | MEDIUM | Manually wait for FlashBlade eradication timer to expire (typically 24h default), then re-run `terraform apply` |
| State diverged from API due to missing Read-after-write | MEDIUM | `terraform refresh` to sync state; if state is too corrupted, `terraform state rm` and re-import |
| Token leaked in state file | HIGH | Rotate the API token in FlashBlade admin UI; revoke and reissue; audit CI logs; regenerate state with new token |
| False-positive drift loop on policy rules | LOW | Switch `ListNestedAttribute` to `SetNestedAttribute` in schema; bump provider minor version; users re-apply |
| Import produces incorrect state | MEDIUM | `terraform state rm` the imported resource; fix the `Read` function; re-import |

---

## Pitfall-to-Phase Mapping

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| FlashBlade soft-delete two-phase destroy | Phase: Bucket + File System resources | Acceptance test: `destroy` followed by `apply` with the same name succeeds |
| Inconsistent result from computed attributes | Phase 1: Provider scaffold + first resource | Every resource has a unit test validating schema attribute types (Computed vs Optional) |
| False drift from unordered policy rules | Phase: Policy resources (NFS, SMB, snapshot) | Idempotency acceptance test: `apply → plan → 0 diff` after rule reorder |
| Incomplete import | Every resource phase | Acceptance test: `import → plan → 0 diff` runs in CI |
| Sensitive value leakage | Phase 1: Provider configuration | Security check: `terraform show planfile` confirms token is `(sensitive)` |
| Stale state from missing Read-after-write | Phase 1: Establish resource template | Code review checklist: every Create/Update ends with `read(ctx, ...)` |
| No retry on transient errors | Phase 1: HTTP client scaffold | Integration test: mock 503 response, verify retry and eventual success |
| Hardcoded API version | Phase 1: Provider configuration | `grep -r "2\.22"` finds only the version constant file and the URL builder |
| Missing pagination | Every data source phase | Acceptance test: data source tested against array with > 100 resources |
| OAuth2 token expiry | Phase 1: Auth implementation | Integration test with synthetic short TTL verifies token refresh mid-apply |

---

## Sources

- [Resources - Data Consistency Errors | Terraform | HashiCorp Developer](https://developer.hashicorp.com/terraform/plugin/sdkv2/resources/data-consistency-errors)
- [Plan modification | Terraform | HashiCorp Developer](https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification)
- [Sensitive state best practices | Terraform | HashiCorp Developer](https://developer.hashicorp.com/terraform/plugin/best-practices/sensitive-state)
- [Resource import | Terraform | HashiCorp Developer](https://developer.hashicorp.com/terraform/plugin/framework/resources/import)
- [Timeouts | Terraform | HashiCorp Developer](https://developer.hashicorp.com/terraform/plugin/framework/resources/timeouts)
- [Implement logging | Terraform | HashiCorp Developer](https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-logging)
- [FlashBlade REST API 2.22 — FLASHBLADE_API.md in repo root](./../../FLASHBLADE_API.md) — `destroyed` field on buckets and file systems; `(ro ...)` field annotations; `continuation_token` pagination; `/api/api_version` negotiation endpoint
- [GitHub Issue: `computed` field producing spurious plan changes with framework](https://github.com/hashicorp/terraform-plugin-framework/issues/628) — MEDIUM confidence (community-verified pattern)
- [GitHub Issue: Provider produced unexpected value after apply for a Computed attribute](https://github.com/hashicorp/terraform-plugin-framework/issues/840) — MEDIUM confidence
- [Dealing with unordered sets of complex objects / SetNestedAttribute with Computed attributes](https://discuss.hashicorp.com/t/dealing-with-unordered-sets-of-complex-objects-setnestedattribute-with-computed-attributes/61874) — MEDIUM confidence
- [How to Handle Terraform API Rate Limiting](https://oneuptime.com/blog/post/2026-02-23-how-to-handle-terraform-api-rate-limiting/view) — LOW confidence (single source)

---

*Pitfalls research for: Terraform provider for Pure Storage FlashBlade (REST API v2.22, terraform-plugin-framework)*
*Researched: 2026-03-26*
