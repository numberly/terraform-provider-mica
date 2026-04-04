package client

import (
	"context"
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
	return postOne[TargetPost, Target](c, ctx, "/targets?names="+url.QueryEscape(name), body, "PostTarget")
}

// PatchTarget updates an existing replication target identified by name.
// Only non-nil pointer fields in body are sent (PATCH semantics).
func (c *FlashBladeClient) PatchTarget(ctx context.Context, name string, body TargetPatch) (*Target, error) {
	return patchOne[TargetPatch, Target](c, ctx, "/targets?names="+url.QueryEscape(name), body, "PatchTarget")
}

// DeleteTarget permanently deletes a replication target by name.
func (c *FlashBladeClient) DeleteTarget(ctx context.Context, name string) error {
	path := "/targets?names=" + url.QueryEscape(name)
	return c.delete(ctx, path)
}
