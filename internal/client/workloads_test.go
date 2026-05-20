package client_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/numberly/terraform-provider-mica/internal/client"
	"github.com/numberly/terraform-provider-mica/internal/testmock/handlers"
)

func newWorkloadServer(t *testing.T) (*httptest.Server, *workloadStoreFacade) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("x-auth-token", "tok")
		w.WriteHeader(http.StatusOK)
	})
	store := handlers.RegisterWorkloadHandlers(mux)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv, &workloadStoreFacade{store: store}
}

// workloadStoreFacade wraps the opaque store so tests can call Seed.
type workloadStoreFacade struct {
	store interface {
		Seed(w *client.Workload)
	}
}

func TestUnit_Workload_Get_Found(t *testing.T) {
	srv, facade := newWorkloadServer(t)
	facade.store.Seed(&client.Workload{
		ID:     "wl-seed-1",
		Name:   "my-workload",
		Status: "ready",
		Preset: &client.WorkloadPreset{
			ID:   "preset-1",
			Name: "my-preset",
		},
	})

	c := newTestClient(t, srv)
	got, err := c.GetWorkload(context.Background(), "my-workload")
	if err != nil {
		t.Fatalf("GetWorkload: %v", err)
	}
	if got.ID != "wl-seed-1" {
		t.Errorf("expected ID %q, got %q", "wl-seed-1", got.ID)
	}
	if got.Name != "my-workload" {
		t.Errorf("expected Name %q, got %q", "my-workload", got.Name)
	}
	if got.Status != "ready" {
		t.Errorf("expected Status %q, got %q", "ready", got.Status)
	}
	if got.Preset == nil {
		t.Fatal("expected Preset to be set, got nil")
	}
	if got.Preset.Name != "my-preset" {
		t.Errorf("expected Preset.Name %q, got %q", "my-preset", got.Preset.Name)
	}
}

func TestUnit_Workload_Get_NotFound(t *testing.T) {
	srv, _ := newWorkloadServer(t)
	c := newTestClient(t, srv)

	_, err := c.GetWorkload(context.Background(), "nonexistent-workload")
	if err == nil {
		t.Fatal("expected error for unknown workload, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound true, got false; err: %v", err)
	}
}

func TestUnit_Workload_Post(t *testing.T) {
	srv, _ := newWorkloadServer(t)
	c := newTestClient(t, srv)

	str := "value-a"
	got, err := c.PostWorkload(context.Background(), "new-workload", "my-preset", client.WorkloadPost{
		Parameters: []client.WorkloadParameter{
			{
				Name: "param1",
				Value: client.WorkloadParameterValue{
					String: &str,
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("PostWorkload: %v", err)
	}
	if got.Name != "new-workload" {
		t.Errorf("expected Name %q, got %q", "new-workload", got.Name)
	}
	if got.ID == "" {
		t.Error("expected non-empty ID after POST")
	}
	if got.Status != "ready" {
		t.Errorf("expected Status %q, got %q", "ready", got.Status)
	}
	if got.Preset == nil {
		t.Fatal("expected Preset to be set, got nil")
	}
	if got.Preset.Name != "my-preset" {
		t.Errorf("expected Preset.Name %q, got %q", "my-preset", got.Preset.Name)
	}
}

func TestUnit_Workload_Patch_Destroyed(t *testing.T) {
	srv, facade := newWorkloadServer(t)
	facade.store.Seed(&client.Workload{
		ID:     "wl-patch-1",
		Name:   "patch-workload",
		Status: "ready",
	})

	c := newTestClient(t, srv)
	destroyed := true
	got, err := c.PatchWorkload(context.Background(), "patch-workload", client.WorkloadPatch{
		Destroyed: &destroyed,
	})
	if err != nil {
		t.Fatalf("PatchWorkload destroyed: %v", err)
	}
	if !got.Destroyed {
		t.Error("expected Destroyed=true after patch")
	}
	if got.Status != "destroying" {
		t.Errorf("expected Status %q, got %q", "destroying", got.Status)
	}
	if got.TimeRemaining == 0 {
		t.Error("expected non-zero TimeRemaining when destroyed")
	}
}

func TestUnit_Workload_Delete(t *testing.T) {
	srv, facade := newWorkloadServer(t)
	facade.store.Seed(&client.Workload{
		ID:     "wl-del-1",
		Name:   "del-workload",
		Status: "ready",
	})

	c := newTestClient(t, srv)
	if err := c.DeleteWorkload(context.Background(), "del-workload"); err != nil {
		t.Fatalf("DeleteWorkload: %v", err)
	}

	// Subsequent GET must return IsNotFound.
	_, err := c.GetWorkload(context.Background(), "del-workload")
	if err == nil {
		t.Fatal("expected error after delete, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound true after delete, got false; err: %v", err)
	}
}
