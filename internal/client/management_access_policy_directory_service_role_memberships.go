package client

import (
	"context"
	"fmt"
	"net/url"
)

// dsrmPath builds "/management-access-policies/directory-services/roles?policy_names=<p>&member_names=<r>".
// FlashBlade API v2.22 expects `member_names` (not `role_names`) on this endpoint — the role IS the
// member of the membership relation. Using `role_names` yields HTTP 400 "Member identifier is required".
func dsrmPath(policyName, roleName string) string {
	v := url.Values{}
	v.Set("policy_names", policyName)
	v.Set("member_names", roleName)
	return "/management-access-policies/directory-services/roles?" + v.Encode()
}

// GetManagementAccessPolicyDirectoryServiceRoleMembership verifies that the association
// policy↔role exists. Returns IsNotFound when the list is empty (association removed).
// The composite key follows D-05: role_name/policy_name (role first, special chars in policy name).
func (c *FlashBladeClient) GetManagementAccessPolicyDirectoryServiceRoleMembership(ctx context.Context, policyName, roleName string) (*ManagementAccessPolicyDirectoryServiceRoleMembership, error) {
	label := "management_access_policy_directory_service_role_membership"
	key := fmt.Sprintf("%s/%s", roleName, policyName)
	return getOneByName[ManagementAccessPolicyDirectoryServiceRoleMembership](c, ctx, dsrmPath(policyName, roleName), label, key)
}

// PostManagementAccessPolicyDirectoryServiceRoleMembership creates the association.
// The FlashBlade API requires an empty object body on POST to this endpoint.
func (c *FlashBladeClient) PostManagementAccessPolicyDirectoryServiceRoleMembership(ctx context.Context, policyName, roleName string) (*ManagementAccessPolicyDirectoryServiceRoleMembership, error) {
	return postOne[struct{}, ManagementAccessPolicyDirectoryServiceRoleMembership](c, ctx, dsrmPath(policyName, roleName), struct{}{}, "PostManagementAccessPolicyDirectoryServiceRoleMembership")
}

// DeleteManagementAccessPolicyDirectoryServiceRoleMembership removes the association.
// Neither the policy nor the role is deleted — only the link is severed.
func (c *FlashBladeClient) DeleteManagementAccessPolicyDirectoryServiceRoleMembership(ctx context.Context, policyName, roleName string) error {
	return c.delete(ctx, dsrmPath(policyName, roleName))
}
