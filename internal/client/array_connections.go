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
	return listAll[ArrayConnection](c, ctx, "/array-connections", params)
}

// PostArrayConnection creates a new array connection.
// The remote name is passed via ?remote_names= query parameter.
func (c *FlashBladeClient) PostArrayConnection(ctx context.Context, remoteName string, body ArrayConnectionPost) (*ArrayConnection, error) {
	return postOne[ArrayConnectionPost, ArrayConnection](c, ctx, "/array-connections?remote_names="+url.QueryEscape(remoteName), body, "PostArrayConnection")
}

// PatchArrayConnection updates an existing array connection identified by remote name.
// Only non-nil pointer fields in body are sent (PATCH semantics).
func (c *FlashBladeClient) PatchArrayConnection(ctx context.Context, remoteName string, body ArrayConnectionPatch) (*ArrayConnection, error) {
	return patchOne[ArrayConnectionPatch, ArrayConnection](c, ctx, "/array-connections?remote_names="+url.QueryEscape(remoteName), body, "PatchArrayConnection")
}

// DeleteArrayConnection permanently removes an array connection by remote name.
func (c *FlashBladeClient) DeleteArrayConnection(ctx context.Context, remoteName string) error {
	return c.delete(ctx, "/array-connections?remote_names="+url.QueryEscape(remoteName))
}
