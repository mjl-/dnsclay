// Package miab implements a DNS record management client compatible
// with the libdns interfaces for https://mailinabox.email/ custom DNS Endpoints.
// The mailinabox DNS API is limited in that it only works with one zone.
package mailinabox

import (
	"context"
	"fmt"
	"net/netip"
	"strings"

	"github.com/libdns/libdns"
	miab "github.com/luv2code/gomiabdns"
)

// Provider facilitates DNS record manipulation with Mail-In-A-Box.
type Provider struct {
	// APIURL is the URL provided by the mailinabox admin interface, found
	// on your box here: https://box.[your-domain.com]/admin#custom_dns
	// https://box.[your-domain.com]/admin/dns/custom
	APIURL string `json:"api_url,omitempty"`
	// EmailAddress of an admin account.
	// It's recommended that a dedicated account
	// be created especially for managing DNS.
	EmailAddress string `json:"email_address,omitempty"`
	// Password of the admin account that corresponds to the email.
	Password string `json:"password,omitempty"`
	// TOTP Secret Key of the admin account. Needed if multi factor authentication
	// is enabled.
	TOTPSecret string `json:"totp_secret,omitempty"`
}

func (p *Provider) getClient() *miab.Client {
	return miab.New(p.APIURL, p.EmailAddress, p.Password, p.TOTPSecret)
}

func removeTrailingDot(zone string) string {
	if strings.HasSuffix(zone, ".") {
		zone = zone[:len(zone)-1]
	}
	return zone
}
func (p *Provider) zoneCheck(zone string, client *miab.Client) error {
	zone = removeTrailingDot(zone)
	ControlledZones, err := client.GetZones(context.TODO())
	if err != nil {
		return fmt.Errorf("This DNS provider (%s) gives an error (%s) when retrieving zones", p.APIURL, err.Error())
	}

	isOk := false
	for _, dz := range ControlledZones {
		isOk = isOk || strings.Contains(zone, string(dz))
	}

	if !isOk {
		return fmt.Errorf("This DNS provider (%s) does not control the specified zone (%s)", p.APIURL, zone)
	}
	return nil
}
func toLibDnsRecords(zone string, miabRecords []miab.DNSRecord) ([]libdns.Record, error) {
	libDNSRecords := []libdns.Record{}
	zone = removeTrailingDot(zone)
	for _, mr := range miabRecords {
		partialName := strings.ReplaceAll(mr.QualifiedName, zone, "")
		partialName = removeTrailingDot(partialName)
		var rr libdns.Record
		switch mr.RecordType {
		case miab.A, miab.AAAA:
			addr, err := netip.ParseAddr(mr.Value)
			if err != nil {
				return nil, fmt.Errorf("Error parsing IP address. Error: %s, input: %s, record: %s", err, mr.Value, mr.QualifiedName)
			}
			rr = &libdns.Address{Name: partialName, IP: addr}
		case miab.CAA:
			rr = &libdns.CAA{Name: partialName, Value: mr.Value}
		case miab.CNAME:
			rr = &libdns.CNAME{Name: partialName, Target: mr.Value}
		case miab.MX:
			rr = &libdns.MX{Name: partialName, Target: mr.Value}
		case miab.NS:
			rr = &libdns.NS{Name: partialName, Target: mr.Value}
		case miab.SRV:
			rr = &libdns.SRV{Name: partialName, Target: mr.Value}
		case miab.TXT:
			rr = &libdns.TXT{Name: partialName, Text: mr.Value}
		}
		libDNSRecords = append(libDNSRecords, rr)
	}
	return libDNSRecords, nil
}

// GetRecords lists all the records in the zone.
func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	client := p.getClient()
	if err := p.zoneCheck(zone, client); err != nil {
		return nil, err
	}
	miabRecords, err := client.GetHosts(ctx, "", "")
	if err != nil {
		return nil, err
	}
	return toLibDnsRecords(zone, miabRecords)
}

// AppendRecords adds records to the zone. It returns the records that were added.
func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	client := p.getClient()
	if err := p.zoneCheck(zone, client); err != nil {
		return nil, err
	}
	zone = removeTrailingDot(zone)
	for _, r := range records {
		rr := r.RR()
		if err := client.AddHost(ctx, rr.Name+"."+zone, miab.RecordType(rr.Type), rr.Data); err != nil {
			return nil, err
		}
	}
	return records, nil
}

// SetRecords sets the records in the zone, either by updating existing records or creating new ones.
// It returns the updated records.
func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	client := p.getClient()
	if err := p.zoneCheck(zone, client); err != nil {
		return nil, err
	}
	zone = removeTrailingDot(zone)
	for _, r := range records {
		rr := r.RR()
		if err := client.UpdateHost(ctx, rr.Name+"."+zone, miab.RecordType(rr.Type), rr.Data); err != nil {
			return nil, err
		}
	}
	return records, nil
}

// DeleteRecords deletes the records from the zone. It returns the records that were deleted.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	client := p.getClient()
	if err := p.zoneCheck(zone, client); err != nil {
		return nil, err
	}
	zone = removeTrailingDot(zone)
	for _, r := range records {
		rr := r.RR()
		if err := client.DeleteHost(ctx, rr.Name+"."+zone, miab.RecordType(rr.Type), rr.Data); err != nil {
			return nil, err
		}
	}
	return records, nil
}

// Interface guards
var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
