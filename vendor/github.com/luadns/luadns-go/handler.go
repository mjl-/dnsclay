package luadns

import (
	"net/http"
	"strconv"
)

// HandlerFunc represents a response handler.
type HandlerFunc func(*http.Response)

// GetListMeta parses response headers and fills pagination details.
func GetListMeta(meta *ListMeta) HandlerFunc {
	set := func(n *uint64, s string) {
		v, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return
		}
		*n = v
	}

	return func(resp *http.Response) {
		set(&meta.Page, resp.Header.Get("X-Page"))
		set(&meta.Limit, resp.Header.Get("X-Limit"))
		set(&meta.TotalCount, resp.Header.Get("X-Total-Count"))
		set(&meta.PagesCount, resp.Header.Get("X-Pages-Count"))
	}
}
