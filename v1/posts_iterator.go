package v1

import (
	"context"
	"net/url"
	"strconv"
	"time"
)

// PostPageFetcher implements PageFetcher for posts
type PostPageFetcher struct {
	client  *Client
	request ListPostsRequest
}

// FetchPage fetches a page of posts
func (f *PostPageFetcher) FetchPage(ctx context.Context, pageNum int) (*Page[Post], error) {
	// Create a copy of the request with the specific page number
	request := f.request
	request.Page = pageNum

	// Build query parameters
	params := url.Values{}
	if request.State != "" {
		params.Set("state", request.State)
	}
	for _, state := range request.States {
		params.Add("state[]", state)
	}
	if !request.From.IsZero() {
		params.Set("from", request.From.Format(time.RFC3339))
	}
	if !request.To.IsZero() {
		params.Set("to", request.To.Format(time.RFC3339))
	}
	if pageNum > 0 {
		params.Set("page", strconv.Itoa(pageNum))
	}
	for _, accountID := range request.AccountIDs {
		params.Add("account_ids[]", accountID)
	}
	if request.Query != "" {
		params.Set("query", request.Query)
	}
	if request.PostType != "" {
		params.Set("postType", request.PostType)
	}
	if request.MemberID != "" {
		params.Set("member_id", request.MemberID)
	}

	// Make API call to get posts
	var response ListPostsResponse
	err := f.client.do(ctx, "GET", "posts?"+params.Encode(), nil, &response)
	if err != nil {
		return nil, err
	}

	// Map ListPostsResponse to Page[Post] structure
	return &Page[Post]{
		Items:      response.Posts,
		Total:      response.Total,
		Page:       response.Page,
		PerPage:    response.PerPage,
		TotalPages: response.TotalPages,
	}, nil
}

// NewPostIterator creates a new iterator for posts
func NewPostIterator(client *Client, request ListPostsRequest) Iterator[Post] {
	fetcher := &PostPageFetcher{
		client:  client,
		request: request,
	}
	return NewGenericIterator(fetcher)
}