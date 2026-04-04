package client

import (
	"context"
	"fmt"
	"net/url"
)

// GetSnapshotPolicy retrieves a snapshot policy by name.
// Returns an IsNotFound error if the policy does not exist.
func (c *FlashBladeClient) GetSnapshotPolicy(ctx context.Context, name string) (*SnapshotPolicy, error) {
	return getOneByName[SnapshotPolicy](c, ctx, "/policies?names="+url.QueryEscape(name), "snapshot policy", name)
}

// PostSnapshotPolicy creates a new snapshot policy.
// The name is passed as a query parameter; optional fields (including inline rules) are in the body.
func (c *FlashBladeClient) PostSnapshotPolicy(ctx context.Context, name string, body SnapshotPolicyPost) (*SnapshotPolicy, error) {
	return postOne[SnapshotPolicyPost, SnapshotPolicy](c, ctx, "/policies?names="+url.QueryEscape(name), body, "PostSnapshotPolicy")
}

// PatchSnapshotPolicy updates an existing snapshot policy identified by name.
// Name is read-only for snapshot policies — do not include name in the body.
// Use AddSnapshotPolicyRule and RemoveSnapshotPolicyRule for rule management.
func (c *FlashBladeClient) PatchSnapshotPolicy(ctx context.Context, name string, body SnapshotPolicyPatch) (*SnapshotPolicy, error) {
	return patchOne[SnapshotPolicyPatch, SnapshotPolicy](c, ctx, "/policies?names="+url.QueryEscape(name), body, "PatchSnapshotPolicy")
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
// FlashBlade identifies rules by their scheduling fields (every, at, keep_for), not by name.
func (c *FlashBladeClient) RemoveSnapshotPolicyRule(ctx context.Context, policyName string, rule SnapshotPolicyRuleRemove) (*SnapshotPolicy, error) {
	body := SnapshotPolicyPatch{
		RemoveRules: []SnapshotPolicyRuleRemove{rule},
	}
	return c.PatchSnapshotPolicy(ctx, policyName, body)
}

// ReplaceSnapshotPolicyRule atomically removes an existing rule and adds a new one via PATCH.
// This is used for in-place rule updates since snapshot rules have no dedicated PATCH endpoint.
func (c *FlashBladeClient) ReplaceSnapshotPolicyRule(ctx context.Context, policyName string, oldRule SnapshotPolicyRuleRemove, newRule SnapshotPolicyRulePost) (*SnapshotPolicy, error) {
	body := SnapshotPolicyPatch{
		RemoveRules: []SnapshotPolicyRuleRemove{oldRule},
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
