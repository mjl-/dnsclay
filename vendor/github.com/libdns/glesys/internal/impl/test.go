package impl

import (
	"context"
	"encoding/json"
)

type mockClient struct {
	body       string
	lastPath   string
	lastMethod string
}

func (c *mockClient) get(ctx context.Context, path string, v interface{}) error {
	c.lastPath = path
	c.lastMethod = "GET"
	return json.Unmarshal([]byte(c.body), v)
}

func (c *mockClient) post(ctx context.Context, path string, v interface{}, params interface{}) error {
	c.lastPath = path
	c.lastMethod = "POST"
	return json.Unmarshal([]byte(c.body), v)
}
