package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// certificateStore is the thread-safe in-memory state for certificate handlers.
type certificateStore struct {
	mu     sync.Mutex
	byName map[string]*client.Certificate
	nextID int
}

// RegisterCertificateHandlers registers CRUD handlers for /api/2.22/certificates
// against the provided ServeMux. The store pointer is returned for test setup.
func RegisterCertificateHandlers(mux *http.ServeMux) *certificateStore {
	store := &certificateStore{
		byName: make(map[string]*client.Certificate),
		nextID: 1,
	}
	mux.HandleFunc("/api/2.22/certificates", store.handle)
	return store
}

// Seed adds a certificate directly to the store for test setup.
func (s *certificateStore) Seed(item *client.Certificate) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.byName[item.Name] = item
}

// handle dispatches certificate requests by HTTP method.
func (s *certificateStore) handle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleGet(w, r)
	case http.MethodPost:
		s.handlePost(w, r)
	case http.MethodPatch:
		s.handlePatch(w, r)
	case http.MethodDelete:
		s.handleDelete(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGet handles GET /api/2.22/certificates with optional ?names= param.
// Returns empty list (HTTP 200) when name not found — matches real API behavior.
func (s *certificateStore) handleGet(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"names"}) {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	namesFilter := r.URL.Query().Get("names")

	var items []client.Certificate

	if namesFilter != "" {
		cert, ok := s.byName[namesFilter]
		if ok {
			items = append(items, *cert)
		}
	} else {
		for _, cert := range s.byName {
			items = append(items, *cert)
		}
	}

	if items == nil {
		items = []client.Certificate{}
	}

	WriteJSONListResponse(w, http.StatusOK, items)
}

// handlePost handles POST /api/2.22/certificates?names={name}.
// Requires non-empty certificate (PEM) in body. Returns 409 if name already exists.
// Populates computed fields; private_key and passphrase are NOT stored (write-only).
func (s *certificateStore) handlePost(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"names"}) {
		return
	}

	name, ok := RequireQueryParam(w, r, "names")
	if !ok {
		return
	}

	var body client.CertificatePost
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	if body.Certificate == "" {
		WriteJSONError(w, http.StatusBadRequest, "certificate is required")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.byName[name]; exists {
		WriteJSONError(w, http.StatusConflict, fmt.Sprintf("certificate %q already exists", name))
		return
	}

	id := fmt.Sprintf("cert-%d", s.nextID)
	s.nextID++

	// Match real FlashBlade API: certificate_type defaults to "external" when
	// omitted at creation time (see swagger _certificateBase.certificate_type).
	certType := body.CertificateType
	if certType == "" {
		certType = "external"
	}

	cert := &client.Certificate{
		ID:                      id,
		Name:                    name,
		Certificate:             body.Certificate,
		CertificateType:         certType,
		IntermediateCertificate: body.IntermediateCertificate,
		CommonName:              "test-cert",
		IssuedBy:                "CN=Test CA",
		IssuedTo:                "CN=test-cert",
		Status:                  "imported",
		ValidFrom:               1700000000000,
		ValidTo:                 1731536000000,
		KeyAlgorithm:            "RSA",
		KeySize:                 2048,
	}

	s.byName[name] = cert

	WriteJSONListResponse(w, http.StatusOK, []client.Certificate{*cert})
}

// handlePatch handles PATCH /api/2.22/certificates?names={name}.
// Applies non-nil pointer fields. Returns 404 if not found.
// private_key and passphrase are accepted but not stored (write-only).
func (s *certificateStore) handlePatch(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"names"}) {
		return
	}

	name, ok := RequireQueryParam(w, r, "names")
	if !ok {
		return
	}

	var body client.CertificatePatch
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	cert, exists := s.byName[name]
	if !exists {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("certificate %q not found", name))
		return
	}

	if body.Certificate != nil {
		cert.Certificate = *body.Certificate
	}
	if body.IntermediateCertificate != nil {
		cert.IntermediateCertificate = *body.IntermediateCertificate
	}
	// Passphrase and PrivateKey are write-only — accepted but not stored or returned.

	WriteJSONListResponse(w, http.StatusOK, []client.Certificate{*cert})
}

// handleDelete handles DELETE /api/2.22/certificates?names={name}.
func (s *certificateStore) handleDelete(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"names"}) {
		return
	}

	name, ok := RequireQueryParam(w, r, "names")
	if !ok {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.byName[name]; !exists {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("certificate %q not found", name))
		return
	}

	delete(s.byName, name)

	w.WriteHeader(http.StatusOK)
}
