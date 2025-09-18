package v1

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

var postIDRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// Client-side validation is necessary to prevent path traversal attacks when constructing URLs.
// Without validation, malicious PostIDs like "../admin" could access unintended endpoints.
func validatePostID(postID string) error {
	if postID == "" {
		return fmt.Errorf("post ID cannot be empty")
	}
	if strings.Contains(postID, "..") || strings.Contains(postID, "/") || strings.Contains(postID, "\\") {
		return fmt.Errorf("post ID contains invalid characters")
	}
	if !postIDRegex.MatchString(postID) {
		return fmt.Errorf("post ID must contain only alphanumeric characters, hyphens, and underscores")
	}
	return nil
}

// GetPost retrieves a single post by ID
func (c *Client) GetPost(ctx context.Context, req GetPostRequest, resp *GetPostResponse) error {
	if err := validatePostID(req.PostID); err != nil {
		return fmt.Errorf("invalid post ID: %w", err)
	}
	path := fmt.Sprintf("posts/%s", req.PostID)
	return c.do(ctx, "GET", path, nil, resp)
}

// UpdatePost updates an existing post
func (c *Client) UpdatePost(ctx context.Context, req UpdatePostRequest, resp *UpdatePostResponse) error {
	if err := validatePostID(req.PostID); err != nil {
		return fmt.Errorf("invalid post ID: %w", err)
	}
	path := fmt.Sprintf("posts/%s", req.PostID)
	return c.do(ctx, "PATCH", path, req, resp)
}

// DeletePost deletes a post
func (c *Client) DeletePost(ctx context.Context, req DeletePostRequest, resp *DeletePostResponse) error {
	if err := validatePostID(req.PostID); err != nil {
		return fmt.Errorf("invalid post ID: %w", err)
	}
	path := fmt.Sprintf("posts/%s", req.PostID)
	return c.do(ctx, "DELETE", path, nil, resp)
}
