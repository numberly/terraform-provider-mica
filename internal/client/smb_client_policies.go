package client

import (
	"context"
	"net/url"
)

// GetSmbClientPolicy retrieves an SMB client policy by name.
// Returns an IsNotFound error if the policy does not exist.
func (c *FlashBladeClient) GetSmbClientPolicy(ctx context.Context, name string) (*SmbClientPolicy, error) {
	return getOneByName[SmbClientPolicy](c, ctx, "/smb-client-policies?names="+url.QueryEscape(name), "SMB client policy", name)
}

// PostSmbClientPolicy creates a new SMB client policy.
// The name is passed as a query parameter; optional fields are in the body.
func (c *FlashBladeClient) PostSmbClientPolicy(ctx context.Context, name string, body SmbClientPolicyPost) (*SmbClientPolicy, error) {
	return postOne[SmbClientPolicyPost, SmbClientPolicy](c, ctx, "/smb-client-policies?names="+url.QueryEscape(name), body, "PostSmbClientPolicy")
}

// PatchSmbClientPolicy updates an existing SMB client policy identified by name.
// When renaming (body.Name != nil), the OLD name must be passed as the name argument.
func (c *FlashBladeClient) PatchSmbClientPolicy(ctx context.Context, name string, body SmbClientPolicyPatch) (*SmbClientPolicy, error) {
	return patchOne[SmbClientPolicyPatch, SmbClientPolicy](c, ctx, "/smb-client-policies?names="+url.QueryEscape(name), body, "PatchSmbClientPolicy")
}

// DeleteSmbClientPolicy permanently deletes an SMB client policy.
func (c *FlashBladeClient) DeleteSmbClientPolicy(ctx context.Context, name string) error {
	path := "/smb-client-policies?names=" + url.QueryEscape(name)
	return c.delete(ctx, path)
}

// ListSmbClientPolicyRules returns all rules for the given SMB client policy.
// It automatically follows continuation_token pagination to collect all results.
func (c *FlashBladeClient) ListSmbClientPolicyRules(ctx context.Context, policyName string) ([]SmbClientPolicyRule, error) {
	params := url.Values{}
	params.Set("policy_names", policyName)

	var all []SmbClientPolicyRule
	for {
		path := "/smb-client-policies/rules?" + params.Encode()
		var resp ListResponse[SmbClientPolicyRule]
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

// GetSmbClientPolicyRuleByName retrieves an SMB client policy rule by name within a policy.
// Synthesizes a 404 APIError if the rule does not exist.
func (c *FlashBladeClient) GetSmbClientPolicyRuleByName(ctx context.Context, policyName, ruleName string) (*SmbClientPolicyRule, error) {
	return getOneByName[SmbClientPolicyRule](c, ctx, "/smb-client-policies/rules?policy_names="+url.QueryEscape(policyName)+"&names="+url.QueryEscape(ruleName), "SMB client policy rule", ruleName)
}

// PostSmbClientPolicyRule creates a new rule in an SMB client policy.
func (c *FlashBladeClient) PostSmbClientPolicyRule(ctx context.Context, policyName string, body SmbClientPolicyRulePost) (*SmbClientPolicyRule, error) {
	return postOne[SmbClientPolicyRulePost, SmbClientPolicyRule](c, ctx, "/smb-client-policies/rules?policy_names="+url.QueryEscape(policyName), body, "PostSmbClientPolicyRule")
}

// PatchSmbClientPolicyRule updates an existing SMB client policy rule.
func (c *FlashBladeClient) PatchSmbClientPolicyRule(ctx context.Context, policyName, ruleName string, body SmbClientPolicyRulePatch) (*SmbClientPolicyRule, error) {
	return patchOne[SmbClientPolicyRulePatch, SmbClientPolicyRule](c, ctx, "/smb-client-policies/rules?names="+url.QueryEscape(ruleName)+"&policy_names="+url.QueryEscape(policyName), body, "PatchSmbClientPolicyRule")
}

// DeleteSmbClientPolicyRule deletes an SMB client policy rule by name.
func (c *FlashBladeClient) DeleteSmbClientPolicyRule(ctx context.Context, policyName, ruleName string) error {
	path := "/smb-client-policies/rules?names=" + url.QueryEscape(ruleName) + "&policy_names=" + url.QueryEscape(policyName)
	return c.delete(ctx, path)
}

// ListSmbClientPolicyMembers returns the file systems that use the given SMB client policy.
// Used for delete-guard checks before removing a policy.
func (c *FlashBladeClient) ListSmbClientPolicyMembers(ctx context.Context, policyName string) ([]PolicyMember, error) {
	filter := "smb.client_policy.name='" + policyName + "'"
	path := "/file-systems?filter=" + url.QueryEscape(filter)
	var resp ListResponse[PolicyMember]
	if err := c.get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}
