package cryptokeys

import "context"

// Client defines method for interacting with the PowerDNS "Cryptokeys" endpoints
type Client interface {

	// ListCryptokeys lists all CryptoKeys, except its privatekey
	ListCryptokeys(ctx context.Context, serverID, zoneID string) ([]Cryptokey, error)

	// GetCryptokey returns all data about the CryptoKey, including the privatekey.
	// If the server with the given "serverID" does not exist,
	// the error return value will contain a pdnshttp.ErrNotFound error (see example)
	GetCryptokey(ctx context.Context, serverID, zoneID string, cryptokeyID int) (*Cryptokey, error)

	// CreateCryptokey creates a new CryptoKey
	CreateCryptokey(ctx context.Context, serverID, zoneID string, opts Cryptokey) (*Cryptokey, error)

	// ToggleCryptokey (de)activates a CryptoKey for the given zone
	ToggleCryptokey(ctx context.Context, serverID, zoneID string, cryptokeyID int) error

	// DeleteCryptokey deletes a CryptoKey from the given zone
	DeleteCryptokey(ctx context.Context, serverID, zoneID string, cryptokeyID int) error
}
