# Pitfalls Research

**Domain:** Terraform provider for storage appliance (Pure Storage FlashBlade, REST API v2.22)
**Researched:** 2026-03-26 (updated 2026-03-30 — VIP milestone addition)
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

## VIP Milestone Pitfalls (v2.1.1 — Network Interfaces)

These pitfalls are specific to adding network interface (VIP) resource management. VIPs are critical infrastructure — a wrong IP assignment or unexpected deletion causes a service outage.

---

### Pitfall V1: Treating `name` as User-Specified (Auto-Generated Name Confusion)

**What goes wrong:**
The developer treats `name` as a `Required` attribute like other FlashBlade resources (file systems, servers, buckets). But for network interfaces, `name` is read-only (marked `ro` in the API spec) — it is auto-generated by the array (e.g., `vip0`, `vip1`). If `name` is `Required`, Terraform will try to send it in the POST body (or as a `?names=` query param), and the API will either reject it or generate a different name. The state will never match the plan, causing perpetual diffs.

**Why it happens:**
All other resources in this provider use user-specified names as the primary identity and as the `?names=` query parameter for POST. The virtual host resource (`object_store_virtual_host`) is the closest analog — but even there the `hostname` is user-specified while the API-assigned `name` is `Computed`. Developers may copy the server resource scaffold and forget that the name field role is completely different here.

**How to avoid:**
- Schema: `name` must be `Computed: true` with `UseStateForUnknown()`, never `Required` or `Optional`.
- The user-specified identity for VIPs is the `address` (IP address) + `subnet` combination, not a name.
- POST body must NOT include a `?names=` query parameter since there is no user-provided name. Verify the actual POST parameter signature from the API — `address` in the body body may be sufficient for the API to accept the request.
- After POST, read the auto-generated `name` from the response and store it in state. All subsequent PATCH and DELETE calls use this auto-generated name as the `?names=` identifier.
- Follow the `object_store_virtual_host_resource.go` pattern: `name` is `Computed: true` with `UseStateForUnknown()`, and import uses the server-assigned name.
- Import ID must be the auto-generated name (e.g., `vip0`), not an IP address. Document this explicitly in the schema `Description`.

**Warning signs:**
- `terraform plan` shows the `name` field as `(known after apply)` on the first apply but then shows a diff on every subsequent plan.
- POST to `/api/2.22/network-interfaces` returns a different name than what was sent.
- State file contains a `name` that differs from what the array admin sees.

**Phase to address:**
Phase implementing the network interface resource (this milestone). Must be defined before writing any CRUD methods.

---

### Pitfall V2: Sending Immutable Fields in PATCH Body (type, subnet)

**What goes wrong:**
The API spec clearly shows that `type` and `subnet` are writable on POST but absent from `NetworkInterfacePatch`. If the provider's PATCH body struct includes these fields (even with `omitempty`), and the user changes them in their config, one of two things happens:
1. The API returns a 422 or 400 error refusing the change, and `terraform apply` fails with a cryptic API error.
2. Worse: the API silently ignores the field, but Terraform state still reflects the user's desired value — now diverged from the actual array state. Every plan shows a diff.

**Why it happens:**
Developers copy the Go struct from the GET response model and use it for PATCH requests. The GET model includes all fields; the PATCH model is a strict subset. Without explicit separate struct types for POST/GET/PATCH, it's easy to send fields that the PATCH endpoint does not accept.

**How to avoid:**
- Define separate Go structs for POST body, PATCH body, and the GET response: `NetworkInterfacePost`, `NetworkInterfacePatch`, `NetworkInterface` (matching the pattern used in `models_storage.go` for `FileSystemPost`, `FileSystemPatch`, `FileSystem`).
- `NetworkInterfacePatch` must only contain `address`, `attached_servers`, `services` — exactly what the API spec shows for `NetworkInterfacePatch`.
- For `type` and `subnet`: mark as `Computed: true` + `RequiresReplace()` in the schema. If the user changes either, Terraform will destroy and recreate the VIP rather than attempting an in-place update. This is safe behavior and matches the API contract.
- Never send `gateway`, `mtu`, `netmask`, `vlan`, `realms`, `enabled`, `name`, `id` in a PATCH body — these are all `ro` fields.

**Warning signs:**
- PATCH call returns `422 Unprocessable Entity` or `400 Bad Request` with a message about a read-only field.
- State shows `subnet` or `type` with `(known after apply)` after an update that didn't require replace.
- A single Go struct is used for both POST body and PATCH body.

**Phase to address:**
Phase implementing the network interface resource (this milestone). Struct definition must be complete before CRUD methods are written.

---

### Pitfall V3: Owning VIPs vs. Server-Managed VIPs — Responsibility Boundary Confusion

**What goes wrong:**
VIPs can be created by the provider as a `flashblade_network_interface` resource, but they can also be attached to/detached from servers independently. The server's `attached_servers` field in the VIP (and the VIP list visible from a server's perspective) creates a bidirectional relationship. If both the `flashblade_network_interface` resource AND the `flashblade_server` resource attempt to manage this relationship, Terraform will fight itself:
- `flashblade_network_interface` PATCH sets `attached_servers = ["server1"]`
- `flashblade_server` enrichment reads the VIP list and tries to set it back on the server side

This results in perpetual diffs or apply loops because each resource's `Read` reflects the other resource's write.

**Why it happens:**
The developer adds VIP info to the server resource/data source for consumer endpoint discovery (a legitimate goal), but inadvertently makes the server resource manage the VIP relationship rather than just exposing it as read-only.

**How to avoid:**
- **Ownership rule**: `attached_servers` is owned exclusively by `flashblade_network_interface`. The server resource and data source expose VIPs as `Computed`-only, read-only attributes — they never write to the VIP relationship.
- Server resource enrichment: add a `network_interfaces` attribute that is `Computed: true` only, populated from a `GET /api/2.22/network-interfaces?attached_servers=<name>` call during server Read. This attribute is NOT settable by the user in the server resource config.
- Server data source enrichment: same pattern — `Computed: true`, populated on Read.
- Document in both resources' schema descriptions which resource owns the relationship.
- Test: creating a VIP with `attached_servers = ["server1"]` then running `terraform plan` on the server resource should produce zero diff on the `network_interfaces` attribute.

**Warning signs:**
- `terraform plan` shows changes on the server resource after a `flashblade_network_interface` apply, or vice versa.
- Two resources are both marked as "updating" in a single apply for a config change in one of them.
- The server PATCH body ever contains VIP or network interface fields.

**Phase to address:**
Phase implementing the network interface resource and the server schema enrichment (this milestone). The ownership boundary must be documented in the milestone plan before coding begins.

---

### Pitfall V4: IP Address Drift and Service Outage from Incorrect `address` Handling

**What goes wrong:**
VIPs are critical infrastructure. If the `address` field in state diverges from the actual array VIP address (e.g., from a manual change on the array, or from a PATCH that failed mid-apply), the next `terraform apply` will try to PATCH the VIP to the state's `address`. If the IP is already in use by another host on the network, the PATCH may succeed on the array but break network routing — causing a service outage that is not immediately obvious from Terraform's output.

Additionally, if a VIP is deleted and recreated with the same `address` (e.g., due to a destroy-create cycle when `subnet` changes), there is a brief window where the IP is unassigned. Any S3 or NFS clients using that VIP will see connection failures during this window.

**Why it happens:**
- IP addresses are not validated against the subnet range before being sent to the API. The API may accept any IP in the subnet, including ones already assigned.
- Destroy-create cycles are not flagged as potentially disruptive in the provider's plan output.

**How to avoid:**
- Validate `address` against `subnet.prefix` (netmask) during plan time using a custom validator. If the address is outside the subnet's CIDR, surface an error before the apply. This catches typos early.
- Add a custom warning diagnostic (not an error, since it's hard to detect at plan time) when `address` changes: `"Warning: changing the VIP address will briefly interrupt services using this IP."`.
- For `subnet` changes (which require destroy-create via `RequiresReplace`): surface a clear plan message. Consider adding a `lifecycle { prevent_destroy = true }` example in documentation.
- In the API client, treat 422 on PATCH with an address conflict message as a distinct error type (not a generic API error) so the operator sees a useful message.

**Warning signs:**
- The `address` field differs between state and the array admin view.
- `terraform apply` on a VIP address change produces no error but clients start failing.
- No subnet-range validation exists for the `address` field.

**Phase to address:**
Phase implementing the network interface resource (this milestone). Validator must be written before acceptance tests.

---

### Pitfall V5: Server Schema Version Not Bumped After Adding VIP Attributes

**What goes wrong:**
Adding a `network_interfaces` computed attribute to the existing `flashblade_server` resource changes its schema. If the provider is deployed to users who have existing Terraform state for `flashblade_server`, and the schema version is not bumped with a corresponding `UpgradeState` entry, Terraform will fail to load the old state with an error about an unexpected attribute. Users will be unable to plan or apply until they manually edit their state file.

**Why it happens:**
The current `serverResource.UpgradeState` returns an empty map (version 0, no migrations). Adding new computed attributes to a schema is sometimes treated as backwards-compatible, but in terraform-plugin-framework, adding `Computed: true` attributes that were absent from old state can still cause state deserialization failures for users on the old schema version.

**How to avoid:**
- Bump the schema `Version` from `0` to `1` when adding the `network_interfaces` attribute to `flashblade_server`.
- Add a `StateUpgrader` entry in `UpgradeState` that migrates v0 state by setting `network_interfaces` to an empty list (or null).
- The state upgrader is the provider's contract to existing users: their state remains loadable after the upgrade.
- Write a unit test that verifies the v0 → v1 migration produces valid state (following the pattern in `object_store_access_policy_rule_resource_test.go`).
- Same applies to `flashblade_server` data source if it gains new attributes — though data source state is less risky since it is always refreshed.

**Warning signs:**
- `terraform plan` after a provider upgrade fails with `An unexpected error occurred while verifying that the provider understands the current Terraform state.`
- The `serverResource.Schema` `Version` field stays at `0` despite a schema change.
- `UpgradeState` returns an empty map while the schema version was bumped.

**Phase to address:**
Phase implementing server schema enrichment (this milestone, alongside the network interface resource). Must be verified before release.

---

### Pitfall V6: attached_servers Full-Replace Semantics — Accidental Server Detachment

**What goes wrong:**
The FlashBlade API for `attached_servers` uses full-replace semantics: a PATCH with `attached_servers: ["server1"]` replaces the entire list, not appends to it. If the provider reads partial state (e.g., only the servers managed by this Terraform workspace) and sends that partial list in a PATCH, any servers attached outside of Terraform will be silently detached. This can remove VIP access for non-Terraform-managed servers.

This mirrors the same risk on `object_store_virtual_host_resource.go`, which already implements full-replace semantics in `Update` (line 237-244).

**Why it happens:**
- Operators managing mixed environments (some servers Terraform-managed, some not) do not expect Terraform to touch the non-managed attachments.
- The provider never reads "what's currently attached" before sending the PATCH — it just sends what's in the plan.
- An operator adds a server attachment out-of-band, then runs `terraform apply` on an unrelated change. The PATCH resets `attached_servers` to the Terraform-managed list, detaching the out-of-band server.

**How to avoid:**
- The provider must implement full-replace semantics (matching the existing virtual host pattern) — this is the correct API behavior. Do not attempt merge semantics.
- In the schema description for `attached_servers`, explicitly document: "This list is the complete set of servers attached to this VIP. Any server not in this list will be detached on apply. Do not use if VIP attachments are managed outside of Terraform."
- On `Read`, always populate `attached_servers` from the API response (current full list). This ensures drift detection catches out-of-band additions.
- If drift is detected on `attached_servers` during Read, log it via `tflog.Info` (following the virtual host pattern for `hostname` drift at line 191-199 of `object_store_virtual_host_resource.go`).
- Operators who cannot use full Terraform ownership should use the data source (read-only) instead of the resource.

**Warning signs:**
- A server that was manually attached to a VIP disappears from `attached_servers` after a `terraform apply` that changed a different attribute.
- `Read` populates `attached_servers` from plan values rather than the API response.
- `attached_servers` is `Optional` but not `Computed` — then drift is invisible.

**Phase to address:**
Phase implementing the network interface resource (this milestone). Document prominently in the resource schema and example configs.

---

### Pitfall V7: Subnet Reference as NamedReference — Incorrect Serialization

**What goes wrong:**
The `subnet` field on a VIP is a `NamedReference` object (it has both `name` and `id` fields). If the provider models `subnet` as a plain `types.String` (just the name), then:
1. On POST: the API receives `{"subnet": "default"}` instead of `{"subnet": {"name": "default"}}` — likely a 422 error.
2. On GET: the API response `{"subnet": {"name": "default", "id": "..."}}` cannot be deserialized into a string attribute.

Alternatively, if the provider models it as a full nested object with both `name` and `id`, the user must specify an `id` they don't know, causing unnecessary complexity.

**Why it happens:**
Other `NamedReference` fields in the provider (e.g., `attached_servers`) are handled as `[]NamedReference` in Go structs and as `types.List` of strings (names only) in the schema. The `subnet` field is a single reference, not a list, which is a different shape.

**How to avoid:**
- Schema: expose `subnet` as a single `types.String` attribute containing just the subnet name (user-facing). Mark it `Required` on create, `Computed: false` for update (it triggers `RequiresReplace`).
- Client model: in the Go struct, `Subnet` is a `*NamedReference` (same type as used elsewhere in the codebase). The serialization to/from JSON handles the object shape automatically.
- In `mapNetworkInterfaceToModel`: extract `Subnet.Name` from the `NamedReference` and store it in the `types.String` schema attribute.
- In `expandNetworkInterfacePost`: build the `NamedReference{Name: data.Subnet.ValueString()}` when sending to the API.
- Test the round-trip: POST with `subnet.name = "default"` → GET returns `subnet.name = "default"` → no diff.

**Warning signs:**
- POST returns `422 Unprocessable Entity` with a message about `subnet`.
- The `subnet` attribute in state shows `null` after create even though the VIP was created on the array.
- The `subnet` schema type is `types.String` but the Go client struct field is a plain `string`.

**Phase to address:**
Phase implementing the network interface client model (this milestone, before writing the resource).

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
| Reuse GET struct for PATCH body (VIPs) | Single struct to maintain | Sends `ro` fields in PATCH → 422 errors or silent state divergence | Never — use separate POST/PATCH/GET structs |
| Model `name` as `Required` on VIP resource | Consistent with other resources | API ignores or rejects user name; auto-generated name never matches plan | Never for auto-named resources |
| Skip schema version bump when adding VIP fields to server resource | Avoids migration code | Existing state fails to load; breaks all users on provider upgrade | Never — bump version + add upgrader |
| Expose `attached_servers` as append-only (not full-replace) | Avoids detachment risk | Diverges from API semantics; partial list sends wrong state to array | Never — document full-replace, not work around it |

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
| VIP `name` query param on POST | Send `?names=<user-value>` on POST like other resources | VIP names are auto-generated — do NOT send `?names=` on POST; send `address` in body only |
| VIP `subnet` field serialization | Send `subnet` as a plain string | `subnet` is a `NamedReference` object in the API — serialize as `{"subnet": {"name": "..."}}` |
| VIP `attached_servers` on PATCH | Send only added/removed servers | API uses full-replace semantics — always send the complete desired list |
| Server resource enrichment with VIPs | Make server resource write to VIP attachment | Server resource should only read VIP data (via `GET /network-interfaces?attached_servers=name`) — never write |

---

## Performance Traps

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| Full resource list to find one resource | Slow plans when array has hundreds of file systems | Always use `?names=<name>` or `?ids=<id>` filter param; never fetch all and filter in Go | At ~50+ resources in the array |
| Missing pagination on data sources | Data source silently returns partial list | Follow `continuation_token` on every list call in data sources | At ~100 resources (default API page size) |
| No connection pooling | New TLS handshake per resource operation | Reuse a single `*http.Client` per provider instance — the framework creates one provider per workspace | At parallelism > 5 resources |
| Polling without context cancellation | `terraform destroy -target=...` hangs indefinitely | All polling loops must check `ctx.Done()` and respect the context deadline | Any apply with a user interrupt (Ctrl-C) |
| Server Read making extra GET for VIPs | Each `terraform plan` on a server triggers an extra API call | Cache the VIP lookup in the same Read call; use `?attached_servers=<name>` to filter | At scale with many servers; also impacts plan time |

---

## Security Mistakes

| Mistake | Risk | Prevention |
|---------|------|------------|
| `api_token` not marked `Sensitive: true` | Token appears in plan output, CI logs, state file | Mark all credential attributes `Sensitive: true` in schema |
| Auth token included in error messages | Token logged to terminal and CI system | Strip `x-auth-token` and `Authorization` headers in error-wrapping layer before passing to framework diagnostics |
| TLS verification disabled by default | Man-in-the-middle against the FlashBlade management interface | `insecure` attribute must default to `false`; emit a warning diagnostic when set to `true` |
| `expose_api_token` param on `GET /admins/api-tokens` | Accidentally fetching and logging tokens during Read | Never call `?expose_api_token=true` in the provider; that is for the UI only |
| Wildcard resource permissions in acceptance tests | Test credentials have over-privileged access | Document minimum required FlashBlade role for each operation; acceptance test setup uses a dedicated low-privilege test user |
| VIP address change without change control | Wrong IP PATCH causes service outage silently | Surface a warning diagnostic on `address` changes; document blast radius in schema description |

---

## UX Pitfalls (Operator Experience)

| Pitfall | Operator Impact | Better Approach |
|---------|-----------------|-----------------|
| Import ID format undocumented | Operator guesses UUID when the ID is a resource name | Document import syntax in every resource's schema `MarkdownDescription`; use `flashblade_bucket.example "bucket-name"` in examples |
| Vague "resource not found" on Read | Operator doesn't know if it's a permissions issue or the resource truly missing | Return `resp.Diagnostics.AddError` with a message that includes the resource name and the HTTP status code |
| No diff detail for drift in audit logging | Compliance team can't trace what changed outside Terraform | Use `tflog.Info` with structured fields (`{resource: "fs-01", old_provisioned: X, new_provisioned: Y}`) in the Read function when drift is detected |
| Perpetual diff on policy rules | Every `terraform plan` shows changes, operator stops trusting plans | Ensure set-based modelling for rules; idempotency test must be in CI |
| `terraform destroy` takes minutes with no output | Operator thinks the apply is hung during eradication polling | Log polling progress via `tflog.Debug` with expected wait time |
| VIP import ID unclear (auto-generated name not obvious) | Operator tries to import by IP address instead of auto-assigned name like `vip0` | Schema description must state: "Import using the auto-generated interface name (e.g., `vip0`), not the IP address." |
| No warning on VIP address change | Operator changes `address` and is surprised when clients briefly lose connectivity | Add a warning diagnostic in the Update plan modifier: "Changing the VIP address will interrupt network connectivity to this interface." |
| Server data source shows stale VIP list | Operator uses server data source to discover endpoints but VIP list is from previous refresh | Document that VIP list on server reflects state at last `terraform refresh` or `terraform apply` — not real-time |

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
- [ ] **VIP `name` is `Computed`:** The network interface resource schema has `name: Computed: true, RequiresReplace: false` — verify there is no `Required: true` or `Optional: true` on `name`.
- [ ] **VIP POST does not send `?names=`:** The `PostNetworkInterface` client function sends `address` in the body, not as a query parameter — verify by inspecting the constructed URL.
- [ ] **VIP separate PATCH struct:** `NetworkInterfacePatch` Go struct contains only `address`, `attached_servers`, `services` — verify by grep that `type` and `subnet` are absent from the struct.
- [ ] **VIP `type` and `subnet` have `RequiresReplace()`:** Changing either forces a destroy-create, not an in-place update — verify in unit tests.
- [ ] **Server schema version bumped:** `flashblade_server` schema `Version` is `1` (or higher) after adding VIP attributes, and `UpgradeState` has a v0→v1 migrator.
- [ ] **Server VIP attributes are `Computed`-only:** No user can set `network_interfaces` on the server resource — verify the schema has no `Optional: true` on VIP-related fields.
- [ ] **VIP `attached_servers` full-replace documented:** Schema description explicitly warns about full-replace semantics — verify by reading the generated docs.

---

## Recovery Strategies

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| Soft-delete not handled, resource recreated with same name fails | MEDIUM | Manually wait for FlashBlade eradication timer to expire (typically 24h default), then re-run `terraform apply` |
| State diverged from API due to missing Read-after-write | MEDIUM | `terraform refresh` to sync state; if state is too corrupted, `terraform state rm` and re-import |
| Token leaked in state file | HIGH | Rotate the API token in FlashBlade admin UI; revoke and reissue; audit CI logs; regenerate state with new token |
| False-positive drift loop on policy rules | LOW | Switch `ListNestedAttribute` to `SetNestedAttribute` in schema; bump provider minor version; users re-apply |
| Import produces incorrect state | MEDIUM | `terraform state rm` the imported resource; fix the `Read` function; re-import |
| VIP address changed by mistake, clients broken | HIGH | Immediately PATCH the VIP back to the original address via the FlashBlade admin UI or CLI; `terraform refresh` to sync; investigate why the change was applied without change control |
| `attached_servers` PATCH detaches unexpected server | MEDIUM | Immediately PATCH VIP with `?names=<vip>` to add the server back via CLI; add the server to the Terraform config; `terraform apply` to reconcile |
| Server state fails to load after provider upgrade (missing schema migration) | HIGH | Users must run `terraform state rm flashblade_server.<name>` and re-import; provider release must include the migration in next patch |

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
| VIP `name` treated as user-specified | v2.1.1: network interface resource | Unit test: schema verifies `name` has `Computed: true`, no `Required`/`Optional` |
| Immutable fields in PATCH body | v2.1.1: network interface client model | Unit test: `NetworkInterfacePatch` struct has no `type` or `subnet` field |
| Server-VIP ownership conflict | v2.1.1: server enrichment + network interface | Acceptance test: VIP attach via NI resource → plan on server resource shows 0 diff |
| VIP address drift → service outage | v2.1.1: network interface resource | Validator test: `address` outside subnet prefix is rejected at plan time |
| Server schema version not bumped | v2.1.1: server enrichment | Integration test: load v0 state file → provider upgrade → plan succeeds with 0 diff |
| `attached_servers` accidental detachment | v2.1.1: network interface resource | Acceptance test: out-of-band server attachment → plan shows diff (drift detected, not silently preserved) |
| Subnet NamedReference serialization error | v2.1.1: network interface client model | Unit test: POST body JSON includes `{"subnet": {"name": "..."}}` not `{"subnet": "..."}` |

---

## Sources

- [Resources - Data Consistency Errors | Terraform | HashiCorp Developer](https://developer.hashicorp.com/terraform/plugin/sdkv2/resources/data-consistency-errors)
- [Plan modification | Terraform | HashiCorp Developer](https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification)
- [Sensitive state best practices | Terraform | HashiCorp Developer](https://developer.hashicorp.com/terraform/plugin/best-practices/sensitive-state)
- [Resource import | Terraform | HashiCorp Developer](https://developer.hashicorp.com/terraform/plugin/framework/resources/import)
- [Timeouts | Terraform | HashiCorp Developer](https://developer.hashicorp.com/terraform/plugin/framework/resources/timeouts)
- [Implement logging | Terraform | HashiCorp Developer](https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-logging)
- [State Upgrade | Terraform Plugin Framework | HashiCorp Developer](https://developer.hashicorp.com/terraform/plugin/framework/resources/state-upgrade)
- [FlashBlade REST API 2.22 — FLASHBLADE_API.md in repo root](./../../FLASHBLADE_API.md) — `NetworkInterface`, `NetworkInterfacePatch`, `NetworkInterfacePost` schema; `ro` field annotations; `attached_servers` list semantics
- `internal/provider/object_store_virtual_host_resource.go` — reference implementation for auto-named resource pattern (`name: Computed`, import by server-assigned name, `attached_servers` full-replace)
- `internal/provider/server_resource.go` — reference for server schema structure; `Version: 0` in `UpgradeState` must be bumped when adding VIP fields
- `internal/client/models_common.go` — `NamedReference` struct used for `subnet` and `attached_servers` serialization
- [GitHub Issue: `computed` field producing spurious plan changes with framework](https://github.com/hashicorp/terraform-plugin-framework/issues/628) — MEDIUM confidence (community-verified pattern)
- [GitHub Issue: Provider produced unexpected value after apply for a Computed attribute](https://github.com/hashicorp/terraform-plugin-framework/issues/840) — MEDIUM confidence
- [Dealing with unordered sets of complex objects / SetNestedAttribute with Computed attributes](https://discuss.hashicorp.com/t/dealing-with-unordered-sets-of-complex-objects-setnestedattribute-with-computed-attributes/61874) — MEDIUM confidence
- [How to Handle Terraform API Rate Limiting](https://oneuptime.com/blog/post/2026-02-23-how-to-handle-terraform-api-rate-limiting/view) — LOW confidence (single source)

---

*Pitfalls research for: Terraform provider for Pure Storage FlashBlade (REST API v2.22, terraform-plugin-framework)*
*Researched: 2026-03-26*
*Updated: 2026-03-30 — Added VIP milestone pitfalls V1–V7 for network interface resource (v2.1.1)*

---

## Pulumi Bridge Milestone Pitfalls (pulumi-2.22.3 — Pulumi Bridge Alpha)

These pitfalls are specific to bridging THIS provider to Pulumi via `pkg/pf`. They enrich
pulumi-bridge.md Section 10 with provider-specific failure modes at the intersection of bridge
bugs and our concrete code patterns. Do not duplicate generic bridge advice already in Section 10.

---

### Pitfall PB1: Bucket/Filesystem `delete` Timeout Silent Truncation at 5 Minutes

**What goes wrong:**
`flashblade_bucket.Delete` and `flashblade_filesystem.Delete` use a 30-minute timeout from the TF
`timeouts` block (`data.Timeouts.Delete(ctx, 30*time.Minute)`). The Pulumi bridge default delete
timeout is 5 minutes (bridge issue [#1652](https://github.com/pulumi/pulumi-terraform-bridge/issues/1652)).
Without an explicit `DeleteTimeout` in `ResourceInfo`, Pulumi's context is cancelled after 5 minutes.
The `pollUntilGone` loop inside `DestroyAndEradicateBucket` receives a cancelled context, returns an
error, and Pulumi marks the resource as failed — leaving a tombstoned bucket in the array. A
subsequent `pulumi up` will fail with a name-collision conflict (the array returns 409 because the
bucket name is already in use in soft-deleted state).

**Why it happens:**
The TF `timeouts` block is stripped from the Pulumi schema (`Fields["timeouts"].Omit = true`). Without
it, there is no per-operation timeout visible to Pulumi users. The bridge falls back to its own default
of 5 minutes — far below the 30-minute eradication window the API may need on large buckets.

**How to avoid:**
In `resources.go`, for every resource that uses soft-delete (bucket, filesystem) and any other resource
with a delete timeout > 5 minutes, set `DeleteTimeout` explicitly in `ResourceInfo`:

```go
"flashblade_bucket": {
    Tok: tfbridge.MakeResource(mainPkg, "bucket", "Bucket"),
    DeleteTimeout: 30 * time.Minute,
    Fields: map[string]*tfbridge.SchemaInfo{
        "timeouts": {Omit: true},
    },
},
"flashblade_filesystem": {
    Tok: tfbridge.MakeResource(mainPkg, "filesystem", "Filesystem"),
    DeleteTimeout: 30 * time.Minute,
    Fields: map[string]*tfbridge.SchemaInfo{
        "timeouts": {Omit: true},
    },
},
```

Also set `CreateTimeout: 20*time.Minute` and `UpdateTimeout: 20*time.Minute` on all resources to
match the TF defaults. Document `customTimeouts` in the SDK examples for bucket destroy.

**Warning signs:**
- `pulumi destroy` on a bucket fails after ~5 minutes with "context deadline exceeded" or "context
  canceled" inside `pollUntilGone`.
- The array still shows the bucket in `destroyed=true` state after the Pulumi destroy fails.
- No `DeleteTimeout` set in `ResourceInfo` for any resource.

**Phase to address:**
Phase 1 (ProviderInfo scaffold + POC 3 resources). Every resource in `resources.go` must have explicit
`DeleteTimeout`/`CreateTimeout`/`UpdateTimeout`. Add a `resources_test.go` assertion that verifies
`DeleteTimeout >= 25*time.Minute` for soft-delete resources.

---

### Pitfall PB2: Composite ID `ComputeID` Asymmetry Breaks `pulumi import`

**What goes wrong:**
Three resource families use composite IDs:
- `flashblade_object_store_access_policy_rule`: TF ID = `policy_name/rule_name` (slash separator,
  synthetic, set in `readIntoState` as `policyName + "/" + ruleName`).
- `flashblade_bucket_access_policy_rule` and `flashblade_network_access_policy_rule`: same pattern.
- `flashblade_management_access_policy_directory_service_role_membership`: TF ID = `role_name/policy_name`
  (role FIRST — because built-in policy names contain colons like `pure:policy/array_admin`).

The bridge's `ComputeID` callback builds the Pulumi resource ID from output state properties. If
`ComputeID` builds `policy_name + ":" + rule_index` (an integer-based key from pulumi-bridge.md
Section 10.3) but the TF provider's actual `id` attribute contains `policy_name + "/" + rule_name`
(a string-based key from the actual code), the IDs are mismatched. Bridge issue
[#2272](https://github.com/pulumi/pulumi-terraform-bridge/issues/2272) surfaces this as "inputs to
import do not match" — `pulumi import` fails or imports with wrong ID that can never be refreshed.

**Why it happens:**
pulumi-bridge.md Section 10.3 uses `policy_name:rule_index` as an example. The actual provider code
uses `policy_name/rule_name` (string name, slash separator). The membership resource uses
`role_name/policy_name` with role FIRST to avoid ambiguity with colon-containing policy names.
Copying the example verbatim produces a `ComputeID` that diverges from the actual TF ID in state.

**How to avoid:**
In `resources.go`, `ComputeID` must mirror exactly how the TF resource sets `data.ID` in its `Read`
function. Read the actual `readIntoState` / `mapXxxToModel` code before writing `ComputeID`:

```go
// flashblade_object_store_access_policy_rule: ID = "policyName/ruleName"
"flashblade_object_store_access_policy_rule": {
    ComputeID: func(ctx context.Context, state resource.PropertyMap) (resource.ID, error) {
        policy := state["policyName"].StringValue()
        rule   := state["name"].StringValue()
        return resource.ID(policy + "/" + rule), nil
    },
},

// flashblade_management_access_policy_directory_service_role_membership: ID = "role/policy"
"flashblade_management_access_policy_directory_service_role_membership": {
    ComputeID: func(ctx context.Context, state resource.PropertyMap) (resource.ID, error) {
        role   := state["role"].StringValue()
        policy := state["policy"].StringValue()
        return resource.ID(role + "/" + policy), nil
    },
},
```

Verify the separator by reading `readIntoState` for each resource — never infer from the schema
description alone. Add a `ProgramTest` or `resources_test.go` round-trip: create -> `pulumi import`
with the same ID -> `pulumi preview` shows 0 diff.

**Warning signs:**
- `pulumi import flashblade:policy:AccessPolicyRule my-rule "mypolicy:0"` succeeds but subsequent
  `pulumi preview` shows replace.
- `ComputeID` uses an integer `ruleIndex` but the TF `id` attribute is a string containing the rule
  name.
- `pulumi import` error: "inputs to import do not match state".

**Phase to address:**
Phase 1 (POC — target, remote_credentials, bucket). Phase 2 (full coverage of all composite-ID
resources). Each resource with a composite ID must have an explicit `ComputeID` AND a `pulumi import`
round-trip test verifying symmetry.

---

### Pitfall PB3: `secret_access_key` Secret-ness Lost in Bridge State

**What goes wrong:**
`flashblade_object_store_remote_credentials` has `secret_access_key` marked `Sensitive: true` in the
TF schema. The bridge auto-promotes TF sensitive fields to Pulumi secrets. However, bridge issue
[#1028](https://github.com/pulumi/pulumi-terraform-bridge/issues/1028) documents that secret-ness can
be lost when the bridge writes state between operations — specifically, the field may appear as
plaintext in the Pulumi state file after an `update` operation that does not re-emit the secret flag.

Additionally, `secret_access_key` is a write-once field: the FlashBlade API returns it on POST but
never on GET. After the first Read, the value in TF state is whatever was set in config (the provider
reads from state, not from the API, for write-once fields). This means `AdditionalSecretOutputs` is
the only guaranteed defense — the bridge's auto-promotion from `Sensitive: true` may not survive
state round-trips.

**How to avoid:**
Belt-and-braces approach for ALL sensitive write-once fields:

1. `Fields["secret_access_key"].Secret = tfbridge.True()` in `resources.go`.
2. `AdditionalSecretOutputs: []string{"secretAccessKey"}` in the `ResourceInfo` (camelCase Pulumi name).
3. Adopt the Pulumi Write-Only Fields model for `secretAccessKey` per
   [Pulumi Write-Only Fields docs](https://www.pulumi.com/docs/iac/concepts/secrets/write-only-fields/).

Apply the same pattern to:
- `flashblade_object_store_remote_credentials.secret_access_key` and `access_key_id`
- `flashblade_directory_service.bind_password` (LDAP bind credential)
- `flashblade_certificate.private_key` and `private_key_passphrase`
- Any field with `Sensitive: true` in the TF schema — audit ALL 28 resources systematically.

**Warning signs:**
- `pulumi stack export` shows a `secretAccessKey` value in plaintext (not `[secret]`).
- A `pulumi up` that updates an unrelated field on `remote_credentials` drops the secret flag.
- `AdditionalSecretOutputs` not set in any `ResourceInfo`.

**Phase to address:**
Phase 1 (POC — remote_credentials). Write a `resources_test.go` assertion that verifies
`AdditionalSecretOutputs` is set for every resource containing a sensitive field. Run `pulumi stack
export` after a `ProgramTest` and grep for any known-sensitive value in plaintext.

---

### Pitfall PB4: State Upgrader `RawState` Distortion for `server` (v0->v1->v2)

**What goes wrong:**
`flashblade_server` has SchemaVersion 2 with upgraders v0->v1 and v1->v2. When Pulumi reads old TF
state (e.g., from a stack that was originally managed with the TF provider), the bridge processes the
raw state through its own schema-aware transformation BEFORE passing it to `UpgradeState`. Bridge
issue [#1667](https://github.com/pulumi/pulumi-terraform-bridge/issues/1667) documents that this
pre-transformation can corrupt `RawState` when list attributes (like `dns`, `network_interfaces`,
`directory_services` — all `types.List` in the server schema) are present. The upgrader receives a
`req.State` that does not match the `PriorSchema` shape, causing `req.State.Get(ctx, &old)` to fail
with a type mismatch diagnostic.

This is highest risk for `flashblade_server` (v0->v1->v2, lists at every version) and
`flashblade_object_store_remote_credentials` (v0->v1, simpler but still has the timeouts block).
`flashblade_directory_service_role` (v1, fixed in Phase 50.1) is also in scope.

**How to avoid:**
- Write a `pulumi refresh` round-trip test for each resource with a state upgrader. The test must:
  1. Create a fake Pulumi state snapshot at the prior schema version (JSON).
  2. Run `pulumi refresh` with the new provider binary.
  3. Verify the state upgrades without errors and the resource plan shows 0 diff.
- If [#1667](https://github.com/pulumi/pulumi-terraform-bridge/issues/1667) is not yet fixed in the
  pinned bridge version, add explicit `TransformFromState` callbacks in `ResourceInfo` to normalize
  list fields before the bridge passes them to the TF upgrader.
- Check the pinned `pulumi-terraform-bridge` version's changelog for #1667 fix status before
  shipping — if unfixed, flag in release notes.

**Warning signs:**
- `pulumi refresh` on a stack with an older server state fails with "value is not a valid object type".
- `req.State.Get(ctx, &old)` in the v0->v1 upgrader returns diagnostics with type errors.
- The bridge version in `pulumi/go.mod` predates the #1667 fix.

**Phase to address:**
Phase 2 (full resource coverage). For each resource with SchemaVersion > 0, add a state-snapshot
based `pulumi refresh` test. High-risk resources: `flashblade_server` (v2), `flashblade_directory_service_role` (v1).

---

### Pitfall PB5: `destroy_eradicate_on_delete` Bool Omitted from Pulumi UX Path

**What goes wrong:**
`flashblade_bucket` has `destroy_eradicate_on_delete: Optional+Computed, default false`. This is a
Terraform-specific control attribute — it controls provider behavior, not an API field. When bridged,
it appears in the Pulumi schema as a normal boolean input. Pulumi users who omit it get `false`
(safe default). But the risk is user confusion: `pulumi destroy` appears to succeed, but the bucket
is only soft-deleted. Recreation of a same-named bucket via `pulumi up` will fail 409 until the
FlashBlade eradication timer expires — typically hours to days depending on array configuration.

**Why it happens:**
Pulumi does not prompt for confirmation on destroy the way Terraform's `-target` flow does. Users
may not realize the resource lingers in soft-deleted state on the array.

**How to avoid:**
- Document `destroyEradicateOnDelete` prominently in the Python/Go SDK generated docs and the
  hand-written `examples/bucket-py` example.
- In the `ResourceInfo` for bucket, add a `Docs` override with a warning about soft-delete semantics:
  "By default, `pulumi destroy` soft-deletes the bucket. Set `destroyEradicateOnDelete=True` to
  eradicate immediately. Recreation of a soft-deleted bucket with the same name will fail until the
  array's eradication timer expires."
- Add a `ProgramTest` that: creates bucket -> destroys -> immediately tries to create same-named
  bucket -> verifies it fails with a clear 409 error (not a silent hang or timeout).

**Warning signs:**
- No mention of `destroyEradicateOnDelete` in the generated Python/Go SDK docs.
- `pulumi destroy` + `pulumi up` with same bucket name fails with 409 and users file a bug.
- The hand-written `examples/bucket-py` does not show `destroy_eradicate_on_delete=True` as an option.

**Phase to address:**
Phase 1 (POC — bucket). Add the `Docs` override in Phase 1. Add the recreation smoke test in Phase 2.

---

### Pitfall PB6: `pulumi import` Passes Name as ID — Must Not Pass UUID

**What goes wrong:**
All FlashBlade `ImportState` implementations identify resources by NAME, not UUID. `req.ID` is the
name string. The TF UUID (stored as `id` in state) is an array-internal identifier that users never
see. When the bridge surfaces `pulumi import`, the Pulumi ID becomes the TF ID. The risk is that
auto-generated SDK documentation says "import using the resource ID" — which users may interpret as
the UUID from `pulumi stack export`, not the name. They run `pulumi import flashblade:bucket:Bucket
my-bucket <uuid>` and the `ImportState` handler passes the UUID to `GetBucket(ctx, uuid)` which
returns 404 (the API's `?names=` filter requires a name, not a UUID).

**How to avoid:**
Override `DocInfo.ImportDetails` for every resource to explicitly state that the import ID is the
resource name:

```go
"flashblade_target": {
    Tok: tfbridge.MakeResource(mainPkg, "index", "Target"),
    Docs: &tfbridge.DocInfo{
        ImportDetails: "Import using the target name: `$ pulumi import flashblade:index:Target my-target my-target-name`",
    },
},
```

Verify that `pulumi import flashblade:index:Target my-target my-target-name` (passing the name as
both the Pulumi resource name AND the TF import ID) correctly calls `ImportState` with the name.
Test this for the POC resources: `target`, `bucket`, `remote_credentials`.

**Warning signs:**
- `pulumi import` using a UUID fails with a "not found" error inside `ImportState`.
- Generated SDK docs say "Import using the resource ID" without specifying name vs UUID.
- No `ImportDetails` override in any `ResourceInfo`.

**Phase to address:**
Phase 1 (POC — target, remote_credentials, bucket). Phase 2 (full coverage): add `ImportDetails`
to all 28 resources.

---

### Pitfall PB7: `**NamedReference` PATCH Null — Pulumi Null Treated as Omit

**What goes wrong:**
Several resources use `**NamedReference` in their PATCH structs to distinguish "omit" from "set to
null" from "set value". In TF, clearing an optional reference means assigning the outer pointer
non-nil and the inner pointer nil (sends `"ca_cert_group": null` in JSON body). In Pulumi, a user
who sets the attribute to `None` (Python) or `nil` (Go) triggers an update.

The pitfall: if the bridge translates a Pulumi `null` input as "omit the field" rather than "set to
null", the PATCH body omits the field entirely (outer pointer stays nil), and the reference is NOT
cleared on the array — silent no-op. The `pkg/pf` path's handling of `null` vs. absent is less
tested than the SDK v2 path (bridge issue [#744](https://github.com/pulumi/pulumi-terraform-bridge/issues/744)
pf epic, general maturity concern).

**How to avoid:**
- Write a `ProgramTest` for any resource with nullable reference fields that: (1) sets the ref on
  create, (2) sets it to `None`/`null` on update, (3) reads back and verifies the ref is cleared.
- If the test fails (null treated as omit), override the field in `ResourceInfo` with a custom
  `SchemaInfo` that forces the bridge to transmit the null.
- Priority resources: `flashblade_target` (CA cert group ref), `flashblade_bucket_replica_link`
  (remote credentials ref), any resource with `Optional: true` `NamedReference` fields.

**Warning signs:**
- Setting a reference field to `None` in Python SDK causes no change on the array.
- `pulumi preview` shows the field changing to `null` but `pulumi up` leaves the old value.
- No `ProgramTest` covers the null-update path for any nullable reference resource.

**Phase to address:**
Phase 2 (full resource coverage). Identify all resources with `**NamedReference` PATCH fields via
inspection of `internal/client/models_*.go` and add null-update tests for each affected resource.

---

### Pitfall PB8: `timeouts` Block Leaks into Pulumi Schema or Fails `tfgen` Validation

**What goes wrong:**
Every FlashBlade resource has a `timeouts` block from `terraform-plugin-framework-timeouts`. The
bridge's `tfgen` step introspects this as a regular schema attribute. Two failure modes:
1. `tfgen` emits the `timeouts` block as a Pulumi input attribute (type: object with `create`,
   `update`, `delete`, `read` string fields). Users see a confusing `timeouts` input in the Python/Go
   SDK that does nothing at the Pulumi level — Pulumi uses `customTimeouts` option instead.
2. In some bridge versions, the `timeouts` block's schema shape (complex nested object with duration
   strings) fails `tfgen` validation, producing a build error that blocks all SDK generation.

**How to avoid:**
Apply `Fields["timeouts"].Omit = true` in `ResourceInfo` for ALL 28 resources — not just bucket.
This must be done systematically via a helper, not discovered resource by resource:

```go
func omitTimeouts() map[string]*tfbridge.SchemaInfo {
    return map[string]*tfbridge.SchemaInfo{
        "timeouts": {Omit: true},
    }
}
// Then in each ResourceInfo:
"flashblade_target": {Fields: omitTimeouts()},
"flashblade_server": {Fields: omitTimeouts()},
// etc.
```

Verify by running `make tfgen` and asserting:
```bash
jq '[.resources | to_entries[] | select(.value.inputProperties.timeouts != null)] | length' schema.json
# must output: 0
```

Add this assertion to `resources_test.go`.

**Warning signs:**
- `schema.json` contains `"timeouts"` as an input property on any resource.
- `make tfgen` exits with a validation error mentioning duration string parsing.
- Python SDK generates a `timeouts` parameter on any resource constructor.

**Phase to address:**
Phase 1 (ProviderInfo scaffold — before first `make tfgen` run). The `omitTimeouts()` helper must
exist before any `ResourceInfo` is written to prevent the "fix it resource by resource" trap.

---

### Pitfall PB9: Python SDK Name Collision on `target` and `server` Module Tokens

**What goes wrong:**
`flashblade_target` and `flashblade_server` are mapped to module tokens by `MustComputeTokens`. If
the default tokenization assigns them to modules named `target` and `server`, the Python SDK generates:

```python
# sdk/python/pulumi_flashblade/target/target.py  — module name == class name
from pulumi_flashblade import target
target.Target(...)
```

This is a Python anti-pattern where the module and the class share the same name. In Python, after
`from pulumi_flashblade import target`, the name `target` refers to the module — importing the class
requires `target.Target(...)`, which works, but user code that does `import target` collides with a
common local variable name. Additionally, `server.Server` has the same issue.

**How to avoid:**
- Use `KnownModules` with differentiated module names, or place singleton/ambiguous resources into
  the flat `"index"` module:
  - `flashblade_target` -> `flashblade:index:Target` (flat index, no module collision)
  - `flashblade_server` -> `flashblade:index:Server`
- After first `make generate_python`, inspect `sdk/python/pulumi_flashblade/` for directories where
  `__init__.py` exports a class with the same name as the directory.
- Token structure changes are breaking changes for SDK users — lock the structure in Phase 1 before
  any alpha release.

**Warning signs:**
- `sdk/python/pulumi_flashblade/target/` directory contains `target.py`.
- `sdk/python/pulumi_flashblade/server/` directory contains `server.py`.
- `make generate_python` does not warn about this — inspection is manual.

**Phase to address:**
Phase 1 (ProviderInfo scaffold — `MustComputeTokens` configuration). Token structure must be locked
before any SDK generation or alpha release.

---

### Pitfall PB10: Stale `bridge-metadata.json` Embedded After `resources.go` Changes

**What goes wrong:**
`bridge-metadata.json` is generated by `make tfgen`, committed to the repo, and embedded via
`//go:embed` in the runtime binary. It encodes schema mapping metadata (tokens, aliases, computed
IDs). If a developer changes `resources.go` (adds a resource, modifies a `ComputeID`, changes a
token) but does not re-run `make tfgen` before committing, the embedded metadata is stale. The
runtime plugin uses the old metadata to serve resource RPCs:
- New resources are unavailable in Pulumi programs.
- Modified `ComputeID` callbacks are ignored (old ID format used).
- SDK users get an old `schema.json` that does not match the live provider binary.

This is a CI consistency problem — it bites during development when PRs are merged without running
`make tfgen`.

**How to avoid:**
- Add a CI gate in `pull-request.yml` that runs `make tfgen` then:
  ```bash
  git diff --exit-code provider/cmd/pulumi-resource-flashblade/schema.json
  git diff --exit-code provider/cmd/pulumi-resource-flashblade/bridge-metadata.json
  ```
  Fail the PR if either file has uncommitted changes.
- The `prerequisites.yml` workflow generated by `ci-mgmt` already does this — ensure it is not
  bypassed via `[skip ci]` or direct pushes to main.
- Add to the local `Makefile`: `check-schema: make tfgen && git diff --exit-code ...` target.
- Never run `go build ./provider/cmd/pulumi-resource-flashblade/...` without `make tfgen` first.

**Warning signs:**
- `schema.json` in the repo has a different resource count than `resources.go` has entries.
- `pulumi plugin inspect flashblade` shows an old resource list after `make provider`.
- A new resource added to `resources.go` is not visible in the Python SDK after `make build_sdks`.

**Phase to address:**
Phase 1 (CI setup — `prerequisites.yml` and Makefile). Must be enforced before any resource coverage
work, or the stale-embed trap will be hit on every PR in Phase 2.

---

## Pulumi Bridge — Technical Debt Patterns

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| `timeouts` omit only on bucket | Saves 30 seconds | All other 27 resources leak a useless `timeouts` input to SDK users | Never — apply via helper to all resources |
| Copy `policy_name:rule_index` from pulumi-bridge.md Section 10.3 | Fast `ComputeID` draft | Mismatches actual ID format (`policy_name/rule_name`); `pulumi import` fails | Never — read `readIntoState` first |
| Skip `AdditionalSecretOutputs`, rely on `Sensitive: true` auto-promotion only | No extra code | Secret-ness lost on state update (#1028); credentials in plaintext state | Never for write-once secrets |
| Skip `DeleteTimeout`, rely on user setting `customTimeouts` | Simpler `resources.go` | Default 5-minute context kills bucket eradication polling; obscure "context canceled" error | Never for soft-delete resources |
| Build SDKs without running `make tfgen` first | Faster local iteration | Stale schema embedded in binary; divergence caught only at runtime | Never in CI |
| Use same module name as resource class (e.g., `target.Target`) | Auto-tokenization default | Python import confusion; token changes are breaking SDK changes if caught late | Never — use differentiated or flat index tokens |

---

## Pulumi Bridge — Integration Gotchas

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| `pulumi import` for FlashBlade resources | Pass UUID as import ID | FlashBlade `ImportState` uses resource NAME — pass the name as the Pulumi import ID; override `DocInfo.ImportDetails` |
| Soft-delete + `pulumi destroy` | Trust Pulumi 5-minute destroy timeout | Set `DeleteTimeout: 30*time.Minute` in `ResourceInfo`; document `customTimeouts` for extra-large buckets |
| `ComputeID` for policy rules | Copy the `:` separator from generic bridge docs | Actual TF ID uses `/` separator (`policy_name/rule_name`); membership uses `role_name/policy_name` (role FIRST) |
| `bridge-metadata.json` in CI | Build provider binary directly | Always run `make tfgen` before `make provider`; CI must diff schema + metadata files |
| Nullable reference fields (`**NamedReference`) | Trust Pulumi null propagates to TF null PATCH | Verify with a `ProgramTest` that sets ref then sets to null; `pkg/pf` null handling is less tested than SDK v2 |
| Python SDK module naming | Accept `MustComputeTokens` defaults | Inspect `sdk/python/` after first `make generate_python` for `target/target.py`-style collisions |

---

## Pulumi Bridge — "Looks Done But Isn't" Checklist

- [ ] **All 28 resources have `timeouts` omitted:** Verify with `jq '[.resources | to_entries[] | select(.value.inputProperties.timeouts != null)] | length == 0' schema.json`.
- [ ] **All soft-delete resources have `DeleteTimeout: 30*time.Minute`:** `resources_test.go` asserts `DeleteTimeout >= 25m` for `flashblade_bucket` and `flashblade_filesystem`.
- [ ] **All composite-ID resources have `ComputeID`:** `resources_test.go` asserts no resource with a `/`-composite TF ID is missing a `ComputeID` override.
- [ ] **All sensitive fields have `AdditionalSecretOutputs`:** `resources_test.go` scans all resources for `Sensitive: true` TF fields and asserts corresponding camelCase Pulumi name appears in `AdditionalSecretOutputs`.
- [ ] **`bridge-metadata.json` is fresh:** CI `git diff --exit-code` gate passes on every PR.
- [ ] **`pulumi import` works for POC resources:** ProgramTest or manual test verifies create -> import -> preview (0 diff) for `target`, `remote_credentials`, `bucket`.
- [ ] **`pulumi destroy` on bucket completes within DeleteTimeout:** ProgramTest includes a bucket with `destroyEradicateOnDelete=true` and verifies destroy completes without timeout error.
- [ ] **State upgraders survive `pulumi refresh`:** For `flashblade_server` (v2) and `flashblade_directory_service_role` (v1), a fake prior-version state snapshot is tested via `pulumi refresh` without errors.
- [ ] **Python SDK has no `module.Module` naming collision:** `sdk/python/pulumi_flashblade/` inspected for directories where the class name equals the directory name.
- [ ] **Nullable ref fields cleared correctly:** For each resource with `**NamedReference` PATCH fields, a test verifies setting to `null` in Pulumi actually clears the ref on the array (not a no-op).

---

## Pulumi Bridge — Pitfall-to-Phase Mapping

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| PB1: Bucket delete timeout truncation | Phase 1 — ProviderInfo scaffold | `resources_test.go`: `DeleteTimeout >= 25m` for soft-delete resources |
| PB2: Composite ID asymmetry | Phase 1 (POC 3 resources); Phase 2 (full coverage) | `pulumi import` round-trip test per composite-ID resource |
| PB3: `secret_access_key` secret-ness lost | Phase 1 — POC (remote_credentials) | `pulumi stack export` grep for plaintext sensitive values |
| PB4: State upgrader RawState distortion | Phase 2 — full resource coverage | `pulumi refresh` with prior-version state snapshot for server + dsr_role |
| PB5: Soft-delete UX confusion in Pulumi | Phase 1 (bucket POC); Phase 2 (docs) | Hand-written bucket example shows `destroyEradicateOnDelete`; ProgramTest recreates same-name bucket |
| PB6: Import ID = name, not UUID | Phase 1 (POC); Phase 2 (all 28 resources) | `DocInfo.ImportDetails` on every resource; import test for POC 3 + all 28 |
| PB7: Nullable ref null treated as omit | Phase 2 — full coverage | ProgramTest: set ref -> update to null -> verify array cleared |
| PB8: `timeouts` block in Pulumi schema | Phase 1 — before first `make tfgen` | `jq` assertion on `schema.json` + `resources_test.go` |
| PB9: Python `target.Target` name collision | Phase 1 — `MustComputeTokens` config | Inspect `sdk/python/` directory structure after first `make generate_python` |
| PB10: Stale `bridge-metadata.json` | Phase 1 — CI setup | CI `git diff --exit-code` on schema + metadata files on every PR |

---

## Sources (Pulumi Bridge additions)

- [pulumi-terraform-bridge #1652 — timeout propagation bugs](https://github.com/pulumi/pulumi-terraform-bridge/issues/1652) — MEDIUM confidence (issue open, workaround = explicit `DeleteTimeout`)
- [pulumi-terraform-bridge #1028 — secret bits lost in nested structs](https://github.com/pulumi/pulumi-terraform-bridge/issues/1028) — MEDIUM confidence
- [pulumi-terraform-bridge #1667 — RawState distortion in `pkg/pf`](https://github.com/pulumi/pulumi-terraform-bridge/issues/1667) — MEDIUM confidence
- [pulumi-terraform-bridge #2272 — import input mismatch](https://github.com/pulumi/pulumi-terraform-bridge/issues/2272) — MEDIUM confidence
- [pulumi-terraform-bridge #2428 — SchemaVersion handling](https://github.com/pulumi/pulumi-terraform-bridge/issues/2428) — MEDIUM confidence (noted as fixed; verify in pinned version)
- [Pulumi Write-Only Fields](https://www.pulumi.com/docs/iac/concepts/secrets/write-only-fields/) — HIGH confidence (official docs)
- [Pulumi customTimeouts option](https://www.pulumi.com/docs/iac/concepts/resources/options/customtimeouts/) — HIGH confidence (official docs)
- `internal/provider/bucket_resource.go` — actual Delete timeout = 30 min; two-phase via `DestroyAndEradicateBucket` with `pollUntilGone`
- `internal/provider/remote_credentials_resource.go` — `secret_access_key` Sensitive=true; v0->v1 upgrader adds `target_name`
- `internal/provider/object_store_access_policy_rule_resource.go` — ID = `policyName + "/" + ruleName` (slash, not colon)
- `internal/provider/management_access_policy_directory_service_role_membership_resource.go` — ID = `role_name + "/" + policy_name` (role FIRST; policy names contain colons)
- `internal/provider/server_resource.go` — SchemaVersion 2; upgraders v0->v1->v2; list attributes (`dns`, `network_interfaces`, `directory_services`) at every version
- `pulumi-bridge.md` Section 10 — source of known bridge issues; this section enriches with provider-specific incidents and corrects the `:` vs `/` composite ID example

---

*Updated: 2026-04-21 — Added Pulumi Bridge milestone pitfalls PB1-PB10 for pulumi-2.22.3*
