package handlers

import (
	"encoding/json"
	"net/http"
)

// WriteJSONListResponse writes a JSON list response envelope with the given items.
// statusCode is used as the HTTP status code. items must be a slice value.
func WriteJSONListResponse(w http.ResponseWriter, statusCode int, items interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	// Build the envelope. We use a map so we can set total_item_count generically.
	// The items value is embedded directly (the caller passes a typed slice).
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"items":            items,
		"total_item_count": countItems(items),
	})
}

// WriteJSONError writes a JSON error response in FlashBlade API format.
func WriteJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"errors": []map[string]string{
			{"message": message},
		},
	})
}

// countItems returns the length of the items slice using JSON round-trip.
// This is a helper that avoids reflection — items is re-marshaled and counted.
func countItems(items interface{}) int {
	b, err := json.Marshal(items)
	if err != nil {
		return 0
	}
	var arr []json.RawMessage
	if err := json.Unmarshal(b, &arr); err != nil {
		return 0
	}
	return len(arr)
}
