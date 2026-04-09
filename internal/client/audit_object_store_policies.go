package client

import (
	"context"
	"net/url"
)

// GetAuditObjectStorePolicy retrieves an audit object store policy by name.
// Returns an IsNotFound error if the policy does not exist.
func (c *FlashBladeClient) GetAuditObjectStorePolicy(ctx context.Context, name string) (*AuditObjectStorePolicy, error) {
	return getOneByName[AuditObjectStorePolicy](c, ctx, "/audit-object-store-policies?names="+url.QueryEscape(name), "audit object store policy", name)
}

// PostAuditObjectStorePolicy creates a new audit object store policy.
// The name is passed as a query parameter.
func (c *FlashBladeClient) PostAuditObjectStorePolicy(ctx context.Context, name string, body AuditObjectStorePolicyPost) (*AuditObjectStorePolicy, error) {
	return postOne[AuditObjectStorePolicyPost, AuditObjectStorePolicy](c, ctx, "/audit-object-store-policies?names="+url.QueryEscape(name), body, "PostAuditObjectStorePolicy")
}

// PatchAuditObjectStorePolicy updates an existing audit object store policy identified by name.
func (c *FlashBladeClient) PatchAuditObjectStorePolicy(ctx context.Context, name string, body AuditObjectStorePolicyPatch) (*AuditObjectStorePolicy, error) {
	return patchOne[AuditObjectStorePolicyPatch, AuditObjectStorePolicy](c, ctx, "/audit-object-store-policies?names="+url.QueryEscape(name), body, "PatchAuditObjectStorePolicy")
}

// DeleteAuditObjectStorePolicy permanently deletes an audit object store policy by name.
func (c *FlashBladeClient) DeleteAuditObjectStorePolicy(ctx context.Context, name string) error {
	path := "/audit-object-store-policies?names=" + url.QueryEscape(name)
	return c.delete(ctx, path)
}

// ListAuditObjectStorePolicyMembers returns all members of an audit object store policy.
func (c *FlashBladeClient) ListAuditObjectStorePolicyMembers(ctx context.Context, policyName string) ([]AuditObjectStorePolicyMember, error) {
	params := url.Values{}
	params.Set("policy_names", policyName)
	var all []AuditObjectStorePolicyMember
	for {
		path := "/audit-object-store-policies/members?" + params.Encode()

		var resp ListResponse[AuditObjectStorePolicyMember]
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

// PostAuditObjectStorePolicyMember adds a bucket as a member of an audit object store policy.
func (c *FlashBladeClient) PostAuditObjectStorePolicyMember(ctx context.Context, policyName string, memberName string) (*AuditObjectStorePolicyMember, error) {
	path := "/audit-object-store-policies/members?policy_names=" + url.QueryEscape(policyName) + "&member_names=" + url.QueryEscape(memberName)
	return postOne[struct{}, AuditObjectStorePolicyMember](c, ctx, path, struct{}{}, "PostAuditObjectStorePolicyMember")
}

// DeleteAuditObjectStorePolicyMember removes a bucket from an audit object store policy.
func (c *FlashBladeClient) DeleteAuditObjectStorePolicyMember(ctx context.Context, policyName string, memberName string) error {
	path := "/audit-object-store-policies/members?policy_names=" + url.QueryEscape(policyName) + "&member_names=" + url.QueryEscape(memberName)
	return c.delete(ctx, path)
}
