package client

import (
	"context"
	"fmt"
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
	path := "/network-interfaces?names=" + url.QueryEscape(name) + "&subnet_names=" + url.QueryEscape(subnetName)
	var resp ListResponse[NetworkInterface]
	if err := c.post(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PostNetworkInterface: empty response from server")
	}
	return &resp.Items[0], nil
}

// PatchNetworkInterface updates an existing network interface identified by name.
// Only fields set in body are updated; Services and AttachedServers are always sent (full-replace).
func (c *FlashBladeClient) PatchNetworkInterface(ctx context.Context, name string, body NetworkInterfacePatch) (*NetworkInterface, error) {
	path := "/network-interfaces?names=" + url.QueryEscape(name)
	var resp ListResponse[NetworkInterface]
	if err := c.patch(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PatchNetworkInterface: empty response from server")
	}
	return &resp.Items[0], nil
}

// DeleteNetworkInterface removes a network interface by name.
func (c *FlashBladeClient) DeleteNetworkInterface(ctx context.Context, name string) error {
	return c.delete(ctx, "/network-interfaces?names="+url.QueryEscape(name))
}
