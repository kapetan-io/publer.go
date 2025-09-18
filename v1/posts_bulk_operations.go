package v1

import (
	"context"
)

// BulkPublishPosts publishes multiple posts immediately
func (c *Client) BulkPublishPosts(ctx context.Context, req BulkPublishPostsRequest, resp *BulkPublishPostsResponse) error {
	return c.do(ctx, "POST", "posts/schedule/publish", req, resp)
}

// BulkSchedulePosts schedules multiple posts
func (c *Client) BulkSchedulePosts(ctx context.Context, req BulkSchedulePostsRequest, resp *BulkSchedulePostsResponse) error {
	return c.do(ctx, "POST", "posts/schedule", req, resp)
}