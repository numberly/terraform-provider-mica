package client

import (
	"context"
	"fmt"
	"net/url"
)

// GetS3ExportPolicy retrieves an S3 export policy by name.
// Returns an IsNotFound error if the policy does not exist.
func (c *FlashBladeClient) GetS3ExportPolicy(ctx context.Context, name string) (*S3ExportPolicy, error) {
	return getOneByName[S3ExportPolicy](c, ctx, "/s3-export-policies?names="+url.QueryEscape(name), "S3 export policy", name)
}

// PostS3ExportPolicy creates a new S3 export policy.
// The name is passed as a query parameter; optional fields are in the body.
func (c *FlashBladeClient) PostS3ExportPolicy(ctx context.Context, name string, body S3ExportPolicyPost) (*S3ExportPolicy, error) {
	path := "/s3-export-policies?names=" + url.QueryEscape(name)
	var resp ListResponse[S3ExportPolicy]
	if err := c.post(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PostS3ExportPolicy: empty response from server")
	}
	return &resp.Items[0], nil
}

// PatchS3ExportPolicy updates an existing S3 export policy identified by name.
// When renaming (body.Name != nil), the OLD name must be passed as the name argument.
func (c *FlashBladeClient) PatchS3ExportPolicy(ctx context.Context, name string, body S3ExportPolicyPatch) (*S3ExportPolicy, error) {
	path := "/s3-export-policies?names=" + url.QueryEscape(name)
	var resp ListResponse[S3ExportPolicy]
	if err := c.patch(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PatchS3ExportPolicy: empty response from server")
	}
	return &resp.Items[0], nil
}

// DeleteS3ExportPolicy permanently deletes an S3 export policy.
func (c *FlashBladeClient) DeleteS3ExportPolicy(ctx context.Context, name string) error {
	path := "/s3-export-policies?names=" + url.QueryEscape(name)
	return c.delete(ctx, path)
}

// ListS3ExportPolicyRules returns all rules for the given S3 export policy.
// It automatically follows continuation_token pagination to collect all results.
func (c *FlashBladeClient) ListS3ExportPolicyRules(ctx context.Context, policyName string) ([]S3ExportPolicyRule, error) {
	params := url.Values{}
	params.Set("policy_names", policyName)

	var all []S3ExportPolicyRule
	for {
		path := "/s3-export-policies/rules?" + params.Encode()
		var resp ListResponse[S3ExportPolicyRule]
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

// GetS3ExportPolicyRuleByIndex retrieves an S3 export policy rule by its index within the policy.
// Synthesizes a 404 APIError if no rule with the given index is found.
func (c *FlashBladeClient) GetS3ExportPolicyRuleByIndex(ctx context.Context, policyName string, index int) (*S3ExportPolicyRule, error) {
	rules, err := c.ListS3ExportPolicyRules(ctx, policyName)
	if err != nil {
		return nil, err
	}
	for i := range rules {
		if rules[i].Index == index {
			return &rules[i], nil
		}
	}
	return nil, &APIError{StatusCode: 404, Message: fmt.Sprintf("S3 export policy rule at index %d not found in policy %q", index, policyName)}
}

// GetS3ExportPolicyRuleByName retrieves an S3 export policy rule by name within a policy.
// Synthesizes a 404 APIError if the rule does not exist.
func (c *FlashBladeClient) GetS3ExportPolicyRuleByName(ctx context.Context, policyName, ruleName string) (*S3ExportPolicyRule, error) {
	return getOneByName[S3ExportPolicyRule](c, ctx, "/s3-export-policies/rules?policy_names="+url.QueryEscape(policyName)+"&names="+url.QueryEscape(ruleName), "S3 export policy rule", ruleName)
}

// PostS3ExportPolicyRule creates a new rule in an S3 export policy.
// The policy is identified via the policy_names query parameter only.
func (c *FlashBladeClient) PostS3ExportPolicyRule(ctx context.Context, policyName, ruleName string, body S3ExportPolicyRulePost) (*S3ExportPolicyRule, error) {
	path := "/s3-export-policies/rules?policy_names=" + url.QueryEscape(policyName) + "&names=" + url.QueryEscape(ruleName)
	var resp ListResponse[S3ExportPolicyRule]
	if err := c.post(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PostS3ExportPolicyRule: empty response from server")
	}
	return &resp.Items[0], nil
}

// PatchS3ExportPolicyRule updates an existing S3 export policy rule.
func (c *FlashBladeClient) PatchS3ExportPolicyRule(ctx context.Context, policyName, ruleName string, body S3ExportPolicyRulePatch) (*S3ExportPolicyRule, error) {
	path := "/s3-export-policies/rules?names=" + url.QueryEscape(ruleName) + "&policy_names=" + url.QueryEscape(policyName)
	var resp ListResponse[S3ExportPolicyRule]
	if err := c.patch(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PatchS3ExportPolicyRule: empty response from server")
	}
	return &resp.Items[0], nil
}

// DeleteS3ExportPolicyRule deletes an S3 export policy rule by name.
func (c *FlashBladeClient) DeleteS3ExportPolicyRule(ctx context.Context, policyName, ruleName string) error {
	path := "/s3-export-policies/rules?names=" + url.QueryEscape(ruleName) + "&policy_names=" + url.QueryEscape(policyName)
	return c.delete(ctx, path)
}
