# Phase 2: Iterator Implementation for Posts

## Implementation Instructions for Claude Code

**IMPORTANT CONTEXT**: This phase adapts the generic iterator for post-specific pagination. You are building on the generic iterator from Phase 0 to create post-specific implementations.

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
- **Phase 1**: Post operations and job handling (includes ListPosts method)
- **Phase 2-6**: All previous phases for complete post functionality

### Context
- **Generic Iterator**: Already implemented in Phase 0 (`v1/iterator.go`)
- **ListPosts Method**: Already implemented in Phase 1 (`v1/posts_operations.go`)
- **Goal**: Create post-specific PageFetcher implementation for the generic iterator

### Changes Required

#### 1. Post Iterator Implementation
**File**: `v1/posts_iterator.go`
**Purpose**: Create post-specific iterator implementation

```go
// PostPageFetcher implements PageFetcher for posts
type PostPageFetcher struct {
    client  *Client
    request ListPostsRequest
}

// FetchPage fetches a page of posts
func (f *PostPageFetcher) FetchPage(ctx context.Context, pageNum int) (*Page[Post], error) {
    // Create a copy of the request with the specific page number
    req := f.request
    req.Page = pageNum
    
    // Make API call to get posts
    var resp ListPostsResponse
    err := f.client.do(ctx, "GET", "/posts", req, &resp)
    if err != nil {
        return nil, err
    }
    
    // Map ListPostsResponse to Page[Post] structure
    return &Page[Post]{
        Items:      resp.Posts,
        Total:      resp.Total,
        Page:       resp.Page,
        PerPage:    resp.PerPage,
        TotalPages: resp.TotalPages,
    }, nil
}

// NewPostIterator creates a new iterator for posts
func NewPostIterator(client *Client, request ListPostsRequest) Iterator[Post] {
    fetcher := &PostPageFetcher{
        client:  client,
        request: request,
    }
    return NewGenericIterator(fetcher)
}
```

#### 2. Update ListPosts Method
**File**: `v1/posts_operations.go` (modify existing)
**Purpose**: Update ListPosts to use the new iterator implementation

```go
// ListPosts retrieves posts with filtering options
func (c *Client) ListPosts(ctx context.Context, req ListPostsRequest) Iterator[Post] {
    return NewPostIterator(c, req)
}
```

**Function Responsibilities:**
- Implement PageFetcher interface for posts
- Handle page parameter in ListPostsRequest
- Map ListPostsResponse to Page[Post] structure
- Return configured iterator instance using generic iterator from Phase 0
- Support all filtering options from ListPostsRequest

### Testing Requirements
**CRITICAL**: Do NOT use `t.Parallel()` in any tests

```go
func TestListPostsIterator(t *testing.T)
func TestPostIteratorPagination(t *testing.T)
func TestPostIteratorError(t *testing.T)
func TestPostIteratorContext(t *testing.T)
```

**Test Objectives:**
- Test iteration through multiple pages of posts
- Verify lazy loading on first Next() call
- Test error propagation during iteration
- Validate context cancellation support
- Test with various ListPostsRequest filters

**Test Requirements from CLAUDE.md:**
- Test MUST always be in the test package `package v1_test` and not `package v1`
- Test names should be in camelCase and start with a capital letter
- Use `github.com/stretchr/testify/require` and `github.com/stretchr/testify/assert`
- Use `require.ErrorContains(t, err, test.wantErr)` instead of `require.Contains(t, err.Error(), test.wantErr)`
- Use `require` for critical assertions, `assert` for non-critical ones
- No explanatory messages in assertions

### Implementation Context
- Iterator should respect all ListPostsRequest filters (state, date range, query, etc.)
- Page fetching should be lazy (on-demand, not prefetched)
- Context cancellation should stop iteration and set Err() to context.Cancelled
- Configure multi-page responses in mock server for pagination testing
- Each test should call server.Reset() before configuring responses
- Tests must run sequentially to avoid mock server state conflicts
- Reuse generic iterator logic from Phase 0, don't reimplement pagination

### Mock Server Configuration for Testing
```go
// Configure multi-page post response
server.Reset()
posts := []Post{
    {ID: "1", Text: "Post 1"},
    {ID: "2", Text: "Post 2"},
    // ... more posts
}
server.AddPosts(posts)

// Configure pagination metadata
server.SetResponse("GET", "/api/v1/posts", 200, ListPostsResponse{
    Posts: posts[:10], // First 10 posts
    Total: len(posts),
    Page: 1,
    PerPage: 10,
    TotalPages: (len(posts) + 9) / 10, // Ceil division
})
```

### Validation Commands
After implementation, run these commands to verify:
- `go test ./v1 -run TestPostIterator`

### Success Criteria
This phase is complete when:
1. Post-specific iterator is properly implemented
2. Iterator works with all ListPostsRequest filtering options
3. Pagination works correctly across multiple pages
4. Context cancellation is handled properly
5. Error scenarios are propagated correctly
6. Lazy loading works (first Next() initializes)
7. All tests pass without race conditions
8. Integration with existing ListPosts method is seamless