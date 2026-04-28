package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

// newGenericHelperServer creates a test server with /api/login + configurable
// /api/2.22/test-resource endpoint for testing postOne/patchOne via PostTarget/PatchTarget.
func newGenericHelperServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("x-auth-token", "tok")
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/api/2.22/targets", handler)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

// TestUnit_PostOne_Success verifies postOne returns the first item from a list response.
func TestUnit_PostOne_Success(t *testing.T) {
	srv := newGenericHelperServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{
				{"id": "tgt-1", "name": "test-target", "address": "10.0.0.1", "status": "connected"},
			},
		})
	})

	c := newTestClient(t, srv)
	got, err := c.PostTarget(context.Background(), "test-target", client.TargetPost{Address: "10.0.0.1"})
	if err != nil {
		t.Fatalf("PostTarget: %v", err)
	}
	if got.ID != "tgt-1" {
		t.Errorf("expected ID %q, got %q", "tgt-1", got.ID)
	}
	if got.Name != "test-target" {
		t.Errorf("expected Name %q, got %q", "test-target", got.Name)
	}
}

// TestUnit_PostOne_EmptyResponse verifies postOne returns an error when the API
// returns an empty items list (exercising the "empty response from server" path).
func TestUnit_PostOne_EmptyResponse(t *testing.T) {
	srv := newGenericHelperServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"items": []any{}})
	})

	c := newTestClient(t, srv)
	_, err := c.PostTarget(context.Background(), "empty-target", client.TargetPost{Address: "10.0.0.1"})
	if err == nil {
		t.Fatal("expected error for empty response, got nil")
	}
	if got := err.Error(); got != "PostTarget: empty response from server" {
		t.Errorf("expected 'PostTarget: empty response from server', got %q", got)
	}
}

// TestUnit_PostOne_APIError verifies postOne propagates HTTP errors from the API.
func TestUnit_PostOne_APIError(t *testing.T) {
	srv := newGenericHelperServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errors": []map[string]any{{"message": "target already exists"}},
		})
	})

	c := newTestClient(t, srv)
	_, err := c.PostTarget(context.Background(), "dup-target", client.TargetPost{Address: "10.0.0.1"})
	if err == nil {
		t.Fatal("expected error for 409 conflict, got nil")
	}
}

// TestUnit_PatchOne_Success verifies patchOne returns the updated item.
func TestUnit_PatchOne_Success(t *testing.T) {
	srv := newGenericHelperServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{
				{"id": "tgt-1", "name": "test-target", "address": "10.0.0.2", "status": "connected"},
			},
		})
	})

	c := newTestClient(t, srv)
	newAddr := "10.0.0.2"
	got, err := c.PatchTarget(context.Background(), "test-target", client.TargetPatch{Address: &newAddr})
	if err != nil {
		t.Fatalf("PatchTarget: %v", err)
	}
	if got.Address != "10.0.0.2" {
		t.Errorf("expected Address %q, got %q", "10.0.0.2", got.Address)
	}
}

// TestUnit_PatchOne_EmptyResponse verifies patchOne returns an error when the API
// returns an empty items list.
func TestUnit_PatchOne_EmptyResponse(t *testing.T) {
	srv := newGenericHelperServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"items": []any{}})
	})

	c := newTestClient(t, srv)
	newAddr := "10.0.0.2"
	_, err := c.PatchTarget(context.Background(), "missing-target", client.TargetPatch{Address: &newAddr})
	if err == nil {
		t.Fatal("expected error for empty response, got nil")
	}
	if got := err.Error(); got != "PatchTarget: empty response from server" {
		t.Errorf("expected 'PatchTarget: empty response from server', got %q", got)
	}
}

// TestUnit_PatchOne_APIError verifies patchOne propagates HTTP errors from the API.
func TestUnit_PatchOne_APIError(t *testing.T) {
	srv := newGenericHelperServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errors": []map[string]any{{"message": "target not found"}},
		})
	})

	c := newTestClient(t, srv)
	newAddr := "10.0.0.2"
	_, err := c.PatchTarget(context.Background(), "gone-target", client.TargetPatch{Address: &newAddr})
	if err == nil {
		t.Fatal("expected error for 404, got nil")
	}
}
