package leaseweb

// Structs for easy json marshalling.
// Only declare fields that are used.
type leasewebRecordSet struct {
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	Content []string `json:"content"`
	TTL     int      `json:"ttl"`
}

type leasewebRecordSets struct {
	ResourceRecordSets []leasewebRecordSet `json:"resourceRecordSets"`
}

// Errors
type leasewebHttpError struct {
	ErrorMessage  string `json:"errorMessage"`
	UserMessage   string `json:"userMessage"`
	CorrelationId string `json:"correlationId"`
	// Its not a string but an json object.
	ErrorDetails string `json:"errorDetails"`
}

// updateRecordSet
// https://developer.leaseweb.com/api-docs/domains_v2.html#operation/put/domains/{domainName}/resourceRecordSets/{name}/{type}
type updateRecordSetRequest struct {
	Content []string `json:"content"`
	TTL     int      `json:"ttl"`
}
