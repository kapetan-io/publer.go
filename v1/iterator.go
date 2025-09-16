package v1

import (
	"context"
)

// Page represents a page of results from paginated API
type Page[T any] struct {
	Items      []T `json:"items"`
	Total      int `json:"total"`
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	TotalPages int `json:"total_pages"`
}

// Iterator provides iteration over paginated API resources
type Iterator[T any] interface {
	Next(ctx context.Context, page *Page[T]) bool
	Err() error
}

// PageFetcher defines how to fetch pages of data
type PageFetcher[T any] interface {
	FetchPage(ctx context.Context, pageNum int) (*Page[T], error)
}

// GenericIterator implements Iterator for any paginated resource
type GenericIterator[T any] struct {
	fetcher     PageFetcher[T]
	currentPage int
	totalPages  int
	err         error
	initialized bool
}

// NewGenericIterator creates a new iterator for paginated resources
func NewGenericIterator[T any](fetcher PageFetcher[T]) *GenericIterator[T] {
	return &GenericIterator[T]{
		fetcher: fetcher,
	}
}

// Next fetches the next page of results
// Returns false when no more pages or context cancelled
// Check Err() for context cancellation or other errors
func (it *GenericIterator[T]) Next(ctx context.Context, page *Page[T]) bool {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		it.err = ctx.Err()
		return false
	default:
	}

	// Lazy initialization on first call
	if !it.initialized {
		it.currentPage = 0
		it.initialized = true
	}

	// Check if we've reached the end
	if it.totalPages > 0 && it.currentPage >= it.totalPages {
		return false
	}

	// Fetch the next page
	it.currentPage++
	fetchedPage, err := it.fetcher.FetchPage(ctx, it.currentPage)
	if err != nil {
		it.err = err
		return false
	}

	// Update total pages if this is the first page
	if it.currentPage == 1 {
		it.totalPages = fetchedPage.TotalPages
	}

	// Copy the fetched page data to the provided page
	*page = *fetchedPage

	// Check if we have more pages
	return it.currentPage < it.totalPages
}

// Err returns any error encountered during iteration
func (it *GenericIterator[T]) Err() error {
	return it.err
}