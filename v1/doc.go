// Package publer provides a Go client library for the Publer.com API v1.
//
// The client supports all major Publer API operations including post management,
// scheduling, bulk operations, and workspace management.
//
// Basic usage:
//
//	import v1 "github.com/thrawn/publer.go/v1"
//
//	config := v1.Config{
//	    APIKey:      "your-api-key",
//	    WorkspaceID: "your-workspace-id",
//	}
//	client, err := v1.NewClient(config)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// List posts
//	iter := client.ListPosts(ctx, v1.ListPostsRequest{})
//	for {
//		var page v1.Page[v1.Post]
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
//	if rateLimitErr, ok := err.(*v1.RateLimitError); ok {
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
//	server := v1.SpawnMockServer()
//	defer server.Stop()
//
//	client := server.Client()
package v1