package client_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

func TestUnit_NetworkAccessPolicy_List_Paginated(t *testing.T) {
	page1 := []client.NetworkAccessPolicy{
		{ID: "nap-1", Name: "default", Enabled: true},
		{ID: "nap-2", Name: "restricted", Enabled: true},
	}
	page2 := []client.NetworkAccessPolicy{
		{ID: "nap-3", Name: "open", Enabled: false},
	}

	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/network-access-policies":
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
	items, err := c.ListNetworkAccessPolicies(context.Background())
	if err != nil {
		t.Fatalf("ListNetworkAccessPolicies: %v", err)
	}
	if len(items) != 3 {
		t.Errorf("expected 3 items across 2 pages, got %d", len(items))
	}
	if callCount != 2 {
		t.Errorf("expected 2 GET requests, got %d", callCount)
	}
	if items[0].Name != "default" {
		t.Errorf("expected first item default, got %q", items[0].Name)
	}
}

func TestUnit_NetworkAccessPolicy_List_SinglePage(t *testing.T) {
	policies := []client.NetworkAccessPolicy{
		{ID: "nap-sp-1", Name: "policy-1", Enabled: true},
		{ID: "nap-sp-2", Name: "policy-2", Enabled: false},
	}

	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/network-access-policies":
			callCount++
			writeJSON(w, http.StatusOK, map[string]any{
				"items":            policies,
				"total_item_count": 2,
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	items, err := c.ListNetworkAccessPolicies(context.Background())
	if err != nil {
		t.Fatalf("ListNetworkAccessPolicies: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
	if callCount != 1 {
		t.Errorf("expected exactly 1 GET request, got %d", callCount)
	}
}
