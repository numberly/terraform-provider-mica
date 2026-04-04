package client

import (
	"context"
	"fmt"
	"net/url"
)

// GetBucketAccessPolicy retrieves a bucket access policy by bucket name.
// Returns an IsNotFound error if no policy exists for the bucket.
func (c *FlashBladeClient) GetBucketAccessPolicy(ctx context.Context, bucketName string) (*BucketAccessPolicy, error) {
	return getOneByName[BucketAccessPolicy](c, ctx, "/buckets/bucket-access-policies?bucket_names="+url.QueryEscape(bucketName), "bucket access policy", bucketName)
}

// PostBucketAccessPolicy creates a bucket access policy for the given bucket.
func (c *FlashBladeClient) PostBucketAccessPolicy(ctx context.Context, bucketName string, body BucketAccessPolicyPost) (*BucketAccessPolicy, error) {
	path := "/buckets/bucket-access-policies?bucket_names=" + url.QueryEscape(bucketName)
	var resp ListResponse[BucketAccessPolicy]
	if err := c.post(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PostBucketAccessPolicy: empty response from server")
	}
	return &resp.Items[0], nil
}

// DeleteBucketAccessPolicy deletes the bucket access policy for the given bucket.
func (c *FlashBladeClient) DeleteBucketAccessPolicy(ctx context.Context, bucketName string) error {
	path := "/buckets/bucket-access-policies?bucket_names=" + url.QueryEscape(bucketName)
	return c.delete(ctx, path)
}

// ListBucketAccessPolicyRules returns all rules for a bucket's access policy.
// It automatically follows continuation_token pagination to collect all results.
func (c *FlashBladeClient) ListBucketAccessPolicyRules(ctx context.Context, bucketName string) ([]BucketAccessPolicyRule, error) {
	params := url.Values{}
	params.Set("bucket_names", bucketName)
	var all []BucketAccessPolicyRule
	for {
		path := "/buckets/bucket-access-policies/rules?" + params.Encode()

		var resp ListResponse[BucketAccessPolicyRule]
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

// GetBucketAccessPolicyRule retrieves a specific rule by bucket name and rule name.
// Returns an IsNotFound error if the rule does not exist.
func (c *FlashBladeClient) GetBucketAccessPolicyRule(ctx context.Context, bucketName string, ruleName string) (*BucketAccessPolicyRule, error) {
	return getOneByName[BucketAccessPolicyRule](c, ctx, "/buckets/bucket-access-policies/rules?bucket_names="+url.QueryEscape(bucketName)+"&names="+url.QueryEscape(ruleName), "bucket access policy rule", ruleName)
}

// PostBucketAccessPolicyRule creates a new rule on the bucket's access policy.
// The API requires ?names= with the bucket name on POST.
func (c *FlashBladeClient) PostBucketAccessPolicyRule(ctx context.Context, bucketName string, body BucketAccessPolicyRulePost) (*BucketAccessPolicyRule, error) {
	path := "/buckets/bucket-access-policies/rules?names=" + url.QueryEscape(bucketName) + "&bucket_names=" + url.QueryEscape(bucketName)
	var resp ListResponse[BucketAccessPolicyRule]
	if err := c.post(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PostBucketAccessPolicyRule: empty response from server")
	}
	return &resp.Items[0], nil
}

// DeleteBucketAccessPolicyRule deletes a specific rule from the bucket's access policy.
func (c *FlashBladeClient) DeleteBucketAccessPolicyRule(ctx context.Context, bucketName string, ruleName string) error {
	path := "/buckets/bucket-access-policies/rules?bucket_names=" + url.QueryEscape(bucketName) + "&names=" + url.QueryEscape(ruleName)
	return c.delete(ctx, path)
}
