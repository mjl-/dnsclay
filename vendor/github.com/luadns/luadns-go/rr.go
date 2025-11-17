package luadns

// RR represents a DNS resource record.
type RR struct {
	Name    string `json:"name"`
	Type    string `json:"type,omitempty"`
	Content string `json:"content,omitempty"`
	TTL     uint32 `json:"ttl,omitempty"`
}
