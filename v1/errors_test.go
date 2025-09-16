package v1_test

import (
	"errors"
	"testing"

	v1 "github.com/thrawn/publer.go/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPIError(t *testing.T) {
	// Test APIError formatting
	err := &v1.APIError{
		Method:     "GET",
		URL:        "https://app.publer.com/api/v1/posts",
		StatusCode: 404,
		Message:    "Not found",
	}

	expected := `GET https://app.publer.com/api/v1/posts with 404 returned "Not found"`
	assert.Equal(t, expected, err.Error())

	// Test error type assertion
	var apiErr *v1.APIError
	require.True(t, errors.As(err, &apiErr))
	assert.Equal(t, 404, apiErr.StatusCode)
	assert.Equal(t, "Not found", apiErr.Message)
}

func TestRateLimitError(t *testing.T) {
	// Test RateLimitError formatting
	err := &v1.RateLimitError{
		APIError: v1.APIError{
			Method:     "POST",
			URL:        "https://app.publer.com/api/v1/posts",
			StatusCode: 429,
			Message:    "Rate limit exceeded",
		},
		Limit:     100,
		Remaining: 0,
		Reset:     1640995200,
	}

	expected := `POST https://app.publer.com/api/v1/posts with 429 returned "Rate limit exceeded"`
	assert.Equal(t, expected, err.Error())

	// Test error type assertion
	var rateLimitErr *v1.RateLimitError
	require.True(t, errors.As(err, &rateLimitErr))
	assert.Equal(t, 429, rateLimitErr.StatusCode)
	assert.Equal(t, 100, rateLimitErr.Limit)
	assert.Equal(t, 0, rateLimitErr.Remaining)
	assert.Equal(t, int64(1640995200), rateLimitErr.Reset)

	// Test that RateLimitError is also an APIError
	var apiErr *v1.APIError
	require.True(t, errors.As(err, &apiErr))
	assert.Equal(t, 429, apiErr.StatusCode)
}

func TestErrorFormatting(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name: "simple API error",
			err: &v1.APIError{
				Method:     "DELETE",
				URL:        "https://app.publer.com/api/v1/posts/123",
				StatusCode: 403,
				Message:    "Forbidden",
			},
			expected: `DELETE https://app.publer.com/api/v1/posts/123 with 403 returned "Forbidden"`,
		},
		{
			name: "API error with long message",
			err: &v1.APIError{
				Method:     "PUT",
				URL:        "https://app.publer.com/api/v1/posts/456",
				StatusCode: 400,
				Message:    "Invalid request: missing required field 'text'",
			},
			expected: `PUT https://app.publer.com/api/v1/posts/456 with 400 returned "Invalid request: missing required field 'text'"`,
		},
		{
			name: "rate limit error",
			err: &v1.RateLimitError{
				APIError: v1.APIError{
					Method:     "GET",
					URL:        "https://app.publer.com/api/v1/accounts",
					StatusCode: 429,
					Message:    "Too many requests",
				},
				Limit:     100,
				Remaining: 0,
				Reset:     1640995200,
			},
			expected: `GET https://app.publer.com/api/v1/accounts with 429 returned "Too many requests"`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, test.err.Error())
		})
	}
}