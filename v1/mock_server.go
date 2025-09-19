package v1

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"time"
)

const defaultPerPage = 10

// MockServer provides a test HTTP server that mimics Publer API
type MockServer struct {
	mu               *sync.RWMutex
	server           *httptest.Server
	apiKey           string
	workspaceID      string
	jobDelay         time.Duration
	jobs             map[string]*JobStatus
	jobProgression   map[string][]JobStatus
	jobProgressIndex map[string]int
	posts            []Post
	accounts         []Account
	workspaces       []Workspace
	currentUser      *User
	responses        map[string]MockResponse
	errorResponses   map[string]MockErrorResponse
	callCounts       map[string]int
	bulkOpLimit      int
}

// MockResponse holds configured response data
type MockResponse struct {
	StatusCode int
	Body       any
}

// MockErrorResponse holds configured error response data
type MockErrorResponse struct {
	StatusCode    int
	Body          any
	Headers       map[string]string
	CallThreshold int // Return error after N calls
	CallCount     int // Current call count for this endpoint
}

// SpawnMockServer creates and starts a new mock server instance
func SpawnMockServer() *MockServer {
	m := &MockServer{
		mu:               &sync.RWMutex{},
		apiKey:           "mock-api-key-" + strconv.FormatInt(time.Now().UnixNano(), 36),
		workspaceID:      "mock-workspace-" + strconv.FormatInt(time.Now().UnixNano(), 36),
		jobs:             make(map[string]*JobStatus),
		jobProgression:   make(map[string][]JobStatus),
		jobProgressIndex: make(map[string]int),
		responses:        make(map[string]MockResponse),
		errorResponses:   make(map[string]MockErrorResponse),
		callCounts:       make(map[string]int),
	}

	m.server = httptest.NewServer(http.HandlerFunc(m.handleRequest))
	return m
}

// Client returns a new Client instance configured to use this mock server
func (m *MockServer) Client() *Client {
	client, _ := NewClient(Config{
		APIKey:      m.apiKey,
		WorkspaceID: m.workspaceID,
		BaseURL:     m.server.URL + "/api/v1/",
	})
	return client
}

// Stop stops the mock HTTP server
func (m *MockServer) Stop() error {
	if m.server == nil {
		return fmt.Errorf("server not started")
	}

	m.server.Close()
	m.server = nil
	return nil
}

// Reset clears all mock server state for next test
func (m *MockServer) Reset() {

	m.mu.Lock()
	defer m.mu.Unlock()

	m.jobs = make(map[string]*JobStatus)
	m.jobProgression = make(map[string][]JobStatus)
	m.jobProgressIndex = make(map[string]int)
	m.posts = []Post{}
	m.accounts = []Account{}
	m.workspaces = []Workspace{}
	m.currentUser = nil
	m.responses = make(map[string]MockResponse)
	m.errorResponses = make(map[string]MockErrorResponse)
	m.callCounts = make(map[string]int)
	m.jobDelay = 0
}

// SetResponse configures expected response for specific endpoint
func (m *MockServer) SetResponse(method, path string, statusCode int, body any) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s %s", method, path)
	m.responses[key] = MockResponse{
		StatusCode: statusCode,
		Body:       body,
	}
}

// SetErrorResponse configures error response after N calls to endpoint
func (m *MockServer) SetErrorResponse(method, path string, callThreshold int, statusCode int, body any, headers map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s %s", method, path)
	m.errorResponses[key] = MockErrorResponse{
		StatusCode:    statusCode,
		Body:          body,
		Headers:       headers,
		CallThreshold: callThreshold,
		CallCount:     0,
	}
}

// SetJobStatus configures job status response for job ID
func (m *MockServer) SetJobStatus(jobID, status string, progress int, result *JobResult, err string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.jobs[jobID] = &JobStatus{
		ID:       jobID,
		Status:   status,
		Progress: progress,
		Result:   result,
		Error:    err,
	}
}

// SetJobProgression configures automatic job state progression
func (m *MockServer) SetJobProgression(jobID string, states []JobStatus) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.jobProgression[jobID] = states
	m.jobProgressIndex[jobID] = 0
}

// AdvanceJobState manually advances job to next state in progression
func (m *MockServer) AdvanceJobState(jobID string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	states, exists := m.jobProgression[jobID]
	if !exists {
		return false
	}

	index := m.jobProgressIndex[jobID]
	if index < len(states)-1 {
		m.jobProgressIndex[jobID]++
		m.jobs[jobID] = &states[m.jobProgressIndex[jobID]]
		return true
	}

	return false
}

// SetDelay adds artificial delay to responses (bypassed in fast test mode)
func (m *MockServer) SetDelay(delay time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.jobDelay = delay
}

// AddPosts adds posts to mock data for listing endpoints
func (m *MockServer) AddPosts(posts []Post) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.posts = append(m.posts, posts...)
}

// AddAccounts adds accounts to mock data for listing endpoints
func (m *MockServer) AddAccounts(accounts []Account) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.accounts = append(m.accounts, accounts...)
}

// AddWorkspaces adds workspaces to mock data for listing endpoints
func (m *MockServer) AddWorkspaces(workspaces []Workspace) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.workspaces = append(m.workspaces, workspaces...)
}

// SetCurrentUser sets the mock current user
func (m *MockServer) SetCurrentUser(user User) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.currentUser = &user
}

// AddWorkspace adds a workspace to mock data
func (m *MockServer) AddWorkspace(workspace Workspace) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.workspaces = append(m.workspaces, workspace)
}

// AddScheduledPost adds a scheduled post to mock data
func (m *MockServer) AddScheduledPost(post Post) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.posts = append(m.posts, post)
}

// handleRequest routes requests to appropriate handlers
func (m *MockServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	// Apply delay before acquiring lock to avoid holding lock during sleep
	m.mu.RLock()
	delay := m.jobDelay
	m.mu.RUnlock()

	if delay > 0 {
		time.Sleep(delay)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate authentication headers
	authHeader := r.Header.Get("Authorization")
	expectedAuth := "Bearer-API " + m.apiKey
	if authHeader != expectedAuth {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "unauthorized",
			Message: "Missing or invalid API key",
		})
		return
	}

	workspaceHeader := r.Header.Get("Publer-Workspace-Id")
	if workspaceHeader != m.workspaceID {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "bad_request",
			Message: "Missing or invalid workspace ID",
		})
		return
	}

	// Track call counts
	key := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
	m.callCounts[key]++

	// Check for error response configuration
	if errResp, exists := m.errorResponses[key]; exists {
		if m.callCounts[key] >= errResp.CallThreshold {
			// Write error headers
			for k, v := range errResp.Headers {
				w.Header().Set(k, v)
			}

			w.WriteHeader(errResp.StatusCode)
			if errResp.Body != nil {
				json.NewEncoder(w).Encode(errResp.Body)
			}
			return
		}
	}

	// Check for configured response
	if resp, exists := m.responses[key]; exists {
		w.WriteHeader(resp.StatusCode)
		if resp.Body != nil {
			json.NewEncoder(w).Encode(resp.Body)
		}
		return
	}

	// Handle job status requests
	if strings.HasPrefix(r.URL.Path, "/api/v1/job_status/") {
		m.handleJobStatus(w, r)
		return
	}

	// Handle posts operations
	if r.URL.Path == "/api/v1/posts" && r.Method == "GET" {
		m.handleListPosts(w, r)
		return
	}

	// Handle post publishing
	if r.URL.Path == "/api/v1/posts/schedule/publish" && r.Method == "POST" {
		m.handlePublishPost(w, r)
		return
	}

	// Handle post scheduling and drafts
	if r.URL.Path == "/api/v1/posts/schedule" && r.Method == "POST" {
		m.handleSchedulePost(w, r)
		return
	}

	// Handle recurring posts
	if r.URL.Path == "/api/v1/posts/recurring" && r.Method == "POST" {
		m.handleRecurringPost(w, r)
		return
	}

	// Handle auto-scheduling
	if r.URL.Path == "/api/v1/posts/auto-schedule" && r.Method == "POST" {
		m.handleAutoSchedulePost(w, r)
		return
	}

	// Handle post recycling
	if r.URL.Path == "/api/v1/posts/recycle" && r.Method == "POST" {
		m.handleRecyclePost(w, r)
		return
	}

	// Handle post management operations
	if strings.HasPrefix(r.URL.Path, "/api/v1/posts/") && len(strings.Split(r.URL.Path, "/")) == 5 {
		// Extract post ID from path: /api/v1/posts/{id}
		parts := strings.Split(r.URL.Path, "/")
		postID := parts[4]

		switch r.Method {
		case "GET":
			m.handleGetPost(w, r, postID)
			return
		case "PATCH":
			m.handleUpdatePost(w, r, postID)
			return
		case "DELETE":
			m.handleDeletePost(w, r, postID)
			return
		}
	}

	// Handle user operations
	if r.URL.Path == "/api/v1/users/me" && r.Method == "GET" {
		m.handleGetMe(w, r)
		return
	}

	// Handle workspace operations
	if r.URL.Path == "/api/v1/workspaces" && r.Method == "GET" {
		m.handleListWorkspaces(w, r)
		return
	}

	// Handle account operations
	if r.URL.Path == "/api/v1/accounts" && r.Method == "GET" {
		m.handleListAccounts(w, r)
		return
	}

	// Default 404
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   "not_found",
		Message: "Endpoint not found",
	})
}

// handleListPosts handles GET /api/v1/posts
func (m *MockServer) handleListPosts(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	page := 1
	if pageStr != "" {
		page, _ = strconv.Atoi(pageStr)
	}

	// Apply filters
	filteredPosts := m.filterPosts(r)

	perPage := defaultPerPage
	total := len(filteredPosts)
	totalPages := (total + perPage - 1) / perPage
	if totalPages == 0 {
		totalPages = 1
	}

	start := (page - 1) * perPage
	end := start + perPage
	if end > total {
		end = total
	}

	var posts []Post
	if start < total {
		posts = filteredPosts[start:end]
	} else {
		posts = []Post{}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ListPostsResponse{
		Posts:      posts,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	})
}

// filterPosts applies query parameter filters to posts
func (m *MockServer) filterPosts(r *http.Request) []Post {
	var filtered []Post

	state := r.URL.Query().Get("state")
	states := r.URL.Query()["state[]"]
	query := r.URL.Query().Get("query")
	accountIDs := r.URL.Query()["account_ids[]"]
	postType := r.URL.Query().Get("postType")
	memberID := r.URL.Query().Get("member_id")
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	var fromTime, toTime time.Time
	var err error
	if fromStr != "" {
		fromTime, err = time.Parse(time.RFC3339, fromStr)
		if err != nil {
			fromTime = time.Time{}
		}
	}
	if toStr != "" {
		toTime, err = time.Parse(time.RFC3339, toStr)
		if err != nil {
			toTime = time.Time{}
		}
	}

	for _, post := range m.posts {
		// Filter by state (single state)
		if state != "" && post.State != state {
			continue
		}

		// Filter by states (multiple states)
		if len(states) > 0 {
			found := false
			for _, s := range states {
				if post.State == s {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Filter by query (simple text search)
		if query != "" && !strings.Contains(strings.ToLower(post.Text), strings.ToLower(query)) {
			continue
		}

		// Filter by account IDs
		if len(accountIDs) > 0 {
			found := false
			for _, accountID := range accountIDs {
				if post.AccountID == accountID {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Filter by post type
		if postType != "" && post.Type != postType {
			continue
		}

		// Filter by member ID (in this mock, we'll assume User.ID represents member)
		if memberID != "" && post.User.ID != memberID {
			continue
		}

		// Filter by date range
		if !fromTime.IsZero() && post.ScheduledAt.Before(fromTime) {
			continue
		}
		if !toTime.IsZero() && post.ScheduledAt.After(toTime) {
			continue
		}

		filtered = append(filtered, post)
	}

	return filtered
}

// handlePublishPost handles POST /api/v1/posts/schedule/publish
func (m *MockServer) handlePublishPost(w http.ResponseWriter, r *http.Request) {
	// Read the entire request body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "bad_request",
			Message: "Failed to read request body",
		})
		return
	}

	var requestData map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &requestData); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid JSON payload",
		})
		return
	}

	// Check if this is a bulk operation (has "posts" array)
	if postsData, isBulk := requestData["posts"]; isBulk {
		m.handleBulkPublish(w, r, bodyBytes, postsData)
		return
	}

	// Handle single post publish
	jobID := "job-" + strconv.FormatInt(time.Now().UnixNano(), 36)

	// Set default job status
	m.jobs[jobID] = &JobStatus{
		ID:       jobID,
		Status:   "pending",
		Progress: 0,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(PublishPostResponse{
		JobID: jobID,
	})
}

// handleBulkPublish handles bulk publishing requests
func (m *MockServer) handleBulkPublish(w http.ResponseWriter, r *http.Request, bodyBytes []byte, postsData interface{}) {
	var bulkReq BulkPublishPostsRequest
	if err := json.Unmarshal(bodyBytes, &bulkReq); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid bulk publish request format",
		})
		return
	}

	// Check bulk operation limit
	if m.bulkOpLimit > 0 && len(bulkReq.Posts) > m.bulkOpLimit {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "bad_request",
			Message: fmt.Sprintf("Bulk operation limit exceeded. Maximum %d posts allowed", m.bulkOpLimit),
		})
		return
	}

	jobID := "job-" + strconv.FormatInt(time.Now().UnixNano(), 36)

	// Set default job status
	m.jobs[jobID] = &JobStatus{
		ID:       jobID,
		Status:   "pending",
		Progress: 0,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(BulkPublishPostsResponse{
		JobID: jobID,
	})
}

// handleJobStatus handles GET /api/v1/job_status/{job_id}
func (m *MockServer) handleJobStatus(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid job ID",
		})
		return
	}

	jobID := parts[4]

	// Check job progression first
	if states, exists := m.jobProgression[jobID]; exists {
		index := m.jobProgressIndex[jobID]
		if index < len(states) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(GetJobStatusResponse{
				JobStatus: states[index],
			})
			return
		}
	}

	// Check regular job status
	if job, exists := m.jobs[jobID]; exists {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(GetJobStatusResponse{
			JobStatus: *job,
		})
		return
	}

	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   "not_found",
		Message: "Job not found",
	})
}

// handleSchedulePost handles POST /api/v1/posts/schedule
func (m *MockServer) handleSchedulePost(w http.ResponseWriter, r *http.Request) {
	var scheduleReq SchedulePostRequest
	var draftReq CreateDraftPostRequest

	// Read the entire request body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "bad_request",
			Message: "Failed to read request body",
		})
		return
	}

	var requestData map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &requestData); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid JSON payload",
		})
		return
	}

	// Check if this is a bulk operation (has "posts" array)
	if postsData, isBulk := requestData["posts"]; isBulk {
		m.handleBulkSchedule(w, r, bodyBytes, postsData)
		return
	}

	jobID := "job-" + strconv.FormatInt(time.Now().UnixNano(), 36)

	// Set default job status
	m.jobs[jobID] = &JobStatus{
		ID:       jobID,
		Status:   "pending",
		Progress: 0,
	}

	// Check if this is a draft request (has visibility field)
	if _, hasDraft := requestData["visibility"]; hasDraft {
		if err := json.Unmarshal(bodyBytes, &draftReq); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "bad_request",
				Message: "Invalid draft request format",
			})
			return
		}

		// Validate visibility
		if draftReq.Visibility != "draft_private" && draftReq.Visibility != "draft_public" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "bad_request",
				Message: "Invalid visibility. Must be draft_private or draft_public",
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(CreateDraftPostResponse{
			JobID: jobID,
		})
		return
	}

	// Otherwise, treat as schedule request
	if err := json.Unmarshal(bodyBytes, &scheduleReq); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid schedule request format",
		})
		return
	}

	// Validate that scheduled_at is in the future
	if !scheduleReq.ScheduledAt.After(time.Now()) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "bad_request",
			Message: "Scheduled time must be in the future",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SchedulePostResponse{
		JobID: jobID,
	})
}

// handleBulkSchedule handles bulk scheduling requests
func (m *MockServer) handleBulkSchedule(w http.ResponseWriter, r *http.Request, bodyBytes []byte, postsData interface{}) {
	var bulkReq BulkSchedulePostsRequest
	if err := json.Unmarshal(bodyBytes, &bulkReq); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid bulk schedule request format",
		})
		return
	}

	// Check bulk operation limit
	if m.bulkOpLimit > 0 && len(bulkReq.Posts) > m.bulkOpLimit {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "bad_request",
			Message: fmt.Sprintf("Bulk operation limit exceeded. Maximum %d posts allowed", m.bulkOpLimit),
		})
		return
	}

	// Validate that all scheduled posts have future timestamps
	for i, post := range bulkReq.Posts {
		if !post.ScheduledAt.IsZero() && !post.ScheduledAt.After(time.Now()) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "bad_request",
				Message: fmt.Sprintf("Post %d: Scheduled time must be in the future", i+1),
			})
			return
		}
	}

	jobID := "job-" + strconv.FormatInt(time.Now().UnixNano(), 36)

	// Set default job status
	m.jobs[jobID] = &JobStatus{
		ID:       jobID,
		Status:   "pending",
		Progress: 0,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(BulkSchedulePostsResponse{
		JobID: jobID,
	})
}

// SetJobDelay configures job completion delay
func (m *MockServer) SetJobDelay(delay time.Duration) {
	m.SetDelay(delay)
}

// handleGetMe handles GET /api/v1/users/me
func (m *MockServer) handleGetMe(w http.ResponseWriter, r *http.Request) {
	if m.currentUser == nil {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "not_found",
			Message: "User not found",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(GetMeResponse{
		User: *m.currentUser,
	})
}

// handleListWorkspaces handles GET /api/v1/workspaces
func (m *MockServer) handleListWorkspaces(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	page := 1
	if pageStr != "" {
		page, _ = strconv.Atoi(pageStr)
	}

	perPage := defaultPerPage
	total := len(m.workspaces)
	totalPages := (total + perPage - 1) / perPage

	start := (page - 1) * perPage
	end := start + perPage
	if end > total {
		end = total
	}

	var workspaces []Workspace
	if start < total {
		workspaces = m.workspaces[start:end]
	} else {
		workspaces = []Workspace{}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ListWorkspacesResponse{
		Workspaces: workspaces,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	})
}

// handleListAccounts handles GET /api/v1/accounts
func (m *MockServer) handleListAccounts(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	page := 1
	if pageStr != "" {
		page, _ = strconv.Atoi(pageStr)
	}

	perPage := defaultPerPage
	total := len(m.accounts)
	totalPages := (total + perPage - 1) / perPage

	start := (page - 1) * perPage
	end := start + perPage
	if end > total {
		end = total
	}

	var accounts []Account
	if start < total {
		accounts = m.accounts[start:end]
	} else {
		accounts = []Account{}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ListAccountsResponse{
		Accounts:   accounts,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	})
}

// AddAccount adds a social media account to mock data
func (m *MockServer) AddAccount(account Account) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.accounts = append(m.accounts, account)
}

// SetAccountsByProvider sets accounts filtered by provider
func (m *MockServer) SetAccountsByProvider(provider string, accounts []Account) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Clear existing accounts for this provider and add new ones
	var filteredAccounts []Account
	for _, acc := range m.accounts {
		if acc.Provider != provider {
			filteredAccounts = append(filteredAccounts, acc)
		}
	}

	m.accounts = append(filteredAccounts, accounts...)
}

// SetBulkOperationLimit sets maximum posts per bulk operation
func (m *MockServer) SetBulkOperationLimit(limit int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.bulkOpLimit = limit
}

// handleGetPost handles GET /api/v1/posts/{id}
func (m *MockServer) handleGetPost(w http.ResponseWriter, r *http.Request, postID string) {
	// Find post by ID
	for _, post := range m.posts {
		if post.ID == postID {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(GetPostResponse{Post: post})
			return
		}
	}

	// Post not found
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   "not_found",
		Message: "Post not found",
	})
}

// handleUpdatePost handles PATCH /api/v1/posts/{id}
func (m *MockServer) handleUpdatePost(w http.ResponseWriter, r *http.Request, postID string) {
	// Read request body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "bad_request",
			Message: "Failed to read request body",
		})
		return
	}

	var updateReq UpdatePostRequest
	if err := json.Unmarshal(bodyBytes, &updateReq); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid JSON in request body",
		})
		return
	}

	// Find and update post
	for i, post := range m.posts {
		if post.ID == postID {
			// Apply partial updates
			if updateReq.Text != "" {
				m.posts[i].Text = updateReq.Text
			}
			if !updateReq.ScheduledAt.IsZero() {
				m.posts[i].ScheduledAt = updateReq.ScheduledAt
			}
			if updateReq.Media != nil {
				m.posts[i].HasMedia = len(updateReq.Media) > 0
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(UpdatePostResponse{Post: m.posts[i]})
			return
		}
	}

	// Post not found
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   "not_found",
		Message: "Post not found",
	})
}

// handleDeletePost handles DELETE /api/v1/posts/{id}
func (m *MockServer) handleDeletePost(w http.ResponseWriter, r *http.Request, postID string) {
	// Find post index to remove
	foundIndex := -1
	for i, post := range m.posts {
		if post.ID == postID {
			foundIndex = i
			break
		}
	}

	if foundIndex == -1 {
		// Post not found
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "not_found",
			Message: "Post not found",
		})
		return
	}

	// Remove post from slice safely
	if foundIndex == len(m.posts)-1 {
		// Last element - just truncate
		m.posts = m.posts[:foundIndex]
	} else {
		// Copy remaining elements
		copy(m.posts[foundIndex:], m.posts[foundIndex+1:])
		m.posts = m.posts[:len(m.posts)-1]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(DeletePostResponse{
		Success: true,
		Message: "Post deleted successfully",
	})
}

// UpdateMockPost updates a post in mock data
func (m *MockServer) UpdateMockPost(id string, updates map[string]any) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, post := range m.posts {
		if post.ID == id {
			// Apply updates based on map
			if text, ok := updates["text"].(string); ok {
				m.posts[i].Text = text
			}
			if scheduledAt, ok := updates["scheduled_at"].(time.Time); ok {
				m.posts[i].ScheduledAt = scheduledAt
			}
			if state, ok := updates["state"].(string); ok {
				m.posts[i].State = state
			}
			break
		}
	}
}

// handleRecurringPost handles POST /api/v1/posts/recurring
func (m *MockServer) handleRecurringPost(w http.ResponseWriter, r *http.Request) {
	var req RecurringPostRequest

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "bad_request",
			Message: "Failed to read request body",
		})
		return
	}

	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid JSON payload",
		})
		return
	}

	if req.Text == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "bad_request",
			Message: "Text field is required",
		})
		return
	}

	if len(req.Accounts) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "bad_request",
			Message: "At least one account is required",
		})
		return
	}

	if req.Recurrence.Frequency == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "bad_request",
			Message: "Recurrence frequency is required",
		})
		return
	}

	jobID := fmt.Sprintf("recurring-%d", time.Now().UnixNano())

	response := RecurringPostResponse{
		JobID: jobID,
	}

	m.jobs[jobID] = &JobStatus{
		ID:       jobID,
		Status:   "in_progress",
		Progress: 0,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// handleAutoSchedulePost handles POST /api/v1/posts/auto-schedule
func (m *MockServer) handleAutoSchedulePost(w http.ResponseWriter, r *http.Request) {
	var req AutoScheduleRequest

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "bad_request",
			Message: "Failed to read request body",
		})
		return
	}

	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid JSON payload",
		})
		return
	}

	if req.Text == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "bad_request",
			Message: "Text field is required",
		})
		return
	}

	if len(req.Accounts) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "bad_request",
			Message: "At least one account is required",
		})
		return
	}

	if req.Slots <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "bad_request",
			Message: "Slots must be greater than 0",
		})
		return
	}

	if req.EndDate.Before(req.StartDate) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "bad_request",
			Message: "End date must be after start date",
		})
		return
	}

	jobID := fmt.Sprintf("auto-schedule-%d", time.Now().UnixNano())

	response := AutoScheduleResponse{
		JobID: jobID,
	}

	m.jobs[jobID] = &JobStatus{
		ID:       jobID,
		Status:   "in_progress",
		Progress: 0,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// handleRecyclePost handles POST /api/v1/posts/recycle
func (m *MockServer) handleRecyclePost(w http.ResponseWriter, r *http.Request) {
	var req RecyclePostRequest

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "bad_request",
			Message: "Failed to read request body",
		})
		return
	}

	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid JSON payload",
		})
		return
	}

	if req.PostID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "bad_request",
			Message: "Post ID is required",
		})
		return
	}

	if req.Frequency == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "bad_request",
			Message: "Frequency is required",
		})
		return
	}

	if req.MaxCount <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "bad_request",
			Message: "Max count must be greater than 0",
		})
		return
	}

	if req.EndDate.Before(req.StartDate) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "bad_request",
			Message: "End date must be after start date",
		})
		return
	}

	found := false
	for _, post := range m.posts {
		if post.ID == req.PostID {
			found = true
			break
		}
	}

	if !found {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "not_found",
			Message: "Post not found",
		})
		return
	}

	jobID := fmt.Sprintf("recycle-%d", time.Now().UnixNano())

	response := RecyclePostResponse{
		JobID: jobID,
	}

	m.jobs[jobID] = &JobStatus{
		ID:       jobID,
		Status:   "in_progress",
		Progress: 0,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// SimulateScheduleGeneration creates mock scheduled posts for advanced features
func (m *MockServer) SimulateScheduleGeneration(count int, interval time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	baseTime := time.Now()
	for i := 0; i < count; i++ {
		post := Post{
			ID:          fmt.Sprintf("scheduled-%d-%d", time.Now().UnixNano(), i),
			Text:        fmt.Sprintf("Scheduled post %d", i+1),
			State:       "scheduled",
			Type:        "post",
			AccountID:   "test-account",
			ScheduledAt: baseTime.Add(time.Duration(i) * interval),
			Network:     "twitter",
		}
		m.posts = append(m.posts, post)
	}
}
