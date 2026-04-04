package client

import (
	"context"
	"net/url"
)

// GetObjectStoreAccessPolicy retrieves an object store access policy by name.
// Returns an IsNotFound error if the policy does not exist.
func (c *FlashBladeClient) GetObjectStoreAccessPolicy(ctx context.Context, name string) (*ObjectStoreAccessPolicy, error) {
	return getOneByName[ObjectStoreAccessPolicy](c, ctx, "/object-store-access-policies?names="+url.QueryEscape(name), "object store access policy", name)
}

// ListObjectStoreAccessPolicies returns all object store access policies.
// It automatically follows continuation_token pagination to collect all results.
func (c *FlashBladeClient) ListObjectStoreAccessPolicies(ctx context.Context) ([]ObjectStoreAccessPolicy, error) {
	params := url.Values{}

	var all []ObjectStoreAccessPolicy
	for {
		path := "/object-store-access-policies"
		if len(params) > 0 {
			path += "?" + params.Encode()
		}

		var resp ListResponse[ObjectStoreAccessPolicy]
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

// PostObjectStoreAccessPolicy creates a new object store access policy.
// The name is passed as a query parameter; optional fields are in the body.
func (c *FlashBladeClient) PostObjectStoreAccessPolicy(ctx context.Context, name string, body ObjectStoreAccessPolicyPost) (*ObjectStoreAccessPolicy, error) {
	return postOne[ObjectStoreAccessPolicyPost, ObjectStoreAccessPolicy](c, ctx, "/object-store-access-policies?names="+url.QueryEscape(name), body, "PostObjectStoreAccessPolicy")
}

// PatchObjectStoreAccessPolicy updates an existing object store access policy identified by name.
// When renaming (body.Name != nil), the OLD name must be passed as the name argument.
func (c *FlashBladeClient) PatchObjectStoreAccessPolicy(ctx context.Context, name string, body ObjectStoreAccessPolicyPatch) (*ObjectStoreAccessPolicy, error) {
	return patchOne[ObjectStoreAccessPolicyPatch, ObjectStoreAccessPolicy](c, ctx, "/object-store-access-policies?names="+url.QueryEscape(name), body, "PatchObjectStoreAccessPolicy")
}

// DeleteObjectStoreAccessPolicy permanently deletes an object store access policy.
func (c *FlashBladeClient) DeleteObjectStoreAccessPolicy(ctx context.Context, name string) error {
	path := "/object-store-access-policies?names=" + url.QueryEscape(name)
	return c.delete(ctx, path)
}

// GetObjectStoreAccessPolicyRuleByName retrieves an object store access policy rule by name within a policy.
// Synthesizes a 404 APIError if the rule does not exist.
func (c *FlashBladeClient) GetObjectStoreAccessPolicyRuleByName(ctx context.Context, policyName, ruleName string) (*ObjectStoreAccessPolicyRule, error) {
	return getOneByName[ObjectStoreAccessPolicyRule](c, ctx, "/object-store-access-policies/rules?policy_names="+url.QueryEscape(policyName)+"&names="+url.QueryEscape(ruleName), "object store access policy rule", ruleName)
}

// PostObjectStoreAccessPolicyRule creates a new rule in an object store access policy.
// Both policyName and ruleName are passed as query parameters.
func (c *FlashBladeClient) PostObjectStoreAccessPolicyRule(ctx context.Context, policyName, ruleName string, body ObjectStoreAccessPolicyRulePost) (*ObjectStoreAccessPolicyRule, error) {
	return postOne[ObjectStoreAccessPolicyRulePost, ObjectStoreAccessPolicyRule](c, ctx, "/object-store-access-policies/rules?policy_names="+url.QueryEscape(policyName)+"&names="+url.QueryEscape(ruleName), body, "PostObjectStoreAccessPolicyRule")
}

// PatchObjectStoreAccessPolicyRule updates an existing object store access policy rule.
func (c *FlashBladeClient) PatchObjectStoreAccessPolicyRule(ctx context.Context, policyName, ruleName string, body ObjectStoreAccessPolicyRulePatch) (*ObjectStoreAccessPolicyRule, error) {
	return patchOne[ObjectStoreAccessPolicyRulePatch, ObjectStoreAccessPolicyRule](c, ctx, "/object-store-access-policies/rules?names="+url.QueryEscape(ruleName)+"&policy_names="+url.QueryEscape(policyName), body, "PatchObjectStoreAccessPolicyRule")
}

// DeleteObjectStoreAccessPolicyRule deletes an object store access policy rule by name.
func (c *FlashBladeClient) DeleteObjectStoreAccessPolicyRule(ctx context.Context, policyName, ruleName string) error {
	path := "/object-store-access-policies/rules?names=" + url.QueryEscape(ruleName) + "&policy_names=" + url.QueryEscape(policyName)
	return c.delete(ctx, path)
}

// ListObjectStoreAccessPolicyMembers returns the users attached to the given object store access policy.
// Used for delete-guard checks before removing a policy.
func (c *FlashBladeClient) ListObjectStoreAccessPolicyMembers(ctx context.Context, policyName string) ([]PolicyMember, error) {
	path := "/object-store-access-policies/object-store-users?policy_names=" + url.QueryEscape(policyName)
	var resp ListResponse[PolicyMember]
	if err := c.get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}
