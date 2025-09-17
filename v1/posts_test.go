package v1_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "github.com/thrawn/publer.go/v1"
)

func TestListPosts(t *testing.T) {
	server := v1.SpawnMockServer()
	defer server.Stop()

	client := server.Client()

	posts := []v1.Post{
		{
			ID:        "post-1",
			Text:      "First post",
			State:     "published",
			AccountID: "account-1",
			Network:   "twitter",
		},
		{
			ID:        "post-2",
			Text:      "Second post",
			State:     "scheduled",
			AccountID: "account-2",
			Network:   "facebook",
		},
	}

	server.Reset()
	server.AddPosts(posts)

	iterator := client.ListPosts(context.Background(), v1.ListPostsRequest{})

	var page v1.Page[v1.Post]
	hasMore := iterator.Next(context.Background(), &page)
	require.NoError(t, iterator.Err())

	// Validate response structure matches API specification
	assert.Equal(t, 2, page.Total)
	assert.Equal(t, 1, page.Page)
	assert.Equal(t, 10, page.PerPage)
	assert.Equal(t, 1, page.TotalPages)
	assert.Len(t, page.Items, 2)

	// Validate post data
	assert.Equal(t, "post-1", page.Items[0].ID)
	assert.Equal(t, "First post", page.Items[0].Text)
	assert.Equal(t, "post-2", page.Items[1].ID)
	assert.Equal(t, "Second post", page.Items[1].Text)
	assert.False(t, hasMore)
}


func TestGetJobStatus(t *testing.T) {
	server := v1.SpawnMockServer()
	defer server.Stop()

	client := server.Client()

	const jobID = "test-job-123"
	server.Reset()
	server.SetJobStatus(jobID, "completed", 100, &v1.JobResult{
		Success: true,
		Data:    map[string]interface{}{"post_id": "12345"},
	}, "")

	req := v1.GetJobStatusRequest{JobID: jobID}
	var resp v1.GetJobStatusResponse

	err := client.GetJobStatus(context.Background(), req, &resp)
	require.NoError(t, err)
	assert.Equal(t, jobID, resp.ID)
	assert.Equal(t, "completed", resp.Status)
	assert.Equal(t, 100, resp.Progress)
	assert.True(t, resp.Result.Success)
}

func TestWaitForJob(t *testing.T) {
	server := v1.SpawnMockServer()
	defer server.Stop()

	client := server.Client()

	const jobID = "test-job-wait"
	server.Reset()

	// Set final completed status directly
	server.SetJobStatus(jobID, "completed", 100, &v1.JobResult{Success: true}, "")

	opts := v1.WaitOptions{
		JobID:        jobID,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Jitter:       5 * time.Millisecond,
	}

	var result v1.JobResult
	err := client.WaitForJob(context.Background(), opts, &result)
	require.NoError(t, err)
	assert.True(t, result.Success)
}

func TestWaitForJobFailed(t *testing.T) {
	server := v1.SpawnMockServer()
	defer server.Stop()

	client := server.Client()

	const jobID = "test-job-failed"
	server.Reset()
	server.SetJobStatus(jobID, "failed", 0, &v1.JobResult{
		Success: false,
		Error:   "Processing failed",
	}, "Processing failed")

	opts := v1.WaitOptions{
		JobID:        jobID,
		InitialDelay: 10 * time.Millisecond,
	}

	var result v1.JobResult
	err := client.WaitForJob(context.Background(), opts, &result)
	require.Error(t, err)
	require.ErrorContains(t, err, "failed")
	assert.False(t, result.Success)
}

func TestWaitForJobTimeout(t *testing.T) {
	server := v1.SpawnMockServer()
	defer server.Stop()

	client := server.Client()

	const jobID = "test-job-timeout"
	server.Reset()
	server.SetJobStatus(jobID, "working", 50, nil, "")

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	opts := v1.WaitOptions{
		JobID:        jobID,
		InitialDelay: 100 * time.Millisecond,
	}

	var result v1.JobResult
	err := client.WaitForJob(ctx, opts, &result)
	require.Error(t, err)
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestPublishPost(t *testing.T) {
	server := v1.SpawnMockServer()
	defer server.Stop()

	client := server.Client()

	req := v1.PublishPostRequest{
		Text:     "Test post content",
		Accounts: []string{"account-1", "account-2"},
	}

	var resp v1.PublishPostResponse
	server.Reset()

	err := client.PublishPost(context.Background(), req, &resp)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.JobID)

	// Verify job status endpoint returns status for the created job
	jobReq := v1.GetJobStatusRequest{JobID: resp.JobID}
	var jobResp v1.GetJobStatusResponse
	err = client.GetJobStatus(context.Background(), jobReq, &jobResp)
	require.NoError(t, err)
	assert.Equal(t, resp.JobID, jobResp.ID)
	assert.Equal(t, "pending", jobResp.Status)
	assert.Equal(t, 0, jobResp.Progress)
}
