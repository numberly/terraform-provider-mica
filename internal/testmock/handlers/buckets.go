package handlers

import (
	"net/http"

	"github.com/soulkyu/terraform-provider-flashblade/internal/client"
)

// RegisterBucketHandlers registers a minimal /api/2.22/buckets handler for testing.
// This stub returns an empty list for all GET requests.
// Full bucket CRUD will be implemented in Plan 02-02.
func RegisterBucketHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/api/2.22/buckets", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			WriteJSONListResponse(w, http.StatusOK, []client.Bucket{})
		} else {
			http.Error(w, "not implemented", http.StatusNotImplemented)
		}
	})
}
