package client

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// ListObjectStoreAccessKeysOpts contains optional query parameters for ListObjectStoreAccessKeys.
type ListObjectStoreAccessKeysOpts struct {
	// Names filters results to specific access key names.
	Names []string
	// Filter is a free-form filter expression.
	Filter string
}

// GetObjectStoreAccessKey retrieves an object store access key by name.
// Returns an IsNotFound error if the key does not exist.
// NOTE: The API does NOT return secret_access_key in GET responses — it is only available on POST.
func (c *FlashBladeClient) GetObjectStoreAccessKey(ctx context.Context, name string) (*ObjectStoreAccessKey, error) {
	return getOneByName[ObjectStoreAccessKey](c, ctx, "/object-store-access-keys?names="+url.QueryEscape(name), "object store access key", name)
}

// ListObjectStoreAccessKeys returns all object store access keys matching the optional opts filters.
// It automatically follows continuation_token pagination to collect all results.
func (c *FlashBladeClient) ListObjectStoreAccessKeys(ctx context.Context, opts ListObjectStoreAccessKeysOpts) ([]ObjectStoreAccessKey, error) {
	params := url.Values{}
	if len(opts.Names) > 0 {
		params.Set("names", strings.Join(opts.Names, ","))
	}
	if opts.Filter != "" {
		params.Set("filter", opts.Filter)
	}

	var all []ObjectStoreAccessKey
	for {
		path := "/object-store-access-keys"
		if len(params) > 0 {
			path += "?" + params.Encode()
		}

		var resp ListResponse[ObjectStoreAccessKey]
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

// PostObjectStoreAccessKey creates a new object store access key.
// The response includes secret_access_key — callers MUST capture it immediately as it is
// never returned again by any subsequent API call.
func (c *FlashBladeClient) PostObjectStoreAccessKey(ctx context.Context, body ObjectStoreAccessKeyPost) (*ObjectStoreAccessKey, error) {
	var resp ListResponse[ObjectStoreAccessKey]
	if err := c.post(ctx, "/object-store-access-keys", body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PostObjectStoreAccessKey: empty response from server")
	}
	return &resp.Items[0], nil
}

// DeleteObjectStoreAccessKey permanently deletes an object store access key identified by name.
// Access keys have no soft-delete — deletion is immediate.
func (c *FlashBladeClient) DeleteObjectStoreAccessKey(ctx context.Context, name string) error {
	path := "/object-store-access-keys?names=" + url.QueryEscape(name)
	return c.delete(ctx, path)
}
