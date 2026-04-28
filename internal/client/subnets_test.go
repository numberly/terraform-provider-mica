package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

func TestUnit_Subnet_Create(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/api/2.22/subnets":
			name := r.URL.Query().Get("names")
			if name == "" {
				http.Error(w, "names query parameter is required", http.StatusBadRequest)
				return
			}
			var body client.SubnetPost
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			sub := client.Subnet{
				ID:      "subnet-id-001",
				Name:    name,
				Prefix:  body.Prefix,
				Gateway: body.Gateway,
				MTU:     body.MTU,
				Enabled: true,
			}
			if body.VLAN != nil {
				sub.VLAN = *body.VLAN
			}
			if body.LinkAggregationGroup != nil {
				sub.LinkAggregationGroup = body.LinkAggregationGroup
			}
			writeJSON(w, http.StatusOK, listResponse([]client.Subnet{sub}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	vlan999 := int64(999)
	sub, err := c.PostSubnet(context.Background(), "test-subnet", client.SubnetPost{
		Prefix:  "10.99.99.0/24",
		Gateway: "10.99.99.1",
		MTU:     1500,
		VLAN:    &vlan999,
		LinkAggregationGroup: &client.NamedReference{Name: "uplink"},
	})
	if err != nil {
		t.Fatalf("PostSubnet: %v", err)
	}
	if sub.Name != "test-subnet" {
		t.Errorf("expected Name test-subnet, got %q", sub.Name)
	}
	if sub.Prefix != "10.99.99.0/24" {
		t.Errorf("expected Prefix 10.99.99.0/24, got %q", sub.Prefix)
	}
	if sub.VLAN != 999 {
		t.Errorf("expected VLAN 999, got %d", sub.VLAN)
	}
}

func TestUnit_Subnet_Read(t *testing.T) {
	expected := client.Subnet{
		ID:      "subnet-id-002",
		Name:    "read-subnet",
		Prefix:  "10.0.0.0/24",
		Gateway: "10.0.0.1",
		MTU:     9000,
		VLAN:    100,
		Enabled: true,
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/subnets":
			name := r.URL.Query().Get("names")
			if name != "read-subnet" {
				writeJSON(w, http.StatusOK, listResponse([]client.Subnet{}))
				return
			}
			writeJSON(w, http.StatusOK, listResponse([]client.Subnet{expected}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	sub, err := c.GetSubnet(context.Background(), "read-subnet")
	if err != nil {
		t.Fatalf("GetSubnet: %v", err)
	}
	if sub.ID != expected.ID {
		t.Errorf("expected ID %q, got %q", expected.ID, sub.ID)
	}
	if sub.MTU != 9000 {
		t.Errorf("expected MTU 9000, got %d", sub.MTU)
	}
}

func TestUnit_Subnet_Read_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/subnets":
			writeJSON(w, http.StatusOK, listResponse([]client.Subnet{}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetSubnet(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound true, got false; err: %v", err)
	}
}

func TestUnit_Subnet_Update(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPatch && r.URL.Path == "/api/2.22/subnets":
			name := r.URL.Query().Get("names")
			if name != "patch-subnet" {
				http.Error(w, "unexpected name", http.StatusBadRequest)
				return
			}
			var body client.SubnetPatch
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			sub := client.Subnet{
				ID:   "subnet-id-003",
				Name: "patch-subnet",
			}
			if body.Gateway != nil {
				sub.Gateway = *body.Gateway
			}
			if body.MTU != nil {
				sub.MTU = *body.MTU
			}
			writeJSON(w, http.StatusOK, listResponse([]client.Subnet{sub}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	newGw := "10.0.0.254"
	newMTU := int64(9000)
	sub, err := c.PatchSubnet(context.Background(), "patch-subnet", client.SubnetPatch{
		Gateway: &newGw,
		MTU:     &newMTU,
	})
	if err != nil {
		t.Fatalf("PatchSubnet: %v", err)
	}
	if sub.Gateway != "10.0.0.254" {
		t.Errorf("expected Gateway 10.0.0.254, got %q", sub.Gateway)
	}
}

func TestUnit_Subnet_Delete(t *testing.T) {
	var deleteCalled bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/2.22/subnets":
			name := r.URL.Query().Get("names")
			if name != "delete-subnet" {
				http.Error(w, "unexpected name", http.StatusBadRequest)
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
	if err := c.DeleteSubnet(context.Background(), "delete-subnet"); err != nil {
		t.Fatalf("DeleteSubnet: %v", err)
	}
	if !deleteCalled {
		t.Error("expected DELETE to be called")
	}
}

func TestUnit_Subnet_List(t *testing.T) {
	subnets := []client.Subnet{
		{ID: "s1", Name: "sub1", Prefix: "10.0.0.0/24"},
		{ID: "s2", Name: "sub2", Prefix: "10.1.0.0/24"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/subnets":
			writeJSON(w, http.StatusOK, map[string]any{
				"items":            subnets,
				"total_item_count": 2,
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	items, err := c.ListSubnets(context.Background())
	if err != nil {
		t.Fatalf("ListSubnets: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
}
