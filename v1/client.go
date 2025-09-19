package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const defaultBaseURL = "https://app.publer.com/api/v1/"

// Config holds client configuration options
type Config struct {
	APIKey      string
	WorkspaceID string
	BaseURL     string
	Client      *http.Client
}

// Client represents the Publer API client
type Client struct {
	config     Config
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new Publer API client
func NewClient(config Config) (*Client, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}
	if config.WorkspaceID == "" {
		return nil, fmt.Errorf("workspace ID is required")
	}

	httpClient := config.Client
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}

	return &Client{
		config:     config,
		httpClient: httpClient,
		baseURL:    baseURL,
	}, nil
}

// do performs HTTP requests with authentication
func (c *Client) do(ctx context.Context, method, path string, body any, result any) error {
	// Build the full URL
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return fmt.Errorf("invalid base URL: %w", err)
	}

	// Ensure path doesn't start with /
	path = strings.TrimPrefix(path, "/")

	rel, err := url.Parse(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	fullURL := u.ResolveReference(rel).String()

	// Prepare request body
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, fullURL, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer-API %s", c.config.APIKey))
	req.Header.Set("Publer-Workspace-Id", c.config.WorkspaceID)

	// Add content type for JSON
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle errors
	if resp.StatusCode >= 400 {
		if resp.StatusCode == 429 {
			// Rate limit error
			rateLimitErr := &RateLimitError{
				APIError: APIError{
					Method:     method,
					URL:        fullURL,
					StatusCode: resp.StatusCode,
				},
			}

			// Parse rate limit headers safely
			if limit := resp.Header.Get("X-RateLimit-Limit"); limit != "" {
				if n, err := fmt.Sscanf(limit, "%d", &rateLimitErr.Limit); n != 1 || err != nil {
					rateLimitErr.Limit = 0
				}
			}
			if remaining := resp.Header.Get("X-RateLimit-Remaining"); remaining != "" {
				if n, err := fmt.Sscanf(remaining, "%d", &rateLimitErr.Remaining); n != 1 || err != nil {
					rateLimitErr.Remaining = 0
				}
			}
			if reset := resp.Header.Get("X-RateLimit-Reset"); reset != "" {
				if n, err := fmt.Sscanf(reset, "%d", &rateLimitErr.Reset); n != 1 || err != nil {
					rateLimitErr.Reset = 0
				}
			}

			// Try to parse error message from body
			var errResp ErrorResponse
			if err := json.Unmarshal(respBody, &errResp); err == nil {
				rateLimitErr.Message = errResp.Message
				if rateLimitErr.Message == "" {
					rateLimitErr.Message = errResp.Error
				}
			}

			return rateLimitErr
		}

		// Regular API error
		apiErr := &APIError{
			Method:     method,
			URL:        fullURL,
			StatusCode: resp.StatusCode,
		}

		// Try to parse error message from body
		var errResp ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil {
			apiErr.Message = errResp.Message
			if apiErr.Message == "" {
				apiErr.Message = errResp.Error
			}
		}

		if apiErr.Message == "" {
			apiErr.Message = string(respBody)
		}

		return apiErr
	}

	// Parse successful response
	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// Test performs a test request to verify connectivity (for testing purposes only)
func (c *Client) Test(ctx context.Context) error {
	var result map[string]interface{}
	return c.do(ctx, "GET", "test", nil, &result)
}