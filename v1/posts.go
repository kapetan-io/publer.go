package v1

import "time"

// ListPostsRequest represents request for listing posts
type ListPostsRequest struct {
	State      string    `json:"state,omitempty"`
	States     []string  `json:"state[],omitempty"`
	From       time.Time `json:"from,omitempty"`
	To         time.Time `json:"to,omitempty"`
	Page       int       `json:"page,omitempty"`
	AccountIDs []string  `json:"account_ids[],omitempty"`
	Query      string    `json:"query,omitempty"`
	PostType   string    `json:"postType,omitempty"`
	MemberID   string    `json:"member_id,omitempty"`
}

// ListPostsResponse represents paginated posts response
type ListPostsResponse struct {
	Posts      []Post `json:"posts"`
	Total      int    `json:"total"`
	Page       int    `json:"page"`
	PerPage    int    `json:"per_page"`
	TotalPages int    `json:"total_pages"`
}

// PublishRequest represents immediate post publishing
type PublishRequest struct {
	Text     string   `json:"text"`
	Accounts []string `json:"accounts"`
	Media    []Media  `json:"media,omitempty"`
}

// PublishResponse contains job ID for async processing
type PublishResponse struct {
	JobID string `json:"job_id"`
}
