package client

import (
	"context"
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
	return postOne[CertificatePost, Certificate](c, ctx, "/certificates?names="+url.QueryEscape(name), body, "PostCertificate")
}

// PatchCertificate updates an existing certificate identified by name.
// Only non-nil pointer fields in body are sent (PATCH semantics).
func (c *FlashBladeClient) PatchCertificate(ctx context.Context, name string, body CertificatePatch) (*Certificate, error) {
	return patchOne[CertificatePatch, Certificate](c, ctx, "/certificates?names="+url.QueryEscape(name), body, "PatchCertificate")
}

// DeleteCertificate permanently deletes a certificate by name.
func (c *FlashBladeClient) DeleteCertificate(ctx context.Context, name string) error {
	path := "/certificates?names=" + url.QueryEscape(name)
	return c.delete(ctx, path)
}
