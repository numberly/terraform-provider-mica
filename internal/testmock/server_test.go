package testmock_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/numberly/opentofu-provider-flashblade/internal/testmock"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock/handlers"
)

// doJSON is a convenience helper for making JSON requests to the mock server.
func doJSON(t *testing.T, method, url string, body any) *http.Response {
	t.Helper()
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		bodyReader = bytes.NewReader(data)
	}
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	return resp
}

// decodeJSON decodes a JSON response body into v.
func decodeJSON(t *testing.T, resp *http.Response, v any) {
	t.Helper()
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		t.Fatalf("decode JSON: %v", err)
	}
}

func TestUnit_MockServer_FullCRUDLifecycle(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterFileSystemHandlers(ms.Mux)

	base := ms.URL()

	// Step 1: POST /api/login — verify x-auth-token header.
	resp := doJSON(t, http.MethodPost, base+"/api/login", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /login: expected 200, got %d", resp.StatusCode)
	}
	token := resp.Header.Get("x-auth-token")
	if token == "" {
		t.Fatal("POST /login: expected x-auth-token header, got empty")
	}
	resp.Body.Close()

	// Step 2: GET /api/api_version — verify "2.22" present.
	resp = doJSON(t, http.MethodGet, base+"/api/api_version", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /api/api_version: expected 200, got %d", resp.StatusCode)
	}
	var versionResp struct {
		Versions []string `json:"versions"`
	}
	decodeJSON(t, resp, &versionResp)
	found := false
	for _, v := range versionResp.Versions {
		if v == "2.22" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("GET /api/api_version: expected 2.22 in versions, got %v", versionResp.Versions)
	}

	// Step 3: POST /api/2.22/file-systems?names=test-fs — create file system.
	// The FlashBlade API requires the name as a ?names= query parameter, not in the body.
	resp = doJSON(t, http.MethodPost, base+"/api/2.22/file-systems?names=test-fs", map[string]any{
		"provisioned": 1073741824,
	})
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("POST /api/2.22/file-systems: expected 200, got %d: %s", resp.StatusCode, body)
	}
	var createResp struct {
		Items []struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			Provisioned int64  `json:"provisioned"`
			Created     int64  `json:"created"`
		} `json:"items"`
	}
	decodeJSON(t, resp, &createResp)
	if len(createResp.Items) == 0 {
		t.Fatal("POST /api/2.22/file-systems: expected items in response")
	}
	fs := createResp.Items[0]
	if fs.ID == "" {
		t.Error("expected non-empty ID")
	}
	if fs.Name != "test-fs" {
		t.Errorf("expected name test-fs, got %q", fs.Name)
	}
	if fs.Created == 0 {
		t.Error("expected non-zero created timestamp")
	}
	fsID := fs.ID

	// Step 4: GET /api/2.22/file-systems?names=test-fs — verify file system returned.
	resp = doJSON(t, http.MethodGet, base+"/api/2.22/file-systems?names=test-fs", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET file-systems?names=test-fs: expected 200, got %d", resp.StatusCode)
	}
	var getResp struct {
		Items []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"items"`
	}
	decodeJSON(t, resp, &getResp)
	if len(getResp.Items) == 0 {
		t.Fatal("GET file-systems?names=test-fs: expected items")
	}
	if getResp.Items[0].Name != "test-fs" {
		t.Errorf("expected name test-fs, got %q", getResp.Items[0].Name)
	}

	// Step 5: PATCH /api/2.22/file-systems?ids={id} — update provisioned size.
	resp = doJSON(t, http.MethodPatch, fmt.Sprintf("%s/api/2.22/file-systems?ids=%s", base, fsID), map[string]any{
		"provisioned": 2147483648,
	})
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("PATCH provisioned: expected 200, got %d: %s", resp.StatusCode, body)
	}
	var patchResp struct {
		Items []struct {
			Provisioned int64 `json:"provisioned"`
		} `json:"items"`
	}
	decodeJSON(t, resp, &patchResp)
	if len(patchResp.Items) == 0 || patchResp.Items[0].Provisioned != 2147483648 {
		t.Errorf("expected provisioned 2147483648 after patch, got %d", func() int64 {
			if len(patchResp.Items) > 0 {
				return patchResp.Items[0].Provisioned
			}
			return 0
		}())
	}

	// Step 6: PATCH destroyed=true — soft-delete.
	resp = doJSON(t, http.MethodPatch, fmt.Sprintf("%s/api/2.22/file-systems?ids=%s", base, fsID), map[string]any{
		"destroyed": true,
	})
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("PATCH destroyed=true: expected 200, got %d: %s", resp.StatusCode, body)
	}
	var softDeleteResp struct {
		Items []struct {
			Destroyed bool `json:"destroyed"`
		} `json:"items"`
	}
	decodeJSON(t, resp, &softDeleteResp)
	if len(softDeleteResp.Items) == 0 || !softDeleteResp.Items[0].Destroyed {
		t.Error("expected destroyed=true after soft-delete patch")
	}

	// Step 7: DELETE /api/2.22/file-systems?ids={id} — eradicate.
	resp = doJSON(t, http.MethodDelete, fmt.Sprintf("%s/api/2.22/file-systems?ids=%s", base, fsID), nil)
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("DELETE: expected 200, got %d: %s", resp.StatusCode, body)
	}
	resp.Body.Close()

	// Step 8: GET after eradication — verify empty items.
	resp = doJSON(t, http.MethodGet, base+"/api/2.22/file-systems?names=test-fs", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET after eradicate: expected 200, got %d", resp.StatusCode)
	}
	var afterResp struct {
		Items []any `json:"items"`
	}
	decodeJSON(t, resp, &afterResp)
	if len(afterResp.Items) != 0 {
		t.Errorf("expected empty items after eradication, got %d items", len(afterResp.Items))
	}
}

func TestUnit_MockServer_LoginAndVersion(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	base := ms.URL()

	// POST /api/login
	resp := doJSON(t, http.MethodPost, base+"/api/login", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /login: expected 200, got %d", resp.StatusCode)
	}
	token := resp.Header.Get("x-auth-token")
	if token != "mock-session-token" {
		t.Errorf("expected x-auth-token mock-session-token, got %q", token)
	}
	resp.Body.Close()

	// GET /api/api_version
	resp = doJSON(t, http.MethodGet, base+"/api/api_version", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /api/api_version: expected 200, got %d", resp.StatusCode)
	}
	var vr struct {
		Versions []string `json:"versions"`
	}
	decodeJSON(t, resp, &vr)
	found := false
	for _, v := range vr.Versions {
		if v == "2.22" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 2.22 in versions, got %v", vr.Versions)
	}
}

func TestUnit_MockServer_DeleteRequiresDestroyed(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterFileSystemHandlers(ms.Mux)
	base := ms.URL()

	// Create a file system (not destroyed).
	// The FlashBlade API requires the name as a ?names= query parameter, not in the body.
	resp := doJSON(t, http.MethodPost, base+"/api/2.22/file-systems?names=no-destroy-fs", map[string]any{
		"provisioned": 1073741824,
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("create fs: expected 200, got %d", resp.StatusCode)
	}
	var createResp struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
	}
	decodeJSON(t, resp, &createResp)
	if len(createResp.Items) == 0 {
		t.Fatal("expected items in create response")
	}
	fsID := createResp.Items[0].ID

	// Attempt DELETE without soft-deleting first — should return 400.
	resp = doJSON(t, http.MethodDelete, fmt.Sprintf("%s/api/2.22/file-systems?ids=%s", base, fsID), nil)
	if resp.StatusCode != http.StatusBadRequest {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("DELETE non-destroyed: expected 400, got %d: %s", resp.StatusCode, body)
	}
	resp.Body.Close()
}
