package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

// APIVersion is the FlashBlade REST API version this client targets.
const APIVersion = "2.22"

// Config holds all parameters for constructing a FlashBladeClient.
type Config struct {
	// Endpoint is the base URL of the FlashBlade array, e.g. "https://flashblade.example.com".
	Endpoint string

	// APIToken is the long-lived API token for session-based auth.
	APIToken string

	// OAuth2ClientID, OAuth2KeyID, OAuth2Issuer are used for OAuth2 token-exchange auth.
	OAuth2ClientID string
	OAuth2KeyID    string
	OAuth2Issuer   string

	// MaxRetries is the number of retry attempts for transient failures (default: 3).
	// The retry delay is fixed at 1000ms.
	MaxRetries int

	// CACertFile is the path to a PEM-encoded CA certificate file for TLS verification.
	CACertFile string

	// CACert is an inline PEM-encoded CA certificate string for TLS verification.
	CACert string

	// InsecureSkipVerify disables TLS certificate verification (for testing only).
	InsecureSkipVerify bool
}

// FlashBladeClient is the HTTP client for the FlashBlade REST API.
type FlashBladeClient struct {
	httpClient   *http.Client
	baseURL      string // endpoint + "/api/" + APIVersion
	sessionToken string
	useOAuth2    bool
}

// NewClient constructs a FlashBladeClient from the given Config.
// The provided context is used for authentication requests and can be
// used by callers to propagate cancellation from Terraform operations.
func NewClient(ctx context.Context, cfg Config) (*FlashBladeClient, error) {
	if cfg.Endpoint == "" {
		return nil, fmt.Errorf("client: Endpoint is required")
	}

	// Apply defaults.
	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = 3
	}
	const defaultRetryDelay = 1000 * time.Millisecond

	transport, err := buildTransport(cfg, defaultRetryDelay)
	if err != nil {
		return nil, fmt.Errorf("client: build transport: %w", err)
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	endpoint := strings.TrimRight(cfg.Endpoint, "/")
	baseURL := endpoint + "/api/" + APIVersion

	c := &FlashBladeClient{
		httpClient: httpClient,
		baseURL:    baseURL,
	}

	// Authenticate using API token session login (preferred for simplicity).
	if cfg.APIToken != "" {
		sessionToken, err := LoginWithAPIToken(ctx, httpClient, endpoint, cfg.APIToken)
		if err != nil {
			return nil, fmt.Errorf("client: login: %w", err)
		}
		c.sessionToken = sessionToken
		return c, nil
	}

	// Fall back to OAuth2 token exchange.
	if cfg.OAuth2ClientID != "" || cfg.OAuth2KeyID != "" {
		ts := NewFlashBladeTokenSource(ctx, endpoint, cfg.OAuth2ClientID, httpClient)
		oauthHTTPClient := oauth2.NewClient(ctx, ts)
		// Wrap the oauth2 client's transport with the retry transport too.
		oauthHTTPClient.Transport = &retryTransport{
			base:       oauthHTTPClient.Transport,
			maxRetries: cfg.MaxRetries,
			baseDelay:  defaultRetryDelay,
		}
		c.httpClient = oauthHTTPClient
		c.useOAuth2 = true
	}

	return c, nil
}

// buildTransport creates a TLS-aware http.RoundTripper with retry logic.
func buildTransport(cfg Config, retryDelay time.Duration) (http.RoundTripper, error) {
	tlsCfg := &tls.Config{
		InsecureSkipVerify: cfg.InsecureSkipVerify, //nolint:gosec
	}

	// Build a custom CA cert pool if a CA cert is provided.
	if cfg.CACertFile != "" || cfg.CACert != "" {
		certPool, err := x509.SystemCertPool()
		if err != nil {
			// SystemCertPool may fail on some platforms — fall back to empty pool.
			certPool = x509.NewCertPool()
		}

		var pemBytes []byte
		if cfg.CACertFile != "" {
			var err error
			pemBytes, err = os.ReadFile(cfg.CACertFile)
			if err != nil {
				return nil, fmt.Errorf("read CA cert file %q: %w", cfg.CACertFile, err)
			}
		} else {
			pemBytes = []byte(cfg.CACert)
		}

		if !certPool.AppendCertsFromPEM(pemBytes) {
			return nil, fmt.Errorf("failed to append CA certificate — check PEM format")
		}
		tlsCfg.RootCAs = certPool
	}

	base := &http.Transport{
		TLSClientConfig: tlsCfg,
	}

	return &retryTransport{
		base:       base,
		maxRetries: cfg.MaxRetries,
		baseDelay:  retryDelay,
	}, nil
}

// NegotiateVersion fetches the API version list and verifies that v2.22 is supported.
func (c *FlashBladeClient) NegotiateVersion(ctx context.Context) error {
	// Extract base endpoint from baseURL (strip "/api/2.22").
	endpoint := strings.TrimSuffix(c.baseURL, "/api/"+APIVersion)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"/api/api_version", nil)
	if err != nil {
		return fmt.Errorf("negotiate version: build request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("negotiate version: GET /api/api_version: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("negotiate version: unexpected status %d", resp.StatusCode)
	}

	var vr VersionResponse
	if err := json.NewDecoder(resp.Body).Decode(&vr); err != nil {
		return fmt.Errorf("negotiate version: decode response: %w", err)
	}

	for _, v := range vr.Versions {
		if v == APIVersion {
			return nil
		}
	}
	return fmt.Errorf("negotiate version: array does not support API version %s (available: %v)", APIVersion, vr.Versions)
}

// HTTPClient exposes the underlying *http.Client for use in tests.
func (c *FlashBladeClient) HTTPClient() *http.Client {
	return c.httpClient
}

// do is the shared HTTP execution helper.
func (c *FlashBladeClient) do(ctx context.Context, method, path string, body any) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.sessionToken != "" {
		req.Header.Set("X-Auth-Token", c.sessionToken)
	}

	return c.httpClient.Do(req)
}

// get performs a GET request and decodes the JSON response into result.
func (c *FlashBladeClient) get(ctx context.Context, path string, result any) error {
	resp, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if apiErr := ParseAPIError(resp); apiErr != nil {
		return apiErr
	}
	return json.NewDecoder(resp.Body).Decode(result)
}

// post performs a POST request and decodes the JSON response into result.
func (c *FlashBladeClient) post(ctx context.Context, path string, body, result any) error {
	resp, err := c.do(ctx, http.MethodPost, path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if apiErr := ParseAPIError(resp); apiErr != nil {
		return apiErr
	}
	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}

// patch performs a PATCH request and decodes the JSON response into result.
func (c *FlashBladeClient) patch(ctx context.Context, path string, body, result any) error {
	resp, err := c.do(ctx, http.MethodPatch, path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if apiErr := ParseAPIError(resp); apiErr != nil {
		return apiErr
	}
	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}

// delete performs a DELETE request.
func (c *FlashBladeClient) delete(ctx context.Context, path string) error {
	resp, err := c.do(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return ParseAPIError(resp)
}

// getOneByName queries a FlashBlade list endpoint filtered by name and returns the single
// matching item. Returns a 404 APIError if no items match.
// path is the API endpoint with ?names= query param already set (e.g., "/buckets?names=mybucket").
// label is a human-readable resource type for the error message (e.g., "bucket").
func getOneByName[T any](c *FlashBladeClient, ctx context.Context, path, label, name string) (*T, error) {
	var resp ListResponse[T]
	if err := c.get(ctx, path, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, &APIError{StatusCode: 404, Message: fmt.Sprintf("%s %q not found", label, name)}
	}
	return &resp.Items[0], nil
}

// pollUntilGone polls a FlashBlade list endpoint with ?destroyed=true until the item
// disappears (empty items response), indicating eradication is complete.
// basePath is the endpoint without query params (e.g., "/buckets").
// label is a human-readable resource type for error messages.
func pollUntilGone[T any](c *FlashBladeClient, ctx context.Context, basePath, label, name string) error {
	path := basePath + "?names=" + url.QueryEscape(name) + "&destroyed=true"
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("pollUntilGone(%s): context cancelled while waiting for %q to eradicate: %w", label, name, ctx.Err())
		default:
		}

		var resp ListResponse[T]
		err := c.get(ctx, path, &resp)
		if err != nil {
			if IsNotFound(err) {
				return nil
			}
			return fmt.Errorf("pollUntilGone(%s): GET error: %w", label, err)
		}

		if len(resp.Items) == 0 {
			return nil
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("pollUntilGone(%s): context cancelled while waiting for %q to eradicate: %w", label, name, ctx.Err())
		case <-time.After(2 * time.Second):
		}
	}
}
