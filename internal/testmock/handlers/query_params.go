package handlers

import (
	"fmt"
	"net/http"
)

// globalQueryParams are framework-level query parameters that the FlashBlade API
// accepts on every list endpoint. They are always allowed in addition to
// endpoint-specific parameters.
var globalQueryParams = []string{
	"continuation_token",
	"limit",
	"offset",
	"sort",
	"filter",
	"total_item_count",
}

// ValidateQueryParams checks that all query parameters in the request are in the
// allowed set. If an unknown param is found, it writes a 400 JSON error response
// and returns false. The caller should return immediately when false is returned.
//
// allowedParams is the set of param names valid for this endpoint+method
// combination. The global framework params (continuation_token, limit, offset,
// sort, filter, total_item_count) are always allowed automatically.
func ValidateQueryParams(w http.ResponseWriter, r *http.Request, allowedParams []string) bool {
	allowed := make(map[string]bool, len(allowedParams)+len(globalQueryParams))
	for _, p := range globalQueryParams {
		allowed[p] = true
	}
	for _, p := range allowedParams {
		allowed[p] = true
	}

	for key := range r.URL.Query() {
		if !allowed[key] {
			WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("unknown query parameter: %s", key))
			return false
		}
	}
	return true
}

// RequireQueryParam checks that the given param is present and non-empty.
// If missing, it writes a 400 JSON error and returns ("", false).
func RequireQueryParam(w http.ResponseWriter, r *http.Request, param string) (string, bool) {
	val := r.URL.Query().Get(param)
	if val == "" {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("%s query parameter is required", param))
		return "", false
	}
	return val, true
}
