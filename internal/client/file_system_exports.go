package client

import (
	"context"
	"fmt"
	"net/url"
)

// GetFileSystemExport retrieves a file system export by its combined name (e.g. "fs/export").
// Returns an IsNotFound error if the export does not exist.
func (c *FlashBladeClient) GetFileSystemExport(ctx context.Context, name string) (*FileSystemExport, error) {
	path := "/file-system-exports?names=" + url.QueryEscape(name)
	var resp ListResponse[FileSystemExport]
	if err := c.get(ctx, path, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, &APIError{StatusCode: 404, Message: fmt.Sprintf("file system export %q not found", name)}
	}
	return &resp.Items[0], nil
}

// PostFileSystemExport creates a new file system export.
// The memberName (file system name) is passed as ?member_names= query parameter.
// The policyName (NFS export policy name) is passed as ?policy_names= query parameter.
func (c *FlashBladeClient) PostFileSystemExport(ctx context.Context, memberName, policyName string, body FileSystemExportPost) (*FileSystemExport, error) {
	path := "/file-system-exports?member_names=" + url.QueryEscape(memberName) + "&policy_names=" + url.QueryEscape(policyName)
	var resp ListResponse[FileSystemExport]
	if err := c.post(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PostFileSystemExport: empty response from server")
	}
	return &resp.Items[0], nil
}

// PatchFileSystemExport updates an existing file system export identified by its ID.
// Only non-nil pointer fields in body are sent (PATCH semantics).
func (c *FlashBladeClient) PatchFileSystemExport(ctx context.Context, id string, body FileSystemExportPatch) (*FileSystemExport, error) {
	path := "/file-system-exports?ids=" + url.QueryEscape(id)
	var resp ListResponse[FileSystemExport]
	if err := c.patch(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PatchFileSystemExport: empty response from server")
	}
	return &resp.Items[0], nil
}

// DeleteFileSystemExport deletes a file system export by member name and export name.
func (c *FlashBladeClient) DeleteFileSystemExport(ctx context.Context, memberName, exportName string) error {
	path := "/file-system-exports?member_names=" + url.QueryEscape(memberName) + "&names=" + url.QueryEscape(exportName)
	return c.delete(ctx, path)
}
