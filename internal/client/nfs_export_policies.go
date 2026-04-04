package client

import (
	"context"
	"fmt"
	"net/url"
)

// GetNfsExportPolicy retrieves an NFS export policy by name.
// Returns an IsNotFound error if the policy does not exist.
func (c *FlashBladeClient) GetNfsExportPolicy(ctx context.Context, name string) (*NfsExportPolicy, error) {
	return getOneByName[NfsExportPolicy](c, ctx, "/nfs-export-policies?names="+url.QueryEscape(name), "NFS export policy", name)
}

// PostNfsExportPolicy creates a new NFS export policy.
// The name is passed as a query parameter; optional fields are in the body.
func (c *FlashBladeClient) PostNfsExportPolicy(ctx context.Context, name string, body NfsExportPolicyPost) (*NfsExportPolicy, error) {
	path := "/nfs-export-policies?names=" + url.QueryEscape(name)
	var resp ListResponse[NfsExportPolicy]
	if err := c.post(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PostNfsExportPolicy: empty response from server")
	}
	return &resp.Items[0], nil
}

// PatchNfsExportPolicy updates an existing NFS export policy identified by name.
// When renaming (body.Name != nil), the OLD name must be passed as the name argument.
func (c *FlashBladeClient) PatchNfsExportPolicy(ctx context.Context, name string, body NfsExportPolicyPatch) (*NfsExportPolicy, error) {
	path := "/nfs-export-policies?names=" + url.QueryEscape(name)
	var resp ListResponse[NfsExportPolicy]
	if err := c.patch(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PatchNfsExportPolicy: empty response from server")
	}
	return &resp.Items[0], nil
}

// DeleteNfsExportPolicy permanently deletes an NFS export policy.
func (c *FlashBladeClient) DeleteNfsExportPolicy(ctx context.Context, name string) error {
	path := "/nfs-export-policies?names=" + url.QueryEscape(name)
	return c.delete(ctx, path)
}

// ListNfsExportPolicyRules returns all rules for the given NFS export policy.
// It automatically follows continuation_token pagination to collect all results.
func (c *FlashBladeClient) ListNfsExportPolicyRules(ctx context.Context, policyName string) ([]NfsExportPolicyRule, error) {
	params := url.Values{}
	params.Set("policy_names", policyName)

	var all []NfsExportPolicyRule
	for {
		path := "/nfs-export-policies/rules?" + params.Encode()
		var resp ListResponse[NfsExportPolicyRule]
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

// GetNfsExportPolicyRuleByIndex retrieves an NFS export policy rule by its index within the policy.
// Synthesizes a 404 APIError if no rule with the given index is found.
func (c *FlashBladeClient) GetNfsExportPolicyRuleByIndex(ctx context.Context, policyName string, index int) (*NfsExportPolicyRule, error) {
	rules, err := c.ListNfsExportPolicyRules(ctx, policyName)
	if err != nil {
		return nil, err
	}
	for i := range rules {
		if rules[i].Index == index {
			return &rules[i], nil
		}
	}
	return nil, &APIError{StatusCode: 404, Message: fmt.Sprintf("NFS export policy rule at index %d not found in policy %q", index, policyName)}
}

// GetNfsExportPolicyRuleByName retrieves an NFS export policy rule by name within a policy.
// Synthesizes a 404 APIError if the rule does not exist.
func (c *FlashBladeClient) GetNfsExportPolicyRuleByName(ctx context.Context, policyName, ruleName string) (*NfsExportPolicyRule, error) {
	return getOneByName[NfsExportPolicyRule](c, ctx, "/nfs-export-policies/rules?policy_names="+url.QueryEscape(policyName)+"&names="+url.QueryEscape(ruleName), "NFS export policy rule", ruleName)
}

// PostNfsExportPolicyRule creates a new rule in an NFS export policy.
// The policy is identified via the policy_names query parameter only — it must NOT appear in the body.
func (c *FlashBladeClient) PostNfsExportPolicyRule(ctx context.Context, policyName string, body NfsExportPolicyRulePost) (*NfsExportPolicyRule, error) {
	path := "/nfs-export-policies/rules?policy_names=" + url.QueryEscape(policyName)
	var resp ListResponse[NfsExportPolicyRule]
	if err := c.post(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PostNfsExportPolicyRule: empty response from server")
	}
	return &resp.Items[0], nil
}

// PatchNfsExportPolicyRule updates an existing NFS export policy rule.
func (c *FlashBladeClient) PatchNfsExportPolicyRule(ctx context.Context, policyName, ruleName string, body NfsExportPolicyRulePatch) (*NfsExportPolicyRule, error) {
	path := "/nfs-export-policies/rules?names=" + url.QueryEscape(ruleName) + "&policy_names=" + url.QueryEscape(policyName)
	var resp ListResponse[NfsExportPolicyRule]
	if err := c.patch(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PatchNfsExportPolicyRule: empty response from server")
	}
	return &resp.Items[0], nil
}

// DeleteNfsExportPolicyRule deletes an NFS export policy rule by name.
func (c *FlashBladeClient) DeleteNfsExportPolicyRule(ctx context.Context, policyName, ruleName string) error {
	path := "/nfs-export-policies/rules?names=" + url.QueryEscape(ruleName) + "&policy_names=" + url.QueryEscape(policyName)
	return c.delete(ctx, path)
}

// ListNfsExportPolicyMembers returns the file systems that use the given NFS export policy.
// Used for delete-guard checks before removing a policy.
func (c *FlashBladeClient) ListNfsExportPolicyMembers(ctx context.Context, policyName string) ([]PolicyMember, error) {
	filter := "nfs.export_policy.name='" + policyName + "'"
	path := "/file-systems?filter=" + url.QueryEscape(filter)
	var resp ListResponse[PolicyMember]
	if err := c.get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}
