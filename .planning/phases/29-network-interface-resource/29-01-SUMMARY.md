---
phase: 29-network-interface-resource
plan: "01"
subsystem: client-layer
tags: [network-interfaces, client, mock, models]
one_liner: "NetworkInterface structs + CRUD client methods + mock handler for /api/2.22/network-interfaces"

dependency_graph:
  requires: []
  provides:
    - internal/client/models_network.go (NetworkInterface, NetworkInterfacePost, NetworkInterfacePatch)
    - internal/client/network_interfaces.go (Get/List/Post/Patch/Delete client methods)
    - internal/testmock/handlers/network_interfaces.go (RegisterNetworkInterfaceHandlers, AddNetworkInterface)
  affects:
    - Plan 29-02 (resource, data source, and acceptance tests depend on these foundations)

tech_stack:
  added: []
  patterns:
    - "?names= query param for user-provided resource names (same as subnets)"
    - "Full-replace semantics for services/attached_servers (no omitempty on Patch fields)"
    - "Raw map[string]json.RawMessage for PATCH decoding in mock handler"
    - "*string + omitempty for Address in Patch (true optional PATCH semantics)"

key_files:
  created:
    - internal/client/network_interfaces.go
    - internal/testmock/handlers/network_interfaces.go
  modified:
    - internal/client/models_network.go

decisions:
  - "NetworkInterfacePatch.Services and .AttachedServers have NO omitempty — clearing requires sending [] in JSON"
  - "Address in NetworkInterfacePatch uses *string + omitempty for true PATCH semantics"
  - "Mock handlePatch uses full-replace on services and attached_servers (consistent with API spec)"
  - "AddNetworkInterface seeder defaults: Gateway=10.21.200.1, MTU=1500, Netmask=255.255.255.0, VLAN=0"

metrics:
  duration: "2 minutes"
  completed_date: "2026-03-30"
  tasks_completed: 2
  tasks_total: 2
  files_created: 2
  files_modified: 1
---

# Phase 29 Plan 01: Network Interface Client Layer Summary

NetworkInterface structs + CRUD client methods + mock handler for /api/2.22/network-interfaces, following the exact same patterns as Phase 28's subnet implementation.

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Add NetworkInterface model structs and client CRUD methods | 91ff984 | models_network.go, network_interfaces.go |
| 2 | Create mock handler for network interfaces | dcc00d9 | handlers/network_interfaces.go |

## Deliverables

### `internal/client/models_network.go` (modified)

Three new structs appended after `LinkAggregationGroup`:

- `NetworkInterface` — GET response with all fields (ID, Name, Address, Enabled, Gateway, MTU, Netmask, Services, Subnet, Type, VLAN, AttachedServers, Realms)
- `NetworkInterfacePost` — writable fields for POST (Address, Services, Subnet, Type, AttachedServers); name passed via ?names= only
- `NetworkInterfacePatch` — mutable fields for PATCH; Services/AttachedServers **without** omitempty for full-replace; Address as *string + omitempty

### `internal/client/network_interfaces.go` (created)

Five CRUD methods following subnets.go pattern exactly:
- `GetNetworkInterface(ctx, name)` — getOneByName helper
- `ListNetworkInterfaces(ctx)` — GET all
- `PostNetworkInterface(ctx, name, body)` — POST ?names=
- `PatchNetworkInterface(ctx, name, body)` — PATCH ?names=
- `DeleteNetworkInterface(ctx, name)` — DELETE ?names=

### `internal/testmock/handlers/network_interfaces.go` (created)

- `RegisterNetworkInterfaceHandlers(mux)` — registers on `/api/2.22/network-interfaces`
- `AddNetworkInterface(name, address, subnetName, niType, service)` — seeder for tests
- `handlePost` — reads name from ?names=, creates NI with body fields + computed defaults
- `handlePatch` — raw map decoding, full-replace on services/attached_servers
- `handleGet` — optional ?names= filter
- `handleDelete` — removes from both byName and byID maps

## Deviations from Plan

None - plan executed exactly as written.

## Verification

```
go build ./internal/client/           PASS
go build ./internal/testmock/...      PASS
go vet ./internal/client/ ./internal/testmock/...   PASS
```
