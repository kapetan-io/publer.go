package v1_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "github.com/thrawn/publer.go/v1"
)

func TestGetMe(t *testing.T) {
	server := v1.SpawnMockServer()
	defer server.Stop()

	client := server.Client()

	user := v1.User{
		ID:        "user-123",
		Email:     "test@example.com",
		Name:      "Test User",
		FirstName: "Test",
		Picture:   "https://example.com/avatar.jpg",
	}

	server.Reset()
	server.SetCurrentUser(user)

	var resp v1.GetMeResponse
	err := client.GetMe(context.Background(), v1.GetMeRequest{}, &resp)
	require.NoError(t, err)

	assert.Equal(t, user.ID, resp.ID)
	assert.Equal(t, user.Email, resp.Email)
	assert.Equal(t, user.Name, resp.Name)
	assert.Equal(t, user.FirstName, resp.FirstName)
	assert.Equal(t, user.Picture, resp.Picture)
}

func TestGetMeNotFound(t *testing.T) {
	server := v1.SpawnMockServer()
	defer server.Stop()

	client := server.Client()

	server.Reset()

	var resp v1.GetMeResponse
	err := client.GetMe(context.Background(), v1.GetMeRequest{}, &resp)
	require.Error(t, err)

	var apiErr *v1.APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 404, apiErr.StatusCode)
}

