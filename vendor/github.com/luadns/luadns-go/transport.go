package luadns

import (
	"net/http"
)

// Transport represents a HTTP transport using Basic authentication and accepts `application/json`.
type Transport struct {
	*http.Transport
	username string
	password string
}

// RoundTrip implements `http.RoundTrip` interface.
func (t Transport) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Set("Accept", jsonMime)
	r.SetBasicAuth(t.username, t.password)
	return t.Transport.RoundTrip(r)
}
