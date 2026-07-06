# Testing

The provider is tested in three tiers, and almost everything runs **without a
real FlashBlade array** thanks to an in-memory mock HTTP server. Run
`make test` + `make lint` before every commit.

## Three tiers

1. **Unit (client)** — `internal/client/*_test.go`. Each CRUD method is tested
   against `httptest.NewServer()` with `/api/login` mocked to return an
   `x-auth-token`. Minimum per resource: `Get_Found`, `Get_NotFound`, `Post`,
   `Patch`, `Delete`.
2. **Mocked integration (provider)** — `internal/provider/*_test.go`. These drive
   the *real* provider through the mock server via `ProtoV6ProviderFactories`,
   using `resource.UnitTest` (no `TF_ACC` needed). The setup helper is
   `setupAcceptanceTest(t)` in `internal/provider/acceptance_test.go`: it spins up
   the mock, registers handlers, sets `FLASHBLADE_HOST`/`FLASHBLADE_API_TOKEN`,
   and returns the provider factory.
   > Note: `CONVENTIONS.md` refers to this helper as `testNewMockedProvider()`;
   > the actual symbol is `setupAcceptanceTest`.
3. **Acceptance (real array)** — `make testacc` (`TF_ACC=1`, 120 m timeout).
   `TestAcc_*` lifecycle tests can run against the mock *or* a real array. Per
   project memory, mock + swagger have both been wrong before, so **verify
   against a real array before tagging a release** (dev_overrides + real apply).

## Mock server (`internal/testmock/`)

- `server.go` — `MockServer` wraps `httptest.Server` + `http.ServeMux`.
  `NewMockServer` auto-registers `/api/login` (returns `mock-session-token`) and
  `/api/api_version` (returns a version list incl. `2.23`).
- `handlers/` — 46 files, one per resource group. Each exposes
  `Register<Resource>Handlers(mux)` returning a **thread-safe** store
  (`sync.Mutex` + `byName`/`byID` maps) whose `handle` method simulates real CRUD:
  UUID generation, query-param validation, cross-references (e.g. buckets validate
  account refs). Handlers register on the **versioned** path (`/api/2.23/buckets`).
- The empty-list-on-miss rule (return `[]` + 200, never 404) is what lets
  `getOneByName[T]` detect not-found — see [conventions.md](conventions.md#mock-handlers).

## Regression gate

`make test` enforces a `TEST_BASELINE` in `GNUmakefile` (currently ~811) — the
build **fails if the total test count drops**. A new resource must add ≥8 tests
(≥4 client, ≥3 resource, ≥1 data source). Total is ~817 `func Test*` across the
packages.

## Naming

`TestUnit_<Resource>_<Operation>[_<Variant>]` — e.g. `TestUnit_Target_Get_Found`,
`TestUnit_BucketResource_Lifecycle`. Acceptance-style lifecycle tests use
`TestAcc_*`. The Pulumi bridge tests intentionally use `TestProviderInfo_*`
(they test bridge config, not TF resource logic). Use `t.Fatalf` for setup
failures, `t.Errorf` for assertions.
