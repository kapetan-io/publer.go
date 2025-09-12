# Phase 10: Integration Testing and Final Validation

## Implementation Instructions for Claude Code

**IMPORTANT CONTEXT**: This is the final phase that validates the complete implementation works as a cohesive system. You are implementing integration tests that verify all phases work together correctly.

### What You Should NOT Do
- Do NOT create CLI tools or command-line interface
- Do NOT use `pkg/` directory structure
- Do NOT implement request timeout or retry logic (handled by caller via context)
- Do NOT create unit tests with HTTP client mocking (all tests use mock server)
- Do NOT use `t.Parallel()` in ANY tests - all tests must run sequentially

### Prerequisites
You must have completed ALL previous phases (0-9):
- **Phase 0**: Foundation (client, errors, iterator, mock server, core types)
- **Phase 1**: List posts and publish post with job tracking
- **Phase 2**: Post iterator implementation
- **Phase 3**: Schedule and draft posts  
- **Phase 4**: User and workspace operations
- **Phase 5**: Account operations
- **Phase 6**: Bulk operations
- **Phase 7**: Post management (update/delete/get)
- **Phase 8**: Advanced features and convenience methods
- **Phase 9**: Documentation and examples

### Goal
Create comprehensive integration tests that validate the entire client library works correctly as a unified system.

### Changes Required

#### 1. Integration Test Suite
**File**: `v1/integration_test.go`
**Purpose**: End-to-end integration tests

```go
// TestFullWorkflow tests a complete user workflow
func TestFullWorkflow(t *testing.T) {
    // This test validates the entire client library by:
    // 1. Creating client and authenticating
    // 2. Listing workspaces and accounts
    // 3. Creating and publishing posts
    // 4. Managing posts (update, delete)
    // 5. Using convenience methods
    // 6. Testing error scenarios
}

// TestConcurrentOperations tests multiple operations without race conditions
func TestConcurrentOperations(t *testing.T) {
    // Test that multiple sequential operations work correctly
    // Note: NOT using t.Parallel() but testing sequential operations
}

// TestErrorHandling validates all error scenarios work correctly
func TestErrorHandling(t *testing.T) {
    // Test rate limiting, authentication failures, network errors
}

// TestIteratorPerformance validates iterator performance with large datasets
func TestIteratorPerformance(t *testing.T) {
    // Test iteration over multiple pages with large datasets
}

// TestJobTracking validates job tracking across all operations
func TestJobTracking(t *testing.T) {
    // Test job tracking for all async operations (publish, schedule, bulk, etc.)
}
```

#### 2. Performance Benchmarks
**File**: `v1/benchmark_test.go`
**Purpose**: Performance testing

```go
// BenchmarkListPosts benchmarks post listing performance
func BenchmarkListPosts(b *testing.B)

// BenchmarkIteratorPagination benchmarks iterator performance
func BenchmarkIteratorPagination(b *testing.B)

// BenchmarkJobPolling benchmarks WaitForJob performance
func BenchmarkJobPolling(b *testing.B)

// BenchmarkBulkOperations benchmarks bulk post operations
func BenchmarkBulkOperations(b *testing.B)
```

#### 3. Mock Server Stress Testing
**File**: `v1/mock_server_stress_test.go`
**Purpose**: Validate mock server handles complex scenarios

```go
// TestMockServerComplexScenarios tests complex mock server usage
func TestMockServerComplexScenarios(t *testing.T) {
    // Test mock server with:
    // - Multiple job progressions
    // - Rate limiting scenarios
    // - Large paginated responses
    // - Error condition simulation
}

// TestMockServerStateManagement tests state consistency
func TestMockServerStateManagement(t *testing.T) {
    // Validate mock server state remains consistent across operations
}
```

#### 4. Example Validation
**File**: `v1/examples_validation_test.go`
**Purpose**: Ensure all examples work correctly

```go
// TestAllExamples validates that all documentation examples work
func TestAllExamples(t *testing.T) {
    // Run all examples from Phase 9 documentation
    // Validate they produce expected output
}

// TestExampleErrorHandling tests error scenarios in examples
func TestExampleErrorHandling(t *testing.T) {
    // Test that examples handle errors appropriately
}
```

#### 5. Client Configuration Testing
**File**: `v1/client_configuration_test.go`
**Purpose**: Validate client configuration options

```go
// TestClientConfigurationVariations tests different client configs
func TestClientConfigurationVariations(t *testing.T) {
    // Test with different:
    // - Base URLs
    // - HTTP clients
    // - Timeout configurations
    // - Authentication scenarios
}

// TestClientResilience tests client resilience
func TestClientResilience(t *testing.T) {
    // Test client behavior with:
    // - Network timeouts
    // - Connection failures  
    // - Invalid responses
    // - Context cancellation
}
```

### Integration Test Scenarios

#### Scenario 1: Complete User Journey
```go
func TestCompleteUserJourney(t *testing.T) {
    server := NewMockServer()
    url, err := server.Start()
    require.NoError(t, err)
    defer server.Stop()
    
    // Setup comprehensive mock data
    server.Reset()
    // ... configure accounts, posts, jobs, etc.
    
    client, err := NewClient(Config{
        APIKey:      "test-key",
        WorkspaceID: "test-workspace",
        BaseURL:     url,
    })
    require.NoError(t, err)
    
    ctx := context.Background()
    
    // 1. Get user info
    var userResp GetMeResponse
    err = client.GetMe(ctx, GetMeRequest{}, &userResp)
    require.NoError(t, err)
    
    // 2. List workspaces
    workspaceIter := client.ListWorkspaces(ctx, ListWorkspacesRequest{})
    // ... iterate and validate
    
    // 3. List accounts
    accountIter := client.ListAccounts(ctx, ListAccountsRequest{})
    // ... iterate and validate
    
    // 4. List existing posts
    postIter := client.ListPosts(ctx, ListPostsRequest{})
    // ... iterate and validate
    
    // 5. Create and publish new post
    publishReq := PublishPostRequest{
        Text:     "Integration test post",
        Accounts: []string{"account-1"},
    }
    var publishResp PublishPostResponse
    err = client.PublishPost(ctx, publishReq, &publishResp)
    require.NoError(t, err)
    
    // 6. Wait for job completion
    var jobResult JobResult
    err = client.WaitForJob(ctx, publishResp.JobID, &jobResult)
    require.NoError(t, err)
    
    // 7. Schedule a post for later
    scheduleReq := SchedulePostRequest{
        Text:        "Scheduled post",
        Accounts:    []string{"account-1"},
        ScheduledAt: time.Now().Add(time.Hour),
    }
    var scheduleResp SchedulePostResponse
    err = client.SchedulePost(ctx, scheduleReq, &scheduleResp)
    require.NoError(t, err)
    
    // 8. Use convenience methods
    draftIter := client.GetPostsByState("draft")
    // ... iterate and validate
    
    // 9. Test bulk operations
    bulkReq := BulkPublishPostsRequest{
        Posts: []BulkPost{
            {Text: "Bulk post 1", Accounts: []string{"account-1"}},
            {Text: "Bulk post 2", Accounts: []string{"account-1"}},
        },
    }
    var bulkResp BulkPublishPostsResponse
    err = client.BulkPublishPosts(ctx, bulkReq, &bulkResp)
    require.NoError(t, err)
    
    // 10. Test advanced features
    recurringReq := RecurringPostRequest{
        Text:     "Recurring post",
        Accounts: []string{"account-1"},
        Recurrence: RecurrenceRule{
            Frequency: "daily",
            Interval:  1,
            Count:     5,
        },
    }
    var recurringResp RecurringPostResponse
    err = client.CreateRecurringPost(ctx, recurringReq, &recurringResp)
    require.NoError(t, err)
}
```

### Testing Requirements
**CRITICAL**: Do NOT use `t.Parallel()` in any tests

**Test Objectives:**
- Validate complete user workflows work end-to-end
- Test all operations work together without conflicts
- Verify error handling across all components
- Validate performance with realistic data sizes
- Ensure mock server handles complex scenarios
- Test all examples from documentation work correctly

**Test Requirements from CLAUDE.md:**
- Test MUST always be in the test package `package v1_test` and not `package v1`
- Test names should be in camelCase and start with a capital letter
- Use `github.com/stretchr/testify/require` and `github.com/stretchr/testify/assert`
- Use `require.ErrorContains(t, err, test.wantErr)` instead of `require.Contains(t, err.Error(), test.wantErr)`
- Use `require` for critical assertions, `assert` for non-critical ones
- No explanatory messages in assertions

### Performance Requirements
- Iterator should handle 1000+ posts efficiently
- Job polling should not create excessive load
- Bulk operations should handle 100+ posts
- Memory usage should remain reasonable with large datasets

### Validation Commands
After implementation, run these commands to verify:
- `go test ./v1 -run Integration`
- `go test ./v1 -bench=.`
- `go test ./v1 -race`
- `go test ./v1 -cover`
- `go test ./v1 -run Example`

### Success Criteria
This final phase is complete when:
1. Complete user workflows work end-to-end
2. All operations integrate correctly without conflicts
3. Error handling works consistently across components
4. Performance benchmarks meet requirements
5. Mock server handles complex scenarios reliably
6. All documentation examples execute correctly
7. Integration tests pass with race detection enabled
8. Code coverage exceeds 80% across all phases
9. Memory usage remains reasonable with large datasets
10. Client resilience tests pass for network/timeout scenarios

### Final Project Validation
After this phase, the complete Publer Go client library should:
- Export `publer.NewClient()` from `v1/` package
- Support all major Publer API operations
- Provide comprehensive error handling with rate limit support
- Include custom iterators for paginated resources
- Offer convenience methods for common use cases
- Handle async operations with job tracking
- Include comprehensive mock server for testing
- Have complete documentation and examples
- Pass all integration and performance tests