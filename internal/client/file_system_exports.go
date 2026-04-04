package client

import (
	"context"
	"net/url"
)

// GetFileSystemExport retrieves a file system export by its combined name (e.g. "fs/export").
// Returns an IsNotFound error if the export does not exist.
func (c *FlashBladeClient) GetFileSystemExport(ctx context.Context, name string) (*FileSystemExport, error) {
	return getOneByName[FileSystemExport](c, ctx, "/file-system-exports?names="+url.QueryEscape(name), "file system export", name)
}

// PostFileSystemExport creates a new file system export.
// The memberName (file system name) is passed as ?member_names= query parameter.
// The policyName (NFS export policy name) is passed as ?policy_names= query parameter.
func (c *FlashBladeClient) PostFileSystemExport(ctx context.Context, memberName, policyName string, body FileSystemExportPost) (*FileSystemExport, error) {
	return postOne[FileSystemExportPost, FileSystemExport](c, ctx, "/file-system-exports?member_names="+url.QueryEscape(memberName)+"&policy_names="+url.QueryEscape(policyName), body, "PostFileSystemExport")
}

// PatchFileSystemExport updates an existing file system export identified by its ID.
// Only non-nil pointer fields in body are sent (PATCH semantics).
func (c *FlashBladeClient) PatchFileSystemExport(ctx context.Context, id string, body FileSystemExportPatch) (*FileSystemExport, error) {
	return patchOne[FileSystemExportPatch, FileSystemExport](c, ctx, "/file-system-exports?ids="+url.QueryEscape(id), body, "PatchFileSystemExport")
}

// DeleteFileSystemExport deletes a file system export by member name and export name.
func (c *FlashBladeClient) DeleteFileSystemExport(ctx context.Context, memberName, exportName string) error {
	path := "/file-system-exports?member_names=" + url.QueryEscape(memberName) + "&names=" + url.QueryEscape(exportName)
	return c.delete(ctx, path)
}
