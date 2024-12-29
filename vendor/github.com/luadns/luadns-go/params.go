package luadns

import (
	"net/url"
	"strconv"
)

// ListParams represents list options accepted by a list endpoints.
type ListParams struct {
	Query     string // q=example.com
	SortBy    string // sort_by=name
	SortOrder string // sort_order=asc/desc
	Limit     uint64 // limit=10
	Page      uint64 // page=1
}

// QueryString convert the list options to a query string.
func (opts *ListParams) QueryString() string {
	values := url.Values{}

	if opts.Query != "" {
		values.Set("query", opts.Query)
	}

	if opts.SortBy != "" {
		values.Set("sort_by", opts.SortBy)
	}

	if opts.SortOrder != "" {
		values.Set("sort_order", opts.SortOrder)
	}

	if opts.Limit != 0 {
		values.Set("limit", strconv.FormatUint(opts.Limit, 10))
	}

	if opts.Page != 0 {
		values.Set("page", strconv.FormatUint(opts.Page, 10))
	}

	return values.Encode()
}
