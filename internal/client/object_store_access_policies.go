package client

import (
	"context"
	"fmt"
	"net/url"
)

// GetObjectStoreAccessPolicy retrieves an object store access policy by name.
// Returns an IsNotFound error if the policy does not exist.
func (c *FlashBladeClient) GetObjectStoreAccessPolicy(ctx context.Context, name string) (*ObjectStoreAccessPolicy, error) {
	path := "/object-store-access-policies?names=" + url.QueryEscape(name)
	var resp ListResponse[ObjectStoreAccessPolicy]
	if err := c.get(ctx, path, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, &APIError{StatusCode: 404, Message: fmt.Sprintf("object store access policy %q not found", name)}
	}
	return &resp.Items[0], nil
}

// ListObjectStoreAccessPolicies returns all object store access policies.
func (c *FlashBladeClient) ListObjectStoreAccessPolicies(ctx context.Context) ([]ObjectStoreAccessPolicy, error) {
	var resp ListResponse[ObjectStoreAccessPolicy]
	if err := c.get(ctx, "/object-store-access-policies", &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

// PostObjectStoreAccessPolicy creates a new object store access policy.
// The name is passed as a query parameter; optional fields are in the body.
func (c *FlashBladeClient) PostObjectStoreAccessPolicy(ctx context.Context, name string, body ObjectStoreAccessPolicyPost) (*ObjectStoreAccessPolicy, error) {
	path := "/object-store-access-policies?names=" + url.QueryEscape(name)
	var resp ListResponse[ObjectStoreAccessPolicy]
	if err := c.post(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PostObjectStoreAccessPolicy: empty response from server")
	}
	return &resp.Items[0], nil
}

// PatchObjectStoreAccessPolicy updates an existing object store access policy identified by name.
// When renaming (body.Name != nil), the OLD name must be passed as the name argument.
func (c *FlashBladeClient) PatchObjectStoreAccessPolicy(ctx context.Context, name string, body ObjectStoreAccessPolicyPatch) (*ObjectStoreAccessPolicy, error) {
	path := "/object-store-access-policies?names=" + url.QueryEscape(name)
	var resp ListResponse[ObjectStoreAccessPolicy]
	if err := c.patch(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PatchObjectStoreAccessPolicy: empty response from server")
	}
	return &resp.Items[0], nil
}

// DeleteObjectStoreAccessPolicy permanently deletes an object store access policy.
func (c *FlashBladeClient) DeleteObjectStoreAccessPolicy(ctx context.Context, name string) error {
	path := "/object-store-access-policies?names=" + url.QueryEscape(name)
	return c.delete(ctx, path)
}

// GetObjectStoreAccessPolicyRuleByName retrieves an object store access policy rule by name within a policy.
// Synthesizes a 404 APIError if the rule does not exist.
func (c *FlashBladeClient) GetObjectStoreAccessPolicyRuleByName(ctx context.Context, policyName, ruleName string) (*ObjectStoreAccessPolicyRule, error) {
	path := "/object-store-access-policies/rules?policy_names=" + url.QueryEscape(policyName) + "&names=" + url.QueryEscape(ruleName)
	var resp ListResponse[ObjectStoreAccessPolicyRule]
	if err := c.get(ctx, path, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, &APIError{StatusCode: 404, Message: fmt.Sprintf("object store access policy rule %q not found in policy %q", ruleName, policyName)}
	}
	return &resp.Items[0], nil
}

// PostObjectStoreAccessPolicyRule creates a new rule in an object store access policy.
// Both policyName and ruleName are passed as query parameters.
func (c *FlashBladeClient) PostObjectStoreAccessPolicyRule(ctx context.Context, policyName, ruleName string, body ObjectStoreAccessPolicyRulePost) (*ObjectStoreAccessPolicyRule, error) {
	path := "/object-store-access-policies/rules?policy_names=" + url.QueryEscape(policyName) + "&names=" + url.QueryEscape(ruleName)
	var resp ListResponse[ObjectStoreAccessPolicyRule]
	if err := c.post(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PostObjectStoreAccessPolicyRule: empty response from server")
	}
	return &resp.Items[0], nil
}

// PatchObjectStoreAccessPolicyRule updates an existing object store access policy rule.
func (c *FlashBladeClient) PatchObjectStoreAccessPolicyRule(ctx context.Context, policyName, ruleName string, body ObjectStoreAccessPolicyRulePatch) (*ObjectStoreAccessPolicyRule, error) {
	path := "/object-store-access-policies/rules?names=" + url.QueryEscape(ruleName) + "&policy_names=" + url.QueryEscape(policyName)
	var resp ListResponse[ObjectStoreAccessPolicyRule]
	if err := c.patch(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PatchObjectStoreAccessPolicyRule: empty response from server")
	}
	return &resp.Items[0], nil
}

// DeleteObjectStoreAccessPolicyRule deletes an object store access policy rule by name.
func (c *FlashBladeClient) DeleteObjectStoreAccessPolicyRule(ctx context.Context, policyName, ruleName string) error {
	path := "/object-store-access-policies/rules?names=" + url.QueryEscape(ruleName) + "&policy_names=" + url.QueryEscape(policyName)
	return c.delete(ctx, path)
}

// ListObjectStoreAccessPolicyMembers returns the buckets that use the given object store access policy.
// Used for delete-guard checks before removing a policy.
func (c *FlashBladeClient) ListObjectStoreAccessPolicyMembers(ctx context.Context, policyName string) ([]PolicyMember, error) {
	filter := "access_policy.name='" + policyName + "'"
	path := "/buckets?filter=" + url.QueryEscape(filter)
	var resp ListResponse[PolicyMember]
	if err := c.get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}
