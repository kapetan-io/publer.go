package v1_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "github.com/thrawn/publer.go/v1"
)

func TestGetPost(t *testing.T) {
	server := v1.SpawnMockServer()
	defer server.Stop()

	client := server.Client()

	expectedPost := v1.Post{
		ScheduledAt: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
		PostLink:    "https://twitter.com/user/status/123",
		URL:         "https://example.com/post",
		Text:        "Test post content",
		AccountID:   "account-1",
		State:       "published",
		Network:     "twitter",
		Type:        "post",
		ID:          "post-123",
		HasMedia:    false,
		User: v1.User{
			Picture:   "https://example.com/avatar.jpg",
			Email:     "test@example.com",
			FirstName: "Test",
			Name:      "Test User",
			ID:        "user-1",
		},
	}

	server.Reset()
	server.AddPosts([]v1.Post{expectedPost})

	var resp v1.GetPostResponse
	err := client.GetPost(context.Background(), v1.GetPostRequest{
		PostID: "post-123",
	}, &resp)

	require.NoError(t, err)
	assert.Equal(t, expectedPost.ID, resp.ID)
	assert.Equal(t, expectedPost.Text, resp.Text)
	assert.Equal(t, expectedPost.State, resp.State)
	assert.Equal(t, expectedPost.AccountID, resp.AccountID)
	assert.Equal(t, expectedPost.Network, resp.Network)
}

func TestUpdatePost(t *testing.T) {
	server := v1.SpawnMockServer()
	defer server.Stop()

	client := server.Client()

	originalPost := v1.Post{
		ID:          "post-456",
		Text:        "Original text",
		State:       "draft",
		AccountID:   "account-1",
		Network:     "facebook",
		HasMedia:    false,
		ScheduledAt: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
	}

	server.Reset()
	server.AddPosts([]v1.Post{originalPost})

	var resp v1.UpdatePostResponse
	err := client.UpdatePost(context.Background(), v1.UpdatePostRequest{
		PostID:      "post-456",
		Text:        "Updated text content",
		ScheduledAt: time.Date(2024, 2, 20, 15, 30, 0, 0, time.UTC),
		Media: []v1.Media{
			{URL: "https://example.com/image.jpg", Type: "image"},
		},
	}, &resp)

	require.NoError(t, err)
	assert.Equal(t, "post-456", resp.ID)
	assert.Equal(t, "Updated text content", resp.Text)
	assert.Equal(t, time.Date(2024, 2, 20, 15, 30, 0, 0, time.UTC), resp.ScheduledAt)
	assert.True(t, resp.HasMedia)
}

func TestDeletePost(t *testing.T) {
	server := v1.SpawnMockServer()
	defer server.Stop()

	client := server.Client()

	postToDelete := v1.Post{
		ID:        "post-789",
		Text:      "Post to be deleted",
		State:     "scheduled",
		AccountID: "account-1",
		Network:   "twitter",
	}

	server.Reset()
	server.AddPosts([]v1.Post{postToDelete})

	var resp v1.DeletePostResponse
	err := client.DeletePost(context.Background(), v1.DeletePostRequest{
		PostID: "post-789",
	}, &resp)

	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "Post deleted successfully", resp.Message)

	// Verify post is actually deleted by trying to get it
	var getResp v1.GetPostResponse
	err = client.GetPost(context.Background(), v1.GetPostRequest{
		PostID: "post-789",
	}, &getResp)

	require.Error(t, err)
}

func TestPostNotFound(t *testing.T) {
	server := v1.SpawnMockServer()
	defer server.Stop()

	client := server.Client()

	for _, test := range []struct {
		name      string
		operation string
		testFunc  func() error
	}{
		{
			name:      "GetPostNotFound",
			operation: "get",
			testFunc: func() error {
				var resp v1.GetPostResponse
				return client.GetPost(context.Background(), v1.GetPostRequest{
					PostID: "nonexistent-post",
				}, &resp)
			},
		},
		{
			name:      "UpdatePostNotFound",
			operation: "update",
			testFunc: func() error {
				var resp v1.UpdatePostResponse
				return client.UpdatePost(context.Background(), v1.UpdatePostRequest{
					PostID: "nonexistent-post",
					Text:   "New text",
				}, &resp)
			},
		},
		{
			name:      "DeletePostNotFound",
			operation: "delete",
			testFunc: func() error {
				var resp v1.DeletePostResponse
				return client.DeletePost(context.Background(), v1.DeletePostRequest{
					PostID: "nonexistent-post",
				}, &resp)
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			server.Reset()

			err := test.testFunc()
			require.Error(t, err)
			require.ErrorContains(t, err, "Post not found")
		})
	}
}

func TestUpdatePostPartialUpdates(t *testing.T) {
	server := v1.SpawnMockServer()
	defer server.Stop()

	client := server.Client()

	originalPost := v1.Post{
		ID:          "post-partial",
		Text:        "Original text",
		State:       "draft",
		AccountID:   "account-1",
		Network:     "twitter",
		HasMedia:    false,
		ScheduledAt: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
	}

	server.Reset()
	server.AddPosts([]v1.Post{originalPost})

	// Test updating only text
	var resp1 v1.UpdatePostResponse
	err := client.UpdatePost(context.Background(), v1.UpdatePostRequest{
		PostID: "post-partial",
		Text:   "Only text updated",
	}, &resp1)

	require.NoError(t, err)
	assert.Equal(t, "Only text updated", resp1.Text)
	assert.Equal(t, originalPost.ScheduledAt, resp1.ScheduledAt)
	assert.Equal(t, originalPost.HasMedia, resp1.HasMedia)

	// Test updating only scheduled time
	var resp2 v1.UpdatePostResponse
	err = client.UpdatePost(context.Background(), v1.UpdatePostRequest{
		PostID:      "post-partial",
		ScheduledAt: time.Date(2024, 3, 25, 14, 45, 0, 0, time.UTC),
	}, &resp2)

	require.NoError(t, err)
	assert.Equal(t, "Only text updated", resp2.Text) // Should keep previous update
	assert.Equal(t, time.Date(2024, 3, 25, 14, 45, 0, 0, time.UTC), resp2.ScheduledAt)
	assert.Equal(t, originalPost.HasMedia, resp2.HasMedia)

	// Test updating only media
	var resp3 v1.UpdatePostResponse
	err = client.UpdatePost(context.Background(), v1.UpdatePostRequest{
		PostID: "post-partial",
		Media: []v1.Media{
			{URL: "https://example.com/video.mp4", Type: "video"},
		},
	}, &resp3)

	require.NoError(t, err)
	assert.Equal(t, "Only text updated", resp3.Text) // Should keep previous updates
	assert.Equal(t, time.Date(2024, 3, 25, 14, 45, 0, 0, time.UTC), resp3.ScheduledAt)
	assert.True(t, resp3.HasMedia) // Should be updated to true
}

func TestPostManagementWithDifferentStates(t *testing.T) {
	server := v1.SpawnMockServer()
	defer server.Stop()

	client := server.Client()

	for _, test := range []struct {
		name  string
		state string
	}{
		{"PublishedPost", "published"},
		{"DraftPost", "draft"},
		{"ScheduledPost", "scheduled"},
		{"FailedPost", "failed"},
	} {
		t.Run(test.name, func(t *testing.T) {
			postID := "post-" + test.state
			testPost := v1.Post{
				ID:        postID,
				Text:      "Test post with " + test.state + " state",
				State:     test.state,
				AccountID: "account-1",
				Network:   "twitter",
			}

			server.Reset()
			server.AddPosts([]v1.Post{testPost})

			// Test get operation
			var getResp v1.GetPostResponse
			err := client.GetPost(context.Background(), v1.GetPostRequest{
				PostID: postID,
			}, &getResp)
			require.NoError(t, err)
			assert.Equal(t, test.state, getResp.State)

			// Test update operation
			var updateResp v1.UpdatePostResponse
			err = client.UpdatePost(context.Background(), v1.UpdatePostRequest{
				PostID: postID,
				Text:   "Updated " + test.state + " post",
			}, &updateResp)
			require.NoError(t, err)
			assert.Equal(t, "Updated "+test.state+" post", updateResp.Text)

			// Test delete operation
			var deleteResp v1.DeletePostResponse
			err = client.DeletePost(context.Background(), v1.DeletePostRequest{
				PostID: postID,
			}, &deleteResp)
			require.NoError(t, err)
			assert.True(t, deleteResp.Success)
		})
	}
}

func TestPostIDValidation(t *testing.T) {
	server := v1.SpawnMockServer()
	defer server.Stop()

	client := server.Client()

	for _, test := range []struct {
		name   string
		postID string
	}{
		{"EmptyPostID", ""},
		{"PathTraversalDots", "../admin"},
		{"PathTraversalSlash", "post/../../admin"},
		{"InvalidCharacters", "post@#$%"},
		{"BackslashCharacter", "post\\admin"},
	} {
		t.Run(test.name, func(t *testing.T) {
			server.Reset()

			var resp v1.GetPostResponse
			err := client.GetPost(context.Background(), v1.GetPostRequest{
				PostID: test.postID,
			}, &resp)

			require.Error(t, err)
			require.ErrorContains(t, err, "invalid post ID")
		})
	}
}