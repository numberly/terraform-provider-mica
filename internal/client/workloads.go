package client

import (
	"context"
	"net/url"
)

// GetWorkload retrieves a workload by name.
// Returns an IsNotFound error if the workload does not exist.
func (c *FlashBladeClient) GetWorkload(ctx context.Context, name string) (*Workload, error) {
	return getOneByName[Workload](c, ctx, "/workloads?names="+url.QueryEscape(name), "workload", name)
}

// GetDestroyedWorkload retrieves a workload by name, including destroyed (pending eradication) ones.
func (c *FlashBladeClient) GetDestroyedWorkload(ctx context.Context, name string) (*Workload, error) {
	return getOneByName[Workload](c, ctx, "/workloads?names="+url.QueryEscape(name)+"&destroyed=true", "workload", name)
}

// PostWorkload creates a new workload from the named preset.
// The workload name is passed via ?names= and the preset via ?preset_names=.
func (c *FlashBladeClient) PostWorkload(ctx context.Context, name, presetName string, body WorkloadPost) (*Workload, error) {
	path := "/workloads?names=" + url.QueryEscape(name) + "&preset_names=" + url.QueryEscape(presetName)
	return postOne[WorkloadPost, Workload](c, ctx, path, body, "PostWorkload")
}

// PatchWorkload updates an existing workload identified by name.
// Only non-nil pointer fields in body are sent (PATCH semantics).
func (c *FlashBladeClient) PatchWorkload(ctx context.Context, name string, body WorkloadPatch) (*Workload, error) {
	return patchOne[WorkloadPatch, Workload](c, ctx, "/workloads?names="+url.QueryEscape(name), body, "PatchWorkload")
}

// DeleteWorkload permanently eradicates a destroyed workload by name.
func (c *FlashBladeClient) DeleteWorkload(ctx context.Context, name string) error {
	return c.delete(ctx, "/workloads?names="+url.QueryEscape(name))
}

// PollWorkloadUntilEradicated polls until the named workload is gone from the
// destroyed queue (i.e. fully eradicated).
func (c *FlashBladeClient) PollWorkloadUntilEradicated(ctx context.Context, name string) error {
	return pollUntilGone[Workload](c, ctx, "/workloads", "workload", name)
}

// DestroyAndEradicateWorkload destroys a workload (soft-delete) and, if eradicate
// is true, waits for the eradication window to close and permanently removes it.
func (c *FlashBladeClient) DestroyAndEradicateWorkload(ctx context.Context, name string, eradicate bool) error {
	destroyed := true
	if _, err := c.PatchWorkload(ctx, name, WorkloadPatch{Destroyed: &destroyed}); err != nil {
		if IsNotFound(err) {
			return nil
		}
		return err
	}
	if !eradicate {
		return nil
	}
	if err := c.DeleteWorkload(ctx, name); err != nil && !IsNotFound(err) {
		return err
	}
	return c.PollWorkloadUntilEradicated(ctx, name)
}
