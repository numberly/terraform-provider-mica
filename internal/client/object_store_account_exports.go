package client

import (
	"context"
	"fmt"
	"net/url"
)

// GetObjectStoreAccountExport retrieves an object store account export by its combined name (e.g. "account/export").
// Returns an IsNotFound error if the export does not exist.
func (c *FlashBladeClient) GetObjectStoreAccountExport(ctx context.Context, name string) (*ObjectStoreAccountExport, error) {
	return getOneByName[ObjectStoreAccountExport](c, ctx, "/object-store-account-exports?names="+url.QueryEscape(name), "object store account export", name)
}

// PostObjectStoreAccountExport creates a new object store account export.
// The memberName (account name) is passed as ?member_names= query parameter.
// The policyName (S3 export policy name) is passed as ?policy_names= query parameter.
func (c *FlashBladeClient) PostObjectStoreAccountExport(ctx context.Context, memberName, policyName string, body ObjectStoreAccountExportPost) (*ObjectStoreAccountExport, error) {
	path := "/object-store-account-exports?member_names=" + url.QueryEscape(memberName) + "&policy_names=" + url.QueryEscape(policyName)
	var resp ListResponse[ObjectStoreAccountExport]
	if err := c.post(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PostObjectStoreAccountExport: empty response from server")
	}
	return &resp.Items[0], nil
}

// PatchObjectStoreAccountExport updates an existing object store account export identified by its ID.
// Only non-nil pointer fields in body are sent (PATCH semantics).
func (c *FlashBladeClient) PatchObjectStoreAccountExport(ctx context.Context, id string, body ObjectStoreAccountExportPatch) (*ObjectStoreAccountExport, error) {
	path := "/object-store-account-exports?ids=" + url.QueryEscape(id)
	var resp ListResponse[ObjectStoreAccountExport]
	if err := c.patch(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PatchObjectStoreAccountExport: empty response from server")
	}
	return &resp.Items[0], nil
}

// DeleteObjectStoreAccountExport deletes an object store account export by member name and export name.
func (c *FlashBladeClient) DeleteObjectStoreAccountExport(ctx context.Context, memberName, exportName string) error {
	path := "/object-store-account-exports?member_names=" + url.QueryEscape(memberName) + "&names=" + url.QueryEscape(exportName)
	return c.delete(ctx, path)
}
