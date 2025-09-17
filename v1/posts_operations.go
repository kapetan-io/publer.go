package v1

import (
	"context"
)

// ListPosts retrieves posts with filtering options
func (c *Client) ListPosts(ctx context.Context, request ListPostsRequest) Iterator[Post] {
	return NewPostIterator(c, request)
}

// PublishPost publishes content immediately
func (c *Client) PublishPost(ctx context.Context, request PublishPostRequest, response *PublishPostResponse) error {
	return c.do(ctx, "POST", "posts/schedule/publish", request, response)
}