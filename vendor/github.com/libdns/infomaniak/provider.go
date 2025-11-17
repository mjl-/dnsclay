package infomaniak

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/libdns/libdns"
)

// Provider facilitates DNS record manipulation with infomaniak.
type Provider struct {
	//infomaniak API token
	APIToken string `json:"api_token,omitempty"`

	//infomaniak client used to call API
	client IkClient

	//mutex to prevent race conditions when initializing client
	mu_client sync.Mutex

	//mutex to prevent race conditions when performing request
	mu_req sync.Mutex
}

// GetRecords returns all the records in the DNS zone.
//
// DNSKEY and DS are DNSSEC-related record types that are included in the output.
func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	zones, err := p.getZoneMapping(ctx, zone)
	if err != nil {
		return []libdns.Record{}, err
	}

	ikRecords, err := p.getRecordsInZone(ctx, zones)
	if err != nil {
		return []libdns.Record{}, err
	}

	libdnsRecords := make([]libdns.Record, 0, len(ikRecords))
	for _, rec := range ikRecords {
		r, err := rec.ToLibDnsRecord(zones)
		if err != nil {
			return []libdns.Record{}, err
		}
		libdnsRecords = append(libdnsRecords, r)
	}

	return libdnsRecords, nil
}

// AppendRecords creates the inputted records in the given zone and returns
// the populated records that were created. It never changes existing records.
//
// Therefore, it makes little sense to use this method with CNAME-type
// records since if there are no existing records with the same name, it
// behaves the same as [libdns.RecordSetter.SetRecords], and if there are
// existing records with the same name, it will either fail or leave the
// zone in an invalid state.
func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	zones, err := p.getZoneMapping(ctx, zone)
	if err != nil {
		return []libdns.Record{}, err
	}

	p.mu_req.Lock()
	defer p.mu_req.Unlock()

	createdRecs := make([]libdns.Record, 0)
	for _, rec := range records {
		createdIkRec, err := p.getClient().CreateOrUpdateRecord(ctx, zones.InfomaniakManagedZone, ToInfomaniakRecord(rec.RR(), zones))
		if err != nil {
			return []libdns.Record{}, err
		}
		createdRec, err := createdIkRec.ToLibDnsRecord(zones)
		if err != nil {
			return []libdns.Record{}, err
		}
		createdRecs = append(createdRecs, createdRec)
	}
	return createdRecs, nil
}

// SetRecords updates the zone so that the records described in the input are
// reflected in the output.
//
// For any (name, type) pair in the input, SetRecords ensures that the only
// records in the output zone with that (name, type) pair are those that were
// provided in the input.
//
// In RFC 9499 terms, SetRecords appends, modifies, or deletes records in the
// zone so that for each RRset in the input, the records provided in the input
// are the only members of their RRset in the output zone.
//
// DNSSEC-related records of type DS are supported.
//
// Calls to SetRecords are not to be presumed atomic; that is, if err == nil,
// then all of the requested changes were made; if err != nil, then some changes
// might have been applied already. In other words, errors may result in partial
// changes to the zone.
//
// If SetRecords is used to add a CNAME record to a name with other existing
// non-DNSSEC records, implementations may either fail with an error, add
// the CNAME and leave the other records in place (in violation of the DNS
// standards), or add the CNAME and remove the other preexisting records.
// Therefore, users should proceed with caution when using SetRecords with
// CNAME records.
func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	zones, err := p.getZoneMapping(ctx, zone)
	if err != nil {
		return []libdns.Record{}, err
	}

	p.mu_req.Lock()
	defer p.mu_req.Unlock()

	recordIdsByCoords, err := p.getExistingRecordIdsByCoordinates(ctx, zones)
	if err != nil {
		return []libdns.Record{}, err
	}

	setRecs := make([]libdns.Record, 0)
	for _, rec := range records {
		rr := rec.RR()
		coords := fmt.Sprintf("%s-%s", libdns.AbsoluteName(rr.Name, zone), rr.Type)
		existingRecordIds := recordIdsByCoords[coords]

		if existingRecordIds != nil {
			for _, id := range existingRecordIds {
				err := p.getClient().DeleteRecord(ctx, zones.InfomaniakManagedZone, id)
				if err != nil {
					return setRecs, err
				}
			}
			recordIdsByCoords[coords] = nil
		}

		updatedIkRec, err := p.getClient().CreateOrUpdateRecord(ctx, zones.InfomaniakManagedZone, ToInfomaniakRecord(rec, zones))
		if err != nil {
			return setRecs, err
		}

		setRec, err := updatedIkRec.ToLibDnsRecord(zones)
		if err != nil {
			return setRecs, err
		}
		setRecs = append(setRecs, setRec)
	}
	return setRecs, nil
}

// getExistingRecordIdsByCoordinates returns the existing records in this zone by their fqdn-type
func (p *Provider) getExistingRecordIdsByCoordinates(ctx context.Context, zones *ZoneMapping) (map[string][]string, error) {
	records, err := p.getRecordsInZone(ctx, zones)
	if err != nil {
		return nil, err
	}

	result := make(map[string][]string)
	for _, rec := range records {
		coordinates := fmt.Sprintf("%s-%s", libdns.AbsoluteName(rec.Source, zones.InfomaniakManagedZone), rec.Type)
		recordsWithSameCoordinates := result[coordinates]
		if recordsWithSameCoordinates == nil {
			recordsWithSameCoordinates = make([]string, 0)
		}
		recordsWithSameCoordinates = append(recordsWithSameCoordinates, strconv.Itoa(rec.ID))
		result[coordinates] = recordsWithSameCoordinates
	}
	return result, nil
}

// DeleteRecords deletes the given records from the zone if they exist in the
// zone and exactly match the input. If the input records do not exist in the
// zone, they are silently ignored. DeleteRecords returns only the the records
// that were deleted, and does not return any records that were provided in the
// input but did not exist in the zone.
//
// DeleteRecords only deletes records from the zone that *exactly* match the
// input records—that is, the name, type, TTL, and value all must be identical
// to a record in the zone for it to be deleted.
//
// As a special case, you may leave any of the fields [libdns.Record.Type],
// [libdns.Record.TTL], or [libdns.Record.Value] empty ("", 0, and ""
// respectively). In this case, DeleteRecords will delete any records that
// match the other fields, regardless of the value of the fields that were left
// empty. Note that this behavior does *not* apply to the [libdns.Record.Name]
// field, which must always be specified.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	zones, err := p.getZoneMapping(ctx, zone)
	if err != nil {
		return []libdns.Record{}, err
	}

	p.mu_req.Lock()
	defer p.mu_req.Unlock()

	existingRecs, err := p.getRecordsInZone(ctx, zones)
	if err != nil {
		return []libdns.Record{}, err
	}

	deletedRecs := make([]libdns.Record, 0)
	for _, recToDelete := range records {
		rrToDelete := recToDelete.RR()
		remainingRecs := make([]IkRecord, 0)
		for _, existingRec := range existingRecs {
			if !isDeleteRecord(zones, &rrToDelete, &existingRec) {
				remainingRecs = append(remainingRecs, existingRec)
			} else {
				resultRec, err := existingRec.ToLibDnsRecord(zones)
				if err != nil {
					return deletedRecs, err
				}

				err = p.getClient().DeleteRecord(ctx, zones.InfomaniakManagedZone, strconv.Itoa(existingRec.ID))
				if err != nil {
					return deletedRecs, err
				}
				deletedRecs = append(deletedRecs, resultRec)
			}
		}
		existingRecs = remainingRecs
	}
	return deletedRecs, nil
}

// isDeleteRecord returns true when the existing record matches the record that should be deleted
func isDeleteRecord(zoneMapping *ZoneMapping, rrToDelete *libdns.RR, existingRec *IkRecord) bool {
	return libdns.AbsoluteName(rrToDelete.Name, zoneMapping.LibDnsZone) == libdns.AbsoluteName(existingRec.Source, zoneMapping.InfomaniakManagedZone) &&
		(rrToDelete.TTL == 0 || rrToDelete.TTL == existingRec.getTtlAsTimeDuration()) &&
		(rrToDelete.Type == "" || rrToDelete.Type == existingRec.Type) &&
		(rrToDelete.Data == "" || rrToDelete.Data == existingRec.Target)
}

// getClient returns a new instance of the infomaniak API client
func (p *Provider) getClient() IkClient {
	p.mu_client.Lock()
	defer p.mu_client.Unlock()
	if p.client == nil {
		p.client = &Client{Token: p.APIToken, HttpClient: http.DefaultClient}
	}
	return p.client
}

// getZoneMapping returns the DNS zone that is mangaged by infomaniak and the input zone
// from the libdns caller without a trailing dot
func (p *Provider) getZoneMapping(ctx context.Context, zone string) (*ZoneMapping, error) {
	libdnsZone := strings.TrimSuffix(zone, ".")
	infomaniakZone, err := p.getClient().GetFqdnOfZoneForDomain(ctx, libdnsZone)
	if err != nil {
		return nil, err
	}
	return &ZoneMapping{
		InfomaniakManagedZone: infomaniakZone,
		LibDnsZone:            libdnsZone,
	}, nil
}

// getRecordsInZone returns the records that are in theinput zone from the libdns caller
func (p *Provider) getRecordsInZone(ctx context.Context, zones *ZoneMapping) ([]IkRecord, error) {
	recs, err := p.getClient().GetDnsRecordsForZone(ctx, zones.InfomaniakManagedZone)
	if err != nil {
		return []IkRecord{}, err
	}

	result := make([]IkRecord, 0)
	for _, rec := range recs {
		if strings.HasSuffix(libdns.AbsoluteName(rec.Source, zones.InfomaniakManagedZone), zones.LibDnsZone) {
			result = append(result, rec)
		}
	}
	return result, nil
}

// Interface guards
var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
