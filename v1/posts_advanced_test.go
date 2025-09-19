package v1_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "github.com/thrawn/publer.go/v1"
)

func TestCreateRecurringPost(t *testing.T) {
	server := v1.SpawnMockServer()
	defer func() { _ = server.Stop() }()

	client := server.Client()

	req := v1.RecurringPostRequest{
		Text:     "Daily recurring post",
		Accounts: []string{"account-1", "account-2"},
		Media: []v1.Media{
			{URL: "https://example.com/image.jpg", Type: "image"},
		},
		Recurrence: v1.RecurrenceRule{
			Frequency: "daily",
			Interval:  1,
			Count:     5,
		},
	}

	var resp v1.RecurringPostResponse
	err := client.CreateRecurringPost(context.Background(), req, &resp)

	require.NoError(t, err)
	assert.NotEmpty(t, resp.JobID)
	assert.Contains(t, resp.JobID, "recurring-")
}

func TestCreateRecurringPostWeekly(t *testing.T) {
	server := v1.SpawnMockServer()
	defer func() { _ = server.Stop() }()

	client := server.Client()

	req := v1.RecurringPostRequest{
		Text:     "Weekly recurring post",
		Accounts: []string{"account-1"},
		Recurrence: v1.RecurrenceRule{
			Frequency:  "weekly",
			Interval:   2,
			DaysOfWeek: []string{"monday", "friday"},
			EndDate:    time.Now().Add(30 * 24 * time.Hour),
		},
	}

	var resp v1.RecurringPostResponse
	err := client.CreateRecurringPost(context.Background(), req, &resp)

	require.NoError(t, err)
	assert.NotEmpty(t, resp.JobID)
}

func TestCreateRecurringPostValidation(t *testing.T) {
	server := v1.SpawnMockServer()
	defer func() { _ = server.Stop() }()

	client := server.Client()

	tests := []struct {
		name    string
		req     v1.RecurringPostRequest
		wantErr string
	}{
		{
			name: "missing text",
			req: v1.RecurringPostRequest{
				Accounts: []string{"account-1"},
				Recurrence: v1.RecurrenceRule{
					Frequency: "daily",
					Interval:  1,
					Count:     5,
				},
			},
			wantErr: "Text field is required",
		},
		{
			name: "missing accounts",
			req: v1.RecurringPostRequest{
				Text: "Post text",
				Recurrence: v1.RecurrenceRule{
					Frequency: "daily",
					Interval:  1,
					Count:     5,
				},
			},
			wantErr: "At least one account is required",
		},
		{
			name: "missing frequency",
			req: v1.RecurringPostRequest{
				Text:     "Post text",
				Accounts: []string{"account-1"},
				Recurrence: v1.RecurrenceRule{
					Interval: 1,
					Count:    5,
				},
			},
			wantErr: "Recurrence frequency is required",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var resp v1.RecurringPostResponse
			err := client.CreateRecurringPost(context.Background(), test.req, &resp)
			require.Error(t, err)
			require.ErrorContains(t, err, test.wantErr)
		})
	}
}

func TestAutoSchedulePost(t *testing.T) {
	server := v1.SpawnMockServer()
	defer func() { _ = server.Stop() }()

	client := server.Client()

	now := time.Now()
	req := v1.AutoScheduleRequest{
		Text:      "Auto-scheduled post",
		Accounts:  []string{"account-1", "account-2"},
		Media:     []v1.Media{{URL: "https://example.com/image.jpg", Type: "image"}},
		StartDate: now,
		EndDate:   now.Add(7 * 24 * time.Hour),
		Slots:     3,
	}

	var resp v1.AutoScheduleResponse
	err := client.AutoSchedulePost(context.Background(), req, &resp)

	require.NoError(t, err)
	assert.NotEmpty(t, resp.JobID)
	assert.Contains(t, resp.JobID, "auto-schedule-")
}

func TestAutoSchedulePostValidation(t *testing.T) {
	server := v1.SpawnMockServer()
	defer func() { _ = server.Stop() }()

	client := server.Client()

	now := time.Now()
	tests := []struct {
		name    string
		req     v1.AutoScheduleRequest
		wantErr string
	}{
		{
			name: "missing text",
			req: v1.AutoScheduleRequest{
				Accounts:  []string{"account-1"},
				StartDate: now,
				EndDate:   now.Add(24 * time.Hour),
				Slots:     3,
			},
			wantErr: "Text field is required",
		},
		{
			name: "missing accounts",
			req: v1.AutoScheduleRequest{
				Text:      "Post text",
				StartDate: now,
				EndDate:   now.Add(24 * time.Hour),
				Slots:     3,
			},
			wantErr: "At least one account is required",
		},
		{
			name: "invalid slots",
			req: v1.AutoScheduleRequest{
				Text:      "Post text",
				Accounts:  []string{"account-1"},
				StartDate: now,
				EndDate:   now.Add(24 * time.Hour),
				Slots:     0,
			},
			wantErr: "Slots must be greater than 0",
		},
		{
			name: "invalid date range",
			req: v1.AutoScheduleRequest{
				Text:      "Post text",
				Accounts:  []string{"account-1"},
				StartDate: now,
				EndDate:   now.Add(-24 * time.Hour),
				Slots:     3,
			},
			wantErr: "End date must be after start date",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var resp v1.AutoScheduleResponse
			err := client.AutoSchedulePost(context.Background(), test.req, &resp)
			require.Error(t, err)
			require.ErrorContains(t, err, test.wantErr)
		})
	}
}

func TestRecyclePost(t *testing.T) {
	server := v1.SpawnMockServer()
	defer func() { _ = server.Stop() }()

	client := server.Client()

	server.Reset()
	posts := []v1.Post{
		{ID: "existing-1", Text: "Original post", State: "published"},
	}
	server.AddPosts(posts)

	now := time.Now()
	req := v1.RecyclePostRequest{
		PostID:    "existing-1",
		StartDate: now,
		EndDate:   now.Add(30 * 24 * time.Hour),
		Frequency: "weekly",
		MaxCount:  4,
	}

	var resp v1.RecyclePostResponse
	err := client.RecyclePost(context.Background(), req, &resp)

	require.NoError(t, err)
	assert.NotEmpty(t, resp.JobID)
	assert.Contains(t, resp.JobID, "recycle-")
}

func TestRecyclePostValidation(t *testing.T) {
	server := v1.SpawnMockServer()
	defer func() { _ = server.Stop() }()

	client := server.Client()

	server.Reset()
	posts := []v1.Post{
		{ID: "existing-1", Text: "Original post", State: "published"},
	}
	server.AddPosts(posts)

	now := time.Now()
	tests := []struct {
		name    string
		req     v1.RecyclePostRequest
		wantErr string
	}{
		{
			name: "missing post ID",
			req: v1.RecyclePostRequest{
				StartDate: now,
				EndDate:   now.Add(24 * time.Hour),
				Frequency: "daily",
				MaxCount:  3,
			},
			wantErr: "Post ID is required",
		},
		{
			name: "missing frequency",
			req: v1.RecyclePostRequest{
				PostID:    "existing-1",
				StartDate: now,
				EndDate:   now.Add(24 * time.Hour),
				MaxCount:  3,
			},
			wantErr: "Frequency is required",
		},
		{
			name: "invalid max count",
			req: v1.RecyclePostRequest{
				PostID:    "existing-1",
				StartDate: now,
				EndDate:   now.Add(24 * time.Hour),
				Frequency: "daily",
				MaxCount:  0,
			},
			wantErr: "Max count must be greater than 0",
		},
		{
			name: "invalid date range",
			req: v1.RecyclePostRequest{
				PostID:    "existing-1",
				StartDate: now,
				EndDate:   now.Add(-24 * time.Hour),
				Frequency: "daily",
				MaxCount:  3,
			},
			wantErr: "End date must be after start date",
		},
		{
			name: "post not found",
			req: v1.RecyclePostRequest{
				PostID:    "nonexistent",
				StartDate: now,
				EndDate:   now.Add(24 * time.Hour),
				Frequency: "daily",
				MaxCount:  3,
			},
			wantErr: "Post not found",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var resp v1.RecyclePostResponse
			err := client.RecyclePost(context.Background(), test.req, &resp)
			require.Error(t, err)
			require.ErrorContains(t, err, test.wantErr)
		})
	}
}

func TestRecurrencePatterns(t *testing.T) {
	server := v1.SpawnMockServer()
	defer func() { _ = server.Stop() }()

	client := server.Client()

	tests := []struct {
		name       string
		recurrence v1.RecurrenceRule
	}{
		{
			name: "daily with count",
			recurrence: v1.RecurrenceRule{
				Frequency: "daily",
				Interval:  1,
				Count:     7,
			},
		},
		{
			name: "weekly with end date",
			recurrence: v1.RecurrenceRule{
				Frequency: "weekly",
				Interval:  2,
				EndDate:   time.Now().Add(60 * 24 * time.Hour),
			},
		},
		{
			name: "monthly pattern",
			recurrence: v1.RecurrenceRule{
				Frequency: "monthly",
				Interval:  1,
				Count:     6,
			},
		},
		{
			name: "weekly with specific days",
			recurrence: v1.RecurrenceRule{
				Frequency:  "weekly",
				Interval:   1,
				DaysOfWeek: []string{"monday", "wednesday", "friday"},
				Count:      10,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := v1.RecurringPostRequest{
				Text:       "Test recurring post",
				Accounts:   []string{"account-1"},
				Recurrence: test.recurrence,
			}

			var resp v1.RecurringPostResponse
			err := client.CreateRecurringPost(context.Background(), req, &resp)

			require.NoError(t, err)
			assert.NotEmpty(t, resp.JobID)
		})
	}
}