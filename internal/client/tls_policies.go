package client

import (
	"context"
	"fmt"
	"net/url"
)

// GetTlsPolicy retrieves a TLS policy by name.
// Returns an IsNotFound error if no policy exists with the given name.
func (c *FlashBladeClient) GetTlsPolicy(ctx context.Context, name string) (*TlsPolicy, error) {
	path := "/tls-policies?names=" + url.QueryEscape(name)
	return getOneByName[TlsPolicy](c, ctx, path, "TLS policy", name)
}

// PostTlsPolicy creates a new TLS policy. The name is passed via ?names= query parameter.
func (c *FlashBladeClient) PostTlsPolicy(ctx context.Context, name string, body TlsPolicyPost) (*TlsPolicy, error) {
	path := "/tls-policies?names=" + url.QueryEscape(name)
	var resp ListResponse[TlsPolicy]
	if err := c.post(ctx, path, body, &resp); err != nil {
		return nil, fmt.Errorf("PostTlsPolicy: %w", err)
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PostTlsPolicy: empty response from server")
	}
	return &resp.Items[0], nil
}

// PatchTlsPolicy updates an existing TLS policy identified by name.
func (c *FlashBladeClient) PatchTlsPolicy(ctx context.Context, name string, body TlsPolicyPatch) (*TlsPolicy, error) {
	path := "/tls-policies?names=" + url.QueryEscape(name)
	var resp ListResponse[TlsPolicy]
	if err := c.patch(ctx, path, body, &resp); err != nil {
		return nil, fmt.Errorf("PatchTlsPolicy: %w", err)
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PatchTlsPolicy: empty response from server")
	}
	return &resp.Items[0], nil
}

// DeleteTlsPolicy deletes a TLS policy by name.
func (c *FlashBladeClient) DeleteTlsPolicy(ctx context.Context, name string) error {
	path := "/tls-policies?names=" + url.QueryEscape(name)
	return c.delete(ctx, path)
}

// ListTlsPolicyMembers returns all network interfaces assigned to a TLS policy.
// It automatically follows continuation_token pagination to collect all results.
func (c *FlashBladeClient) ListTlsPolicyMembers(ctx context.Context, policyName string) ([]TlsPolicyMember, error) {
	params := url.Values{}
	params.Set("policy_names", policyName)
	var all []TlsPolicyMember
	for {
		path := "/tls-policies/members?" + params.Encode()
		var resp ListResponse[TlsPolicyMember]
		if err := c.get(ctx, path, &resp); err != nil {
			return nil, fmt.Errorf("ListTlsPolicyMembers: %w", err)
		}
		all = append(all, resp.Items...)
		if resp.ContinuationToken == "" {
			break
		}
		params.Set("continuation_token", resp.ContinuationToken)
	}
	return all, nil
}

// PostTlsPolicyMember assigns a network interface to a TLS policy.
// Uses the /network-interfaces/tls-policies endpoint per FlashBlade API spec.
func (c *FlashBladeClient) PostTlsPolicyMember(ctx context.Context, policyName string, memberName string) (*TlsPolicyMember, error) {
	path := "/network-interfaces/tls-policies?policy_names=" + url.QueryEscape(policyName) +
		"&member_names=" + url.QueryEscape(memberName)
	var resp ListResponse[TlsPolicyMember]
	if err := c.post(ctx, path, struct{}{}, &resp); err != nil {
		return nil, fmt.Errorf("PostTlsPolicyMember: %w", err)
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PostTlsPolicyMember: empty response from server")
	}
	return &resp.Items[0], nil
}

// DeleteTlsPolicyMember removes a network interface from a TLS policy.
// Uses the /network-interfaces/tls-policies endpoint per FlashBlade API spec.
func (c *FlashBladeClient) DeleteTlsPolicyMember(ctx context.Context, policyName string, memberName string) error {
	path := "/network-interfaces/tls-policies?policy_names=" + url.QueryEscape(policyName) +
		"&member_names=" + url.QueryEscape(memberName)
	return c.delete(ctx, path)
}
