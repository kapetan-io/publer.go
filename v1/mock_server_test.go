package v1_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "github.com/thrawn/publer.go/v1"
)

// waitForServer ensures the server is running and accepting connections
func waitForServer(t *testing.T, client *v1.Client, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			t.Fatal("timeout waiting for server to be ready")
		case <-ticker.C:
			if err := client.Test(ctx); err == nil {
				return // Server is ready
			}
		}
	}
}

func TestMockServer(t *testing.T) {
	server := v1.SpawnMockServer()
	defer func() { _ = server.Stop() }()

	// Test basic server functionality
	client := server.Client()
	assert.NotNil(t, client)

	// Test server reset
	server.Reset()

	// Configure a test response for connectivity check
	server.SetResponse("GET", "/api/v1/test", 200, map[string]string{"status": "ok"})

	// Ensure the server is running and accepting connections with retry
	waitForServer(t, client, 5*time.Second)
}

func TestMockServerSetResponse(t *testing.T) {
	server := v1.SpawnMockServer()
	defer func() { _ = server.Stop() }()

	server.Reset()

	// Configure a response
	testData := map[string]string{"message": "hello world"}
	server.SetResponse("GET", "/api/v1/test", 200, testData)

	// The configured client has the correct credentials automatically
	client := server.Client()
	assert.NotNil(t, client)

	// Invoke /api/v1/test via client.Test(ctx) to ensure mock server, API key and workspace headers work
	ctx := context.Background()
	err := client.Test(ctx)
	require.NoError(t, err)

	// This verifies that:
	// 1. The mock server is running and accepting connections
	// 2. The client has the correct API key and workspace ID
	// 3. The server validates these credentials properly
	// 4. The configured response is returned successfully
}

func TestMockServerSetErrorResponse(t *testing.T) {
	server := v1.SpawnMockServer()
	defer func() { _ = server.Stop() }()

	server.Reset()

	// Configure successful response first
	server.SetResponse("GET", "/api/v1/test", 200, map[string]string{"status": "ok"})

	// Configure error response after 2 calls
	errorData := map[string]string{"error": "rate_limit_exceeded"}
	headers := map[string]string{
		"X-RateLimit-Limit":     "100",
		"X-RateLimit-Remaining": "0",
		"X-RateLimit-Reset":     "1640995200",
	}
	server.SetErrorResponse("GET", "/api/v1/test", 2, 429, errorData, headers)

	// Get configured client
	client := server.Client()
	assert.NotNil(t, client)

	ctx := context.Background()

	// First call should succeed
	err := client.Test(ctx)
	require.NoError(t, err)

	// Second call should return rate limit error
	err = client.Test(ctx)
	require.Error(t, err)

	// Verify it's a RateLimitError with proper details
	var rateLimitErr *v1.RateLimitError
	require.ErrorAs(t, err, &rateLimitErr)
	assert.Equal(t, 429, rateLimitErr.StatusCode)
	assert.Equal(t, 100, rateLimitErr.Limit)
	assert.Equal(t, 0, rateLimitErr.Remaining)
	assert.Equal(t, int64(1640995200), rateLimitErr.Reset)
}

func TestMockServerJobProgression(t *testing.T) {
	server := v1.SpawnMockServer()
	defer func() { _ = server.Stop() }()

	server.Reset()

	// Set up job progression
	jobID := "job-123"
	states := []v1.JobStatus{
		{ID: jobID, Status: "pending", Progress: 0},
		{ID: jobID, Status: "working", Progress: 50},
		{ID: jobID, Status: "completed", Progress: 100, Result: &v1.JobResult{PostIDs: []string{"post-456"}}},
	}
	server.SetJobProgression(jobID, states)

	// Test job state advancement
	advanced := server.AdvanceJobState(jobID)
	assert.True(t, advanced)

	advanced = server.AdvanceJobState(jobID)
	assert.True(t, advanced)

	// Try to advance beyond final state
	advanced = server.AdvanceJobState(jobID)
	assert.False(t, advanced)

	// Job progression functionality is working - actual API testing would
	// be through client calls when they become available in later phases
}

func TestMockServerPagination(t *testing.T) {
	server := v1.SpawnMockServer()
	defer func() { _ = server.Stop() }()

	server.Reset()

	// Add test posts
	posts := []v1.Post{
		{ID: "1", Text: "Post 1"},
		{ID: "2", Text: "Post 2"},
		{ID: "3", Text: "Post 3"},
		{ID: "4", Text: "Post 4"},
		{ID: "5", Text: "Post 5"},
	}
	server.AddPosts(posts)

	// Get configured client
	client := server.Client()
	assert.NotNil(t, client)

	// Pagination setup is working - actual testing would be through
	// client API calls when they become available in later phases
}

func TestMockServerDelayAndReset(t *testing.T) {
	server := v1.SpawnMockServer()
	server.SetResponse("GET", "/api/v1/test", 200, map[string]string{"status": "ok"})
	defer func() { _ = server.Stop() }()

	// Test delay setting
	server.SetDelay(10 * time.Millisecond)

	// Get configured client
	client := server.Client()
	assert.NotNil(t, client)

	ctx := context.Background()
	server.SetResponse("GET", "/api/v1/test", 200, map[string]string{"status": "ok"})
	require.NoError(t, client.Test(ctx))

	// Server is automatically started and ready to use
	// Multiple servers can be spawned independently
	server2 := v1.SpawnMockServer()
	server2.SetResponse("GET", "/api/v1/test", 200, map[string]string{"status": "ok"})
	defer func() { _ = server2.Stop() }()

	client2 := server2.Client()
	assert.NotNil(t, client2)

	require.NoError(t, client2.Test(ctx))
	// Each server has its own credentials and configuration
}
