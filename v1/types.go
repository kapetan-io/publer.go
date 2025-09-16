package v1

// User represents a Publer user
type User struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	FirstName string `json:"first_name"`
	Picture   string `json:"picture"`
}

// Post represents a Publer post (basic definition, extended in Phase 1)
type Post struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

// Account represents a social media account (basic definition, extended in Phase 4)
type Account struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Workspace represents a Publer workspace (basic definition, extended in Phase 3)
type Workspace struct {
	ID   string `json:"id"`
	Name string `json:"name"`
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
	PostIDs []string `json:"post_ids"`
	Message string   `json:"message"`
	Error   string   `json:"error,omitempty"`
}

// Media represents media attachment
type Media struct {
	URL  string `json:"url"`
	Type string `json:"type"`
}