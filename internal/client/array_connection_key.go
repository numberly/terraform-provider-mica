package client

import (
	"context"
	"fmt"
)

// GetArrayConnectionKey retrieves the current connection key.
// The API wraps the response in {"items": [...]}.
func (c *FlashBladeClient) GetArrayConnectionKey(ctx context.Context) (*ArrayConnectionKey, error) {
	var resp ListResponse[ArrayConnectionKey]
	if err := c.get(ctx, "/array-connections/connection-key", &resp); err != nil {
		return nil, fmt.Errorf("GetArrayConnectionKey: %w", err)
	}
	if len(resp.Items) == 0 {
		return &ArrayConnectionKey{}, nil
	}
	return &resp.Items[0], nil
}

// PostArrayConnectionKey generates a new connection key, replacing any existing one.
// The POST takes no body. The API wraps the response in {"items": [...]}.
func (c *FlashBladeClient) PostArrayConnectionKey(ctx context.Context) (*ArrayConnectionKey, error) {
	var resp ListResponse[ArrayConnectionKey]
	if err := c.post(ctx, "/array-connections/connection-key", nil, &resp); err != nil {
		return nil, fmt.Errorf("PostArrayConnectionKey: %w", err)
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PostArrayConnectionKey: empty response")
	}
	return &resp.Items[0], nil
}
