package v1

import "context"

// CreateRecurringPost creates a recurring post schedule
func (c *Client) CreateRecurringPost(ctx context.Context, req RecurringPostRequest, resp *RecurringPostResponse) error {
	return c.do(ctx, "POST", "posts/recurring", req, resp)
}

// AutoSchedulePost uses AI to determine optimal posting times
func (c *Client) AutoSchedulePost(ctx context.Context, req AutoScheduleRequest, resp *AutoScheduleResponse) error {
	return c.do(ctx, "POST", "posts/auto-schedule", req, resp)
}

// RecyclePost configures content recycling schedule
func (c *Client) RecyclePost(ctx context.Context, req RecyclePostRequest, resp *RecyclePostResponse) error {
	return c.do(ctx, "POST", "posts/recycle", req, resp)
}