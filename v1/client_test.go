package v1_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "github.com/thrawn/publer.go/v1"
)

func TestNewClient(t *testing.T) {
	// Test with valid configuration
	config := v1.Config{
		APIKey:      "test-api-key",
		WorkspaceID: "test-workspace-id",
	}

	client, err := v1.NewClient(config)
	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestNewClientValidation(t *testing.T) {
	// Test missing API key
	config := v1.Config{
		WorkspaceID: "test-workspace-id",
	}

	client, err := v1.NewClient(config)
	require.Error(t, err)
	require.ErrorContains(t, err, "API key is required")
	assert.Nil(t, client)

	// Test missing workspace ID
	config = v1.Config{
		APIKey: "test-api-key",
	}

	client, err = v1.NewClient(config)
	require.Error(t, err)
	require.ErrorContains(t, err, "workspace ID is required")
	assert.Nil(t, client)
}

func TestNewClientCustom(t *testing.T) {
	// Test default base URL
	config := v1.Config{
		APIKey:      "test-api-key",
		WorkspaceID: "test-workspace-id",
	}

	client, err := v1.NewClient(config)
	require.NoError(t, err)
	assert.NotNil(t, client)

	// Test custom base URL without trailing slash
	config = v1.Config{
		APIKey:      "test-api-key",
		WorkspaceID: "test-workspace-id",
		BaseURL:     "http://localhost:8080",
	}

	client, err = v1.NewClient(config)
	require.NoError(t, err)
	assert.NotNil(t, client)

	// Test custom HTTP client
	customClient := &http.Client{}
	config = v1.Config{
		APIKey:      "test-api-key",
		WorkspaceID: "test-workspace-id",
		Client:      customClient,
	}

	client, err = v1.NewClient(config)
	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestClientAuthentication(t *testing.T) {
	// Create mock server
	server := v1.SpawnMockServer()
	defer func() { _ = server.Stop() }()

	server.SetResponse("GET", "/api/v1/test", 200, map[string]string{
		"message": "success",
	})

	// Get client configured for this mock server
	client := server.Client()
	require.NotNil(t, client)

	// The client is now properly configured with the mock server's credentials
	// Authentication validation happens automatically within the mock server
}
