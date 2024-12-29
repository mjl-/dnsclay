package luadns

import "context"

func (c *Client) Me(ctx context.Context) (*User, error) {
	var user User

	req := func(ctx context.Context) ([]byte, error) {
		return c.client.Get(ctx, c.endpoint("/users/me"))
	}

	err := c.do(ctx, req, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
