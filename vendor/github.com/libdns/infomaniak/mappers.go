package infomaniak

import (
	"time"

	"github.com/libdns/libdns"
)

// Default priority applied by infomaniak
const defaultPriority = 10

// Default TTL that is applied if none is provided - infomaniak requires a TTL
const defaultTtlSecs = 300

// ToLibDnsRecord maps a infomaniak dns record to a libdns record
func (ikr *IkRecord) ToLibDnsRecord(zone string) libdns.Record {
	return libdns.Record{
		ID:       ikr.ID,
		Type:     ikr.Type,
		Name:     libdns.RelativeName(ikr.SourceIdn, zone),
		Value:    ikr.Target,
		TTL:      time.Duration(ikr.TtlInSec),
		Priority: ikr.Priority,
	}
}

// ToInfomaniakRecord maps a libdns record to a infomaniak dns record
func ToInfomaniakRecord(libdnsRec *libdns.Record, zone string) IkRecord {
	ikRec := IkRecord{
		ID:        libdnsRec.ID,
		Type:      libdnsRec.Type,
		SourceIdn: libdns.AbsoluteName(libdnsRec.Name, zone),
		Target:    libdnsRec.Value,
		TtlInSec:  uint(libdnsRec.TTL),
		Priority:  libdnsRec.Priority,
	}

	if ikRec.TtlInSec <= 0 {
		ikRec.TtlInSec = defaultTtlSecs
	}

	if ikRec.Priority <= 0 {
		ikRec.Priority = defaultPriority
	}

	return ikRec
}
