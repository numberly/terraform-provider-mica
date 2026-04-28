package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

// decodeJSON decodes the request body into v or fails the test.
func decodeJSON(t *testing.T, r *http.Request, v any) {
	t.Helper()
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		t.Fatalf("decode request body: %v", err)
	}
}

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

func TestUnit_Bucket_Get_Found(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/buckets":
			if r.URL.Query().Get("names") != "my-bucket" {
				t.Errorf("expected names=my-bucket, got %q", r.URL.Query().Get("names"))
			}
			writeJSON(w, http.StatusOK, map[string]any{
				"items": []client.Bucket{
					{ID: "bkt-1", Name: "my-bucket", Account: client.NamedReference{Name: "acct"}},
				},
				"total_item_count": 1,
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	b, err := c.GetBucket(context.Background(), "my-bucket")
	if err != nil {
		t.Fatalf("GetBucket: %v", err)
	}
	if b.ID != "bkt-1" || b.Name != "my-bucket" {
		t.Errorf("unexpected bucket %+v", b)
	}
}

func TestUnit_Bucket_Get_NotFound(t *testing.T) {
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
	_, err := c.GetBucket(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error for missing bucket, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound error, got %v", err)
	}
}

func TestUnit_Bucket_Post(t *testing.T) {
	var gotBody client.BucketPost
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/api/2.22/buckets":
			if r.URL.Query().Get("names") != "new-bucket" {
				t.Errorf("expected names=new-bucket, got %q", r.URL.Query().Get("names"))
			}
			decodeJSON(t, r, &gotBody)
			writeJSON(w, http.StatusOK, map[string]any{
				"items": []client.Bucket{
					{ID: "bkt-new", Name: "new-bucket", Account: client.NamedReference{Name: gotBody.Account.Name}},
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	b, err := c.PostBucket(context.Background(), "new-bucket", client.BucketPost{
		Account: client.NamedReference{Name: "acct"},
	})
	if err != nil {
		t.Fatalf("PostBucket: %v", err)
	}
	if b.ID != "bkt-new" {
		t.Errorf("expected bkt-new, got %q", b.ID)
	}
	if gotBody.Account.Name != "acct" {
		t.Errorf("expected account=acct in body, got %q", gotBody.Account.Name)
	}
}

func TestUnit_Bucket_Patch_Destroyed(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPatch && r.URL.Path == "/api/2.22/buckets":
			if r.URL.Query().Get("ids") != "bkt-1" {
				t.Errorf("expected ids=bkt-1, got %q", r.URL.Query().Get("ids"))
			}
			decodeJSON(t, r, &gotBody)
			writeJSON(w, http.StatusOK, map[string]any{
				"items": []client.Bucket{{ID: "bkt-1", Name: "b", Destroyed: true}},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	destroyed := true
	b, err := c.PatchBucket(context.Background(), "bkt-1", client.BucketPatch{Destroyed: &destroyed})
	if err != nil {
		t.Fatalf("PatchBucket: %v", err)
	}
	if !b.Destroyed {
		t.Errorf("expected destroyed=true, got %+v", b)
	}
	if v, ok := gotBody["destroyed"].(bool); !ok || !v {
		t.Errorf("expected destroyed=true in PATCH body, got %v", gotBody)
	}
}

func TestUnit_Bucket_Delete(t *testing.T) {
	var hitIDs string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/2.22/buckets":
			hitIDs = r.URL.Query().Get("ids")
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	if err := c.DeleteBucket(context.Background(), "bkt-1"); err != nil {
		t.Fatalf("DeleteBucket: %v", err)
	}
	if hitIDs != "bkt-1" {
		t.Errorf("expected DELETE with ids=bkt-1, got %q", hitIDs)
	}
}

func TestUnit_Bucket_PollUntilEradicated(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/buckets":
			calls++
			if calls < 2 {
				writeJSON(w, http.StatusOK, map[string]any{
					"items": []client.Bucket{{ID: "bkt-1", Name: "b", Destroyed: true}},
				})
			} else {
				writeJSON(w, http.StatusOK, map[string]any{"items": []client.Bucket{}})
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := c.PollBucketUntilEradicated(ctx, "b"); err != nil {
		t.Fatalf("PollBucketUntilEradicated: %v", err)
	}
	if calls < 2 {
		t.Errorf("expected at least 2 GET calls, got %d", calls)
	}
}
