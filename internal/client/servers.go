package client

import (
	"context"
	"net/url"
	"strings"
)

// GetServer retrieves a server by name.
// Returns an IsNotFound error if the server does not exist.
func (c *FlashBladeClient) GetServer(ctx context.Context, name string) (*Server, error) {
	return getOneByName[Server](c, ctx, "/servers?names="+url.QueryEscape(name), "server", name)
}

// PostServer creates a new server. The API requires both ?names= (server name)
// and ?create_ds= (directory service name, conventionally name + "_nfs").
func (c *FlashBladeClient) PostServer(ctx context.Context, name string, body ServerPost) (*Server, error) {
	path := "/servers?names=" + url.QueryEscape(name) + "&create_ds=" + url.QueryEscape(name+"_nfs")
	return postOne[ServerPost, Server](c, ctx, path, body, "PostServer")
}

// PatchServer updates an existing server identified by name.
func (c *FlashBladeClient) PatchServer(ctx context.Context, name string, body ServerPatch) (*Server, error) {
	return patchOne[ServerPatch, Server](c, ctx, "/servers?names="+url.QueryEscape(name), body, "PatchServer")
}

// DeleteServer removes a server by name. If cascadeDelete is non-empty,
// the comma-joined export names are passed via the ?cascade_delete= query parameter.
func (c *FlashBladeClient) DeleteServer(ctx context.Context, name string, cascadeDelete []string) error {
	path := "/servers?names=" + url.QueryEscape(name)
	if len(cascadeDelete) > 0 {
		path += "&cascade_delete=" + url.QueryEscape(strings.Join(cascadeDelete, ","))
	}
	return c.delete(ctx, path)
}
