package v1_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "github.com/thrawn/publer.go/v1"
)

func TestListPostsIterator(t *testing.T) {
	server := v1.SpawnMockServer()
	defer server.Stop()

	client := server.Client()

	const totalPosts = 35
	posts := make([]v1.Post, totalPosts)
	for i := 0; i < totalPosts; i++ {
		posts[i] = v1.Post{
			ID:          fmt.Sprintf("post-%d", i+1),
			Text:        fmt.Sprintf("Post content %d", i+1),
			State:       "published",
			Type:        "regular",
			AccountID:   fmt.Sprintf("account-%d", (i%3)+1),
			Network:     "twitter",
			ScheduledAt: time.Now().Add(time.Hour * time.Duration(i)),
			HasMedia:    i%2 == 0,
			User: v1.User{
				ID:    fmt.Sprintf("user-%d", (i%2)+1),
				Name:  fmt.Sprintf("User %d", (i%2)+1),
				Email: fmt.Sprintf("user%d@example.com", (i%2)+1),
			},
		}
	}

	server.Reset()
	server.AddPosts(posts)

	// Test basic iteration through all pages
	iterator := client.ListPosts(context.Background(), v1.ListPostsRequest{})

	var collectedPosts []v1.Post
	pageCount := 0
	var page v1.Page[v1.Post]

	for {
		hasMore := iterator.Next(context.Background(), &page)
		require.NoError(t, iterator.Err())
		pageCount++

		// Validate page metadata
		assert.Equal(t, totalPosts, page.Total)
		assert.Equal(t, pageCount, page.Page)
		assert.Equal(t, 10, page.PerPage)
		assert.Equal(t, 4, page.TotalPages) // 35 posts / 10 per page = 4 pages

		// Validate page size
		expectedSize := 10
		if pageCount == 4 {
			expectedSize = 5 // Last page has 5 items
		}
		assert.Len(t, page.Items, expectedSize)

		collectedPosts = append(collectedPosts, page.Items...)

		if !hasMore {
			break
		}
	}

	assert.Equal(t, 4, pageCount)
	assert.Len(t, collectedPosts, totalPosts)

	// Verify posts are in correct order
	for i, post := range collectedPosts {
		assert.Equal(t, posts[i].ID, post.ID)
		assert.Equal(t, posts[i].Text, post.Text)
		assert.Equal(t, posts[i].State, post.State)
		assert.Equal(t, posts[i].AccountID, post.AccountID)
	}
}

func TestPostIteratorPagination(t *testing.T) {
	server := v1.SpawnMockServer()
	defer server.Stop()

	client := server.Client()

	// Test with exact page boundary (20 posts = 2 full pages)
	const totalPosts = 20
	posts := make([]v1.Post, totalPosts)
	for i := 0; i < totalPosts; i++ {
		posts[i] = v1.Post{
			ID:        fmt.Sprintf("iter-post-%d", i+1),
			Text:      fmt.Sprintf("Iterator test post %d", i+1),
			State:     "scheduled",
			AccountID: "test-account",
			Network:   "facebook",
		}
	}

	server.Reset()
	server.AddPosts(posts)

	iterator := client.ListPosts(context.Background(), v1.ListPostsRequest{})

	// Page 1
	var page1 v1.Page[v1.Post]
	hasMore := iterator.Next(context.Background(), &page1)
	require.NoError(t, iterator.Err())
	assert.True(t, hasMore)
	assert.Equal(t, 1, page1.Page)
	assert.Len(t, page1.Items, 10)
	assert.Equal(t, "iter-post-1", page1.Items[0].ID)
	assert.Equal(t, "iter-post-10", page1.Items[9].ID)

	// Page 2
	var page2 v1.Page[v1.Post]
	hasMore = iterator.Next(context.Background(), &page2)
	require.NoError(t, iterator.Err())
	assert.False(t, hasMore) // No more pages after this
	assert.Equal(t, 2, page2.Page)
	assert.Len(t, page2.Items, 10)
	assert.Equal(t, "iter-post-11", page2.Items[0].ID)
	assert.Equal(t, "iter-post-20", page2.Items[9].ID)

	// Verify no more pages
	var page3 v1.Page[v1.Post]
	hasMore = iterator.Next(context.Background(), &page3)
	require.NoError(t, iterator.Err())
	assert.False(t, hasMore)
	assert.Empty(t, page3.Items)
}

func TestPostIteratorError(t *testing.T) {
	server := v1.SpawnMockServer()
	defer server.Stop()

	client := server.Client()

	// Don't add any posts, which should trigger an error
	server.Reset()
	server.SetErrorResponse("GET", "/api/v1/posts", 0, 500, map[string]string{"error": "Internal Server Error"}, nil)

	iterator := client.ListPosts(context.Background(), v1.ListPostsRequest{})

	var page v1.Page[v1.Post]
	hasMore := iterator.Next(context.Background(), &page)
	assert.False(t, hasMore)

	err := iterator.Err()
	require.Error(t, err)
	require.ErrorContains(t, err, "Internal Server Error")
}

func TestPostIteratorContext(t *testing.T) {
	server := v1.SpawnMockServer()
	defer server.Stop()

	client := server.Client()

	// Add many posts to ensure multiple pages
	const totalPosts = 50
	posts := make([]v1.Post, totalPosts)
	for i := 0; i < totalPosts; i++ {
		posts[i] = v1.Post{
			ID:        fmt.Sprintf("ctx-post-%d", i+1),
			Text:      fmt.Sprintf("Context test post %d", i+1),
			State:     "draft",
			AccountID: "ctx-account",
			Network:   "linkedin",
		}
	}

	server.Reset()
	server.AddPosts(posts)

	// Create context that will be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	iterator := client.ListPosts(ctx, v1.ListPostsRequest{})

	// Get first page successfully
	var page1 v1.Page[v1.Post]
	hasMore := iterator.Next(ctx, &page1)
	require.NoError(t, iterator.Err())
	assert.True(t, hasMore)
	assert.Len(t, page1.Items, 10)

	// Cancel context
	cancel()

	// Next call should fail with context cancellation
	var page2 v1.Page[v1.Post]
	hasMore = iterator.Next(ctx, &page2)
	assert.False(t, hasMore)

	err := iterator.Err()
	require.Error(t, err)
	require.ErrorIs(t, err, context.Canceled)
}

func TestPostIteratorWithFilters(t *testing.T) {
	server := v1.SpawnMockServer()
	defer server.Stop()

	client := server.Client()

	// Create posts with different states and accounts
	posts := []v1.Post{
		{ID: "f1", Text: "Filter 1", State: "published", AccountID: "acc1", Network: "twitter"},
		{ID: "f2", Text: "Filter 2", State: "scheduled", AccountID: "acc2", Network: "facebook"},
		{ID: "f3", Text: "Filter 3", State: "published", AccountID: "acc1", Network: "twitter"},
		{ID: "f4", Text: "Filter 4", State: "draft", AccountID: "acc3", Network: "linkedin"},
		{ID: "f5", Text: "Filter 5", State: "published", AccountID: "acc2", Network: "facebook"},
		{ID: "f6", Text: "Filter 6", State: "scheduled", AccountID: "acc1", Network: "twitter"},
		{ID: "f7", Text: "Query match", State: "published", AccountID: "acc1", Network: "twitter"},
		{ID: "f8", Text: "Another post", State: "published", AccountID: "acc1", Network: "instagram"},
		{ID: "f9", Text: "Query match too", State: "scheduled", AccountID: "acc2", Network: "facebook"},
		{ID: "f10", Text: "Last post", State: "draft", AccountID: "acc3", Network: "linkedin"},
		{ID: "f11", Text: "Extra post", State: "published", AccountID: "acc1", Network: "twitter"},
		{ID: "f12", Text: "Final post", State: "scheduled", AccountID: "acc2", Network: "facebook"},
	}

	server.Reset()
	server.AddPosts(posts)

	// Test with state filter
	iterator := client.ListPosts(context.Background(), v1.ListPostsRequest{
		State: "published",
	})

	var page v1.Page[v1.Post]
	hasMore := iterator.Next(context.Background(), &page)
	require.NoError(t, iterator.Err())

	// Published posts: f1, f3, f5, f7, f8, f11 = 6 posts total, fits on one page
	assert.False(t, hasMore)
	assert.Len(t, page.Items, 6)

	// Test with multiple states
	iterator = client.ListPosts(context.Background(), v1.ListPostsRequest{
		States: []string{"published", "scheduled"},
	})

	hasMore = iterator.Next(context.Background(), &page)
	require.NoError(t, iterator.Err())
	// Published + scheduled posts: f1,f2,f3,f5,f6,f7,f8,f9,f11,f12 = 10 posts, fits on one page
	assert.False(t, hasMore)
	assert.Len(t, page.Items, 10)

	// Test with account IDs filter
	iterator = client.ListPosts(context.Background(), v1.ListPostsRequest{
		AccountIDs: []string{"acc1", "acc2"},
	})

	hasMore = iterator.Next(context.Background(), &page)
	require.NoError(t, iterator.Err())
	// Posts from acc1 and acc2: f1,f2,f3,f5,f6,f7,f8,f9,f11,f12 = 10 posts, fits on one page
	assert.False(t, hasMore)
	assert.Len(t, page.Items, 10)

	// Test with query filter
	iterator = client.ListPosts(context.Background(), v1.ListPostsRequest{
		Query: "Query match",
	})

	hasMore = iterator.Next(context.Background(), &page)
	require.NoError(t, iterator.Err())
	// Posts matching "Query match": f7, f9 = 2 posts
	assert.False(t, hasMore)
	assert.Len(t, page.Items, 2)

	// Test with date range filter
	now := time.Now()
	iterator = client.ListPosts(context.Background(), v1.ListPostsRequest{
		From: now.Add(-24 * time.Hour),
		To:   now.Add(24 * time.Hour),
	})

	hasMore = iterator.Next(context.Background(), &page)
	require.NoError(t, iterator.Err())
	// No posts have scheduled times set (zero time), so no matches
	assert.False(t, hasMore)
	assert.Len(t, page.Items, 0)

	// Test with post type filter
	iterator = client.ListPosts(context.Background(), v1.ListPostsRequest{
		PostType: "regular",
	})

	hasMore = iterator.Next(context.Background(), &page)
	require.NoError(t, iterator.Err())
	// No posts have Type set to "regular" (they're empty), so no matches
	assert.False(t, hasMore)
	assert.Len(t, page.Items, 0)

	// Test with member ID filter
	iterator = client.ListPosts(context.Background(), v1.ListPostsRequest{
		MemberID: "member123",
	})

	hasMore = iterator.Next(context.Background(), &page)
	require.NoError(t, iterator.Err())
	// No posts have User.ID set to "member123" (they're empty), so no matches
	assert.False(t, hasMore)
	assert.Len(t, page.Items, 0)

	// Test with combined filters
	iterator = client.ListPosts(context.Background(), v1.ListPostsRequest{
		State:      "published",
		AccountIDs: []string{"acc1"},
		Query:      "Filter",
		From:       now.Add(-48 * time.Hour),
		To:         now,
		PostType:   "regular",
		MemberID:   "member456",
	})

	hasMore = iterator.Next(context.Background(), &page)
	require.NoError(t, iterator.Err())
	// Combined filters will find no matches due to strict filtering
	assert.False(t, hasMore)
	assert.Len(t, page.Items, 0)
}

func TestPostIteratorLazyLoading(t *testing.T) {
	server := v1.SpawnMockServer()
	defer server.Stop()

	client := server.Client()

	// Setup posts
	posts := make([]v1.Post, 15)
	for i := 0; i < 15; i++ {
		posts[i] = v1.Post{
			ID:        fmt.Sprintf("lazy-%d", i+1),
			Text:      fmt.Sprintf("Lazy load test %d", i+1),
			State:     "published",
			AccountID: "lazy-account",
			Network:   "twitter",
		}
	}

	server.Reset()
	server.AddPosts(posts)

	// Create iterator but don't call Next yet
	iterator := client.ListPosts(context.Background(), v1.ListPostsRequest{})

	// Verify Err() returns nil before any calls
	require.NoError(t, iterator.Err())

	// First Next() call should initialize and fetch first page
	var page1 v1.Page[v1.Post]
	hasMore := iterator.Next(context.Background(), &page1)
	require.NoError(t, iterator.Err())
	assert.True(t, hasMore)
	assert.Equal(t, 1, page1.Page)
	assert.Len(t, page1.Items, 10)
	assert.Equal(t, 15, page1.Total)
	assert.Equal(t, 2, page1.TotalPages)

	// Second Next() call fetches second page
	var page2 v1.Page[v1.Post]
	hasMore = iterator.Next(context.Background(), &page2)
	require.NoError(t, iterator.Err())
	assert.False(t, hasMore)
	assert.Equal(t, 2, page2.Page)
	assert.Len(t, page2.Items, 5)

	// Further Next() calls return false
	var page3 v1.Page[v1.Post]
	hasMore = iterator.Next(context.Background(), &page3)
	require.NoError(t, iterator.Err())
	assert.False(t, hasMore)
	assert.Empty(t, page3.Items)
}

func TestPostIteratorEmptyResult(t *testing.T) {
	server := v1.SpawnMockServer()
	defer server.Stop()

	client := server.Client()

	// No posts added - empty result set
	server.Reset()

	iterator := client.ListPosts(context.Background(), v1.ListPostsRequest{
		State: "nonexistent",
	})

	var page v1.Page[v1.Post]
	hasMore := iterator.Next(context.Background(), &page)
	require.NoError(t, iterator.Err())
	assert.False(t, hasMore)
	assert.Empty(t, page.Items)
	assert.Equal(t, 0, page.Total)
}

func TestPostIteratorSinglePage(t *testing.T) {
	server := v1.SpawnMockServer()
	defer server.Stop()

	client := server.Client()

	// Add exactly 10 posts (one full page)
	posts := make([]v1.Post, 10)
	for i := 0; i < 10; i++ {
		posts[i] = v1.Post{
			ID:        fmt.Sprintf("single-%d", i+1),
			Text:      fmt.Sprintf("Single page post %d", i+1),
			State:     "published",
			AccountID: "single-account",
			Network:   "twitter",
		}
	}

	server.Reset()
	server.AddPosts(posts)

	iterator := client.ListPosts(context.Background(), v1.ListPostsRequest{})

	var page v1.Page[v1.Post]
	hasMore := iterator.Next(context.Background(), &page)
	require.NoError(t, iterator.Err())
	assert.False(t, hasMore) // Only one page, so no more
	assert.Len(t, page.Items, 10)
	assert.Equal(t, 1, page.Page)
	assert.Equal(t, 1, page.TotalPages)
	assert.Equal(t, 10, page.Total)

	// Next call should return false with empty page
	var page2 v1.Page[v1.Post]
	hasMore = iterator.Next(context.Background(), &page2)
	require.NoError(t, iterator.Err())
	assert.False(t, hasMore)
	assert.Empty(t, page2.Items)
}