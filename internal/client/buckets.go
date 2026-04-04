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
	// Destroyed, when set, filters by soft-deleted state.
	Destroyed *bool
}

// ListBuckets returns all buckets matching the optional opts filters.
// It automatically follows continuation_token pagination to collect all results.
func (c *FlashBladeClient) ListBuckets(ctx context.Context, opts ListBucketsOpts) ([]Bucket, error) {
	params := url.Values{}
	if len(opts.Names) > 0 {
		params.Set("names", strings.Join(opts.Names, ","))
	}
	if len(opts.AccountNames) > 0 {
		// FlashBlade API does not have an account_names query param on /buckets.
		// Use filter expression: account.name='name1' or account.name='name2'
		var parts []string
		for _, acct := range opts.AccountNames {
			parts = append(parts, fmt.Sprintf("account.name='%s'", acct))
		}
		acctFilter := strings.Join(parts, " or ")
		if opts.Filter != "" {
			opts.Filter = "(" + acctFilter + ") and (" + opts.Filter + ")"
		} else {
			opts.Filter = acctFilter
		}
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

	var all []Bucket
	for {
		path := "/buckets"
		if len(params) > 0 {
			path += "?" + params.Encode()
		}

		var resp ListResponse[Bucket]
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

// GetBucket retrieves a bucket by name.
// Returns an IsNotFound error if the bucket does not exist.
func (c *FlashBladeClient) GetBucket(ctx context.Context, name string) (*Bucket, error) {
	return getOneByName[Bucket](c, ctx, "/buckets?names="+url.QueryEscape(name), "bucket", name)
}

// PostBucket creates a new bucket with the given name.
// The bucket name is passed as a query parameter (?names=), matching FlashBlade API semantics.
func (c *FlashBladeClient) PostBucket(ctx context.Context, name string, body BucketPost) (*Bucket, error) {
	return postOne[BucketPost, Bucket](c, ctx, "/buckets?names="+url.QueryEscape(name), body, "PostBucket")
}

// PatchBucket updates an existing bucket identified by its ID.
// Only non-nil pointer fields in body are sent (PATCH semantics).
// Uses ID (not name) for stability across the resource lifecycle.
func (c *FlashBladeClient) PatchBucket(ctx context.Context, id string, body BucketPatch) (*Bucket, error) {
	return patchOne[BucketPatch, Bucket](c, ctx, "/buckets?ids="+url.QueryEscape(id), body, "PatchBucket")
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
	return pollUntilGone[Bucket](c, ctx, "/buckets", "bucket", name)
}
