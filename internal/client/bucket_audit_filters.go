package client

import (
	"context"
	"net/url"
)

// GetBucketAuditFilter retrieves a bucket audit filter by filter name and bucket name.
func (c *FlashBladeClient) GetBucketAuditFilter(ctx context.Context, filterName string, bucketName string) (*BucketAuditFilter, error) {
	return getOneByName[BucketAuditFilter](c, ctx, "/buckets/audit-filters?names="+url.QueryEscape(filterName)+"&bucket_names="+url.QueryEscape(bucketName), "bucket audit filter", filterName)
}

// GetBucketAuditFilterByBucket retrieves a bucket audit filter by bucket name only.
func (c *FlashBladeClient) GetBucketAuditFilterByBucket(ctx context.Context, bucketName string) (*BucketAuditFilter, error) {
	return getOneByName[BucketAuditFilter](c, ctx, "/buckets/audit-filters?bucket_names="+url.QueryEscape(bucketName), "bucket audit filter", bucketName)
}

// PostBucketAuditFilter creates a bucket audit filter.
func (c *FlashBladeClient) PostBucketAuditFilter(ctx context.Context, filterName string, bucketName string, body BucketAuditFilterPost) (*BucketAuditFilter, error) {
	return postOne[BucketAuditFilterPost, BucketAuditFilter](c, ctx, "/buckets/audit-filters?names="+url.QueryEscape(filterName)+"&bucket_names="+url.QueryEscape(bucketName), body, "PostBucketAuditFilter")
}

// PatchBucketAuditFilter updates a bucket audit filter.
func (c *FlashBladeClient) PatchBucketAuditFilter(ctx context.Context, filterName string, bucketName string, body BucketAuditFilterPatch) (*BucketAuditFilter, error) {
	return patchOne[BucketAuditFilterPatch, BucketAuditFilter](c, ctx, "/buckets/audit-filters?names="+url.QueryEscape(filterName)+"&bucket_names="+url.QueryEscape(bucketName), body, "PatchBucketAuditFilter")
}

// DeleteBucketAuditFilter deletes a bucket audit filter.
func (c *FlashBladeClient) DeleteBucketAuditFilter(ctx context.Context, filterName string, bucketName string) error {
	path := "/buckets/audit-filters?names=" + url.QueryEscape(filterName) + "&bucket_names=" + url.QueryEscape(bucketName)
	return c.delete(ctx, path)
}
