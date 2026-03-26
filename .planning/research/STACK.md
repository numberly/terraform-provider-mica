# Stack Research

**Domain:** Terraform Provider (Go) — REST API wrapping, storage infrastructure
**Researched:** 2026-03-26
**Confidence:** HIGH — all versions verified against official scaffolding go.mod (March 2026) and pkg.go.dev

## Recommended Stack

### Core Technologies

| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| Go | 1.25.x | Implementation language | Only supported language for Terraform providers. v1.25 is current LTS and required minimum for all terraform-plugin-* modules as of March 2026. The official scaffolding go.mod pins `go 1.25.5`. |
| terraform-plugin-framework | v1.19.0 | Provider CRUD scaffold, schema, plan modifiers | HashiCorp's current recommended SDK. Protocol v6, better type safety than SDKv2, native plan modifiers, built-in diagnostics, required for all new providers. SDKv2 is in maintenance-only mode. |
| terraform-plugin-go | v0.31.0 | Low-level Terraform protocol bindings | Indirect dependency of framework; occasionally needed directly for custom type implementations or protocol-level hooks. Listed explicitly for version pinning. |
| terraform-plugin-testing | v1.15.0 | Acceptance and integration test runner | The official test harness that runs real `terraform` CLI against provider code. Required for acceptance tests. Provides `resource.Test`, `TestCase`, `TestStep` constructs. |
| terraform-plugin-log | v0.10.0 | Structured logging inside provider | HashiCorp's structured logging for providers. Integrates with Terraform's log streaming. Use `tflog` package, not `log` stdlib. |

### Supporting Libraries

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| golang.org/x/oauth2 | v0.27+ | OAuth2 client_credentials token flow | FlashBlade `POST /oauth2/1.0/token` token exchange. `oauth2/clientcredentials.Config.Client()` returns an auto-refreshing `*http.Client`. Use for production auth. |
| golang.org/x/oauth2/clientcredentials | (same module) | Two-legged OAuth2 subpackage | Specifically for client_credentials grant. Avoids hand-rolling token refresh logic. |
| net/http (stdlib) | — | HTTP transport base | All API calls. Build a custom `http.RoundTripper` chain: TLS config → auth injection → request tracing. Do NOT use third-party HTTP clients. |
| crypto/tls (stdlib) | — | Custom CA certificate support | Enterprise environments with private PKI. `tls.Config{RootCAs: pool}` injected into `http.Transport`. |
| crypto/x509 (stdlib) | — | CA cert pool construction | `x509.SystemCertPool()` + `pool.AppendCertsFromPEM()` for custom CA. |
| encoding/json (stdlib) | — | JSON serialisation/deserialisation | FlashBlade API speaks JSON throughout. Keep stdlib; no need for third-party JSON library. |
| github.com/hashicorp/terraform-plugin-docs | v0.21+ | Registry documentation generation | `tfplugindocs generate` produces Terraform Registry-compatible Markdown from schema descriptions and templates. Run as part of CI/pre-release. |
| net/http/httptest (stdlib) | — | Integration test mock HTTP server | `httptest.NewServer()` for mocked-API integration tests (the "middle tier" without a real FlashBlade). Zero external dependencies. |

### Development Tools

| Tool | Purpose | Notes |
|------|---------|-------|
| goreleaser | Cross-platform binary release | HashiCorp-mandated release pipeline. Builds linux/darwin/windows amd64/arm64 binaries, signs checksums with GPG. Copy `.goreleaser.yml` from `hashicorp/terraform-provider-scaffolding-framework`. |
| tfplugindocs (`github.com/hashicorp/terraform-plugin-docs`) | Docs generation CLI | `go generate ./...` with `//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs`. Install via `tools.go` pattern. |
| golangci-lint v2 | Go linting | 40+ linters in one binary. v2 released March 2025 with revamped config (`.golangci.yml`). Use `linters.default: standard` then enable extras. |
| tfproviderlint (`github.com/bflad/tfproviderlint`) | Terraform-specific Go lint | Catches provider-specific mistakes: missing Read calls, import state patterns, schema issues. Complements golangci-lint. Run `tfproviderlintx` for extended checks. |
| Terraform CLI | Acceptance test runner | `terraform-plugin-testing` shells out to a real `terraform` binary. Set `TF_ACC=1` env var to enable acceptance tests. Pin version via `TF_CLI_PATH` or GitHub Actions setup step. |
| GitHub Actions (`hashicorp/ghaction-terraform-provider-release`) | CI/CD | HashiCorp's reusable workflow for provider release. Handles GoReleaser, GPG signing, Registry publish. |

## Installation

```bash
# Initialize module (greenfield)
go mod init github.com/soulkyu/terraform-provider-flashblade

# Core framework + testing
go get github.com/hashicorp/terraform-plugin-framework@v1.19.0
go get github.com/hashicorp/terraform-plugin-testing@v1.15.0
go get github.com/hashicorp/terraform-plugin-log@v0.10.0
go get github.com/hashicorp/terraform-plugin-go@v0.31.0

# Auth
go get golang.org/x/oauth2@latest

# Dev tools (tools.go pattern — keeps versions pinned in go.mod)
cat >> tools/tools.go << 'EOF'
//go:build tools

package tools

import (
    _ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"
)
EOF
go get github.com/hashicorp/terraform-plugin-docs@latest

# Install dev binaries
go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest
go install github.com/bflad/tfproviderlint/cmd/tfproviderlintx@latest

# golangci-lint v2
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.x
```

## Alternatives Considered

| Recommended | Alternative | When to Use Alternative |
|-------------|-------------|-------------------------|
| terraform-plugin-framework v1.19 | terraform-plugin-sdk/v2 | Never for new providers. SDKv2 is feature-frozen, maintenance-only. Only valid when maintaining a legacy provider that cannot be migrated. |
| stdlib net/http | resty, go-resty, cleanhttp | Only if the API requires complex retry logic not worth handrolling. hashicorp/go-retryablehttp is an acceptable complement for retry-with-backoff on transient errors. |
| httptest.NewServer (stdlib) | WireMock, httpmock jarcoal | WireMock is Java-native, adds container overhead. jarcoal/httpmock is reasonable but httptest.NewServer is zero-dep and idiomatic Go for this use case. |
| golang.org/x/oauth2 | Hand-rolled token refresh | Never hand-roll. x/oauth2 handles refresh, concurrency, and expiry correctly. Its `clientcredentials` sub-package covers the FlashBlade token exchange exactly. |
| goreleaser | Manual Makefile multi-platform build | Only for internal-only providers that never publish to Registry. If Registry publication is planned (it is here), use goreleaser from day one. |

## What NOT to Use

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| terraform-plugin-sdk/v2 | Feature-frozen, no new framework features (plan modifiers, write-only attrs, identity). HashiCorp's official position: migrate away. | terraform-plugin-framework v1.19 |
| terraform-plugin-mux | Only needed when combining framework + SDKv2 providers in one binary. Adds complexity with no benefit for pure-framework providers. | Direct framework provider, no mux needed |
| Third-party HTTP clients (resty, fiber, etc.) | Adds deps, hides transport layer — breaks custom CA/auth RoundTripper chain. stdlib net/http is sufficient and transparent. | stdlib net/http with RoundTripper chain |
| `log` stdlib inside provider | Terraform has its own structured log capture. stdlib `log` writes bypass Terraform's log routing and appear as raw stderr noise. | `github.com/hashicorp/terraform-plugin-log/tflog` |
| Hardcoded API version paths | FlashBlade exposes `/api/2.22/...` — hardcoding across 538 endpoints creates drift risk when upgrading. | Const for API version, assembled in HTTP client constructor |

## Stack Patterns by Variant

**For API token auth (dev/local):**
- Build an `http.RoundTripper` that injects `x-auth-token` header
- Configured via provider block `api_token` attribute
- Use `stringplanmodifier.UseStateForUnknown()` on computed-only fields

**For OAuth2 client_credentials auth (production):**
- Use `golang.org/x/oauth2/clientcredentials.Config{...}.TokenSource(ctx)`
- Wrap with `oauth2.NewClient(ctx, tokenSource)` which auto-refreshes
- FlashBlade token endpoint: `POST /oauth2/1.0/token` with `grant_type=urn:ietf:params:oauth:grant-type:token-exchange`

**For custom CA certificate (enterprise TLS):**
```go
pool, _ := x509.SystemCertPool()
pool.AppendCertsFromPEM(customCAPEM)
transport := &http.Transport{
    TLSClientConfig: &tls.Config{RootCAs: pool},
}
```
Layer this transport under the auth RoundTripper via composition.

**For mocked integration tests (CI without FlashBlade):**
- Use `httptest.NewServer(mux)` where `mux` is an `http.ServeMux` recording expected calls
- Provider configured to point at `httptest.Server.URL`
- Run with `go test ./... -run TestIntegration` (no `TF_ACC=1`)

**For acceptance tests (real FlashBlade required):**
- Set `TF_ACC=1`, `FLASHBLADE_ENDPOINT`, `FLASHBLADE_API_TOKEN`
- Use `resource.Test(t, resource.TestCase{...})` from `terraform-plugin-testing`
- Gate in CI: run only on manual trigger or dedicated FlashBlade environment

## Version Compatibility

| Package | Compatible With | Notes |
|---------|-----------------|-------|
| terraform-plugin-framework v1.19.0 | Go 1.25+, terraform-plugin-go v0.31.0 | Sourced from official scaffolding go.mod (March 10, 2026) |
| terraform-plugin-testing v1.15.0 | Go 1.25+, Terraform CLI 1.0+ | Same release date as framework v1.19.0 — released together |
| terraform-plugin-log v0.10.0 | Go 1.25+ | Indirect via framework; pin explicitly to avoid transitive bumps |
| golang.org/x/oauth2 | Go 1.21+ | Follows x/ library policy; no breaking changes expected |
| golangci-lint v2 | Go 1.22+ for targets | v2 config format incompatible with v1 `.golangci.yml` — write fresh config |

## Sources

- [terraform-plugin-framework v1.19.0 — pkg.go.dev](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-framework) — version and Go 1.25 requirement confirmed HIGH confidence
- [terraform-plugin-testing v1.15.0 — pkg.go.dev](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-testing) — version and release date confirmed HIGH confidence
- [terraform-provider-scaffolding-framework go.mod](https://raw.githubusercontent.com/hashicorp/terraform-provider-scaffolding-framework/main/go.mod) — authoritative dependency pinning, March 2026 HIGH confidence
- [HashiCorp Plugin Framework benefits](https://developer.hashicorp.com/terraform/plugin/framework-benefits) — SDKv2 deprecation rationale HIGH confidence
- [golang.org/x/oauth2/clientcredentials](https://pkg.go.dev/golang.org/x/oauth2/clientcredentials) — token flow pattern HIGH confidence
- [terraform-plugin-docs GitHub](https://github.com/hashicorp/terraform-plugin-docs) — tfplugindocs CLI usage MEDIUM confidence (version unverified)
- [bflad/tfproviderlint GitHub](https://github.com/bflad/tfproviderlint) — provider-specific linter MEDIUM confidence (last verified active)
- [golangci-lint v2 announcement](https://ldez.github.io/blog/2025/03/23/golangci-lint-v2/) — v2 config format change MEDIUM confidence
- [hashicorp/ghaction-terraform-provider-release](https://github.com/hashicorp/ghaction-terraform-provider-release) — official release workflow HIGH confidence

---
*Stack research for: Terraform Provider for Pure Storage FlashBlade*
*Researched: 2026-03-26*
