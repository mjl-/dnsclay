package luadns

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	jsonMime = "application/json"              // Accept, Content-Type header
	timeout  = time.Duration(10 * time.Second) // Request timeout
)

// JSONClient represents a REST client using JSON format.
type JSONClient struct {
	client   *http.Client
	username string
	password string
}

// NewJSONClient initializes JSON client.
func NewJSONClient(client *http.Client) *JSONClient {
	return &JSONClient{
		client: client,
	}
}

// NewAuthJSONClient initializes JSON client using Basic authentication.
func NewAuthJSONClient(username, password string) *JSONClient {
	client := &http.Client{
		Transport: &Transport{
			Transport: &http.Transport{},
			username:  username,
			password:  password,
		},
		Timeout: timeout,
	}
	return NewJSONClient(client)
}

// Post executes a POST request using JSON body and returns JSON response.
func (c *JSONClient) Post(ctx context.Context, url string, attrs interface{}, handlers ...HandlerFunc) ([]byte, error) {
	json, err := c.marshalJSON(attrs)
	if err != nil {
		return nil, err
	}
	payload := bytes.NewBuffer(json)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, payload)
	if err != nil {
		return nil, err
	}

	return c.do(req, handlers...)
}

// Get executes a GET request and returns JSON response.
func (c *JSONClient) Get(ctx context.Context, url string, handlers ...HandlerFunc) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, err
	}

	return c.do(req, handlers...)
}

// Put executes a PUT request using JSON body and returns JSON response.
func (c *JSONClient) Put(ctx context.Context, url string, data interface{}, handlers ...HandlerFunc) ([]byte, error) {
	json, err := c.marshalJSON(data)
	if err != nil {
		return nil, err
	}
	payload := bytes.NewBuffer(json)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, payload)
	if err != nil {
		return nil, err
	}

	return c.do(req, handlers...)
}

// Delete executes a DELETE request and returns JSON response.
func (c *JSONClient) Delete(ctx context.Context, url string, handlers ...HandlerFunc) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, http.NoBody)
	if err != nil {
		return nil, err
	}
	return c.do(req, handlers...)
}

// do executes HTTP request, checks for proper response and returns the response body.
func (c *JSONClient) do(req *http.Request, handlers ...HandlerFunc) ([]byte, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = c.checkStatusCode(resp)
	if err != nil {
		return nil, err
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, jsonMime) {
		return nil, &ErrBadContentType{ContentType: contentType}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Run response handlers.
	for _, h := range handlers {
		h(resp)
	}

	return body, nil
}

// checkStatusCode checks the HTTP status code and maps to corresponding error.
func (c *JSONClient) checkStatusCode(resp *http.Response) error {
	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusBadRequest:
		var herr BadRequestError
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		err = json.Unmarshal(body, &herr)
		if err != nil {
			return err
		}
		return &herr
	case http.StatusForbidden:
		var herr ForbiddenRequestError
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		err = json.Unmarshal(body, &herr)
		if err != nil {
			return err
		}
		return &herr
	case http.StatusTooManyRequests:
		limit, err := c.getRatelimitValue(resp, "X-Ratelimit-Limit")
		if err != nil {
			return err
		}
		reset, err := c.getRatelimitValue(resp, "X-Ratelimit-Reset")
		if err != nil {
			return err
		}
		return &ErrTooManyRequests{
			Limit: limit,
			Reset: reset,
		}
	default:
		return &ErrBadStatusCode{StatusCode: resp.StatusCode}
	}
}

// getRatelimitValue parses values from X-Ratelimit-* headers.
func (c *JSONClient) getRatelimitValue(resp *http.Response, key string) (int64, error) {
	value := resp.Header.Get(key)

	n, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, err
	}

	return n, nil
}

func (c *JSONClient) marshalJSON(payload interface{}) ([]byte, error) {
	return json.Marshal(payload)
}
