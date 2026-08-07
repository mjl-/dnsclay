// libdns implementation for IONOS DNS API.
// IONOS API documentaion: https://developer.hosting.ionos.de/docs/dns
// libdns: https://github.com/libdns/libdns
package ionos

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/libdns/libdns"
)

// Provider implements the libdns interfaces for IONOS
type Provider struct {
	// AuthAPIToken is the IONOS Auth API token -
	// see https://dns.ionos.com/api-docs#section/Authentication/Auth-API-Token
	AuthAPIToken string `json:"auth_api_token"`
}

func toIonosRecord(r libdns.Record, zoneName string) record {
	return record{
		Type:    r.Type,
		Name:    libdns.AbsoluteName(r.Name, zoneName),
		Content: r.Value,
		TTL:     ionosTTL(r.TTL.Seconds()),
	}
}

func fromIonosRecord(r zoneRecord, zoneName string) libdns.Record {
	// IONOS returns TXT records quoted: remove quotes
	var value string
	if strings.ToUpper(r.Type) == "TXT" {
		value, _ = strconv.Unquote(r.Content)
	} else {
		value = r.Content
	}
	return libdns.Record{
		ID:   r.ID,
		Type: r.Type,
		// libdns Name is partially qualified, relative to zone, Ionos absoulte
		Name:  libdns.RelativeName(r.Name, zoneName), // use r.rootName for zoneName TODO?
		Value: value,
		TTL:   time.Duration(r.TTL) * time.Second,
	}
}

func (p *Provider) findZoneByName(ctx context.Context, zoneName string) (zoneDescriptor, error) {
	// obtain list of all zones
	zones, err := ionosGetAllZones(ctx, p.AuthAPIToken)
	if err != nil {
		return zoneDescriptor{}, fmt.Errorf("get all zones: %w", err)
	}

	// find the desired zone
	for _, zone := range zones.Zones {
		if zone.Name == unFQDN(zoneName) {
			return zone, nil
		}
	}
	return zoneDescriptor{}, fmt.Errorf("zone named not found (%s)", zoneName)
}

// GetRecords lists all the records in the zone.
func (p *Provider) GetRecords(ctx context.Context, zoneName string) ([]libdns.Record, error) {

	zoneDes, err := p.findZoneByName(ctx, zoneName)
	if err != nil {
		return nil, fmt.Errorf("find zone: %w", err)
	}

	// obtain list of all records in zone
	zoneResp, err := ionosGetZone(ctx, p.AuthAPIToken, zoneDes.ID, "", "")
	if err != nil {
		return nil, fmt.Errorf("get zone records: %w", err)
	}

	records := make([]libdns.Record, len(zoneResp.Records))
	for i, r := range zoneResp.Records {
		records[i] = fromIonosRecord(r, zoneName)
	}
	return records, nil
}

// AppendRecords adds records to the zone. It returns the records that were added.
func (p *Provider) AppendRecords(
	ctx context.Context,
	zone string,
	records []libdns.Record) ([]libdns.Record, error) {

	zoneDes, err := p.findZoneByName(ctx, zone)
	if err != nil {
		return nil, fmt.Errorf("find zone: %w", err)
	}

	// populate ionos request
	reqs := make([]record, len(records))
	for i, r := range records {
		reqs[i] = toIonosRecord(r, zoneDes.Name)
	}

	newRecords, err := ionosCreateRecords(ctx, p.AuthAPIToken, zoneDes.ID, reqs)
	if err != nil {
		return nil, fmt.Errorf("create records: %w", err)
	}

	// populate libdns response
	res := make([]libdns.Record, len(records))
	for i, r := range newRecords {
		res[i] = fromIonosRecord(r, zoneDes.Name)
	}
	return res, nil
}

// DeleteRecords deletes the records from the zone. Returns the list of
// records acutally deleted. Fails fast on first error, but in this case
func (p *Provider) DeleteRecords(
	ctx context.Context,
	zone string,
	records []libdns.Record) ([]libdns.Record, error) {

	zoneDes, err := p.findZoneByName(ctx, zone)
	if err != nil {
		return nil, fmt.Errorf("find zone: %w", err)
	}

	// ionos api has no batch-delete, delete one record at a time
	var deleteQueue []libdns.Record // list of record IDs to delete

	// collect IDs
	for _, r := range records {
		// no ID provided, search record first
		if r.ID == "" {
			// safety: avoid to delete the whole zone
			if r.Type == "" || r.Name == "" {
				continue
			}

			name := libdns.AbsoluteName(r.Name, zoneDes.Name)
			existing, err := ionosFindRecordsInZone(ctx, p.AuthAPIToken, zoneDes.ID, r.Type, name)
			if err != nil {
				return nil, fmt.Errorf("find record for deletion: %w", err)
			}
			for _, found := range existing {
				deleteQueue = append(deleteQueue, fromIonosRecord(found, zoneDes.Name))
			}
		} else {
			deleteQueue = append(deleteQueue, r)
		}
	}
	// delete all collected records
	for _, r := range deleteQueue {
		err := ionosDeleteRecord(ctx, p.AuthAPIToken, zoneDes.ID, r.ID)
		if err != nil {
			return nil, fmt.Errorf("delete record by ID: %w", err)
		}
	}

	return deleteQueue, nil
}

func (p *Provider) createOrUpdateRecord(
	ctx context.Context,
	zoneDes zoneDescriptor,
	r libdns.Record) (libdns.Record, error) {

	// an ID is provided, we can directly call the ionos api
	if r.ID != "" {
		err := ionosUpdateRecord(ctx, p.AuthAPIToken, zoneDes.ID, r.ID, toIonosRecord(r, zoneDes.Name))
		if err != nil {
			return r, fmt.Errorf("update record: %w", err)
		}
		return r, nil
	}

	// before we create a new record, make sure there is no existing record
	// of same (type, name). In this case we only update the record
	name := libdns.AbsoluteName(r.Name, zoneDes.Name)
	existing, err := ionosFindRecordsInZone(ctx, p.AuthAPIToken, zoneDes.ID, r.Type, name)
	if err == nil {
		if len(existing) != 1 {
			return r, fmt.Errorf("found unexpected number of records during delete, expected 1 (%d)", len(existing))
		}
		err := ionosUpdateRecord(ctx, p.AuthAPIToken, zoneDes.ID, existing[0].ID, toIonosRecord(r, zoneDes.Name))
		if err != nil {
			return r, fmt.Errorf("update found record: %w", err)
		}
		r.ID = existing[0].ID
		return r, nil
	}

	created, err := ionosCreateRecords(ctx, p.AuthAPIToken, zoneDes.ID, []record{toIonosRecord(r, zoneDes.Name)})
	if err != nil {
		return r, fmt.Errorf("create new record: %w", err)
	}
	if len(created) != 1 {
		return r, fmt.Errorf("expected one record to be created, got %d", len(created))
	}
	return fromIonosRecord(created[0], zoneDes.Name), nil
}

// SetRecords sets the records in the zone, either by updating existing records
// or creating new ones. It returns the updated records.
func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	var res []libdns.Record

	zoneDes, err := p.findZoneByName(ctx, zone)
	if err != nil {
		return nil, fmt.Errorf("find zone: %w", err)
	}

	for _, r := range records {
		newRecord, err := p.createOrUpdateRecord(ctx, zoneDes, r)
		if err != nil {
			return res, err
		}
		res = append(res, newRecord)
	}
	return res, nil
}

// unFQDN trims any trailing "." from fqdn. IONOS's API does not use FQDNs.
func unFQDN(fqdn string) string {
	return strings.TrimSuffix(fqdn, ".")
}

// Interface guards
var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
