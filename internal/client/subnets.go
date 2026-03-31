package client

import (
	"context"
	"fmt"
	"net/url"
)

// GetSubnet retrieves a subnet by name.
// Returns an IsNotFound error if the subnet does not exist.
func (c *FlashBladeClient) GetSubnet(ctx context.Context, name string) (*Subnet, error) {
	return getOneByName[Subnet](c, ctx, "/subnets?names="+url.QueryEscape(name), "subnet", name)
}

// ListSubnets returns all subnets.
func (c *FlashBladeClient) ListSubnets(ctx context.Context) ([]Subnet, error) {
	var resp ListResponse[Subnet]
	if err := c.get(ctx, "/subnets", &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

// PostSubnet creates a new subnet. The name is passed via the ?names= query parameter
// (not in the request body) because name is a read-only field in the API model.
func (c *FlashBladeClient) PostSubnet(ctx context.Context, name string, body SubnetPost) (*Subnet, error) {
	path := "/subnets?names=" + url.QueryEscape(name)
	var resp ListResponse[Subnet]
	if err := c.post(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PostSubnet: empty response from server")
	}
	return &resp.Items[0], nil
}

// PatchSubnet updates an existing subnet identified by name.
// Only fields set in body are updated (partial PATCH semantics via pointer types).
func (c *FlashBladeClient) PatchSubnet(ctx context.Context, name string, body SubnetPatch) (*Subnet, error) {
	path := "/subnets?names=" + url.QueryEscape(name)
	var resp ListResponse[Subnet]
	if err := c.patch(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PatchSubnet: empty response from server")
	}
	return &resp.Items[0], nil
}

// DeleteSubnet removes a subnet by name.
func (c *FlashBladeClient) DeleteSubnet(ctx context.Context, name string) error {
	return c.delete(ctx, "/subnets?names="+url.QueryEscape(name))
}
