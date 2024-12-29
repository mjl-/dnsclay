package dnspod

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/libdns/libdns"
	d "github.com/nrdcg/dnspod-go"
)

// Client ...
type Client struct {
	client     *d.Client
	mutex      sync.Mutex
	domainList []d.Domain
}

func (p *Provider) getClient() error {
	if p.client == nil {
		params := d.CommonParams{LoginToken: p.APIToken, Format: "json"}
		p.client = d.NewClient(params)
	}

	return nil
}
func (p *Provider) getDomains() ([]d.Domain, error) {
	if len(p.domainList) > 0 {
		return p.domainList, nil
	}
	domains, _, err := p.client.Domains.List()
	if nil != err {
		return p.domainList, err
	}
	p.domainList = domains
	return p.domainList, nil
}
func (p *Provider) getDomainIDByDomainName(domainName string) (string, error) {
	domains, err := p.getDomains()
	if nil != err {
		return "", err
	}
	domainName = strings.Trim(domainName, ".")
	for _, domain := range domains {
		if domain.Name == domainName {
			return string(domain.ID), nil
		}
	}
	return "", fmt.Errorf("Domain %s not found in your dnspod account", domainName)
}

func (p *Provider) getDNSEntries(ctx context.Context, zone string) ([]libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.getClient()

	var records []libdns.Record
	domainID, err := p.getDomainIDByDomainName(zone)
	if nil != err {
		//debug
		// fmt.Printf("%s, %s", zone, err.Error())
		return records, fmt.Errorf("Get records err.Zone:%s, Error:%s", zone, err.Error())
	}
	//todo now can only return 100 records
	reqRecords, _, err := p.client.Records.List(string(domainID), "")
	if err != nil {
		// fmt.Printf("%s, %s", zone, err.Error())
		return records, fmt.Errorf("Get records err.Zone:%s, Error:%s", zone, err.Error())
	}

	for _, entry := range reqRecords {
		ttl, _ := strconv.ParseInt(entry.TTL, 10, 64)
		record := libdns.Record{
			Name:  entry.Name + "." + strings.Trim(zone, ".") + ".",
			Value: entry.Value,
			Type:  entry.Type,
			TTL:   time.Duration(ttl) * time.Second,
			ID:    entry.ID,
		}
		records = append(records, record)
	}

	return records, nil
}

func extractRecordName(name string, zone string) string {
	if idx := strings.Index(name, "."+strings.Trim(zone, ".")); idx != -1 {
		return name[:idx]
	}
	return name
}

func (p *Provider) addDNSEntry(ctx context.Context, zone string, record libdns.Record) (libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.getClient()

	entry := d.Record{
		Name:  extractRecordName(record.Name, zone),
		Value: record.Value,
		Type:  record.Type,
		Line:  "默认",
		TTL:   strconv.Itoa(int(record.TTL.Seconds())),
	}
	domainID, err := p.getDomainIDByDomainName(zone)
	if nil != err {
		// fmt.Printf("%s, %s, %s, %s, %v", zone, entry.Name, entry.Value, err.Error(), record)
		return record, fmt.Errorf("Create record err.Zone:%s, Name: %s, Value: %s, Error:%s, %v", zone, entry.Name, entry.Value, err.Error(), record)
	}
	rec, _, err := p.client.Records.Create(domainID, entry)
	if err != nil {
		// fmt.Printf("%s, %s, %s, %s, %v", zone, entry.Name, entry.Value, err.Error(), record)
		return record, fmt.Errorf("Create record err.Zone:%s, Name: %s, Value: %s, Error:%s, %v", zone, entry.Name, entry.Value, err.Error(), record)
	}
	record.ID = rec.ID

	return record, nil
}

func (p *Provider) removeDNSEntry(ctx context.Context, zone string, record libdns.Record) (libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.getClient()

	domainID, err := p.getDomainIDByDomainName(zone)
	if nil != err {
		// fmt.Printf("%s, %s, %s, %s, %v", zone, record.Name, record.Value, err.Error(), record)
		return record, fmt.Errorf("Remove record err.Zone:%s, Name: %s, Value: %s, Error:%s", zone, record.Name, record.Value, err.Error())
	}
	_, err = p.client.Records.Delete(domainID, record.ID)
	if err != nil {
		// fmt.Printf("%s, %s, %s, %s, %v", zone, record.Name, record.Value, err.Error(), record)
		return record, fmt.Errorf("Remove record err.Zone:%s, Name: %s, Value: %s, Error:%s", zone, record.Name, record.Value, err.Error())
	}

	return record, nil
}

func (p *Provider) updateDNSEntry(ctx context.Context, zone string, record libdns.Record) (libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.getClient()

	entry := d.Record{
		Name:  extractRecordName(record.Name, zone),
		Value: record.Value,
		Type:  record.Type,
		Line:  "默认",
		TTL:   strconv.Itoa(int(record.TTL.Seconds())),
	}
	domainID, err := p.getDomainIDByDomainName(zone)
	if nil != err {
		// fmt.Printf("%s, %s, %s, %s, %v", zone, entry.Name, entry.Value, err.Error(), record)
		return record, fmt.Errorf("Update record err.Zone:%s, Name: %s, Value: %s, Error:%s, %v", zone, entry.Name, entry.Value, err.Error(), record)
	}
	_, _, err = p.client.Records.Update(domainID, record.ID, entry)
	if err != nil {
		// fmt.Printf("%s, %s, %s, %s, %v", zone, entry.Name, entry.Value, err.Error(), record)
		return record, fmt.Errorf("Update record err.Zone:%s, Name: %s, Value: %s, Error:%s, %v", zone, entry.Name, entry.Value, err.Error(), record)
	}

	return record, nil
}
