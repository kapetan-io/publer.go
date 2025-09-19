package v1_test

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	v1 "github.com/thrawn/publer.go/v1"
)

func ExampleClient_ListPosts() {
	// Create mock server for example
	server := v1.SpawnMockServer()
	defer func() { _ = server.Stop() }()

	client := server.Client()

	// Configure mock data
	server.Reset()
	server.AddPosts([]v1.Post{
		{ID: "1", Text: "Example post", State: "published"},
		{ID: "2", Text: "Another post", State: "scheduled"},
	})

	ctx := context.Background()
	iter := client.ListPosts(ctx, v1.ListPostsRequest{})

	var page v1.Page[v1.Post]
	iter.Next(ctx, &page)

	if err := iter.Err(); err != nil {
		log.Printf("Error: %v", err)
		return
	}

	for _, post := range page.Items {
		fmt.Printf("Post: %s - %s\n", post.ID, post.Text)
	}

	// Output:
	// Post: 1 - Example post
	// Post: 2 - Another post
}

func ExampleClient_PublishPost() {
	server := v1.SpawnMockServer()
	defer func() { _ = server.Stop() }()

	client := server.Client()

	// Configure mock response and job completion
	server.Reset()
	server.SetResponse("POST", "/api/v1/posts/schedule/publish", 200, v1.PublishPostResponse{
		JobID: "job-123",
	})
	server.SetJobStatus("job-123", "completed", 100, &v1.JobResult{
		PostIDs: []string{"post-456"},
	}, "")

	ctx := context.Background()
	req := v1.PublishPostRequest{
		Text:     "Hello, world!",
		Accounts: []string{"account-1"},
	}

	var resp v1.PublishPostResponse
	err := client.PublishPost(ctx, req, &resp)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Job ID: %s\n", resp.JobID)

	// Wait for job completion
	var result v1.JobResult
	opts := v1.WaitOptions{JobID: resp.JobID}
	err = client.WaitForJob(ctx, opts, &result)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Published posts: %v\n", result.PostIDs)

	// Output:
	// Job ID: job-123
	// Published posts: [post-456]
}

func ExampleClient_GetPostsByState() {
	server := v1.SpawnMockServer()
	defer func() { _ = server.Stop() }()

	client := server.Client()

	server.Reset()
	server.AddPosts([]v1.Post{
		{ID: "1", Text: "Draft post", State: "draft"},
		{ID: "2", Text: "Published post", State: "published"},
	})

	ctx := context.Background()
	iter := client.GetPostsByState("draft")

	var page v1.Page[v1.Post]
	iter.Next(ctx, &page)

	if err := iter.Err(); err != nil {
		log.Printf("Error: %v", err)
		return
	}

	for _, post := range page.Items {
		fmt.Printf("Draft: %s\n", post.Text)
	}

	// Output:
	// Draft: Draft post
}

func ExampleClient_WaitForJob() {
	server := v1.SpawnMockServer()
	defer func() { _ = server.Stop() }()

	client := server.Client()

	// Configure job completion immediately
	server.Reset()
	server.SetJobStatus("job-123", "completed", 100, &v1.JobResult{
		PostIDs: []string{"post-456"},
		Message: "Post published successfully",
	}, "")

	ctx := context.Background()
	var result v1.JobResult
	opts := v1.WaitOptions{JobID: "job-123"}
	err := client.WaitForJob(ctx, opts, &result)
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
	server := v1.SpawnMockServer()
	defer func() { _ = server.Stop() }()

	client := server.Client()

	server.Reset()
	server.SetResponse("POST", "/api/v1/posts/schedule", 200, v1.BulkSchedulePostsResponse{
		JobID: "bulk-job-123",
	})
	server.SetJobStatus("bulk-job-123", "completed", 100, &v1.JobResult{
		PostIDs: []string{"post-1", "post-2"},
	}, "")

	ctx := context.Background()
	req := v1.BulkSchedulePostsRequest{
		Posts: []v1.BulkPost{
			{
				ScheduledAt: time.Now().Add(time.Hour),
				Accounts:    []string{"account-1"},
				Text:        "First post",
			},
			{
				ScheduledAt: time.Now().Add(2 * time.Hour),
				Accounts:    []string{"account-1"},
				Text:        "Second post",
			},
		},
	}

	var resp v1.BulkSchedulePostsResponse
	err := client.BulkSchedulePosts(ctx, req, &resp)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Bulk job ID: %s\n", resp.JobID)

	// Output:
	// Bulk job ID: bulk-job-123
}

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