package pdnshttp

import "fmt"

type ErrNotFound struct {
	URL string
}

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("not found: %s", e.URL)
}

type ErrUnexpectedStatus struct {
	URL        string
	StatusCode int
	ErrResponse
}

func (e ErrUnexpectedStatus) Error() string {
	if len(e.ErrResponse.Messages) > 0 {
		return fmt.Sprintf("unexpected status code %d: %s %s %s", e.StatusCode, e.URL, e.ErrResponse.Message, e.ErrResponse.Messages)
	}
	return fmt.Sprintf("unexpected status code %d: %s %s", e.StatusCode, e.URL, e.ErrResponse.Message)
}

// ErrResponse represents error response from PowerDNS HTTP API
type ErrResponse struct {
	Message  string   `json:"error"`
	Messages []string `json:"errors,omitempty"`
}

func IsNotFound(err error) bool {
	switch err.(type) {
	case ErrNotFound:
		return true
	case *ErrNotFound:
		return true
	}

	return false
}
