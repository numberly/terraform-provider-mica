package client

import (
	"context"
	"fmt"
	"net/url"
)

// GetNetworkAccessPolicy retrieves a network access policy by name.
// Returns an IsNotFound error if the policy does not exist.
func (c *FlashBladeClient) GetNetworkAccessPolicy(ctx context.Context, name string) (*NetworkAccessPolicy, error) {
	return getOneByName[NetworkAccessPolicy](c, ctx, "/network-access-policies?names="+url.QueryEscape(name), "network access policy", name)
}

// ListNetworkAccessPolicies returns all network access policies.
// It automatically follows continuation_token pagination to collect all results.
func (c *FlashBladeClient) ListNetworkAccessPolicies(ctx context.Context) ([]NetworkAccessPolicy, error) {
	params := url.Values{}

	var all []NetworkAccessPolicy
	for {
		path := "/network-access-policies"
		if len(params) > 0 {
			path += "?" + params.Encode()
		}

		var resp ListResponse[NetworkAccessPolicy]
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

// PatchNetworkAccessPolicy updates an existing network access policy identified by name.
// Network access policies are singletons — no POST or DELETE is available at the policy level.
func (c *FlashBladeClient) PatchNetworkAccessPolicy(ctx context.Context, name string, body NetworkAccessPolicyPatch) (*NetworkAccessPolicy, error) {
	path := "/network-access-policies?names=" + url.QueryEscape(name)
	var resp ListResponse[NetworkAccessPolicy]
	if err := c.patch(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PatchNetworkAccessPolicy: empty response from server")
	}
	return &resp.Items[0], nil
}

// ListNetworkAccessPolicyRules returns all rules for the given network access policy.
// It automatically follows continuation_token pagination to collect all results.
func (c *FlashBladeClient) ListNetworkAccessPolicyRules(ctx context.Context, policyName string) ([]NetworkAccessPolicyRule, error) {
	params := url.Values{}
	params.Set("policy_names", policyName)

	var all []NetworkAccessPolicyRule
	for {
		path := "/network-access-policies/rules?" + params.Encode()
		var resp ListResponse[NetworkAccessPolicyRule]
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

// GetNetworkAccessPolicyRuleByName retrieves a network access policy rule by name within a policy.
// Synthesizes a 404 APIError if the rule does not exist.
func (c *FlashBladeClient) GetNetworkAccessPolicyRuleByName(ctx context.Context, policyName, ruleName string) (*NetworkAccessPolicyRule, error) {
	return getOneByName[NetworkAccessPolicyRule](c, ctx, "/network-access-policies/rules?policy_names="+url.QueryEscape(policyName)+"&names="+url.QueryEscape(ruleName), "network access policy rule", ruleName)
}

// GetNetworkAccessPolicyRuleByIndex retrieves a network access policy rule by its index within the policy.
// Synthesizes a 404 APIError if no rule with the given index is found.
func (c *FlashBladeClient) GetNetworkAccessPolicyRuleByIndex(ctx context.Context, policyName string, index int) (*NetworkAccessPolicyRule, error) {
	rules, err := c.ListNetworkAccessPolicyRules(ctx, policyName)
	if err != nil {
		return nil, err
	}
	for i := range rules {
		if rules[i].Index == index {
			return &rules[i], nil
		}
	}
	return nil, &APIError{StatusCode: 404, Message: fmt.Sprintf("network access policy rule at index %d not found in policy %q", index, policyName)}
}

// PostNetworkAccessPolicyRule creates a new rule in a network access policy.
func (c *FlashBladeClient) PostNetworkAccessPolicyRule(ctx context.Context, policyName string, body NetworkAccessPolicyRulePost) (*NetworkAccessPolicyRule, error) {
	path := "/network-access-policies/rules?policy_names=" + url.QueryEscape(policyName)
	var resp ListResponse[NetworkAccessPolicyRule]
	if err := c.post(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PostNetworkAccessPolicyRule: empty response from server")
	}
	return &resp.Items[0], nil
}

// PatchNetworkAccessPolicyRule updates an existing network access policy rule.
func (c *FlashBladeClient) PatchNetworkAccessPolicyRule(ctx context.Context, policyName, ruleName string, body NetworkAccessPolicyRulePatch) (*NetworkAccessPolicyRule, error) {
	path := "/network-access-policies/rules?names=" + url.QueryEscape(ruleName) + "&policy_names=" + url.QueryEscape(policyName)
	var resp ListResponse[NetworkAccessPolicyRule]
	if err := c.patch(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PatchNetworkAccessPolicyRule: empty response from server")
	}
	return &resp.Items[0], nil
}

// DeleteNetworkAccessPolicyRule deletes a network access policy rule by name.
func (c *FlashBladeClient) DeleteNetworkAccessPolicyRule(ctx context.Context, policyName, ruleName string) error {
	path := "/network-access-policies/rules?names=" + url.QueryEscape(ruleName) + "&policy_names=" + url.QueryEscape(policyName)
	return c.delete(ctx, path)
}
