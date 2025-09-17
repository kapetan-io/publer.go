package v1

import (
	"context"
	"fmt"
	"strconv"
)

// ListWorkspacesRequest represents request for listing workspaces
type ListWorkspacesRequest struct{}

// ListWorkspacesResponse represents workspace list response
type ListWorkspacesResponse struct {
	Workspaces []Workspace `json:"workspaces"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	PerPage    int         `json:"per_page"`
	TotalPages int         `json:"total_pages"`
}

// workspacePageFetcher implements PageFetcher for workspaces
type workspacePageFetcher struct {
	client *Client
}

// FetchPage fetches a page of workspaces
func (f *workspacePageFetcher) FetchPage(ctx context.Context, pageNum int) (*Page[Workspace], error) {
	path := "workspaces"
	if pageNum > 1 {
		path = fmt.Sprintf("workspaces?page=%s", strconv.Itoa(pageNum))
	}

	var resp ListWorkspacesResponse
	if err := f.client.do(ctx, "GET", path, nil, &resp); err != nil {
		return nil, err
	}

	return &Page[Workspace]{
		Items:      resp.Workspaces,
		Total:      resp.Total,
		Page:       resp.Page,
		PerPage:    resp.PerPage,
		TotalPages: resp.TotalPages,
	}, nil
}

// ListWorkspaces retrieves all workspaces for the authenticated user
func (c *Client) ListWorkspaces(ctx context.Context, req ListWorkspacesRequest) Iterator[Workspace] {
	fetcher := &workspacePageFetcher{client: c}
	return NewGenericIterator(fetcher)
}