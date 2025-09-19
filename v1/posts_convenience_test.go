package v1_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "github.com/thrawn/publer.go/v1"
)

func TestGetPostsByState(t *testing.T) {
	server := v1.SpawnMockServer()
	defer server.Stop()

	client := server.Client()

	server.Reset()
	posts := []v1.Post{
		{ID: "1", Text: "Draft post", State: "draft", AccountID: "acc-1"},
		{ID: "2", Text: "Published post", State: "published", AccountID: "acc-2"},
		{ID: "3", Text: "Scheduled post", State: "scheduled", AccountID: "acc-3"},
	}
	server.AddPosts(posts)

	iter := client.GetPostsByState("draft")
	require.NotNil(t, iter)

	var page v1.Page[v1.Post]
	hasMore := iter.Next(context.Background(), &page)
	require.NoError(t, iter.Err())

	require.Len(t, page.Items, 1)
	assert.Equal(t, "1", page.Items[0].ID)
	assert.Equal(t, "draft", page.Items[0].State)
	assert.False(t, hasMore)
}

func TestGetPostsByDateRange(t *testing.T) {
	server := v1.SpawnMockServer()
	defer server.Stop()

	client := server.Client()

	server.Reset()
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	tomorrow := now.Add(24 * time.Hour)

	posts := []v1.Post{
		{ID: "1", Text: "Past post", State: "published", ScheduledAt: yesterday},
		{ID: "2", Text: "Current post", State: "scheduled", ScheduledAt: now},
		{ID: "3", Text: "Future post", State: "scheduled", ScheduledAt: tomorrow},
	}
	server.AddPosts(posts)

	from := now.Add(-1 * time.Hour)
	to := now.Add(1 * time.Hour)
	iter := client.GetPostsByDateRange(from, to)
	require.NotNil(t, iter)

	var page v1.Page[v1.Post]
	hasMore := iter.Next(context.Background(), &page)
	require.NoError(t, iter.Err())

	require.Len(t, page.Items, 1)
	assert.Equal(t, "2", page.Items[0].ID)
	assert.False(t, hasMore)
}

func TestGetPostsByAccount(t *testing.T) {
	server := v1.SpawnMockServer()
	defer server.Stop()

	client := server.Client()

	server.Reset()
	posts := []v1.Post{
		{ID: "1", Text: "Post 1", State: "published", AccountID: "account-1"},
		{ID: "2", Text: "Post 2", State: "published", AccountID: "account-2"},
		{ID: "3", Text: "Post 3", State: "published", AccountID: "account-1"},
	}
	server.AddPosts(posts)

	iter := client.GetPostsByAccount("account-1")
	require.NotNil(t, iter)

	var page v1.Page[v1.Post]
	hasMore := iter.Next(context.Background(), &page)
	require.NoError(t, iter.Err())

	require.Len(t, page.Items, 2)
	assert.Equal(t, "1", page.Items[0].ID)
	assert.Equal(t, "3", page.Items[1].ID)
	assert.Equal(t, "account-1", page.Items[0].AccountID)
	assert.Equal(t, "account-1", page.Items[1].AccountID)
	assert.False(t, hasMore)
}

func TestGetPostsByQuery(t *testing.T) {
	server := v1.SpawnMockServer()
	defer server.Stop()

	client := server.Client()

	server.Reset()
	posts := []v1.Post{
		{ID: "1", Text: "Hello world", State: "published"},
		{ID: "2", Text: "Goodbye universe", State: "published"},
		{ID: "3", Text: "Hello everyone", State: "published"},
	}
	server.AddPosts(posts)

	iter := client.GetPostsByQuery("Hello")
	require.NotNil(t, iter)

	var page v1.Page[v1.Post]
	hasMore := iter.Next(context.Background(), &page)
	require.NoError(t, iter.Err())

	require.Len(t, page.Items, 2)
	assert.Equal(t, "1", page.Items[0].ID)
	assert.Equal(t, "3", page.Items[1].ID)
	assert.False(t, hasMore)
}