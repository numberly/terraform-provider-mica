# Architecture

The provider is a single Go binary (`main.go`) that Terraform/OpenTofu launch
over gRPC. All logic lives under `internal/`:

```
internal/
  client/          # HTTP client, auth, transport, generics, model structs
  provider/        # Terraform resources, data sources, validators, helpers
  testmock/        # In-memory mock HTTP server + per-resource handlers (tests)
```

Data flow: a Terraform config → the provider's `Configure` builds a
`*client.FlashBladeClient` → each resource/data source re-casts that client and
calls typed CRUD methods → the client speaks the FlashBlade REST API.

## Client HTTP layer (`internal/client/`)

**`client.go`** is the core.

- `Config` holds `endpoint`, `APIToken`, the OAuth2 triple
  (`OAuth2ClientID`/`OAuth2KeyID`/`OAuth2Issuer`), `MaxRetries`, and TLS options.
- `FlashBladeClient` holds the `httpClient`, a `baseURL`, the `sessionToken`,
  and `useOAuth2`.
- **API-version prefixing is central**: `baseURL = endpoint + "/api/" + APIVersion`
  is built once (`APIVersion` const = `2.23`). Every path passed to the internal
  `do()` is appended to it, so resource code never writes `/api/2.23/...` — it
  passes bare paths like `buckets`. This is why upgrading the API version is a
  one-line change plus schema work (see [api-versioning.md](api-versioning.md)).
- `do()` (via `get`/`post`/`patch`/`delete` helpers) sets `Content-Type`, adds
  the `X-Auth-Token` header when a session exists, and runs `ParseAPIError`.
- `NegotiateVersion` GETs `/api/api_version` and asserts `2.23` is offered;
  called by the provider at configure time.

**`auth.go`** — two mechanisms:

- **API token** (`LoginWithAPIToken`): POST `/api/login` with the `api-token`
  request header; the session token is read from the **`x-auth-token`** response
  header. Preferred path.
- **OAuth2 token exchange** (`FlashBladeTokenSource`, an `oauth2.TokenSource`):
  POSTs to `/oauth2/1.0/token` with grant type
  `urn:ietf:params:oauth:grant-type:token-exchange`; caches the access token
  (mutex-guarded) until expiry.

**`transport.go`** — `retryTransport` (a `RoundTripper`): injects a random
`X-Request-ID`, snapshots the request body for replay, and retries retryable
responses up to `maxRetries`. Backoff is exponential (`baseDelay * 2^attempt`),
capped at 30 s, with ±20 % jitter, and honours context cancellation.

**`errors.go`** — `APIError{StatusCode, Message, Errors[]}`. Helpers:
`IsNotFound` (404 **and** 400 whose sub-error message ends with "does not
exist"), `IsConflict` (409), `isAlreadyExists`, `IsRetryable` (429 or ≥500).

> ⚠️ `IsNotFound` is coupled to the API's error-message shape. It's documented in
> the source; if the API changes wording, not-found detection can silently break.

## Generic CRUD helpers (`client.go`)

Resource client methods are thin wrappers over Go-generic helpers — **never
hand-roll these**:

| Helper | Purpose |
|--------|---------|
| `getOneByName[T](c, ctx, path, label, name)` | GET a `ListResponse[T]`, return first item or synthetic 404 |
| `postOne[TBody,TResp]` / `patchOne[...]` | POST/PATCH returning first item of the list response |
| `listAll[T](c, ctx, basePath, params)` | Pagination — follows `continuation_token` until empty |
| `pollUntilGone[T](c, ctx, basePath, label, name)` | Poll `?names=...&destroyed=true` every 2 s until the item disappears — the eradication wait |

## Model structs — the three-struct pattern

Each resource defines **three** structs in `internal/client/models_<domain>.go`
because the API takes different fields per verb (example: `Bucket` in
`models_storage.go`):

| Struct | Suffix | Use | Field style |
|--------|--------|-----|-------------|
| `Bucket` | (none) | GET response | plain types, `NamedReference` for refs |
| `BucketPost` | `Post` | POST body | plain + `omitempty`; `*bool`/`*int64` only where zero is meaningful |
| `BucketPatch` | `Patch` | PATCH body | **every field a pointer** — `nil` = omit, non-nil = send |

The pointer rules (and the `**NamedReference` / `*[]T` subtleties) are the most
error-prone part of the codebase — see [conventions.md](conventions.md#model-structs).

## Provider registration (`internal/provider/provider.go`)

- `FlashBladeProvider` implements `provider.Provider`; `New(version)` is the factory.
- **Schema**: top-level `endpoint`, TLS options (`ca_cert_file`, `ca_cert`,
  `insecure_skip_verify`), `max_retries`, and a nested `auth` block
  (`api_token` sensitive + nested `oauth2`).
- **`Configure`** resolves endpoint/auth with env fallbacks (`FLASHBLADE_HOST`,
  `FLASHBLADE_API_TOKEN`, `FLASHBLADE_OAUTH2_*`), validates at least one auth
  method, builds the client, calls `NegotiateVersion`, and injects the client
  into **both** `resp.ResourceData` and `resp.DataSourceData`. Every resource
  re-casts it: `req.ProviderData.(*client.FlashBladeClient)`.
- **`Resources()`** returns 56 constructors; **`DataSources()`** returns 44,
  grouped by domain. Registering a new resource = appending its constructor here.

## Resource pattern

Every resource implements four interfaces: `Resource`,
`ResourceWithConfigure`, `ResourceWithImportState`, `ResourceWithUpgradeState`
(some add `ResourceWithValidateConfig`). Canonical reference:
`internal/provider/bucket_resource.go`.

- **Plan modifiers**: `UseStateForUnknown()` on stable computed fields (`id`,
  `created`); `RequiresReplace()` on immutable fields (`name`); **nothing** on
  volatile fields (masking those would hide drift).
- **Drift detection**: on `Read`, every mutable/computed field is compared and
  logged with `tflog.Debug(ctx, "drift detected", …)` — logged, never errored.
- **ImportState**: import by **name** (not UUID); always null the timeouts and
  write-once fields.
- **Schema versioning**: `SchemaVersion` starts at 0; changing attributes bumps
  it and adds a state upgrader with an exact `PriorSchema`. Real example:
  `directory_service_role_resource.go` (v0→v1). See
  [conventions.md](conventions.md#state-upgraders).
- **Soft-delete** (buckets, file systems only): two-phase destroy — PATCH
  `{destroyed:true}`, then, if `destroy_eradicate_on_delete`, DELETE +
  `pollUntilGone`. In `bucket_resource.go` this is `DestroyAndEradicateBucket`
  in `internal/client/buckets.go`.

## Data source pattern

Data sources implement only `DataSource` + `DataSourceWithConfigure` — no
timeouts, no plan modifiers. `name` is Required, everything else Computed.
Not-found is an `AddError` (not `RemoveResource`). Reference:
`internal/provider/target_data_source.go`.

See also: **[Testing](testing.md)** for how all of this is exercised without a
real array, and **[Conventions](conventions.md)** for the full rule set.
