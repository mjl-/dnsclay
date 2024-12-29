package servers

import "context"

func (c *client) ListServers(ctx context.Context) ([]Server, error) {
	servers := make([]Server, 0)

	err := c.httpClient.Get(ctx, "/servers", &servers)
	if err != nil {
		return nil, err
	}

	return servers, nil
}
