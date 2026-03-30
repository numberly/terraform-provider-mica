package client

import (
	"context"
	"fmt"
	"net/url"
)

// GetBucketAuditFilter retrieves a bucket audit filter by name.
// Returns an IsNotFound error if no audit filter exists with the given name.
func (c *FlashBladeClient) GetBucketAuditFilter(ctx context.Context, filterName string) (*BucketAuditFilter, error) {
	path := "/buckets/audit-filters?names=" + url.QueryEscape(filterName)
	var resp ListResponse[BucketAuditFilter]
	if err := c.get(ctx, path, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, &APIError{StatusCode: 404, Message: fmt.Sprintf("bucket audit filter %q not found", filterName)}
	}
	return &resp.Items[0], nil
}

// GetBucketAuditFilterByBucket retrieves a bucket audit filter by bucket name.
// Used by the data source where the user queries by bucket name.
func (c *FlashBladeClient) GetBucketAuditFilterByBucket(ctx context.Context, bucketName string) (*BucketAuditFilter, error) {
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

// PostBucketAuditFilter creates a bucket audit filter.
// The API requires both ?names=<filter_name> and ?bucket_names=<bucket_name> on POST.
func (c *FlashBladeClient) PostBucketAuditFilter(ctx context.Context, filterName string, bucketName string, body BucketAuditFilterPost) (*BucketAuditFilter, error) {
	path := "/buckets/audit-filters?names=" + url.QueryEscape(filterName) + "&bucket_names=" + url.QueryEscape(bucketName)
	var resp ListResponse[BucketAuditFilter]
	if err := c.post(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PostBucketAuditFilter: empty response from server")
	}
	return &resp.Items[0], nil
}

// PatchBucketAuditFilter updates a bucket audit filter by name.
func (c *FlashBladeClient) PatchBucketAuditFilter(ctx context.Context, filterName string, body BucketAuditFilterPatch) (*BucketAuditFilter, error) {
	path := "/buckets/audit-filters?names=" + url.QueryEscape(filterName)
	var resp ListResponse[BucketAuditFilter]
	if err := c.patch(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PatchBucketAuditFilter: empty response from server")
	}
	return &resp.Items[0], nil
}

// DeleteBucketAuditFilter deletes a bucket audit filter by name.
func (c *FlashBladeClient) DeleteBucketAuditFilter(ctx context.Context, filterName string) error {
	path := "/buckets/audit-filters?names=" + url.QueryEscape(filterName)
	return c.delete(ctx, path)
}
