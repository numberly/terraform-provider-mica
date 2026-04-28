package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

func TestUnit_FileSystemExport_Get(t *testing.T) {
	expected := client.FileSystemExport{
		ID:         "fse-id-001",
		Name:       "my-fs/my-export",
		ExportName: "my-export",
		Enabled:    true,
		Member:     &client.NamedReference{Name: "my-fs"},
		Server:     &client.NamedReference{Name: "array-server-1"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/file-system-exports":
			name := r.URL.Query().Get("names")
			if name != "my-fs/my-export" {
				writeJSON(w, http.StatusOK, listResponse([]client.FileSystemExport{}))
				return
			}
			writeJSON(w, http.StatusOK, listResponse([]client.FileSystemExport{expected}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.GetFileSystemExport(context.Background(), "my-fs/my-export")
	if err != nil {
		t.Fatalf("GetFileSystemExport: %v", err)
	}
	if got.ID != expected.ID {
		t.Errorf("expected ID %q, got %q", expected.ID, got.ID)
	}
	if got.ExportName != "my-export" {
		t.Errorf("expected ExportName my-export, got %q", got.ExportName)
	}
	if !got.Enabled {
		t.Errorf("expected Enabled true")
	}
	if got.Member == nil || got.Member.Name != "my-fs" {
		t.Errorf("expected Member.Name my-fs, got %v", got.Member)
	}
}

func TestUnit_FileSystemExport_Get_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/file-system-exports":
			writeJSON(w, http.StatusOK, listResponse([]client.FileSystemExport{}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetFileSystemExport(context.Background(), "nonexistent/export")
	if err == nil {
		t.Fatal("expected error for not found, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound true, got false; err: %v", err)
	}
}

func TestUnit_FileSystemExport_Post(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/api/2.22/file-system-exports":
			// POST uses ?member_names= and ?policy_names=
			memberName := r.URL.Query().Get("member_names")
			policyName := r.URL.Query().Get("policy_names")
			if memberName == "" || policyName == "" {
				http.Error(w, "member_names and policy_names required", http.StatusBadRequest)
				return
			}
			var body client.FileSystemExportPost
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			export := client.FileSystemExport{
				ID:         "fse-id-002",
				Name:       memberName + "/" + body.ExportName,
				ExportName: body.ExportName,
				Enabled:    true,
				Member:     &client.NamedReference{Name: memberName},
				Policy:     &client.NamedReference{Name: policyName},
			}
			writeJSON(w, http.StatusOK, listResponse([]client.FileSystemExport{export}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.PostFileSystemExport(context.Background(), "my-fs", "nfs-export-policy", client.FileSystemExportPost{
		ExportName: "new-export",
		Server:     &client.NamedReference{Name: "array-server-1"},
	})
	if err != nil {
		t.Fatalf("PostFileSystemExport: %v", err)
	}
	if got.ID != "fse-id-002" {
		t.Errorf("expected ID fse-id-002, got %q", got.ID)
	}
	if got.ExportName != "new-export" {
		t.Errorf("expected ExportName new-export, got %q", got.ExportName)
	}
	if got.Member == nil || got.Member.Name != "my-fs" {
		t.Errorf("expected Member.Name my-fs, got %v", got.Member)
	}
}

func TestUnit_FileSystemExport_Patch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPatch && r.URL.Path == "/api/2.22/file-system-exports":
			// PATCH uses ?ids=
			id := r.URL.Query().Get("ids")
			if id != "fse-id-003" {
				http.Error(w, "unexpected ids param", http.StatusBadRequest)
				return
			}
			var body client.FileSystemExportPatch
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			exportName := "original-export"
			if body.ExportName != nil {
				exportName = *body.ExportName
			}
			export := client.FileSystemExport{
				ID:         id,
				Name:       "my-fs/" + exportName,
				ExportName: exportName,
				Enabled:    true,
			}
			writeJSON(w, http.StatusOK, listResponse([]client.FileSystemExport{export}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	newExportName := "renamed-export"
	got, err := c.PatchFileSystemExport(context.Background(), "fse-id-003", client.FileSystemExportPatch{
		ExportName: &newExportName,
	})
	if err != nil {
		t.Fatalf("PatchFileSystemExport: %v", err)
	}
	if got.ExportName != "renamed-export" {
		t.Errorf("expected ExportName renamed-export, got %q", got.ExportName)
	}
}

func TestUnit_FileSystemExport_Delete(t *testing.T) {
	var deleteCalled bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/2.22/file-system-exports":
			// DELETE uses ?member_names= and ?names=
			memberName := r.URL.Query().Get("member_names")
			exportName := r.URL.Query().Get("names")
			if memberName != "my-fs" || exportName != "my-export" {
				http.Error(w, "unexpected params", http.StatusBadRequest)
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
	if err := c.DeleteFileSystemExport(context.Background(), "my-fs", "my-export"); err != nil {
		t.Fatalf("DeleteFileSystemExport: %v", err)
	}
	if !deleteCalled {
		t.Error("expected DELETE to be called")
	}
}
