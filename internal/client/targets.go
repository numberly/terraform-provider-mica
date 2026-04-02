package client

import (
	"context"
	"fmt"
	"net/url"
)

// GetTarget retrieves a replication target by name.
// Returns an IsNotFound error if the target does not exist.
func (c *FlashBladeClient) GetTarget(ctx context.Context, name string) (*Target, error) {
	return getOneByName[Target](c, ctx, "/targets?names="+url.QueryEscape(name), "target", name)
}

// PostTarget creates a new replication target.
// The name is passed via ?names= query parameter.
func (c *FlashBladeClient) PostTarget(ctx context.Context, name string, body TargetPost) (*Target, error) {
	path := "/targets?names=" + url.QueryEscape(name)
	var resp ListResponse[Target]
	if err := c.post(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PostTarget: empty response from server")
	}
	return &resp.Items[0], nil
}

// PatchTarget updates an existing replication target identified by name.
// Only non-nil pointer fields in body are sent (PATCH semantics).
func (c *FlashBladeClient) PatchTarget(ctx context.Context, name string, body TargetPatch) (*Target, error) {
	path := "/targets?names=" + url.QueryEscape(name)
	var resp ListResponse[Target]
	if err := c.patch(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PatchTarget: empty response from server")
	}
	return &resp.Items[0], nil
}

// DeleteTarget permanently deletes a replication target by name.
func (c *FlashBladeClient) DeleteTarget(ctx context.Context, name string) error {
	path := "/targets?names=" + url.QueryEscape(name)
	return c.delete(ctx, path)
}
