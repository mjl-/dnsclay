package cryptokeys

// Cryptokey represents a Cryptokey model of the API
// More information: https://doc.powerdns.com/authoritative/http-api/cryptokey.html#cryptokey
type Cryptokey struct {
	ID         int      `json:"id,omitempty"`
	Type       string   `json:"type,omitempty"`
	KeyType    string   `json:"keytype,omitempty"`
	Active     bool     `json:"active,omitempty"`
	Published  bool     `json:"published,omitempty"`
	DNSKey     string   `json:"dnskey,omitempty"`
	DS         []string `json:"ds,omitempty"`
	PrivateKey string   `json:"privatekey,omitempty"`
	Algorithm  string   `json:"algorithm,omitempty"`
	Bits       int      `json:"bits,omitempty"`
}
