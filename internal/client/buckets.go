package client

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// ListBucketsOpts contains optional query parameters for ListBuckets.
type ListBucketsOpts struct {
	// Names filters results to specific bucket names.
	Names []string
	// AccountNames filters results to buckets belonging to specific accounts.
	AccountNames []string
	// Filter is a free-form filter expression.
	Filter string
	// Destroyed, when set, filters by soft-deleted state.
	Destroyed *bool
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
	if opts.Destroyed != nil {
		if *opts.Destroyed {
			params.Set("destroyed", "true")
		} else {
			params.Set("destroyed", "false")
		}
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

// PostBucket creates a new bucket with the given name.
// The bucket name is passed as a query parameter (?names=), matching FlashBlade API semantics.
func (c *FlashBladeClient) PostBucket(ctx context.Context, name string, body BucketPost) (*Bucket, error) {
	path := "/buckets?names=" + url.QueryEscape(name)
	var resp ListResponse[Bucket]
	if err := c.post(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PostBucket: empty response from server")
	}
	return &resp.Items[0], nil
}

// PatchBucket updates an existing bucket identified by its ID.
// Only non-nil pointer fields in body are sent (PATCH semantics).
// Uses ID (not name) for stability across the resource lifecycle.
func (c *FlashBladeClient) PatchBucket(ctx context.Context, id string, body BucketPatch) (*Bucket, error) {
	path := "/buckets?ids=" + url.QueryEscape(id)
	var resp ListResponse[Bucket]
	if err := c.patch(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PatchBucket: empty response from server")
	}
	return &resp.Items[0], nil
}

// DeleteBucket eradicates a soft-deleted bucket identified by its ID.
// The bucket must already be soft-deleted (destroyed=true) before calling this.
func (c *FlashBladeClient) DeleteBucket(ctx context.Context, id string) error {
	path := "/buckets?ids=" + url.QueryEscape(id)
	return c.delete(ctx, path)
}

// PollBucketUntilEradicated polls GET /buckets?names={name}&destroyed=true until the
// bucket is fully eradicated (empty items response). Respects context deadline.
// The caller should provide a context with an appropriate timeout.
func (c *FlashBladeClient) PollBucketUntilEradicated(ctx context.Context, name string) error {
	for {
		// Check context before polling.
		select {
		case <-ctx.Done():
			return fmt.Errorf("PollBucketUntilEradicated: context cancelled while waiting for %q to eradicate: %w", name, ctx.Err())
		default:
		}

		path := "/buckets?names=" + url.QueryEscape(name) + "&destroyed=true"
		var resp ListResponse[Bucket]
		err := c.get(ctx, path, &resp)
		if err != nil {
			if IsNotFound(err) {
				return nil
			}
			return fmt.Errorf("PollBucketUntilEradicated: GET error: %w", err)
		}

		if len(resp.Items) == 0 {
			// Eradication complete.
			return nil
		}

		// Still present — wait before retrying.
		select {
		case <-ctx.Done():
			return fmt.Errorf("PollBucketUntilEradicated: context cancelled while waiting for %q to eradicate: %w", name, ctx.Err())
		case <-time.After(2 * time.Second):
			// Continue polling.
		}
	}
}
