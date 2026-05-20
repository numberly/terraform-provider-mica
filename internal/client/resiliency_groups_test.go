package client_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

func TestUnit_ResiliencyGroup_Get_Found(t *testing.T) {
	expected := client.ResiliencyGroup{
		ID:            "rg-id-001",
		Name:          "rg0",
		Status:        "healthy",
		StatusDetails: "",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.23/resiliency-groups":
			name := r.URL.Query().Get("names")
			if name != "rg0" {
				writeJSON(w, http.StatusOK, listResponse([]client.ResiliencyGroup{}))
				return
			}
			writeJSON(w, http.StatusOK, listResponse([]client.ResiliencyGroup{expected}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	rg, err := c.GetResiliencyGroup(context.Background(), "rg0")
	if err != nil {
		t.Fatalf("GetResiliencyGroup: %v", err)
	}
	if rg.Name != "rg0" {
		t.Errorf("expected Name rg0, got %q", rg.Name)
	}
	if rg.ID != "rg-id-001" {
		t.Errorf("expected ID rg-id-001, got %q", rg.ID)
	}
	if rg.Status != "healthy" {
		t.Errorf("expected Status healthy, got %q", rg.Status)
	}
}

func TestUnit_ResiliencyGroup_Get_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.23/resiliency-groups":
			writeJSON(w, http.StatusOK, listResponse([]client.ResiliencyGroup{}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetResiliencyGroup(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound true; err: %v", err)
	}
}

func TestUnit_ResiliencyGroup_List(t *testing.T) {
	groups := []client.ResiliencyGroup{
		{ID: "rg1", Name: "rg0", Status: "healthy"},
		{ID: "rg2", Name: "rg1", Status: "unhealthy", StatusDetails: "blade missing"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.23/resiliency-groups":
			writeJSON(w, http.StatusOK, map[string]any{
				"items":            groups,
				"total_item_count": 2,
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	items, err := c.ListResiliencyGroups(context.Background())
	if err != nil {
		t.Fatalf("ListResiliencyGroups: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
	if items[1].StatusDetails != "blade missing" {
		t.Errorf("expected StatusDetails 'blade missing', got %q", items[1].StatusDetails)
	}
}
