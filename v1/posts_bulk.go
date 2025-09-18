package v1

import "time"

// BulkPost represents a single post in bulk operation
type BulkPost struct {
	Text        string    `json:"text"`
	Accounts    []string  `json:"accounts"`
	ScheduledAt time.Time `json:"scheduled_at,omitempty"`
	Media       []Media   `json:"media,omitempty"`
}

// BulkPublishPostsRequest represents bulk immediate publishing
type BulkPublishPostsRequest struct {
	Posts []BulkPost `json:"posts"`
}

// BulkPublishPostsResponse contains job ID for async processing
type BulkPublishPostsResponse struct {
	JobID string `json:"job_id"`
}

// BulkSchedulePostsRequest represents bulk scheduled publishing
type BulkSchedulePostsRequest struct {
	Posts []BulkPost `json:"posts"`
}

// BulkSchedulePostsResponse contains job ID for async processing
type BulkSchedulePostsResponse struct {
	JobID string `json:"job_id"`
}