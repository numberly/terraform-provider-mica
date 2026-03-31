package client_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

func TestUnit_LAG_Read(t *testing.T) {
	expected := client.LinkAggregationGroup{
		ID:        "lag-id-001",
		Name:      "uplink",
		Status:    "healthy",
		Ports:     []string{"CH1.ETH1.1", "CH1.ETH1.2"},
		PortSpeed: 25000000000,
		LagSpeed:  50000000000,
		MacAddress:   "24:a9:37:f5:da:64",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/link-aggregation-groups":
			name := r.URL.Query().Get("names")
			if name != "uplink" {
				writeJSON(w, http.StatusOK, listResponse([]client.LinkAggregationGroup{}))
				return
			}
			writeJSON(w, http.StatusOK, listResponse([]client.LinkAggregationGroup{expected}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	lag, err := c.GetLinkAggregationGroup(context.Background(), "uplink")
	if err != nil {
		t.Fatalf("GetLinkAggregationGroup: %v", err)
	}
	if lag.Name != "uplink" {
		t.Errorf("expected Name uplink, got %q", lag.Name)
	}
	if lag.Status != "healthy" {
		t.Errorf("expected Status healthy, got %q", lag.Status)
	}
	if len(lag.Ports) != 2 {
		t.Errorf("expected 2 ports, got %d", len(lag.Ports))
	}
	if lag.MacAddress != "24:a9:37:f5:da:64" {
		t.Errorf("expected MacAddress 24:a9:37:f5:da:64, got %q", lag.MacAddress)
	}
}

func TestUnit_LAG_Read_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/link-aggregation-groups":
			writeJSON(w, http.StatusOK, listResponse([]client.LinkAggregationGroup{}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetLinkAggregationGroup(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound true; err: %v", err)
	}
}

func TestUnit_LAG_List(t *testing.T) {
	lags := []client.LinkAggregationGroup{
		{ID: "lag1", Name: "uplink", Status: "healthy"},
		{ID: "lag2", Name: "replication", Status: "healthy"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/link-aggregation-groups":
			writeJSON(w, http.StatusOK, map[string]any{
				"items":            lags,
				"total_item_count": 2,
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	items, err := c.ListLinkAggregationGroups(context.Background())
	if err != nil {
		t.Fatalf("ListLinkAggregationGroups: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
}
