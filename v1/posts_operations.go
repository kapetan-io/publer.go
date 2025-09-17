package v1

import (
	"context"
	"net/url"
	"strconv"
	"time"
)

// PostFetcher implements PageFetcher for Post resources
type PostFetcher struct {
	client  *Client
	request ListPostsRequest
}

// FetchPage implements PageFetcher interface
func (pf *PostFetcher) FetchPage(ctx context.Context, pageNum int) (*Page[Post], error) {
	params := url.Values{}
	if pf.request.State != "" {
		params.Set("state", pf.request.State)
	}
	for _, state := range pf.request.States {
		params.Add("state[]", state)
	}
	if !pf.request.From.IsZero() {
		params.Set("from", pf.request.From.Format(time.RFC3339))
	}
	if !pf.request.To.IsZero() {
		params.Set("to", pf.request.To.Format(time.RFC3339))
	}
	if pageNum > 0 {
		params.Set("page", strconv.Itoa(pageNum))
	}
	for _, accountID := range pf.request.AccountIDs {
		params.Add("account_ids[]", accountID)
	}
	if pf.request.Query != "" {
		params.Set("query", pf.request.Query)
	}
	if pf.request.PostType != "" {
		params.Set("postType", pf.request.PostType)
	}
	if pf.request.MemberID != "" {
		params.Set("member_id", pf.request.MemberID)
	}

	var response ListPostsResponse
	if err := pf.client.do(ctx, "GET", "posts?"+params.Encode(), nil, &response); err != nil {
		return nil, err
	}

	return &Page[Post]{
		TotalPages: response.TotalPages,
		Items:      response.Posts,
		Total:      response.Total,
		PerPage:    response.PerPage,
		Page:       response.Page,
	}, nil
}

// ListPosts retrieves posts with filtering options
func (c *Client) ListPosts(ctx context.Context, req ListPostsRequest) Iterator[Post] {
	fetcher := &PostFetcher{
		request: req,
		client:  c,
	}
	return NewGenericIterator(fetcher)
}

// PublishPost publishes content immediately
func (c *Client) PublishPost(ctx context.Context, req PublishPostRequest, resp *PublishPostResponse) error {
	return c.do(ctx, "POST", "posts/schedule/publish", req, resp)
}