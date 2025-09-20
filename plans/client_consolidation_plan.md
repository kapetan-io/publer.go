# v1/Client Method Consolidation Implementation Plan

## Overview

Consolidate all 23 Client methods from 10 separate files into `v1/client.go` to centralize the Client implementation while maintaining logical organization and removing empty files.

## Current State Analysis

The v1 package currently has Client methods distributed across multiple files:

### Files with Client Methods:
- `v1/client.go` - Core Client type, constructor, `do()` method, `Test()` method
- `v1/posts_operations.go` - Basic post operations (2 methods)
- `v1/posts_management_operations.go` - CRUD operations (3 methods + validation helper)
- `v1/posts_schedule_operations.go` - Scheduling operations (2 methods)
- `v1/posts_bulk_operations.go` - Bulk operations (2 methods)
- `v1/posts_advanced_operations.go` - Advanced features (3 methods)
- `v1/posts_convenience.go` - Convenience methods (4 methods)
- `v1/accounts.go` - Account operations (1 method + page fetcher type)
- `v1/users.go` - User operations (1 method)
- `v1/workspaces.go` - Workspace operations (1 method + page fetcher type)
- `v1/jobs.go` - Job management (2 methods)

### Key Discoveries:
- Current modular approach follows clean separation of concerns at `v1/posts_operations.go:8-15`
- Validation helper `validatePostID` at `v1/posts_management_operations.go:14-25` prevents path traversal
- Page fetcher types implement `PageFetcher[T]` interface for iterator pattern
- All methods follow consistent error handling via `c.do()` at `v1/client.go:62-189`

## Desired End State

A single `v1/client.go` file containing:
1. All Client type definitions and constructor
2. All 23 Client methods organized by functionality
3. All helper functions and validation logic
4. All page fetcher implementations
5. Publishing methods prioritized at the top

**Verification**: `go test ./v1` and `make ci` should pass with no behavioral changes.

## What We're NOT Doing

- Changing any method signatures or behavior
- Modifying test files or test logic
- Altering import dependencies outside of v1 package
- Changing the Client struct definition or constructor logic
- Modifying iterator interfaces or `types.go` definitions

## Implementation Approach

Single-phase consolidation with careful import management and logical grouping to maintain code readability while centralizing all Client functionality.

## Phase 1: Consolidate All Client Methods

### Overview
Move all Client methods, helper functions, and supporting types into `v1/client.go` with logical organization and remove empty files.

### Changes Required:

#### 1. Core Client File Enhancement
**File**: `v1/client.go`
**Changes**: Add all Client methods with organized sections and required imports

```go
// Package-level variables (from posts_management_operations.go)
var postIDRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// Core Client functionality (existing)
func NewClient(config Config) (*Client, error)
func (c *Client) do(ctx context.Context, method, path string, body any, result any) error
func (c *Client) Test(ctx context.Context) error

// Post Publishing Operations
func (c *Client) PublishPost(ctx context.Context, request PublishPostRequest, response *PublishPostResponse) error
func (c *Client) BulkPublishPosts(ctx context.Context, req BulkPublishPostsRequest, resp *BulkPublishPostsResponse) error

// Post Scheduling Operations
func (c *Client) SchedulePost(ctx context.Context, req SchedulePostRequest, resp *SchedulePostResponse) error
func (c *Client) CreateDraftPost(ctx context.Context, req CreateDraftPostRequest, resp *CreateDraftPostResponse) error
func (c *Client) BulkSchedulePosts(ctx context.Context, req BulkSchedulePostsRequest, resp *BulkSchedulePostsResponse) error

// Post Management Operations
func validatePostID(postID string) error
func (c *Client) GetPost(ctx context.Context, req GetPostRequest, resp *GetPostResponse) error
func (c *Client) UpdatePost(ctx context.Context, req UpdatePostRequest, resp *UpdatePostResponse) error
func (c *Client) DeletePost(ctx context.Context, req DeletePostRequest, resp *DeletePostResponse) error

// Post Listing Operations
func (c *Client) ListPosts(ctx context.Context, request ListPostsRequest) Iterator[Post]

// Post Advanced Operations
func (c *Client) CreateRecurringPost(ctx context.Context, req RecurringPostRequest, resp *RecurringPostResponse) error
func (c *Client) AutoSchedulePost(ctx context.Context, req AutoScheduleRequest, resp *AutoScheduleResponse) error
func (c *Client) RecyclePost(ctx context.Context, req RecyclePostRequest, resp *RecyclePostResponse) error

// Post Convenience Operations
func (c *Client) GetPostsByState(state string) Iterator[Post]
func (c *Client) GetPostsByDateRange(from, to time.Time) Iterator[Post]
func (c *Client) GetPostsByAccount(accountID string) Iterator[Post]
func (c *Client) GetPostsByQuery(query string) Iterator[Post]

// Account Operations
type accountFetcher struct
func (f *accountFetcher) FetchPage(ctx context.Context, pageNum int) (*Page[Account], error)
func (c *Client) ListAccounts(ctx context.Context, req ListAccountsRequest) Iterator[Account]

// User Operations
func (c *Client) GetMe(ctx context.Context, req GetMeRequest, resp *GetMeResponse) error

// Workspace Operations
type workspacePageFetcher struct
func (f *workspacePageFetcher) FetchPage(ctx context.Context, pageNum int) (*Page[Workspace], error)
func (c *Client) ListWorkspaces(ctx context.Context, req ListWorkspacesRequest) Iterator[Workspace]

// Job Management Operations
func (c *Client) GetJobStatus(ctx context.Context, req GetJobStatusRequest, resp *GetJobStatusResponse) error
func (c *Client) WaitForJob(ctx context.Context, opts WaitOptions, result *JobResult) error
```

**Function Responsibilities:**
- Move all method implementations exactly as they exist
- Preserve all validation logic including `validatePostID` pattern and `postIDRegex` variable
- Maintain all page fetcher struct definitions and implementations
- Add required imports: `regexp` (for postIDRegex), `time` (for convenience methods)
- Move package-level variable: `var postIDRegex = regexp.MustCompile(^[a-zA-Z0-9_-]+$)` before validatePostID function
- Organize methods in logical sections with clear comment headers
- Ensure publishing operations appear first as requested

**Testing Requirements:**
```go
// All existing tests should continue to pass without modification
func TestClientMethods(t *testing.T)
func TestValidatePostID(t *testing.T)
```

**Test Objectives:**
- Verify all Client methods work identically after consolidation
- Confirm method signatures and behavior remain unchanged
- Validate import resolution and compilation
- Ensure iterator patterns continue functioning

**Context for implementation:**
- Follow existing import organization pattern from `v1/client.go:3-12`
- Preserve exact method implementations from source files
- Maintain existing error handling patterns via `c.do()` method
- Keep all struct field ordering and JSON tags intact
- **⚠️ Critical**: Must move `postIDRegex` variable before `validatePostID` function to maintain compilation order
- **⚠️ Critical**: Move package-level variables from `posts_management_operations.go:10` to maintain validation functionality
- Follow CLAUDE.md guidelines for Go code organization

#### 2. File Removal Operations
**Files to Remove**: Files that become empty after method extraction

**Changes**: Delete empty files after confirming no remaining content

**Function Responsibilities:**
- Remove `v1/posts_operations.go` after moving `ListPosts`, `PublishPost`
- Remove `v1/posts_management_operations.go` after moving methods and validation
- Remove `v1/posts_schedule_operations.go` after moving scheduling methods
- Remove `v1/posts_bulk_operations.go` after moving bulk operations
- Remove `v1/posts_advanced_operations.go` after moving advanced methods
- Remove `v1/posts_convenience.go` after moving convenience methods
- Remove `v1/accounts.go` after moving account operations and fetcher
- Remove `v1/users.go` after moving user operations
- Remove `v1/workspaces.go` after moving workspace operations and fetcher
- Remove `v1/jobs.go` after moving job operations

**Testing Requirements:**
```go
// Verify package builds and imports resolve correctly
func TestPackageCompilation(t *testing.T)
```

**Test Objectives:**
- Confirm package compiles without import errors
- Verify no missing symbols or undefined references
- Validate all existing tests continue to pass

**Context for implementation:**
- Only remove files that contain exclusively Client methods and supporting types
- Verify each file is completely empty of useful code before deletion
- Check that no other files import from the files being removed
- Ensure test files are not affected by consolidation

### Validation Commands
- `go build -v ./v1` - Verify package compiles correctly with verbose output
- `go test ./v1` - Ensure all tests pass
- `make test` - Run full test suite
- `make lint` - Verify code quality standards
- `go mod tidy && make lint` - Check for unused imports after consolidation
- `make ci` - Complete CI validation (tidy, lint, test)

### Dependencies and Integration Points
- Iterator pattern via `v1/iterator.go` interfaces
- Type definitions from `v1/types.go` remain unchanged
- Error types from `v1/errors.go` remain unchanged
- Test files reference Client methods by name - no changes needed
- Mock server in `v1/mock_server.go` returns `*Client` - unchanged

## Success Criteria
The consolidation serves as a code organization improvement providing:
- Single location for all Client method definitions
- Logical grouping with publishing operations prioritized
- Elimination of scattered Client method files
- Preserved functionality with no behavioral changes
- Maintained test coverage and code quality standards
- Clean package structure with minimal files