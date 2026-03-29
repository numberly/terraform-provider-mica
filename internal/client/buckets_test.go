package client_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

func TestUnit_Bucket_List_Paginated(t *testing.T) {
	page1 := []client.Bucket{
		{ID: "bkt-p1-1", Name: "paginated-bkt-1", Account: client.NamedReference{Name: "acct1"}},
		{ID: "bkt-p1-2", Name: "paginated-bkt-2", Account: client.NamedReference{Name: "acct1"}},
	}
	page2 := []client.Bucket{
		{ID: "bkt-p2-1", Name: "paginated-bkt-3", Account: client.NamedReference{Name: "acct1"}},
	}

	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/buckets":
			callCount++
			token := r.URL.Query().Get("continuation_token")
			switch token {
			case "":
				writeJSON(w, http.StatusOK, map[string]any{
					"items":              page1,
					"total_item_count":   3,
					"continuation_token": "page2-token",
				})
			case "page2-token":
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
	items, err := c.ListBuckets(context.Background(), client.ListBucketsOpts{})
	if err != nil {
		t.Fatalf("ListBuckets: %v", err)
	}
	if len(items) != 3 {
		t.Errorf("expected 3 items across 2 pages, got %d", len(items))
	}
	if callCount != 2 {
		t.Errorf("expected 2 GET requests, got %d", callCount)
	}
	if items[0].Name != "paginated-bkt-1" {
		t.Errorf("expected first item paginated-bkt-1, got %q", items[0].Name)
	}
	if items[2].Name != "paginated-bkt-3" {
		t.Errorf("expected third item paginated-bkt-3, got %q", items[2].Name)
	}
}

func TestUnit_Bucket_List_SinglePage(t *testing.T) {
	bktList := []client.Bucket{
		{ID: "bkt-sp-1", Name: "single-page-bkt-1"},
		{ID: "bkt-sp-2", Name: "single-page-bkt-2"},
	}

	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/buckets":
			callCount++
			writeJSON(w, http.StatusOK, map[string]any{
				"items":            bktList,
				"total_item_count": 2,
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	items, err := c.ListBuckets(context.Background(), client.ListBucketsOpts{})
	if err != nil {
		t.Fatalf("ListBuckets: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
	if callCount != 1 {
		t.Errorf("expected exactly 1 GET request, got %d", callCount)
	}
}

func TestUnit_Bucket_List_Empty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/buckets":
			writeJSON(w, http.StatusOK, map[string]any{
				"items":            []client.Bucket{},
				"total_item_count": 0,
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	items, err := c.ListBuckets(context.Background(), client.ListBucketsOpts{})
	if err != nil {
		t.Fatalf("ListBuckets: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items, got %d", len(items))
	}
}
