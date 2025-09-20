package v1

import "time"

// ScheduleRequest represents scheduled post creation
type ScheduleRequest struct {
	ScheduledAt time.Time `json:"scheduled_at"`
	TimeZone    string    `json:"timezone,omitempty"`
	Accounts    []string  `json:"accounts"`
	Media       []Media   `json:"media,omitempty"`
	Text        string    `json:"text"`
}

// ScheduleResponse contains job ID for async processing
type ScheduleResponse struct {
	JobID string `json:"job_id"`
}

// CreateDraftRequest represents draft post creation
type CreateDraftRequest struct {
	Visibility string   `json:"visibility"` // draft_private or draft_public
	Accounts   []string `json:"accounts"`
	Media      []Media  `json:"media,omitempty"`
	Text       string   `json:"text"`
}

// CreateDraftResponse contains job ID for async processing
type CreateDraftResponse struct {
	JobID string `json:"job_id"`
}
