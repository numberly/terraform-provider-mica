package client

import (
	"context"
	"fmt"
	"net/url"
)

// ListResiliencyGroupMembers returns all member rows for the given resiliency
// group name. The endpoint is a flat list filtered by parent via the
// `resiliency_group_names` query param (API 2.23).
func (c *FlashBladeClient) ListResiliencyGroupMembers(ctx context.Context, groupName string) ([]ResiliencyGroupMember, error) {
	path := "/resiliency-groups/members?resiliency_group_names=" + url.QueryEscape(groupName)
	var resp ListResponse[ResiliencyGroupMember]
	if err := c.get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

// GetResiliencyGroupMember retrieves the single membership row whose group name
// matches groupName and whose member name matches memberName. Members have no
// global name field, so we list rows scoped by parent and filter client-side.
// Returns an IsNotFound error if no matching row exists.
func (c *FlashBladeClient) GetResiliencyGroupMember(ctx context.Context, groupName, memberName string) (*ResiliencyGroupMember, error) {
	items, err := c.ListResiliencyGroupMembers(ctx, groupName)
	if err != nil {
		return nil, err
	}
	for i := range items {
		if items[i].Member.Name == memberName {
			return &items[i], nil
		}
	}
	return nil, &APIError{
		StatusCode: 404,
		Message:    fmt.Sprintf("resiliency group member %q in group %q not found", memberName, groupName),
	}
}
