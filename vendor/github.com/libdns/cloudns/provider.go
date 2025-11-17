package cloudns

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/libdns/libdns"
)

// ClouDNS API docs: https://www.cloudns.net/wiki/article/41/

// Default configuration values for DNS operations
const (
	// DefaultOperationRetries is the default number of retries for DNS operations
	DefaultOperationRetries = 5

	// DefaultInitialBackoff is the default initial backoff duration for retries
	DefaultInitialBackoff = 1 * time.Second

	// DefaultMaxBackoff is the default maximum backoff duration for retries
	DefaultMaxBackoff = 30 * time.Second
)

// Provider facilitates DNS record manipulation with ClouDNS.
type Provider struct {
	AuthId           string        `json:"auth_id,omitempty"`
	SubAuthId        string        `json:"sub_auth_id,omitempty"`
	AuthPassword     string        `json:"auth_password"`
	OperationRetries int           `json:"operation_retries,omitempty"`
	InitialBackoff   time.Duration `json:"initial_backoff,omitempty"`
	MaxBackoff       time.Duration `json:"max_backoff,omitempty"`
}

// GetRecords lists all the records in the zone.
func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	zone = strings.TrimSuffix(zone, ".")

	// Use retry mechanism for the GetRecords operation
	var records []libdns.Record
	err := RetryWithBackoff(ctx, func() error {
		var e error

		records, e = UseClient(p.AuthId, p.SubAuthId, p.AuthPassword).GetRecords(ctx, zone)
		return e
	}, p.getOperationRetries(), p.getInitialBackoff(), p.getMaxBackoff())
	if err != nil {
		return nil, fmt.Errorf("failed to get records after retries: %w", err)
	}

	return records, nil
}

// AppendRecords adds records to the zone. It returns the records that were added.
func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	zone = strings.TrimSuffix(zone, ".")

	createdRecords := make([]libdns.Record, 0, cap(records))
	for _, record := range records {
		// Use retry mechanism for the AddRecord operation
		var r libdns.Record
		err := RetryWithBackoff(ctx, func() error {
			var err error
			r, err = UseClient(p.AuthId, p.SubAuthId, p.AuthPassword).AddRecord(ctx, zone, fromLibdnsRecord(record, ""))

			return err
		}, p.getOperationRetries(), p.getInitialBackoff(), p.getMaxBackoff())
		if err != nil {
			return nil, fmt.Errorf("failed to add record %q: %w", record.RR().Name, err)
		}

		createdRecords = append(createdRecords, r)
	}

	return createdRecords, nil
}

func (p *Provider) processOperation(ctx context.Context, c *Client, zone string, oplist operationEntry) (libdns.Record, error) {
	var (
		r   libdns.Record
		err error
	)

	switch oplist.op {
	case addRecord:
		err = RetryWithBackoff(ctx, func() error {
			var e error
			r, e = c.AddRecord(ctx, zone, oplist.record)

			return e
		}, p.getOperationRetries(), p.getInitialBackoff(), p.getMaxBackoff())

	case modifyRecord:
		err = RetryWithBackoff(ctx, func() error {
			var e error
			r, e = c.UpdateRecord(ctx, zone, oplist.record)

			return e
		}, p.getOperationRetries(), p.getInitialBackoff(), p.getMaxBackoff())
	case deleteRecord:
		err = RetryWithBackoff(ctx, func() error {
			var e error
			r, e = nil, c.DeleteRecord(ctx, zone, oplist.record.Id)

			return e
		}, p.getOperationRetries(), p.getInitialBackoff(), p.getMaxBackoff())
	default:
		return nil, fmt.Errorf("unknown operation: %v", oplist.op)
	}

	return r, err
}

// SetRecords sets the records in the zone, either by updating existing records or creating new ones.
// ClouDNS does not offer an atomic update, so updates here can leave the zone
// in an inconsistent state upon error. No rollback is attempted.
//
// All updates are attempted, even if an error is encountered. All successfully
// updated records are returned.
func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	zone = strings.TrimSuffix(zone, ".")

	c := UseClient(p.AuthId, p.SubAuthId, p.AuthPassword)
	upstreamRecords, err := c.GetClouDNSRecords(ctx, zone)
	if err != nil {
		return nil, fmt.Errorf("Could not get records for zone %q: %w", zone, err)
	}

	ret := make([]libdns.Record, 0, cap(records))
	var retErr error
	existing := clouDNSRecordsToMap(upstreamRecords)
	rrsets := libdnsRecordsToMap(records)
	oplist := makeOperationList(rrsets, existing)

	for _, op := range oplist {
		rec, err := p.processOperation(ctx, c, zone, op)
		retErr = errors.Join(retErr, err)
		if rec != nil {
			ret = append(ret, rec)
		}
	}

	return ret, retErr
}

func matchDeleteTarget(target, matched libdns.Record) bool {
	matchedRR := matched.RR()
	targetRR := target.RR()

	if targetRR.Type != "" && targetRR.Type != matchedRR.Type {
		return false
	}

	if targetRR.TTL != 0 && targetRR.TTL != matchedRR.TTL {
		return false
	}

	if targetRR.Data != "" && targetRR.Data != matchedRR.Data {
		return false
	}

	return true
}

// DeleteRecords deletes the records from the zone. It returns the records that were deleted.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	zone = strings.TrimSuffix(zone, ".")

	c := UseClient(p.AuthId, p.SubAuthId, p.AuthPassword)
	upstreamRecords, err := c.GetClouDNSRecords(ctx, zone)
	if err != nil {
		return nil, fmt.Errorf("Could not get records for zone %q: %w", zone, err)
	}

	keyedRecords := clouDNSRecordsToMap(upstreamRecords)

	var deletedRecords []libdns.Record
	for _, record := range records {
		rr := record.RR()
		matchingRecords := keyedRecords[nameAndType{name: rr.Name, type_: rr.Type}]
		for _, matchingRecord := range matchingRecords {
			matchedLibdnsRecord, err := matchingRecord.toLibdnsRecord()
			if err != nil {
				return nil, err
			}

			if !matchDeleteTarget(record, matchedLibdnsRecord) {
				continue
			}

			// Use retry mechanism for the DeleteRecord operation
			err = RetryWithBackoff(ctx, func() error {
				return c.DeleteRecord(ctx, zone, matchingRecord.Id)
			}, p.getOperationRetries(), p.getInitialBackoff(), p.getMaxBackoff())
			if err != nil {
				return nil, fmt.Errorf("failed to delete record %q: %w", matchingRecord.Host, err)
			}

			deletedRecords = append(deletedRecords, matchedLibdnsRecord)
		}
	}

	return deletedRecords, nil
}

// Helper methods to get configuration values with defaults

// getOperationRetries returns the configured operation retries or the default value
func (p *Provider) getOperationRetries() int {
	if p.OperationRetries <= 0 {
		return DefaultOperationRetries
	}
	return p.OperationRetries
}

// getInitialBackoff returns the configured initial backoff or the default value
func (p *Provider) getInitialBackoff() time.Duration {
	if p.InitialBackoff <= 0 {
		return DefaultInitialBackoff
	}
	return p.InitialBackoff
}

// getMaxBackoff returns the configured max backoff or the default value
func (p *Provider) getMaxBackoff() time.Duration {
	if p.MaxBackoff <= 0 {
		return DefaultMaxBackoff
	}
	return p.MaxBackoff
}

// Interface guards
var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
