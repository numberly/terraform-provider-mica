package client

import (
	"context"
	"fmt"
	"net/url"
)

// GetQosPolicy retrieves a QoS policy by name.
// Returns an IsNotFound error if no policy exists with the given name.
func (c *FlashBladeClient) GetQosPolicy(ctx context.Context, name string) (*QosPolicy, error) {
	return getOneByName[QosPolicy](c, ctx, "/qos-policies?names="+url.QueryEscape(name), "QoS policy", name)
}

// PostQosPolicy creates a new QoS policy. The name is passed via ?names= query parameter.
func (c *FlashBladeClient) PostQosPolicy(ctx context.Context, name string, body QosPolicyPost) (*QosPolicy, error) {
	path := "/qos-policies?names=" + url.QueryEscape(name)
	var resp ListResponse[QosPolicy]
	if err := c.post(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PostQosPolicy: empty response from server")
	}
	return &resp.Items[0], nil
}

// PatchQosPolicy updates an existing QoS policy identified by name.
func (c *FlashBladeClient) PatchQosPolicy(ctx context.Context, name string, body QosPolicyPatch) (*QosPolicy, error) {
	path := "/qos-policies?names=" + url.QueryEscape(name)
	var resp ListResponse[QosPolicy]
	if err := c.patch(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PatchQosPolicy: empty response from server")
	}
	return &resp.Items[0], nil
}

// DeleteQosPolicy deletes a QoS policy by name.
func (c *FlashBladeClient) DeleteQosPolicy(ctx context.Context, name string) error {
	path := "/qos-policies?names=" + url.QueryEscape(name)
	return c.delete(ctx, path)
}

// ListQosPolicyMembers returns all members of a QoS policy.
// It automatically follows continuation_token pagination to collect all results.
func (c *FlashBladeClient) ListQosPolicyMembers(ctx context.Context, policyName string) ([]QosPolicyMember, error) {
	params := url.Values{}
	params.Set("policy_names", policyName)
	var all []QosPolicyMember
	for {
		path := "/qos-policies/members?" + params.Encode()

		var resp ListResponse[QosPolicyMember]
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

// PostQosPolicyMember adds a member to a QoS policy.
// The API requires both ?policy_names= and ?member_names= on POST.
func (c *FlashBladeClient) PostQosPolicyMember(ctx context.Context, policyName string, memberName string, memberType string) (*QosPolicyMember, error) {
	path := "/qos-policies/members?policy_names=" + url.QueryEscape(policyName) + "&member_names=" + url.QueryEscape(memberName) + "&member_types=" + url.QueryEscape(memberType)
	var resp ListResponse[QosPolicyMember]
	if err := c.post(ctx, path, struct{}{}, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("PostQosPolicyMember: empty response from server")
	}
	return &resp.Items[0], nil
}

// DeleteQosPolicyMember removes a member from a QoS policy.
func (c *FlashBladeClient) DeleteQosPolicyMember(ctx context.Context, policyName string, memberName string) error {
	path := "/qos-policies/members?policy_names=" + url.QueryEscape(policyName) + "&member_names=" + url.QueryEscape(memberName)
	return c.delete(ctx, path)
}
