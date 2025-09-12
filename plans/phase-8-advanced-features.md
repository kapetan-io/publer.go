# Phase 8: Convenience Methods and Advanced Features

## Implementation Instructions for Claude Code

**IMPORTANT CONTEXT**: This phase implements both convenience methods for common operations AND advanced post features (recurring posts, auto-scheduling, content recycling). You are building on all previous phases' foundation.

### What You Should NOT Do
- Do NOT create CLI tools or command-line interface
- Do NOT use `pkg/` directory structure
- Do NOT implement request timeout or retry logic (handled by caller via context)
- Do NOT add rate limit headers in successful responses
- Do NOT create unit tests with HTTP client mocking (all tests use mock server)
- Do NOT use `t.Parallel()` in ANY tests - all tests must run sequentially

### Prerequisites
You must have completed:
- **Phase 0**: Core client, errors, iterator, mock server foundation (includes core types)
- **Phase 1**: Post operations and job handling
- **Phase 2**: Post iterator implementation
- **Phase 3**: Schedule and draft operations
- **Phase 4**: User and workspace operations
- **Phase 5**: Account operations
- **Phase 6**: Bulk operations
- **Phase 7**: Post management (update/delete/get)

### Context
This phase combines two sets of features:
1. **Convenience Methods**: Simple wrappers for common filtering use cases
2. **Advanced Features**: Recurring posts, auto-scheduling, and content recycling

## Part A: Convenience Methods

### Changes Required

#### 1. Post Convenience Methods
**File**: `v1/posts_convenience.go`
**Purpose**: Add convenience methods for common operations

```go
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
```

**Function Responsibilities:**
- Create pre-configured ListPostsRequest instances with specific filters
- Return iterators with appropriate filters applied
- Provide simple interfaces for common use cases
- Use context.Background() for simplicity (caller can still cancel via iterator context)

## Part B: Advanced Post Features

### Changes Required

#### 2. Advanced Post Types
**File**: `v1/posts_advanced.go`
**Purpose**: Define advanced posting types

```go
// RecurringPostRequest represents recurring post configuration
type RecurringPostRequest struct {
    Text        string         `json:"text"`
    Accounts    []string       `json:"accounts"`
    Media       []Media        `json:"media,omitempty"`  // Media type from Phase 0
    Recurrence  RecurrenceRule `json:"recurrence"`
}

// RecurrenceRule defines how posts repeat
type RecurrenceRule struct {
    Frequency  string    `json:"frequency"` // daily, weekly, monthly
    Interval   int       `json:"interval"`  // every N days/weeks/months
    DaysOfWeek []string  `json:"days_of_week,omitempty"` // for weekly: ["monday", "friday"]
    EndDate    time.Time `json:"end_date,omitempty"`
    Count      int       `json:"count,omitempty"` // alternative to end_date
}

// AutoScheduleRequest represents auto-scheduling configuration
type AutoScheduleRequest struct {
    Text      string    `json:"text"`
    Accounts  []string  `json:"accounts"`
    Media     []Media   `json:"media,omitempty"`  // Media type from Phase 0
    StartDate time.Time `json:"start_date"`
    EndDate   time.Time `json:"end_date"`
    Slots     int       `json:"slots"` // number of times to post in date range
}

// RecyclePostRequest represents content recycling configuration
type RecyclePostRequest struct {
    PostID    string    `json:"post_id"`
    StartDate time.Time `json:"start_date"`
    EndDate   time.Time `json:"end_date"`
    Frequency string    `json:"frequency"`
    MaxCount  int       `json:"max_count"` // maximum times to recycle
}

// RecurringPostResponse contains job ID for recurring post setup
type RecurringPostResponse struct {
    JobID string `json:"job_id"`
}

// AutoScheduleResponse contains job ID for auto-scheduling
type AutoScheduleResponse struct {
    JobID string `json:"job_id"`
}

// RecyclePostResponse contains job ID for recycling setup
type RecyclePostResponse struct {
    JobID string `json:"job_id"`
}
```

#### 3. Advanced Operations Implementation
**File**: `v1/posts_advanced_operations.go`
**Purpose**: Implement advanced methods

```go
// CreateRecurringPost creates a recurring post schedule
func (c *Client) CreateRecurringPost(ctx context.Context, req RecurringPostRequest, resp *RecurringPostResponse) error

// AutoSchedulePost uses AI to determine optimal posting times
func (c *Client) AutoSchedulePost(ctx context.Context, req AutoScheduleRequest, resp *AutoScheduleResponse) error

// RecyclePost configures content recycling schedule
func (c *Client) RecyclePost(ctx context.Context, req RecyclePostRequest, resp *RecyclePostResponse) error
```

**Function Responsibilities:**
- CreateRecurringPost: POST to `/api/v1/posts/recurring` with recurrence rules
- AutoSchedulePost: POST to `/api/v1/posts/auto-schedule` with time range and slot count
- RecyclePost: POST to `/api/v1/posts/recycle` with existing post ID and recycling rules
- Handle complex scheduling rules and validation
- Support various frequency options (daily, weekly, monthly)
- Return job IDs for tracking like other post operations

#### 4. Mock Server Updates
**File**: `v1/mock_server.go` (extend existing)
**Purpose**: Handle both convenience and advanced features

```go
// handleRecurringPost handles POST /api/v1/posts/recurring
func (m *MockServer) handleRecurringPost(w http.ResponseWriter, r *http.Request)

// handleAutoSchedulePost handles POST /api/v1/posts/auto-schedule
func (m *MockServer) handleAutoSchedulePost(w http.ResponseWriter, r *http.Request)

// handleRecyclePost handles POST /api/v1/posts/recycle
func (m *MockServer) handleRecyclePost(w http.ResponseWriter, r *http.Request)

// SimulateScheduleGeneration creates mock scheduled posts for advanced features
func (m *MockServer) SimulateScheduleGeneration(count int, interval time.Duration)
```

**Function Responsibilities:**
- Process complex scheduling requests
- Validate recurrence patterns and date ranges
- Create jobs for advanced feature tracking
- Simulate schedule generation for testing
- Support different frequency patterns

### Testing Requirements
**CRITICAL**: Do NOT use `t.Parallel()` in any tests

```go
// Convenience method tests
func TestGetPostsByState(t *testing.T)
func TestGetPostsByDateRange(t *testing.T)
func TestGetPostsByAccount(t *testing.T)
func TestGetPostsByQuery(t *testing.T)

// Advanced feature tests
func TestCreateRecurringPost(t *testing.T)
func TestAutoSchedulePost(t *testing.T)
func TestRecyclePost(t *testing.T)
func TestRecurrencePatterns(t *testing.T)
```

**Test Objectives:**
- Convenience Methods:
  - Verify correct filter application for each convenience method
  - Test iterator functionality with convenience methods
  - Validate parameter handling
  - Ensure pagination works
- Advanced Features:
  - Test recurring post patterns (daily, weekly, monthly)
  - Verify auto-scheduling slot distribution
  - Test recycling configuration with existing posts
  - Validate complex scheduling rules
  - Test date range validation

**Test Requirements from CLAUDE.md:**
- Test MUST always be in the test package `package v1_test` and not `package v1`
- Test names should be in camelCase and start with a capital letter
- Use `github.com/stretchr/testify/require` and `github.com/stretchr/testify/assert`
- Use `require.ErrorContains(t, err, test.wantErr)` instead of `require.Contains(t, err.Error(), test.wantErr)`
- Use `require` for critical assertions, `assert` for non-critical ones
- No explanatory messages in assertions

### Implementation Context
- Convenience methods should be thin wrappers around ListPosts - do not duplicate logic
- Mock server should simulate realistic schedule generation
- Support daily, weekly, monthly frequency patterns  
- Validate date ranges and counts appropriately
- Each test should call server.Reset() before configuring responses
- Tests must run sequentially to avoid mock server state conflicts
- Reuse existing job tracking system from Phase 1
- Media type is already defined in Phase 0

### Mock Server Configuration for Testing
```go
// For convenience methods - configure posts with different attributes
server.Reset()
posts := []Post{
    {ID: "1", Text: "Draft post", State: "draft", AccountID: "acc-1"},
    {ID: "2", Text: "Published post", State: "published", AccountID: "acc-2"},
    {ID: "3", Text: "Scheduled post", State: "scheduled", ScheduledAt: time.Now().Add(2 * time.Hour)},
}
server.AddPosts(posts)

// For advanced features - configure existing posts for recycling
server.Reset()
posts := []Post{
    {ID: "existing-1", Text: "Original post", State: "published"},
}
server.AddPosts(posts)

// Configure job tracking for advanced features
server.SetJobStatus("recurring-job-123", "completed", 100, &JobResult{
    PostIDs: []string{"scheduled-1", "scheduled-2", "scheduled-3"},
    Message: "Recurring posts created successfully",
}, "")
```

### Validation Commands
After implementation, run these commands to verify:
- `go test ./v1 -run TestConvenience`
- `go test ./v1 -run TestRecurring`
- `go test ./v1 -run TestAutoSchedule`  
- `go test ./v1 -run TestRecycle`

### Success Criteria
This phase is complete when:
1. All convenience methods correctly filter posts
2. Convenience methods return working iterators with pagination support
3. Recurring posts can be created with various frequency patterns
4. Auto-scheduling distributes posts across date ranges
5. Content recycling works with existing posts
6. Complex scheduling rules are validated properly
7. Job tracking works for all advanced features
8. Mock server simulates realistic advanced processing
9. All tests pass without race conditions
10. Error scenarios are handled appropriately