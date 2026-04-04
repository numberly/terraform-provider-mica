package client

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// ListFileSystemsOpts contains optional query parameters for ListFileSystems.
type ListFileSystemsOpts struct {
	// Names filters results to specific file system names (comma-separated when multiple).
	Names []string
	// Filter is a free-form filter expression.
	Filter string
	// Destroyed, when set to true, returns only soft-deleted file systems.
	Destroyed *bool
	// ContinuationToken is used for paginated results.
	ContinuationToken string
	// Limit restricts the number of results returned.
	Limit int
}

// GetFileSystem retrieves a file system by name.
// Returns an IsNotFound error if the file system does not exist.
func (c *FlashBladeClient) GetFileSystem(ctx context.Context, name string) (*FileSystem, error) {
	return getOneByName[FileSystem](c, ctx, "/file-systems?names="+url.QueryEscape(name), "file system", name)
}

// ListFileSystems returns all file systems matching the optional opts filters.
// It automatically follows continuation_token pagination to collect all results.
func (c *FlashBladeClient) ListFileSystems(ctx context.Context, opts ListFileSystemsOpts) ([]FileSystem, error) {
	params := url.Values{}
	if len(opts.Names) > 0 {
		params.Set("names", strings.Join(opts.Names, ","))
	}
	if opts.Filter != "" {
		params.Set("filter", opts.Filter)
	}
	if opts.Destroyed != nil {
		if *opts.Destroyed {
			params.Set("destroyed", "true")
		} else {
			params.Set("destroyed", "false")
		}
	}
	if opts.ContinuationToken != "" {
		params.Set("continuation_token", opts.ContinuationToken)
	}
	if opts.Limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", opts.Limit))
	}

	var all []FileSystem
	for {
		path := "/file-systems"
		if len(params) > 0 {
			path += "?" + params.Encode()
		}

		var resp ListResponse[FileSystem]
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

// PostFileSystem creates a new file system.
// The name is passed as a ?names= query parameter as required by the FlashBlade API.
func (c *FlashBladeClient) PostFileSystem(ctx context.Context, body FileSystemPost) (*FileSystem, error) {
	return postOne[FileSystemPost, FileSystem](c, ctx, "/file-systems?names="+url.QueryEscape(body.Name), body, "PostFileSystem")
}

// PatchFileSystem updates an existing file system identified by its ID.
// Only non-nil pointer fields in body are sent (PATCH semantics).
// Uses ID (not name) for stability across renames.
func (c *FlashBladeClient) PatchFileSystem(ctx context.Context, id string, body FileSystemPatch) (*FileSystem, error) {
	return patchOne[FileSystemPatch, FileSystem](c, ctx, "/file-systems?ids="+url.QueryEscape(id), body, "PatchFileSystem")
}

// DeleteFileSystem eradicates a soft-deleted file system identified by its ID.
// The file system must already be soft-deleted (destroyed=true) before calling this.
func (c *FlashBladeClient) DeleteFileSystem(ctx context.Context, id string) error {
	path := "/file-systems?ids=" + url.QueryEscape(id)
	return c.delete(ctx, path)
}

// PollUntilEradicated polls GET /file-systems?names={name}&destroyed=true until the
// file system is fully eradicated (empty items response). Respects context deadline.
// The caller should provide a context with an appropriate timeout (e.g., Terraform resource timeout).
func (c *FlashBladeClient) PollUntilEradicated(ctx context.Context, name string) error {
	return pollUntilGone[FileSystem](c, ctx, "/file-systems", "file system", name)
}
