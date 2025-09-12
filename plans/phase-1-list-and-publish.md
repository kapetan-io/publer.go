# Phase 1: List Posts and Publish Post

## Implementation Instructions for Claude Code

**IMPORTANT CONTEXT**: This phase implements post listing with pagination and immediate post publishing with job status checking. You are building on Phase 0's foundation (client, errors, iterator, mock server).

### What You Should NOT Do
- Do NOT create CLI tools or command-line interface  
- Do NOT use `pkg/` directory structure
- Do NOT implement request timeout or retry logic (handled by caller via context)
- Do NOT add rate limit headers in successful responses
- Do NOT create unit tests with HTTP client mocking (all tests use mock server)
- Do NOT use `t.Parallel()` in ANY tests - all tests must run sequentially

### Prerequisites 
You must have completed Phase 0, which provides:
- Core client structure in `v1/client.go`
- Error handling in `v1/errors.go` 
- Generic iterator in `v1/iterator.go`
- Mock server foundation in `v1/mock_server.go`

### API Context
- **Post listing**: GET `/api/v1/posts` with query parameters for filtering
- **Post publishing**: POST `/api/v1/posts/schedule/publish` (immediate publication)
- **Job status**: GET `/api/v1/job_status/{job_id}` for async operation tracking
- **Authentication**: Uses Bearer API key + Workspace ID headers from Phase 0
- **Rate limits**: 100 requests per 2 minutes with `X-RateLimit-*` headers

### Changes Required

#### 1. Post Types and Structures
**File**: `v1/posts.go`
**Purpose**: Define post-related types

```go
// Post represents a Publer post (extends basic definition from Phase 0)
type Post struct {
    ID          string    `json:"id"`
    Text        string    `json:"text"`
    URL         string    `json:"url"`
    State       string    `json:"state"`
    Type        string    `json:"type"`
    AccountID   string    `json:"account_id"`
    User        User      `json:"user"`        // User type from Phase 0 v1/types.go
    ScheduledAt time.Time `json:"scheduled_at"`
    PostLink    string    `json:"post_link"`    // Published post URL on the social platform
    HasMedia    bool      `json:"has_media"`
    Network     string    `json:"network"`
}

// Note: User type is defined in Phase 0 (v1/types.go)

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
    Media    []Media  `json:"media,omitempty"`  // Media type from Phase 0 v1/types.go
}

// PublishPostResponse contains job ID for async processing
type PublishPostResponse struct {
    JobID string `json:"job_id"`
}

// Note: Media type is defined in Phase 0 (v1/types.go)
```

#### 2. Post Operations
**File**: `v1/posts_operations.go`
**Purpose**: Implement ListPosts and PublishPost methods

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
- Apply standard error handling from Phase 0

#### 3. Job Operations
**File**: `v1/jobs.go`
**Purpose**: Job status checking and waiting

```go
// Note: JobStatus and JobResult types are defined in Phase 0 (v1/types.go)
// This phase extends them with additional fields if needed

// JobStatus (from Phase 0) is used for async job tracking
// Additional fields for this phase:
// - CreatedAt time.Time `json:"created_at"`
// - UpdatedAt time.Time `json:"updated_at"`

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

**WaitForJob Implementation Details:**
- **Initial delay**: 1 second
- **Backoff strategy**: Exponential with jitter (delay *= 2, max 30 seconds)
- **Maximum attempts**: No limit (relies on context timeout)
- **Jitter**: Add random 0-500ms to prevent thundering herd
- **Final statuses**: "completed", "failed", "cancelled"
- **Continue polling on**: "pending", "working", "processing"
- **Context handling**: Return context.Err() if context is cancelled
- **Implementation pattern**:
```go
delay := time.Second
for {
    select {
    case <-ctx.Done():
        return ctx.Err()
    case <-time.After(delay):
        // Check job status
        // If final status, return
        // Otherwise continue
        if delay < 30*time.Second {
            delay *= 2
        }
        delay += time.Duration(rand.Intn(500)) * time.Millisecond
    }
}
```

#### 4. Mock Server Updates
**File**: `v1/mock_server.go` (extend existing)
**Purpose**: Add post and job endpoints

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

### Testing Requirements
**CRITICAL**: Do NOT use `t.Parallel()` in any tests

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

**Test Requirements from CLAUDE.md:**
- Test MUST always be in the test package `package v1_test` and not `package v1`
- Test names should be in camelCase and start with a capital letter
- Use `github.com/stretchr/testify/require` and `github.com/stretchr/testify/assert`
- Use `require.ErrorContains(t, err, test.wantErr)` instead of `require.Contains(t, err.Error(), test.wantErr)`
- Use `require` for critical assertions, `assert` for non-critical ones
- No explanatory messages in assertions

### Implementation Context
- Mock server should simulate realistic pagination (configure multi-page responses)
- Job delay should be configurable for testing (use server.SetDelay())
- WaitForJob should respect context cancellation
- Each test should call server.Reset() before configuring responses
- Tests must run sequentially to avoid mock server state conflicts

### Validation Commands
After implementation, run these commands to verify:
- `go test ./v1 -run TestListPosts`
- `go test ./v1 -run TestPublishPost`
- `go test ./v1 -run TestJob`

### Success Criteria
This phase is complete when:
1. Posts can be listed using iterator with filtering support
2. Iterator properly handles pagination across multiple pages
3. Posts can be published immediately with job ID returned
4. Job status can be checked and polling works with WaitForJob
5. Mock server simulates realistic post and job scenarios
6. All tests pass without race conditions
7. Error scenarios are properly handled including rate limiting