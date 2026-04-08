package client

import (
	"context"
	"fmt"
)

// GetArrayConnectionKey retrieves the current connection key.
// Returns the key or an error. The endpoint returns a single object, not a list.
func (c *FlashBladeClient) GetArrayConnectionKey(ctx context.Context) (*ArrayConnectionKey, error) {
	var result ArrayConnectionKey
	if err := c.get(ctx, "/array-connections/connection-key", &result); err != nil {
		return nil, fmt.Errorf("GetArrayConnectionKey: %w", err)
	}
	return &result, nil
}

// PostArrayConnectionKey generates a new connection key, replacing any existing one.
// The POST takes no body. Returns the newly generated key.
func (c *FlashBladeClient) PostArrayConnectionKey(ctx context.Context) (*ArrayConnectionKey, error) {
	var result ArrayConnectionKey
	if err := c.post(ctx, "/array-connections/connection-key", nil, &result); err != nil {
		return nil, fmt.Errorf("PostArrayConnectionKey: %w", err)
	}
	return &result, nil
}
