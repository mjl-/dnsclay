package cryptokeys

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

func (c *client) ToggleCryptokey(ctx context.Context, serverID, zoneID string, cryptokeyID int) error {
	path := fmt.Sprintf("/servers/%s/zones/%s/cryptokeys/%s",
		url.PathEscape(serverID), url.PathEscape(zoneID), url.PathEscape(strconv.Itoa(cryptokeyID)))

	return c.httpClient.Put(ctx, path, nil)
}
