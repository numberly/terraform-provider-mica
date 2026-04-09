package client

import (
	"context"
	"net/url"
	"strings"
)

// ListObjectStoreVirtualHostsOpts contains optional query parameters for ListObjectStoreVirtualHosts.
type ListObjectStoreVirtualHostsOpts struct {
	// Names filters results to specific virtual host names.
	Names []string
	// Filter is a free-form filter expression.
	Filter string
}

// GetObjectStoreVirtualHost retrieves an object store virtual host by name.
// Returns an IsNotFound error if the virtual host does not exist.
func (c *FlashBladeClient) GetObjectStoreVirtualHost(ctx context.Context, name string) (*ObjectStoreVirtualHost, error) {
	return getOneByName[ObjectStoreVirtualHost](c, ctx, "/object-store-virtual-hosts?names="+url.QueryEscape(name), "object store virtual host", name)
}

// ListObjectStoreVirtualHosts returns all object store virtual hosts matching the optional opts filters.
// It automatically follows continuation_token pagination to collect all results.
func (c *FlashBladeClient) ListObjectStoreVirtualHosts(ctx context.Context, opts ListObjectStoreVirtualHostsOpts) ([]ObjectStoreVirtualHost, error) {
	params := url.Values{}
	if len(opts.Names) > 0 {
		params.Set("names", strings.Join(opts.Names, ","))
	}
	if opts.Filter != "" {
		params.Set("filter", opts.Filter)
	}

	var all []ObjectStoreVirtualHost
	for {
		path := "/object-store-virtual-hosts"
		if len(params) > 0 {
			path += "?" + params.Encode()
		}

		var resp ListResponse[ObjectStoreVirtualHost]
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

// PostObjectStoreVirtualHost creates a new object store virtual host.
// The name is passed as the ?names= query parameter; body contains hostname and attached_servers.
func (c *FlashBladeClient) PostObjectStoreVirtualHost(ctx context.Context, name string, body ObjectStoreVirtualHostPost) (*ObjectStoreVirtualHost, error) {
	return postOne[ObjectStoreVirtualHostPost, ObjectStoreVirtualHost](c, ctx, "/object-store-virtual-hosts?names="+url.QueryEscape(name), body, "PostObjectStoreVirtualHost")
}

// PatchObjectStoreVirtualHost updates an existing object store virtual host identified by its server-assigned name.
// When renaming (body.Name != nil), the OLD name must be passed as the name argument.
func (c *FlashBladeClient) PatchObjectStoreVirtualHost(ctx context.Context, name string, body ObjectStoreVirtualHostPatch) (*ObjectStoreVirtualHost, error) {
	return patchOne[ObjectStoreVirtualHostPatch, ObjectStoreVirtualHost](c, ctx, "/object-store-virtual-hosts?names="+url.QueryEscape(name), body, "PatchObjectStoreVirtualHost")
}

// DeleteObjectStoreVirtualHost permanently deletes an object store virtual host.
func (c *FlashBladeClient) DeleteObjectStoreVirtualHost(ctx context.Context, name string) error {
	path := "/object-store-virtual-hosts?names=" + url.QueryEscape(name)
	return c.delete(ctx, path)
}
