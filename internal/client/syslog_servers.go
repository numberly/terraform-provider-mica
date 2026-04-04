package client

import (
	"context"
	"net/url"
	"strings"
)

// ListSyslogServersOpts contains optional query parameters for ListSyslogServers.
type ListSyslogServersOpts struct {
	// Names filters results to specific syslog server names.
	Names []string
	// Filter is a free-form filter expression.
	Filter string
}

// GetSyslogServer retrieves a syslog server by name.
// Returns an IsNotFound error if the server does not exist.
func (c *FlashBladeClient) GetSyslogServer(ctx context.Context, name string) (*SyslogServer, error) {
	return getOneByName[SyslogServer](c, ctx, "/syslog-servers?names="+url.QueryEscape(name), "syslog server", name)
}

// ListSyslogServers returns all syslog servers matching the optional opts filters.
// It automatically follows continuation_token pagination to collect all results.
func (c *FlashBladeClient) ListSyslogServers(ctx context.Context, opts ListSyslogServersOpts) ([]SyslogServer, error) {
	params := url.Values{}
	if len(opts.Names) > 0 {
		params.Set("names", strings.Join(opts.Names, ","))
	}
	if opts.Filter != "" {
		params.Set("filter", opts.Filter)
	}

	var all []SyslogServer
	for {
		path := "/syslog-servers"
		if len(params) > 0 {
			path += "?" + params.Encode()
		}

		var resp ListResponse[SyslogServer]
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

// PostSyslogServer creates a new syslog server.
// The name is passed as a query parameter; URI/services/sources are in the body.
func (c *FlashBladeClient) PostSyslogServer(ctx context.Context, name string, body SyslogServerPost) (*SyslogServer, error) {
	return postOne[SyslogServerPost, SyslogServer](c, ctx, "/syslog-servers?names="+url.QueryEscape(name), body, "PostSyslogServer")
}

// PatchSyslogServer updates an existing syslog server identified by name.
func (c *FlashBladeClient) PatchSyslogServer(ctx context.Context, name string, body SyslogServerPatch) (*SyslogServer, error) {
	return patchOne[SyslogServerPatch, SyslogServer](c, ctx, "/syslog-servers?names="+url.QueryEscape(name), body, "PatchSyslogServer")
}

// DeleteSyslogServer permanently deletes a syslog server.
func (c *FlashBladeClient) DeleteSyslogServer(ctx context.Context, name string) error {
	path := "/syslog-servers?names=" + url.QueryEscape(name)
	return c.delete(ctx, path)
}
