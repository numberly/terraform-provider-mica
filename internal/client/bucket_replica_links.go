package client

import (
	"context"
	"fmt"
	"net/url"
)

// GetBucketReplicaLink retrieves a bucket replica link by local and remote bucket names.
// Returns an IsNotFound error if the link does not exist.
func (c *FlashBladeClient) GetBucketReplicaLink(ctx context.Context, localBucketName string, remoteBucketName string) (*BucketReplicaLink, error) {
	return getOneByName[BucketReplicaLink](c, ctx, "/bucket-replica-links?local_bucket_names="+url.QueryEscape(localBucketName)+"&remote_bucket_names="+url.QueryEscape(remoteBucketName), "bucket replica link", localBucketName+"->"+remoteBucketName)
}

// GetBucketReplicaLinkByID retrieves a bucket replica link by its unique ID.
// Returns an IsNotFound error if the link does not exist.
func (c *FlashBladeClient) GetBucketReplicaLinkByID(ctx context.Context, id string) (*BucketReplicaLink, error) {
	return getOneByName[BucketReplicaLink](c, ctx, "/bucket-replica-links?ids="+url.QueryEscape(id), "bucket replica link", id)
}

// ListBucketReplicaLinks returns all bucket replica links.
// It automatically follows continuation_token pagination to collect all results.
func (c *FlashBladeClient) ListBucketReplicaLinks(ctx context.Context) ([]BucketReplicaLink, error) {
	params := url.Values{}
	var all []BucketReplicaLink
	for {
		path := "/bucket-replica-links"
		if len(params) > 0 {
			path += "?" + params.Encode()
		}

		var resp ListResponse[BucketReplicaLink]
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

// PostBucketReplicaLink creates a new bucket replica link.
// Local bucket, remote bucket, and remote credentials are passed as query parameters.
// If remoteCredentialsName is empty, the remote_credentials_names param is omitted (FB-to-FB case).
func (c *FlashBladeClient) PostBucketReplicaLink(ctx context.Context, localBucketName string, remoteBucketName string, remoteCredentialsName string, body BucketReplicaLinkPost) (*BucketReplicaLink, error) {
	path := "/bucket-replica-links?local_bucket_names=" + url.QueryEscape(localBucketName) + "&remote_bucket_names=" + url.QueryEscape(remoteBucketName)
	if remoteCredentialsName != "" {
		path += "&remote_credentials_names=" + url.QueryEscape(remoteCredentialsName)
	}
	var resp ListResponse[BucketReplicaLink]
	if err := c.post(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PostBucketReplicaLink: empty response from server")
	}
	return &resp.Items[0], nil
}

// PatchBucketReplicaLink updates an existing bucket replica link identified by its ID.
// Uses ID for PATCH stability (same pattern as PatchBucket).
func (c *FlashBladeClient) PatchBucketReplicaLink(ctx context.Context, id string, body BucketReplicaLinkPatch) (*BucketReplicaLink, error) {
	path := "/bucket-replica-links?ids=" + url.QueryEscape(id)
	var resp ListResponse[BucketReplicaLink]
	if err := c.patch(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PatchBucketReplicaLink: empty response from server")
	}
	return &resp.Items[0], nil
}

// DeleteBucketReplicaLink permanently deletes a bucket replica link by its ID.
func (c *FlashBladeClient) DeleteBucketReplicaLink(ctx context.Context, id string) error {
	path := "/bucket-replica-links?ids=" + url.QueryEscape(id)
	return c.delete(ctx, path)
}
