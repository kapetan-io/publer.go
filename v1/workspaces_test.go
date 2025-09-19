package v1_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "github.com/thrawn/publer.go/v1"
)

func TestListWorkspaces(t *testing.T) {
	server := v1.SpawnMockServer()
	defer func() { _ = server.Stop() }()

	client := server.Client()

	owner := v1.User{
		ID:        "owner-123",
		Email:     "owner@example.com",
		Name:      "Owner User",
		FirstName: "Owner",
		Picture:   "https://example.com/owner.jpg",
	}

	member1 := v1.User{
		ID:        "member-456",
		Email:     "member1@example.com",
		Name:      "Member One",
		FirstName: "Member",
		Picture:   "https://example.com/member1.jpg",
	}

	member2 := v1.User{
		ID:        "member-789",
		Email:     "member2@example.com",
		Name:      "Member Two",
		FirstName: "Member",
		Picture:   "https://example.com/member2.jpg",
	}

	workspaces := []v1.Workspace{
		{
			ID:      "workspace-1",
			Owner:   owner,
			Name:    "Personal Workspace",
			Members: []v1.User{owner},
			Plan:    "free",
			Picture: "https://example.com/workspace1.jpg",
		},
		{
			ID:      "workspace-2",
			Owner:   owner,
			Name:    "Team Workspace",
			Members: []v1.User{owner, member1, member2},
			Plan:    "pro",
			Picture: "https://example.com/workspace2.jpg",
		},
	}

	server.Reset()
	server.AddWorkspaces(workspaces)

	iterator := client.ListWorkspaces(context.Background(), v1.ListWorkspacesRequest{})

	var page v1.Page[v1.Workspace]
	hasMore := iterator.Next(context.Background(), &page)
	require.NoError(t, iterator.Err())

	assert.Equal(t, 2, page.Total)
	assert.Equal(t, 1, page.Page)
	assert.Equal(t, 10, page.PerPage)
	assert.Equal(t, 1, page.TotalPages)
	assert.Len(t, page.Items, 2)
	assert.False(t, hasMore)

	first := page.Items[0]
	assert.Equal(t, "workspace-1", first.ID)
	assert.Equal(t, "Personal Workspace", first.Name)
	assert.Equal(t, "free", first.Plan)
	assert.Equal(t, owner.ID, first.Owner.ID)
	assert.Len(t, first.Members, 1)
	assert.Equal(t, owner.ID, first.Members[0].ID)

	second := page.Items[1]
	assert.Equal(t, "workspace-2", second.ID)
	assert.Equal(t, "Team Workspace", second.Name)
	assert.Equal(t, "pro", second.Plan)
	assert.Equal(t, owner.ID, second.Owner.ID)
	assert.Len(t, second.Members, 3)
}

func TestWorkspaceMembers(t *testing.T) {
	server := v1.SpawnMockServer()
	defer func() { _ = server.Stop() }()

	client := server.Client()

	owner := v1.User{
		ID:        "owner-123",
		Email:     "owner@example.com",
		Name:      "Owner User",
		FirstName: "Owner",
		Picture:   "https://example.com/owner.jpg",
	}

	member := v1.User{
		ID:        "member-456",
		Email:     "member@example.com",
		Name:      "Member User",
		FirstName: "Member",
		Picture:   "https://example.com/member.jpg",
	}

	workspace := v1.Workspace{
		ID:      "workspace-1",
		Owner:   owner,
		Name:    "Test Workspace",
		Members: []v1.User{owner, member},
		Plan:    "pro",
		Picture: "https://example.com/workspace.jpg",
	}

	server.Reset()
	server.AddWorkspace(workspace)

	iterator := client.ListWorkspaces(context.Background(), v1.ListWorkspacesRequest{})

	var page v1.Page[v1.Workspace]
	iterator.Next(context.Background(), &page)
	require.NoError(t, iterator.Err())

	require.Len(t, page.Items, 1)

	assert.Equal(t, owner.ID, page.Items[0].Owner.ID)
	assert.Equal(t, owner.Email, page.Items[0].Owner.Email)
	assert.Equal(t, owner.Name, page.Items[0].Owner.Name)

	require.Len(t, page.Items[0].Members, 2)
	assert.Equal(t, owner.ID, page.Items[0].Members[0].ID)
	assert.Equal(t, member.ID, page.Items[0].Members[1].ID)
	assert.Equal(t, member.Email, page.Items[0].Members[1].Email)
	assert.Equal(t, member.Name, page.Items[0].Members[1].Name)
}

func TestListWorkspacesEmpty(t *testing.T) {
	server := v1.SpawnMockServer()
	defer func() { _ = server.Stop() }()

	client := server.Client()

	server.Reset()

	iterator := client.ListWorkspaces(context.Background(), v1.ListWorkspacesRequest{})

	var page v1.Page[v1.Workspace]
	hasMore := iterator.Next(context.Background(), &page)
	require.NoError(t, iterator.Err())

	assert.Equal(t, 0, page.Total)
	assert.Equal(t, 1, page.Page)
	assert.Equal(t, 10, page.PerPage)
	assert.Equal(t, 0, page.TotalPages)
	assert.Len(t, page.Items, 0)
	assert.False(t, hasMore)
}