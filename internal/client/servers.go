package client

import (
	"context"
	"fmt"
	"net/url"
)

// GetServer retrieves a server by name.
// Returns an IsNotFound error if the server does not exist.
func (c *FlashBladeClient) GetServer(ctx context.Context, name string) (*Server, error) {
	path := "/servers?names=" + url.QueryEscape(name)
	var resp ListResponse[Server]
	if err := c.get(ctx, path, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, &APIError{StatusCode: 404, Message: fmt.Sprintf("server %q not found", name)}
	}
	return &resp.Items[0], nil
}
