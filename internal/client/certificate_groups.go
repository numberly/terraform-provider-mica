package client

import (
	"context"
	"fmt"
	"net/url"
)

// GetCertificateGroup retrieves a certificate group by name.
// Returns an IsNotFound error if no group exists with the given name.
func (c *FlashBladeClient) GetCertificateGroup(ctx context.Context, name string) (*CertificateGroup, error) {
	path := "/certificate-groups?names=" + url.QueryEscape(name)
	return getOneByName[CertificateGroup](c, ctx, path, "certificate group", name)
}

// PostCertificateGroup creates a new certificate group. The name is passed via ?names= query parameter.
// No body fields are required — the API creates the group from the name alone.
func (c *FlashBladeClient) PostCertificateGroup(ctx context.Context, name string) (*CertificateGroup, error) {
	return postOne[CertificateGroupPost, CertificateGroup](c, ctx, "/certificate-groups?names="+url.QueryEscape(name), CertificateGroupPost{}, "PostCertificateGroup")
}

// DeleteCertificateGroup deletes a certificate group by name.
func (c *FlashBladeClient) DeleteCertificateGroup(ctx context.Context, name string) error {
	path := "/certificate-groups?names=" + url.QueryEscape(name)
	return c.delete(ctx, path)
}

// ListCertificateGroupMembers returns all certificates assigned to a certificate group.
// It automatically follows continuation_token pagination to collect all results.
func (c *FlashBladeClient) ListCertificateGroupMembers(ctx context.Context, groupName string) ([]CertificateGroupMember, error) {
	params := url.Values{}
	params.Set("certificate_group_names", groupName)
	var all []CertificateGroupMember
	for {
		path := "/certificate-groups/certificates?" + params.Encode()
		var resp ListResponse[CertificateGroupMember]
		if err := c.get(ctx, path, &resp); err != nil {
			return nil, fmt.Errorf("ListCertificateGroupMembers: %w", err)
		}
		all = append(all, resp.Items...)
		if resp.ContinuationToken == "" {
			break
		}
		params.Set("continuation_token", resp.ContinuationToken)
	}
	return all, nil
}

// GetCertificateGroupMember checks if a specific certificate is in a certificate group.
// Filters by both group and certificate name. Returns true if the membership exists.
func (c *FlashBladeClient) GetCertificateGroupMember(ctx context.Context, groupName string, certName string) (bool, error) {
	path := "/certificate-groups/certificates?certificate_group_names=" + url.QueryEscape(groupName) +
		"&certificate_names=" + url.QueryEscape(certName)
	var resp ListResponse[CertificateGroupMember]
	if err := c.get(ctx, path, &resp); err != nil {
		return false, fmt.Errorf("GetCertificateGroupMember: %w", err)
	}
	return len(resp.Items) > 0, nil
}

// PostCertificateGroupMember adds a certificate to a certificate group.
// Both names are passed as query parameters — no request body.
func (c *FlashBladeClient) PostCertificateGroupMember(ctx context.Context, groupName string, certName string) (*CertificateGroupMember, error) {
	path := "/certificate-groups/certificates?certificate_group_names=" + url.QueryEscape(groupName) +
		"&certificate_names=" + url.QueryEscape(certName)
	return postOne[struct{}, CertificateGroupMember](c, ctx, path, struct{}{}, "PostCertificateGroupMember")
}

// DeleteCertificateGroupMember removes a certificate from a certificate group.
// Both names are passed as query parameters.
func (c *FlashBladeClient) DeleteCertificateGroupMember(ctx context.Context, groupName string, certName string) error {
	path := "/certificate-groups/certificates?certificate_group_names=" + url.QueryEscape(groupName) +
		"&certificate_names=" + url.QueryEscape(certName)
	return c.delete(ctx, path)
}
