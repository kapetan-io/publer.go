# Phase 6: Bulk Operations

## Implementation Instructions for Claude Code

**IMPORTANT CONTEXT**: This phase implements bulk post publishing and scheduling operations. You are building on previous phases' foundation.

### What You Should NOT Do
- Do NOT create CLI tools or command-line interface
- Do NOT use `pkg/` directory structure
- Do NOT implement request timeout or retry logic (handled by caller via context)
- Do NOT add rate limit headers in successful responses
- Do NOT create unit tests with HTTP client mocking (all tests use mock server)
- Do NOT use `t.Parallel()` in ANY tests - all tests must run sequentially

### Prerequisites
You must have completed:
- **Phase 0**: Core client, errors, iterator, mock server foundation
- **Phase 1**: Post operations and job handling
- **Phase 2**: Scheduling and draft operations
- **Phase 3**: User and workspace operations
- **Phase 4**: Account operations

### API Context
- **Bulk publish**: POST `/api/v1/posts/schedule/publish` with multiple posts
- **Bulk schedule**: POST `/api/v1/posts/schedule` with multiple posts
- **Authentication**: Uses Bearer API key + Workspace ID headers from Phase 0
- **Job tracking**: Both operations return single job ID for tracking all posts
- **Media support**: Each post can have media attachments (from Phase 1)

### Changes Required

#### 1. Bulk Operation Types
**File**: `v1/posts_bulk.go`
**Purpose**: Define bulk operation types

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
**Purpose**: Implement bulk methods

```go
// BulkPublishPosts publishes multiple posts immediately
func (c *Client) BulkPublishPosts(ctx context.Context, req BulkPublishPostsRequest, resp *BulkPublishPostsResponse) error

// BulkSchedulePosts schedules multiple posts
func (c *Client) BulkSchedulePosts(ctx context.Context, req BulkSchedulePostsRequest, resp *BulkSchedulePostsResponse) error
```

**Function Responsibilities:**
- BulkPublishPosts: POST to `/api/v1/posts/schedule/publish` with multiple posts
- BulkSchedulePosts: POST to `/api/v1/posts/schedule` with multiple posts
- Validate bulk operation limits (if any)
- Handle partial failures in job results
- Use client.do() method from Phase 0 for HTTP handling
- Reuse job tracking system from Phase 1

#### 3. Mock Server Updates
**File**: `v1/mock_server.go` (extend existing)
**Purpose**: Handle bulk operations

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
- Create job for bulk operation tracking (single job ID for all posts)
- Simulate partial success scenarios
- Enforce bulk operation limits (if configured)
- Differentiate between bulk publish and bulk schedule requests

### Testing Requirements
**CRITICAL**: Do NOT use `t.Parallel()` in any tests

```go
func TestBulkPublishPosts(t *testing.T)
func TestBulkSchedulePosts(t *testing.T)
func TestBulkOperationLimits(t *testing.T)
func TestBulkPartialFailure(t *testing.T)
```

**Test Objectives:**
- Test bulk publishing of multiple posts
- Verify bulk scheduling with different times
- Test bulk operation limits (if enforced)
- Validate partial failure handling in job results
- Test job tracking for bulk operations

**Test Requirements from CLAUDE.md:**
- Test MUST always be in the test package `package v1_test` and not `package v1`
- Test names should be in camelCase and start with a capital letter
- Use `github.com/stretchr/testify/require` and `github.com/stretchr/testify/assert`
- Use `require.ErrorContains(t, err, test.wantErr)` instead of `require.Contains(t, err.Error(), test.wantErr)`
- Use `require` for critical assertions, `assert` for non-critical ones
- No explanatory messages in assertions

### Implementation Context
- Bulk operations return single job ID for tracking all posts
- Job results should indicate success/failure per individual post
- Mock server should simulate realistic processing times with SetDelay (from Phase 0)
- Each test should call server.Reset() before configuring responses
- Tests must run sequentially to avoid mock server state conflicts
- Reuse existing job progression system (SetJobStatus) from Phase 1
- Media type is already defined in Phase 1 (`v1/posts.go`)

### Validation Commands
After implementation, run these commands to verify:
- `go test ./v1 -run TestBulk`

### Success Criteria
This phase is complete when:
1. Multiple posts can be published immediately in bulk
2. Multiple posts can be scheduled in bulk with different times
3. Single job ID is returned for tracking entire bulk operation
4. Job results properly indicate success/failure per post
5. Bulk operation limits are enforced (if applicable)
6. Partial failure scenarios are handled correctly
7. All tests pass without race conditions
8. Mock server simulates realistic bulk processing