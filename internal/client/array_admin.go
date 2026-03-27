package client

import (
	"context"
	"fmt"
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
	var resp ListResponse[ArrayDns]
	if err := c.post(ctx, "/dns", body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PostArrayDns: empty response from server")
	}
	return &resp.Items[0], nil
}

// PatchArrayDns updates the DNS configuration of the array.
func (c *FlashBladeClient) PatchArrayDns(ctx context.Context, body ArrayDnsPatch) (*ArrayDns, error) {
	var resp ListResponse[ArrayDns]
	if err := c.patch(ctx, "/dns", body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PatchArrayDns: empty response from server")
	}
	return &resp.Items[0], nil
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
	var resp ListResponse[ArrayInfo]
	if err := c.patch(ctx, "/arrays", body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PatchArrayNtp: empty response from server")
	}
	return &resp.Items[0], nil
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
	var resp ListResponse[SmtpServer]
	if err := c.patch(ctx, "/smtp-servers", body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PatchSmtpServer: empty response from server")
	}
	return &resp.Items[0], nil
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
	path := "/alert-watchers?names=" + url.QueryEscape(email)
	var resp ListResponse[AlertWatcher]
	if err := c.post(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PostAlertWatcher: empty response from server")
	}
	return &resp.Items[0], nil
}

// PatchAlertWatcher updates an existing alert watcher identified by email address.
func (c *FlashBladeClient) PatchAlertWatcher(ctx context.Context, email string, body AlertWatcherPatch) (*AlertWatcher, error) {
	path := "/alert-watchers?names=" + url.QueryEscape(email)
	var resp ListResponse[AlertWatcher]
	if err := c.patch(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PatchAlertWatcher: empty response from server")
	}
	return &resp.Items[0], nil
}

// DeleteAlertWatcher removes an alert watcher identified by email address.
func (c *FlashBladeClient) DeleteAlertWatcher(ctx context.Context, email string) error {
	path := "/alert-watchers?names=" + url.QueryEscape(email)
	return c.delete(ctx, path)
}
