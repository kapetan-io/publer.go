package v1

import "time"

// RecurringPostRequest represents recurring post configuration
type RecurringPostRequest struct {
	Text       string         `json:"text"`
	Accounts   []string       `json:"accounts"`
	Media      []Media        `json:"media,omitempty"`
	Recurrence RecurrenceRule `json:"recurrence"`
}

// RecurrenceRule defines how posts repeat
type RecurrenceRule struct {
	Frequency  string    `json:"frequency"`           // daily, weekly, monthly
	Interval   int       `json:"interval"`            // every N days/weeks/months
	DaysOfWeek []string  `json:"days_of_week,omitempty"` // for weekly: ["monday", "friday"]
	EndDate    time.Time `json:"end_date,omitempty"`
	Count      int       `json:"count,omitempty"` // alternative to end_date
}

// AutoScheduleRequest represents auto-scheduling configuration
type AutoScheduleRequest struct {
	Text      string    `json:"text"`
	Accounts  []string  `json:"accounts"`
	Media     []Media   `json:"media,omitempty"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Slots     int       `json:"slots"` // number of times to post in date range
}

// RecyclePostRequest represents content recycling configuration
type RecyclePostRequest struct {
	PostID    string    `json:"post_id"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Frequency string    `json:"frequency"`
	MaxCount  int       `json:"max_count"` // maximum times to recycle
}

// RecurringPostResponse contains job ID for recurring post setup
type RecurringPostResponse struct {
	JobID string `json:"job_id"`
}

// AutoScheduleResponse contains job ID for auto-scheduling
type AutoScheduleResponse struct {
	JobID string `json:"job_id"`
}

// RecyclePostResponse contains job ID for recycling setup
type RecyclePostResponse struct {
	JobID string `json:"job_id"`
}