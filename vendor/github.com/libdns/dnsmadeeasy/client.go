package dnsmadeeasy

import (
	"context"
	"fmt"
	"strconv"
	"time"

	dme "github.com/john-k/dnsmadeeasy"
	"github.com/libdns/libdns"
)

func (p *Provider) init(ctx context.Context) {
	p.once.Do(func() {
		p.client = *dme.GetClient(
			p.APIKey,
			p.SecretKey,
			p.APIEndpoint,
		)
	})
}

func recordFromDmeRecord(dmeRecord dme.Record) libdns.Record {
	var rec libdns.Record
	rec.ID = fmt.Sprint(dmeRecord.ID)
	rec.Type = dmeRecord.Type
	rec.Name = dmeRecord.Name
	rec.Value = dmeRecord.Value
	rec.TTL = time.Duration(dmeRecord.Ttl)

	// TODO: enable support for SRV weight field and embedding
	// "<port> <target>" in value when libdns releases support
	if dmeRecord.Type == "MX" {
		rec.Priority = uint(dmeRecord.MxLevel)
	} else if dmeRecord.Type == "SRV" {
		rec.Priority = uint(dmeRecord.Priority)
		//rec.Weight = dmeRecord.Weight
		//rec.Value = fmt.Sprintf("%d %s", dmeRecord.Port, dmeRecord.Value)
	}

	return rec
}

func dmeRecordFromRecord(record libdns.Record) (dme.Record, error) {
	var dmeRecord dme.Record
	var id int
	var err error
	// Since dmeRecord.ID is set to `json:"id,omitempty"`, this properly preserves empty values
	if record.ID == "" {
		id = 0
	} else {
		id, err = strconv.Atoi(record.ID)
		if err != nil {
			return dme.Record{}, err
		}
	}
	dmeRecord.ID = id
	dmeRecord.Name = record.Name
	dmeRecord.Type = record.Type
	dmeRecord.Value = record.Value
	dmeRecord.Ttl = int(record.TTL)
	// DNSMadeEasy fails to accept zero TTL, so use a default value
	if dmeRecord.Ttl == 0 {
		dmeRecord.Ttl = 120
	}
	// Likewise, DNSMadeEasy doesn't accept a blank GtdLocation
	dmeRecord.GtdLocation = "DEFAULT"
	if record.Type == "MX" {
		dmeRecord.MxLevel = int(record.Priority)
	} else if record.Type == "SRV" {
		dmeRecord.Priority = int(record.Priority)
		/*
			// TODO: enable support for SRV weight field and extracting
			// "<port> <target>" from value when libdns releases support
			dmeRecord.Weight = record.Weight
			fields := strings.Fields(record.Value)
			if len(fields) != 2 {
				return dme.Record{}, fmt.Errorf("malformed SRV value '%s'; expected: '<port> <target>'", record.Value)
			}

			port, err := strconv.Atoi(fields[0])
			if err != nil {
				return dme.Record{}, fmt.Errorf("invalid port %s: %v", fields[0], err)
			}
			if port < 0 {
				return dme.Record{}, fmt.Errorf("port cannot be < 0: %d", port)
			}
			dmeRecord.Port = port
			dmeRecord.Value = fields[1]
		*/
	}
	return dmeRecord, nil

}
