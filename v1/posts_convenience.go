package v1

import (
	"context"
	"time"
)

// GetPostsByState returns an iterator for posts filtered by state
func (c *Client) GetPostsByState(state string) Iterator[Post] {
	req := ListPostsRequest{
		State: state,
	}
	return c.ListPosts(context.Background(), req)
}

// GetPostsByDateRange returns an iterator for posts within date range
func (c *Client) GetPostsByDateRange(from, to time.Time) Iterator[Post] {
	req := ListPostsRequest{
		From: from,
		To:   to,
	}
	return c.ListPosts(context.Background(), req)
}

// GetPostsByAccount returns posts for specific account
func (c *Client) GetPostsByAccount(accountID string) Iterator[Post] {
	req := ListPostsRequest{
		AccountIDs: []string{accountID},
	}
	return c.ListPosts(context.Background(), req)
}

// GetPostsByQuery returns posts matching search query
func (c *Client) GetPostsByQuery(query string) Iterator[Post] {
	req := ListPostsRequest{
		Query: query,
	}
	return c.ListPosts(context.Background(), req)
}