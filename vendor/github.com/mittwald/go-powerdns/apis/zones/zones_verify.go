package zones

import (
	"context"
	"fmt"
	"net/url"
)

func (c *client) VerifyZone(ctx context.Context, serverID string, zoneID string) error {
	path := fmt.Sprintf("/servers/%s/zones/%s/check", url.PathEscape(serverID), url.PathEscape(zoneID))

	return c.httpClient.Get(ctx, path, nil)
}
