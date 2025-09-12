# Phase 3: Schedule Post and Draft Post

## Implementation Instructions for Claude Code

**IMPORTANT CONTEXT**: This phase implements scheduled post creation and draft post management. You are building on Phase 0 (foundation) and Phase 1 (list posts and publish post).

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
- **Phase 1**: Post listing, immediate publishing, job operations

### API Context
- **Schedule posts**: POST `/api/v1/posts/schedule` with scheduled_at timestamp
- **Draft posts**: POST `/api/v1/posts/schedule` with draft state (draft_private or draft_public)
- **Authentication**: Uses Bearer API key + Workspace ID headers from Phase 0
- **Async processing**: Both operations return job IDs for tracking (like Phase 1)

### Changes Required

#### 1. Schedule Post Types
**File**: `v1/posts_schedule.go`
**Purpose**: Define scheduling-related types

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
**Purpose**: Implement scheduling methods

```go
// SchedulePost schedules a post for future publication
func (c *Client) SchedulePost(ctx context.Context, req SchedulePostRequest, resp *SchedulePostResponse) error

// CreateDraftPost creates a draft post
func (c *Client) CreateDraftPost(ctx context.Context, req CreateDraftPostRequest, resp *CreateDraftPostResponse) error
```

**Function Responsibilities:**
- SchedulePost: POST to `/api/v1/posts/schedule` with scheduled_at
- CreateDraftPost: POST to `/api/v1/posts/schedule` with draft state
- Validate scheduling parameters (future dates)
- Handle timezone conversions if needed
- Use client.do() method from Phase 0 for HTTP handling

#### 3. Mock Server Updates
**File**: `v1/mock_server.go` (extend existing)
**Purpose**: Add scheduling endpoints

```go
// handleSchedulePost handles POST /api/v1/posts/schedule
func (m *MockServer) handleSchedulePost(w http.ResponseWriter, r *http.Request)

// AddScheduledPost adds a scheduled post to mock data
func (m *MockServer) AddScheduledPost(post Post)
```

**Function Responsibilities:**
- Parse and validate scheduling requests
- Store scheduled posts with proper timestamps
- Return job IDs for tracking (using existing job system from Phase 1)
- Simulate draft visibility settings
- Differentiate between scheduled and draft posts based on request payload

### Testing Requirements
**CRITICAL**: Do NOT use `t.Parallel()` in any tests

```go
func TestSchedulePost(t *testing.T)
func TestSchedulePostValidation(t *testing.T)
func TestCreateDraftPost(t *testing.T)
func TestDraftVisibility(t *testing.T)
```

**Test Objectives:**
- Test post scheduling with various future times
- Verify draft creation with visibility options (private/public)
- Test timezone handling if implemented
- Validate error scenarios (past dates, invalid visibility)

**Test Requirements from CLAUDE.md:**
- Test MUST always be in the test package `package v1_test` and not `package v1`
- Test names should be in camelCase and start with a capital letter
- Use `github.com/stretchr/testify/require` and `github.com/stretchr/testify/assert`
- Use `require.ErrorContains(t, err, test.wantErr)` instead of `require.Contains(t, err.Error(), test.wantErr)`
- Use `require` for critical assertions, `assert` for non-critical ones
- No explanatory messages in assertions

### Implementation Context
- Mock server should validate scheduled times are in future
- Draft posts should have appropriate state values
- Support both private and public draft visibility
- Each test should call server.Reset() before configuring responses
- Tests must run sequentially to avoid mock server state conflicts
- Reuse job tracking system from Phase 1 for both scheduling and drafts

### Validation Commands
After implementation, run these commands to verify:
- `go test ./v1 -run TestSchedulePost`
- `go test ./v1 -run TestDraft`

### Success Criteria
This phase is complete when:
1. Posts can be scheduled for future publication
2. Draft posts can be created with visibility settings
3. Timezone handling works correctly (if implemented)
4. Job tracking works for both scheduled and draft posts
5. Mock server validates scheduling parameters properly
6. All tests pass without race conditions
7. Error scenarios are handled appropriately