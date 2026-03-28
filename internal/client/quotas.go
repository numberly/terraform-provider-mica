package client

import (
	"context"
	"fmt"
	"net/url"
)

// GetQuotaUser retrieves the quota for a specific user on a file system.
// Returns an IsNotFound error if the quota does not exist.
func (c *FlashBladeClient) GetQuotaUser(ctx context.Context, fileSystemName, uid string) (*QuotaUser, error) {
	path := "/quotas/users?file_system_names=" + url.QueryEscape(fileSystemName) + "&uids=" + url.QueryEscape(uid)
	var resp ListResponse[QuotaUser]
	if err := c.get(ctx, path, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, &APIError{StatusCode: 404, Message: fmt.Sprintf("user quota for uid %q on file system %q not found", uid, fileSystemName)}
	}
	return &resp.Items[0], nil
}

// ListQuotaUsers returns all user quotas for a given file system.
// It automatically follows continuation_token pagination to collect all results.
func (c *FlashBladeClient) ListQuotaUsers(ctx context.Context, fileSystemName string) ([]QuotaUser, error) {
	params := url.Values{}
	params.Set("file_system_names", fileSystemName)

	var all []QuotaUser
	for {
		path := "/quotas/users?" + params.Encode()
		var resp ListResponse[QuotaUser]
		if err := c.get(ctx, path, &resp); err != nil {
			return nil, err
		}
		all = append(all, resp.Items...)
		if resp.ContinuationToken == "" {
			break
		}
		params.Set("continuation_token", resp.ContinuationToken)
	}
	return all, nil
}

// PostQuotaUser creates a new user quota on a file system.
func (c *FlashBladeClient) PostQuotaUser(ctx context.Context, fileSystemName, uid string, body QuotaUserPost) (*QuotaUser, error) {
	path := "/quotas/users?file_system_names=" + url.QueryEscape(fileSystemName) + "&uids=" + url.QueryEscape(uid)
	var resp ListResponse[QuotaUser]
	if err := c.post(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PostQuotaUser: empty response from server")
	}
	return &resp.Items[0], nil
}

// PatchQuotaUser updates an existing user quota on a file system.
func (c *FlashBladeClient) PatchQuotaUser(ctx context.Context, fileSystemName, uid string, body QuotaUserPatch) (*QuotaUser, error) {
	path := "/quotas/users?file_system_names=" + url.QueryEscape(fileSystemName) + "&uids=" + url.QueryEscape(uid)
	var resp ListResponse[QuotaUser]
	if err := c.patch(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PatchQuotaUser: empty response from server")
	}
	return &resp.Items[0], nil
}

// DeleteQuotaUser removes a user quota from a file system.
func (c *FlashBladeClient) DeleteQuotaUser(ctx context.Context, fileSystemName, uid string) error {
	path := "/quotas/users?file_system_names=" + url.QueryEscape(fileSystemName) + "&uids=" + url.QueryEscape(uid)
	return c.delete(ctx, path)
}

// GetQuotaGroup retrieves the quota for a specific group on a file system.
// Returns an IsNotFound error if the quota does not exist.
func (c *FlashBladeClient) GetQuotaGroup(ctx context.Context, fileSystemName, gid string) (*QuotaGroup, error) {
	path := "/quotas/groups?file_system_names=" + url.QueryEscape(fileSystemName) + "&gids=" + url.QueryEscape(gid)
	var resp ListResponse[QuotaGroup]
	if err := c.get(ctx, path, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, &APIError{StatusCode: 404, Message: fmt.Sprintf("group quota for gid %q on file system %q not found", gid, fileSystemName)}
	}
	return &resp.Items[0], nil
}

// ListQuotaGroups returns all group quotas for a given file system.
// It automatically follows continuation_token pagination to collect all results.
func (c *FlashBladeClient) ListQuotaGroups(ctx context.Context, fileSystemName string) ([]QuotaGroup, error) {
	params := url.Values{}
	params.Set("file_system_names", fileSystemName)

	var all []QuotaGroup
	for {
		path := "/quotas/groups?" + params.Encode()
		var resp ListResponse[QuotaGroup]
		if err := c.get(ctx, path, &resp); err != nil {
			return nil, err
		}
		all = append(all, resp.Items...)
		if resp.ContinuationToken == "" {
			break
		}
		params.Set("continuation_token", resp.ContinuationToken)
	}
	return all, nil
}

// PostQuotaGroup creates a new group quota on a file system.
func (c *FlashBladeClient) PostQuotaGroup(ctx context.Context, fileSystemName, gid string, body QuotaGroupPost) (*QuotaGroup, error) {
	path := "/quotas/groups?file_system_names=" + url.QueryEscape(fileSystemName) + "&gids=" + url.QueryEscape(gid)
	var resp ListResponse[QuotaGroup]
	if err := c.post(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PostQuotaGroup: empty response from server")
	}
	return &resp.Items[0], nil
}

// PatchQuotaGroup updates an existing group quota on a file system.
func (c *FlashBladeClient) PatchQuotaGroup(ctx context.Context, fileSystemName, gid string, body QuotaGroupPatch) (*QuotaGroup, error) {
	path := "/quotas/groups?file_system_names=" + url.QueryEscape(fileSystemName) + "&gids=" + url.QueryEscape(gid)
	var resp ListResponse[QuotaGroup]
	if err := c.patch(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PatchQuotaGroup: empty response from server")
	}
	return &resp.Items[0], nil
}

// DeleteQuotaGroup removes a group quota from a file system.
func (c *FlashBladeClient) DeleteQuotaGroup(ctx context.Context, fileSystemName, gid string) error {
	path := "/quotas/groups?file_system_names=" + url.QueryEscape(fileSystemName) + "&gids=" + url.QueryEscape(gid)
	return c.delete(ctx, path)
}
