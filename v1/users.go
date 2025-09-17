package v1

import "context"

// GetMeRequest represents request for current user
type GetMeRequest struct{}

// GetMeResponse represents current user response
type GetMeResponse struct {
	User
}

// GetMe retrieves information about the currently authenticated user
func (c *Client) GetMe(ctx context.Context, req GetMeRequest, resp *GetMeResponse) error {
	return c.do(ctx, "GET", "users/me", nil, resp)
}