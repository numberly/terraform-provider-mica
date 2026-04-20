package client

import (
	"context"
	"net/url"
)

// GetDirectoryServiceRole retrieves a single directory service role by name.
// Returns IsNotFound when the role does not exist (empty list + HTTP 200 from API).
func (c *FlashBladeClient) GetDirectoryServiceRole(ctx context.Context, name string) (*DirectoryServiceRole, error) {
	return getOneByName[DirectoryServiceRole](c, ctx, "/directory-services/roles?names="+url.QueryEscape(name), "directory_service_role", name)
}

// PostDirectoryServiceRole creates a directory service role.
// The name is passed via ?names= query parameter (per D-01).
func (c *FlashBladeClient) PostDirectoryServiceRole(ctx context.Context, name string, body DirectoryServiceRolePost) (*DirectoryServiceRole, error) {
	return postOne[DirectoryServiceRolePost, DirectoryServiceRole](c, ctx, "/directory-services/roles?names="+url.QueryEscape(name), body, "PostDirectoryServiceRole")
}

// PatchDirectoryServiceRole updates a directory service role's mutable fields
// (group, group_base, deprecated role). management_access_policies is readonly
// on PATCH — mutate via DSRM membership resource instead.
func (c *FlashBladeClient) PatchDirectoryServiceRole(ctx context.Context, name string, body DirectoryServiceRolePatch) (*DirectoryServiceRole, error) {
	return patchOne[DirectoryServiceRolePatch, DirectoryServiceRole](c, ctx, "/directory-services/roles?names="+url.QueryEscape(name), body, "PatchDirectoryServiceRole")
}

// DeleteDirectoryServiceRole removes a directory service role by name.
func (c *FlashBladeClient) DeleteDirectoryServiceRole(ctx context.Context, name string) error {
	return c.delete(ctx, "/directory-services/roles?names="+url.QueryEscape(name))
}
