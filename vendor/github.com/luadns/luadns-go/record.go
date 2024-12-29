package luadns

import "time"

type RecordType string

const (
	TypeA        = "A"
	TypeAAAA     = "AAAA"
	TypeALIAS    = "ALIAS"
	TypeCAA      = "CAAA"
	TypeCNAME    = "CNAME"
	TypeDS       = "DS"
	TypeFORWARD  = "FORWARD"
	TypeMX       = "MX"
	TypeNS       = "NS"
	TypePTR      = "PTR"
	TypeREDIRECT = "REDIRECT"
	TypeSLAVE    = "SLAVE"
	TypeSOA      = "SOA"
	TypeSPF      = "SPF"
	TypeSRV      = "SRV"
	TypeSSHFP    = "SSHFP"
	TypeTLSA     = "TLSA"
	TypeTXT      = "TXT"
)

type Record struct {
	ID        int64     `json:"id,omitempty"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Content   string    `json:"content"`
	TTL       uint32    `json:"ttl"`
	ZoneID    int64     `json:"zone_id"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}
