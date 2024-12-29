package cryptokeys

import (
	"context"
	"fmt"
	"net/url"
)

func (c *client) ListCryptokeys(ctx context.Context, serverID, zoneID string) ([]Cryptokey, error) {
	cryptokeys := []Cryptokey{}
	path := fmt.Sprintf("/servers/%s/zones/%s/cryptokeys", url.PathEscape(serverID), url.PathEscape(zoneID))

	err := c.httpClient.Get(ctx, path, &cryptokeys)
	if err != nil {
		return nil, err
	}

	return cryptokeys, nil
}
