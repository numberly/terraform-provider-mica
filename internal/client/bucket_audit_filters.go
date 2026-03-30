package client

import (
	"context"
	"fmt"
	"net/url"
)

// GetBucketAuditFilter retrieves a bucket audit filter by bucket name.
// Returns an IsNotFound error if no audit filter exists for the bucket.
func (c *FlashBladeClient) GetBucketAuditFilter(ctx context.Context, bucketName string) (*BucketAuditFilter, error) {
	path := "/buckets/audit-filters?bucket_names=" + url.QueryEscape(bucketName)
	var resp ListResponse[BucketAuditFilter]
	if err := c.get(ctx, path, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, &APIError{StatusCode: 404, Message: fmt.Sprintf("bucket audit filter for bucket %q not found", bucketName)}
	}
	return &resp.Items[0], nil
}

// PostBucketAuditFilter creates a bucket audit filter for the given bucket.
func (c *FlashBladeClient) PostBucketAuditFilter(ctx context.Context, bucketName string, body BucketAuditFilterPost) (*BucketAuditFilter, error) {
	path := "/buckets/audit-filters?bucket_names=" + url.QueryEscape(bucketName)
	var resp ListResponse[BucketAuditFilter]
	if err := c.post(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PostBucketAuditFilter: empty response from server")
	}
	return &resp.Items[0], nil
}

// PatchBucketAuditFilter updates a bucket audit filter for the given bucket.
func (c *FlashBladeClient) PatchBucketAuditFilter(ctx context.Context, bucketName string, body BucketAuditFilterPatch) (*BucketAuditFilter, error) {
	path := "/buckets/audit-filters?bucket_names=" + url.QueryEscape(bucketName)
	var resp ListResponse[BucketAuditFilter]
	if err := c.patch(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PatchBucketAuditFilter: empty response from server")
	}
	return &resp.Items[0], nil
}

// DeleteBucketAuditFilter deletes the bucket audit filter for the given bucket.
func (c *FlashBladeClient) DeleteBucketAuditFilter(ctx context.Context, bucketName string) error {
	path := "/buckets/audit-filters?bucket_names=" + url.QueryEscape(bucketName)
	return c.delete(ctx, path)
}
