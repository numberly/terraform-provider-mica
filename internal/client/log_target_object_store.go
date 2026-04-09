package client

import (
	"context"
	"net/url"
)

// GetLogTargetObjectStore retrieves a log target object store by name.
// Returns an IsNotFound error if the entry does not exist.
func (c *FlashBladeClient) GetLogTargetObjectStore(ctx context.Context, name string) (*LogTargetObjectStore, error) {
	return getOneByName[LogTargetObjectStore](c, ctx, "/log-targets/object-store?names="+url.QueryEscape(name), "log target object store", name)
}

// PostLogTargetObjectStore creates a new log target object store.
// The name is passed via ?names= query parameter; bucket and log settings are in the body.
func (c *FlashBladeClient) PostLogTargetObjectStore(ctx context.Context, name string, body LogTargetObjectStorePost) (*LogTargetObjectStore, error) {
	return postOne[LogTargetObjectStorePost, LogTargetObjectStore](c, ctx, "/log-targets/object-store?names="+url.QueryEscape(name), body, "PostLogTargetObjectStore")
}

// PatchLogTargetObjectStore updates an existing log target object store identified by name.
// Only non-nil pointer fields in body are sent (PATCH semantics).
func (c *FlashBladeClient) PatchLogTargetObjectStore(ctx context.Context, name string, body LogTargetObjectStorePatch) (*LogTargetObjectStore, error) {
	return patchOne[LogTargetObjectStorePatch, LogTargetObjectStore](c, ctx, "/log-targets/object-store?names="+url.QueryEscape(name), body, "PatchLogTargetObjectStore")
}

// DeleteLogTargetObjectStore permanently deletes a log target object store by name.
func (c *FlashBladeClient) DeleteLogTargetObjectStore(ctx context.Context, name string) error {
	path := "/log-targets/object-store?names=" + url.QueryEscape(name)
	return c.delete(ctx, path)
}
