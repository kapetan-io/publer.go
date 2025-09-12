# Phase 5: Account Operations

## Implementation Instructions for Claude Code

**IMPORTANT CONTEXT**: This phase implements social media account listing and management. You are building on previous phases' foundation.

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

### API Context
- **Account listing**: GET `/api/v1/accounts` returns paginated social media accounts
- **Authentication**: Uses Bearer API key + Workspace ID headers from Phase 0
- **Account scope**: Accounts are workspace-specific
- **Providers**: Support for facebook, instagram, twitter, linkedin, etc.

### Changes Required

#### 1. Account Types
**File**: `v1/accounts.go`
**Purpose**: Define account-related types

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
    Accounts   []Account `json:"accounts"`
    Total      int       `json:"total"`
    Page       int       `json:"page"`
    PerPage    int       `json:"per_page"`
    TotalPages int       `json:"total_pages"`
}

// ListAccounts retrieves all social media accounts in the workspace
func (c *Client) ListAccounts(ctx context.Context, req ListAccountsRequest) Iterator[Account]
```

**Function Responsibilities:**
- ListAccounts: Return iterator that fetches from GET `/api/v1/accounts`
- Handle various provider types (facebook, instagram, twitter, linkedin, etc.)
- Parse account type variations (page, profile, group, etc.)
- Use generic iterator from Phase 0
- Apply standard error handling

#### 2. Mock Server Updates
**File**: `v1/mock_server.go` (extend existing)
**Purpose**: Add account endpoints

```go
// handleListAccounts handles GET /api/v1/accounts
func (m *MockServer) handleListAccounts(w http.ResponseWriter, r *http.Request)

// AddAccount adds a social media account to mock data
func (m *MockServer) AddAccount(account Account)

// SetAccountsByProvider sets accounts filtered by provider
func (m *MockServer) SetAccountsByProvider(provider string, accounts []Account)
```

**Function Responsibilities:**
- Return configured social media accounts with pagination
- Support filtering by provider type (if needed for testing)
- Validate workspace context (workspace ID header)
- Simulate various account types and providers

### Testing Requirements
**CRITICAL**: Do NOT use `t.Parallel()` in any tests

```go
func TestListAccounts(t *testing.T)
func TestAccountProviders(t *testing.T)
func TestAccountTypes(t *testing.T)
```

**Test Objectives:**
- Test account listing for various providers (facebook, instagram, twitter, etc.)
- Verify account type handling (page, profile, group)
- Test empty account scenarios
- Validate workspace-specific accounts
- Test iterator pagination for accounts

**Test Requirements from CLAUDE.md:**
- Test MUST always be in the test package `package v1_test` and not `package v1`
- Test names should be in camelCase and start with a capital letter
- Use `github.com/stretchr/testify/require` and `github.com/stretchr/testify/assert`
- Use `require.ErrorContains(t, err, test.wantErr)` instead of `require.Contains(t, err.Error(), test.wantErr)`
- Use `require` for critical assertions, `assert` for non-critical ones
- No explanatory messages in assertions

### Implementation Context
- Mock server should support all major provider types (facebook, instagram, twitter, linkedin, youtube, etc.)
- Account types vary by provider (facebook: page/profile, instagram: business/personal, etc.)
- Accounts are workspace-specific and should respect workspace ID header
- Each test should call server.Reset() before configuring responses
- Tests must run sequentially to avoid mock server state conflicts
- Use existing generic iterator pattern from Phase 0

### Validation Commands
After implementation, run these commands to verify:
- `go test ./v1 -run TestAccount`

### Success Criteria
This phase is complete when:
1. Social media accounts can be listed using iterator with pagination
2. Various provider types are properly handled (facebook, instagram, etc.)
3. Account types are correctly parsed (page, profile, group, etc.)
4. Empty account scenarios work correctly
5. Workspace-specific account filtering works
6. Mock server supports realistic account scenarios
7. All tests pass without race conditions