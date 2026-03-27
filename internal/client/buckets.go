package client

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// ListBucketsOpts contains optional query parameters for ListBuckets.
type ListBucketsOpts struct {
	// Names filters results to specific bucket names.
	Names []string
	// AccountNames filters results to buckets belonging to specific accounts.
	AccountNames []string
	// Filter is a free-form filter expression.
	Filter string
}

// ListBuckets returns all buckets matching the optional opts filters.
func (c *FlashBladeClient) ListBuckets(ctx context.Context, opts ListBucketsOpts) ([]Bucket, error) {
	params := url.Values{}
	if len(opts.Names) > 0 {
		params.Set("names", strings.Join(opts.Names, ","))
	}
	if len(opts.AccountNames) > 0 {
		params.Set("account_names", strings.Join(opts.AccountNames, ","))
	}
	if opts.Filter != "" {
		params.Set("filter", opts.Filter)
	}

	path := "/buckets"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	var resp ListResponse[Bucket]
	if err := c.get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

// GetBucket retrieves a bucket by name.
// Returns an IsNotFound error if the bucket does not exist.
func (c *FlashBladeClient) GetBucket(ctx context.Context, name string) (*Bucket, error) {
	path := "/buckets?names=" + url.QueryEscape(name)
	var resp ListResponse[Bucket]
	if err := c.get(ctx, path, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, &APIError{StatusCode: 404, Message: fmt.Sprintf("bucket %q not found", name)}
	}
	return &resp.Items[0], nil
}
