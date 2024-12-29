package cryptokeys

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

func (c *client) DeleteCryptokey(ctx context.Context, serverID, zoneID string, cryptokeyID int) error {
	path := fmt.Sprintf("/servers/%s/zones/%s/cryptokeys/%s",
		url.PathEscape(serverID), url.PathEscape(zoneID), url.PathEscape(strconv.Itoa(cryptokeyID)))

	return c.httpClient.Delete(ctx, path, nil)
}
