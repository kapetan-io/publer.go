package v1

import (
	"fmt"
)

// ErrorResponse represents the JSON error response from Publer API
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    string `json:"code,omitempty"`
}

// APIError represents an error response from the Publer API
type APIError struct {
	Method     string
	URL        string
	StatusCode int
	Message    string
}

// Error returns the formatted error message
func (e *APIError) Error() string {
	return fmt.Sprintf("%s %s with %d returned \"%s\"", e.Method, e.URL, e.StatusCode, e.Message)
}

// RateLimitError represents a rate limit exceeded error
type RateLimitError struct {
	APIError
	Limit     int
	Remaining int
	Reset     int64
}

// Error returns the formatted rate limit error message
func (e *RateLimitError) Error() string {
	return fmt.Sprintf("%s %s with %d returned \"%s\"", e.Method, e.URL, e.StatusCode, e.Message)
}

// As implements error unwrapping for errors.As
func (e *RateLimitError) As(target interface{}) bool {
	switch t := target.(type) {
	case **APIError:
		*t = &e.APIError
		return true
	default:
		return false
	}
}