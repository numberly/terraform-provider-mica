package client

import (
	"context"
	"fmt"
	"net/url"
)

// GetObjectStoreUser retrieves an object store user by name (format: "account/username").
// Returns an IsNotFound error if the user does not exist.
func (c *FlashBladeClient) GetObjectStoreUser(ctx context.Context, name string) (*ObjectStoreUser, error) {
	return getOneByName[ObjectStoreUser](c, ctx, "/object-store-users?names="+url.QueryEscape(name), "object store user", name)
}

// PostObjectStoreUser creates a new object store user.
// The user name is passed as a query parameter (?names=); optional fields (e.g. full_access) are in body.
func (c *FlashBladeClient) PostObjectStoreUser(ctx context.Context, name string, body ObjectStoreUserPost) (*ObjectStoreUser, error) {
	path := "/object-store-users?names=" + url.QueryEscape(name)
	if body.FullAccess != nil {
		path += fmt.Sprintf("&full_access=%t", *body.FullAccess)
	}
	var resp ListResponse[ObjectStoreUser]
	if err := c.post(ctx, path, nil, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PostObjectStoreUser: empty response from server")
	}
	return &resp.Items[0], nil
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
	_, err := c.PostObjectStoreUser(ctx, name, ObjectStoreUserPost{})
	if err == nil {
		return nil
	}
	// If user already exists, that's fine. FlashBlade may return 409 Conflict
	// or 400 with "already exists" message depending on the firmware version.
	if IsConflict(err) || isAlreadyExists(err) {
		return nil
	}
	return fmt.Errorf("ensuring object store user %q: %w", name, err)
}

// ListObjectStoreUserPolicies returns all access policies attached to the given object store user.
func (c *FlashBladeClient) ListObjectStoreUserPolicies(ctx context.Context, userName string) ([]ObjectStoreUserPolicyMember, error) {
	path := "/object-store-users/object-store-access-policies?member_names=" + url.QueryEscape(userName)
	var resp ListResponse[ObjectStoreUserPolicyMember]
	if err := c.get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

// PostObjectStoreUserPolicy attaches an access policy to an object store user.
// Both userName and policyName are passed as query parameters; no request body is needed.
func (c *FlashBladeClient) PostObjectStoreUserPolicy(ctx context.Context, userName, policyName string) (*ObjectStoreUserPolicyMember, error) {
	path := "/object-store-users/object-store-access-policies?member_names=" + url.QueryEscape(userName) +
		"&policy_names=" + url.QueryEscape(policyName)
	var resp ListResponse[ObjectStoreUserPolicyMember]
	if err := c.post(ctx, path, nil, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PostObjectStoreUserPolicy: empty response from server")
	}
	return &resp.Items[0], nil
}

// DeleteObjectStoreUserPolicy detaches an access policy from an object store user.
func (c *FlashBladeClient) DeleteObjectStoreUserPolicy(ctx context.Context, userName, policyName string) error {
	path := "/object-store-users/object-store-access-policies?member_names=" + url.QueryEscape(userName) +
		"&policy_names=" + url.QueryEscape(policyName)
	return c.delete(ctx, path)
}
