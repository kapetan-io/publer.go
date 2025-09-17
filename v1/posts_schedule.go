package v1

import "time"

// SchedulePostRequest represents scheduled post creation
type SchedulePostRequest struct {
	ScheduledAt time.Time `json:"scheduled_at"`
	TimeZone    string    `json:"timezone,omitempty"`
	Accounts    []string  `json:"accounts"`
	Media       []Media   `json:"media,omitempty"`
	Text        string    `json:"text"`
}

// SchedulePostResponse contains job ID for async processing
type SchedulePostResponse struct {
	JobID string `json:"job_id"`
}

// CreateDraftPostRequest represents draft post creation
type CreateDraftPostRequest struct {
	Visibility string   `json:"visibility"` // draft_private or draft_public
	Accounts   []string `json:"accounts"`
	Media      []Media  `json:"media,omitempty"`
	Text       string   `json:"text"`
}

// CreateDraftPostResponse contains job ID for async processing
type CreateDraftPostResponse struct {
	JobID string `json:"job_id"`
}