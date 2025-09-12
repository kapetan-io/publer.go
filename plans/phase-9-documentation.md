# Phase 9: Documentation and Examples

## Implementation Instructions for Claude Code

**IMPORTANT CONTEXT**: This phase completes the implementation with comprehensive documentation and usage examples. This is the final phase that makes the library ready for use.

### What You Should NOT Do
- Do NOT create CLI tools or command-line interface
- Do NOT use `pkg/` directory structure
- Do NOT implement request timeout or retry logic (handled by caller via context)
- Do NOT create unit tests with HTTP client mocking (all tests use mock server)
- Do NOT use `t.Parallel()` in ANY tests - all tests must run sequentially

### Prerequisites
You must have completed ALL previous phases (0-9):
- **Phase 0**: Foundation (client, errors, iterator, mock server)
- **Phase 1**: List posts and publish post
- **Phase 2**: Schedule and draft posts  
- **Phase 3**: User and workspace operations
- **Phase 4**: Account operations
- **Phase 5**: Bulk operations
- **Phase 6**: Post management (update/delete/get)
- **Phase 7**: Post iterator implementation
- **Phase 8**: Convenience methods
- **Phase 9**: Advanced post features

### Goal
Create comprehensive package documentation and runnable examples that demonstrate all major functionality.

### Changes Required

#### 1. Package Documentation
**File**: `v1/doc.go`
**Purpose**: Package-level documentation

```go
// Package publer provides a Go client library for the Publer.com API v1.
//
// The client supports all major Publer API operations including post management,
// scheduling, bulk operations, and workspace management.
//
// Basic usage:
//
//	config := publer.Config{
//	    APIKey:      "your-api-key",
//	    WorkspaceID: "your-workspace-id",
//	}
//	client, err := publer.NewClient(config)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// List posts
//	iter := client.ListPosts(ctx, publer.ListPostsRequest{})
//	for {
//		var page publer.Page[publer.Post]
//		if !iter.Next(ctx, &page) {
//			break
//		}
//		// Process page.Items...
//		log.Printf("Page %d of %d, %d posts\n", page.Page, page.TotalPages, len(page.Items))
//	}
//	if err := iter.Err(); err != nil {
//		log.Fatal(err)
//	}
//
// Authentication:
//
// The client requires a Bearer API key and workspace ID from your Publer account.
// Both are required for all API operations.
//
// Rate Limiting:
//
// The API has rate limits of 100 requests per 2 minutes. Rate limit exceeded
// errors include rate limit information that can be accessed via type assertion:
//
//	if rateLimitErr, ok := err.(*publer.RateLimitError); ok {
//	    log.Printf("Rate limited: %d/%d, resets at %v",
//	        rateLimitErr.Remaining, rateLimitErr.Limit,
//	        time.Unix(rateLimitErr.Reset, 0))
//	}
//
// Context Support:
//
// All operations support context.Context for cancellation and timeouts:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//	
//	err := client.PublishPost(ctx, req, &resp)
//
// Mock Server:
//
// The package includes a comprehensive mock server for testing:
//
//	server := publer.NewMockServer()
//	url, _ := server.Start()
//	defer server.Stop()
//	
//	client, _ := publer.NewClient(publer.Config{
//	    APIKey: "test-key",
//	    WorkspaceID: "test-workspace",
//	    BaseURL: url,
//	})
package publer
```

#### 2. Example Files
**File**: `v1/examples_test.go`
**Purpose**: Runnable examples for godoc

```go
func ExampleClient_ListPosts() {
    // Create mock server for example
    server := NewMockServer()
    url, _ := server.Start()
    defer server.Stop()
    
    // Configure mock data
    server.AddPosts([]Post{
        {ID: "1", Text: "Example post", State: "published"},
        {ID: "2", Text: "Another post", State: "scheduled"},
    })
    
    client, _ := NewClient(Config{
        APIKey:      "example-key",
        WorkspaceID: "example-workspace",
        BaseURL:     url,
    })
    
    ctx := context.Background()
    iter := client.ListPosts(ctx, ListPostsRequest{})
    
    for {
        var page Page[Post]
        if !iter.Next(ctx, &page) {
            break
        }
        for _, post := range page.Items {
            fmt.Printf("Post: %s - %s\n", post.ID, post.Text)
        }
    }
    
    if err := iter.Err(); err != nil {
        log.Printf("Error: %v", err)
    }
    
    // Output:
    // Post: 1 - Example post
    // Post: 2 - Another post
}

func ExampleClient_PublishPost() {
    server := NewMockServer()
    url, _ := server.Start()
    defer server.Stop()
    
    // Configure job completion
    server.SetJobStatus("job-123", "completed", 100, &JobResult{
        PostIDs: []string{"post-456"},
    }, "")
    
    client, _ := NewClient(Config{
        APIKey:      "example-key",
        WorkspaceID: "example-workspace", 
        BaseURL:     url,
    })
    
    ctx := context.Background()
    req := PublishPostRequest{
        Text:     "Hello, world!",
        Accounts: []string{"account-1"},
    }
    
    var resp PublishPostResponse
    err := client.PublishPost(ctx, req, &resp)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Job ID: %s\n", resp.JobID)
    
    // Wait for job completion
    var result JobResult
    err = client.WaitForJob(ctx, resp.JobID, &result)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Published posts: %v\n", result.PostIDs)
    
    // Output:
    // Job ID: job-123
    // Published posts: [post-456]
}

func ExampleClient_GetPostsByState() {
    server := NewMockServer()
    url, _ := server.Start()
    defer server.Stop()
    
    server.AddPosts([]Post{
        {ID: "1", Text: "Draft post", State: "draft"},
        {ID: "2", Text: "Published post", State: "published"},
    })
    
    client, _ := NewClient(Config{
        APIKey:      "example-key",
        WorkspaceID: "example-workspace",
        BaseURL:     url,
    })
    
    ctx := context.Background()
    iter := client.GetPostsByState("draft")
    
    for {
        var page Page[Post]
        if !iter.Next(ctx, &page) {
            break
        }
        for _, post := range page.Items {
            fmt.Printf("Draft: %s\n", post.Text)
        }
    }
    
    // Output:
    // Draft: Draft post
}

func ExampleClient_WaitForJob() {
    server := NewMockServer()
    url, _ := server.Start()
    defer server.Stop()
    
    // Simulate job progression
    server.SetJobStatus("job-123", "working", 50, nil, "")
    // Later the job completes
    server.SetJobStatus("job-123", "completed", 100, &JobResult{
        PostIDs: []string{"post-456"},
        Message: "Post published successfully",
    }, "")
    
    client, _ := NewClient(Config{
        APIKey:      "example-key",
        WorkspaceID: "example-workspace",
        BaseURL:     url,
    })
    
    ctx := context.Background()
    var result JobResult
    err := client.WaitForJob(ctx, "job-123", &result)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Result: %s\n", result.Message)
    fmt.Printf("Posts: %v\n", result.PostIDs)
    
    // Output:
    // Result: Post published successfully
    // Posts: [post-456]
}

func ExampleClient_BulkSchedulePosts() {
    server := NewMockServer()
    url, _ := server.Start()
    defer server.Stop()
    
    server.SetJobStatus("bulk-job-123", "completed", 100, &JobResult{
        PostIDs: []string{"post-1", "post-2"},
    }, "")
    
    client, _ := NewClient(Config{
        APIKey:      "example-key",
        WorkspaceID: "example-workspace",
        BaseURL:     url,
    })
    
    ctx := context.Background()
    req := BulkSchedulePostsRequest{
        Posts: []BulkPost{
            {
                Text:        "First post",
                Accounts:    []string{"account-1"},
                ScheduledAt: time.Now().Add(time.Hour),
            },
            {
                Text:        "Second post", 
                Accounts:    []string{"account-1"},
                ScheduledAt: time.Now().Add(2 * time.Hour),
            },
        },
    }
    
    var resp BulkSchedulePostsResponse
    err := client.BulkSchedulePosts(ctx, req, &resp)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Bulk job ID: %s\n", resp.JobID)
    
    // Output:
    // Bulk job ID: bulk-job-123
}
```

#### 3. Example Test Validation
**File**: `v1/examples_test.go` (add to same file)
**Purpose**: Ensure examples work correctly

```go
func TestExamples(t *testing.T) {
    // Test that all examples compile and run without errors
    // This validates the examples are correct and up-to-date
    
    t.Run("ListPostsExample", func(t *testing.T) {
        ExampleClient_ListPosts()
    })
    
    t.Run("PublishPostExample", func(t *testing.T) {
        ExampleClient_PublishPost()
    })
    
    t.Run("GetPostsByStateExample", func(t *testing.T) {
        ExampleClient_GetPostsByState()
    })
    
    t.Run("WaitForJobExample", func(t *testing.T) {
        ExampleClient_WaitForJob()
    })
    
    t.Run("BulkSchedulePostsExample", func(t *testing.T) {
        ExampleClient_BulkSchedulePosts()
    })
}
```

### Testing Requirements
**CRITICAL**: Do NOT use `t.Parallel()` in any tests

**Test Objectives:**
- Ensure all examples compile and run successfully
- Verify example output matches expectations
- Test that examples use mock server correctly
- Validate that examples demonstrate key functionality

**Test Requirements from CLAUDE.md:**
- Test MUST always be in the test package `package v1_test` and not `package v1`
- Test names should be in camelCase and start with a capital letter
- Use `github.com/stretchr/testify/require` and `github.com/stretchr/testify/assert`
- No explanatory messages in assertions

### Implementation Context
- Examples should be self-contained and runnable
- Include proper error handling in examples
- Show both simple and complex use cases
- Examples should use mock server, not real API calls
- Each example should call server.Reset() if needed
- Tests must run sequentially to avoid mock server state conflicts

### Validation Commands
After implementation, run these commands to verify:
- `go test ./v1 -run Example`
- `go doc -all ./v1`
- `go test ./v1 -run TestExamples`

### Success Criteria
This phase is complete when:
1. Package documentation is comprehensive and clear
2. All examples compile and run correctly
3. Examples demonstrate major functionality
4. Error handling patterns are shown
5. Mock server usage is documented
6. Godoc output looks professional and complete
7. All tests pass without race conditions
8. Examples are self-contained and educational