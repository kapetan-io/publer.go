# Phase 7: Post Management (Update/Delete/Get)

## Implementation Instructions for Claude Code

**IMPORTANT CONTEXT**: This phase implements post update, deletion, and individual post retrieval. You are building on previous phases' foundation.

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
- **Phase 5**: Bulk operations

### API Context
- **Get post**: GET `/api/v1/posts/{id}` returns single post
- **Update post**: PATCH `/api/v1/posts/{id}` updates existing post
- **Delete post**: DELETE `/api/v1/posts/{id}` removes post
- **Authentication**: Uses Bearer API key + Workspace ID headers from Phase 0
- **Post type**: Use existing Post struct from Phase 1 (`v1/posts.go`)

### Changes Required

#### 1. Post Management Types
**File**: `v1/posts_management.go`
**Purpose**: Define management operation types

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
**Purpose**: Implement management methods

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
- UpdatePost: PATCH to `/api/v1/posts/{id}` with partial updates
- DeletePost: DELETE to `/api/v1/posts/{id}`
- Validate post exists before update/delete (return 404 if not found)
- Handle post state transitions appropriately
- Use client.do() method from Phase 0 for HTTP handling

#### 3. Mock Server Updates
**File**: `v1/mock_server.go` (extend existing)
**Purpose**: Add management endpoints

```go
// handleGetPost handles GET /api/v1/posts/{id}
func (m *MockServer) handleGetPost(w http.ResponseWriter, r *http.Request)

// handleUpdatePost handles PATCH /api/v1/posts/{id}
func (m *MockServer) handleUpdatePost(w http.ResponseWriter, r *http.Request)

// handleDeletePost handles DELETE /api/v1/posts/{id}
func (m *MockServer) handleDeletePost(w http.ResponseWriter, r *http.Request)

// UpdateMockPost updates a post in mock data
func (m *MockServer) UpdateMockPost(id string, updates map[string]any)
```

**Function Responsibilities:**
- Find and return individual posts by ID
- Apply partial updates to existing posts in mock data
- Remove posts from mock data when deleted
- Return 404 for non-existent posts
- Maintain post state consistency in mock data

### Testing Requirements
**CRITICAL**: Do NOT use `t.Parallel()` in any tests

```go
func TestGetPost(t *testing.T)
func TestUpdatePost(t *testing.T)
func TestDeletePost(t *testing.T)
func TestPostNotFound(t *testing.T)
```

**Test Objectives:**
- Test individual post retrieval by ID
- Verify partial post updates (text, scheduled_at, media)
- Test post deletion and success response
- Validate 404 handling for missing posts
- Test update and delete of different post states

**Test Requirements from CLAUDE.md:**
- Test MUST always be in the test package `package v1_test` and not `package v1`
- Test names should be in camelCase and start with a capital letter
- Use `github.com/stretchr/testify/require` and `github.com/stretchr/testify/assert`
- Use `require.ErrorContains(t, err, test.wantErr)` instead of `require.Contains(t, err.Error(), test.wantErr)`
- Use `require` for critical assertions, `assert` for non-critical ones
- No explanatory messages in assertions

### Implementation Context
- Update should support partial updates (only provided fields are updated)
- Delete should handle scheduled vs published posts appropriately
- Mock server should maintain post state consistency across operations
- Each test should call server.Reset() before configuring responses
- Tests must run sequentially to avoid mock server state conflicts
- Use existing Post and Media types from Phase 1
- URL path extraction for post ID (e.g., `/api/v1/posts/{id}`)

### Validation Commands
After implementation, run these commands to verify:
- `go test ./v1 -run TestGetPost`
- `go test ./v1 -run TestUpdatePost`
- `go test ./v1 -run TestDeletePost`

### Success Criteria
This phase is complete when:
1. Individual posts can be retrieved by ID
2. Posts can be partially updated (text, scheduled_at, media)
3. Posts can be deleted successfully
4. 404 errors are returned for non-existent posts
5. Mock server maintains consistent post state
6. All tests pass without race conditions
7. State transitions are handled appropriately