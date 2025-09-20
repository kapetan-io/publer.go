package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const defaultBaseURL = "https://app.publer.com/api/v1/"

// Package-level variables for validation
var postIDRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// Config holds client configuration options
type Config struct {
	APIKey      string
	WorkspaceID string
	BaseURL     string
	Client      *http.Client
}

// Client represents the Publer API client
type Client struct {
	config     Config
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new Publer API client
func NewClient(config Config) (*Client, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}
	if config.WorkspaceID == "" {
		return nil, fmt.Errorf("workspace ID is required")
	}

	httpClient := config.Client
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}

	return &Client{
		config:     config,
		httpClient: httpClient,
		baseURL:    baseURL,
	}, nil
}

// do performs HTTP requests with authentication
func (c *Client) do(ctx context.Context, method, path string, body any, result any) error {
	// Build the full URL
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return fmt.Errorf("invalid base URL: %w", err)
	}

	// Ensure path doesn't start with /
	path = strings.TrimPrefix(path, "/")

	rel, err := url.Parse(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	fullURL := u.ResolveReference(rel).String()

	// Prepare request body
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, fullURL, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer-API %s", c.config.APIKey))
	req.Header.Set("Publer-Workspace-Id", c.config.WorkspaceID)

	// Add content type for JSON
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle errors
	if resp.StatusCode >= 400 {
		if resp.StatusCode == 429 {
			// Rate limit error
			rateLimitErr := &RateLimitError{
				APIError: APIError{
					Method:     method,
					URL:        fullURL,
					StatusCode: resp.StatusCode,
				},
			}

			// Parse rate limit headers safely
			if limit := resp.Header.Get("X-RateLimit-Limit"); limit != "" {
				if n, err := fmt.Sscanf(limit, "%d", &rateLimitErr.Limit); n != 1 || err != nil {
					rateLimitErr.Limit = 0
				}
			}
			if remaining := resp.Header.Get("X-RateLimit-Remaining"); remaining != "" {
				if n, err := fmt.Sscanf(remaining, "%d", &rateLimitErr.Remaining); n != 1 || err != nil {
					rateLimitErr.Remaining = 0
				}
			}
			if reset := resp.Header.Get("X-RateLimit-Reset"); reset != "" {
				if n, err := fmt.Sscanf(reset, "%d", &rateLimitErr.Reset); n != 1 || err != nil {
					rateLimitErr.Reset = 0
				}
			}

			// Try to parse error message from body
			var errResp ErrorResponse
			if err := json.Unmarshal(respBody, &errResp); err == nil {
				rateLimitErr.Message = errResp.Message
				if rateLimitErr.Message == "" {
					rateLimitErr.Message = errResp.Error
				}
			}

			return rateLimitErr
		}

		// Regular API error
		apiErr := &APIError{
			Method:     method,
			URL:        fullURL,
			StatusCode: resp.StatusCode,
		}

		// Try to parse error message from body
		var errResp ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil {
			apiErr.Message = errResp.Message
			if apiErr.Message == "" {
				apiErr.Message = errResp.Error
			}
		}

		if apiErr.Message == "" {
			apiErr.Message = string(respBody)
		}

		return apiErr
	}

	// Parse successful response
	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// Test performs a test request to verify connectivity (for testing purposes only)
func (c *Client) Test(ctx context.Context) error {
	var result map[string]interface{}
	return c.do(ctx, "GET", "test", nil, &result)
}

// ============================================================================
// Post Publishing Operations
// ============================================================================

// PublishPost publishes content immediately
func (c *Client) PublishPost(ctx context.Context, request PublishPostRequest, response *PublishPostResponse) error {
	return c.do(ctx, "POST", "posts/schedule/publish", request, response)
}

// BulkPublishPosts publishes multiple posts immediately
func (c *Client) BulkPublishPosts(ctx context.Context, req BulkPublishPostsRequest, resp *BulkPublishPostsResponse) error {
	return c.do(ctx, "POST", "posts/schedule/publish", req, resp)
}

// ============================================================================
// Post Scheduling Operations
// ============================================================================

// SchedulePost schedules a post for future publication
func (c *Client) SchedulePost(ctx context.Context, req SchedulePostRequest, resp *SchedulePostResponse) error {
	return c.do(ctx, "POST", "posts/schedule", req, resp)
}

// CreateDraftPost creates a draft post
func (c *Client) CreateDraftPost(ctx context.Context, req CreateDraftPostRequest, resp *CreateDraftPostResponse) error {
	return c.do(ctx, "POST", "posts/schedule", req, resp)
}

// BulkSchedulePosts schedules multiple posts
func (c *Client) BulkSchedulePosts(ctx context.Context, req BulkSchedulePostsRequest, resp *BulkSchedulePostsResponse) error {
	return c.do(ctx, "POST", "posts/schedule", req, resp)
}

// ============================================================================
// Post Management Operations
// ============================================================================

// Client-side validation is necessary to prevent path traversal attacks when constructing URLs.
// Without validation, malicious PostIDs like "../admin" could access unintended endpoints.
func validatePostID(postID string) error {
	if postID == "" {
		return fmt.Errorf("post ID cannot be empty")
	}
	if strings.Contains(postID, "..") || strings.Contains(postID, "/") || strings.Contains(postID, "\\") {
		return fmt.Errorf("post ID contains invalid characters")
	}
	if !postIDRegex.MatchString(postID) {
		return fmt.Errorf("post ID must contain only alphanumeric characters, hyphens, and underscores")
	}
	return nil
}

// GetPost retrieves a single post by ID
func (c *Client) GetPost(ctx context.Context, req GetPostRequest, resp *GetPostResponse) error {
	if err := validatePostID(req.PostID); err != nil {
		return fmt.Errorf("invalid post ID: %w", err)
	}
	path := fmt.Sprintf("posts/%s", req.PostID)
	return c.do(ctx, "GET", path, nil, resp)
}

// UpdatePost updates an existing post
func (c *Client) UpdatePost(ctx context.Context, req UpdatePostRequest, resp *UpdatePostResponse) error {
	if err := validatePostID(req.PostID); err != nil {
		return fmt.Errorf("invalid post ID: %w", err)
	}
	path := fmt.Sprintf("posts/%s", req.PostID)
	return c.do(ctx, "PATCH", path, req, resp)
}

// DeletePost deletes a post
func (c *Client) DeletePost(ctx context.Context, req DeletePostRequest, resp *DeletePostResponse) error {
	if err := validatePostID(req.PostID); err != nil {
		return fmt.Errorf("invalid post ID: %w", err)
	}
	path := fmt.Sprintf("posts/%s", req.PostID)
	return c.do(ctx, "DELETE", path, nil, resp)
}

// ============================================================================
// Post Listing Operations
// ============================================================================

// ListPosts retrieves posts with filtering options
func (c *Client) ListPosts(ctx context.Context, request ListPostsRequest) Iterator[Post] {
	return NewPostIterator(c, request)
}

// ============================================================================
// Post Advanced Operations
// ============================================================================

// CreateRecurringPost creates a recurring post schedule
func (c *Client) CreateRecurringPost(ctx context.Context, req RecurringPostRequest, resp *RecurringPostResponse) error {
	return c.do(ctx, "POST", "posts/recurring", req, resp)
}

// AutoSchedulePost uses AI to determine optimal posting times
func (c *Client) AutoSchedulePost(ctx context.Context, req AutoScheduleRequest, resp *AutoScheduleResponse) error {
	return c.do(ctx, "POST", "posts/auto-schedule", req, resp)
}

// RecyclePost configures content recycling schedule
func (c *Client) RecyclePost(ctx context.Context, req RecyclePostRequest, resp *RecyclePostResponse) error {
	return c.do(ctx, "POST", "posts/recycle", req, resp)
}

// ============================================================================
// Post Convenience Operations
// ============================================================================

// GetPostsByState returns an iterator for posts filtered by state
func (c *Client) GetPostsByState(state string) Iterator[Post] {
	req := ListPostsRequest{
		State: state,
	}
	return c.ListPosts(context.Background(), req)
}

// GetPostsByDateRange returns an iterator for posts within date range
func (c *Client) GetPostsByDateRange(from, to time.Time) Iterator[Post] {
	req := ListPostsRequest{
		From: from,
		To:   to,
	}
	return c.ListPosts(context.Background(), req)
}

// GetPostsByAccount returns posts for specific account
func (c *Client) GetPostsByAccount(accountID string) Iterator[Post] {
	req := ListPostsRequest{
		AccountIDs: []string{accountID},
	}
	return c.ListPosts(context.Background(), req)
}

// GetPostsByQuery returns posts matching search query
func (c *Client) GetPostsByQuery(query string) Iterator[Post] {
	req := ListPostsRequest{
		Query: query,
	}
	return c.ListPosts(context.Background(), req)
}

// ============================================================================
// Account Operations
// ============================================================================

// ListAccountsRequest represents request for listing accounts
type ListAccountsRequest struct{}

// ListAccountsResponse represents account list response
type ListAccountsResponse struct {
	Accounts   []Account `json:"accounts"`
	Total      int       `json:"total"`
	Page       int       `json:"page"`
	PerPage    int       `json:"per_page"`
	TotalPages int       `json:"total_pages"`
}

// accountFetcher implements PageFetcher for accounts
type accountFetcher struct {
	client *Client
	req    ListAccountsRequest
}

// FetchPage implements PageFetcher interface
func (f *accountFetcher) FetchPage(ctx context.Context, pageNum int) (*Page[Account], error) {
	path := "accounts"
	if pageNum > 1 {
		path = fmt.Sprintf("accounts?page=%d", pageNum)
	}

	var resp ListAccountsResponse
	if err := f.client.do(ctx, "GET", path, nil, &resp); err != nil {
		return nil, err
	}

	return &Page[Account]{
		Items:      resp.Accounts,
		Total:      resp.Total,
		Page:       resp.Page,
		PerPage:    resp.PerPage,
		TotalPages: resp.TotalPages,
	}, nil
}

// ListAccounts retrieves all social media accounts in the workspace
func (c *Client) ListAccounts(ctx context.Context, req ListAccountsRequest) Iterator[Account] {
	fetcher := &accountFetcher{
		client: c,
		req:    req,
	}
	return NewGenericIterator[Account](fetcher)
}

// ============================================================================
// User Operations
// ============================================================================

// GetMeRequest represents request for current user
type GetMeRequest struct{}

// GetMeResponse represents current user response
type GetMeResponse struct {
	User
}

// GetMe retrieves information about the currently authenticated user
func (c *Client) GetMe(ctx context.Context, req GetMeRequest, resp *GetMeResponse) error {
	return c.do(ctx, "GET", "users/me", nil, resp)
}

// ============================================================================
// Workspace Operations
// ============================================================================

// ListWorkspacesRequest represents request for listing workspaces
type ListWorkspacesRequest struct{}

// ListWorkspacesResponse represents workspace list response
type ListWorkspacesResponse struct {
	Workspaces []Workspace `json:"workspaces"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	PerPage    int         `json:"per_page"`
	TotalPages int         `json:"total_pages"`
}

// workspacePageFetcher implements PageFetcher for workspaces
type workspacePageFetcher struct {
	client *Client
}

// FetchPage fetches a page of workspaces
func (f *workspacePageFetcher) FetchPage(ctx context.Context, pageNum int) (*Page[Workspace], error) {
	path := "workspaces"
	if pageNum > 1 {
		path = fmt.Sprintf("workspaces?page=%s", strconv.Itoa(pageNum))
	}

	var resp ListWorkspacesResponse
	if err := f.client.do(ctx, "GET", path, nil, &resp); err != nil {
		return nil, err
	}

	return &Page[Workspace]{
		Items:      resp.Workspaces,
		Total:      resp.Total,
		Page:       resp.Page,
		PerPage:    resp.PerPage,
		TotalPages: resp.TotalPages,
	}, nil
}

// ListWorkspaces retrieves all workspaces for the authenticated user
func (c *Client) ListWorkspaces(ctx context.Context, req ListWorkspacesRequest) Iterator[Workspace] {
	fetcher := &workspacePageFetcher{client: c}
	return NewGenericIterator(fetcher)
}

// ============================================================================
// Job Management Operations
// ============================================================================

// GetJobStatusRequest requests job status
type GetJobStatusRequest struct {
	JobID string
}

// GetJobStatusResponse contains job status
type GetJobStatusResponse struct {
	JobStatus
}

// WaitOptions configures job polling behavior
type WaitOptions struct {
	JobID        string
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Jitter       time.Duration
}

// GetJobStatus checks status of async job
func (c *Client) GetJobStatus(ctx context.Context, req GetJobStatusRequest, resp *GetJobStatusResponse) error {
	path := fmt.Sprintf("job_status/%s", req.JobID)
	return c.do(ctx, "GET", path, nil, resp)
}

// WaitForJob polls job status until completion with configurable timing
func (c *Client) WaitForJob(ctx context.Context, opts WaitOptions, result *JobResult) error {
	initialDelay := opts.InitialDelay
	if initialDelay == 0 {
		initialDelay = time.Second
	}
	maxDelay := opts.MaxDelay
	if maxDelay == 0 {
		maxDelay = 30 * time.Second
	}
	jitter := opts.Jitter
	if jitter == 0 {
		jitter = 500 * time.Millisecond
	}

	delay := initialDelay
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			var statusResp GetJobStatusResponse
			err := c.GetJobStatus(ctx, GetJobStatusRequest{JobID: opts.JobID}, &statusResp)
			if err != nil {
				return err
			}

			switch statusResp.Status {
			case "completed":
				if statusResp.Result != nil {
					*result = *statusResp.Result
				} else {
					*result = JobResult{Success: true}
				}
				return nil
			case "failed", "cancelled":
				if statusResp.Result != nil {
					*result = *statusResp.Result
				} else {
					*result = JobResult{Success: false, Error: statusResp.Error}
				}
				return fmt.Errorf("job %s: %s", statusResp.Status, statusResp.Error)
			case "pending", "working", "processing":
				if delay < maxDelay {
					delay *= 2
					if delay > maxDelay {
						delay = maxDelay
					}
				}
				r := rand.New(rand.NewSource(time.Now().UnixNano()))
				delay += time.Duration(r.Intn(int(jitter/time.Millisecond))) * time.Millisecond
			default:
				return fmt.Errorf("unknown job status: %s", statusResp.Status)
			}
		}
	}
}