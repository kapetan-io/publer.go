package v1

import "time"

// GetPostRequest represents request for single post
type GetPostRequest struct {
	PostID string
}

// GetPostResponse represents single post response
type GetPostResponse struct {
	Post
}

// UpdatePostRequest represents post update request
type UpdatePostRequest struct {
	ScheduledAt time.Time `json:"scheduled_at,omitempty"`
	Media       []Media   `json:"media,omitempty"`
	Text        string    `json:"text,omitempty"`
	PostID      string    `json:"-"`
}

// UpdatePostResponse represents post update response
type UpdatePostResponse struct {
	Post
}

// DeletePostRequest represents post deletion request
type DeletePostRequest struct {
	PostID string
}

// DeletePostResponse represents post deletion response
type DeletePostResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}