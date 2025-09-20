# publer.go

Go client library for Publer.com API

[![Go Version](https://img.shields.io/github/go-mod/go-version/kapetan-io/publer.go)](https://golang.org/dl/)
[![CI Status](https://github.com/kapetan-io/publer.go/workflows/CI/badge.svg)](https://github.com/kapetan-io/publer.go/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/kapetan-io/publer.go)](https://goreportcard.com/report/github.com/kapetan-io/publer.go)


publer.go is designed for Go developers who need to integrate with Publer.com's social media management platform. This
library emphasizes:

- **Zero external dependencies** (except test libraries)
- **Context support** throughout all operations
- **Rate limit awareness** with proper handling and backoff
- **Iterator based pagination** through an elegant iterator pattern
- **Async job tracking** for post creation and publishing operations

Perfect for building social media management tools, automation scripts, or integrating Publer.com into larger
applications without complex dependency reviews.

## Installation

```bash
go get github.com/kapetan-io/publer.go/v1
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/kapetan-io/publer.go/v1"
)

func main() {
    ctx := context.Background()

    // Create the client
    c, err := publer.NewClient(publer.Config{
        APIKey:      "your-api-key-here",
        WorkspaceID: "your-workspace-id",
    })

    if err != nil {
        log.Fatal(err)
    }

    var resp publer.PublishResponse
    req := publer.PublishRequest{
        Accounts: []string{"account-id-1", "account-id-2"}
        Text:     "Hello from publer.go! 🚀",
    }

    // Publish a post immediately
    if err := c.Publish(ctx, req, &resp); err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Job ID: %s\n", resp.JobID)

    // Wait for job completion
    var result publer.JobResult
    if err = c.WaitForJob(ctx, publer.WaitOptions{JobID: resp.JobID}, &result); err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Published posts: %v\n", result.PostIDs)
}
```

## Core Features

### Custom Iterator Pattern for Pagination

Navigate through paginated results with ease using our custom iterator pattern:

```go
// List all posts with automatic pagination
iter := client.ListPosts(ctx, publer.ListPostsRequest{
    Limit: 50, // Optional: control page size
})

var page publer.Page[publer.Post]
for iter.Next(ctx, &page) {
    for _, post := range page.Items {
        fmt.Printf("Post: %s (State: %s)\n", post.Text, post.State)
    }
}

if err := iter.Err(); err != nil {
    log.Fatal(err)
}

fmt.Printf("Total posts: %d\n", page.Total)
```

### Rate Limit Handling

The client automatically handles Publer.com's rate limits (100 requests per 2 minutes) with intelligent backoff:

```go
// The client will automatically retry with exponential backoff
// when rate limits are encountered
for i := 0; i < 200; i++ {
    var resp publer.PublishResponse
    err := client.PublishPost(ctx, publer.PublishRequest{
        Text:     fmt.Sprintf("Post %d", i+1),
        Accounts: []string{"account-1"},
    }, &resp)
    if err != nil {
        // Check if it's a rate limit error
        if rateLimitErr, ok := err.(*publer.RateLimitError); ok {
            fmt.Printf("Rate limited. Retry after: %v\n", rateLimitErr.RetryAfter)
            // Client handles this automatically on next request
        }
        continue
    }

    fmt.Printf("Created post %d with job ID: %s\n", i+1, resp.JobID)
}
```

### Context Support

All operations support Go's context package for cancellation and timeouts:

```go
// Create a context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// Create an iterator
iter := client.ListPosts(ctx, publer.ListPostsRequest{})
var page publer.Page[publer.Post]

// Iterate
if iter.Next(ctx, &page) {
    fmt.Printf("Got %d posts\n", len(page.Items))
} else if err := iter.Err(); err != nil {
    if ctx.Err() == context.DeadlineExceeded {
        fmt.Println("Operation timed out")
    }
    return
}
```

### Async Job Tracking

Monitor long-running operations like post creation:

```go
// Publish a post (returns immediately with job ID)
var resp publer.PublishPostResponse
err := client.PublishPost(ctx, publer.PublishPostRequest{
    Text:     "Async post creation",
    Accounts: []string{"account-1"},
}, &resp)
if err != nil {
    log.Fatal(err)
}

// Poll job status manually
for {
    var status publer.JobStatus
    err := client.GetJobStatus(ctx, publer.GetJobStatusRequest{JobID: resp.JobID}, &status)
    if err != nil {
        log.Fatal(err)
    }

    switch status.Status {
    case "completed":
        fmt.Printf("Post created! IDs: %v\n", status.Result.PostIDs)
        return
    case "failed":
        fmt.Printf("Job failed: %s\n", status.Result.Message)
        return
    case "processing":
        fmt.Println("Still processing...")
        time.Sleep(2 * time.Second)
    }
}

// Or use the convenience method with timeout
var result publer.JobResult
err = client.WaitForJob(ctx, publer.WaitOptions{JobID: resp.JobID}, &result)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Posts created: %v\n", result.PostIDs)
```

## API Operations

### Posts

#### Basic Operations
```go
// List posts with filtering
iter := client.ListPosts(ctx, publer.ListPostsRequest{
    State: "scheduled",
    Limit: 25,
})
var page publer.Page[publer.Post]
iter.Next(ctx, &page)

// Get a specific post
var post publer.GetPostResponse
err := client.GetPost(ctx, publer.GetPostRequest{ID: "post-id"}, &post)

// Delete a post
var deleteResp publer.DeletePostResponse
err = client.DeletePost(ctx, publer.DeletePostRequest{ID: "post-id"}, &deleteResp)
```

#### Scheduling
```go
// Schedule a post for later
var resp publer.SchedulePostResponse
err := client.SchedulePost(ctx, publer.SchedulePostRequest{
    Text:        "Scheduled post",
    Accounts:    []string{"account-id"},
    ScheduledAt: time.Now().Add(2 * time.Hour),
}, &resp)
```

### Bulk Operations

Handle multiple posts efficiently:

```go
// Schedule multiple posts at once
var resp publer.BulkSchedulePostsResponse
err := client.BulkSchedulePosts(ctx, publer.BulkSchedulePostsRequest{
    Posts: []publer.BulkPost{
        {Text: "Post 1", Accounts: accounts, ScheduledAt: time.Now().Add(time.Hour)},
        {Text: "Post 2", Accounts: accounts, ScheduledAt: time.Now().Add(2 * time.Hour)},
        {Text: "Post 3", Accounts: accounts, ScheduledAt: time.Now().Add(3 * time.Hour)},
    },
}, &resp)
if err != nil {
    log.Fatal(err)
}

// Track the bulk job
var result publer.JobResult
err = client.WaitForJob(ctx, publer.WaitOptions{JobID: resp.JobID}, &result)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Created posts: %v\n", result.PostIDs)
```

### Advanced Features

#### Recurring Posts
```go
var resp publer.CreateRecurringPostResponse
err := client.CreateRecurringPost(ctx, publer.CreateRecurringPostRequest{
    Text:      "Daily motivation! 💪",
    Accounts:  accounts,
    Frequency: "daily",
    EndDate:   time.Now().AddDate(0, 1, 0), // Run for 1 month
}, &resp)
```

#### Auto-Scheduling
```go
var resp publer.AutoSchedulePostResponse
err := client.AutoSchedulePost(ctx, publer.AutoScheduleRequest{
    Text:     "Auto-scheduled post",
    Accounts: accounts,
}, &resp)
```

### Users and Workspaces

```go
// Get current user info
var user publer.GetMeResponse
err := client.GetMe(ctx, &user)

// List all workspaces
iter := client.ListWorkspaces(ctx, publer.ListWorkspacesRequest{})
var workspacePage publer.Page[publer.Workspace]
iter.Next(ctx, &workspacePage)

// List social media accounts in workspace
accountIter := client.ListAccounts(ctx, publer.ListAccountsRequest{
    WorkspaceID: "workspace-id",
})
var accountPage publer.Page[publer.Account]
accountIter.Next(ctx, &accountPage)
```

## Error Handling

The library provides detailed error types for different scenarios:

```go
var resp publer.PublishPostResponse
err := client.PublishPost(ctx, publer.PublishPostRequest{
    Text:     "Test post",
    Accounts: []string{"account-1"},
}, &resp)
if err != nil {
    switch e := err.(type) {
    case *publer.APIError:
        fmt.Printf("API error: %s (Status: %d)\n", e.Message, e.StatusCode)
        if e.StatusCode == 422 {
            fmt.Println("Validation failed")
        }
    case *publer.RateLimitError:
        fmt.Printf("Rate limited. Retry after: %v\n", e.RetryAfter)
    default:
        fmt.Printf("Unexpected error: %v\n", err)
    }
}
```

## Project Status

⚠️ **Early Testing Phase**: This library is currently in early testing and development. While the core functionality is
implemented and tested, the API may still evolve based on feedback and real-world usage.

## Development

This project was written entirely by [Claude Code AI](https://claude.ai/code), under the direction
of [Derrick Wippler](https://wippler.dev) following the detailed implementation
plans found in the [`plans/`](./plans/) directory. The plans document the complete development process, architecture
decisions, and feature implementation strategy.

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines on how to get involved.

## License

This project is licensed under the MIT License - see the [LICENSE](./LICENSE) file for details.

## Links

- [Publer.com](https://publer.com) - Social media management platform
- [API Documentation](https://app.publer.com/api/v1/docs) - Official Publer API docs
- [Go Documentation](https://pkg.go.dev/github.com/kapetan-io/publer.go/v1) - Package documentation