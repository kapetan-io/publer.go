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
	responses        map[string]MockResponse
	errorResponses   map[string]MockErrorResponse
	callCounts       map[string]int
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

	perPage := 10
	total := len(m.posts)
	totalPages := (total + perPage - 1) / perPage

	start := (page - 1) * perPage
	end := start + perPage
	if end > total {
		end = total
	}

	var posts []Post
	if start < total {
		posts = m.posts[start:end]
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

// handlePublishPost handles POST /api/v1/posts/schedule/publish
func (m *MockServer) handlePublishPost(w http.ResponseWriter, r *http.Request) {
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

// SetJobDelay configures job completion delay
func (m *MockServer) SetJobDelay(delay time.Duration) {
	m.SetDelay(delay)
}
