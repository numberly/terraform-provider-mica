package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/oauth2"
)

// LoginWithAPIToken performs a session login using a long-lived API token.
// It POSTs to {endpoint}/api/login with the api-token header and extracts the
// resulting session token from the x-auth-token response header.
func LoginWithAPIToken(ctx context.Context, httpClient *http.Client, endpoint, apiToken string) (string, error) {
	if httpClient == nil {
		return "", fmt.Errorf("login: httpClient must not be nil")
	}

	loginURL := strings.TrimRight(endpoint, "/") + "/api/login"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, loginURL, nil)
	if err != nil {
		return "", fmt.Errorf("login: build request: %w", err)
	}
	req.Header.Set("api-token", apiToken)

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("login: POST /api/login: %w", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("login: unexpected status %d", resp.StatusCode)
	}

	sessionToken := resp.Header.Get("x-auth-token")
	if sessionToken == "" {
		return "", fmt.Errorf("login: x-auth-token header missing from response")
	}
	return sessionToken, nil
}

// oauthTokenResponse is the parsed body from /oauth2/1.0/token.
type oauthTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

// FlashBladeTokenSource is an oauth2.TokenSource that exchanges a FlashBlade
// API token for a short-lived Bearer token using the non-standard
// urn:ietf:params:oauth:grant-type:token-exchange grant type.
//
// Tokens are cached until they expire to avoid unnecessary round-trips.
// The context provided at construction is used for all token exchanges, ensuring
// that cancelling a Terraform operation also cancels the in-flight token exchange.
type FlashBladeTokenSource struct {
	ctx         context.Context
	endpoint    string
	apiToken    string
	httpClient  *http.Client
	mu          sync.Mutex
	cachedToken *oauth2.Token
}

// NewFlashBladeTokenSource creates a new FlashBladeTokenSource.
// The provided ctx is stored and used for all token exchanges via Token().
func NewFlashBladeTokenSource(ctx context.Context, endpoint, apiToken string, httpClient *http.Client) *FlashBladeTokenSource {
	if httpClient == nil {
		panic("NewFlashBladeTokenSource: httpClient must not be nil")
	}
	return &FlashBladeTokenSource{
		ctx:        ctx,
		endpoint:   strings.TrimRight(endpoint, "/"),
		apiToken:   apiToken,
		httpClient: httpClient,
	}
}

// Token returns a valid OAuth2 token, refreshing it if necessary.
// Satisfies the oauth2.TokenSource interface.
func (ts *FlashBladeTokenSource) Token() (*oauth2.Token, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.cachedToken != nil && ts.cachedToken.Valid() {
		return ts.cachedToken, nil
	}

	tok, err := ts.fetchToken(ts.ctx)
	if err != nil {
		return nil, err
	}
	ts.cachedToken = tok
	return tok, nil
}

// FetchTokenWithContext exposes context-aware token fetching for callers
// that have a context available (e.g. tests, direct usage).
func (ts *FlashBladeTokenSource) FetchTokenWithContext(ctx context.Context) (*oauth2.Token, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.fetchToken(ctx)
}

func (ts *FlashBladeTokenSource) fetchToken(ctx context.Context) (*oauth2.Token, error) {
	tokenURL := ts.endpoint + "/oauth2/1.0/token"

	form := url.Values{
		"grant_type":         {"urn:ietf:params:oauth:grant-type:token-exchange"},
		"subject_token":      {ts.apiToken},
		"subject_token_type": {"urn:ietf:params:oauth:token-type:jwt"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("token exchange: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := ts.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token exchange: POST /oauth2/1.0/token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("token exchange: read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange: unexpected HTTP %d from %s/oauth2/1.0/token", resp.StatusCode, ts.endpoint)
	}

	var tr oauthTokenResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		return nil, fmt.Errorf("token exchange: parse response: %w", err)
	}

	expiry := time.Now().Add(time.Duration(tr.ExpiresIn) * time.Second)
	return &oauth2.Token{
		AccessToken: tr.AccessToken,
		TokenType:   tr.TokenType,
		Expiry:      expiry,
	}, nil
}
