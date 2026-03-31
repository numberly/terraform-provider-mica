---
phase: quick-7
plan: 01
subsystem: server
tags: [refactor, schema-migration, dns, directory-services, state-upgrade]
dependency_graph:
  requires: []
  provides: [server-dns-namedreference, server-directory-services, schema-v2]
  affects: [server_resource, server_data_source, server_client_models]
tech_stack:
  added: []
  patterns: [flat-string-list-for-api-references, state-upgrader-chain]
key_files:
  created: []
  modified:
    - internal/client/models_exports.go
    - internal/testmock/handlers/servers.go
    - internal/client/servers_test.go
    - internal/provider/server_resource.go
    - internal/provider/server_data_source.go
    - internal/provider/server_resource_test.go
    - internal/provider/server_data_source_test.go
decisions:
  - "ServerDNS struct deleted — DNS field is []NamedReference matching real API response"
  - "directory_services added as Computed-only []NamedReference field in Server struct"
  - "Schema version bumped 1->2; v1 nested DNS objects lack name field so reset to null on upgrade"
  - "v0->v1 upgrader outputs serverV1StateModel (intermediate nested format); framework chains to v1->v2"
  - "dnsNamesToRefs() replaces mapModelDNSToClient; DNS conversion is now 4 lines not 50"
metrics:
  duration: ~20 minutes
  completed_date: "2026-03-30"
  tasks_completed: 2
  files_modified: 7
---

# Quick Task 7: Refactor Server DNS to NamedReference Summary

**One-liner:** Fixed server DNS schema from incorrect inline objects (domain/nameservers/services) to flat list of DNS config names using NamedReference, matching the real FlashBlade API response format.

## What Was Built

The `flashblade_server` resource and data source had a wrong DNS model — the provider stored nested objects with domain/nameservers/services fields, but the FlashBlade API returns `[{name, id, resource_type}]` references. This task corrected the data model end-to-end.

## Changes Made

### Task 1: Client models, mock handler, client tests (commit: 8ae93db)

- `ServerDNS` struct deleted from `models_exports.go`
- `Server.DNS` type changed from `[]ServerDNS` to `[]NamedReference`
- `Server.DirectoryServices []NamedReference` added (new field)
- `ServerPost.DNS` and `ServerPatch.DNS` changed to `[]NamedReference`
- Mock `AddServer` now seeds `DNS: [{Name: "management"}]` and `DirectoryServices: [{Name: "srv-backup_nfs"}]`
- Mock `handlePatch` DNS unmarshal updated to `[]NamedReference`
- Client unit tests updated to assert on `.Name` field instead of `.Domain`

### Task 2: Resource, data source, provider tests (commit: d9dc0bd)

**server_resource.go:**
- `serverDNSModel`, `serverDNSAttrTypes()`, `serverDNSObjectType()`, `mapModelDNSToClient()` deleted
- `serverResourceModel` updated: `DNS types.List` now holds `StringType` elements; `DirectoryServices types.List` added
- Schema bumped to `Version: 2`
- `dns` attribute: `ListNestedAttribute` → `ListAttribute{ElementType: types.StringType}`
- `directory_services`: new `Computed` `ListAttribute{ElementType: types.StringType}`
- v0→v1 upgrader updated to output `serverV1StateModel` (nested DNS intermediate format)
- v1→v2 upgrader added: reads nested v1 state, resets DNS to `ListNull(StringType)`, adds `DirectoryServices: ListNull(StringType)`, preserves NetworkInterfaces/CascadeDelete/Timeouts
- `dnsNamesToRefs()` helper replaces `mapModelDNSToClient()`
- `mapServerToModel()` simplified: 50-line DNS block → 12 lines

**server_data_source.go:**
- `serverDataSourceModel` gets `DirectoryServices types.List`
- Schema: `dns` → `ListAttribute{StringType}`, `directory_services` added as Computed
- `mapServerDNSToDataSourceModel()` deleted; inline mapping in Read
- DNS and DirectoryServices mapped inline

**Tests:**
- `buildServerType()` / `nullServerConfig()`: `dnsType` object removed, `dns` → `tftypes.List{ElementType: tftypes.String}`, `directory_services` added
- `serverPlanWithDNS()` rewritten to accept `[]string` names (not domain/nameservers)
- `buildServerDSType()` / `nullServerDSConfig()`: same changes + `directory_services` added
- `TestUnit_Server_SchemaVersion`: updated assertion from 1 to 2
- `TestUnit_Server_StateUpgradeV0ToV1`: updated target schema to use v1 PriorSchema
- `TestUnit_Server_StateUpgradeV1ToV2`: new test for v1→v2 path
- `TestUnit_ServerDataSource`: added assertions for DNS names and directory_services names from mock seed

## Verification

```
go build ./...   — OK
go vet ./...     — OK
go test ./... -count=1 — 668 passed
```

## Deviations from Plan

None — plan executed exactly as written.

## Self-Check

- `internal/client/models_exports.go` — ServerDNS deleted, Server has DirectoryServices: confirmed
- `internal/provider/server_resource.go` — Version: 2, directory_services, dnsNamesToRefs: confirmed
- `internal/provider/server_data_source.go` — directory_services schema and mapping: confirmed
- Commits 8ae93db and d9dc0bd — both exist in git log: confirmed

## Self-Check: PASSED
