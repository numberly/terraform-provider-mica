package client_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// newTestClient creates a FlashBladeClient pointing at a test server with a fake session token.
func newTestClient(t *testing.T, srv *httptest.Server) *client.FlashBladeClient {
	t.Helper()
	// We build a client manually using a minimal login mock.
	// The test server must handle /api/login.
	c, err := client.NewClient(context.Background(), client.Config{
		Endpoint:       srv.URL,
		APIToken:       "test-api-token",
		MaxRetries:     1,
	})
	if err != nil {
		t.Fatalf("newTestClient: %v", err)
	}
	return c
}

// writeJSON writes a JSON response to the ResponseWriter.
func writeJSON(w http.ResponseWriter, statusCode int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(v)
}

// listResponse wraps items in a ListResponse JSON envelope.
func listResponse(items any) map[string]any {
	return map[string]any{
		"items":            items,
		"total_item_count": 1,
	}
}

func TestUnit_FileSystem_Create(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/api/2.22/file-systems":
			name := r.URL.Query().Get("names")
			if name == "" {
				http.Error(w, "Names query parameter is missing", http.StatusBadRequest)
				return
			}
			var body client.FileSystemPost
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			fs := client.FileSystem{
				ID:          "fs-id-001",
				Name:        name,
				Provisioned: body.Provisioned,
				Created:     time.Now().UnixMilli(),
				Space:       client.Space{},
			}
			writeJSON(w, http.StatusOK, listResponse([]client.FileSystem{fs}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	fs, err := c.PostFileSystem(context.Background(), client.FileSystemPost{
		Name:        "test-fs",
		Provisioned: 1073741824,
	})
	if err != nil {
		t.Fatalf("PostFileSystem: %v", err)
	}
	if fs.ID != "fs-id-001" {
		t.Errorf("expected ID fs-id-001, got %q", fs.ID)
	}
	if fs.Name != "test-fs" {
		t.Errorf("expected Name test-fs, got %q", fs.Name)
	}
	if fs.Provisioned != 1073741824 {
		t.Errorf("expected Provisioned 1073741824, got %d", fs.Provisioned)
	}
}

func TestUnit_FileSystem_Read(t *testing.T) {
	expected := client.FileSystem{
		ID:          "fs-id-002",
		Name:        "read-fs",
		Provisioned: 2147483648,
		Created:     1700000000000,
		Space: client.Space{
			Unique:  100,
			Virtual: 200,
		},
		NFS: client.NFSConfig{Enabled: true},
		SMB: client.SMBConfig{Enabled: false},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/file-systems":
			name := r.URL.Query().Get("names")
			if name != "read-fs" {
				writeJSON(w, http.StatusOK, listResponse([]client.FileSystem{}))
				return
			}
			writeJSON(w, http.StatusOK, listResponse([]client.FileSystem{expected}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	fs, err := c.GetFileSystem(context.Background(), "read-fs")
	if err != nil {
		t.Fatalf("GetFileSystem: %v", err)
	}
	if fs.ID != expected.ID {
		t.Errorf("expected ID %q, got %q", expected.ID, fs.ID)
	}
	if fs.Created != expected.Created {
		t.Errorf("expected Created %d, got %d", expected.Created, fs.Created)
	}
	if !fs.NFS.Enabled {
		t.Errorf("expected NFS.Enabled true")
	}
}

func TestUnit_FileSystem_Read_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/file-systems":
			// Return empty items — FlashBlade behavior for not found
			writeJSON(w, http.StatusOK, listResponse([]client.FileSystem{}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetFileSystem(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound true, got false; err: %v", err)
	}
}

func TestUnit_FileSystem_Update(t *testing.T) {
	var gotBody client.FileSystemPatch

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPatch && r.URL.Path == "/api/2.22/file-systems":
			id := r.URL.Query().Get("ids")
			if id != "fs-id-003" {
				http.Error(w, "unexpected id", http.StatusBadRequest)
				return
			}
			if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			newProvisioned := int64(4294967296)
			fs := client.FileSystem{
				ID:          "fs-id-003",
				Name:        "update-fs",
				Provisioned: newProvisioned,
			}
			writeJSON(w, http.StatusOK, listResponse([]client.FileSystem{fs}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	newProv := int64(4294967296)
	fs, err := c.PatchFileSystem(context.Background(), "fs-id-003", client.FileSystemPatch{
		Provisioned: &newProv,
	})
	if err != nil {
		t.Fatalf("PatchFileSystem: %v", err)
	}
	if fs.Provisioned != 4294967296 {
		t.Errorf("expected Provisioned 4294967296, got %d", fs.Provisioned)
	}
	// Verify PATCH semantics: Name should be absent (nil) in body
	if gotBody.Name != nil {
		t.Errorf("expected Name to be absent in PATCH body, got %q", *gotBody.Name)
	}
}

func TestUnit_FileSystem_Update_Rename(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPatch && r.URL.Path == "/api/2.22/file-systems":
			var body client.FileSystemPatch
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			newName := "renamed-fs"
			if body.Name == nil || *body.Name != newName {
				http.Error(w, "expected name in patch", http.StatusBadRequest)
				return
			}
			fs := client.FileSystem{
				ID:   "fs-id-004",
				Name: *body.Name,
			}
			writeJSON(w, http.StatusOK, listResponse([]client.FileSystem{fs}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	newName := "renamed-fs"
	fs, err := c.PatchFileSystem(context.Background(), "fs-id-004", client.FileSystemPatch{
		Name: &newName,
	})
	if err != nil {
		t.Fatalf("PatchFileSystem rename: %v", err)
	}
	if fs.Name != "renamed-fs" {
		t.Errorf("expected Name renamed-fs, got %q", fs.Name)
	}
}

func TestUnit_FileSystem_SoftDelete(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPatch && r.URL.Path == "/api/2.22/file-systems":
			var body client.FileSystemPatch
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			if body.Destroyed == nil || !*body.Destroyed {
				http.Error(w, "expected destroyed=true", http.StatusBadRequest)
				return
			}
			destroyed := true
			fs := client.FileSystem{
				ID:        "fs-id-005",
				Name:      "soft-delete-fs",
				Destroyed: destroyed,
			}
			writeJSON(w, http.StatusOK, listResponse([]client.FileSystem{fs}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	destroyed := true
	fs, err := c.PatchFileSystem(context.Background(), "fs-id-005", client.FileSystemPatch{
		Destroyed: &destroyed,
	})
	if err != nil {
		t.Fatalf("PatchFileSystem soft-delete: %v", err)
	}
	if !fs.Destroyed {
		t.Errorf("expected Destroyed true after soft-delete")
	}
}

func TestUnit_FileSystem_Eradicate(t *testing.T) {
	var deleteCalled bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/2.22/file-systems":
			id := r.URL.Query().Get("ids")
			if id != "fs-id-006" {
				http.Error(w, "unexpected id", http.StatusBadRequest)
				return
			}
			deleteCalled = true
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	if err := c.DeleteFileSystem(context.Background(), "fs-id-006"); err != nil {
		t.Fatalf("DeleteFileSystem: %v", err)
	}
	if !deleteCalled {
		t.Error("expected DELETE to be called")
	}
}

func TestUnit_FileSystem_PollEradicated(t *testing.T) {
	callCount := 0

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/file-systems":
			callCount++
			destroyed := r.URL.Query().Get("destroyed")
			if destroyed != "true" {
				http.Error(w, "expected destroyed=true query param", http.StatusBadRequest)
				return
			}
			// First 2 calls: resource still present; 3rd call: empty (eradicated)
			if callCount < 3 {
				fs := client.FileSystem{ID: "fs-id-007", Name: "poll-fs", Destroyed: true}
				writeJSON(w, http.StatusOK, listResponse([]client.FileSystem{fs}))
			} else {
				writeJSON(w, http.StatusOK, map[string]any{"items": []client.FileSystem{}})
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	ctx := context.Background()
	if err := c.PollUntilEradicated(ctx, "poll-fs"); err != nil {
		t.Fatalf("PollUntilEradicated: %v", err)
	}
	if callCount < 3 {
		t.Errorf("expected at least 3 GET calls, got %d", callCount)
	}
}

func TestUnit_FileSystem_PollEradicated_Timeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/file-systems":
			// Always return the file system as present (never eradicated)
			fs := client.FileSystem{ID: "fs-id-008", Name: "timeout-fs", Destroyed: true}
			writeJSON(w, http.StatusOK, listResponse([]client.FileSystem{fs}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	// Very short deadline — should timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := c.PollUntilEradicated(ctx, "timeout-fs")
	if err == nil {
		t.Fatal("expected error due to context timeout, got nil")
	}
}

func TestUnit_FileSystem_List(t *testing.T) {
	fsList := []client.FileSystem{
		{ID: "fs-id-009", Name: "list-fs-1", Provisioned: 1073741824},
		{ID: "fs-id-010", Name: "list-fs-2", Provisioned: 2147483648},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/file-systems":
			writeJSON(w, http.StatusOK, map[string]any{
				"items":            fsList,
				"total_item_count": 2,
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	items, err := c.ListFileSystems(context.Background(), client.ListFileSystemsOpts{})
	if err != nil {
		t.Fatalf("ListFileSystems: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
	if items[0].Name != "list-fs-1" {
		t.Errorf("expected first item name list-fs-1, got %q", items[0].Name)
	}
}

// TestUnit_FileSystem_List_Paginated verifies that ListFileSystems auto-paginates:
// page 1 returns 2 items + continuation_token, page 2 returns 1 item + no token.
// ListFileSystems must return all 3 items combined.
func TestUnit_FileSystem_List_Paginated(t *testing.T) {
	page1 := []client.FileSystem{
		{ID: "fs-p1-1", Name: "paginated-fs-1", Provisioned: 1073741824},
		{ID: "fs-p1-2", Name: "paginated-fs-2", Provisioned: 2147483648},
	}
	page2 := []client.FileSystem{
		{ID: "fs-p2-1", Name: "paginated-fs-3", Provisioned: 4294967296},
	}

	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/file-systems":
			callCount++
			token := r.URL.Query().Get("continuation_token")
			switch token {
			case "":
				// First page: return 2 items + continuation_token.
				writeJSON(w, http.StatusOK, map[string]any{
					"items":              page1,
					"total_item_count":   3,
					"continuation_token": "page2-token",
				})
			case "page2-token":
				// Second page: return 1 item, no token.
				writeJSON(w, http.StatusOK, map[string]any{
					"items":            page2,
					"total_item_count": 3,
				})
			default:
				http.Error(w, "unexpected continuation_token", http.StatusBadRequest)
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	items, err := c.ListFileSystems(context.Background(), client.ListFileSystemsOpts{})
	if err != nil {
		t.Fatalf("ListFileSystems: %v", err)
	}
	if len(items) != 3 {
		t.Errorf("expected 3 items across 2 pages, got %d", len(items))
	}
	if callCount != 2 {
		t.Errorf("expected 2 GET requests (one per page), got %d", callCount)
	}
	if items[0].Name != "paginated-fs-1" {
		t.Errorf("expected first item paginated-fs-1, got %q", items[0].Name)
	}
	if items[2].Name != "paginated-fs-3" {
		t.Errorf("expected third item paginated-fs-3, got %q", items[2].Name)
	}
}

// TestUnit_FileSystem_List_SinglePage verifies that ListFileSystems does NOT make
// extra requests when no continuation_token is present in the response.
func TestUnit_FileSystem_List_SinglePage(t *testing.T) {
	fsList := []client.FileSystem{
		{ID: "fs-sp-1", Name: "single-page-fs-1"},
		{ID: "fs-sp-2", Name: "single-page-fs-2"},
	}

	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/file-systems":
			callCount++
			// No continuation_token in response — single page.
			writeJSON(w, http.StatusOK, map[string]any{
				"items":            fsList,
				"total_item_count": 2,
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	items, err := c.ListFileSystems(context.Background(), client.ListFileSystemsOpts{})
	if err != nil {
		t.Fatalf("ListFileSystems: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
	if callCount != 1 {
		t.Errorf("expected exactly 1 GET request, got %d", callCount)
	}
}

func TestUnit_FileSystem_List_WithFilter(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/file-systems":
			names := r.URL.Query().Get("names")
			if names != "specific-fs" {
				writeJSON(w, http.StatusOK, map[string]any{"items": []client.FileSystem{}})
				return
			}
			fs := client.FileSystem{ID: "fs-id-011", Name: "specific-fs", Provisioned: 1073741824}
			writeJSON(w, http.StatusOK, listResponse([]client.FileSystem{fs}))
		default:
			fmt.Fprintf(w, "not found")
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	items, err := c.ListFileSystems(context.Background(), client.ListFileSystemsOpts{
		Names: []string{"specific-fs"},
	})
	if err != nil {
		t.Fatalf("ListFileSystems with filter: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("expected 1 item, got %d", len(items))
	}
}
