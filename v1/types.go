package v1

import "time"

// User represents a Publer user
type User struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	FirstName string `json:"first_name"`
	Picture   string `json:"picture"`
}

// Post represents a Publer post
type Post struct {
	ID          string    `json:"id"`
	Text        string    `json:"text"`
	URL         string    `json:"url"`
	State       string    `json:"state"`
	Type        string    `json:"type"`
	AccountID   string    `json:"account_id"`
	User        User      `json:"user"`
	ScheduledAt time.Time `json:"scheduled_at"`
	PostLink    string    `json:"post_link"`
	HasMedia    bool      `json:"has_media"`
	Network     string    `json:"network"`
}

// Account represents a social media account
type Account struct {
	ID       string `json:"id"`
	Provider string `json:"provider"`
	Name     string `json:"name"`
	SocialID string `json:"social_id"`
	Picture  string `json:"picture"`
	Type     string `json:"type"`
}

// Workspace represents a Publer workspace
type Workspace struct {
	ID      string `json:"id"`
	Owner   User   `json:"owner"`
	Name    string `json:"name"`
	Members []User `json:"members"`
	Plan    string `json:"plan"`
	Picture string `json:"picture"`
}

// JobStatus represents async job status (basic definition, extended in Phase 1)
type JobStatus struct {
	ID       string     `json:"id"`
	Status   string     `json:"status"`
	Progress int        `json:"progress"`
	Result   *JobResult `json:"result,omitempty"`
	Error    string     `json:"error,omitempty"`
}

// JobResult contains job completion data
type JobResult struct {
	Success bool                   `json:"success"`
	PostIDs []string               `json:"post_ids"`
	Message string                 `json:"message"`
	Error   string                 `json:"error,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

// Media represents media attachment
type Media struct {
	URL  string `json:"url"`
	Type string `json:"type"`
}