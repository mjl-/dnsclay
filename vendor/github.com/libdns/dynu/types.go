package dynu

import "fmt"

type APIException struct {
	StatusCode int32  `json:"statusCode,omitempty"`
	Type       string `json:"type,omitempty"`
	Message    string `json:"message,omitempty"`
}

func (a APIException) Error() string {
	return fmt.Sprintf("%d: %s: %s", a.StatusCode, a.Type, a.Message)
}

type DNSRecord struct {
	ID          int64  `json:"id,omitempty"`
	Type        string `json:"recordType,omitempty"`
	DomainID    int64  `json:"domainId,omitempty"`
	DomainName  string `json:"domainName,omitempty"`
	NodeName    string `json:"nodeName,omitempty"`
	Hostname    string `json:"hostname,omitempty"`
	State       bool   `json:"state,omitempty"`
	Content     string `json:"content,omitempty"`
	Ipv4Address string `json:"ipv4Address,omitempty"`
	Ipv6Address string `json:"ipv6Address,omitempty"`
	Host        string `json:"host,omitempty"`
	TextData    string `json:"textData,omitempty"`
	TTL         int    `json:"ttl,omitempty"`
	Priority    int    `json:"priority,omitempty"`
	StatusCode  int32  `json:"statusCode,omitempty"`
}

type DNSHostname struct {
	StatusCode int32  `json:"statusCode,omitempty"`
	ID         int64  `json:"id,omitempty"`
	DomainName string `json:"domainName,omitempty"`
	Hostname   string `json:"hostname,omitempty"`
	Node       string `json:"node,omitempty"`
}

type RecordsResponse struct {
	StatusCode int32       `json:"statusCode,omitempty"`
	DNSRecords []DNSRecord `json:"dnsRecords,omitempty"`
}

type UpdateResponse struct {
	StatusCode int32 `json:"statusCode,omitempty"`
	ID         int64 `json:"id,omitempty"`
}

type DeleteResponse struct {
	StatusCode int32 `json:"statusCode,omitempty"`
}
