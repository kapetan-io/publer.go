package v1_test

import (
	"context"
	"errors"
	"testing"
	"time"

	v1 "github.com/thrawn/publer.go/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockPageFetcher struct {
	pages []v1.Page[v1.Post]
	err   error
}

func (m *mockPageFetcher) FetchPage(ctx context.Context, pageNum int) (*v1.Page[v1.Post], error) {
	if m.err != nil {
		return nil, m.err
	}

	if pageNum <= 0 || pageNum > len(m.pages) {
		return &v1.Page[v1.Post]{
			Items:      []v1.Post{},
			Total:      0,
			Page:       pageNum,
			PerPage:    10,
			TotalPages: 0,
		}, nil
	}

	return &m.pages[pageNum-1], nil
}

func TestGenericIterator(t *testing.T) {
	// Test with multiple pages
	pages := []v1.Page[v1.Post]{
		{
			Items:      []v1.Post{{ID: "1", Text: "First post"}},
			Total:      3,
			Page:       1,
			PerPage:    1,
			TotalPages: 3,
		},
		{
			Items:      []v1.Post{{ID: "2", Text: "Second post"}},
			Total:      3,
			Page:       2,
			PerPage:    1,
			TotalPages: 3,
		},
		{
			Items:      []v1.Post{{ID: "3", Text: "Third post"}},
			Total:      3,
			Page:       3,
			PerPage:    1,
			TotalPages: 3,
		},
	}

	fetcher := &mockPageFetcher{pages: pages}
	iterator := v1.NewGenericIterator[v1.Post](fetcher)

	ctx := context.Background()

	// First page
	var page1 v1.Page[v1.Post]
	hasMore := iterator.Next(ctx, &page1)
	require.True(t, hasMore)
	require.NoError(t, iterator.Err())
	assert.Len(t, page1.Items, 1)
	assert.Equal(t, "1", page1.Items[0].ID)
	assert.Equal(t, 1, page1.Page)
	assert.Equal(t, 3, page1.TotalPages)

	// Second page
	var page2 v1.Page[v1.Post]
	hasMore = iterator.Next(ctx, &page2)
	require.True(t, hasMore)
	require.NoError(t, iterator.Err())
	assert.Len(t, page2.Items, 1)
	assert.Equal(t, "2", page2.Items[0].ID)
	assert.Equal(t, 2, page2.Page)

	// Third page (last)
	var page3 v1.Page[v1.Post]
	hasMore = iterator.Next(ctx, &page3)
	require.False(t, hasMore)
	require.NoError(t, iterator.Err())
	assert.Len(t, page3.Items, 1)
	assert.Equal(t, "3", page3.Items[0].ID)
	assert.Equal(t, 3, page3.Page)

	// No more pages
	var page4 v1.Page[v1.Post]
	hasMore = iterator.Next(ctx, &page4)
	require.False(t, hasMore)
	require.NoError(t, iterator.Err())
}

func TestGenericIteratorEmptyResult(t *testing.T) {
	// Test with empty result
	pages := []v1.Page[v1.Post]{
		{
			Items:      []v1.Post{},
			Total:      0,
			Page:       1,
			PerPage:    10,
			TotalPages: 0,
		},
	}

	fetcher := &mockPageFetcher{pages: pages}
	iterator := v1.NewGenericIterator[v1.Post](fetcher)

	ctx := context.Background()

	var page v1.Page[v1.Post]
	hasMore := iterator.Next(ctx, &page)
	require.False(t, hasMore)
	require.NoError(t, iterator.Err())
	assert.Empty(t, page.Items)
	assert.Equal(t, 0, page.TotalPages)
}

func TestGenericIteratorError(t *testing.T) {
	// Test with error
	expectedErr := errors.New("fetch error")
	fetcher := &mockPageFetcher{err: expectedErr}
	iterator := v1.NewGenericIterator[v1.Post](fetcher)

	ctx := context.Background()

	var page v1.Page[v1.Post]
	hasMore := iterator.Next(ctx, &page)
	require.False(t, hasMore)
	assert.ErrorIs(t, iterator.Err(), expectedErr)
}

func TestGenericIteratorContextCancellation(t *testing.T) {
	// Test with cancelled context
	pages := []v1.Page[v1.Post]{
		{
			Items:      []v1.Post{{ID: "1", Text: "First post"}},
			Total:      1,
			Page:       1,
			PerPage:    10,
			TotalPages: 1,
		},
	}

	fetcher := &mockPageFetcher{pages: pages}
	iterator := v1.NewGenericIterator[v1.Post](fetcher)

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var page v1.Page[v1.Post]
	hasMore := iterator.Next(ctx, &page)
	require.False(t, hasMore)
	assert.ErrorIs(t, iterator.Err(), context.Canceled)
}

func TestGenericIteratorContextTimeout(t *testing.T) {
	// Test with timeout context
	pages := []v1.Page[v1.Post]{
		{
			Items:      []v1.Post{{ID: "1", Text: "First post"}},
			Total:      1,
			Page:       1,
			PerPage:    10,
			TotalPages: 1,
		},
	}

	fetcher := &mockPageFetcher{pages: pages}
	iterator := v1.NewGenericIterator[v1.Post](fetcher)

	// Create timeout context that expires immediately
	ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancel()

	// Wait for context to timeout
	time.Sleep(time.Millisecond)

	var page v1.Page[v1.Post]
	hasMore := iterator.Next(ctx, &page)
	require.False(t, hasMore)
	assert.ErrorIs(t, iterator.Err(), context.DeadlineExceeded)
}

func TestGenericIteratorSinglePage(t *testing.T) {
	// Test with single page
	pages := []v1.Page[v1.Post]{
		{
			Items: []v1.Post{
				{ID: "1", Text: "First post"},
				{ID: "2", Text: "Second post"},
			},
			Total:      2,
			Page:       1,
			PerPage:    10,
			TotalPages: 1,
		},
	}

	fetcher := &mockPageFetcher{pages: pages}
	iterator := v1.NewGenericIterator[v1.Post](fetcher)

	ctx := context.Background()

	// First and only page
	var page v1.Page[v1.Post]
	hasMore := iterator.Next(ctx, &page)
	require.False(t, hasMore)
	require.NoError(t, iterator.Err())
	assert.Len(t, page.Items, 2)
	assert.Equal(t, 1, page.Page)
	assert.Equal(t, 1, page.TotalPages)

	// No more pages
	var page2 v1.Page[v1.Post]
	hasMore = iterator.Next(ctx, &page2)
	require.False(t, hasMore)
	require.NoError(t, iterator.Err())
}