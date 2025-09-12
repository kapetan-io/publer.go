# Publer.com Go Client Library Implementation Plan

## Overview

This plan details the implementation of a Go client library for the Publer.com HTTP API v1. The library will provide a type-safe, context-aware client with custom iterator support for paginated resources, proper rate limit handling, and a comprehensive mock server for testing.

## Current State Analysis

**What exists now:** Empty project with only `.git` and `plans/` directories
**What's missing:** Complete Go client library implementation
**Key constraints discovered:**
- API requires Bearer API key + Workspace ID for authentication
- Business users only (no public API access)
- RESTful JSON API with consistent patterns
- Standard HTTP status codes with JSON error responses
- Post creation is asynchronous, returns job IDs

### Key Discoveries:
- API base URL: `https://app.publer.com/api/v1/`
- Authentication headers: `Authorization: Bearer-API {key}` and `Publer-Workspace-Id: {id}`
- Rate limits: 100 requests per 2 minutes with `X-RateLimit-*` headers
- Four main resource areas: Users, Workspaces, Accounts, Posts
- Posts endpoint supports extensive filtering via query parameters
- Post creation returns job IDs for async processing

## Desired End State

A complete Go client library that:
- Exports as `publer.NewClient()` from versioned `/v1/` package
- Supports custom HTTP client injection
- Provides context-based request cancellation for all operations  
- Implements custom iterators for paginated resources
- Surfaces rate limit information during 429 errors
- Follows established error format: `POST https://host:port/slug with 500 returned "Error"`
- Includes comprehensive mock server for testing

**Verification:** Client can successfully authenticate, list resources, handle pagination, create posts, check job status, and surface rate limits.

## What We're NOT Doing

- CLI tools or command-line interface
- Using `pkg/` directory structure
- Request timeout or retry logic (handled by caller via context)
- Rate limit headers in successful responses
- Unit tests with HTTP client mocking (all tests use mock server)

## Complete Public Interface

```go
package publer

// Client creation
func NewClient(config Config) (*Client, error)

// User operations
func (c *Client) GetMe(ctx context.Context, req GetMeRequest, resp *GetMeResponse) error

// Workspace operations
func (c *Client) ListWorkspaces(ctx context.Context, req ListWorkspacesRequest) Iterator[Workspace]

// Account operations  
func (c *Client) ListAccounts(ctx context.Context, req ListAccountsRequest) Iterator[Account]

// Post operations - Read
func (c *Client) ListPosts(ctx context.Context, req ListPostsRequest) Iterator[Post]
func (c *Client) GetPost(ctx context.Context, req GetPostRequest, resp *GetPostResponse) error

// Post operations - Create
func (c *Client) PublishPost(ctx context.Context, req PublishPostRequest, resp *PublishPostResponse) error
func (c *Client) SchedulePost(ctx context.Context, req SchedulePostRequest, resp *SchedulePostResponse) error
func (c *Client) CreateDraftPost(ctx context.Context, req CreateDraftPostRequest, resp *CreateDraftPostResponse) error

// Post operations - Bulk
func (c *Client) BulkPublishPosts(ctx context.Context, req BulkPublishPostsRequest, resp *BulkPublishPostsResponse) error
func (c *Client) BulkSchedulePosts(ctx context.Context, req BulkSchedulePostsRequest, resp *BulkSchedulePostsResponse) error

// Post operations - Update/Delete
func (c *Client) UpdatePost(ctx context.Context, req UpdatePostRequest, resp *UpdatePostResponse) error
func (c *Client) DeletePost(ctx context.Context, req DeletePostRequest, resp *DeletePostResponse) error

// Job operations
func (c *Client) GetJobStatus(ctx context.Context, req GetJobStatusRequest, resp *GetJobStatusResponse) error
func (c *Client) WaitForJob(ctx context.Context, jobID string, result *JobResult) error
func (c *Client) GetPostsByState(state string) Iterator[Post]
func (c *Client) GetPostsByDateRange(from, to time.Time) Iterator[Post]

// Mock server for testing
func NewMockServer() *MockServer
func (m *MockServer) Start() (string, error)
func (m *MockServer) Stop() error
func (m *MockServer) Reset()
func (m *MockServer) SetResponse(method, path string, statusCode int, body interface{})
func (m *MockServer) SetRateLimit(limit, remaining int, reset time.Time)
func (m *MockServer) SetJobStatus(jobID, status string, progress int, result *JobResult, err string)
func (m *MockServer) SetDelay(delay time.Duration)
func (m *MockServer) AddPosts(posts []Post)
func (m *MockServer) AddAccounts(accounts []Account)
func (m *MockServer) AddWorkspaces(workspaces []Workspace)
```

## Implementation Approach

1. **Incremental Delivery**: Each phase builds on the previous with working, testable code
2. **Mock-First Testing**: Mock server evolves alongside client implementation
3. **Iterator Pattern**: Generic iterator established early, adapted throughout
4. **Error Handling**: Consistent error types with rate limit metadata
5. **TDD Approach**: Functional tests drive implementation using mock server
6. **Sequential Testing**: All tests run sequentially without t.Parallel() for mock server consistency

## Phase 0: Foundation - Client, Errors, Mock Server, Iterator

### Overview
Establishes core client structure, error handling, basic mock server, and generic iterator interface.

### Changes Required:

#### 1. Core Client Structure
**File**: `v1/client.go`
**Changes**: Create main client with configuration and HTTP handling

```go
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
func NewClient(config Config) (*Client, error)

// do performs HTTP requests with authentication
func (c *Client) do(ctx context.Context, method, path string, body interface{}, result interface{}) error
```

**Function Responsibilities:**
- Validate required configuration (API key, workspace ID)
- Set default HTTP client if none provided
- Set default base URL to `https://app.publer.com/api/v1/`
- Add authentication headers to all requests
- Handle JSON marshaling/unmarshaling
- Convert HTTP errors to custom error types

#### 2. Error Handling
**File**: `v1/errors.go`
**Changes**: Custom error types for API and rate limit errors

```go
// APIError represents an error response from the Publer API
type APIError struct {
    Method     string
    URL        string
    StatusCode int
    Message    string
}

// RateLimitError represents a rate limit exceeded error
type RateLimitError struct {
    APIError
    Limit     int
    Remaining int
    Reset     int64
}

// Error returns the formatted error message
func (e *APIError) Error() string

// Error returns the formatted rate limit error message  
func (e *RateLimitError) Error() string
```

**Function Responsibilities:**
- Format error messages as `METHOD https://host:port/slug with STATUS returned "Message"`
- Parse rate limit headers from 429 responses
- Support `errors.As()` for type assertions
- Extract error messages from JSON response bodies

#### 3. Generic Iterator
**File**: `v1/iterator.go`
**Changes**: Generic iterator interface compatible with Publer pagination

```go
// Page represents a page of results from paginated API
type Page[T any] struct {
    Items      []T `json:"items"`
    Total      int `json:"total"`
    Page       int `json:"page"`
    PerPage    int `json:"per_page"`
    TotalPages int `json:"total_pages"`
}

// Iterator provides iteration over paginated API resources
type Iterator[T any] interface {
    Next(ctx context.Context, page *Page[T]) bool
    Err() error
}

// PageFetcher defines how to fetch pages of data
type PageFetcher[T any] interface {
    FetchPage(ctx context.Context, pageNum int) (*Page[T], error)
}

// GenericIterator implements Iterator for any paginated resource
type GenericIterator[T any] struct {
    fetcher     PageFetcher[T]
    currentPage int
    totalPages  int
    err         error
    initialized bool
}

// NewGenericIterator creates a new iterator for paginated resources
func NewGenericIterator[T any](fetcher PageFetcher[T]) *GenericIterator[T]

// Next fetches the next page of results
// Returns false when no more pages or context cancelled
// Check Err() for context cancellation or other errors
func (it *GenericIterator[T]) Next(ctx context.Context, page *Page[T]) bool

// Err returns any error encountered during iteration
func (it *GenericIterator[T]) Err() error
```

**Function Responsibilities:**
- Provide reusable iteration logic for any paginated resource
- Lazy initialization on first Next() call
- Track pagination state across calls
- Store and return errors appropriately

#### 4. Mock Server Foundation
**File**: `v1/mock_server.go`
**Changes**: Basic HTTP mock server for testing

```go
// MockServer provides a test HTTP server that mimics Publer API
type MockServer struct {
    server      *httptest.Server
    scenario    string
    rateLimit   RateLimitConfig
    jobDelay    time.Duration
    jobs        map[string]*JobStatus
    posts       []Post
    accounts    []Account
    workspaces  []Workspace
}

// RateLimitConfig holds rate limit simulation settings
type RateLimitConfig struct {
    Limit     int
    Remaining int
    Reset     time.Time
}

// NewMockServer creates a new mock server instance
func NewMockServer() *MockServer

// Start starts the mock HTTP server
func (m *MockServer) Start() (string, error)

// Stop stops the mock HTTP server
func (m *MockServer) Stop() error

// Reset clears all mock server state for next test
func (m *MockServer) Reset()

// SetResponse configures expected response for specific endpoint
func (m *MockServer) SetResponse(method, path string, statusCode int, body interface{})

// SetRateLimit configures rate limit headers for next response
func (m *MockServer) SetRateLimit(limit, remaining int, reset time.Time)

// SetJobStatus configures job status response for job ID
func (m *MockServer) SetJobStatus(jobID, status string, progress int, result *JobResult, err string)

// SetDelay adds artificial delay to responses
func (m *MockServer) SetDelay(delay time.Duration)

// AddPosts adds posts to mock data for listing endpoints
func (m *MockServer) AddPosts(posts []Post)

// AddAccounts adds accounts to mock data for listing endpoints
func (m *MockServer) AddAccounts(accounts []Account)

// AddWorkspaces adds workspaces to mock data for listing endpoints
func (m *MockServer) AddWorkspaces(workspaces []Workspace)

// SetRateLimit configures rate limit simulation
func (m *MockServer) SetRateLimit(limit, remaining int, reset time.Time)

// handleRequest routes requests to appropriate handlers
func (m *MockServer) handleRequest(w http.ResponseWriter, r *http.Request)
```

**Function Responsibilities:**
- Start HTTP test server with configurable URL
- Route requests based on path and method
- Return configured responses per endpoint
- Validate authentication headers (unless explicitly disabled)
- Support pagination metadata in responses
- Always return full pages to caller (API doesn't support configurable page sizes)

**Mock Server Usage Pattern:**
```go
// Each test starts with reset
server.Reset()

// Configure expected responses
server.SetResponse("GET", "/api/v1/posts", 200, ListPostsResponse{
    Posts: []Post{{ID: "1", Text: "test"}},
    Total: 1, Page: 1, PerPage: 10, TotalPages: 1,
})

// Configure rate limiting if needed
server.SetRateLimit(100, 50, time.Now().Add(time.Minute))

// Configure job status progression
server.SetJobStatus("job-123", "working", 50, nil, "")
// Later in test:
server.SetJobStatus("job-123", "completed", 100, &JobResult{PostIDs: []string{"post-456"}}, "")

// Add artificial delay for timeout testing
server.SetDelay(3 * time.Second)
```

**Testing Requirements:**
```go
func TestNewClient(t *testing.T)         // NOTE: Do NOT use t.Parallel()
func TestClientAuthentication(t *testing.T) // NOTE: Do NOT use t.Parallel()
func TestAPIError(t *testing.T)          // NOTE: Do NOT use t.Parallel()
func TestRateLimitError(t *testing.T)    // NOTE: Do NOT use t.Parallel()
func TestGenericIterator(t *testing.T)   // NOTE: Do NOT use t.Parallel()
func TestMockServer(t *testing.T)        // NOTE: Do NOT use t.Parallel()
```

**Test Objectives:**
- Verify client creation and configuration
- Test authentication header injection
- Validate error formatting and rate limit parsing
- Test iterator initialization and pagination
- Verify mock server routing and responses

**Context for implementation:**
- Mock server should validate auth headers on all requests by default
- Iterator should work with any paginated response structure using Page[T]
- Each test should call server.Reset() before configuring responses
- Tests must run sequentially to avoid mock server state conflicts
- Page size is fixed by API, not configurable by client
- Context cancellation in Next() should return false and set Err() to context.Cancelled

### Validation Commands
- `go build ./v1`
- `go test ./v1 -run TestClient`
- `go test ./v1 -run TestMockServer`

## Phase 1: List Posts and Publish Post

### Overview
Implements post listing with pagination and immediate post publishing with job status checking.

### Changes Required:

#### 1. Post Types and Structures
**File**: `v1/posts.go`
**Changes**: Define post-related types

```go
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
    PostLink    string    `json:"post_link"` // Published post URL on the social platform
    HasMedia    bool      `json:"has_media"`
    Network     string    `json:"network"`
}

// User represents a Publer user (simplified for posts)
type User struct {
    ID        string `json:"id"`
    Email     string `json:"email"`
    Name      string `json:"name"`
    FirstName string `json:"first_name"`
    Picture   string `json:"picture"`
}

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

// PublishPostRequest represents immediate post publishing
type PublishPostRequest struct {
    Text     string   `json:"text"`
    Accounts []string `json:"accounts"`
    Media    []Media  `json:"media,omitempty"`
}

// PublishPostResponse contains job ID for async processing
type PublishPostResponse struct {
    JobID string `json:"job_id"`
}

// Media represents media attachment
type Media struct {
    URL  string `json:"url"`
    Type string `json:"type"`
}
```

#### 2. Post Operations
**File**: `v1/posts_operations.go`
**Changes**: Implement ListPosts and PublishPost methods

```go
// ListPosts retrieves posts with filtering options
func (c *Client) ListPosts(ctx context.Context, req ListPostsRequest) Iterator[Post]

// PublishPost publishes content immediately
func (c *Client) PublishPost(ctx context.Context, req PublishPostRequest, resp *PublishPostResponse) error
```

**Function Responsibilities:**
- ListPosts: Return iterator that fetches from GET `/api/v1/posts` with query parameters
- PublishPost: POST to `/api/v1/posts/schedule/publish`
- Handle JSON marshaling for requests and responses
- Apply standard error handling

#### 3. Job Operations
**File**: `v1/jobs.go`
**Changes**: Job status checking and waiting

```go
// JobStatus represents async job status
type JobStatus struct {
    ID        string    `json:"id"`
    Status    string    `json:"status"`
    Progress  int       `json:"progress"`
    Result    JobResult `json:"result,omitempty"`
    Error     string    `json:"error,omitempty"` // Job system errors (not business validation)
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// JobResult contains job completion data
type JobResult struct {
    PostIDs []string `json:"post_ids"`
    Message string   `json:"message"`
    Error   string   `json:"error,omitempty"`
}

// GetJobStatusRequest requests job status
type GetJobStatusRequest struct {
    JobID string
}

// GetJobStatusResponse contains job status
type GetJobStatusResponse struct {
    JobStatus
}

// GetJobStatus checks status of async job
func (c *Client) GetJobStatus(ctx context.Context, req GetJobStatusRequest, resp *GetJobStatusResponse) error

// WaitForJob polls job status until completion
func (c *Client) WaitForJob(ctx context.Context, jobID string, result *JobResult) error
```

**Function Responsibilities:**
- GetJobStatus: GET to `/api/v1/job_status/{job_id}`
- WaitForJob: Poll status with exponential backoff, populate result parameter
- Handle job completion, failure, and timeout
- Return appropriate errors for job failures, include error info in JobResult

#### 4. Mock Server Updates
**File**: `v1/mock_server.go`
**Changes**: Add post and job endpoints

```go
// handleListPosts handles GET /api/v1/posts
func (m *MockServer) handleListPosts(w http.ResponseWriter, r *http.Request)

// handlePublishPost handles POST /api/v1/posts/schedule/publish
func (m *MockServer) handlePublishPost(w http.ResponseWriter, r *http.Request)

// handleJobStatus handles GET /api/v1/job_status/{job_id}
func (m *MockServer) handleJobStatus(w http.ResponseWriter, r *http.Request)

// SetJobDelay configures job completion delay
func (m *MockServer) SetJobDelay(delay time.Duration)
```

**Function Responsibilities:**
- Simulate paginated post responses
- Create job IDs for post publishing
- Simulate job progression with configurable delays
- Return appropriate job status based on scenario

**Testing Requirements:**
```go
func TestListPosts(t *testing.T)
func TestListPostsPagination(t *testing.T)
func TestPublishPost(t *testing.T)
func TestGetJobStatus(t *testing.T)
func TestWaitForJob(t *testing.T)
```

**Test Objectives:**
- Test post listing iterator with various filters
- Verify pagination handling through iterator
- Test immediate post publishing
- Validate job status checking and waiting
- Test error scenarios and rate limiting
- NOTE: All tests must run sequentially, do NOT use t.Parallel()

**Context for implementation:**
- Mock server should simulate realistic pagination (configure multi-page responses)
- Job delay should be configurable for testing (use server.SetDelay())
- WaitForJob should respect context cancellation
- Each test should call server.Reset() before configuring responses

### Validation Commands
- `go test ./v1 -run TestListPosts`
- `go test ./v1 -run TestPublishPost`
- `go test ./v1 -run TestJob`

## Phase 2: Schedule Post and Draft Post

### Overview
Implements scheduled post creation and draft post management.

### Changes Required:

#### 1. Schedule Post Types
**File**: `v1/posts_schedule.go`
**Changes**: Define scheduling-related types

```go
// SchedulePostRequest represents scheduled post creation
type SchedulePostRequest struct {
    Text        string    `json:"text"`
    Accounts    []string  `json:"accounts"`
    ScheduledAt time.Time `json:"scheduled_at"`
    Media       []Media   `json:"media,omitempty"`
    TimeZone    string    `json:"timezone,omitempty"`
}

// SchedulePostResponse contains job ID for async processing
type SchedulePostResponse struct {
    JobID string `json:"job_id"`
}

// CreateDraftPostRequest represents draft post creation
type CreateDraftPostRequest struct {
    Text       string   `json:"text"`
    Accounts   []string `json:"accounts"`
    Media      []Media  `json:"media,omitempty"`
    Visibility string   `json:"visibility"` // draft_private or draft_public
}

// CreateDraftPostResponse contains job ID for async processing
type CreateDraftPostResponse struct {
    JobID string `json:"job_id"`
}
```

#### 2. Schedule Operations
**File**: `v1/posts_schedule_operations.go`
**Changes**: Implement scheduling methods

```go
// SchedulePost schedules a post for future publication
func (c *Client) SchedulePost(ctx context.Context, req SchedulePostRequest, resp *SchedulePostResponse) error

// CreateDraftPost creates a draft post
func (c *Client) CreateDraftPost(ctx context.Context, req CreateDraftPostRequest, resp *CreateDraftPostResponse) error
```

**Function Responsibilities:**
- SchedulePost: POST to `/api/v1/posts/schedule` with scheduled_at
- CreateDraftPost: POST to `/api/v1/posts/schedule` with draft state
- Validate scheduling parameters
- Handle timezone conversions if needed

#### 3. Mock Server Updates
**File**: `v1/mock_server.go`
**Changes**: Add scheduling endpoints

```go
// handleSchedulePost handles POST /api/v1/posts/schedule
func (m *MockServer) handleSchedulePost(w http.ResponseWriter, r *http.Request)

// AddScheduledPost adds a scheduled post to mock data
func (m *MockServer) AddScheduledPost(post Post)
```

**Function Responsibilities:**
- Parse and validate scheduling requests
- Store scheduled posts with proper timestamps
- Return job IDs for tracking
- Simulate draft visibility settings

**Testing Requirements:**
```go
func TestSchedulePost(t *testing.T)
func TestSchedulePostValidation(t *testing.T)
func TestCreateDraftPost(t *testing.T)
func TestDraftVisibility(t *testing.T)
```

**Test Objectives:**
- Test post scheduling with various times
- Verify draft creation with visibility options
- Test timezone handling
- Validate error scenarios

**Context for implementation:**
- Mock server should validate scheduled times are in future
- Draft posts should have appropriate state values
- Support both private and public draft visibility
- Each test should call server.Reset() before configuring responses

### Validation Commands
- `go test ./v1 -run TestSchedulePost`
- `go test ./v1 -run TestDraft`

## Phase 3: User and Workspace Operations

### Overview
Implements user profile retrieval and workspace listing operations.

### Changes Required:

#### 1. User and Workspace Types
**File**: `v1/users.go`
**Changes**: Define user-related types

```go
// GetMeRequest represents request for current user
type GetMeRequest struct{}

// GetMeResponse represents current user response
type GetMeResponse struct {
    User
}

// GetMe retrieves information about the currently authenticated user
func (c *Client) GetMe(ctx context.Context, req GetMeRequest, resp *GetMeResponse) error
```

**File**: `v1/workspaces.go`
**Changes**: Define workspace-related types

```go
// Workspace represents a Publer workspace
type Workspace struct {
    ID      string `json:"id"`
    Owner   User   `json:"owner"`
    Name    string `json:"name"`
    Members []User `json:"members"`
    Plan    string `json:"plan"`
    Picture string `json:"picture"`
}

// ListWorkspacesRequest represents request for listing workspaces
type ListWorkspacesRequest struct{}

// ListWorkspacesResponse represents workspace list response
type ListWorkspacesResponse struct {
    Workspaces []Workspace
}

// ListWorkspaces retrieves all workspaces for the authenticated user
func (c *Client) ListWorkspaces(ctx context.Context, req ListWorkspacesRequest) Iterator[Workspace]
```

**Function Responsibilities:**
- GetMe: GET to `/api/v1/users/me`
- ListWorkspaces: Return iterator that fetches from GET `/api/v1/workspaces`
- Handle nested user objects in workspace responses
- Apply standard error handling

#### 2. Mock Server Updates
**File**: `v1/mock_server.go`
**Changes**: Add user and workspace endpoints

```go
// handleGetMe handles GET /api/v1/users/me
func (m *MockServer) handleGetMe(w http.ResponseWriter, r *http.Request)

// handleListWorkspaces handles GET /api/v1/workspaces
func (m *MockServer) handleListWorkspaces(w http.ResponseWriter, r *http.Request)

// SetCurrentUser sets the mock current user
func (m *MockServer) SetCurrentUser(user User)

// AddWorkspace adds a workspace to mock data
func (m *MockServer) AddWorkspace(workspace Workspace)
```

**Function Responsibilities:**
- Return configured user profile
- Return list of workspaces with members
- Validate workspace ID header presence
- Support multiple workspace scenarios

**Testing Requirements:**
```go
func TestGetMe(t *testing.T)
func TestListWorkspaces(t *testing.T)
func TestWorkspaceMembers(t *testing.T)
```

**Test Objectives:**
- Test user profile retrieval
- Verify workspace listing with members
- Test authentication failures
- Validate workspace switching scenarios

**Context for implementation:**
- Mock server should return consistent user data
- Workspaces should include owner and member details
- Support testing of multi-workspace scenarios
- Each test should call server.Reset() before configuring responses

### Validation Commands
- `go test ./v1 -run TestGetMe`
- `go test ./v1 -run TestWorkspace`

## Phase 4: Account Operations

### Overview
Implements social media account listing and management.

### Changes Required:

#### 1. Account Types
**File**: `v1/accounts.go`
**Changes**: Define account-related types

```go
// Account represents a social media account
type Account struct {
    ID       string `json:"id"`
    Provider string `json:"provider"`
    Name     string `json:"name"`
    SocialID string `json:"social_id"`
    Picture  string `json:"picture"`
    Type     string `json:"type"`
}

// ListAccountsRequest represents request for listing accounts
type ListAccountsRequest struct{}

// ListAccountsResponse represents account list response
type ListAccountsResponse struct {
    Accounts []Account
}

// ListAccounts retrieves all social media accounts in the workspace
func (c *Client) ListAccounts(ctx context.Context, req ListAccountsRequest) Iterator[Account]
```

**Function Responsibilities:**
- ListAccounts: Return iterator that fetches from GET `/api/v1/accounts`
- Handle various provider types (facebook, instagram, twitter, etc.)
- Parse account type variations (page, profile, group)
- Apply standard error handling

#### 2. Mock Server Updates
**File**: `v1/mock_server.go`
**Changes**: Add account endpoints

```go
// handleListAccounts handles GET /api/v1/accounts
func (m *MockServer) handleListAccounts(w http.ResponseWriter, r *http.Request)

// AddAccount adds a social media account to mock data
func (m *MockServer) AddAccount(account Account)

// SetAccountsByProvider sets accounts filtered by provider
func (m *MockServer) SetAccountsByProvider(provider string, accounts []Account)
```

**Function Responsibilities:**
- Return configured social media accounts
- Support filtering by provider type
- Validate workspace context
- Simulate various account types

**Testing Requirements:**
```go
func TestListAccounts(t *testing.T)
func TestAccountProviders(t *testing.T)
func TestAccountTypes(t *testing.T)
```

**Test Objectives:**
- Test account listing for various providers
- Verify account type handling
- Test empty account scenarios
- Validate workspace-specific accounts

**Context for implementation:**
- Mock server should support all provider types
- Account types vary by provider (page, profile, etc.)
- Accounts are workspace-specific
- Each test should call server.Reset() before configuring responses

### Validation Commands
- `go test ./v1 -run TestAccount`

## Phase 5: Bulk Operations

### Overview
Implements bulk post publishing and scheduling operations.

### Changes Required:

#### 1. Bulk Operation Types
**File**: `v1/posts_bulk.go`
**Changes**: Define bulk operation types

```go
// BulkPost represents a single post in bulk operation
type BulkPost struct {
    Text        string    `json:"text"`
    Accounts    []string  `json:"accounts"`
    ScheduledAt time.Time `json:"scheduled_at,omitempty"`
    Media       []Media   `json:"media,omitempty"`
}

// BulkPublishPostsRequest represents bulk immediate publishing
type BulkPublishPostsRequest struct {
    Posts []BulkPost `json:"posts"`
}

// BulkPublishPostsResponse contains job ID for async processing
type BulkPublishPostsResponse struct {
    JobID string `json:"job_id"`
}

// BulkSchedulePostsRequest represents bulk scheduled publishing
type BulkSchedulePostsRequest struct {
    Posts []BulkPost `json:"posts"`
}

// BulkSchedulePostsResponse contains job ID for async processing
type BulkSchedulePostsResponse struct {
    JobID string `json:"job_id"`
}
```

#### 2. Bulk Operations Implementation
**File**: `v1/posts_bulk_operations.go`
**Changes**: Implement bulk methods

```go
// BulkPublishPosts publishes multiple posts immediately
func (c *Client) BulkPublishPosts(ctx context.Context, req BulkPublishPostsRequest, resp *BulkPublishPostsResponse) error

// BulkSchedulePosts schedules multiple posts
func (c *Client) BulkSchedulePosts(ctx context.Context, req BulkSchedulePostsRequest, resp *BulkSchedulePostsResponse) error
```

**Function Responsibilities:**
- BulkPublishPosts: POST to `/api/v1/posts/schedule/publish` with multiple posts
- BulkSchedulePosts: POST to `/api/v1/posts/schedule` with multiple posts
- Validate bulk operation limits
- Handle partial failures in job results

#### 3. Mock Server Updates
**File**: `v1/mock_server.go`
**Changes**: Handle bulk operations

```go
// handleBulkPublish handles bulk publishing requests
func (m *MockServer) handleBulkPublish(w http.ResponseWriter, r *http.Request)

// handleBulkSchedule handles bulk scheduling requests
func (m *MockServer) handleBulkSchedule(w http.ResponseWriter, r *http.Request)

// SetBulkOperationLimit sets maximum posts per bulk operation
func (m *MockServer) SetBulkOperationLimit(limit int)
```

**Function Responsibilities:**
- Process multiple posts in single request
- Create job for bulk operation tracking
- Simulate partial success scenarios
- Enforce bulk operation limits

**Testing Requirements:**
```go
func TestBulkPublishPosts(t *testing.T)
func TestBulkSchedulePosts(t *testing.T)
func TestBulkOperationLimits(t *testing.T)
func TestBulkPartialFailure(t *testing.T)
```

**Test Objectives:**
- Test bulk publishing of multiple posts
- Verify bulk scheduling with different times
- Test bulk operation limits
- Validate partial failure handling

**Context for implementation:**
- Bulk operations return single job ID (configure job progression with SetJobStatus)
- Job results should indicate success/failure per post
- Mock server should simulate realistic processing times with SetDelay
- Each test should call server.Reset() before configuring responses

### Validation Commands
- `go test ./v1 -run TestBulk`

## Phase 6: Post Management (Update/Delete/Get)

### Overview
Implements post update, deletion, and individual post retrieval.

### Changes Required:

#### 1. Post Management Types
**File**: `v1/posts_management.go`
**Changes**: Define management operation types

```go
// GetPostRequest represents request for single post
type GetPostRequest struct {
    PostID string
}

// GetPostResponse represents single post response
type GetPostResponse struct {
    Post
}

// UpdatePostRequest represents post update request
type UpdatePostRequest struct {
    PostID      string    `json:"-"`
    Text        string    `json:"text,omitempty"`
    ScheduledAt time.Time `json:"scheduled_at,omitempty"`
    Media       []Media   `json:"media,omitempty"`
}

// UpdatePostResponse represents post update response
type UpdatePostResponse struct {
    Post
}

// DeletePostRequest represents post deletion request
type DeletePostRequest struct {
    PostID string
}

// DeletePostResponse represents post deletion response
type DeletePostResponse struct {
    Success bool   `json:"success"`
    Message string `json:"message"`
}
```

#### 2. Management Operations Implementation
**File**: `v1/posts_management_operations.go`
**Changes**: Implement management methods

```go
// GetPost retrieves a single post by ID
func (c *Client) GetPost(ctx context.Context, req GetPostRequest, resp *GetPostResponse) error

// UpdatePost updates an existing post
func (c *Client) UpdatePost(ctx context.Context, req UpdatePostRequest, resp *UpdatePostResponse) error

// DeletePost deletes a post
func (c *Client) DeletePost(ctx context.Context, req DeletePostRequest, resp *DeletePostResponse) error
```

**Function Responsibilities:**
- GetPost: GET to `/api/v1/posts/{id}`
- UpdatePost: PATCH to `/api/v1/posts/{id}`
- DeletePost: DELETE to `/api/v1/posts/{id}`
- Validate post exists before update/delete
- Handle post state transitions

#### 3. Mock Server Updates
**File**: `v1/mock_server.go`
**Changes**: Add management endpoints

```go
// handleGetPost handles GET /api/v1/posts/{id}
func (m *MockServer) handleGetPost(w http.ResponseWriter, r *http.Request)

// handleUpdatePost handles PATCH /api/v1/posts/{id}
func (m *MockServer) handleUpdatePost(w http.ResponseWriter, r *http.Request)

// handleDeletePost handles DELETE /api/v1/posts/{id}
func (m *MockServer) handleDeletePost(w http.ResponseWriter, r *http.Request)

// UpdateMockPost updates a post in mock data
func (m *MockServer) UpdateMockPost(id string, updates map[string]interface{})
```

**Function Responsibilities:**
- Find and return individual posts
- Apply partial updates to existing posts
- Remove posts from mock data
- Return 404 for non-existent posts

**Testing Requirements:**
```go
func TestGetPost(t *testing.T)
func TestUpdatePost(t *testing.T)
func TestDeletePost(t *testing.T)
func TestPostNotFound(t *testing.T)
```

**Test Objectives:**
- Test individual post retrieval
- Verify partial post updates
- Test post deletion
- Validate 404 handling for missing posts

**Context for implementation:**
- Update should support partial updates
- Delete should handle scheduled vs published posts differently
- Mock server should maintain post state consistency
- Each test should call server.Reset() before configuring responses

### Validation Commands
- `go test ./v1 -run TestGetPost`
- `go test ./v1 -run TestUpdatePost`
- `go test ./v1 -run TestDeletePost`

## Phase 7: Iterator Implementation for Posts

### Overview
Adapts the generic iterator for post-specific pagination.

### Changes Required:

#### 1. Post Iterator Implementation
**File**: `v1/posts_iterator.go`
**Changes**: Create post-specific iterator

```go
// PostPageFetcher implements PageFetcher for posts
type PostPageFetcher struct {
    client  *Client
    request ListPostsRequest
}

// FetchPage fetches a page of posts
func (f *PostPageFetcher) FetchPage(ctx context.Context, pageNum int) (*Page[Post], error)
```

**Function Responsibilities:**
- Implement PageFetcher interface for posts
- Handle page parameter in ListPostsRequest
- Map ListPostsResponse to Page[Post] structure
- Return configured iterator instance

**Testing Requirements:**
```go
func TestListPostsIterator(t *testing.T)
func TestPostIteratorPagination(t *testing.T)
func TestPostIteratorError(t *testing.T)
func TestPostIteratorContext(t *testing.T)
```

**Test Objectives:**
- Test iteration through multiple pages
- Verify lazy loading on first Next()
- Test error propagation during iteration
- Validate context cancellation support

**Context for implementation:**
- Iterator should respect all ListPostsRequest filters
- Page fetching should be lazy (on-demand)
- Context cancellation should stop iteration and set Err() to context.Cancelled
- Configure multi-page responses for pagination testing
- Each test should call server.Reset() before configuring responses

### Validation Commands
- `go test ./v1 -run TestPostIterator`

## Phase 8: Convenience Methods

### Overview
Implements high-level convenience methods for common operations.

### Changes Required:

#### 1. Post Convenience Methods
**File**: `v1/posts_convenience.go`
**Changes**: Add convenience methods

```go
// GetPostsByState returns an iterator for posts filtered by state
func (c *Client) GetPostsByState(state string) Iterator[Post]

// GetPostsByDateRange returns an iterator for posts within date range
func (c *Client) GetPostsByDateRange(from, to time.Time) Iterator[Post]
```

**Function Responsibilities:**
- Create pre-configured ListPostsRequest instances
- Return iterators with appropriate filters applied
- Provide simple interfaces for common use cases

**Testing Requirements:**
```go
func TestGetPostsByState(t *testing.T)
func TestGetPostsByDateRange(t *testing.T)
```

**Test Objectives:**
- Verify correct filter application
- Test iterator functionality with convenience methods
- Validate parameter handling

**Context for implementation:**
- Methods should be thin wrappers around ListPostsIterator
- Keep interfaces simple and focused
- Document common use cases

### Validation Commands
- `go test ./v1 -run TestConvenience`

## Phase 9: Advanced Post Features

### Overview
Implements recurring posts, auto-scheduling, and content recycling.

### Changes Required:

#### 1. Advanced Post Types
**File**: `v1/posts_advanced.go`
**Changes**: Define advanced posting types

```go
// RecurringPostRequest represents recurring post configuration
type RecurringPostRequest struct {
    Text        string         `json:"text"`
    Accounts    []string       `json:"accounts"`
    Media       []Media        `json:"media,omitempty"`
    Recurrence  RecurrenceRule `json:"recurrence"`
}

// RecurrenceRule defines how posts repeat
type RecurrenceRule struct {
    Frequency string    `json:"frequency"` // daily, weekly, monthly
    Interval  int       `json:"interval"`
    DaysOfWeek []string `json:"days_of_week,omitempty"`
    EndDate   time.Time `json:"end_date,omitempty"`
    Count     int       `json:"count,omitempty"`
}

// AutoScheduleRequest represents auto-scheduling configuration
type AutoScheduleRequest struct {
    Text      string    `json:"text"`
    Accounts  []string  `json:"accounts"`
    Media     []Media   `json:"media,omitempty"`
    StartDate time.Time `json:"start_date"`
    EndDate   time.Time `json:"end_date"`
    Slots     int       `json:"slots"`
}

// RecyclePostRequest represents content recycling configuration
type RecyclePostRequest struct {
    PostID    string    `json:"post_id"`
    StartDate time.Time `json:"start_date"`
    EndDate   time.Time `json:"end_date"`
    Frequency string    `json:"frequency"`
    MaxCount  int       `json:"max_count"`
}
```

#### 2. Advanced Operations Implementation
**File**: `v1/posts_advanced_operations.go`
**Changes**: Implement advanced methods

```go
// CreateRecurringPost creates a recurring post
func (c *Client) CreateRecurringPost(ctx context.Context, req RecurringPostRequest, resp *SchedulePostResponse) error

// AutoSchedulePost uses AI to determine optimal posting times
func (c *Client) AutoSchedulePost(ctx context.Context, req AutoScheduleRequest, resp *SchedulePostResponse) error

// RecyclePost configures content recycling
func (c *Client) RecyclePost(ctx context.Context, req RecyclePostRequest, resp *SchedulePostResponse) error
```

**Function Responsibilities:**
- Handle complex scheduling rules
- Validate recurrence patterns
- Support various frequency options
- Return job IDs for tracking

**Testing Requirements:**
```go
func TestCreateRecurringPost(t *testing.T)
func TestAutoSchedulePost(t *testing.T)
func TestRecyclePost(t *testing.T)
```

**Test Objectives:**
- Test recurring post patterns
- Verify auto-scheduling slot distribution
- Test recycling configuration
- Validate complex scheduling rules

**Context for implementation:**
- Mock server should simulate schedule generation
- Support daily, weekly, monthly patterns
- Validate date ranges and counts
- Each test should call server.Reset() before configuring responses

### Validation Commands
- `go test ./v1 -run TestRecurring`
- `go test ./v1 -run TestAutoSchedule`
- `go test ./v1 -run TestRecycle`

## Phase 10: Documentation and Examples

### Overview
Completes the implementation with comprehensive documentation and usage examples.

### Changes Required:

#### 1. Package Documentation
**File**: `v1/doc.go`
**Changes**: Package-level documentation

```go
// Package publer provides a Go client library for the Publer.com API v1.
//
// The client supports all major Publer API operations including post management,
// scheduling, bulk operations, and workspace management.
//
// Basic usage:
//
//	config := publer.Config{
//	    APIKey:      "your-api-key",
//	    WorkspaceID: "your-workspace-id",
//	}
//	client, err := publer.NewClient(config)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// List posts
//	iter := client.ListPosts(ctx, publer.ListPostsRequest{})
//	for {
//		var page publer.Page[publer.Post]
//		if !iter.Next(ctx, &page) {
//			break
//		}
//		// Process page.Items...
//		log.Printf("Page %d of %d, %d posts\n", page.Page, page.TotalPages, len(page.Items))
//	}
//	if err := iter.Err(); err != nil {
//		log.Fatal(err)
//	}
package publer
```

#### 2. Example Files
**File**: `v1/examples_test.go`
**Changes**: Runnable examples

```go
func ExampleClient_ListPosts()
func ExampleClient_PublishPost()
func ExampleClient_GetPostsByState()
func ExampleClient_WaitForJob()
func ExampleClient_BulkSchedulePosts()
```

**Function Responsibilities:**
- Demonstrate common usage patterns
- Show error handling approaches
- Illustrate iterator usage
- Document authentication setup

**Testing Requirements:**
```go
func TestExamples(t *testing.T)
```

**Test Objectives:**
- Ensure all examples compile and run
- Verify example output is correct
- Test that examples use mock server

**Context for implementation:**
- Examples should be self-contained
- Include error handling in examples
- Show both simple and complex use cases

### Validation Commands
- `go test ./v1 -run Example`
- `go doc -all ./v1`

## Final Validation

### Complete Test Suite
```bash
# Run all tests with race detection and coverage
go test ./v1 -race -cover

# Run benchmarks
go test ./v1 -bench=.

# Check documentation
go doc -all ./v1

# Verify examples
go test ./v1 -run Example
```

### Integration Testing
- Test against mock server with all scenarios
- Verify rate limit handling
- Test context cancellation
- Validate iterator behavior
- Ensure job polling works correctly

## Success Criteria

The implementation is complete when:
1. All phases are implemented and tested
2. Mock server supports all API operations
3. Iterators work correctly with pagination
4. Rate limiting is properly handled
5. Job status polling functions correctly
6. All convenience methods are implemented
7. Documentation and examples are comprehensive
8. Test coverage exceeds 80%
9. All tests pass with race detection enabled
10. The client can be imported as `import "github.com/user/publer.go/v1"` and used as `publer.NewClient()`