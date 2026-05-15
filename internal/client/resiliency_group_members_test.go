package client_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

func TestUnit_ResiliencyGroupMember_Get_Found(t *testing.T) {
	rows := []client.ResiliencyGroupMember{
		{
			Group:  client.ResiliencyReference{ID: "rg-id-1", Name: "rg0", ResourceType: "resiliency-groups"},
			Member: client.ResiliencyReference{ID: "fs-id-1", Name: "fs-alpha", ResourceType: "file-systems"},
		},
		{
			Group:  client.ResiliencyReference{ID: "rg-id-1", Name: "rg0", ResourceType: "resiliency-groups"},
			Member: client.ResiliencyReference{ID: "fs-id-2", Name: "fs-beta", ResourceType: "file-systems"},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.23/resiliency-groups/members":
			if r.URL.Query().Get("resiliency_group_names") != "rg0" {
				writeJSON(w, http.StatusOK, listResponse([]client.ResiliencyGroupMember{}))
				return
			}
			writeJSON(w, http.StatusOK, listResponse(rows))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	m, err := c.GetResiliencyGroupMember(context.Background(), "rg0", "fs-beta")
	if err != nil {
		t.Fatalf("GetResiliencyGroupMember: %v", err)
	}
	if m.Group.Name != "rg0" {
		t.Errorf("expected Group.Name rg0, got %q", m.Group.Name)
	}
	if m.Member.Name != "fs-beta" {
		t.Errorf("expected Member.Name fs-beta, got %q", m.Member.Name)
	}
	if m.Member.ID != "fs-id-2" {
		t.Errorf("expected Member.ID fs-id-2, got %q", m.Member.ID)
	}
	if m.Member.ResourceType != "file-systems" {
		t.Errorf("expected Member.ResourceType file-systems, got %q", m.Member.ResourceType)
	}
}

func TestUnit_ResiliencyGroupMember_Get_NotFound(t *testing.T) {
	// Empty parent: API returns 200 + []; client must promote to 404.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.23/resiliency-groups/members":
			writeJSON(w, http.StatusOK, listResponse([]client.ResiliencyGroupMember{
				{
					Group:  client.ResiliencyReference{Name: "rg0"},
					Member: client.ResiliencyReference{Name: "fs-alpha"},
				},
			}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetResiliencyGroupMember(context.Background(), "rg0", "fs-missing")
	if err == nil {
		t.Fatal("expected not-found error, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound true; err: %v", err)
	}
}

func TestUnit_ResiliencyGroupMember_List(t *testing.T) {
	rows := []client.ResiliencyGroupMember{
		{
			Group:  client.ResiliencyReference{ID: "rg-id-1", Name: "rg0"},
			Member: client.ResiliencyReference{ID: "fs-1", Name: "fs-alpha", ResourceType: "file-systems"},
		},
		{
			Group:  client.ResiliencyReference{ID: "rg-id-1", Name: "rg0"},
			Member: client.ResiliencyReference{ID: "fs-2", Name: "fs-beta", ResourceType: "file-systems"},
		},
	}

	var seenParam string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.23/resiliency-groups/members":
			seenParam = r.URL.Query().Get("resiliency_group_names")
			writeJSON(w, http.StatusOK, listResponse(rows))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	items, err := c.ListResiliencyGroupMembers(context.Background(), "rg0")
	if err != nil {
		t.Fatalf("ListResiliencyGroupMembers: %v", err)
	}
	if seenParam != "rg0" {
		t.Errorf("expected resiliency_group_names=rg0, got %q", seenParam)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].Member.Name != "fs-alpha" || items[1].Member.Name != "fs-beta" {
		t.Errorf("unexpected member ordering: %+v", items)
	}
}

func TestUnit_ResiliencyGroupMember_List_EmptyParent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.23/resiliency-groups/members":
			writeJSON(w, http.StatusOK, listResponse([]client.ResiliencyGroupMember{}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	items, err := c.ListResiliencyGroupMembers(context.Background(), "rg-empty")
	if err != nil {
		t.Fatalf("ListResiliencyGroupMembers: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected empty list, got %d items", len(items))
	}
}
