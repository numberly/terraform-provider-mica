package client

import (
	"context"
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
	return postOne[ObjectStoreAccountExportPost, ObjectStoreAccountExport](c, ctx, "/object-store-account-exports?member_names="+url.QueryEscape(memberName)+"&policy_names="+url.QueryEscape(policyName), body, "PostObjectStoreAccountExport")
}

// PatchObjectStoreAccountExport updates an existing object store account export identified by its ID.
// Only non-nil pointer fields in body are sent (PATCH semantics).
func (c *FlashBladeClient) PatchObjectStoreAccountExport(ctx context.Context, id string, body ObjectStoreAccountExportPatch) (*ObjectStoreAccountExport, error) {
	return patchOne[ObjectStoreAccountExportPatch, ObjectStoreAccountExport](c, ctx, "/object-store-account-exports?ids="+url.QueryEscape(id), body, "PatchObjectStoreAccountExport")
}

// DeleteObjectStoreAccountExport deletes an object store account export by member name and export name.
func (c *FlashBladeClient) DeleteObjectStoreAccountExport(ctx context.Context, memberName, exportName string) error {
	path := "/object-store-account-exports?member_names=" + url.QueryEscape(memberName) + "&names=" + url.QueryEscape(exportName)
	return c.delete(ctx, path)
}
