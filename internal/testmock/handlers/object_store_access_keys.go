package handlers

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// accessKeyStore is the thread-safe in-memory state for object store access key handlers.
type accessKeyStore struct {
	mu       sync.Mutex
	byName   map[string]*client.ObjectStoreAccessKey
	accounts *objectStoreAccountStore
}

// RegisterObjectStoreAccessKeyHandlers registers CRUD handlers for /api/2.22/object-store-access-keys
// against the provided ServeMux. The accounts store is used to validate user account existence.
// The store pointer is returned for cross-reference if needed.
func RegisterObjectStoreAccessKeyHandlers(mux *http.ServeMux, accounts *objectStoreAccountStore) *accessKeyStore {
	store := &accessKeyStore{
		byName:   make(map[string]*client.ObjectStoreAccessKey),
		accounts: accounts,
	}
	mux.HandleFunc("/api/2.22/object-store-access-keys", store.handle)
	return store
}

func (s *accessKeyStore) handle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleGet(w, r)
	case http.MethodPost:
		s.handlePost(w, r)
	case http.MethodDelete:
		s.handleDelete(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGet handles GET /api/2.22/object-store-access-keys with optional ?names= param.
// IMPORTANT: secret_access_key is NOT returned in GET responses — it is set to empty string.
func (s *accessKeyStore) handleGet(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	namesFilter := r.URL.Query().Get("names")

	var items []client.ObjectStoreAccessKey

	if namesFilter != "" {
		key, ok := s.byName[namesFilter]
		if ok {
			// Return a copy with secret_access_key stripped — simulates real API behavior.
			redacted := *key
			redacted.SecretAccessKey = ""
			items = append(items, redacted)
		}
	} else {
		for _, key := range s.byName {
			redacted := *key
			redacted.SecretAccessKey = ""
			items = append(items, redacted)
		}
	}

	if items == nil {
		items = []client.ObjectStoreAccessKey{}
	}

	WriteJSONListResponse(w, http.StatusOK, items)
}

// handlePost handles POST /api/2.22/object-store-access-keys.
// Body must contain {user: {name: "<account>/admin"}}.
// Response includes secret_access_key — it will never be returned again on subsequent GETs.
func (s *accessKeyStore) handlePost(w http.ResponseWriter, r *http.Request) {
	var body client.ObjectStoreAccessKeyPost
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	if body.User.Name == "" {
		WriteJSONError(w, http.StatusBadRequest, "user.name is required")
		return
	}

	// Extract account name from user name (format: "<account>/admin").
	accountName := extractAccountName(body.User.Name)
	if accountName == "" {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid user name format %q: expected <account>/admin", body.User.Name))
		return
	}

	// Validate that the referenced account exists.
	s.accounts.mu.Lock()
	_, accountExists := s.accounts.byName[accountName]
	s.accounts.mu.Unlock()

	if !accountExists {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("object store account %q not found", accountName))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate random access key ID. Use caller-provided secret if present.
	accessKeyID := generateAccessKeyID()
	secretAccessKey := body.SecretAccessKey
	if secretAccessKey == "" {
		secretAccessKey = generateSecretAccessKey()
	}

	// Access key name format: <account>/admin/<key-id>
	keyName := fmt.Sprintf("%s/admin/%s", accountName, accessKeyID)

	key := &client.ObjectStoreAccessKey{
		Name:            keyName,
		AccessKeyID:     accessKeyID,
		SecretAccessKey: secretAccessKey,
		Created:         time.Now().UnixMilli(),
		Enabled:         true,
		User: client.NamedReference{
			Name: body.User.Name,
		},
	}

	s.byName[keyName] = key

	// Return the full key including secret_access_key (POST only).
	WriteJSONListResponse(w, http.StatusOK, []client.ObjectStoreAccessKey{*key})
}

// handleDelete handles DELETE /api/2.22/object-store-access-keys?names={name}.
func (s *accessKeyStore) handleDelete(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("names")
	if name == "" {
		WriteJSONError(w, http.StatusBadRequest, "names query parameter is required for DELETE")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.byName[name]; !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("object store access key %q not found", name))
		return
	}

	delete(s.byName, name)

	w.WriteHeader(http.StatusOK)
}

// extractAccountName extracts the account portion from a user name formatted as "<account>/admin".
// Returns empty string if the format is not recognized.
func extractAccountName(userName string) string {
	parts := strings.SplitN(userName, "/", 2)
	if len(parts) != 2 {
		return ""
	}
	return parts[0]
}

// generateAccessKeyID generates a random access key ID in the style of FlashBlade.
func generateAccessKeyID() string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 16)
	for i := range b {
		b[i] = chars[rand.IntN(len(chars))] //nolint:gosec // G404: test mock generating fake key IDs — crypto strength not needed
	}
	return "PSFB" + string(b)
}

// generateSecretAccessKey generates a random 40-character secret access key.
func generateSecretAccessKey() string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789+/"
	b := make([]byte, 40)
	for i := range b {
		b[i] = chars[rand.IntN(len(chars))] //nolint:gosec // G404: test mock generating fake secret keys — crypto strength not needed
	}
	return string(b)
}
