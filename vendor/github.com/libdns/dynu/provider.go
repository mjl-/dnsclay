// Package dynu implements a DNS record management client compatible
// with the libdns interfaces for dynu.
package dynu

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"strings"
	"sync"
	"time"

	"github.com/libdns/libdns"
)

// Provider facilitates DNS record manipulation with dynu.
type Provider struct {
	// config fields (with snake_case json struct tags on exported fields)
	APIToken  string `json:"api_token,omitempty"`
	OwnDomain string `json:"own_domain,omitempty"`

	Once   sync.Once
	Client *Client
}

func (p *Provider) init() {
	p.Client = NewClient(p.APIToken)
}

// GetRecords lists all the records in the zone.
func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	p.Once.Do(func() { p.init() })

	var libRecords []libdns.Record

	domain := zoneToFqdn(zone)

	// GET /dns/getroot/{hostname}
	dnsHostName, err := p.Client.GetRootDomain(ctx, p.OwnDomain)
	if err != nil {
		return nil, err
	}

	// GET /dns/{id}/record
	dnsRecords, err := p.Client.GetRecords(ctx, dnsHostName.ID)
	if err != nil {
		return nil, err
	}

	for _, dnsRecord := range dnsRecords {
		libRecords = append(libRecords, dnsRecordToLibdnsRecord(dnsRecord, domain))
	}

	return libRecords, nil
}

func dnsRecordToLibdnsRecord(dnsRecord DNSRecord, domain string) libdns.Record {
	var fqdn = dnsRecord.Hostname

	// sub.owndomain.domain.com -> sub.owndomain
	var relativeName string = libdns.RelativeName(fqdn, domain)

	var ttl time.Duration = time.Duration(dnsRecord.TTL) * time.Second

	// store ID and DomainID in ProviderData of specific RR types for efficiency (not available for base RR type)
	dynuProviderData := DynuProviderData{
		ID:       dnsRecord.ID,
		DomainID: dnsRecord.DomainID,
	}

	var libRecord libdns.Record

	switch dnsRecord.Type {
	case "A":
		libRecord = libdns.Address{
			Name:         relativeName,
			TTL:          ttl,
			IP:           netip.MustParseAddr(dnsRecord.Ipv4Address),
			ProviderData: dynuProviderData,
		}
	case "AAAA":
		libRecord = libdns.Address{
			Name:         relativeName,
			TTL:          ttl,
			IP:           netip.MustParseAddr(dnsRecord.Ipv6Address),
			ProviderData: dynuProviderData,
		}
	case "CAA":
		libRecord = libdns.CAA{
			Name:         relativeName,
			TTL:          ttl,
			Flags:        uint8(dnsRecord.Flags),
			Tag:          dnsRecord.Tag,
			Value:        dnsRecord.Value,
			ProviderData: dynuProviderData,
		}
	case "CNAME":
		libRecord = libdns.CNAME{
			Name:         relativeName,
			TTL:          ttl,
			Target:       dnsRecord.Host,
			ProviderData: dynuProviderData,
		}
	case "MX":
		libRecord = libdns.MX{
			Name:         relativeName,
			TTL:          ttl,
			Preference:   uint16(dnsRecord.Priority),
			Target:       dnsRecord.Host,
			ProviderData: dynuProviderData,
		}
	case "NS":
		libRecord = libdns.NS{
			Name:         relativeName,
			TTL:          ttl,
			Target:       dnsRecord.Host,
			ProviderData: dynuProviderData,
		}
	case "PTR":
		libRecord = libdns.RR{
			Type: "PTR",
			Name: dnsRecord.Host,
			TTL:  ttl,
			Data: dnsRecord.Hostname,
		}
	case "SRV":
		service, transportAndName, _ := strings.Cut(dnsRecord.NodeName, ".")
		transport, name, nameFound := strings.Cut(transportAndName, ".")
		if !nameFound {
			name = "@"
		}
		name = libdns.RelativeName(libdns.AbsoluteName(name, dnsRecord.DomainName), domain)
		libRecord = libdns.SRV{
			Service:      strings.TrimPrefix(service, "_"),
			Transport:    strings.TrimPrefix(transport, "_"),
			Name:         name,
			TTL:          ttl,
			Priority:     uint16(dnsRecord.Priority),
			Weight:       uint16(dnsRecord.Weight),
			Port:         uint16(dnsRecord.Port),
			Target:       dnsRecord.Host,
			ProviderData: dynuProviderData,
		}
	case "TXT":
		libRecord = libdns.TXT{
			Name:         relativeName,
			TTL:          ttl,
			Text:         dnsRecord.TextData,
			ProviderData: dynuProviderData,
		}
	default:
		libRecord = libdns.RR{
			Type: dnsRecord.Type,
			Name: relativeName,
			TTL:  ttl,
			Data: dnsRecord.Content,
		}
	}

	return libRecord
}

// AppendRecords adds records to the zone. It returns the records that were added.
func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	return p.appendOrSetRecords(ctx, zone, records, true)
}

// SetRecords sets the records in the zone, either by updating existing records or creating new ones.
// It returns the updated records.
func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	return p.appendOrSetRecords(ctx, zone, records, false)
}

// if appendOnlyMode is true, the records will be added even if record id is provided
func (p *Provider) appendOrSetRecords(ctx context.Context, zone string, records []libdns.Record, appendOnlyMode bool) ([]libdns.Record, error) {
	p.Once.Do(func() { p.init() })

	var updatedRecords []libdns.Record
	var updateErrors []error

	domain := zoneToFqdn(zone)

	// GET /dns/getroot/{hostname}
	dnsHostName, err := p.Client.GetRootDomain(ctx, p.OwnDomain)
	if err != nil {
		return nil, err
	}

	for _, rec := range records {
		dnsRecord, err := libdnsRecordToDnsRecord(rec, domain, p.OwnDomain)
		if err != nil {
			updateErrors = append(updateErrors, err)
			continue
		}

		if !appendOnlyMode {
			// delete existing record(s) before add

			var err error

			if dnsRecord.ID == 0 {
				// record id not available, search and delete ALL records with matching type and name
				err = p.Client.DeleteRecords(ctx, dnsHostName.ID, dnsRecord.Type, dnsRecord.NodeName, "")
			} else {
				// record id is available, delete SINGLE record by existing record id
				err = p.Client.DeleteRecord(ctx, dnsHostName.ID, fmt.Sprint(dnsRecord.ID))
			}

			if err != nil {
				updateErrors = append(updateErrors, fmt.Errorf("failed to delete existing records for %+v: %w", dnsRecord, err))
				continue
			}
		}

		// POST /dns/{id}/record[/{dnsRecordId}]
		updateResponse, err := p.Client.AddRecord(ctx, dnsHostName.ID, dnsRecord)

		if err != nil {
			updateErrors = append(updateErrors, fmt.Errorf("dnsRecord %+v: %w", rec, err))
			continue
		}

		updatedRecords = append(updatedRecords, dnsRecordToLibdnsRecord(*updateResponse, domain))
	}

	return updatedRecords, errors.Join(updateErrors...)
}

func libdnsRecordToDnsRecord(record libdns.Record, domain string, ownDomain string) (DNSRecord, error) {
	var rr = record.RR()

	// sub.owndomain -> sub.owndomain.domain.com -> sub
	var fqdn = libdns.AbsoluteName(rr.Name, domain)
	var relativeName = libdns.RelativeName(fqdn, ownDomain)
	if relativeName == "@" {
		relativeName = ""
	}

	dnsRecord := DNSRecord{
		Type:     rr.Type,
		NodeName: relativeName,
		TTL:      int(rr.TTL.Seconds()),
		State:    true, // must be set to true to take effect
	}

	var err error

	if rr, ok := record.(libdns.RR); ok {
		// if passed in variable is an RR, parse to get the specific type
		record, err = rr.Parse()
		if err != nil {
			return dnsRecord, err
		}
	}

	switch rr := record.(type) {
	case libdns.Address:
		dnsRecord.populateIDsFromProviderData(rr.ProviderData)
		if rr.IP.Is6() {
			dnsRecord.Ipv6Address = rr.IP.String()
		} else {
			dnsRecord.Ipv4Address = rr.IP.String()
		}
	case libdns.CAA:
		dnsRecord.populateIDsFromProviderData(rr.ProviderData)
		dnsRecord.Flags = int(rr.Flags)
		dnsRecord.Tag = rr.Tag
		dnsRecord.Value = rr.Value
	case libdns.CNAME:
		dnsRecord.populateIDsFromProviderData(rr.ProviderData)
		dnsRecord.Host = rr.Target
	case libdns.MX:
		dnsRecord.populateIDsFromProviderData(rr.ProviderData)
		dnsRecord.Host = rr.Target
		dnsRecord.Priority = int(rr.Preference)
	case libdns.NS:
		dnsRecord.populateIDsFromProviderData(rr.ProviderData)
		dnsRecord.Host = rr.Target
	case libdns.SRV:
		dnsRecord.populateIDsFromProviderData(rr.ProviderData)
		dnsRecord.Priority = int(rr.Priority)
		dnsRecord.Weight = int(rr.Weight)
		dnsRecord.Port = int(rr.Port)
		dnsRecord.Host = rr.Target

		serviceTransportName := fmt.Sprintf("_%s._%s", rr.Service, rr.Transport)
		if rr.Name != "@" {
			serviceTransportName = fmt.Sprintf("%s.%s", serviceTransportName, libdns.RelativeName(libdns.AbsoluteName(rr.Name, domain), ownDomain))
		}
		dnsRecord.NodeName = serviceTransportName
	case libdns.TXT:
		dnsRecord.populateIDsFromProviderData(rr.ProviderData)
		dnsRecord.TextData = rr.Text
	default:
		if record.RR().Type == "PTR" {
			dnsRecord.Host = record.RR().Name
			dnsRecord.NodeName = libdns.RelativeName(record.RR().Data, ownDomain) // seems Dynu can only point to subdomain; get relative name from input
		} else {
			err = fmt.Errorf("dnsRecord %+v: record type not implemented", record)
		}
	}

	return dnsRecord, err
}

// DeleteRecords deletes the records from the zone. It returns the records that were deleted.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	p.Once.Do(func() { p.init() })

	var deletedRecords []libdns.Record
	var deleteErrors []error

	// GET /dns/getroot/{hostname}
	dnsHostName, err := p.Client.GetRootDomain(ctx, p.OwnDomain)
	if err != nil {
		return nil, err
	}

	// DELETE /dns/{id}/record/{dnsRecordId}
	for _, rec := range records {
		dnsRecord, err := libdnsRecordToDnsRecord(rec, zoneToFqdn(zone), p.OwnDomain)

		if err != nil {
			deleteErrors = append(deleteErrors, fmt.Errorf("dns record %+v: %w", rec, err))
			continue
		}

		if dnsRecord.ID == 0 {
			// record id is not available, search and delete
			err = p.Client.DeleteRecords(ctx, dnsHostName.ID, dnsRecord.Type, dnsRecord.NodeName, dnsRecord.Content)
		} else {
			// record id is available, delete by record id
			err = p.Client.DeleteRecord(ctx, dnsHostName.ID, fmt.Sprint(dnsRecord.ID))
		}

		if err != nil {
			deleteErrors = append(deleteErrors, fmt.Errorf("dns record %+v: %w", rec, err))
			continue
		}

		deletedRecords = append(deletedRecords, rec)
	}

	return deletedRecords, errors.Join(deleteErrors...)
}

func zoneToFqdn(zone string) string {
	// we trim the dot at the end of the zone name to get the fqdn
	return strings.TrimRight(zone, ".")
}

// Interface guards
var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
