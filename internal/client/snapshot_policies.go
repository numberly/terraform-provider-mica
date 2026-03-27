package client

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// ListSnapshotPoliciesOpts contains optional query parameters for ListSnapshotPolicies.
type ListSnapshotPoliciesOpts struct {
	// Names filters results to specific policy names.
	Names []string
	// Filter is a free-form filter expression.
	Filter string
}

// GetSnapshotPolicy retrieves a snapshot policy by name.
// Returns an IsNotFound error if the policy does not exist.
func (c *FlashBladeClient) GetSnapshotPolicy(ctx context.Context, name string) (*SnapshotPolicy, error) {
	path := "/policies?names=" + url.QueryEscape(name)
	var resp ListResponse[SnapshotPolicy]
	if err := c.get(ctx, path, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, &APIError{StatusCode: 404, Message: fmt.Sprintf("snapshot policy %q not found", name)}
	}
	return &resp.Items[0], nil
}

// ListSnapshotPolicies returns all snapshot policies matching the optional opts filters.
func (c *FlashBladeClient) ListSnapshotPolicies(ctx context.Context, opts ListSnapshotPoliciesOpts) ([]SnapshotPolicy, error) {
	params := url.Values{}
	if len(opts.Names) > 0 {
		params.Set("names", strings.Join(opts.Names, ","))
	}
	if opts.Filter != "" {
		params.Set("filter", opts.Filter)
	}

	path := "/policies"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	var resp ListResponse[SnapshotPolicy]
	if err := c.get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

// PostSnapshotPolicy creates a new snapshot policy.
// The name is passed as a query parameter; optional fields (including inline rules) are in the body.
func (c *FlashBladeClient) PostSnapshotPolicy(ctx context.Context, name string, body SnapshotPolicyPost) (*SnapshotPolicy, error) {
	path := "/policies?names=" + url.QueryEscape(name)
	var resp ListResponse[SnapshotPolicy]
	if err := c.post(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PostSnapshotPolicy: empty response from server")
	}
	return &resp.Items[0], nil
}

// PatchSnapshotPolicy updates an existing snapshot policy identified by name.
// Name is read-only for snapshot policies — do not include name in the body.
// Use AddSnapshotPolicyRule and RemoveSnapshotPolicyRule for rule management.
func (c *FlashBladeClient) PatchSnapshotPolicy(ctx context.Context, name string, body SnapshotPolicyPatch) (*SnapshotPolicy, error) {
	path := "/policies?names=" + url.QueryEscape(name)
	var resp ListResponse[SnapshotPolicy]
	if err := c.patch(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PatchSnapshotPolicy: empty response from server")
	}
	return &resp.Items[0], nil
}

// DeleteSnapshotPolicy permanently deletes a snapshot policy.
func (c *FlashBladeClient) DeleteSnapshotPolicy(ctx context.Context, name string) error {
	path := "/policies?names=" + url.QueryEscape(name)
	return c.delete(ctx, path)
}

// AddSnapshotPolicyRule adds a rule to a snapshot policy via PATCH add_rules.
func (c *FlashBladeClient) AddSnapshotPolicyRule(ctx context.Context, policyName string, rule SnapshotPolicyRulePost) (*SnapshotPolicy, error) {
	body := SnapshotPolicyPatch{
		AddRules: []SnapshotPolicyRulePost{rule},
	}
	return c.PatchSnapshotPolicy(ctx, policyName, body)
}

// RemoveSnapshotPolicyRule removes a rule from a snapshot policy via PATCH remove_rules.
func (c *FlashBladeClient) RemoveSnapshotPolicyRule(ctx context.Context, policyName, ruleName string) (*SnapshotPolicy, error) {
	body := SnapshotPolicyPatch{
		RemoveRules: []SnapshotPolicyRuleRemove{{Name: ruleName}},
	}
	return c.PatchSnapshotPolicy(ctx, policyName, body)
}

// ReplaceSnapshotPolicyRule atomically removes an existing rule and adds a new one via PATCH.
// This is used for in-place rule updates since snapshot rules have no dedicated PATCH endpoint.
func (c *FlashBladeClient) ReplaceSnapshotPolicyRule(ctx context.Context, policyName, oldRuleName string, newRule SnapshotPolicyRulePost) (*SnapshotPolicy, error) {
	body := SnapshotPolicyPatch{
		RemoveRules: []SnapshotPolicyRuleRemove{{Name: oldRuleName}},
		AddRules:    []SnapshotPolicyRulePost{newRule},
	}
	return c.PatchSnapshotPolicy(ctx, policyName, body)
}

// GetSnapshotPolicyRuleByIndex retrieves an embedded rule from a snapshot policy by its position index.
// Synthesizes a 404 APIError if the index is out of range.
func (c *FlashBladeClient) GetSnapshotPolicyRuleByIndex(ctx context.Context, policyName string, index int) (*SnapshotPolicyRuleInPolicy, error) {
	policy, err := c.GetSnapshotPolicy(ctx, policyName)
	if err != nil {
		return nil, err
	}
	if index < 0 || index >= len(policy.Rules) {
		return nil, &APIError{StatusCode: 404, Message: fmt.Sprintf("snapshot policy rule at index %d not found in policy %q", index, policyName)}
	}
	rule := policy.Rules[index]
	return &rule, nil
}

// ListSnapshotPolicyMembers returns the file systems attached to the given snapshot policy.
// Used for delete-guard checks before removing a policy.
func (c *FlashBladeClient) ListSnapshotPolicyMembers(ctx context.Context, policyName string) ([]PolicyMember, error) {
	path := "/policies/file-systems?policy_names=" + url.QueryEscape(policyName)
	var resp ListResponse[PolicyMember]
	if err := c.get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}
