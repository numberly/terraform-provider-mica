package client

import (
	"context"
	"fmt"
	"net/url"
)

// GetLifecycleRule retrieves a lifecycle rule by bucket name and rule ID.
// Returns an IsNotFound error if the rule does not exist.
func (c *FlashBladeClient) GetLifecycleRule(ctx context.Context, bucketName string, ruleID string) (*LifecycleRule, error) {
	path := "/lifecycle-rules?bucket_names=" + url.QueryEscape(bucketName)
	var resp ListResponse[LifecycleRule]
	if err := c.get(ctx, path, &resp); err != nil {
		return nil, err
	}
	for i := range resp.Items {
		if resp.Items[i].RuleID == ruleID {
			return &resp.Items[i], nil
		}
	}
	return nil, &APIError{StatusCode: 404, Message: fmt.Sprintf("lifecycle rule %q on bucket %q not found", ruleID, bucketName)}
}

// ListLifecycleRulesByBucket returns all lifecycle rules for a given bucket.
// It automatically follows continuation_token pagination to collect all results.
func (c *FlashBladeClient) ListLifecycleRulesByBucket(ctx context.Context, bucketName string) ([]LifecycleRule, error) {
	params := url.Values{}
	params.Set("bucket_names", bucketName)
	return listAll[LifecycleRule](c, ctx, "/lifecycle-rules", params)
}

// PostLifecycleRule creates a new lifecycle rule.
// If confirmDate is true, the confirm_date query parameter is added.
func (c *FlashBladeClient) PostLifecycleRule(ctx context.Context, body LifecycleRulePost, confirmDate bool) (*LifecycleRule, error) {
	path := "/lifecycle-rules"
	if confirmDate {
		path += "?confirm_date=true"
	}
	return postOne[LifecycleRulePost, LifecycleRule](c, ctx, path, body, "PostLifecycleRule")
}

// PatchLifecycleRule updates an existing lifecycle rule identified by bucket name and rule ID.
// The API identifies the rule by composite name "bucketName/ruleID".
// If confirmDate is true, the confirm_date query parameter is added.
func (c *FlashBladeClient) PatchLifecycleRule(ctx context.Context, bucketName string, ruleID string, body LifecycleRulePatch, confirmDate bool) (*LifecycleRule, error) {
	compositeName := bucketName + "/" + ruleID
	path := "/lifecycle-rules?names=" + url.QueryEscape(compositeName)
	if confirmDate {
		path += "&confirm_date=true"
	}
	return patchOne[LifecycleRulePatch, LifecycleRule](c, ctx, path, body, "PatchLifecycleRule")
}

// DeleteLifecycleRule permanently deletes a lifecycle rule identified by bucket name and rule ID.
// The API identifies the rule by composite name "bucketName/ruleID".
func (c *FlashBladeClient) DeleteLifecycleRule(ctx context.Context, bucketName string, ruleID string) error {
	compositeName := bucketName + "/" + ruleID
	path := "/lifecycle-rules?names=" + url.QueryEscape(compositeName)
	return c.delete(ctx, path)
}
