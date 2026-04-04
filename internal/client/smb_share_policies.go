package client

import (
	"context"
	"fmt"
	"net/url"
)

// GetSmbSharePolicy retrieves an SMB share policy by name.
// Returns an IsNotFound error if the policy does not exist.
func (c *FlashBladeClient) GetSmbSharePolicy(ctx context.Context, name string) (*SmbSharePolicy, error) {
	return getOneByName[SmbSharePolicy](c, ctx, "/smb-share-policies?names="+url.QueryEscape(name), "SMB share policy", name)
}

// PostSmbSharePolicy creates a new SMB share policy.
// The name is passed as a query parameter; optional fields are in the body.
func (c *FlashBladeClient) PostSmbSharePolicy(ctx context.Context, name string, body SmbSharePolicyPost) (*SmbSharePolicy, error) {
	path := "/smb-share-policies?names=" + url.QueryEscape(name)
	var resp ListResponse[SmbSharePolicy]
	if err := c.post(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PostSmbSharePolicy: empty response from server")
	}
	return &resp.Items[0], nil
}

// PatchSmbSharePolicy updates an existing SMB share policy identified by name.
// When renaming (body.Name != nil), the OLD name must be passed as the name argument.
func (c *FlashBladeClient) PatchSmbSharePolicy(ctx context.Context, name string, body SmbSharePolicyPatch) (*SmbSharePolicy, error) {
	path := "/smb-share-policies?names=" + url.QueryEscape(name)
	var resp ListResponse[SmbSharePolicy]
	if err := c.patch(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PatchSmbSharePolicy: empty response from server")
	}
	return &resp.Items[0], nil
}

// DeleteSmbSharePolicy permanently deletes an SMB share policy.
func (c *FlashBladeClient) DeleteSmbSharePolicy(ctx context.Context, name string) error {
	path := "/smb-share-policies?names=" + url.QueryEscape(name)
	return c.delete(ctx, path)
}

// ListSmbSharePolicyRules returns all rules for the given SMB share policy.
// It automatically follows continuation_token pagination to collect all results.
func (c *FlashBladeClient) ListSmbSharePolicyRules(ctx context.Context, policyName string) ([]SmbSharePolicyRule, error) {
	params := url.Values{}
	params.Set("policy_names", policyName)

	var all []SmbSharePolicyRule
	for {
		path := "/smb-share-policies/rules?" + params.Encode()
		var resp ListResponse[SmbSharePolicyRule]
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

// GetSmbSharePolicyRuleByName retrieves an SMB share policy rule by name within a policy.
// Synthesizes a 404 APIError if the rule does not exist.
func (c *FlashBladeClient) GetSmbSharePolicyRuleByName(ctx context.Context, policyName, ruleName string) (*SmbSharePolicyRule, error) {
	return getOneByName[SmbSharePolicyRule](c, ctx, "/smb-share-policies/rules?policy_names="+url.QueryEscape(policyName)+"&names="+url.QueryEscape(ruleName), "SMB share policy rule", ruleName)
}

// PostSmbSharePolicyRule creates a new rule in an SMB share policy.
func (c *FlashBladeClient) PostSmbSharePolicyRule(ctx context.Context, policyName string, body SmbSharePolicyRulePost) (*SmbSharePolicyRule, error) {
	path := "/smb-share-policies/rules?policy_names=" + url.QueryEscape(policyName)
	var resp ListResponse[SmbSharePolicyRule]
	if err := c.post(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PostSmbSharePolicyRule: empty response from server")
	}
	return &resp.Items[0], nil
}

// PatchSmbSharePolicyRule updates an existing SMB share policy rule.
func (c *FlashBladeClient) PatchSmbSharePolicyRule(ctx context.Context, policyName, ruleName string, body SmbSharePolicyRulePatch) (*SmbSharePolicyRule, error) {
	path := "/smb-share-policies/rules?names=" + url.QueryEscape(ruleName) + "&policy_names=" + url.QueryEscape(policyName)
	var resp ListResponse[SmbSharePolicyRule]
	if err := c.patch(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PatchSmbSharePolicyRule: empty response from server")
	}
	return &resp.Items[0], nil
}

// DeleteSmbSharePolicyRule deletes an SMB share policy rule by name.
func (c *FlashBladeClient) DeleteSmbSharePolicyRule(ctx context.Context, policyName, ruleName string) error {
	path := "/smb-share-policies/rules?names=" + url.QueryEscape(ruleName) + "&policy_names=" + url.QueryEscape(policyName)
	return c.delete(ctx, path)
}

// ListSmbSharePolicyMembers returns the file systems that use the given SMB share policy.
// Used for delete-guard checks before removing a policy.
func (c *FlashBladeClient) ListSmbSharePolicyMembers(ctx context.Context, policyName string) ([]PolicyMember, error) {
	filter := "smb.share_policy.name='" + policyName + "'"
	path := "/file-systems?filter=" + url.QueryEscape(filter)
	var resp ListResponse[PolicyMember]
	if err := c.get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}
