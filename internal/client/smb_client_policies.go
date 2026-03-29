package client

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// ListSmbClientPoliciesOpts contains optional query parameters for ListSmbClientPolicies.
type ListSmbClientPoliciesOpts struct {
	// Names filters results to specific policy names.
	Names []string
	// Filter is a free-form filter expression.
	Filter string
}

// GetSmbClientPolicy retrieves an SMB client policy by name.
// Returns an IsNotFound error if the policy does not exist.
func (c *FlashBladeClient) GetSmbClientPolicy(ctx context.Context, name string) (*SmbClientPolicy, error) {
	return getOneByName[SmbClientPolicy](c, ctx, "/smb-client-policies?names="+url.QueryEscape(name), "SMB client policy", name)
}

// ListSmbClientPolicies returns all SMB client policies matching the optional opts filters.
// It automatically follows continuation_token pagination to collect all results.
func (c *FlashBladeClient) ListSmbClientPolicies(ctx context.Context, opts ListSmbClientPoliciesOpts) ([]SmbClientPolicy, error) {
	params := url.Values{}
	if len(opts.Names) > 0 {
		params.Set("names", strings.Join(opts.Names, ","))
	}
	if opts.Filter != "" {
		params.Set("filter", opts.Filter)
	}

	var all []SmbClientPolicy
	for {
		path := "/smb-client-policies"
		if len(params) > 0 {
			path += "?" + params.Encode()
		}

		var resp ListResponse[SmbClientPolicy]
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

// PostSmbClientPolicy creates a new SMB client policy.
// The name is passed as a query parameter; optional fields are in the body.
func (c *FlashBladeClient) PostSmbClientPolicy(ctx context.Context, name string, body SmbClientPolicyPost) (*SmbClientPolicy, error) {
	path := "/smb-client-policies?names=" + url.QueryEscape(name)
	var resp ListResponse[SmbClientPolicy]
	if err := c.post(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PostSmbClientPolicy: empty response from server")
	}
	return &resp.Items[0], nil
}

// PatchSmbClientPolicy updates an existing SMB client policy identified by name.
// When renaming (body.Name != nil), the OLD name must be passed as the name argument.
func (c *FlashBladeClient) PatchSmbClientPolicy(ctx context.Context, name string, body SmbClientPolicyPatch) (*SmbClientPolicy, error) {
	path := "/smb-client-policies?names=" + url.QueryEscape(name)
	var resp ListResponse[SmbClientPolicy]
	if err := c.patch(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PatchSmbClientPolicy: empty response from server")
	}
	return &resp.Items[0], nil
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
	path := "/smb-client-policies/rules?policy_names=" + url.QueryEscape(policyName) + "&names=" + url.QueryEscape(ruleName)
	var resp ListResponse[SmbClientPolicyRule]
	if err := c.get(ctx, path, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, &APIError{StatusCode: 404, Message: fmt.Sprintf("SMB client policy rule %q not found in policy %q", ruleName, policyName)}
	}
	return &resp.Items[0], nil
}

// PostSmbClientPolicyRule creates a new rule in an SMB client policy.
func (c *FlashBladeClient) PostSmbClientPolicyRule(ctx context.Context, policyName string, body SmbClientPolicyRulePost) (*SmbClientPolicyRule, error) {
	path := "/smb-client-policies/rules?policy_names=" + url.QueryEscape(policyName)
	var resp ListResponse[SmbClientPolicyRule]
	if err := c.post(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PostSmbClientPolicyRule: empty response from server")
	}
	return &resp.Items[0], nil
}

// PatchSmbClientPolicyRule updates an existing SMB client policy rule.
func (c *FlashBladeClient) PatchSmbClientPolicyRule(ctx context.Context, policyName, ruleName string, body SmbClientPolicyRulePatch) (*SmbClientPolicyRule, error) {
	path := "/smb-client-policies/rules?names=" + url.QueryEscape(ruleName) + "&policy_names=" + url.QueryEscape(policyName)
	var resp ListResponse[SmbClientPolicyRule]
	if err := c.patch(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PatchSmbClientPolicyRule: empty response from server")
	}
	return &resp.Items[0], nil
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
