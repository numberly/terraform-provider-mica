package client

import (
	"context"
	"net/url"
	"strings"
)

// ListObjectStoreAccountsOpts contains optional query parameters for ListObjectStoreAccounts.
type ListObjectStoreAccountsOpts struct {
	// Names filters results to specific account names (comma-separated when multiple).
	Names []string
	// Filter is a free-form filter expression.
	Filter string
}

// GetObjectStoreAccount retrieves an object store account by name.
// Returns an IsNotFound error if the account does not exist.
func (c *FlashBladeClient) GetObjectStoreAccount(ctx context.Context, name string) (*ObjectStoreAccount, error) {
	return getOneByName[ObjectStoreAccount](c, ctx, "/object-store-accounts?names="+url.QueryEscape(name), "object store account", name)
}

// ListObjectStoreAccounts returns all object store accounts matching the optional opts filters.
// It automatically follows continuation_token pagination to collect all results.
func (c *FlashBladeClient) ListObjectStoreAccounts(ctx context.Context, opts ListObjectStoreAccountsOpts) ([]ObjectStoreAccount, error) {
	params := url.Values{}
	if len(opts.Names) > 0 {
		params.Set("names", strings.Join(opts.Names, ","))
	}
	if opts.Filter != "" {
		params.Set("filter", opts.Filter)
	}

	return listAll[ObjectStoreAccount](c, ctx, "/object-store-accounts", params)
}

// PostObjectStoreAccount creates a new object store account.
// The name is passed as a query parameter; optional fields are in the body.
func (c *FlashBladeClient) PostObjectStoreAccount(ctx context.Context, name string, body ObjectStoreAccountPost) (*ObjectStoreAccount, error) {
	return postOne[ObjectStoreAccountPost, ObjectStoreAccount](c, ctx, "/object-store-accounts?names="+url.QueryEscape(name), body, "PostObjectStoreAccount")
}

// PatchObjectStoreAccount updates an existing object store account identified by name.
// Only non-nil pointer fields in body are sent (PATCH semantics).
func (c *FlashBladeClient) PatchObjectStoreAccount(ctx context.Context, name string, body ObjectStoreAccountPatch) (*ObjectStoreAccount, error) {
	return patchOne[ObjectStoreAccountPatch, ObjectStoreAccount](c, ctx, "/object-store-accounts?names="+url.QueryEscape(name), body, "PatchObjectStoreAccount")
}

// DeleteObjectStoreAccount permanently deletes an object store account (single-phase, no soft-delete).
func (c *FlashBladeClient) DeleteObjectStoreAccount(ctx context.Context, name string) error {
	path := "/object-store-accounts?names=" + url.QueryEscape(name)
	return c.delete(ctx, path)
}
