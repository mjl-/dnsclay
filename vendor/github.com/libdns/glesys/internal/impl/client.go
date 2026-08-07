// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2017 GleSYS Internet Services AB

package impl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const version = "8.1.0"

type httpClientInterface interface {
	Do(*http.Request) (*http.Response, error)
}

type clientInterface interface {
	get(ctx context.Context, path string, v interface{}) error
	post(ctx context.Context, path string, v interface{}, params interface{}) error
}

// Client is used to interact with the GleSYS API
type Client struct {
	apiKey     string
	BaseURL    *url.URL
	httpClient httpClientInterface
	project    string
	userAgent  string

	DNSDomains *DNSDomainService
}

// NewClient creates a new Client for interacting with the GleSYS API. This is
// the main entrypoint for API interactions.
func NewClient(project, apiKey, userAgent string) *Client {
	BaseURL, _ := url.Parse("https://api.glesys.com")

	c := &Client{
		apiKey:     apiKey,
		BaseURL:    BaseURL,
		httpClient: http.DefaultClient,
		project:    project,
		userAgent:  userAgent,
	}

	c.DNSDomains = &DNSDomainService{client: c}
	return c
}

// SetBaseURL can be used to set a custom BaseURL
func (c *Client) SetBaseURL(bu string) error {
	url, err := url.Parse(bu)
	if err != nil {
		return err
	}
	c.BaseURL = url
	return nil
}

func (c *Client) get(ctx context.Context, path string, v interface{}) error {
	request, err := c.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return err
	}
	return c.do(request, v)
}

func (c *Client) post(ctx context.Context, path string, v interface{}, params interface{}) error {
	request, err := c.newRequest(ctx, "POST", path, params)
	if err != nil {
		return err
	}
	return c.do(request, v)
}

func (c *Client) newRequest(ctx context.Context, method, path string, params interface{}) (*http.Request, error) {
	u, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	if c.BaseURL != nil {
		u = c.BaseURL.ResolveReference(u)
	}

	buffer := new(bytes.Buffer)

	if params != nil {
		err = json.NewEncoder(buffer).Encode(params)
		if err != nil {
			return nil, err
		}
	}

	request, err := http.NewRequestWithContext(ctx, method, u.String(), buffer)
	if err != nil {
		return nil, err
	}

	userAgent := strings.TrimSpace(fmt.Sprintf("%s glesys-go/%s", c.userAgent, version))

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", userAgent)
	request.SetBasicAuth(c.project, c.apiKey)

	return request, nil
}

func (c *Client) do(request *http.Request, v interface{}) error {
	response, err := c.httpClient.Do(request)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusOK {
		return handleResponseError(response)
	}

	return parseResponseBody(response, v)
}

func handleResponseError(response *http.Response) error {
	data := struct {
		Response struct {
			Status struct {
				Text string `json:"text"`
			} `json:"status"`
		} `json:"response"`
	}{}

	err := parseResponseBody(response, &data)
	if err != nil {
		return err
	}

	return fmt.Errorf("request failed with HTTP error: %v (%v)", response.StatusCode, strings.TrimSpace(data.Response.Status.Text))
}

func parseResponseBody(response *http.Response, v interface{}) error {
	if v == nil {
		return nil
	}

	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, v)
}
