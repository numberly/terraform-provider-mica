package client

import (
	"context"
	"net/url"
)

// GetSnmpManager retrieves an SNMP manager by name.
// Returns an IsNotFound error if no manager matches.
func (c *FlashBladeClient) GetSnmpManager(ctx context.Context, name string) (*SnmpManager, error) {
	return getOneByName[SnmpManager](c, ctx, "/snmp-managers?names="+url.QueryEscape(name), "snmp_manager", name)
}

// ListSnmpManagers returns all configured SNMP managers.
func (c *FlashBladeClient) ListSnmpManagers(ctx context.Context) ([]SnmpManager, error) {
	var resp ListResponse[SnmpManager]
	if err := c.get(ctx, "/snmp-managers", &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

// PostSnmpManager creates a new SNMP manager.
// The name is passed via ?names= query parameter.
func (c *FlashBladeClient) PostSnmpManager(ctx context.Context, name string, body SnmpManagerPost) (*SnmpManager, error) {
	return postOne[SnmpManagerPost, SnmpManager](c, ctx, "/snmp-managers?names="+url.QueryEscape(name), body, "PostSnmpManager")
}

// PatchSnmpManager updates an existing SNMP manager identified by name.
// Only non-nil pointer fields in body are sent (PATCH semantics).
func (c *FlashBladeClient) PatchSnmpManager(ctx context.Context, name string, body SnmpManagerPatch) (*SnmpManager, error) {
	return patchOne[SnmpManagerPatch, SnmpManager](c, ctx, "/snmp-managers?names="+url.QueryEscape(name), body, "PatchSnmpManager")
}

// DeleteSnmpManager permanently removes an SNMP manager by name.
func (c *FlashBladeClient) DeleteSnmpManager(ctx context.Context, name string) error {
	path := "/snmp-managers?names=" + url.QueryEscape(name)
	return c.delete(ctx, path)
}
