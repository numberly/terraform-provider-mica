package client

import (
	"context"
	"net/url"
)

// GetResiliencyGroup retrieves a resiliency group by name.
// Returns an IsNotFound error if the resiliency group does not exist.
// Resiliency groups are hardware-managed; no POST/PATCH/DELETE methods are provided.
func (c *FlashBladeClient) GetResiliencyGroup(ctx context.Context, name string) (*ResiliencyGroup, error) {
	return getOneByName[ResiliencyGroup](c, ctx, "/resiliency-groups?names="+url.QueryEscape(name), "resiliency group", name)
}

// ListResiliencyGroups returns all resiliency groups.
func (c *FlashBladeClient) ListResiliencyGroups(ctx context.Context) ([]ResiliencyGroup, error) {
	var resp ListResponse[ResiliencyGroup]
	if err := c.get(ctx, "/resiliency-groups", &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}
