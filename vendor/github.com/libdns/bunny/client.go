package bunny

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/libdns/libdns"
	"golang.org/x/net/publicsuffix"
)

type getAllRecordsResponse struct {
	Records []bunnyRecord `json:"Records"`
}

type getAllZonesResponse struct {
	Zones []bunnyZone `json:"Items"`
}

type bunnyZone struct {
	ID            int    `json:"Id"`
	Domain        string `json:"Domain"`
	DnsSecEnabled bool   `json:"DnsSecEnabled"`
	nameBase      string `json:"-"`
}

type bunnyRecord struct {
	ID       int    `json:"Id,omitempty"`
	Type     int    `json:"Type"`
	TTL      int    `json:"Ttl"`
	Value    string `json:"Value"`
	Name     string `json:"Name"`
	Weight   int32  `json:"Weight,omitempty"`
	Priority int32  `json:"Priority,omitempty"`
	Flags    int    `json:"Flags,omitempty"`
	Tag      string `json:"Tag,omitempty"`
	Port     int32  `json:"Port,omitempty"`
}

func (p *Provider) doRequest(request *http.Request) ([]byte, error) {
	request.Header.Add("accept", "application/json")
	request.Header.Add("AccessKey", p.AccessKey)

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("%s (%d)", http.StatusText(response.StatusCode), response.StatusCode)
	}

	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (p *Provider) getAllZones(ctx context.Context) ([]bunnyZone, error) {
	p.log("fetching all zones")

	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.bunny.net/dnszone", nil)
	if err != nil {
		return nil, err
	}

	data, err := p.doRequest(req)
	if err != nil {
		return nil, err
	}

	result := getAllZonesResponse{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	p.log(fmt.Sprintf("retrieved %d zone(s)", len(result.Zones)))

	return result.Zones, nil
}

func (p *Provider) getZone(ctx context.Context, domain string) (bunnyZone, error) {
	p.zonesMu.Lock()
	defer p.zonesMu.Unlock()

	if domain == "" {
		return bunnyZone{}, fmt.Errorf("domain is an empty string")
	}

	// If we already got the zone info, reuse it
	if p.zones == nil {
		p.zones = make(map[string]bunnyZone)
	}
	if zone, ok := p.zones[domain]; ok {
		return zone, nil
	}

	p.log(fmt.Sprintf("fetching zone for %s", domain))

	zone, err := publicsuffix.EffectiveTLDPlusOne(domain)
	if err != nil {
		zone = domain
	}

	// The API can only return up to 1000 records. So we need to search for the
	// apex domain to be safe and then filter from there to get an exact result.
	req, err := http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf("https://api.bunny.net/dnszone?search=%s", url.QueryEscape(zone)), nil)
	if err != nil {
		return bunnyZone{}, err
	}

	data, err := p.doRequest(req)
	if err != nil {
		return bunnyZone{}, err
	}

	result := getAllZonesResponse{}
	if err := json.Unmarshal(data, &result); err != nil {
		return bunnyZone{}, err
	}

	// Get all possible parent domains to check
	zoneGuesses := getBaseDomainNameGuesses(domain)

	// Iterate through domain guesses (most specific to least specific)
	for _, zoneGuess := range zoneGuesses {
		for _, zone := range result.Zones {
			if strings.EqualFold(zone.Domain, zoneGuess) {
				if len(domain) > len(zone.Domain) {
					zone.nameBase = strings.ToLower(domain[:len(domain)-len(zone.Domain)-1])
					p.log(fmt.Sprintf("found zone ID %d (%s) for %s",
						zone.ID, zone.Domain, domain))
				} else {
					p.log(fmt.Sprintf("found zone ID %d for %s",
						zone.ID, domain))
				}

				// cache this zone for possible reuse
				p.zones[domain] = zone

				return zone, nil
			}
		}
	}

	return bunnyZone{}, fmt.Errorf("zone not found for domain: %s", zone)
}

// getBaseDomainNameGuesses returns a slice of possible parent domain names
// ordered from most specific to least specific. For a domain "sub.example.com",
// it returns ["sub.example.com", "example.com", "com"]. If the input domain has
// fewer than 2 parts (no dots or just one dot), it returns a slice containing
// only the input domain.
func getBaseDomainNameGuesses(domain string) []string {
	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return []string{domain}
	}

	var guesses []string
	for i := 0; i < len(parts)-1; i++ {
		guess := strings.Join(parts[i:], ".")
		guesses = append(guesses, guess)
	}
	return guesses
}

func (p *Provider) getDNSRecords(ctx context.Context, zone bunnyZone) ([]bunnyRecord, error) {
	p.log(fmt.Sprintf("fetching all records in zone %s", zone.Domain))

	req, err := http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf("https://api.bunny.net/dnszone/%d", zone.ID), nil)
	if err != nil {
		return nil, err
	}

	data, err := p.doRequest(req)
	if err != nil {
		return nil, err
	}

	result := getAllRecordsResponse{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	p.log(fmt.Sprintf("done fetching %d record(s) in zone %s", len(result.Records), zone.Domain))

	return result.Records, nil
}

func (p *Provider) getAllRecords(ctx context.Context, zone bunnyZone) ([]libdns.Record, error) {
	bunnyRecords, err := p.getDNSRecords(ctx, zone)
	if err != nil {
		return nil, err
	}

	records := []libdns.Record{}
	for _, resData := range bunnyRecords {
		record, err := zone.libdnsRecord(resData)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	p.log(fmt.Sprintf("retrieved %d record(s)", len(records)), records...)

	return records, nil
}

func (p *Provider) createRecord(ctx context.Context, zone bunnyZone, record libdns.Record) (libdns.Record, error) {
	rr := record.RR()

	p.log(fmt.Sprintf("creating %s record in zone %s", rr.Type, zone.Domain), record)

	reqData, err := zone.bunnyRecord(record)
	if err != nil {
		return nil, err
	}

	reqBuffer, err := json.Marshal(reqData)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "PUT",
		fmt.Sprintf("https://api.bunny.net/dnszone/%d/records", zone.ID), bytes.NewBuffer(reqBuffer))
	if err != nil {
		return nil, err
	}

	req.Header.Add("content-type", "application/json")
	data, err := p.doRequest(req)
	if err != nil {
		return nil, err
	}

	result := bunnyRecord{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	record, err = zone.libdnsRecord(result)
	if err != nil {
		return nil, err
	}

	p.log(fmt.Sprintf("done creating %s record %d in zone %s", rr.Type, result.ID, zone.Domain), record)

	return record, nil
}

func (p *Provider) updateRecord(ctx context.Context, zone bunnyZone, record libdns.Record, id int) error {
	rr := record.RR()

	p.log(fmt.Sprintf("updating %s record in zone %s", rr.Type, zone.Domain), record)

	reqData, err := zone.bunnyRecord(record)
	if err != nil {
		return err
	}

	reqBuffer, err := json.Marshal(reqData)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("https://api.bunny.net/dnszone/%d/records/%d", zone.ID, id), bytes.NewBuffer(reqBuffer))
	if err != nil {
		return err
	}

	req.Header.Add("content-type", "application/json")

	_, err = p.doRequest(req)
	if err != nil {
		return err
	}

	p.log(fmt.Sprintf("done updating %s record %s in zone %s", rr.Type, rr.Name, zone.Domain), record)

	return nil
}

// Creates a new record if it does not exist, or updates an existing one.
func (p *Provider) createOrUpdateRecord(ctx context.Context, zone bunnyZone, record libdns.Record) (libdns.Record, error) {
	bunnyRecords, err := p.getDNSRecords(ctx, zone)
	if err != nil {
		return nil, err
	}

	matchingRecords, err := zone.filterBunnyRecords(bunnyRecords, record)
	if err != nil {
		return nil, err
	}
	if len(matchingRecords) == 0 {
		return p.createRecord(ctx, zone, record)
	}
	if len(matchingRecords) > 1 {
		return nil, fmt.Errorf("unexpectedly found more than 1 record for %s in zone %s", record.RR().Name, zone.Domain)
	}
	err = p.updateRecord(ctx, zone, record, matchingRecords[0].ID)
	return record, err
}

func (p *Provider) deleteRecord(ctx context.Context, zone bunnyZone, record libdns.Record) error {
	rr := record.RR()

	p.log(fmt.Sprintf("deleting %s record in zone %s", rr.Type, zone.Domain))

	bunnyRecords, err := p.getDNSRecords(ctx, zone)
	if err != nil {
		return err
	}

	matchingRecords, err := zone.filterBunnyRecords(bunnyRecords, record)
	if err != nil {
		return err
	}

	if len(matchingRecords) == 0 {
		p.log(fmt.Sprintf("no matching record found for %s in zone %s, skipping deletion", rr.Name, zone.Domain))
		return nil
	}

	req, err := http.NewRequestWithContext(ctx, "DELETE",
		fmt.Sprintf("https://api.bunny.net/dnszone/%d/records/%d", zone.ID, matchingRecords[0].ID), nil)
	if err != nil {
		return err
	}

	_, err = p.doRequest(req)
	if err != nil {
		return err
	}

	p.log(fmt.Sprintf("done deleting %s record %d in zone %s", rr.Type, matchingRecords[0].ID, zone.Domain))

	return nil
}

func (p *Provider) log(msg string, records ...libdns.Record) {
	if p.Logger != nil {
		p.Logger(msg, records)
	} else if p.Debug {
		fmt.Printf("[bunny] %s\n", msg)
		for _, record := range records {
			rr := record.RR()
			fmt.Printf("[bunny]   %s: Name=%s, Value=%s TTL=%s\n", rr.Type, rr.Name, rr.Data, rr.TTL)
		}
	}
}

const (
	// The Bunny.net API uses integers to represent record types.
	bunnyTypeA        = 0
	bunnyTypeAAAA     = 1
	bunnyTypeCNAME    = 2
	bunnyTypeTXT      = 3
	bunnyTypeMX       = 4
	bunnyTypeRedirect = 5
	bunnyTypeFlatten  = 6
	bunnyTypePullZone = 7
	bunnyTypeSRV      = 8
	bunnyTypeCAA      = 9
	bunnyTypePTR      = 10
	bunnyTypeScript   = 11
	bunnyTypeNS       = 12
)

// Converts the Bunny.net record type to the libdns record type.
func fromBunnyType(t int) (string, error) {
	switch t {
	case bunnyTypeA:
		return "A", nil
	case bunnyTypeAAAA:
		return "AAAA", nil
	case bunnyTypeCNAME:
		return "CNAME", nil
	case bunnyTypeTXT:
		return "TXT", nil
	case bunnyTypeMX:
		return "MX", nil
	case bunnyTypeRedirect:
		return "Redirect", nil
	case bunnyTypeFlatten:
		return "Flatten", nil
	case bunnyTypePullZone:
		return "PullZone", nil
	case bunnyTypeSRV:
		return "SRV", nil
	case bunnyTypeCAA:
		return "CAA", nil
	case bunnyTypePTR:
		return "PTR", nil
	case bunnyTypeScript:
		return "Script", nil
	case bunnyTypeNS:
		return "NS", nil
	default:
		return "", fmt.Errorf("unknown record type ID: %d", t)
	}
}

// Converts the libdns record type to the Bunny.net record type.
func toBunnyType(t string) (int, error) {
	switch t {
	case "A":
		return bunnyTypeA, nil
	case "AAAA":
		return bunnyTypeAAAA, nil
	case "CNAME":
		return bunnyTypeCNAME, nil
	case "TXT":
		return bunnyTypeTXT, nil
	case "MX":
		return bunnyTypeMX, nil
	case "Redirect":
		return bunnyTypeRedirect, nil
	case "Flatten":
		return bunnyTypeFlatten, nil
	case "PullZone":
		return bunnyTypePullZone, nil
	case "SRV":
		return bunnyTypeSRV, nil
	case "CAA":
		return bunnyTypeCAA, nil
	case "PTR":
		return bunnyTypePTR, nil
	case "Script":
		return bunnyTypeScript, nil
	case "NS":
		return bunnyTypeNS, nil
	default:
		return -1, fmt.Errorf("unknown record type: %s", t)
	}
}

func (zone bunnyZone) bunnyRecord(record libdns.Record) (bunnyRecord, error) {
	rr := record.RR()

	rType, err := toBunnyType(rr.Type)
	if err != nil {
		return bunnyRecord{}, err
	}
	r := bunnyRecord{
		Type:  rType,
		Name:  rr.Name,
		Value: rr.Data,
		TTL:   int(rr.TTL.Seconds()),
	}
	if r.Name == "@" {
		r.Name = ""
	}
	if zone.nameBase != "" {
		if r.Name == "" {
			r.Name = zone.nameBase
		} else {
			r.Name = fmt.Sprintf("%s.%s", r.Name, zone.nameBase)
		}
	}
	switch rec := record.(type) {
	case libdns.CAA:
		r.Flags = int(rec.Flags)
		r.Tag = rec.Tag
		r.Value = rec.Value
	case libdns.MX:
		r.Priority = int32(rec.Preference)
		r.Value = rec.Target
	case libdns.SRV:
		r.Priority = int32(rec.Priority)
		r.Weight = int32(rec.Weight)
		r.Port = int32(rec.Port)
		r.Value = rec.Target
	}
	return r, nil
}

func (zone bunnyZone) libdnsRecord(record bunnyRecord) (libdns.Record, error) {
	rType, err := fromBunnyType(record.Type)
	if err != nil {
		return nil, err
	}
	r := libdns.RR{
		Type: rType,
		Name: record.Name,
		Data: record.Value,
		TTL:  time.Duration(record.TTL) * time.Second,
	}

	if zone.nameBase != "" {
		if r.Name == zone.nameBase {
			r.Name = ""
		} else if strings.HasSuffix(r.Name, "."+zone.nameBase) {
			r.Name = strings.TrimSuffix(r.Name, "."+zone.nameBase)
		}
	}
	if r.Name == "" {
		r.Name = "@"
	}
	switch r.Type {
	// Types that are compatible with RR.Parse()
	case "A", "AAAA", "CNAME", "NS", "TXT":
		return r.Parse()
	case "CAA":
		return libdns.CAA{
			Name:  r.Name,
			TTL:   r.TTL,
			Flags: uint8(record.Flags),
			Tag:   record.Tag,
			Value: record.Value,
		}, nil
	case "MX":
		return libdns.MX{
			Name:       r.Name,
			TTL:        r.TTL,
			Preference: uint16(record.Priority),
			Target:     record.Value,
		}, nil
	case "SRV":
		parts := strings.SplitN(r.Name, ".", 3)
		if len(parts) < 2 {
			return libdns.SRV{}, fmt.Errorf("name %v does not contain enough fields; expected format: '_service._proto.name' or '_service._proto'", r.Name)
		}
		name := "@"
		if len(parts) == 3 {
			name = parts[2]
		}
		return libdns.SRV{
			Service:   strings.TrimPrefix(parts[0], "_"),
			Transport: strings.TrimPrefix(parts[1], "_"),
			Name:      name,
			TTL:       r.TTL,
			Priority:  uint16(record.Priority),
			Weight:    uint16(record.Weight),
			Port:      uint16(record.Port),
			Target:    record.Value,
		}, nil
	default:
		return r, nil
	}
}

func (zone bunnyZone) filterBunnyRecords(haystack []bunnyRecord, record libdns.Record) ([]bunnyRecord, error) {
	needle, err := zone.bunnyRecord(record)
	if err != nil {
		return nil, err
	}
	records := []bunnyRecord{}
	for _, r := range haystack {
		if r.Name == needle.Name && r.Type == needle.Type {
			records = append(records, r)
		}
	}
	return records, nil
}
