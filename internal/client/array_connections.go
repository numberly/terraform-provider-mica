package client

import (
	"context"
	"fmt"
	"net/url"
)

// GetArrayConnection retrieves an array connection by remote name.
// Returns an IsNotFound error if no connection matches.
func (c *FlashBladeClient) GetArrayConnection(ctx context.Context, remoteName string) (*ArrayConnection, error) {
	path := "/array-connections?remote_names=" + url.QueryEscape(remoteName)
	var resp ListResponse[ArrayConnection]
	if err := c.get(ctx, path, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, &APIError{StatusCode: 404, Message: fmt.Sprintf("array connection with remote name %q not found", remoteName)}
	}
	return &resp.Items[0], nil
}

// ListArrayConnections returns all array connections with automatic pagination.
func (c *FlashBladeClient) ListArrayConnections(ctx context.Context) ([]ArrayConnection, error) {
	params := url.Values{}
	var all []ArrayConnection
	for {
		path := "/array-connections"
		if len(params) > 0 {
			path += "?" + params.Encode()
		}

		var resp ListResponse[ArrayConnection]
		if err := c.get(ctx, path, &resp); err != nil {
			return nil, err
		}
		all = append(all, resp.Items...)
		if resp.ContinuationToken == "" {
			break
		}
		params.Set("continuation_token", resp.ContinuationToken)
	}
	return all, nil
}
