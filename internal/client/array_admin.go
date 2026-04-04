package client

import (
	"context"
	"net/url"
)

// GetArrayDns retrieves the DNS configuration of the array.
// Returns the first item from the list response.
func (c *FlashBladeClient) GetArrayDns(ctx context.Context) (*ArrayDns, error) {
	var resp ListResponse[ArrayDns]
	if err := c.get(ctx, "/dns", &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, &APIError{StatusCode: 404, Message: "DNS configuration not found"}
	}
	return &resp.Items[0], nil
}

// PostArrayDns creates a new DNS configuration entry.
func (c *FlashBladeClient) PostArrayDns(ctx context.Context, body ArrayDnsPost) (*ArrayDns, error) {
	return postOne[ArrayDnsPost, ArrayDns](c, ctx, "/dns", body, "PostArrayDns")
}

// PatchArrayDns updates the DNS configuration of the array.
func (c *FlashBladeClient) PatchArrayDns(ctx context.Context, body ArrayDnsPatch) (*ArrayDns, error) {
	return patchOne[ArrayDnsPatch, ArrayDns](c, ctx, "/dns", body, "PatchArrayDns")
}

// GetArrayNtp retrieves the NTP servers configured on the array.
// Returns the array info struct with ntp_servers populated.
func (c *FlashBladeClient) GetArrayNtp(ctx context.Context) (*ArrayInfo, error) {
	var resp ListResponse[ArrayInfo]
	if err := c.get(ctx, "/arrays", &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, &APIError{StatusCode: 404, Message: "array info not found"}
	}
	return &resp.Items[0], nil
}

// PatchArrayNtp updates the NTP server list on the array.
// Only the ntp_servers field is sent to avoid unintentional modification of other array settings.
func (c *FlashBladeClient) PatchArrayNtp(ctx context.Context, body ArrayNtpPatch) (*ArrayInfo, error) {
	return patchOne[ArrayNtpPatch, ArrayInfo](c, ctx, "/arrays", body, "PatchArrayNtp")
}

// GetSmtpServer retrieves the SMTP server configuration of the array.
// Returns the first item from the list response.
func (c *FlashBladeClient) GetSmtpServer(ctx context.Context) (*SmtpServer, error) {
	var resp ListResponse[SmtpServer]
	if err := c.get(ctx, "/smtp-servers", &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, &APIError{StatusCode: 404, Message: "SMTP server configuration not found"}
	}
	return &resp.Items[0], nil
}

// PatchSmtpServer updates the SMTP server configuration of the array.
func (c *FlashBladeClient) PatchSmtpServer(ctx context.Context, body SmtpServerPatch) (*SmtpServer, error) {
	return patchOne[SmtpServerPatch, SmtpServer](c, ctx, "/smtp-servers", body, "PatchSmtpServer")
}

// GetAlertWatchers returns all configured alert watchers (email recipients).
func (c *FlashBladeClient) GetAlertWatchers(ctx context.Context) ([]AlertWatcher, error) {
	var resp ListResponse[AlertWatcher]
	if err := c.get(ctx, "/alert-watchers", &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

// PostAlertWatcher creates a new alert watcher for the given email address.
// The email is passed as the names query parameter.
func (c *FlashBladeClient) PostAlertWatcher(ctx context.Context, email string, body AlertWatcherPost) (*AlertWatcher, error) {
	return postOne[AlertWatcherPost, AlertWatcher](c, ctx, "/alert-watchers?names="+url.QueryEscape(email), body, "PostAlertWatcher")
}

// PatchAlertWatcher updates an existing alert watcher identified by email address.
func (c *FlashBladeClient) PatchAlertWatcher(ctx context.Context, email string, body AlertWatcherPatch) (*AlertWatcher, error) {
	return patchOne[AlertWatcherPatch, AlertWatcher](c, ctx, "/alert-watchers?names="+url.QueryEscape(email), body, "PatchAlertWatcher")
}

// DeleteAlertWatcher removes an alert watcher identified by email address.
func (c *FlashBladeClient) DeleteAlertWatcher(ctx context.Context, email string) error {
	path := "/alert-watchers?names=" + url.QueryEscape(email)
	return c.delete(ctx, path)
}
