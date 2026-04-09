package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// arrayDnsStore is a thread-safe in-memory store for DNS configuration entries.
type arrayDnsStore struct {
	mu     sync.Mutex
	byName map[string]*client.ArrayDns
	nextID int
}

// Seed adds a DNS entry to the store for test setup.
func (s *arrayDnsStore) Seed(item *client.ArrayDns) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.byName[item.Name] = item
}

// arrayAdminStore is the thread-safe in-memory state for array admin handlers.
type arrayAdminStore struct {
	mu            sync.Mutex
	DnsStore      *arrayDnsStore
	arrayInfo     *client.ArrayInfo
	smtp          *client.SmtpServer
	alertWatchers map[string]*client.AlertWatcher // email -> watcher
}

// SeedDns is a convenience method that delegates to the embedded DNS store.
func (s *arrayAdminStore) SeedDns(item *client.ArrayDns) {
	s.DnsStore.Seed(item)
}

// RegisterArrayAdminHandlers registers CRUD handlers for DNS, NTP, SMTP, and alert watchers.
func RegisterArrayAdminHandlers(mux *http.ServeMux) *arrayAdminStore {
	dnsStore := &arrayDnsStore{
		byName: make(map[string]*client.ArrayDns),
		nextID: 1,
	}

	store := &arrayAdminStore{
		DnsStore: dnsStore,
		arrayInfo: &client.ArrayInfo{
			ID:         uuid.New().String(),
			Name:       "array0",
			NtpServers: []string{},
		},
		smtp: &client.SmtpServer{
			ID:   uuid.New().String(),
			Name: "smtp",
		},
		alertWatchers: make(map[string]*client.AlertWatcher),
	}

	mux.HandleFunc("/api/2.22/dns", dnsStore.handleDns)
	mux.HandleFunc("/api/2.22/arrays", store.handleArrays)
	mux.HandleFunc("/api/2.22/smtp-servers", store.handleSmtp)
	mux.HandleFunc("/api/2.22/alert-watchers", store.handleAlertWatchers)
	return store
}

func (s *arrayDnsStore) handleDns(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"names"}) {
		return
	}
	switch r.Method {
	case http.MethodGet:
		s.handleDnsGet(w, r)
	case http.MethodPost:
		s.handleDnsPost(w, r)
	case http.MethodPatch:
		s.handleDnsPatch(w, r)
	case http.MethodDelete:
		s.handleDnsDelete(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *arrayDnsStore) handleDnsGet(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	namesFilter := r.URL.Query().Get("names")

	var items []client.ArrayDns
	if namesFilter != "" {
		item, ok := s.byName[namesFilter]
		if ok {
			items = append(items, *item)
		}
	} else {
		for _, item := range s.byName {
			items = append(items, *item)
		}
	}
	if items == nil {
		items = []client.ArrayDns{}
	}
	WriteJSONListResponse(w, http.StatusOK, items)
}

func (s *arrayDnsStore) handleDnsPost(w http.ResponseWriter, r *http.Request) {
	name, ok := RequireQueryParam(w, r, "names")
	if !ok {
		return
	}

	var body client.ArrayDnsPost
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.byName[name]; exists {
		WriteJSONError(w, http.StatusConflict, fmt.Sprintf("DNS configuration %q already exists", name))
		return
	}

	item := &client.ArrayDns{
		ID:          fmt.Sprintf("dns-%d", s.nextID),
		Name:        name,
		Domain:      body.Domain,
		Nameservers: body.Nameservers,
		Services:    body.Services,
		Sources:     body.Sources,
	}
	s.nextID++
	if item.Nameservers == nil {
		item.Nameservers = []string{}
	}
	if item.Services == nil {
		item.Services = []string{}
	}
	if item.Sources == nil {
		item.Sources = []string{}
	}

	s.byName[name] = item
	WriteJSONListResponse(w, http.StatusOK, []client.ArrayDns{*item})
}

func (s *arrayDnsStore) handleDnsPatch(w http.ResponseWriter, r *http.Request) {
	name, ok := RequireQueryParam(w, r, "names")
	if !ok {
		return
	}

	var rawPatch map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&rawPatch); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	item, exists := s.byName[name]
	if !exists {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("DNS configuration %q not found", name))
		return
	}

	if v, ok := rawPatch["domain"]; ok {
		var domain string
		if err := json.Unmarshal(v, &domain); err == nil {
			item.Domain = domain
		}
	}
	if v, ok := rawPatch["nameservers"]; ok {
		var nameservers []string
		if err := json.Unmarshal(v, &nameservers); err == nil {
			item.Nameservers = nameservers
		}
	}
	if v, ok := rawPatch["services"]; ok {
		var services []string
		if err := json.Unmarshal(v, &services); err == nil {
			item.Services = services
		}
	}
	if v, ok := rawPatch["sources"]; ok {
		var sources []string
		if err := json.Unmarshal(v, &sources); err == nil {
			item.Sources = sources
		}
	}

	WriteJSONListResponse(w, http.StatusOK, []client.ArrayDns{*item})
}

func (s *arrayDnsStore) handleDnsDelete(w http.ResponseWriter, r *http.Request) {
	name, ok := RequireQueryParam(w, r, "names")
	if !ok {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.byName[name]; !exists {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("DNS configuration %q not found", name))
		return
	}

	delete(s.byName, name)
	w.WriteHeader(http.StatusOK)
}

func (s *arrayAdminStore) handleArrays(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleArraysGet(w, r)
	case http.MethodPatch:
		s.handleArraysPatch(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *arrayAdminStore) handleSmtp(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleSmtpGet(w, r)
	case http.MethodPatch:
		s.handleSmtpPatch(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *arrayAdminStore) handleAlertWatchers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleAlertWatchersGet(w, r)
	case http.MethodPost:
		s.handleAlertWatchersPost(w, r)
	case http.MethodPatch:
		s.handleAlertWatchersPatch(w, r)
	case http.MethodDelete:
		s.handleAlertWatchersDelete(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *arrayAdminStore) handleArraysGet(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	WriteJSONListResponse(w, http.StatusOK, []client.ArrayInfo{*s.arrayInfo})
}

func (s *arrayAdminStore) handleArraysPatch(w http.ResponseWriter, r *http.Request) {
	var rawPatch map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&rawPatch); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Only process ntp_servers — other array fields are intentionally ignored in mock.
	if v, ok := rawPatch["ntp_servers"]; ok {
		var ntpServers []string
		if err := json.Unmarshal(v, &ntpServers); err == nil {
			s.arrayInfo.NtpServers = ntpServers
		}
	}

	WriteJSONListResponse(w, http.StatusOK, []client.ArrayInfo{*s.arrayInfo})
}

func (s *arrayAdminStore) handleSmtpGet(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	WriteJSONListResponse(w, http.StatusOK, []client.SmtpServer{*s.smtp})
}

func (s *arrayAdminStore) handleSmtpPatch(w http.ResponseWriter, r *http.Request) {
	var rawPatch map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&rawPatch); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if v, ok := rawPatch["relay_host"]; ok {
		var relayHost string
		if err := json.Unmarshal(v, &relayHost); err == nil {
			s.smtp.RelayHost = relayHost
		}
	}
	if v, ok := rawPatch["sender_domain"]; ok {
		var senderDomain string
		if err := json.Unmarshal(v, &senderDomain); err == nil {
			s.smtp.SenderDomain = senderDomain
		}
	}
	if v, ok := rawPatch["encryption_mode"]; ok {
		var encryptionMode string
		if err := json.Unmarshal(v, &encryptionMode); err == nil {
			s.smtp.EncryptionMode = encryptionMode
		}
	}

	WriteJSONListResponse(w, http.StatusOK, []client.SmtpServer{*s.smtp})
}

func (s *arrayAdminStore) handleAlertWatchersGet(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var items []client.AlertWatcher
	for _, watcher := range s.alertWatchers {
		items = append(items, *watcher)
	}
	if items == nil {
		items = []client.AlertWatcher{}
	}

	WriteJSONListResponse(w, http.StatusOK, items)
}

func (s *arrayAdminStore) handleAlertWatchersPost(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("names")
	if email == "" {
		WriteJSONError(w, http.StatusBadRequest, "names query parameter (email) is required for POST")
		return
	}

	var body client.AlertWatcherPost
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.alertWatchers[email]; exists {
		WriteJSONError(w, http.StatusConflict, fmt.Sprintf("alert watcher %q already exists", email))
		return
	}

	watcher := &client.AlertWatcher{
		ID:                          uuid.New().String(),
		Name:                        email,
		Enabled:                     true,
		MinimumNotificationSeverity: body.MinimumNotificationSeverity,
	}

	s.alertWatchers[email] = watcher

	WriteJSONListResponse(w, http.StatusOK, []client.AlertWatcher{*watcher})
}

func (s *arrayAdminStore) handleAlertWatchersPatch(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("names")
	if email == "" {
		WriteJSONError(w, http.StatusBadRequest, "names query parameter (email) is required for PATCH")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	watcher, ok := s.alertWatchers[email]
	if !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("alert watcher %q not found", email))
		return
	}

	var rawPatch map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&rawPatch); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	if v, ok := rawPatch["enabled"]; ok {
		var enabled bool
		if err := json.Unmarshal(v, &enabled); err == nil {
			watcher.Enabled = enabled
		}
	}
	if v, ok := rawPatch["minimum_notification_severity"]; ok {
		var severity string
		if err := json.Unmarshal(v, &severity); err == nil {
			watcher.MinimumNotificationSeverity = severity
		}
	}

	WriteJSONListResponse(w, http.StatusOK, []client.AlertWatcher{*watcher})
}

func (s *arrayAdminStore) handleAlertWatchersDelete(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("names")
	if email == "" {
		WriteJSONError(w, http.StatusBadRequest, "names query parameter (email) is required for DELETE")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.alertWatchers[email]; !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("alert watcher %q not found", email))
		return
	}

	delete(s.alertWatchers, email)

	w.WriteHeader(http.StatusOK)
}
