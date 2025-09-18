package v1_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "github.com/thrawn/publer.go/v1"
)

func TestBulkPublishPosts(t *testing.T) {
	server := v1.SpawnMockServer()
	defer server.Stop()

	client := server.Client()

	req := v1.BulkPublishPostsRequest{
		Posts: []v1.BulkPost{
			{
				Text:     "First bulk post",
				Accounts: []string{"account-1", "account-2"},
				Media: []v1.Media{
					{URL: "https://example.com/image1.jpg", Type: "image"},
				},
			},
			{
				Text:     "Second bulk post",
				Accounts: []string{"account-3"},
			},
		},
	}

	var resp v1.BulkPublishPostsResponse
	server.Reset()

	err := client.BulkPublishPosts(context.Background(), req, &resp)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.JobID)

	// Verify job status endpoint returns status for the created job
	var jobResp v1.GetJobStatusResponse
	err = client.GetJobStatus(context.Background(), v1.GetJobStatusRequest{JobID: resp.JobID}, &jobResp)
	require.NoError(t, err)
	assert.Equal(t, resp.JobID, jobResp.ID)
	assert.Equal(t, "pending", jobResp.Status)
	assert.Equal(t, 0, jobResp.Progress)
}

func TestBulkSchedulePosts(t *testing.T) {
	server := v1.SpawnMockServer()
	defer server.Stop()

	client := server.Client()

	futureTime1 := time.Now().Add(2 * time.Hour)
	futureTime2 := time.Now().Add(4 * time.Hour)

	req := v1.BulkSchedulePostsRequest{
		Posts: []v1.BulkPost{
			{
				Text:        "First scheduled post",
				Accounts:    []string{"account-1", "account-2"},
				ScheduledAt: futureTime1,
				Media: []v1.Media{
					{URL: "https://example.com/image1.jpg", Type: "image"},
				},
			},
			{
				Text:        "Second scheduled post",
				Accounts:    []string{"account-3"},
				ScheduledAt: futureTime2,
			},
		},
	}

	var resp v1.BulkSchedulePostsResponse
	server.Reset()

	err := client.BulkSchedulePosts(context.Background(), req, &resp)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.JobID)

	// Verify job status endpoint returns status for the created job
	var jobResp v1.GetJobStatusResponse
	err = client.GetJobStatus(context.Background(), v1.GetJobStatusRequest{JobID: resp.JobID}, &jobResp)
	require.NoError(t, err)
	assert.Equal(t, resp.JobID, jobResp.ID)
	assert.Equal(t, "pending", jobResp.Status)
	assert.Equal(t, 0, jobResp.Progress)
}

func TestBulkOperationLimits(t *testing.T) {
	server := v1.SpawnMockServer()
	defer server.Stop()

	client := server.Client()

	const bulkLimit = 3

	// Create more posts than the limit
	posts := make([]v1.BulkPost, bulkLimit+1)
	for i := range posts {
		posts[i] = v1.BulkPost{
			Text:     "Test post",
			Accounts: []string{"account-1"},
		}
	}

	for _, test := range []struct {
		name    string
		posts   []v1.BulkPost
		wantErr string
	}{
		{
			name:    "WithinLimit",
			posts:   posts[:bulkLimit],
			wantErr: "",
		},
		{
			name:    "ExceedsLimit",
			posts:   posts,
			wantErr: "Bulk operation limit exceeded",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			server.Reset()
			server.SetBulkOperationLimit(bulkLimit)

			req := v1.BulkPublishPostsRequest{Posts: test.posts}
			var resp v1.BulkPublishPostsResponse

			err := client.BulkPublishPosts(context.Background(), req, &resp)
			if test.wantErr == "" {
				require.NoError(t, err)
				assert.NotEmpty(t, resp.JobID)
			} else {
				require.Error(t, err)
				require.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}

func TestBulkPartialFailure(t *testing.T) {
	server := v1.SpawnMockServer()
	defer server.Stop()

	client := server.Client()

	const jobID = "test-bulk-job"
	server.Reset()

	// Set up job with partial success
	server.SetJobStatus(jobID, "completed", 100, &v1.JobResult{
		Success: false,
		PostIDs: []string{"post-1"},
		Error:   "Post 2 failed: invalid account",
		Data: map[string]interface{}{
			"successful_posts": 1,
			"failed_posts":     1,
			"errors": []interface{}{
				map[string]interface{}{
					"post_index": 1,
					"error":      "invalid account",
				},
			},
		},
	}, "")

	// Configure mock to return this job ID
	server.SetResponse("POST", "/api/v1/posts/schedule/publish", 200, v1.BulkPublishPostsResponse{
		JobID: jobID,
	})

	req := v1.BulkPublishPostsRequest{
		Posts: []v1.BulkPost{
			{
				Text:     "Post that succeeds",
				Accounts: []string{"valid-account"},
			},
			{
				Text:     "Post that fails",
				Accounts: []string{"invalid-account"},
			},
		},
	}

	var resp v1.BulkPublishPostsResponse
	err := client.BulkPublishPosts(context.Background(), req, &resp)
	require.NoError(t, err)
	assert.Equal(t, jobID, resp.JobID)

	// Check job result for partial failure
	var jobResp v1.GetJobStatusResponse
	err = client.GetJobStatus(context.Background(), v1.GetJobStatusRequest{JobID: resp.JobID}, &jobResp)
	require.NoError(t, err)
	assert.Equal(t, "completed", jobResp.Status)
	assert.False(t, jobResp.Result.Success)
	assert.Contains(t, jobResp.Result.Error, "Post 2 failed")
	assert.Len(t, jobResp.Result.PostIDs, 1)
}

func TestBulkSchedulePostsValidation(t *testing.T) {
	server := v1.SpawnMockServer()
	defer server.Stop()

	client := server.Client()

	for _, test := range []struct {
		name    string
		posts   []v1.BulkPost
		wantErr string
	}{
		{
			name: "ValidFutureTimes",
			posts: []v1.BulkPost{
				{
					Text:        "Future post 1",
					Accounts:    []string{"account-1"},
					ScheduledAt: time.Now().Add(time.Hour),
				},
				{
					Text:        "Future post 2",
					Accounts:    []string{"account-2"},
					ScheduledAt: time.Now().Add(2 * time.Hour),
				},
			},
			wantErr: "",
		},
		{
			name: "FirstPostPastTime",
			posts: []v1.BulkPost{
				{
					Text:        "Past post",
					Accounts:    []string{"account-1"},
					ScheduledAt: time.Now().Add(-time.Hour),
				},
				{
					Text:        "Future post",
					Accounts:    []string{"account-2"},
					ScheduledAt: time.Now().Add(time.Hour),
				},
			},
			wantErr: "Post 1: Scheduled time must be in the future",
		},
		{
			name: "SecondPostPastTime",
			posts: []v1.BulkPost{
				{
					Text:        "Future post",
					Accounts:    []string{"account-1"},
					ScheduledAt: time.Now().Add(time.Hour),
				},
				{
					Text:        "Past post",
					Accounts:    []string{"account-2"},
					ScheduledAt: time.Now().Add(-time.Hour),
				},
			},
			wantErr: "Post 2: Scheduled time must be in the future",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			server.Reset()

			req := v1.BulkSchedulePostsRequest{Posts: test.posts}
			var resp v1.BulkSchedulePostsResponse

			err := client.BulkSchedulePosts(context.Background(), req, &resp)
			if test.wantErr == "" {
				require.NoError(t, err)
				assert.NotEmpty(t, resp.JobID)
			} else {
				require.Error(t, err)
				require.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}