package client

import (
	"context"
	"net/url"
)

// GetNetworkInterface retrieves a network interface by name.
// Returns an IsNotFound error if the network interface does not exist.
func (c *FlashBladeClient) GetNetworkInterface(ctx context.Context, name string) (*NetworkInterface, error) {
	return getOneByName[NetworkInterface](c, ctx, "/network-interfaces?names="+url.QueryEscape(name), "network-interface", name)
}

// ListNetworkInterfaces returns all network interfaces.
func (c *FlashBladeClient) ListNetworkInterfaces(ctx context.Context) ([]NetworkInterface, error) {
	var resp ListResponse[NetworkInterface]
	if err := c.get(ctx, "/network-interfaces", &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

// PostNetworkInterface creates a new network interface. The name is passed via the ?names= query
// parameter and the subnet via ?subnet_names= (neither is in the request body).
func (c *FlashBladeClient) PostNetworkInterface(ctx context.Context, name string, subnetName string, body NetworkInterfacePost) (*NetworkInterface, error) {
	return postOne[NetworkInterfacePost, NetworkInterface](c, ctx, "/network-interfaces?names="+url.QueryEscape(name)+"&subnet_names="+url.QueryEscape(subnetName), body, "PostNetworkInterface")
}

// PatchNetworkInterface updates an existing network interface identified by name.
// Only fields set in body are updated; Services and AttachedServers are always sent (full-replace).
func (c *FlashBladeClient) PatchNetworkInterface(ctx context.Context, name string, body NetworkInterfacePatch) (*NetworkInterface, error) {
	return patchOne[NetworkInterfacePatch, NetworkInterface](c, ctx, "/network-interfaces?names="+url.QueryEscape(name), body, "PatchNetworkInterface")
}

// DeleteNetworkInterface removes a network interface by name.
func (c *FlashBladeClient) DeleteNetworkInterface(ctx context.Context, name string) error {
	return c.delete(ctx, "/network-interfaces?names="+url.QueryEscape(name))
}
