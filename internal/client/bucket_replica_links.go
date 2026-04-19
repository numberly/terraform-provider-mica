package client

import (
	"context"
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
	return listAll[BucketReplicaLink](c, ctx, "/bucket-replica-links", params)
}

// PostBucketReplicaLink creates a new bucket replica link.
// Local bucket, remote bucket, and remote credentials are passed as query parameters.
// If remoteCredentialsName is empty, the remote_credentials_names param is omitted (FB-to-FB case).
func (c *FlashBladeClient) PostBucketReplicaLink(ctx context.Context, localBucketName string, remoteBucketName string, remoteCredentialsName string, body BucketReplicaLinkPost) (*BucketReplicaLink, error) {
	path := "/bucket-replica-links?local_bucket_names=" + url.QueryEscape(localBucketName) + "&remote_bucket_names=" + url.QueryEscape(remoteBucketName)
	if remoteCredentialsName != "" {
		path += "&remote_credentials_names=" + url.QueryEscape(remoteCredentialsName)
	}
	return postOne[BucketReplicaLinkPost, BucketReplicaLink](c, ctx, path, body, "PostBucketReplicaLink")
}

// PatchBucketReplicaLink updates an existing bucket replica link identified by its ID.
// Uses ID for PATCH stability (same pattern as PatchBucket).
func (c *FlashBladeClient) PatchBucketReplicaLink(ctx context.Context, id string, body BucketReplicaLinkPatch) (*BucketReplicaLink, error) {
	return patchOne[BucketReplicaLinkPatch, BucketReplicaLink](c, ctx, "/bucket-replica-links?ids="+url.QueryEscape(id), body, "PatchBucketReplicaLink")
}

// DeleteBucketReplicaLink permanently deletes a bucket replica link by its ID.
func (c *FlashBladeClient) DeleteBucketReplicaLink(ctx context.Context, id string) error {
	path := "/bucket-replica-links?ids=" + url.QueryEscape(id)
	return c.delete(ctx, path)
}
