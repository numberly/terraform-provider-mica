package client

import (
	"context"
	"fmt"
	"net/url"
)

// GetCertificate retrieves a certificate by name.
// Returns an IsNotFound error if the certificate does not exist.
func (c *FlashBladeClient) GetCertificate(ctx context.Context, name string) (*Certificate, error) {
	return getOneByName[Certificate](c, ctx, "/certificates?names="+url.QueryEscape(name), "certificate", name)
}

// PostCertificate imports a new certificate.
// The name is passed via ?names= query parameter.
func (c *FlashBladeClient) PostCertificate(ctx context.Context, name string, body CertificatePost) (*Certificate, error) {
	path := "/certificates?names=" + url.QueryEscape(name)
	var resp ListResponse[Certificate]
	if err := c.post(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PostCertificate: empty response from server")
	}
	return &resp.Items[0], nil
}

// PatchCertificate updates an existing certificate identified by name.
// Only non-nil pointer fields in body are sent (PATCH semantics).
func (c *FlashBladeClient) PatchCertificate(ctx context.Context, name string, body CertificatePatch) (*Certificate, error) {
	path := "/certificates?names=" + url.QueryEscape(name)
	var resp ListResponse[Certificate]
	if err := c.patch(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PatchCertificate: empty response from server")
	}
	return &resp.Items[0], nil
}

// DeleteCertificate permanently deletes a certificate by name.
func (c *FlashBladeClient) DeleteCertificate(ctx context.Context, name string) error {
	path := "/certificates?names=" + url.QueryEscape(name)
	return c.delete(ctx, path)
}
