package cryptokeys

import (
	"context"
	"fmt"
	"github.com/mittwald/go-powerdns/pdnshttp"
	"net/url"
)

func (c *client) CreateCryptokey(ctx context.Context, serverID, zoneID string, opts Cryptokey) (*Cryptokey, error) {
	cryptokey := Cryptokey{}
	path := fmt.Sprintf("/servers/%s/zones/%s/cryptokeys",
		url.PathEscape(serverID), url.PathEscape(zoneID))

	err := c.httpClient.Post(ctx, path, &cryptokey, pdnshttp.WithJSONRequestBody(opts))
	if err != nil {
		return nil, err
	}

	return &cryptokey, nil
}
