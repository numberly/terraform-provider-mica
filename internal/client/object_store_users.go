package client

import (
	"context"
	"fmt"
	"net/url"
)

// GetObjectStoreUser retrieves an object store user by name (format: "account/username").
// Returns an IsNotFound error if the user does not exist.
func (c *FlashBladeClient) GetObjectStoreUser(ctx context.Context, name string) error {
	path := "/object-store-users?names=" + url.QueryEscape(name)
	var resp ListResponse[map[string]any]
	if err := c.get(ctx, path, &resp); err != nil {
		return err
	}
	if len(resp.Items) == 0 {
		return &APIError{StatusCode: 404, Message: fmt.Sprintf("object store user %q not found", name)}
	}
	return nil
}

// PostObjectStoreUser creates a new object store user.
// The user name is passed as a query parameter (?names=).
func (c *FlashBladeClient) PostObjectStoreUser(ctx context.Context, name string) error {
	path := "/object-store-users?names=" + url.QueryEscape(name)
	var resp ListResponse[map[string]any]
	return c.post(ctx, path, nil, &resp)
}

// DeleteObjectStoreUser deletes an object store user by name.
func (c *FlashBladeClient) DeleteObjectStoreUser(ctx context.Context, name string) error {
	path := "/object-store-users?names=" + url.QueryEscape(name)
	return c.delete(ctx, path)
}

// EnsureObjectStoreUser creates the object store user if it does not exist.
// This is needed before creating access keys, as keys reference a user.
// Uses POST-first strategy: FlashBlade returns HTTP 400 (not 404) for
// non-existent users on GET, so we try POST and tolerate "already exists" (409).
func (c *FlashBladeClient) EnsureObjectStoreUser(ctx context.Context, name string) error {
	err := c.PostObjectStoreUser(ctx, name)
	if err == nil {
		return nil
	}
	// If user already exists (409 Conflict), that's fine.
	if IsConflict(err) {
		return nil
	}
	return fmt.Errorf("ensuring object store user %q: %w", name, err)
}
