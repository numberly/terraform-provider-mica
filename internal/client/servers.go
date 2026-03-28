package client

import (
	"context"
	"fmt"
	"net/url"
	"strings"
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

// PostServer creates a new server. The name is passed via the ?create_ds= query parameter.
func (c *FlashBladeClient) PostServer(ctx context.Context, name string, body ServerPost) (*Server, error) {
	path := "/servers?create_ds=" + url.QueryEscape(name)
	var resp ListResponse[Server]
	if err := c.post(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PostServer: empty response from server")
	}
	return &resp.Items[0], nil
}

// PatchServer updates an existing server identified by name.
func (c *FlashBladeClient) PatchServer(ctx context.Context, name string, body ServerPatch) (*Server, error) {
	path := "/servers?names=" + url.QueryEscape(name)
	var resp ListResponse[Server]
	if err := c.patch(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PatchServer: empty response from server")
	}
	return &resp.Items[0], nil
}

// DeleteServer removes a server by name. If cascadeDelete is non-empty,
// the comma-joined export names are passed via the ?cascade_delete= query parameter.
func (c *FlashBladeClient) DeleteServer(ctx context.Context, name string, cascadeDelete []string) error {
	path := "/servers?names=" + url.QueryEscape(name)
	if len(cascadeDelete) > 0 {
		path += "&cascade_delete=" + url.QueryEscape(strings.Join(cascadeDelete, ","))
	}
	return c.delete(ctx, path)
}
