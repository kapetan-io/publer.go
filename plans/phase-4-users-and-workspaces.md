# Phase 4: User and Workspace Operations

## Implementation Instructions for Claude Code

**IMPORTANT CONTEXT**: This phase implements user profile retrieval and workspace listing operations. You are building on previous phases' foundation.

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

### API Context
- **User profile**: GET `/api/v1/users/me` returns current user information
- **Workspace listing**: GET `/api/v1/workspaces` returns paginated workspaces
- **Authentication**: Uses Bearer API key + Workspace ID headers from Phase 0
- **User context**: The User type is already defined in Phase 1 (`v1/posts.go`)

### Changes Required

#### 1. User Operations
**File**: `v1/users.go`
**Purpose**: Define user-related operations

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

**Function Responsibilities:**
- GetMe: GET to `/api/v1/users/me`
- Use existing User struct from Phase 1 (`v1/posts.go`)
- Apply standard error handling from Phase 0

#### 2. Workspace Operations
**File**: `v1/workspaces.go`
**Purpose**: Define workspace-related operations

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
    Workspaces []Workspace `json:"workspaces"`
    Total      int         `json:"total"`
    Page       int         `json:"page"`
    PerPage    int         `json:"per_page"`
    TotalPages int         `json:"total_pages"`
}

// ListWorkspaces retrieves all workspaces for the authenticated user
func (c *Client) ListWorkspaces(ctx context.Context, req ListWorkspacesRequest) Iterator[Workspace]
```

**Function Responsibilities:**
- ListWorkspaces: Return iterator that fetches from GET `/api/v1/workspaces`
- Handle nested user objects in workspace responses (owner, members)
- Use generic iterator from Phase 0
- Apply standard error handling

#### 3. Mock Server Updates
**File**: `v1/mock_server.go` (extend existing)
**Purpose**: Add user and workspace endpoints

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
- Return paginated list of workspaces with members
- Validate workspace ID header presence (from Phase 0 authentication)
- Support multiple workspace scenarios

### Testing Requirements
**CRITICAL**: Do NOT use `t.Parallel()` in any tests

```go
func TestGetMe(t *testing.T)
func TestListWorkspaces(t *testing.T)
func TestWorkspaceMembers(t *testing.T)
```

**Test Objectives:**
- Test user profile retrieval
- Verify workspace listing with members using iterator
- Test authentication failures
- Validate workspace switching scenarios

**Test Requirements from CLAUDE.md:**
- Test MUST always be in the test package `package v1_test` and not `package v1`
- Test names should be in camelCase and start with a capital letter
- Use `github.com/stretchr/testify/require` and `github.com/stretchr/testify/assert`
- Use `require.ErrorContains(t, err, test.wantErr)` instead of `require.Contains(t, err.Error(), test.wantErr)`
- Use `require` for critical assertions, `assert` for non-critical ones
- No explanatory messages in assertions

### Implementation Context
- Mock server should return consistent user data across endpoints
- Workspaces should include owner and member details using User type
- Support testing of multi-workspace scenarios
- Each test should call server.Reset() before configuring responses
- Tests must run sequentially to avoid mock server state conflicts
- User type is already defined in Phase 1, reuse it

### Validation Commands
After implementation, run these commands to verify:
- `go test ./v1 -run TestGetMe`
- `go test ./v1 -run TestWorkspace`

### Success Criteria
This phase is complete when:
1. Current user profile can be retrieved successfully
2. Workspaces can be listed using iterator with pagination support
3. Workspace members are properly included in responses
4. Authentication errors are handled appropriately
5. Mock server supports user and workspace scenarios consistently
6. All tests pass without race conditions
7. Iterator works correctly for workspace pagination