package client

import (
	"context"
	"net/url"
)

// GetLinkAggregationGroup retrieves a link aggregation group by name.
// Returns an IsNotFound error if the LAG does not exist.
// LAGs are hardware-managed; no POST/PATCH/DELETE methods are provided.
func (c *FlashBladeClient) GetLinkAggregationGroup(ctx context.Context, name string) (*LinkAggregationGroup, error) {
	return getOneByName[LinkAggregationGroup](c, ctx, "/link-aggregation-groups?names="+url.QueryEscape(name), "link aggregation group", name)
}

// ListLinkAggregationGroups returns all link aggregation groups.
func (c *FlashBladeClient) ListLinkAggregationGroups(ctx context.Context) ([]LinkAggregationGroup, error) {
	var resp ListResponse[LinkAggregationGroup]
	if err := c.get(ctx, "/link-aggregation-groups", &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}
