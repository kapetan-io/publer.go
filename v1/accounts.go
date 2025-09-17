package v1

import (
	"context"
	"fmt"
)

// ListAccountsRequest represents request for listing accounts
type ListAccountsRequest struct{}

// ListAccountsResponse represents account list response
type ListAccountsResponse struct {
	Accounts   []Account `json:"accounts"`
	Total      int       `json:"total"`
	Page       int       `json:"page"`
	PerPage    int       `json:"per_page"`
	TotalPages int       `json:"total_pages"`
}

// accountFetcher implements PageFetcher for accounts
type accountFetcher struct {
	client *Client
	req    ListAccountsRequest
}

// FetchPage implements PageFetcher interface
func (f *accountFetcher) FetchPage(ctx context.Context, pageNum int) (*Page[Account], error) {
	path := "accounts"
	if pageNum > 1 {
		path = fmt.Sprintf("accounts?page=%d", pageNum)
	}

	var resp ListAccountsResponse
	if err := f.client.do(ctx, "GET", path, nil, &resp); err != nil {
		return nil, err
	}

	return &Page[Account]{
		Items:      resp.Accounts,
		Total:      resp.Total,
		Page:       resp.Page,
		PerPage:    resp.PerPage,
		TotalPages: resp.TotalPages,
	}, nil
}

// ListAccounts retrieves all social media accounts in the workspace
func (c *Client) ListAccounts(ctx context.Context, req ListAccountsRequest) Iterator[Account] {
	fetcher := &accountFetcher{
		client: c,
		req:    req,
	}
	return NewGenericIterator[Account](fetcher)
}