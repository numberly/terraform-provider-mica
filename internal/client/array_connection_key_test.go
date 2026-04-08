package client_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock/handlers"
)

func newArrayConnectionKeyServer(t *testing.T) (*httptest.Server, *arrayConnectionKeyStoreFacade) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("x-auth-token", "tok")
		w.WriteHeader(http.StatusOK)
	})
	store := handlers.RegisterArrayConnectionKeyHandlers(mux)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv, &arrayConnectionKeyStoreFacade{store: store}
}

// arrayConnectionKeyStoreFacade wraps the opaque store so tests can call Seed.
type arrayConnectionKeyStoreFacade struct {
	store interface {
		Seed(key *client.ArrayConnectionKey)
	}
}

func TestUnit_ArrayConnectionKey_Get(t *testing.T) {
	srv, facade := newArrayConnectionKeyServer(t)
	facade.store.Seed(&client.ArrayConnectionKey{
		ConnectionKey: "seeded-key-abc",
		Created:       1000000000000,
		Expires:       1000003600000,
	})

	c := newTestClient(t, srv)
	got, err := c.GetArrayConnectionKey(context.Background())
	if err != nil {
		t.Fatalf("GetArrayConnectionKey: %v", err)
	}
	if got.ConnectionKey != "seeded-key-abc" {
		t.Errorf("expected ConnectionKey %q, got %q", "seeded-key-abc", got.ConnectionKey)
	}
	if got.Created != 1000000000000 {
		t.Errorf("expected Created %d, got %d", 1000000000000, got.Created)
	}
	if got.Expires != 1000003600000 {
		t.Errorf("expected Expires %d, got %d", 1000003600000, got.Expires)
	}
}

func TestUnit_ArrayConnectionKey_Post(t *testing.T) {
	srv, _ := newArrayConnectionKeyServer(t)
	c := newTestClient(t, srv)

	got, err := c.PostArrayConnectionKey(context.Background())
	if err != nil {
		t.Fatalf("PostArrayConnectionKey: %v", err)
	}
	if got.ConnectionKey == "" {
		t.Error("expected non-empty ConnectionKey after POST")
	}
	if got.Created == 0 {
		t.Error("expected non-zero Created after POST")
	}
	if got.Expires == 0 {
		t.Error("expected non-zero Expires after POST")
	}
}

func TestUnit_ArrayConnectionKey_Get_AfterPost(t *testing.T) {
	srv, _ := newArrayConnectionKeyServer(t)
	c := newTestClient(t, srv)

	// POST generates the key.
	posted, err := c.PostArrayConnectionKey(context.Background())
	if err != nil {
		t.Fatalf("PostArrayConnectionKey: %v", err)
	}

	// GET must return the same key.
	got, err := c.GetArrayConnectionKey(context.Background())
	if err != nil {
		t.Fatalf("GetArrayConnectionKey after POST: %v", err)
	}
	if got.ConnectionKey != posted.ConnectionKey {
		t.Errorf("GET returned %q but POST returned %q", got.ConnectionKey, posted.ConnectionKey)
	}
	if got.Created != posted.Created {
		t.Errorf("GET Created=%d but POST Created=%d", got.Created, posted.Created)
	}
	if got.Expires != posted.Expires {
		t.Errorf("GET Expires=%d but POST Expires=%d", got.Expires, posted.Expires)
	}
}
