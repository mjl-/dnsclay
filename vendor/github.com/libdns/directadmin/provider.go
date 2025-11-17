// Package directadmin implements a DNS record management client compatible
// with the libdns interfaces for DirectAdmin.
package directadmin

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"

	"github.com/libdns/libdns"
	"go.uber.org/zap"
)

// Provider facilitates DNS record manipulation with DirectAdmin.
type Provider struct {
	Logger *zap.Logger `json:"-"`
	mutex  sync.Mutex

	// ServerURL should be the hostname (with port if necessary) of the DirectAdmin instance
	// you are trying to use
	ServerURL string `json:"host,omitempty"`

	// User should be the DirectAdmin username that the Login Key is created under
	User string `json:"user,omitempty"`

	// LoginKey is used for authentication
	//
	// The key will need two permissions:
	//
	// `CMD_API_SHOW_DOMAINS` - Required for automatic zone detection
	//
	// `CMD_API_DNS_CONTROL` - Required for DNS record management
	//
	// Both permissions are required for all operations as the provider
	// uses automatic zone detection to handle subdomains correctly
	LoginKey string `json:"login_key,omitempty"`

	// InsecureRequests is an optional parameter used to ignore SSL related errors on the
	// DirectAdmin host
	InsecureRequests bool `json:"insecure_requests,omitempty"`
}

// getLogger returns the logger with caller location context, creating a default one if none is set
func (p *Provider) getLogger() *zap.Logger {
	baseLogger := p.Logger
	if baseLogger == nil {
		baseLogger, _ = zap.NewProduction()
	}
	return baseLogger.With(zap.String("location", p.caller()))
}

func (p *Provider) caller() string {
	pc := make([]uintptr, 15)
	n := runtime.Callers(4, pc) // Fixed skip depth: runtime.Callers -> caller -> getLogger -> actual caller
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	// Extract just the filename from the full path
	parts := strings.Split(frame.File, "/")
	filename := parts[len(parts)-1]
	return fmt.Sprintf("%s:%d", filename, frame.Line)
}

// GetRecords lists all the records in the zone.
func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	p.getLogger().Debug("GetRecords called",
		zap.String("zone", zone))

	managedZone, err := p.findManageableZone(ctx, zone)
	if err != nil {
		return nil, err
	}

	records, err := p.getZoneRecords(ctx, managedZone)
	if err != nil {
		return nil, err
	}

	return records, nil
}

// AppendRecords adds records to the zone. It returns the records that were added.
func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	p.getLogger().Debug("AppendRecords called",
		zap.String("zone", zone),
		zap.Int("record_count", len(records)))

	managedZone, err := p.findManageableZone(ctx, zone)
	if err != nil {
		return nil, err
	}

	if zone != managedZone {
		p.getLogger().Debug("Using managed zone",
			zap.String("managed_zone", managedZone),
			zap.String("requested_zone", zone))
	}

	var created []libdns.Record
	for _, rec := range records {
		// Adjust record name if managedZone differs from requested zone
		adjustedRecord := rec
		if managedZone != strings.TrimSuffix(zone, ".") {
			adjustedRecord = p.adjustRecordForZone(rec, zone, managedZone)
		}

		adjustedRR := adjustedRecord.RR()
		p.getLogger().Debug("Creating record",
			zap.String("name", adjustedRR.Name),
			zap.String("type", adjustedRR.Type),
			zap.String("value", adjustedRR.Data))

		result, err := p.appendZoneRecord(ctx, managedZone, adjustedRecord)
		if err != nil {
			return nil, err
		}
		created = append(created, result)
	}

	return created, nil
}

// SetRecords sets the records in the zone, either by updating existing records or creating new ones.
// It returns the updated records.
func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	p.getLogger().Debug("SetRecords called",
		zap.String("zone", zone),
		zap.Int("record_count", len(records)))

	managedZone, err := p.findManageableZone(ctx, zone)
	if err != nil {
		return nil, err
	}

	if zone != managedZone {
		p.getLogger().Debug("Using managed zone",
			zap.String("managed_zone", managedZone),
			zap.String("requested_zone", zone))
	}

	var updated []libdns.Record
	var errors []error

	for _, rec := range records {
		// Adjust record name if managedZone differs from requested zone
		adjustedRecord := rec
		if managedZone != strings.TrimSuffix(zone, ".") {
			adjustedRecord = p.adjustRecordForZone(rec, zone, managedZone)
		}

		adjustedRR := adjustedRecord.RR()
		p.getLogger().Debug("Creating record",
			zap.String("name", adjustedRR.Name),
			zap.String("type", adjustedRR.Type),
			zap.String("value", adjustedRR.Data))

		result, err := p.setZoneRecord(ctx, managedZone, adjustedRecord)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		updated = append(updated, result)
	}

	if len(errors) > 0 {
		if len(updated) == 0 {
			// No records were updated, return AtomicErr
			return nil, libdns.AtomicErr(fmt.Errorf("all records failed to update: %v", errors))
		}
		// Some records were updated, return a combined error
		return updated, fmt.Errorf("partial update failed: %v", errors)
	}

	return updated, nil
}

// DeleteRecords deletes the records from the zone. It returns the records that were deleted.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	p.getLogger().Debug("DeleteRecords called",
		zap.String("zone", zone),
		zap.Int("record_count", len(records)))

	managedZone, err := p.findManageableZone(ctx, zone)
	if err != nil {
		return nil, err
	}

	if zone != managedZone {
		p.getLogger().Debug("Using managed zone",
			zap.String("managed_zone", managedZone),
			zap.String("requested_zone", zone))
	}

	var deleted []libdns.Record
	for _, rec := range records {
		// Adjust record name if managedZone differs from requested zone
		adjustedRecord := rec
		if managedZone != strings.TrimSuffix(zone, ".") {
			adjustedRecord = p.adjustRecordForZone(rec, zone, managedZone)
		}

		adjustedRR := adjustedRecord.RR()
		p.getLogger().Debug("Deleting record",
			zap.String("name", adjustedRR.Name),
			zap.String("type", adjustedRR.Type),
			zap.String("value", adjustedRR.Data))

		result, err := p.deleteZoneRecord(ctx, managedZone, adjustedRecord)
		if err != nil {
			return nil, err
		}
		deleted = append(deleted, result)
	}

	return deleted, nil
}

// adjustRecordForZone adjusts the record name when the managed zone differs from the requested zone
func (p *Provider) adjustRecordForZone(record libdns.Record, requestedZone, managedZone string) libdns.Record {
	requestedZone = strings.TrimSuffix(requestedZone, ".")
	managedZone = strings.TrimSuffix(managedZone, ".")

	// Calculate the subdomain portion that was stripped during zone detection
	// Example: requestedZone="test.domain.com", managedZone="domain.com" -> subdomain="test"
	if !strings.HasSuffix(requestedZone, managedZone) {
		return record // Safety check - shouldn't happen with proper zone detection
	}

	var subdomain string
	if requestedZone == managedZone {
		subdomain = ""
	} else {
		subdomain = strings.TrimSuffix(requestedZone, "."+managedZone)
	}

	if subdomain == "" {
		return record
	}

	rr := record.RR()

	// Check if the record name has already been adjusted by seeing if it already ends with the subdomain
	if strings.HasSuffix(rr.Name, "."+subdomain) {
		p.getLogger().Debug("Record name already adjusted, skipping",
			zap.String("name", rr.Name),
			zap.String("subdomain", subdomain))
		return record
	}

	// Adjust the record name to include the subdomain
	// Example: "_acme-challenge.libdns" -> "_acme-challenge.libdns.test"
	adjustedName := rr.Name + "." + subdomain

	p.getLogger().Debug("Adjusting record name",
		zap.String("original_name", rr.Name),
		zap.String("adjusted_name", adjustedName),
		zap.String("subdomain", subdomain))

	adjustedRR := &libdns.RR{
		Type: rr.Type,
		Name: adjustedName,
		Data: rr.Data,
		TTL:  rr.TTL,
	}

	return adjustedRR
}

// Interface guards
var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
