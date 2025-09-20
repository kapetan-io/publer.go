package v1

import "time"

// BulkPost represents a single post in bulk operation
type BulkPost struct {
	Text        string    `json:"text"`
	Accounts    []string  `json:"accounts"`
	ScheduledAt time.Time `json:"scheduled_at,omitempty"`
	Media       []Media   `json:"media,omitempty"`
}

// BulkPublishRequest represents bulk immediate publishing
type BulkPublishRequest struct {
	Posts []BulkPost `json:"posts"`
}

// BulkPublishResponse contains job ID for async processing
type BulkPublishResponse struct {
	JobID string `json:"job_id"`
}

// BulkScheduleRequest represents bulk scheduled publishing
type BulkScheduleRequest struct {
	Posts []BulkPost `json:"posts"`
}

// BulkScheduleResponse contains job ID for async processing
type BulkScheduleResponse struct {
	JobID string `json:"job_id"`
}
