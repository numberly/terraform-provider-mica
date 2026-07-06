package handlers

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strconv"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

// APIPrefix is the versioned API path prefix, derived from the client's single
// source of truth so an API version bump touches only client.APIVersion.
const APIPrefix = "/api/" + client.APIVersion

// WriteJSONListResponse writes a JSON list response envelope with the given items.
// statusCode is used as the HTTP status code. items must be a slice value.
func WriteJSONListResponse(w http.ResponseWriter, statusCode int, items any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	// Build the envelope. We use a map so we can set total_item_count generically.
	// The items value is embedded directly (the caller passes a typed slice).
	_ = json.NewEncoder(w).Encode(map[string]any{
		"items":            items,
		"total_item_count": countItems(items),
	})
}

// WriteJSONError writes a JSON error response in FlashBlade API format.
func WriteJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"errors": []map[string]string{
			{"message": message},
		},
	})
}

// countItems returns the length of the items slice using reflect.
func countItems(items any) int {
	v := reflect.ValueOf(items)
	if !v.IsValid() || v.Kind() != reflect.Slice {
		return 0
	}
	return v.Len()
}

// parseQuotaLimit converts a string quota_limit (as sent by POST/PATCH) to int64 (as stored internally).
func parseQuotaLimit(s string) int64 {
	if s == "" {
		return 0
	}
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
}
