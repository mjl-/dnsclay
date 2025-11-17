package infomaniak

import (
	"context"
	"encoding/json"
)

// IkRecord infomaniak API record return type
type IkRecord struct {
	// ID of this record on infomaniak's side
	ID int `json:"id,omitempty"`

	// Type of this record
	Type string `json:"type,omitempty"`

	// Absolute Source / Name
	Source string `json:"source,omitempty"`

	// Value of this record
	Target string `json:"target,omitempty"`

	// TTL in seconds
	TtlInSec int `json:"ttl,omitempty"`

	// Record Description
	Description IkRecordDescription `json:"description,omitempty"`
}

type IkRecordDescription struct {
	// Only available for SRV and MX records
	Priority IkIntValueAttribute `json:"priority,omitempty"`

	// Only available for SRV records
	Port IkIntValueAttribute `json:"port,omitempty"`

	// Only available for SRV records
	Weight IkIntValueAttribute `json:"weight,omitempty"`

	// Only available for SRV and DNSKEY records
	Protocol IkStringValueAttribute `json:"protocol,omitempty"`

	// Only available for CAA and DNSKEY records
	Flags IkIntValueAttribute `json:"flags,omitempty"`

	// Only available for CAA records
	Tag IkStringValueAttribute `json:"tag,omitempty"`
}

type IkIntValueAttribute struct {
	// Attribute value
	Value int `json:"value,omitempty"`
	// Human readable value of attribute
	Label string `json:"label,omitempty"`
}

type IkStringValueAttribute struct {
	// Attribute value
	Value string `json:"value,omitempty"`
	// Human readable value of attribute
	Label string `json:"label,omitempty"`
}

// IkResponse infomaniak API response
type IkResponse struct {
	// Result of the API call: either "success" or "error"
	Result string `json:"result"`

	// Data is set if API call was successful and contains the actual response
	Data json.RawMessage `json:"data,omitempty"`

	// Error is set if the API call failed and contains all errors that occurred
	Error json.RawMessage `json:"error,omitempty"`
}

// IkZone infomaniak API zone return type
type IkZone struct {
	// Zone's FQDN on infomaniak's side
	Fqdn string `json:"fqdn"`
}

// ZoneMapping represents input zone coming from the caller and the zone
// that is actually managed by infomaniak
type ZoneMapping struct {
	// Zone that is mangaged by infomaniak
	InfomaniakManagedZone string `json:"fqdn"`

	// Zone that is provided by libdns
	LibDnsZone string
}

// IkClient interface to abstract infomaniak client
type IkClient interface {
	// DeleteRecord deletes record with given ID
	DeleteRecord(ctx context.Context, zone string, id string) error

	// CreateOrUpdateRecord creates record if it has no ID property set, otherwise it updates the record with the given ID
	CreateOrUpdateRecord(ctx context.Context, zone string, record IkRecord) (*IkRecord, error)

	// GetDnsRecordsForZone returns all records of the given zone
	GetDnsRecordsForZone(ctx context.Context, zone string) ([]IkRecord, error)

	// GetFqdnOfZoneForDomain returns the FQDN of the zone managed by infomaniak
	GetFqdnOfZoneForDomain(ctx context.Context, domain string) (string, error)
}
