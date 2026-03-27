package client

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// ListNfsExportPoliciesOpts contains optional query parameters for ListNfsExportPolicies.
type ListNfsExportPoliciesOpts struct {
	// Names filters results to specific policy names.
	Names []string
	// Filter is a free-form filter expression.
	Filter string
}

// GetNfsExportPolicy retrieves an NFS export policy by name.
// Returns an IsNotFound error if the policy does not exist.
func (c *FlashBladeClient) GetNfsExportPolicy(ctx context.Context, name string) (*NfsExportPolicy, error) {
	path := "/nfs-export-policies?names=" + url.QueryEscape(name)
	var resp ListResponse[NfsExportPolicy]
	if err := c.get(ctx, path, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, &APIError{StatusCode: 404, Message: fmt.Sprintf("NFS export policy %q not found", name)}
	}
	return &resp.Items[0], nil
}

// ListNfsExportPolicies returns all NFS export policies matching the optional opts filters.
func (c *FlashBladeClient) ListNfsExportPolicies(ctx context.Context, opts ListNfsExportPoliciesOpts) ([]NfsExportPolicy, error) {
	params := url.Values{}
	if len(opts.Names) > 0 {
		params.Set("names", strings.Join(opts.Names, ","))
	}
	if opts.Filter != "" {
		params.Set("filter", opts.Filter)
	}

	path := "/nfs-export-policies"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	var resp ListResponse[NfsExportPolicy]
	if err := c.get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
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
func (c *FlashBladeClient) ListNfsExportPolicyRules(ctx context.Context, policyName string) ([]NfsExportPolicyRule, error) {
	path := "/nfs-export-policies/rules?policy_names=" + url.QueryEscape(policyName)
	var resp ListResponse[NfsExportPolicyRule]
	if err := c.get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
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
	path := "/nfs-export-policies/rules?policy_names=" + url.QueryEscape(policyName) + "&names=" + url.QueryEscape(ruleName)
	var resp ListResponse[NfsExportPolicyRule]
	if err := c.get(ctx, path, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, &APIError{StatusCode: 404, Message: fmt.Sprintf("NFS export policy rule %q not found in policy %q", ruleName, policyName)}
	}
	return &resp.Items[0], nil
}

// PostNfsExportPolicyRule creates a new rule in an NFS export policy.
func (c *FlashBladeClient) PostNfsExportPolicyRule(ctx context.Context, policyName string, body NfsExportPolicyRulePost) (*NfsExportPolicyRule, error) {
	// Ensure the policy reference is set in the body.
	if body.Policy == nil {
		body.Policy = &NamedReference{Name: policyName}
	}
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
