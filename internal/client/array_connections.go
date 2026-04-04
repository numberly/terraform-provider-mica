package client

import (
	"context"
	"net/url"
)

// GetArrayConnection retrieves an array connection by remote name.
// Returns an IsNotFound error if no connection matches.
func (c *FlashBladeClient) GetArrayConnection(ctx context.Context, remoteName string) (*ArrayConnection, error) {
	return getOneByName[ArrayConnection](c, ctx, "/array-connections?remote_names="+url.QueryEscape(remoteName), "array connection", remoteName)
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
