package cloudns

// ApiDnsRecord represents a DNS record retrieved from or sent to the API.
// It includes fields for record identification, configuration, and status.
type ApiDnsRecord struct {
	Id       string `json:"id"`
	Type     string `json:"type"`
	Host     string `json:"host"`
	Record   string `json:"record"`
	Failover string `json:"failover"`
	Ttl      string `json:"ttl"`
	Status   int    `json:"status"`
}

// ApiResponse represents the structure of a standard response from the API, including status and optional data.
type ApiResponse struct {
	Status            string `json:"status"`
	StatusDescription string `json:"statusDescription"`
	Data              struct {
		Id int `json:"id"`
	} `json:"data,omitempty"`
}
