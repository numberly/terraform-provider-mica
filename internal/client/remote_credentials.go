package client

import (
	"context"
	"net/url"
)

// GetRemoteCredentials retrieves an object store remote credentials entry by name.
// Returns an IsNotFound error if the entry does not exist.
func (c *FlashBladeClient) GetRemoteCredentials(ctx context.Context, name string) (*ObjectStoreRemoteCredentials, error) {
	return getOneByName[ObjectStoreRemoteCredentials](c, ctx, "/object-store-remote-credentials?names="+url.QueryEscape(name), "remote credentials", name)
}

// ListRemoteCredentials returns all object store remote credentials.
// It automatically follows continuation_token pagination to collect all results.
func (c *FlashBladeClient) ListRemoteCredentials(ctx context.Context) ([]ObjectStoreRemoteCredentials, error) {
	params := url.Values{}
	var all []ObjectStoreRemoteCredentials
	for {
		path := "/object-store-remote-credentials"
		if len(params) > 0 {
			path += "?" + params.Encode()
		}

		var resp ListResponse[ObjectStoreRemoteCredentials]
		if err := c.get(ctx, path, &resp); err != nil {
			return nil, err
		}
		all = append(all, resp.Items...)
		if resp.ContinuationToken == "" {
			break
		}
		params.Set("continuation_token", resp.ContinuationToken)
	}
	return all, nil
}

// PostRemoteCredentials creates a new remote credentials entry.
// The name is passed via ?names=. Either targetName or remoteName must be non-empty (not both).
// When targetName is non-empty, routes to ?target_names=; otherwise routes to ?remote_names=.
func (c *FlashBladeClient) PostRemoteCredentials(ctx context.Context, name string, remoteName string, targetName string, body ObjectStoreRemoteCredentialsPost) (*ObjectStoreRemoteCredentials, error) {
	path := "/object-store-remote-credentials?names=" + url.QueryEscape(name)
	if targetName != "" {
		path += "&target_names=" + url.QueryEscape(targetName)
	} else if remoteName != "" {
		path += "&remote_names=" + url.QueryEscape(remoteName)
	}
	return postOne[ObjectStoreRemoteCredentialsPost, ObjectStoreRemoteCredentials](c, ctx, path, body, "PostRemoteCredentials")
}

// PatchRemoteCredentials updates an existing remote credentials entry identified by name.
// Only non-nil pointer fields in body are sent (PATCH semantics).
func (c *FlashBladeClient) PatchRemoteCredentials(ctx context.Context, name string, body ObjectStoreRemoteCredentialsPatch) (*ObjectStoreRemoteCredentials, error) {
	return patchOne[ObjectStoreRemoteCredentialsPatch, ObjectStoreRemoteCredentials](c, ctx, "/object-store-remote-credentials?names="+url.QueryEscape(name), body, "PatchRemoteCredentials")
}

// DeleteRemoteCredentials permanently deletes a remote credentials entry by name.
func (c *FlashBladeClient) DeleteRemoteCredentials(ctx context.Context, name string) error {
	path := "/object-store-remote-credentials?names=" + url.QueryEscape(name)
	return c.delete(ctx, path)
}
