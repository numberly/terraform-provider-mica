# Feature Research

**Domain:** Terraform provider for enterprise storage (Pure Storage FlashBlade REST API v2.22)
**Researched:** 2026-03-30 (VIP milestone update; original 2026-03-26)
**Confidence:** HIGH (API reference verified in FLASHBLADE_API.md + existing provider patterns)

---

## Milestone v2.1.1 — Network Interface (VIP) Feature Scope

This section focuses exclusively on the new network interface (VIP) features.
The sections below it cover the full provider feature landscape from the original research.

### API Capability Summary (v2.22)

```
POST   /api/2.22/network-interfaces
  Writable:  address (string), type (string — only valid value: "vip"),
             subnet (object — reference by name), services (array),
             attached_servers (array — list of server name references)
  Read-only at create: enabled, gateway, mtu, netmask, vlan, realms, name, id

PATCH  /api/2.22/network-interfaces  ?names=[name]
  Writable:  address (string), attached_servers (array), services (array)
  NOT patchable: subnet, type, name

GET    /api/2.22/network-interfaces  ?names=[name]
  Returns all fields (writable + read-only)

DELETE /api/2.22/network-interfaces  ?names=[name]
```

Key API constraints:
- `name` is read-only (assigned by the API, not user-defined at POST time — needs investigation)
- `subnet` is set at create, cannot be changed (RequiresReplace on update)
- `type` is always `"vip"` — no other valid value at v2.22
- `gateway`, `mtu`, `netmask`, `vlan` are derived from the subnet (read-only on VIP)
- `services` controls what protocols the VIP serves (e.g., `data-s3`, `data-nfs`, `management`)
- `attached_servers` follows the same pattern as `object_store_virtual_host.attached_servers`

### Table Stakes (Users Expect These)

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| `flashblade_network_interface` resource (CRUD) | Core deliverable of this milestone — VIPs cannot be managed without it | MEDIUM | POST + PATCH + GET + DELETE. Pattern mirrors `server_resource.go` + `object_store_virtual_host_resource.go`. |
| `flashblade_network_interface` data source | Consumers need to reference existing VIPs (not Terraform-managed) by name for cross-stack composition | LOW | Read-only, single lookup by name. Mirrors `server_data_source.go` pattern exactly. |
| Import support (`terraform import flashblade_network_interface.x <name>`) | Team has existing VIPs; can't adopt into state without import | LOW | ImportState by name — identical to every other resource in this provider. |
| Drift detection on all writable fields | Ops compliance requirement — already present on all 29+ resources; VIP must be consistent | MEDIUM | Read after Create/Update; log diffs via tflog for `address`, `services`, `attached_servers`. |
| Read-only computed fields exposed in state | Consumers need `gateway`, `mtu`, `netmask`, `vlan`, `enabled` to discover network properties | LOW | All read-only fields from GET response must be in schema as `Computed: true`. Essential for data source consumers. |
| `attached_servers` on server resource/data source | Server consumers (e.g., bucket workflows) need to discover which VIPs are reachable via a given server | MEDIUM | Enrichment of existing `flashblade_server` resource and data source. Add `network_interfaces` computed list attribute. |
| `subnet` as a named reference (string) | Subnet is required at creation and is a reference object in the API; must be addressable by name | LOW | Expose as `subnet_name` string (Required, RequiresReplace). Do not attempt to manage the subnet itself. |
| `services` as a list of strings | Consumers need to declare which protocols a VIP serves | LOW | `["data-s3"]`, `["data-nfs"]`, `["management"]` are the expected valid values. Validate against known enum. |
| Timeouts on all operations | Provider-wide convention — all resources have configurable timeouts | LOW | Use `timeouts.Attributes(ctx, timeouts.Opts{Create, Read, Update, Delete})` — copy from server_resource.go. |

### Differentiators (Competitive Advantage)

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Server data source enrichment with VIP list | Consumers can discover all VIPs attached to a server in one data source lookup — enables endpoint discovery patterns in workflows | MEDIUM | Add `network_interfaces` computed list of objects (`name`, `address`, `services`, `enabled`) to `flashblade_server` data source. Requires additional GET `/api/2.22/network-interfaces?names=...` call filtered by server. |
| `services` enum validator | Catches invalid service names at plan time (before apply) | LOW | Validate `services` list elements against `["data-s3", "data-nfs", "management", "replication"]`. Consistent with existing provider validator pattern (`HostnameNoDotValidator`, etc.). |
| Explicit `RequiresReplace` on `subnet` | Makes immutability of subnet visible in plan output — prevents surprise destroys | LOW | `stringplanmodifier.RequiresReplace()` on `subnet_name`. Documents the API constraint clearly. |
| Expose VIP name as computed (API-assigned) | If VIP `name` is truly read-only (API-assigned), exposing it as `Computed` with `UseStateForUnknown` prevents spurious plan diffs | LOW | Needs verification: does POST accept a `name` parameter or is it always derived? FLASHBLADE_API.md shows `name(ro string)` — treat as computed. |

### Anti-Features (Commonly Requested, Often Problematic)

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| Subnet management as part of VIP resource | "One resource to create both subnet and VIP together" | Subnets are lower-level network infrastructure, typically managed by network team separately; mixing creates blast-radius issues | Manage subnets via the FlashBlade admin UI or a separate `flashblade_subnet` resource (future, out of scope for v2.1.1). Reference by name with `subnet_name`. |
| Network interface connectors management | "I see `/api/2.22/network-interfaces/connectors` in the API" | Physical connector settings (lane speed, port count, transceiver type) are hardware configuration, not declarative IaC. Values depend on physical hardware and should not be Terraform-managed. | Out of scope. Read-only data source if needed, but connector settings are changed by hardware upgrades, not code. |
| Auto-detect `services` from subnet | "Let the provider figure out what services are enabled" | Services are user intent, not derived state. Auto-detection hides configuration from the plan and creates hidden dependencies. | User explicitly declares `services`. Provider validates against known enum at plan time. |
| Ping/trace diagnostic resources | "Expose `GET /network-interfaces/ping` and `/trace` as data sources" | Diagnostics are operational actions, not infrastructure state. Results change every run, causing permanent drift in data source results. | Use the FlashBlade admin UI or direct API calls for network diagnostics. |
| Managing `enabled` state of a VIP | "I want to disable a VIP via Terraform" | `enabled` is read-only in the API — it reflects physical link state, not a configurable flag. Setting it would require modifying the subnet, not the VIP. | If VIP needs to be disabled, delete it and recreate, or manage the subnet's enabled state separately. |

---

## Feature Dependencies

```
flashblade_network_interface
    └──references──> subnet (by name, pre-existing — not managed by provider in v2.1.1)
    └──references──> flashblade_server (via attached_servers — optional, by name)

flashblade_server (data source enrichment)
    └──reads──> flashblade_network_interface (to populate network_interfaces computed list)

flashblade_object_store_virtual_host
    └──same attached_servers pattern (already implemented — reference, not dependency)
```

### Dependency Notes

- **`flashblade_network_interface` requires a pre-existing subnet:** The `subnet_name` field references a subnet that must exist on the FlashBlade. The provider does not manage subnets. If the subnet is absent, the POST will fail with an API error (not a plan error). No Terraform-level dependency can be declared without a subnet resource.
- **`attached_servers` is optional:** VIPs can exist without any server attached. Servers can be attached after creation via PATCH. This means `attached_servers` should be Optional+Computed, not Required — mirrors `object_store_virtual_host.attached_servers`.
- **Server data source enrichment is independent of VIP resource:** The `flashblade_server` data source can expose VIP info as a secondary lookup (GET network-interfaces filtered by server name) without any change to the VIP resource itself. These are parallel changes to separate files.
- **`type` is a constant:** The API documents `type: "vip"` as the only valid value. It should be Required+Computed or set as a constant in the schema description. Simplest approach: Required with a validator that only accepts `"vip"`, or Computed with hardcoded value set during Read.

---

## MVP Definition (v2.1.1)

### Launch With

- [ ] `flashblade_network_interface` resource — full CRUD + import + drift detection
  - Schema: `name` (Computed), `address` (Required), `type` (Computed, always "vip"), `subnet_name` (Required, RequiresReplace), `services` (Optional+Computed list), `attached_servers` (Optional+Computed list), computed: `enabled`, `gateway`, `mtu`, `netmask`, `vlan`, `id`
  - Client methods: PostNetworkInterface, GetNetworkInterface, PatchNetworkInterface, DeleteNetworkInterface
- [ ] `flashblade_network_interface` data source — read by name
  - All fields Computed; `name` Required as lookup key
- [ ] `flashblade_server` data source enrichment — add `network_interfaces` computed list
- [ ] `flashblade_server` resource enrichment — add `network_interfaces` computed list (read-only, no write)
- [ ] Unit tests for schema, validators, plan modifiers (new resource + data source)

### Defer (Not in v2.1.1)

- [ ] `flashblade_subnet` resource — full subnet management is a separate, larger feature
- [ ] Network interface connector management — hardware config, not IaC
- [ ] Ping/trace diagnostic data sources — operational tools, not infrastructure state
- [ ] TLS policy attachment to network interfaces (`/network-interfaces/tls-policies`) — security hardening, defer to v2.2

---

## Feature Prioritization Matrix (v2.1.1 scope)

| Feature | User Value | Implementation Cost | Priority |
|---------|------------|---------------------|----------|
| `flashblade_network_interface` resource (CRUD) | HIGH | MEDIUM | P1 |
| `flashblade_network_interface` data source | HIGH | LOW | P1 |
| Import support | HIGH | LOW | P1 |
| Drift detection + tflog | HIGH | LOW | P1 |
| Computed fields (gateway, mtu, netmask, vlan) | MEDIUM | LOW | P1 |
| `flashblade_server` data source enrichment | MEDIUM | MEDIUM | P2 |
| `flashblade_server` resource enrichment | LOW | LOW | P2 |
| `services` enum validator | MEDIUM | LOW | P2 |
| Unit tests | HIGH | LOW | P1 |

**Priority key:**
- P1: Must have for milestone completion
- P2: Should have — adds significant value with low cost
- P3: Nice to have, future consideration

---

## Implementation Notes (VIP-specific)

### Name Assignment

The API schema shows `name(ro string)` on NetworkInterface, which means the provider likely cannot specify the name at POST time (unlike `flashblade_server` where name is Required+RequiresReplace). This needs verification:

- If POST accepts `?names=[name]` query param (like PATCH/DELETE do): `name` is Required+RequiresReplace (same as server)
- If POST does not accept name: `name` is Computed+UseStateForUnknown; the API assigns it, and import must use the API-assigned name

Recommendation: Verify with actual API call. The FLASHBLADE_API.md POST line does not list `name` in the body, but GET/PATCH/DELETE use `?names=` query param — strongly suggests name IS user-supplied at POST via query param, not body. **Check POST endpoint behavior before writing the client method.**

### `subnet` Field Modeling

The API `subnet` field is an object reference (not a flat string). The provider should expose it as `subnet_name` (string) to avoid nested object complexity — consistent with how `attached_servers` is a flat list of name strings rather than objects.

During Read, extract `subnet.name` and map to `subnet_name` in state.

### `attached_servers` Pattern

Follows the exact same pattern as `object_store_virtual_host.attached_servers`:
- Optional+Computed list of strings (server names)
- On Create: pass as named references to POST body
- On Read: map API response array of objects to flat list of name strings
- On Update: include in PATCH body only when changed

Reuse `modelServersToNamedRefs()` helper if applicable — or create equivalent `modelServersToNetworkInterfaceRefs()`.

---

## Original Provider Feature Landscape (v1.0 research — preserved)

### Table Stakes (Users Expect These)

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Full CRUD for every resource | Terraform's core contract — resources without full lifecycle are broken | MEDIUM | Create/Read/Update/Delete + destroy for all 12 resource families in PROJECT.md |
| Accurate drift detection (Read) | Users run `terraform plan` to detect out-of-band changes; stale Read = silent corruption | HIGH | Must capture ALL attributes in Read, including computed/backend-assigned ones. Call Read after Create/Update to sync state. |
| `terraform import` for all resources | Ops team has existing FlashBlade infra; can't manage without import | MEDIUM | Implement `ResourceWithImportState`; use `resource.ImportStatePassthroughID()` where name is the natural key. Policy rules need composite ID (`policy_name/rule_index`). |
| Data sources for every resource | Users need to reference existing infra (not Terraform-managed) in configs | MEDIUM | Parallel to each resource: `flashblade_file_system`, `flashblade_bucket`, etc. List data sources for multi-result lookups. |
| Dual authentication: API token + OAuth2 | API token = dev/local; OAuth2 client_credentials = production CI/CD | MEDIUM | API token via session login (`POST /api/login` → `x-auth-token`). OAuth2 via `POST /oauth2/1.0/token` with `token-exchange` grant type. |
| TLS support with custom CA certificates | Enterprise environments use internal CAs; hard failure without this | LOW | Configurable `ca_cert_file` or `ca_cert` in provider schema. Standard Go `tls.Config`. |
| Sensitive attribute flags on secrets | API tokens, passwords, access keys must not appear in plan/apply output | LOW | Mark `api_token`, `secret_access_key`, `password` fields as `Sensitive: true` in schema. |
| `sensitive` flag on object store access keys | Access keys stored in state must be redacted from CLI output | LOW | `flashblade_object_store_access_key` — secret_access_key is sensitive |
| Provider-level environment variable configuration | CI/CD pipelines pass credentials via env vars; hardcoding is anti-pattern | LOW | `FLASHBLADE_ENDPOINT`, `FLASHBLADE_API_TOKEN`, `FLASHBLADE_OAUTH2_*` env vars as fallback to config block |
| Correct plan modifiers on computed attributes | Without `UseStateForUnknown` on stable computed attrs, every plan shows "(known after apply)" noise | MEDIUM | All `id`, `eui`, `wwn`, `created`, `space` fields need `UseStateForUnknown`. Immutable fields (account on bucket, nfs_v3 on create) need `RequiresReplace`. |
| Resource validators | Invalid input (negative quota, invalid policy rule syntax) must fail at plan time, not at apply | MEDIUM | Custom `validator.String`/`validator.Int64` for quota values, policy effect enum, retention period ranges |
| Structured logging via `tflog` | Required for debugging provider issues in production; standard across ecosystem | LOW | Use `tflog.Info/Debug/Error` with structured fields (resource name, operation, API path). Critical for the audit logging requirement. |
| Unit tests for schema and validators | Schema regressions and validator bugs are common; catch before acceptance tests | MEDIUM | Test plan modifiers, validators, schema defaults in isolation — no API required |
| Acceptance tests against real FlashBlade | The only way to verify full CRUD behavior including API responses and state convergence | HIGH | Use `resource.Test` + `resource.TestStep` sequences (Create → Read → Update → Read → Destroy) |
| API versioning header | FlashBlade requires `/api/2.22/` prefix; future-proof version handling | LOW | Pin to `2.22` in HTTP client. Version negotiation via `GET /api/api_version` on startup. |

### Differentiators (Competitive Advantage)

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Drift detection with structured audit log output | Ops compliance requirement: when Read detects a diff, log exactly which attributes changed from what to what at INFO level via tflog | MEDIUM | Log field-by-field diffs in Read when prior state differs from API response. |
| Mocked API integration tests for CI (no FlashBlade required) | Enables fast feedback in CI pipelines where real FlashBlade access is unavailable or expensive | HIGH | HTTP mock server (e.g., `net/http/httptest`) implementing FlashBlade API responses. |
| Full policy family coverage in v1 | Competitors ship partial policy support, forcing click-ops fallback | HIGH | All 6 policy types: NFS export, SMB share, snapshot, object store access, network access, quota. |
| Composite import IDs for policy rules | Policy rules have no standalone ID — naive import breaks | MEDIUM | Convention: `policy_name:rule_index` or `policy_name:rule_name`. |
| Quota policy resource with hard/soft limits | Storage quota enforcement is a top ops requirement | MEDIUM | `flashblade_quota_policy` and `flashblade_quota_policy_rule` with `quota_limit`, `hard_limit_enabled` |
| Object Lock configuration on buckets | Compliance/WORM requirements | HIGH | Map `retention_lock` enum, `object_lock_config.default_retention_mode` |
| Eradication config management | Controls how quickly destroyed resources are permanently deleted | MEDIUM | `eradication_config.eradication_delay` on filesystem/bucket resources. |
| Destroyed state lifecycle (soft delete) | FlashBlade buckets/filesystems support `destroyed=true` before permanent deletion | HIGH | Two-phase delete: PATCH `destroyed=true`, then DELETE. |

### Anti-Features (Commonly Requested, Often Problematic)

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| Performance metrics resources | "We want dashboards in Terraform" | Time-series data, not configuration state. Every plan shows spurious diffs. | Datadog FlashBlade integration or Prometheus purestorage-exporter. |
| Snapshot management as resources | "We want on-demand snapshots via Terraform" | Snapshots are operational artifacts, not declarative infrastructure. | Snapshot policies declare the schedule; execution is API-driven. |
| Multi-array management in one provider block | "We have 5 FlashBlades" | Provider is a single API client; mixing breaks state isolation. | Terraform `provider` alias pattern. |
| Automatic resource name generation | "Let the provider generate unique names" | Generated names = unstable state. | User supplies names. |
| Hardware management (blades, drives) | "I see `/api/2.22/blades` in the API" | Hardware state is read-only observation, not declarative configuration. | Out of scope. |
| Session/client management | "We want to manage active sessions" | Sessions are ephemeral runtime state. | Operational runbooks, not IaC. |

---

## Feature Dependencies (full provider)

```
flashblade_object_store_account
    └──required by──> flashblade_bucket
                          └──required by──> flashblade_object_store_access_key (via account)
                          └──optional──> flashblade_object_store_access_policy

flashblade_nfs_export_policy
    └──optional attachment──> flashblade_file_system (via policy reference)

flashblade_smb_share_policy
    └──optional attachment──> flashblade_file_system (via policy reference)

flashblade_snapshot_policy
    └──optional attachment──> flashblade_file_system (via policy attachment)

flashblade_quota_policy
    └──optional attachment──> flashblade_file_system (via qos_policy reference)

flashblade_network_access_policy
    └──optional attachment──> flashblade_bucket (via policy reference)
    └──optional attachment──> flashblade_file_system (via policy reference)

flashblade_network_interface (NEW — v2.1.1)
    └──references──> subnet (pre-existing, by name)
    └──references──> flashblade_server (via attached_servers, optional)

flashblade_array_dns / flashblade_array_ntp / flashblade_array_smtp
    └──singleton resources──> no dependencies, managed independently
```

---

## Sources

- FlashBlade REST API 2.22 reference (`FLASHBLADE_API.md` in repo root)
- Existing provider implementation: `internal/provider/server_resource.go`, `internal/provider/object_store_virtual_host_resource.go`
- [HashiCorp Terraform Provider Best Practices](https://developer.hashicorp.com/terraform/plugin/best-practices)
- [terraform-plugin-framework Resources](https://developer.hashicorp.com/terraform/plugin/framework/resources)
- [Plan Modification — terraform-plugin-framework](https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification)

---
*Feature research for: Terraform provider for Pure Storage FlashBlade — network interface (VIP) milestone v2.1.1*
*Researched: 2026-03-30*
