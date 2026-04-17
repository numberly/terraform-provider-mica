package client

import (
	"context"
	"net/url"
)

// GetDirectoryServiceManagement retrieves the management directory service configuration.
// The singleton is always queried with name == "management".
// Returns an IsNotFound error when the configuration has never been set.
func (c *FlashBladeClient) GetDirectoryServiceManagement(ctx context.Context, name string) (*DirectoryService, error) {
	return getOneByName[DirectoryService](c, ctx, "/directory-services?names="+url.QueryEscape(name), "directory_service", name)
}

// PatchDirectoryServiceManagement updates the management directory service configuration.
// Only non-nil pointer fields in body are sent (PATCH semantics). See DirectoryServicePatch
// for the **NamedReference discipline used on ca_certificate and ca_certificate_group.
func (c *FlashBladeClient) PatchDirectoryServiceManagement(ctx context.Context, name string, body DirectoryServicePatch) (*DirectoryService, error) {
	return patchOne[DirectoryServicePatch, DirectoryService](c, ctx, "/directory-services?names="+url.QueryEscape(name), body, "PatchDirectoryServiceManagement")
}
