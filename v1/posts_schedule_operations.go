package v1

import (
	"context"
)

// SchedulePost schedules a post for future publication
func (c *Client) SchedulePost(ctx context.Context, req SchedulePostRequest, resp *SchedulePostResponse) error {
	return c.do(ctx, "POST", "posts/schedule", req, resp)
}

// CreateDraftPost creates a draft post
func (c *Client) CreateDraftPost(ctx context.Context, req CreateDraftPostRequest, resp *CreateDraftPostResponse) error {
	return c.do(ctx, "POST", "posts/schedule", req, resp)
}