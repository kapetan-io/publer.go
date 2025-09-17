package v1_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "github.com/thrawn/publer.go/v1"
)

func TestListAccounts(t *testing.T) {
	server := v1.SpawnMockServer()
	defer server.Stop()

	client := server.Client()

	accounts := []v1.Account{
		{
			Picture:  "https://example.com/fb-pic.jpg",
			Name:     "My Facebook Page",
			SocialID: "fb-123456",
			Provider: "facebook",
			ID:       "account-1",
			Type:     "page",
		},
		{
			Picture:  "https://example.com/ig-pic.jpg",
			Name:     "My Instagram Business",
			SocialID: "ig-789012",
			Provider: "instagram",
			ID:       "account-2",
			Type:     "business",
		},
	}

	server.Reset()
	server.AddAccounts(accounts)

	iterator := client.ListAccounts(context.Background(), v1.ListAccountsRequest{})

	var page v1.Page[v1.Account]
	hasMore := iterator.Next(context.Background(), &page)
	require.NoError(t, iterator.Err())

	// Validate response structure matches API specification
	assert.Equal(t, 2, page.Total)
	assert.Equal(t, 1, page.Page)
	assert.Equal(t, 10, page.PerPage)
	assert.Equal(t, 1, page.TotalPages)
	assert.Len(t, page.Items, 2)

	// Validate account data
	assert.Equal(t, "account-1", page.Items[0].ID)
	assert.Equal(t, "facebook", page.Items[0].Provider)
	assert.Equal(t, "My Facebook Page", page.Items[0].Name)
	assert.Equal(t, "fb-123456", page.Items[0].SocialID)
	assert.Equal(t, "page", page.Items[0].Type)

	assert.Equal(t, "account-2", page.Items[1].ID)
	assert.Equal(t, "instagram", page.Items[1].Provider)
	assert.Equal(t, "My Instagram Business", page.Items[1].Name)
	assert.Equal(t, "ig-789012", page.Items[1].SocialID)
	assert.Equal(t, "business", page.Items[1].Type)

	// Should be no more pages
	assert.False(t, hasMore)
}

func TestAccountProviders(t *testing.T) {
	server := v1.SpawnMockServer()
	defer server.Stop()

	client := server.Client()

	accounts := []v1.Account{
		{
			Name:     "Facebook Page",
			SocialID: "fb-123",
			Provider: "facebook",
			ID:       "fb-1",
			Type:     "page",
		},
		{
			Name:     "Instagram Account",
			SocialID: "ig-456",
			Provider: "instagram",
			Type:     "personal",
			ID:       "ig-1",
		},
		{
			Name:     "Twitter Account",
			SocialID: "tw-789",
			Provider: "twitter",
			Type:     "profile",
			ID:       "tw-1",
		},
		{
			Name:     "LinkedIn Company",
			SocialID: "li-012",
			Provider: "linkedin",
			Type:     "company",
			ID:       "li-1",
		},
		{
			Name:     "YouTube Channel",
			SocialID: "yt-345",
			Provider: "youtube",
			Type:     "channel",
			ID:       "yt-1",
		},
	}

	server.Reset()
	server.AddAccounts(accounts)

	iterator := client.ListAccounts(context.Background(), v1.ListAccountsRequest{})

	var page v1.Page[v1.Account]
	iterator.Next(context.Background(), &page)
	require.NoError(t, iterator.Err())

	// Validate all providers are supported
	providers := make(map[string]bool)
	for _, account := range page.Items {
		providers[account.Provider] = true
	}

	assert.True(t, providers["facebook"])
	assert.True(t, providers["instagram"])
	assert.True(t, providers["twitter"])
	assert.True(t, providers["linkedin"])
	assert.True(t, providers["youtube"])
}

func TestAccountTypes(t *testing.T) {
	server := v1.SpawnMockServer()
	defer server.Stop()

	client := server.Client()

	accounts := []v1.Account{
		{
			Name:     "Facebook Page",
			Provider: "facebook",
			ID:       "page-1",
			Type:     "page",
		},
		{
			Name:     "Facebook Profile",
			Provider: "facebook",
			ID:       "profile-1",
			Type:     "profile",
		},
		{
			Name:     "Facebook Group",
			Provider: "facebook",
			ID:       "group-1",
			Type:     "group",
		},
		{
			Name:     "Instagram Business",
			Provider: "instagram",
			ID:       "business-1",
			Type:     "business",
		},
		{
			Name:     "Instagram Personal",
			Provider: "instagram",
			ID:       "personal-1",
			Type:     "personal",
		},
	}

	server.Reset()
	server.AddAccounts(accounts)

	iterator := client.ListAccounts(context.Background(), v1.ListAccountsRequest{})

	var page v1.Page[v1.Account]
	iterator.Next(context.Background(), &page)
	require.NoError(t, iterator.Err())

	// Validate account types are correctly parsed
	types := make(map[string]bool)
	for _, account := range page.Items {
		types[account.Type] = true
	}

	assert.True(t, types["page"])
	assert.True(t, types["profile"])
	assert.True(t, types["group"])
	assert.True(t, types["business"])
	assert.True(t, types["personal"])
}

func TestListAccountsEmpty(t *testing.T) {
	server := v1.SpawnMockServer()
	defer server.Stop()

	client := server.Client()

	server.Reset()

	iterator := client.ListAccounts(context.Background(), v1.ListAccountsRequest{})

	var page v1.Page[v1.Account]
	hasMore := iterator.Next(context.Background(), &page)
	require.NoError(t, iterator.Err())

	// Should handle empty results correctly
	assert.Equal(t, 0, page.Total)
	assert.Equal(t, 1, page.Page)
	assert.Equal(t, 10, page.PerPage)
	assert.Equal(t, 0, page.TotalPages)
	assert.Len(t, page.Items, 0)
	assert.False(t, hasMore)
}

func TestListAccountsPagination(t *testing.T) {
	server := v1.SpawnMockServer()
	defer server.Stop()

	client := server.Client()

	// Create 15 accounts to test pagination (more than default 10 per page)
	accounts := make([]v1.Account, 15)
	for i := 0; i < 15; i++ {
		accounts[i] = v1.Account{
			Name:     fmt.Sprintf("Account %d", i+1),
			SocialID: fmt.Sprintf("fb-%d", i+1),
			ID:       fmt.Sprintf("account-%d", i+1),
			Provider: "facebook",
			Type:     "page",
		}
	}

	server.Reset()
	server.AddAccounts(accounts)

	iterator := client.ListAccounts(context.Background(), v1.ListAccountsRequest{})

	// First page
	var page1 v1.Page[v1.Account]
	hasMore := iterator.Next(context.Background(), &page1)
	require.NoError(t, iterator.Err())

	assert.Equal(t, 15, page1.Total)
	assert.Equal(t, 1, page1.Page)
	assert.Equal(t, 10, page1.PerPage)
	assert.Equal(t, 2, page1.TotalPages)
	assert.Len(t, page1.Items, 10)
	assert.True(t, hasMore)

	// Second page
	var page2 v1.Page[v1.Account]
	hasMore = iterator.Next(context.Background(), &page2)
	require.NoError(t, iterator.Err())

	assert.Equal(t, 15, page2.Total)
	assert.Equal(t, 2, page2.Page)
	assert.Equal(t, 10, page2.PerPage)
	assert.Equal(t, 2, page2.TotalPages)
	assert.Len(t, page2.Items, 5)
	assert.False(t, hasMore)
}

func TestListAccountsContextCancellation(t *testing.T) {
	server := v1.SpawnMockServer()
	defer server.Stop()

	client := server.Client()

	accounts := []v1.Account{
		{
			Name:     "Test Account",
			Provider: "facebook",
			ID:       "account-1",
		},
	}

	server.Reset()
	server.AddAccounts(accounts)

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	iterator := client.ListAccounts(ctx, v1.ListAccountsRequest{})

	var page v1.Page[v1.Account]
	hasMore := iterator.Next(ctx, &page)

	assert.False(t, hasMore)
	require.ErrorContains(t, iterator.Err(), "context canceled")
}