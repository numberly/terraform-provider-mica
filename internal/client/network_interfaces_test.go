package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

func TestUnit_NetworkInterface_Create(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/api/2.22/network-interfaces":
			name := r.URL.Query().Get("names")
			if name == "" {
				http.Error(w, "names required", http.StatusBadRequest)
				return
			}
			var body client.NetworkInterfacePost
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			ni := client.NetworkInterface{
				ID:              "ni-id-001",
				Name:            name,
				Address:         body.Address,
				Type:            body.Type,
				Services:        body.Services,
				Subnet:          body.Subnet,
				AttachedServers: body.AttachedServers,
				Enabled:         true,
				Gateway:         "10.99.99.1",
				Netmask:         "255.255.255.0",
				MTU:             1500,
				VLAN:            999,
			}
			writeJSON(w, http.StatusOK, listResponse([]client.NetworkInterface{ni}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	ni, err := c.PostNetworkInterface(context.Background(), "test-data-vip", client.NetworkInterfacePost{
		Address:         "10.99.99.10",
		Type:            "vip",
		Services:        []string{"data"},
		Subnet:          &client.NamedReference{Name: "test-subnet"},
		AttachedServers: []client.NamedReference{{Name: "srv-test"}},
	})
	if err != nil {
		t.Fatalf("PostNetworkInterface: %v", err)
	}
	if ni.Name != "test-data-vip" {
		t.Errorf("expected Name test-data-vip, got %q", ni.Name)
	}
	if ni.Address != "10.99.99.10" {
		t.Errorf("expected Address 10.99.99.10, got %q", ni.Address)
	}
	if ni.Gateway != "10.99.99.1" {
		t.Errorf("expected Gateway derived from subnet, got %q", ni.Gateway)
	}
	if !ni.Enabled {
		t.Error("expected Enabled true")
	}
}

func TestUnit_NetworkInterface_Read(t *testing.T) {
	expected := client.NetworkInterface{
		ID:       "ni-id-002",
		Name:     "read-vip",
		Address:  "10.0.0.10",
		Type:     "vip",
		Services: []string{"sts"},
		Subnet:   &client.NamedReference{Name: "mgmt"},
		Enabled:  true,
		Gateway:  "10.0.0.1",
		Netmask:  "255.255.255.0",
		MTU:      1500,
		VLAN:     100,
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/network-interfaces":
			name := r.URL.Query().Get("names")
			if name != "read-vip" {
				writeJSON(w, http.StatusOK, listResponse([]client.NetworkInterface{}))
				return
			}
			writeJSON(w, http.StatusOK, listResponse([]client.NetworkInterface{expected}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	ni, err := c.GetNetworkInterface(context.Background(), "read-vip")
	if err != nil {
		t.Fatalf("GetNetworkInterface: %v", err)
	}
	if ni.ID != expected.ID {
		t.Errorf("expected ID %q, got %q", expected.ID, ni.ID)
	}
	if len(ni.Services) != 1 || ni.Services[0] != "sts" {
		t.Errorf("expected Services [sts], got %v", ni.Services)
	}
}

func TestUnit_NetworkInterface_Read_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/network-interfaces":
			writeJSON(w, http.StatusOK, listResponse([]client.NetworkInterface{}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetNetworkInterface(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound true; err: %v", err)
	}
}

func TestUnit_NetworkInterface_Update(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPatch && r.URL.Path == "/api/2.22/network-interfaces":
			name := r.URL.Query().Get("names")
			if name != "patch-vip" {
				http.Error(w, "unexpected name", http.StatusBadRequest)
				return
			}
			var body client.NetworkInterfacePatch
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			ni := client.NetworkInterface{
				ID:   "ni-id-003",
				Name: "patch-vip",
				Type: "vip",
			}
			if body.Address != nil {
				ni.Address = *body.Address
			}
			ni.Services = body.Services
			ni.AttachedServers = body.AttachedServers
			writeJSON(w, http.StatusOK, listResponse([]client.NetworkInterface{ni}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	newAddr := "10.99.99.20"
	ni, err := c.PatchNetworkInterface(context.Background(), "patch-vip", client.NetworkInterfacePatch{
		Address:         &newAddr,
		Services:        []string{"data"},
		AttachedServers: []client.NamedReference{{Name: "new-server"}},
	})
	if err != nil {
		t.Fatalf("PatchNetworkInterface: %v", err)
	}
	if ni.Address != "10.99.99.20" {
		t.Errorf("expected Address 10.99.99.20, got %q", ni.Address)
	}
}

func TestUnit_NetworkInterface_Delete(t *testing.T) {
	var deleteCalled bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/2.22/network-interfaces":
			name := r.URL.Query().Get("names")
			if name != "delete-vip" {
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
	if err := c.DeleteNetworkInterface(context.Background(), "delete-vip"); err != nil {
		t.Fatalf("DeleteNetworkInterface: %v", err)
	}
	if !deleteCalled {
		t.Error("expected DELETE to be called")
	}
}

func TestUnit_NetworkInterface_List(t *testing.T) {
	nis := []client.NetworkInterface{
		{ID: "ni1", Name: "vip1", Address: "10.0.0.10", Services: []string{"data"}},
		{ID: "ni2", Name: "vip2", Address: "10.0.0.11", Services: []string{"sts"}},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/network-interfaces":
			writeJSON(w, http.StatusOK, map[string]any{
				"items":            nis,
				"total_item_count": 2,
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	items, err := c.ListNetworkInterfaces(context.Background())
	if err != nil {
		t.Fatalf("ListNetworkInterfaces: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
}
