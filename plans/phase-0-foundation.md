# Phase 0: Foundation - Client, Errors, Mock Server, Iterator

## Implementation Instructions for Claude Code

**IMPORTANT CONTEXT**: This phase establishes the foundation for a complete Go client library for the Publer.com HTTP API v1. You are implementing the core client structure, error handling, mock server for testing, and generic iterator interface.

### What You Should NOT Do
- Do NOT create CLI tools or command-line interface
- Do NOT use `pkg/` directory structure
- Do NOT implement request timeout or retry logic (handled by caller via context)
- Do NOT add rate limit headers in successful responses
- Do NOT create unit tests with HTTP client mocking (all tests use mock server)
- Do NOT use `t.Parallel()` in ANY tests - all tests must run sequentially

### Project Context
- **Current State**: Empty project with only `.git` and `plans/` directories
- **Target**: Foundation for Go client library that exports as `publer.NewClient()` from versioned `/v1/` package
- **Go Version Requirement**: Go 1.18+ (required for generic types)
- **API Details**:
  - Base URL: `https://app.publer.com/api/v1/`
  - Authentication: `Authorization: Bearer-API {key}` and `Publer-Workspace-Id: {id}` headers
  - Rate limits: 100 requests per 2 minutes with headers:
    - `X-RateLimit-Limit`: Total requests allowed
    - `X-RateLimit-Remaining`: Requests remaining
    - `X-RateLimit-Reset`: Unix timestamp when limit resets
  - Business users only (no public API access)
  - RESTful JSON API with standard HTTP status codes

### Changes Required

#### 1. Core Client Structure
**File**: `v1/client.go`
**Purpose**: Create main client with configuration and HTTP handling

**Required imports:**
```go
import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "net/url"
    "bytes"
    "io"
    "strings"
)
```

**Implementation:**
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
func (c *Client) do(ctx context.Context, method, path string, body any, result any) error
```

**Function Responsibilities:**
- Validate required configuration (API key, workspace ID)
- Set default HTTP client if none provided
- Set default base URL to `https://app.publer.com/api/v1/`
- Add authentication headers to all requests
- Handle JSON marshaling/unmarshaling
- Convert HTTP errors to custom error types

#### 2. Core Types
**File**: `v1/types.go`
**Purpose**: Core types used throughout the library

**Required imports:**
```go
import (
    "time"
)
```

**Implementation:**
```go
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
```

#### 3. Error Handling
**File**: `v1/errors.go`
**Purpose**: Custom error types for API and rate limit errors

```go
// ErrorResponse represents the JSON error response from Publer API
type ErrorResponse struct {
    Error   string `json:"error"`
    Message string `json:"message,omitempty"`
    Code    string `json:"code,omitempty"`
}

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
- Parse rate limit headers from 429 responses (`X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`)
- Support `errors.As()` for type assertions
- Extract error messages from JSON response bodies using ErrorResponse struct
- Parse error response JSON: `{"error": "invalid_request", "message": "Details here"}`

#### 4. Generic Iterator
**File**: `v1/iterator.go`
**Purpose**: Generic iterator interface compatible with Publer pagination

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

#### 5. Mock Server Foundation
**File**: `v1/mock_server.go`
**Purpose**: Basic HTTP mock server for testing

**Required imports:**
```go
import (
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "sync"
    "time"
    "fmt"
    "strings"
    "strconv"
)
```

**Implementation:**
```go
// MockServer provides a test HTTP server that mimics Publer API
type MockServer struct {
    mu                *sync.RWMutex
    server            *httptest.Server
    jobDelay          time.Duration
    jobs              map[string]*JobStatus
    jobProgression    map[string][]JobStatus
    jobProgressIndex  map[string]int
    posts             []Post
    accounts          []Account
    workspaces        []Workspace
    responses         map[string]MockResponse
    errorResponses    map[string]MockErrorResponse
    callCounts        map[string]int
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

// NewMockServer creates a new mock server instance
func NewMockServer() *MockServer

// Start starts the mock HTTP server
func (m *MockServer) Start() (string, error)

// Stop stops the mock HTTP server
func (m *MockServer) Stop() error

// Reset clears all mock server state for next test
func (m *MockServer) Reset()

// SetResponse configures expected response for specific endpoint
func (m *MockServer) SetResponse(method, path string, statusCode int, body any)

// SetErrorResponse configures error response after N calls to endpoint
func (m *MockServer) SetErrorResponse(method, path string, callThreshold int, statusCode int, body any, headers map[string]string)

// SetJobStatus configures job status response for job ID
func (m *MockServer) SetJobStatus(jobID, status string, progress int, result *JobResult, err string)

// SetJobProgression configures automatic job state progression
func (m *MockServer) SetJobProgression(jobID string, states []JobStatus)

// AdvanceJobState manually advances job to next state in progression
func (m *MockServer) AdvanceJobState(jobID string) bool

// SetDelay adds artificial delay to responses (bypassed in fast test mode)
func (m *MockServer) SetDelay(delay time.Duration)

// AddPosts adds posts to mock data for listing endpoints
func (m *MockServer) AddPosts(posts []Post)

// AddAccounts adds accounts to mock data for listing endpoints
func (m *MockServer) AddAccounts(accounts []Account)

// AddWorkspaces adds workspaces to mock data for listing endpoints
func (m *MockServer) AddWorkspaces(workspaces []Workspace)

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
- **Job State Management**: Track job progression with state machine validation
- **Error Simulation**: Support call-based error responses with custom headers (e.g., return 429 rate limit error on 3rd call)
- **Call Tracking**: Track call counts per endpoint to trigger errors at specified thresholds
- **Thread Safety**: Use mutex for concurrent access to server state

**Mock Server Usage Patterns:**

**Basic Test Setup:**
```go
// Each test starts with reset (clears all state and call counts)
server.Reset()

// Configure expected responses
server.SetResponse("GET", "/api/v1/posts", 200, ListPostsResponse{
    Posts: []Post{{ID: "1", Text: "test"}},
    Total: 1, Page: 1, PerPage: 10, TotalPages: 1,
})

// Configure error after N successful calls (optional)
server.SetErrorResponse("POST", "/api/v1/posts", 5, 429, map[string]any{
    "error": "rate_limit_exceeded",
}, map[string]string{
    "X-RateLimit-Limit":     "100",
    "X-RateLimit-Remaining": "0",
})
```

**Job Polling with State Progression:**
```go
// Configure automatic job progression
server.SetJobProgression("job-123", []JobStatus{
    {ID: "job-123", Status: "pending", Progress: 0},
    {ID: "job-123", Status: "working", Progress: 50},
    {ID: "job-123", Status: "completed", Progress: 100, Result: &JobResult{PostIDs: []string{"post-456"}}},
})

// Test advances job state explicitly
server.AdvanceJobState("job-123") // pending -> working
server.AdvanceJobState("job-123") // working -> completed
```

**Rate Limit Error Simulation:**
```go
// First configure the successful response for calls 1-2
server.SetResponse("GET", "/api/v1/posts", 200, map[string]any{
    "posts": []Post{{ID: "1", Text: "Hello"}},
    "total": 1, "page": 1, "per_page": 10, "total_pages": 1,
})

// Then configure error response starting on call 3
server.SetErrorResponse("GET", "/api/v1/posts", 3, 429, map[string]any{
    "error": "rate_limit_exceeded",
    "message": "Too many requests",
}, map[string]string{
    "X-RateLimit-Limit":     "100",
    "X-RateLimit-Remaining": "0", 
    "X-RateLimit-Reset":     "1640995200", // Unix timestamp
})

// Calls 1-2 return 200 with posts data, call 3+ returns 429 rate limit error
```

**Context-Based Polling Pattern:**
```go
// All polling should use bounded context
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// Polling with state machine validation
for {
    select {
    case <-ctx.Done():
        return ctx.Err() // Prevents infinite loops
    default:
        job, err := client.GetJobStatus(ctx, jobID)
        if err != nil {
            return err
        }
        if job.Status == "completed" || job.Status == "failed" {
            return nil // Terminal states
        }
        server.AdvanceJobState(jobID) // Test controls progression
    }
}
```

### Testing Requirements
**CRITICAL**: Do NOT use `t.Parallel()` in any tests

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

**Test Requirements from CLAUDE.md:**
- Test MUST always be in the test package `package v1_test` and not `package v1`
- Test names should be in camelCase and start with a capital letter
- Use `github.com/stretchr/testify/require` and `github.com/stretchr/testify/assert`
- Use `require.ErrorContains(t, err, test.wantErr)` instead of `require.Contains(t, err.Error(), test.wantErr)`
- Use `require` for critical assertions, `assert` for non-critical ones
- No explanatory messages in assertions

### Implementation Context
- Mock server should validate auth headers on all requests by default
- Iterator should work with any paginated response structure using Page[T]
- Each test should call server.Reset() before configuring responses
- Tests must run sequentially to avoid mock server state conflicts
- Page size is fixed by API, not configurable by client
- Context cancellation in Next() should return false and set Err() to context.Cancelled

### Advanced Testing Strategies

**Job Polling Infinite Loop Prevention:**
- Use `context.WithTimeout()` for all polling operations (5-10 second max)
- Implement state machine validation: `pending → working → completed/failed`
- Test-controlled progression via `AdvanceJobState()` instead of time-based
- Iteration counter with 100 poll maximum as safety net
- Exponential backoff: 10ms → 20ms → 40ms → ... → 1s max

**Error Response Test Strategies:**
- Simulate any HTTP error (429 rate limit, 500 server error, 404 not found) after N successful calls
- Configure custom headers for rate limit errors (X-RateLimit-*) or other metadata
- No actual timing delays - purely call-count based
- Reset call counts with `server.Reset()` between tests

**Mock Server State Management:**
- Thread-safe with mutex protection for concurrent access
- State isolation via `Reset()` method clears all server state and call counts
- Resource management: Always use `defer server.Stop()` in tests
- Job progression tracking with automatic state advancement
- Configurable response mapping per endpoint
- Call count tracking per endpoint for error threshold simulation

### Validation Commands
After implementation, run these commands to verify:
- `go build ./v1`
- `go test ./v1 -run TestClient`
- `go test ./v1 -run TestMockServer`

### Success Criteria
This phase is complete when:
1. Client can be created with proper configuration validation
2. Authentication headers are added to all requests
3. Errors are properly formatted and rate limit errors include headers
4. Generic iterator works with any type using Page[T]
5. Mock server can start, handle requests, and return configured responses
6. All tests pass without race conditions
7. Test excercise SetErrorResponse() and SetResponse()
8. Code follows the established patterns for subsequent phases